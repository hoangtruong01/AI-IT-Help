package model

import "time"

// Problem represents an ITIL Problem record aggregating multiple incidents.
type Problem struct {
	ID            string     `json:"id" db:"id"`
	ProblemNumber string     `json:"problem_number" db:"problem_number"`
	Title         string     `json:"title" db:"title"`
	Description   string     `json:"description" db:"description"`
	Category      string     `json:"category" db:"category"`
	Priority      string     `json:"priority" db:"priority"`
	Status        string     `json:"status" db:"status"` // OPEN, UNDER_INVESTIGATION, WORKAROUND_FOUND, KNOWN_ERROR, RESOLVED, CLOSED
	Impact        string     `json:"impact" db:"impact"`
	Urgency       string     `json:"urgency" db:"urgency"`
	AssigneeID    *string    `json:"assignee_id,omitempty" db:"assignee_id"`
	AssigneeName  *string    `json:"assignee_name,omitempty" db:"assignee_name"`
	RootCause     *string    `json:"root_cause,omitempty" db:"root_cause"`
	Workaround    *string    `json:"workaround,omitempty" db:"workaround"`
	Resolution    *string    `json:"resolution,omitempty" db:"resolution"`
	IsKnownError  bool       `json:"is_known_error" db:"is_known_error"`
	LinkedCount   int        `json:"linked_count" db:"linked_count"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty" db:"resolved_at"`
	ClosedAt      *time.Time `json:"closed_at,omitempty" db:"closed_at"`
}

// ProblemIncidentLink represents a link between a problem and an incident ticket.
type ProblemIncidentLink struct {
	ID           string    `json:"id" db:"id"`
	ProblemID    string    `json:"problem_id" db:"problem_id"`
	TicketID     string    `json:"ticket_id" db:"ticket_id"`
	TicketNumber string    `json:"ticket_number" db:"ticket_number"`
	TicketTitle  string    `json:"ticket_title" db:"ticket_title"`
	LinkedBy     string    `json:"linked_by" db:"linked_by"`
	LinkedAt     time.Time `json:"linked_at" db:"linked_at"`
}

// CreateProblemPayload is the request DTO for creating a new Problem.
type CreateProblemPayload struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Category     string   `json:"category"`
	Priority     string   `json:"priority"`
	Impact       string   `json:"impact"`
	Urgency      string   `json:"urgency"`
	AssigneeID   *string  `json:"assignee_id,omitempty"`
	AssigneeName *string  `json:"assignee_name,omitempty"`
	RootCause    *string  `json:"root_cause,omitempty"`
	Workaround   *string  `json:"workaround,omitempty"`
	IsKnownError bool     `json:"is_known_error"`
	TicketIDs    []string `json:"ticket_ids,omitempty"`
}

// UpdateProblemStatusPayload is the request DTO for updating a Problem's status.
type UpdateProblemStatusPayload struct {
	Status     string  `json:"status"` // OPEN, UNDER_INVESTIGATION, WORKAROUND_FOUND, KNOWN_ERROR, RESOLVED, CLOSED
	Resolution *string `json:"resolution,omitempty"`
	Notes      *string `json:"notes,omitempty"`
}

// UpdateProblemRCAPayload is the request DTO for updating Root Cause and Workaround.
type UpdateProblemRCAPayload struct {
	RootCause    *string `json:"root_cause,omitempty"`
	Workaround   *string `json:"workaround,omitempty"`
	IsKnownError bool    `json:"is_known_error"`
}

// LinkIncidentPayload is the request DTO for linking an incident to a problem.
type LinkIncidentPayload struct {
	TicketID string `json:"ticket_id"`
	LinkedBy string `json:"linked_by"`
}

// ProblemStats contains aggregated KPIs for the Problem Management dashboard.
type ProblemStats struct {
	TotalProblems      int `json:"total_problems"`
	UnderInvestigation int `json:"under_investigation"`
	KnownErrors        int `json:"known_errors"`
	ResolvedProblems   int `json:"resolved_problems"`
	TotalLinkedTickets int `json:"total_linked_tickets"`
}
