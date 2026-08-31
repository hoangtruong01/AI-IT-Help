package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	appErrors "eomp/packages/shared/pkg/errors"
	"eomp/packages/shared/pkg/middleware"
	"eomp/services/helpdesk/internal/model"
)

// Repository interface for helpdesk data operations
type Repository interface {
	ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error)
	FindTicketByID(ctx context.Context, id string) (*model.Ticket, error)
	FindTicketByIDForActor(ctx context.Context, id string, actor middleware.Actor) (*model.Ticket, error)
	FindTicketByNumber(ctx context.Context, ticketNumber string) (*model.Ticket, error)
	CreateTicket(ctx context.Context, ticket *model.Ticket) error
	UpdateTicketStatus(ctx context.Context, id, status string, assigneeID, assigneeName *string, resolvedAt, closedAt *time.Time, expectedVersion *int) error
	AssignTicket(ctx context.Context, id, assigneeID, assigneeName string, expectedVersion *int) error

	AddComment(ctx context.Context, comment *model.TicketComment) error
	ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error)

	AddTimelineRecord(ctx context.Context, timeline *model.TicketTimeline) error
	ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error)

	ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error)
	ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error)
	FindServiceCatalogItemByID(ctx context.Context, id string) (*model.ServiceCatalogItem, error)
	NextTicketNumber(ctx context.Context) (string, error)
	ListTicketsByAssetID(ctx context.Context, assetID string) ([]model.Ticket, error)
	ListTicketsByAssetIDForActor(ctx context.Context, assetID string, actor middleware.Actor) ([]model.Ticket, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository constructs a PostgreSQL Helpdesk repository
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) NextTicketNumber(ctx context.Context) (string, error) {
	var number int64
	err := r.db.QueryRowContext(ctx, "SELECT nextval('ticket_number_seq')").Scan(&number)
	if err != nil {
		return "", fmt.Errorf("failed to allocate ticket number: %w", err)
	}
	return fmt.Sprintf("TK-%04d", number), nil
}

func (r *postgresRepository) ListTickets(ctx context.Context, query model.TicketListQuery) (*model.TicketListResponse, error) {
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

	if query.Priority != "" && query.Priority != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("priority = $%d", argIndex))
		args = append(args, query.Priority)
		argIndex++
	}

	if query.Category != "" && query.Category != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, query.Category)
		argIndex++
	}

	// Scope-based authorization filters (Gate B-01)
	switch query.ActorRole {
	case "ROLE_EMPLOYEE":
		whereClauses = append(whereClauses, fmt.Sprintf("requester_id = $%d", argIndex))
		args = append(args, query.ActorID)
		argIndex++
	case "ROLE_AGENT":
		whereClauses = append(whereClauses, fmt.Sprintf("(assignee_id = $%d OR assignee_id IS NULL OR assignee_id = '')", argIndex))
		args = append(args, query.ActorID)
		argIndex++
	case "ROLE_MANAGER":
		whereClauses = append(whereClauses, fmt.Sprintf("department_id = $%d", argIndex))
		args = append(args, query.ActorDepartmentID)
		argIndex++
	}

	if query.DepartmentID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("department_id = $%d", argIndex))
		args = append(args, query.DepartmentID)
		argIndex++
	}

	if query.AssigneeID != "" && query.ActorRole != "ROLE_EMPLOYEE" {
		whereClauses = append(whereClauses, fmt.Sprintf("assignee_id = $%d", argIndex))
		args = append(args, query.AssigneeID)
		argIndex++
	}

	if query.RequesterID != "" && query.ActorRole != "ROLE_EMPLOYEE" {
		whereClauses = append(whereClauses, fmt.Sprintf("requester_id = $%d", argIndex))
		args = append(args, query.RequesterID)
		argIndex++
	}

	if query.Search != "" {
		pattern := "%" + strings.ToLower(query.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(ticket_number) LIKE $%d OR LOWER(title) LIKE $%d OR LOWER(requester_name) LIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, pattern)
		argIndex++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets WHERE %s", whereSQL)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count tickets: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT 
			id, ticket_number, title, description, service_item_id, category, priority, status,
			requester_id, requester_name, requester_email,
			assignee_id, assignee_name, department_id, affected_ci_id,
			sla_response_deadline, sla_resolution_deadline, responded_at, resolved_at, closed_at,
			sla_status, COALESCE(version, 1) AS version, created_at, updated_at
		FROM tickets
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIndex, argIndex+1)

	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets: %w", err)
	}
	defer rows.Close()

	tickets := []model.Ticket{}
	for rows.Next() {
		var t model.Ticket
		var assigneeID, assigneeName, deptID, ciID sql.NullString
		var serviceItemID sql.NullString
		var respondedAt, resolvedAt, closedAt sql.NullTime

		err := rows.Scan(
			&t.ID, &t.TicketNumber, &t.Title, &t.Description, &serviceItemID, &t.Category, &t.Priority, &t.Status,
			&t.RequesterID, &t.RequesterName, &t.RequesterEmail,
			&assigneeID, &assigneeName, &deptID, &ciID,
			&t.SLAResponseDeadline, &t.SLAResolutionDeadline, &respondedAt, &resolvedAt, &closedAt,
			&t.SLAStatus, &t.Version, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket: %w", err)
		}

		if serviceItemID.Valid {
			t.ServiceItemID = &serviceItemID.String
		}
		if assigneeID.Valid {
			t.AssigneeID = &assigneeID.String
		}
		if assigneeName.Valid {
			t.AssigneeName = &assigneeName.String
		}
		if deptID.Valid {
			t.DepartmentID = &deptID.String
		}
		if ciID.Valid {
			t.AffectedCIID = &ciID.String
		}
		if respondedAt.Valid {
			t.RespondedAt = &respondedAt.Time
		}
		if resolvedAt.Valid {
			t.ResolvedAt = &resolvedAt.Time
		}
		if closedAt.Valid {
			t.ClosedAt = &closedAt.Time
		}

		tickets = append(tickets, t)
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))

	return &model.TicketListResponse{
		Data:       tickets,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *postgresRepository) FindTicketByID(ctx context.Context, id string) (*model.Ticket, error) {
	return r.findTicketByID(ctx, id, nil)
}

