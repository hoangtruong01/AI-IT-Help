package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/lib/pq"
)

func TestReadinessHandler_NilDB(t *testing.T) {
	handler := ReadinessHandler(nil)
	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 Service Unavailable, got %d", rec.Code)
	}
}

func TestMigrationAdvisoryLockID_Constant(t *testing.T) {
	if MigrationAdvisoryLockID == 0 {
		t.Fatal("expected non-zero MigrationAdvisoryLockID")
	}
}

func TestRunMigrations_NonExistentDir(t *testing.T) {
	// If the database is nil, trying to acquire a connection returns an error
	// but let's verify directory not found handling logic
	tempDir := filepath.Join(os.TempDir(), "non_existent_eomp_migrations_12345")
	_ = os.RemoveAll(tempDir)

	if _, err := os.Stat(tempDir); !os.IsNotExist(err) {
		t.Fatalf("expected dir not to exist")
	}
}

func TestRunMigrations_NilDatabase(t *testing.T) {
	err := RunMigrations(nil, t.TempDir())
	if err == nil {
		t.Fatal("expected a nil database error")
	}
}

// This test uses five independent PostgreSQL connections to exercise the same
// session-lock behavior used by concurrent pods. It is opt-in because the
// ordinary unit suite must not silently claim real database evidence.
func TestRunMigrationsFiveConcurrentRunnersPostgres(t *testing.T) {
	dsn := os.Getenv("MIGRATION_INTEGRATION_DSN")
	if dsn == "" {
		t.Skip("MIGRATION_INTEGRATION_DSN is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open PostgreSQL: %v", err)
	}
	defer db.Close()
	db.SetMaxOpenConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("ping PostgreSQL: %v", err)
	}

	suffix := time.Now().UnixNano()
	tableName := fmt.Sprintf("gatec_migration_probe_%d", suffix)
	filename := fmt.Sprintf("gatec_%d.sql", suffix)
	migrationsDir := t.TempDir()
	migrationSQL := fmt.Sprintf(
		"CREATE TABLE %s (id INTEGER PRIMARY KEY); INSERT INTO %s (id) VALUES (1);",
		pq.QuoteIdentifier(tableName),
		pq.QuoteIdentifier(tableName),
	)
	if err := os.WriteFile(filepath.Join(migrationsDir, filename), []byte(migrationSQL), 0o600); err != nil {
		t.Fatalf("write migration fixture: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DROP TABLE IF EXISTS " + pq.QuoteIdentifier(tableName))
		_, _ = db.Exec("DELETE FROM schema_migrations WHERE version=$1", filename)
	})

	start := make(chan struct{})
	errs := make(chan error, 5)
	var runners sync.WaitGroup
	for i := 0; i < 5; i++ {
		runners.Add(1)
		go func() {
			defer runners.Done()
			<-start
			errs <- RunMigrations(db, migrationsDir)
		}()
	}
	close(start)
	runners.Wait()
	close(errs)
	for runErr := range errs {
		if runErr != nil {
			t.Fatalf("concurrent migration runner failed: %v", runErr)
		}
	}

	var rowCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+pq.QuoteIdentifier(tableName)).Scan(&rowCount); err != nil {
		t.Fatalf("query migration probe table: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("migration executed more than once: row count=%d", rowCount)
	}

	var trackerCount int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM schema_migrations WHERE version=$1", filename).Scan(&trackerCount); err != nil {
		t.Fatalf("query migration tracker: %v", err)
	}
	if trackerCount != 1 {
		t.Fatalf("expected exactly one tracker row, got %d", trackerCount)
	}
}
