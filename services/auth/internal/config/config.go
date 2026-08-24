package config

import (
	"errors"
	"fmt"
	"time"

	"eomp/packages/shared/pkg/config"
)

// Config represents auth service configuration loaded from environment variables.
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
	JWTSecret      string
	JWTAccessTTL   time.Duration
	JWTRefreshTTL  time.Duration
}

// Load reads configuration from the environment with validation.
func Load() *Config {
	return &Config{
		ServiceName:    "auth",
		Port:           config.GetEnvInt("PORT", 8081),
		Environment:    config.GetEnv("APP_ENV", "development"),
		Version:        "0.1.0",
		DBHost:         config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:         config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:         config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:     config.GetEnv("POSTGRES_PASSWORD", "eomp_dev_password"),
		DBName:         config.GetEnv("AUTH_DB_NAME", "auth_db"),
		DBSSLMode:      config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath: config.GetEnv("AUTH_MIGRATIONS_PATH", "migrations"),
		JWTSecret:      config.GetEnv("JWT_SECRET", "eomp-enterprise-super-secret-jwt-key-2026"),
		JWTAccessTTL:   time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)) * time.Minute,
		JWTRefreshTTL:  time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour,
	}
}

// Validate performs fail-fast verification on critical configuration items.
func (c *Config) Validate() error {
	if c.JWTSecret == "" {
		return errors.New("security violation: JWT_SECRET environment variable must not be empty")
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("security violation: JWT_SECRET must be at least 16 characters long for HMAC-SHA256, got length %d", len(c.JWTSecret))
	}
	if c.Environment == "production" {
		if c.JWTSecret == "eomp-enterprise-super-secret-jwt-key-2026" {
			return errors.New("security violation: default dev JWT_SECRET is strictly prohibited in production")
		}
		if c.DBPassword == "eomp_dev_password" || c.DBPassword == "" {
			return errors.New("security violation: default dev DB_PASSWORD is strictly prohibited in production")
		}
	}
	return nil
}
