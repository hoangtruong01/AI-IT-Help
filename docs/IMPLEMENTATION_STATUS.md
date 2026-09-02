# EOMP implementation status

Last audited: 2026-09-02

Code version: `0.1.0`

Readiness decision: **not yet approved for production**

This is the current evidence baseline. "Implemented" means the code exists and the listed local checks pass; it does not replace deployment, security-owner or disaster-recovery acceptance.

## Gate A status (Chốt luật và khóa bề mặt tấn công)

| Task | Current status | Evidence & Details |
|---|---|---|
| A-01 Authorization matrix | **DONE (Awaiting Owner Sign-off)** | Ma trận 4 role × 11 resources và 8 quyết định sản phẩm đã chốt trong `docs/AUTHORIZATION_MATRIX.md` |
| A-02 Identity trust boundary | **DONE & VERIFIED** | Gateway `StripIdentityHeaders` loại bỏ toàn bộ `X-User-*` và `X-Department-ID` từ client; Bearer token bắt buộc cho route bảo vệ; 4 unit tests pass |
| A-03 Secret policy | **DONE & VERIFIED** | Bắt buộc `JWT_SECRET` ≥ 32 chars ở mọi môi trường; blacklist secret public; cấm mật khẩu dev mặc định; `.env.example` sạch credential |
| A-04 Error sanitization | **DONE & VERIFIED** | 35 điểm rò rỉ 5xx `InternalServerError(fmt.Sprintf(...))` đã được thay bằng generic error kèm `request_id` |
| A-05 Claims reconciliation | **DONE & VERIFIED** | `IMPLEMENTATION_STATUS.md` và `docs/task.md` là nguồn trạng thái duy nhất; Phase 8 claims gắn nhãn `[HISTORICAL ARCHIVE]` |

## Gate B status (Ranh giới dữ liệu và tính trung thực)

| Task | Current status | Evidence & Details |
|---|---|---|
| B-01 Row-level authorization | **IMPLEMENTED — POSTGRESQL RUNTIME VERIFIED** | Helpdesk, Workflow/Approval, Employee và Knowledge áp scope fail-closed theo `Actor`; list/get quan trọng được lọc trong SQL, ngoài scope trả `404`. Asset/CMDB chặn direct call thiếu actor và chỉ cho Agent/Manager/Admin. Ma trận PostgreSQL đã pass cho Employee/Agent/Manager/Admin với own, assigned, unassigned, same/other department; directory Employee chỉ trả name+email; Knowledge internal không tăng view ngoài scope. |
| B-02 User lifecycle for pilot | **IMPLEMENTED — RUNTIME VERIFIED** | PostgreSQL runtime test đã chứng minh create/login/refresh rotation, replay trả 401, admin reset/deactivate thu hồi session và ghi security audit. Integration test chủ động làm audit insert lỗi đã xác nhận user/audit cùng rollback; auth migration `004` được áp dụng thành công trên database có sẵn. Public register bị tắt ở production và không nhận department. |
| B-03 Real reporting & filters | **IMPLEMENTED — RUNTIME E2E VERIFIED** | PostgreSQL + RabbitMQ test đã chứng minh create/assign/resolve cập nhật projection và KPI, queue được tiêu thụ hết; publish hai lần cùng event ID chỉ tạo một projection; date filter, invalid range, CSV và PDF payload đều pass. Assignee ID được lưu như opaque identifier và DLQ binding đã sửa. PDF vẫn là generator nội bộ nếu tiêu chí bắt buộc thư viện được duyệt. |
| B-04 Frontend API contract | **IMPLEMENTED — LOCAL VERIFIED** | Chuẩn hóa `get(url, params)`; dashboard role-aware. `ApiStatePanel` và state classifier phân biệt rõ `empty`, `403 forbidden`, `backend unavailable` trên các data page. Frontend Vitest 14/14, Nuxt typecheck và ESLint đều pass. |
| B-05 Clean baseline migration | **IMPLEMENTED — FRESH MIGRATION VERIFIED** | Baseline Asset/Helpdesk/Workflow/Knowledge/Notification không còn operational INSERT; reference taxonomy/catalog/template/definition được giữ. Fresh migration trên 5 database tạm cho kết quả operational `0`; reference lần lượt Helpdesk=5, Workflow=3, Knowledge=5, Notification=4. `scripts/dev_seed.ps1`/`.sh` chỉ cho development/test và chạy hai lần vẫn đúng một fixture mỗi loại. |

Gate B now meets its **technical implementation and runtime exit criteria**. Formal closure remains **pending owner acceptance** under the evidence rule in `docs/task.md`; this document does not self-approve that sign-off. Local evidence from 2026-09-01: relevant Go test/vet suites pass; frontend Vitest 66/66, Nuxt typecheck and ESLint pass; PostgreSQL authorization/fresh-migration matrices and RabbitMQ reporting tests pass.

