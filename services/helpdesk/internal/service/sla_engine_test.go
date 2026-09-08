package service

import (
	"testing"
	"time"

	"eomp/services/helpdesk/internal/model"
)

func TestCalculateDeadlines(t *testing.T) {
	engine := NewSLAEngine()

	// High priority: 30m response, 4h resolution
	before := time.Now()
	respDeadline, resolDeadline := engine.CalculateDeadlines(model.PriorityHigh, 0, 0)
	after := time.Now()

	expectedResp := 30 * time.Minute
	expectedResol := 4 * time.Hour

	actualRespDur := respDeadline.Sub(before)
	actualResolDur := resolDeadline.Sub(before)

	if actualRespDur < expectedResp || actualRespDur > expectedResp+time.Second {
		t.Errorf("expected response deadline ~%v, got %v", expectedResp, actualRespDur)
	}
	if actualResolDur < expectedResol || actualResolDur > expectedResol+time.Second {
		t.Errorf("expected resolution deadline ~%v, got %v", expectedResol, actualResolDur)
	}

	// Custom SLA minutes
	respDeadlineCustom, resolDeadlineCustom := engine.CalculateDeadlines("CUSTOM", 15, 60)
	actualCustomResp := respDeadlineCustom.Sub(after)
	actualCustomResol := resolDeadlineCustom.Sub(after)

	if actualCustomResp < 15*time.Minute || actualCustomResp > 16*time.Minute {
		t.Errorf("expected custom response deadline ~15m, got %v", actualCustomResp)
	}
	if actualCustomResol < 60*time.Minute || actualCustomResol > 61*time.Minute {
		t.Errorf("expected custom resolution deadline ~60m, got %v", actualCustomResol)
	}
}

func TestEvaluateSLAStatus_ResponseBreached_Unresponded(t *testing.T) {
	engine := NewSLAEngine()
	now := time.Now()

	// Ticket created 1 hour ago with 30m response deadline, unresponded
	ticket := &model.Ticket{
		Status:                model.StatusOpen,
		CreatedAt:             now.Add(-1 * time.Hour),
		SLAResponseDeadline:   now.Add(-30 * time.Minute), // Breached 30m ago
		SLAResolutionDeadline: now.Add(3 * time.Hour),     // Resolution still valid
		RespondedAt:           nil,
	}

	status := engine.EvaluateSLAStatus(ticket)
	if status != model.SLABreached {
		t.Fatalf("expected SLA status '%s' for overdue unresponded ticket, got '%s'", model.SLABreached, status)
	}
}

func TestEvaluateSLAStatus_ResponseBreached_LateResponse(t *testing.T) {
	engine := NewSLAEngine()
	now := time.Now()

	lateResp := now.Add(-20 * time.Minute)
	ticket := &model.Ticket{
		Status:                model.StatusInProgress,
		CreatedAt:             now.Add(-2 * time.Hour),
		SLAResponseDeadline:   now.Add(-1 * time.Hour), // Deadline was 1h ago
		SLAResolutionDeadline: now.Add(2 * time.Hour),
		RespondedAt:           &lateResp, // Responded 40 minutes late!
	}

	status := engine.EvaluateSLAStatus(ticket)
	if status != model.SLABreached {
		t.Fatalf("expected SLA status '%s' for late response, got '%s'", model.SLABreached, status)
	}
}

func TestEvaluateSLAStatus_ResponseWithin_OnTimeResponse(t *testing.T) {
	engine := NewSLAEngine()
	now := time.Now()

	onTimeResp := now.Add(-50 * time.Minute)
	ticket := &model.Ticket{
		Status:                model.StatusInProgress,
		CreatedAt:             now.Add(-1 * time.Hour),
		SLAResponseDeadline:   now.Add(-30 * time.Minute), // Deadline was 30m ago
		SLAResolutionDeadline: now.Add(3 * time.Hour),
		RespondedAt:           &onTimeResp, // Responded 50m ago, which was 20m before deadline!
	}

	status := engine.EvaluateSLAStatus(ticket)
	if status != model.SLAWithinSLA {
		t.Fatalf("expected SLA status '%s' for on-time response, got '%s'", model.SLAWithinSLA, status)
	}
}

func TestEvaluateSLAStatus_ResolutionBreached(t *testing.T) {
	engine := NewSLAEngine()
	now := time.Now()

	onTimeResp := now.Add(-3 * time.Hour)
	ticket := &model.Ticket{
		Status:                model.StatusInProgress,
		CreatedAt:             now.Add(-5 * time.Hour),
		SLAResponseDeadline:   now.Add(-4 * time.Hour),
		SLAResolutionDeadline: now.Add(-1 * time.Hour), // Breached 1 hour ago!
		RespondedAt:           &onTimeResp,
	}

	status := engine.EvaluateSLAStatus(ticket)
	if status != model.SLABreached {
		t.Fatalf("expected SLA status '%s' for breached resolution, got '%s'", model.SLABreached, status)
	}
}

func TestEvaluateSLAStatus_ResolvedTicket_Late(t *testing.T) {
	engine := NewSLAEngine()
	now := time.Now()

	onTimeResp := now.Add(-4 * time.Hour)
	lateResolved := now.Add(-30 * time.Minute)
	ticket := &model.Ticket{
		Status:                model.StatusResolved,
		CreatedAt:             now.Add(-5 * time.Hour),
		SLAResponseDeadline:   now.Add(-4 * time.Hour),
		SLAResolutionDeadline: now.Add(-1 * time.Hour),
		RespondedAt:           &onTimeResp,
		ResolvedAt:            &lateResolved, // Resolved 30m late
	}

	status := engine.EvaluateSLAStatus(ticket)
	if status != model.SLABreached {
		t.Fatalf("expected SLA status '%s' for late resolved ticket, got '%s'", model.SLABreached, status)
	}
}

func TestEvaluateSLAStatus_ResolvedTicket_OnTime(t *testing.T) {
	engine := NewSLAEngine()
	now := time.Now()

	onTimeResp := now.Add(-4 * time.Hour)
	onTimeResolved := now.Add(-2 * time.Hour)
	ticket := &model.Ticket{
		Status:                model.StatusResolved,
		CreatedAt:             now.Add(-5 * time.Hour),
		SLAResponseDeadline:   now.Add(-4 * time.Hour),
		SLAResolutionDeadline: now.Add(-1 * time.Hour),
		RespondedAt:           &onTimeResp,
		ResolvedAt:            &onTimeResolved, // Resolved 1h before deadline
	}

	status := engine.EvaluateSLAStatus(ticket)
	if status != model.SLAWithinSLA {
		t.Fatalf("expected SLA status '%s' for on-time resolved ticket, got '%s'", model.SLAWithinSLA, status)
	}
}

func TestEvaluateSLAStatus_Warning_LowRemainingTime(t *testing.T) {
	engine := NewSLAEngine()
	now := time.Now()

	// Total duration: 10 hours. Remaining: 1 hour (10% remaining <= 20% threshold)
	onTimeResp := now.Add(-8 * time.Hour)
	ticket := &model.Ticket{
		Status:                model.StatusInProgress,
		CreatedAt:             now.Add(-9 * time.Hour),
		SLAResponseDeadline:   now.Add(-7 * time.Hour),
		SLAResolutionDeadline: now.Add(1 * time.Hour),
		RespondedAt:           &onTimeResp,
	}

	status := engine.EvaluateSLAStatus(ticket)
	if status != model.SLAWarning {
		t.Fatalf("expected SLA status '%s' for warning threshold, got '%s'", model.SLAWarning, status)
	}
}
