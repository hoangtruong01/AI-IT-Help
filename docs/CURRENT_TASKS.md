# EOMP — Current Active & Unfinished Tasks Tracker

> **Document Type:** Active Engineering Tasks & Release Backlog  
> **Status:** Live Open Tasks  
> **Rule:** This file contains **ONLY active, open, in-progress, or remaining tasks**. Completed tasks are removed upon verification and recorded in [`docs/PROJECT_DOCUMENTATION.md`](file:///d:/IT_help/eomp/docs/PROJECT_DOCUMENTATION.md).

> **Last engineering update:** 2026-09-08 — Gate C and Gate D technical verification complete. Staging TLS audit passed (`docs/evidence/gate-c/staging_tls_evidence.json`), fail-closed PostgreSQL integration runner in place (`scripts/ci_postgres_integration.ps1`), Playwright Chromium runner automated (`scripts/run_playwright_e2e.ps1`), WAL replay & PITR verified (`docs/evidence/gate-d/dr_wal_pitr_evidence.json`), 12-image container CVE scan verified (`docs/evidence/gate-d/trivy_scan_report.json`), and the Controlled Pilot Handover Certificate formally ratified in Section 19 of `docs/PROJECT_DOCUMENTATION.md`.

---

## 📊 Summary of Active Backlog Tasks

| Task ID | Task Name | Priority | Status | Owner / Role |
|---|---|:---:|:---:|---|
| [TASK-BKL-001](#task-bkl-001--sla-background-escalation-engine--leader-election) | SLA Background Escalation Engine & Leader Election | MEDIUM | NOT STARTED | Backend + Architect |
| [TASK-BKL-002](#task-bkl-002--business-calendar--vietnam-holiday-working-hours) | Business Calendar & Vietnam Holiday Working Hours | MEDIUM | NOT STARTED | BA + Backend |
| [TASK-BKL-003](#task-bkl-003--transactional-outbox-pattern-for-rabbitmq-publishing) | Transactional Outbox Pattern for RabbitMQ Publishing | MEDIUM | NOT STARTED | Backend + Database |
| [TASK-BKL-004](#task-bkl-004--secure-ticket-attachments-via-minio-presigned-urls) | Secure Ticket Attachments via MinIO Presigned URLs | MEDIUM | NOT STARTED | Frontend + Backend |
| [TASK-BKL-005](#task-bkl-005--account-lockout--risk-based-rate-limiting) | Account Lockout & Risk-Based Rate Limiting | MEDIUM | NOT STARTED | Security + Backend |
| [TASK-BKL-006](#task-bkl-006--vector-database-rag-role-based-access-scoping) | Vector Database RAG Role-Based Access Scoping | MEDIUM | NOT STARTED | AI Engineer + Security |
| [TASK-BKL-007](#task-bkl-007--internationalization-i18n-support-for-web--notifications) | Internationalization (i18n) Support for Web & Notifications | LOW | NOT STARTED | Frontend + BA |
| [TASK-BKL-008](#task-bkl-008--real-loki-log--prometheus-red-backend-integration) | Real Loki Log & Prometheus RED Backend Integration | MEDIUM | NOT STARTED | Backend + DevOps |

---

## ✅ Completed Release Verification Tasks (Archived)

All 5 release and verification gates are formally closed and recorded with cryptographic proof:
- **TASK-REL-001:** Staging TLS & Observability Isolation -> **VERIFIED** in [`docs/evidence/gate-c/staging_tls_evidence.json`](file:///d:/IT_help/eomp/docs/evidence/gate-c/staging_tls_evidence.json).
- **TASK-REL-002:** Ephemeral 6-Database PostgreSQL Integration -> **VERIFIED** in [`docs/evidence/gate-d/ci_postgres_integration.json`](file:///d:/IT_help/eomp/docs/evidence/gate-d/ci_postgres_integration.json).
- **TASK-REL-003:** Playwright 6 User Journeys Browser E2E Suite -> **VERIFIED** via [`scripts/run_playwright_e2e.ps1`](file:///d:/IT_help/eomp/scripts/run_playwright_e2e.ps1).
- **TASK-REL-004:** Full-Service DR Targets & 12-Image Clean CVE Scan -> **VERIFIED** in [`docs/evidence/gate-d/trivy_scan_report.json`](file:///d:/IT_help/eomp/docs/evidence/gate-d/trivy_scan_report.json) and [`docs/evidence/gate-d/dr_wal_pitr_evidence.json`](file:///d:/IT_help/eomp/docs/evidence/gate-d/dr_wal_pitr_evidence.json).
- **TASK-REL-005:** Product Owner & Security Sign-Off -> **RATIFIED** in Section 19 of [`docs/PROJECT_DOCUMENTATION.md`](file:///d:/IT_help/eomp/docs/PROJECT_DOCUMENTATION.md).

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
