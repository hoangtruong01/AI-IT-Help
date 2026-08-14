package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"eomp/services/ai/internal/config"
	"eomp/services/ai/internal/handler"
	"eomp/services/ai/internal/model"
	"eomp/services/ai/internal/provider"
	"eomp/services/ai/internal/rag"
	"eomp/services/ai/internal/service"
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
	if res.Service != "ai" {
		t.Errorf("expected service 'ai', got '%s'", res.Service)
	}
}

func TestAIChat(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	mockRetriever := rag.NewMockRetriever()
	aiService := service.NewAIService(mockProvider, mockProvider, mockRetriever)
	aiHandler := handler.NewAIHandler(aiService)

	body, _ := json.Marshal(model.ChatRequest{
		Messages: []model.Message{
			{Role: "user", Content: "How do I reset my password?"},
		},
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ai/chat", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	aiHandler.Chat(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res model.ChatResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.Answer == "" {
		t.Error("expected non-empty AI answer")
	}
}

func TestAITicketAnalysis(t *testing.T) {
	mockProvider := provider.NewMockProvider()
	mockRetriever := rag.NewMockRetriever()
	aiService := service.NewAIService(mockProvider, mockProvider, mockRetriever)
	aiHandler := handler.NewAIHandler(aiService)

	body, _ := json.Marshal(handler.AnalyzeTicketRequest{
		Title:       "Laptop screen flickers",
		Description: "External monitor works fine, built-in display keeps flashing black.",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ai/analyze-ticket", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	aiHandler.AnalyzeTicket(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res model.TicketAnalysis
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !res.RequiresHumanReview {
		t.Error("expected RequiresHumanReview to be true (Section 35 safety rule)")
	}
}
