# EOMP — Kế hoạch remediation đã hiệu chỉnh

**Ngày đối chiếu:** 2026-08-31  
**Baseline:** commit `997a3d3` cộng với working tree hiện tại  
**Nguồn đầu vào:** `EOMP_AUDIT_REPORT.md`, bản kế hoạch 29 task do người dùng cung cấp, mã nguồn và tài liệu trong repo  
**Quyết định readiness:** **chưa được phép production; chưa đủ bằng chứng để pilot với dữ liệu thật**

## 1. Kết luận ngắn

Bản audit và cách ưu tiên ban đầu đúng hướng: đặt identity, authorization, dữ liệu thật và kiểm thử trước hạ tầng mở rộng. Tuy nhiên kế hoạch hiện tại không nên dùng nguyên trạng vì:

1. Audit được lập trên một commit cũ; một số hạng mục đã được làm toàn bộ hoặc một phần.
2. Task quá lớn nhưng chỉ có một trạng thái `done/not done`, nên dễ đánh dấu hoàn thành khi acceptance criteria quan trọng vẫn thiếu.
3. Kiểm thử bảo mật bị dồn tới Sprint 4, trong khi phải đi cùng từng bản sửa P0.
4. Câu “production ready sau Sprint 6” không có release evidence tương ứng.
5. Một số nhận định kỹ thuật đã không còn đúng hoặc chưa đủ, đặc biệt ở reporting, migration và E2E.

Mục tiêu hợp lý hơn là **pilot có kiểm soát sau khi qua Gate A–D**. Production chỉ được quyết định bằng evidence, không quyết định theo số sprint.

## 2. Những gì đã kiểm chứng trên baseline hiện tại

- Repo có 332 file được theo dõi bởi `rg`: 152 Go, 28 SQL, 17 Vue và 39 Markdown.
- Có 11 Go service, một shared module, một module `tests/e2e` và một Nuxt app.
- `go test ./...` đã pass ở cả 13 Go module.
- Frontend đã pass 3 Vitest, `nuxt typecheck`, ESLint và production build.
- OpenAPI gate pass `102/102` runtime operations.
- Các check trên **không** chứng minh runtime integration: không có test PostgreSQL thật, không có stack E2E thật, không có browser journey thật và chưa có deployment evidence trong lần kiểm tra này.

## 3. Điều chỉnh quan trọng so với kế hoạch cũ

### 3.1 Hạng mục đã xong hoặc gần xong

- SQL injection cụ thể trong notification repository đã được parameterize.
- Ticket, problem, workflow instance và change number đã dùng PostgreSQL sequence.
- Audit đã có HMAC chain và endpoint kiểm tra integrity.
- CI đã fail-closed khi thiếu gosec, govulncheck hoặc Trivy.
- OpenAPI đã đạt operation parity 102/102; domain schema và contract conformance vẫn còn thiếu.
- Frontend dashboard/reporting đã bỏ phần lớn array KPI giả; ticket mutation đã gửi `version`.

Các hạng mục trên vẫn cần integration/regression evidence trước khi đóng hoàn toàn.

### 3.2 Hạng mục P0 vẫn còn nguyên

- Gateway chưa strip identity headers.
- Prefix `/api/v1/auth/` vẫn public ở gateway; auth service vẫn tin `X-User-*` và có nhánh bỏ qua Bearer.
- `X-User-Department` chỉ được set khi claim không rỗng.
- CORS vẫn cho client gửi identity headers.
- Auth và gateway vẫn có JWT secret public làm fallback; chín service DB vẫn có password development mặc định.
- Có 35 chỗ tạo HTTP 500 bằng `InternalServerError(fmt.Sprintf(...err))`; error writer chưa có correlation ID.
- Chưa có ma trận quyền được duyệt và chưa có API quản trị user.

### 3.3 Phát hiện mới cần thêm vào backlog

