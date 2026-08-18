package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eomp/services/workflow/internal/config"
	"eomp/services/workflow/internal/handler"
	"eomp/services/workflow/internal/model"
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

	if body["service"] != "workflow" {
		t.Errorf("expected service 'workflow', got '%v'", body["service"])
	}
}

func TestWorkflowConstants(t *testing.T) {
	if model.InstanceStatusWaitingApproval != "WAITING_APPROVAL" {
		t.Errorf("expected WAITING_APPROVAL, got %s", model.InstanceStatusWaitingApproval)
	}
	if model.ApprovalStatusApproved != "APPROVED" {
		t.Errorf("expected APPROVED, got %s", model.ApprovalStatusApproved)
	}
	if model.ApprovalStatusRejected != "REJECTED" {
		t.Errorf("expected REJECTED, got %s", model.ApprovalStatusRejected)
	}
}
