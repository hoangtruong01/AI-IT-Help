# EOMP — System Architecture Overview

> **Tài Liệu Tổng Quan Kiến Trúc Hệ Thống (High-Level Architecture)**  
> **Áp dụng cho:** Solution Architects, Tech Leads & Full-Stack Developers.  
> **Xem thêm hồ sơ chuyên sâu:** [C4 Model Diagrams](architecture/c4_model_diagrams.md)

---

## 1. System Overview & Monorepo Topology

Hệ thống EOMP áp dụng mô hình **Distributed Microservices Monorepo** tuân thủ nghiêm ngặt nguyên lý **Clean Architecture** (Ports & Adapters) tại từng service và Event-Driven Architecture (EDA) qua CloudEvents v1.0.

```
                            ┌──────────────────────────────────────────┐
                            │        Nuxt 4 Web Frontend (:3000)       │
                            │   Vue 3 · Tailwind CSS · Pinia · SSR    │
                            └────────────────────┬─────────────────────┘
                                                 │ REST / SSE / WebSockets
                            ┌────────────────────▼─────────────────────┐
                            │         API Gateway Go (:8080)           │
                            │  Reverse Proxy · JWT · Rate Limit · RED  │
                            └───────┬────────────┬─────────────┬───────┘
                                    │            │             │
        ┌───────────────────────────┼────────────┴─────────────┼───────────────────────────┐
        ▼                           ▼                          ▼                           ▼
 ┌──────────────┐            ┌──────────────┐           ┌──────────────┐            ┌──────────────┐
 │     Auth     │            │   Employee   │           │    Asset     │            │   Helpdesk   │
 │    :8081     │            │    :8082     │           │    :8083     │            │    :8084     │
 └──────┬───────┘            └──────┬───────┘           └──────┬───────┘            └──────┬───────┘
        │                           │                          │                           │
        ▼                           ▼                          ▼                           ▼
 ┌──────────────┐            ┌──────────────┐           ┌──────────────┐            ┌──────────────┐
 │   Workflow   │            │ Notification │           │  Knowledge   │            │  AI Copilot  │
 │    :8085     │            │    :8086     │           │    :8087     │            │    :8088     │
 └──────┬───────┘            └──────┬───────┘           └──────┬───────┘            └──────┬───────┘
        │                           │                          │                           │
        ▼                           ▼                          ▼                           ▼
 ┌──────────────┐            ┌──────────────┐           ┌──────────────────────────────────────────┐
 │ Audit Trail  │            │ Reporting/BI │           │          CloudEvents EventBus            │
 │    :8089     │            │    :8090     │           │       RabbitMQ 4 · Asynchronous          │
 └──────┬───────┘            └──────┬───────┘           └──────────────────────────────────────────┘
        │                           │
        └───────────────────────────┴──────────────────────────┐
                                                               ▼
 ┌─────────────────────────────────────────────────────────────────────────────────────────────┐
 │                                   Distributed Infrastructure                                │
 │   PostgreSQL 17 (8 Isolated DBs) · Redis 7 · MinIO S3 · Qdrant Vector Store                 │
 │   Prometheus Metrics · Grafana Dashboards (:3002) · Loki Log Aggregator                     │
 └─────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Microservices Responsibilities Matrix

| Microservice | Bounded Context & Trách Nhiệm Kỹ Thuật | Port | Database |
|---|---|:---:|---|
| **API Gateway** | Reverse Proxy Routing, JWT Header Injection, Sliding Window Rate Limiting (100 req/min/IP), Correlation ID, RED Metrics Aggregator | `:8080` | *(Stateless / Redis)* |
| **Auth Service** | Quản lý định danh người dùng, cấp phát JWT HS256, Refresh Token, Bcrypt password hashing, phân quyền Strict RBAC 4 vai trò | `:8081` | `auth_db` |
| **Employee Service** | Danh bạ nhân viên, sơ đồ tổ chức, phân cấp phòng ban, chức danh & thông tin liên hệ | `:8082` | `employee_db` |
| **Asset & CMDB** | Quản lý vòng đời thiết bị phần cứng, bàn giao/thu hồi, bản quyền phần mềm, và đồ thị quan hệ cấu hình hạ tầng CMDB | `:8083` | `asset_db` |
| **Helpdesk & SLA** | Quản lý sự cố (Incident Ticketing), tính toán hạn cam kết SLA tự động, ITIL Problem Management (RCA 5-Whys, Known Errors) | `:8084` | `helpdesk_db` |
| **Workflow Engine** | State Machine Workflow Engine, quy trình phê duyệt đa cấp, Change Advisory Board (CAB) & Ma trận rủi ro (Risk Matrix 3x3) | `:8085` | `workflow_db` |
| **Notification** | Lắng nghe Domain Events từ EventBus, gửi thông báo in-app realtime, quản lý biên nhận đọc (Read Receipt) | `:8086` | `notification_db` |
| **Knowledge Base** | Cẩm nang kỹ thuật, tài liệu SOP, Runbooks, Full-text Search và lưu trữ Document Embeddings | `:8087` | `knowledge_db` |
| **AI Copilot** | Trợ lý ảo IT Ops Copilot, RAG Engine, phân loại Ticket tự động (Auto-triage), tìm kiếm ngữ nghĩa với Qdrant Vector Store | `:8088` | *(Stateless / Qdrant)* |
| **Audit Service** | Ghi nhận nhật ký append-only với chuỗi **HMAC-SHA256**, kiểm tra toàn vẹn và tự động ẩn thông tin nhạy cảm | `:8089` | `audit_db` |
| **Reporting & BI** | Thống kê MTTR/MTTD, tỷ lệ vi phạm SLA, hiệu suất kỹ thuật viên (Agent Scorecard), xuất báo cáo tốc độ cao PDF & Excel | `:8090` | `reporting_db` |

---

## 3. Communication Patterns

1. **Frontend ➔ Backend (External):** REST APIs duy nhất qua API Gateway (`http://localhost:8080/api/v1/*`). Frontend không bao giờ gọi trực tiếp microservices nội bộ.
2. **Synchronous Inter-Service:** REST/gRPC với timeout nghiêm ngặt và Circuit Breaker.
3. **Asynchronous Event-Driven (EDA):** Chuẩn **CloudEvents v1.0** qua RabbitMQ 4 (Topics: `ticket.created`, `ticket.sla_warning`, `approval.requested`, `asset.assigned`, `security.alert`).
4. **Data Ownership:** Mỗi microservice sở hữu cơ sở dữ liệu cô lập tuyệt đối. Không chia sẻ schema giữa các service.

---

## 4. Observability & SRE Standards (RED Method)

Mỗi microservice đều cung cấp endpoint `/metrics` theo chuẩn Prometheus text format đo lường 3 chỉ số vàng:
* **Rate:** Tốc độ request (`http_requests_total`).
* **Errors:** Số lượng request lỗi mã 4xx/5xx.
* **Duration:** Thời gian phản hồi p50, p90, p95, p99 (`http_request_duration_seconds`).
