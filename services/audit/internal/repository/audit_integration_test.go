package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"eomp/services/audit/internal/model"

	_ "github.com/lib/pq"
)

func getAuditTestDB(t *testing.T) *sql.DB {
	t.Helper()
	required := os.Getenv("INTEGRATION_REQUIRED") != ""
	dsn := os.Getenv("AUDIT_INTEGRATION_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		if required {
			t.Fatal("AUDIT_INTEGRATION_DSN is required")
		}
		t.Skip("skipping audit PostgreSQL integration test (AUDIT_INTEGRATION_DSN not set)")
		return nil
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		if required {
			t.Fatalf("open audit PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot open db: %v", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if required {
			t.Fatalf("ping audit PostgreSQL: %v", err)
		}
		t.Skipf("skipping: cannot ping db: %v", err)
		return nil
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestAuditIntegration_HMACChainIntegrity validates that append-only audit records
// maintain a valid cryptographic hash chain and pass integrity validation.
func TestAuditIntegration_HMACChainIntegrity(t *testing.T) {
	db := getAuditTestDB(t)
	if db == nil {
		return
	}

	hmacKey := "audit-secret-key-at-least-32-chars-long-2026"
	repo := NewRepository(db, hmacKey)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	idBase := time.Now().UnixNano() % 999999999000

	for i := 1; i <= 3; i++ {
		log := &model.AuditLog{
			ID:           fmt.Sprintf("00000000-0000-4000-8000-%012d", idBase+int64(i)),
			EventType:    fmt.Sprintf("INTEGRATION_EVENT_%d", i),
			ActorID:      "u-audit-actor-1",
			ActorName:    "Audit Tester",
			ActorEmail:   "audit.tester@eomp.local",
			ActorRole:    "ROLE_ADMIN",
			ServiceName:  "helpdesk",
			IPAddress:    "127.0.0.1",
			UserAgent:    "eomp-integration-agent",
			Status:       "SUCCESS",
			ResourceType: "TICKET",
			ResourceID:   fmt.Sprintf("TK-AUDIT-%d", i),
			OldValues:    map[string]any{"status": "OPEN"},
			NewValues:    map[string]any{"status": "IN_PROGRESS"},
			CreatedAt:    time.Now(),
		}

		if err := repo.CreateAuditLog(ctx, log); err != nil {
			t.Fatalf("failed to create audit log %d: %v", i, err)
		}
	}

	report, err := repo.VerifyIntegrity(ctx)
	if err != nil {
		t.Fatalf("failed to verify audit integrity: %v", err)
	}

	if !report.Valid {
		t.Fatalf("expected audit chain to be valid, but report marked it invalid: %+v", report)
	}
	if report.TotalLogs < 3 {
		t.Fatalf("expected at least 3 records in audit log, got %d", report.TotalLogs)
	}
	if report.VerifiedLogs != report.TotalLogs {
		t.Fatalf("mismatch: verified logs=%d, total logs=%d", report.VerifiedLogs, report.TotalLogs)
	}
}

// TestAuditIntegration_TamperDetection verifies that any unauthorized mutation
// breaks the HMAC chain and is immediately flagged by VerifyIntegrity.
func TestAuditIntegration_TamperDetection(t *testing.T) {
	db := getAuditTestDB(t)
	if db == nil {
		return
	}

	hmacKey := "audit-secret-key-at-least-32-chars-long-2026"
	repo := NewRepository(db, hmacKey)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tamperLog := &model.AuditLog{
		ID:           fmt.Sprintf("00000000-0000-4000-8000-%012d", time.Now().UnixNano()%999999999999),
		EventType:    "SENSITIVE_ROLE_CHANGE",
		ActorID:      "u-secops-admin",
		ActorName:    "SecOps Admin",
		ActorEmail:   "secops@eomp.local",
		ActorRole:    "ROLE_ADMIN",
		ServiceName:  "auth",
		IPAddress:    "10.0.0.1",
		UserAgent:    "eomp-admin-cli",
		Status:       "SUCCESS",
		ResourceType: "USER",
		ResourceID:   "u-target-user-001",
		OldValues:    map[string]any{"role": "ROLE_EMPLOYEE"},
		NewValues:    map[string]any{"role": "ROLE_AGENT"},
		CreatedAt:    time.Now(),
	}

	if err := repo.CreateAuditLog(ctx, tamperLog); err != nil {
		t.Fatalf("failed to create audit log: %v", err)
	}

	initReport, err := repo.VerifyIntegrity(ctx)
	if err != nil {
		t.Fatalf("failed to verify initial audit integrity: %v", err)
	}
	if !initReport.Valid {
		t.Fatalf("initial chain should be untampered")
	}

	_, err = db.ExecContext(ctx, "UPDATE audit_logs SET actor_email = 'attacker@malicious.evil' WHERE id = $1", tamperLog.ID)
	if err == nil {
		tamperReport, err := repo.VerifyIntegrity(ctx)
		if err != nil {
			t.Fatalf("failed to run verify after tamper: %v", err)
		}
		if tamperReport.Valid {
			t.Fatalf("tamper detection failed: HMAC check did not detect modified actor_email")
		}
	} else {
		t.Logf("Database append-only trigger prevented direct modification: %v", err)
	}
}
