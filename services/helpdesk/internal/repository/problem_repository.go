package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	appErrors "eomp/packages/shared/pkg/errors"
	"eomp/services/helpdesk/internal/model"
)

// ProblemRepository defines methods for ITIL Problem data access.
type ProblemRepository interface {
	ListProblems(ctx context.Context, category, status string, page, pageSize int) ([]model.Problem, int, error)
	GetProblemByID(ctx context.Context, id string) (*model.Problem, error)
	CreateProblem(ctx context.Context, p *model.Problem) error
	UpdateProblem(ctx context.Context, p *model.Problem, expectedVersion int) error
	UpdateProblemStatus(ctx context.Context, id, status string, resolution *string, resolvedAt, closedAt *time.Time, expectedVersion int) error
	UpdateProblemRCA(ctx context.Context, id string, rootCause, workaround *string, isKnownError bool, expectedVersion int) error
	LinkIncident(ctx context.Context, link *model.ProblemIncidentLink) error
	UnlinkIncident(ctx context.Context, problemID, ticketID string) error
	GetLinkedIncidents(ctx context.Context, problemID string) ([]model.ProblemIncidentLink, error)
	CascadeResolveLinkedTickets(ctx context.Context, problemID, problemNumber, resolution string) ([]string, error)
	GetProblemStats(ctx context.Context) (*model.ProblemStats, error)
	NextProblemNumber(ctx context.Context) (string, error)
}

type postgresProblemRepository struct {
	db *sql.DB
}

// NewProblemRepository creates a new instance of ProblemRepository.
func NewProblemRepository(db *sql.DB) ProblemRepository {
	return &postgresProblemRepository{db: db}
}

func (r *postgresProblemRepository) NextProblemNumber(ctx context.Context) (string, error) {
	if r.db == nil {
		return "", errors.New("database connection is nil")
	}
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM problems").Scan(&count)
	if err != nil {
		return fmt.Sprintf("PRB-%d", 1001+count), nil
	}
	return fmt.Sprintf("PRB-%04d", 1000+count+1), nil
}

