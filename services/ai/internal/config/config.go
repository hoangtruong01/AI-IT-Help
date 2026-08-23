package config

import (
	"errors"

	"eomp/packages/shared/pkg/config"
)

// Config holds AI service and provider configuration.
type Config struct {
	ServiceName    string
	Port           int
	Environment    string
	Version        string
	AIProvider     string
	AIAPIKey       string
	AIModel        string
	EmbeddingModel string
	QdrantURL      string
}

// Load loads AI service configuration from environment variables.
func Load() *Config {
	return &Config{
		ServiceName:    "ai",
		Port:           config.GetEnvInt("PORT", 8088),
		Environment:    config.GetEnv("APP_ENV", "development"),
		Version:        "0.1.0",
		AIProvider:     config.GetEnv("AI_PROVIDER", "mock"),
		AIAPIKey:       config.GetEnv("AI_API_KEY", ""),
		AIModel:        config.GetEnv("AI_MODEL", "gemini-2.0-flash"),
		EmbeddingModel: config.GetEnv("EMBEDDING_MODEL", "text-embedding-004"),
		QdrantURL:      config.GetEnv("QDRANT_URL", "http://localhost:6333"),
	}
}

// Validate performs fail-fast configuration checks.
func (c *Config) Validate() error {
	if c.Environment == "production" && c.AIProvider == "openai" && c.AIAPIKey == "" {
		return errors.New("security violation: AI_API_KEY must be provided in production when using openai provider")
	}
	return nil
}
