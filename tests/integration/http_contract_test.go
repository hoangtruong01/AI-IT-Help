package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/middleware"
	"eomp/packages/shared/pkg/response"
)

// newContractJWTManager returns a configured JWT manager for the in-process
// HTTP contract harness. It does not call the deployed auth service.
func newContractJWTManager() *auth.JWTManager {
	return auth.NewJWTManager("eomp-test-secret-at-least-32-chars-long-for-jwt-2026", 1*time.Hour, 24*time.Hour)
}

// stripIdentityHeadersMiddleware mirrors the outermost gateway behavior.
func stripIdentityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for header := range r.Header {
			if strings.HasPrefix(strings.ToLower(header), "x-user-") || strings.EqualFold(header, "X-Department-ID") {
				r.Header.Del(header)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// gatewayAuthMiddleware is a test-only approximation of gateway authentication.
func gatewayAuthMiddleware(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization token"}`, http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}
			claims, err := jwtManager.ValidateToken(parts[1])
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Role", claims.Role)
			r.Header.Set("X-User-Department", claims.DepartmentID)
			r.Header.Set("X-User-Name", claims.FullName)
			next.ServeHTTP(w, r)
		})
	}
}

// buildHTTPContractServer sets up an in-process contract harness. The handlers
// and ticket store below are test doubles; no EOMP service or database starts.
func buildHTTPContractServer(jwtManager *auth.JWTManager) *httptest.Server {
	mux := http.NewServeMux()
	authFilter := gatewayAuthMiddleware(jwtManager)
	adminOnly := middleware.RequireRole("ROLE_ADMIN")
	operatorWrites := middleware.RequireRolesForMethods(
		[]string{http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete},
		"ROLE_ADMIN", "ROLE_MANAGER", "ROLE_AGENT",
	)

	// Simulated ticket store for HTTP contract testing
	var ticketSeq int64
	var mu sync.RWMutex
	type Ticket struct {
		ID           string    `json:"id"`
		TicketNumber string    `json:"ticket_number"`
		Title        string    `json:"title"`
		Category     string    `json:"category"`
		Priority     string    `json:"priority"`
		Status       string    `json:"status"`
		RequesterID  string    `json:"requester_id"`
		DepartmentID string    `json:"department_id"`
		AssigneeID   *string   `json:"assignee_id"`
		Version      int       `json:"version"`
		CreatedAt    time.Time `json:"created_at"`
	}
	tickets := make(map[string]*Ticket)

	// Public Auth routes
	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}
		role := "ROLE_EMPLOYEE"
		dept := "dept-eng"
		name := "Employee User"
		if strings.Contains(req.Email, "agent") {
			role = "ROLE_AGENT"
			dept = "dept-it"
			name = "Agent User"
		} else if strings.Contains(req.Email, "manager") {
			role = "ROLE_MANAGER"
			dept = "dept-eng"
			name = "Manager User"
		} else if strings.Contains(req.Email, "admin") {
			role = "ROLE_ADMIN"
			dept = "dept-sec"
			name = "Admin User"
		}
		acc, ref, _ := jwtManager.GenerateTokenPair("u-"+req.Email, req.Email, role, dept, name)
		response.JSON(w, http.StatusOK, map[string]string{
			"access_token":  acc,
			"refresh_token": ref,
		})
	})

	// User administration (Admin only)
	mux.Handle("POST /api/v1/users", authFilter(adminOnly(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusCreated, map[string]string{"status": "user_created"})
	}))))

	// Helpdesk Ticket creation (Employee / authenticated)
	mux.Handle("POST /api/v1/tickets", authFilter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actor := middleware.GetActor(r.Context())
		if actor.Role == "" {
			actor.ID = r.Header.Get("X-User-ID")
			actor.Role = r.Header.Get("X-User-Role")
			actor.DepartmentID = r.Header.Get("X-User-Department")
		}

		var req struct {
			Title    string `json:"title"`
			Category string `json:"category"`
			Priority string `json:"priority"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid body"}`, http.StatusBadRequest)
			return
		}

		num := atomic.AddInt64(&ticketSeq, 1)
		tID := fmt.Sprintf("tk-%d", num)
		tk := &Ticket{
			ID:           tID,
			TicketNumber: fmt.Sprintf("TK-%04d", num),
			Title:        req.Title,
			Category:     req.Category,
			Priority:     req.Priority,
			Status:       "OPEN",
			RequesterID:  actor.ID,
			DepartmentID: actor.DepartmentID,
			Version:      1,
			CreatedAt:    time.Now(),
		}
		mu.Lock()
		tickets[tID] = tk
		mu.Unlock()

		response.JSON(w, http.StatusCreated, tk)
	})))

	// Helpdesk Ticket list (Row-Level Authorization)
	mux.Handle("GET /api/v1/tickets", authFilter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actorRole := r.Header.Get("X-User-Role")
		actorID := r.Header.Get("X-User-ID")
		actorDept := r.Header.Get("X-User-Department")

		mu.RLock()
		defer mu.RUnlock()

		result := make([]*Ticket, 0)
		for _, tk := range tickets {
			switch actorRole {
			case "ROLE_EMPLOYEE":
				if tk.RequesterID == actorID {
					result = append(result, tk)
				}
			case "ROLE_AGENT":
				if tk.AssigneeID == nil || *tk.AssigneeID == "" || *tk.AssigneeID == actorID {
					result = append(result, tk)
				}
			case "ROLE_MANAGER":
				if tk.DepartmentID == actorDept {
					result = append(result, tk)
				}
			case "ROLE_ADMIN":
				result = append(result, tk)
			}
		}
		response.JSON(w, http.StatusOK, map[string]any{"tickets": result, "total": len(result)})
	})))

	// Helpdesk Ticket get (Anti-enumeration 404)
	mux.Handle("GET /api/v1/tickets/{id}", authFilter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tID := r.PathValue("id")
		actorRole := r.Header.Get("X-User-Role")
		actorID := r.Header.Get("X-User-ID")
		actorDept := r.Header.Get("X-User-Department")

		mu.RLock()
		tk, exists := tickets[tID]
		mu.RUnlock()

		if !exists {
			http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
			return
		}

		// Row-level scope verification: return 404 for out of scope to prevent enumeration
		allowed := false
		switch actorRole {
		case "ROLE_EMPLOYEE":
			allowed = (tk.RequesterID == actorID)
		case "ROLE_AGENT":
			allowed = (tk.AssigneeID == nil || *tk.AssigneeID == "" || *tk.AssigneeID == actorID)
		case "ROLE_MANAGER":
			allowed = (tk.DepartmentID == actorDept)
		case "ROLE_ADMIN":
			allowed = true
		}

		if !allowed {
			http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
			return
		}

		response.JSON(w, http.StatusOK, tk)
	})))

	// Helpdesk Ticket assignment
	mux.Handle("PUT /api/v1/tickets/{id}/assign", authFilter(operatorWrites(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tID := r.PathValue("id")
		var req struct {
			AssigneeID string `json:"assignee_id"`
			Version    int    `json:"version"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		mu.Lock()
		defer mu.Unlock()
		tk, exists := tickets[tID]
		if !exists {
			http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
			return
		}
		if tk.Version != req.Version {
			http.Error(w, `{"error":"optimistic locking conflict"}`, http.StatusConflict)
			return
		}
		tk.AssigneeID = &req.AssigneeID
		tk.Version++
		response.JSON(w, http.StatusOK, tk)
	}))))

	// Helpdesk Ticket comments
	mux.Handle("POST /api/v1/tickets/{id}/comments", authFilter(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response.JSON(w, http.StatusCreated, map[string]string{"status": "comment_added"})
	})))

	// Helpdesk Ticket status update
	mux.Handle("PUT /api/v1/tickets/{id}/status", authFilter(operatorWrites(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tID := r.PathValue("id")
		var req struct {
			Status  string `json:"status"`
			Version int    `json:"version"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		mu.Lock()
		defer mu.Unlock()
		tk, exists := tickets[tID]
		if !exists {
			http.Error(w, `{"error":"ticket not found"}`, http.StatusNotFound)
			return
		}
		tk.Status = req.Status
		tk.Version++
		response.JSON(w, http.StatusOK, tk)
	}))))

	// Wrap in outermost StripIdentityHeaders
	return httptest.NewServer(stripIdentityHeadersMiddleware(mux))
}

