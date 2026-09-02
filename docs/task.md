# 📋 EOMP — Task Tracker & Báo Cáo Nâng Cấp (Gate-based)

> **Baseline Commit:** `997a3d3` | **Cập nhật:** 2026-09-02
> **Trạng thái:** **Gate A** đã hoàn tất; **Gate B chờ owner acceptance; Gate C chờ C-02 staging acceptance**
> **Mục tiêu:** Pilot có kiểm soát → chỉ khi Gate A–D pass 100%  
> **Quy tắc:** Task chỉ đóng khi có đủ 4 bằng chứng: `Code evidence` + `Automated test` + `Runtime evidence` + `Owner acceptance`.

---

## 📊 Bảng Tổng Hợp Tiến Độ Các Gate

| Gate | Mô tả phạm vi | Trọng số | Trạng thái | Số task |
|---|---|:---:|:---:|:---:|
| **Gate A** | **Chốt luật và khóa bề mặt tấn công** (Trust boundary, secrets, errors, matrix) | P0 | ✅ **DONE** (100%) | 5/5 |
| **Gate B** | **Ranh giới dữ liệu & tính trung thực** (Row-level auth, user lifecycle, real reports) | P0 | 🟡 **TECHNICALLY VERIFIED — OWNER ACCEPTANCE PENDING** | 5/5 technical |
| **Gate C** | **Session, transport và migration safety** (Cookie BFF, refresh mutex, TLS, locks) | P1 | 🟡 **IMPLEMENTED — C-02 STAGING ACCEPTANCE PENDING** | 3/3 code; 2/3 verified |
| **Gate D** | **Bằng chứng pilot** (Postgres integration suite, real E2E, staging verification) | P0 | 🟡 **LOCAL 90% / FORMAL EVIDENCE 70%** | 0/3 formally accepted |

---

## 🛡️ Gate A — Chốt Luật & Khóa Bề Mặt Tấn Công

> 🚫 **Điều kiện tiên quyết:** Không bắt đầu pilot nếu bất kỳ task nào trong Gate A chưa đạt.

