package handler

import (
	"encoding/json"
	"net/http"

	"eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
	"eomp/packages/shared/pkg/response"
	"eomp/services/auth/internal/model"
	"eomp/services/auth/internal/service"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler constructs a new AuthHandler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

// Login authenticates a user and returns a token pair
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// RefreshToken renews an access token using a valid refresh token
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req model.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	resp, err := h.authService.RefreshToken(r.Context(), &req)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// GetMe returns the authenticated user's profile
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteHTTP(w, errors.Unauthorized("unauthorized: missing user context"))
		return
	}

	resp, err := h.authService.GetMe(r.Context(), userID)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}
