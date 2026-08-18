package model

import (
	"time"
)

// Notification Categories
const (
	CategoryIncident = "INCIDENT"
	CategoryApproval = "APPROVAL"
	CategoryAsset    = "ASSET"
	CategorySLA      = "SLA"
	CategorySecurity = "SECURITY"
	CategorySystem   = "SYSTEM"
)

// Notification Priorities
const (
	PriorityLow    = "LOW"
	PriorityMedium = "MEDIUM"
	PriorityHigh   = "HIGH"
	PriorityUrgent = "URGENT"
)

// Notification Channels
const (
	ChannelInApp   = "IN_APP"
	ChannelEmail   = "EMAIL"
	ChannelSlack   = "SLACK"
	ChannelWebhook = "WEBHOOK"
)

// Notification entity
type Notification struct {
	ID             string     `json:"id"`
	RecipientID    string     `json:"recipient_id"`
	RecipientEmail string     `json:"recipient_email"`
	Title          string     `json:"title"`
	Message        string     `json:"message"`
	Category       string     `json:"category"`
	Priority       string     `json:"priority"`
	IsRead         bool       `json:"is_read"`
	ReadAt         *time.Time `json:"read_at,omitempty"`
	Channel        string     `json:"channel"`
	Metadata       *string    `json:"metadata,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// CreateNotificationRequest DTO
type CreateNotificationRequest struct {
	RecipientID    string  `json:"recipient_id"`
	RecipientEmail string  `json:"recipient_email"`
	Title          string  `json:"title"`
	Message        string  `json:"message"`
	Category       string  `json:"category"`
	Priority       string  `json:"priority"`
	Channel        string  `json:"channel"`
	Metadata       *string `json:"metadata,omitempty"`
}

// NotificationListQuery parameters
type NotificationListQuery struct {
	RecipientID string `json:"recipient_id"`
	IsRead      *bool  `json:"is_read,omitempty"`
	Category    string `json:"category"`
	Page        int    `json:"page"`
	PageSize    int    `json:"page_size"`
}

// NotificationListResponse envelope
type NotificationListResponse struct {
	Data        []Notification `json:"data"`
	Total       int            `json:"total"`
	UnreadCount int            `json:"unread_count"`
	Page        int            `json:"page"`
	PageSize    int            `json:"page_size"`
	TotalPages  int            `json:"total_pages"`
}

// NotificationStats summary
type NotificationStats struct {
	Total     int `json:"total"`
	Unread    int `json:"unread"`
	Incidents int `json:"incidents"`
	Approvals int `json:"approvals"`
}
