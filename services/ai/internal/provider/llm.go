package provider

import (
	"context"

	"eomp/services/ai/internal/model"
)

// LLMProvider defines the contract for large language model operations.
type LLMProvider interface {
	// Generate generates a text response for the given conversation messages.
	Generate(ctx context.Context, messages []model.Message) (string, error)

	// AnalyzeTicket evaluates a helpdesk ticket and produces recommendations.
	AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error)
}
