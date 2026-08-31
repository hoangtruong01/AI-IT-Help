package middleware

import (
	"net/http"
	"os"
	"strings"
)

// DynamicCORS returns a CORS middleware configured with a specific whitelist of allowed origins.
func DynamicCORS(allowedOrigins []string) func(http.Handler) http.Handler {
	// Normalize allowed origins
	originMap := make(map[string]bool)
	for _, o := range allowedOrigins {
		trimmed := strings.TrimRight(strings.TrimSpace(o), "/")
		if trimmed != "" {
			originMap[trimmed] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimRight(strings.TrimSpace(r.Header.Get("Origin")), "/")

			// Check if origin is allowed or if wildcard is enabled
			isAllowed := false
			if origin != "" {
				if originMap["*"] || originMap[origin] {
					isAllowed = true
				}
			}

			if isAllowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
			w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID, X-RateLimit-Limit, X-RateLimit-Remaining, Retry-After")
			w.Header().Set("Access-Control-Max-Age", "86400")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// CORS provides standard dynamic cross-origin resource sharing reading from CORS_ALLOWED_ORIGINS.
func CORS(next http.Handler) http.Handler {
	defaultOrigins := []string{"http://localhost:3000", "http://127.0.0.1:3000", "http://localhost:8080"}
	if envOrigins := os.Getenv("CORS_ALLOWED_ORIGINS"); envOrigins != "" {
		parts := strings.Split(envOrigins, ",")
		var origins []string
		for _, p := range parts {
			if t := strings.TrimSpace(p); t != "" {
				origins = append(origins, t)
			}
		}
		if len(origins) > 0 {
			defaultOrigins = origins
		}
	}
	return DynamicCORS(defaultOrigins)(next)
}
