package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eomp/packages/shared/pkg/auth"
	"eomp/services/auth/internal/config"
	"eomp/services/auth/internal/handler"
	"eomp/services/auth/internal/model"
)

func TestHealthHandler(t *testing.T) {
	cfg := config.Load()
	h := handler.NewHealthHandler(cfg)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.Check(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var body map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body["service"] != "auth" {
		t.Errorf("expected service 'auth', got '%v'", body["service"])
	}
	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%v'", body["status"])
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "SecretP@ssw0rd!"
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	if !auth.CheckPassword(password, hash) {
		t.Errorf("expected password to match hash")
	}

	if auth.CheckPassword("WrongPassword", hash) {
		t.Errorf("expected wrong password to fail check")
	}
}

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key-1234567890123456"
	manager := auth.NewJWTManager(secret, 15*time.Minute, 24*time.Hour)

	userID := "user-123"
	email := "test@eomp.local"
	role := model.RoleManager
	deptID := "dept-456"
	fullName := "John Doe"

	accessToken, refreshToken, err := manager.GenerateTokenPair(userID, email, role, deptID, fullName)
	if err != nil {
		t.Fatalf("failed to generate tokens: %v", err)
	}

	if accessToken == "" || refreshToken == "" {
		t.Errorf("tokens must not be empty")
	}
	if _, err := manager.ValidateToken(refreshToken); err == nil {
		t.Fatal("refresh token must not be accepted as an access token")
	}
	if _, err := manager.ValidateRefreshToken(refreshToken); err != nil {
		t.Fatalf("signed refresh token should validate: %v", err)
	}

	claims, err := manager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("failed to validate access token: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %s, got %s", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected email %s, got %s", email, claims.Email)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}
	if claims.FullName != fullName {
		t.Errorf("expected fullName %s, got %s", fullName, claims.FullName)
	}
}
