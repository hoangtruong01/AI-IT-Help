package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Test Case 10.1: Strict RBAC enforcement.
func TestRBAC_RequireRoles(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ACCESS_GRANTED"))
	})

	rbacAdminManager := RequireRoles("ROLE_ADMIN", "ROLE_MANAGER")(dummyHandler)

	// 1. User with ROLE_EMPLOYEE tries to access Admin route -> 403 Forbidden
	reqUser := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqUser.Header.Set("X-User-Role", "ROLE_EMPLOYEE")
	recUser := httptest.NewRecorder()

	rbacAdminManager.ServeHTTP(recUser, reqUser)

	if recUser.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for ROLE_EMPLOYEE, got %d", recUser.Code)
	}

	// 2. User with ROLE_ADMIN -> 200 OK
	reqAdmin := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqAdmin.Header.Set("X-User-Role", "ROLE_ADMIN")
	recAdmin := httptest.NewRecorder()

	rbacAdminManager.ServeHTTP(recAdmin, reqAdmin)

	if recAdmin.Code != http.StatusOK {
		t.Errorf("expected 200 OK for ROLE_ADMIN, got %d", recAdmin.Code)
	}

	// 3. User with ROLE_MANAGER -> 200 OK
	reqManager := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqManager.Header.Set("X-User-Role", "ROLE_MANAGER")
	recManager := httptest.NewRecorder()

	rbacAdminManager.ServeHTTP(recManager, reqManager)

	if recManager.Code != http.StatusOK {
		t.Errorf("expected 200 OK for ROLE_MANAGER, got %d", recManager.Code)
	}
}

// Test Case 10.2: IP Rate Limiting.
func TestRateLimiter_IPLimit(t *testing.T) {
	dummyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Limit to 5 requests per second
	limiter := IPRateLimiter(5, 1*time.Second)(dummyHandler)

	// Send 5 valid requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.Header.Set("X-Forwarded-For", "192.168.1.100")
		rec := httptest.NewRecorder()
		limiter.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request #%d should pass, got status %d", i+1, rec.Code)
		}
	}

	// 6th request should be blocked with 429 Too Many Requests
	req6 := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	req6.Header.Set("X-Forwarded-For", "192.168.1.100")
	rec6 := httptest.NewRecorder()

	limiter.ServeHTTP(rec6, req6)

	if rec6.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests on rate limit breach, got %d", rec6.Code)
	}
	if rec6.Header().Get("Retry-After") != "60" {
		t.Errorf("expected Retry-After: 60 header, got %s", rec6.Header().Get("Retry-After"))
	}
}

// Test Case 10.3: Data Masking of sensitive fields.
func TestDataMasker(t *testing.T) {
	sensitivePayload := map[string]interface{}{
		"username":    "admin@eomp.local",
		"password":    "SuperSecretPassword123!",
		"api_key":     "sk_live_9921820491820391203",
		"credit_card": "4111 2222 3333 4444",
		"profile": map[string]interface{}{
			"auth_token": "secret-token-value",
			"city":       "Hanoi",
		},
	}

	masked := MaskSensitiveData(sensitivePayload)

	if masked["password"] != "********" {
		t.Errorf("expected masked password, got %v", masked["password"])
	}
	if masked["api_key"] != "********" {
		t.Errorf("expected masked api_key, got %v", masked["api_key"])
	}
	if masked["credit_card"] != "********" {
		t.Errorf("expected masked credit_card, got %v", masked["credit_card"])
	}
	profileMap := masked["profile"].(map[string]interface{})
	if profileMap["auth_token"] != "********" {
		t.Errorf("expected masked nested auth_token, got %v", profileMap["auth_token"])
	}
	if profileMap["city"] != "Hanoi" {
		t.Errorf("expected non-sensitive city preserved, got %v", profileMap["city"])
	}
}
