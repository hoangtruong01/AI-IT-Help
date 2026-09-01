package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eomp/packages/shared/pkg/requestctx"
)

func TestRequestLoggerAddsRequestIDToContextAndResponse(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := RequestLogger(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := requestctx.RequestID(r.Context()); got != "client-request-id" {
			t.Fatalf("expected request ID in context, got %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("X-Request-ID", "client-request-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Request-ID"); got != "client-request-id" {
		t.Fatalf("expected response request ID, got %q", got)
	}
}

// Test Case 10.1: Strict RBAC enforcement.
func TestRBAC_RequireRoles(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ACCESS_GRANTED"))
	})

	rbacAdminManager := RequireRoles("ROLE_ADMIN", "ROLE_MANAGER")(dummyHandler)

	// 1. User with ROLE_EMPLOYEE tries to access Admin route -> 403 Forbidden
	reqUser := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqUser.Header.Set("X-User-Role", "ROLE_EMPLOYEE")
	recUser := httptest.NewRecorder()

	rbacAdminManager.ServeHTTP(recUser, reqUser)

	if recUser.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for ROLE_EMPLOYEE, got %d", recUser.Code)
	}

	// 2. User with ROLE_ADMIN -> 200 OK
	reqAdmin := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqAdmin.Header.Set("X-User-Role", "ROLE_ADMIN")
	recAdmin := httptest.NewRecorder()

	rbacAdminManager.ServeHTTP(recAdmin, reqAdmin)

	if recAdmin.Code != http.StatusOK {
		t.Errorf("expected 200 OK for ROLE_ADMIN, got %d", recAdmin.Code)
	}

	// 3. User with ROLE_MANAGER -> 200 OK
	reqManager := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqManager.Header.Set("X-User-Role", "ROLE_MANAGER")
	recManager := httptest.NewRecorder()

	rbacAdminManager.ServeHTTP(recManager, reqManager)

	if recManager.Code != http.StatusOK {
		t.Errorf("expected 200 OK for ROLE_MANAGER, got %d", recManager.Code)
	}
}

func TestRBAC_RequireRolesForMethods(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := RequireRolesForMethods([]string{http.MethodPost, http.MethodPatch}, "ROLE_MANAGER")(next)

	read := httptest.NewRecorder()
	handler.ServeHTTP(read, httptest.NewRequest(http.MethodGet, "/resource", nil))
	if read.Code != http.StatusOK {
		t.Fatalf("read request should remain available to authenticated callers, got %d", read.Code)
	}

	writeReq := httptest.NewRequest(http.MethodPost, "/resource", nil)
	writeReq.Header.Set("X-User-Role", "ROLE_EMPLOYEE")
	write := httptest.NewRecorder()
	handler.ServeHTTP(write, writeReq)
	if write.Code != http.StatusForbidden {
		t.Fatalf("employee write should be forbidden, got %d", write.Code)
	}
}

func TestRequireValidActor(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := ExtractGatewayHeaders()(RequireValidActor()(next))

	tests := []struct {
		name       string
		userID     string
		role       string
		wantStatus int
	}{
		{name: "missing identity", wantStatus: http.StatusUnauthorized},
		{name: "role only", role: "ROLE_AGENT", wantStatus: http.StatusUnauthorized},
		{name: "id only", userID: "agent-1", wantStatus: http.StatusUnauthorized},
		{name: "unknown role", userID: "agent-1", role: "ROLE_ROOT", wantStatus: http.StatusUnauthorized},
		{name: "valid actor", userID: "agent-1", role: "ROLE_AGENT", wantStatus: http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/v1/assets", nil)
			if tt.userID != "" {
				req.Header.Set("X-User-ID", tt.userID)
			}
			if tt.role != "" {
				req.Header.Set("X-User-Role", tt.role)
			}
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

// Test Case 10.2: IP Rate Limiting.
func TestRateLimiter_IPLimit(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Limit to 5 requests per second
	limiter := IPRateLimiter(5, 1*time.Second)(dummyHandler)

	// Send 5 valid requests from trusted loopback
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "127.0.0.1:45123"
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		rec := httptest.NewRecorder()
		limiter.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request #%d should pass, got status %d", i+1, rec.Code)
		}
	}

	// 6th request should be blocked with 429 Too Many Requests
	req6 := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req6.RemoteAddr = "127.0.0.1:45123"
	req6.Header.Set("X-Forwarded-For", "192.168.1.100")
	rec6 := httptest.NewRecorder()

	limiter.ServeHTTP(rec6, req6)

	if rec6.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests on rate limit breach, got %d", rec6.Code)
	}
	if rec6.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After: 60 header, got %s", rec6.Header().Get("Retry-After"))
	}
}

