package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"eomp/services/ai/internal/model"
	"eomp/services/ai/internal/prompt"
)

// OllamaProvider implements LLMProvider and EmbeddingProvider for local Ollama instances.
type OllamaProvider struct {
	baseURL        string
	model          string
	embeddingModel string
	httpClient     *http.Client
}

// NewOllamaProvider creates a new Ollama client for local LLM inference.
func NewOllamaProvider(baseURL, modelName, embeddingModelName string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	if modelName == "" {
		modelName = "llama3.2"
	}
	if embeddingModelName == "" {
		embeddingModelName = "nomic-embed-text"
	}

	return &OllamaProvider{
		baseURL:        baseURL,
		model:          modelName,
		embeddingModel: embeddingModelName,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *OllamaProvider) Generate(ctx context.Context, messages []model.Message) (string, error) {
	type ollamaMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	ollamaMessages := make([]ollamaMsg, len(messages))
	for i, m := range messages {
		ollamaMessages[i] = ollamaMsg{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	reqBody := map[string]any{
		"model":    p.model,
		"messages": ollamaMessages,
		"stream":   false,
		"options": map[string]any{
			"temperature": 0.2,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("ollama marshal error: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/chat", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("ollama request creation error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama connection error (is Ollama running on %s?): %w", p.baseURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama error status %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("ollama decode error: %w", err)
	}

	return res.Message.Content, nil
}

func (p *OllamaProvider) AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error) {
	triagePrompt := prompt.HelpdeskTriagePrompt(title, description)
	messages := []model.Message{
		{Role: "system", Content: prompt.SystemPrompt},
		{Role: "user", Content: triagePrompt},
	}

	raw, err := p.Generate(ctx, messages)
	if err != nil {
		return nil, err
	}

	return prompt.ParseTriageJSON(raw, title)
}

func (p *OllamaProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
	reqBody := map[string]any{
		"model":  p.embeddingModel,
		"prompt": text,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("ollama embedding marshal error: %w", err)
	}

	endpoint := fmt.Sprintf("%s/api/embeddings", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embedding connection error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embedding error status %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Embedding []float32 `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("ollama embedding decode error: %w", err)
	}

	return res.Embedding, nil
}

func (p *OllamaProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := p.EmbedText(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("batch embedding index %d error: %w", i, err)
		}
		result[i] = vec
	}
	return result, nil
}

func (p *OllamaProvider) Dimensions() int {
	return 768
}
