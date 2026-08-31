package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"eomp/services/auth/internal/model"
)

var ErrInvalidRefreshToken = errors.New("refresh token is no longer active")

// UserRepository defines database operations for users and tokens
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	CreateWithAudit(ctx context.Context, user *model.User, audit model.SecurityAuditLog) error
	Update(ctx context.Context, user *model.User) error
	UpdateWithAudit(ctx context.Context, user *model.User, revokeTokens bool, audit model.SecurityAuditLog) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
	UpdatePasswordAndRevoke(ctx context.Context, userID, passwordHash string, audit model.SecurityAuditLog) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	ListUsers(ctx context.Context, query model.UserListQuery) (*model.UserListResponse, error)
	SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	RevokeAllUserRefreshTokens(ctx context.Context, userID string) error
	RotateRefreshTokenAtomic(ctx context.Context, oldTokenHash, userID, newTokenHash string, expiresAt time.Time) error
	ValidateRefreshToken(ctx context.Context, tokenHash string) (string, error)
	RecordLoginAudit(ctx context.Context, log *model.LoginAuditLog) error
	GetLoginHistory(ctx context.Context, email string, limit int) ([]model.LoginAuditLog, error)
}

type postgresUserRepository struct {
	db *sql.DB
}

// NewUserRepository constructs a new PostgreSQL UserRepository
func NewUserRepository(db *sql.DB) UserRepository {
	return &postgresUserRepository{db: db}
}

func (r *postgresUserRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (email, password_hash, full_name, role, department_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := r.db.QueryRowContext(
		ctx, query,
		user.Email, user.PasswordHash, user.FullName, user.Role, user.DepartmentID, user.IsActive, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.ID)

	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) CreateWithAudit(ctx context.Context, user *model.User, audit model.SecurityAuditLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin user creation transaction: %w", err)
	}
	defer tx.Rollback()

	now := time.Now()
	user.CreatedAt, user.UpdatedAt = now, now
	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, full_name, role, department_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
	`, user.Email, user.PasswordHash, user.FullName, user.Role, user.DepartmentID, user.IsActive, now, now).Scan(&user.ID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}
	audit.TargetUserID = user.ID
	if err := insertSecurityAudit(ctx, tx, audit); err != nil {
		return err
	}
	return tx.Commit()
}

type auditExecer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

func insertSecurityAudit(ctx context.Context, execer auditExecer, audit model.SecurityAuditLog) error {
	_, err := execer.ExecContext(ctx, `
		INSERT INTO security_audit_logs (actor_id, actor_email, action, target_user_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, audit.ActorID, audit.ActorEmail, audit.Action, audit.TargetUserID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to record security audit event: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, department_id, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.DepartmentID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to query user by email: %w", err)
	}
	return &user, nil
}

func (r *postgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, full_name, role, department_id, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.PasswordHash, &user.FullName, &user.Role, &user.DepartmentID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query user by id: %w", err)
	}
	return &user, nil
}

func (r *postgresUserRepository) SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, FALSE, $4)
	`
	_, err := r.db.ExecContext(ctx, query, userID, tokenHash, expiresAt, time.Now())
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE token_hash = $1
	`
	_, err := r.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) ValidateRefreshToken(ctx context.Context, tokenHash string) (string, error) {
	query := `
		SELECT user_id
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked = FALSE AND expires_at > CURRENT_TIMESTAMP
	`
	var userID string
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(&userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", nil
		}
		return "", fmt.Errorf("failed to validate refresh token: %w", err)
	}
	return userID, nil
}

