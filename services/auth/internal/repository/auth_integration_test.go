package repository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"testing"
	"time"

	"eomp/services/auth/internal/model"

	_ "github.com/lib/pq"
)

func getAuthTestDB(t *testing.T) *sql.DB {
	t.Helper()
	required := os.Getenv("INTEGRATION_REQUIRED") != ""
	dsn := os.Getenv("AUTH_INTEGRATION_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		if required {
			t.Fatal("AUTH_INTEGRATION_DSN is required")
		}
		t.Skip("skipping auth PostgreSQL integration test (AUTH_INTEGRATION_DSN not set)")
		return nil
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		if required {
			t.Fatalf("open auth PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot open db: %v", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if required {
			t.Fatalf("ping auth PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot ping db: %v", err)
		return nil
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// TestAuthIntegration_AtomicRefreshTokenRotation validates that refresh token rotation
// revokes the previous token and activates the new token within a single atomic SQL transaction.
func TestAuthIntegration_AtomicRefreshTokenRotation(t *testing.T) {
	db := getAuthTestDB(t)
	if db == nil {
		return
	}

	repo := NewUserRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uniqueEmail := fmt.Sprintf("test.rotate.%d@eomp.local", time.Now().UnixNano())
	dept := "00000000-0000-4000-8000-000000000101"
	user := &model.User{
		Email:        uniqueEmail,
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
		FullName:     "Integration Test User",
		Role:         "ROLE_EMPLOYEE",
		DepartmentID: &dept,
		IsActive:     true,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	token1 := fmt.Sprintf("token_v1_%d", time.Now().UnixNano())
	hash1 := hashToken(token1)
	exp1 := time.Now().Add(24 * time.Hour)

	if err := repo.SaveRefreshToken(ctx, user.ID, hash1, exp1); err != nil {
		t.Fatalf("failed to save initial refresh token: %v", err)
	}

	validatedUID, err := repo.ValidateRefreshToken(ctx, hash1)
	if err != nil || validatedUID != user.ID {
		t.Fatalf("expected token 1 to be active, got uid=%s, err=%v", validatedUID, err)
	}

	// Rotate token 1 -> token 2 atomically
	token2 := fmt.Sprintf("token_v2_%d", time.Now().UnixNano())
	hash2 := hashToken(token2)
	exp2 := time.Now().Add(24 * time.Hour)

	if err := repo.RotateRefreshTokenAtomic(ctx, hash1, user.ID, hash2, exp2); err != nil {
		t.Fatalf("failed to rotate refresh token atomically: %v", err)
	}

	// 1. Verify token 2 is active
	validatedUID2, err := repo.ValidateRefreshToken(ctx, hash2)
	if err != nil || validatedUID2 != user.ID {
		t.Fatalf("expected token 2 to be active, got uid=%s, err=%v", validatedUID2, err)
	}

	// 2. Replay test: Token 1 must now be rejected
	replayedUID, err := repo.ValidateRefreshToken(ctx, hash1)
	if err != nil {
		t.Fatalf("failed to validate revoked token state: %v", err)
	}
	if replayedUID != "" {
		t.Fatalf("expected token 1 to be revoked upon replay, got active user %s", replayedUID)
	}
}

// TestAuthIntegration_UserDeactivationRevokesAllSessions verifies that deactivating
// a user revokes all active refresh tokens for that user.
func TestAuthIntegration_UserDeactivationRevokesAllSessions(t *testing.T) {
	db := getAuthTestDB(t)
	if db == nil {
		return
	}

	repo := NewUserRepository(db)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uniqueEmail := fmt.Sprintf("test.deactivate.%d@eomp.local", time.Now().UnixNano())
	dept := "00000000-0000-4000-8000-000000000102"
	user := &model.User{
		Email:        uniqueEmail,
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
		FullName:     "Deactivate Test User",
		Role:         "ROLE_EMPLOYEE",
		DepartmentID: &dept,
		IsActive:     true,
	}

	if err := repo.Create(ctx, user); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	h1 := hashToken(fmt.Sprintf("t1_%d", time.Now().UnixNano()))
	h2 := hashToken(fmt.Sprintf("t2_%d", time.Now().UnixNano()))
	exp := time.Now().Add(24 * time.Hour)

	if err := repo.SaveRefreshToken(ctx, user.ID, h1, exp); err != nil {
		t.Fatalf("failed to save token 1: %v", err)
	}
	if err := repo.SaveRefreshToken(ctx, user.ID, h2, exp); err != nil {
		t.Fatalf("failed to save token 2: %v", err)
	}

	user.IsActive = false
	actorID := user.ID
	audit := model.SecurityAuditLog{
		ActorID:      &actorID,
		ActorEmail:   "admin@eomp.local",
		Action:       "DEACTIVATE_USER",
		TargetUserID: user.ID,
	}

	if err := repo.UpdateWithAudit(ctx, user, true, audit); err != nil {
		t.Fatalf("failed to update user with audit and token revocation: %v", err)
	}

	if uid, err := repo.ValidateRefreshToken(ctx, h1); err != nil || uid != "" {
		t.Fatalf("expected token 1 to be revoked after deactivation, uid=%q err=%v", uid, err)
	}
	if uid, err := repo.ValidateRefreshToken(ctx, h2); err != nil || uid != "" {
		t.Fatalf("expected token 2 to be revoked after deactivation, uid=%q err=%v", uid, err)
	}
}

// TestAuthIntegration_TransactionRollbackSafety verifies that failure in audit logging
// rolls back the entire transaction, leaving no partial state.
func TestAuthIntegration_TransactionRollbackSafety(t *testing.T) {
	db := getAuthTestDB(t)
	if db == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	uniqueEmail := fmt.Sprintf("test.rollback.%d@eomp.local", time.Now().UnixNano())

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	var newUserID string
	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (email, password_hash, full_name, role, department_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id
	`, uniqueEmail, "hash123", "Rollback User", "ROLE_EMPLOYEE", "00000000-0000-4000-8000-000000000101", true, time.Now(), time.Now()).Scan(&newUserID)
	if err != nil {
		tx.Rollback()
		t.Fatalf("failed to insert user in tx: %v", err)
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO non_existent_audit_table (dummy) VALUES (1)`)
	if err == nil {
		tx.Commit()
		t.Fatalf("expected error from non existent table")
	}
	tx.Rollback()

	var count int
	err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", uniqueEmail).Scan(&count)
	if err != nil {
		t.Fatalf("failed to query users: %v", err)
	}
	if count != 0 {
		t.Fatalf("transaction rollback failed: user still exists in database (count=%d)", count)
	}
}
