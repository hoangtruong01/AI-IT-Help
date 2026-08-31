package errors_test

import (
	"bytes"
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	appErrors "eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
)

func TestWriteHTTPDoesNotExposeUnexpectedError(t *testing.T) {
	rec := httptest.NewRecorder()
	rec.Header().Set("X-Request-ID", "request-123")
	appErrors.WriteHTTP(rec, stderrors.New("pq: password authentication failed for db.internal"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "pq:") || strings.Contains(rec.Body.String(), "db.internal") {
		t.Fatalf("response exposed internal error details: %s", rec.Body.String())
	}

	var payload struct {
		Error struct {
			Code      string `json:"code"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if payload.Error.Code != appErrors.CodeInternalServerError || payload.Error.RequestID != "request-123" {
		t.Fatalf("unexpected error payload: %+v", payload.Error)
	}
}

func TestWriteHTTPSanitizesExplicitInternalAppError(t *testing.T) {
	rec := httptest.NewRecorder()
	appErrors.WriteHTTP(rec, appErrors.InternalServerError("relation secret_table does not exist"))

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "secret_table") || !strings.Contains(rec.Body.String(), "an internal error occurred") {
		t.Fatalf("expected a generic 500 response, got %s", rec.Body.String())
	}
}

func TestInternalErrorUsesSameRequestIDForLogAndResponse(t *testing.T) {
	var logOutput bytes.Buffer
	previousLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&logOutput, nil)))
	t.Cleanup(func() { slog.SetDefault(previousLogger) })

	handler := middleware.RequestLogger(slog.Default())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := stderrors.New("pq: relation secret_table does not exist")
		appErrors.WriteHTTP(w, appErrors.Internal(r.Context(), "load ticket", raw))
	}))
	req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets/1", nil)
	req.Header.Set("X-Request-ID", "correlation-456")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !strings.Contains(logOutput.String(), "correlation-456") || !strings.Contains(logOutput.String(), "secret_table") {
		t.Fatalf("server log must retain the request ID and original error: %s", logOutput.String())
	}
	if !strings.Contains(rec.Body.String(), "correlation-456") {
		t.Fatalf("response must include the same request ID: %s", rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "secret_table") || strings.Contains(rec.Body.String(), "pq:") {
		t.Fatalf("response exposed internal error details: %s", rec.Body.String())
	}
}
