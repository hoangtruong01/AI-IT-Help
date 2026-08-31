package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"eomp/packages/shared/pkg/requestctx"
)

// Common domain error codes
const (
	CodeBadRequest          = "BAD_REQUEST"
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeForbidden           = "FORBIDDEN"
	CodeNotFound            = "NOT_FOUND"
	CodeConflict            = "CONFLICT"
	CodePayloadTooLarge     = "PAYLOAD_TOO_LARGE"
	CodeInternalServerError = "INTERNAL_SERVER_ERROR"
	CodeUnprocessableEntity = "UNPROCESSABLE_ENTITY"
)

// AppError represents a structured, enterprise-grade domain error
type AppError struct {
	StatusCode int    `json:"status_code"`
	Code       string `json:"code"`
	Message    string `json:"message"`
	Details    any    `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s (status: %d)", e.Code, e.Message, e.StatusCode)
}

// New creates a generic AppError
func New(statusCode int, code, message string) *AppError {
	return &AppError{
		StatusCode: statusCode,
		Code:       code,
		Message:    message,
	}
}

// WithDetails attaches extra metadata/validation errors
func (e *AppError) WithDetails(details any) *AppError {
	return &AppError{
		StatusCode: e.StatusCode,
		Code:       e.Code,
		Message:    e.Message,
		Details:    details,
	}
}

// BadRequest helper
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, message)
}

// Unauthorized helper
func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

// Forbidden helper
func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message)
}

// NotFound helper
func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message)
}

// Conflict helper
func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message)
}

// PayloadTooLarge helper
func PayloadTooLarge(message string) *AppError {
	return New(http.StatusRequestEntityTooLarge, CodePayloadTooLarge, message)
}

// InternalServerError helper — use ONLY for safe, generic messages visible to clients.
// Do NOT pass raw error strings from database/driver/runtime.
func InternalServerError(message string) *AppError {
	return New(http.StatusInternalServerError, CodeInternalServerError, message)
}

// Internal records the original server-side error with its correlation ID and
// returns a generic error that is safe to expose to clients.
func Internal(ctx context.Context, operation string, cause error) *AppError {
	attrs := []any{
		slog.String("operation", operation),
		slog.Any("error", cause),
	}
	if requestID := requestctx.RequestID(ctx); requestID != "" {
		attrs = append(attrs, slog.String("request_id", requestID))
	}
	slog.ErrorContext(ctx, "internal service error", attrs...)
	return InternalServerError("an internal error occurred")
}

// WriteHTTP sends structured JSON error response to http.ResponseWriter.
// For 5xx errors, includes the X-Request-ID as request_id in the response
// so that clients can reference it when contacting support.
func WriteHTTP(w http.ResponseWriter, err error) {
	appErr, ok := err.(*AppError)
	if !ok {
		slog.Error("unhandled HTTP error",
			slog.String("request_id", w.Header().Get("X-Request-ID")),
			slog.Any("error", err),
		)
		appErr = InternalServerError("an unexpected error occurred")
	}
	if appErr.StatusCode >= 500 {
		appErr = InternalServerError("an internal error occurred")
	}

	// Build response payload
	errPayload := map[string]any{
		"code":        appErr.Code,
		"message":     appErr.Message,
		"status_code": appErr.StatusCode,
	}
	if appErr.StatusCode < 500 && appErr.Details != nil {
		errPayload["details"] = appErr.Details
	}

	// Include request_id for 5xx errors to help with log correlation
	if appErr.StatusCode >= 500 {
		if reqID := w.Header().Get("X-Request-ID"); reqID != "" {
			errPayload["request_id"] = reqID
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": errPayload,
	})
}
