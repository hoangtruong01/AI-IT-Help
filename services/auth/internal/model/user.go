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
	Role         string  `json:"role,omitempty"`
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