1. **Reporting chưa có nguồn dữ liệu thật.** `SLAAggregator` chỉ roll up `raw_incident_records` trong `reporting_db`, nhưng ngoài migration không có code nào ghi bảng này. Sau cleanup migration, report có thể trả 0 mãi mãi dù helpdesk có ticket.
2. **Bộ lọc thời gian của reporting không được áp dụng.** Repository nhận `DateFilterQuery` nhưng các SQL query không dùng range/start/end; chỉ nhãn kỳ báo cáo thay đổi.
3. **PDF vẫn hardcode KPI** `31.8/7.2/97.4/4.86`, tự ghi byte PDF, không escape dữ liệu và chỉ hiện 25 dòng.
4. **Frontend truyền query param không nhất quán.** `useApi.get()` đã tự bọc `params`, nhưng một số trang lại truyền `{ params: {...} }`, làm filter/phân trang có thể bị serialize sai.
5. **Refresh rotation chưa atomic.** Auth lưu refresh token mới trước rồi mới revoke token cũ, không nằm trong cùng transaction.
6. **Tài liệu tự mâu thuẫn nghiêm trọng.** `IMPLEMENTATION_STATUS.md` nói chưa production, nhưng nhiều tài liệu khác vẫn ghi “100% production certified/approved”. Đây là rủi ro bàn giao, không phải cleanup P3.

## 4. Trạng thái backlog cũ

Quy ước: `DONE` = code và evidence phù hợp; `PARTIAL` = có code nhưng thiếu scope hoặc verification; `OPEN` = chưa có implementation đáng kể.

| Task cũ | Trạng thái | Nhận định hiện tại |
|---|---|---|
| BA-01 | OPEN | Chưa có ma trận 4 role × resource × action × scope được duyệt |
| SEC-01 | OPEN | Lỗ hổng header spoofing vẫn hiện diện |
| SEC-02 | OPEN | Default JWT/DB secret vẫn còn; task cũ nói “11 service dùng JWT” là sai |
| BE-01 | PARTIAL | SQLi đã sửa; thiếu PostgreSQL regression test và security rule có bằng chứng |
| BE-02 | PARTIAL | Employee bị giới hạn theo requester ở handler; agent/manager/department/query-level scope chưa có |
| BE-03 | DONE | Bốn loại mã đã dùng sequence; `%04d` là minimum width, không hỏng sau 9999 |
| BE-04 | OPEN | 35 điểm leak lỗi nội bộ; chưa có error ID trong JSON |
| DB-01 | PARTIAL | Có exact cleanup migrations nhưng demo vẫn nằm trong migration đầu; chưa có `seeds/`/`make seed` |
| BE-05 | OPEN | Chưa có users API, change password, admin reset hoặc revoke-all-session |
| FE-01 | PARTIAL | Đã nối nhiều API; còn nested params, role-unaware calls và route prerender cần quyết định theo auth design |
| RPT-01 | PARTIAL | Repository bỏ mock overview; PDF và data pipeline/filter vẫn sai |
| FE-02 | PARTIAL | Working tree đã lọc menu theo role; middleware chưa global, chưa refresh mutex/BFF/HttpOnly |
| BE-06 | OPEN | Không có SLA scanner, escalation, deduplication hoặc leader lock |
| QA-01 | OPEN | Không có testcontainers/PostgreSQL integration test |
| QA-02 | OPEN | `tests/e2e` là unit/simulation bằng struct/httptest, không phải E2E thật |
| DEVOPS-01 | OPEN | Nginx vẫn HTTP, public Grafana/Prometheus, CSP có unsafe-inline/unsafe-eval |
| DEVOPS-02 | PARTIAL | CI gate đã sửa; migration advisory lock và coverage policy chưa có |
| FE-03 | OPEN | Chỉ có hai component; 12 page dài hơn 500 dòng, lớn nhất 1.293 dòng |
| BE-07 | OPEN | Publish/timeline vẫn nuốt lỗi; chưa có outbox |
| BE-08 | PARTIAL | Có JWT ID, token type và refresh issuer; chưa lockout, access issuer check hay access-token revocation |
| BE-09 | OPEN | Chưa có business calendar/pause semantics |
| BE-10 | OPEN | MinIO chưa được helpdesk sử dụng |
| CLEAN-01 | PARTIAL | Audit mock đã bỏ; còn 62 `.gitkeep`, alias AI, path fallback, `deployment/` và proto không dùng |
| Audit integrity API | DONE | Endpoint `/api/v1/audit/integrity` đã có |
| OpenAPI parity | PARTIAL | Operation parity done; domain schemas/contract tests chưa done |
| Knowledge ACL | OPEN | Chưa có visibility/scope ở SQL và RAG |
| AI quota/audit | OPEN | Chưa có quota, input budget hoặc usage audit |
| CSAT | OPEN | Chưa có nguồn thu thập; PDF vẫn hiển thị số giả |

## 5. Backlog mới theo release gate

