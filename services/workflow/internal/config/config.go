package config

import (
	"eomp/packages/shared/pkg/config"
)

// Config represents workflow service configuration
type Config struct {
	ServiceName    string
	Port           int
	Environment    string
	Version        string
	DBHost         string
	DBPort         int
	DBUser         string
	DBPassword     string
	DBName         string
	DBSSLMode      string
	MigrationsPath string
}

// Load reads configuration from environment
func Load() *Config {
	return &Config{
		ServiceName:    "workflow",
		Port:           config.GetEnvInt("PORT", 8085),
		Environment:    config.GetEnv("APP_ENV", "development"),
		Version:        "0.1.0",
		DBHost:         config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:         config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:         config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:     config.GetEnv("POSTGRES_PASSWORD", "eomp_dev_password"),
		DBName:         config.GetEnv("WORKFLOW_DB_NAME", "workflow_db"),
		DBSSLMode:      config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath: config.GetEnv("WORKFLOW_MIGRATIONS_PATH", "migrations"),
	}
}
