package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/eventbus"
	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/service"
)

// =============================================================================
// PHASE 3 ITSM / HELPDESK & CONCURRENCY CONTROL TESTS
// =============================================================================

// 1. In-Memory Mock Repository with Atomic CAS for Concurrency Testing
type concurrentMockTicketRepo struct {
	mu      sync.RWMutex
	tickets map[string]*model.Ticket
}

func newConcurrentMockTicketRepo() *concurrentMockTicketRepo {
	return &concurrentMockTicketRepo{
		tickets: make(map[string]*model.Ticket),
	}
}

func (r *concurrentMockTicketRepo) ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var list []model.Ticket
	for _, t := range r.tickets {
		list = append(list, *t)
	}
	return &model.TicketListResponse{Data: list, Total: len(list)}, nil
}

func (r *concurrentMockTicketRepo) FindTicketByID(ctx context.Context, id string) (*model.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if t, exists := r.tickets[id]; exists {
		copy := *t
		return &copy, nil
	}
	return nil, nil
}

func (r *concurrentMockTicketRepo) FindTicketByNumber(ctx context.Context, ticketNumber string) (*model.Ticket, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, t := range r.tickets {
		if t.TicketNumber == ticketNumber {
			copy := *t
			return &copy, nil
		}
	}
	return nil, nil
}

func (r *concurrentMockTicketRepo) CreateTicket(ctx context.Context, ticket *model.Ticket) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ticket.ID = fmt.Sprintf("t-%d", len(r.tickets)+1)
	ticket.Version = 1
	ticket.CreatedAt = time.Now()
	ticket.UpdatedAt = time.Now()
	r.tickets[ticket.ID] = ticket
	return nil
}

func (r *concurrentMockTicketRepo) UpdateTicketStatus(ctx context.Context, id, status string, assigneeID, assigneeName *string, resolvedAt, closedAt *time.Time, expectedVersion *int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, exists := r.tickets[id]
	if !exists {
		return errors.NotFound("ticket not found")
	}

	// Compare-And-Swap check
	if expectedVersion != nil && t.Version != *expectedVersion {
		return errors.Conflict(fmt.Sprintf("optimistic lock conflict: current version %d does not match expected %d", t.Version, *expectedVersion))
	}

	t.Status = status
	if assigneeID != nil {
		t.AssigneeID = assigneeID
	}
	if assigneeName != nil {
		t.AssigneeName = assigneeName
	}
	if resolvedAt != nil {
		t.ResolvedAt = resolvedAt
	}
	if closedAt != nil {
		t.ClosedAt = closedAt
	}
	t.Version++
	t.UpdatedAt = time.Now()
	return nil
}

func (r *concurrentMockTicketRepo) AssignTicket(ctx context.Context, id, assigneeID, assigneeName string, expectedVersion *int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, exists := r.tickets[id]
	if !exists {
		return errors.NotFound("ticket not found")
	}

	if expectedVersion != nil && t.Version != *expectedVersion {
		return errors.Conflict(fmt.Sprintf("optimistic lock conflict: current version %d does not match expected %d", t.Version, *expectedVersion))
	}

	t.AssigneeID = &assigneeID
	t.AssigneeName = &assigneeName
	if t.Status == model.StatusOpen {
		t.Status = model.StatusAssigned
	}
	t.Version++
	t.UpdatedAt = time.Now()
	return nil
}

func (r *concurrentMockTicketRepo) AddComment(ctx context.Context, comment *model.TicketComment) error {
	return nil
}
func (r *concurrentMockTicketRepo) ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error) {
	return nil, nil
}
func (r *concurrentMockTicketRepo) AddTimelineRecord(ctx context.Context, timeline *model.TicketTimeline) error {
	return nil
}
func (r *concurrentMockTicketRepo) ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error) {
	return nil, nil
}
func (r *concurrentMockTicketRepo) ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error) {
	return nil, nil
}
func (r *concurrentMockTicketRepo) ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error) {
	return nil, nil
}
func (r *concurrentMockTicketRepo) FindServiceCatalogItemByID(ctx context.Context, id string) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (r *concurrentMockTicketRepo) NextTicketNumber(ctx context.Context) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return fmt.Sprintf("TK-%04d", 1000+len(r.tickets)+1), nil
}
func (r *concurrentMockTicketRepo) ListTicketsByAssetID(ctx context.Context, assetID string) ([]model.Ticket, error) {
	return nil, nil
}

