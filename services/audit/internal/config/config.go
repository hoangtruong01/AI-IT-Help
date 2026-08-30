package config

import (
	"errors"

	"eomp/packages/shared/pkg/config"
)

// Config represents service configuration loaded from environment variables.
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
	RabbitMQURL    string
	AuditHMACKey   string
}

// Load reads configuration from the environment with sensible defaults.
func Load() *Config {
	return &Config{
		ServiceName:    "audit",
		Port:           config.GetEnvInt("PORT", 8089),
		Environment:    config.GetEnv("APP_ENV", "development"),
		Version:        "0.1.0",
		DBHost:         config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:         config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:         config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:     config.GetEnv("POSTGRES_PASSWORD", "eomp_dev_password"),
		DBName:         config.GetEnv("AUDIT_DB_NAME", "audit_db"),
		DBSSLMode:      config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath: config.GetEnv("AUDIT_MIGRATIONS_PATH", "migrations"),
		RabbitMQURL:    config.RabbitMQURL(),
		AuditHMACKey:   config.GetEnv("AUDIT_HMAC_KEY", "eomp-development-audit-hmac-key-change-me"),
	}
}

// Validate performs fail-fast configuration checks.
func (c *Config) Validate() error {
	if len(c.AuditHMACKey) < 32 {
		return errors.New("security violation: AUDIT_HMAC_KEY must be at least 32 characters")
	}
	if c.Environment == "production" {
		if c.DBPassword == "eomp_dev_password" || c.DBPassword == "" {
			return errors.New("security violation: default dev DB_PASSWORD is prohibited in production")
		}
		if c.AuditHMACKey == "eomp-development-audit-hmac-key-change-me" {
			return errors.New("security violation: default development AUDIT_HMAC_KEY is prohibited in production")
		}
	}
	return nil
}
