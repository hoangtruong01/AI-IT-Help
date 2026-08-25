package main

import (
	"context"
	"testing"
	"time"

	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/service"
)

// mockProblemRepo implements repository.ProblemRepository for in-memory testing.
type mockProblemRepo struct {
	problems []model.Problem
	links    map[string][]model.ProblemIncidentLink
}

func newMockProblemRepo() *mockProblemRepo {
	return &mockProblemRepo{
		problems: make([]model.Problem, 0),
		links:    make(map[string][]model.ProblemIncidentLink),
	}
}

func (m *mockProblemRepo) NextProblemNumber(ctx context.Context) (string, error) {
	return "PRB-1001", nil
}

func (m *mockProblemRepo) ListProblems(ctx context.Context, category, status string, page, pageSize int) ([]model.Problem, int, error) {
	var filtered []model.Problem
	for _, p := range m.problems {
		if (category == "" || category == "All" || p.Category == category) &&
			(status == "" || status == "All" || p.Status == status) {
			filtered = append(filtered, p)
		}
	}
	return filtered, len(filtered), nil
}

func (m *mockProblemRepo) GetProblemByID(ctx context.Context, id string) (*model.Problem, error) {
	for _, p := range m.problems {
		if p.ID == id || p.ProblemNumber == id {
			cp := p
			cp.LinkedCount = len(m.links[p.ID])
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *mockProblemRepo) CreateProblem(ctx context.Context, p *model.Problem) error {
	m.problems = append(m.problems, *p)
	return nil
}

func (m *mockProblemRepo) UpdateProblem(ctx context.Context, p *model.Problem) error {
	for i, existing := range m.problems {
		if existing.ID == p.ID {
			m.problems[i] = *p
			return nil
		}
	}
	return nil
}

func (m *mockProblemRepo) UpdateProblemStatus(ctx context.Context, id, status string, resolution *string, resolvedAt, closedAt *time.Time) error {
	for i, existing := range m.problems {
		if existing.ID == id {
			m.problems[i].Status = status
			if resolution != nil {
				m.problems[i].Resolution = resolution
			}
			m.problems[i].ResolvedAt = resolvedAt
			m.problems[i].ClosedAt = closedAt
			return nil
		}
	}
	return nil
}

func (m *mockProblemRepo) UpdateProblemRCA(ctx context.Context, id string, rootCause, workaround *string, isKnownError bool) error {
	for i, existing := range m.problems {
		if existing.ID == id {
			m.problems[i].RootCause = rootCause
			m.problems[i].Workaround = workaround
			m.problems[i].IsKnownError = isKnownError
			return nil
		}
	}
	return nil
}

func (m *mockProblemRepo) LinkIncident(ctx context.Context, link *model.ProblemIncidentLink) error {
	m.links[link.ProblemID] = append(m.links[link.ProblemID], *link)
	return nil
}

func (m *mockProblemRepo) UnlinkIncident(ctx context.Context, problemID, ticketID string) error {
	var updated []model.ProblemIncidentLink
	for _, l := range m.links[problemID] {
		if l.TicketID != ticketID && l.TicketNumber != ticketID {
			updated = append(updated, l)
		}
	}
	m.links[problemID] = updated
	return nil
}

func (m *mockProblemRepo) GetLinkedIncidents(ctx context.Context, problemID string) ([]model.ProblemIncidentLink, error) {
	return m.links[problemID], nil
}

func (m *mockProblemRepo) CascadeResolveLinkedTickets(ctx context.Context, problemID, problemNumber, resolution string) ([]string, error) {
	var resolved []string
	for _, l := range m.links[problemID] {
		resolved = append(resolved, l.TicketNumber)
	}
	return resolved, nil
}

func (m *mockProblemRepo) GetProblemStats(ctx context.Context) (*model.ProblemStats, error) {
	return &model.ProblemStats{
		TotalProblems:      len(m.problems),
		UnderInvestigation: 1,
		KnownErrors:        1,
		ResolvedProblems:   0,
		TotalLinkedTickets: 3,
	}, nil
}

// mockTicketRepo for Problem testing.
type mockTicketRepoForProblem struct {
	tickets map[string]*model.Ticket
}

func (m *mockTicketRepoForProblem) ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error) {
	return &model.TicketListResponse{}, nil
}
func (m *mockTicketRepoForProblem) FindTicketByID(ctx context.Context, id string) (*model.Ticket, error) {
	if t, ok := m.tickets[id]; ok {
		return t, nil
	}
	return nil, nil
}
func (m *mockTicketRepoForProblem) FindTicketByNumber(ctx context.Context, ticketNumber string) (*model.Ticket, error) {
	for _, t := range m.tickets {
		if t.TicketNumber == ticketNumber {
			return t, nil
		}
	}
	return nil, nil
}
func (m *mockTicketRepoForProblem) CreateTicket(ctx context.Context, ticket *model.Ticket) error {
	return nil
}
func (m *mockTicketRepoForProblem) UpdateTicketStatus(ctx context.Context, id, status string, assigneeID, assigneeName *string, resolvedAt, closedAt *time.Time, expectedVersion *int) error {
	if t, ok := m.tickets[id]; ok {
		t.Status = status
		t.Version++
	}
	return nil
}
func (m *mockTicketRepoForProblem) AssignTicket(ctx context.Context, id, assigneeID, assigneeName string, expectedVersion *int) error {
	if t, ok := m.tickets[id]; ok {
		t.AssigneeID = &assigneeID
		t.AssigneeName = &assigneeName
		t.Version++
	}
	return nil
}
func (m *mockTicketRepoForProblem) AddComment(ctx context.Context, comment *model.TicketComment) error {
	return nil
}
func (m *mockTicketRepoForProblem) ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error) {
	return nil, nil
}
func (m *mockTicketRepoForProblem) AddTimelineRecord(ctx context.Context, timeline *model.TicketTimeline) error {
	return nil
}
func (m *mockTicketRepoForProblem) ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error) {
	return nil, nil
}
func (m *mockTicketRepoForProblem) ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error) {
	return nil, nil
}
func (m *mockTicketRepoForProblem) ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error) {
	return nil, nil
}
func (m *mockTicketRepoForProblem) FindServiceCatalogItemByID(ctx context.Context, id string) (*model.ServiceCatalogItem, error) {
	return nil, nil
}
func (m *mockTicketRepoForProblem) NextTicketNumber(ctx context.Context) (string, error) {
	return "TK-1001", nil
}
func (m *mockTicketRepoForProblem) ListTicketsByAssetID(ctx context.Context, assetID string) ([]model.Ticket, error) {
	var list []model.Ticket
	for _, t := range m.tickets {
		if t.AffectedCIID != nil && *t.AffectedCIID == assetID {
			list = append(list, *t)
		}
	}
	return list, nil
}


