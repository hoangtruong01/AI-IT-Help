package model

import (
	"time"
)

// Asset categories
const (
	CategoryLaptop  = "LAPTOP"
	CategoryDesktop = "DESKTOP"
	CategoryServer  = "SERVER"
	CategoryMonitor = "MONITOR"
	CategoryMobile  = "MOBILE"
	CategoryNetwork = "NETWORK"
	CategoryLicense = "LICENSE"
	CategoryOther   = "OTHER"
)

// Asset lifecycle statuses
const (
	StatusProcurement = "PROCUREMENT"
	StatusReceived    = "RECEIVED"
	StatusInStock     = "IN_STOCK"
	StatusAssigned    = "ASSIGNED"
	StatusInUse       = "IN_USE"
	StatusMaintenance = "MAINTENANCE"
	StatusRetired     = "RETIRED"
	StatusDisposed    = "DISPOSED"
)

// Asset entity
type Asset struct {
	ID                 string     `json:"id"`
	AssetTag           string     `json:"asset_tag"`
	Name               string     `json:"name"`
	Category           string     `json:"category"`
	Model              *string    `json:"model,omitempty"`
	SerialNumber       *string    `json:"serial_number,omitempty"`
	PurchaseDate       *string    `json:"purchase_date,omitempty"`
	PurchaseCost       float64    `json:"purchase_cost"`
	WarrantyExpiry     *string    `json:"warranty_expiry,omitempty"`
	CurrentValue       float64    `json:"current_value"`
	Status             string     `json:"status"`
	Location           string     `json:"location"`
	AssignedToUserID   *string    `json:"assigned_to_user_id,omitempty"`
	AssignedToUserName *string    `json:"assigned_to_user_name,omitempty"`
	AssignedAt         *time.Time `json:"assigned_at,omitempty"`
	Notes              *string    `json:"notes,omitempty"`
	Version            int        `json:"version"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CreateAssetRequest DTO
type CreateAssetRequest struct {
	AssetTag       string  `json:"asset_tag"`
	Name           string  `json:"name"`
	Category       string  `json:"category"`
	Model          *string `json:"model,omitempty"`
	SerialNumber   *string `json:"serial_number,omitempty"`
	PurchaseDate   *string `json:"purchase_date,omitempty"`
	PurchaseCost   float64 `json:"purchase_cost"`
	WarrantyExpiry *string `json:"warranty_expiry,omitempty"`
	Location       string  `json:"location"`
	Notes          *string `json:"notes,omitempty"`
}

// AssignAssetRequest DTO
type AssignAssetRequest struct {
	UserID            string  `json:"user_id"`
	UserName          string  `json:"user_name"`
	DepartmentID      *string `json:"department_id,omitempty"`
	ConditionOnAssign string  `json:"condition_on_assign"`
	Notes             *string `json:"notes,omitempty"`
	ExpectedVersion   int     `json:"version"`
}

// AssetAssignment entity
type AssetAssignment struct {
	ID                string     `json:"id"`
	AssetID           string     `json:"asset_id"`
	UserID            string     `json:"user_id"`
	UserName          string     `json:"user_name"`
	DepartmentID      *string    `json:"department_id,omitempty"`
	AssignedAt        time.Time  `json:"assigned_at"`
	ReturnedAt        *time.Time `json:"returned_at,omitempty"`
	ConditionOnAssign string     `json:"condition_on_assign"`
	ConditionOnReturn *string    `json:"condition_on_return,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
}

// AssetListQuery DTO
type AssetListQuery struct {
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
	Category         string `json:"category"`
	Status           string `json:"status"`
	Location         string `json:"location"`
	AssignedToUserID string `json:"assigned_to_user_id"`
	Search           string `json:"search"`
}

// AssetListResponse envelope
type AssetListResponse struct {
	Data       []Asset `json:"data"`
	Total      int     `json:"total"`
	Page       int     `json:"page"`
	PageSize   int     `json:"page_size"`
	TotalPages int     `json:"total_pages"`
}

// AssetStatsResponse summary metrics
type AssetStatsResponse struct {
	TotalAssets   int     `json:"total_assets"`
	InUse         int     `json:"in_use"`
	InStock       int     `json:"in_stock"`
	InMaintenance int     `json:"in_maintenance"`
	TotalValue    float64 `json:"total_value"`
}

// EmployeeAssetHistoryItem represents an asset assignment record with joined asset details
type EmployeeAssetHistoryItem struct {
	AssignmentID      string     `json:"assignment_id"`
	AssetID           string     `json:"asset_id"`
	AssetTag          string     `json:"asset_tag"`
	AssetName         string     `json:"asset_name"`
	Category          string     `json:"category"`
	Model             *string    `json:"model,omitempty"`
	SerialNumber      *string    `json:"serial_number,omitempty"`
	AssetStatus       string     `json:"asset_status"`
	AssignedAt        time.Time  `json:"assigned_at"`
	ReturnedAt        *time.Time `json:"returned_at,omitempty"`
	ConditionOnAssign string     `json:"condition_on_assign"`
	ConditionOnReturn *string    `json:"condition_on_return,omitempty"`
	Notes             *string    `json:"notes,omitempty"`
}

// AssetIncidentHistoryItem represents an incident/ticket linked to an asset
type AssetIncidentHistoryItem struct {
	TicketID      string     `json:"ticket_id"`
	TicketNumber  string     `json:"ticket_number"`
	Title         string     `json:"title"`
	Category      string     `json:"category"`
	Priority      string     `json:"priority"`
	Status        string     `json:"status"`
	RequesterID   string     `json:"requester_id"`
	RequesterName string     `json:"requester_name"`
	AssigneeName  *string    `json:"assignee_name,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
}
