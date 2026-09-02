package integration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"eomp/packages/shared/pkg/database"

	_ "github.com/lib/pq"
)

func getIntegrationPostgresDB(t *testing.T) *sql.DB {
	t.Helper()
	required := os.Getenv("INTEGRATION_REQUIRED") != ""
	dsn := os.Getenv("INTEGRATION_POSTGRES_DSN")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		if required {
			t.Fatal("INTEGRATION_POSTGRES_DSN is required")
		}
		t.Skip("skipping PostgreSQL integration test (INTEGRATION_POSTGRES_DSN not set)")
		return nil
	}
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		if required {
			t.Fatalf("open integration PostgreSQL: %v", err)
		}
		t.Skipf("cannot open PostgreSQL: %v", err)
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		if required {
			t.Fatalf("ping integration PostgreSQL: %v", err)
		}
		t.Skipf("cannot ping PostgreSQL: %v", err)
		return nil
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestPostgresIntegration_MigrationConcurrency applies migrations with 5 concurrent runners.
func TestPostgresIntegration_MigrationConcurrency(t *testing.T) {
	db := getIntegrationPostgresDB(t)
	if db == nil {
		return
	}

	suffix := time.Now().UnixNano()
	tableName := fmt.Sprintf("gate_d_migration_%d", suffix)
	filename := fmt.Sprintf("gate_d_%d.sql", suffix)
	tmpDir := t.TempDir()

	migrationSQL := fmt.Sprintf(
		"CREATE TABLE %s (id SERIAL PRIMARY KEY, note TEXT); INSERT INTO %s (note) VALUES ('gate_d_verified');",
		tableName, tableName,
	)
	if err := os.WriteFile(filepath.Join(tmpDir, filename), []byte(migrationSQL), 0o600); err != nil {
		t.Fatalf("write migration file: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DROP TABLE IF EXISTS " + tableName)
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
			errs <- database.RunMigrations(db, tmpDir)
		}()
	}
	close(start)
	runners.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent migration failed: %v", err)
		}
	}

	var rowCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM " + tableName).Scan(&rowCount); err != nil {
		t.Fatalf("query test table: %v", err)
	}
	if rowCount != 1 {
		t.Fatalf("migration executed multiple times: count=%d", rowCount)
	}
}

// TestPostgresIntegration_SQLInjectionResistance verifies parameterized query safety.
func TestPostgresIntegration_SQLInjectionResistance(t *testing.T) {
	db := getIntegrationPostgresDB(t)
	if db == nil {
		return
	}

	tableName := fmt.Sprintf("gate_d_sqli_%d", time.Now().UnixNano())
	_, err := db.Exec(fmt.Sprintf("CREATE TABLE %s (id SERIAL PRIMARY KEY, title TEXT, category TEXT);", tableName))
	if err != nil {
		t.Fatalf("create test table: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DROP TABLE IF EXISTS " + tableName)
	})

	payloads := []string{
		"' OR '1'='1",
		"'; DROP TABLE " + tableName + "; --",
		"1' UNION SELECT 1, 'injected', 'data'--",
		"admin'--",
	}

	for _, p := range payloads {
		// Parameterized INSERT
		_, err := db.Exec(fmt.Sprintf("INSERT INTO %s (title, category) VALUES ($1, $2)", tableName), p, "hardware")
		if err != nil {
			t.Fatalf("parameterized insert failed for payload %q: %v", p, err)
		}

		// Parameterized SELECT
		var title string
		err = db.QueryRow(fmt.Sprintf("SELECT title FROM %s WHERE title = $1", tableName), p).Scan(&title)
		if err != nil {
			t.Fatalf("parameterized select failed for payload %q: %v", p, err)
		}
		if title != p {
			t.Fatalf("data mismatch: expected %q, got %q", p, title)
		}
	}

	// Verify table still exists
	var count int
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count); err != nil {
		t.Fatalf("table was corrupted by injection payload: %v", err)
	}
	if count != len(payloads) {
		t.Fatalf("expected %d rows, got %d", len(payloads), count)
	}
}
