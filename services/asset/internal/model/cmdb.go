package model

import (
	"time"
)

// CI Types
const (
	CITypeApplication   = "APPLICATION"
	CITypeAPIService    = "API_SERVICE"
	CITypeServer        = "SERVER"
	CITypeDatabase      = "DATABASE"
	CITypeNetworkDevice = "NETWORK_DEVICE"
	CITypeCloudResource = "CLOUD_RESOURCE"
)

// Environments
const (
	EnvProduction  = "PRODUCTION"
	EnvStaging     = "STAGING"
	EnvDevelopment = "DEVELOPMENT"
	EnvDR          = "DR"
)

// CI Statuses
const (
	CIStatusOperational = "OPERATIONAL"
	CIStatusDegraded    = "DEGRADED"
	CIStatusOffline     = "OFFLINE"
	CIStatusMaintenance = "MAINTENANCE"
)

// Relationship Types
const (
	RelDependsOn  = "DEPENDS_ON"
	RelRunsOn     = "RUNS_ON"
	RelConnectsTo = "CONNECTS_TO"
	RelBackedUpBy = "BACKED_UP_BY"
)

// ConfigurationItem (CI) entity
type ConfigurationItem struct {
	ID          string    `json:"id"`
	CICode      string    `json:"ci_code"`
	Name        string    `json:"name"`
	CIType      string    `json:"ci_type"`
	Environment string    `json:"environment"`
	OwnerID     *string   `json:"owner_id,omitempty"`
	OwnerName   *string   `json:"owner_name,omitempty"`
	Status      string    `json:"status"`
	IPAddress   *string   `json:"ip_address,omitempty"`
	AssetID     *string   `json:"asset_id,omitempty"`
	Description *string   `json:"description,omitempty"`
	Version     int       `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CIRelationship edge between two CIs
type CIRelationship struct {
	ID               string    `json:"id"`
	ParentCIID       string    `json:"parent_ci_id"`
	ParentCIName     *string   `json:"parent_ci_name,omitempty"`
	ParentCIType     *string   `json:"parent_ci_type,omitempty"`
	ChildCIID        string    `json:"child_ci_id"`
	ChildCIName      *string   `json:"child_ci_name,omitempty"`
	ChildCIType      *string   `json:"child_ci_type,omitempty"`
	RelationshipType string    `json:"relationship_type"`
	ImpactWeight     string    `json:"impact_weight"`
	CreatedAt        time.Time `json:"created_at"`
}

// CreateCIRequest DTO
type CreateCIRequest struct {
	CICode      string  `json:"ci_code"`
	Name        string  `json:"name"`
	CIType      string  `json:"ci_type"`
	Environment string  `json:"environment"`
	OwnerID     *string `json:"owner_id,omitempty"`
	OwnerName   *string `json:"owner_name,omitempty"`
	IPAddress   *string `json:"ip_address,omitempty"`
	AssetID     *string `json:"asset_id,omitempty"`
	Description *string `json:"description,omitempty"`
}

// CreateCIRelationshipRequest DTO
type CreateCIRelationshipRequest struct {
	ParentCIID       string `json:"parent_ci_id"`
	ChildCIID        string `json:"child_ci_id"`
	RelationshipType string `json:"relationship_type"`
	ImpactWeight     string `json:"impact_weight"`
}

// CMDBTopologyGraph represents topology structure
type CMDBTopologyGraph struct {
	Nodes []ConfigurationItem `json:"nodes"`
	Edges []CIRelationship    `json:"edges"`
}
