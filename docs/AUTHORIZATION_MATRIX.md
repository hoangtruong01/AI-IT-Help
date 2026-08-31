# EOMP — Ma trận Authorization (Authorization Matrix)

> **Trạng thái:** READY FOR OWNER APPROVAL — nội dung không còn quyết định mở; chưa có hiệu lực cho đến khi owner ký
> **Cập nhật:** 2026-08-31
> **Tham chiếu:** Gate A-01 trong `EOMP_REMEDIATION_PLAN_REVISED.md`
> **Revision:** 1.0
> **Owner phê duyệt:** Chưa ký
> **Ngày phê duyệt:** Chưa có

## 1. Bốn Role của hệ thống

| Role | Mô tả | Scope mặc định |
|------|--------|----------------|
| `ROLE_EMPLOYEE` | Nhân viên — người báo sự cố, tạo yêu cầu | Own (chỉ dữ liệu mình tạo/liên quan) |
| `ROLE_AGENT` | Nhân viên IT — xử lý ticket, quản lý tài sản | Assigned + Queue (ticket được giao + chưa assign) |
| `ROLE_MANAGER` | Trưởng phòng IT — giám sát phòng ban | Department (toàn bộ phòng ban mình quản lý) |
| `ROLE_ADMIN` | Quản trị hệ thống — toàn quyền | All (toàn bộ hệ thống) |

## 2. Scope definitions

| Scope | Ý nghĩa |
|-------|---------|
| `own` | Chỉ bản ghi mà `requester_id = actor_id` hoặc `created_by = actor_id` |
| `assigned` | Bản ghi mà `assignee_id = actor_id` |
| `queue` | Bản ghi chưa assign (`assignee_id IS NULL`) |
| `department` | Bản ghi thuộc `department_id = actor.department_id` |
| `all` | Toàn bộ bản ghi, không giới hạn |
| `—` | Không có quyền |

## 3. Ma trận quyền chi tiết

### 3.1 Ticket (Helpdesk)

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List** | `own` | `assigned` + `queue` | `department` | `all` |
| **Get** | `own` | `assigned` + `queue` | `department` | `all` |
| **Create** | ✅ (requester = self) | ✅ (on-behalf allowed) | ✅ (on-behalf allowed) | ✅ (on-behalf allowed) |
| **Update status** | — | `assigned` | `department` | `all` |
| **Assign** | — | `queue` (self-assign) | `department` | `all` |
| **Add comment** | `own` | `assigned` + `queue` | `department` | `all` |
| **View timeline** | `own` | `assigned` + `queue` | `department` | `all` |
| **View comments** | `own` | `assigned` + `queue` | `department` | `all` |
| **Delete** | — | — | — | `all` |

> **Quyết định:** Employee chỉ được xem ticket do chính mình tạo; không được xem ticket của đồng nghiệp cùng phòng ban. Truy vấn ngoài scope phải trả `404` để tránh lộ sự tồn tại của tài nguyên.
>
> **Quyết định queue:** Agent được xem và self-assign mọi ticket chưa được gán trong queue chung. Khi ticket đã được gán cho agent khác, agent không còn quyền đọc/cập nhật ticket đó; Manager/Admin vẫn theo scope trong bảng.

### 3.2 Problem Management

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List** | — | `all` | `all` | `all` |
| **Get** | — | `all` | `all` | `all` |
| **Create** | — | ✅ | ✅ | ✅ |
| **Update** | — | `own` (created_by) | `department` | `all` |
| **Delete** | — | — | — | `all` |

### 3.3 Asset & CMDB

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List** | — | `all` (read-only) | `all` | `all` |
| **Get** | — | `all` | `all` | `all` |
| **Create** | — | ✅ | ✅ | ✅ |
| **Update** | — | ✅ | ✅ | ✅ |
| **Assign/Revoke** | — | ✅ | ✅ | ✅ |
| **Delete** | — | — | — | `all` |

### 3.4 Employee & Department

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List employees** | `department` (read-only, name+email) | `all` (read-only) | `all` | `all` |
| **Get employee** | `own` profile | `all` | `all` | `all` |
| **Create** | — | — | ✅ (department) | ✅ |
| **Update** | — | — | `department` | `all` |
| **Delete** | — | — | — | `all` |
| **List departments** | ✅ (read-only) | ✅ | ✅ | ✅ |
| **Create department** | — | — | — | ✅ |

