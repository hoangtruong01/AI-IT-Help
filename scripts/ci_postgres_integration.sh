#!/usr/bin/env bash
# Fail-closed PostgreSQL integration runner for CI. It creates isolated databases,
# applies every service migration from scratch, runs all Gate D repository suites,
# archives evidence, and always removes the ephemeral PostgreSQL container.

set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

command -v docker >/dev/null 2>&1 || { echo "docker is required" >&2; exit 1; }
command -v go >/dev/null 2>&1 || { echo "go is required" >&2; exit 1; }

RUN_KEY="${BUILD_TAG:-local-$$}"
RUN_KEY="$(printf '%s' "$RUN_KEY" | tr -cd '[:alnum:]_.-' | cut -c1-40)"
CONTAINER="eomp-ci-postgres-${RUN_KEY:-$$}"
POSTGRES_USER="eomp_ci"
POSTGRES_PASSWORD="eomp-ci-${RANDOM}-${RANDOM}"
EVIDENCE_DIR="$PROJECT_ROOT/docs/evidence/gate-d"
LOG_FILE="$EVIDENCE_DIR/ci_postgres_integration.log"
RESULT_FILE="$EVIDENCE_DIR/ci_postgres_integration.json"

mkdir -p "$EVIDENCE_DIR"
: > "$LOG_FILE"

cleanup() {
    docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT INT TERM

docker run -d --rm \
    --name "$CONTAINER" \
    -e POSTGRES_USER="$POSTGRES_USER" \
    -e POSTGRES_PASSWORD="$POSTGRES_PASSWORD" \
    -e POSTGRES_DB=postgres \
    -p 127.0.0.1::5432 \
    postgres:17-alpine >/dev/null

for _ in $(seq 1 60); do
    if docker exec "$CONTAINER" pg_isready -U "$POSTGRES_USER" -d postgres >/dev/null 2>&1; then
        break
    fi
    sleep 1
done
docker exec "$CONTAINER" pg_isready -U "$POSTGRES_USER" -d postgres >/dev/null

HOST_PORT="$(docker port "$CONTAINER" 5432/tcp | head -n1 | sed 's/.*://')"
test -n "$HOST_PORT" || { echo "failed to resolve ephemeral PostgreSQL port" >&2; exit 1; }

DATABASES="auth_ci helpdesk_ci audit_ci migration_ci notification_ci reporting_ci"
for database in $DATABASES; do
    docker exec "$CONTAINER" createdb -U "$POSTGRES_USER" "$database"
done

apply_migrations() {
    local database="$1"
    local directory="$2"
    local file
    while IFS= read -r file; do
        docker exec -i "$CONTAINER" psql -v ON_ERROR_STOP=1 -U "$POSTGRES_USER" -d "$database" < "$file"
    done < <(find "$directory" -maxdepth 1 -type f -name '*.sql' | sort)
}

apply_migrations auth_ci services/auth/migrations
apply_migrations helpdesk_ci services/helpdesk/migrations
apply_migrations audit_ci services/audit/migrations
apply_migrations notification_ci services/notification/migrations
apply_migrations reporting_ci services/reporting/migrations

BASE_DSN="postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@127.0.0.1:${HOST_PORT}"
export INTEGRATION_REQUIRED=1
export AUTH_INTEGRATION_DSN="${BASE_DSN}/auth_ci?sslmode=disable"
export HELPDESK_INTEGRATION_DSN="${BASE_DSN}/helpdesk_ci?sslmode=disable"
export AUDIT_INTEGRATION_DSN="${BASE_DSN}/audit_ci?sslmode=disable"
export INTEGRATION_POSTGRES_DSN="${BASE_DSN}/migration_ci?sslmode=disable"
export NOTIFICATION_INTEGRATION_DSN="${BASE_DSN}/notification_ci?sslmode=disable"
export REPORTING_INTEGRATION_DSN="${BASE_DSN}/reporting_ci?sslmode=disable"

{
    (cd services/auth && go test -count=1 -v ./internal/repository)
    (cd services/helpdesk && go test -count=1 -v ./internal/repository)
    (cd services/audit && go test -count=1 -v ./internal/repository)
    (cd services/notification && go test -count=1 -v ./internal/repository)
    (cd services/reporting && go test -count=1 -v ./internal/repository)
    (cd tests/integration && go test -count=1 -v ./...)
} 2>&1 | tee "$LOG_FILE"

SOURCE_REVISION="$(git rev-parse HEAD)"
POSTGRES_VERSION="$(docker exec "$CONTAINER" psql -U "$POSTGRES_USER" -d postgres -Atc "SHOW server_version")"
COMPLETED_AT="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
BUILD_URL_VALUE="${BUILD_URL:-local}"
printf '{\n  "schema_version": 1,\n  "result": "PASS",\n  "source_revision": "%s",\n  "completed_at_utc": "%s",\n  "postgres_version": "%s",\n  "database_count": 6,\n  "integration_required": true,\n  "ci_build_url": "%s"\n}\n' \
    "$SOURCE_REVISION" "$COMPLETED_AT" "$POSTGRES_VERSION" "$BUILD_URL_VALUE" > "$RESULT_FILE"

echo "PostgreSQL integration PASS: six isolated databases, no skips allowed"
