package handler

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"strings"

	"eomp/packages/shared/pkg/response"
	"eomp/services/audit/internal/model"
	"eomp/services/audit/internal/service"
)

// AuditHandler provides HTTP REST endpoints for Audit Trail and Compliance.
type AuditHandler struct {
	svc service.Service
}

// NewAuditHandler creates a new AuditHandler.
func NewAuditHandler(svc service.Service) *AuditHandler {
	return &AuditHandler{svc: svc}
}

// ListAuditLogs handles GET /api/v1/audit/logs
func (h *AuditHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	filter := model.AuditFilterQuery{
		EventType: r.URL.Query().Get("event_type"),
		Status:    r.URL.Query().Get("status"),
		Service:   r.URL.Query().Get("service"),
		Actor:     r.URL.Query().Get("actor"),
		Search:    r.URL.Query().Get("search"),
		Page:      page,
		PageSize:  pageSize,
	}

	logs, total, err := h.svc.ListAuditLogs(r.Context(), filter)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":        logs,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetAuditLogByID handles GET /api/v1/audit/logs/{id}
func (h *AuditHandler) GetAuditLogByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	log, err := h.svc.GetAuditLogByID(r.Context(), id)
	if err != nil {
		response.JSON(w, http.StatusNotFound, map[string]string{"error": "audit log not found"})
		return
	}

	response.JSON(w, http.StatusOK, log)
}

// CreateAuditLog handles POST /api/v1/audit/logs
func (h *AuditHandler) CreateAuditLog(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAuditLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	if req.IPAddress == "" {
		req.IPAddress = r.RemoteAddr
	}
	if req.UserAgent == "" {
		req.UserAgent = r.UserAgent()
	}

	log, err := h.svc.CreateAuditLog(r.Context(), req)
	if err != nil {
		response.JSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusCreated, log)
}

// GetStats handles GET /api/v1/audit/stats
func (h *AuditHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, stats)
}

// GetSecurityEvents handles GET /api/v1/audit/security-events
func (h *AuditHandler) GetSecurityEvents(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	events, err := h.svc.GetSecurityEvents(r.Context(), limit)
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	response.JSON(w, http.StatusOK, events)
}

// VerifyIntegrity handles GET /api/v1/audit/integrity.
func (h *AuditHandler) VerifyIntegrity(w http.ResponseWriter, r *http.Request) {
	report, err := h.svc.VerifyIntegrity(r.Context())
	if err != nil {
		response.JSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	status := http.StatusOK
	if !report.Valid {
		status = http.StatusConflict
	}
	response.JSON(w, status, report)
}