func (r *postgresUserRepository) RecordLoginAudit(ctx context.Context, log *model.LoginAuditLog) error {
	query := `
		INSERT INTO login_audit_logs (user_id, email, ip_address, user_agent, status, failure_reason, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	now := time.Now()
	log.CreatedAt = now
	err := r.db.QueryRowContext(
		ctx, query,
		log.UserID, log.Email, log.IPAddress, log.UserAgent, log.Status, log.FailureReason, now,
	).Scan(&log.ID)
	if err != nil {
		return fmt.Errorf("failed to record login audit log: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) GetLoginHistory(ctx context.Context, email string, limit int) ([]model.LoginAuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var query string
	var args []any
	if email != "" {
		query = `
			SELECT id, user_id, email, ip_address, user_agent, status, failure_reason, created_at
			FROM login_audit_logs
			WHERE email = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		args = []any{email, limit}
	} else {
		query = `
			SELECT id, user_id, email, ip_address, user_agent, status, failure_reason, created_at
			FROM login_audit_logs
			ORDER BY created_at DESC
			LIMIT $1
		`
		args = []any{limit}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query login audit logs: %w", err)
	}
	defer rows.Close()

	logs := []model.LoginAuditLog{}
	for rows.Next() {
		var l model.LoginAuditLog
		var uid, uAgent, fReason sql.NullString
		err := rows.Scan(
			&l.ID, &uid, &l.Email, &l.IPAddress, &uAgent, &l.Status, &fReason, &l.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan login audit log: %w", err)
		}
		if uid.Valid {
			l.UserID = &uid.String
		}
		if uAgent.Valid {
			l.UserAgent = &uAgent.String
		}
		if fReason.Valid {
			l.FailureReason = &fReason.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (r *postgresUserRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users
		SET full_name = $1, role = $2, department_id = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`
	user.UpdatedAt = time.Now()
	res, err := r.db.ExecContext(ctx, query, user.FullName, user.Role, user.DepartmentID, user.IsActive, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *postgresUserRepository) UpdateWithAudit(ctx context.Context, user *model.User, revokeTokens bool, audit model.SecurityAuditLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin user update transaction: %w", err)
	}
	defer tx.Rollback()

	user.UpdatedAt = time.Now()
	res, err := tx.ExecContext(ctx, `
		UPDATE users SET full_name=$1, role=$2, department_id=$3, is_active=$4, updated_at=$5 WHERE id=$6
	`, user.FullName, user.Role, user.DepartmentID, user.IsActive, user.UpdatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		return errors.New("user not found")
	}
	if revokeTokens {
		if _, err := tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked=TRUE WHERE user_id=$1`, user.ID); err != nil {
			return fmt.Errorf("failed to revoke user refresh tokens: %w", err)
		}
	}
	if err := insertSecurityAudit(ctx, tx, audit); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *postgresUserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, updated_at = $2
		WHERE id = $3
	`
	res, err := r.db.ExecContext(ctx, query, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

func (r *postgresUserRepository) UpdatePasswordAndRevoke(ctx context.Context, userID, passwordHash string, audit model.SecurityAuditLog) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin password transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `UPDATE users SET password_hash=$1, updated_at=$2 WHERE id=$3`, passwordHash, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil || rows != 1 {
		return errors.New("user not found")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE refresh_tokens SET revoked=TRUE WHERE user_id=$1`, userID); err != nil {
		return fmt.Errorf("failed to revoke user refresh tokens: %w", err)
	}
	if err := insertSecurityAudit(ctx, tx, audit); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *postgresUserRepository) RevokeAllUserRefreshTokens(ctx context.Context, userID string) error {
	query := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE user_id = $1
	`
	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke all refresh tokens for user: %w", err)
	}
	return nil
}

func (r *postgresUserRepository) RotateRefreshTokenAtomic(ctx context.Context, oldTokenHash, userID, newTokenHash string, expiresAt time.Time) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Consume exactly one active token owned by this user. The predicate makes
	// concurrent/replayed refresh attempts fail before a replacement is issued.
	revokeQuery := `
		UPDATE refresh_tokens
		SET revoked = TRUE
		WHERE token_hash = $1 AND user_id = $2 AND revoked = FALSE AND expires_at > CURRENT_TIMESTAMP
	`
	res, err := tx.ExecContext(ctx, revokeQuery, oldTokenHash, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke old refresh token: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to verify refresh token consumption: %w", err)
	}
	if rows != 1 {
		return ErrInvalidRefreshToken
	}

	// 2. Insert new token
	insertQuery := `
		INSERT INTO refresh_tokens (user_id, token_hash, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, FALSE, $4)
	`
	if _, err := tx.ExecContext(ctx, insertQuery, userID, newTokenHash, expiresAt, time.Now()); err != nil {
		return fmt.Errorf("failed to save new refresh token: %w", err)
	}

	return tx.Commit()
}

func (r *postgresUserRepository) ListUsers(ctx context.Context, query model.UserListQuery) (*model.UserListResponse, error) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 || query.PageSize > 100 {
		query.PageSize = 20
	}

	whereClauses := []string{"1=1"}
	var args []any
	argIdx := 1

	if query.Role != "" && query.Role != "All" {
		whereClauses = append(whereClauses, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, query.Role)
		argIdx++
	}

	if query.DepartmentID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("department_id = $%d", argIdx))
		args = append(args, query.DepartmentID)
		argIdx++
	}

	if query.Search != "" {
		pattern := "%" + strings.ToLower(query.Search) + "%"
		whereClauses = append(whereClauses, fmt.Sprintf("(LOWER(email) LIKE $%d OR LOWER(full_name) LIKE $%d)", argIdx, argIdx))
		args = append(args, pattern)
		argIdx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users WHERE %s", whereSQL)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	offset := (query.Page - 1) * query.PageSize
	dataQuery := fmt.Sprintf(`
		SELECT id, email, full_name, role, department_id, is_active, created_at
		FROM users
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereSQL, argIdx, argIdx+1)
	args = append(args, query.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := []model.UserResponse{}
	for rows.Next() {
		var u model.UserResponse
		var deptID sql.NullString
		err := rows.Scan(
			&u.ID, &u.Email, &u.FullName, &u.Role, &deptID, &u.IsActive, &u.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		if deptID.Valid {
			u.DepartmentID = &deptID.String
		}
		users = append(users, u)
	}

	totalPages := (total + query.PageSize - 1) / query.PageSize
	return &model.UserListResponse{
		Data:       users,
		Total:      total,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}, nil
}
