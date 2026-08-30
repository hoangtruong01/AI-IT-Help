package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"eomp/packages/shared/pkg/errors"
	pkgRedis "eomp/packages/shared/pkg/redis"
	"github.com/redis/go-redis/v9"
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
			trustedProxies = []string{"127.0.0.1", "::1", "localhost"}
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

// IPRateLimiterWithProxies creates a standalone in-memory rate limiter with custom trusted proxy configuration.
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

// Lua Script for Atomic Redis Sliding Window Rate Limiting
var slidingWindowLuaScript = redis.NewScript(`
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local clearBefore = now - window

redis.call('ZREMRANGEBYSCORE', key, 0, clearBefore)
local currentRequests = redis.call('ZCARD', key)

if currentRequests < limit then
    redis.call('ZADD', key, now, ARGV[4])
    redis.call('EXPIRE', key, math.ceil(window / 1000) + 5)
    return {1, limit - currentRequests - 1}
else
    return {0, 0}
end
`)

var slidingWindowSequence uint64

// RedisSlidingWindowRateLimiter creates a distributed rate limiter backed by Redis.
// If Redis is unavailable or nil, it gracefully falls back to the in-memory sliding window limiter.
func RedisSlidingWindowRateLimiter(
	redisClient *pkgRedis.Client,
	limit int,
	window time.Duration,
	trustedProxies []string,
	scope string,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	inMemoryFallback := IPRateLimiterWithProxies(limit, window, trustedProxies)

	if scope == "" {
		scope = "global"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rawRdb := redisClient.Raw()
			if rawRdb == nil {
				// Graceful in-memory fallback
				inMemoryFallback(next).ServeHTTP(w, r)
				return
			}

			clientIP := ExtractClientIP(r, trustedProxies)
			redisKey := fmt.Sprintf("eomp:ratelimit:%s:%s", scope, clientIP)

			nowMillis := time.Now().UnixMilli()
			windowMillis := window.Milliseconds()
			member := fmt.Sprintf("%d-%d", nowMillis, atomic.AddUint64(&slidingWindowSequence, 1))

			ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
			defer cancel()

			res, err := slidingWindowLuaScript.Run(ctx, rawRdb, []string{redisKey}, nowMillis, windowMillis, limit, member).Slice()
			if err != nil {
				// Redis error / timeout -> Graceful in-memory fallback
				if logger != nil {
					logger.Warn("redis rate limit failed, falling back to in-memory",
						slog.String("ip", clientIP),
						slog.String("key", redisKey),
						slog.Any("error", err),
					)
				}
				inMemoryFallback(next).ServeHTTP(w, r)
				return
			}

			allowed := res[0].(int64) == 1
			remaining := res[1].(int64)

			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			if !allowed {
				w.Header().Set("Retry-After", "60")
				errors.WriteHTTP(w, errors.New(http.StatusTooManyRequests, "TOO_MANY_REQUESTS", "rate limit exceeded: too many requests from your IP across cluster, please retry in 60 seconds"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// StrictRedisAuthRateLimiter creates a strict brute-force rate limiter (10 req/min default) for auth endpoints with Redis backing.
func StrictRedisAuthRateLimiter(
	redisClient *pkgRedis.Client,
	limit int,
	window time.Duration,
	trustedProxies []string,
	logger *slog.Logger,
) func(http.Handler) http.Handler {
	return RedisSlidingWindowRateLimiter(redisClient, limit, window, trustedProxies, "auth", logger)
}
