package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/response"
	"eomp/services/helpdesk/internal/model"
	"eomp/services/helpdesk/internal/service"
)

// ProblemHandler handles HTTP endpoints for ITIL Problem Management.
type ProblemHandler struct {
	svc service.ProblemService
}

// NewProblemHandler constructs a new ProblemHandler instance.
func NewProblemHandler(svc service.ProblemService) *ProblemHandler {
	return &ProblemHandler{svc: svc}
}

// ListProblems returns paginated problems.
func (h *ProblemHandler) ListProblems(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	category := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	problems, total, err := h.svc.ListProblems(r.Context(), category, status, page, pageSize)
	if err != nil {
		errors.WriteHTTP(w, errors.InternalServerError("failed to list problems"))
		return
	}

	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"data":        problems,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	})
}

// GetProblem returns problem details and linked incidents.
func (h *ProblemHandler) GetProblem(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("problem id is required"))
		return
	}

	problem, links, err := h.svc.GetProblem(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, errors.NotFound("problem not found"))
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"problem":          problem,
		"linked_incidents": links,
	})
}

// CreateProblem creates a new Problem record.
func (h *ProblemHandler) CreateProblem(w http.ResponseWriter, r *http.Request) {
	var payload model.CreateProblemPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json payload"))
		return
	}

	problem, err := h.svc.CreateProblem(r.Context(), payload)
	if err != nil {
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusCreated, problem)
}

// UpdateStatus updates problem status and cascades to linked incidents if resolved.
func (h *ProblemHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("problem id is required"))
		return
	}

	var payload model.UpdateProblemStatusPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json payload"))
		return
	}

	problem, cascadedTickets, err := h.svc.UpdateStatus(r.Context(), id, payload)
	if err != nil {
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"problem":           problem,
		"cascaded_tickets":  cascadedTickets,
		"cascaded_count":    len(cascadedTickets),
		"cascade_completed": len(cascadedTickets) > 0,
	})
}

// UpdateRCA updates problem root cause analysis and workaround.
func (h *ProblemHandler) UpdateRCA(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("problem id is required"))
		return
	}

	var payload model.UpdateProblemRCAPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json payload"))
		return
	}

	problem, err := h.svc.UpdateRCA(r.Context(), id, payload)
	if err != nil {
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusOK, problem)
}

// LinkIncident links an incident ticket to the problem.
func (h *ProblemHandler) LinkIncident(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("problem id is required"))
		return
	}

	var payload model.LinkIncidentPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json payload"))
		return
	}

	link, err := h.svc.LinkIncident(r.Context(), id, payload)
	if err != nil {
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusCreated, link)
}

// UnlinkIncident unlinks an incident ticket from the problem.
func (h *ProblemHandler) UnlinkIncident(w http.ResponseWriter, r *http.Request) {
	problemID := r.PathValue("id")
	ticketID := r.PathValue("ticketId")

	if problemID == "" || ticketID == "" {
		errors.WriteHTTP(w, errors.BadRequest("problem id and ticket id are required"))
		return
	}

	if err := h.svc.UnlinkIncident(r.Context(), problemID, ticketID); err != nil {
		errors.WriteHTTP(w, errors.BadRequest(err.Error()))
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "incident unlinked successfully",
	})
}

// GetStats returns KPI statistics for Problem Management.
func (h *ProblemHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		errors.WriteHTTP(w, errors.InternalServerError("failed to load problem statistics"))
		return
	}
	response.JSON(w, http.StatusOK, stats)
}
