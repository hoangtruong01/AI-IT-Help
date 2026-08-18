package middleware

import (
	"context"
	"net/http"
	"strings"

	"eomp/packages/shared/pkg/auth"
	"eomp/packages/shared/pkg/errors"
)

type contextKey string

const (
	UserIDKey       contextKey = "user_id"
	UserEmailKey    contextKey = "user_email"
	UserRoleKey     contextKey = "user_role"
	UserClaimsKey   contextKey = "user_claims"
	DepartmentIDKey contextKey = "department_id"
)

// Authenticate verifies Bearer token and attaches UserClaims to Context
func Authenticate(jwtManager *auth.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				errors.WriteHTTP(w, errors.Unauthorized("missing authorization header"))
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				errors.WriteHTTP(w, errors.Unauthorized("invalid authorization header format (expected Bearer <token>)"))
				return
			}

			claims, err := jwtManager.ValidateToken(parts[1])
			if err != nil {
				errors.WriteHTTP(w, errors.Unauthorized("invalid or expired token"))
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
			ctx = context.WithValue(ctx, DepartmentIDKey, claims.DepartmentID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExtractGatewayHeaders reads identity headers injected by API Gateway
func ExtractGatewayHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			if uid := r.Header.Get("X-User-ID"); uid != "" {
				ctx = context.WithValue(ctx, UserIDKey, uid)
			}
			if email := r.Header.Get("X-User-Email"); email != "" {
				ctx = context.WithValue(ctx, UserEmailKey, email)
			}
			if role := r.Header.Get("X-User-Role"); role != "" {
				ctx = context.WithValue(ctx, UserRoleKey, role)
			}
			if dept := r.Header.Get("X-User-Department"); dept != "" {
				ctx = context.WithValue(ctx, DepartmentIDKey, dept)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole checks if user role matches one of allowed roles
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role, _ := r.Context().Value(UserRoleKey).(string)
			if role == "" {
				errors.WriteHTTP(w, errors.Forbidden("access denied: missing user role"))
				return
			}

			allowed := false
			for _, r := range allowedRoles {
				if r == role || role == "ROLE_ADMIN" {
					allowed = true
					break
				}
			}

			if !allowed {
				errors.WriteHTTP(w, errors.Forbidden("insufficient permissions for this resource"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper context getters
func GetUserID(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDKey).(string); ok {
		return val
	}
	return ""
}

func GetUserRole(ctx context.Context) string {
	if val, ok := ctx.Value(UserRoleKey).(string); ok {
		return val
	}
	return ""
}

func GetUserEmail(ctx context.Context) string {
	if val, ok := ctx.Value(UserEmailKey).(string); ok {
		return val
	}
	return ""
}
