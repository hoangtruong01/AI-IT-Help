package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"eomp/services/workflow/internal/model"
)

// ChangeRepository defines data access methods for Change Management & CAB.
type ChangeRepository interface {
	ListChanges(ctx context.Context, changeType, status, risk string, page, pageSize int) ([]model.ChangeRequest, int, error)
	GetChangeByID(ctx context.Context, id string) (*model.ChangeRequest, error)
	CreateChange(ctx context.Context, c *model.ChangeRequest) error
	UpdateChange(ctx context.Context, c *model.ChangeRequest) error
	UpdateChangeStatus(ctx context.Context, id, status string, actualStart, actualEnd *time.Time, expectedVersion int) error
	AddCABReview(ctx context.Context, r *model.CABReview) error
	GetCABReviews(ctx context.Context, changeID string) ([]model.CABReview, error)
	RecalculateCABApprovedCount(ctx context.Context, changeID string) (int, error)
	GetChangeCalendar(ctx context.Context, startDate, endDate time.Time) ([]model.ChangeCalendarItem, error)
	GetChangeStats(ctx context.Context) (*model.ChangeStats, error)
	NextChangeNumber(ctx context.Context) (string, error)
}

var ErrVersionConflict = errors.New("change request version conflict")
var errChangeDatabaseUnavailable = errors.New("change database is unavailable")

type postgresChangeRepository struct {
	db *sql.DB
}

// NewChangeRepository creates a new instance of ChangeRepository.
func NewChangeRepository(db *sql.DB) ChangeRepository {
	return &postgresChangeRepository{db: db}
}

func (r *postgresChangeRepository) NextChangeNumber(ctx context.Context) (string, error) {
	if r.db == nil {
		return "", errChangeDatabaseUnavailable
	}
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM change_requests").Scan(&count)
	if err != nil {
		return fmt.Sprintf("CHG-%d", 2001+count), nil
	}
	return fmt.Sprintf("CHG-%04d", 2000+count+1), nil
}

func (r *postgresChangeRepository) ListChanges(ctx context.Context, changeType, status, risk string, page, pageSize int) ([]model.ChangeRequest, int, error) {
	if r.db == nil {
		return nil, 0, errChangeDatabaseUnavailable
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var conditions []string
	var args []interface{}
	argIdx := 1

	if changeType != "" && changeType != "All" {
		conditions = append(conditions, fmt.Sprintf("change_type = $%d", argIdx))
		args = append(args, changeType)
		argIdx++
	}
	if status != "" && status != "All" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if risk != "" && risk != "All" {
		conditions = append(conditions, fmt.Sprintf("risk_level = $%d", argIdx))
		args = append(args, risk)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM change_requests %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count changes: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT 
			id, change_number, title, description, change_type, category,
			priority, risk_level, impact_level, probability_level, status,
			requester_id, requester_name, requester_email, assigned_to_id, assigned_to_name,
			reason_for_change, implementation_plan, rollback_plan, test_plan,
			scheduled_start_time, scheduled_end_time, actual_start_time, actual_end_time,
			downtime_required, downtime_minutes, cab_required_count, cab_approved_count,
			COALESCE(version, 1) AS version, created_at, updated_at
		FROM change_requests
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query changes: %w", err)
	}
	defer rows.Close()

	var changes []model.ChangeRequest
	for rows.Next() {
		var c model.ChangeRequest
		err := rows.Scan(
			&c.ID, &c.ChangeNumber, &c.Title, &c.Description, &c.ChangeType, &c.Category,
			&c.Priority, &c.RiskLevel, &c.ImpactLevel, &c.ProbabilityLevel, &c.Status,
			&c.RequesterID, &c.RequesterName, &c.RequesterEmail, &c.AssignedToID, &c.AssignedToName,
			&c.ReasonForChange, &c.ImplementationPlan, &c.RollbackPlan, &c.TestPlan,
			&c.ScheduledStartTime, &c.ScheduledEndTime, &c.ActualStartTime, &c.ActualEndTime,
			&c.DowntimeRequired, &c.DowntimeMinutes, &c.CABRequiredCount, &c.CABApprovedCount,
			&c.Version, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan change: %w", err)
		}
		changes = append(changes, c)
	}

	return changes, total, nil
}

func (r *postgresChangeRepository) GetChangeByID(ctx context.Context, id string) (*model.ChangeRequest, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT 
			id, change_number, title, description, change_type, category,
			priority, risk_level, impact_level, probability_level, status,
			requester_id, requester_name, requester_email, assigned_to_id, assigned_to_name,
			reason_for_change, implementation_plan, rollback_plan, test_plan,
			scheduled_start_time, scheduled_end_time, actual_start_time, actual_end_time,
			downtime_required, downtime_minutes, cab_required_count, cab_approved_count,
			COALESCE(version, 1) AS version, created_at, updated_at
		FROM change_requests
		WHERE id = $1 OR change_number = $1
	`

	var c model.ChangeRequest
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.ChangeNumber, &c.Title, &c.Description, &c.ChangeType, &c.Category,
		&c.Priority, &c.RiskLevel, &c.ImpactLevel, &c.ProbabilityLevel, &c.Status,
		&c.RequesterID, &c.RequesterName, &c.RequesterEmail, &c.AssignedToID, &c.AssignedToName,
		&c.ReasonForChange, &c.ImplementationPlan, &c.RollbackPlan, &c.TestPlan,
		&c.ScheduledStartTime, &c.ScheduledEndTime, &c.ActualStartTime, &c.ActualEndTime,
		&c.DowntimeRequired, &c.DowntimeMinutes, &c.CABRequiredCount, &c.CABApprovedCount,
		&c.Version, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("change request not found")
		}
		return nil, fmt.Errorf("failed to get change by id: %w", err)
	}

	return &c, nil
}

func (r *postgresChangeRepository) CreateChange(ctx context.Context, c *model.ChangeRequest) error {
	if r.db == nil {
		return errChangeDatabaseUnavailable
	}

	query := `
		INSERT INTO change_requests (
			id, change_number, title, description, change_type, category,
			priority, risk_level, impact_level, probability_level, status,
			requester_id, requester_name, requester_email, assigned_to_id, assigned_to_name,
			reason_for_change, implementation_plan, rollback_plan, test_plan,
			scheduled_start_time, scheduled_end_time, actual_start_time, actual_end_time,
			downtime_required, downtime_minutes, cab_required_count, cab_approved_count,
			version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11,
			$12, $13, $14, $15, $16,
			$17, $18, $19, $20,
			$21, $22, $23, $24,
			$25, $26, $27, $28,
			1, $29, $30
		)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		c.ID, c.ChangeNumber, c.Title, c.Description, c.ChangeType, c.Category,
		c.Priority, c.RiskLevel, c.ImpactLevel, c.ProbabilityLevel, c.Status,
		c.RequesterID, c.RequesterName, c.RequesterEmail, c.AssignedToID, c.AssignedToName,
		c.ReasonForChange, c.ImplementationPlan, c.RollbackPlan, c.TestPlan,
		c.ScheduledStartTime, c.ScheduledEndTime, c.ActualStartTime, c.ActualEndTime,
		c.DowntimeRequired, c.DowntimeMinutes, c.CABRequiredCount, c.CABApprovedCount,
		c.CreatedAt, c.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert change request: %w", err)
	}
	return nil
}

