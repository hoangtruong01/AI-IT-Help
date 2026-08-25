package middleware

import (
	"fmt"
	"net/http"

	"eomp/packages/shared/pkg/errors"
)

// MaxBodySize restricts the maximum allowed request body size to protect against DoS attacks.
func MaxBodySize(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ContentLength > maxBytes {
				errors.WriteHTTP(w, errors.PayloadTooLarge(fmt.Sprintf("request payload exceeds maximum limit of %d bytes", maxBytes)))
				return
			}

			// Wrap request body to enforce limit during streaming reads
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}
