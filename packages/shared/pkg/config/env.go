package config

import (
	"os"
	"strconv"
)

// GetEnv reads an environment variable with a fallback default.
func GetEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return fallback
}

// GetEnvInt reads an integer environment variable with a fallback default.
func GetEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return fallback
}
