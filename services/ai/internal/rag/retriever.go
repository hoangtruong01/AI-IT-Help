package rag

import (
	"context"

	"eomp/services/ai/internal/model"
)

// Retriever defines operations for retrieving context from vector storage (Qdrant).
type Retriever interface {
	// Search retrieves top-K relevant knowledge base articles for a query vector.
	Search(ctx context.Context, vector []float32, limit int) ([]model.Citation, error)
}

// MockRetriever provides dummy citations when vector database is empty or offline.
type MockRetriever struct{}

// NewMockRetriever creates a new mock retriever.
func NewMockRetriever() *MockRetriever {
	return &MockRetriever{}
}

func (r *MockRetriever) Search(ctx context.Context, vector []float32, limit int) ([]model.Citation, error) {
	return []model.Citation{
		{
			ArticleID: "kb-001",
			Title:     "Getting Started with EOMP Operations",
			Score:     0.92,
		},
	}, nil
}
