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

// OpenAIProvider implements LLMProvider and EmbeddingProvider for OpenAI API.
type OpenAIProvider struct {
	apiKey         string
	baseURL        string
	model          string
	embeddingModel string
	httpClient     *http.Client
}

// NewOpenAIProvider creates an OpenAI API client.
func NewOpenAIProvider(apiKey, modelName, embeddingModelName string) *OpenAIProvider {
	if modelName == "" {
		modelName = "gpt-4o-mini"
	}
	if embeddingModelName == "" {
		embeddingModelName = "text-embedding-3-small"
	}

	return &OpenAIProvider{
		apiKey:         apiKey,
		baseURL:        "https://api.openai.com/v1",
		model:          modelName,
		embeddingModel: embeddingModelName,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (p *OpenAIProvider) Generate(ctx context.Context, messages []model.Message) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("openai API key is not configured")
	}

	type openAIMsg struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}

	openAIMessages := make([]openAIMsg, len(messages))
	for i, m := range messages {
		openAIMessages[i] = openAIMsg{
			Role:    m.Role,
			Content: m.Content,
		}
	}

	reqBody := map[string]any{
		"model":       p.model,
		"messages":    openAIMessages,
		"temperature": 0.2,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("openai marshal error: %w", err)
	}

	endpoint := fmt.Sprintf("%s/chat/completions", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai API returned status %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Choices []struct {
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("openai decode error: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("openai returned empty choices")
	}

	return res.Choices[0].Message.Content, nil
}

func (p *OpenAIProvider) AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error) {
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

func (p *OpenAIProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
	res, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(res) == 0 {
		return nil, fmt.Errorf("openai embedding returned empty response")
	}
	return res[0], nil
}

func (p *OpenAIProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("openai API key is not configured")
	}

	reqBody := map[string]any{
		"model": p.embeddingModel,
		"input": texts,
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("openai embedding marshal error: %w", err)
	}

	endpoint := fmt.Sprintf("%s/embeddings", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("openai embedding API returned status %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("openai embedding decode error: %w", err)
	}

	result := make([][]float32, len(texts))
	for _, item := range res.Data {
		if item.Index < len(result) {
			result[item.Index] = item.Embedding
		}
	}

	return result, nil
}

func (p *OpenAIProvider) Dimensions() int {
	if strings.Contains(p.embeddingModel, "large") {
		return 3072
	}
	return 1536
}
