package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// UserClaims defines custom claims embedded in JWT
type UserClaims struct {
	UserID       string `json:"user_id"`
	Email        string `json:"email"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id,omitempty"`
	FullName     string `json:"full_name"`
	TokenType    string `json:"token_type"`
	jwt.RegisteredClaims
}

// RefreshClaims contains the minimum identity needed to rotate a refresh token.
type RefreshClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager handles token issuance and verification
type JWTManager struct {
	secretKey  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

// NewJWTManager constructs a new JWT manager instance
func NewJWTManager(secretKey string, accessTTL, refreshTTL time.Duration) *JWTManager {
	if accessTTL == 0 {
		accessTTL = 60 * time.Minute
	}
	if refreshTTL == 0 {
		refreshTTL = 7 * 24 * time.Hour
	}
	return &JWTManager{
		secretKey:  []byte(secretKey),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func randomTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate token id: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

// AccessTTL returns the configured access-token lifetime.
func (m *JWTManager) AccessTTL() time.Duration { return m.accessTTL }

// RefreshTTL returns the configured refresh-token lifetime.
func (m *JWTManager) RefreshTTL() time.Duration { return m.refreshTTL }

// GenerateTokenPair issues both an Access Token and a Refresh Token
func (m *JWTManager) GenerateTokenPair(userID, email, role, departmentID, fullName string) (accessToken string, refreshToken string, err error) {
	now := time.Now()
	accessID, err := randomTokenID()
	if err != nil {
		return "", "", err
	}
	refreshID, err := randomTokenID()
	if err != nil {
		return "", "", err
	}

	// 1. Access Token
	accessClaims := UserClaims{
		UserID:       userID,
		Email:        email,
		Role:         role,
		DepartmentID: departmentID,
		FullName:     fullName,
		TokenType:    "access",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        accessID,
			Subject:   userID,
			Issuer:    "eomp-auth-service",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	accessObj := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessToken, err = accessObj.SignedString(m.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign access token: %w", err)
	}

	// 2. Refresh Token
	refreshClaims := RefreshClaims{
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        refreshID,
			Subject:   userID,
			Issuer:    "eomp-auth-service",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
		},
	}
	refreshObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err = refreshObj.SignedString(m.secretKey)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// ValidateToken verifies token signature and expiration, returning user claims
func (m *JWTManager) ValidateToken(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid || claims.TokenType != "access" || claims.UserID == "" || claims.ID == "" {
		return nil, errors.New("invalid or expired token claims")
	}

	return claims, nil
}

// ValidateRefreshToken verifies the signature, expiry, issuer and token type.
func (m *JWTManager) ValidateRefreshToken(tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	}, jwt.WithIssuer("eomp-auth-service"))
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid || claims.TokenType != "refresh" || claims.Subject == "" || claims.ID == "" {
		return nil, errors.New("invalid or expired refresh token claims")
	}
	return claims, nil
}

// HashPassword hashes a raw password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(bytes), nil
}

// CheckPassword compares a plaintext password against a bcrypt hash
func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
