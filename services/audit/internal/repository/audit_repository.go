package repository

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"eomp/services/audit/internal/model"
)

const (
	checksumAlgorithm = "HMAC-SHA256"
	chainGenesis      = "0000000000000000000000000000000000000000000000000000000000000000"
)

// Repository defines the contract for append-only audit persistence.
type Repository interface {
	ListAuditLogs(ctx context.Context, filter model.AuditFilterQuery) ([]model.AuditLog, int, error)
	GetAuditLogByID(ctx context.Context, id string) (*model.AuditLog, error)
	CreateAuditLog(ctx context.Context, log *model.AuditLog) error
	GetStats(ctx context.Context) (*model.AuditStats, error)
	GetSecurityEvents(ctx context.Context, limit int) ([]model.SecurityEvent, error)
	VerifyIntegrity(ctx context.Context) (*model.IntegrityReport, error)
}

type postgresRepository struct {
	db      *sql.DB
	hmacKey []byte
}

// NewRepository initializes the PostgreSQL audit repository.
func NewRepository(db *sql.DB, hmacKey string) Repository {
	return &postgresRepository{db: db, hmacKey: []byte(hmacKey)}
}

// ComputeHMAC seals the canonical, masked audit payload and its predecessor.
func ComputeHMAC(log *model.AuditLog, key []byte) error {
	if len(key) < 32 {
		return fmt.Errorf("audit HMAC key must contain at least 32 bytes")
	}
	// PostgreSQL timestamptz persists microsecond precision. Normalize before
	// signing so a record verifies identically after a database round trip.
	log.CreatedAt = log.CreatedAt.UTC().Truncate(time.Microsecond)
	oldJSON, err := json.Marshal(log.OldValues)
	if err != nil {
		return fmt.Errorf("marshal old audit values: %w", err)
	}
	newJSON, err := json.Marshal(log.NewValues)
	if err != nil {
		return fmt.Errorf("marshal new audit values: %w", err)
	}
	payload, err := json.Marshal(struct {
		ID, EventType, ActorID, ActorName, ActorEmail, ActorRole string
		ServiceName, IPAddress, UserAgent, Status                string
		ResourceType, ResourceID, PreviousChecksum               string
		OldValues, NewValues                                     json.RawMessage
		CreatedAt                                                time.Time
	}{
		log.ID, log.EventType, log.ActorID, log.ActorName, log.ActorEmail, log.ActorRole,
		log.ServiceName, log.IPAddress, log.UserAgent, log.Status, log.ResourceType, log.ResourceID,
		log.PreviousChecksum, oldJSON, newJSON, log.CreatedAt.UTC(),
	})
	if err != nil {
		return fmt.Errorf("marshal audit HMAC payload: %w", err)
	}
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(payload)
	log.ChecksumSHA256 = hex.EncodeToString(mac.Sum(nil))
	log.ChecksumAlgorithm = checksumAlgorithm
	return nil
}

func (r *postgresRepository) ListAuditLogs(ctx context.Context, filter model.AuditFilterQuery) ([]model.AuditLog, int, error) {
	if r.db == nil {
		return nil, 0, fmt.Errorf("audit database is unavailable")
	}
	query := `
		SELECT id, event_type, COALESCE(actor_id::text, ''), actor_name, actor_email, actor_role,
		       service_name, ip_address, user_agent, status, resource_type, resource_id,
		       COALESCE(old_values::text, '{}'), COALESCE(new_values::text, '{}'),
		       previous_checksum, checksum_sha256, checksum_algorithm, created_at,
		       COUNT(*) OVER()
		FROM audit_logs WHERE 1=1`
	args := make([]any, 0)
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
	if filter.Actor != "" {
		query += fmt.Sprintf(" AND LOWER(actor_email) = LOWER($%d)", idx)
		args = append(args, filter.Actor)
		idx++
	}
	if filter.Search != "" {
		query += fmt.Sprintf(" AND (LOWER(actor_email) LIKE $%d OR LOWER(event_type) LIKE $%d OR LOWER(ip_address) LIKE $%d OR LOWER(resource_id) LIKE $%d)", idx, idx, idx, idx)
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
	}
	if filter.PageSize <= 0 || filter.PageSize > 200 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	query += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT %d OFFSET %d", filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("query audit logs: %w", err)
	}
	defer rows.Close()

	logs := make([]model.AuditLog, 0)
	total := int64(0)
	for rows.Next() {
		var log model.AuditLog
		var oldJSON, newJSON string
		if err := rows.Scan(
			&log.ID, &log.EventType, &log.ActorID, &log.ActorName, &log.ActorEmail, &log.ActorRole,
			&log.ServiceName, &log.IPAddress, &log.UserAgent, &log.Status, &log.ResourceType, &log.ResourceID,
			&oldJSON, &newJSON, &log.PreviousChecksum, &log.ChecksumSHA256, &log.ChecksumAlgorithm,
			&log.CreatedAt, &total,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		if err := decodeValues(oldJSON, newJSON, &log); err != nil {
			return nil, 0, err
		}
		logs = append(logs, log)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audit logs: %w", err)
	}
	return logs, int(total), nil
}

