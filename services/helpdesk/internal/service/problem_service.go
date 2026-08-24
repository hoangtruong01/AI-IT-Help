package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/repository"
)

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ProblemService handles business logic for ITIL Problem Management.
type ProblemService interface {
	ListProblems(ctx context.Context, category, status string, page, pageSize int) ([]model.Problem, int, error)
	GetProblem(ctx context.Context, id string) (*model.Problem, []model.ProblemIncidentLink, error)
	CreateProblem(ctx context.Context, payload model.CreateProblemPayload) (*model.Problem, error)
	UpdateStatus(ctx context.Context, id string, payload model.UpdateProblemStatusPayload) (*model.Problem, []string, error)
	UpdateRCA(ctx context.Context, id string, payload model.UpdateProblemRCAPayload) (*model.Problem, error)
	LinkIncident(ctx context.Context, problemID string, payload model.LinkIncidentPayload) (*model.ProblemIncidentLink, error)
	UnlinkIncident(ctx context.Context, problemID, ticketID string) error
	GetStats(ctx context.Context) (*model.ProblemStats, error)
}

type problemService struct {
	problemRepo repository.ProblemRepository
	ticketRepo  repository.Repository
}

// NewProblemService creates a new instance of ProblemService.
func NewProblemService(problemRepo repository.ProblemRepository, ticketRepo repository.Repository) ProblemService {
	return &problemService{
		problemRepo: problemRepo,
		ticketRepo:  ticketRepo,
	}
}

func (s *problemService) ListProblems(ctx context.Context, category, status string, page, pageSize int) ([]model.Problem, int, error) {
	return s.problemRepo.ListProblems(ctx, category, status, page, pageSize)
}

func (s *problemService) GetProblem(ctx context.Context, id string) (*model.Problem, []model.ProblemIncidentLink, error) {
	problem, err := s.problemRepo.GetProblemByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	links, err := s.problemRepo.GetLinkedIncidents(ctx, problem.ID)
	if err != nil {
		links = []model.ProblemIncidentLink{}
	}

	return problem, links, nil
}

func (s *problemService) CreateProblem(ctx context.Context, payload model.CreateProblemPayload) (*model.Problem, error) {
	if payload.Title == "" {
		return nil, errors.New("problem title is required")
	}

	num, err := s.problemRepo.NextProblemNumber(ctx)
	if err != nil {
		num = fmt.Sprintf("PRB-%d", time.Now().Unix()%10000)
	}

	now := time.Now()
	category := payload.Category
	if category == "" {
		category = "Infrastructure"
	}
	priority := payload.Priority
	if priority == "" {
		priority = "HIGH"
	}
	impact := payload.Impact
	if impact == "" {
		impact = "HIGH"
	}
	urgency := payload.Urgency
	if urgency == "" {
		urgency = "HIGH"
	}

	p := &model.Problem{
		ID:            newUUID(),
		ProblemNumber: num,
		Title:         payload.Title,
		Description:   payload.Description,
		Category:      category,
		Priority:      priority,
		Status:        "OPEN",
		Impact:        impact,
		Urgency:       urgency,
		AssigneeID:    payload.AssigneeID,
		AssigneeName:  payload.AssigneeName,
		RootCause:     payload.RootCause,
		Workaround:    payload.Workaround,
		IsKnownError:  payload.IsKnownError,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.problemRepo.CreateProblem(ctx, p); err != nil {
		return nil, err
	}

	// Link any initial tickets provided
	for _, ticketID := range payload.TicketIDs {
		if ticketID == "" {
			continue
		}
		// Look up ticket details
		ticket, err := s.ticketRepo.FindTicketByID(ctx, ticketID)
		if err == nil && ticket != nil {
			link := &model.ProblemIncidentLink{
				ID:           newUUID(),
				ProblemID:    p.ID,
				TicketID:     ticket.ID,
				TicketNumber: ticket.TicketNumber,
				TicketTitle:  ticket.Title,
				LinkedBy:     "Problem Manager",
				LinkedAt:     now,
			}
			_ = s.problemRepo.LinkIncident(ctx, link)
		}
	}

	return s.problemRepo.GetProblemByID(ctx, p.ID)
}

func (s *problemService) UpdateStatus(ctx context.Context, id string, payload model.UpdateProblemStatusPayload) (*model.Problem, []string, error) {
	problem, err := s.problemRepo.GetProblemByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	var resolvedAt, closedAt *time.Time
	if payload.Status == "RESOLVED" {
		resolvedAt = &now
	} else if payload.Status == "CLOSED" {
		closedAt = &now
	}

	if err := s.problemRepo.UpdateProblemStatus(ctx, problem.ID, payload.Status, payload.Resolution, resolvedAt, closedAt); err != nil {
		return nil, nil, err
	}

	var cascadedTickets []string
	// ITIL Business Rule & Test Case 7.1: If Problem transitions to RESOLVED, cascade resolution to all linked incidents
	if payload.Status == "RESOLVED" {
		resolutionText := "Root cause identified and permanent fix verified."
		if payload.Resolution != nil && *payload.Resolution != "" {
			resolutionText = *payload.Resolution
		}
		cascaded, err := s.problemRepo.CascadeResolveLinkedTickets(ctx, problem.ID, problem.ProblemNumber, resolutionText)
		if err == nil {
			cascadedTickets = cascaded
		}
	}

	updated, err := s.problemRepo.GetProblemByID(ctx, problem.ID)
	return updated, cascadedTickets, err
}

func (s *problemService) UpdateRCA(ctx context.Context, id string, payload model.UpdateProblemRCAPayload) (*model.Problem, error) {
	problem, err := s.problemRepo.GetProblemByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.problemRepo.UpdateProblemRCA(ctx, problem.ID, payload.RootCause, payload.Workaround, payload.IsKnownError); err != nil {
		return nil, err
	}

	return s.problemRepo.GetProblemByID(ctx, problem.ID)
}

func (s *problemService) LinkIncident(ctx context.Context, problemID string, payload model.LinkIncidentPayload) (*model.ProblemIncidentLink, error) {
	problem, err := s.problemRepo.GetProblemByID(ctx, problemID)
	if err != nil {
		return nil, err
	}

	ticket, err := s.ticketRepo.FindTicketByID(ctx, payload.TicketID)
	if err != nil || ticket == nil {
		// Try finding by ticket number
		ticket, err = s.ticketRepo.FindTicketByNumber(ctx, payload.TicketID)
		if err != nil || ticket == nil {
			return nil, fmt.Errorf("incident ticket %s not found", payload.TicketID)
		}
	}

	linkedBy := payload.LinkedBy
	if linkedBy == "" {
		linkedBy = "IT Support Agent"
	}

	link := &model.ProblemIncidentLink{
		ID:           newUUID(),
		ProblemID:    problem.ID,
		TicketID:     ticket.ID,
		TicketNumber: ticket.TicketNumber,
		TicketTitle:  ticket.Title,
		LinkedBy:     linkedBy,
		LinkedAt:     time.Now(),
	}

	if err := s.problemRepo.LinkIncident(ctx, link); err != nil {
		return nil, err
	}

	return link, nil
}

func (s *problemService) UnlinkIncident(ctx context.Context, problemID, ticketID string) error {
	problem, err := s.problemRepo.GetProblemByID(ctx, problemID)
	if err != nil {
		return err
	}
	return s.problemRepo.UnlinkIncident(ctx, problem.ID, ticketID)
}

func (s *problemService) GetStats(ctx context.Context) (*model.ProblemStats, error) {
	return s.problemRepo.GetProblemStats(ctx)
}
