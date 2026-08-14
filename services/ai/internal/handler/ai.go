package handler

import (
	"encoding/json"
	"net/http"

	"eomp/packages/shared/pkg/response"
	"eomp/services/ai/internal/model"
	"eomp/services/ai/internal/service"
)

// AIHandler handles incoming HTTP requests for AI operations.
type AIHandler struct {
	svc service.AIService
}

// NewAIHandler constructs a new AIHandler instance.
func NewAIHandler(svc service.AIService) *AIHandler {
	return &AIHandler{svc: svc}
}

// Chat handles interactive conversation with the AI assistant.
func (h *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req model.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	res, err := h.svc.Chat(r.Context(), &req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, res)
}

// AnalyzeTicketRequest payload for ticket triage endpoint.
type AnalyzeTicketRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// AnalyzeTicket evaluates helpdesk ticket content for classification and suggestions.
func (h *AIHandler) AnalyzeTicket(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req AnalyzeTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	res, err := h.svc.AnalyzeTicket(r.Context(), req.Title, req.Description)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, err.Error())
		return
	}

	response.JSON(w, http.StatusOK, res)
}
