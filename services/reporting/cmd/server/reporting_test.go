package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"eomp/packages/shared/pkg/eventbus"
	"eomp/services/reporting/internal/handler"
	"eomp/services/reporting/internal/model"
	"eomp/services/reporting/internal/service"
)

type fixtureReportingRepository struct{}

func (fixtureReportingRepository) ProjectTicketEvent(context.Context, eventbus.Event) error {
	return nil
}

func (fixtureReportingRepository) GetExecutiveOverview(context.Context, model.DateFilterQuery) (*model.ExecutiveOverview, error) {
	return &model.ExecutiveOverview{}, nil
}
func (fixtureReportingRepository) GetIncidentTrends(context.Context, model.DateFilterQuery) ([]model.IncidentTrend, error) {
	return []model.IncidentTrend{}, nil
}
func (fixtureReportingRepository) GetCategoryBreakdowns(context.Context, model.DateFilterQuery) ([]model.CategoryBreakdown, error) {
	return []model.CategoryBreakdown{}, nil
}
func (fixtureReportingRepository) GetDepartmentSLAMetrics(context.Context, model.DateFilterQuery) ([]model.DepartmentSLAMetric, error) {
	return []model.DepartmentSLAMetric{}, nil
}
func (fixtureReportingRepository) GetAgentScorecards(context.Context, model.DateFilterQuery) ([]model.AgentScorecard, error) {
	return []model.AgentScorecard{}, nil
}
func (fixtureReportingRepository) GetRawRecords(_ context.Context, _ model.DateFilterQuery, limit int) ([]model.RawIncidentRecord, error) {
	records := make([]model.RawIncidentRecord, limit)
	for i := range records {
		records[i] = model.RawIncidentRecord{TicketNumber: "TST-1", Title: "Test incident", Status: "RESOLVED", CreatedAt: time.Unix(0, 0)}
	}
	return records, nil
}

func setupTestApp() *handler.ReportingHandler {
	repo := fixtureReportingRepository{}
	svc := service.NewService(repo)
	return handler.NewReportingHandler(svc)
}

// Test Case 9.1: PDF & Excel Export Benchmark (Exporting up to 10,000 records in < 3s).
func TestReporting_TestCase_9_1_ExportPerformance(t *testing.T) {
	h := setupTestApp()

	// 1. Test PDF Export
	reqBody := `{"format":"pdf","title":"Executive Operations Report","range":"30d","limit_rows":10000}`
	req := httptest.NewRequest("POST", "/api/v1/reports/export", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	start := time.Now()
	h.ExportReport(rec, req)
	elapsed := time.Since(start)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	if elapsed > 3*time.Second {
		t.Errorf("PDF export took too long: %v (expected < 3s)", elapsed)
	}

	var res model.ExportReportResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode export response: %v", err)
	}

	if res.MimeType != "application/pdf" {
		t.Errorf("expected application/pdf, got: %s", res.MimeType)
	}
	if len(res.ContentBase64) == 0 {
		t.Errorf("expected non-empty base64 content")
	}
	pdf, err := base64.StdEncoding.DecodeString(res.ContentBase64)
	if err != nil {
		t.Fatalf("invalid PDF base64: %v", err)
	}
	marker := []byte("startxref\n")
	markerAt := bytes.LastIndex(pdf, marker)
	if markerAt < 0 {
		t.Fatal("PDF is missing startxref")
	}
	xrefText := strings.SplitN(string(pdf[markerAt+len(marker):]), "\n", 2)[0]
	xrefAt, err := strconv.Atoi(xrefText)
	if err != nil || xrefAt < 0 || xrefAt >= len(pdf) || !bytes.HasPrefix(pdf[xrefAt:], []byte("xref\n")) {
		t.Fatalf("PDF startxref does not point to the xref table: %q", xrefText)
	}
	if res.TotalRecords != 10000 {
		t.Errorf("expected 10000 records processed, got %d", res.TotalRecords)
	}

	// 2. Test CSV / Excel Export
	reqCSVBody := `{"format":"csv","title":"Raw Operations Data","range":"30d","limit_rows":10000}`
	reqCSV := httptest.NewRequest("POST", "/api/v1/reports/export", strings.NewReader(reqCSVBody))
	reqCSV.Header.Set("Content-Type", "application/json")
	recCSV := httptest.NewRecorder()

	startCSV := time.Now()
	h.ExportReport(recCSV, reqCSV)
	elapsedCSV := time.Since(startCSV)

	if recCSV.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for CSV, got %d", recCSV.Code)
	}
	if elapsedCSV > 3*time.Second {
		t.Errorf("CSV export took too long: %v (expected < 3s)", elapsedCSV)
	}

	var resCSV model.ExportReportResponse
	if err := json.Unmarshal(recCSV.Body.Bytes(), &resCSV); err != nil {
		t.Fatalf("failed to decode CSV export response: %v", err)
	}
	if resCSV.MimeType != "text/csv" {
		t.Errorf("expected text/csv, got: %s", resCSV.MimeType)
	}
}

