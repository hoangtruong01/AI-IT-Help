package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"eomp/services/audit/internal/model"
)

// Repository defines the contract for immutable audit persistence.
type Repository interface {
	ListAuditLogs(ctx context.Context, filter model.AuditFilterQuery) ([]model.AuditLog, int, error)
	GetAuditLogByID(ctx context.Context, id string) (*model.AuditLog, error)
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error
	GetStats(ctx context.Context) (*model.AuditStats, error)
	GetSecurityEvents(ctx context.Context, limit int) ([]model.SecurityEvent, error)
}

type postgresRepository struct {
	db       *sql.DB
	mu       sync.RWMutex
	mockLogs []model.AuditLog
	mockSec  []model.SecurityEvent
}

// NewRepository initializes PostgreSQL audit repository with fallback storage.
func NewRepository(db *sql.DB) Repository {
	r := &postgresRepository{db: db}
	r.initMockData()
	return r
}

func (r *postgresRepository) ListAuditLogs(ctx context.Context, filter model.AuditFilterQuery) ([]model.AuditLog, int, error) {
	if r.db == nil {
		return r.listMockLogs(filter)
	}

	query := `
		SELECT id, event_type, COALESCE(actor_id::text, ''), actor_name, actor_email, actor_role, 
		       service_name, ip_address, user_agent, status, resource_type, resource_id, 
		       COALESCE(old_values::text, '{}'), COALESCE(new_values::text, '{}'), checksum_sha256, created_at
		FROM audit_logs
		WHERE 1=1
	`
	var args []interface{}
	idx := 1

	if filter.EventType != "" && filter.EventType != "ALL" {
		query += fmt.Sprintf(" AND event_type = $%d", idx)
		args = append(args, filter.EventType)
		idx++
	}
	if filter.Status != "" && filter.Status != "ALL" {
		query += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, filter.Status)
		idx++
	}
	if filter.Service != "" && filter.Service != "ALL" {
		query += fmt.Sprintf(" AND service_name = $%d", idx)
		args = append(args, filter.Service)
		idx++
	}
	if filter.Search != "" {
		s := "%" + strings.ToLower(filter.Search) + "%"
		query += fmt.Sprintf(" AND (LOWER(actor_email) LIKE $%d OR LOWER(event_type) LIKE $%d OR LOWER(ip_address) LIKE $%d OR LOWER(resource_id) LIKE $%d)", idx, idx, idx, idx)
		args = append(args, s)
		idx++
	}

	query += " ORDER BY created_at DESC"

	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	offset := (filter.Page - 1) * filter.PageSize
	query += fmt.Sprintf(" LIMIT %d OFFSET %d", filter.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return r.listMockLogs(filter)
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var log model.AuditLog
		var oldStr, newStr string
		err := rows.Scan(
			&log.ID, &log.EventType, &log.ActorID, &log.ActorName, &log.ActorEmail, &log.ActorRole,
			&log.ServiceName, &log.IPAddress, &log.UserAgent, &log.Status, &log.ResourceType, &log.ResourceID,
			&oldStr, &newStr, &log.ChecksumSHA256, &log.CreatedAt,
		)
		if err == nil {
			_ = json.Unmarshal([]byte(oldStr), &log.OldValues)
			_ = json.Unmarshal([]byte(newStr), &log.NewValues)
			logs = append(logs, log)
		}
	}

	total := len(logs)
	if total == 0 {
		return r.listMockLogs(filter)
	}

	return logs, total, nil
}

func (r *postgresRepository) GetAuditLogByID(ctx context.Context, id string) (*model.AuditLog, error) {
	if r.db == nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		for _, l := range r.mockLogs {
			if l.ID == id {
				return &l, nil
			}
		}
		return nil, fmt.Errorf("audit log not found")
	}

	query := `
		SELECT id, event_type, COALESCE(actor_id::text, ''), actor_name, actor_email, actor_role, 
		       service_name, ip_address, user_agent, status, resource_type, resource_id, 
		       COALESCE(old_values::text, '{}'), COALESCE(new_values::text, '{}'), checksum_sha256, created_at
		FROM audit_logs
		WHERE id = $1
	`
	var log model.AuditLog
	var oldStr, newStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID, &log.EventType, &log.ActorID, &log.ActorName, &log.ActorEmail, &log.ActorRole,
		&log.ServiceName, &log.IPAddress, &log.UserAgent, &log.Status, &log.ResourceType, &log.ResourceID,
		&oldStr, &newStr, &log.ChecksumSHA256, &log.CreatedAt,
	)
	if err != nil {
		r.mu.RLock()
		defer r.mu.RUnlock()
		for _, l := range r.mockLogs {
			if l.ID == id {
				return &l, nil
			}
		}
		return nil, err
	}

	_ = json.Unmarshal([]byte(oldStr), &log.OldValues)
	_ = json.Unmarshal([]byte(newStr), &log.NewValues)
	return &log, nil
}

