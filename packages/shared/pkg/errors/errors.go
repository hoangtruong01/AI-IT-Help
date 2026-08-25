package errors

import (
	"encoding/json"
	"fmt"
	"net/http"
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

// InternalServerError helper
func InternalServerError(message string) *AppError {
	return New(http.StatusInternalServerError, CodeInternalServerError, message)
}

// WriteHTTP sends structured JSON error response to http.ResponseWriter
func WriteHTTP(w http.ResponseWriter, err error) {
	appErr, ok := err.(*AppError)
	if !ok {
		appErr = InternalServerError(err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(appErr.StatusCode)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{
			"code":        appErr.Code,
			"message":     appErr.Message,
			"status_code": appErr.StatusCode,
			"details":     appErr.Details,
		},
	})
}
