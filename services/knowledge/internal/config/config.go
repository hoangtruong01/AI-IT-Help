package config

import (
	"errors"

	"eomp/packages/shared/pkg/config"
)

// Config represents service configuration loaded from environment variables.
type Config struct {
	ServiceName      string
	Port             int
	Environment      string
	Version          string
	DBHost           string
	DBPort           int
	DBUser           string
	DBPassword       string
	DBName           string
	DBSSLMode        string
	MigrationsPath   string
	QdrantHost       string
	QdrantPort       int
	QdrantCollection string
}

// Load reads configuration from the environment with sensible defaults.
func Load() *Config {
	return &Config{
		ServiceName:      "knowledge",
		Port:             config.GetEnvInt("PORT", 8087),
		Environment:      config.GetEnv("APP_ENV", "development"),
		Version:          "0.1.0",
		DBHost:           config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:           config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:           config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:       config.GetEnv("POSTGRES_PASSWORD", "eomp_dev_password"),
		DBName:           config.GetEnv("KNOWLEDGE_DB_NAME", "knowledge_db"),
		DBSSLMode:        config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath:   config.GetEnv("KNOWLEDGE_MIGRATIONS_PATH", "migrations"),
		QdrantHost:       config.GetEnv("QDRANT_HOST", "localhost"),
		QdrantPort:       config.GetEnvInt("QDRANT_PORT", 6333),
		QdrantCollection: config.GetEnv("QDRANT_COLLECTION", "knowledge_base"),
	}
}

// Validate performs fail-fast configuration checks.
func (c *Config) Validate() error {
	if c.Environment == "production" {
		if c.DBPassword == "eomp_dev_password" || c.DBPassword == "" {
			return errors.New("security violation: default dev DB_PASSWORD is prohibited in production")
		}
	}
	return nil
}