// TEST: ITIL v4 State Machine
func TestPhase3_ITILv4StateMachineTransitions(t *testing.T) {
	t.Log("===> [PHASE 3 - Task 3.1] Testing Strict ITIL v4 State Machine Transition Rules...")

	repo := newConcurrentMockTicketRepo()
	slaEng := service.NewSLAEngine()
	bus := eventbus.NewMemoryEventBus()
	svc := service.NewTicketService(repo, slaEng, bus)

	ctx := context.Background()
	ticket, err := svc.CreateTicket(ctx, &model.CreateTicketRequest{
		Title:          "Core Switch Port Flapping in Server Room B",
		Description:    "Intermittent packet loss on VLAN 10 uplink",
		Category:       "NETWORK",
		Priority:       model.PriorityHigh,
		RequesterID:    "u-emp-01",
		RequesterName:  "Kenji Sato",
		RequesterEmail: "kenji.sato@eomp.local",
	})
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}

	if ticket.Status != model.StatusOpen || ticket.Version != 1 {
		t.Fatalf("expected status OPEN, version 1; got %s, version %d", ticket.Status, ticket.Version)
	}
	t.Logf("  [+] Initial State: Ticket %s | Status: %s | Version: %d", ticket.TicketNumber, ticket.Status, ticket.Version)

	// Step 1: Valid Transition OPEN -> ASSIGNED
	v := 1
	ticket, err = svc.UpdateStatus(ctx, ticket.ID, &model.UpdateTicketStatusRequest{
		Status:  model.StatusAssigned,
		Version: &v,
	}, "u-agent-01", "Marcus Vance")
	if err != nil {
		t.Fatalf("valid transition OPEN -> ASSIGNED failed: %v", err)
	}
	if ticket.Status != model.StatusAssigned || ticket.Version != 2 {
		t.Fatalf("expected status ASSIGNED, version 2; got %s, version %d", ticket.Status, ticket.Version)
	}
	t.Logf("  [+] Valid Transition: OPEN ➔ ASSIGNED | Version: %d", ticket.Version)

	// Step 2: Valid Transition ASSIGNED -> IN_PROGRESS
	v = 2
	ticket, err = svc.UpdateStatus(ctx, ticket.ID, &model.UpdateTicketStatusRequest{
		Status:  model.StatusInProgress,
		Version: &v,
	}, "u-agent-01", "Marcus Vance")
	if err != nil {
		t.Fatalf("valid transition ASSIGNED -> IN_PROGRESS failed: %v", err)
	}
	t.Logf("  [+] Valid Transition: ASSIGNED ➔ IN_PROGRESS | Version: %d", ticket.Version)

	// Step 3: Valid Transition IN_PROGRESS -> RESOLVED
	v = 3
	ticket, err = svc.UpdateStatus(ctx, ticket.ID, &model.UpdateTicketStatusRequest{
		Status:  model.StatusResolved,
		Version: &v,
	}, "u-agent-01", "Marcus Vance")
	if err != nil {
		t.Fatalf("valid transition IN_PROGRESS -> RESOLVED failed: %v", err)
	}
	t.Logf("  [+] Valid Transition: IN_PROGRESS ➔ RESOLVED | Version: %d", ticket.Version)

	// Step 4: Valid Transition RESOLVED -> CLOSED
	v = 4
	ticket, err = svc.UpdateStatus(ctx, ticket.ID, &model.UpdateTicketStatusRequest{
		Status:  model.StatusClosed,
		Version: &v,
	}, "u-agent-01", "Marcus Vance")
	if err != nil {
		t.Fatalf("valid transition RESOLVED -> CLOSED failed: %v", err)
	}
	t.Logf("  [+] Valid Transition: RESOLVED ➔ CLOSED (Terminal) | Version: %d", ticket.Version)

	// Step 5: INVALID Transition CLOSED -> IN_PROGRESS (Must be rejected with 400 Bad Request)
	v = 5
	_, err = svc.UpdateStatus(ctx, ticket.ID, &model.UpdateTicketStatusRequest{
		Status:  model.StatusInProgress,
		Version: &v,
	}, "u-agent-01", "Marcus Vance")
	if err == nil {
		t.Fatalf("security/business violation: invalid transition CLOSED -> IN_PROGRESS was permitted!")
	}
	if appErr, ok := err.(*errors.AppError); ok {
		if appErr.StatusCode != http.StatusBadRequest {
			t.Fatalf("expected status code 400, got %d", appErr.StatusCode)
		}
		t.Logf("  [+] Invalid Transition Correctly Blocked: CLOSED ➔ IN_PROGRESS rejected with error [%s]: %s", appErr.Code, appErr.Message)
	} else {
		t.Fatalf("expected *errors.AppError, got %T: %v", err, err)
	}

	t.Log("  [✓] ITIL v4 State Machine Transition Rules fully verified.")
}