// Test Phase 1 Task 1.3: Dynamic CORS Whitelist Protection
func TestDynamicCORS(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	whitelist := []string{"http://localhost:3000", "https://app.eomp.local"}
	corsMiddleware := DynamicCORS(whitelist)(dummyHandler)

	// 1. Whitelisted Origin
	reqAllowed := httptest.NewRequest("GET", "/api/v1/tickets", nil)
	reqAllowed.Header.Set("Origin", "http://localhost:3000")
	recAllowed := httptest.NewRecorder()

	corsMiddleware.ServeHTTP(recAllowed, reqAllowed)

	if recAllowed.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin: http://localhost:3000, got '%s'", recAllowed.Header().Get("Access-Control-Allow-Origin"))
	}
	if recAllowed.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials: true")
	}

	// 2. Untrusted / Attacker Origin
	reqBlocked := httptest.NewRequest("GET", "/api/v1/tickets", nil)
	reqBlocked.Header.Set("Origin", "http://evil-phishing-site.com")
	recBlocked := httptest.NewRecorder()

	corsMiddleware.ServeHTTP(recBlocked, reqBlocked)

	if recBlocked.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected no Access-Control-Allow-Origin for untrusted origin, got '%s'", recBlocked.Header().Get("Access-Control-Allow-Origin"))
	}

	// 3. Preflight OPTIONS request
	reqOptions := httptest.NewRequest("OPTIONS", "/api/v1/tickets", nil)
	reqOptions.Header.Set("Origin", "http://localhost:3000")
	recOptions := httptest.NewRecorder()

	corsMiddleware.ServeHTTP(recOptions, reqOptions)

	if recOptions.Code != http.StatusNoContent {
		t.Errorf("expected 204 No Content for OPTIONS preflight, got %d", recOptions.Code)
	}
	if recOptions.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("expected Access-Control-Allow-Methods header present on preflight")
	}
}

// Test Phase 1 Task 1.4: Anti-Spoofing Client IP Extraction
func TestAntiSpoofingClientIP(t *testing.T) {
	trustedProxies := []string{"10.0.0.1", "127.0.0.1"}

	// 1. Direct request from untrusted IP trying to forge X-Forwarded-For
	reqUntrusted := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	reqUntrusted.RemoteAddr = "203.0.113.50:54321"        // External untrusted caller
	reqUntrusted.Header.Set("X-Forwarded-For", "8.8.8.8") // Forged header

	extractedIP := ExtractClientIP(reqUntrusted, trustedProxies)
	if extractedIP != "203.0.113.50" {
		t.Errorf("anti-spoofing failed: expected remote host '203.0.113.50', got forged '%s'", extractedIP)
	}

	// 2. Request through a trusted proxy
	reqTrusted := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	reqTrusted.RemoteAddr = "10.0.0.1:45678" // Internal trusted load balancer
	reqTrusted.Header.Set("X-Forwarded-For", "198.51.100.42, 10.0.0.1")

	trustedExtractedIP := ExtractClientIP(reqTrusted, trustedProxies)
	if trustedExtractedIP != "198.51.100.42" {
		t.Errorf("expected trusted client IP '198.51.100.42', got '%s'", trustedExtractedIP)
	}
}

// Test Phase 1 Task 1.5: Strict Rate Limiting on Auth Endpoints (10 req/min)
func TestStrictAuthRateLimiter(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("LOGIN_ATTEMPT_PROCESSED"))
	})

	// 10 req / 1 second window for test
	authLimiter := StrictAuthRateLimiter(10, 1*time.Second)(dummyHandler)

	// Send 10 login requests
	for i := 1; i <= 10; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.RemoteAddr = "192.168.1.55:12345"
		rec := httptest.NewRecorder()
		authLimiter.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request #%d should pass, got %d", i, rec.Code)
		}
	}

	// 11th login attempt within window should be BLOCKED with 429
	req11 := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req11.RemoteAddr = "192.168.1.55:12345"
	rec11 := httptest.NewRecorder()

	authLimiter.ServeHTTP(rec11, req11)

	if rec11.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests for 11th login attempt, got %d", rec11.Code)
	}
}

// Test Case 10.3: Data Masking of sensitive fields.
func TestDataMasker(t *testing.T) {
	sensitivePayload := map[string]interface{}{
		"username":    "admin@eomp.local",
		"password":    "SuperSecretPassword123!",
		"api_key":     "sk_live_9921820491820391203",
		"credit_card": "4111 2222 3333 4444",
		"profile": map[string]interface{}{
			"auth_token": "secret-token-value",
			"city":       "Hanoi",
		},
	}

	masked := MaskSensitiveData(sensitivePayload)

	if masked["password"] != "********" {
		t.Errorf("expected masked password, got %v", masked["password"])
	}
	if masked["api_key"] != "********" {
		t.Errorf("expected masked api_key, got %v", masked["api_key"])
	}
	if masked["credit_card"] != "********" {
		t.Errorf("expected masked credit_card, got %v", masked["credit_card"])
	}
	profileMap := masked["profile"].(map[string]interface{})
	if profileMap["auth_token"] != "********" {
		t.Errorf("expected masked nested auth_token, got %v", profileMap["auth_token"])
	}
	if profileMap["city"] != "Hanoi" {
		t.Errorf("expected non-sensitive city preserved, got %v", profileMap["city"])
	}
}
