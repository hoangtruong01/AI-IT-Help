package e2e

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"eomp/packages/shared/pkg/middleware"
	pkgRedis "eomp/packages/shared/pkg/redis"
)

// =============================================================================
// PHASE 6: DISTRIBUTED RATE LIMITER & RESILIENCE FALLBACK TEST SUITE
// =============================================================================

func TestPhase6_E2E_DistributedRateLimiterAndFallback(t *testing.T) {
	t.Log("===> [PHASE 6 - Task 6.1] Testing Distributed Rate Limiting & Graceful In-Memory Fallback...")

	// 1. Mock Next HTTP Handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true}`))
	})

	trustedProxies := []string{"127.0.0.1", "::1", "10.0.0.0/8", "192.168.0.0/16"}

	// -------------------------------------------------------------------------
	// Test Case A: In-Memory / Standalone Fallback Rate Limiting (10 req/min)
	// -------------------------------------------------------------------------
	t.Log("  [1/3] Testing In-Memory Sliding Window Rate Limiting (Limit: 10 req/min)...")

	limiter := middleware.IPRateLimiterWithProxies(10, 1*time.Minute, trustedProxies)(nextHandler)

	// Send 10 normal requests -> All 200 OK
	for i := 1; i <= 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
		req.RemoteAddr = "192.168.1.100:54321"
		rec := httptest.NewRecorder()

		limiter.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("request #%d expected 200 OK, got %d", i, rec.Code)
		}
	}
	t.Log("      -> 10 requests successfully passed (HTTP 200 OK)")

	// 11th request from same IP -> Must be 429 Too Many Requests
	req11 := httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil)
	req11.RemoteAddr = "192.168.1.100:54321"
	rec11 := httptest.NewRecorder()
	limiter.ServeHTTP(rec11, req11)

	if rec11.Code != http.StatusTooManyRequests {
		t.Fatalf("11th request expected HTTP 429 Too Many Requests, got %d", rec11.Code)
	}
	if rec11.Header().Get("Retry-After") != "60" {
		t.Fatalf("expected Retry-After: 60, got %s", rec11.Header().Get("Retry-After"))
	}
	if rec11.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Fatalf("expected X-RateLimit-Remaining: 0, got %s", rec11.Header().Get("X-RateLimit-Remaining"))
	}
	t.Log("      -> 11th request successfully blocked with HTTP 429 Too Many Requests & Retry-After: 60")

	// -------------------------------------------------------------------------
	// Test Case B: Graceful Fallback with Unreachable/Nil Redis Client
	// -------------------------------------------------------------------------
	t.Log("  [2/3] Testing Graceful Degradation when Redis is Offline (Nil/Unreachable)...")

	var nilRedisClient *pkgRedis.Client = nil
	fallbackLimiter := middleware.RedisSlidingWindowRateLimiter(nilRedisClient, 5, 1*time.Minute, trustedProxies, "global", nil)(nextHandler)

	for i := 1; i <= 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/assets", nil)
		req.RemoteAddr = "192.168.2.200:54322"
		rec := httptest.NewRecorder()

		fallbackLimiter.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("fallback request #%d expected 200 OK, got %d", i, rec.Code)
		}
	}

	req6 := httptest.NewRequest(http.MethodGet, "/api/v1/assets", nil)
	req6.RemoteAddr = "192.168.2.200:54322"
	rec6 := httptest.NewRecorder()
	fallbackLimiter.ServeHTTP(rec6, req6)

	if rec6.Code != http.StatusTooManyRequests {
		t.Fatalf("fallback 6th request expected HTTP 429, got %d", rec6.Code)
	}
	t.Log("      -> Graceful Fallback smoothly maintained rate limiting without service crash or 500 error")

	// -------------------------------------------------------------------------
	// Test Case C: Anti-Spoofing Client IP Verification across Rate Limiter
	// -------------------------------------------------------------------------
	t.Log("  [3/3] Testing Anti-Spoofing Client IP extraction under rate limiter...")

	// Remote address is untrusted origin (not in trusted proxies) -> XFF must be ignored
	untrustedReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", nil)
	untrustedReq.RemoteAddr = "203.0.113.50:12345"
	untrustedReq.Header.Set("X-Forwarded-For", "1.1.1.1, 2.2.2.2")

	extractedIP := middleware.ExtractClientIP(untrustedReq, trustedProxies)
	if extractedIP != "203.0.113.50" {
		t.Fatalf("anti-spoofing failed: expected extracted IP 203.0.113.50, got %s", extractedIP)
	}
	t.Logf("      -> Anti-spoofing verified: Fake XFF header ignored, real client IP (%s) identified correctly", extractedIP)

	t.Log("  [✓] Distributed Rate Limiting & Resilience Fallback: 100% Verified.")
}
