package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eomp/services/audit/internal/config"
	"eomp/services/audit/internal/handler"
)

func TestHealthCheck(t *testing.T) {
	cfg := config.Load()
	h := handler.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res handler.HealthResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", res.Status)
	}
	if res.Service != "audit" {
		t.Errorf("expected service 'audit', got '%s'", res.Service)
	}
	if res.Version != cfg.Version {
		t.Errorf("expected version '%s', got '%s'", cfg.Version, res.Version)
	}
}