// Test Case 9.2: Date filter with zero data returns clean empty state without NaN.
func TestReporting_TestCase_9_2_EmptyState_NoNaN(t *testing.T) {
	h := setupTestApp()

	// Query overview with empty range
	req := httptest.NewRequest("GET", "/api/v1/reports/overview?range=30d", nil)
	rec := httptest.NewRecorder()

	h.GetOverview(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", rec.Code)
	}

	var ov model.ExecutiveOverview
	if err := json.Unmarshal(rec.Body.Bytes(), &ov); err != nil {
		t.Fatalf("failed to decode overview: %v", err)
	}

	// Assert that values are valid 0.0 and NOT NaN or crashed
	if ov.TotalIncidents != 0 {
		t.Errorf("expected 0 total incidents in empty state, got %d", ov.TotalIncidents)
	}
	if ov.SLACompliancePct != 0.0 {
		t.Errorf("expected 0.0 SLA compliance in empty state, got %f", ov.SLACompliancePct)
	}
	if ov.AvgMTTRMinutes != 0.0 {
		t.Errorf("expected 0.0 MTTR in empty state, got %f", ov.AvgMTTRMinutes)
	}
}

// Test reporting sub-endpoints: Trends, Categories, Departments, Agents.
func TestReporting_Endpoints(t *testing.T) {
	h := setupTestApp()

	// Trends
	reqTrends := httptest.NewRequest("GET", "/api/v1/reports/trends", nil)
	recTrends := httptest.NewRecorder()
	h.GetTrends(recTrends, reqTrends)
	if recTrends.Code != http.StatusOK {
		t.Errorf("expected 200 OK for trends, got %d", recTrends.Code)
	}

	// Categories
	reqCat := httptest.NewRequest("GET", "/api/v1/reports/categories", nil)
	recCat := httptest.NewRecorder()
	h.GetCategories(recCat, reqCat)
	if recCat.Code != http.StatusOK {
		t.Errorf("expected 200 OK for categories, got %d", recCat.Code)
	}

	// Departments SLA
	reqDepts := httptest.NewRequest("GET", "/api/v1/reports/departments-sla", nil)
	recDepts := httptest.NewRecorder()
	h.GetDepartmentsSLA(recDepts, reqDepts)
	if recDepts.Code != http.StatusOK {
		t.Errorf("expected 200 OK for departments-sla, got %d", recDepts.Code)
	}

	// Agents Scorecard
	reqAgents := httptest.NewRequest("GET", "/api/v1/reports/agents", nil)
	recAgents := httptest.NewRecorder()
	h.GetAgents(recAgents, reqAgents)
	if recAgents.Code != http.StatusOK {
		t.Errorf("expected 200 OK for agents, got %d", recAgents.Code)
	}
}
