package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"eomp/services/workflow/internal/model"
	"eomp/services/workflow/internal/repository"
)

func newUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// ChangeService defines business logic for ITIL Change Management & CAB.
type ChangeService interface {
	ListChanges(ctx context.Context, changeType, status, risk string, page, pageSize int) ([]model.ChangeRequest, int, error)
	GetChange(ctx context.Context, id string) (*model.ChangeRequest, []model.CABReview, error)
	CreateChange(ctx context.Context, payload model.CreateChangePayload) (*model.ChangeRequest, error)
	UpdateStatus(ctx context.Context, id string, payload model.UpdateChangeStatusPayload) (*model.ChangeRequest, error)
	SubmitCABVote(ctx context.Context, changeID string, payload model.SubmitCABVotePayload) (*model.CABReview, *model.ChangeRequest, error)
	GetCalendar(ctx context.Context, startDate, endDate time.Time) ([]model.ChangeCalendarItem, error)
	GetStats(ctx context.Context) (*model.ChangeStats, error)
}

type changeService struct {
	repo repository.ChangeRepository
}

// NewChangeService constructs a new ChangeService instance.
func NewChangeService(repo repository.ChangeRepository) ChangeService {
	return &changeService{repo: repo}
}

func (s *changeService) calculateRiskLevel(probability, impact string) string {
	prob := probability
	imp := impact

	if imp == "CRITICAL" || (imp == "HIGH" && prob == "HIGH") {
		return "CRITICAL"
	}
	if (imp == "HIGH" && prob == "MEDIUM") || (imp == "MEDIUM" && prob == "HIGH") {
		return "HIGH"
	}
	if (imp == "MEDIUM" && prob == "MEDIUM") || (imp == "HIGH" && prob == "LOW") || (imp == "LOW" && prob == "HIGH") {
		return "MEDIUM"
	}
	return "LOW"
}

func (s *changeService) ListChanges(ctx context.Context, changeType, status, risk string, page, pageSize int) ([]model.ChangeRequest, int, error) {
	return s.repo.ListChanges(ctx, changeType, status, risk, page, pageSize)
}

func (s *changeService) GetChange(ctx context.Context, id string) (*model.ChangeRequest, []model.CABReview, error) {
	change, err := s.repo.GetChangeByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	reviews, err := s.repo.GetCABReviews(ctx, change.ID)
	if err != nil {
		reviews = []model.CABReview{}
	}

	return change, reviews, nil
}

