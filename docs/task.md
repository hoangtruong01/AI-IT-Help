# 📋 EOMP — Task Tracker & Báo Cáo Nâng Cấp (Gate-based)

> **Baseline Commit:** `997a3d3` | **Cập nhật:** 2026-08-31  
> **Trạng thái:** Hoàn tất **Gate A** (Commit `c1ae962`) & **Gate B** (Commit `c6a9b59`)  
> **Mục tiêu:** Pilot có kiểm soát → chỉ khi Gate A–D pass 100%  
> **Quy tắc:** Task chỉ đóng khi có đủ 4 bằng chứng: `Code evidence` + `Automated test` + `Runtime evidence` + `Owner acceptance`.

---

## 📊 Bảng Tổng Hợp Tiến Độ Các Gate

| Gate | Mô tả phạm vi | Trọng số | Trạng thái | Số task |
|---|---|:---:|:---:|:---:|
| **Gate A** | **Chốt luật và khóa bề mặt tấn công** (Trust boundary, secrets, errors, matrix) | P0 | ✅ **DONE** (100%) | 5/5 |
| **Gate B** | **Ranh giới dữ liệu & tính trung thực** (Row-level auth, user lifecycle, real reports) | P0 | ✅ **DONE** (100%) | 5/5 |
| **Gate C** | **Session, transport và migration safety** (Cookie BFF, refresh mutex, TLS, locks) | P1 | 🟡 **READY** | 0/3 |
| **Gate D** | **Bằng chứng pilot** (Postgres integration suite, real E2E, staging verification) | P0 | ⚪ **OPEN** | 0/3 |

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

### ✅ B-01 — Authorization theo bản ghi (Row-Level / Scope-Based)
- **Priority:** P0 | **Effort:** 5–8 ngày | **Status:** `DONE`
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
  - [x] `GetTicket`: Kiểm tra qua `actorCanAccessTicket`, trả `404 Not Found` (thay vì 403) khi ngoài scope (chống enumeration).
  - [x] `AddComment`, `ListComments`, `ListTimeline`: Kế thừa xác thực quyền từ ticket.
  - [x] `CreateTicket`: Employee luôn bị gán `requester_id = actor.ID`; Operators được tạo on-behalf.
- [x] **Workflow Scope:** Instance visibility theo requester (`ROLE_EMPLOYEE` chỉ xem instance mình tạo, 404 cho truy cập trái phép).

---

### ✅ B-02 — User lifecycle tối thiểu cho pilot
- **Priority:** P0 | **Effort:** 5–7 ngày | **Status:** `DONE`
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
- [x] **Automated Tests:** 7/7 tests pass (**PASS 100%**).

---

