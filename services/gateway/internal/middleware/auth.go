package middleware

import (
	"net/http"
	"strings"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/errors"
)

// identityHeaders lists all X-User-* headers that must be stripped from
// incoming requests to prevent client-side identity spoofing.
var identityHeaders = []string{
	"X-User-ID",
	"X-User-Email",
	"X-User-Role",
	"X-User-Department",
	"X-User-Name",
	"X-Department-ID",
}

// StripIdentityHeaders removes all identity headers from incoming requests.
// This MUST be the outermost middleware in the gateway stack so that no
// downstream handler ever sees client-supplied identity values.
func StripIdentityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for header := range r.Header {
			if strings.HasPrefix(strings.ToLower(header), "x-user-") || strings.EqualFold(header, "X-Department-ID") {
				r.Header.Del(header)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// GatewayAuth verifies JWT and attaches X-User-* headers before proxying.
// All identity headers are set unconditionally (including empty values) to
// ensure downstream services never see stale or spoofed values.
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

			// Attach identity headers unconditionally — always overwrite,
			// even with empty string, to prevent any residual spoofed values.
			r.Header.Set("X-User-ID", claims.UserID)
			r.Header.Set("X-User-Email", claims.Email)
			r.Header.Set("X-User-Role", claims.Role)
			r.Header.Set("X-User-Department", claims.DepartmentID)
			r.Header.Set("X-User-Name", claims.FullName)

			next.ServeHTTP(w, r)
		})
	}
}
