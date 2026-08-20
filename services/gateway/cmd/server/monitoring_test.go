package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"eomp/packages/shared/pkg/metrics"
	"eomp/services/gateway/internal/handler"
)

// Test Case 8.1: Verify Prometheus Exposition Format across microservices.
func TestObservability_TestCase_8_1_PrometheusExposition(t *testing.T) {
	// 1. Record metrics on gateway
	reg := metrics.InitRegistry("eomp-gateway")
	reg.RecordRequest("GET", "/api/v1/tickets", 200, 15*time.Millisecond)
	reg.RecordRequest("POST", "/api/v1/changes", 201, 35*time.Millisecond)

	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()

	promHandler := metrics.PrometheusHandler()
	promHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("expected text/plain content type, got: %s", contentType)
	}

	body := rec.Body.String()

	// Assert mandatory Prometheus RED and runtime metrics
	if !strings.Contains(body, "http_requests_total") {
		t.Errorf("missing http_requests_total metric")
	}
	if !strings.Contains(body, "http_request_duration_seconds") {
		t.Errorf("missing http_request_duration_seconds metric")
	}
	if !strings.Contains(body, "service_uptime_seconds") {
		t.Errorf("missing service_uptime_seconds metric")
	}
	if !strings.Contains(body, "service_memory_bytes") {
		t.Errorf("missing service_memory_bytes metric")
	}
}

// Test Case 8.2: Verify Outage Detection in Cluster Monitoring in < 5 seconds.
func TestObservability_TestCase_8_2_OutageDetection(t *testing.T) {
	monHandler := handler.NewMonitoringHandler()

	// 1. Check initial cluster overview -> All 11 services operational
	req := httptest.NewRequest("GET", "/api/v1/monitoring/overview", nil)
	rec := httptest.NewRecorder()
	monHandler.GetOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for overview, got %d", rec.Code)
	}

	var overview handler.ClusterOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("failed to decode overview JSON: %v", err)
	}
	if overview.TotalServices < 11 {
		t.Errorf("expected at least 11 total services, got %d", overview.TotalServices)
	}
	if overview.OnlineServices < 11 {
		t.Errorf("expected at least 11 online services initially, got %d", overview.OnlineServices)
	}

	// 2. Simulate AI Service Disruption / Outage
	startTime := time.Now()
	monHandler.SetServiceStatusForTest("ai", "OFFLINE")

	// 3. Query updated overview
	rec2 := httptest.NewRecorder()
	monHandler.GetOverview(rec2, req)

	elapsed := time.Since(startTime)
	if elapsed > 5*time.Second {
		t.Errorf("outage detection took too long: %v (expected < 5s)", elapsed)
	}

	var updatedOverview handler.ClusterOverview
	if err := json.Unmarshal(rec2.Body.Bytes(), &updatedOverview); err != nil {
		t.Fatalf("failed to decode updated overview: %v", err)
	}

	if updatedOverview.OfflineServices != 1 {
		t.Errorf("expected 1 offline service, got %d", updatedOverview.OfflineServices)
	}
	if updatedOverview.OnlineServices != overview.TotalServices-1 {
		t.Errorf("expected %d online services, got %d", overview.TotalServices-1, updatedOverview.OnlineServices)
	}
	if updatedOverview.ClusterHealthPct >= 100.0 {
		t.Errorf("expected cluster health < 100%%, got %.2f%%", updatedOverview.ClusterHealthPct)
	}

	// 4. Test Live Log Streamer
	logReq := httptest.NewRequest("GET", "/api/v1/monitoring/logs?limit=10", nil)
	logRec := httptest.NewRecorder()
	monHandler.GetLogs(logRec, logReq)

	if logRec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for logs, got %d", logRec.Code)
	}

	var logs []handler.LogEntry
	if err := json.Unmarshal(logRec.Body.Bytes(), &logs); err != nil {
		t.Fatalf("failed to decode logs JSON: %v", err)
	}
	if len(logs) == 0 {
		t.Errorf("expected non-empty log entries, got 0")
	}
}
