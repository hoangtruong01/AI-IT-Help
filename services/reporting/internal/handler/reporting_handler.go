package handler

import (
	"encoding/json"
	"net/http"

	"eomp/packages/shared/pkg/response"
	"eomp/services/reporting/internal/model"
	"eomp/services/reporting/internal/service"
)

// ReportingHandler provides HTTP REST endpoints for BI reporting.
type ReportingHandler struct {
	svc service.Service
}

// NewReportingHandler creates a new ReportingHandler.
func NewReportingHandler(svc service.Service) *ReportingHandler {
	return &ReportingHandler{svc: svc}
}

// GetOverview handles GET /api/v1/reports/overview
func (h *ReportingHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	rangeVal := r.URL.Query().Get("range")
	if rangeVal == "" {
		rangeVal = "30d"
	}

	filter := model.DateFilterQuery{Range: rangeVal}
	overview, err := h.svc.GetOverview(r.Context(), filter)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, overview)
}

// GetTrends handles GET /api/v1/reports/trends
func (h *ReportingHandler) GetTrends(w http.ResponseWriter, r *http.Request) {
	rangeVal := r.URL.Query().Get("range")
	if rangeVal == "" {
		rangeVal = "30d"
	}

	filter := model.DateFilterQuery{Range: rangeVal}
	trends, err := h.svc.GetTrends(r.Context(), filter)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, trends)
}

// GetCategories handles GET /api/v1/reports/categories
func (h *ReportingHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	rangeVal := r.URL.Query().Get("range")
	filter := model.DateFilterQuery{Range: rangeVal}

	categories, err := h.svc.GetCategories(r.Context(), filter)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, categories)
}

// GetDepartmentsSLA handles GET /api/v1/reports/departments-sla
func (h *ReportingHandler) GetDepartmentsSLA(w http.ResponseWriter, r *http.Request) {
	rangeVal := r.URL.Query().Get("range")
	filter := model.DateFilterQuery{Range: rangeVal}

	depts, err := h.svc.GetDepartmentsSLA(r.Context(), filter)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, depts)
}

// GetAgents handles GET /api/v1/reports/agents
func (h *ReportingHandler) GetAgents(w http.ResponseWriter, r *http.Request) {
	rangeVal := r.URL.Query().Get("range")
	filter := model.DateFilterQuery{Range: rangeVal}

	agents, err := h.svc.GetAgents(r.Context(), filter)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, agents)
}

// ExportReport handles POST /api/v1/reports/export (Test Case 9.1)
func (h *ReportingHandler) ExportReport(w http.ResponseWriter, r *http.Request) {
	var req model.ExportReportRequest

	if r.Header.Get("Content-Type") == "application/json" && r.ContentLength > 0 {
		_ = json.NewDecoder(r.Body).Decode(&req)
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

	doc, err := h.svc.ExportReport(r.Context(), req)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, doc)
}
