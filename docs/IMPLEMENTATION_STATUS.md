# EOMP implementation status

Last audited: 2026-08-30
Code version: 0.1.0
Readiness decision: **not yet approved for production**

This file is the current status baseline. “Implemented” means code exists and its local automated checks pass; it does not mean the feature has completed production acceptance.

## Verified inventory

- 11 Go services plus Nuxt web application.
- Service ports: gateway 8080, auth 8081, employee 8082, asset 8083, helpdesk 8084, workflow 8085, notification 8086, knowledge 8087, AI 8088, audit 8089, reporting 8090.
- Nine PostgreSQL databases: auth, employee, asset, helpdesk, workflow, notification, knowledge, audit and reporting.
- 96 runtime `/api/v1` operations were found in the Go route registrations during the 2026-08-30 audit.
- The checked-in OpenAPI document describes only 17 operations and is therefore not yet a complete API contract.

## Remediation completed on 2026-08-30

- Public registration always creates `ROLE_EMPLOYEE`; caller-selected privileged roles were removed.
- Registration validates email and strong passwords. JWTs now have unique IDs and explicit access/refresh token types; refresh rotation verifies JWT signature, issuer, expiry and subject.
- Fixed demo administrator seeds were removed. Existing historical fixed-ID demo accounts are disabled by migration. Optional initial admin bootstrap requires explicit environment variables.
- Gateway write routes now enforce role-based authorization; monitoring and reporting endpoints are restricted.
- Frontend route authentication was enabled, login defaults/demo buttons were removed and production cookies use `Secure`.
- Production Compose/Kubernetes variable names now match the service configuration contract. PostgreSQL, Redis, RabbitMQ and Nuxt API variables were corrected.
- Plaintext Kubernetes/Helm production secrets were removed. Deployments now expect an externally managed `eomp-secrets` Secret.
- Docker builds use pinned Go/pnpm versions, a valid workspace build context, conditional migrations and `.dockerignore`.
- RabbitMQ configuration URL-escapes credentials, reconnects after boot failure and returns publish errors while disconnected instead of reporting false success.
- Redis sliding-window members are unique per request; trusted-proxy defaults are loopback-only.
- Database-backed services fail fast when DB/migrations fail and expose `/ready` with a real DB ping.
- Audit/reporting repositories no longer return fabricated production data after DB errors. Test fixtures now live only in test code.
- Asset assign/return/status and Change status transitions use optimistic compare-and-swap and return HTTP 409 on stale versions.
- Backup restore validation no longer creates simulated artifacts; DR acceptance requires evidence from a real temporary-database restore.
- Frontend now has a real Vitest command. QA distinguishes PASS, FAIL and SKIP.

## Verification evidence

Completed locally on 2026-08-30:

- Go tests passed for all 13 workspace modules, including `tests/e2e`.
- Frontend Vitest: 1 file, 3 tests passed.
- Frontend Nuxt typecheck passed.
- Frontend ESLint passed after remediation.
- Production Nuxt build passed during the audit; it must be rerun after every subsequent UI change.

Environment-limited checks:

- Docker/Kubernetes runtime integration was not executed because the local Docker daemon was unavailable.
- Helm rendering was not executed because Helm was not installed.
- No real PostgreSQL restore drill was available; RPO/RTO are unverified until `scripts/backup_restore.ps1 test-restore` produces `backups/dr_evidence.json`.
- Go coverage could not be collected because this local Go installation lacks the `covdata` tool. No coverage percentage is claimed.

## Remaining release blockers

| Priority | Blocker | Acceptance condition |
|---|---|---|
| P0 | Previously committed secrets may already be compromised | Rotate PostgreSQL, JWT, RabbitMQ, MinIO and Grafana credentials in every environment and secret store |
| P0 | OpenAPI covers only a fraction of runtime endpoints | Document and validate all 96 operations in CI |
| P0 | No real deployment/runtime proof | Build all images, render Helm, apply manifests in a test cluster and run smoke/integration tests |
| P0 | DR targets are unverified | Complete a measured nine-database backup/restore drill and retain evidence |
| P1 | Access/refresh tokens remain readable by browser JavaScript | Introduce a BFF/session design with HttpOnly refresh cookie and CSRF protection |
| P1 | Audit checksum is SHA-256, not an HMAC hash chain or database-enforced immutability | Add keyed chained hashes, verification endpoint and restricted append-only DB role/permissions |
| P1 | Monitoring handler still exposes synthetic/static values | Replace it with Prometheus/service probes or label and isolate it as demo data |
| P1 | Optimistic locking is not yet consistent across every mutable aggregate | Add CAS to tickets, problems, workflow instances and remaining update paths with DB integration tests |
| P1 | No browser E2E suite or measured load run | Add Playwright for critical user journeys and run k6 against a deployed environment |
| P1 | Static Kubernetes defaults explicitly acknowledge the mock AI provider | Set a real provider and securely inject its API key before production acceptance |

## Release rule

Do not describe the platform as “100% complete”, “fully certified” or “production ready” until every P0 item has objective evidence and the acceptance document is signed by the accountable owners.
