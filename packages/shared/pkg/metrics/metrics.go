package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	startTime = time.Now()
)

// MetricRegistry manages in-memory Prometheus metrics.
type MetricRegistry struct {
	mu           sync.RWMutex
	serviceName  string
	requestCount map[string]*int64   // key: method_path_status
	requestTime  map[string]*float64 // key: method_path
	activeReqs   int64
}

var (
	defaultRegistry *MetricRegistry
	registryOnce    sync.Once
)

// InitRegistry initializes the global metric registry for a service.
func InitRegistry(serviceName string) *MetricRegistry {
	registryOnce.Do(func() {
		defaultRegistry = &MetricRegistry{
			serviceName:  serviceName,
			requestCount: make(map[string]*int64),
			requestTime:  make(map[string]*float64),
		}
	})
	return defaultRegistry
}

// GetRegistry returns the default registry.
func GetRegistry() *MetricRegistry {
	if defaultRegistry == nil {
		return InitRegistry("eomp-service")
	}
	return defaultRegistry
}

// RecordRequest records RED metrics for an incoming HTTP request.
func (r *MetricRegistry) RecordRequest(method, path string, status int, duration time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Normalize path to avoid high cardinality
	normalizedPath := normalizePath(path)

	// Increment request count counter
	countKey := fmt.Sprintf("%s_%s_%d", method, normalizedPath, status)
	if _, ok := r.requestCount[countKey]; !ok {
		var cnt int64
		r.requestCount[countKey] = &cnt
	}
	atomic.AddInt64(r.requestCount[countKey], 1)

	// Accumulate request duration
	timeKey := fmt.Sprintf("%s_%s", method, normalizedPath)
	if _, ok := r.requestTime[timeKey]; !ok {
		var dur float64
		r.requestTime[timeKey] = &dur
	}
	*r.requestTime[timeKey] += duration.Seconds()
}

func normalizePath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}
	// Strip trailing slash
	p := strings.TrimSuffix(path, "/")
	if strings.HasPrefix(p, "/api/v1/tickets/") && len(p) > len("/api/v1/tickets/") {
		return "/api/v1/tickets/{id}"
	}
	if strings.HasPrefix(p, "/api/v1/problems/") && len(p) > len("/api/v1/problems/") {
		return "/api/v1/problems/{id}"
	}
	if strings.HasPrefix(p, "/api/v1/changes/") && len(p) > len("/api/v1/changes/") {
		return "/api/v1/changes/{id}"
	}
	if strings.HasPrefix(p, "/api/v1/workflows/") && len(p) > len("/api/v1/workflows/") {
		return "/api/v1/workflows/{id}"
	}
	if strings.HasPrefix(p, "/api/v1/employees/") && len(p) > len("/api/v1/employees/") {
		return "/api/v1/employees/{id}"
	}
	if strings.HasPrefix(p, "/api/v1/assets/") && len(p) > len("/api/v1/assets/") {
		return "/api/v1/assets/{id}"
	}
	if strings.HasPrefix(p, "/api/v1/knowledge/") && len(p) > len("/api/v1/knowledge/") {
		return "/api/v1/knowledge/{id}"
	}
	return p
}

type statusTrackingWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusTrackingWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

// HTTPMetricsMiddleware intercepts HTTP requests to collect RED metrics.
func HTTPMetricsMiddleware(serviceName string) func(http.Handler) http.Handler {
	reg := InitRegistry(serviceName)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip internal metrics endpoint from polluting request counters
			if r.URL.Path == "/metrics" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			atomic.AddInt64(&reg.activeReqs, 1)
			defer atomic.AddInt64(&reg.activeReqs, -1)

			tw := &statusTrackingWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(tw, r)

			duration := time.Since(start)
			reg.RecordRequest(r.Method, r.URL.Path, tw.statusCode, duration)
		})
	}
}

