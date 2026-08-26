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

// GeminiProvider implements LLMProvider and EmbeddingProvider for Google Gemini API.
type GeminiProvider struct {
	apiKey         string
	model          string
	embeddingModel string
	httpClient     *http.Client
}

// NewGeminiProvider creates a Google Gemini API client.
func NewGeminiProvider(apiKey, modelName, embeddingModelName string) *GeminiProvider {
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}
	if embeddingModelName == "" {
		embeddingModelName = "text-embedding-004"
	}

	return &GeminiProvider{
		apiKey:         apiKey,
		model:          modelName,
		embeddingModel: embeddingModelName,
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (p *GeminiProvider) Generate(ctx context.Context, messages []model.Message) (string, error) {
	if p.apiKey == "" {
		return "", fmt.Errorf("gemini API key is not configured")
	}

	type geminiPart struct {
		Text string `json:"text"`
	}
	type geminiContent struct {
		Role  string       `json:"role"`
		Parts []geminiPart `json:"parts"`
	}

	var contents []geminiContent
	for _, m := range messages {
		role := "user"
		if m.Role == "assistant" {
			role = "model"
		}
		// Gemini standard format
		contents = append(contents, geminiContent{
			Role: role,
			Parts: []geminiPart{
				{Text: m.Content},
			},
		})
	}

	reqBody := map[string]any{
		"contents": contents,
		"generationConfig": map[string]any{
			"temperature": 0.2,
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("gemini marshal error: %w", err)
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", p.model, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("gemini API returned status %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("gemini decode error: %w", err)
	}

	if len(res.Candidates) == 0 || len(res.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini returned empty response")
	}

	var sb strings.Builder
	for _, part := range res.Candidates[0].Content.Parts {
		sb.WriteString(part.Text)
	}
	return sb.String(), nil
}

func (p *GeminiProvider) AnalyzeTicket(ctx context.Context, title, description string) (*model.TicketAnalysis, error) {
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

func (p *GeminiProvider) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("gemini API key is not configured")
	}

	reqBody := map[string]any{
		"model": fmt.Sprintf("models/%s", p.embeddingModel),
		"content": map[string]any{
			"parts": []map[string]string{
				{"text": text},
			},
		},
	}

	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s", p.embeddingModel, p.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini embedding request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("gemini embedding returned status %d: %s", resp.StatusCode, string(body))
	}

	var res struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("gemini embedding decode error: %w", err)
	}

	return res.Embedding.Values, nil
}

func (p *GeminiProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, text := range texts {
		vec, err := p.EmbedText(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("gemini batch embedding index %d error: %w", i, err)
		}
		result[i] = vec
	}
	return result, nil
}

func (p *GeminiProvider) Dimensions() int {
	return 768
}