// TEST: 50 Goroutines Optimistic Locking Concurrency Test
func TestPhase3_OptimisticLocking_50Goroutines_Concurrency(t *testing.T) {
	t.Log("===> [PHASE 3 - Task 3.2] Testing Optimistic Locking under 50 Concurrent Goroutines...")

	repo := newConcurrentMockTicketRepo()
	slaEng := service.NewSLAEngine()
	bus := eventbus.NewMemoryEventBus()
	svc := service.NewTicketService(repo, slaEng, bus)

	ctx := context.Background()
	ticket, err := svc.CreateTicket(ctx, &model.CreateTicketRequest{
		Title:          "Database High Memory Alert on Cluster 01",
		Description:    "PostgreSQL RAM usage exceeded 92%",
		Category:       "DATABASE",
		Priority:       model.PriorityUrgent,
		RequesterID:    "u-sre-01",
		RequesterName:  "Sarah Jenkins",
		RequesterEmail: "sarah.jenkins@eomp.local",
	})
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}

	initialVersion := ticket.Version // Version 1
	t.Logf("  [+] Target Ticket Created: %s | ID: %s | Base Version: %d", ticket.TicketNumber, ticket.ID, initialVersion)

	const totalGoroutines = 50
	var successCount int32
	var conflictCount int32

	var wg sync.WaitGroup
	startSignal := make(chan struct{})

	for i := 0; i < totalGoroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			// Wait for synchronization signal
			<-startSignal

			expectedVer := initialVersion // Everyone attempts to update using Version 1
			_, updateErr := svc.UpdateStatus(ctx, ticket.ID, &model.UpdateTicketStatusRequest{
				Status:  model.StatusAssigned,
				Version: &expectedVer,
				Notes:   fmt.Sprintf("Triage attempt by worker #%d", workerID),
			}, fmt.Sprintf("u-worker-%d", workerID), fmt.Sprintf("Worker %d", workerID))

			if updateErr == nil {
				atomic.AddInt32(&successCount, 1)
			} else if appErr, ok := updateErr.(*errors.AppError); ok && appErr.StatusCode == http.StatusConflict {
				atomic.AddInt32(&conflictCount, 1)
			}
		}(i)
	}

	// Release all 50 workers simultaneously
	close(startSignal)
	wg.Wait()

	t.Logf("  [+] Concurrency Execution Results across %d parallel goroutines:", totalGoroutines)
	t.Logf("      - Successful Updates (200 OK): %d", successCount)
	t.Logf("      - Blocked Conflicts (409 Conflict): %d", conflictCount)

	// Strict assertion: Exactly 1 update succeeded, exactly 49 received 409 Conflict
	if successCount != 1 {
		t.Fatalf("concurrency anomaly: expected exactly 1 winner, got %d", successCount)
	}
	if conflictCount != int32(totalGoroutines-1) {
		t.Fatalf("concurrency anomaly: expected %d conflict rejections, got %d", totalGoroutines-1, conflictCount)
	}

	// Inspect final ticket state
	finalTicket, _ := svc.GetTicket(ctx, ticket.ID)
	if finalTicket.Version != initialVersion+1 {
		t.Fatalf("expected final version %d, got %d", initialVersion+1, finalTicket.Version)
	}

	t.Logf("  [✓] Optimistic Locking successfully prevented Lost Update Anomaly! Final Version: %d", finalTicket.Version)
}
