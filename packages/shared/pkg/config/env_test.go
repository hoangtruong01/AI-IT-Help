package config

import (
	"os"
	"reflect"
	"strings"
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

func TestGetEnvBool(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")

	if val := GetEnvBool("TEST_BOOL", false); !val {
		t.Errorf("expected true, got false")
	}

	if val := GetEnvBool("NON_EXISTING_BOOL", true); !val {
		t.Errorf("expected default true, got false")
	}
}

func TestGetEnvSlice(t *testing.T) {
	os.Setenv("TEST_SLICE", "http://localhost:3000, http://127.0.0.1:3000 , https://app.eomp.local")
	defer os.Unsetenv("TEST_SLICE")

	expected := []string{"http://localhost:3000", "http://127.0.0.1:3000", "https://app.eomp.local"}
	val := GetEnvSlice("TEST_SLICE", nil)
	if !reflect.DeepEqual(val, expected) {
		t.Errorf("expected %v, got %v", expected, val)
	}

	defaultSlice := []string{"default1", "default2"}
	fallbackVal := GetEnvSlice("NON_EXISTING_SLICE", defaultSlice)
	if !reflect.DeepEqual(fallbackVal, defaultSlice) {
		t.Errorf("expected %v, got %v", defaultSlice, fallbackVal)
	}
}

func TestRequireEnv(t *testing.T) {
	os.Setenv("TEST_REQ", "secret_value")
	defer os.Unsetenv("TEST_REQ")

	val, err := RequireEnv("TEST_REQ")
	if err != nil || val != "secret_value" {
		t.Fatalf("expected 'secret_value', got error: %v", err)
	}

	_, errMissing := RequireEnv("MISSING_REQ_KEY")
	if errMissing == nil {
		t.Fatalf("expected error for missing required env, got nil")
	}
}

func TestRabbitMQURL_EscapesCredentials(t *testing.T) {
	t.Setenv("RABBITMQ_URL", "")
	t.Setenv("RABBITMQ_HOST", "rabbitmq-service")
	t.Setenv("RABBITMQ_PORT", "5672")
	t.Setenv("RABBITMQ_USER", "service@example")
	t.Setenv("RABBITMQ_PASSWORD", "p@ss:word/with spaces")

	value := RabbitMQURL()
	if strings.Contains(value, "p@ss:word/with spaces") {
		t.Fatal("RabbitMQ password was not URL-escaped")
	}
	if !strings.Contains(value, "rabbitmq-service:5672") {
		t.Fatalf("unexpected RabbitMQ URL: %s", value)
	}
}
