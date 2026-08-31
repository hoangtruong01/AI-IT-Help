package config

import "eomp/packages/shared/pkg/config"

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
}

// Load reads configuration from the environment with sensible defaults.
func Load() *Config {
	return &Config{
		ServiceName:    "reporting",
		Port:           config.GetEnvInt("PORT", 8090),
		Environment:    config.GetEnv("APP_ENV", "development"),
		Version:        "1.0.0",
		DBHost:         config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:         config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:         config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:     config.GetEnv("POSTGRES_PASSWORD", ""),
		DBName:         config.GetEnv("REPORTING_DB_NAME", "reporting_db"),
		DBSSLMode:      config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath: config.GetEnv("REPORTING_MIGRATIONS_PATH", "migrations"),
		RabbitMQURL:    config.RabbitMQURL(),
	}
}

// Validate performs fail-fast configuration checks.
func (c *Config) Validate() error {
	return config.ValidateRequiredSecret("POSTGRES_PASSWORD", c.DBPassword, 12, "eomp_dev_password")
}
