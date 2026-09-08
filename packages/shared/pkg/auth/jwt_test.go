package auth_test

import (
	"testing"
	"time"

	"eomp/packages/shared/pkg/auth"
	"github.com/golang-jwt/jwt/v5"
)

func TestJWT_ValidTokenLifecycle(t *testing.T) {
	manager := auth.NewJWTManager("super-secret-key-32-characters-minimum!", 15*time.Minute, 24*time.Hour)

	accessToken, refreshToken, err := manager.GenerateTokenPair("usr-123", "user@eomp.local", "AGENT", "dep-it", "Support Agent")
	if err != nil {
		t.Fatalf("unexpected error generating token pair: %v", err)
	}

	claims, err := manager.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("expected access token to be valid, got error: %v", err)
	}
	if claims.UserID != "usr-123" || claims.Role != "AGENT" || claims.Email != "user@eomp.local" {
		t.Errorf("unexpected claims: %+v", claims)
	}

	refreshClaims, err := manager.ValidateRefreshToken(refreshToken)
	if err != nil {
		t.Fatalf("expected refresh token to be valid, got error: %v", err)
	}
	if refreshClaims.Subject != "usr-123" {
		t.Errorf("unexpected refresh subject: %s", refreshClaims.Subject)
	}
}

func TestJWT_RejectsWrongIssuer(t *testing.T) {
	key := []byte("super-secret-key-32-characters-minimum!")
	manager := auth.NewJWTManager(string(key), 15*time.Minute, 24*time.Hour)

	claims := auth.UserClaims{
		UserID:    "usr-rogue",
		Email:     "rogue@attacker.com",
		Role:      "ADMIN",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "tok-1",
			Subject:   "usr-rogue",
			Issuer:    "malicious-third-party-issuer",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	forgedToken, err := tokenObj.SignedString(key)
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	_, err = manager.ValidateToken(forgedToken)
	if err == nil {
		t.Fatal("expected ValidateToken to reject token with invalid issuer")
	}
}

func TestJWT_RejectsNonHS256Algorithm(t *testing.T) {
	key := []byte("super-secret-key-32-characters-minimum!")
	manager := auth.NewJWTManager(string(key), 15*time.Minute, 24*time.Hour)

	claims := auth.UserClaims{
		UserID:    "usr-rogue",
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        "tok-2",
			Subject:   "usr-rogue",
			Issuer:    "eomp-auth-service",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
		},
	}
	// Sign with HS384 instead of pinned HS256
	tokenObj := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	hs384Token, err := tokenObj.SignedString(key)
	if err != nil {
		t.Fatalf("failed to create test token: %v", err)
	}

	_, err = manager.ValidateToken(hs384Token)
	if err == nil {
		t.Fatal("expected ValidateToken to reject non-HS256 algorithm (HS384)")
	}
}
