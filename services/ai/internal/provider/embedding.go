package provider

import "context"

// EmbeddingProvider defines the contract for vector embedding generation.
type EmbeddingProvider interface {
	// EmbedText creates a vector embedding for a single text snippet.
	EmbedText(ctx context.Context, text string) ([]float32, error)

	// EmbedBatch creates vector embeddings for multiple text snippets in batch.
	EmbedBatch(ctx context.Context, texts []string) ([][]float32, error)

	// Dimensions returns the dimensionality of the generated vectors.
	Dimensions() int
}
