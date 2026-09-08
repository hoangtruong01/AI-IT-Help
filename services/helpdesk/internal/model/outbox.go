package model

import (
	"time"
)

// Outbox status constants
const (
	OutboxStatusPending   = "PENDING"
	OutboxStatusPublished = "PUBLISHED"
	OutboxStatusFailed    = "FAILED"
)

// OutboxEvent represents a domain event queued within a local database transaction.
type OutboxEvent struct {
	ID          string     `json:"id"`
	EventType   string     `json:"event_type"`
	Source      string     `json:"source"`
	Payload     string     `json:"payload"` // Raw JSON string
	Status      string     `json:"status"`
	RetryCount  int        `json:"retry_count"`
	LastError   *string    `json:"last_error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
}