func (r *postgresChangeRepository) UpdateChange(ctx context.Context, c *model.ChangeRequest) error {
	if r.db == nil {
		return errChangeDatabaseUnavailable
	}

	query := `
		UPDATE change_requests SET
			title = $1, description = $2, change_type = $3, category = $4,
			priority = $5, risk_level = $6, impact_level = $7, probability_level = $8,
			assigned_to_id = $9, assigned_to_name = $10, reason_for_change = $11,
			implementation_plan = $12, rollback_plan = $13, test_plan = $14,
			scheduled_start_time = $15, scheduled_end_time = $16,
			downtime_required = $17, downtime_minutes = $18,
			version = version + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $19
	`
	_, err := r.db.ExecContext(
		ctx, query,
		c.Title, c.Description, c.ChangeType, c.Category,
		c.Priority, c.RiskLevel, c.ImpactLevel, c.ProbabilityLevel,
		c.AssignedToID, c.AssignedToName, c.ReasonForChange,
		c.ImplementationPlan, c.RollbackPlan, c.TestPlan,
		c.ScheduledStartTime, c.ScheduledEndTime,
		c.DowntimeRequired, c.DowntimeMinutes,
		c.ID,
	)
	return err
}

func (r *postgresChangeRepository) UpdateChangeStatus(ctx context.Context, id, status string, actualStart, actualEnd *time.Time, expectedVersion int) error {
	if r.db == nil {
		return errChangeDatabaseUnavailable
	}

	query := `
		UPDATE change_requests SET
			status = $1,
			actual_start_time = COALESCE($2, actual_start_time),
			actual_end_time = COALESCE($3, actual_end_time),
			version = version + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND version = $5
	`
	result, err := r.db.ExecContext(ctx, query, status, actualStart, actualEnd, id, expectedVersion)
	if err != nil {
		return err
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return ErrVersionConflict
	}
	return nil
}