### Gate A — Chốt luật và khóa bề mặt tấn công

Không bắt đầu pilot nếu bất kỳ task nào trong gate này chưa đạt.

#### A-01 — Ma trận authorization và quyết định sản phẩm

**Priority:** P0  
**Effort:** 2–3 ngày làm việc có BA/owner tham gia

Chốt bằng văn bản:

- Bốn role trên từng resource và action.
- Scope `own`, `assigned`, `department`, `all`.
- Employee có được xem ticket cùng phòng ban hay chỉ ticket mình tạo.
- Agent có được xem mọi queue hay chỉ queue/department được gán.
- Chính sách self-register, admin-created, domain allowlist và mapping phòng ban.
- Quyền đóng ticket, SLA pause, lịch làm việc, CSAT và ticket–CI.

**Exit criteria:** không còn `TBD`; có owner phê duyệt và có test matrix được sinh trực tiếp từ ma trận.

#### A-02 — Sửa identity trust boundary

**Priority:** P0  
**Effort:** 2–3 ngày

- Strip tất cả biến thể `X-User-*` ở outermost gateway middleware.
- Tách route auth public và protected; `/me` và `/login-history` bắt buộc access token.
- Ghi đè department/name kể cả giá trị rỗng.
- Xóa identity headers khỏi CORS allowlist.
- Bỏ nhánh auth service chấp nhận header client để skip Bearer.
- Xác minh service port không expose ra untrusted network bằng Compose/K8s/NetworkPolicy test.
- Tạo follow-up P1 cho signed internal identity hoặc downstream JWT validation; không coi header trần là trust boundary dài hạn.

**Exit criteria:** bốn case spoofing trong audit trả 401/403 đúng; có automated gateway test và deployed-network test.

#### A-03 — Secret policy và rotation

**Priority:** P0  
**Effort code:** 1–2 ngày; rotation là external operation

- Auth và gateway phải nhận JWT secret bắt buộc, tối thiểu 32 byte entropy phù hợp.
- Chín DB-backed service phải nhận password bắt buộc trong non-test runtime.
- Test phải tự inject secret riêng, không dùng secret public của repo.
- `.env.example` chỉ ghi placeholder/hướng dẫn tạo, không chứa credential dùng được.
- Rotate JWT, PostgreSQL, RabbitMQ, MinIO, Grafana và provider key ở mọi environment đã từng dùng.

**Exit criteria:** missing/weak/known secret fail ở startup; có rotation record từ secret owner. Grep sạch không thay thế rotation evidence.

#### A-04 — Sanitize lỗi và correlation ID

**Priority:** P0  
**Effort:** 2–3 ngày

- Handler/service log lỗi gốc ở server; response 5xx chỉ có thông điệp chung.
- Đưa request ID vào context và response error envelope.
- Loại toàn bộ `InternalServerError(fmt.Sprintf(...err))`.
- Thêm test bảo đảm tên table/column/driver không lọt ra client.

**Exit criteria:** 0 pattern leak; log và response có cùng request ID; fault-injection tests pass.

#### A-05 — Reconcile tài liệu và claims

**Priority:** P0 cho governance/release  
**Effort:** 1–2 ngày

- `IMPLEMENTATION_STATUS.md` là nguồn trạng thái duy nhất.
- Gắn nhãn historical/obsolete hoặc sửa các tài liệu đang nói “100% production certified”.
- Đổi tên `tests/e2e` hoặc ghi rõ đây là simulation/unit test.
- CI message không được nói E2E pass khi chỉ chạy unit simulation.

**Exit criteria:** không còn tài liệu active mâu thuẫn với readiness decision.

### Gate B — Ranh giới dữ liệu và tính trung thực

#### B-01 — Authorization theo bản ghi

**Priority:** P0  
**Effort:** 5–8 ngày

- Tạo kiểu `Actor`/`AccessScope` dùng nhất quán thay vì truyền các string rời.
- Áp scope trong SQL cho list/get; không tải toàn bộ rồi lọc.
- Helpdesk: list/get/comments/timeline/create-on-behalf/asset-linked tickets.
- Workflow: instance/log/approval.
- Asset/employee/knowledge theo ma trận A-01.
- Ngoài scope trả 404 khi cần chống enumeration.

**Exit criteria:** integration test PostgreSQL cho từng role, own/other-department/unassigned/assigned; không chỉ test handler mock.

#### B-02 — User lifecycle tối thiểu cho pilot

