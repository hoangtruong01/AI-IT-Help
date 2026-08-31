package model

import (
	"time"
)

// Ticket priorities
const (
	PriorityUrgent = "URGENT"
	PriorityHigh   = "HIGH"
	PriorityMedium = "MEDIUM"
	PriorityLow    = "LOW"
)

// Ticket statuses (ITIL lifecycle)
const (
	StatusOpen        = "OPEN"
	StatusAssigned    = "ASSIGNED"
	StatusInProgress  = "IN_PROGRESS"
	StatusWaitingUser = "WAITING_USER"
	StatusResolved    = "RESOLVED"
	StatusClosed      = "CLOSED"
)

// SLA status indicators
const (
	SLAWithinSLA = "WITHIN_SLA"
	SLAWarning   = "WARNING"
	SLABreached  = "BREACHED"
)

// ValidTicketTransitions defines ITIL v4 compliant ticket lifecycle state transitions
var ValidTicketTransitions = map[string][]string{
	StatusOpen:        {StatusAssigned, StatusInProgress, StatusClosed},
	StatusAssigned:    {StatusInProgress, StatusWaitingUser, StatusOpen},
	StatusInProgress:  {StatusWaitingUser, StatusResolved, StatusAssigned},
	StatusWaitingUser: {StatusInProgress, StatusResolved},
	StatusResolved:    {StatusClosed, StatusInProgress},
	StatusClosed:      {}, // Terminal State - No further transitions allowed
}

// IsValidTransition checks if moving from one status to another is permitted
func IsValidTransition(from, to string) bool {
	if from == to {
		return true
	}
	allowed, exists := ValidTicketTransitions[from]
	if !exists {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// Ticket entity
type Ticket struct {
	ID                    string     `json:"id"`
	TicketNumber          string     `json:"ticket_number"`
	Title                 string     `json:"title"`
	Description           string     `json:"description"`
	ServiceItemID         *string    `json:"service_item_id,omitempty"`
	Category              string     `json:"category"`
	Priority              string     `json:"priority"`
	Status                string     `json:"status"`
	RequesterID           string     `json:"requester_id"`
	RequesterName         string     `json:"requester_name"`
	RequesterEmail        string     `json:"requester_email"`
	AssigneeID            *string    `json:"assignee_id,omitempty"`
	AssigneeName          *string    `json:"assignee_name,omitempty"`
	DepartmentID          *string    `json:"department_id,omitempty"`
	AffectedCIID          *string    `json:"affected_ci_id,omitempty"`
	SLAResponseDeadline   time.Time  `json:"sla_response_deadline"`
	SLAResolutionDeadline time.Time  `json:"sla_resolution_deadline"`
	RespondedAt           *time.Time `json:"responded_at,omitempty"`
	ResolvedAt            *time.Time `json:"resolved_at,omitempty"`
	ClosedAt              *time.Time `json:"closed_at,omitempty"`
	SLAStatus             string     `json:"sla_status"`
	Version               int        `json:"version"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

// CreateTicketRequest DTO
type CreateTicketRequest struct {
	Title          string  `json:"title"`
	Description    string  `json:"description"`
	ServiceItemID  *string `json:"service_item_id,omitempty"`
	Category       string  `json:"category"`
	Priority       string  `json:"priority"`
	RequesterID    string  `json:"requester_id"`
	RequesterName  string  `json:"requester_name"`
	RequesterEmail string  `json:"requester_email"`
	DepartmentID   *string `json:"department_id,omitempty"`
	AffectedCIID   *string `json:"affected_ci_id,omitempty"`
}

// UpdateTicketStatusRequest DTO
type UpdateTicketStatusRequest struct {
	Status       string  `json:"status"`
	Notes        string  `json:"notes,omitempty"`
	AssigneeID   *string `json:"assignee_id,omitempty"`
	AssigneeName *string `json:"assignee_name,omitempty"`
	Version      *int    `json:"version"`
}

// AssignTicketRequest DTO
type AssignTicketRequest struct {
	AssigneeID   string `json:"assignee_id"`
	AssigneeName string `json:"assignee_name"`
	Version      *int   `json:"version"`
}

// TicketComment entity
type TicketComment struct {
	ID         string    `json:"id"`
	TicketID   string    `json:"ticket_id"`
	AuthorID   string    `json:"author_id"`
	AuthorName string    `json:"author_name"`
	AuthorRole string    `json:"author_role"`
	Content    string    `json:"content"`
	IsInternal bool      `json:"is_internal"`
	CreatedAt  time.Time `json:"created_at"`
}

// AddCommentRequest DTO
type AddCommentRequest struct {
	Content    string `json:"content"`
	IsInternal bool   `json:"is_internal"`
}

// TicketTimeline audit record
type TicketTimeline struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticket_id"`
	ActorID   string    `json:"actor_id"`
	ActorName string    `json:"actor_name"`
	Action    string    `json:"action"`
	OldValue  *string   `json:"old_value,omitempty"`
	NewValue  *string   `json:"new_value,omitempty"`
	Notes     *string   `json:"notes,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// TicketListQuery parameters
type TicketListQuery struct {
	Page              int    `json:"page"`
	PageSize          int    `json:"page_size"`
	Status            string `json:"status"`
	Priority          string `json:"priority"`
	Category          string `json:"category"`
	AssigneeID        string `json:"assignee_id"`
	RequesterID       string `json:"requester_id"`
	DepartmentID      string `json:"department_id"`
	Search            string `json:"search"`
	ActorRole         string `json:"actor_role"`
	ActorID           string `json:"actor_id"`
	ActorDepartmentID string `json:"actor_department_id"`
}

// TicketListResponse paginated envelope
type TicketListResponse struct {
	Data       []Ticket `json:"data"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}
