package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"eomp/packages/shared/pkg/eventbus"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/service"
)

type mockFullRepo struct {
	tickets       map[string]*model.Ticket
	timelines     []*model.TicketTimeline
	outbox        []model.OutboxEvent
	serviceItems  map[string]*model.ServiceCatalogItem
	nextTicketNum int
}

func newMockFullRepo() *mockFullRepo {
	return &mockFullRepo{
		tickets:      make(map[string]*model.Ticket),
		serviceItems: make(map[string]*model.ServiceCatalogItem),
	}
}

func (m *mockFullRepo) ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error) {
	return &model.TicketListResponse{}, nil
}
func (m *mockFullRepo) FindTicketByID(ctx context.Context, id string) (*model.Ticket, error) {
	if t, ok := m.tickets[id]; ok {
		copy := *t
		return &copy, nil
	}
	return nil, nil
}
func (m *mockFullRepo) FindTicketByIDForActor(ctx context.Context, id string, actor middleware.Actor) (*model.Ticket, error) {
	return m.FindTicketByID(ctx, id)
}
func (m *mockFullRepo) FindTicketByNumber(ctx context.Context, num string) (*model.Ticket, error) {
	for _, t := range m.tickets {
		if t.TicketNumber == num {
			copy := *t
			return &copy, nil
		}
	}
	return nil, nil
}
func (m *mockFullRepo) CreateTicket(ctx context.Context, t *model.Ticket) error {
	m.nextTicketNum++
	t.ID = fmt.Sprintf("tk-%d", m.nextTicketNum)
	copy := *t
	m.tickets[t.ID] = &copy
	return nil
}
func (m *mockFullRepo) UpdateTicketStatus(ctx context.Context, id, status string, assigneeID, assigneeName *string, resolvedAt, closedAt *time.Time, expectedVersion *int) error {
	if t, ok := m.tickets[id]; ok {
		t.Status = status
		t.Version++
	}
	return nil
}
func (m *mockFullRepo) AssignTicket(ctx context.Context, id, assigneeID, assigneeName string, expectedVersion *int) error {
	if t, ok := m.tickets[id]; ok {
		t.AssigneeID = &assigneeID
		t.AssigneeName = &assigneeName
		t.Version++
	}
	return nil
}
func (m *mockFullRepo) RecordFirstResponse(ctx context.Context, ticketID string, respondedAt time.Time) error {
	if t, ok := m.tickets[ticketID]; ok && t.RespondedAt == nil {
		t.RespondedAt = &respondedAt
	}
	return nil
}
func (m *mockFullRepo) AddComment(ctx context.Context, comment *model.TicketComment) error {
	return nil
}
func (m *mockFullRepo) ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error) {
	return nil, nil
}
func (m *mockFullRepo) AddTimelineRecord(ctx context.Context, timeline *model.TicketTimeline) error {
	m.timelines = append(m.timelines, timeline)
	return nil
}
func (m *mockFullRepo) ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error) {
	return nil, nil
}
func (m *mockFullRepo) ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error) {
	return nil, nil
}
func (m *mockFullRepo) ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error) {
	return nil, nil
}
func (m *mockFullRepo) FindServiceCatalogItemByID(ctx context.Context, id string) (*model.ServiceCatalogItem, error) {
	if item, ok := m.serviceItems[id]; ok {
		return item, nil
	}
	return nil, nil
}
func (m *mockFullRepo) NextTicketNumber(ctx context.Context) (string, error) {
	m.nextTicketNum++
	return fmt.Sprintf("TK-%04d", m.nextTicketNum), nil
}
func (m *mockFullRepo) ListTicketsByAssetID(ctx context.Context, assetID string) ([]model.Ticket, error) {
	return nil, nil
}
func (m *mockFullRepo) ListTicketsByAssetIDForActor(ctx context.Context, assetID string, actor middleware.Actor) ([]model.Ticket, error) {
	return nil, nil
}

func (m *mockFullRepo) CreateTicketWithOutbox(ctx context.Context, ticket *model.Ticket, timeline *model.TicketTimeline, outbox *model.OutboxEvent) error {
	if err := m.CreateTicket(ctx, ticket); err != nil {
		return err
	}
	if timeline != nil {
		m.timelines = append(m.timelines, timeline)
	}
	if outbox != nil {
		m.outbox = append(m.outbox, *outbox)
	}
	return nil
}

func (m *mockFullRepo) UpdateTicketStatusWithOutbox(ctx context.Context, id, status string, assigneeID, assigneeName *string, resolvedAt, closedAt *time.Time, expectedVersion *int, timeline *model.TicketTimeline, outbox *model.OutboxEvent) error {
	if err := m.UpdateTicketStatus(ctx, id, status, assigneeID, assigneeName, resolvedAt, closedAt, expectedVersion); err != nil {
		return err
	}
	if timeline != nil {
		m.timelines = append(m.timelines, timeline)
	}
	if outbox != nil {
		m.outbox = append(m.outbox, *outbox)
	}
	return nil
}

func (m *mockFullRepo) UpdateTicketApprovalWithOutbox(ctx context.Context, id, status string, slaRespDeadline, slaResolDeadline *time.Time, closedAt *time.Time, timeline *model.TicketTimeline, outbox *model.OutboxEvent) error {
	if t, ok := m.tickets[id]; ok {
		t.Status = status
		if slaRespDeadline != nil {
			t.SLAResponseDeadline = *slaRespDeadline
		}
		if slaResolDeadline != nil {
			t.SLAResolutionDeadline = *slaResolDeadline
		}
		if closedAt != nil {
			t.ClosedAt = closedAt
		}
		t.Version++
	}
	if timeline != nil {
		m.timelines = append(m.timelines, timeline)
	}
	if outbox != nil {
		m.outbox = append(m.outbox, *outbox)
	}
	return nil
}

