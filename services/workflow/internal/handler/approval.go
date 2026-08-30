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

// ApprovalHandler handles HTTP endpoints for Approval Requests
type ApprovalHandler struct {
	svc service.WorkflowService
}

// NewApprovalHandler constructs a new ApprovalHandler
func NewApprovalHandler(svc service.WorkflowService) *ApprovalHandler {
	return &ApprovalHandler{svc: svc}
}

// ListApprovals returns pending or historical approval requests
func (h *ApprovalHandler) ListApprovals(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	approverID := r.URL.Query().Get("approver_id")
	role := middleware.GetUserRole(r.Context())
	if role == "ROLE_EMPLOYEE" {
		errors.WriteHTTP(w, errors.Forbidden("approval access requires manager privileges"))
		return
	}
	approverRole := ""
	if role != "ROLE_ADMIN" {
		approverID = middleware.GetUserID(r.Context())
		approverRole = role
	}
	status := r.URL.Query().Get("status")

	resp, err := h.svc.ListApprovals(r.Context(), approverID, approverRole, status, page, pageSize)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ProcessDecision approves or rejects a pending request
func (h *ApprovalHandler) ProcessDecision(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req model.ApprovalDecisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	actorID := middleware.GetUserID(r.Context())
	if actorID == "" {
		actorID = "system"
	}
	actorName := middleware.GetUserEmail(r.Context())
	if actorName == "" {
		actorName = "Approver"
	}

	actorRole := middleware.GetUserRole(r.Context())
	if err := h.svc.ProcessApprovalDecision(r.Context(), id, &req, actorID, actorName, actorRole); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "approval decision recorded successfully",
		"status":  req.Decision,
	})
}