func (r *postgresRepository) FindTicketByIDForActor(ctx context.Context, id string, actor middleware.Actor) (*model.Ticket, error) {
	return r.findTicketByID(ctx, id, &actor)
}

func (r *postgresRepository) findTicketByID(ctx context.Context, id string, actor *middleware.Actor) (*model.Ticket, error) {
	where := "id = $1"
	args := []any{id}
	if actor != nil {
		switch {
		case actor.IsAdmin():
		case actor.IsEmployee() && actor.ID != "":
			where += " AND requester_id = $2"
			args = append(args, actor.ID)
		case actor.IsAgent() && actor.ID != "":
			where += " AND (assignee_id = $2 OR assignee_id IS NULL OR assignee_id = '')"
			args = append(args, actor.ID)
		case actor.IsManager() && actor.DepartmentID != "":
			where += " AND department_id = $2"
			args = append(args, actor.DepartmentID)
		default:
			return nil, nil
		}
	}

	query := fmt.Sprintf(`
		SELECT 
			id, ticket_number, title, description, service_item_id, category, priority, status,
			requester_id, requester_name, requester_email,
			assignee_id, assignee_name, department_id, affected_ci_id,
			sla_response_deadline, sla_resolution_deadline, responded_at, resolved_at, closed_at,
			sla_status, COALESCE(version, 1) AS version, created_at, updated_at
		FROM tickets
		WHERE %s
	`, where)
	var t model.Ticket
	var assigneeID, assigneeName, deptID, ciID, serviceItemID sql.NullString
	var respondedAt, resolvedAt, closedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&t.ID, &t.TicketNumber, &t.Title, &t.Description, &serviceItemID, &t.Category, &t.Priority, &t.Status,
		&t.RequesterID, &t.RequesterName, &t.RequesterEmail,
		&assigneeID, &assigneeName, &deptID, &ciID,
		&t.SLAResponseDeadline, &t.SLAResolutionDeadline, &respondedAt, &resolvedAt, &closedAt,
		&t.SLAStatus, &t.Version, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query ticket by id: %w", err)
	}

	if serviceItemID.Valid {
		t.ServiceItemID = &serviceItemID.String
	}
	if assigneeID.Valid {
		t.AssigneeID = &assigneeID.String
	}
	if assigneeName.Valid {
		t.AssigneeName = &assigneeName.String
	}
	if deptID.Valid {
		t.DepartmentID = &deptID.String
	}
	if ciID.Valid {
		t.AffectedCIID = &ciID.String
	}
	if respondedAt.Valid {
		t.RespondedAt = &respondedAt.Time
	}
	if resolvedAt.Valid {
		t.ResolvedAt = &resolvedAt.Time
	}
	if closedAt.Valid {
		t.ClosedAt = &closedAt.Time
	}

	return &t, nil
}

