package config

import (
	"errors"
	"time"

	"eomp/packages/shared/pkg/config"
)

// Config represents auth service configuration loaded from environment variables.
type Config struct {
	ServiceName            string
	Port                   int
	Environment            string
	Version                string
	DBHost                 string
	DBPort                 int
	DBUser                 string
	DBPassword             string
	DBName                 string
	DBSSLMode              string
	MigrationsPath         string
	JWTSecret              string
	JWTAccessTTL           time.Duration
	JWTRefreshTTL          time.Duration
	BootstrapAdminEmail    string
	BootstrapAdminPassword string
	BootstrapAdminName     string
}

// Load reads configuration from the environment.
func Load() *Config {
	return &Config{
		ServiceName:            "auth",
		Port:                   config.GetEnvInt("PORT", 8081),
		Environment:            config.GetEnv("APP_ENV", "development"),
		Version:                "0.1.0",
		DBHost:                 config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:                 config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:                 config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:             config.GetEnv("POSTGRES_PASSWORD", ""),
		DBName:                 config.GetEnv("AUTH_DB_NAME", "auth_db"),
		DBSSLMode:              config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath:         config.GetEnv("AUTH_MIGRATIONS_PATH", "migrations"),
		JWTSecret:              config.GetEnv("JWT_SECRET", ""),
		JWTAccessTTL:           time.Duration(config.GetEnvInt("JWT_ACCESS_TTL_MINUTES", 60)) * time.Minute,
		JWTRefreshTTL:          time.Duration(config.GetEnvInt("JWT_REFRESH_TTL_DAYS", 7)) * 24 * time.Hour,
		BootstrapAdminEmail:    config.GetEnv("BOOTSTRAP_ADMIN_EMAIL", ""),
		BootstrapAdminPassword: config.GetEnv("BOOTSTRAP_ADMIN_PASSWORD", ""),
		BootstrapAdminName:     config.GetEnv("BOOTSTRAP_ADMIN_NAME", "Initial Administrator"),
	}
}

// Validate performs fail-fast verification in every runtime environment.
// Tests must inject explicit test-only credentials.
func (c *Config) Validate() error {
	if (c.BootstrapAdminEmail == "") != (c.BootstrapAdminPassword == "") {
		return errors.New("BOOTSTRAP_ADMIN_EMAIL and BOOTSTRAP_ADMIN_PASSWORD must be provided together")
	}

	if err := config.ValidateRequiredSecret(
		"JWT_SECRET",
		c.JWTSecret,
		32,
		"eomp-enterprise-super-secret-jwt-key-2026",
	); err != nil {
		return err
	}

	return config.ValidateRequiredSecret("POSTGRES_PASSWORD", c.DBPassword, 12, "eomp_dev_password")
}
