package handler

import (
	"io"
	"net/http"
	"strings"
	"time"

	"eomp/packages/shared/pkg/response"
	"eomp/services/ai/internal/config"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewHealthHandler creates a new HealthHandler instance.
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: 3 * time.Second},
	}
}

// HealthResponse represents standard health check response payload.
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
}

// RuntimeStatus reports configured AI capabilities and an actively observed
// Qdrant status without exposing provider credentials.
type RuntimeStatus struct {
	ServiceStatus       string    `json:"service_status"`
	Provider            string    `json:"provider"`
	Model               string    `json:"model"`
	EmbeddingModel      string    `json:"embedding_model"`
	MockFallbackEnabled bool      `json:"mock_fallback_enabled"`
	AutoIngestEnabled   bool      `json:"auto_ingest_enabled"`
	QdrantStatus        string    `json:"qdrant_status"`
	QdrantCollection    string    `json:"qdrant_collection"`
	LastCheckedAt       time.Time `json:"last_checked_at"`
}

// Check responds with service health status.
func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, HealthResponse{
		Status:  "ok",
		Service: h.cfg.ServiceName,
		Version: h.cfg.Version,
	})
}

// Status probes the configured Qdrant collection and returns observed runtime configuration.
func (h *HealthHandler) Status(w http.ResponseWriter, r *http.Request) {
	status := RuntimeStatus{
		ServiceStatus:       "ONLINE",
		Provider:            h.cfg.AIProvider,
		Model:               h.cfg.AIModel,
		EmbeddingModel:      h.cfg.EmbeddingModel,
		MockFallbackEnabled: h.cfg.AIProvider == "mock" || h.cfg.AllowMockAI,
		AutoIngestEnabled:   h.cfg.AutoIngest,
		QdrantStatus:        "OFFLINE",
		QdrantCollection:    h.cfg.QdrantCollection,
		LastCheckedAt:       time.Now().UTC(),
	}

	probeURL := strings.TrimRight(h.cfg.QdrantURL, "/") + "/collections/" + h.cfg.QdrantCollection
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, probeURL, nil)
	if err == nil {
		resp, probeErr := h.httpClient.Do(req)
		if probeErr == nil {
			defer resp.Body.Close()
			_, _ = io.Copy(io.Discard, resp.Body)
			if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices {
				status.QdrantStatus = "ONLINE"
			}
		}
	}

	response.JSON(w, http.StatusOK, status)
}
