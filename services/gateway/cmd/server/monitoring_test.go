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

func TestObservabilityPrometheusExposition(t *testing.T) {
	reg := metrics.InitRegistry("eomp-gateway")
	reg.RecordRequest("GET", "/api/v1/tickets", http.StatusOK, 15*time.Millisecond)
	reg.RecordRequest("POST", "/api/v1/changes", http.StatusCreated, 35*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	metrics.PrometheusHandler()(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}
	if contentType := rec.Header().Get("Content-Type"); !strings.Contains(contentType, "text/plain") {
		t.Errorf("expected text/plain content type, got: %s", contentType)
	}
	body := rec.Body.String()
	for _, metricName := range []string{
		"http_requests_total",
		"http_request_duration_seconds",
		"service_uptime_seconds",
		"service_memory_bytes",
	} {
		if !strings.Contains(body, metricName) {
			t.Errorf("missing %s metric", metricName)
		}
	}
}

func TestMonitoringUsesActiveProbesWithoutSyntheticMetrics(t *testing.T) {
	healthy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","version":"0.1.0"}`))
	}))
	defer healthy.Close()

	offline := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))
	offlineURL := offline.URL
	offline.Close()

	urls := make(map[string]string)
	for _, id := range []string{"gateway", "auth", "employee", "asset", "helpdesk", "workflow", "notification", "knowledge", "ai", "audit", "reporting"} {
		urls[id] = healthy.URL
	}
	urls["ai"] = offlineURL

	monitoring := handler.NewMonitoringHandlerWithURLs(urls)
	startedAt := time.Now()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/monitoring/overview", nil)
	rec := httptest.NewRecorder()
	monitoring.GetOverview(rec, req)
	if elapsed := time.Since(startedAt); elapsed >= 5*time.Second {
		t.Fatalf("outage probe took %v; expected less than 5 seconds", elapsed)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected overview 200, got %d", rec.Code)
	}

	var overview handler.ClusterOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &overview); err != nil {
		t.Fatalf("decode overview: %v", err)
	}
	if overview.TotalServices != 11 || overview.OnlineServices != 10 || overview.OfflineServices != 1 {
		t.Fatalf("unexpected probe summary: %+v", overview)
	}
	if overview.MetricsAvailable {
		t.Fatal("RED/resource metrics must not be marked available without a metrics backend")
	}
	if overview.DataSource != "active_health_probes" {
		t.Fatalf("unexpected data source %q", overview.DataSource)
	}

	servicesReq := httptest.NewRequest(http.MethodGet, "/api/v1/monitoring/services", nil)
	servicesRec := httptest.NewRecorder()
	monitoring.ListServices(servicesRec, servicesReq)
	var services []handler.ServiceHealthStatus
	if err := json.Unmarshal(servicesRec.Body.Bytes(), &services); err != nil {
		t.Fatalf("decode service probes: %v", err)
	}
	for _, service := range services {
		if service.MetricsAvailable {
			t.Fatalf("service %s unexpectedly reported synthetic metrics", service.ID)
		}
	}

	logsReq := httptest.NewRequest(http.MethodGet, "/api/v1/monitoring/logs", nil)
	logsRec := httptest.NewRecorder()
	monitoring.GetLogs(logsRec, logsReq)
	if logsRec.Code != http.StatusNotImplemented {
		t.Fatalf("expected logs to be explicitly unavailable, got %d", logsRec.Code)
	}
}