**Priority:** P0  
**Effort:** 5–7 ngày backend + UI tối thiểu

Tách scope cũ BE-05 thành ba phần:

1. Admin list/create/update/deactivate user, role và department.
2. User change password + admin reset; mọi refresh token của user bị revoke atomically.
3. Forgot/reset qua email là P1 riêng nếu pilot có helpdesk-assisted reset.

Public register phải theo quyết định A-01 và không được tự chọn department.

**Exit criteria:** tạo được agent không chạm DB; self-promotion bị chặn; deactivate/password change vô hiệu mọi refresh session; có audit event đáng tin cậy.

#### B-03 — Reporting read model và filter thật

**Priority:** P0 nếu Reports xuất hiện trong pilot; nếu chưa làm thì ẩn module  
**Effort:** 5–8 ngày

- Chọn một nguồn sự thật: event-driven projection có idempotency hoặc truy vấn Helpdesk API; không tiếp tục roll up một bảng không có producer.
- Áp dụng `range/start_date/end_date` vào tất cả query.
- PDF nhận overview thật, escape dữ liệu và dùng thư viện PDF chuẩn.
- Bỏ CSAT cho tới khi có survey.
- Xác định semantics cho SLA denominator và trạng thái chưa resolved.

**Exit criteria:** tạo/resolve ticket thật làm report thay đổi; các khoảng ngày trả kết quả khác nhau đúng fixture; DB rỗng trả 0; PDF không chứa hằng KPI.

#### B-04 — Frontend API contract và role-aware data loading

**Priority:** P1  
**Effort:** 3–5 ngày

- Chuẩn hóa `useApi.get(url, params)` và bỏ mọi `{ params: { ... } }` lồng sai.
- Dashboard chỉ gọi endpoint role đó có quyền; 403 không bị hiểu thành “0 dữ liệu”.
- Mọi page có loading/empty/error/unavailable state riêng.
- Quyết định giữ/bỏ prerender dựa trên BFF/SSR auth; prerender shell tự nó không phải dữ liệu giả.

**Exit criteria:** contract tests kiểm tra query string; DB empty, backend down và 403 hiển thị ba trạng thái khác nhau.

#### B-05 — Migration và seed sạch

**Priority:** P1  
**Effort:** 2–4 ngày

- Baseline mới không insert demo rồi delete ở migration sau.
- Tách reference data bắt buộc khỏi demo operational data.
- Seed command chỉ chạy explicit trong development/test và idempotent.
- Giữ cleanup migration cũ để upgrade, nhưng test không xóa nhầm dữ liệu production có cùng thuộc tính.

**Exit criteria:** clean production migration tạo schema/reference data đúng và 0 operational record; dev seed có thể chạy hai lần.

### Gate C — Session, transport và migration safety

#### C-01 — Frontend session architecture

**Priority:** P1 release blocker  
**Effort:** 4–6 ngày

- Dùng BFF/server route hoặc auth server set HttpOnly Secure SameSite cookie.
- Thêm CSRF protection nếu cookie được gửi tự động.
- Access token không persist trong JS-readable cookie.
- Refresh mutex, single retry và loop protection.
- Refresh rotation/revoke trong một DB transaction; phát hiện reuse nếu scope cho phép.
- Route middleware global và page-level role guard chỉ là UX defense; backend vẫn là authority.

**Exit criteria:** `document.cookie` không đọc được refresh token; 5 request 401 chỉ refresh một lần; token rotation rollback đúng khi DB fault.

#### C-02 — TLS và private monitoring

**Priority:** P0 trước khi truyền credential ngoài localhost  
**Effort:** 2–4 ngày

- HTTPS redirect, HSTS sau khi TLS ổn định.
- Không public Prometheus/Grafana; dùng private ingress/VPN/SSO.
- Bỏ `unsafe-eval`; xây CSP tương thích Nuxt bằng nonce/hash.
- Rotate Grafana credential.

**Exit criteria:** test từ untrusted network không vào monitoring; TLS scanner đạt policy tổ chức; auth cookie chỉ đi qua HTTPS.

#### C-03 — Migration concurrency

**Priority:** P1  
**Effort:** 1–2 ngày

- Advisory lock bao toàn bộ discovery/check/apply migration, hoặc chỉ một migration Job/initContainer có quyền DDL.
- Test năm process/pod khởi động đồng thời trên DB mới và DB upgrade.

