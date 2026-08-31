package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

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

	clientIP := middleware.ExtractClientIP(r, nil)
	userAgent := r.UserAgent()

	resp, err := h.authService.LoginWithAudit(r.Context(), &req, clientIP, userAgent)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// Logout revokes the provided refresh token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req model.LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	if err := h.authService.Logout(r.Context(), &req); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "logged out successfully and token revoked",
	})
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

// GetLoginHistory returns login audit logs
func (h *AuthHandler) GetLoginHistory(w http.ResponseWriter, r *http.Request) {
	userRole := middleware.GetUserRole(r.Context())
	userEmail := middleware.GetUserEmail(r.Context())
	if middleware.GetUserID(r.Context()) == "" || userEmail == "" {
		errors.WriteHTTP(w, errors.Unauthorized("authentication required"))
		return
	}

	// If not admin, restrict to caller's email
	queryEmail := r.URL.Query().Get("email")
	if userRole != model.RoleAdmin && userRole != model.RoleManager {
		queryEmail = userEmail
	}

	logs, err := h.authService.GetLoginHistory(r.Context(), queryEmail, 50)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, logs)
}

// ListUsers returns paginated users (Admin and Manager only)
func (h *AuthHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	query := model.UserListQuery{
		Page:         page,
		PageSize:     pageSize,
		Search:       r.URL.Query().Get("search"),
		Role:         r.URL.Query().Get("role"),
		DepartmentID: r.URL.Query().Get("department_id"),
	}

	resp, err := h.authService.ListUsers(r.Context(), query, middleware.GetActor(r.Context()))
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// CreateUser handles admin user provisioning
func (h *AuthHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	resp, err := h.authService.CreateUser(r.Context(), &req, middleware.GetActor(r.Context()))
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, resp)
}

// UpdateUser handles admin modifications to role/department/active status
func (h *AuthHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	var req model.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	actor := middleware.GetActor(r.Context())
	resp, err := h.authService.UpdateUser(r.Context(), id, &req, actor)
	if err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

// ChangePassword allows authenticated users to change their own password
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		errors.WriteHTTP(w, errors.Unauthorized("missing user identity"))
		return
	}

	var req model.ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	if err := h.authService.ChangePassword(r.Context(), userID, &req); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "password changed successfully; all previous sessions revoked",
	})
}

// AdminResetPassword allows administrators to reset another user's password
func (h *AuthHandler) AdminResetPassword(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) > 0 {
			id = parts[len(parts)-1]
		}
	}

	var req model.AdminResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errors.WriteHTTP(w, errors.BadRequest("invalid json request body"))
		return
	}

	if err := h.authService.AdminResetPassword(r.Context(), id, &req, middleware.GetActor(r.Context())); err != nil {
		errors.WriteHTTP(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]any{
		"message": "user password reset successfully and sessions revoked",
	})
}
