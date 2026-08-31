package config

import "eomp/packages/shared/pkg/config"

// Config represents helpdesk service configuration
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
}

// Load reads configuration from environment
func Load() *Config {
	return &Config{
		ServiceName:    "helpdesk",
		Port:           config.GetEnvInt("PORT", 8084),
		Environment:    config.GetEnv("APP_ENV", "development"),
		Version:        "0.1.0",
		DBHost:         config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:         config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:         config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:     config.GetEnv("POSTGRES_PASSWORD", ""),
		DBName:         config.GetEnv("HELPDESK_DB_NAME", "helpdesk_db"),
		DBSSLMode:      config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath: config.GetEnv("HELPDESK_MIGRATIONS_PATH", "migrations"),
		RabbitMQURL:    config.RabbitMQURL(),
	}
}

// Validate performs fail-fast configuration checks.
func (c *Config) Validate() error {
	return config.ValidateRequiredSecret("POSTGRES_PASSWORD", c.DBPassword, 12, "eomp_dev_password")
}