### 3.5 Workflow & Approvals & Changes

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List instances** | `own` (requester) | `own` + `assigned` | `department` | `all` |
| **Create instance** | ✅ | ✅ | ✅ | ✅ |
| **Get instance** | `own` | `own` + `assigned` | `department` | `all` |
| **Approve/Reject** | — | — | `assigned` (approver) | `all` |
| **View definitions** | ✅ (read-only) | ✅ | ✅ | ✅ |
| **Manage definitions** | — | — | ✅ | ✅ |
| **List changes (CAB)** | — | — | ✅ | ✅ |
| **Manage changes** | — | — | ✅ | ✅ |

### 3.6 Knowledge Base

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List/Search** | ✅ (public articles only) | ✅ (public + internal) | ✅ (all) | ✅ (all) |
| **Get article** | ✅ (public) | ✅ (public + internal) | ✅ (all) | ✅ (all) |
| **Create** | — | ✅ | ✅ | ✅ |
| **Update** | — | `own` (author) | `department` | `all` |
| **Delete** | — | — | — | `all` |

### 3.7 Notification

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List own** | ✅ | ✅ | ✅ | ✅ |
| **Mark as read** | `own` | `own` | `own` | ✅ |
| **Create (system)** | — | — | — | ✅ (broadcast) |

### 3.8 Audit Trail

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List** | — | — | ✅ | ✅ |
| **Get** | — | — | ✅ | ✅ |
| **Verify integrity** | — | — | ✅ | ✅ |

### 3.9 Reports & BI

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **View dashboard** | — | — | `department` | `all` |
| **Export PDF/CSV** | — | — | `department` | `all` |
| **View stats** | — | — | ✅ | ✅ |

### 3.10 AI Copilot

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **Chat** | ✅ | ✅ | ✅ | ✅ |
| **Triage ticket** | — | ✅ | ✅ | ✅ |

### 3.11 User Management (mới — BE-05)

| Action | EMPLOYEE | AGENT | MANAGER | ADMIN |
|--------|----------|-------|---------|-------|
| **List users** | — | — | `department` | `all` |
| **Create user** | — | — | — | ✅ |
| **Update user (role/dept)** | — | — | — | ✅ |
| **Deactivate user** | — | — | — | ✅ |
| **Change own password** | ✅ | ✅ | ✅ | ✅ |
| **Reset other's password** | — | — | — | ✅ |

## 4. Quyết định sản phẩm

### 4.1 Chính sách đăng ký
- **Quyết định:** Admin-created. Public registration bị **tắt** trong production.
- **Lý do:** Đây là hệ thống vận hành nội bộ, không ai ngoài tổ chức được tạo tài khoản.
- **Development:** Cho phép register để tiện test, nhưng không gán department.
- **Domain allowlist:** Không dùng allowlist như một lớp authorization trong production vì production không self-register. Admin chỉ được tạo email thuộc danh sách domain doanh nghiệp cấu hình; backend phải kiểm tra, không chỉ UI.
- **Department mapping:** Khi Admin tạo user production, `department_id` là bắt buộc và phải tham chiếu department đang active. Không suy luận department từ email domain. Development self-register luôn nhận `department_id = NULL` và `ROLE_EMPLOYEE`.

### 4.2 Luồng đóng ticket
- **Quyết định:** Agent chuyển sang `RESOLVED` → Nếu người báo không phản hồi trong 5 ngày làm việc → tự động `CLOSED`.
- **Quyền:** Employee không tự đóng ticket. Agent chỉ resolve ticket được assign cho mình; Manager có thể resolve/close trong department; Admin có thể resolve/close mọi ticket. Tác vụ auto-close chạy bằng system actor và phải ghi audit event.
- **Mở rộng sau:** Thêm `PENDING_CUSTOMER_CONFIRMATION` nếu cần.

### 4.3 SLA
- **Giờ làm việc:** 08:00–17:30, Thứ Hai–Thứ Sáu, timezone `Asia/Ho_Chi_Minh`.
- **Ngày lễ:** Theo lịch ngày lễ Việt Nam.
- **URGENT/CRITICAL:** Tính 24/7.
- **Trạng thái dừng đồng hồ:** `PENDING_CUSTOMER` (chờ phản hồi người báo).
- **Lưu:** `sla_paused_duration` trên ticket.

