package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/response"
	"eomp/services/reporting/internal/model"
	"eomp/services/reporting/internal/service"
)

// ReportingHandler provides HTTP REST endpoints for BI reporting.
type ReportingHandler struct {
	svc service.Service
}

func dateFilterFromRequest(r *http.Request) (model.DateFilterQuery, error) {
	rangeVal := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("range")))
	if rangeVal == "" {
		rangeVal = "30d"
	}
	filter := model.DateFilterQuery{Range: rangeVal}
	if rangeVal != "custom" {
		switch rangeVal {
		case "today", "7d", "30d", "quarter":
			return filter, nil
		default:
			return filter, errors.BadRequest("range must be one of today, 7d, 30d, quarter, or custom")
		}
	}
	start, err := time.Parse("2006-01-02", r.URL.Query().Get("start_date"))
	if err != nil {
		return filter, errors.BadRequest("custom range requires start_date in YYYY-MM-DD format")
	}
	end, err := time.Parse("2006-01-02", r.URL.Query().Get("end_date"))
	if err != nil {
		return filter, errors.BadRequest("custom range requires end_date in YYYY-MM-DD format")
	}
	end = end.Add(24*time.Hour - time.Nanosecond)
	if start.After(end) {
		return filter, errors.BadRequest("start_date must not be after end_date")
	}
	filter.StartDate, filter.EndDate = &start, &end
	return filter, nil
}

func writeReportingError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	if _, ok := err.(*errors.AppError); ok {
		errors.WriteHTTP(w, err)
		return
	}
	errors.WriteHTTP(w, errors.Internal(r.Context(), operation, err))
}

// NewReportingHandler creates a new ReportingHandler.
func NewReportingHandler(svc service.Service) *ReportingHandler {
	return &ReportingHandler{svc: svc}
}

// GetOverview handles GET /api/v1/reports/overview
func (h *ReportingHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	filter, err := dateFilterFromRequest(r)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	overview, err := h.svc.GetOverview(r.Context(), filter)
	if err != nil {
		writeReportingError(w, r, "reporting overview", err)
		return
	}

	response.JSON(w, http.StatusOK, overview)
}

// GetTrends handles GET /api/v1/reports/trends
func (h *ReportingHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	filter, err := dateFilterFromRequest(r)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	trends, err := h.svc.GetTrends(r.Context(), filter)
	if err != nil {
		writeReportingError(w, r, "reporting trends", err)
		return
	}

	response.JSON(w, http.StatusOK, trends)
}

// GetCategories handles GET /api/v1/reports/categories
func (h *ReportingHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	filter, err := dateFilterFromRequest(r)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	categories, err := h.svc.GetCategories(r.Context(), filter)
	if err != nil {
		writeReportingError(w, r, "reporting categories", err)
		return
	}

	response.JSON(w, http.StatusOK, categories)
}

// GetDepartmentsSLA handles GET /api/v1/reports/departments-sla
func (h *ReportingHandler) GetDepartmentsSLA(w http.ResponseWriter, r *http.Request) {
	filter, err := dateFilterFromRequest(r)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	depts, err := h.svc.GetDepartmentsSLA(r.Context(), filter)
	if err != nil {
		writeReportingError(w, r, "reporting department SLA", err)
		return
	}

	response.JSON(w, http.StatusOK, depts)
}

// GetAgents handles GET /api/v1/reports/agents
func (h *ReportingHandler) GetAgents(w http.ResponseWriter, r *http.Request) {
	filter, err := dateFilterFromRequest(r)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	agents, err := h.svc.GetAgents(r.Context(), filter)
	if err != nil {
		writeReportingError(w, r, "reporting agents", err)
		return
	}

	response.JSON(w, http.StatusOK, agents)
}

// ExportReport handles POST /api/v1/reports/export (Test Case 9.1)
func (h *ReportingHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	var req model.ExportReportRequest

	if strings.Contains(r.Header.Get("Content-Type"), "application/json") && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
			return
		}
	}

	if req.Format == "" {
		req.Format = r.URL.Query().Get("format")
	}
	if req.Title == "" {
		req.Title = "Enterprise IT Operations BI Performance Summary"
	}
	if req.Range == "" {
		req.Range = r.URL.Query().Get("range")
	}
	if req.Range == "" {
		req.Range = "30d"
	}
	switch strings.ToLower(req.Format) {
	case "pdf", "csv", "excel", "xlsx":
	default:
		errors.WriteHTTP(w, errors.BadRequest("format must be pdf or csv"))
		return
	}
	if req.Range == "custom" && (req.StartDate == nil || req.EndDate == nil || req.StartDate.After(*req.EndDate)) {
		errors.WriteHTTP(w, errors.BadRequest("custom export range requires valid start_date and end_date"))
		return
	}

	doc, err := h.svc.ExportReport(r.Context(), req)
	if err != nil {
		writeReportingError(w, r, "reporting export", err)
		return
	}

	response.JSON(w, http.StatusOK, doc)
}
