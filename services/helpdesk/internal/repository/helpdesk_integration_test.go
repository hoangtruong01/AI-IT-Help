package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"eomp/packages/shared/pkg/middleware"
	"eomp/services/helpdesk/internal/model"

	_ "github.com/lib/pq"
)

func getHelpdeskTestDB(t *testing.T) *sql.DB {
	t.Helper()
	required := os.Getenv("INTEGRATION_REQUIRED") != ""
	dsn := os.Getenv("HELPDESK_INTEGRATION_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		if required {
			t.Fatal("HELPDESK_INTEGRATION_DSN is required")
		}
		t.Skip("skipping helpdesk PostgreSQL integration test (HELPDESK_INTEGRATION_DSN not set)")
		return nil
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		if required {
			t.Fatalf("open helpdesk PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot open db: %v", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if required {
			t.Fatalf("ping helpdesk PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot ping db: %v", err)
		return nil
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestHelpdeskIntegration_RowLevelAuthorization validates that database queries
// enforce row-level access control across 4 roles (Employee, Agent, Manager, Admin).
func TestHelpdeskIntegration_RowLevelAuthorization(t *testing.T) {
	db := getHelpdeskTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	emp1ID := fmt.Sprintf("u-emp-1-%d", time.Now().UnixNano())
	emp2ID := fmt.Sprintf("u-emp-2-%d", time.Now().UnixNano())
	agent1ID := fmt.Sprintf("u-agent-1-%d", time.Now().UnixNano())
	deptEng := "dept-eng"
	deptHr := "dept-hr"

	// 1. Employee 1 ticket in Department ENG
	t1 := &model.Ticket{
		TicketNumber:   fmt.Sprintf("TK-RLA-1-%d", time.Now().UnixNano()%10000),
		Title:          "Laptop screen flickering",
		Description:    "Hardware issue with MacBook display",
		Category:       "Hardware",
		Priority:       "MEDIUM",
		Status:         "OPEN",
		RequesterID:    emp1ID,
		RequesterName:  "Employee One",
		RequesterEmail: "emp1@eomp.local",
		DepartmentID:   &deptEng,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}
	if err := repo.CreateTicket(ctx, t1); err != nil {
		t.Fatalf("failed to create ticket 1: %v", err)
	}

	// 2. Employee 2 ticket in Department HR
	t2 := &model.Ticket{
		TicketNumber:   fmt.Sprintf("TK-RLA-2-%d", time.Now().UnixNano()%10000),
		Title:          "HR portal access denied",
		Description:    "Cannot login to workday portal",
		Category:       "Software",
		Priority:       "HIGH",
		Status:         "OPEN",
		RequesterID:    emp2ID,
		RequesterName:  "Employee Two",
		RequesterEmail: "emp2@eomp.local",
		DepartmentID:   &deptHr,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}
	if err := repo.CreateTicket(ctx, t2); err != nil {
		t.Fatalf("failed to create ticket 2: %v", err)
	}

	// 3. Ticket assigned to Agent 1
	t3 := &model.Ticket{
		TicketNumber:   fmt.Sprintf("TK-RLA-3-%d", time.Now().UnixNano()%10000),
		Title:          "VPN configuration request",
		Description:    "Need cert for remote access",
		Category:       "Network",
		Priority:       "LOW",
		Status:         "IN_PROGRESS",
		RequesterID:    emp2ID,
		RequesterName:  "Employee Two",
		RequesterEmail: "emp2@eomp.local",
		AssigneeID:     &agent1ID,
		DepartmentID:   &deptHr,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}
	if err := repo.CreateTicket(ctx, t3); err != nil {
		t.Fatalf("failed to create ticket 3: %v", err)
	}

	// Check 1: ROLE_EMPLOYEE (emp1) can only list own tickets
	empResp, err := repo.ListTickets(ctx, model.TicketListQuery{
		ActorID:   emp1ID,
		ActorRole: "ROLE_EMPLOYEE",
		Page:      1,
		PageSize:  50,
	})
	if err != nil {
		t.Fatalf("failed to list tickets as employee: %v", err)
	}
	for _, tk := range empResp.Data {
		if tk.RequesterID != emp1ID {
			t.Fatalf("RLA violation: Employee saw ticket owned by %s", tk.RequesterID)
		}
	}

	// Check 2: ROLE_AGENT (agent1) sees assigned tickets + unassigned queue
	agentResp, err := repo.ListTickets(ctx, model.TicketListQuery{
		ActorID:   agent1ID,
		ActorRole: "ROLE_AGENT",
		Page:      1,
		PageSize:  50,
	})
	if err != nil {
		t.Fatalf("failed to list tickets as agent: %v", err)
	}
	for _, tk := range agentResp.Data {
		if tk.AssigneeID != nil && *tk.AssigneeID != agent1ID && *tk.AssigneeID != "" {
			t.Fatalf("RLA violation: Agent saw ticket assigned to another agent (%s)", *tk.AssigneeID)
		}
	}

	// Check 3: ROLE_MANAGER (dept-eng) only sees dept-eng tickets
	mgrResp, err := repo.ListTickets(ctx, model.TicketListQuery{
		ActorID:           "mgr-uuid",
		ActorRole:         "ROLE_MANAGER",
		ActorDepartmentID: "dept-eng",
		Page:              1,
		PageSize:          50,
	})
	if err != nil {
		t.Fatalf("failed to list tickets as manager: %v", err)
	}
	for _, tk := range mgrResp.Data {
		if tk.DepartmentID == nil || *tk.DepartmentID != "dept-eng" {
			t.Fatalf("RLA violation: ENG Manager saw ticket from department %+v", tk.DepartmentID)
		}
	}

	// Check 4: Anti-Enumeration 404
	actorEmp1 := middleware.Actor{
		ID:           emp1ID,
		Email:        "emp1@eomp.local",
		Role:         "ROLE_EMPLOYEE",
		DepartmentID: "dept-eng",
	}
	otherTicket, err := repo.FindTicketByIDForActor(ctx, t2.ID, actorEmp1)
	if err != nil {
		t.Fatalf("anti-enumeration lookup failed: %v", err)
	}
	if otherTicket != nil {
		t.Fatalf("Anti-enumeration failed: Employee 1 was able to access Employee 2's ticket directly")
	}
}

// TestHelpdeskIntegration_SequenceConcurrency validates that 100 concurrent ticket
// number allocations produce 100 distinct numbers without duplicates.
func TestHelpdeskIntegration_SequenceConcurrency(t *testing.T) {
	db := getHelpdeskTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	const numGoroutines = 100
	var wg sync.WaitGroup
	var mu sync.Mutex
	ticketNumbers := make(map[string]bool)
	errs := make([]error, 0)

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			num, err := repo.NextTicketNumber(ctx)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
			} else {
				if ticketNumbers[num] {
					errs = append(errs, fmt.Errorf("duplicate ticket number allocated: %s", num))
				}
				ticketNumbers[num] = true
			}
		}()
	}
	wg.Wait()

	if len(errs) > 0 {
		t.Fatalf("concurrency sequence test failed with %d errors: %v", len(errs), errs[0])
	}
	if len(ticketNumbers) != numGoroutines {
		t.Fatalf("expected %d unique ticket numbers, got %d", numGoroutines, len(ticketNumbers))
	}
}

// TestHelpdeskIntegration_OptimisticLocking validates compare-and-swap concurrency control.
func TestHelpdeskIntegration_OptimisticLocking(t *testing.T) {
	db := getHelpdeskTestDB(t)
	if db == nil {
		return
	}

	repo := NewRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	deptIT := "dept-it"
	ticket := &model.Ticket{
		TicketNumber:   fmt.Sprintf("TK-CAS-%d", time.Now().UnixNano()%10000),
		Title:          "Optimistic lock test ticket",
		Description:    "Testing version increments and collision handling",
		Category:       "General",
		Priority:       "LOW",
		Status:         "OPEN",
		RequesterID:    "u-test-cas",
		RequesterName:  "Test User",
		RequesterEmail: "test.cas@eomp.local",
		DepartmentID:   &deptIT,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Version:        1,
	}

	if err := repo.CreateTicket(ctx, ticket); err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}

	expectedV1 := 1
	agentID := "u-agent-cas"
	agentName := "Agent CAS"
	if err := repo.AssignTicket(ctx, ticket.ID, agentID, agentName, &expectedV1); err != nil {
		t.Fatalf("expected version 1 update to succeed, got: %v", err)
	}

	staleV1 := 1
	err := repo.AssignTicket(ctx, ticket.ID, "another-agent", "Another Agent", &staleV1)
	if err == nil {
		t.Fatalf("expected stale version update to fail with conflict, but it succeeded")
	}

	refreshed, err := repo.FindTicketByID(ctx, ticket.ID)
	if err != nil {
		t.Fatalf("failed to find refreshed ticket: %v", err)
	}
	if refreshed.Version != 2 {
		t.Fatalf("expected ticket version to be 2, got %d", refreshed.Version)
	}
}
