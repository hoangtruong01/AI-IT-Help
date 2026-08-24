package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/response"
	"eomp/services/asset/internal/model"
	"eomp/services/asset/internal/service"
)

// AssetHandler handles HTTP requests for hardware and software assets
type AssetHandler struct {
	svc service.AssetService
}

// NewAssetHandler constructs a new AssetHandler
func NewAssetHandler(svc service.AssetService) *AssetHandler {
	return &AssetHandler{svc: svc}
}

// ListAssets returns paginated assets
func (h *AssetHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	query := model.AssetListQuery{
		Page:             page,
		PageSize:         pageSize,
		Category:         r.URL.Query().Get("category"),
		Status:           r.URL.Query().Get("status"),
		Location:         r.URL.Query().Get("location"),
		AssignedToUserID: r.URL.Query().Get("assigned_to_user_id"),
		Search:           r.URL.Query().Get("search"),
	}

	resp, err := h.svc.ListAssets(r.Context(), query)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// CreateAsset registers a new asset
func (h *AssetHandler) CreateAsset(w http.ResponseWriter, r *http.Request) {
	var req model.CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	asset, err := h.svc.CreateAsset(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, asset)
}

// GetAsset returns asset by ID
func (h *AssetHandler) GetAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	asset, err := h.svc.GetAsset(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, asset)
}

// UpdateStatus updates lifecycle status and location
func (h *AssetHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req struct {
		Status   string  `json:"status"`
		Location string  `json:"location"`
		Notes    *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, req.Status, req.Location, req.Notes); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	asset, _ := h.svc.GetAsset(r.Context(), id)
	response.JSON(w, http.StatusOK, asset)
}

// AssignAsset assigns asset to employee
func (h *AssetHandler) AssignAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req model.AssignAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	if err := h.svc.AssignAsset(r.Context(), id, &req); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	asset, _ := h.svc.GetAsset(r.Context(), id)
	response.JSON(w, http.StatusOK, asset)
}

// ReturnAsset returns assigned asset to stock
func (h *AssetHandler) ReturnAsset(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	var req struct {
		Condition string  `json:"condition"`
		Notes     *string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid request body"))
		return
	}

	if err := h.svc.ReturnAsset(r.Context(), id, req.Condition, req.Notes); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	asset, _ := h.svc.GetAsset(r.Context(), id)
	response.JSON(w, http.StatusOK, asset)
}

// GetStats returns summary metrics
func (h *AssetHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.svc.GetStats(r.Context())
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}
	response.JSON(w, http.StatusOK, stats)
}

// ListAssignments returns assignment history for an asset
func (h *AssetHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	assignments, err := h.svc.ListAssignments(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, assignments)
}

// GetEmployeeAssetHistory returns all asset assignments for a given employee
func (h *AssetHandler) GetEmployeeAssetHistory(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("employee id is required"))
		return
	}

	history, err := h.svc.GetEmployeeAssetHistory(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, history)
}

// GetAssetIncidents returns all incident tickets associated with an asset
func (h *AssetHandler) GetAssetIncidents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) >= 2 {
			id = parts[len(parts)-2]
		}
	}

	if id == "" {
		errors.WriteHTTP(w, errors.BadRequest("asset id is required"))
		return
	}

	incidents, err := h.svc.GetAssetIncidents(r.Context(), id)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, incidents)
}

