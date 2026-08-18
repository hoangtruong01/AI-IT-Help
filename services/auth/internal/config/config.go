package config

import (
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

// Load reads configuration from the environment with sensible defaults.
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
