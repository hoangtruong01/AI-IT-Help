package config

import (
	"eomp/packages/shared/pkg/config"
)

// Config holds AI service and provider configuration.
type Config struct {
	ServiceName    string
	Port           int
	Environment    string
	Version        string
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
		AIAPIKey:       config.GetEnv("AI_API_KEY", ""),
		AIModel:        config.GetEnv("AI_MODEL", "gemini-2.0-flash"),
		EmbeddingModel: config.GetEnv("EMBEDDING_MODEL", "text-embedding-004"),
		QdrantURL:      config.GetEnv("QDRANT_URL", "http://localhost:6333"),
	}
}
