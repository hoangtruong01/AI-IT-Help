package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"eomp/services/auth/internal/model"
)

// UserRepository defines database operations for users and tokens
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	SaveRefreshToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	ValidateRefreshToken(ctx context.Context, tokenHash string) (string, error)
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