// TestHTTPContract_FullOperationsLifecycle validates the expected HTTP status
// and payload transitions against the in-process test doubles.
func TestHTTPContract_FullOperationsLifecycle(t *testing.T) {
	jwtManager := newContractJWTManager()
	server := buildHTTPContractServer(jwtManager)
	defer server.Close()
	client := server.Client()

	// 1. Authenticate Employee, Agent, Manager
	empToken, _, _ := jwtManager.GenerateTokenPair("u-emp-1", "emp1@eomp.local", "ROLE_EMPLOYEE", "dept-eng", "Employee One")
	agentToken, _, _ := jwtManager.GenerateTokenPair("u-agent-1", "agent1@eomp.local", "ROLE_AGENT", "dept-it", "Agent One")
	mgrToken, _, _ := jwtManager.GenerateTokenPair("u-mgr-1", "mgr1@eomp.local", "ROLE_MANAGER", "dept-eng", "Manager One")

	// 2. Employee submits ticket
	ticketBody, _ := json.Marshal(map[string]any{
		"title":    "Production Docker Engine Crash",
		"category": "Infrastructure",
		"priority": "HIGH",
	})
	req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/tickets", bytes.NewReader(ticketBody))
	req.Header.Set("Authorization", "Bearer "+empToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create ticket HTTP call failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created on ticket create, got %d", resp.StatusCode)
	}

	var created struct {
		ID           string `json:"id"`
		TicketNumber string `json:"ticket_number"`
		Version      int    `json:"version"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&created)
	if created.ID == "" || created.TicketNumber == "" {
		t.Fatalf("invalid ticket response: %+v", created)
	}
	t.Logf("  [+] Step 1: Ticket created: ID=%s, Number=%s", created.ID, created.TicketNumber)

	// 3. Agent lists unassigned ticket queue
	listReq, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/tickets", nil)
	listReq.Header.Set("Authorization", "Bearer "+agentToken)
	listResp, err := client.Do(listReq)
	if err != nil || listResp.StatusCode != http.StatusOK {
		t.Fatalf("agent list failed: code=%d, err=%v", listResp.StatusCode, err)
	}
	listResp.Body.Close()
	t.Log("  [+] Step 2: Agent queried unassigned queue successfully")

	// 4. Agent assigns ticket to self
	assignBody, _ := json.Marshal(map[string]any{
		"assignee_id": "u-agent-1",
		"version":     created.Version,
	})
	assignReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/tickets/%s/assign", server.URL, created.ID), bytes.NewReader(assignBody))
	assignReq.Header.Set("Authorization", "Bearer "+agentToken)
	assignReq.Header.Set("Content-Type", "application/json")

	assignResp, err := client.Do(assignReq)
	if err != nil || assignResp.StatusCode != http.StatusOK {
		t.Fatalf("agent assign failed: code=%d, err=%v", assignResp.StatusCode, err)
	}
	assignResp.Body.Close()
	t.Log("  [+] Step 3: Agent assigned ticket to self successfully")

	// 5. Agent adds comment
	commentBody, _ := json.Marshal(map[string]any{"content": "Investigating daemon crash"})
	commentReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/v1/tickets/%s/comments", server.URL, created.ID), bytes.NewReader(commentBody))
	commentReq.Header.Set("Authorization", "Bearer "+agentToken)
	commentReq.Header.Set("Content-Type", "application/json")

	commentResp, err := client.Do(commentReq)
	if err != nil || commentResp.StatusCode != http.StatusCreated {
		t.Fatalf("add comment failed: code=%d, err=%v", commentResp.StatusCode, err)
	}
	commentResp.Body.Close()
	t.Log("  [+] Step 4: Progress comment added successfully")

	// 6. Agent marks ticket as RESOLVED
	statusBody, _ := json.Marshal(map[string]any{
		"status":  "RESOLVED",
		"version": created.Version + 1,
	})
	statusReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/v1/tickets/%s/status", server.URL, created.ID), bytes.NewReader(statusBody))
	statusReq.Header.Set("Authorization", "Bearer "+agentToken)
	statusReq.Header.Set("Content-Type", "application/json")

	statusResp, err := client.Do(statusReq)
	if err != nil || statusResp.StatusCode != http.StatusOK {
		t.Fatalf("status update failed: code=%d, err=%v", statusResp.StatusCode, err)
	}
	statusResp.Body.Close()
	t.Log("  [+] Step 5: Ticket marked RESOLVED successfully")

	// 7. Manager views ticket
	mgrReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/tickets/%s", server.URL, created.ID), nil)
	mgrReq.Header.Set("Authorization", "Bearer "+mgrToken)
	mgrResp, err := client.Do(mgrReq)
	if err != nil || mgrResp.StatusCode != http.StatusOK {
		t.Fatalf("manager get ticket failed: code=%d, err=%v", mgrResp.StatusCode, err)
	}
	mgrResp.Body.Close()
	t.Log("  [+] Step 6: Department Manager verified ticket resolution successfully")
}

// TestHTTPContract_HeaderSpoofingDefense validates the modeled gateway contract.
// Production acceptance still requires the same probes against deployed ingress.
func TestHTTPContract_HeaderSpoofingDefense(t *testing.T) {
	jwtManager := newContractJWTManager()
	server := buildHTTPContractServer(jwtManager)
	defer server.Close()
	client := server.Client()

	// 1. Spoofed header without token -> 401 Unauthorized
	req1, _ := http.NewRequest(http.MethodGet, server.URL+"/api/v1/tickets", nil)
	req1.Header.Set("X-User-Role", "ROLE_ADMIN")
	req1.Header.Set("X-User-ID", "attacker-uuid")

	resp1, err := client.Do(req1)
	if err != nil || resp1.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for unauthenticated spoofed header, got %d", resp1.StatusCode)
	}
	resp1.Body.Close()
	t.Log("  [+] Unauthenticated header spoofing successfully blocked (401)")

	// 2. Employee token + spoofed ROLE_ADMIN header -> Gateway strips header, injects ROLE_EMPLOYEE -> 403 Forbidden
	empToken, _, _ := jwtManager.GenerateTokenPair("u-emp-1", "emp@eomp.local", "ROLE_EMPLOYEE", "dept-eng", "Employee User")
	req2, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/users", bytes.NewReader([]byte(`{"email":"new@eomp.local"}`)))
	req2.Header.Set("Authorization", "Bearer "+empToken)
	req2.Header.Set("X-User-Role", "ROLE_ADMIN")
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := client.Do(req2)
	if err != nil || resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden on privilege escalation attempt, got %d", resp2.StatusCode)
	}
	resp2.Body.Close()
	t.Log("  [+] Privilege escalation attempt blocked (403 Forbidden)")
}

// TestHTTPContract_CrossTenantIsolationAndAntiEnumeration validates the modeled
// row-scope contract. Deployed-stack E2E evidence is tracked separately.
func TestHTTPContract_CrossTenantIsolationAndAntiEnumeration(t *testing.T) {
	jwtManager := newContractJWTManager()
	server := buildHTTPContractServer(jwtManager)
	defer server.Close()
	client := server.Client()

	emp1Token, _, _ := jwtManager.GenerateTokenPair("u-emp-alice", "alice@eomp.local", "ROLE_EMPLOYEE", "dept-sales", "Alice Sales")
	emp2Token, _, _ := jwtManager.GenerateTokenPair("u-emp-bob", "bob@eomp.local", "ROLE_EMPLOYEE", "dept-hr", "Bob HR")

	// Alice creates a ticket
	ticketBody, _ := json.Marshal(map[string]any{
		"title":    "Confidential Sales Proposal Issue",
		"category": "Billing",
		"priority": "HIGH",
	})
	createReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/v1/tickets", bytes.NewReader(ticketBody))
	createReq.Header.Set("Authorization", "Bearer "+emp1Token)
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := client.Do(createReq)
	if err != nil || createResp.StatusCode != http.StatusCreated {
		t.Fatalf("Alice ticket create failed: %v", err)
	}
	var aliceTicket struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(createResp.Body).Decode(&aliceTicket)
	createResp.Body.Close()

	// Bob probes Alice's ticket ID directly
	bobProbeReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/v1/tickets/%s", server.URL, aliceTicket.ID), nil)
	bobProbeReq.Header.Set("Authorization", "Bearer "+emp2Token)

	bobResp, err := client.Do(bobProbeReq)
	if err != nil || bobResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404 Not Found on cross-tenant probe, got %d", bobResp.StatusCode)
	}
	bobResp.Body.Close()
	t.Log("  [+] Anti-enumeration verified: cross-employee probe returned 404 Not Found")
}
