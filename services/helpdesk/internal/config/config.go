package config

import (
	"eomp/packages/shared/pkg/config"
)

// Config represents service configuration loaded from environment variables.
type Config struct {
	ServiceName string
	Port        int
	Environment string
	Version     string
}

// Load reads configuration from the environment with sensible defaults.
func Load() *Config {
	return &Config{
		ServiceName: "helpdesk",
		Port:        config.GetEnvInt("PORT", 8084),
		Environment: config.GetEnv("APP_ENV", "development"),
		Version:     "0.1.0",
	}
}
