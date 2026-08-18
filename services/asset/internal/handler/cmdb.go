package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/response"
	"eomp/services/asset/internal/model"
	"eomp/services/asset/internal/service"
)

// CMDBHandler handles HTTP endpoints for Configuration Items and Topology
type CMDBHandler struct {
	svc service.CMDBService
}

// NewCMDBHandler constructs a new CMDBHandler
func NewCMDBHandler(svc service.CMDBService) *CMDBHandler {
	return &CMDBHandler{svc: svc}
}

// ListCIs returns all configuration items with filtering
func (h *CMDBHandler) ListCIs(w http.ResponseWriter, r *http.Request) {
	env := r.URL.Query().Get("environment")
	ciType := r.URL.Query().Get("ci_type")
	status := r.URL.Query().Get("status")

	cis, err := h.svc.ListCIs(r.Context(), env, ciType, status)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, cis)
}

// CreateCI registers a new configuration item
func (h *CMDBHandler) CreateCI(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCIRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	ci, err := h.svc.CreateCI(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, ci)
}

// GetCI returns CI by ID
func (h *CMDBHandler) GetCI(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	ci, err := h.svc.GetCI(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, ci)
}

// UpdateCIStatus updates operational status of a CI
func (h *CMDBHandler) UpdateCIStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	if err := h.svc.UpdateCIStatus(r.Context(), id, req.Status); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	ci, _ := h.svc.GetCI(r.Context(), id)
	response.JSON(w, http.StatusOK, ci)
}

// ListRelationships returns all dependency edges
func (h *CMDBHandler) ListRelationships(w http.ResponseWriter, r *http.Request) {
	rels, err := h.svc.ListRelationships(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, rels)
}

// CreateRelationship adds a dependency edge
func (h *CMDBHandler) CreateRelationship(w http.ResponseWriter, r *http.Request) {
	var req model.CreateCIRelationshipRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	if err := h.svc.CreateRelationship(r.Context(), &req); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]string{"message": "relationship created"})
}

// GetTopology returns full node and edge topology graph
func (h *CMDBHandler) GetTopology(w http.ResponseWriter, r *http.Request) {
	topology, err := h.svc.GetTopology(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, topology)
}
