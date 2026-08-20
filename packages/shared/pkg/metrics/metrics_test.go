package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsRegistry_RecordAndSnapshot(t *testing.T) {
	reg := InitRegistry("test-service")

	// Record sample requests
	reg.RecordRequest("GET", "/api/v1/tickets", 200, 25*time.Millisecond)
	reg.RecordRequest("GET", "/api/v1/tickets", 200, 35*time.Millisecond)
	reg.RecordRequest("POST", "/api/v1/tickets", 201, 50*time.Millisecond)
	reg.RecordRequest("GET", "/api/v1/tickets/999", 404, 10*time.Millisecond)

	snapshot := reg.GetSnapshot()

	if snapshot.TotalRequests != 4 {
		t.Errorf("expected total requests = 4, got %d", snapshot.TotalRequests)
	}
	if snapshot.ErrorCount != 1 {
		t.Errorf("expected error count = 1, got %d", snapshot.ErrorCount)
	}
	if snapshot.ErrorRatePct != 25.0 {
		t.Errorf("expected error rate = 25%%, got %.2f%%", snapshot.ErrorRatePct)
	}
	if snapshot.AvgDurationMs <= 0 {
		t.Errorf("expected positive avg duration ms, got %.2f", snapshot.AvgDurationMs)
	}
}

// Test Case 8.1: Verify Prometheus Exposition Format.
func TestPrometheusHandler_TestCase_8_1(t *testing.T) {
	reg := InitRegistry("eomp-gateway")
	reg.RecordRequest("GET", "/api/v1/health", 200, 5*time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := PrometheusHandler()
	handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("expected text/plain content-type, got %s", contentType)
	}

	body := rec.Body.String()

	// Verify required Prometheus metrics per Test Case 8.1
	requiredMetrics := []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"service_uptime_seconds",
		"service_memory_bytes",
		"service_goroutines_count",
	}

	for _, m := range requiredMetrics {
		if !strings.Contains(body, m) {
			t.Errorf("Prometheus output missing expected metric: %s", m)
		}
	}
}