func (r *postgresProblemRepository) ListProblems(ctx context.Context, category, status string, page, pageSize int) ([]model.Problem, int, error) {
	if r.db == nil {
		return nil, 0, errors.New("database connection is nil")
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

	if category != "" && category != "All" {
		conditions = append(conditions, fmt.Sprintf("p.category = $%d", argIdx))
		args = append(args, category)
		argIdx++
	}
	if status != "" && status != "All" {
		conditions = append(conditions, fmt.Sprintf("p.status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM problems p %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count problems: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT 
			p.id, p.problem_number, p.title, p.description, p.category, p.priority,
			p.status, p.impact, p.urgency, p.assignee_id, p.assignee_name,
			p.root_cause, p.workaround, p.resolution, p.is_known_error,
			COALESCE((SELECT COUNT(*) FROM problem_incident_links pil WHERE pil.problem_id = p.id), 0) AS linked_count,
			p.version, p.created_at, p.updated_at, p.resolved_at, p.closed_at
		FROM problems p
		%s
		ORDER BY p.created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIdx, argIdx+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query problems: %w", err)
	}
	defer rows.Close()

	var problems []model.Problem
	for rows.Next() {
		var p model.Problem
		err := rows.Scan(
			&p.ID, &p.ProblemNumber, &p.Title, &p.Description, &p.Category, &p.Priority,
			&p.Status, &p.Impact, &p.Urgency, &p.AssigneeID, &p.AssigneeName,
			&p.RootCause, &p.Workaround, &p.Resolution, &p.IsKnownError,
			&p.LinkedCount, &p.Version, &p.CreatedAt, &p.UpdatedAt, &p.ResolvedAt, &p.ClosedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan problem: %w", err)
		}
		problems = append(problems, p)
	}

	return problems, total, nil
}

func (r *postgresProblemRepository) GetProblemByID(ctx context.Context, id string) (*model.Problem, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT 
			p.id, p.problem_number, p.title, p.description, p.category, p.priority,
			p.status, p.impact, p.urgency, p.assignee_id, p.assignee_name,
			p.root_cause, p.workaround, p.resolution, p.is_known_error,
			COALESCE((SELECT COUNT(*) FROM problem_incident_links pil WHERE pil.problem_id = p.id), 0) AS linked_count,
			p.version, p.created_at, p.updated_at, p.resolved_at, p.closed_at
		FROM problems p
		WHERE p.id = $1 OR p.problem_number = $1
	`

	var p model.Problem
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.ProblemNumber, &p.Title, &p.Description, &p.Category, &p.Priority,
		&p.Status, &p.Impact, &p.Urgency, &p.AssigneeID, &p.AssigneeName,
		&p.RootCause, &p.Workaround, &p.Resolution, &p.IsKnownError,
		&p.LinkedCount, &p.Version, &p.CreatedAt, &p.UpdatedAt, &p.ResolvedAt, &p.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("problem record not found")
		}
		return nil, fmt.Errorf("failed to get problem by id: %w", err)
	}

	return &p, nil
}

func (r *postgresProblemRepository) CreateProblem(ctx context.Context, p *model.Problem) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
		INSERT INTO problems (
			id, problem_number, title, description, category, priority, status,
			impact, urgency, assignee_id, assignee_name, root_cause, workaround,
			resolution, is_known_error, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13,
			$14, $15, 1, $16, $17
		)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		p.ID, p.ProblemNumber, p.Title, p.Description, p.Category, p.Priority, p.Status,
		p.Impact, p.Urgency, p.AssigneeID, p.AssigneeName, p.RootCause, p.Workaround,
		p.Resolution, p.IsKnownError, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert problem: %w", err)
	}
	return nil
}

func (r *postgresProblemRepository) UpdateProblem(ctx context.Context, p *model.Problem, expectedVersion int) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
		UPDATE problems SET
			title = $1, description = $2, category = $3, priority = $4,
			impact = $5, urgency = $6, assignee_id = $7, assignee_name = $8,
			root_cause = $9, workaround = $10, resolution = $11, is_known_error = $12,
			version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $13 AND version = $14
	`
	result, err := r.db.ExecContext(
		ctx, query,
		p.Title, p.Description, p.Category, p.Priority,
		p.Impact, p.Urgency, p.AssigneeID, p.AssigneeName,
		p.RootCause, p.Workaround, p.Resolution, p.IsKnownError,
		p.ID, expectedVersion,
	)
	return ensureProblemUpdated(result, err)
}

func (r *postgresProblemRepository) UpdateProblemStatus(ctx context.Context, id, status string, resolution *string, resolvedAt, closedAt *time.Time, expectedVersion int) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
		UPDATE problems SET
			status = $1,
			resolution = COALESCE($2, resolution),
			resolved_at = COALESCE($3, resolved_at),
			closed_at = COALESCE($4, closed_at),
			version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $5 AND version = $6
	`
	result, err := r.db.ExecContext(ctx, query, status, resolution, resolvedAt, closedAt, id, expectedVersion)
	return ensureProblemUpdated(result, err)
}

func (r *postgresProblemRepository) UpdateProblemRCA(ctx context.Context, id string, rootCause, workaround *string, isKnownError bool, expectedVersion int) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
		UPDATE problems SET
			root_cause = $1,
			workaround = $2,
			is_known_error = $3,
			version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $4 AND version = $5
	`
	result, err := r.db.ExecContext(ctx, query, rootCause, workaround, isKnownError, id, expectedVersion)
	return ensureProblemUpdated(result, err)
}

func ensureProblemUpdated(result sql.Result, err error) error {
	if err != nil {
		return fmt.Errorf("update problem: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read problem update result: %w", err)
	}
	if rows != 1 {
		return appErrors.Conflict("problem was updated by another request; reload and retry")
	}
	return nil
}

func (r *postgresProblemRepository) LinkIncident(ctx context.Context, link *model.ProblemIncidentLink) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	query := `
		INSERT INTO problem_incident_links (id, problem_id, ticket_id, ticket_number, ticket_title, linked_by, linked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (problem_id, ticket_id) DO NOTHING
	`
	_, err := r.db.ExecContext(ctx, query, link.ID, link.ProblemID, link.TicketID, link.TicketNumber, link.TicketTitle, link.LinkedBy, link.LinkedAt)
	return err
}

func (r *postgresProblemRepository) UnlinkIncident(ctx context.Context, problemID, ticketID string) error {
	if r.db == nil {
		return errors.New("database connection is nil")
	}

	query := `DELETE FROM problem_incident_links WHERE problem_id = $1 AND (ticket_id = $2 OR ticket_number = $2)`
	_, err := r.db.ExecContext(ctx, query, problemID, ticketID)
	return err
}

func (r *postgresProblemRepository) GetLinkedIncidents(ctx context.Context, problemID string) ([]model.ProblemIncidentLink, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT id, problem_id, ticket_id, ticket_number, ticket_title, linked_by, linked_at
		FROM problem_incident_links
		WHERE problem_id = $1
		ORDER BY linked_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked incidents: %w", err)
	}
	defer rows.Close()

	var links []model.ProblemIncidentLink
	for rows.Next() {
		var l model.ProblemIncidentLink
		if err := rows.Scan(&l.ID, &l.ProblemID, &l.TicketID, &l.TicketNumber, &l.TicketTitle, &l.LinkedBy, &l.LinkedAt); err != nil {
			return nil, fmt.Errorf("failed to scan link: %w", err)
		}
		links = append(links, l)
	}
	return links, nil
}

