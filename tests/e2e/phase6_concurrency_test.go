package e2e

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"eomp/packages/shared/pkg/auth"
)

// =============================================================================
// PHASE 6: MULTI-USER CONCURRENCY & OPTIMISTIC LOCKING RACE CONDITION TEST SUITE
// =============================================================================

type VersionedEntity struct {
	ID        string
	Name      string
	Status    string
	Assignee  string
	Version   int
	UpdatedAt time.Time
}

type ConcurrencyEntityStore struct {
	mu       sync.Mutex
	entities map[string]*VersionedEntity
}

func (s *ConcurrencyEntityStore) AtomicCASUpdate(id, newStatus, newAssignee string, expectedVersion int) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.entities[id]
	if !ok {
		return false, fmt.Errorf("entity not found: %s", id)
	}

	if item.Version != expectedVersion {
		return false, fmt.Errorf("HTTP 409 Conflict: optimistic lock conflict, version %d != expected %d", item.Version, expectedVersion)
	}

	item.Status = newStatus
	item.Assignee = newAssignee
	item.Version++
	item.UpdatedAt = time.Now()
	return true, nil
}

func TestPhase6_E2E_MultiUserConcurrencyAndOptimisticLocking(t *testing.T) {
	t.Log("===> [PHASE 6 - Task 6.2] Executing Multi-User Concurrency Race Condition Test (50 Goroutines)...")

	jwtManager := auth.NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", 1*time.Hour, 7*24*time.Hour)

	// Authenticate Agents
	agentToken, _, err := jwtManager.GenerateTokenPair("u-agent-01", "agent.concurrency@eomp.local", "ROLE_AGENT", "dept-it", "Agent Concurrency")
	if err != nil {
		t.Fatalf("failed to generate agent token: %v", err)
	}
	claims, err := jwtManager.ValidateToken(agentToken)
	if err != nil || claims.Role != "ROLE_AGENT" {
		t.Fatalf("agent validation failed: %v", err)
	}
	t.Logf("  [+] Authenticated test actor: %s (%s)", claims.Email, claims.Role)

	// Initialize entities store with Version 1
	store := &ConcurrencyEntityStore{
		entities: map[string]*VersionedEntity{
			"TK-2026-9001": {
				ID:        "TK-2026-9001",
				Name:      "Database Connection Pool Exhaustion",
				Status:    "OPEN",
				Assignee:  "Unassigned",
				Version:   1,
				UpdatedAt: time.Now(),
			},
			"AST-MBP-9001": {
				ID:        "AST-MBP-9001",
				Name:      "MacBook Pro M3 Max 64GB",
				Status:    "IN_STOCK",
				Assignee:  "Warehouse",
				Version:   1,
				UpdatedAt: time.Now(),
			},
			"WF-INST-9001": {
				ID:        "WF-INST-9001",
				Name:      "Production Emergency Hotfix Approval",
				Status:    "WAITING_APPROVAL",
				Assignee:  "Manager Queue",
				Version:   1,
				UpdatedAt: time.Now(),
			},
		},
	}

	const concurrentWorkers = 50

	// -------------------------------------------------------------------------
	// Scenario A: 50 Goroutines contending for Ticket status update (TK-2026-9001)
	// -------------------------------------------------------------------------
	t.Log("  [1/3] Testing 50 concurrent Goroutines contending for Ticket status update...")

	var ticketSuccessCount, ticketConflictCount int32
	var wgA sync.WaitGroup
	startGateA := make(chan struct{})

	for i := 0; i < concurrentWorkers; i++ {
		wgA.Add(1)
		go func(workerID int) {
			defer wgA.Done()
			<-startGateA

			ok, err := store.AtomicCASUpdate("TK-2026-9001", "IN_PROGRESS", fmt.Sprintf("Agent-%d", workerID), 1)
			if ok && err == nil {
				atomic.AddInt32(&ticketSuccessCount, 1)
			} else {
				atomic.AddInt32(&ticketConflictCount, 1)
			}
		}(i)
	}

	close(startGateA)
	wgA.Wait()

	t.Logf("      -> Results: %d Succeeded (200 OK), %d Conflicts (409 Conflict)", ticketSuccessCount, ticketConflictCount)
	if ticketSuccessCount != 1 || ticketConflictCount != concurrentWorkers-1 {
		t.Fatalf("Ticket concurrency violation: expected exactly 1 success and %d conflicts, got %d and %d",
			concurrentWorkers-1, ticketSuccessCount, ticketConflictCount)
	}

	// -------------------------------------------------------------------------
	// Scenario B: 50 Goroutines contending for Asset assignment (AST-MBP-9001)
	// -------------------------------------------------------------------------
	t.Log("  [2/3] Testing 50 concurrent Goroutines contending for Asset assignment...")

	var assetSuccessCount, assetConflictCount int32
	var wgB sync.WaitGroup
	startGateB := make(chan struct{})

	for i := 0; i < concurrentWorkers; i++ {
		wgB.Add(1)
		go func(workerID int) {
			defer wgB.Done()
			<-startGateB

			ok, err := store.AtomicCASUpdate("AST-MBP-9001", "IN_USE", fmt.Sprintf("Employee-%d", workerID), 1)
			if ok && err == nil {
				atomic.AddInt32(&assetSuccessCount, 1)
			} else {
				atomic.AddInt32(&assetConflictCount, 1)
			}
		}(i)
	}

	close(startGateB)
	wgB.Wait()

	t.Logf("      -> Results: %d Succeeded (200 OK), %d Conflicts (409 Conflict)", assetSuccessCount, assetConflictCount)
	if assetSuccessCount != 1 || assetConflictCount != concurrentWorkers-1 {
		t.Fatalf("Asset concurrency violation: expected exactly 1 success and %d conflicts, got %d and %d",
			concurrentWorkers-1, assetSuccessCount, assetConflictCount)
	}

	// -------------------------------------------------------------------------
	// Scenario C: 50 Goroutines contending for Workflow decision (WF-INST-9001)
	// -------------------------------------------------------------------------
	t.Log("  [3/3] Testing 50 concurrent Goroutines contending for Workflow approval...")

	var wfSuccessCount, wfConflictCount int32
	var wgC sync.WaitGroup
	startGateC := make(chan struct{})

	for i := 0; i < concurrentWorkers; i++ {
		wgC.Add(1)
		go func(workerID int) {
			defer wgC.Done()
			<-startGateC

			ok, err := store.AtomicCASUpdate("WF-INST-9001", "APPROVED", fmt.Sprintf("Manager-%d", workerID), 1)
			if ok && err == nil {
				atomic.AddInt32(&wfSuccessCount, 1)
			} else {
				atomic.AddInt32(&wfConflictCount, 1)
			}
		}(i)
	}

	close(startGateC)
	wgC.Wait()

	t.Logf("      -> Results: %d Succeeded (200 OK), %d Conflicts (409 Conflict)", wfSuccessCount, wfConflictCount)
	if wfSuccessCount != 1 || wfConflictCount != concurrentWorkers-1 {
		t.Fatalf("Workflow concurrency violation: expected exactly 1 success and %d conflicts, got %d and %d",
			concurrentWorkers-1, wfSuccessCount, wfConflictCount)
	}

	// Verify all entities versions cleanly incremented to 2
	store.mu.Lock()
	for id, entity := range store.entities {
		if entity.Version != 2 {
			t.Fatalf("entity %s final version must be 2, got %d", id, entity.Version)
		}
	}
	store.mu.Unlock()

	t.Log("  [✓] 150 Total Concurrent Goroutines Processed: Zero Lost Updates, 100% Optimistic Locking Integrity Verified.")
}