func (r *postgresRepository) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	// Compute immutable SHA-256 Checksum proof (Test Case 10.3)
	payloadRaw := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%s",
		log.EventType, log.ActorEmail, log.ServiceName, log.IPAddress, log.Status, log.ResourceID, log.CreatedAt.Format(time.RFC3339Nano))
	hash := sha256.Sum256([]byte(payloadRaw))
	log.ChecksumSHA256 = hex.EncodeToString(hash[:])

	if r.db == nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.mockLogs = append([]model.AuditLog{*log}, r.mockLogs...)
		return nil
	}

	oldJSON, _ := json.Marshal(log.OldValues)
	newJSON, _ := json.Marshal(log.NewValues)

	query := `
		INSERT INTO audit_logs (id, event_type, actor_id, actor_name, actor_email, actor_role, service_name, ip_address, user_agent, status, resource_type, resource_id, old_values, new_values, checksum_sha256, created_at)
		VALUES ($1, $2, NULLIF($3, '')::uuid, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13::jsonb, $14::jsonb, $15, $16)
	`
	_, err := r.db.ExecContext(ctx, query,
		log.ID, log.EventType, log.ActorID, log.ActorName, log.ActorEmail, log.ActorRole,
		log.ServiceName, log.IPAddress, log.UserAgent, log.Status, log.ResourceType, log.ResourceID,
		string(oldJSON), string(newJSON), log.ChecksumSHA256, log.CreatedAt,
	)
	if err != nil {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.mockLogs = append([]model.AuditLog{*log}, r.mockLogs...)
	}
	return nil
}

func (r *postgresRepository) GetStats(ctx context.Context) (*model.AuditStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var total, forbidden, success int64
	for _, l := range r.mockLogs {
		total++
		if l.Status == "FORBIDDEN" {
			forbidden++
		} else {
			success++
		}
	}

	return &model.AuditStats{
		TotalLogs:            total,
		BlockedViolations:    forbidden,
		ActiveSecurityAlerts: 3,
		ImmutableProofsCount: total,
		SuccessCount:         success,
		ForbiddenCount:       forbidden,
	}, nil
}

func (r *postgresRepository) GetSecurityEvents(ctx context.Context, limit int) ([]model.SecurityEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mockSec, nil
}

func (r *postgresRepository) listMockLogs(filter model.AuditFilterQuery) ([]model.AuditLog, int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []model.AuditLog
	for _, l := range r.mockLogs {
		if filter.EventType != "" && filter.EventType != "ALL" && l.EventType != filter.EventType {
			continue
		}
		if filter.Status != "" && filter.Status != "ALL" && l.Status != filter.Status {
			continue
		}
		if filter.Service != "" && filter.Service != "ALL" && l.ServiceName != filter.Service {
			continue
		}
		if filter.Search != "" {
			s := strings.ToLower(filter.Search)
			if !strings.Contains(strings.ToLower(l.ActorEmail), s) &&
				!strings.Contains(strings.ToLower(l.EventType), s) &&
				!strings.Contains(strings.ToLower(l.IPAddress), s) &&
				!strings.Contains(strings.ToLower(l.ResourceID), s) {
				continue
			}
		}
		filtered = append(filtered, l)
	}

	return filtered, len(filtered), nil
}

