package middleware

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"eomp/packages/shared/pkg/errors"
)

type ipBucket struct {
	timestamps []time.Time
}

// ExtractClientIP securely extracts client IP address with anti-spoofing validation.
// It only trusts X-Forwarded-For / X-Real-IP if the direct request origin (RemoteAddr)
// is verified to be in the trusted proxy whitelist.
func ExtractClientIP(r *http.Request, trustedProxies []string) string {
	remoteHost := r.RemoteAddr
	if host, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		remoteHost = host
	}
	remoteHost = strings.TrimSpace(remoteHost)

	if len(trustedProxies) == 0 {
		if envProxies := os.Getenv("TRUSTED_PROXIES"); envProxies != "" {
			parts := strings.Split(envProxies, ",")
			for _, p := range parts {
				if t := strings.TrimSpace(p); t != "" {
					trustedProxies = append(trustedProxies, t)
				}
			}
		}
		if len(trustedProxies) == 0 {
			trustedProxies = []string{"127.0.0.1", "::1", "localhost", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
		}
	}

	// Check if RemoteAddr is a trusted proxy
	isTrusted := isTrustedIP(remoteHost, trustedProxies)

	if isTrusted {
		// Trust X-Forwarded-For or X-Real-IP from trusted proxy
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			parts := strings.Split(xff, ",")
			if len(parts) > 0 {
				clientIP := strings.TrimSpace(parts[0])
				if clientIP != "" {
					return clientIP
				}
			}
		}
		if xRealIP := strings.TrimSpace(r.Header.Get("X-Real-IP")); xRealIP != "" {
			return xRealIP
		}
	}

	// Fallback to RemoteAddr directly (Anti-Spoofing: ignore untrusted XFF)
	if remoteHost != "" {
		return remoteHost
	}
	return "127.0.0.1"
}

func isTrustedIP(ipStr string, trustedList []string) bool {
	if ipStr == "127.0.0.1" || ipStr == "::1" || ipStr == "localhost" {
		return true
	}
	parsedIP := net.ParseIP(ipStr)

	for _, entry := range trustedList {
		entry = strings.TrimSpace(entry)
		if entry == ipStr {
			return true
		}
		// Check CIDR block
		if strings.Contains(entry, "/") {
			if _, cidrNet, err := net.ParseCIDR(entry); err == nil && parsedIP != nil {
				if cidrNet.Contains(parsedIP) {
					return true
				}
			}
		}
	}
	return false
}

// IPRateLimiter creates an in-memory sliding window rate limiter per client IP.
func IPRateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
	return IPRateLimiterWithProxies(limit, window, nil)
}

// StrictAuthRateLimiter creates a strict brute-force rate limiter (10 req/min default) for auth endpoints.
func StrictAuthRateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
	return IPRateLimiterWithProxies(limit, window, nil)
}

// IPRateLimiterWithProxies creates a rate limiter with custom trusted proxy configuration.
func IPRateLimiterWithProxies(limit int, window time.Duration, trustedProxies []string) func(http.Handler) http.Handler {
	var mu sync.Mutex
	clients := make(map[string]*ipBucket)

	// Background routine to clean up stale IP buckets
	go func() {
		ticker := time.NewTicker(window * 2)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			now := time.Now()
			for ip, bucket := range clients {
				var valid []time.Time
				for _, t := range bucket.timestamps {
					if now.Sub(t) <= window {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(clients, ip)
				} else {
					bucket.timestamps = valid
				}
			}
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := ExtractClientIP(r, trustedProxies)
			now := time.Now()

			mu.Lock()
			bucket, exists := clients[clientIP]
			if !exists {
				bucket = &ipBucket{}
				clients[clientIP] = bucket
			}

			// Filter out timestamps outside current window
			var valid []time.Time
			for _, t := range bucket.timestamps {
				if now.Sub(t) <= window {
					valid = append(valid, t)
				}
			}
			bucket.timestamps = valid

			// Check if threshold exceeded
			if len(bucket.timestamps) >= limit {
				mu.Unlock()
				w.Header().Set("Retry-After", "60")
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				errors.WriteHTTP(w, errors.New(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "rate limit exceeded: too many requests from your IP, please retry in 60 seconds"))
				return
			}

			// Record current request
			bucket.timestamps = append(bucket.timestamps, now)
			remaining := limit - len(bucket.timestamps)
			mu.Unlock()

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			next.ServeHTTP(w, r)
		})
	}
}