// Test Case 7.1: Aggregate 3 duplicate Incidents and verify Cascade Resolution when Problem is Resolved.
func TestProblemManagement_TestCase_7_1(t *testing.T) {
	ctx := context.Background()
	probRepo := newMockProblemRepo()
	ticketRepo := &mockTicketRepoForProblem{
		tickets: map[string]*model.Ticket{
			"tk-1": {ID: "tk-1", TicketNumber: "INC-1001", Title: "VPN Gateway handshake timeout on 10.8.0.1", Status: "OPEN"},
			"tk-2": {ID: "tk-2", TicketNumber: "INC-1002", Title: "WireGuard drops every 15 mins for engineering", Status: "OPEN"},
			"tk-3": {ID: "tk-3", TicketNumber: "INC-1003", Title: "Cannot reach staging via VPN after 200 users connect", Status: "OPEN"},
		},
	}

	probSvc := service.NewProblemService(probRepo, ticketRepo)

	// 1. Create Problem Record
	p, err := probSvc.CreateProblem(ctx, model.CreateProblemPayload{
		Title:       "Intermittent WireGuard Gateway Packet Loss under Concurrency",
		Description: "Multiple users experience disconnections when concurrent peers exceed 200.",
		Category:    "Network & Access",
		Priority:    "CRITICAL",
		TicketIDs:   []string{"tk-1"},
	})
	if err != nil {
		t.Fatalf("failed to create problem: %v", err)
	}
	if p.ProblemNumber != "PRB-1001" {
		t.Errorf("expected problem number PRB-1001, got %s", p.ProblemNumber)
	}

	// 2. Link 2 additional incident tickets
	_, err = probSvc.LinkIncident(ctx, p.ID, model.LinkIncidentPayload{TicketID: "tk-2", LinkedBy: "Problem Manager"})
	if err != nil {
		t.Fatalf("failed to link tk-2: %v", err)
	}
	_, err = probSvc.LinkIncident(ctx, p.ID, model.LinkIncidentPayload{TicketID: "tk-3", LinkedBy: "Problem Manager"})
	if err != nil {
		t.Fatalf("failed to link tk-3: %v", err)
	}

	// 3. Verify all 3 incidents are linked
	_, links, err := probSvc.GetProblem(ctx, p.ID)
	if err != nil {
		t.Fatalf("failed to get problem details: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 linked incidents, got %d", len(links))
	}

	// 4. Update Problem status to RESOLVED -> Verify Cascade Resolution to all 3 tickets
	resText := "Root Cause: sysctl net.core.rmem_max increased to 26MB. Verified zero packet drops."
	updatedProblem, cascaded, err := probSvc.UpdateStatus(ctx, p.ID, model.UpdateProblemStatusPayload{
		Status:     "RESOLVED",
		Resolution: &resText,
	})
	if err != nil {
		t.Fatalf("failed to resolve problem: %v", err)
	}

	if updatedProblem.Status != "RESOLVED" {
		t.Errorf("expected problem status RESOLVED, got %s", updatedProblem.Status)
	}
	if len(cascaded) != 3 {
		t.Errorf("expected 3 cascaded tickets, got %d (%v)", len(cascaded), cascaded)
	}
}
