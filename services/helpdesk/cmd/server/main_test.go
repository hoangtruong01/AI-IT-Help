package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eomp/services/helpdesk/internal/config"
	"eomp/services/helpdesk/internal/handler"
	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/service"
)

func TestHealthHandler(t *testing.T) {
	cfg := config.Load()
	h := handler.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["service"] != "helpdesk" {
		t.Errorf("expected service 'helpdesk', got '%v'", body["service"])
	}
}

func TestSLAEngineDeadlines(t *testing.T) {
	engine := service.NewSLAEngine()

	// 1. Urgent Priority: 15m response, 2h resolution
	now := time.Now()
	respDeadline, resolDeadline := engine.CalculateDeadlines(model.PriorityUrgent, 0, 0)
	if respDeadline.Sub(now) < 14*time.Minute || respDeadline.Sub(now) > 16*time.Minute {
		t.Errorf("expected ~15m response deadline for URGENT, got %v", respDeadline.Sub(now))
	}
	if resolDeadline.Sub(now) < 118*time.Minute || resolDeadline.Sub(now) > 122*time.Minute {
		t.Errorf("expected ~2h resolution deadline for URGENT, got %v", resolDeadline.Sub(now))
	}

	// 2. High Priority: 30m response, 4h resolution
	respDeadline, resolDeadline = engine.CalculateDeadlines(model.PriorityHigh, 0, 0)
	if respDeadline.Sub(now) < 29*time.Minute || respDeadline.Sub(now) > 31*time.Minute {
		t.Errorf("expected ~30m response deadline for HIGH, got %v", respDeadline.Sub(now))
	}
	if resolDeadline.Sub(now) < 238*time.Minute || resolDeadline.Sub(now) > 242*time.Minute {
		t.Errorf("expected ~4h resolution deadline for HIGH, got %v", resolDeadline.Sub(now))
	}

	// 3. Custom SLA
	respDeadline, resolDeadline = engine.CalculateDeadlines("CUSTOM", 60, 180)
	if respDeadline.Sub(now) < 59*time.Minute || respDeadline.Sub(now) > 61*time.Minute {
		t.Errorf("expected ~60m response deadline for custom, got %v", respDeadline.Sub(now))
	}
	if resolDeadline.Sub(now) < 178*time.Minute || resolDeadline.Sub(now) > 182*time.Minute {
		t.Errorf("expected ~180m resolution deadline for custom, got %v", resolDeadline.Sub(now))
	}
}

func TestSLAEngineStatusEvaluation(t *testing.T) {
	engine := service.NewSLAEngine()
	now := time.Now()

	// 1. Within SLA
	ticket1 := &model.Ticket{
		Status:                model.StatusOpen,
		CreatedAt:             now.Add(-10 * time.Minute),
		SLAResolutionDeadline: now.Add(2 * time.Hour),
	}
	if status := engine.EvaluateSLAStatus(ticket1); status != model.SLAWithinSLA {
		t.Errorf("expected WITHIN_SLA, got %s", status)
	}

	// 2. Breached SLA
	ticket2 := &model.Ticket{
		Status:                model.StatusOpen,
		CreatedAt:             now.Add(-3 * time.Hour),
		SLAResolutionDeadline: now.Add(-10 * time.Minute),
	}
	if status := engine.EvaluateSLAStatus(ticket2); status != model.SLABreached {
		t.Errorf("expected BREACHED, got %s", status)
	}

	// 3. Resolved on time
	resolvedTime := now.Add(-30 * time.Minute)
	ticket3 := &model.Ticket{
		Status:                model.StatusResolved,
		CreatedAt:             now.Add(-2 * time.Hour),
		ResolvedAt:            &resolvedTime,
		SLAResolutionDeadline: now.Add(1 * time.Hour),
	}
	if status := engine.EvaluateSLAStatus(ticket3); status != model.SLAWithinSLA {
		t.Errorf("expected WITHIN_SLA for on-time resolved ticket, got %s", status)
	}
}