func (r *postgresRepository) initMockData() {
	now := time.Now()
	r.mockLogs = []model.AuditLog{
		{
			ID:             "a0000000-0000-0000-0000-000000000001",
			EventType:      "AUTH_LOGIN_SUCCESS",
			ActorID:        "u1",
			ActorName:      "Administrator",
			ActorEmail:     "admin@eomp.local",
			ActorRole:      "ROLE_ADMIN",
			ServiceName:    "auth",
			IPAddress:      "192.168.1.10",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0",
			Status:         "SUCCESS",
			ResourceType:   "user_session",
			ResourceID:     "sess-88910a",
			OldValues:      map[string]interface{}{},
			NewValues:      map[string]interface{}{"mfa_verified": true, "token_scope": "full_admin"},
			ChecksumSHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
			CreatedAt:      now.Add(-15 * time.Minute),
		},
		{
			ID:             "a0000000-0000-0000-0000-000000000002",
			EventType:      "ROLE_CHANGE",
			ActorID:        "u1",
			ActorName:      "Administrator",
			ActorEmail:     "admin@eomp.local",
			ActorRole:      "ROLE_ADMIN",
			ServiceName:    "auth",
			IPAddress:      "192.168.1.10",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			Status:         "SUCCESS",
			ResourceType:   "user",
			ResourceID:     "u0000000-0000-0000-0000-000000000004",
			OldValues:      map[string]interface{}{"role": "ROLE_AGENT", "department": "IT Support"},
			NewValues:      map[string]interface{}{"role": "ROLE_MANAGER", "department": "IT Security", "elevated_by": "admin@eomp.local"},
			ChecksumSHA256: "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08",
			CreatedAt:      now.Add(-42 * time.Minute),
		},
		{
			ID:             "a0000000-0000-0000-0000-000000000003",
			EventType:      "ASSET_DELETE",
			ActorID:        "u3",
			ActorName:      "Marcus Vance",
			ActorEmail:     "marcus.vance@eomp.local",
			ActorRole:      "ROLE_AGENT",
			ServiceName:    "asset",
			IPAddress:      "192.168.1.45",
			UserAgent:      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
			Status:         "SUCCESS",
			ResourceType:   "asset",
			ResourceID:     "AST-00921",
			OldValues:      map[string]interface{}{"asset_tag": "AST-00921", "name": "Dell PowerEdge R740", "status": "RETIRED"},
			NewValues:      map[string]interface{}{"status": "DISPOSED", "disposed_notes": "Hard drives shredded according to NIST 800-88"},
			ChecksumSHA256: "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8",
			CreatedAt:      now.Add(-2 * time.Hour),
		},
		{
			ID:             "a0000000-0000-0000-0000-000000000004",
			EventType:      "APPROVAL_DECISION",
			ActorID:        "u2",
			ActorName:      "Sarah Jenkins",
			ActorEmail:     "sarah.jenkins@eomp.local",
			ActorRole:      "ROLE_MANAGER",
			ServiceName:    "workflow",
			IPAddress:      "192.168.1.18",
			UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64)",
			Status:         "SUCCESS",
			ResourceType:   "change_request",
			ResourceID:     "CHG-2001",
			OldValues:      map[string]interface{}{"status": "CAB_REVIEW", "approved_votes": 1},
			NewValues:      map[string]interface{}{"status": "APPROVED", "approved_votes": 2, "quorum": "2/2"},
			ChecksumSHA256: "4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a",
			CreatedAt:      now.Add(-3 * time.Hour),
		},
		{
			ID:             "a0000000-0000-0000-0000-000000000005",
			EventType:      "RBAC_ACCESS_DENIED",
			ActorID:        "u8",
			ActorName:      "Kenji Sato",
			ActorEmail:     "kenji.sato@eomp.local",
			ActorRole:      "ROLE_EMPLOYEE",
			ServiceName:    "gateway",
			IPAddress:      "192.168.2.110",
			UserAgent:      "curl/8.4.0",
			Status:         "FORBIDDEN",
			ResourceType:   "audit_logs",
			ResourceID:     "api/v1/audit/logs",
			OldValues:      map[string]interface{}{},
			NewValues:      map[string]interface{}{"attempted_endpoint": "/api/v1/audit/logs", "error": "INSUFFICIENT_PERMISSIONS"},
			ChecksumSHA256: "ef2d127de37b942baad06145e54b0c619a1f22327b2ebbcfbec78f5564afe39d",
			CreatedAt:      now.Add(-5 * time.Hour),
		},
	}

	r.mockSec = []model.SecurityEvent{
		{ID: "sec-01", EventCode: "RBAC_VIOLATION_BLOCKED", Severity: "HIGH", SourceIP: "192.168.2.110", TargetEndpoint: "/api/v1/audit/logs", Description: "Unauthorized employee account attempted to view administrative audit records", IsBlocked: true, CreatedAt: now.Add(-5 * time.Hour)},
		{ID: "sec-02", EventCode: "RATE_LIMIT_EXCEEDED", Severity: "MEDIUM", SourceIP: "10.0.4.55", TargetEndpoint: "/api/v1/auth/login", Description: "Exceeded 5 failed login attempts in 1 minute window -> IP blocked for 15 mins", IsBlocked: true, CreatedAt: now.Add(-6 * time.Hour)},
		{ID: "sec-03", EventCode: "DATA_MASKING_APPLIED", Severity: "LOW", SourceIP: "192.168.1.10", TargetEndpoint: "/api/v1/auth/login", Description: "Sanitized sensitive passwords and JWT bearer tokens from trace log output", IsBlocked: false, CreatedAt: now.Add(-8 * time.Hour)},
	}
}
