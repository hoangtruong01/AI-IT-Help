#!/usr/bin/env bash
# ==============================================================================
# EOMP local release precheck (Linux / CI)
# ==============================================================================

set -e

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo ""
echo "================================================================="
echo "          EOMP LOCAL RELEASE PRECHECK (NON-ACCEPTANCE)           "
echo "================================================================="
echo ""

if [ "${REQUIRE_POSTGRES:-0}" = "1" ]; then
    MISSING_DSN=""
    for name in AUTH_INTEGRATION_DSN HELPDESK_INTEGRATION_DSN AUDIT_INTEGRATION_DSN INTEGRATION_POSTGRES_DSN; do
        if [ -z "${!name:-}" ]; then MISSING_DSN="$MISSING_DSN $name"; fi
    done
    if [ -n "$MISSING_DSN" ]; then
        echo "[-] PostgreSQL verification requested, but required DSN(s) are missing:$MISSING_DSN" >&2
        exit 1
    fi
fi

# 1. Go formatting & vet
echo "[1/6] Verifying Go Code Quality, Formatting & Vet..."
UNFORMATTED=$(gofmt -l "$PROJECT_ROOT/packages" "$PROJECT_ROOT/services" "$PROJECT_ROOT/tests" "$PROJECT_ROOT/scripts" 2>/dev/null || true)
if [ -n "$UNFORMATTED" ]; then
    echo "  [-] Unformatted Go files detected:"
    echo "$UNFORMATTED"
    exit 1
fi
echo "  [+] Go Formatting: PASS"

for mod in packages/shared services/* tests/e2e tests/integration; do
    if [ -f "$PROJECT_ROOT/$mod/go.mod" ]; then
        (cd "$PROJECT_ROOT/$mod" && go vet ./...)
    fi
done
echo "  [+] Go Vet: PASS"

# 2. OpenAPI Parity
echo "[2/6] Verifying OpenAPI Contract Coverage..."
(cd "$PROJECT_ROOT" && go run scripts/check_openapi_coverage.go)
echo "  [+] OpenAPI Parity: PASS"

# 3. Unit & Simulation
echo "[3/6] Running Backend Unit & Simulation Suites (13 modules)..."
(cd "$PROJECT_ROOT" && go test -skip '(Integration|Postgres)' ./packages/shared/... ./services/... ./tests/e2e/...)
echo "  [+] Unit & Simulation Tests: PASS"

# 4. HTTP contracts and optional fail-closed PostgreSQL integration
if [ "${REQUIRE_POSTGRES:-0}" = "1" ]; then
    echo "[4/6] Running fail-closed PostgreSQL integration suites and HTTP contracts..."
    MISSING_DSN=""
    for name in AUTH_INTEGRATION_DSN HELPDESK_INTEGRATION_DSN AUDIT_INTEGRATION_DSN INTEGRATION_POSTGRES_DSN; do
        if [ -z "${!name:-}" ]; then MISSING_DSN="$MISSING_DSN $name"; fi
    done
    if [ -n "$MISSING_DSN" ]; then
        echo "  [-] Missing required DSN(s):$MISSING_DSN" >&2
        exit 1
    fi
    export INTEGRATION_REQUIRED=1
    for mod in services/auth services/helpdesk services/audit tests/integration; do
        (cd "$PROJECT_ROOT/$mod" && go test -count=1 -v ./...)
    done
    echo "  [+] PostgreSQL integration: PASS (required; no skips allowed)"
else
    echo "[4/6] Running in-process HTTP contract harness..."
    (cd "$PROJECT_ROOT/tests/integration" && go test -count=1 -v -run '^TestHTTPContract_' ./...)
    echo "  [+] HTTP contract harness: PASS (test doubles; not deployed E2E)"
fi

# 5. Frontend Suite
echo "[5/6] Verifying Frontend Application (apps/web)..."
(cd "$PROJECT_ROOT/apps/web" && pnpm test && pnpm typecheck && pnpm lint && pnpm build)
echo "  [+] Frontend unit/contracts, typecheck & production build: PASS (no browser launched)"

# 6. Container Hardening
echo "[6/6] Running static Container/Kubernetes hardening prechecks..."
if grep -q "FROM .*latest" "$PROJECT_ROOT/deploy/docker/Dockerfile.go-service" || grep -q "FROM .*latest" "$PROJECT_ROOT/deploy/docker/Dockerfile.web"; then
    echo "  [-] Unpinned ':latest' base image found!"
    exit 1
fi
grep -q "USER 10001:10001" "$PROJECT_ROOT/deploy/docker/Dockerfile.go-service"
grep -q "USER 10001:10001" "$PROJECT_ROOT/deploy/docker/Dockerfile.web"
test -f "$PROJECT_ROOT/deploy/kubernetes/manifests/09-network-policies.yaml"
test -f "$PROJECT_ROOT/deploy/kubernetes/manifests/10-pod-disruption-budgets.yaml"
echo "  [+] Static hardening precheck: PASS (not a Trivy/CIS runtime scan)"

echo ""
echo "================================================================="
echo "  OVERALL STATUS: LOCAL PRECHECK PASSED"
echo "  GATE D: NOT APPROVED BY THIS SCRIPT"
echo "  Pending: deployed API E2E, Playwright, Trivy/image build,"
echo "           Helm staging smoke, k6 results, and DR evidence."
echo "================================================================="
echo ""
