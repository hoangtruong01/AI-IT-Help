package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"eomp/packages/shared/pkg/errors"
)

type ipBucket struct {
	timestamps []time.Time
}

// IPRateLimiter creates an in-memory sliding window rate limiter per IP address (Test Case 10.2).
// When requests exceed the limit within window, it returns HTTP 429 Too Many Requests.
func IPRateLimiter(limit int, window time.Duration) func(http.Handler) http.Handler {
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
			// Extract client IP address
			clientIP := r.Header.Get("X-Forwarded-For")
			if clientIP == "" {
				clientIP = strings.Split(r.RemoteAddr, ":")[0]
			} else {
				clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
			}
			if clientIP == "" {
				clientIP = "127.0.0.1"
			}

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
