package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"eomp/services/workflow/internal/model"
)

// Repository interface for Workflow Engine and Approvals
type Repository interface {
	ListDefinitions(ctx context.Context) ([]model.WorkflowDefinition, error)
	FindDefinitionByID(ctx context.Context, id string) (*model.WorkflowDefinition, error)
	FindDefinitionByCode(ctx context.Context, code string) (*model.WorkflowDefinition, error)

	ListInstances(ctx context.Context, query model.WorkflowListQuery) (*model.WorkflowListResponse, error)
	FindInstanceByID(ctx context.Context, id string) (*model.WorkflowInstance, error)
	CreateInstance(ctx context.Context, inst *model.WorkflowInstance) error
	UpdateInstanceStatus(ctx context.Context, id, status, currentStep string, completedAt *time.Time) error
	NextInstanceNumber(ctx context.Context) (string, error)

	ListApprovals(ctx context.Context, approverID, status string, page, pageSize int) (*model.ApprovalListResponse, error)
	FindApprovalByID(ctx context.Context, id string) (*model.ApprovalRequest, error)
	CreateApproval(ctx context.Context, app *model.ApprovalRequest) error
	UpdateApprovalDecision(ctx context.Context, id, status, notes string, decidedAt *time.Time) error

	AddLog(ctx context.Context, log *model.WorkflowLog) error
	ListLogs(ctx context.Context, instanceID string) ([]model.WorkflowLog, error)
	GetStats(ctx context.Context) (*model.WorkflowStats, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository constructs a PostgreSQL Workflow repository
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) NextInstanceNumber(ctx context.Context) (string, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM workflow_instances").Scan(&count)
	if err != nil {
		return "", fmt.Errorf("failed to count instances: %w", err)
	}
	return fmt.Sprintf("WFI-%04d", 1000+count+1), nil
}

func (r *postgresRepository) ListDefinitions(ctx context.Context) ([]model.WorkflowDefinition, error) {
	query := `
		SELECT id, code, name, description, category, trigger_type, is_active, steps_config::text, created_at, updated_at
		FROM workflow_definitions
		WHERE is_active = TRUE
		ORDER BY name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query definitions: %w", err)
	}
	defer rows.Close()

	defs := []model.WorkflowDefinition{}
	for rows.Next() {
		var d model.WorkflowDefinition
		var desc sql.NullString
		err := rows.Scan(&d.ID, &d.Code, &d.Name, &desc, &d.Category, &d.TriggerType, &d.IsActive, &d.StepsConfig, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan definition: %w", err)
		}
		if desc.Valid {
			d.Description = &desc.String
		}
		defs = append(defs, d)
	}
	return defs, nil
}

func (r *postgresRepository) FindDefinitionByID(ctx context.Context, id string) (*model.WorkflowDefinition, error) {
	query := `
		SELECT id, code, name, description, category, trigger_type, is_active, steps_config::text, created_at, updated_at
		FROM workflow_definitions
		WHERE id = $1
	`
	var d model.WorkflowDefinition
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(&d.ID, &d.Code, &d.Name, &desc, &d.Category, &d.TriggerType, &d.IsActive, &d.StepsConfig, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}
	if desc.Valid {
		d.Description = &desc.String
	}
	return &d, nil
}

func (r *postgresRepository) FindDefinitionByCode(ctx context.Context, code string) (*model.WorkflowDefinition, error) {
	query := `
		SELECT id, code, name, description, category, trigger_type, is_active, steps_config::text, created_at, updated_at
		FROM workflow_definitions
		WHERE code = $1
	`
	var d model.WorkflowDefinition
	var desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, code).Scan(&d.ID, &d.Code, &d.Name, &desc, &d.Category, &d.TriggerType, &d.IsActive, &d.StepsConfig, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get definition by code: %w", err)
	}
	if desc.Valid {
		d.Description = &desc.String
	}
	return &d, nil
}

func (r *postgresRepository) ListInstances(ctx context.Context, query model.WorkflowListQuery) (*model.WorkflowListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	whereClauses := []string{"1=1"}
	args := []any{}
	argIndex := 1

	if query.Status != "" && query.Status != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, query.Status)
		argIndex++
	}

	if query.Search != "" {
		pat := "%" + strings.ToLower(query.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(instance_number) LIKE $%d OR LOWER(title) LIKE $%d OR LOWER(requester_name) LIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, pat)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM workflow_instances WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count workflow instances: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, instance_number, definition_id, definition_name, entity_type, entity_id,
			title, requester_id, requester_name, requester_email,
			current_step_name, status, context_data::text, started_at, completed_at,
			COALESCE(version, 1) AS version, created_at, updated_at
		FROM workflow_instances
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query instances: %w", err)
	}
	defer rows.Close()

	instances := []model.WorkflowInstance{}
	for rows.Next() {
		var inst model.WorkflowInstance
		var ctxStr sql.NullString
		var completedAt sql.NullTime

		err := rows.Scan(
			&inst.ID, &inst.InstanceNumber, &inst.DefinitionID, &inst.DefinitionName, &inst.EntityType, &inst.EntityID,
			&inst.Title, &inst.RequesterID, &inst.RequesterName, &inst.RequesterEmail,
			&inst.CurrentStepName, &inst.Status, &ctxStr, &inst.StartedAt, &completedAt,
			&inst.Version, &inst.CreatedAt, &inst.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan instance: %w", err)
		}

		if ctxStr.Valid {
			inst.ContextData = &ctxStr.String
		}
		if completedAt.Valid {
			inst.CompletedAt = &completedAt.Time
		}

		instances = append(instances, inst)
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))

	return &model.WorkflowListResponse{
		Data:       instances,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindInstanceByID(ctx context.Context, id string) (*model.WorkflowInstance, error) {
	query := `
		SELECT 
			id, instance_number, definition_id, definition_name, entity_type, entity_id,
			title, requester_id, requester_name, requester_email,
			current_step_name, status, context_data::text, started_at, completed_at,
			COALESCE(version, 1) AS version, created_at, updated_at
		FROM workflow_instances
		WHERE id = $1
	`
	var inst model.WorkflowInstance
	var ctxStr sql.NullString
	var completedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&inst.ID, &inst.InstanceNumber, &inst.DefinitionID, &inst.DefinitionName, &inst.EntityType, &inst.EntityID,
		&inst.Title, &inst.RequesterID, &inst.RequesterName, &inst.RequesterEmail,
		&inst.CurrentStepName, &inst.Status, &ctxStr, &inst.StartedAt, &completedAt,
		&inst.Version, &inst.CreatedAt, &inst.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query instance: %w", err)
	}

	if ctxStr.Valid {
		inst.ContextData = &ctxStr.String
	}
	if completedAt.Valid {
		inst.CompletedAt = &completedAt.Time
	}

	return &inst, nil
}

func (r *postgresRepository) CreateInstance(ctx context.Context, inst *model.WorkflowInstance) error {
	query := `
		INSERT INTO workflow_instances (
			instance_number, definition_id, definition_name, entity_type, entity_id,
			title, requester_id, requester_name, requester_email,
			current_step_name, status, context_data, started_at, version, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12::jsonb, $13, 1, $14, $15
		)
		RETURNING id, version, created_at, updated_at
	`
	now := time.Now()
	ctxData := "{}"
	if inst.ContextData != nil && *inst.ContextData != "" {
		ctxData = *inst.ContextData
	}

	err := r.db.QueryRowContext(
		ctx, query,
		inst.InstanceNumber, inst.DefinitionID, inst.DefinitionName, inst.EntityType, inst.EntityID,
		inst.Title, inst.RequesterID, inst.RequesterName, inst.RequesterEmail,
		inst.CurrentStepName, inst.Status, ctxData, now, now, now,
	).Scan(&inst.ID, &inst.Version, &inst.CreatedAt, &inst.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert workflow instance: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateInstanceStatus(ctx context.Context, id, status, currentStep string, completedAt *time.Time) error {
	query := `
		UPDATE workflow_instances
		SET status = $2, current_step_name = $3, completed_at = COALESCE($4, completed_at), version = version + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, status, currentStep, completedAt)
	if err != nil {
		return fmt.Errorf("failed to update instance: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListApprovals(ctx context.Context, approverID, status string, page, pageSize int) (*model.ApprovalListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}

	whereClauses := []string{"1=1"}
	args := []any{}
	idx := 1

	if approverID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("approver_id = $%d", idx))
		args = append(args, approverID)
		idx++
	}

	if status != "" && status != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("status = $%d", idx))
		args = append(args, status)
		idx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM approval_requests WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count approvals: %w", err)
	}

	offset := (page - 1) * pageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, instance_id, step_id, title, approver_id, approver_name, approver_role,
			approval_level, status, decision_notes, decided_at, sla_deadline, created_at
		FROM approval_requests
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)

	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query approvals: %w", err)
	}
	defer rows.Close()

	list := []model.ApprovalRequest{}
	for rows.Next() {
		var a model.ApprovalRequest
		var stepID, notesStr sql.NullString
		var decidedAt sql.NullTime

		err := rows.Scan(
			&a.ID, &a.InstanceID, &stepID, &a.Title, &a.ApproverID, &a.ApproverName, &a.ApproverRole,
			&a.ApprovalLevel, &a.Status, &notesStr, &decidedAt, &a.SLADeadline, &a.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan approval: %w", err)
		}

		if stepID.Valid {
			a.StepID = &stepID.String
		}
		if notesStr.Valid {
			a.DecisionNotes = &notesStr.String
		}
		if decidedAt.Valid {
			a.DecidedAt = &decidedAt.Time
		}

		list = append(list, a)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

	return &model.ApprovalListResponse{
		Data:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindApprovalByID(ctx context.Context, id string) (*model.ApprovalRequest, error) {
	query := `
		SELECT 
			id, instance_id, step_id, title, approver_id, approver_name, approver_role,
			approval_level, status, decision_notes, decided_at, sla_deadline, created_at
		FROM approval_requests
		WHERE id = $1
	`
	var a model.ApprovalRequest
	var stepID, notesStr sql.NullString
	var decidedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.InstanceID, &stepID, &a.Title, &a.ApproverID, &a.ApproverName, &a.ApproverRole,
		&a.ApprovalLevel, &a.Status, &notesStr, &decidedAt, &a.SLADeadline, &a.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get approval: %w", err)
	}

	if stepID.Valid {
		a.StepID = &stepID.String
	}
	if notesStr.Valid {
		a.DecisionNotes = &notesStr.String
	}
	if decidedAt.Valid {
		a.DecidedAt = &decidedAt.Time
	}

	return &a, nil
}

func (r *postgresRepository) CreateApproval(ctx context.Context, app *model.ApprovalRequest) error {
	query := `
		INSERT INTO approval_requests (
			instance_id, step_id, title, approver_id, approver_name, approver_role,
			approval_level, status, sla_deadline, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10
		)
		RETURNING id, created_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		app.InstanceID, app.StepID, app.Title, app.ApproverID, app.ApproverName, app.ApproverRole,
		app.ApprovalLevel, app.Status, app.SLADeadline, now,
	).Scan(&app.ID, &app.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert approval request: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateApprovalDecision(ctx context.Context, id, status, notes string, decidedAt *time.Time) error {
	query := `
		UPDATE approval_requests
		SET status = $2, decision_notes = $3, decided_at = $4
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query, id, status, notes, decidedAt)
	if err != nil {
		return fmt.Errorf("failed to update approval decision: %w", err)
	}
	return nil
}

