package database

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// ReadinessHandler reports whether the database dependency is currently usable.
func ReadinessHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ready"}`))
	}
}

// Config holds DB connection settings
type Config struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

// Connect initializes a connection pool to PostgreSQL
func Connect(cfg Config) (*sql.DB, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	if cfg.MaxOpenConns == 0 {
		cfg.MaxOpenConns = 25
	}
	if cfg.MaxIdleConns == 0 {
		cfg.MaxIdleConns = 10
	}
	if cfg.ConnMaxLifetime == 0 {
		cfg.ConnMaxLifetime = 5 * time.Minute
	}
	if cfg.ConnMaxIdleTime == 0 {
		cfg.ConnMaxIdleTime = 2 * time.Minute
	}

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database handle: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Verify connection with timeout
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database %s: %w", cfg.DBName, err)
	}

	return db, nil
}

// MigrationAdvisoryLockID is a deterministic 64-bit key used for synchronizing schema migrations across pods
const MigrationAdvisoryLockID int64 = 8424119472649191

// RunMigrations executes all .sql migration files in ascending order from migrationsDir
// protected by a PostgreSQL session-level advisory lock to guarantee concurrency safety.
func RunMigrations(db *sql.DB, migrationsDir string) error {
	if db == nil {
		return fmt.Errorf("database handle is required for migrations")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Obtain a dedicated single connection from the pool for advisory lock session affinity
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire dedicated connection for migration: %w", err)
	}
	defer conn.Close()

	// Acquire advisory lock
	if _, err := conn.ExecContext(ctx, "SELECT pg_advisory_lock($1)", MigrationAdvisoryLockID); err != nil {
		return fmt.Errorf("failed to acquire migration advisory lock: %w", err)
	}
	defer func() {
		// Release advisory lock on exit
		unlockCtx, unlockCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer unlockCancel()
		_, _ = conn.ExecContext(unlockCtx, "SELECT pg_advisory_unlock($1)", MigrationAdvisoryLockID)
	}()

	// 1. Create schema_migrations tracker table if not exists
	trackerSQL := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);
	`
	if _, err := conn.ExecContext(ctx, trackerSQL); err != nil {
		return fmt.Errorf("failed to ensure schema_migrations table: %w", err)
	}

	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		return nil // No migrations directory found, skip
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir %s: %w", migrationsDir, err)
	}

	var sqlFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			sqlFiles = append(sqlFiles, entry.Name())
		}
	}
	sort.Strings(sqlFiles)

	for _, filename := range sqlFiles {
		var exists bool
		err := conn.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", filename).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", filename, err)
		}

		if exists {
			continue // Already applied
		}

		content, err := os.ReadFile(filepath.Join(migrationsDir, filename))
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", filename, err)
		}

		tx, err := conn.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", filename, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", filename); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record applied migration %s: %w", filename, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", filename, err)
		}
	}

	return nil
}
