package model

import (
	"time"
)

// Roles defined in EOMP
const (
	RoleAdmin    = "ROLE_ADMIN"
	RoleManager  = "ROLE_MANAGER"
	RoleAgent    = "ROLE_AGENT"
	RoleEmployee = "ROLE_EMPLOYEE"
)

// User represents the persistent user account in auth_db
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	DepartmentID *string   `json:"department_id,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest contains payload to create a new user account
type RegisterRequest struct {
	Email        string  `json:"email"`
	Password     string  `json:"password"`
	FullName     string  `json:"full_name"`
	DepartmentID *string `json:"department_id,omitempty"`
}

// LoginRequest contains login credentials
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshTokenRequest contains the refresh token payload
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest contains the refresh token to revoke
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LoginAuditLog represents an entry in login_audit_logs table
type LoginAuditLog struct {
	ID            string    `json:"id"`
	UserID        *string   `json:"user_id,omitempty"`
	Email         string    `json:"email"`
	IPAddress     string    `json:"ip_address"`
	UserAgent     *string   `json:"user_agent,omitempty"`
	Status        string    `json:"status"` // SUCCESS, FAILED, LOCKED
	FailureReason *string   `json:"failure_reason,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// Login audit status constants
const (
	LoginStatusSuccess = "SUCCESS"
	LoginStatusFailed  = "FAILED"
	LoginStatusLocked  = "LOCKED"
)

// UserResponse is the public sanitized user representation
type UserResponse struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	FullName     string    `json:"full_name"`
	Role         string    `json:"role"`
	DepartmentID *string   `json:"department_id,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// AuthResponse is returned on successful login or token refresh
type AuthResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	TokenType    string       `json:"token_type"`
	ExpiresIn    int64        `json:"expires_in"` // seconds
	User         UserResponse `json:"user"`
}

// CreateUserRequest contains payload for admin user creation
type CreateUserRequest struct {
	Email        string  `json:"email"`
	Password     string  `json:"password"`
	FullName     string  `json:"full_name"`
	Role         string  `json:"role"`
	DepartmentID *string `json:"department_id,omitempty"`
}

// UpdateUserRequest contains payload for updating user account details
type UpdateUserRequest struct {
	FullName     *string `json:"full_name,omitempty"`
	Role         *string `json:"role,omitempty"`
	DepartmentID *string `json:"department_id,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
}

// ChangePasswordRequest contains payload for self-service password change
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// AdminResetPasswordRequest contains payload for administrator password reset
type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// UserListQuery parameters
type UserListQuery struct {
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	Search       string `json:"search"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id"`
}

// UserListResponse paginated envelope
type UserListResponse struct {
	Data       []UserResponse `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// ToResponse converts User entity to sanitized UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		FullName:     u.FullName,
		Role:         u.Role,
		DepartmentID: u.DepartmentID,
		IsActive:     u.IsActive,
		CreatedAt:    u.CreatedAt,
	}
}
