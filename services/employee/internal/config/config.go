package config

import (
	"eomp/packages/shared/pkg/config"
)

// Config represents employee service configuration
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

// Load reads configuration from the environment
func Load() *Config {
	return &Config{
		ServiceName:    "employee",
		Port:           config.GetEnvInt("PORT", 8082),
		Environment:    config.GetEnv("APP_ENV", "development"),
		Version:        "0.1.0",
		DBHost:         config.GetEnv("POSTGRES_HOST", "localhost"),
		DBPort:         config.GetEnvInt("POSTGRES_PORT", 5432),
		DBUser:         config.GetEnv("POSTGRES_USER", "eomp"),
		DBPassword:     config.GetEnv("POSTGRES_PASSWORD", "eomp_dev_password"),
		DBName:         config.GetEnv("EMPLOYEE_DB_NAME", "employee_db"),
		DBSSLMode:      config.GetEnv("POSTGRES_SSLMODE", "disable"),
		MigrationsPath: config.GetEnv("EMPLOYEE_MIGRATIONS_PATH", "migrations"),
	}
}
