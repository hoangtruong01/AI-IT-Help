package model

import (
	"time"
)

// ServiceCategory represents a group of IT service catalog items
type ServiceCategory struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Icon        string               `json:"icon"`
	Description *string              `json:"description,omitempty"`
	Items       []ServiceCatalogItem `json:"items,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
}

// ServiceCatalogItem represents an individual service or request template
type ServiceCatalogItem struct {
	ID                   string    `json:"id"`
	CategoryID           string    `json:"category_id"`
	CategoryName         *string   `json:"category_name,omitempty"`
	Name                 string    `json:"name"`
	Code                 string    `json:"code"`
	Description          *string   `json:"description,omitempty"`
	DefaultPriority      string    `json:"default_priority"`
	SLAResponseMinutes   int       `json:"sla_response_minutes"`
	SLAResolutionMinutes int       `json:"sla_resolution_minutes"`
	RequiresApproval     bool      `json:"requires_approval"`
	IsActive             bool      `json:"is_active"`
	CreatedAt            time.Time `json:"created_at"`
}

// CreateCategoryRequest DTO
type CreateCategoryRequest struct {
	Name        string  `json:"name"`
	Icon        string  `json:"icon,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateCatalogItemRequest DTO
type CreateCatalogItemRequest struct {
	CategoryID           string  `json:"category_id"`
	Name                 string  `json:"name"`
	Code                 string  `json:"code"`
	Description          *string `json:"description,omitempty"`
	DefaultPriority      string  `json:"default_priority,omitempty"`
	SLAResponseMinutes   int     `json:"sla_response_minutes,omitempty"`
	SLAResolutionMinutes int     `json:"sla_resolution_minutes,omitempty"`
	RequiresApproval     bool    `json:"requires_approval,omitempty"`
}
