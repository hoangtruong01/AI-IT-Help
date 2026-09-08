package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/eventbus"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/repository"
)

// TicketService defines ITSM ticket business logic
type TicketService interface {
	ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error)
	GetTicket(ctx context.Context, id string) (*model.Ticket, error)
	GetTicketForActor(ctx context.Context, id string, actor middleware.Actor) (*model.Ticket, error)
	CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error)
	UpdateStatus(ctx context.Context, id string, req *model.UpdateTicketStatusRequest, actorID, actorName string) (*model.Ticket, error)
	AssignTicket(ctx context.Context, id string, req *model.AssignTicketRequest, actorID, actorName string) (*model.Ticket, error)

	AddComment(ctx context.Context, ticketID string, req *model.AddCommentRequest, authorID, authorName, authorRole string) (*model.TicketComment, error)
	ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error)
	ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error)

	ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error)
	ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error)
	GetTicketsByAssetID(ctx context.Context, assetID string) ([]model.Ticket, error)
	GetTicketsByAssetIDForActor(ctx context.Context, assetID string, actor middleware.Actor) ([]model.Ticket, error)

	HandleApprovalDecided(ctx context.Context, event eventbus.Event) error
}

type ticketService struct {
	repo      repository.Repository
	slaEngine SLAEngine
	bus       eventbus.EventBus
}

// NewTicketService constructs a new TicketService
func NewTicketService(repo repository.Repository, slaEngine SLAEngine, bus eventbus.EventBus) TicketService {
	return &ticketService{
		repo:      repo,
		slaEngine: slaEngine,
		bus:       bus,
	}
}

func (s *ticketService) ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error) {
	if query.ActorID == "" {
		return nil, errors.Unauthorized("missing user identity")
	}
	switch query.ActorRole {
	case "ROLE_ADMIN", "ROLE_AGENT", "ROLE_EMPLOYEE":
	case "ROLE_MANAGER":
		if query.ActorDepartmentID == "" {
			return nil, errors.Forbidden("manager department scope is required")
		}
	default:
		return nil, errors.Forbidden("unknown user role")
	}
	resp, err := s.repo.ListTickets(ctx, query)
	if err != nil {
		return nil, errors.Internal(ctx, "helpdesk list tickets", err)
	}

	// Dynamic SLA evaluation for each returned ticket
	for i := range resp.Data {
		resp.Data[i].SLAStatus = s.slaEngine.EvaluateSLAStatus(&resp.Data[i])
	}

	return resp, nil
}

func (s *ticketService) GetTicket(ctx context.Context, id string) (*model.Ticket, error) {
	if id == "" {
		return nil, errors.BadRequest("ticket id is required")
	}

	ticket, err := s.repo.FindTicketByID(ctx, id)
	if err != nil {
		return nil, errors.Internal(ctx, "helpdesk get ticket", err)
	}
	if ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}

	ticket.SLAStatus = s.slaEngine.EvaluateSLAStatus(ticket)
	return ticket, nil
}

func (s *ticketService) GetTicketForActor(ctx context.Context, id string, actor middleware.Actor) (*model.Ticket, error) {
	if id == "" {
		return nil, errors.BadRequest("ticket id is required")
	}
	if !actor.IsValid() {
		return nil, errors.Unauthorized("valid user identity and role are required")
	}
	if actor.IsManager() && actor.DepartmentID == "" {
		return nil, errors.Forbidden("manager department scope is required")
	}

	ticket, err := s.repo.FindTicketByIDForActor(ctx, id, actor)
	if err != nil {
		return nil, errors.Internal(ctx, "helpdesk get scoped ticket", err)
	}
	if ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}
	ticket.SLAStatus = s.slaEngine.EvaluateSLAStatus(ticket)
	return ticket, nil
}

