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
	llm          provider.LLMProvider
	embedder     provider.EmbeddingProvider
	retriever    rag.Retriever
	mockFallback *provider.MockProvider
}

// NewAIService constructs a new AIService instance.
func NewAIService(llm provider.LLMProvider, embedder provider.EmbeddingProvider, retriever rag.Retriever) AIService {
	return &aiService{
		llm:          llm,
		embedder:     embedder,
		retriever:    retriever,
		mockFallback: provider.NewMockProvider(),
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

	// Prepare messages with system prompt prepend and RAG context augmentation
	var augmentedMessages []model.Message
	augmentedMessages = append(augmentedMessages, model.Message{
		Role:    "system",
		Content: prompt.SystemPrompt,
	})

	for i, m := range req.Messages {
		if i == len(req.Messages)-1 && m.Role == "user" && len(citations) > 0 {
			// Augment the last user query with RAG context
			augmentedMessages = append(augmentedMessages, model.Message{
				Role:    "user",
				Content: prompt.FormatRAGPrompt(m.Content, citations),
			})
		} else {
			augmentedMessages = append(augmentedMessages, m)
		}
	}

	// Get response from primary LLM, fallback to mock if error occurs
	ans, err := s.llm.Generate(ctx, augmentedMessages)
	if err != nil {
		// Graceful fallback to MockProvider
		ans, _ = s.mockFallback.Generate(ctx, req.Messages)
		isFallback = true
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

	combined := title + " " + description

	// 1. Vector Search for relevant SOP Runbooks
	var queryVec []float32
	if s.embedder != nil {
		vec, err := s.embedder.EmbedText(ctx, combined)
		if err == nil {
			queryVec = vec
		}
	}
	citations, _, _ := s.retriever.Search(ctx, combined, queryVec, 3)

	// 2. Perform AI Triage via LLM
	analysis, err := s.llm.AnalyzeTicket(ctx, title, description)
	if err != nil {
		// Fallback to domain mock analysis if LLM fails
		analysis, err = s.mockFallback.AnalyzeTicket(ctx, title, description)
		if err != nil {
			return nil, err
		}
	}

	// 3. Attach citations if not present
	if len(analysis.Citations) == 0 && len(citations) > 0 {
		analysis.Citations = citations
	}

	return analysis, nil
}
