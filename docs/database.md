# EOMP — Kiến trúc và chiến lược dữ liệu

Last audited: 2026-08-30

EOMP áp dụng mô hình database-per-service. Mỗi service chỉ truy cập PostgreSQL database do mình sở hữu; trao đổi liên service đi qua HTTP API hoặc CloudEvents/RabbitMQ, không dùng cross-database join.

## Ma trận dữ liệu

| Service | Database | Bảng do service sở hữu | Ghi chú chính |
|---|---|---|---|
| Auth | `auth_db` | `users`, `refresh_tokens`, `login_audit_logs` | bcrypt, token rotation và nhật ký đăng nhập |
| Employee | `employee_db` | `departments`, `employees` | cây tổ chức; `employees.version` dùng optimistic locking |
| Asset | `asset_db` | `assets`, `asset_assignments`, `configuration_items`, `ci_relationships` | vòng đời tài sản và CMDB; asset/CI có `version` |
| Helpdesk | `helpdesk_db` | `service_categories`, `service_catalog_items`, `tickets`, `ticket_comments`, `ticket_timeline`, `problems`, `problem_incident_links` | ticket/problem có `version`; số TK/PRB dùng sequence; dữ liệu nhân viên được giới hạn theo requester |
| Workflow | `workflow_db` | `workflow_definitions`, `workflow_instances`, `workflow_steps`, `approval_requests`, `workflow_logs`, `change_requests`, `cab_reviews` | approval theo user/role; workflow/change có `version`; số WFI/CHG dùng sequence |
| Notification | `notification_db` | `notifications`, `notification_templates`, `notification_reads` | trạng thái đọc theo từng recipient, kể cả notification broadcast |
| Knowledge | `knowledge_db` | `knowledge_categories`, `knowledge_articles`, `runbooks`, `document_embeddings` | PostgreSQL full-text search; article có `version` |
| Audit | `audit_db` | `audit_logs`, `security_events` | chuỗi HMAC-SHA256, sequence ổn định và trigger append-only |
| Reporting | `reporting_db` | `sla_metrics_daily`, `agent_performance`, `category_metrics`, `department_sla_metrics`, `raw_incident_records` | dữ liệu KPI tổng hợp; không còn seed telemetry giả sau migration cleanup |

## Migration runner

`packages/shared/pkg/database` thực hiện các migration `*.sql` theo thứ tự tên file và ghi phiên bản đã áp dụng. Service phụ thuộc database sẽ dừng khởi động nếu kết nối hoặc migration thất bại.

Quy tắc migration:

- Chỉ thêm migration mới cho môi trường đã tồn tại; không sửa ý nghĩa migration đã phát hành nếu có thể tránh được.
- DDL/DML phải chạy lại an toàn khi phù hợp (`IF NOT EXISTS`, `ON CONFLICT`).
- Cleanup fixture phải định danh chính xác bằng ID và thuộc tính fixture; không dùng điều kiện rộng có thể xóa dữ liệu thật.
- Mọi thay đổi CAS phải cập nhật đồng bộ schema, repository, API contract và frontend DTO.
- Trước release phải thử cả fresh install và upgrade từ bản gần nhất trên PostgreSQL thật.

## Lưu trữ ngoài PostgreSQL

- **Redis 7 (`:6379`)**: rate limiting và dữ liệu tạm thời. Redis không phải nguồn dữ liệu nghiệp vụ chuẩn.
- **RabbitMQ (`:5672`)**: vận chuyển CloudEvents giữa service; publisher báo lỗi khi mất kết nối.
- **Qdrant (`:6333`)**: vector store tùy chọn cho AI/RAG. Trạng thái được probe qua `/api/v1/ai/status`; PostgreSQL search vẫn là nguồn tìm kiếm Knowledge hiện hành.
- **MinIO (`:9000`)**: object storage theo cấu hình triển khai. Chưa có bằng chứng E2E production cho luồng upload/restore trong lần audit này.

## Tính nhất quán và an toàn

- Các aggregate quan trọng dùng trường `version` và compare-and-swap; request stale trả HTTP 409.
- Workflow tạo instance, approval đầu tiên và log đầu tiên trong cùng transaction.
- Notification broadcast không dùng cờ đọc toàn cục; `notification_reads(notification_id, recipient_id)` lưu receipt riêng.
- Audit chain khóa đầu chuỗi khi append và chặn UPDATE/DELETE thông thường bằng trigger. Production vẫn cần dedicated DB role với grant append-only.

Chi tiết bảng và quan hệ: [Master ERD & Data Dictionary](architecture/database_erd_and_data_dictionary.md).