func (s *ticketService) CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error) {
	req.Title = strings.TrimSpace(req.Title)
	req.Description = strings.TrimSpace(req.Description)
	req.RequesterEmail = strings.TrimSpace(req.RequesterEmail)

	if len(req.Title) < 5 || len(req.Title) > 255 {
		return nil, errors.BadRequest("title must be between 5 and 255 characters")
	}
	if len(req.Description) < 10 || len(req.Description) > 5000 {
		return nil, errors.BadRequest("description must be between 10 and 5000 characters")
	}
	if req.RequesterEmail == "" {
		return nil, errors.BadRequest("requester_email is required")
	}

	// Validate priority enum
	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	} else {
		req.Priority = strings.ToUpper(strings.TrimSpace(req.Priority))
		switch req.Priority {
		case model.PriorityLow, model.PriorityMedium, model.PriorityHigh, model.PriorityUrgent:
		default:
			return nil, errors.BadRequest("invalid priority: must be LOW, MEDIUM, HIGH, or URGENT")
		}
	}

	if req.Category == "" {
		req.Category = "General IT"
	}

	// Generate ticket number
	ticketNumber, err := s.repo.NextTicketNumber(ctx)
	if err != nil {
		return nil, errors.Internal(ctx, "helpdesk generate ticket number", err)
	}

	// Calculate SLA Deadlines and validate service catalog item if specified
	var customResponseMins, customResolutionMins int
	if req.ServiceItemID != nil && *req.ServiceItemID != "" {
		item, err := s.repo.FindServiceCatalogItemByID(ctx, *req.ServiceItemID)
		if err != nil {
			return nil, errors.Internal(ctx, "helpdesk find service catalog item", err)
		}
		if item == nil {
			return nil, errors.BadRequest("service catalog item not found")
		}
		if !item.IsActive {
			return nil, errors.BadRequest("service catalog item is inactive")
		}
		customResponseMins = item.SLAResponseMinutes
		customResolutionMins = item.SLAResolutionMinutes
	}

	respDeadline, resolDeadline := s.slaEngine.CalculateDeadlines(req.Priority, customResponseMins, customResolutionMins)
	initialStatus := model.StatusOpen
	timelineAction := "TICKET_CREATED"
	if req.ServiceItemID != nil && *req.ServiceItemID != "" {
		item, _ := s.repo.FindServiceCatalogItemByID(ctx, *req.ServiceItemID)
		if item != nil && item.RequiresApproval {
			initialStatus = model.StatusWaitingApproval
			timelineAction = "APPROVAL_REQUESTED"
			respDeadline = time.Time{}
			resolDeadline = time.Time{}
		}
	}

	ticket := &model.Ticket{
		TicketNumber:          ticketNumber,
		Title:                 req.Title,
		Description:           req.Description,
		ServiceItemID:         req.ServiceItemID,
		Category:              req.Category,
		Priority:              req.Priority,
		Status:                initialStatus,
		RequesterID:           req.RequesterID,
		RequesterName:         req.RequesterName,
		RequesterEmail:        req.RequesterEmail,
		DepartmentID:          req.DepartmentID,
		AffectedCIID:          req.AffectedCIID,
		SLAResponseDeadline:   respDeadline,
		SLAResolutionDeadline: resolDeadline,
		SLAStatus:             model.SLAWithinSLA,
	}

	timeline := &model.TicketTimeline{
		ActorID:   req.RequesterID,
		ActorName: req.RequesterName,
		Action:    timelineAction,
		NewValue:  &ticket.Status,
	}

	eventData := ticketEventData(ticket)
	eventPayload, _ := json.Marshal(eventData)
	outbox := &model.OutboxEvent{
		EventType: eventbus.TopicTicketCreated,
		Source:    "helpdesk",
		Payload:   string(eventPayload),
	}

	if err := s.repo.CreateTicketWithOutbox(ctx, ticket, timeline, outbox); err != nil {
		return nil, errors.Internal(ctx, "helpdesk create ticket", err)
	}

	// Publish ticket.created event via EventBus immediately if available
	if s.bus != nil {
		_ = s.bus.Publish(ctx, eventbus.Event{
			Source: "helpdesk",
			Type:   eventbus.TopicTicketCreated,
			Data:   eventData,
		})
		if initialStatus == model.StatusWaitingApproval {
			_ = s.bus.Publish(ctx, eventbus.Event{
				Source: "helpdesk",
				Type:   eventbus.TopicApprovalRequested,
				Data: map[string]any{
					"entity_type":     "ticket",
					"entity_id":       ticket.ID,
					"ticket_number":   ticket.TicketNumber,
					"title":           ticket.Title,
					"requester_id":    ticket.RequesterID,
					"requester_email": ticket.RequesterEmail,
				},
			})
		}
	}

	return s.repo.FindTicketByID(ctx, ticket.ID)
}