func (r *postgresRepository) FindTicketByNumber(ctx context.Context, ticketNumber string) (*model.Ticket, error) {
	query := "SELECT id FROM tickets WHERE ticket_number = $1"
	var id string
	err := r.db.QueryRowContext(ctx, query, ticketNumber).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return r.FindTicketByID(ctx, id)
}

func (r *postgresRepository) ListTicketsByAssetID(ctx context.Context, assetID string) ([]model.Ticket, error) {
	return r.listTicketsByAssetID(ctx, assetID, nil)
}

func (r *postgresRepository) ListTicketsByAssetIDForActor(ctx context.Context, assetID string, actor middleware.Actor) ([]model.Ticket, error) {
	return r.listTicketsByAssetID(ctx, assetID, &actor)
}

func (r *postgresRepository) listTicketsByAssetID(ctx context.Context, assetID string, actor *middleware.Actor) ([]model.Ticket, error) {
	where := "affected_ci_id = $1"
	args := []any{assetID}
	if actor != nil {
		switch {
		case actor.IsAdmin():
		case actor.IsAgent() && actor.ID != "":
			where += " AND (assignee_id = $2 OR assignee_id IS NULL OR assignee_id = '')"
			args = append(args, actor.ID)
		case actor.IsManager() && actor.DepartmentID != "":
			where += " AND department_id = $2"
			args = append(args, actor.DepartmentID)
		default:
			return []model.Ticket{}, nil
		}
	}

	query := fmt.Sprintf(`
		SELECT 
			id, ticket_number, title, description, service_item_id, category, priority, status,
			requester_id, requester_name, requester_email,
			assignee_id, assignee_name, department_id, affected_ci_id,
			sla_response_deadline, sla_resolution_deadline, responded_at, resolved_at, closed_at,
			sla_status, COALESCE(version, 1) AS version, created_at, updated_at
		FROM tickets
		WHERE %s
		ORDER BY created_at DESC
	`, where)
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tickets by asset: %w", err)
	}
	defer rows.Close()

	tickets := []model.Ticket{}
	for rows.Next() {
		var t model.Ticket
		var assigneeID, assigneeName, deptID, ciID sql.NullString
		var serviceItemID sql.NullString
		var respondedAt, resolvedAt, closedAt sql.NullTime

		err := rows.Scan(
			&t.ID, &t.TicketNumber, &t.Title, &t.Description, &serviceItemID, &t.Category, &t.Priority, &t.Status,
			&t.RequesterID, &t.RequesterName, &t.RequesterEmail,
			&assigneeID, &assigneeName, &deptID, &ciID,
			&t.SLAResponseDeadline, &t.SLAResolutionDeadline, &respondedAt, &resolvedAt, &closedAt,
			&t.SLAStatus, &t.Version, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan ticket by asset: %w", err)
		}

		if serviceItemID.Valid {
			t.ServiceItemID = &serviceItemID.String
		}
		if assigneeID.Valid {
			t.AssigneeID = &assigneeID.String
		}
		if assigneeName.Valid {
			t.AssigneeName = &assigneeName.String
		}
		if deptID.Valid {
			t.DepartmentID = &deptID.String
		}
		if ciID.Valid {
			t.AffectedCIID = &ciID.String
		}
		if respondedAt.Valid {
			t.RespondedAt = &respondedAt.Time
		}
		if resolvedAt.Valid {
			t.ResolvedAt = &resolvedAt.Time
		}
		if closedAt.Valid {
			t.ClosedAt = &closedAt.Time
		}

		tickets = append(tickets, t)
	}
	return tickets, nil
}