## Gate C status (Session, transport và migration safety)

| Task | Current status | Evidence & Details |
|---|---|---|
| C-01 Frontend session & Cookie BFF | **IMPLEMENTED — LOCAL RUNTIME VERIFIED** | Edge route `/api/auth/*` được ưu tiên chuyển tới Nuxt BFF; Nitro gọi Gateway qua server-only `NUXT_API_BASE_URL`. Refresh cookie có `HttpOnly`, `Secure` ở production, `SameSite=Strict`, path `/api/auth`; access token chỉ ở Pinia memory. POST auth kiểm tra exact Origin chống CSRF. Refresh mutex dùng chính utility production và cô lập theo auth-store/SSR context; route guard không rotate cookie trong SSR subrequest. Frontend Vitest `73/73`, Nuxt typecheck, ESLint và production build pass; Nitro smoke test trả `403` khi thiếu Origin và `401` khi same-origin nhưng thiếu cookie. |
| C-02 TLS & Private Monitoring | **IMPLEMENTED — STAGING ACCEPTANCE PENDING** | Compose publish `443`, yêu cầu mount certificate/key, redirect toàn bộ `80→443`, thêm HSTS và security headers. Static Ingress và Helm ưu tiên BFF route, bật TLS/HSTS và không còn route public Grafana/Prometheus; observability service chỉ ở internal network. Compose render validation pass. Chưa có Helm render trên máy này, `nginx -t` với certificate thật, TLS scanner hoặc kiểm tra từ untrusted network. |
| C-03 Migration Concurrency Control | **IMPLEMENTED — POSTGRESQL RUNTIME VERIFIED** | `pg_advisory_lock` (Lock ID `8424119472649191`) giữ trên dedicated connection quanh tracker/discovery/apply. Bài test PostgreSQL 17 thật chạy 5 runner đồng thời đã pass: migration effect đúng một lần và chỉ có một tracker row. Shared Go suite pass. |

Gate C has **3/3 implementations complete and 2/3 tasks locally/runtime verified**. Formal closure remains pending C-02 staging transport evidence: render/deploy the Helm chart with a real TLS secret, run `nginx -t`/TLS policy scan, and prove Grafana/Prometheus are unreachable from an untrusted network. This repository must not self-approve those environment-level checks.

## Gate D status (Bằng chứng pilot & Verification Evidence)

| Task | Current status | Evidence & Details |
|---|---|---|
| D-01 PostgreSQL Integration Suite | **PARTIAL — CORE POSTGRESQL RUNTIME VERIFIED** | A strict run against PostgreSQL 17.11 used four isolated temporary databases and allowed no skips. Auth rotation/replay/session revocation/rollback, Helpdesk row scope/100-way sequence/CAS, Audit HMAC/append-only protection, migration concurrency and parameterized-query payloads passed. Test fixture and assertion defects exposed by the first real run were corrected. Coverage required by the plan for Notification, Reporting and reporting date filters is still absent, and CI does not yet provision ephemeral PostgreSQL automatically. |
| D-02 Real API E2E & Browser E2E | **PARTIAL — LOCAL DEPLOYED-STACK E2E PASSED** | The real Docker Gateway/Auth/services completed login, create → assign → comment → resolve, stale-version `409`, refresh rotation and logout/revocation. Spoofed identity returned `401`, cross-department access returned `404`, and PostgreSQL assertions confirmed Helpdesk, Audit, Reporting and Notification effects. Evidence is retained at `docs/evidence/gate-d/deployed_stack_e2e.json`. Playwright and a release/staging run remain open. |
| D-03 Staging Release Evidence & DR | **PARTIAL — LOCAL IMAGE/LOAD/DR VERIFIED** | One `eomp` Docker project runs 20 healthy containers. Twelve application images have immutable local digests and run as `10001:10001` (`docs/evidence/gate-d/image_manifest.json`). Controlled local load produced 2,928 requests with zero failures, p95 30.3439ms and p99 65.0058ms (`docs/evidence/gate-d/local_load_summary.json`). Nine-database restore passed in 6.602s (`docs/evidence/gate-d/dr_evidence_20260902.json`). CVE scan, k6/TLS staging, Helm/Kubernetes rollout, network isolation and WAL/full-service RTO remain open. |

Gate D is **not closed**. Current engineering estimate is **90% local technical readiness** and **70% formal Gate D evidence**; all three tasks remain formally partial. Gate B owner acceptance and Gate C staging acceptance also remain prerequisites for pilot approval.

### Gate D audit evidence — 2026-09-02

