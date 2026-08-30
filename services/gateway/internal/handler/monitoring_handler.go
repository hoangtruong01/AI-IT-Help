package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"eomp/packages/shared/pkg/response"
)

const probeCacheTTL = 2 * time.Second

// ServiceHealthStatus contains only data observed by an active health probe.
// Resource and RED metrics are intentionally omitted until a metrics backend is connected.
type ServiceHealthStatus struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Category         string    `json:"category"`
	Port             int       `json:"port"`
	Status           string    `json:"status"` // UNKNOWN, ONLINE, DEGRADED, OFFLINE
	LatencyMs        float64   `json:"latency_ms"`
	Version          string    `json:"version,omitempty"`
	LastProbeTime    time.Time `json:"last_probe_time,omitempty"`
	ProbeError       string    `json:"probe_error,omitempty"`
	MetricsAvailable bool      `json:"metrics_available"`
}

// ClusterOverview summarizes the latest active service probes.
type ClusterOverview struct {
	TotalServices      int       `json:"total_services"`
	OnlineServices     int       `json:"online_services"`
	DegradedServices   int       `json:"degraded_services"`
	OfflineServices    int       `json:"offline_services"`
	UnknownServices    int       `json:"unknown_services"`
	ClusterHealthPct   float64   `json:"cluster_health_pct"`
	AvgProbeLatencyMs  float64   `json:"avg_probe_latency_ms"`
	MetricsAvailable   bool      `json:"metrics_available"`
	DataSource         string    `json:"data_source"`
	LastMeasurementUTC time.Time `json:"last_measurement_utc"`
}

type serviceDefinition struct {
	ID       string
	Name     string
	Category string
	Port     int
}

var monitoredServices = []serviceDefinition{
	{ID: "gateway", Name: "API Gateway", Category: "Core Edge", Port: 8080},
	{ID: "auth", Name: "Auth & Identity Service", Category: "Security", Port: 8081},
	{ID: "employee", Name: "Employee & Org Service", Category: "Master Data", Port: 8082},
	{ID: "asset", Name: "Asset & CMDB Service", Category: "Inventory", Port: 8083},
	{ID: "helpdesk", Name: "IT Helpdesk & Problem Service", Category: "Operations", Port: 8084},
	{ID: "workflow", Name: "Workflow Engine & CAB", Category: "Automation", Port: 8085},
	{ID: "notification", Name: "Notification Service", Category: "Messaging", Port: 8086},
	{ID: "knowledge", Name: "Knowledge Base & SOPs", Category: "Intelligence", Port: 8087},
	{ID: "ai", Name: "AI Copilot & Triage Engine", Category: "Intelligence", Port: 8088},
	{ID: "audit", Name: "Audit & Compliance Service", Category: "Governance", Port: 8089},
	{ID: "reporting", Name: "Reporting & BI Analytics", Category: "Analytics", Port: 8090},
}

// MonitoringHandler actively probes configured service health endpoints.
type MonitoringHandler struct {
	mu          sync.RWMutex
	probeMu     sync.Mutex
	services    []ServiceHealthStatus
	serviceURLs map[string]string
	httpClient  *http.Client
	lastRefresh time.Time
}

// NewMonitoringHandler creates a local-development probe set.
func NewMonitoringHandler() *MonitoringHandler {
	return NewMonitoringHandlerWithURLs(nil)
}

// NewMonitoringHandlerWithURLs creates a probe set using configured service base URLs.
func NewMonitoringHandlerWithURLs(baseURLs map[string]string) *MonitoringHandler {
	h := &MonitoringHandler{
		serviceURLs: make(map[string]string, len(monitoredServices)),
		httpClient: &http.Client{
			Timeout: 2 * time.Second,
		},
	}

	for _, definition := range monitoredServices {
		baseURL := fmt.Sprintf("http://localhost:%d", definition.Port)
		if configuredURL := strings.TrimSpace(baseURLs[definition.ID]); configuredURL != "" {
			baseURL = configuredURL
		}
		h.serviceURLs[definition.ID] = healthURL(baseURL)
		h.services = append(h.services, ServiceHealthStatus{
			ID:               definition.ID,
			Name:             definition.Name,
			Category:         definition.Category,
			Port:             definition.Port,
			Status:           "UNKNOWN",
			MetricsAvailable: false,
		})
	}

	return h
}

func healthURL(baseURL string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if strings.HasSuffix(trimmed, "/health") {
		return trimmed
	}
	return trimmed + "/health"
}

