package config

import (
	"eomp/packages/shared/pkg/config"
)

// Config represents service configuration loaded from environment variables.
type Config struct {
	ServiceName     string
	Port            int
	Environment     string
	Version         string
	DBHost          string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	DBSSLMode       string
	MigrationsPath  string
	QdrantHost      string
	QdrantPort      int
	QdrantCollection string
}

// Load reads configuration from the environment with sensible defaults.
func Load() *Config {
	return &Config{
		ServiceName:      "knowledge",
		Port:             config.GetEnvInt("PORT", 8087),
		Environment:      config.GetEnv("APP_ENV", "development"),
		Version:          "0.1.0",
		DBHost:           config.GetEnv("DB_HOST", "localhost"),
		DBPort:           config.GetEnvInt("DB_PORT", 5432),
		DBUser:           config.GetEnv("DB_USER", "eomp"),
		DBPassword:       config.GetEnv("DB_PASSWORD", "eomp_secret"),
		DBName:           config.GetEnv("DB_NAME", "knowledge_db"),
		DBSSLMode:        config.GetEnv("DB_SSLMODE", "disable"),
		MigrationsPath:   config.GetEnv("MIGRATIONS_PATH", "migrations"),
		QdrantHost:       config.GetEnv("QDRANT_HOST", "localhost"),
		QdrantPort:       config.GetEnvInt("QDRANT_PORT", 6333),
		QdrantCollection: config.GetEnv("QDRANT_COLLECTION", "knowledge_base"),
	}
}
