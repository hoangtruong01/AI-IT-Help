package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eomp/packages/shared/pkg/auth"
	sharedMiddleware "eomp/packages/shared/pkg/middleware"
)

func TestStripIdentityHeaders(t *testing.T) {
	handler := StripIdentityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, h := range identityHeaders {
			if val := r.Header.Get(h); val != "" {
				t.Errorf("expected header %s to be stripped, got %q", h, val)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
	req.Header.Set("X-User-ID", "attacker-uuid")
	req.Header.Set("X-User-Email", "attacker@evil.com")
	req.Header.Set("X-User-Role", "ROLE_ADMIN")
	req.Header.Set("X-User-Department", "dept-finance")
	req.Header.Set("X-User-Name", "Fake Admin")
	req.Header.Set("X-User-Future-Claim", "must-also-be-removed")
	req.Header.Set("X-Department-ID", "dept-finance")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestStripIdentityHeaders_RemovesAnyUserHeader(t *testing.T) {
	handler := StripIdentityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-User-Future-Claim"); got != "" {
			t.Fatalf("expected future identity header to be stripped, got %q", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
	req.Header.Set("X-User-Future-Claim", "attacker-controlled")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestGatewayAuth_MissingToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", time.Hour, 24*time.Hour)
	mw := GatewayAuth(jwtManager)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for missing token, got %d", rec.Code)
	}
}

func TestGatewayAuth_InvalidToken(t *testing.T) {
	jwtManager := auth.NewJWTManager("test-secret-key-that-is-at-least-32-chars-long", time.Hour, 24*time.Hour)
	mw := GatewayAuth(jwtManager)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token-string")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401 Unauthorized for invalid token, got %d", rec.Code)
	}
}

func TestGatewayAuth_ValidToken_SetsHeadersUnconditionally(t *testing.T) {
	secret := "test-secret-key-that-is-at-least-32-chars-long"
	jwtManager := auth.NewJWTManager(secret, time.Hour, 24*time.Hour)

	accessToken, _, err := jwtManager.GenerateTokenPair("user-123", "user@eomp.com", "ROLE_EMPLOYEE", "", "Regular User")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	var capturedHeaders struct {
		userID     string
		email      string
		role       string
		department string
		name       string
	}

	mw := GatewayAuth(jwtManager)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeaders.userID = r.Header.Get("X-User-ID")
		capturedHeaders.email = r.Header.Get("X-User-Email")
		capturedHeaders.role = r.Header.Get("X-User-Role")
		capturedHeaders.department = r.Header.Get("X-User-Department")
		capturedHeaders.name = r.Header.Get("X-User-Name")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	// Try spoofing a department on an account with no department
	req.Header.Set("X-User-Department", "spoofed-dept")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	if capturedHeaders.userID != "user-123" {
		t.Errorf("expected X-User-ID to be 'user-123', got %q", capturedHeaders.userID)
	}
	if capturedHeaders.email != "user@eomp.com" {
		t.Errorf("expected X-User-Email to be 'user@eomp.com', got %q", capturedHeaders.email)
	}
	if capturedHeaders.role != "ROLE_EMPLOYEE" {
		t.Errorf("expected X-User-Role to be 'ROLE_EMPLOYEE', got %q", capturedHeaders.role)
	}
	// Crucial check: Empty department in token MUST overwrite spoofed department with empty string
	if capturedHeaders.department != "" {
		t.Errorf("expected X-User-Department to be overwritten with empty string, got %q", capturedHeaders.department)
	}
	if capturedHeaders.name != "Regular User" {
		t.Errorf("expected X-User-Name to be 'Regular User', got %q", capturedHeaders.name)
	}
}

func TestGatewayChain_SpoofedAdminHeaderCannotBypassRBAC(t *testing.T) {
	secret := "test-secret-key-that-is-at-least-32-chars-long"
	jwtManager := auth.NewJWTManager(secret, time.Hour, 24*time.Hour)
	accessToken, _, err := jwtManager.GenerateTokenPair("employee-123", "employee@eomp.com", "ROLE_EMPLOYEE", "dept-it", "Employee")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	protected := sharedMiddleware.RequireRoles("ROLE_ADMIN")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler := StripIdentityHeaders(GatewayAuth(jwtManager)(protected))

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/admin-only", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("X-User-Role", "ROLE_ADMIN")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected spoofed admin header to be rejected with 403, got %d", rec.Code)
	}
}
