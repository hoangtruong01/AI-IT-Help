package config

import (
	"errors"

	"eomp/packages/shared/pkg/config"
)

// Config holds AI service and provider configuration.
type Config struct {
	ServiceName      string
	Port             int
	Environment      string
	Version          string
	AIProvider       string // "mock", "ollama", "openai", "gemini"
	AIAPIKey         string
	AIModel          string
	EmbeddingModel   string
	OllamaBaseURL    string
	OpenAIAPIKey     string
	GeminiAPIKey     string
	QdrantURL        string
	QdrantHost       string
	QdrantPort       int
	QdrantCollection string
	AutoIngest       bool
}

// Load loads AI service configuration from environment variables.
func Load() *Config {
	aiProvider := config.GetEnv("AI_PROVIDER", "mock")
	apiKey := config.GetEnv("AI_API_KEY", "")
	openAIKey := config.GetEnv("OPENAI_API_KEY", apiKey)
	geminiKey := config.GetEnv("GEMINI_API_KEY", apiKey)

	return &Config{
		ServiceName:      "ai",
		Port:             config.GetEnvInt("PORT", 8088),
		Environment:      config.GetEnv("APP_ENV", "development"),
		Version:          "2.0.0",
		AIProvider:       aiProvider,
		AIAPIKey:         apiKey,
		AIModel:          config.GetEnv("AI_MODEL", "llama3.2"),
		EmbeddingModel:   config.GetEnv("EMBEDDING_MODEL", "nomic-embed-text"),
		OllamaBaseURL:    config.GetEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		OpenAIAPIKey:     openAIKey,
		GeminiAPIKey:     geminiKey,
		QdrantURL:        config.GetEnv("QDRANT_URL", "http://localhost:6333"),
		QdrantHost:       config.GetEnv("QDRANT_HOST", "localhost"),
		QdrantPort:       config.GetEnvInt("QDRANT_PORT", 6333),
		QdrantCollection: config.GetEnv("QDRANT_COLLECTION", "knowledge_base"),
		AutoIngest:       config.GetEnvBool("AUTO_INGEST", false),
	}
}

// Validate performs fail-fast configuration checks.
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.AIProvider == "openai" && c.OpenAIAPIKey == "" {
			return errors.New("security violation: OPENAI_API_KEY must be provided in production when using openai provider")
		}
		if c.AIProvider == "gemini" && c.GeminiAPIKey == "" {
			return errors.New("security violation: GEMINI_API_KEY must be provided in production when using gemini provider")
		}
	}
	return nil
}
