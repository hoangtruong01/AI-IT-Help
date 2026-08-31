package config

import "testing"

func validConfig() *Config {
	return &Config{
		JWTSecret:  "test-only-jwt-secret-that-is-long-enough",
		DBPassword: "test-db-password",
	}
}

func TestValidateRequiresSecretsInEveryEnvironment(t *testing.T) {
	for _, environment := range []string{"test", "development", "production"} {
		t.Run(environment, func(t *testing.T) {
			cfg := validConfig()
			cfg.Environment = environment
			cfg.DBPassword = ""
			if err := cfg.Validate(); err == nil {
				t.Fatalf("expected empty database password to fail in %s", environment)
			}
		})
	}
}

func TestValidateRejectsPublishedSecrets(t *testing.T) {
	cfg := validConfig()
	cfg.JWTSecret = "eomp-enterprise-super-secret-jwt-key-2026"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected published JWT secret to be rejected")
	}

	cfg = validConfig()
	cfg.DBPassword = "eomp_dev_password"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected published database password to be rejected")
	}
}

func TestValidateAcceptsExplicitTestCredentials(t *testing.T) {
	cfg := validConfig()
	cfg.Environment = "test"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected explicit test credentials to pass: %v", err)
	}
}
