package rag

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"eomp/services/ai/internal/provider"
	_ "github.com/lib/pq"
)

// QdrantPoint represents a point inserted into Qdrant collection.
type QdrantPoint struct {
	ID      string         `json:"id"`
	Vector  []float32      `json:"vector"`
	Payload map[string]any `json:"payload"`
}

// IngestionPipeline orchestrates chunking, embedding, and storing articles into Qdrant.
type IngestionPipeline struct {
	qdrantURL      string
	collectionName string
	embedder       provider.EmbeddingProvider
	httpClient     *http.Client
}

// NewIngestionPipeline creates an ingestion pipeline instance.
func NewIngestionPipeline(qdrantURL, collectionName string, embedder provider.EmbeddingProvider) *IngestionPipeline {
	if qdrantURL == "" {
		qdrantURL = "http://localhost:6333"
	}
	qdrantURL = strings.TrimRight(qdrantURL, "/")
	if collectionName == "" {
		collectionName = "knowledge_base"
	}

	return &IngestionPipeline{
		qdrantURL:      qdrantURL,
		collectionName: collectionName,
		embedder:       embedder,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EnsureCollection verifies collection exists or creates it with correct vector dimension.
func (p *IngestionPipeline) EnsureCollection(ctx context.Context, dim int) error {
	if dim <= 0 {
		dim = p.embedder.Dimensions()
		if dim <= 0 {
			dim = 768
		}
	}

	endpoint := fmt.Sprintf("%s/collections/%s", p.qdrantURL, p.collectionName)

	// Check if collection exists
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	resp, err := p.httpClient.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		resp.Body.Close()
		return nil // Collection already exists
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Create collection
	createPayload := map[string]any{
		"vectors": map[string]any{
			"size":     dim,
			"distance": "Cosine",
		},
	}
	data, _ := json.Marshal(createPayload)

	createReq, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	createReq.Header.Set("Content-Type", "application/json")

	createResp, err := p.httpClient.Do(createReq)
	if err != nil {
		return fmt.Errorf("failed to create Qdrant collection %s: %w", p.collectionName, err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusOK && createResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createResp.Body)
		return fmt.Errorf("qdrant create collection failed status %d: %s", createResp.StatusCode, string(body))
	}

	return nil
}

// UpsertPoints inserts a batch of vector points into Qdrant.
func (p *IngestionPipeline) UpsertPoints(ctx context.Context, points []QdrantPoint) error {
	if len(points) == 0 {
		return nil
	}

	endpoint := fmt.Sprintf("%s/collections/%s/points?wait=true", p.qdrantURL, p.collectionName)

	payload := map[string]any{
		"points": points,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("qdrant upsert error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("qdrant upsert status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// IngestArticleChunks extracts, chunks, embeds, and uploads an article to Qdrant.
func (p *IngestionPipeline) IngestArticle(ctx context.Context, id, title, category, summary, content string) (int, error) {
	chunks := ChunkMarkdownText(content, 500, 50)
	if len(chunks) == 0 {
		chunks = []string{summary}
	}

	embeddings, err := p.embedder.EmbedBatch(ctx, chunks)
	if err != nil {
		return 0, fmt.Errorf("failed to embed article chunks: %w", err)
	}

	var points []QdrantPoint
	for i, chunk := range chunks {
		pointID := fmt.Sprintf("%s-%d", id, i)
		points = append(points, QdrantPoint{
			ID:     pointID,
			Vector: embeddings[i],
			Payload: map[string]any{
				"article_id":  id,
				"title":       title,
				"category":    category,
				"type":        "article",
				"chunk_index": i,
				"content":     chunk,
			},
		})
	}

	if err := p.UpsertPoints(ctx, points); err != nil {
		return 0, err
	}

	return len(points), nil
}

// IngestRunbook extracts, embeds, and uploads an SOP Runbook to Qdrant.
func (p *IngestionPipeline) IngestRunbook(ctx context.Context, id, code, title, category, description, prereqs, stepsJSON, rollback string) (int, error) {
	text := fmt.Sprintf("Runbook %s: %s\nCategory: %s\nDescription: %s\nPrerequisites: %s\nSteps: %s\nRollback: %s",
		code, title, category, description, prereqs, stepsJSON, rollback)

	chunks := ChunkMarkdownText(text, 500, 50)
	if len(chunks) == 0 {
		chunks = []string{description}
	}

	embeddings, err := p.embedder.EmbedBatch(ctx, chunks)
	if err != nil {
		return 0, fmt.Errorf("failed to embed runbook chunks: %w", err)
	}

	var points []QdrantPoint
	for i, chunk := range chunks {
		pointID := fmt.Sprintf("%s-%d", id, i)
		points = append(points, QdrantPoint{
			ID:     pointID,
			Vector: embeddings[i],
			Payload: map[string]any{
				"runbook_id":  id,
				"code":        code,
				"title":       fmt.Sprintf("%s: %s", code, title),
				"category":    category,
				"type":        "runbook",
				"chunk_index": i,
				"content":     chunk,
			},
		})
	}

	if err := p.UpsertPoints(ctx, points); err != nil {
		return 0, err
	}

	return len(points), nil
}

// IngestFromKnowledgeDB connects to knowledge_db and synchronizes all articles and runbooks into Qdrant.
func (p *IngestionPipeline) IngestFromKnowledgeDB(ctx context.Context, db *sql.DB) (int, int, error) {
	if err := p.EnsureCollection(ctx, p.embedder.Dimensions()); err != nil {
		return 0, 0, fmt.Errorf("failed ensuring collection: %w", err)
	}

	// 1. Ingest Knowledge Articles
	articleRows, err := db.QueryContext(ctx, `
		SELECT a.id, a.title, c.name, a.summary, a.content
		FROM knowledge_articles a
		JOIN knowledge_categories c ON a.category_id = c.id
		WHERE a.is_published = TRUE
	`)
	if err != nil {
		return 0, 0, fmt.Errorf("failed querying articles: %w", err)
	}
	defer articleRows.Close()

	totalArticles := 0
	totalVectors := 0

	for articleRows.Next() {
		var id, title, cat, summary, content string
		if err := articleRows.Scan(&id, &title, &cat, &summary, &content); err != nil {
			return totalArticles, totalVectors, err
		}
		pts, err := p.IngestArticle(ctx, id, title, cat, summary, content)
		if err != nil {
			return totalArticles, totalVectors, err
		}
		totalArticles++
		totalVectors += pts
	}

	// 2. Ingest Runbooks
	runbookRows, err := db.QueryContext(ctx, `
		SELECT id, code, title, category, description, prerequisites, steps_json::text, rollback_steps
		FROM runbooks
		WHERE is_active = TRUE
	`)
	if err != nil {
		return totalArticles, totalVectors, fmt.Errorf("failed querying runbooks: %w", err)
	}
	defer runbookRows.Close()

	totalRunbooks := 0
	for runbookRows.Next() {
		var id, code, title, cat, desc, prereqs, steps, rollback string
		if err := runbookRows.Scan(&id, &code, &title, &cat, &desc, &prereqs, &steps, &rollback); err != nil {
			return totalArticles, totalVectors, err
		}
		pts, err := p.IngestRunbook(ctx, id, code, title, cat, desc, prereqs, steps, rollback)
		if err != nil {
			return totalArticles, totalVectors, err
		}
		totalRunbooks++
		totalVectors += pts
	}

	return totalArticles + totalRunbooks, totalVectors, nil
}

// ChunkMarkdownText splits markdown text into logical chunks of roughly targetTokens (estimated as words*1.3).
func ChunkMarkdownText(text string, targetTokens, overlapTokens int) []string {
	clean := strings.TrimSpace(text)
	if clean == "" {
		return nil
	}

	// Split by markdown sections first (##, ###)
	sectionRegex := regexp.MustCompile(`(?m)^#{1,3}\s+`)
	sections := sectionRegex.Split(clean, -1)

	var chunks []string
	for _, sec := range sections {
		sec = strings.TrimSpace(sec)
		if sec == "" {
			continue
		}

		words := strings.Fields(sec)
		maxWords := targetTokens
		if maxWords <= 0 {
			maxWords = 200
		}

		if len(words) <= maxWords {
			chunks = append(chunks, sec)
			continue
		}

		// Split large sections into sub-chunks with overlap
		step := maxWords - overlapTokens
		if step <= 0 {
			step = maxWords / 2
		}
		for i := 0; i < len(words); i += step {
			end := i + maxWords
			if end > len(words) {
				end = len(words)
			}
			chunk := strings.Join(words[i:end], " ")
			chunks = append(chunks, chunk)
			if end >= len(words) {
				break
			}
		}
	}

	if len(chunks) == 0 {
		chunks = append(chunks, clean)
	}

	return chunks
}
