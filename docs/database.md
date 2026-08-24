# EOMP — Database Architecture & Data Strategy

> **Tài Liệu Chiến Lược Cơ Sở Dữ Liệu Phân Tán (Database Strategy)**  
> **Xem tài liệu chi tiết:** [Master ERD & Data Dictionary](architecture/database_erd_and_data_dictionary.md)

---

## 1. Database Architecture & Ownership Strategy

EOMP tuân thủ nguyên tắc cốt lõi của Microservices: **Database-per-Service Pattern**. Mỗi service chịu trách nhiệm duy nhất về dữ liệu của mình:

* Tuyệt đối không thực hiện Cross-Database Joins trực tiếp qua SQL.
* Đồng bộ dữ liệu liên service được thực hiện thông qua **Event-Driven Architecture (CloudEvents EventBus)** hoặc **API Gateway Aggregation**.
* Cơ chế **PostgreSQL Auto-Migration Runner** tự động khởi chạy và áp dụng các file SQL migration khi mỗi microservice khởi động.

---

## 2. Master Databases Matrix (9 Dedicated Microservice Databases)

| Service | Database Name | Entities Owned (Bảng Dữ Liệu) | Vai Trò & Ràng Buộc Chính |
|---|---|---|---|
| **Auth Service** | `auth_db` | `users`, `refresh_tokens` | Email Unique, Bcrypt hashed password, Role enum |
| **Employee Service** | `employee_db` | `departments`, `employees` | Cấu trúc cây tổ chức, Quan hệ đệ quy `manager_id` |
| **Asset Service** | `asset_db` | `assets`, `asset_assignments`, `configuration_items`, `ci_relationships` | Vòng đời phần cứng, Đồ thị phụ thuộc hạ tầng CMDB |
| **Helpdesk Service** | `helpdesk_db` | `service_categories`, `service_catalog_items`, `tickets`, `ticket_comments`, `ticket_timeline`, `problems`, `problem_incident_links` | SLA targets, ITIL Problem RCA & KEDB |
| **Workflow Service** | `workflow_db` | `workflow_definitions`, `workflow_instances`, `workflow_steps`, `approval_requests`, `workflow_logs`, `change_requests`, `cab_reviews` | State machine approvals, Risk Matrix 3x3, CAB |
| **Notification Service** | `notification_db` | `notifications`, `notification_templates` | In-app notification queue, Unread counter |
| **Knowledge Base** | `knowledge_db` | `knowledge_categories`, `knowledge_articles`, `runbooks`, `document_embeddings` | SOP Runbooks, Full-text Search + Vector storage |
| **Audit Service** | `audit_db` | `audit_logs`, `security_events` | **SHA-256 Checksum** bất biến, Data masking |
| **Reporting & BI** | `reporting_db` | `sla_metrics_daily`, `agent_performance`, `category_metrics`, `department_sla_metrics`, `raw_incident_records` | Aggregated KPI tables, MTTR/MTTD analytics |

---

## 3. Migration Runner Engine (`packages/shared/pkg/database`)

Tất cả các microservices đều sử dụng module kết nối chuẩn `packages/shared/pkg/database/postgres.go`:

1. **Connection Pool Cấu Hình:** `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, `SetConnMaxLifetime(15m)`.
2. **Auto Migration:** Quét toàn bộ file `migrations/*.sql` theo thứ tự alphabet/số (`001_...`, `002_...`).
3. **Transaction Safety:** Mỗi file SQL được thực thi trong một `sql.Tx` độc lập. Nếu xảy ra lỗi cú pháp, transaction tự động rollback an toàn.

---

## 4. Vector Database & Caching Storage

* **Qdrant Vector DB (`:6333`):** Lưu trữ vector embeddings (1536/768 dimensions) của Knowledge Base articles và SOP Runbooks phục vụ tìm kiếm ngữ nghĩa RAG với cosine similarity scoring.
* **Redis 7 (`:6379`):** Lưu trữ Token Bucket Rate Limiting per IP, Session Cache, và pub/sub tạm thời.
* **MinIO Object Storage (`:9000`):** Lưu trữ ảnh đính kèm Ticket, tài liệu đính kèm Knowledge Base, và tệp sao lưu WAL database nén.
