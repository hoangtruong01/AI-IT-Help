package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
	"eomp/packages/shared/pkg/response"
	"eomp/services/workflow/internal/model"
	"eomp/services/workflow/internal/service"
)

// WorkflowHandler handles HTTP endpoints for Workflow Definitions, Instances, and Logs
type WorkflowHandler struct {
	svc service.WorkflowService
}

// NewWorkflowHandler constructs a new WorkflowHandler
func NewWorkflowHandler(svc service.WorkflowService) *WorkflowHandler {
	return &WorkflowHandler{svc: svc}
}

// ListDefinitions returns all active blueprints
func (h *WorkflowHandler) ListDefinitions(w http.ResponseWriter, r *http.Request) {
	defs, err := h.svc.ListDefinitions(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, defs)
}

// GetDefinition returns definition by ID
func (h *WorkflowHandler) GetDefinition(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	def, err := h.svc.GetDefinition(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, def)
}

// ListInstances returns paginated workflow runs
func (h *WorkflowHandler) ListInstances(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	query := model.WorkflowListQuery{
		Page:     page,
		PageSize: pageSize,
		Status:   r.URL.Query().Get("status"),
		Search:   r.URL.Query().Get("search"),
	}

	resp, err := h.svc.ListInstances(r.Context(), query)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// StartWorkflow initiates a new workflow instance
func (h *WorkflowHandler) StartWorkflow(w http.ResponseWriter, r *http.Request) {
	var req model.CreateInstanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	if req.RequesterID == "" {
		req.RequesterID = middleware.GetUserID(r.Context())
	}
	if req.RequesterEmail == "" {
		req.RequesterEmail = middleware.GetUserEmail(r.Context())
	}

	inst, err := h.svc.StartWorkflow(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, inst)
}

// GetInstance returns workflow execution details
func (h *WorkflowHandler) GetInstance(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	inst, err := h.svc.GetInstance(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, inst)
}

// ListLogs returns audit logs for an execution
func (h *WorkflowHandler) ListLogs(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	logs, err := h.svc.ListLogs(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, logs)
}

// GetStats returns summary metrics
func (h *WorkflowHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}