- All Go modules passed `go test ./...` and `go vet ./...`.
- Strict PostgreSQL integration passed against four isolated temporary databases with `INTEGRATION_REQUIRED=1`; no database test was skipped. All temporary databases were removed afterward.
- The in-process HTTP contract harness passed three lifecycle/security contract tests and remains classified as contract evidence.
- Real local deployed-stack E2E passed through Gateway/Auth/services, including lifecycle, security probes, session rotation/revocation and downstream PostgreSQL assertions; evidence is in `docs/evidence/gate-d/deployed_stack_e2e.json`.
- Frontend Vitest passed `81/81`; Nuxt typecheck, ESLint and production build passed. These are local unit/contract/build checks; no browser was launched.
- OpenAPI parity passed `107/107` runtime operations (`go run scripts/check_openapi_coverage.go`).
- Docker Server 28.4.0 runs one unified `eomp` project with 20 healthy containers. Twelve application images were built, assigned immutable local digests and verified as non-root; evidence is in `docs/evidence/gate-d/image_manifest.json`.
- Controlled local load passed at 100 VU for 30 seconds: 2,928 requests, zero failures, p95 30.3439ms and p99 65.0058ms; evidence is in `docs/evidence/gate-d/local_load_summary.json`.
- Nine-database backup/restore passed on PostgreSQL 17.11: oldest snapshot age 54.718s and database restore duration 6.602s. This does not prove WAL RPO or full-service RTO.
- Docker Scout CVE output was unavailable because Docker Desktop is not logged into Docker ID; k6/TLS staging and Helm deployment evidence also remain unavailable. The local controlled-load result is not counted as formal staging acceptance.
- Detailed gap analysis and the required evidence checklist are recorded in `docs/GATE_D_VERIFICATION_EVIDENCE.md`.

### Gate C verification evidence — 2026-09-02

- Frontend: Vitest `73/73`, Nuxt typecheck, ESLint and production build passed. The built Nitro server registered all four `/api/auth/*` routes; same-origin CSRF smoke cases returned the expected `403/401` results and emitted CSP/frame/content-type/permissions headers.
- Compose: `docker compose -f deploy/docker-compose.prod.yml config --quiet` passed with validation-only secret and certificate paths. No container was started by this syntax check.
- Migration: a disposable PostgreSQL 17 container ran `TestRunMigrationsFiveConcurrentRunnersPostgres`; five concurrent connections completed without duplicate execution or tracker rows. The container was stopped and auto-removed afterward.
- Environment limitation: Helm and local Nginx binaries are not installed; no organization-issued certificate, staging ingress or untrusted-network vantage point was available.

### Gate B verification evidence — 2026-09-01

- Helpdesk PostgreSQL matrix: Employee list only own ticket; Agent list own assignment plus unassigned queue; Manager list only department; Admin list all. Employee/Agent/Manager get or comments outside scope returned `404`; missing actor returned `401`.
- Workflow PostgreSQL matrix: Employee list/get own; Agent list own+assigned but not unassigned; Manager list/approval only department; Admin list all. Employee/Agent approval returned `403`; cross-department approval decision returned `404`; Manager without department returned `403`.
- Employee/Knowledge/Asset: Employee directory returned same-department name+email only and own full profile; peer/other profile returned `404`. Employee Knowledge list/search returned public article only; internal get returned `404` without changing `view_count`. Asset/CMDB returned `401` for missing/role-only identity, `403` for Employee, and `200` for Agent/Manager/Admin.
- Clean migration drill used five disposable `gateb_clean_*` databases, ran every migration in filename order, and removed the databases afterward. Operational counts were Asset `0/0`, Helpdesk `0/0`, Workflow `0/0`, Knowledge `0/0`, Notification `0`; retained reference counts were `5/3/5/4` for Helpdesk/Workflow/Knowledge/Notification.
- Development seed was run twice after migrations. The second run returned `INSERT 0 0` for every fixture and exact-ID counts remained one. All authorization and seed fixtures were then deleted by exact IDs; verification counts returned zero.

## Verified inventory

- 11 Go services plus one Nuxt web application.
- Service ports: gateway 8080, auth 8081, employee 8082, asset 8083, helpdesk 8084, workflow 8085, notification 8086, knowledge 8087, AI 8088, audit 8089 and reporting 8090.
- Nine PostgreSQL databases: auth, employee, asset, helpdesk, workflow, notification, knowledge, audit and reporting.
- 102 registered `/api/v1` operations. All 102 are present in the OpenAPI operation inventory.
- `go run scripts/check_openapi_coverage.go` validates route compatibility and exact runtime/OpenAPI parity. Jenkins now runs this gate.

