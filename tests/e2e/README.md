# EOMP Test Suites & Taxonomy

This workspace organizes tests into three distinct verification tiers:

### 1. In-Memory Simulation Suite (`tests/e2e`)
Unit and in-memory business-flow simulation tests. They exercise core domain flows, state machines, middleware, distributed rate limiting, event bus models, and concurrency primitives using mocked backends. Fast regression feedback during local development.

### 2. Database Integration & HTTP Contract Suites (`tests/integration` & `services/*/internal/repository`)
- **PostgreSQL Integration Suite:** Database-dependent tests for atomic refresh token rotation, token replay protection, user lifecycle & rollback, row-level authorization across 4 roles, ticket sequence allocation concurrency, optimistic locking, HMAC chain integrity, migration concurrency and parameterized queries. They skip by default when DSNs are absent. Gate evidence must run with `INTEGRATION_REQUIRED=1` (or the strict verification-script option) so a missing database fails closed.
- **HTTP Contract Harness:** `tests/integration/http_contract_test.go` uses `httptest`, test-only handlers and an in-memory ticket map to verify lifecycle/status and security contracts. It does not start the EOMP gateway or services and is not deployed API E2E evidence.

### 3. Frontend Unit & Contract Suite (`apps/web/tests`)
Vitest tests verify authentication state management, cookie BFF security, exact-origin CSRF guards, 401 refresh mutex concurrency, global route-policy logic and optimistic-lock state modeling. Vitest does not launch a browser; Playwright coverage remains required for browser E2E acceptance.

Current release and pilot acceptance status is maintained only in `docs/IMPLEMENTATION_STATUS.md`.