func (r *postgresRepository) GetAuditLogByID(ctx context.Context, id string) (*model.AuditLog, error) {
	if r.db == nil {
		return nil, fmt.Errorf("audit database is unavailable")
	}
	var log model.AuditLog
	var oldJSON, newJSON string
	err := r.db.QueryRowContext(ctx, `
		SELECT id, event_type, COALESCE(actor_id::text, ''), actor_name, actor_email, actor_role,
		       service_name, ip_address, user_agent, status, resource_type, resource_id,
		       COALESCE(old_values::text, '{}'), COALESCE(new_values::text, '{}'),
		       previous_checksum, checksum_sha256, checksum_algorithm, created_at
		FROM audit_logs WHERE id = $1`, id).Scan(
		&log.ID, &log.EventType, &log.ActorID, &log.ActorName, &log.ActorEmail, &log.ActorRole,
		&log.ServiceName, &log.IPAddress, &log.UserAgent, &log.Status, &log.ResourceType, &log.ResourceID,
		&oldJSON, &newJSON, &log.PreviousChecksum, &log.ChecksumSHA256, &log.ChecksumAlgorithm, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if err := decodeValues(oldJSON, newJSON, &log); err != nil {
		return nil, err
	}
	return &log, nil
}

func decodeValues(oldJSON, newJSON string, log *model.AuditLog) error {
	if err := json.Unmarshal([]byte(oldJSON), &log.OldValues); err != nil {
		return fmt.Errorf("decode old audit values: %w", err)
	}
	if err := json.Unmarshal([]byte(newJSON), &log.NewValues); err != nil {
		return fmt.Errorf("decode new audit values: %w", err)
	}
	return nil
}

func (r *postgresRepository) CreateAuditLog(ctx context.Context, log *model.AuditLog) error {
	if r.db == nil {
		return fmt.Errorf("audit database is unavailable")
	}
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return fmt.Errorf("begin audit transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext('eomp_audit_hmac_chain'))`); err != nil {
		return fmt.Errorf("lock audit chain: %w", err)
	}
	previous := chainGenesis
	err = tx.QueryRowContext(ctx, `SELECT checksum_sha256 FROM audit_logs ORDER BY audit_sequence DESC LIMIT 1`).Scan(&previous)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read audit chain head: %w", err)
	}
	log.PreviousChecksum = previous
	if err := ComputeHMAC(log, r.hmacKey); err != nil {
		return err
	}
	oldJSON, err := json.Marshal(log.OldValues)
	if err != nil {
		return fmt.Errorf("marshal old audit values: %w", err)
	}
	newJSON, err := json.Marshal(log.NewValues)
	if err != nil {
		return fmt.Errorf("marshal new audit values: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO audit_logs (
			id, event_type, actor_id, actor_name, actor_email, actor_role, service_name,
			ip_address, user_agent, status, resource_type, resource_id, old_values, new_values,
			previous_checksum, checksum_sha256, checksum_algorithm, created_at
		) VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6, $7, $8, $9, $10, $11, $12,
		          $13::jsonb, $14::jsonb, $15, $16, $17, $18)`,
		log.ID, log.EventType, log.ActorID, log.ActorName, log.ActorEmail, log.ActorRole,
		log.ServiceName, log.IPAddress, log.UserAgent, log.Status, log.ResourceType, log.ResourceID,
		string(oldJSON), string(newJSON), log.PreviousChecksum, log.ChecksumSHA256,
		log.ChecksumAlgorithm, log.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert audit log: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit audit log: %w", err)
	}
	return nil
}

func (r *postgresRepository) VerifyIntegrity(ctx context.Context) (*model.IntegrityReport, error) {
	if r.db == nil {
		return nil, fmt.Errorf("audit database is unavailable")
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_type, COALESCE(actor_id::text, ''), actor_name, actor_email, actor_role,
		       service_name, ip_address, user_agent, status, resource_type, resource_id,
		       COALESCE(old_values::text, '{}'), COALESCE(new_values::text, '{}'),
		       previous_checksum, checksum_sha256, checksum_algorithm, created_at
		FROM audit_logs ORDER BY audit_sequence ASC`)
	if err != nil {
		return nil, fmt.Errorf("query audit chain: %w", err)
	}
	defer rows.Close()

	report := &model.IntegrityReport{Valid: true, Message: "audit HMAC chain is valid"}
	previous := chainGenesis
	for rows.Next() {
		var log model.AuditLog
		var oldJSON, newJSON string
		if err := rows.Scan(
			&log.ID, &log.EventType, &log.ActorID, &log.ActorName, &log.ActorEmail, &log.ActorRole,
			&log.ServiceName, &log.IPAddress, &log.UserAgent, &log.Status, &log.ResourceType, &log.ResourceID,
			&oldJSON, &newJSON, &log.PreviousChecksum, &log.ChecksumSHA256, &log.ChecksumAlgorithm, &log.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan audit chain: %w", err)
		}
		report.TotalLogs++
		if err := decodeValues(oldJSON, newJSON, &log); err != nil {
			return nil, err
		}
		if log.ChecksumAlgorithm != checksumAlgorithm {
			report.LegacyLogs++
			previous = log.ChecksumSHA256
			continue
		}
		actual := log.ChecksumSHA256
		if log.PreviousChecksum != previous {
			return invalidReport(report, log.ID, "audit chain predecessor does not match"), nil
		}
		if err := ComputeHMAC(&log, r.hmacKey); err != nil {
			return nil, err
		}
		expectedBytes, expectedErr := hex.DecodeString(log.ChecksumSHA256)
		actualBytes, actualErr := hex.DecodeString(actual)
		if expectedErr != nil || actualErr != nil || !hmac.Equal(expectedBytes, actualBytes) {
			return invalidReport(report, log.ID, "audit record HMAC does not match its persisted payload"), nil
		}
		report.VerifiedLogs++
		previous = actual
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate audit chain: %w", err)
	}
	if report.LegacyLogs > 0 {
		report.Message = fmt.Sprintf("HMAC chain is valid; %d legacy SHA-256 records cannot be cryptographically reverified", report.LegacyLogs)
	}
	return report, nil
}

