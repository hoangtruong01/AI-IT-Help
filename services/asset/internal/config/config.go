package config

import (
	"errors"

	"eomp/packages/shared/pkg/config"
)

// Config represents asset service configuration
type Config struct {
	ServiceName        string
	Port               int
	Environment        string
	Version            string
	DBHost             string
	DBPort             int
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	MigrationsPath     string
	HelpdeskServiceURL string
	RabbitMQURL        string
}

// Load reads configuration from environment
func Load() *Config {
	return &Config{
		ServiceName:        "asset",
		Port:               config.GetEnvInt("PORT", 8083),
		Environment:        config.GetEnv("APP_ENV", "development"),
		Version:            "0.1.0",
		DBHost:             config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:             config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:             config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:         config.GetEnv("POSTGRES_PASSWORD", "eomp_dev_password"),
		DBName:             config.GetEnv("ASSET_DB_NAME", "asset_db"),
		DBSSLMode:          config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath:     config.GetEnv("ASSET_MIGRATIONS_PATH", "migrations"),
		HelpdeskServiceURL: config.GetEnv("HELPDESK_SERVICE_URL", "http://localhost:8084"),
		RabbitMQURL:        config.GetEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
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
