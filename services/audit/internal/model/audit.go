package model

import "time"

// AuditLog represents an immutable audit trail entry.
type AuditLog struct {
	ID                string                 `json:"id"`
	EventType         string                 `json:"event_type"` // AUTH_LOGIN_SUCCESS, TICKET_UPDATE, ASSET_DELETE, APPROVAL_DECISION, ROLE_CHANGE, CONFIG_UPDATE
	ActorID           string                 `json:"actor_id"`
	ActorName         string                 `json:"actor_name"`
	ActorEmail        string                 `json:"actor_email"`
	ActorRole         string                 `json:"actor_role"` // ROLE_ADMIN, ROLE_MANAGER, ROLE_AGENT, ROLE_EMPLOYEE
	ServiceName       string                 `json:"service_name"`
	IPAddress         string                 `json:"ip_address"`
	UserAgent         string                 `json:"user_agent,omitempty"`
	Status            string                 `json:"status"` // SUCCESS, FORBIDDEN, FAILED
	ResourceType      string                 `json:"resource_type"`
	ResourceID        string                 `json:"resource_id"`
	OldValues         map[string]interface{} `json:"old_values,omitempty"`
	NewValues         map[string]interface{} `json:"new_values,omitempty"`
	PreviousChecksum  string                 `json:"previous_checksum"`
	ChecksumSHA256    string                 `json:"checksum_sha256"`
	ChecksumAlgorithm string                 `json:"checksum_algorithm"`
	CreatedAt         time.Time              `json:"created_at"`
}

// IntegrityReport is the result of verifying the ordered audit hash chain.
type IntegrityReport struct {
	Valid          bool   `json:"valid"`
	TotalLogs      int64  `json:"total_logs"`
	VerifiedLogs   int64  `json:"verified_logs"`
	LegacyLogs     int64  `json:"legacy_logs"`
	FirstInvalidID string `json:"first_invalid_id,omitempty"`
	Message        string `json:"message"`
}

// SecurityEvent represents a security warning or blocked violation.
type SecurityEvent struct {
	ID             string    `json:"id"`
	EventCode      string    `json:"event_code"`
	Severity       string    `json:"severity"` // LOW, MEDIUM, HIGH, CRITICAL
	SourceIP       string    `json:"source_ip"`
	TargetEndpoint string    `json:"target_endpoint"`
	Description    string    `json:"description"`
	IsBlocked      bool      `json:"is_blocked"`
	CreatedAt      time.Time `json:"created_at"`
}

// AuditStats captures high-level compliance and security KPIs.
type AuditStats struct {
	TotalLogs            int64 `json:"total_logs"`
	BlockedViolations    int64 `json:"blocked_violations"`
	ActiveSecurityAlerts int64 `json:"active_security_alerts"`
	ImmutableProofsCount int64 `json:"immutable_proofs_count"`
	SuccessCount         int64 `json:"success_count"`
	ForbiddenCount       int64 `json:"forbidden_count"`
}

// AuditFilterQuery captures search and filter parameters.
type AuditFilterQuery struct {
	EventType string `json:"event_type,omitempty"`
	Status    string `json:"status,omitempty"`
	Service   string `json:"service,omitempty"`
	Actor     string `json:"actor,omitempty"`
	Search    string `json:"search,omitempty"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

// CreateAuditLogRequest defines the payload for creating an audit record.
type CreateAuditLogRequest struct {
	EventType    string                 `json:"event_type"`
	ActorID      string                 `json:"actor_id,omitempty"`
	ActorName    string                 `json:"actor_name"`
	ActorEmail   string                 `json:"actor_email"`
	ActorRole    string                 `json:"actor_role"`
	ServiceName  string                 `json:"service_name"`
	IPAddress    string                 `json:"ip_address"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Status       string                 `json:"status"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   string                 `json:"resource_id"`
	OldValues    map[string]interface{} `json:"old_values,omitempty"`
	NewValues    map[string]interface{} `json:"new_values,omitempty"`
}