func (r *postgresRepository) CreateTicket(ctx context.Context, t *model.Ticket) error {
	query := `
		INSERT INTO tickets (
			ticket_number, title, description, service_item_id, category, priority, status,
			requester_id, requester_name, requester_email,
			assignee_id, assignee_name, department_id, affected_ci_id,
			sla_response_deadline, sla_resolution_deadline, sla_status, version,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10,
			$11, $12, $13, $14,
			$15, $16, $17, 1,
			$18, $19
		)
		RETURNING id, version, created_at, updated_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		t.TicketNumber, t.Title, t.Description, t.ServiceItemID, t.Category, t.Priority, t.Status,
		t.RequesterID, t.RequesterName, t.RequesterEmail,
		t.AssigneeID, t.AssigneeName, t.DepartmentID, t.AffectedCIID,
		t.SLAResponseDeadline, t.SLAResolutionDeadline, t.SLAStatus,
		now, now,
	).Scan(&t.ID, &t.Version, &t.CreatedAt, &t.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert ticket: %w", err)
	}
	return nil
}

func (r *postgresRepository) UpdateTicketStatus(ctx context.Context, id, status string, assigneeID, assigneeName *string, resolvedAt, closedAt *time.Time, expectedVersion *int) error {
	var query string
	var res sql.Result
	var err error

	if expectedVersion != nil {
		query = `
			UPDATE tickets
			SET 
				status = $2,
				assignee_id = COALESCE($3, assignee_id),
				assignee_name = COALESCE($4, assignee_name),
				resolved_at = COALESCE($5, resolved_at),
				closed_at = COALESCE($6, closed_at),
				version = version + 1,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND version = $7
		`
		res, err = r.db.ExecContext(ctx, query, id, status, assigneeID, assigneeName, resolvedAt, closedAt, *expectedVersion)
	} else {
		query = `
			UPDATE tickets
			SET 
				status = $2,
				assignee_id = COALESCE($3, assignee_id),
				assignee_name = COALESCE($4, assignee_name),
				resolved_at = COALESCE($5, resolved_at),
				closed_at = COALESCE($6, closed_at),
				version = version + 1,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		res, err = r.db.ExecContext(ctx, query, id, status, assigneeID, assigneeName, resolvedAt, closedAt)
	}

	if err != nil {
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return appErrors.Conflict("optimistic lock conflict: ticket has been updated by another transaction or does not exist")
	}

	return nil
}

func (r *postgresRepository) AssignTicket(ctx context.Context, id, assigneeID, assigneeName string, expectedVersion *int) error {
	var query string
	var res sql.Result
	var err error

	if expectedVersion != nil {
		query = `
			UPDATE tickets
			SET 
				assignee_id = $2,
				assignee_name = $3,
				status = CASE WHEN status = 'OPEN' THEN 'ASSIGNED' ELSE status END,
				version = version + 1,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1 AND version = $4
		`
		res, err = r.db.ExecContext(ctx, query, id, assigneeID, assigneeName, *expectedVersion)
	} else {
		query = `
			UPDATE tickets
			SET 
				assignee_id = $2,
				assignee_name = $3,
				status = CASE WHEN status = 'OPEN' THEN 'ASSIGNED' ELSE status END,
				version = version + 1,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = $1
		`
		res, err = r.db.ExecContext(ctx, query, id, assigneeID, assigneeName)
	}

	if err != nil {
		return fmt.Errorf("failed to assign ticket: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return appErrors.Conflict("optimistic lock conflict: ticket has been updated by another transaction or does not exist")
	}

	return nil
}

func (r *postgresRepository) AddComment(ctx context.Context, c *model.TicketComment) error {
	query := `
		INSERT INTO ticket_comments (ticket_id, author_id, author_name, author_role, content, is_internal, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		c.TicketID, c.AuthorID, c.AuthorName, c.AuthorRole, c.Content, c.IsInternal, now,
	).Scan(&c.ID, &c.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert comment: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListComments(ctx context.Context, ticketID string) ([]model.TicketComment, error) {
	query := `
		SELECT id, ticket_id, author_id, author_name, author_role, content, is_internal, created_at
		FROM ticket_comments
		WHERE ticket_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to query ticket comments: %w", err)
	}
	defer rows.Close()

	comments := []model.TicketComment{}
	for rows.Next() {
		var c model.TicketComment
		if err := rows.Scan(&c.ID, &c.TicketID, &c.AuthorID, &c.AuthorName, &c.AuthorRole, &c.Content, &c.IsInternal, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}
		comments = append(comments, c)
	}
	return comments, nil
}

