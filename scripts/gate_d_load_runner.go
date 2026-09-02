// Command gate_d_load_runner runs an authenticated, dependency-free local load
// check against the deployed Docker Gateway and writes objective JSON evidence.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"sync"
	"time"
)

type authResponse struct {
	AccessToken string `json:"access_token"`
}

type sample struct {
	Endpoint string
	Duration time.Duration
	Status   int
	Failed   bool
}

type endpointSummary struct {
	Requests int     `json:"requests"`
	Failures int     `json:"failures"`
	P95MS    float64 `json:"p95_ms"`
	P99MS    float64 `json:"p99_ms"`
	MaxMS    float64 `json:"max_ms"`
}

type evidence struct {
	SchemaVersion  int                        `json:"schema_version"`
	Result         string                     `json:"result"`
	Scope          string                     `json:"scope"`
	StartedAtUTC   string                     `json:"started_at_utc"`
	CompletedAtUTC string                     `json:"completed_at_utc"`
	SourceRevision string                     `json:"source_revision"`
	BaseURL        string                     `json:"base_url"`
	VirtualUsers   int                        `json:"virtual_users"`
	Duration       string                     `json:"duration"`
	Requests       int                        `json:"requests"`
	Failures       int                        `json:"failures"`
	FailureRate    float64                    `json:"failure_rate"`
	P95MS          float64                    `json:"p95_ms"`
	P99MS          float64                    `json:"p99_ms"`
	Thresholds     map[string]string          `json:"thresholds"`
	Endpoints      map[string]endpointSummary `json:"endpoints"`
}

func main() {
	baseURL := flag.String("base-url", "http://127.0.0.1:8080", "deployed Gateway URL")
	duration := flag.Duration("duration", 30*time.Second, "test duration")
	vus := flag.Int("vus", 100, "concurrent virtual users")
	output := flag.String("output", "docs/evidence/gate-d/local_load_summary.json", "JSON evidence path")
	flag.Parse()

	if *vus <= 0 || *duration <= 0 {
		fatalf("vus and duration must be positive")
	}
	employeeEmail := os.Getenv("LOADTEST_EMPLOYEE_EMAIL")
	employeePassword := os.Getenv("LOADTEST_EMPLOYEE_PASSWORD")
	managerEmail := os.Getenv("LOADTEST_MANAGER_EMAIL")
	managerPassword := os.Getenv("LOADTEST_MANAGER_PASSWORD")
	if employeeEmail == "" || employeePassword == "" || managerEmail == "" || managerPassword == "" {
		fatalf("LOADTEST_EMPLOYEE_* and LOADTEST_MANAGER_* credentials are required")
	}

	transport := &http.Transport{
		MaxIdleConns:        *vus * 2,
		MaxIdleConnsPerHost: *vus * 2,
		IdleConnTimeout:     30 * time.Second,
	}
	client := &http.Client{Transport: transport, Timeout: 5 * time.Second}
	defer transport.CloseIdleConnections()

	root := strings.TrimRight(*baseURL, "/")
	employeeToken := login(client, root, employeeEmail, employeePassword, "198.51.100.1")
	managerToken := login(client, root, managerEmail, managerPassword, "198.51.100.2")
	started := time.Now().UTC()
	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()

	var mu sync.Mutex
	samples := make([]sample, 0, *vus*int(duration.Seconds()))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for worker := 1; worker <= *vus; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			<-start
			clientIP := fmt.Sprintf("198.51.%d.%d", workerID/254, workerID%254+1)
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()
			iteration := 0
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					endpoint, path, token := requestTarget(workerID+iteration, employeeToken, managerToken)
					result := execute(client, root+path, endpoint, token, clientIP)
					mu.Lock()
					samples = append(samples, result)
					mu.Unlock()
					iteration++
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()
	completed := time.Now().UTC()

	summary := summarize(samples)
	summary.SchemaVersion = 1
	summary.Scope = "local Docker authenticated load evidence; not k6/TLS/staging acceptance"
	summary.StartedAtUTC = started.Format(time.RFC3339Nano)
	summary.CompletedAtUTC = completed.Format(time.RFC3339Nano)
	summary.SourceRevision = revision()
	summary.BaseURL = root
	summary.VirtualUsers = *vus
	summary.Duration = duration.String()
	summary.Thresholds = map[string]string{
		"p95": "< 200 ms", "p99": "< 500 ms", "failure_rate": "< 1%",
	}
	if summary.P95MS < 200 && summary.P99MS < 500 && summary.FailureRate < 0.01 {
		summary.Result = "PASS"
	} else {
		summary.Result = "FAIL"
	}

	if err := os.MkdirAll(directory(*output), 0o755); err != nil {
		fatalf("create evidence directory: %v", err)
	}
	data, err := json.MarshalIndent(summary, "", "  ")
	if err != nil {
		fatalf("marshal evidence: %v", err)
	}
	if err := os.WriteFile(*output, append(data, '\n'), 0o644); err != nil {
		fatalf("write evidence: %v", err)
	}
	fmt.Printf("Gate D local load %s: requests=%d failures=%d p95=%.2fms p99=%.2fms\n", summary.Result, summary.Requests, summary.Failures, summary.P95MS, summary.P99MS)
	if summary.Result != "PASS" {
		os.Exit(1)
	}
}