func (m *mockFullRepo) AssignTicketWithOutbox(ctx context.Context, id, assigneeID, assigneeName string, expectedVersion *int, timeline *model.TicketTimeline, outbox *model.OutboxEvent) error {
	return m.AssignTicket(ctx, id, assigneeID, assigneeName, expectedVersion)
}

func (m *mockFullRepo) AddCommentWithOutbox(ctx context.Context, comment *model.TicketComment, timeline *model.TicketTimeline, outbox *model.OutboxEvent) error {
	return m.AddComment(ctx, comment)
}

func (m *mockFullRepo) FetchPendingOutboxEvents(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	var pending []model.OutboxEvent
	for _, ev := range m.outbox {
		if ev.Status == "PENDING" {
			pending = append(pending, ev)
		}
	}
	return pending, nil
}

func (m *mockFullRepo) MarkOutboxEventPublished(ctx context.Context, id string) error {
	for i := range m.outbox {
		if m.outbox[i].ID == id {
			m.outbox[i].Status = "PUBLISHED"
		}
	}
	return nil
}

func (m *mockFullRepo) MarkOutboxEventFailed(ctx context.Context, id string, errStr string) error {
	return nil
}

func TestServiceCatalog_RequiresApproval_Lifecycle(t *testing.T) {
	ctx := context.Background()
	repo := newMockFullRepo()
	bus := eventbus.NewMemoryEventBus()
	slaEngine := service.NewSLAEngine()
	ticketSvc := service.NewTicketService(repo, slaEngine, bus)

	// Configure a service catalog item requiring manager approval
	itemCode := "srv-laptop-replacement"
	repo.serviceItems[itemCode] = &model.ServiceCatalogItem{
		ID:                   itemCode,
		Name:                 "High-Performance Laptop Upgrade",
		RequiresApproval:     true,
		IsActive:             true,
		SLAResponseMinutes:   60,
		SLAResolutionMinutes: 480,
	}

	// 1. Create Ticket for this item
	req := &model.CreateTicketRequest{
		Title:          "Request for M3 MacBook Pro 32GB",
		Description:    "Need additional RAM and GPU power for local LLM inference and Docker pipelines.",
		Priority:       "HIGH",
		RequesterEmail: "dev@eomp.local",
		RequesterName:  "Lead Developer",
		ServiceItemID:  &itemCode,
	}

	ticket, err := ticketSvc.CreateTicket(ctx, req)
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}

	// Status MUST be WAITING_APPROVAL, not OPEN
	if ticket.Status != model.StatusWaitingApproval {
		t.Fatalf("expected status WAITING_APPROVAL, got %s", ticket.Status)
	}

	// Outbox event must be queued
	if len(repo.outbox) == 0 {
		t.Fatal("expected outbox event to be queued within ticket creation transaction")
	}

	// 2. Simulate Approval Granted event from Workflow Engine
	approvalEv := eventbus.Event{
		Source: "workflow",
		Type:   eventbus.TopicApprovalDecided,
		Data: map[string]any{
			"entity_id":  ticket.ID,
			"decision":   "APPROVED",
			"actor_name": "Engineering VP",
			"notes":      "Hardware budget verified and approved.",
		},
	}
	if err := ticketSvc.HandleApprovalDecided(ctx, approvalEv); err != nil {
		t.Fatalf("HandleApprovalDecided failed: %v", err)
	}

	// Ticket status MUST now be OPEN
	approvedTicket, _ := ticketSvc.GetTicket(ctx, ticket.ID)
	if approvedTicket.Status != model.StatusOpen {
		t.Fatalf("expected ticket status to transition to OPEN after approval, got %s", approvedTicket.Status)
	}

	// SLA deadlines must now be activated
	if approvedTicket.SLAResolutionDeadline.IsZero() {
		t.Fatal("expected SLA Resolution Deadline to be computed upon approval")
	}
}

func TestServiceCatalog_RequiresApproval_Rejected(t *testing.T) {
	ctx := context.Background()
	repo := newMockFullRepo()
	bus := eventbus.NewMemoryEventBus()
	slaEngine := service.NewSLAEngine()
	ticketSvc := service.NewTicketService(repo, slaEngine, bus)

	itemCode := "srv-executive-phone"
	repo.serviceItems[itemCode] = &model.ServiceCatalogItem{
		ID:               itemCode,
		Name:             "Corporate iPhone 16 Pro Max",
		RequiresApproval: true,
		IsActive:         true,
	}

	req := &model.CreateTicketRequest{
		Title:          "Request for corporate iPhone upgrade",
		Description:    "Requesting latest model for executive testing and communications.",
		Priority:       "MEDIUM",
		RequesterEmail: "staff@eomp.local",
		RequesterName:  "Junior Staff",
		ServiceItemID:  &itemCode,
	}

	ticket, err := ticketSvc.CreateTicket(ctx, req)
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}

	// Reject the request
	rejectEv := eventbus.Event{
		Source: "workflow",
		Type:   eventbus.TopicApprovalDecided,
		Data: map[string]any{
			"entity_id":  ticket.ID,
			"decision":   "REJECTED",
			"actor_name": "Department Director",
			"notes":      "Executive hardware tier is not eligible for current role.",
		},
	}
	if err := ticketSvc.HandleApprovalDecided(ctx, rejectEv); err != nil {
		t.Fatalf("HandleApprovalDecided failed: %v", err)
	}

	rejectedTicket, _ := ticketSvc.GetTicket(ctx, ticket.ID)
	if rejectedTicket.Status != model.StatusClosed {
		t.Fatalf("expected ticket status to transition to CLOSED upon rejection, got %s", rejectedTicket.Status)
	}
}