### 4.4 CSAT
- **Quyết định:** **Tạm bỏ** khỏi UI và báo cáo cho đến khi có survey sau đóng ticket.
- **Lý do:** Hiển thị CSAT không có nguồn thu thập là gian dối dữ liệu.

### 4.5 Ticket ↔ CMDB
- **Quyết định:** **Tạm không bắt buộc** gắn CI vào ticket. Giữ field `affected_ci_id` là optional.
- **Lý do:** UI chưa hỗ trợ, cần BE-10 (attachment) trước.

### 4.6 Create on behalf
- **Quyết định:** AGENT/ADMIN có thể tạo ticket thay người khác.
- **Bắt buộc:** Ghi `created_on_behalf_by = actor_id` và `requester_id = người được tạo hộ`.
- **EMPLOYEE:** `requester_id` **luôn** = actor, không nhận từ request body.

## 5. Test matrix sinh trực tiếp từ ma trận quyền

Mỗi ô `resource × action × role` ở mục 3 sinh test theo quy tắc xác định sau:

| Giá trị ô quyền | Positive case | Boundary/negative case | Kết quả bắt buộc |
|---|---|---|---|
| `—` | Không có | Actor đã đăng nhập gọi action bị cấm | `403` |
| `own` | Bản ghi của actor | Bản ghi của actor khác | `2xx` / `404` |
| `assigned` | Bản ghi assign cho actor | Bản ghi assign cho actor khác | `2xx` / `404` |
| `queue` | Bản ghi `assignee_id IS NULL` | Bản ghi đã assign cho actor khác | `2xx` / `404` |
| `department` | Bản ghi cùng department | Bản ghi khác department hoặc actor không có department | `2xx` / `404` |
| `all` hoặc `✅` | Một bản ghi/action hợp lệ | Payload/state/version không hợp lệ | `2xx` / `4xx` theo domain contract |

Mọi protected action còn sinh thêm case không có/không hợp lệ Bearer token → `401`. Mọi create/on-behalf case phải assert các identity field do server gán, không tin giá trị actor từ request body.

### 5.1 Fixture chuẩn cho scope-sensitive tests

| Fixture | Actor | Resource | Kỳ vọng |
|---|---|---|---|
| `employee_own` | Employee A / Dept IT | requester hoặc creator = A | Cho phép với `own` |
| `employee_peer` | Employee A / Dept IT | requester = Employee B / Dept IT | `404` với `own` |
| `employee_other_dept` | Employee A / Dept IT | requester = Employee C / Dept HR | `404` với `own`/`department` |
| `agent_assigned` | Agent A | assignee = Agent A | Cho phép với `assigned` |
| `agent_other_assigned` | Agent A | assignee = Agent B | `404` với `assigned + queue` |
| `agent_global_queue` | Agent A | assignee = `NULL` ở bất kỳ department | Cho phép theo quyết định queue chung |
| `manager_same_dept` | Manager IT | resource.department = IT | Cho phép với `department` |
| `manager_other_dept` | Manager IT | resource.department = HR | `404` với `department` |
| `admin_any` | Admin | resource bất kỳ | Cho phép với `all` |

### 5.2 Các assertion bắt buộc theo resource

- Ticket: list/get/comment/timeline không trả record ngoài scope; self-assign chỉ áp dụng queue; employee create luôn ép `requester_id = actor_id`.
- Workflow/Approval: requester chỉ đọc instance của mình; approval chỉ assigned user/role pool; actor từ JWT quyết định CAB identity.
- Employee/Department: employee chỉ đọc profile đầy đủ của mình; dữ liệu đồng nghiệp chỉ gồm field read-only đã phê duyệt.
- Knowledge: public/internal visibility phải được lọc ở query trước khi dữ liệu đi vào Search/RAG.
- Notification: list/count/read receipt cùng dùng một recipient scope.
- Reporting/Audit: Employee và Agent nhận `403`; Manager bị giới hạn department khi bảng quy định `department`.

Bộ test PostgreSQL triển khai từ matrix này thuộc Gate B-01; Gate A chỉ chốt input, fixture và expected status để không còn quyết định ngầm khi viết test.

---

> ⚠️ **Tài liệu này đã sẵn sàng để owner phê duyệt nhưng chưa được ký. Không bắt đầu BE-02 trước khi có chữ ký/phê duyệt có thể truy vết.**
> Mọi thay đổi phải được ghi revision và re-approve.
