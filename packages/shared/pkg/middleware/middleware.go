package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"

	appErrors "eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/requestctx"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// RequestLogger logs incoming HTTP requests with latency, method, path, status, and Request-ID.
func RequestLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" {
				b := make([]byte, 8)
				_, _ = rand.Read(b)
				reqID = hex.EncodeToString(b)
			}
			w.Header().Set("X-Request-ID", reqID)
			r = r.WithContext(requestctx.WithRequestID(r.Context(), reqID))

			rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			logger.Info("http request",
				slog.String("request_id", reqID),
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", rw.statusCode),
				slog.Duration("latency", duration),
				slog.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}

// Recoverer recovers from panics and logs the error gracefully.
func Recoverer(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.Error("panic recovered",
						slog.String("request_id", w.Header().Get("X-Request-ID")),
						slog.Any("error", err),
					)
					appErrors.WriteHTTP(w, appErrors.InternalServerError("an internal error occurred"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
