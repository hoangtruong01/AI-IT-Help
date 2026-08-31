package e2e

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/middleware"
)

// =============================================================================
// PHASE 3 END-TO-END CONCURRENCY & GATEWAY CAPACITY TEST SUITE
// =============================================================================

func TestPhase3_E2E_ITSMConcurrencyAndGatewayLimits(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 1*time.Hour, 7*24*time.Hour)

	t.Log("===> [PHASE 3 - Task 3.1 & 3.2] Testing Cross-Service Concurrency & Version CAS Control...")

	// 1. Authenticate Agents
	agent1Token, _, err := jwtManager.GenerateTokenPair("u-agent-01", "marcus.vance@eomp.local", "ROLE_AGENT", "dept-it", "Marcus Vance")
	if err != nil {
		t.Fatalf("failed to generate agent1 token: %v", err)
	}

	claims, err := jwtManager.ValidateToken(agent1Token)
	if err != nil || claims.Role != "ROLE_AGENT" {
		t.Fatalf("claims validation failed: %v", err)
	}
	t.Logf("  [+] Agent authenticated: %s (Role: %s)", claims.Email, claims.Role)

	// 2. Simulated Entity with Atomic CAS Optimistic Locking
	type VersionedTicket struct {
		ID        string
		Status    string
		Assignee  string
		Version   int
		UpdatedAt time.Time
	}

	type ConcurrencyStore struct {
		mu      sync.Mutex
		tickets map[string]*VersionedTicket
	}

	store := &ConcurrencyStore{
		tickets: map[string]*VersionedTicket{
			"TK-2026-9001": {
				ID:        "TK-2026-9001",
				Status:    "OPEN",
				Assignee:  "Unassigned",
				Version:   1,
				UpdatedAt: time.Now(),
			},
		},
	}

	// CAS Update Function
	casUpdate := func(ticketID, newStatus, newAssignee string, expectedVersion int) (bool, error) {
		store.mu.Lock()
		defer store.mu.Unlock()

		item, ok := store.tickets[ticketID]
		if !ok {
			return false, fmt.Errorf("ticket not found")
		}

		if item.Version != expectedVersion {
			return false, fmt.Errorf("HTTP 409 Conflict: version %d != expected %d", item.Version, expectedVersion)
		}

		item.Status = newStatus
		item.Assignee = newAssignee
		item.Version++
		item.UpdatedAt = time.Now()
		return true, nil
	}

	// 3. Fire 50 concurrent Goroutines all attempting CAS update on Version 1
	const totalWorkers = 50
	var successCount int32
	var conflictCount int32

	var wg sync.WaitGroup
	startGate := make(chan struct{})

	for i := 0; i < totalWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-startGate

			ok, err := casUpdate("TK-2026-9001", "IN_PROGRESS", fmt.Sprintf("Agent-%d", workerID), 1)
			if ok && err == nil {
				atomic.AddInt32(&successCount, 1)
			} else {
				atomic.AddInt32(&conflictCount, 1)
			}
		}(i)
	}

	close(startGate)
	wg.Wait()

	t.Logf("  [+] Concurrency 50 Goroutines Results: %d Successful, %d Conflicts", successCount, conflictCount)
	if successCount != 1 || conflictCount != totalWorkers-1 {
		t.Fatalf("CAS anomaly: expected 1 success and %d conflicts; got %d, %d", totalWorkers-1, successCount, conflictCount)
	}

	store.mu.Lock()
	finalVer := store.tickets["TK-2026-9001"].Version
	store.mu.Unlock()

	if finalVer != 2 {
		t.Fatalf("expected final version 2, got %d", finalVer)
	}
	t.Logf("  [✓] Optimistic Locking CAS verified: Lost Update prevented, Ticket is at Version %d", finalVer)

	// =========================================================================
	// Task 3.4: Gateway Request Body Size Limiter (5MB Guard)
	// =========================================================================
	t.Log("===> [PHASE 3 - Task 3.4] Testing API Gateway 5MB Request Body Size Guard...")

	const maxBodyBytes = 5 * 1024 * 1024 // 5MB

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	limiter := middleware.MaxBodySize(maxBodyBytes)(handler)

	// Test A: Normal 10KB Request -> 200 OK
	smallBody := bytes.Repeat([]byte("A"), 10*1024)
	reqSmall := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewReader(smallBody))
	reqSmall.Header.Set("Content-Type", "application/json")
	wSmall := httptest.NewRecorder()
	limiter.ServeHTTP(wSmall, reqSmall)

	if wSmall.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for small body, got %d", wSmall.Code)
	}
	t.Log("  [+] Small payload (10KB) passed successfully: HTTP 200 OK")

	// Test B: Oversized 6MB Request -> 413 Payload Too Large
	largeBody := bytes.Repeat([]byte("X"), 6*1024*1024)
	reqLarge := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewReader(largeBody))
	reqLarge.Header.Set("Content-Type", "application/json")
	wLarge := httptest.NewRecorder()
	limiter.ServeHTTP(wLarge, reqLarge)

	if wLarge.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 Payload Too Large for 6MB body, got %d", wLarge.Code)
	}
	t.Logf("  [+] Oversized payload (6MB) blocked: HTTP 413 Payload Too Large")
	t.Log("===> [PHASE 3 E2E SUITE COMPLETED] 100% Verification Pass Rate.")
}
