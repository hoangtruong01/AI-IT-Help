package service

import (
	"context"
	"errors"
	"strings"

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

	// Extract the latest user query
	lastUserMsg := ""
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			lastUserMsg = req.Messages[i].Content
			break
		}
	}

	// Generate embedding vector for the query (if embedder available)
	var queryVector []float32
	if s.embedder != nil && lastUserMsg != "" {
		vec, err := s.embedder.EmbedText(ctx, lastUserMsg)
		if err == nil {
			queryVector = vec
		}
	}

	// Retrieve relevant citations via RAG Retriever (Qdrant with Fallback)
	citations, isFallback, _ := s.retriever.Search(ctx, lastUserMsg, queryVector, 3)

	// Prepare messages with system prompt prepend
	messages := append([]model.Message{
		{Role: "system", Content: prompt.SystemPrompt},
	}, req.Messages...)

	// Get response from LLM
	ans, err := s.llm.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	// Compute overall confidence score
	confidence := 0.95
	if len(citations) > 0 {
		var sum float64
		for _, c := range citations {
			sum += c.Score
		}
		confidence = sum / float64(len(citations))
		if confidence > 0.98 {
			confidence = 0.98
		}
	}

	// Estimate tokens used (~4 chars per token)
	tokensUsed := len(ans)/4 + 40

	return &model.ChatResponse{
		Answer:       ans,
		Citations:    citations,
		Confidence:   confidence,
		TokensUsed:   tokensUsed,
		FallbackMode: isFallback,
	}, nil
}

func (s *aiService) AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, errors.New("ticket title cannot be empty")
	}
	return s.llm.AnalyzeTicket(ctx, title, description)
}
