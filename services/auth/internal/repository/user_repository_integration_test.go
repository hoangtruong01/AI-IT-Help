package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"eomp/services/auth/internal/model"
	_ "github.com/lib/pq"
)

func TestCreateWithAuditRollsBackUserWhenAuditInsertFailsPostgres(t *testing.T) {
	dsn := os.Getenv("AUTH_INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("AUTH_INTEGRATION_DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	email := fmt.Sprintf("gateb-rollback-%d@local.test", time.Now().UnixNano())
	defer db.ExecContext(context.Background(), "DELETE FROM users WHERE email=$1", email)

	user := &model.User{
		Email:        email,
		PasswordHash: "integration-test-password-hash",
		FullName:     "Gate B Rollback Test",
		Role:         model.RoleEmployee,
		IsActive:     true,
	}
	invalidActorID := "not-a-uuid"
	err = NewUserRepository(db).CreateWithAudit(ctx, user, model.SecurityAuditLog{
		ActorID:    &invalidActorID,
		ActorEmail: "gateb-rollback-actor@local.test",
		Action:     "USER_CREATED",
	})
	if err == nil {
		t.Fatal("expected audit insert failure")
	}

	var userCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email=$1", email).Scan(&userCount); err != nil {
		t.Fatalf("query rolled-back user: %v", err)
	}
	if userCount != 0 {
		t.Fatalf("user insert was not rolled back: count=%d", userCount)
	}

	var auditCount int
	if user.ID != "" {
		if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM security_audit_logs WHERE target_user_id=$1", user.ID).Scan(&auditCount); err != nil {
			t.Fatalf("query rolled-back audit: %v", err)
		}
	}
	if auditCount != 0 {
		t.Fatalf("audit insert was not rolled back: count=%d", auditCount)
	}
}
