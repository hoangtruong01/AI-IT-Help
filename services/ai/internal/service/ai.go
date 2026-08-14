package service

import (
	"context"
	"errors"

	"eomp/services/ai/internal/model"
	"eomp/services/ai/internal/prompt"
	"eomp/services/ai/internal/provider"
	"eomp/services/ai/internal/rag"
)

// AIService defines core AI business logic methods.
type AIService interface {
	Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error)
	AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error)
}

type aiService struct {
	llm       provider.LLMProvider
	embedder  provider.EmbeddingProvider
	retriever rag.Retriever
}

// NewAIService constructs a new AIService instance.
func NewAIService(llm provider.LLMProvider, embedder provider.EmbeddingProvider, retriever rag.Retriever) AIService {
	return &aiService{
		llm:       llm,
		embedder:  embedder,
		retriever: retriever,
	}
}

func (s *aiService) Chat(ctx context.Context, req *model.ChatRequest) (*model.ChatResponse, error) {
	if req == nil || len(req.Messages) == 0 {
		return nil, errors.New("chat request cannot be empty")
	}

	// Prepare messages with system prompt prepend
	messages := append([]model.Message{
		{Role: "system", Content: prompt.SystemPrompt},
	}, req.Messages...)

	// Get response from LLM
	ans, err := s.llm.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Mock citations retrieval
	citations, _ := s.retriever.Search(ctx, make([]float32, 768), 3)

	return &model.ChatResponse{
		Answer:     ans,
		Citations:  citations,
		Confidence: 0.95,
		TokensUsed: 120,
	}, nil
}

func (s *aiService) AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error) {
	if title == "" {
		return nil, errors.New("ticket title cannot be empty")
	}
	return s.llm.AnalyzeTicket(ctx, title, description)
}
