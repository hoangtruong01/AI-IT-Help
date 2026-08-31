package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// GetEnv reads an environment variable with a fallback default.
func GetEnv(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		return strings.TrimSpace(val)
	}
	return fallback
}

// GetEnvInt reads an integer environment variable with a fallback default.
func GetEnvInt(key string, fallback int) int {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		if intVal, err := strconv.Atoi(strings.TrimSpace(val)); err == nil {
			return intVal
		}
	}
	return fallback
}

// GetEnvBool reads a boolean environment variable with a fallback default.
func GetEnvBool(key string, fallback bool) bool {
	if val, ok := os.LookupEnv(key); ok && strings.TrimSpace(val) != "" {
		if boolVal, err := strconv.ParseBool(strings.TrimSpace(val)); err == nil {
			return boolVal
		}
	}
	return fallback
}

// GetEnvSlice reads a comma-separated environment variable into a string slice.
func GetEnvSlice(key string, fallback []string) []string {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return fallback
	}
	parts := strings.Split(val, ",")
	var result []string
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}

// RequireEnv reads a mandatory environment variable or returns an error (Fail-Fast).
func RequireEnv(key string) (string, error) {
	val, ok := os.LookupEnv(key)
	if !ok || strings.TrimSpace(val) == "" {
		return "", fmt.Errorf("required environment variable %q is not set or empty", key)
	}
	return strings.TrimSpace(val), nil
}

// MustGetEnv reads a mandatory environment variable and panics if not found.
func MustGetEnv(key string) string {
	val, err := RequireEnv(key)
	if err != nil {
		panic(err)
	}
	return val
}

// ValidateRequiredSecret validates a runtime secret without making security
// behavior depend on APP_ENV. Tests should inject explicit test-only values.
func ValidateRequiredSecret(key, value string, minLength int, forbiddenValues ...string) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("security violation: %s environment variable is required but not set", key)
	}
	if minLength > 0 && len(value) < minLength {
		return fmt.Errorf("security violation: %s must be at least %d characters, got %d", key, minLength, len(value))
	}
	for _, forbidden := range forbiddenValues {
		if value == forbidden {
			return fmt.Errorf("security violation: %s matches a known public value and must be rotated", key)
		}
	}
	return nil
}

// RabbitMQURL returns an explicit RABBITMQ_URL when supplied, otherwise it
// safely builds one from component variables and URL-escapes the credentials.
func RabbitMQURL() string {
	if value := GetEnv("RABBITMQ_URL", ""); value != "" {
		return value
	}
	host := GetEnv("RABBITMQ_HOST", "localhost")
	port := GetEnvInt("RABBITMQ_PORT", 5672)
	user := GetEnv("RABBITMQ_USER", "guest")
	password := GetEnv("RABBITMQ_PASSWORD", "")
	u := &url.URL{
		Scheme: "amqp",
		User:   url.UserPassword(user, password),
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   "/",
	}
	return u.String()
}
