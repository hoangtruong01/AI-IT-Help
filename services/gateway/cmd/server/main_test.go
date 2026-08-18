package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eomp/services/gateway/internal/config"
	"eomp/services/gateway/internal/handler"
)

func TestGatewayHealthHandler(t *testing.T) {
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

	if body["service"] != "gateway" {
		t.Errorf("expected service 'gateway', got '%v'", body["service"])
	}
	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%v'", body["status"])
	}
}
