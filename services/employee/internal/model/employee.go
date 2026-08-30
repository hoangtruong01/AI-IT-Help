package model

import (
	"time"
)

// Department entity
type Department struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	ManagerID *string   `json:"manager_id,omitempty"`
	ParentID  *string   `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateDepartmentRequest DTO
type CreateDepartmentRequest struct {
	Name      string  `json:"name"`
	Code      string  `json:"code"`
	ManagerID *string `json:"manager_id,omitempty"`
	ParentID  *string `json:"parent_id,omitempty"`
}

// Employee entity with joined department information
type Employee struct {
	ID             string    `json:"id"`
	UserID         *string   `json:"user_id,omitempty"`
	FirstName      string    `json:"first_name"`
	LastName       string    `json:"last_name"`
	FullName       string    `json:"full_name"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone,omitempty"`
	JobTitle       string    `json:"job_title"`
	DepartmentID   *string   `json:"department_id,omitempty"`
	DepartmentName *string   `json:"department_name,omitempty"`
	DepartmentCode *string   `json:"department_code,omitempty"`
	ManagerID      *string   `json:"manager_id,omitempty"`
	ManagerName    *string   `json:"manager_name,omitempty"`
	Status         string    `json:"status"`
	Location       string    `json:"location"`
	JoinedAt       string    `json:"joined_at"`
	Version        int       `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// CreateEmployeeRequest DTO
type CreateEmployeeRequest struct {
	UserID       *string `json:"user_id,omitempty"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone,omitempty"`
	JobTitle     string  `json:"job_title"`
	DepartmentID *string `json:"department_id,omitempty"`
	ManagerID    *string `json:"manager_id,omitempty"`
	Status       string  `json:"status,omitempty"`
	Location     string  `json:"location,omitempty"`
	JoinedAt     string  `json:"joined_at,omitempty"`
}

// UpdateEmployeeRequest DTO
type UpdateEmployeeRequest struct {
	Version      int     `json:"version"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	Phone        *string `json:"phone,omitempty"`
	JobTitle     *string `json:"job_title,omitempty"`
	DepartmentID *string `json:"department_id,omitempty"`
	ManagerID    *string `json:"manager_id,omitempty"`
	Status       *string `json:"status,omitempty"`
	Location     *string `json:"location,omitempty"`
}

// EmployeeListQuery filters and pagination
type EmployeeListQuery struct {
	Page         int    `json:"page"`
	PageSize     int    `json:"page_size"`
	DepartmentID string `json:"department_id"`
	Status       string `json:"status"`
	Search       string `json:"search"`
}

// EmployeeListResponse paginated envelope
type EmployeeListResponse struct {
	Data       []Employee `json:"data"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