func (h *MonitoringHandler) probe(ctx context.Context, serviceID, targetURL string) ServiceHealthStatus {
	h.mu.RLock()
	var current ServiceHealthStatus
	for _, service := range h.services {
		if service.ID == serviceID {
			current = service
			break
		}
	}
	h.mu.RUnlock()

	startedAt := time.Now()
	current.LastProbeTime = startedAt.UTC()
	current.Version = ""
	current.ProbeError = ""
	current.MetricsAvailable = false

	probeCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, targetURL, nil)
	if err != nil {
		current.Status = "OFFLINE"
		current.ProbeError = "invalid health endpoint"
		return current
	}

	resp, err := h.httpClient.Do(req)
	current.LatencyMs = float64(time.Since(startedAt).Microseconds()) / 1000
	if err != nil {
		current.Status = "OFFLINE"
		current.ProbeError = err.Error()
		return current
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode >= 500:
		current.Status = "OFFLINE"
		current.ProbeError = fmt.Sprintf("health endpoint returned HTTP %d", resp.StatusCode)
	case resp.StatusCode >= 400 || current.LatencyMs > 1000:
		current.Status = "DEGRADED"
		current.ProbeError = fmt.Sprintf("health endpoint returned HTTP %d", resp.StatusCode)
	default:
		current.Status = "ONLINE"
	}

	var health struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&health); err == nil {
		current.Version = health.Version
	}
	return current
}

func (h *MonitoringHandler) refreshAll(ctx context.Context) {
	h.probeMu.Lock()
	defer h.probeMu.Unlock()

	h.mu.RLock()
	cacheFresh := !h.lastRefresh.IsZero() && time.Since(h.lastRefresh) < probeCacheTTL
	h.mu.RUnlock()
	if cacheFresh {
		return
	}

	results := make(chan ServiceHealthStatus, len(h.serviceURLs))
	var wg sync.WaitGroup
	for serviceID, targetURL := range h.serviceURLs {
		wg.Add(1)
		go func(id, url string) {
			defer wg.Done()
			results <- h.probe(ctx, id, url)
		}(serviceID, targetURL)
	}
	wg.Wait()
	close(results)

	h.mu.Lock()
	defer h.mu.Unlock()
	for result := range results {
		for i := range h.services {
			if h.services[i].ID == result.ID {
				h.services[i] = result
				break
			}
		}
	}
	h.lastRefresh = time.Now()
}

// GetOverview returns a summary derived from active health probes.
func (h *MonitoringHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	h.refreshAll(r.Context())

	h.mu.RLock()
	defer h.mu.RUnlock()

	var online, degraded, offline, unknown, measured int
	var totalLatency float64
	for _, service := range h.services {
		switch service.Status {
		case "ONLINE":
			online++
		case "DEGRADED":
			degraded++
		case "OFFLINE":
			offline++
		default:
			unknown++
		}
		if !service.LastProbeTime.IsZero() {
			totalLatency += service.LatencyMs
			measured++
		}
	}

	total := len(h.services)
	healthPct := 0.0
	if total > 0 {
		healthPct = float64(online) / float64(total) * 100
	}
	avgLatency := 0.0
	if measured > 0 {
		avgLatency = totalLatency / float64(measured)
	}

	response.JSON(w, http.StatusOK, ClusterOverview{
		TotalServices:      total,
		OnlineServices:     online,
		DegradedServices:   degraded,
		OfflineServices:    offline,
		UnknownServices:    unknown,
		ClusterHealthPct:   healthPct,
		AvgProbeLatencyMs:  avgLatency,
		MetricsAvailable:   false,
		DataSource:         "active_health_probes",
		LastMeasurementUTC: h.lastRefresh.UTC(),
	})
}

// ListServices returns the latest active health probe for every configured service.
func (h *MonitoringHandler) ListServices(w http.ResponseWriter, r *http.Request) {
	h.refreshAll(r.Context())
	h.mu.RLock()
	defer h.mu.RUnlock()
	response.JSON(w, http.StatusOK, h.services)
}

// ProbeService triggers an immediate active health check on one configured service.
func (h *MonitoringHandler) ProbeService(w http.ResponseWriter, r *http.Request) {
	serviceID := r.PathValue("id")
	targetURL, exists := h.serviceURLs[serviceID]
	if !exists {
		response.JSON(w, http.StatusNotFound, map[string]string{"error": "service not found"})
		return
	}

	result := h.probe(r.Context(), serviceID, targetURL)
	h.mu.Lock()
	for i := range h.services {
		if h.services[i].ID == result.ID {
			h.services[i] = result
			break
		}
	}
	h.lastRefresh = time.Now()
	h.mu.Unlock()
	response.JSON(w, http.StatusOK, result)
}

// GetLogs reports that no log backend is configured instead of fabricating log records.
func (h *MonitoringHandler) GetLogs(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusNotImplemented, map[string]string{
		"error": "live logs are unavailable until a Loki-compatible log query backend is configured",
	})
}