func (s *changeService) CreateChange(ctx context.Context, payload model.CreateChangePayload) (*model.ChangeRequest, error) {
	if payload.Title == "" {
		return nil, errors.New("change title is required")
	}
	if payload.ReasonForChange == "" {
		return nil, errors.New("reason for change is required")
	}
	if payload.ImplementationPlan == "" {
		return nil, errors.New("implementation plan is required")
	}
	if payload.RollbackPlan == "" {
		return nil, errors.New("rollback plan is required")
	}

	changeNum, err := s.repo.NextChangeNumber(ctx)
	if err != nil {
		changeNum = fmt.Sprintf("CHG-%d", time.Now().Unix()%10000)
	}

	changeType := payload.ChangeType
	if changeType == "" {
		changeType = "NORMAL"
	}

	category := payload.Category
	if category == "" {
		category = "Infrastructure"
	}

	priority := payload.Priority
	if priority == "" {
		priority = "MEDIUM"
	}

	impact := payload.ImpactLevel
	if impact == "" {
		impact = "MEDIUM"
	}
	probability := payload.ProbabilityLevel
	if probability == "" {
		probability = "MEDIUM"
	}

	riskLevel := s.calculateRiskLevel(probability, impact)

	// Determine CAB required count based on ITIL change classification
	cabRequired := 1
	initialStatus := "SUBMITTED"
	if changeType == "STANDARD" {
		cabRequired = 0
		initialStatus = "APPROVED" // Standard changes are pre-approved
	} else if changeType == "EMERGENCY" || changeType == "MAJOR" {
		cabRequired = 2
		initialStatus = "CAB_REVIEW"
	} else if changeType == "NORMAL" {
		cabRequired = 1
		initialStatus = "CAB_REVIEW"
	}

	now := time.Now()
	c := &model.ChangeRequest{
		ID:                 newUUID(),
		ChangeNumber:       changeNum,
		Title:              payload.Title,
		Description:        payload.Description,
		ChangeType:         changeType,
		Category:           category,
		Priority:           priority,
		RiskLevel:          riskLevel,
		ImpactLevel:        impact,
		ProbabilityLevel:   probability,
		Status:             initialStatus,
		RequesterID:        payload.RequesterID,
		RequesterName:      payload.RequesterName,
		RequesterEmail:     payload.RequesterEmail,
		AssignedToID:       payload.AssignedToID,
		AssignedToName:     payload.AssignedToName,
		ReasonForChange:    payload.ReasonForChange,
		ImplementationPlan: payload.ImplementationPlan,
		RollbackPlan:       payload.RollbackPlan,
		TestPlan:           payload.TestPlan,
		ScheduledStartTime: payload.ScheduledStartTime,
		ScheduledEndTime:   payload.ScheduledEndTime,
		DowntimeRequired:   payload.DowntimeRequired,
		DowntimeMinutes:    payload.DowntimeMinutes,
		CABRequiredCount:   cabRequired,
		CABApprovedCount:   0,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.repo.CreateChange(ctx, c); err != nil {
		return nil, err
	}

	return s.repo.GetChangeByID(ctx, c.ID)
}

func (s *changeService) UpdateStatus(ctx context.Context, id string, payload model.UpdateChangeStatusPayload) (*model.ChangeRequest, error) {
	change, err := s.repo.GetChangeByID(ctx, id)
	if err != nil {
		return nil, err
	}

	targetStatus := payload.Status

	// ITIL Rule & Test Case 7.2: CAB Quorum Enforcement
	// If moving to APPROVED, SCHEDULED, or IMPLEMENTING, verify CAB approvals
	if targetStatus == "APPROVED" || targetStatus == "SCHEDULED" || targetStatus == "IMPLEMENTING" {
		if change.ChangeType == "EMERGENCY" || change.ChangeType == "MAJOR" {
			if change.CABApprovedCount < change.CABRequiredCount {
				return nil, fmt.Errorf("insufficient CAB approvals: %d of %d required for %s change", change.CABApprovedCount, change.CABRequiredCount, change.ChangeType)
			}
		} else if change.ChangeType == "NORMAL" {
			if change.CABApprovedCount < 1 {
				return nil, fmt.Errorf("insufficient CAB approvals: 1 approval required for NORMAL change")
			}
		}
	}

	now := time.Now()
	var actualStart, actualEnd *time.Time
	if targetStatus == "IMPLEMENTING" && change.ActualStartTime == nil {
		actualStart = &now
	} else if (targetStatus == "COMPLETED" || targetStatus == "FAILED") && change.ActualEndTime == nil {
		actualEnd = &now
	}

	if err := s.repo.UpdateChangeStatus(ctx, change.ID, targetStatus, actualStart, actualEnd); err != nil {
		return nil, err
	}

	return s.repo.GetChangeByID(ctx, change.ID)
}

func (s *changeService) SubmitCABVote(ctx context.Context, changeID string, payload model.SubmitCABVotePayload) (*model.CABReview, *model.ChangeRequest, error) {
	change, err := s.repo.GetChangeByID(ctx, changeID)
	if err != nil {
		return nil, nil, err
	}

	if payload.ReviewerID == "" || payload.ReviewerName == "" {
		return nil, nil, errors.New("reviewer ID and Name are required")
	}
	if payload.Vote != "APPROVED" && payload.Vote != "REJECTED" && payload.Vote != "ABSTAIN" {
		return nil, nil, errors.New("vote must be APPROVED, REJECTED, or ABSTAIN")
	}

	review := &model.CABReview{
		ID:           newUUID(),
		ChangeID:     change.ID,
		ReviewerID:   payload.ReviewerID,
		ReviewerName: payload.ReviewerName,
		ReviewerRole: payload.ReviewerRole,
		Vote:         payload.Vote,
		Comments:     payload.Comments,
		ReviewedAt:   time.Now(),
	}

	if err := s.repo.AddCABReview(ctx, review); err != nil {
		return nil, nil, err
	}

	// Recalculate approval count
	approvedCount, err := s.repo.RecalculateCABApprovedCount(ctx, change.ID)
	if err == nil {
		change.CABApprovedCount = approvedCount
	}

	// If quorum reached and still in CAB_REVIEW, auto-transition to APPROVED
	if change.Status == "CAB_REVIEW" && change.CABApprovedCount >= change.CABRequiredCount {
		_ = s.repo.UpdateChangeStatus(ctx, change.ID, "APPROVED", nil, nil)
	}

	updatedChange, _ := s.repo.GetChangeByID(ctx, change.ID)
	return review, updatedChange, nil
}

func (s *changeService) GetCalendar(ctx context.Context, startDate, endDate time.Time) ([]model.ChangeCalendarItem, error) {
	if startDate.IsZero() {
		startDate = time.Now().AddDate(0, 0, -30)
	}
	if endDate.IsZero() {
		endDate = time.Now().AddDate(0, 2, 0)
	}
	return s.repo.GetChangeCalendar(ctx, startDate, endDate)
}

func (s *changeService) GetStats(ctx context.Context) (*model.ChangeStats, error) {
	return s.repo.GetChangeStats(ctx)
}
