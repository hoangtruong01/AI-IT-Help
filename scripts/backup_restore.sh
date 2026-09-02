#!/usr/bin/env bash
# ==============================================================================
# EOMP — Database Backup & Disaster Recovery CLI (Linux / macOS / CI)
# ==============================================================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BACKUP_DIR="$PROJECT_ROOT/backups"
mkdir -p "$BACKUP_DIR"

DATABASES=("auth_db" "employee_db" "asset_db" "helpdesk_db" "workflow_db" "notification_db" "knowledge_db" "reporting_db" "audit_db")

show_help() {
    echo "================================================================"
    echo "  EOMP - Phase 14 Database Backup & Disaster Recovery CLI"
    echo "================================================================"
    echo ""
    echo "Usage: ./scripts/backup_restore.sh [command]"
    echo ""
    echo "Commands:"
    echo "  backup        Perform full backup of all 9 PostgreSQL databases"
    echo "  list          List all archived backup snapshots"
    echo "  test-restore  Validate integrity of latest backup snapshot"
    echo ""
}

backup_dbs() {
    TIMESTAMP=$(date +%Y%m%d_%H%M%S)
    echo "=== Starting Full Backup of 9 EOMP Databases [$TIMESTAMP] ==="
    CONTAINER="eomp-postgres"
    if ! docker inspect "$CONTAINER" >/dev/null 2>&1; then
        CONTAINER="eomp-prod-postgres"
    fi
    if ! docker inspect "$CONTAINER" >/dev/null 2>&1; then
        echo "Error: PostgreSQL container is not running; backup was not created." >&2
        exit 1
    fi

    for db in "${DATABASES[@]}"; do
        OUT_FILE="$BACKUP_DIR/${db}_${TIMESTAMP}.sql"
        echo "  [BACKUP] Dumping database $db..."
        if ! docker exec "$CONTAINER" pg_dump -U eomp -d "$db" -F p > "$OUT_FILE"; then
            rm -f "$OUT_FILE"
            echo "Error: pg_dump failed for $db" >&2
            exit 1
        fi
        if [ ! -s "$OUT_FILE" ] || [ $(wc -c < "$OUT_FILE") -lt 100 ]; then
            rm -f "$OUT_FILE"
            echo "Error: pg_dump failed or produced an invalid artifact for $db" >&2
            exit 1
        fi
        echo "    -> Saved $OUT_FILE ($(wc -c < "$OUT_FILE") bytes)"
    done
    echo "All 9 databases backed up successfully to $BACKUP_DIR"
}

list_backups() {
    echo "=== Existing Database Backup Archives ==="
    ls -lh "$BACKUP_DIR"
}

test_restore() {
    echo "=== Running Integrity & Disaster Recovery Verification Test ==="
    CONTAINER="eomp-postgres"
    if ! docker inspect "$CONTAINER" >/dev/null 2>&1; then
        CONTAINER="eomp-prod-postgres"
    fi
    if ! docker inspect "$CONTAINER" >/dev/null 2>&1; then
        echo "Error: PostgreSQL container is not running; restore drill cannot run." >&2
        exit 1
    fi

    START_TIME=$(date +%s)
    VERIFIED_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)
    INDEX=0
    CURRENT_RESTORE_DB=""
    cleanup_restore_db() {
        if [ -n "$CURRENT_RESTORE_DB" ]; then
            docker exec "$CONTAINER" dropdb -U eomp --if-exists "$CURRENT_RESTORE_DB" >/dev/null 2>&1 || true
        fi
    }
    trap cleanup_restore_db EXIT
    for db in "${DATABASES[@]}"; do
        LATEST_FILE=$(ls -t "$BACKUP_DIR/${db}"_*.sql 2>/dev/null | head -n 1)
        if [ -z "$LATEST_FILE" ]; then
            echo "Error: No backup file found for $db" >&2
            exit 1
        fi
        INDEX=$((INDEX + 1))
        RESTORE_DB="eomp_restore_verify_${START_TIME}_${INDEX}"
        CURRENT_RESTORE_DB="$RESTORE_DB"
        docker exec "$CONTAINER" createdb -U eomp "$RESTORE_DB"
        docker exec -i "$CONTAINER" psql -v ON_ERROR_STOP=1 -U eomp -d "$RESTORE_DB" < "$LATEST_FILE" >/dev/null
        TABLE_COUNT=$(docker exec "$CONTAINER" psql -U eomp -d "$RESTORE_DB" -Atc "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public';")
        docker exec "$CONTAINER" dropdb -U eomp --if-exists "$RESTORE_DB" >/dev/null 2>&1 || true
        CURRENT_RESTORE_DB=""

        if [ "$TABLE_COUNT" -le 0 ]; then
            echo "Error: Restored database has no public tables for $db" >&2
            exit 1
        fi
    done

    END_TIME=$(date +%s)
    DURATION=$((END_TIME - START_TIME))
    trap - EXIT
    printf '{\n  "status": "passed",\n  "verified_at": "%s",\n  "database_count": %d,\n  "restore_duration_seconds": %d,\n  "restore_scope": "nine PostgreSQL databases only; not full-service RTO"\n}\n' \
        "$VERIFIED_AT" "${#DATABASES[@]}" "$DURATION" > "$BACKUP_DIR/dr_evidence.json"
    echo "[VERIFIED] Real database restore drill completed for ${#DATABASES[@]} databases. DatabaseRestoreDuration=${DURATION}s"
    echo "This measures database restore duration; it does not by itself certify WAL RPO or full-service RTO."
}

case "$1" in
    backup)       backup_dbs ;;
    list)         list_backups ;;
    test-restore) test_restore ;;
    *)            show_help ;;
esac
