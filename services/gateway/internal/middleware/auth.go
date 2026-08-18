package middleware

import (
	"net/http"
	"strings"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/errors"
)

// GatewayAuth verifies JWT and attaches X-User-* headers before proxying
func GatewayAuth(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteHTTP(w, errors.Unauthorized("missing authorization token"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				errors.WriteHTTP(w, errors.Unauthorized("invalid authorization format"))
				return
			}

			claims, err := jwtManager.ValidateToken(parts[1])
			if err != nil {
				errors.WriteHTTP(w, errors.Unauthorized("invalid or expired token"))
				return
			}

			// Attach identity headers to forwarded request
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Role", claims.Role)
			if claims.DepartmentID != "" {
				r.Header.Set("X-User-Department", claims.DepartmentID)
			}

			next.ServeHTTP(w, r)
		})
	}
}