func (r *postgresRepository) AddTimelineRecord(ctx context.Context, t *model.TicketTimeline) error {
	query := `
		INSERT INTO ticket_timeline (ticket_id, actor_id, actor_name, action, old_value, new_value, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`
	now := time.Now()
	err := r.db.QueryRowContext(
		ctx, query,
		t.TicketID, t.ActorID, t.ActorName, t.Action, t.OldValue, t.NewValue, t.Notes, now,
	).Scan(&t.ID, &t.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert timeline: %w", err)
	}
	return nil
}

func (r *postgresRepository) ListTimeline(ctx context.Context, ticketID string) ([]model.TicketTimeline, error) {
	query := `
		SELECT id, ticket_id, actor_id, actor_name, action, old_value, new_value, notes, created_at
		FROM ticket_timeline
		WHERE ticket_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("failed to query timeline: %w", err)
	}
	defer rows.Close()

	records := []model.TicketTimeline{}
	for rows.Next() {
		var t model.TicketTimeline
		var oldVal, newVal, notes sql.NullString
		if err := rows.Scan(&t.ID, &t.TicketID, &t.ActorID, &t.ActorName, &t.Action, &oldVal, &newVal, &notes, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan timeline: %w", err)
		}
		if oldVal.Valid {
			t.OldValue = &oldVal.String
		}
		if newVal.Valid {
			t.NewValue = &newVal.String
		}
		if notes.Valid {
			t.Notes = &notes.String
		}
		records = append(records, t)
	}
	return records, nil
}

func (r *postgresRepository) ListServiceCategories(ctx context.Context) ([]model.ServiceCategory, error) {
	query := "SELECT id, name, icon, description, created_at FROM service_categories ORDER BY name ASC"
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	cats := []model.ServiceCategory{}
	for rows.Next() {
		var c model.ServiceCategory
		var desc sql.NullString
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &desc, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		if desc.Valid {
			c.Description = &desc.String
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (r *postgresRepository) ListServiceCatalogItems(ctx context.Context) ([]model.ServiceCatalogItem, error) {
	query := `
		SELECT 
			i.id, i.category_id, c.name AS category_name, i.name, i.code, i.description,
			i.default_priority, i.sla_response_minutes, i.sla_resolution_minutes,
			i.requires_approval, i.is_active, i.created_at
		FROM service_catalog_items i
		JOIN service_categories c ON i.category_id = c.id
		WHERE i.is_active = TRUE
		ORDER BY i.name ASC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query service items: %w", err)
	}
	defer rows.Close()

	items := []model.ServiceCatalogItem{}
	for rows.Next() {
		var item model.ServiceCatalogItem
		var catName, desc sql.NullString
		err := rows.Scan(
			&item.ID, &item.CategoryID, &catName, &item.Name, &item.Code, &desc,
			&item.DefaultPriority, &item.SLAResponseMinutes, &item.SLAResolutionMinutes,
			&item.RequiresApproval, &item.IsActive, &item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service item: %w", err)
		}
		if catName.Valid {
			item.CategoryName = &catName.String
		}
		if desc.Valid {
			item.Description = &desc.String
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *postgresRepository) FindServiceCatalogItemByID(ctx context.Context, id string) (*model.ServiceCatalogItem, error) {
	query := `
		SELECT 
			i.id, i.category_id, c.name AS category_name, i.name, i.code, i.description,
			i.default_priority, i.sla_response_minutes, i.sla_resolution_minutes,
			i.requires_approval, i.is_active, i.created_at
		FROM service_catalog_items i
		JOIN service_categories c ON i.category_id = c.id
		WHERE i.id = $1
	`
	var item model.ServiceCatalogItem
	var catName, desc sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&item.ID, &item.CategoryID, &catName, &item.Name, &item.Code, &desc,
		&item.DefaultPriority, &item.SLAResponseMinutes, &item.SLAResolutionMinutes,
		&item.RequiresApproval, &item.IsActive, &item.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query service item: %w", err)
	}
	if catName.Valid {
		item.CategoryName = &catName.String
	}
	if desc.Valid {
		item.Description = &desc.String
	}
	return &item, nil
}