### ✅ A-01 — Ma trận authorization và quyết định sản phẩm
- **Priority:** P0 | **Effort:** 2–3 ngày | **Status:** `DONE` (Awaiting Owner Formal Sign-off)
- **Dependency:** Không
- **Owner:** BA + Product Owner
- **Tài liệu bàn giao:** [`docs/AUTHORIZATION_MATRIX.md`](file:///d:/IT_help/eomp/docs/AUTHORIZATION_MATRIX.md)
- [x] **Ma trận 4 role × 11 resources × actions × scope:**
  - [x] Ticket (list/get/create/update/delete/comment/timeline)
  - [x] Asset (list/get/create/update/assign/revoke)
  - [x] Employee (list/get/create/update/department)
  - [x] Workflow (list/get/create/approve/reject/logs)
  - [x] Knowledge (list/get/create/update/delete/runbook)
  - [x] Audit (list/get/verify HMAC chain)
  - [x] Report (view/export CSV/PDF)
- [x] **Chốt 8 quyết định sản phẩm:**
  - [x] *Employee xem ticket:* Chỉ xem ticket mình tạo (`scope: own`).
  - [x] *Agent xem ticket:* Xem ticket được gán + unassigned queue (`scope: assigned + queue`).
  - [x] *Chính sách đăng ký:* Public register chỉ tạo `ROLE_EMPLOYEE`; Production dùng admin-created.
  - [x] *Mapping phòng ban:* Admin gán, không nhận `department_id` tự do từ client.
  - [x] *Quyền đóng ticket:* Agent chuyển `RESOLVED`, auto-close sau 5 ngày nếu không có phản hồi.
  - [x] *SLA pause:* Trạng thái `WAITING_USER` / `PENDING_CUSTOMER` tạm dừng đồng hồ tính SLA.
  - [x] *Lịch làm việc:* 08:00–17:30 Thứ 2 – Thứ 6, nghỉ lễ VN; Sự cố `URGENT` hỗ trợ 24/7.
  - [x] *CSAT:* Tạm ẩn khỏi UI/Report cho tới khi triển khai module khảo sát độc lập.
  - [x] *Ticket ↔ CI (CMDB):* Thuộc tính tùy chọn, không ép buộc.
- [x] Sinh test matrix trực tiếp từ ma trận trong tài liệu đặc tả.

---

### ✅ A-02 — Sửa identity trust boundary & chống header spoofing
- **Priority:** P0 | **Effort:** 2–3 ngày | **Status:** `DONE`
- **Files:**
  - [`services/gateway/internal/middleware/auth.go`](file:///d:/IT_help/eomp/services/gateway/internal/middleware/auth.go)
  - [`services/gateway/cmd/server/main.go`](file:///d:/IT_help/eomp/services/gateway/cmd/server/main.go)
  - [`services/auth/cmd/server/main.go`](file:///d:/IT_help/eomp/services/auth/cmd/server/main.go)
  - [`packages/shared/pkg/middleware/cors.go`](file:///d:/IT_help/eomp/packages/shared/pkg/middleware/cors.go)
  - [`services/gateway/internal/middleware/auth_test.go`](file:///d:/IT_help/eomp/services/gateway/internal/middleware/auth_test.go)
- [x] Viết middleware `StripIdentityHeaders` — loại bỏ toàn bộ biến thể `X-User-*` và `X-Department-ID` từ client.
- [x] Đặt `StripIdentityHeaders` ở outermost middleware trong stack Gateway.
- [x] Tách route auth public vs protected:
  - Public: `/login`, `/register`, `/refresh`, `/logout`
  - Protected (cần Bearer JWT): `/me`, `/login-history`
- [x] `GatewayAuth`: Ghi đè `X-User-ID`, `X-User-Email`, `X-User-Role`, `X-User-Department`, `X-User-Name` vô điều kiện (kể cả rỗng).
- [x] Bỏ hoàn toàn nhánh auth service chấp nhận header client để skip Bearer.
- [x] Xoá `X-User-*` khỏi CORS `Access-Control-Allow-Headers`.
- [x] **Automated Tests:** 4 unit tests trong `auth_test.go` (**PASS 100%**).

---

### ✅ A-03 — Secret policy & loại bỏ hardcoded credentials
- **Priority:** P0 | **Effort:** 1–2 ngày | **Status:** `DONE`
- **Files:**
  - [`services/gateway/internal/config/config.go`](file:///d:/IT_help/eomp/services/gateway/internal/config/config.go)
  - [`services/auth/internal/config/config.go`](file:///d:/IT_help/eomp/services/auth/internal/config/config.go)
  - [`.env.example`](file:///d:/IT_help/eomp/.env.example)
  - [`services/auth/internal/service/auth_service_test.go`](file:///d:/IT_help/eomp/services/auth/internal/service/auth_service_test.go)
  - [`tests/e2e/*.go`](file:///d:/IT_help/eomp/tests/e2e/)
- [x] Bắt buộc `JWT_SECRET` ≥ 32 ký tự, fail-fast ở **mọi** môi trường (`APP_ENV`).
- [x] Blacklist chuỗi public `eomp-enterprise-super-secret-jwt-key-2026` ở runtime.
- [x] Non-test runtime bắt buộc `POSTGRES_PASSWORD`, cấm dùng mật khẩu mặc định `eomp_dev_password`.
- [x] Tách secret dùng trong unit test / E2E test độc lập.
- [x] `.env.example`: Dọn sạch credential, thay bằng placeholder và hướng dẫn sinh ngẫu nhiên `openssl rand -base64 48`.

---

### ✅ A-04 — Sanitize lỗi nội bộ & correlation ID
- **Priority:** P0 | **Effort:** 2–3 ngày | **Status:** `DONE`
- **Files:**
  - [`packages/shared/pkg/errors/errors.go`](file:///d:/IT_help/eomp/packages/shared/pkg/errors/errors.go)
  - Toàn bộ các file `*_service.go` trong 9 microservices
- [x] Sửa error writer: Mọi lỗi 5xx chỉ trả thông điệp chung `an internal error occurred` + đính kèm `request_id`.
- [x] Thay thế toàn bộ **35 điểm rò rỉ** `InternalServerError(fmt.Sprintf("...%v", err))` thành generic error message.
- [x] `grep -rn "InternalServerError(fmt.Sprintf" services` trả về **0 kết quả**.

---

### ✅ A-05 — Reconcile tài liệu & chuẩn hóa claims
- **Priority:** P0 (governance) | **Effort:** 1–2 ngày | **Status:** `DONE`
- **Files:**
  - [`docs/IMPLEMENTATION_STATUS.md`](file:///d:/IT_help/eomp/docs/IMPLEMENTATION_STATUS.md)
  - [`docs/PHASE_8_ENTERPRISE_VALIDATION_AND_PORTFOLIO_CASE_STUDY.md`](file:///d:/IT_help/eomp/docs/PHASE_8_ENTERPRISE_VALIDATION_AND_PORTFOLIO_CASE_STUDY.md)
  - [`docs/testing.md`](file:///d:/IT_help/eomp/docs/testing.md)
- [x] `IMPLEMENTATION_STATUS.md` là nguồn trạng thái duy nhất.
- [x] Gắn nhãn `[HISTORICAL ARCHIVE]` cho các tuyên bố 100% production trong Phase 8 case study.
- [x] Cập nhật `docs/testing.md` làm rõ kết quả kiểm thử là in-memory simulation / unit suite.

---

## 🔒 Gate B — Ranh Giới Dữ Liệu & Tính Trung Thực

> ⚠️ **Điều kiện tiên quyết:** Gate A đã pass 100%.

### 🟢 B-01 — Authorization theo bản ghi (Row-Level / Scope-Based)
- **Priority:** P0 | **Effort:** 5–8 ngày | **Status:** `IMPLEMENTED — POSTGRESQL RUNTIME VERIFIED`
- **Files:**
  - [`packages/shared/pkg/middleware/auth.go`](file:///d:/IT_help/eomp/packages/shared/pkg/middleware/auth.go)
  - [`services/helpdesk/internal/handler/ticket.go`](file:///d:/IT_help/eomp/services/helpdesk/internal/handler/ticket.go)
  - [`services/helpdesk/internal/repository/repository.go`](file:///d:/IT_help/eomp/services/helpdesk/internal/repository/repository.go)
  - [`services/workflow/internal/handler/workflow.go`](file:///d:/IT_help/eomp/services/workflow/internal/handler/workflow.go)
- [x] Tạo kiểu `Actor` (`ID`, `Email`, `Role`, `DepartmentID`, `Name`) và `GetActor(ctx)` trong shared middleware.
- [x] **Helpdesk Row-Level Scope:**
  - [x] `ListTickets`: Áp dụng filter trực tiếp trong SQL `WHERE` theo role:
    - `ROLE_EMPLOYEE`: `requester_id = $N`
    - `ROLE_AGENT`: `assignee_id = $N OR assignee_id IS NULL`
    - `ROLE_MANAGER`: `department_id = $N`
    - `ROLE_ADMIN`: Xem toàn bộ
  - [x] `GetTicket`: Scope được đưa vào SQL; ngoài scope trả `404 Not Found` để chống enumeration.
  - [x] `AddComment`, `ListComments`, `ListTimeline`: Kế thừa xác thực quyền từ ticket.
  - [x] `CreateTicket`: Employee luôn bị gán `requester_id = actor.ID`; Operators được tạo on-behalf.
- [x] Hoàn tất scope cho Workflow/Approval, Asset/CMDB, Employee và Knowledge theo ma trận A-01; direct call thiếu actor fail-closed.
- [x] PostgreSQL runtime matrix pass cho cả 4 role với own/other department/assigned/unassigned; ngoài scope trả `404`, forbidden role trả `403`, thiếu actor trả `401`.

---

### 🟢 B-02 — User lifecycle tối thiểu cho pilot
- **Priority:** P0 | **Effort:** 5–7 ngày | **Status:** `IMPLEMENTED — RUNTIME VERIFIED`
- **Files:**
  - [`services/auth/internal/model/user.go`](file:///d:/IT_help/eomp/services/auth/internal/model/user.go)
  - [`services/auth/internal/handler/auth.go`](file:///d:/IT_help/eomp/services/auth/internal/handler/auth.go)
  - [`services/auth/internal/service/auth_service.go`](file:///d:/IT_help/eomp/services/auth/internal/service/auth_service.go)
  - [`services/auth/internal/repository/user_repository.go`](file:///d:/IT_help/eomp/services/auth/internal/repository/user_repository.go)
  - [`services/auth/internal/service/auth_service_test.go`](file:///d:/IT_help/eomp/services/auth/internal/service/auth_service_test.go)
  - [`services/gateway/cmd/server/main.go`](file:///d:/IT_help/eomp/services/gateway/cmd/server/main.go)
- [x] **Admin User Management:**
  - [x] `GET /api/v1/users`: Phân trang, tìm kiếm email/tên cho Admin & Manager.
  - [x] `POST /api/v1/users`: Admin tạo user, gán role + phòng ban.
  - [x] `PATCH /api/v1/users/{id}`: Admin cập nhật role, phòng ban, kích hoạt/vô hiệu hóa.
  - [x] **Self-promotion prevention:** User tự đổi role của chính mình bị chặn với `403 Forbidden`.
  - [x] Deactivate user: Thu hồi lập tức toàn bộ refresh token và active sessions của user.
- [x] **Password Management:**
  - [x] `POST /api/v1/auth/change-password`: Xác thực mật khẩu cũ + kiểm tra độ mạnh mới; thu hồi session cũ.
  - [x] `POST /api/v1/auth/reset-password/{id}`: Admin reset password cho tài khoản khác.
  - [x] `RotateRefreshTokenAtomic`: Xoay vòng refresh token an toàn trong 1 transaction SQL.
- [x] Security audit + account/password change + refresh-token revoke dùng cùng transaction.
- [x] Public register bị tắt ở production và không được nhận department từ payload.
- [x] PostgreSQL runtime: create/login/rotate/replay/reset/deactivate/audit pass; replay và token đã revoke trả `401`, inactive login trả `403`.
- [x] PostgreSQL integration test cố tình làm audit insert lỗi xác nhận user/audit rollback cùng transaction; migration `004` upgrade thành công trên database có sẵn.

---

### 🟡 B-03 — Reporting read model và filter thật
- **Priority:** P0 | **Effort:** 5–8 ngày | **Status:** `IMPLEMENTED — RUNTIME E2E VERIFIED`
- **Files:**
  - [`services/reporting/internal/repository/reporting_repository.go`](file:///d:/IT_help/eomp/services/reporting/internal/repository/reporting_repository.go)
  - [`services/reporting/internal/service/reporting_service.go`](file:///d:/IT_help/eomp/services/reporting/internal/service/reporting_service.go)
- [x] Áp dụng `range/start_date/end_date` vào mọi aggregate và raw export query.
- [x] RabbitMQ projection idempotent từ ticket create/assign/status event; rollup khớp schema.
- [x] **PDF Export Enhancement:**
  - [x] Tính toán chỉ số KPI động từ danh sách bản ghi thực tế.
  - [x] Escape ký tự đặc biệt PDF `(`, `)`, `\` chống vỡ layout.
  - [x] Xoá bỏ CSAT giả lập và các hằng số KPI hardcode `31.8 / 4.86`.
- [x] PostgreSQL + RabbitMQ runtime: create/assign/resolve cập nhật raw projection, daily/category/department/agent KPI; queue về 0; filter/export pass.
- [x] Runtime redelivery: publish hai lần cùng một event ID chỉ tạo một `reporting_processed_events` và một raw projection.
- [ ] Thay generator PDF nội bộ bằng thư viện PDF được duyệt nếu đây là điều kiện bắt buộc của Gate B.

---

### 🟢 B-04 — Frontend API contract & chuẩn hóa tham số
- **Priority:** P1 | **Effort:** 3–5 ngày | **Status:** `IMPLEMENTED — LOCAL VERIFIED`
- **Files:**
  - [`apps/web/app/composables/useApi.ts`](file:///d:/IT_help/eomp/apps/web/app/composables/useApi.ts)
- [x] Chuẩn hóa duy nhất `useApi.get(url, params)` và xóa toàn bộ caller `{ params: { ... } }`.
- [x] Contract test kiểm tra URL/encoding và từ chối nested object; dashboard chỉ gọi endpoint hợp role.
- [x] `ApiStatePanel` + classifier hiển thị riêng `empty`, `403`, `backend unavailable` trên các data page; Vitest 14/14, typecheck và ESLint pass.

---

### 🟢 B-05 — Migration và seed sạch
- **Priority:** P1 | **Effort:** 2–4 ngày | **Status:** `IMPLEMENTED — FRESH MIGRATION VERIFIED`
- **Files:**
  - [`services/*/migrations/*.sql`](file:///d:/IT_help/eomp/services/)
- [x] Reporting baseline không còn insert telemetry demo; giữ exact cleanup migration cho upgrade.
- [x] Gỡ operational demo INSERT khỏi baseline Asset/Helpdesk/Workflow/Knowledge/Notification; giữ reference data và exact cleanup migrations cho upgrade.
- [x] Fresh migration trên 5 database tạm tạo 0 operational record và đúng reference data.
- [x] `make seed`, `scripts/dev_seed.ps1`/`.sh` chỉ cho development/test; chạy hai lần vẫn đúng một fixture mỗi loại.

---

## ⚡ Gate C — Session, Transport & Migration Safety

> ⚠️ **Điều kiện tiên quyết:** Không đóng Gate C dựa trên giả định Gate B đã pass; các mục còn mở của Gate B vẫn là blocker.

### ✅ C-01 — Frontend session architecture & Cookie BFF
- **Priority:** P1 (release blocker) | **Effort:** 4–6 ngày | **Status:** `IMPLEMENTED — LOCAL RUNTIME VERIFIED`
- **Dependency:** A-03
- **Files:**
  - `apps/web/server/api/auth/*.ts`
  - `apps/web/server/utils/*.ts`
  - `apps/web/app/middleware/auth.global.ts`
  - `apps/web/app/composables/useApi.ts`
  - `apps/web/app/stores/auth.ts`
  - `apps/web/app/utils/refresh-mutex.ts`
- [x] Cookie strategy:
  - [x] BFF / Server Route set HttpOnly Secure SameSite cookie (`eomp_refresh_token`) cho refresh token.
  - [x] Access token lưu trong memory (Pinia store), không persist ra JS-readable storage (chống XSS token theft).
  - [x] Nginx/Ingress route `/api/auth/*` tới Nuxt trước generic `/api/*`; Nitro dùng server-only internal Gateway URL.
  - [x] Exact-Origin CSRF check cho login/refresh/logout; cookie giới hạn path `/api/auth`.
- [x] Refresh Mutex:
  - [x] 5 requests 401 đồng thời chỉ trigger đúng 1 lần gọi refresh token (`auth_mutex.test.ts` pass).
  - [x] Queue các pending requests trong thời gian chờ token mới và tự động retry với token mới.
- [x] Route Guard:
  - [x] Đổi middleware thành `auth.global.ts` tự động bảo vệ mọi trang.
  - [x] Chặn truy cập trái quyền ở giao diện client theo ma trận RBAC (`/audit`, `/reports`, `/monitoring`, `/changes`, `/problems`, `/assets`). Backend vẫn là security authority.
- [x] Evidence: Vitest `73/73`, typecheck, ESLint, production build và Nitro CSRF/header smoke test pass.

---

### 🟡 C-02 — TLS & Private Monitoring
- **Priority:** P0 | **Effort:** 2–4 ngày | **Status:** `IMPLEMENTED — STAGING ACCEPTANCE PENDING`
- **Files:**
  - `deploy/nginx/conf.d/eomp.conf`
  - `deploy/docker-compose.prod.yml`
  - `deploy/kubernetes/manifests/08-ingress.yaml`
  - `deploy/kubernetes/helm/eomp/templates/ingress.yaml`
- [x] HTTPS redirect (80 → 443) & HSTS header (`Strict-Transport-Security: max-age=31536000; includeSubDomains; preload`).
- [x] Gỡ `/monitoring/grafana/` và `/monitoring/prometheus/` khỏi Ingress/Nginx công khai để bảo vệ telemetry.
- [x] Cấu hình CSP: Bỏ `unsafe-eval`, tương thích chuẩn Nuxt SSR.
- [x] Compose yêu cầu certificate/key ngoài repo, publish 443 và render validation pass.
- [ ] Render/deploy Helm với TLS secret thật; chạy `nginx -t` và TLS scanner theo policy tổ chức.
- [ ] Xác minh từ untrusted network không truy cập được Grafana/Prometheus.

---

### ✅ C-03 — Migration Concurrency Control
- **Priority:** P1 | **Effort:** 1–2 ngày | **Status:** `DONE & VERIFIED`
- **Files:**
  - `packages/shared/pkg/database/postgres.go`
  - `packages/shared/pkg/database/postgres_test.go`
- [x] Bọc `pg_advisory_lock` (Lock ID `8424119472649191`) trên dedicated connection cho toàn bộ quá trình kiểm tra & apply migration khi nhiều pod cùng khởi động.
- [x] PostgreSQL 17 runtime test với 5 connection đồng thời pass; migration effect và tracker row đều đúng một lần. Shared Go suite pass.

---

## 🧪 Gate D — Bằng Chứng Pilot (Verification Evidence)

> ⚠️ **Điều kiện tiên quyết:** Gate A, B, C đã hoàn tất.

### 🟡 D-01 — PostgreSQL Integration Suite
- **Priority:** P0 (evidence) | **Effort:** 5–8 ngày | **Status:** `PARTIAL — CORE POSTGRESQL RUNTIME VERIFIED`
- **Files:**
  - `services/auth/internal/repository/auth_integration_test.go`
  - `services/helpdesk/internal/repository/helpdesk_integration_test.go`
  - `services/audit/internal/repository/audit_integration_test.go`
  - `tests/integration/postgres_integration_test.go`
- [x] Đã viết test Auth, Helpdesk, Audit, migration concurrency và parameterized-query payload.
- [x] Runner có chế độ fail-closed: `scripts/staging_verify.ps1 -RequirePostgres` hoặc `REQUIRE_POSTGRES=1 ./scripts/staging_verify.sh`.
- [x] Chạy local trên 4 database PostgreSQL 17.11 tạm, DSN riêng cho Auth/Helpdesk/Audit/migration và không có test `SKIP`; database tạm đã được dọn.
- [ ] Tự động hóa cùng suite bằng PostgreSQL ephemeral trong CI và lưu run URL/log theo commit.
- [ ] Bổ sung coverage Notification, Reporting và date filter theo đúng acceptance plan.
- [ ] Lưu CI run URL/log gắn với commit được nghiệm thu.

---

### 🟡 D-02 — Real API E2E & Browser E2E Suite
- **Priority:** P0 (evidence) | **Effort:** 4–7 ngày | **Status:** `PARTIAL — LOCAL DEPLOYED-STACK E2E PASSED; BROWSER/STAGING OPEN`
- **Files:**
  - `tests/integration/http_contract_test.go`
  - `apps/web/tests/route_permissions.test.ts`
  - `tests/e2e/README.md`
- [x] Có HTTP contract harness bằng `httptest` cho lifecycle/status, header spoofing và anti-enumeration; đây là test double, không phải gateway thật.
- [x] Có Vitest route-policy contracts cho RBAC, safe redirect, optimistic-lock model và refresh mutex; không khởi động browser.
- [x] Stack Docker thật login qua Auth/Gateway, tạo/assign/comment/resolve, kiểm tra `409`, refresh rotation, logout/revoke và xác minh PostgreSQL Helpdesk/Audit/Reporting/Notification; evidence: `docs/evidence/gate-d/deployed_stack_e2e.json`.
- [x] Spoofed identity bị từ chối `401` và cross-department probe được che bằng `404` qua Gateway thật.
- [ ] Chạy lại cùng kịch bản trên staging/release và lưu evidence gắn với release version.
- [ ] Bổ sung Playwright cho login, route guard, create ticket, 409, refresh và logout; lưu report/video/trace theo chính sách CI.

---

### 🟡 D-03 — Staging Release Evidence & Handover
- **Priority:** P0 (evidence) | **Effort:** 3–5 ngày | **Status:** `PARTIAL — LOCAL IMAGE/LOAD/DR EVIDENCE PRESENT; STAGING OPEN`
- **Files:**
  - `deploy/k6/load_test.js`
  - `scripts/backup_restore.ps1` & `.sh`
  - `scripts/staging_verify.ps1` & `.sh`
  - `docs/GATE_D_VERIFICATION_EVIDENCE.md`
  - `docs/evidence/gate-d/dr_evidence_20260902.json`
- [x] Có kịch bản k6 với threshold `p95 < 200ms`, `p99 < 500ms`, error rate `< 1%`; credential thật là bắt buộc và 401 không còn được coi là thành công.
- [x] Có script backup/restore fail-closed cho 9 database và tạo `backups/dr_evidence.json` khi drill thật hoàn tất.
- [x] Có static precheck cho non-root, tag `:latest`, NetworkPolicy và PDB.
- [x] Build 12 application image, ghi local digest, xác minh user `10001:10001` và 20 container healthy; evidence: `docs/evidence/gate-d/image_manifest.json`.
- [ ] Lưu Trivy/Scout JSON chứng minh 0 High/Critical hoặc exception được duyệt; Docker Scout hiện yêu cầu Docker ID login.
- [ ] `helm lint`/render, deploy staging với TLS secret thật, chạy rollout/readiness/migration/smoke và network-isolation checks.
- [ ] Chạy k6 trên endpoint staging thật và lưu summary JSON.
- [x] Controlled load local 100 VU/30s: 2.928 request, 0 lỗi, p95 `30.3439ms`, p99 `65.0058ms`; evidence: `docs/evidence/gate-d/local_load_summary.json` (không thay thế k6 staging).
- [x] Backup/restore đủ 9 database trên PostgreSQL 17.11 pass; database restore 6.602s, evidence tại `docs/evidence/gate-d/dr_evidence_20260902.json`.
- [ ] Chứng minh WAL-based RPO và full-service RTO < 15 phút; database-only restore không thay thế tiêu chí này.
- [ ] Owner ký biên bản sau khi toàn bộ evidence trên gắn với cùng version/commit.

---

## 🎯 Pilot Exit Criteria

1. ✅ **Gate A:** 100% Passed.
2. 🟡 **Gate B:** Technical/runtime criteria đạt 100%; chờ owner acceptance để đóng chính thức.
3. 🟡 **Gate C:** 3/3 implementation; chờ C-02 staging TLS/private-network acceptance.
4. 🟡 **Gate D:** local technical readiness ước tính 90%, formal acceptance evidence 70%; 0/3 task đã được đóng chính thức.
5. 🔴 Vẫn còn P0 về browser/CI, CVE scan, TLS/network staging, WAL/full-service recovery và owner sign-off.
6. 📝 Product Owner chưa ký biên bản bàn giao phạm vi Pilot.

---

## 📦 Backlog Sau Pilot (Post-Pilot Prioritization)

1. SLA background scanner + escalation engine + leader election.
2. Business calendar & cấu hình giờ làm việc ngày lễ VN.
3. Transactional Outbox Pattern cho RabbitMQ event publishing.
4. Ticket Attachment qua MinIO + Presigned URL an toàn.
5. Account lockout & risk-based rate limiting.
6. Vector Database RAG Access Control.
7. Đa ngôn ngữ tiếng Việt (i18n).
