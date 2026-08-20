package middleware

import (
	"net/http"
	"strings"

	"eomp/packages/shared/pkg/errors"
)

// RequireRoles verifies that the authenticated user has at least one of the required roles.
// Returns HTTP 403 Forbidden with code INSUFFICIENT_PERMISSIONS if unauthorized (Test Case 10.1).
func RequireRoles(allowedRoles ...string) func(http.Handler) http.Handler {
	allowedMap := make(map[string]bool, len(allowedRoles))
	for _, r := range allowedRoles {
		allowedMap[strings.ToUpper(r)] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check role from context first, then from header
			role := GetUserRole(r.Context())
			if role == "" {
				role = r.Header.Get("X-User-Role")
			}

			if role == "" {
				errors.WriteHTTP(w, errors.Unauthorized("authentication required before authorization"))
				return
			}

			// ROLE_ADMIN has universal access
			if strings.ToUpper(role) == "ROLE_ADMIN" {
				next.ServeHTTP(w, r)
				return
			}

			if !allowedMap[strings.ToUpper(role)] {
				errors.WriteHTTP(w, errors.Forbidden("access denied: insufficient permissions for this resource"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