func (s *ticketService) UpdateStatus(ctx context.Context, id string, req *model.UpdateTicketStatusRequest, actorID, actorName string) (*model.Ticket, error) {
	if req.Version == nil || *req.Version <= 0 {
		return nil, errors.BadRequest("version is required for ticket status updates")
	}
	ticket, err := s.repo.FindTicketByID(ctx, id)
	if err != nil || ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}

	oldStatus := ticket.Status
	newStatus := req.Status

	// 1. Enforce strict ITIL v4 State Machine Transition Rules
	if !model.IsValidTransition(oldStatus, newStatus) {
		return nil, errors.BadRequest(fmt.Sprintf("invalid ticket state transition from '%s' to '%s'", oldStatus, newStatus))
	}

	var resolvedAt, closedAt *time.Time
	now := time.Now()

	if newStatus == model.StatusResolved && ticket.ResolvedAt == nil {
		resolvedAt = &now
	}
	if newStatus == model.StatusClosed && ticket.ClosedAt == nil {
		closedAt = &now
	}

	expectedVersion := req.Version

	timeline := &model.TicketTimeline{
		TicketID:  ticket.ID,
		ActorID:   actorID,
		ActorName: actorName,
		Action:    "STATUS_CHANGED",
		OldValue:  &oldStatus,
		NewValue:  &newStatus,
		Notes:     &req.Notes,
	}

	outboxPayload, _ := json.Marshal(map[string]any{
		"ticket_id": id, "status": newStatus, "old_status": oldStatus,
		"notes": req.Notes, "actor_id": actorID, "actor_name": actorName,
	})
	outbox := &model.OutboxEvent{
		EventType: eventbus.TopicTicketStatusChanged,
		Source:    "helpdesk",
		Payload:   string(outboxPayload),
	}

	err = s.repo.UpdateTicketStatusWithOutbox(ctx, id, newStatus, req.AssigneeID, req.AssigneeName, resolvedAt, closedAt, expectedVersion, timeline, outbox)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		return nil, errors.Internal(ctx, "helpdesk update ticket status", err)
	}

	// Mark first response if transitioning from OPEN to IN_PROGRESS or ASSIGNED
	if newStatus == model.StatusInProgress || newStatus == model.StatusAssigned {
		_ = s.repo.RecordFirstResponse(ctx, id, time.Now())
	}

	updated, err := s.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.bus != nil {
		_ = s.bus.Publish(ctx, eventbus.Event{
			Source: "helpdesk",
			Type:   eventbus.TopicTicketStatusChanged,
			Data:   ticketEventData(updated),
		})
	}

	return updated, nil
}

func ticketEventData(ticket *model.Ticket) map[string]any {
	assigneeID, assigneeName, departmentID := "", "", ""
	if ticket.AssigneeID != nil {
		assigneeID = *ticket.AssigneeID
	}
	if ticket.AssigneeName != nil {
		assigneeName = *ticket.AssigneeName
	}
	if ticket.DepartmentID != nil {
		departmentID = *ticket.DepartmentID
	}
	return map[string]any{
		"ticket_id": ticket.ID, "ticket_number": ticket.TicketNumber,
		"title": ticket.Title, "category": ticket.Category, "priority": ticket.Priority,
		"status": ticket.Status, "requester_id": ticket.RequesterID,
		"requester_name": ticket.RequesterName, "requester_email": ticket.RequesterEmail,
		"reporter_id": ticket.RequesterID, "reporter_email": ticket.RequesterEmail,
		"assignee_id": assigneeID, "assignee_name": assigneeName,
		"department_id": departmentID, "sla_status": ticket.SLAStatus,
		"created_at": ticket.CreatedAt, "responded_at": ticket.RespondedAt,
		"resolved_at": ticket.ResolvedAt,
	}
}

