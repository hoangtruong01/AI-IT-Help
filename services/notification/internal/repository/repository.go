package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"eomp/services/notification/internal/model"
)

// Repository interface for Notifications
type Repository interface {
	ListNotifications(ctx context.Context, query model.NotificationListQuery) (*model.NotificationListResponse, error)
	FindNotificationByID(ctx context.Context, id string) (*model.Notification, error)
	CreateNotification(ctx context.Context, n *model.Notification) error
	MarkAsRead(ctx context.Context, id string) error
	MarkAllAsRead(ctx context.Context, recipientID string) error
	GetStats(ctx context.Context, recipientID string) (*model.NotificationStats, error)
}

type postgresRepository struct {
	db *sql.DB
}

// NewRepository constructs a PostgreSQL Notification repository
func NewRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) ListNotifications(ctx context.Context, query model.NotificationListQuery) (*model.NotificationListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	whereClauses := []string{"1=1"}
	args := []any{}
	idx := 1

	if query.RecipientID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("(recipient_id = $%d OR recipient_id = 'all')", idx))
		args = append(args, query.RecipientID)
		idx++
	}

	if query.IsRead != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_read = $%d", idx))
		args = append(args, *query.IsRead)
		idx++
	}

	if query.Category != "" && query.Category != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("category = $%d", idx))
		args = append(args, query.Category)
		idx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM notifications WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count notifications: %w", err)
	}

	var unreadCount int
	unreadQuery := "SELECT COUNT(*) FROM notifications WHERE is_read = FALSE"
	if query.RecipientID != "" {
		unreadQuery += fmt.Sprintf(" AND (recipient_id = '%s' OR recipient_id = 'all')", query.RecipientID)
	}
	_ = r.db.QueryRowContext(ctx, unreadQuery).Scan(&unreadCount)

	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, recipient_id, recipient_email, title, message, category, priority, is_read, read_at, channel, metadata::text, created_at
		FROM notifications
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, idx, idx+1)

	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}
	defer rows.Close()

	list := []model.Notification{}
	for rows.Next() {
		var n model.Notification
		var metaStr sql.NullString
		var readAt sql.NullTime

		err := rows.Scan(
			&n.ID, &n.RecipientID, &n.RecipientEmail, &n.Title, &n.Message,
			&n.Category, &n.Priority, &n.IsRead, &readAt, &n.Channel, &metaStr, &n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notification: %w", err)
		}

		if metaStr.Valid {
			n.Metadata = &metaStr.String
		}
		if readAt.Valid {
			n.ReadAt = &readAt.Time
		}

		list = append(list, n)
	}

	totalPages := int(math.Ceil(float64(total) / float64(query.PageSize)))

	return &model.NotificationListResponse{
		Data:        list,
		Total:       total,
		UnreadCount: unreadCount,
		Page:        query.Page,
		PageSize:    query.PageSize,
		TotalPages:  totalPages,
	}, nil
}

func (r *postgresRepository) FindNotificationByID(ctx context.Context, id string) (*model.Notification, error) {
	query := `
		SELECT id, recipient_id, recipient_email, title, message, category, priority, is_read, read_at, channel, metadata::text, created_at
		FROM notifications
		WHERE id = $1
	`
	var n model.Notification
	var metaStr sql.NullString
	var readAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&n.ID, &n.RecipientID, &n.RecipientEmail, &n.Title, &n.Message,
		&n.Category, &n.Priority, &n.IsRead, &readAt, &n.Channel, &metaStr, &n.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get notification: %w", err)
	}

	if metaStr.Valid {
		n.Metadata = &metaStr.String
	}
	if readAt.Valid {
		n.ReadAt = &readAt.Time
	}

	return &n, nil
}

func (r *postgresRepository) CreateNotification(ctx context.Context, n *model.Notification) error {
	query := `
		INSERT INTO notifications (
			recipient_id, recipient_email, title, message, category, priority, is_read, channel, metadata, created_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, $10
		)
		RETURNING id, created_at
	`
	now := time.Now()
	meta := "{}"
	if n.Metadata != nil && *n.Metadata != "" {
		meta = *n.Metadata
	}

	err := r.db.QueryRowContext(
		ctx, query,
		n.RecipientID, n.RecipientEmail, n.Title, n.Message,
		n.Category, n.Priority, n.IsRead, n.Channel, meta, now,
	).Scan(&n.ID, &n.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert notification: %w", err)
	}
	return nil
}

func (r *postgresRepository) MarkAsRead(ctx context.Context, id string) error {
	now := time.Now()
	query := "UPDATE notifications SET is_read = TRUE, read_at = $2 WHERE id = $1"
	_, err := r.db.ExecContext(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to mark notification as read: %w", err)
	}
	return nil
}

func (r *postgresRepository) MarkAllAsRead(ctx context.Context, recipientID string) error {
	now := time.Now()
	query := "UPDATE notifications SET is_read = TRUE, read_at = $2 WHERE (recipient_id = $1 OR recipient_id = 'all') AND is_read = FALSE"
	_, err := r.db.ExecContext(ctx, query, recipientID, now)
	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}
	return nil
}

func (r *postgresRepository) GetStats(ctx context.Context, recipientID string) (*model.NotificationStats, error) {
	query := `
		SELECT 
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE is_read = FALSE) AS unread,
			COUNT(*) FILTER (WHERE category = 'INCIDENT') AS incidents,
			COUNT(*) FILTER (WHERE category = 'APPROVAL') AS approvals
		FROM notifications
		WHERE recipient_id = $1 OR recipient_id = 'all'
	`
	var s model.NotificationStats
	err := r.db.QueryRowContext(ctx, query, recipientID).Scan(&s.Total, &s.Unread, &s.Incidents, &s.Approvals)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification stats: %w", err)
	}
	return &s, nil
}