func (r *postgresRepository) AddLog(ctx context.Context, l *model.WorkflowLog) error {
	query := `
		INSERT INTO workflow_logs (instance_id, actor_id, actor_name, action, message, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(ctx, query, l.InstanceID, l.ActorID, l.ActorName, l.Action, l.Message, now).Scan(&l.ID, &l.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert workflow log: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListLogs(ctx context.Context, instanceID string) ([]model.WorkflowLog, error) {
	query := `
		SELECT id, instance_id, actor_id, actor_name, action, message, created_at
		FROM workflow_logs
		WHERE instance_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query workflow logs: %w", err)
	}
	defer rows.Close()

	logs := []model.WorkflowLog{}
	for rows.Next() {
		var l model.WorkflowLog
		if err := rows.Scan(&l.ID, &l.InstanceID, &l.ActorID, &l.ActorName, &l.Action, &l.Message, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *postgresRepository) GetStats(ctx context.Context) (*model.WorkflowStats, error) {
	query := `
		SELECT 
			(SELECT COUNT(*) FROM workflow_definitions WHERE is_active = TRUE) AS total_defs,
			(SELECT COUNT(*) FROM workflow_instances WHERE status IN ('RUNNING', 'WAITING_APPROVAL')) AS active_instances,
			(SELECT COUNT(*) FROM approval_requests WHERE status = 'PENDING') AS pending_approvals,
			(SELECT COUNT(*) FROM workflow_instances WHERE status = 'COMPLETED') AS completed_today
	`
	var s model.WorkflowStats
	err := r.db.QueryRowContext(ctx, query).Scan(&s.TotalDefinitions, &s.ActiveInstances, &s.PendingApprovals, &s.CompletedToday)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow stats: %w", err)
	}
	return &s, nil
}