func (s *ticketService) AssignTicket(ctx context.Context, id string, req *model.AssignTicketRequest, actorID, actorName string) (*model.Ticket, error) {
	if req.AssigneeID == "" || req.AssigneeName == "" {
		return nil, errors.BadRequest("assignee_id and assignee_name are required")
	}
	if req.Version == nil || *req.Version <= 0 {
		return nil, errors.BadRequest("version is required for ticket assignment")
	}

	ticket, err := s.repo.FindTicketByID(ctx, id)
	if err != nil || ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}

	oldAssignee := "Unassigned"
	if ticket.AssigneeName != nil {
		oldAssignee = *ticket.AssigneeName
	}

	expectedVersion := req.Version

	timeline := &model.TicketTimeline{
		TicketID:  ticket.ID,
		ActorID:   actorID,
		ActorName: actorName,
		Action:    "ASSIGNED",
		OldValue:  &oldAssignee,
		NewValue:  &req.AssigneeName,
	}

	outboxPayload, _ := json.Marshal(map[string]any{
		"ticket_id": id, "assignee_id": req.AssigneeID, "assignee_name": req.AssigneeName,
	})
	outbox := &model.OutboxEvent{
		EventType: eventbus.TopicTicketAssigned,
		Source:    "helpdesk",
		Payload:   string(outboxPayload),
	}

	err = s.repo.AssignTicketWithOutbox(ctx, id, req.AssigneeID, req.AssigneeName, expectedVersion, timeline, outbox)
	if err != nil {
		if appErr, ok := err.(*errors.AppError); ok {
			return nil, appErr
		}
		return nil, errors.Internal(ctx, "helpdesk assign ticket", err)
	}

	updated, err := s.GetTicket(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.bus != nil {
		_ = s.bus.Publish(ctx, eventbus.Event{
			Source: "helpdesk", Type: eventbus.TopicTicketAssigned, Data: ticketEventData(updated),
		})
	}
	return updated, nil
}

func (s *ticketService) AddComment(ctx context.Context, ticketID string, req *model.AddCommentRequest, authorID, authorName, authorRole string) (*model.TicketComment, error) {
	if req.Content == "" {
		return nil, errors.BadRequest("comment content cannot be empty")
	}

	ticket, err := s.repo.FindTicketByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}

	comment := &model.TicketComment{
		TicketID:   ticketID,
		AuthorID:   authorID,
		AuthorName: authorName,
		AuthorRole: authorRole,
		Content:    req.Content,
		IsInternal: req.IsInternal,
	}

	timeline := &model.TicketTimeline{
		TicketID:  ticketID,
		ActorID:   authorID,
		ActorName: authorName,
		Action:    "COMMENT_ADDED",
	}

	outboxPayload, _ := json.Marshal(map[string]any{
		"ticket_id": ticketID, "author_id": authorID, "author_role": authorRole,
	})
	outbox := &model.OutboxEvent{
		EventType: "ticket.comment_added",
		Source:    "helpdesk",
		Payload:   string(outboxPayload),
	}

	if err := s.repo.AddCommentWithOutbox(ctx, comment, timeline, outbox); err != nil {
		return nil, errors.Internal(ctx, "helpdesk add comment", err)
	}

	// Mark first response if the comment author is an IT Agent or Administrator
	if authorRole == "ROLE_AGENT" || authorRole == "ROLE_ADMIN" {
		_ = s.repo.RecordFirstResponse(ctx, ticketID, time.Now())
	}

	return comment, nil
}

