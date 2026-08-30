package handler

import (
	"encoding/json"
	stdErrors "errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/response"
	"eomp/services/workflow/internal/model"
	"eomp/services/workflow/internal/repository"
	"eomp/services/workflow/internal/service"
)

// ChangeHandler handles HTTP endpoints for ITIL Change Management and CAB.
type ChangeHandler struct {
	svc service.ChangeService
}

// NewChangeHandler constructs a new ChangeHandler instance.
func NewChangeHandler(svc service.ChangeService) *ChangeHandler {
	return &ChangeHandler{svc: svc}
}

// ListChanges returns paginated change requests.
func (h *ChangeHandler) ListChanges(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	changeType := r.URL.Query().Get("type")
	status := r.URL.Query().Get("status")
	risk := r.URL.Query().Get("risk")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	changes, total, err := h.svc.ListChanges(r.Context(), changeType, status, risk, page, pageSize)
	if err != nil {
		errors.WriteHTTP(w, errors.InternalServerError("failed to list change requests"))
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":        changes,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetChange returns a single Change Request and its CAB reviews.
func (h *ChangeHandler) GetChange(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("change id is required"))
		return
	}

	change, reviews, err := h.svc.GetChange(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, errors.NotFound("change request not found"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"change":      change,
		"cab_reviews": reviews,
	})
}

// CreateChange creates a new Request for Change (RFC).
func (h *ChangeHandler) CreateChange(w http.ResponseWriter, r *http.Request) {
	var payload model.CreateChangePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json payload"))
		return
	}

	change, err := h.svc.CreateChange(r.Context(), payload)
	if err != nil {
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusCreated, change)
}

// UpdateStatus updates change request status with CAB quorum validation.
func (h *ChangeHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("change id is required"))
		return
	}

	var payload model.UpdateChangeStatusPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json payload"))
		return
	}

	change, err := h.svc.UpdateStatus(r.Context(), id, payload)
	if err != nil {
		if stdErrors.Is(err, repository.ErrVersionConflict) {
			errors.WriteHTTP(w, errors.Conflict("change request was modified by another request; reload and retry"))
			return
		}
		// Test Case 7.2: If insufficient CAB approvals, return HTTP 403 Forbidden
		if strings.Contains(strings.ToLower(err.Error()), "insufficient cab") {
			errors.WriteHTTP(w, errors.Forbidden(err.Error()))
			return
		}
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusOK, change)
}

// SubmitCABVote submits a vote from a CAB member.
func (h *ChangeHandler) SubmitCABVote(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("change id is required"))
		return
	}

	var payload model.SubmitCABVotePayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json payload"))
		return
	}

	review, updatedChange, err := h.svc.SubmitCABVote(r.Context(), id, payload)
	if err != nil {
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"review": review,
		"change": updatedChange,
	})
}

// GetCalendar returns scheduled maintenance window changes.
func (h *ChangeHandler) GetCalendar(w http.ResponseWriter, r *http.Request) {
	startStr := r.URL.Query().Get("start")
	endStr := r.URL.Query().Get("end")

	var start, end time.Time
	if startStr != "" {
		start, _ = time.Parse(time.RFC3339, startStr)
	}
	if endStr != "" {
		end, _ = time.Parse(time.RFC3339, endStr)
	}

	items, err := h.svc.GetCalendar(r.Context(), start, end)
	if err != nil {
		errors.WriteHTTP(w, errors.InternalServerError("failed to load change calendar"))
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// GetStats returns KPI metrics for Change Management dashboard.
func (h *ChangeHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		errors.WriteHTTP(w, errors.InternalServerError("failed to load change stats"))
		return
	}
	response.JSON(w, http.StatusOK, stats)
}
