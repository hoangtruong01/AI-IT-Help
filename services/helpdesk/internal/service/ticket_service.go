package service

import (
	"context"
	"fmt"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/repository"
)

// TicketService defines ITSM ticket business logic
type TicketService interface {
	ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error)
	GetTicket(ctx context.Context, id string) (*model.Ticket, error)
	CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error)
	UpdateStatus(ctx context.Context, id string, req *model.UpdateTicketStatusRequest, actorID, actorName string) (*model.Ticket, error)
	AssignTicket(ctx context.Context, id string, req *model.AssignTicketRequest, actorID, actorName string) (*model.Ticket, error)

	AddComment(ctx context.Context, ticketID string, req *model.AddCommentRequest, authorID, authorName, authorRole string) (*model.TicketComment, error)
	ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error)
	ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error)

	ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error)
	ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error)
}

type ticketService struct {
	repo      repository.Repository
	slaEngine SLAEngine
}

// NewTicketService constructs a new TicketService
func NewTicketService(repo repository.Repository, slaEngine SLAEngine) TicketService {
	return &ticketService{
		repo:      repo,
		slaEngine: slaEngine,
	}
}

func (s *ticketService) ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error) {
	resp, err := s.repo.ListTickets(ctx, query)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to list tickets: %v", err))
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
		return nil, errors.InternalServerError(fmt.Sprintf("failed to get ticket: %v", err))
	}
	if ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}

	ticket.SLAStatus = s.slaEngine.EvaluateSLAStatus(ticket)
	return ticket, nil
}

func (s *ticketService) CreateTicket(ctx context.Context, req *model.CreateTicketRequest) (*model.Ticket, error) {
	if req.Title == "" || req.Description == "" || req.RequesterEmail == "" {
		return nil, errors.BadRequest("title, description, and requester_email are required")
	}

	if req.Priority == "" {
		req.Priority = model.PriorityMedium
	}
	if req.Category == "" {
		req.Category = "General IT"
	}

	// Generate ticket number
	ticketNumber, err := s.repo.NextTicketNumber(ctx)
	if err != nil {
		return nil, errors.InternalServerError("failed to generate ticket number")
	}

	// Calculate SLA Deadlines
	var customResponseMins, customResolutionMins int
	if req.ServiceItemID != nil && *req.ServiceItemID != "" {
		item, _ := s.repo.FindServiceCatalogItemByID(ctx, *req.ServiceItemID)
		if item != nil {
			customResponseMins = item.SLAResponseMinutes
			customResolutionMins = item.SLAResolutionMinutes
		}
	}

	respDeadline, resolDeadline := s.slaEngine.CalculateDeadlines(req.Priority, customResponseMins, customResolutionMins)

	ticket := &model.Ticket{
		TicketNumber:          ticketNumber,
		Title:                 req.Title,
		Description:           req.Description,
		ServiceItemID:         req.ServiceItemID,
		Category:              req.Category,
		Priority:              req.Priority,
		Status:                model.StatusOpen,
		RequesterID:           req.RequesterID,
		RequesterName:         req.RequesterName,
		RequesterEmail:        req.RequesterEmail,
		DepartmentID:          req.DepartmentID,
		AffectedCIID:          req.AffectedCIID,
		SLAResponseDeadline:   respDeadline,
		SLAResolutionDeadline: resolDeadline,
		SLAStatus:             model.SLAWithinSLA,
	}

	if err := s.repo.CreateTicket(ctx, ticket); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to create ticket: %v", err))
	}

	// Record creation in timeline
	_ = s.repo.AddTimelineRecord(ctx, &model.TicketTimeline{
		TicketID:  ticket.ID,
		ActorID:   req.RequesterID,
		ActorName: req.RequesterName,
		Action:    "TICKET_CREATED",
		NewValue:  &ticket.Status,
	})

	return s.repo.FindTicketByID(ctx, ticket.ID)
}

func (s *ticketService) UpdateStatus(ctx context.Context, id string, req *model.UpdateTicketStatusRequest, actorID, actorName string) (*model.Ticket, error) {
	ticket, err := s.repo.FindTicketByID(ctx, id)
	if err != nil || ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}

	oldStatus := ticket.Status
	newStatus := req.Status

	var resolvedAt, closedAt *time.Time
	now := time.Now()

	if newStatus == model.StatusResolved && ticket.ResolvedAt == nil {
		resolvedAt = &now
	}
	if newStatus == model.StatusClosed && ticket.ClosedAt == nil {
		closedAt = &now
	}

	err = s.repo.UpdateTicketStatus(ctx, id, newStatus, req.AssigneeID, req.AssigneeName, resolvedAt, closedAt)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to update ticket status: %v", err))
	}

	// Add timeline entry
	_ = s.repo.AddTimelineRecord(ctx, &model.TicketTimeline{
		TicketID:  ticket.ID,
		ActorID:   actorID,
		ActorName: actorName,
		Action:    "STATUS_CHANGED",
		OldValue:  &oldStatus,
		NewValue:  &newStatus,
		Notes:     &req.Notes,
	})

	return s.GetTicket(ctx, id)
}

func (s *ticketService) AssignTicket(ctx context.Context, id string, req *model.AssignTicketRequest, actorID, actorName string) (*model.Ticket, error) {
	if req.AssigneeID == "" || req.AssigneeName == "" {
		return nil, errors.BadRequest("assignee_id and assignee_name are required")
	}

	ticket, err := s.repo.FindTicketByID(ctx, id)
	if err != nil || ticket == nil {
		return nil, errors.NotFound("ticket not found")
	}

	oldAssignee := "Unassigned"
	if ticket.AssigneeName != nil {
		oldAssignee = *ticket.AssigneeName
	}

	err = s.repo.AssignTicket(ctx, id, req.AssigneeID, req.AssigneeName)
	if err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to assign ticket: %v", err))
	}

	_ = s.repo.AddTimelineRecord(ctx, &model.TicketTimeline{
		TicketID:  ticket.ID,
		ActorID:   actorID,
		ActorName: actorName,
		Action:    "ASSIGNED",
		OldValue:  &oldAssignee,
		NewValue:  &req.AssigneeName,
	})

	return s.GetTicket(ctx, id)
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

	if err := s.repo.AddComment(ctx, comment); err != nil {
		return nil, errors.InternalServerError(fmt.Sprintf("failed to add comment: %v", err))
	}

	_ = s.repo.AddTimelineRecord(ctx, &model.TicketTimeline{
		TicketID:  ticketID,
		ActorID:   authorID,
		ActorName: authorName,
		Action:    "COMMENT_ADDED",
	})

	return comment, nil
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
