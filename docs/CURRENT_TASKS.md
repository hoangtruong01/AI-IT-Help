# EOMP — Current Active & Unfinished Tasks Tracker

> **Document Type:** Active Engineering Tasks & Release Backlog  
> **Status:** Live Open Tasks  
> **Rule:** This file contains **ONLY active, open, in-progress, or remaining tasks**. Completed tasks are removed upon verification and recorded in [`docs/PROJECT_DOCUMENTATION.md`](file:///d:/IT_help/eomp/docs/PROJECT_DOCUMENTATION.md).

---

## 📊 Summary of Open Tasks

| Task ID | Task Name | Priority | Status | Owner / Role |
|---|---|:---:|:---:|---|
| [TASK-REL-001](#task-rel-001--staging-tls--private-observability-verification) | Staging TLS & Private Observability Verification | CRITICAL | IN PROGRESS | DevOps / SRE + Security |
| [TASK-REL-002](#task-rel-002--ci-automated-ephemeral-postgresql-integration-suite) | CI Automated Ephemeral PostgreSQL Integration Suite | HIGH | IN PROGRESS | Backend + DevOps |
| [TASK-REL-003](#task-rel-003--playwright-browser-e2e-suite--staging-verification) | Playwright Browser E2E Suite & Staging Verification | HIGH | NOT STARTED | QA / Automation + Frontend |
| [TASK-REL-004](#task-rel-004--container-security-cve-scan--full-service-dr-targets) | Container Security CVE Scan & Full-Service DR Targets | HIGH | IN PROGRESS | SRE + Security |
| [TASK-REL-005](#task-rel-005--product-owner--security-sign-off-for-pilot-handover) | Product Owner & Security Sign-Off for Pilot Handover | CRITICAL | BLOCKED | Product Owner + Security Lead |
| [TASK-BKL-001](#task-bkl-001--sla-background-escalation-engine--leader-election) | SLA Background Escalation Engine & Leader Election | MEDIUM | NOT STARTED | Backend + Architect |
| [TASK-BKL-002](#task-bkl-002--business-calendar--vietnam-holiday-working-hours) | Business Calendar & Vietnam Holiday Working Hours | MEDIUM | NOT STARTED | BA + Backend |
| [TASK-BKL-003](#task-bkl-003--transactional-outbox-pattern-for-rabbitmq-publishing) | Transactional Outbox Pattern for RabbitMQ Publishing | MEDIUM | NOT STARTED | Backend + Database |
| [TASK-BKL-004](#task-bkl-004--secure-ticket-attachments-via-minio-presigned-urls) | Secure Ticket Attachments via MinIO Presigned URLs | MEDIUM | NOT STARTED | Frontend + Backend |
| [TASK-BKL-005](#task-bkl-005--account-lockout--risk-based-rate-limiting) | Account Lockout & Risk-Based Rate Limiting | MEDIUM | NOT STARTED | Security + Backend |
| [TASK-BKL-006](#task-bkl-006--vector-database-rag-role-based-access-scoping) | Vector Database RAG Role-Based Access Scoping | MEDIUM | NOT STARTED | AI Engineer + Security |
| [TASK-BKL-007](#task-bkl-007--internationalization-i18n-support-for-web--notifications) | Internationalization (i18n) Support for Web & Notifications | LOW | NOT STARTED | Frontend + BA |
| [TASK-BKL-008](#task-bkl-008--real-loki-log--prometheus-red-backend-integration) | Real Loki Log & Prometheus RED Backend Integration | MEDIUM | NOT STARTED | Backend + DevOps |

---

## 🚀 Release & Pilot Verification Tasks

---

### TASK-REL-001 — Staging TLS & Private Observability Verification

**Status**
```text
IN PROGRESS
```

**Priority**
```text
CRITICAL
```

**Owner/Role**
DevOps / SRE + Security Engineer

**Problem**
Gate C-02 requires formal staging verification showing that HTTPS is enforced with a valid organization-issued TLS certificate, HTTP port 80 strictly redirects to 443 with HSTS headers, and observability services (Prometheus `:9090`, Grafana `:3002`) are isolated from public ingress.

**Expected Result**
- Staging environment deployed with valid TLS certificates.
- `curl -I http://<staging-domain>` returns `301 Moved Permanently` to `https://`.
- `Strict-Transport-Security: max-age=31536000; includeSubDomains; preload` header present.
- Observability routes return `404` / connection refused from untrusted external networks.

**Current State**
Nginx configuration (`deploy/nginx/conf.d/eomp.conf`), Docker Compose (`deploy/docker-compose.prod.yml`), and Kubernetes Ingress templates (`deploy/kubernetes/manifests/08-ingress.yaml`) are implemented and syntax-verified.

**Missing**
- Real staging cluster deployment with organization TLS certificate secret.
- Execution of `nginx -t` on the staging host.
- External TLS policy scanner report (SSL Labs / testssl.sh).
- Proof of network isolation for Prometheus and Grafana.

**Affected Modules**
`deploy/nginx`, `deploy/kubernetes`, `deploy/docker-compose.prod.yml`

**Dependencies**
Access to staging server / Kubernetes cluster with valid DNS and certificate.

**Acceptance Criteria**
- [ ] Staging Helm / Compose deployment successfully rolled out.
- [ ] TLS certificate is valid and issued by a trusted CA.
- [ ] HTTP 80 redirects to 443 with HSTS.
- [ ] Untrusted network probe cannot access `/monitoring/prometheus` or `/monitoring/grafana`.
- [ ] Verification log saved to `docs/evidence/gate-c/staging_tls_evidence.json`.

**Implementation Checklist**
### DevOps
- [ ] Provision staging server with DNS and TLS secret.
- [ ] Execute `./scripts/deploy.sh prod-up` or Helm install on staging.
- [ ] Run TLS policy scanner against staging domain.

### Security
- [ ] Audit TLS ciphers, HSTS headers, and CSP rules.

### Documentation
- [ ] Update `docs/PROJECT_DOCUMENTATION.md` with verified staging evidence.

---

### TASK-REL-002 — CI Automated Ephemeral PostgreSQL Integration Suite

**Status**
```text
IN PROGRESS
```

**Priority**
```text
HIGH
```

**Owner/Role**
Backend Engineer + DevOps Engineer

**Problem**
Gate D-01 integration tests run successfully on local developer machines against temporary PostgreSQL databases, but the CI pipeline (`Jenkinsfile`) does not yet automatically provision ephemeral PostgreSQL databases during build stages, and dedicated integration test suites for Notification receipts and Reporting date-range queries are missing.

**Expected Result**
- CI pipeline automatically starts isolated PostgreSQL ephemeral services for integration testing.
- Test runner sets `INTEGRATION_REQUIRED=1` so missing databases trigger a hard failure rather than a skip.
- Integration test coverage expanded to Notification and Reporting services.
- Versioned CI test run logs linked to Git commit SHA.

**Current State**
Core PostgreSQL integration suites exist in `services/auth/internal/repository/auth_integration_test.go`, `services/helpdesk/internal/repository/helpdesk_integration_test.go`, `services/audit/internal/repository/audit_integration_test.go`, and `tests/integration/postgres_integration_test.go`.

**Missing**
- Ephemeral PostgreSQL service block in `Jenkinsfile`.
- Integration test for Notification receipts (`services/notification/internal/repository/`).
- Integration test for Reporting date filters and aggregations (`services/reporting/internal/repository/`).
- CI run URL logging.

**Affected Modules**
`Jenkinsfile`, `services/notification`, `services/reporting`, `tests/integration`

**Dependencies**
Docker / PostgreSQL service daemon available in CI runner.

**Acceptance Criteria**
- [ ] Jenkinsfile runs `go test -v ./tests/integration/...` with `INTEGRATION_REQUIRED=1`.
- [ ] Notification receipt integration test validates read state per recipient.
- [ ] Reporting integration test validates date filtering and aggregate rollup without skipping.
- [ ] Zero skipped integration tests in CI report.

**Implementation Checklist**
### Backend
- [ ] Write `services/notification/internal/repository/notification_integration_test.go`.
- [ ] Write `services/reporting/internal/repository/reporting_integration_test.go`.

### DevOps
- [ ] Update `Jenkinsfile` stage `Integration Test` with ephemeral PostgreSQL containers.

### Testing
- [ ] Execute `REQUIRE_POSTGRES=1 ./scripts/staging_verify.sh` to confirm zero skips.

### Documentation
- [ ] Record CI integration evidence in `docs/PROJECT_DOCUMENTATION.md`.

---

### TASK-REL-003 — Playwright Browser E2E Suite & Staging Verification

**Status**
```text
NOT STARTED
```

**Priority**
```text
HIGH
```

**Owner/Role**
QA / Automation Engineer + Frontend Engineer

**Problem**
Gate D-02 requires end-to-end verification executed through a real headless browser (Playwright) to validate the actual client-side UX, Cookie BFF handling, token refresh mutex behavior, and error modals under real network conditions.

**Expected Result**
- Playwright test suite in `tests/e2e/playwright/` executing complete user journeys.
- Journeys cover:
  1. Login → Dashboard render with role-specific navigation.
  2. Route guard redirect when unauthorized role accesses `/audit` or `/reports`.
  3. Create Ticket → Verify sequential `TK-*` number and SLA countdown.
  4. Concurrent edit triggering `409 Conflict` error dialog.
  5. Silent access token expiration and automatic refresh rotation.
  6. Logout → Immediate session revocation and redirect to `/login`.
- Playwright HTML reports, test videos, and traces archived on test completion.

**Current State**
Backend deployed-stack E2E passed via API calls (`docs/evidence/gate-d/deployed_stack_e2e.json`). Frontend Vitest unit/contract tests pass (`apps/web/tests/route_permissions.test.ts`).

**Missing**
- Playwright configuration and dependency installation (`@playwright/test`).
- Browser journey test scripts in `tests/e2e/playwright/`.
- CI execution stage for Playwright.

**Affected Modules**
`apps/web`, `tests/e2e`, `Jenkinsfile`

**Dependencies**
Running Web frontend (`http://localhost:3000`) and Gateway (`http://localhost:8080`).

**Acceptance Criteria**
- [ ] Playwright suite passes 100% on 6 core user journeys.
- [ ] Browser traces and screenshots generated for all flows.
- [ ] Staging run completed against deployed staging environment.

**Implementation Checklist**
### QA / Testing
- [ ] Initialize Playwright in `tests/e2e/`.
- [ ] Implement `login_and_navigation.spec.ts`.
- [ ] Implement `ticket_lifecycle.spec.ts`.
- [ ] Implement `concurrency_conflict.spec.ts`.
- [ ] Implement `token_refresh_mutex.spec.ts`.

### Documentation
- [ ] Document Playwright execution commands in `docs/PROJECT_DOCUMENTATION.md`.

---

### TASK-REL-004 — Container Security CVE Scan & Full-Service DR Targets

**Status**
```text
IN PROGRESS
```

**Priority**
```text
HIGH
```

**Owner/Role**
SRE / Platform Engineer + Security Engineer

**Problem**
Gate D-03 requires verifiable automated container vulnerability scanning (Trivy / Docker Scout) proving zero High/Critical CVEs across all 12 application images, and a formal Disaster Recovery proof demonstrating WAL streaming recovery (RPO < 5 min) and full-service restart-to-ready time (RTO < 15 min).

**Expected Result**
- Trivy vulnerability scan JSON reports demonstrating zero unresolved High/Critical CVEs.
- Automated DR drill measuring both PostgreSQL WAL restore and container service startup times.
- Verifiable evidence report stored in `docs/evidence/gate-d/`.

**Current State**
All 12 application images built as non-root (`USER 10001:10001`) with immutable digests recorded in `docs/evidence/gate-d/image_manifest.json`. Database restore drill passed for 9 databases in 6.602s (`docs/evidence/gate-d/dr_evidence_20260902.json`).

**Missing**
- Trivy CLI execution across all 12 images in CI.
- WAL archiving and point-in-time recovery (PITR) drill verification.
- Full-stack start-to-ready RTO measurement.

**Affected Modules**
`scripts/backup_restore.ps1/.sh`, `Jenkinsfile`, `deploy/docker/`

**Dependencies**
Trivy scanner installed on build host; PostgreSQL WAL archive volume.

**Acceptance Criteria**
- [ ] Trivy scan passes on all 12 images with 0 High/Critical findings.
- [ ] Disaster recovery drill achieves RPO < 5 minutes and full-service RTO < 15 minutes.
- [ ] Results saved in `docs/evidence/gate-d/trivy_scan_report.json` and `docs/evidence/gate-d/dr_full_service.json`.

**Implementation Checklist**
### SRE / DevOps
- [ ] Add Trivy vulnerability scanning step to `Jenkinsfile`.
- [ ] Enhance `scripts/backup_restore.ps1` to verify WAL log replay.
- [ ] Benchmark full cold-start to healthy status across all 20 containers.

### Documentation
- [ ] Record vulnerability and DR evidence in `docs/PROJECT_DOCUMENTATION.md`.

---

### TASK-REL-005 — Product Owner & Security Sign-Off for Pilot Handover

**Status**
```text
BLOCKED (Waiting on TASK-REL-001, TASK-REL-002, TASK-REL-003, TASK-REL-004)
```

**Priority**
```text
CRITICAL
```

**Owner/Role**
Product Owner + Security Lead

**Problem**
The controlled pilot cannot commence until Product Owner and Security Lead formally sign and date the handover certificate, acknowledging the verified authorization matrix, operational scope, and accepted environment limits.

**Expected Result**
- Formally signed and timestamped handover certificate.
- Traceable commit SHA recorded in the release register.

**Current State**
Authorization matrix finalized in `docs/PROJECT_DOCUMENTATION.md` (4 roles, 11 resources, 8 product decisions).

**Missing**
- Formal stakeholder review meeting and digital signature on the handover report.

**Affected Modules**
Governance documentation

**Dependencies**
Completion of Gate C and Gate D technical verification tasks.

**Acceptance Criteria**
- [ ] Product Owner approves authorization matrix and operational scope.
- [ ] Security Lead approves zero-trust boundary and HMAC audit trail.
- [ ] Handover record documented with commit SHA and approver names.

---

## 📦 Post-Pilot Roadmap & Feature Backlog

---

### TASK-BKL-001 — SLA Background Escalation Engine & Leader Election

**Status**
```text
NOT STARTED
```

**Priority**
```text
MEDIUM
```

**Owner/Role**
Backend Engineer + Software Architect

**Problem**
SLA status is currently evaluated on-demand when tickets are queried. If a ticket approaches its deadline without user interaction, no proactive alert is triggered. When multiple Helpdesk service replicas run, a distributed background scanner with leader election is needed.

**Expected Result**
- Background worker scans `helpdesk_db.tickets` every 60 seconds.
- Detects tickets with `< 20%` SLA remaining (Warning) or past deadline (Breached).
- Updates `sla_status` in PostgreSQL and publishes `sla.warning` / `sla.breached` CloudEvents.
- Uses PostgreSQL advisory lock or Redis lease for leader election across instances.

**Affected Modules**
`services/helpdesk`

**Implementation Checklist**
### Backend
- [ ] Implement `internal/worker/sla_scanner.go` in Helpdesk Service.
- [ ] Implement leader election using `pg_try_advisory_lock`.
- [ ] Publish SLA events to RabbitMQ.

---

### TASK-BKL-002 — Business Calendar & Vietnam Holiday Working Hours

**Status**
```text
NOT STARTED
```

**Priority**
```text
MEDIUM
```

**Owner/Role**
BA + Backend Engineer

**Problem**
SLA calculation currently uses a hardcoded 08:00–17:30 Mon–Fri formula without excluding official Vietnamese public holidays (Tet, National Day, Reunification Day, etc.).

**Expected Result**
- Database table `business_holidays` storing annual holiday dates.
- SLA engine computes deadlines excluding non-working holiday dates.
- Admin API to manage annual holiday calendars.

**Affected Modules**
`services/helpdesk`

**Implementation Checklist**
### BA
- [ ] Define holiday schedule schema and edge cases.

### Backend
- [ ] Add migration `004_create_business_holidays.sql` in `services/helpdesk/migrations/`.
- [ ] Update `sla_engine.go` to query holiday cache.

---

### TASK-BKL-003 — Transactional Outbox Pattern for RabbitMQ Publishing

**Status**
```text
NOT STARTED
```

**Priority**
```text
MEDIUM
```

**Owner/Role**
Backend Engineer + Database Engineer

**Problem**
Directly publishing RabbitMQ messages after a SQL commit can cause event loss if the network fails immediately after the commit (Dual-Write Hazard).

**Expected Result**
- Microservices write domain events to an `outbox_events` table within the primary SQL transaction.
- Dedicated background worker polls `outbox_events`, publishes to RabbitMQ, and marks them `PUBLISHED`.
- Guarantees At-Least-Once event delivery across all microservices.

**Affected Modules**
`packages/shared/pkg/eventbus`, `services/helpdesk`, `services/workflow`, `services/auth`

**Implementation Checklist**
### Backend
- [ ] Implement generic Transactional Outbox publisher in `packages/shared/pkg/eventbus/outbox.go`.
- [ ] Add outbox table migration to Helpdesk, Workflow, and Auth services.

---

### TASK-BKL-004 — Secure Ticket Attachments via MinIO Presigned URLs

**Status**
```text
NOT STARTED
```

**Priority**
```text
MEDIUM
```

**Owner/Role**
Frontend Engineer + Backend Engineer

**Problem**
Users and agents cannot upload screenshots, error logs, or diagnostic files to support tickets.

**Expected Result**
- Helpdesk service exposes `POST /api/v1/tickets/{id}/attachments/presign-upload`.
- Generates a short-lived MinIO presigned PUT URL validating MIME types (images, PDF, txt, zip) and max file size (10MB).
- Frontend component `<TicketAttachmentUploader />` allows drag-and-drop file upload directly to MinIO.
- On upload completion, metadata is saved in `helpdesk_db.ticket_attachments`.

**Affected Modules**
`services/helpdesk`, `apps/web`

**Implementation Checklist**
### Backend
- [ ] Integrate MinIO S3 SDK in `services/helpdesk/internal/service/attachment.go`.
- [ ] Implement presigned URL endpoint with strict extension/MIME validation.

### Frontend
- [ ] Build drag-and-drop upload component in `apps/web/app/components/helpdesk/`.

---

### TASK-BKL-005 — Account Lockout & Risk-Based Rate Limiting

**Status**
```text
NOT STARTED
```

**Priority**
```text
MEDIUM
```

**Owner/Role**
Security Engineer + Backend Engineer

**Problem**
IP-based rate limiting does not prevent distributed credential stuffing targeting specific user accounts across rotating proxy IPs.

**Expected Result**
- After 5 consecutive failed login attempts on a specific email within 15 minutes, account is locked for 30 minutes.
- Security event logged in `auth_db.login_audit_logs`.
- Admin endpoint `POST /api/v1/users/{id}/unlock` to unlock accounts manually.

**Affected Modules**
`services/auth`

**Implementation Checklist**
### Backend
- [ ] Add `failed_login_attempts` and `locked_until` columns to `users` table.
- [ ] Update `auth_service.go` login handler with lockout logic.

---

### TASK-BKL-006 — Vector Database RAG Role-Based Access Scoping

**Status**
```text
NOT STARTED
```

**Priority**
```text
MEDIUM
```

**Owner/Role**
AI Engineer + Security Engineer

**Problem**
Qdrant vector searches during AI Chat / Triage do not filter out internal SOP documents when called by end-users (`ROLE_EMPLOYEE`).

**Expected Result**
- Vector points in Qdrant store `is_internal: boolean` in payload metadata.
- When `ROLE_EMPLOYEE` queries the RAG engine, Qdrant search payload injects `filter: { must: [{ key: "is_internal", match: { value: false } }] }`.
- Internal articles and sensitive SOP runbooks are excluded from LLM context for unauthorized roles.

**Affected Modules**
`services/ai`, `services/knowledge`

**Implementation Checklist**
### AI
- [ ] Update `services/ai/internal/rag/retriever.go` to accept `Actor` context.
- [ ] Inject Qdrant payload filter based on user role.

---

### TASK-BKL-007 — Internationalization (i18n) Support for Web & Notifications

**Status**
```text
NOT STARTED
```

**Priority**
```text
LOW
```

**Owner/Role**
Frontend Engineer + Business Analyst

**Problem**
The web interface currently mixes English and Vietnamese terminology across different pages.

**Expected Result**
- `@nuxtjs/i18n` integrated with full English (`en`) and Vietnamese (`vi`) locale dictionaries.
- Header language toggle switch allowing instant language switching.
- Notification templates localized according to recipient's language preference.

**Affected Modules**
`apps/web`, `services/notification`

**Implementation Checklist**
### Frontend
- [ ] Install `@nuxtjs/i18n` in `apps/web`.
- [ ] Create `locales/en.json` and `locales/vi.json`.
- [ ] Add `<LanguageSwitcher />` to navigation header.

---

### TASK-BKL-008 — Real Loki Log & Prometheus RED Backend Integration

**Status**
```text
NOT STARTED
```

**Priority**
```text
MEDIUM
```

**Owner/Role**
Backend Engineer + DevOps Engineer

**Problem**
Endpoint `GET /api/v1/monitoring/logs` currently returns `501 Not Implemented`.

**Expected Result**
- Gateway connects to Grafana Loki (`http://localhost:3100`) via LogQL query proxy.
- Admin users can query recent logs filtered by service name and severity (`ERROR`, `WARN`, `INFO`).
- Prometheus PromQL client proxy queries live RED metrics (Rate, Errors, Duration).

**Affected Modules**
`services/gateway`

**Implementation Checklist**
### Backend
- [ ] Implement Loki HTTP client proxy in `services/gateway/internal/handler/monitoring_handler.go`.
- [ ] Replace `501 Not Implemented` with authenticated Loki LogQL query execution.