### ✅ B-03 — Reporting read model và filter thật
- **Priority:** P0 | **Effort:** 5–8 ngày | **Status:** `DONE`
- **Files:**
  - [`services/reporting/internal/repository/reporting_repository.go`](file:///d:/IT_help/eomp/services/reporting/internal/repository/reporting_repository.go)
  - [`services/reporting/internal/service/reporting_service.go`](file:///d:/IT_help/eomp/services/reporting/internal/service/reporting_service.go)
- [x] Áp dụng `range/start_date/end_date` vào SQL queries (`GetExecutiveOverview`, `GetIncidentTrends`).
- [x] **PDF Export Enhancement:**
  - [x] Tính toán chỉ số KPI động từ danh sách bản ghi thực tế.
  - [x] Escape ký tự đặc biệt PDF `(`, `)`, `\` chống vỡ layout.
  - [x] Xoá bỏ CSAT giả lập và các hằng số KPI hardcode `31.8 / 4.86`.

---

### ✅ B-04 — Frontend API contract & chuẩn hóa tham số
- **Priority:** P1 | **Effort:** 3–5 ngày | **Status:** `DONE`
- **Files:**
  - [`apps/web/app/composables/useApi.ts`](file:///d:/IT_help/eomp/apps/web/app/composables/useApi.ts)
- [x] Chuẩn hóa `useApi.get(url, params)` — Tự động unwrap `{ params: { ... } }` nếu caller truyền lồng nhau, bảo đảm query string luôn chuẩn xác (`?range=30d`).

---

### ✅ B-05 — Migration và seed sạch
- **Priority:** P1 | **Effort:** 2–4 ngày | **Status:** `DONE`
- **Files:**
  - [`services/*/migrations/*.sql`](file:///d:/IT_help/eomp/services/)
- [x] Rà soát và xác nhận baseline migrations không chứa demo credentials.
- [x] Tài khoản demo mặc định đã được vô hiệu hóa qua migration `003_disable_insecure_demo_accounts.sql`.

---

## ⚡ Gate C — Session, Transport & Migration Safety

> ⚠️ **Điều kiện tiên quyết:** Gate A & B đã hoàn tất.

### 🟡 C-01 — Frontend session architecture & Cookie BFF
- **Priority:** P1 (release blocker) | **Effort:** 4–6 ngày | **Status:** `PARTIAL`
- **Dependency:** A-03
- **Files:**
  - `apps/web/app/middleware/auth.global.ts`
  - `apps/web/app/composables/useApi.ts`
  - `apps/web/app/stores/auth.ts`
  - `apps/web/app/layouts/default.vue`
- [ ] Cookie strategy:
  - [ ] BFF / Server Route set HttpOnly Secure SameSite cookie cho refresh token.
  - [ ] Access token lưu trong memory (Pinia store), không persist ra JS-readable storage.
- [ ] Refresh Mutex:
  - [ ] 5 requests 401 đồng thời chỉ trigger đúng 1 lần gọi refresh token.
  - [ ] Queue các pending requests trong thời gian chờ token mới.
- [ ] Route Guard:
  - [ ] Đổi middleware thành `auth.global.ts`.
  - [ ] Chặn truy cập trái quyền ở giao diện client (`/audit`, `/reports`, `/monitoring`).

---

### ⚪ C-02 — TLS & Private Monitoring
- **Priority:** P0 | **Effort:** 2–4 ngày | **Status:** `OPEN`
- **Files:**
  - `deploy/nginx/conf.d/eomp.conf`
  - `deploy/kubernetes/manifests/08-ingress.yaml`
- [ ] HTTPS redirect (80 → 443) & HSTS header.
- [ ] Gỡ `/monitoring/grafana/` và `/monitoring/prometheus/` khỏi Ingress/Nginx công khai.
- [ ] Cấu hình CSP: Bỏ `unsafe-eval`, tương thích Nuxt SSR.

---

### ⚪ C-03 — Migration Concurrency Control
- **Priority:** P1 | **Effort:** 1–2 ngày | **Status:** `OPEN`
- **Files:**
  - `packages/shared/pkg/database/postgres.go`
- [ ] Bọc `pg_advisory_lock` cho toàn bộ quá trình kiểm tra & apply migration khi nhiều pod cùng khởi động.
- [ ] Đảm bảo mỗi migration chỉ được thực thi đúng 1 lần duy nhất mà không bị crash loop.

---

## 🧪 Gate D — Bằng Chứng Pilot (Verification Evidence)

> ⚠️ **Điều kiện tiên quyết:** Gate A, B, C đã hoàn tất.

### ⚪ D-01 — PostgreSQL Integration Suite
- **Priority:** P0 (evidence) | **Effort:** 5–8 ngày | **Status:** `OPEN`
- **Files:** `tests/integration/`, `Jenkinsfile`
- [ ] Khởi chạy PostgreSQL test container qua testcontainers-go.
- [ ] Kiểm thử tích hợp thực tế:
  - [ ] Auth: Xoay vòng token atomic, đổi pass thu hồi session, CRUD user.
  - [ ] Helpdesk: Phân quyền bản ghi theo 4 role, concurrency sequence (100 goroutines), optimistic locking.
  - [ ] Audit: Kiểm tra tính toàn vẹn chuỗi HMAC append-only.

---

### ⚪ D-02 — Real API E2E & Browser E2E Suite
- **Priority:** P0 (evidence) | **Effort:** 4–7 ngày | **Status:** `OPEN`
- **Files:** `tests/e2e/`, `tests/browser/` (Playwright)
- [ ] Kịch bản E2E toàn diện trên Gateway thật:
  - [ ] Login → Tạo Ticket → Gán Kỹ thuật viên → Thêm Bình luận → Giải quyết.
  - [ ] Giả lập tấn công Header Spoofing → Bị từ chối 401/403.
  - [ ] Truy cập chéo dữ liệu người dùng khác → Trả về 404.

---

### ⚪ D-03 — Staging Release Evidence & Handover
- **Priority:** P0 (evidence) | **Effort:** 3–5 ngày | **Status:** `OPEN`
- [ ] Quét lỗ hổng container image (Trivy Scan: 0 High/Critical).
- [ ] Triển khai Staging qua Helm Chart.
- [ ] Kiểm thử tải k6 & Báo cáo kết quả phục hồi dữ liệu Backup/Restore.

---

## 🎯 Pilot Exit Criteria

1. ✅ **Gate A:** 100% Passed.
2. ✅ **Gate B:** 100% Passed.
3. ⏳ **Gate C:** 100% Passed.
4. ⏳ **Gate D:** 100% Passed.
5. Không còn bất kỳ issue P0 nào mở.
6. Product Owner ký biên bản bàn giao phạm vi Pilot.

---

## 📦 Backlog Sau Pilot (Post-Pilot Prioritization)

1. SLA background scanner + escalation engine + leader election.
2. Business calendar & cấu hình giờ làm việc ngày lễ VN.
3. Transactional Outbox Pattern cho RabbitMQ event publishing.
4. Ticket Attachment qua MinIO + Presigned URL an toàn.
5. Account lockout & risk-based rate limiting.
6. Vector Database RAG Access Control.
7. Đa ngôn ngữ tiếng Việt (i18n).
