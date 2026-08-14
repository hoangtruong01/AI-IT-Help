package handler

import (
	"net/http"

	"eomp/packages/shared/pkg/response"
	"eomp/services/employee/internal/config"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	cfg *config.Config
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

// HealthResponse represents standard health check response payload.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// Check responds with service health status.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: h.cfg.ServiceName,
		Version: h.cfg.Version,
	})
}
