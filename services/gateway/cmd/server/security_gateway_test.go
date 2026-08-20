package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eomp/packages/shared/pkg/middleware"
)

// Test Case 10.1: API Gateway RBAC Enforcement (Employee gets 403 Forbidden, Admin gets 200 OK).
func TestGateway_TestCase_10_1_RBACEnforcement(t *testing.T) {
	mockBackend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	adminOnly := middleware.RequireRoles("ROLE_ADMIN", "ROLE_MANAGER")(mockBackend)

	// 1. Employee tries to access admin audit route -> 403 Forbidden
	reqEmp := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqEmp.Header.Set("X-User-Role", "ROLE_EMPLOYEE")
	recEmp := httptest.NewRecorder()

	adminOnly.ServeHTTP(recEmp, reqEmp)

	if recEmp.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for employee, got %d", recEmp.Code)
	}

	// 2. Admin access -> 200 OK
	reqAdmin := httptest.NewRequest("GET", "/api/v1/audit/logs", nil)
	reqAdmin.Header.Set("X-User-Role", "ROLE_ADMIN")
	recAdmin := httptest.NewRecorder()

	adminOnly.ServeHTTP(recAdmin, reqAdmin)

	if recAdmin.Code != http.StatusOK {
		t.Errorf("expected 200 OK for admin, got %d", recAdmin.Code)
	}
}

// Test Case 10.2: API Gateway Rate Limiting (Excessive requests receive 429 Too Many Requests).
func TestGateway_TestCase_10_2_RateLimiter(t *testing.T) {
	mockBackend := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	limiter := middleware.IPRateLimiter(5, 500*time.Millisecond)(mockBackend)

	// Send 5 requests from same IP -> all OK
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.195")
		rec := httptest.NewRecorder()
		limiter.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request #%d should pass, got %d", i+1, rec.Code)
		}
	}

	// 6th request triggers rate limit -> 429 Too Many Requests
	reqSpam := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
	reqSpam.Header.Set("X-Forwarded-For", "203.0.113.195")
	recSpam := httptest.NewRecorder()

	limiter.ServeHTTP(recSpam, reqSpam)

	if recSpam.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests on rate limit breach, got %d", recSpam.Code)
	}
}