func login(client *http.Client, baseURL, email, password, clientIP string) string {
	payload, _ := json.Marshal(map[string]string{"email": email, "password": password})
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/auth/login", bytes.NewReader(payload))
	if err != nil {
		fatalf("build login request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", clientIP)
	resp, err := client.Do(req)
	if err != nil {
		fatalf("login request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		fatalf("login returned HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	var auth authResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil || auth.AccessToken == "" {
		fatalf("login returned no access token")
	}
	return auth.AccessToken
}

func requestTarget(index int, employeeToken, managerToken string) (string, string, string) {
	switch index % 3 {
	case 0:
		return "health", "/health", ""
	case 1:
		return "tickets", "/api/v1/tickets?page=1&page_size=20&status=OPEN", employeeToken
	default:
		return "reports_overview", "/api/v1/reports/overview?range=30d", managerToken
	}
}

func execute(client *http.Client, url, endpoint, token, clientIP string) sample {
	started := time.Now()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return sample{Endpoint: endpoint, Failed: true}
	}
	req.Header.Set("X-Forwarded-For", clientIP)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := client.Do(req)
	duration := time.Since(started)
	if err != nil {
		return sample{Endpoint: endpoint, Duration: duration, Failed: true}
	}
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
	_ = resp.Body.Close()
	return sample{Endpoint: endpoint, Duration: duration, Status: resp.StatusCode, Failed: resp.StatusCode != http.StatusOK}
}

func summarize(samples []sample) evidence {
	result := evidence{Requests: len(samples), Endpoints: make(map[string]endpointSummary)}
	allDurations := make([]time.Duration, 0, len(samples))
	groups := make(map[string][]sample)
	for _, item := range samples {
		allDurations = append(allDurations, item.Duration)
		groups[item.Endpoint] = append(groups[item.Endpoint], item)
		if item.Failed {
			result.Failures++
		}
	}
	if result.Requests > 0 {
		result.FailureRate = float64(result.Failures) / float64(result.Requests)
	}
	result.P95MS = percentileMS(allDurations, 0.95)
	result.P99MS = percentileMS(allDurations, 0.99)
	for endpoint, items := range groups {
		durations := make([]time.Duration, 0, len(items))
		failures := 0
		for _, item := range items {
			durations = append(durations, item.Duration)
			if item.Failed {
				failures++
			}
		}
		sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
		max := 0.0
		if len(durations) > 0 {
			max = float64(durations[len(durations)-1]) / float64(time.Millisecond)
		}
		result.Endpoints[endpoint] = endpointSummary{Requests: len(items), Failures: failures, P95MS: percentileMS(durations, 0.95), P99MS: percentileMS(durations, 0.99), MaxMS: max}
	}
	return result
}

func percentileMS(values []time.Duration, percentile float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	index := int(float64(len(values)-1) * percentile)
	return float64(values[index]) / float64(time.Millisecond)
}

func revision() string {
	output, err := exec.Command("git", "rev-parse", "HEAD").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

func directory(path string) string {
	index := strings.LastIndexAny(path, `/\\`)
	if index < 0 {
		return "."
	}
	return path[:index]
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
