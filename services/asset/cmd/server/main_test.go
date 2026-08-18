package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eomp/services/asset/internal/config"
	"eomp/services/asset/internal/handler"
	"eomp/services/asset/internal/model"
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

	if body["service"] != "asset" {
		t.Errorf("expected service 'asset', got '%v'", body["service"])
	}
}

func TestAssetConstants(t *testing.T) {
	if model.CategoryLaptop != "LAPTOP" {
		t.Errorf("expected LAPTOP, got %s", model.CategoryLaptop)
	}
	if model.StatusInUse != "IN_USE" {
		t.Errorf("expected IN_USE, got %s", model.StatusInUse)
	}
	if model.RelDependsOn != "DEPENDS_ON" {
		t.Errorf("expected DEPENDS_ON, got %s", model.RelDependsOn)
	}
	if model.CIStatusOperational != "OPERATIONAL" {
		t.Errorf("expected OPERATIONAL, got %s", model.CIStatusOperational)
	}
}
