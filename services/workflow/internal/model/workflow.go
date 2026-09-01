package model

import (
	"time"
)

// Trigger types
const (
	TriggerServiceRequest = "SERVICE_REQUEST"
	TriggerIncident       = "INCIDENT"
	TriggerManual         = "MANUAL"
	TriggerScheduled      = "SCHEDULED"
)

// Workflow instance statuses
const (
	InstanceStatusPending         = "PENDING"
	InstanceStatusRunning         = "RUNNING"
	InstanceStatusWaitingApproval = "WAITING_APPROVAL"
	InstanceStatusApproved        = "APPROVED"
	InstanceStatusRejected        = "REJECTED"
	InstanceStatusCompleted       = "COMPLETED"
	InstanceStatusFailed          = "FAILED"
	InstanceStatusCancelled       = "CANCELLED"
)

// Step types
const (
	StepTypeApproval        = "APPROVAL"
	StepTypeAutomatedAction = "AUTOMATED_ACTION"
	StepTypeNotification    = "NOTIFICATION"
	StepTypeScript          = "SCRIPT"
)

// Approval statuses
const (
	ApprovalStatusPending  = "PENDING"
	ApprovalStatusApproved = "APPROVED"
	ApprovalStatusRejected = "REJECTED"
)

// WorkflowDefinition entity
type WorkflowDefinition struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Category    string    `json:"category"`
	TriggerType string    `json:"trigger_type"`
	IsActive    bool      `json:"is_active"`
	StepsConfig string    `json:"steps_config"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// WorkflowInstance entity
type WorkflowInstance struct {
	ID              string     `json:"id"`
	InstanceNumber  string     `json:"instance_number"`
	DefinitionID    string     `json:"definition_id"`
	DefinitionName  string     `json:"definition_name"`
	EntityType      string     `json:"entity_type"`
	EntityID        string     `json:"entity_id"`
	Title           string     `json:"title"`
	RequesterID     string     `json:"requester_id"`
	RequesterName   string     `json:"requester_name"`
	RequesterEmail  string     `json:"requester_email"`
	DepartmentID    string     `json:"department_id,omitempty"`
	CurrentStepName string     `json:"current_step_name"`
	Status          string     `json:"status"`
	ContextData     *string    `json:"context_data,omitempty"`
	StartedAt       time.Time  `json:"started_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	Version         int        `json:"version"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ApprovalRequest entity
type ApprovalRequest struct {
	ID            string     `json:"id"`
	InstanceID    string     `json:"instance_id"`
	StepID        *string    `json:"step_id,omitempty"`
	Title         string     `json:"title"`
	ApproverID    string     `json:"approver_id"`
	ApproverName  string     `json:"approver_name"`
	ApproverRole  string     `json:"approver_role"`
	ApprovalLevel int        `json:"approval_level"`
	Status        string     `json:"status"`
	DecisionNotes *string    `json:"decision_notes,omitempty"`
	DecidedAt     *time.Time `json:"decided_at,omitempty"`
	SLADeadline   time.Time  `json:"sla_deadline"`
	CreatedAt     time.Time  `json:"created_at"`
}

// WorkflowLog audit record
type WorkflowLog struct {
	ID         string    `json:"id"`
	InstanceID string    `json:"instance_id"`
	ActorID    string    `json:"actor_id"`
	ActorName  string    `json:"actor_name"`
	Action     string    `json:"action"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

// CreateInstanceRequest DTO
type CreateInstanceRequest struct {
	DefinitionID   string  `json:"definition_id"`
	EntityType     string  `json:"entity_type"`
	EntityID       string  `json:"entity_id"`
	Title          string  `json:"title"`
	RequesterID    string  `json:"requester_id"`
	RequesterName  string  `json:"requester_name"`
	RequesterEmail string  `json:"requester_email"`
	DepartmentID   string  `json:"department_id,omitempty"`
	ContextData    *string `json:"context_data,omitempty"`
}

// ApprovalDecisionRequest DTO
type ApprovalDecisionRequest struct {
	Decision string `json:"decision"` // APPROVED / REJECTED
	Notes    string `json:"notes"`
}

// WorkflowListQuery parameters
type WorkflowListQuery struct {
	Page              int    `json:"page"`
	PageSize          int    `json:"page_size"`
	Status            string `json:"status"`
	Search            string `json:"search"`
	RequesterID       string `json:"requester_id,omitempty"`
	ActorRole         string `json:"-"`
	ActorID           string `json:"-"`
	ActorDepartmentID string `json:"-"`
}

// WorkflowListResponse envelope
type WorkflowListResponse struct {
	Data       []WorkflowInstance `json:"data"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// ApprovalListResponse envelope
type ApprovalListResponse struct {
	Data       []ApprovalRequest `json:"data"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// WorkflowStats summary metrics
type WorkflowStats struct {
	TotalDefinitions int `json:"total_definitions"`
	ActiveInstances  int `json:"active_instances"`
	PendingApprovals int `json:"pending_approvals"`
	CompletedToday   int `json:"completed_today"`
}
