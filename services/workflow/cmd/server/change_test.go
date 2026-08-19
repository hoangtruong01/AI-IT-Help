package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"eomp/services/workflow/internal/model"
	"eomp/services/workflow/internal/service"
)

// mockChangeRepo implements repository.ChangeRepository for testing.
type mockChangeRepo struct {
	changes map[string]*model.ChangeRequest
	reviews map[string][]model.CABReview
}

func newMockChangeRepo() *mockChangeRepo {
	return &mockChangeRepo{
		changes: make(map[string]*model.ChangeRequest),
		reviews: make(map[string][]model.CABReview),
	}
}

func (m *mockChangeRepo) NextChangeNumber(ctx context.Context) (string, error) {
	return "CHG-2001", nil
}

func (m *mockChangeRepo) ListChanges(ctx context.Context, changeType, status, risk string, page, pageSize int) ([]model.ChangeRequest, int, error) {
	var list []model.ChangeRequest
	for _, c := range m.changes {
		if (changeType == "" || changeType == "All" || c.ChangeType == changeType) &&
			(status == "" || status == "All" || c.Status == status) &&
			(risk == "" || risk == "All" || c.RiskLevel == risk) {
			list = append(list, *c)
		}
	}
	return list, len(list), nil
}

func (m *mockChangeRepo) GetChangeByID(ctx context.Context, id string) (*model.ChangeRequest, error) {
	for _, c := range m.changes {
		if c.ID == id || c.ChangeNumber == id {
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}

func (m *mockChangeRepo) CreateChange(ctx context.Context, c *model.ChangeRequest) error {
	m.changes[c.ID] = c
	return nil
}

func (m *mockChangeRepo) UpdateChange(ctx context.Context, c *model.ChangeRequest) error {
	m.changes[c.ID] = c
	return nil
}

func (m *mockChangeRepo) UpdateChangeStatus(ctx context.Context, id, status string, actualStart, actualEnd *time.Time) error {
	if c, ok := m.changes[id]; ok {
		c.Status = status
		c.ActualStartTime = actualStart
		c.ActualEndTime = actualEnd
	}
	return nil
}

func (m *mockChangeRepo) AddCABReview(ctx context.Context, r *model.CABReview) error {
	m.reviews[r.ChangeID] = append(m.reviews[r.ChangeID], *r)
	return nil
}

func (m *mockChangeRepo) GetCABReviews(ctx context.Context, changeID string) ([]model.CABReview, error) {
	return m.reviews[changeID], nil
}

func (m *mockChangeRepo) RecalculateCABApprovedCount(ctx context.Context, changeID string) (int, error) {
	count := 0
	for _, r := range m.reviews[changeID] {
		if r.Vote == "APPROVED" {
			count++
		}
	}
	if c, ok := m.changes[changeID]; ok {
		c.CABApprovedCount = count
	}
	return count, nil
}

func (m *mockChangeRepo) GetChangeCalendar(ctx context.Context, startDate, endDate time.Time) ([]model.ChangeCalendarItem, error) {
	var items []model.ChangeCalendarItem
	for _, c := range m.changes {
		items = append(items, model.ChangeCalendarItem{
			ID:             c.ID,
			ChangeNumber:   c.ChangeNumber,
			Title:          c.Title,
			ChangeType:     c.ChangeType,
			Category:       c.Category,
			RiskLevel:      c.RiskLevel,
			Status:         c.Status,
			ScheduledStart: c.ScheduledStartTime,
			ScheduledEnd:   c.ScheduledEndTime,
		})
	}
	return items, nil
}

func (m *mockChangeRepo) GetChangeStats(ctx context.Context) (*model.ChangeStats, error) {
	return &model.ChangeStats{
		ActiveChanges:      1,
		PendingCABReview:   1,
		EmergencyChanges:   1,
		SuccessRatePercent: 100.0,
		TotalThisMonth:     4,
	}, nil
}

// Test Case 7.2: Verify CAB Quorum Enforcement for Emergency / Major Change Requests.
func TestChangeManagement_TestCase_7_2(t *testing.T) {
	ctx := context.Background()
	repo := newMockChangeRepo()
	svc := service.NewChangeService(repo)

	// 1. Create an EMERGENCY Change Request
	c, err := svc.CreateChange(ctx, model.CreateChangePayload{
		Title:              "Emergency Zero-Day Patching on Border Gateway Router",
		Description:        "Apply vendor patch for CVE-2026-8819 vulnerability.",
		ChangeType:         "EMERGENCY",
		Category:           "Network & Access",
		Priority:           "CRITICAL",
		ImpactLevel:        "CRITICAL",
		ProbabilityLevel:   "LOW",
		RequesterID:        "u1",
		RequesterName:      "Sarah Jenkins",
		RequesterEmail:     "sarah@eomp.local",
		ReasonForChange:    "Zero-day remote code execution vulnerability.",
		ImplementationPlan: "1. VRRP failover 2. Flash firmware 3. Reload",
		RollbackPlan:       "Revert to backup firmware image",
		TestPlan:           "Verify BGP route stability",
	})
	if err != nil {
		t.Fatalf("failed to create change: %v", err)
	}

	if c.CABRequiredCount != 2 {
		t.Fatalf("expected CABRequiredCount = 2 for EMERGENCY change, got %d", c.CABRequiredCount)
	}
	if c.Status != "CAB_REVIEW" {
		t.Fatalf("expected status = CAB_REVIEW, got %s", c.Status)
	}

	// 2. Try to transition to IMPLEMENTING with 0 approvals -> Must fail
	_, err = svc.UpdateStatus(ctx, c.ID, model.UpdateChangeStatusPayload{Status: "IMPLEMENTING"})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "insufficient cab") {
		t.Fatalf("expected Insufficient CAB approvals error, got: %v", err)
	}

	// 3. Submit 1st CAB Approval (Reviewer 1)
	_, updated, err := svc.SubmitCABVote(ctx, c.ID, model.SubmitCABVotePayload{
		ReviewerID:   "cab-1",
		ReviewerName: "Sarah Jenkins (Security Officer)",
		ReviewerRole: "Security Lead",
		Vote:         "APPROVED",
	})
	if err != nil {
		t.Fatalf("failed to submit first CAB vote: %v", err)
	}
	if updated.CABApprovedCount != 1 {
		t.Fatalf("expected CABApprovedCount = 1, got %d", updated.CABApprovedCount)
	}

	// 4. Try to transition to IMPLEMENTING with only 1 approval -> Must still fail
	_, err = svc.UpdateStatus(ctx, c.ID, model.UpdateChangeStatusPayload{Status: "IMPLEMENTING"})
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "insufficient cab") {
		t.Fatalf("expected Insufficient CAB approvals error with 1 vote, got: %v", err)
	}

	// 5. Submit 2nd CAB Approval (Reviewer 2 - Quorum Reached)
	_, updated2, err := svc.SubmitCABVote(ctx, c.ID, model.SubmitCABVotePayload{
		ReviewerID:   "cab-2",
		ReviewerName: "Alex Rivera (Infrastructure Lead)",
		ReviewerRole: "Network Lead",
		Vote:         "APPROVED",
	})
	if err != nil {
		t.Fatalf("failed to submit second CAB vote: %v", err)
	}
	if updated2.CABApprovedCount != 2 {
		t.Fatalf("expected CABApprovedCount = 2, got %d", updated2.CABApprovedCount)
	}
	if updated2.Status != "APPROVED" {
		t.Fatalf("expected status auto-promoted to APPROVED, got %s", updated2.Status)
	}

	// 6. Transition to IMPLEMENTING -> Must succeed!
	implementingChange, err := svc.UpdateStatus(ctx, c.ID, model.UpdateChangeStatusPayload{Status: "IMPLEMENTING"})
	if err != nil {
		t.Fatalf("failed to transition to IMPLEMENTING after quorum reached: %v", err)
	}
	if implementingChange.Status != "IMPLEMENTING" {
		t.Errorf("expected status IMPLEMENTING, got %s", implementingChange.Status)
	}
}