func (r *postgresChangeRepository) AddCABReview(ctx context.Context, rev *model.CABReview) error {
	if r.db == nil {
		return errChangeDatabaseUnavailable
	}

	query := `
		INSERT INTO cab_reviews (id, change_id, reviewer_id, reviewer_name, reviewer_role, vote, comments, reviewed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (change_id, reviewer_id) DO UPDATE SET
			vote = EXCLUDED.vote,
			comments = EXCLUDED.comments,
			reviewed_at = EXCLUDED.reviewed_at
	`
	_, err := r.db.ExecContext(ctx, query, rev.ID, rev.ChangeID, rev.ReviewerID, rev.ReviewerName, rev.ReviewerRole, rev.Vote, rev.Comments, rev.ReviewedAt)
	return err
}

func (r *postgresChangeRepository) GetCABReviews(ctx context.Context, changeID string) ([]model.CABReview, error) {
	if r.db == nil {
		return nil, errChangeDatabaseUnavailable
	}

	query := `
		SELECT id, change_id, reviewer_id, reviewer_name, reviewer_role, vote, comments, reviewed_at
		FROM cab_reviews
		WHERE change_id = $1
		ORDER BY reviewed_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, changeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cab reviews: %w", err)
	}
	defer rows.Close()

	var reviews []model.CABReview
	for rows.Next() {
		var rev model.CABReview
		if err := rows.Scan(&rev.ID, &rev.ChangeID, &rev.ReviewerID, &rev.ReviewerName, &rev.ReviewerRole, &rev.Vote, &rev.Comments, &rev.ReviewedAt); err != nil {
			return nil, fmt.Errorf("failed to scan cab review: %w", err)
		}
		reviews = append(reviews, rev)
	}
	return reviews, nil
}

func (r *postgresChangeRepository) RecalculateCABApprovedCount(ctx context.Context, changeID string) (int, error) {
	if r.db == nil {
		return 0, errChangeDatabaseUnavailable
	}

	query := `
		WITH approved AS (
			SELECT COUNT(*) AS cnt
			FROM cab_reviews
			WHERE change_id = $1 AND vote = 'APPROVED'
		)
		UPDATE change_requests
		SET cab_approved_count = (SELECT cnt FROM approved), version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING cab_approved_count
	`
	var count int
	err := r.db.QueryRowContext(ctx, query, changeID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to recalculate cab approvals: %w", err)
	}
	return count, nil
}

func (r *postgresChangeRepository) GetChangeCalendar(ctx context.Context, startDate, endDate time.Time) ([]model.ChangeCalendarItem, error) {
	if r.db == nil {
		return nil, errChangeDatabaseUnavailable
	}

	query := `
		SELECT id, change_number, title, change_type, category, risk_level, status,
		       scheduled_start_time, scheduled_end_time, downtime_required, downtime_minutes
		FROM change_requests
		WHERE scheduled_start_time IS NOT NULL
		  AND (scheduled_start_time BETWEEN $1 AND $2 OR scheduled_end_time BETWEEN $1 AND $2)
		ORDER BY scheduled_start_time ASC
	`
	rows, err := r.db.QueryContext(ctx, query, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query change calendar: %w", err)
	}
	defer rows.Close()

	var items []model.ChangeCalendarItem
	for rows.Next() {
		var it model.ChangeCalendarItem
		if err := rows.Scan(
			&it.ID, &it.ChangeNumber, &it.Title, &it.ChangeType, &it.Category,
			&it.RiskLevel, &it.Status, &it.ScheduledStart, &it.ScheduledEnd,
			&it.DowntimeRequired, &it.DowntimeMinutes,
		); err != nil {
			return nil, fmt.Errorf("failed to scan calendar item: %w", err)
		}
		items = append(items, it)
	}
	return items, nil
}

func (r *postgresChangeRepository) GetChangeStats(ctx context.Context) (*model.ChangeStats, error) {
	if r.db == nil {
		return nil, errChangeDatabaseUnavailable
	}

	query := `
		SELECT
			COUNT(*) FILTER (WHERE status IN ('APPROVED', 'SCHEDULED', 'IMPLEMENTING')) AS active_changes,
			COUNT(*) FILTER (WHERE status = 'CAB_REVIEW') AS pending_cab_review,
			COUNT(*) FILTER (WHERE change_type = 'EMERGENCY') AS emergency_changes,
			COALESCE(
				ROUND(
					(COUNT(*) FILTER (WHERE status = 'COMPLETED')::decimal / 
					 NULLIF(COUNT(*) FILTER (WHERE status IN ('COMPLETED', 'FAILED')), 0)) * 100, 1
				), 100.0
			) AS success_rate,
			COUNT(*) AS total_this_month
		FROM change_requests
	`

	var s model.ChangeStats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&s.ActiveChanges,
		&s.PendingCABReview,
		&s.EmergencyChanges,
		&s.SuccessRatePercent,
		&s.TotalThisMonth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get change stats: %w", err)
	}
	return &s, nil
}