// CascadeResolveLinkedTickets updates all linked tickets to RESOLVED when the problem is resolved.
func (r *postgresProblemRepository) CascadeResolveLinkedTickets(ctx context.Context, problemID, problemNumber, resolution string) ([]string, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// 1. Get all linked ticket IDs
	rows, err := tx.QueryContext(ctx, "SELECT ticket_id, ticket_number FROM problem_incident_links WHERE problem_id = $1", problemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query linked tickets for cascade: %w", err)
	}
	defer rows.Close()

	type tInfo struct {
		id     string
		number string
	}
	var ticketsToResolve []tInfo
	for rows.Next() {
		var t tInfo
		if err := rows.Scan(&t.id, &t.number); err == nil {
			ticketsToResolve = append(ticketsToResolve, t)
		}
	}
	rows.Close()

	var resolvedNumbers []string
	now := time.Now()
	resNote := fmt.Sprintf("Automatically resolved via Problem Record %s. Root Cause Resolution: %s", problemNumber, resolution)

	for _, t := range ticketsToResolve {
		// Update ticket status
		_, err := tx.ExecContext(
			ctx,
			"UPDATE tickets SET status = 'RESOLVED', resolved_at = $1, version = version + 1, updated_at = $1 WHERE id = $2 AND status != 'RESOLVED' AND status != 'CLOSED'",
			now, t.id,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to cascade resolve ticket %s: %w", t.number, err)
		}

		// Insert timeline log
		timelineID := fmt.Sprintf("tl-%d-%s", now.UnixNano(), t.id[:8])
		_, _ = tx.ExecContext(
			ctx,
			"INSERT INTO ticket_timeline (id, ticket_id, actor_id, actor_name, action, old_value, new_value, notes, created_at) VALUES ($1, $2, 'system', 'ITIL Problem Management', 'STATUS_CHANGE', 'OPEN', 'RESOLVED', $3, $4)",
			timelineID, t.id, resNote, now,
		)

		resolvedNumbers = append(resolvedNumbers, t.number)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit cascade resolution: %w", err)
	}

	return resolvedNumbers, nil
}

func (r *postgresProblemRepository) GetProblemStats(ctx context.Context) (*model.ProblemStats, error) {
	if r.db == nil {
		return nil, errors.New("database connection is nil")
	}

	query := `
		SELECT
			COUNT(*) AS total_problems,
			COUNT(*) FILTER (WHERE status = 'UNDER_INVESTIGATION') AS under_investigation,
			COUNT(*) FILTER (WHERE is_known_error = TRUE OR status = 'KNOWN_ERROR') AS known_errors,
			COUNT(*) FILTER (WHERE status = 'RESOLVED' OR status = 'CLOSED') AS resolved_problems,
			(SELECT COUNT(*) FROM problem_incident_links) AS total_linked_tickets
		FROM problems
	`

	var s model.ProblemStats
	err := r.db.QueryRowContext(ctx, query).Scan(
		&s.TotalProblems,
		&s.UnderInvestigation,
		&s.KnownErrors,
		&s.ResolvedProblems,
		&s.TotalLinkedTickets,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get problem stats: %w", err)
	}
	return &s, nil
}
