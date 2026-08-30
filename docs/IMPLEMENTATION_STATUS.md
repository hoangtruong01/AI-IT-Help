# EOMP implementation status

Last audited: 2026-08-30

Code version: `0.1.0`

Readiness decision: **not yet approved for production**

This is the current evidence baseline. “Implemented” means the code exists and the listed local checks pass; it does not replace deployment, security-owner or disaster-recovery acceptance.

## Verified inventory

- 11 Go services plus one Nuxt web application.
- Service ports: gateway 8080, auth 8081, employee 8082, asset 8083, helpdesk 8084, workflow 8085, notification 8086, knowledge 8087, AI 8088, audit 8089 and reporting 8090.
- Nine PostgreSQL databases: auth, employee, asset, helpdesk, workflow, notification, knowledge, audit and reporting.
- 101 registered `/api/v1` operations. All 101 are present in the OpenAPI operation inventory.
- `go run scripts/check_openapi_coverage.go` validates route compatibility and exact runtime/OpenAPI parity. Jenkins now runs this gate.

## Remediation completed on 2026-08-30

- Registration always creates `ROLE_EMPLOYEE`; caller-selected privileged roles and fixed demo administrator seeds were removed. Optional initial-admin bootstrap is explicit.
- Email/password validation, JWT IDs and access/refresh token types were added. Refresh rotation verifies signature, issuer, expiry and subject.
- Gateway mutations enforce role-based authorization. Monitoring, reporting and audit endpoints are restricted.
- Frontend authentication is enabled; demo credentials/buttons were removed and production cookies use `Secure`.
- Production Compose/Kubernetes variables match service configuration. Plaintext production secrets were replaced by an external `eomp-secrets` contract.
- Docker builds use pinned Go/pnpm versions, the correct workspace context, conditional migrations and `.dockerignore`.
- RabbitMQ reconnects after startup failure and returns publish errors while disconnected. Redis rate-limit members are unique and trusted-proxy defaults are loopback-only.
- Database-backed services fail fast on DB/migration failure and expose real readiness checks.
- Production repositories and UI screens no longer replace DB/API failures with fabricated audit, reporting, monitoring, change, knowledge or AI data.
- Monitoring reports live service health probes and latency. RED/resource/log fields are explicitly unavailable until their real backends are configured; log API returns HTTP 501.
- Audit records use a chained HMAC-SHA256 proof with a minimum 32-character key. The repository serializes chain-head writes, exposes an integrity endpoint and migration `003` blocks ordinary UPDATE/DELETE operations.
- Asset, change, ticket, problem and workflow-approval critical transitions use optimistic compare-and-swap and return HTTP 409 for stale versions.
- AI/Qdrant mock fallback is disabled for real production providers unless `ALLOW_MOCK_AI=true` is explicitly set. Production manifests default to OpenAI and require an external key.
- The conflicting Helpdesk routes `/tickets/asset/{assetId}` and `/tickets/{id}/assign` were replaced with an unambiguous asset-ticket route.
- Backup validation no longer creates simulated evidence; release DR acceptance requires an actual temporary-database restore.
- CI security gates are fail-closed for gosec, govulncheck and Trivy, and include frontend tests/typecheck plus Go race tests.

## Verification evidence

Completed locally on 2026-08-30 after the remediation above:

- Go tests passed for all 13 workspace modules, including `tests/e2e`.
- OpenAPI YAML parsed successfully and Redocly reported a valid OpenAPI document.
- Runtime/OpenAPI gate passed: `101/101` operations documented, with no `http.ServeMux` conflicts.
- Frontend Vitest (3 tests), Nuxt typecheck, ESLint and production build passed.

Environment-limited checks:

- Docker/Kubernetes runtime integration was not executed because the local Docker daemon is unavailable.
- Helm rendering was not executed because Helm is not installed.
- No real PostgreSQL restore drill was available; RPO/RTO remain unverified until `scripts/backup_restore.ps1 test-restore` produces real evidence.
- No production AI provider key or external Qdrant instance was available for an end-to-end AI validation.

## Remaining release blockers

| Priority | Blocker | Acceptance condition |
|---|---|---|
| P0 | Previously committed secrets may already be compromised | Rotate PostgreSQL, JWT, RabbitMQ, MinIO, Grafana and any provider credentials in every environment and secret store |
| P0 | No real deployment/runtime proof | Build and scan every image, render Helm, deploy to a test cluster and pass smoke/integration checks |
| P0 | DR targets are unverified | Complete a measured nine-database backup/restore drill and retain objective evidence |
| P1 | Access/refresh tokens remain readable by browser JavaScript | Introduce a BFF/session design with an HttpOnly refresh cookie and CSRF protection |
| P1 | Audit immutability is not yet enforced by a dedicated restricted DB role | Run the audit service with append-only grants, document HMAC key rotation and validate tamper detection against PostgreSQL |
| P1 | Monitoring has probes but no real RED/log backend | Integrate Prometheus queries and a Loki-compatible log backend; keep unavailable fields explicit until then |
| P1 | OpenAPI has complete operation parity but many generic bodies/responses | Replace generic schemas with domain models and add request/response conformance tests |
| P1 | Optimistic locking is not universal | Add CAS and DB integration tests to every remaining mutable aggregate/update path |
| P1 | No browser E2E suite or measured load run | Add Playwright journeys and run k6 against a deployed environment |
| P1 | Production AI has not been exercised with a real provider | Inject a real key/provider, validate Qdrant ingestion and prove failure behavior without fabricated fallback |

## Release rule

Do not describe EOMP as “100% complete”, “fully certified” or “production ready” until every P0 condition has objective evidence and accountable owners sign the handover acceptance.