func invalidReport(report *model.IntegrityReport, id, message string) *model.IntegrityReport {
	report.Valid = false
	report.FirstInvalidID = id
	report.Message = message
	return report
}

func (r *postgresRepository) GetStats(ctx context.Context) (*model.AuditStats, error) {
	if r.db == nil {
		return nil, fmt.Errorf("audit database is unavailable")
	}
	var total, forbidden, success, proofs, alerts int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'FORBIDDEN'),
		       COUNT(*) FILTER (WHERE status = 'SUCCESS'),
		       COUNT(*) FILTER (WHERE checksum_algorithm = 'HMAC-SHA256')
		FROM audit_logs`).Scan(&total, &forbidden, &success, &proofs); err != nil {
		return nil, fmt.Errorf("query audit stats: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM security_events WHERE severity IN ('HIGH', 'CRITICAL')`).Scan(&alerts); err != nil {
		return nil, fmt.Errorf("query security alert count: %w", err)
	}
	return &model.AuditStats{
		TotalLogs: total, BlockedViolations: forbidden, ActiveSecurityAlerts: alerts,
		ImmutableProofsCount: proofs, SuccessCount: success, ForbiddenCount: forbidden,
	}, nil
}

func (r *postgresRepository) GetSecurityEvents(ctx context.Context, limit int) ([]model.SecurityEvent, error) {
	if r.db == nil {
		return nil, fmt.Errorf("audit database is unavailable")
	}
	if limit <= 0 || limit > 1000 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, event_code, severity, source_ip, target_endpoint, description, is_blocked, created_at
		FROM security_events ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query security events: %w", err)
	}
	defer rows.Close()
	events := make([]model.SecurityEvent, 0)
	for rows.Next() {
		var event model.SecurityEvent
		if err := rows.Scan(&event.ID, &event.EventCode, &event.Severity, &event.SourceIP, &event.TargetEndpoint, &event.Description, &event.IsBlocked, &event.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan security event: %w", err)
		}
		events = append(events, event)
	}
	return events, rows.Err()
}
