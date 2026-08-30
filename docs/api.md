# EOMP — Master API Reference & Gateway Routing Hub

> **Tài Liệu Đặc Tả Master REST API (Toàn Bộ 11 Microservices)**  
> **Base URL:** `http://localhost:8080/api/v1`  
> **Xem thêm OpenAPI 3.0 Spec:** [OpenAPI 3.0 Hub](openapi/README.md) | [eomp-openapi-spec.yaml](openapi/eomp-openapi-spec.yaml)

---

## 📑 MỤC LỤC
1. [Bảng Phân Luồng API Gateway (Routing Matrix)](#1-bảng-phân-luồng-api-gateway-routing-matrix)
2. [Quy Chuẩn Request & Response Envelopes](#2-quy-chuẩn-request--response-envelopes)
3. [Chi Tiết Endpoints Từng Microservice](#3-chi-tiết-endpoints-từng-microservice)
   - [3.1 Auth & RBAC Service (:8081)](#31-auth--rbac-service-8081)
   - [3.2 Employee Directory Service (:8082)](#32-employee-directory-service-8082)
   - [3.3 Asset & CMDB Topology Service (:8083)](#33-asset--cmdb-topology-service-8083)
   - [3.4 Helpdesk, SLA & ITIL Problem Service (:8084)](#34-helpdesk-sla--itil-problem-service-8084)
   - [3.5 Workflow Engine & CAB Service (:8085)](#35-workflow-engine--cab-service-8085)
   - [3.6 Notification Service (:8086)](#36-notification-service-8086)
   - [3.7 Knowledge Base & SOP Service (:8087)](#37-knowledge-base--sop-service-8087)
   - [3.8 AI Operations Copilot (:8088)](#38-ai-operations-copilot-8088)
   - [3.9 Audit Trail & Compliance Service (:8089)](#39-audit-trail--compliance-service-8089)
   - [3.10 Reporting & BI Analytics Service (:8090)](#310-reporting--bi-analytics-service-8090)
   - [3.11 SRE Monitoring & Health Probes (:8080)](#311-sre-monitoring--health-probes-8080)

---

## 1. Bảng Phân Luồng API Gateway (Routing Matrix)

Tất cả các cuộc gọi API từ Frontend hoặc bên ngoài đều đi qua cổng **API Gateway (`http://localhost:8080`)**:

| Microservice | Port Nội Bộ | Gateway Prefix URL | Bộ Lọc Bảo Mật & Xác Thực |
|---|:---:|---|:---:|
| **Auth Service** | `:8081` | `/api/v1/auth/*` | Public (Login/Register) / JWT (Me/Refresh) |
| **Employee Service** | `:8082` | `/api/v1/employees/*`, `/api/v1/departments/*` | JWT Bearer Token |
| **Asset Service** | `:8083` | `/api/v1/assets/*`, `/api/v1/cmdb/*` | JWT Bearer Token |
| **Helpdesk Service** | `:8084` | `/api/v1/tickets/*`, `/api/v1/services/*`, `/api/v1/problems/*` | JWT Bearer Token |
| **Workflow Engine** | `:8085` | `/api/v1/workflows/*`, `/api/v1/approvals/*`, `/api/v1/changes/*` | JWT Bearer Token |
| **Notification** | `:8086` | `/api/v1/notifications/*` | JWT Bearer Token |
| **Knowledge Base** | `:8087` | `/api/v1/knowledge/*` | JWT Bearer Token |
| **AI Copilot** | `:8088` | `/api/v1/ai/*` | JWT Bearer Token |
| **Audit Service** | `:8089` | `/api/v1/audit/*` | JWT + Strict RBAC (`ADMIN`, `MANAGER`) |
| **Reporting & BI** | `:8090` | `/api/v1/reports/*` | JWT Bearer Token |
| **Observability SRE**| `:8080` | `/api/v1/monitoring/*` | JWT Bearer Token |

---

## 2. Quy Chuẩn Request & Response Envelopes

### Phản hồi Thành Công Phân Trang (Standard Paginated Envelope):
```json
{
  "data": [ ... ],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

### Phản hồi Lỗi Chuẩn (Standard Error Envelope):
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "ticket TK-1099 does not exist",
    "details": null
  }
}
```

---

## 3. Chi Tiết Endpoints Từng Microservice

### 3.1 Auth & RBAC Service (:8081)
* `POST /api/v1/auth/login` — Đăng nhập hệ thống, trả về `access_token`, `refresh_token`, và `user` object.
* `POST /api/v1/auth/refresh` — Cấp mới access token từ refresh token.
* `GET /api/v1/auth/me` — Lấy thông tin định danh và vai trò của tài khoản đang đăng nhập.

### 3.2 Employee Directory Service (:8082)
* `GET /api/v1/employees?page=1&search=John` — Danh sách nhân viên phân trang & tìm kiếm.
* `GET /api/v1/departments` — Danh sách phòng ban và cây tổ chức phân cấp.
* `POST /api/v1/employees` — Thêm mới nhân viên vào hệ thống.

### 3.3 Asset & CMDB Topology Service (:8083)
* `GET /api/v1/assets?status=IN_USE` — Danh mục phần cứng/phần mềm phân trang.
* `POST /api/v1/assets/{id}/assign` — Bàn giao thiết bị cho nhân viên.
* `POST /api/v1/assets/{id}/return` — Thu hồi thiết bị về kho (`IN_STOCK`).
* `GET /api/v1/cmdb/topology` — Cây đồ thị phân tích phụ thuộc hạ tầng CMDB.

### 3.4 Helpdesk, SLA & ITIL Problem Service (:8084)
* `GET /api/v1/tickets?status=OPEN&priority=URGENT` — Danh sách Ticket kèm SLA Realtime.
* `POST /api/v1/tickets` — Tạo yêu cầu hỗ trợ từ Service Catalog.
* `PATCH /api/v1/tickets/{id}/status` — Cập nhật vòng đời Ticket (`IN_PROGRESS`, `RESOLVED`, `CLOSED`).
* `GET /api/v1/problems` — Danh sách sự cố ITIL Problem.
* `POST /api/v1/problems` — Tạo Problem gom nhóm các Incidents trùng lặp và phân tích RCA.

### 3.5 Workflow Engine & CAB Service (:8085)
* `GET /api/v1/approvals/queue` — Hàng đợi phê duyệt của tôi (My Approvals).
* `POST /api/v1/approvals/{id}/action` — Phê duyệt (`APPROVED`) hoặc Từ chối (`REJECTED`) kèm lý do.
* `GET /api/v1/changes` — Danh sách Yêu cầu Thay đổi (RFC) và Hội đồng CAB.
* `POST /api/v1/changes/{id}/vote` — Bỏ phiếu phê duyệt CAB (Yêu cầu quorum >= 2 votes cho Major/Emergency).

### 3.6 Notification Service (:8086)
* `GET /api/v1/notifications` — Danh sách thông báo in-app realtime.
* `PATCH /api/v1/notifications/{id}/read` — Đánh dấu đã đọc 1 thông báo.
* `POST /api/v1/notifications/read-all` — Đánh dấu đã đọc toàn bộ thông báo.

### 3.7 Knowledge Base & SOP Service (:8087)
* `GET /api/v1/knowledge/articles` — Danh mục bài viết cẩm nang kỹ thuật.
* `GET /api/v1/knowledge/search?q=VPN` — Tìm kiếm Full-Text và ngữ nghĩa Qdrant.
* `GET /api/v1/knowledge/runbooks` — Danh sách quy trình chuẩn SOP Runbooks.

### 3.8 AI Operations Copilot (:8088)
* `POST /api/v1/ai/chat` — Live Chat với AI Copilot hỗ trợ RAG và trích dẫn tài liệu SOP.
* `POST /api/v1/ai/analyze-ticket` — Ticket Auto-Triage (Phân loại danh mục, độ ưu tiên, chẩn đoán nguyên nhân gốc rễ).

### 3.9 Audit Trail & Compliance Service (:8089)
* `GET /api/v1/audit/logs?page=1&action=LOGIN` — Tra cứu nhật ký kiểm toán kèm bằng chứng chuỗi **HMAC-SHA256** (chỉ dành cho `ADMIN`/`MANAGER`).
* `GET /api/v1/audit/stats` — Thống kê an ninh và tuân thủ chuẩn SOC2.
* `GET /api/v1/audit/integrity` — Kiểm tra toàn bộ chuỗi HMAC; trả `409` nếu phát hiện sai lệch.

### 3.10 Reporting & BI Analytics Service (:8090)
* `GET /api/v1/reports/overview?range=30d` — Tổng quan MTTR, MTTD và SLA.
* `GET /api/v1/reports/trends` — Xu hướng ticket theo ngày.
* `GET /api/v1/reports/categories` — Phân bổ theo danh mục.
* `GET /api/v1/reports/departments-sla` — SLA theo phòng ban.
* `GET /api/v1/reports/agents` — Năng suất kỹ thuật viên.
* `POST /api/v1/reports/export` — Xuất báo cáo từ dữ liệu thực; lỗi backend không tạo file mẫu.

### 3.11 SRE Monitoring & Health Probes (:8080)
* `GET /api/v1/monitoring/overview` — Tổng hợp trạng thái và độ trễ health probe thật của 11 services.
* `GET /api/v1/monitoring/services` — Kết quả probe từng service; không sinh CPU/RAM/RPS giả.
* `POST /api/v1/monitoring/probe/{id}` — Kích hoạt health probe tức thì.
* `GET /api/v1/monitoring/logs` — Trả `501 Not Implemented` cho đến khi cấu hình log backend tương thích Loki.
