package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// Test Case 6.1 (Positive): Gửi câu hỏi "How to reset user MFA token?" -> Phản hồi đúng quy trình trong tài liệu nội bộ, trích dẫn đúng bài viết Knowledge Base.
func TestTestCase6_1_Positive_MFAReset(t *testing.T) {
	mockProv := provider.NewMockProvider()
	retriever := rag.NewSmartRetriever("127.0.0.1", 6333, "knowledge_base")
	aiSvc := service.NewAIService(mockProv, mockProv, retriever)

	req := &model.ChatRequest{
		Messages: []model.Message{
			{Role: "user", Content: "How to reset user MFA token?"},
		},
	}

	res, err := aiSvc.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify internal SOP procedure in answer
	if !strings.Contains(res.Answer, "Okta") || !strings.Contains(res.Answer, "Identity Verification") {
		t.Errorf("answer should contain internal Okta / Identity Verification procedure, got: %s", res.Answer)
	}

	// Verify citations grounding
	if len(res.Citations) == 0 {
		t.Fatal("expected citations for MFA question")
	}

	hasMFACitation := false
	for _, c := range res.Citations {
		if strings.Contains(strings.ToLower(c.Title), "mfa") {
			hasMFACitation = true
			if c.Score < 0.90 {
				t.Errorf("expected high similarity score >= 0.90, got %f", c.Score)
			}
			break
		}
	}
	if !hasMFACitation {
		t.Errorf("expected citation for MFA article, got citations: %+v", res.Citations)
	}
}

// Test Case 6.2 (Negative / Fallback): Khi Qdrant Vector Store mất kết nối -> Hệ thống tự động fallback sang In-Memory Knowledge Search mà không crash Gateway.
func TestTestCase6_2_Negative_QdrantFallback(t *testing.T) {
	mockProv := provider.NewMockProvider()
	// Point to an unreachable port to simulate Qdrant being offline
	offlineRetriever := rag.NewSmartRetriever("127.0.0.1", 59999, "knowledge_base")
	aiSvc := service.NewAIService(mockProv, mockProv, offlineRetriever)

	req := &model.ChatRequest{
		Messages: []model.Message{
			{Role: "user", Content: "Cannot connect to VPN Staging Server"},
		},
	}

	// Must not crash or return error
	res, err := aiSvc.Chat(context.Background(), req)
	if err != nil {
		t.Fatalf("service should not fail when Qdrant is offline, got error: %v", err)
	}

	if !res.FallbackMode {
		t.Errorf("expected FallbackMode to be true when Qdrant is unreachable")
	}

	if len(res.Citations) == 0 {
		t.Fatal("expected fallback citations to be returned")
	}

	hasVPNCitation := false
	for _, c := range res.Citations {
		if strings.Contains(strings.ToLower(c.Title), "vpn") {
			hasVPNCitation = true
			break
		}
	}
	if !hasVPNCitation {
		t.Errorf("expected fallback VPN citation, got: %+v", res.Citations)
	}
}

// Test Case 6.3 (Performance): Thời gian nhận Token đầu tiên (Time To First Token - TTFT) phải đạt < 800ms.
func TestTestCase6_3_Performance_TTFT(t *testing.T) {
	mockProv := provider.NewMockProvider()
	retriever := rag.NewSmartRetriever("127.0.0.1", 6333, "knowledge_base")
	aiSvc := service.NewAIService(mockProv, mockProv, retriever)
	aiHandler := handler.NewAIHandler(aiSvc)

	bodyBytes, _ := json.Marshal(model.ChatRequest{
		Messages: []model.Message{
			{Role: "user", Content: "How do I configure PostgreSQL connection pool?"},
		},
	})

	start := time.Now()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/ai/chat", bytes.NewReader(bodyBytes))
	rec := httptest.NewRecorder()

	aiHandler.Chat(rec, req)
	elapsed := time.Since(start)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	if elapsed > 800*time.Millisecond {
		t.Errorf("TTFT / response time exceeded 800ms limit: took %v", elapsed)
	}
}

// User Story Scenario: AI automatically analyzes ticket and suggests action
func TestAnalyzeTicket_VPNScenario(t *testing.T) {
	mockProv := provider.NewMockProvider()
	retriever := rag.NewSmartRetriever("127.0.0.1", 6333, "knowledge_base")
	aiSvc := service.NewAIService(mockProv, mockProv, retriever)

	analysis, err := aiSvc.AnalyzeTicket(
		context.Background(),
		"Cannot connect to VPN Staging Server",
		"Remote engineer is unable to reach internal staging services via WireGuard tunnel.",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if analysis.SuggestedCategory != "Network & Access" {
		t.Errorf("expected category 'Network & Access', got '%s'", analysis.SuggestedCategory)
	}
	if analysis.Priority != "HIGH" {
		t.Errorf("expected priority 'HIGH', got '%s'", analysis.Priority)
	}
	if analysis.Confidence < 0.90 {
		t.Errorf("expected confidence >= 0.90, got %f", analysis.Confidence)
	}
	if !analysis.RequiresHumanReview {
		t.Errorf("expected RequiresHumanReview to be true (Safety Rule)")
	}
}
