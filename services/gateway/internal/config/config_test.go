package config

import "testing"

func TestValidateJWTSecret(t *testing.T) {
	for name, secret := range map[string]string{
		"empty":     "",
		"short":     "short",
		"published": "eomp-enterprise-super-secret-jwt-key-2026",
	} {
		t.Run(name, func(t *testing.T) {
			if err := (&Config{JWTSecret: secret}).Validate(); err == nil {
				t.Fatalf("expected %s JWT secret to be rejected", name)
			}
		})
	}

	if err := (&Config{JWTSecret: "test-only-jwt-secret-that-is-long-enough"}).Validate(); err != nil {
		t.Fatalf("expected explicit test secret to pass: %v", err)
	}
}