**Exit criteria:** mỗi migration applied đúng một lần, không duplicate tracker row, không crash loop.

### Gate D — Bằng chứng pilot

#### D-01 — PostgreSQL integration suite

**Priority:** P0 evidence  
**Effort:** 5–8 ngày để dựng nền và phủ critical paths

- Testcontainers hoặc ephemeral PostgreSQL trong CI.
- Ưu tiên auth, helpdesk, notification, reporting, audit và migration.
- Bao phủ authorization, SQLi payload, sequence concurrency, optimistic lock, refresh rotation, transaction rollback và date filter.
- Coverage chỉ là chỉ báo; release gate dựa trên scenario rủi ro, không dựa riêng vào con số 60%.

#### D-02 — API E2E và browser E2E thật

**Priority:** P0 evidence  
**Effort:** 4–7 ngày

- Đổi `tests/e2e` hiện tại thành `tests/simulation` hoặc `tests/unit`.
- API E2E phải khởi động stack, login HTTP, tạo/assign/comment/resolve ticket và kiểm tra DB/event/audit.
- Playwright kiểm tra login, route guard, create ticket, conflict 409, session refresh và logout.
- Case spoofing/cross-tenant phải chạy trên ingress/gateway thật.

#### D-03 — Staging release evidence

**Priority:** P0 evidence  
**Effort:** 3–5 ngày sau khi CI sẵn sàng

- Build và scan đủ image.
- Render Helm, deploy staging, chạy smoke/readiness/migration tests.
- Chạy k6 trên endpoint thật; lưu cấu hình, dataset và kết quả.
- Chạy backup/restore chín DB và đo RPO/RTO thật.

**Pilot exit:** Gate A–D pass, không còn P0, và owner ký phạm vi pilot.  
**Production exit:** thêm secret rotation, DR evidence, on-call/monitoring evidence và production owner sign-off.

## 6. Backlog sau pilot

Thứ tự sau pilot phụ thuộc A-01 và SLA/product promise:

1. SLA scanner + escalation + dedup + leader election.
2. Business calendar/pause. Nếu chưa làm, không quảng bá SLA compliance theo business hours.
3. Transactional outbox cho helpdesk/asset/workflow; cần trước khi notification/audit được coi là reliable.
4. Ticket attachment với presigned URL, MIME/content validation, malware scanning policy và authorization kế thừa ticket.
5. Account lockout/risk-based throttling và access-token revocation policy.
6. Knowledge ACL áp dụng đồng thời ở PostgreSQL và vector retrieval.
7. AI quota, input/context limit, prompt-injection boundary và usage audit.
8. Frontend component extraction theo feature, không đặt mục tiêu LOC giảm 30% như release criterion.
9. CSAT survey, i18n và các cải tiến UX khác.
10. Dead-code cleanup sau khi test bảo vệ route behavior.

Nếu SLA, attachment hoặc report là điều kiện bắt buộc của pilot, kéo task tương ứng lên Gate B/C và lùi ngày pilot; không giữ cùng timeline rồi bỏ acceptance criteria.

## 7. Timeline thực tế

Với một full-stack developer kiêm DevOps, backlog còn lại vượt quá 60 person-days trước buffer. Kế hoạch 12 tuần cũ vừa chứa implementation vừa chứa security, DBA, QA, browser E2E, staging, DR và production sign-off nên quá chặt.

Ước lượng an toàn hơn sau khi A-01 chốt scope:

- **Pilot có kiểm soát:** khoảng 8–10 tuần cho một người, nếu report/SLA/attachment có thể ẩn khỏi pilot.
- **Production candidate có evidence:** khoảng 14–18 tuần cho một người.
- Có thể rút ngắn khi Backend, Frontend/QA và DevOps làm song song, nhưng release gates không thay đổi.

Không cam kết ngày production trước khi Gate A hoàn tất và D-01 chạy được trong CI.

## 8. Quy tắc cập nhật task

Mỗi task chỉ được đóng khi có đủ bốn trường:

1. **Code evidence:** commit/file/migration cụ thể.
2. **Automated evidence:** tên test và pipeline run.
3. **Runtime evidence:** môi trường, thời điểm, version và kết quả đối với việc không thể chứng minh bằng unit test.
4. **Owner acceptance:** bắt buộc cho authorization, security, DR và production handover.

Không dùng “code tồn tại”, “test pass 100%” hoặc “có Helm/K8s” làm đồng nghĩa với production-ready.