## Remediation completed through 2026-08-31

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
- Asset/CMDB, employee, knowledge article, change, ticket, problem and workflow-approval critical transitions use optimistic compare-and-swap and return HTTP 409 for stale versions.
- Employee ticket access is requester-scoped; requester identity is taken from the gateway identity, internal comments are hidden and cross-owner resource probes return 404.
- Workflow instances and logs are requester-scoped for employees. Approval decisions are limited to the assigned user or configured role pool, CAB identity comes from the authenticated request and initial workflow state is written atomically.
- Notification reads are per recipient, including broadcasts; list/count/read queries are parameterized and read receipts enforce ownership.
- Runtime dashboards now show API results or explicit unavailable/empty states instead of static operational claims. The AI page reads provider/Qdrant state from `/api/v1/ai/status`.
- Invalid non-hex UUID fixtures were corrected. Exact cleanup migrations remove development employees, inventory/CMDB, knowledge content, tickets/problems, workflow runs, notifications and reporting telemetry without broad production-data deletes.
- Asset/CMDB, Problem, Change, Monitoring, Reporting, Audit and workflow aggregate metrics are role-restricted in both gateway routes and frontend navigation; hard-coded navigation counts and the fictional "100% HEALTHY" sidebar were removed.
- Human-readable ticket, problem, workflow and change numbers are allocated by PostgreSQL sequences rather than row counts or timestamp fallbacks, preventing reuse after deletion and common concurrent-create collisions.
- AI/Qdrant mock fallback is disabled for real production providers unless `ALLOW_MOCK_AI=true` is explicitly set. Production manifests default to OpenAI and require an external key.
- The conflicting Helpdesk routes `/tickets/asset/{assetId}` and `/tickets/{id}/assign` were replaced with an unambiguous asset-ticket route.
- Backup validation no longer creates simulated evidence; release DR acceptance requires an actual temporary-database restore.
- CI security gates are fail-closed for gosec, govulncheck and Trivy, and include frontend tests/typecheck plus Go race tests.

## Verification evidence

Completed locally on 2026-08-31 after the remediation above:

- `go test ./...` and `go vet ./...` passed for all 13 workspace modules, including `tests/e2e`.
- OpenAPI YAML parsed successfully and Redocly reported a valid OpenAPI document.
- Runtime/OpenAPI gate passed: `102/102` operations documented, with no `http.ServeMux` conflicts.
- Frontend Vitest (3 tests), Nuxt typecheck, ESLint and production build passed.

Environment-limited checks:

- Docker Desktop runtime integration passed for PostgreSQL 17 and RabbitMQ 4: all infrastructure containers were healthy; Auth/Helpdesk/Reporting readiness returned 200; migrations `auth/004` and `reporting/001..005` applied; B-02 lifecycle and B-03 event/KPI/export flows passed. Runtime fixtures were removed by exact identifiers after verification.
- Helm rendering was not executed because Helm is not installed.
- A nine-database PostgreSQL restore drill passed and is retained at `docs/evidence/gate-d/dr_evidence_20260902.json`. WAL-based RPO and full-service RTO remain unverified.
- No production AI provider key or external Qdrant instance was available for an end-to-end AI validation.

## Remaining release blockers

| Priority | Blocker | Acceptance condition |
|---|---|---|
| P0 | Previously committed secrets may already be compromised | Rotate PostgreSQL, JWT, RabbitMQ, MinIO, Grafana and any provider credentials in every environment and secret store |
| P0 | Authorization matrix is not owner-approved | Record a traceable owner approval for `AUTHORIZATION_MATRIX.md` revision 1.0 before implementing dependent authorization work |
| P0 | No real deployment/runtime proof | Build and scan every image, render Helm, deploy to a test cluster and pass smoke/integration checks |
| P0 | Full-service DR targets remain unverified | Database restore evidence now exists; still prove WAL-based RPO and application/infrastructure recovery within the agreed full-service RTO |
| P1 | Downstream services still trust gateway-injected plaintext identity headers | Add signed internal identity or validate the downstream JWT; keep service ports unreachable from untrusted networks |
| P1 | Audit immutability is not yet enforced by a dedicated restricted DB role | Run the audit service with append-only grants, document HMAC key rotation and validate tamper detection against PostgreSQL |
| P1 | Monitoring has probes but no real RED/log backend | Integrate Prometheus queries and a Loki-compatible log backend; keep unavailable fields explicit until then |
| P1 | OpenAPI has complete operation parity but many generic bodies/responses | Replace generic schemas with domain models and add request/response conformance tests |
| P1 | Optimistic locking and transaction coverage is not universal | Audit every remaining mutable aggregate and add PostgreSQL integration tests for CAS, rollback and migration-upgrade behavior |
| P1 | No browser E2E suite or measured load run | Add Playwright journeys and run k6 against a deployed environment |
| P1 | Production AI has not been exercised with a real provider | Inject a real key/provider, validate Qdrant ingestion and prove failure behavior without fabricated fallback |

## Release rule

Do not describe EOMP as "100% complete", "fully certified" or "production ready" until every P0 condition has objective evidence and accountable owners sign the handover acceptance.