func (s *ticketService) HandleApprovalDecided(ctx context.Context, event eventbus.Event) error {
	data, ok := event.Data.(map[string]any)
	if !ok {
		raw, err := json.Marshal(event.Data)
		if err == nil {
			_ = json.Unmarshal(raw, &data)
		}
	}
	if data == nil {
		return nil
	}

	entityID, _ := data["entity_id"].(string)
	decision, _ := data["decision"].(string)
	actorName, _ := data["actor_name"].(string)
	notes, _ := data["notes"].(string)

	if entityID == "" || decision == "" {
		return nil
	}

	ticket, err := s.repo.FindTicketByID(ctx, entityID)
	if err != nil || ticket == nil {
		return nil
	}

	if ticket.Status != model.StatusWaitingApproval {
		return nil
	}

	now := time.Now()
	if decision == "APPROVED" {
		respDeadline, resolDeadline := s.slaEngine.CalculateDeadlines(ticket.Priority, 0, 0)
		ticket.SLAResponseDeadline = respDeadline
		ticket.SLAResolutionDeadline = resolDeadline

		newOpenStatus := model.StatusOpen
		timelineNotes := fmt.Sprintf("Service catalog request approved by %s. Notes: %s", actorName, notes)
		timeline := &model.TicketTimeline{
			TicketID:  ticket.ID,
			ActorID:   "system",
			ActorName: "Workflow Engine",
			Action:    "APPROVAL_GRANTED",
			OldValue:  &ticket.Status,
			NewValue:  &newOpenStatus,
			Notes:     &timelineNotes,
		}

		payload, _ := json.Marshal(map[string]any{
			"ticket_id": ticket.ID, "status": model.StatusOpen, "decision": decision,
		})
		outbox := &model.OutboxEvent{
			EventType: eventbus.TopicTicketStatusChanged,
			Source:    "helpdesk",
			Payload:   string(payload),
		}

		_ = s.repo.UpdateTicketApprovalWithOutbox(ctx, ticket.ID, model.StatusOpen, &respDeadline, &resolDeadline, nil, timeline, outbox)
	} else if decision == "REJECTED" {
		newClosedStatus := model.StatusClosed
		timelineNotes := fmt.Sprintf("Service catalog request rejected by %s. Reason: %s", actorName, notes)
		timeline := &model.TicketTimeline{
			TicketID:  ticket.ID,
			ActorID:   "system",
			ActorName: "Workflow Engine",
			Action:    "APPROVAL_REJECTED",
			OldValue:  &ticket.Status,
			NewValue:  &newClosedStatus,
			Notes:     &timelineNotes,
		}

		payload, _ := json.Marshal(map[string]any{
			"ticket_id": ticket.ID, "status": model.StatusClosed, "decision": decision,
		})
		outbox := &model.OutboxEvent{
			EventType: eventbus.TopicTicketStatusChanged,
			Source:    "helpdesk",
			Payload:   string(payload),
		}

		_ = s.repo.UpdateTicketApprovalWithOutbox(ctx, ticket.ID, model.StatusClosed, nil, nil, &now, timeline, outbox)
	}

	return nil
}

func (s *ticketService) ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error) {
	return s.repo.ListComments(ctx, ticketID)
}

func (s *ticketService) ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error) {
	return s.repo.ListTimeline(ctx, ticketID)
}

func (s *ticketService) ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error) {
	return s.repo.ListServiceCategories(ctx)
}

func (s *ticketService) ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error) {
	return s.repo.ListServiceCatalogItems(ctx)
}

func (s *ticketService) GetTicketsByAssetID(ctx context.Context, assetID string) ([]model.Ticket, error) {
	if assetID == "" {
		return nil, errors.BadRequest("asset id is required")
	}
	return s.repo.ListTicketsByAssetID(ctx, assetID)
}

func (s *ticketService) GetTicketsByAssetIDForActor(ctx context.Context, assetID string, actor middleware.Actor) ([]model.Ticket, error) {
	if assetID == "" {
		return nil, errors.BadRequest("asset id is required")
	}
	if !actor.IsValid() || actor.IsEmployee() {
		return nil, errors.Forbidden("asset-linked ticket lookup requires an operator role")
	}
	if actor.IsManager() && actor.DepartmentID == "" {
		return nil, errors.Forbidden("manager department scope is required")
	}
	return s.repo.ListTicketsByAssetIDForActor(ctx, assetID, actor)
}
