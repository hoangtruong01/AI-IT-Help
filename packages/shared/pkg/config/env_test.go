package config

import (
	"os"
	"testing"
)

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_VAR", "hello")
	defer os.Unsetenv("TEST_VAR")

	if val := GetEnv("TEST_VAR", "default"); val != "hello" {
		t.Errorf("expected 'hello', got '%s'", val)
	}

	if val := GetEnv("NON_EXISTING_VAR", "fallback"); val != "fallback" {
		t.Errorf("expected 'fallback', got '%s'", val)
	}
}

func TestGetEnvInt(t *testing.T) {
	os.Setenv("TEST_INT", "1234")
	defer os.Unsetenv("TEST_INT")

	if val := GetEnvInt("TEST_INT", 0); val != 1234 {
		t.Errorf("expected 1234, got %d", val)
	}

	if val := GetEnvInt("NON_EXISTING_INT", 99); val != 99 {
		t.Errorf("expected 99, got %d", val)
	}
}
