package model

import "time"

// ChangeRequest represents an ITIL Request for Change (RFC).
type ChangeRequest struct {
	ID                 string     `json:"id" db:"id"`
	ChangeNumber       string     `json:"change_number" db:"change_number"`
	Title              string     `json:"title" db:"title"`
	Description        string     `json:"description" db:"description"`
	ChangeType         string     `json:"change_type" db:"change_type"` // STANDARD, NORMAL, EMERGENCY, MAJOR
	Category           string     `json:"category" db:"category"`
	Priority           string     `json:"priority" db:"priority"`
	RiskLevel          string     `json:"risk_level" db:"risk_level"`               // LOW, MEDIUM, HIGH, CRITICAL
	ImpactLevel        string     `json:"impact_level" db:"impact_level"`           // LOW, MEDIUM, HIGH, CRITICAL
	ProbabilityLevel   string     `json:"probability_level" db:"probability_level"` // LOW, MEDIUM, HIGH, CRITICAL
	Status             string     `json:"status" db:"status"`                       // DRAFT, SUBMITTED, CAB_REVIEW, APPROVED, REJECTED, SCHEDULED, IMPLEMENTING, COMPLETED, FAILED, CANCELLED
	RequesterID        string     `json:"requester_id" db:"requester_id"`
	RequesterName      string     `json:"requester_name" db:"requester_name"`
	RequesterEmail     string     `json:"requester_email" db:"requester_email"`
	AssignedToID       *string    `json:"assigned_to_id,omitempty" db:"assigned_to_id"`
	AssignedToName     *string    `json:"assigned_to_name,omitempty" db:"assigned_to_name"`
	ReasonForChange    string     `json:"reason_for_change" db:"reason_for_change"`
	ImplementationPlan string     `json:"implementation_plan" db:"implementation_plan"`
	RollbackPlan       string     `json:"rollback_plan" db:"rollback_plan"`
	TestPlan           string     `json:"test_plan" db:"test_plan"`
	ScheduledStartTime *time.Time `json:"scheduled_start_time,omitempty" db:"scheduled_start_time"`
	ScheduledEndTime   *time.Time `json:"scheduled_end_time,omitempty" db:"scheduled_end_time"`
	ActualStartTime    *time.Time `json:"actual_start_time,omitempty" db:"actual_start_time"`
	ActualEndTime      *time.Time `json:"actual_end_time,omitempty" db:"actual_end_time"`
	DowntimeRequired   bool       `json:"downtime_required" db:"downtime_required"`
	DowntimeMinutes    int        `json:"downtime_minutes" db:"downtime_minutes"`
	CABRequiredCount   int        `json:"cab_required_count" db:"cab_required_count"`
	CABApprovedCount   int        `json:"cab_approved_count" db:"cab_approved_count"`
	Version            int        `json:"version" db:"version"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// CABReview represents a review and vote from a CAB member.
type CABReview struct {
	ID           string    `json:"id" db:"id"`
	ChangeID     string    `json:"change_id" db:"change_id"`
	ReviewerID   string    `json:"reviewer_id" db:"reviewer_id"`
	ReviewerName string    `json:"reviewer_name" db:"reviewer_name"`
	ReviewerRole string    `json:"reviewer_role" db:"reviewer_role"`
	Vote         string    `json:"vote" db:"vote"` // APPROVED, REJECTED, ABSTAIN
	Comments     *string   `json:"comments,omitempty" db:"comments"`
	ReviewedAt   time.Time `json:"reviewed_at" db:"reviewed_at"`
}

// CreateChangePayload is the request DTO for creating a new Change Request.
type CreateChangePayload struct {
	Title              string     `json:"title"`
	Description        string     `json:"description"`
	ChangeType         string     `json:"change_type"` // STANDARD, NORMAL, EMERGENCY, MAJOR
	Category           string     `json:"category"`
	Priority           string     `json:"priority"`
	ImpactLevel        string     `json:"impact_level"`      // LOW, MEDIUM, HIGH, CRITICAL
	ProbabilityLevel   string     `json:"probability_level"` // LOW, MEDIUM, HIGH, CRITICAL
	RequesterID        string     `json:"requester_id"`
	RequesterName      string     `json:"requester_name"`
	RequesterEmail     string     `json:"requester_email"`
	AssignedToID       *string    `json:"assigned_to_id,omitempty"`
	AssignedToName     *string    `json:"assigned_to_name,omitempty"`
	ReasonForChange    string     `json:"reason_for_change"`
	ImplementationPlan string     `json:"implementation_plan"`
	RollbackPlan       string     `json:"rollback_plan"`
	TestPlan           string     `json:"test_plan"`
	ScheduledStartTime *time.Time `json:"scheduled_start_time,omitempty"`
	ScheduledEndTime   *time.Time `json:"scheduled_end_time,omitempty"`
	DowntimeRequired   bool       `json:"downtime_required"`
	DowntimeMinutes    int        `json:"downtime_minutes"`
}

// UpdateChangeStatusPayload is the request DTO for updating Change status.
type UpdateChangeStatusPayload struct {
	Status string  `json:"status"`
	Notes  *string `json:"notes,omitempty"`
}

// SubmitCABVotePayload is the request DTO for submitting a CAB vote.
type SubmitCABVotePayload struct {
	ReviewerID   string  `json:"reviewer_id"`
	ReviewerName string  `json:"reviewer_name"`
	ReviewerRole string  `json:"reviewer_role"`
	Vote         string  `json:"vote"` // APPROVED, REJECTED, ABSTAIN
	Comments     *string `json:"comments,omitempty"`
}

// ChangeCalendarItem represents a scheduled maintenance window item.
type ChangeCalendarItem struct {
	ID               string     `json:"id"`
	ChangeNumber     string     `json:"change_number"`
	Title            string     `json:"title"`
	ChangeType       string     `json:"change_type"`
	Category         string     `json:"category"`
	RiskLevel        string     `json:"risk_level"`
	Status           string     `json:"status"`
	ScheduledStart   *time.Time `json:"scheduled_start"`
	ScheduledEnd     *time.Time `json:"scheduled_end"`
	DowntimeRequired bool       `json:"downtime_required"`
	DowntimeMinutes  int        `json:"downtime_minutes"`
}

// ChangeStats contains KPIs for Change Management dashboard.
type ChangeStats struct {
	ActiveChanges      int     `json:"active_changes"`
	PendingCABReview   int     `json:"pending_cab_review"`
	EmergencyChanges   int     `json:"emergency_changes"`
	SuccessRatePercent float64 `json:"success_rate_percent"`
	TotalThisMonth     int     `json:"total_this_month"`
}