// PrometheusHandler exports metrics in official Prometheus 2.0 text format.
func PrometheusHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reg := GetRegistry()
		reg.mu.RLock()
		defer reg.mu.RUnlock()

		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

		var b strings.Builder

		// Service runtime metadata
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		uptime := time.Since(startTime).Seconds()
		goroutines := runtime.NumGoroutine()

		b.WriteString("# HELP service_uptime_seconds Total seconds since service started\n")
		b.WriteString("# TYPE service_uptime_seconds gauge\n")
		b.WriteString(fmt.Sprintf("service_uptime_seconds{service=\"%s\"} %.2f\n\n", reg.serviceName, uptime))

		b.WriteString("# HELP service_goroutines_count Current number of active Go goroutines\n")
		b.WriteString("# TYPE service_goroutines_count gauge\n")
		b.WriteString(fmt.Sprintf("service_goroutines_count{service=\"%s\"} %d\n\n", reg.serviceName, goroutines))

		b.WriteString("# HELP service_memory_bytes Memory allocated by the service in bytes\n")
		b.WriteString("# TYPE service_memory_bytes gauge\n")
		b.WriteString(fmt.Sprintf("service_memory_bytes{service=\"%s\",type=\"alloc\"} %d\n", reg.serviceName, mem.Alloc))
		b.WriteString(fmt.Sprintf("service_memory_bytes{service=\"%s\",type=\"sys\"} %d\n\n", reg.serviceName, mem.Sys))

		b.WriteString("# HELP http_requests_total Total number of HTTP requests processed\n")
		b.WriteString("# TYPE http_requests_total counter\n")
		if len(reg.requestCount) == 0 {
			// Emit at least one base counter for initial scraping
			b.WriteString(fmt.Sprintf("http_requests_total{service=\"%s\",method=\"GET\",path=\"/health\",status=\"200\"} 1\n", reg.serviceName))
		} else {
			for key, countPtr := range reg.requestCount {
				parts := strings.Split(key, "_")
				if len(parts) >= 3 {
					method := parts[0]
					status := parts[len(parts)-1]
					path := strings.Join(parts[1:len(parts)-1], "_")
					b.WriteString(fmt.Sprintf("http_requests_total{service=\"%s\",method=\"%s\",path=\"%s\",status=\"%s\"} %d\n",
						reg.serviceName, method, path, status, atomic.LoadInt64(countPtr)))
				}
			}
		}
		b.WriteString("\n")

		b.WriteString("# HELP http_request_duration_seconds Total seconds spent processing HTTP requests\n")
		b.WriteString("# TYPE http_request_duration_seconds summary\n")
		if len(reg.requestTime) == 0 {
			b.WriteString(fmt.Sprintf("http_request_duration_seconds_sum{service=\"%s\",method=\"GET\",path=\"/health\"} 0.001\n", reg.serviceName))
			b.WriteString(fmt.Sprintf("http_request_duration_seconds_count{service=\"%s\",method=\"GET\",path=\"/health\"} 1\n", reg.serviceName))
		} else {
			for key, durPtr := range reg.requestTime {
				parts := strings.Split(key, "_")
				if len(parts) >= 2 {
					method := parts[0]
					path := strings.Join(parts[1:], "_")
					b.WriteString(fmt.Sprintf("http_request_duration_seconds_sum{service=\"%s\",method=\"%s\",path=\"%s\"} %.6f\n",
						reg.serviceName, method, path, *durPtr))
				}
			}
		}
		b.WriteString("\n")

		w.Write([]byte(b.String()))
	}
}

// Snapshot returns summary metrics for dashboard aggregation.
type Snapshot struct {
	ServiceName    string  `json:"service_name"`
	UptimeSeconds  float64 `json:"uptime_seconds"`
	MemoryMB       float64 `json:"memory_mb"`
	Goroutines     int     `json:"goroutines"`
	TotalRequests  int64   `json:"total_requests"`
	ErrorCount     int64   `json:"error_count"`
	ErrorRatePct   float64 `json:"error_rate_pct"`
	AvgDurationMs  float64 `json:"avg_duration_ms"`
	ActiveRequests int64   `json:"active_requests"`
}

// GetSnapshot calculates the aggregated RED metrics snapshot.
func (r *MetricRegistry) GetSnapshot() Snapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	var totalReqs int64
	var errorReqs int64
	var totalDur float64

	for key, countPtr := range r.requestCount {
		cnt := atomic.LoadInt64(countPtr)
		totalReqs += cnt
		parts := strings.Split(key, "_")
		if len(parts) >= 3 {
			statusStr := parts[len(parts)-1]
			if code, err := strconv.Atoi(statusStr); err == nil && code >= 400 {
				errorReqs += cnt
			}
		}
	}

	for _, durPtr := range r.requestTime {
		totalDur += *durPtr
	}

	errRate := 0.0
	if totalReqs > 0 {
		errRate = (float64(errorReqs) / float64(totalReqs)) * 100.0
	}

	avgMs := 0.0
	if totalReqs > 0 {
		avgMs = (totalDur / float64(totalReqs)) * 1000.0
	}

	return Snapshot{
		ServiceName:    r.serviceName,
		UptimeSeconds:  time.Since(startTime).Seconds(),
		MemoryMB:       float64(mem.Alloc) / 1024 / 1024,
		Goroutines:     runtime.NumGoroutine(),
		TotalRequests:  totalReqs,
		ErrorCount:     errorReqs,
		ErrorRatePct:   errRate,
		AvgDurationMs:  avgMs,
		ActiveRequests: atomic.LoadInt64(&r.activeReqs),
	}
}
