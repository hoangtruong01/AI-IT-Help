package provider

import (
	"context"
	"fmt"
	"time"

	"eomp/services/ai/internal/model"
)

// MockProvider implements both LLMProvider and EmbeddingProvider for testing and offline development.
type MockProvider struct{}

// NewMockProvider creates a ready-to-use mock provider.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

func (m *MockProvider) Generate(ctx context.Context, messages []model.Message) (string, error) {
	if len(messages) == 0 {
		return "Hello! How can I assist you with your enterprise operations today?", nil
	}
	lastMsg := messages[len(messages)-1].Content
	return fmt.Sprintf("I received your query: '%s'. AI Service is currently running in development mode.", lastMsg), nil
}

func (m *MockProvider) AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error) {
	return &model.TicketAnalysis{
		TicketID:            "mock-ticket-001",
		SuggestedCategory:   "IT Support",
		Priority:            "medium",
		Summary:             fmt.Sprintf("Issue regarding: %s", title),
		SuggestedResolution: "Check hardware/software configuration and verify user permissions.",
		RequiresHumanReview: true, // Safety rule: AI cannot execute authoritative changes
		CreatedAt:           time.Now(),
	}, nil
}

func (m *MockProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
	vec := make([]float32, 768)
	return vec, nil
}

func (m *MockProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i := range texts {
		result[i] = make([]float32, 768)
	}
	return result, nil
}

func (m *MockProvider) Dimensions() int {
	return 768
}
