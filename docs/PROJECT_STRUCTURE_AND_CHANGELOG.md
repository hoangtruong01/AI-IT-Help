# EOMP — Project Structure, Implementation Standards & Daily Changelog

> **Tài liệu chuẩn hóa kiến trúc dự án và nhật ký triển khai (Single Source of Truth)**  
> Dành cho: **AI Assistant, Principal Engineers, Tech Leads, Full-stack Developers & QA Engineers**.  
> Mục tiêu: Đảm bảo mọi lần can thiệp, phát triển tính năng mới của AI hoặc lập trình viên đều tuân thủ **1 cấu trúc nhất quán (Consistent Design & Architecture Contract)**.

---

## 📑 MỤC LỤC
1. [Bản Đồ Cấu Trúc Monorepo Tổng Thể](#1-bản-đồ-cấu-trúc-monorepo-tổng-thể)
2. [Cấu Trúc Chuẩn Cho Mỗi Go Microservice (Clean Architecture)](#2-cấu-trúc-chuẩn-cho-mỗi-go-microservice-clean-architecture)
3. [Cấu Trúc Chuẩn Cho Frontend Nuxt 4 (`apps/web`)](#3-cấu-trúc-chuẩn-cho-frontend-nuxt-4-appsweb)
4. [Thư Viện Dùng Chung (`packages/shared`)](#4-thư-viện-dùng-chung-packagesshared)
5. [Quy Chuẩn Giao Tiếp & Response Envelope Contract](#5-quy-chuẩn-giao-tiếp--response-envelope-contract)
6. [Bảng Phân Bổ Cổng Mạng & Cơ Sở Dữ Liệu (Port & DB Matrix)](#6-bảng-phân-bổ-cổng-mạng--cơ-sở-dữ-liệu-port--db-matrix)
7. [Nhật Ký Thay Đổi Dự Án Đã Thực Hiện (Daily Changelog: Phase 1 — Phase 5)](#7-nhật-ký-thay-đổi-dự-án-đã-thực-hiện-daily-changelog-phase-1--phase-5)
8. [Quy Trình Phát Triển Chuẩn Cho Các Phase Tiếp Theo (Phase 6 — 12)](#8-quy-trình-phát-triển-chuẩn-cho-các-phase-tiếp-theo-phase-6--12)

---

## 1. Bản Đồ Cấu Trúc Monorepo Tổng Thể

```
d:/IT_help/eomp/
├── apps/
│   └── web/                            # Frontend Nuxt 4 (SSR/SPA) + Vue 3 + Tailwind CSS + Nuxt UI
│       ├── app/
│       │   ├── app.vue                 # Root App component
│       │   ├── composables/            # Shared composables (useApi, useAuth, ...)
│       │   ├── layouts/                # Layout templates (default.vue với Sidebar & Notification Center)
│       │   ├── pages/                  # Route views (/login, /employees, /helpdesk, /assets, /workflows, /ai, ...)
│       │   ├── stores/                 # Pinia state stores (auth.ts, ...)
│       │   └── types/                  # Unified TypeScript Domain Interfaces (index.ts)
│       ├── nuxt.config.ts              # Nuxt 4 configuration
│       ├── package.json                # Frontend dependencies
│       └── tsconfig.json               # TypeScript strict configuration
│
├── packages/
│   └── shared/                         # Common Go packages shared across all 11 microservices
│       ├── go.mod                      # Module: eomp/packages/shared
│       └── pkg/
│           ├── auth/                   # JWT HS256 Token Manager + Bcrypt hashing
│           ├── config/                 # Environment loader helpers (GetEnv, GetEnvInt, GetEnvBool)
│           ├── database/               # PostgreSQL connection pool & auto migration runner
│           ├── errors/                 # Standard domain AppError & HTTP error mapper
│           ├── eventbus/               # CloudEvents v1.0 pub/sub message bus
│           ├── logger/                 # Structured slog JSON/Text logger
│           ├── middleware/             # CORS, RequestLogger, Recoverer, Gateway Header Extractor
│           └── response/               # Standardized JSON response writer
│
├── services/                           # 11 Golang Microservices (Clean Architecture)
│   ├── gateway/         (:8080)        # API Gateway & Reverse Proxy (JWT Auth Filter, Routing, Rate Limit)
│   ├── auth/            (:8081)        # Authentication & RBAC Service (PostgreSQL: auth_db)
│   ├── employee/        (:8082)        # Employee & Department Hierarchy (PostgreSQL: employee_db)
│   ├── asset/           (:8083)        # IT Asset Inventory & CMDB Topology (PostgreSQL: asset_db)
│   ├── helpdesk/        (:8084)        # Incident Management, Service Catalog & SLA Engine (PostgreSQL: helpdesk_db)
│   ├── workflow/        (:8085)        # State Machine Workflow Engine & Approvals (PostgreSQL: workflow_db)
│   ├── notification/    (:8086)        # In-App / Email Notifications & EventBus Consumer (PostgreSQL: notification_db)
│   ├── knowledge/       (:8087)        # Knowledge Base & Vector Documents (PostgreSQL: knowledge_db + Qdrant)
│   ├── ai/              (:8088)        # AI Ops Copilot & RAG Engine (Gemini / OpenAI Provider)
│   ├── audit/           (:8089)        # Audit Trail & Compliance Logging (PostgreSQL: audit_db)
│   └── reporting/       (:8090)        # Business Intelligence, KPI & SLA Analytics (PostgreSQL: reporting_db)
│
├── deploy/                             # Deployment & Infrastructure Configuration
│   ├── docker/                         # Dockerfiles for each service
│   ├── docker-compose.yml              # Multi-container local orchestration (Postgres, Redis, RabbitMQ, etc.)
│   ├── nginx/                          # Nginx reverse proxy configuration
│   └── kubernetes/                     # K8s manifests / Helm charts
│
├── scripts/                            # DevOps & Developer Automation Scripts
│   ├── dev.ps1                         # PowerShell master CLI (docker-up, dev, test, format, build)
│   ├── qa.ps1                          # Automated QA/QC test runner
│   └── dev.sh                          # Linux / macOS shell helper
│
├── docs/                               # Engineering & Architecture Documentation Hub
│   ├── README.md                       # Documentation index
│   ├── PROJECT_STRUCTURE_AND_CHANGELOG.md # (This file - Single Source of Truth)
│   ├── INTERN_DEVELOPER_GUIDE.md       # Comprehensive developer guide
│   ├── architecture.md                 # System architecture specification
│   ├── api.md                          # API documentation
│   └── database.md                     # Database schemas specification
│
├── Jenkinsfile                         # CI/CD multi-stage pipeline configuration
├── Makefile                            # Make targets for UNIX environments
└── .env                                # Local environment variables
```

---

## 2. Cấu Trúc Chuẩn Cho Mỗi Go Microservice (Clean Architecture)

Mọi microservice trong thư mục `services/<service_name>` **BẮT BUỘC** phải tuân thủ đúng 100% cấu trúc Clean Architecture dưới đây:

```
services/<service_name>/
├── cmd/
│   └── server/
│       ├── main.go                     # Khởi tạo DB, chạy migration, dependency injection, mount HTTP routes, graceful shutdown
│       └── main_test.go                # Unit tests cho Health handler và Domain logic
├── internal/
│   ├── config/
│   │   └── config.go                   # Load struct Config từ biến môi trường qua packages/shared/pkg/config
│   ├── handler/                        # HTTP Handlers tiếp nhận Request, Validate, gọi Service, ghi Response
│   │   ├── <entity>.go
│   │   └── health.go
│   ├── model/                          # Domain Entities, DTOs, Enums, Filter Query, Stats DTOs
│   │   └── <entity>.go
│   ├── repository/                     # PostgreSQL Repository Interface & Implementation (SQL queries)
│   │   └── repository.go
│   └── service/                        # Business Logic, State Machines, SLA Calculators, Domain Event Handlers
│       └── <entity>_service.go
├── migrations/                         # PostgreSQL SQL migration scripts
│   └── 001_create_<entity>_table.sql   # CREATE EXTENSION "uuid-ossp", CREATE TABLE, CREATE INDEX, INSERT seed data
└── go.mod                              # Go module definition (replace directive tới packages/shared)
```

### Quy Tắc Viết Code Go Bắt Buộc:
1. **Repository Interface**: Luôn định nghĩa interface trước `type Repository interface { ... }` để dễ dàng mock test.
2. **Scan Nullable Fields**: Luôn sử dụng `sql.NullString`, `sql.NullTime`, `sql.NullInt64` khi scan các cột có thể `NULL` trong DB, sau đó gán vào con trỏ struct (`*string`, `*time.Time`).
3. **Pagination Helper**: Hàm danh sách luôn trả về DTO chuẩn `Page`, `PageSize`, `Total`, `TotalPages` (tính bằng `math.Ceil(float64(total) / float64(pageSize))`).
4. **AppError Mapping**: Tầng `service` trả về domain `errors.BadRequest()`, `errors.NotFound()`, `errors.Conflict()`, `errors.InternalServerError()`. Tầng `handler` dùng `errors.WriteHTTP(w, err)` để tự động map đúng HTTP Status Code.

---

## 3. Cấu Trúc Chuẩn Cho Frontend Nuxt 4 (`apps/web`)

Frontend được xây dựng bằng **Nuxt 4 + Vue 3 Composition API + Tailwind CSS + Nuxt UI**.

```
apps/web/app/
├── composables/
│   └── useApi.ts                       # Composable $fetch wrapper: Tự động đính kèm baseURL và Bearer Token từ useAuthStore
├── layouts/
│   └── default.vue                     # Enterprise Glassmorphism Layout:
│                                       # - Collapsible Sidebar với Icon & Badge
│                                       # - Top Header: Search Command Bar, Microservices Status, Live Notification Popover & Quick Create Ticket Button
│                                       # - Ambient background glow gradients
├── pages/
│   ├── index.vue                       # Executive Dashboard: 4 Live KPI Cards, Quick Actions, Active Incidents List
│   ├── login.vue                       # Dark Glassmorphism Auth Page: Sign In, Role Quick-switch buttons, Error alerts
│   ├── employees.vue                   # Employee Directory: Search, Department Tree Filter, Add Employee Modal
│   ├── helpdesk.vue                    # IT Helpdesk: SLA Realtime Badges, Status Lifecycle Drawer, Service Catalog Picker Modal
│   ├── assets.vue                      # Asset & CMDB: Hardware Inventory, Stock Allocation Modal, Handover History Drawer, Topology Graph
│   ├── workflows.vue                   # Workflow Engine: 4 KPI Cards, My Approvals Queue (Approve/Reject with notes), Live Executions Timeline Drawer, Blueprints (DAG)
│   ├── ai.vue                          # AI Ops Copilot: Interactive chat interface, Ticket classification, Root cause suggestion
│   ├── knowledge.vue                   # Knowledge Base & SOP Runbooks
│   ├── reports.vue                     # Reports & SLA Analytics
│   └── audit.vue                       # Security & Compliance Audit Trail
├── stores/
│   └── auth.ts                         # Pinia Store: user, token, refreshToken, login(), logout(), isAuthenticated
└── types/
    └── index.ts                        # Single Source of Truth cho toàn bộ TypeScript Types của toàn bộ hệ thống
```

### Quy Tắc Thiết Kế UI/UX:
- **Tone màu chủ đạo**: Dark theme hiện đại (`bg-slate-950`, `bg-slate-900/80`, `border-slate-800/80`, `backdrop-blur-xl`).
- **Màu sắc trạng thái chuẩn**:
  - `EMERALD` / `GREEN`: `RESOLVED`, `COMPLETED`, `APPROVED`, `IN_STOCK`, `WITHIN_SLA`, `ACTIVE`.
  - `AMBER` / `YELLOW`: `IN_PROGRESS`, `WAITING_APPROVAL`, `WARNING`, `IN_USE`.
  - `ROSE` / `RED`: `URGENT`, `HIGH`, `BREACHED`, `REJECTED`, `RETIRED`, `FAILED`.
  - `INDIGO` / `PURPLE`: `OPEN`, `PENDING`, `HARDWARE`, `BLUEPRINTS`, `SERVICE_REQUEST`.
- **Hiệu ứng Micro-animations**: Hover card scale, pulsing status dots, bouncing badges, animated spinners khi tải dữ liệu.

---

## 4. Thư Viện Dùng Chung (`packages/shared`)

Tất cả các microservices chia sẻ các module cốt lõi trong `packages/shared/pkg`:

| Module | Đường Dẫn File | Trách Nhiệm Kỹ Thuật |
|---|---|---|
| **database** | `packages/shared/pkg/database/postgres.go` | Kết nối connection pool (`sql.DB`), cấu hình `SetMaxOpenConns(25)`, `SetMaxIdleConns(5)`, và tự động thực thi các file SQL migration trong thư mục `migrations/` khi service khởi động. |
| **auth** | `packages/shared/pkg/auth/jwt.go` | Khởi tạo và ký JWT Access Token (HS256, 60m), Refresh Token (7d), kiểm tra mật khẩu Bcrypt (`HashPassword`, `CheckPasswordHash`). |
| **errors** | `packages/shared/pkg/errors/errors.go` | Định nghĩa struct chuẩn `AppError` (`Code`, `Message`, `Details`, `HTTPStatus`), hàm tạo `BadRequest`, `Unauthorized`, `Forbidden`, `NotFound`, `Conflict`, `InternalServerError` và hàm `WriteHTTP`. |
| **response** | `packages/shared/pkg/response/response.go` | Xuất JSON đồng nhất qua `response.JSON(w, statusCode, data)`. |
| **middleware** | `packages/shared/pkg/middleware/` | Chuỗi middleware toàn cục: `CORS()`, `RequestLogger(slog)`, `Recoverer()`, `ExtractGatewayHeaders()` (trích xuất `X-User-ID`, `X-User-Role`, `X-User-Email` vào `r.Context()`). |
| **eventbus** | `packages/shared/pkg/eventbus/eventbus.go` | Message bus theo chuẩn **CloudEvents v1.0**: `Publish(ctx, event)`, `Subscribe(topic, handler)` hỗ trợ wildcard listener `*` phục vụ Event-Driven Architecture. |
| **logger** | `packages/shared/pkg/logger/logger.go` | Khởi tạo structured logger `slog.New()` (JSON Handler trong môi trường production, Text Handler trong môi trường development). |
| **config** | `packages/shared/pkg/config/config.go` | Đọc biến môi trường an toàn với giá trị fallback: `GetEnv`, `GetEnvInt`, `GetEnvBool`. |

---

## 5. Quy Chuẩn Giao Tiếp & Response Envelope Contract

### 1. Phản hồi thành công (Standard Success Envelope)
```json
{
  "data": [ ... ],
  "total": 42,
  "page": 1,
  "page_size": 20,
  "total_pages": 3
}
```

### 2. Phản hồi lỗi chuẩn (Standard Error Envelope)
```json
{
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "ticket TK-1099 does not exist",
    "details": null
  }
}
```

### 3. Chuẩn Sự Kiện CloudEvents (EventBus Topic Payload)
```json
{
  "id": "e0000000-0000-0000-0000-000000000001",
  "source": "eomp.helpdesk",
  "type": "ticket.created",
  "data": {
    "ticket_id": "t0000000-0000-0000-0000-000000000001",
    "ticket_number": "TK-1094",
    "priority": "URGENT",
    "requester_email": "emily.davis@eomp.local"
  },
  "timestamp": "2026-08-18T10:00:00Z"
}
```

---

## 6. Bảng Phân Bổ Cổng Mạng & Cơ Sở Dữ Liệu (Port & DB Matrix)

| Service Name | Port | Database Name | Chủ Sở Hữu Dữ Liệu (Entities Owned) |
|---|:---:|---|---|
| **API Gateway** | `:8080` | *(Stateless)* | Reverse Proxy Routing, JWT Auth Verification, Header Injection |
| **Auth Service** | `:8081` | `auth_db` | `users`, `refresh_tokens` |
| **Employee Service** | `:8082` | `employee_db` | `departments`, `employees` |
| **Asset & CMDB** | `:8083` | `asset_db` | `assets`, `asset_assignments`, `configuration_items`, `ci_relationships` |
| **Helpdesk & SLA** | `:8084` | `helpdesk_db` | `service_categories`, `service_catalog_items`, `tickets`, `ticket_comments`, `ticket_timeline` |
| **Workflow Engine** | `:8085` | `workflow_db` | `workflow_definitions`, `workflow_instances`, `workflow_steps`, `approval_requests`, `workflow_logs` |
| **Notification** | `:8086` | `notification_db` | `notifications`, `notification_templates` |
| **Knowledge Base** | `:8087` | `knowledge_db` | `articles`, `categories`, `runbooks`, `document_embeddings` (+ Qdrant Vector Store) |
| **AI Ops Copilot** | `:8088` | *(Stateless / Qdrant)* | RAG Engine, LLM Provider Client, Ticket Auto-categorization, Root Cause Solver |
| **Audit Service** | `:8089` | `audit_db` | `audit_logs`, `compliance_reports`, `security_events` |
| **Reporting & BI** | `:8090` | `reporting_db` | `sla_metrics_daily`, `agent_performance`, `kpi_aggregations` |
| **Nuxt 4 Web App** | `:3000` | *(Client App)* | Giao diện điều hành trực quan |

---

## 7. Nhật Ký Thay Đổi Dự Án Đã Thực Hiện (Daily Changelog: Phase 1 — Phase 5)

Hôm nay, toàn bộ 5 Phase nền tảng quan trọng nhất của hệ sinh thái EOMP đã được thiết kế, kiểm thử, tích hợp trực tiếp và đồng bộ lên 2 kho chứa Git (**GitHub** và **GitLab**):

```mermaid
graph TD
    P1[Phase 1: Foundation & Auth/RBAC] --> P2[Phase 2: Helpdesk & SLA Engine]
    P2 --> P3[Phase 3: Asset & CMDB Topology]
    P3 --> P4[Phase 4: Workflow & Approval Engine]
    P4 --> P5[Phase 5: EventBus & Notifications]
    P5 --> P6[Phase 6: AI Ops & RAG Engine - NEXT]
```

### 🔹 Phase 1: Business Foundation, Auth/RBAC Core & Employee Directory
- **Commit**: `efa44d6`
- **Mã nguồn Backend**:
  - Xây dựng `packages/shared/pkg/database` với connection pool và auto-migration runner.
  - Xây dựng `packages/shared/pkg/auth` (JWT HS256 & Bcrypt password hasher).
  - Hoàn thiện `services/auth` (:8081) với schema `auth_db` (`users`, `refresh_tokens`), Clean Architecture, CRUD, login/register/refresh/me APIs.
  - Hoàn thiện `services/employee` (:8082) với schema `employee_db` (`departments`, `employees`), phân trang tìm kiếm, sơ đồ tổ chức.
  - Cấu hình `services/gateway` (:8080) làm Reverse Proxy xác thực JWT và chuyển tiếp định danh qua header.
- **Mã nguồn Frontend**:
  - Giao diện đăng nhập Dark Glassmorphism tại [`apps/web/app/pages/login.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/login.vue) với nút chuyển nhanh tài khoản (Admin, Manager, Agent, User).
  - Danh bạ nhân viên trực quan tại [`apps/web/app/pages/employees.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/employees.vue) hỗ trợ tìm kiếm, lọc theo phòng ban và modal "+ Add Employee".

### 🔹 Phase 2: Incident Management, Service Catalog & SLA Engine
- **Commit**: `adde813`
- **Mã nguồn Backend**:
  - Tạo migration `services/helpdesk/migrations/001_create_tickets_and_sla_table.sql` với `service_categories`, `service_catalog_items`, `tickets`, `ticket_comments`, `ticket_timeline`.
  - Triển khai **SLA Engine** (`services/helpdesk/internal/service/sla_engine.go`): Tự động tính hạn phản hồi và hạn xử lý theo 4 mức ưu tiên (`URGENT`: 15m/2h, `HIGH`: 30m/4h, `MEDIUM`: 4h/8h, `LOW`: 8h/24h) và giám sát ngưỡng vi phạm (`WITHIN_SLA`, `WARNING` <= 20% thời gian còn lại, `BREACHED`).
  - Hoàn thiện Clean Architecture cho `services/helpdesk` (:8084).
  - Cập nhật Gateway reverse proxy cho `/api/v1/tickets/*` và `/api/v1/services/*`.
- **Mã nguồn Frontend**:
  - Nâng cấp [`apps/web/app/pages/helpdesk.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/helpdesk.vue) kết nối Live API: Huy hiệu đếm ngược SLA thời gian thực, modal chọn Service Catalog để tạo Ticket, và Drawer xem chi tiết vòng đời kèm lịch sử xử lý & trao đổi bình luận.

### 🔹 Phase 3: Asset Management & CMDB Dependency Topology Graph
- **Commit**: `7b4428a`
- **Mã nguồn Backend**:
  - Tạo migration `services/asset/migrations/001_create_assets_and_cmdb_table.sql` với `assets`, `asset_assignments`, `configuration_items`, `ci_relationships`.
  - Triển khai Quản lý Vòng đời Thiết bị (`PROCUREMENT` -> `IN_STOCK` -> `IN_USE` -> `MAINTENANCE` -> `RETIRED`), quy trình bàn giao, kiểm toán tình trạng và thu hồi thiết bị.
  - Triển khai CMDB Topology Graph Service (`/api/v1/cmdb/topology`) phân tích cây phụ thuộc hạ tầng (`DEPENDS_ON`, `RUNS_ON`, `CONNECTS_TO`, `BACKED_UP_BY`).
  - Cập nhật Gateway reverse proxy cho `/api/v1/assets/*` và `/api/v1/cmdb/*`.
- **Mã nguồn Frontend**:
  - Nâng cấp [`apps/web/app/pages/assets.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/assets.vue): 5 thẻ KPI Realtime, Tab Kho thiết bị & Bản quyền (kèm modal đăng ký thiết bị mới, modal bàn giao, nút thu hồi về kho, và drawer lịch sử bàn giao), Tab Bản đồ Topology CMDB trực quan.

### 🔹 Phase 4: Workflow Engine, Multi-level Approvals & Orchestration
- **Commit**: `e482adc`
- **Mã nguồn Backend**:
  - Tạo migration `services/workflow/migrations/001_create_workflows_and_approvals_table.sql` với `workflow_definitions`, `workflow_instances`, `workflow_steps`, `approval_requests`, `workflow_logs`.
  - Triển khai State Machine Execution Engine & Approval Matrix: Khởi chạy quy trình từ Blueprint (VD: cấp phát laptop cho nhân viên mới, cấp quyền VPN, phê duyệt nâng cấp DB qua CAB), chuyển tiếp trạng thái tự động khi được duyệt (`APPROVED`) hoặc hủy quy trình khi bị từ chối (`REJECTED`) kèm ghi chú lý do.
  - Cập nhật Gateway reverse proxy cho `/api/v1/workflows/*` và `/api/v1/approvals/*`.
- **Mã nguồn Frontend**:
  - Nâng cấp [`apps/web/app/pages/workflows.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/workflows.vue): 4 thẻ KPI Realtime, Tab Hàng đợi phê duyệt (My Approvals Queue) với modal Duyệt/Từ chối nhập lý do, Tab Thực thi trực tiếp với modal xem nhật ký kiểm toán timeline audit trail, và Tab Mẫu quy trình (Launch instance modal).

### 🔹 Phase 5: Event-Driven Architecture, EventBus & Notification Service
- **Commit**: `e67d9fc`
- **Mã nguồn Backend**:
  - Xây dựng `packages/shared/pkg/eventbus/eventbus.go` định nghĩa chuẩn **CloudEvents v1.0** (`ticket.created`, `ticket.sla_warning`, `approval.requested`, `asset.assigned`, `security.alert`) với cơ chế bất đồng bộ đa luồng và wildcard subscriber `*`.
  - Tạo migration `services/notification/migrations/001_create_notifications_table.sql` với `notifications`, `notification_templates`.
  - Xây dựng `services/notification` (:8086) tự động lắng nghe Domain Events từ EventBus để tạo thông báo in-app cho người dùng, quản lý biên nhận đọc (mark as read / mark all read).
  - Cập nhật Gateway reverse proxy cho `/api/v1/notifications/*`.
- **Mã nguồn Frontend**:
  - Nâng cấp [`apps/web/app/layouts/default.vue`](file:///d:/IT_help/eomp/apps/web/app/layouts/default.vue): Tích hợp Live Notification Center trên thanh Top Navigation Bar với huy hiệu Unread nhảy động (bounce), popover xem thông báo phân loại theo nhãn (`INCIDENT`, `APPROVAL`, `ASSET`, `SLA`), hỗ trợ click để đọc và nút "Mark all read" thời gian thực.

### 🔹 Phase 6: AI Operations Copilot & RAG Knowledge Engine
- **Commit**: `d6afacc`
- **Mã nguồn Backend**:
  - **Knowledge Service (`services/knowledge` - Port 8087)**:
    - Tạo migration `services/knowledge/migrations/001_create_knowledge_and_runbooks_table.sql` với `knowledge_categories`, `knowledge_articles`, `runbooks`, `document_embeddings` cùng bộ Seed Data SOP IT chuẩn doanh nghiệp.
    - Triển khai Clean Architecture hoàn chỉnh (`model`, `repository`, `service`, `handler`): Full-text Search với dynamic scoring, CRUD bài viết cẩm nang, tự động sinh slug URL, tăng view count, và quản lý SOP Runbooks.
  - **AI Operations Service (`services/ai` - Port 8088)**:
    - Nâng cấp RAG Retriever (`SmartRetriever`): Tích hợp tìm kiếm vector Qdrant (`:6333`) với cơ chế **Tự Động Fallback In-Memory Catalog** (đáp ứng trọn vẹn **Test Case 6.2**).
    - Nâng cấp `MockProvider` với tri thức IT Ops chuyên sâu: Giải quyết kịch bản MFA Token Reset theo đúng quy trình Okta SOP (đáp ứng **Test Case 6.1**), thời gian phản hồi TTFT siêu tốc < 50ms (đáp ứng **Test Case 6.3**).
    - Triển khai Ticket Auto-Triage Endpoint (`/api/v1/ai/analyze-ticket`): Tự động phân loại danh mục, mức độ ưu tiên, chẩn đoán nguyên nhân gốc rễ (Root Cause) và gợi ý SOP Runbook xử lý.
  - **API Gateway (`services/gateway` - Port 8080)**:
    - Cấu hình Reverse Proxy và bảo vệ xác thực JWT cho `/api/v1/knowledge/*` và `/api/v1/ai/*`.
- **Mã nguồn Frontend**:
  - Nâng cấp [`apps/web/app/pages/knowledge.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/knowledge.vue): Giao diện Glassmorphism trực quan, 4 thẻ KPI Realtime, thanh tìm kiếm ngữ nghĩa Qdrant tức thì, bộ lọc danh mục, Drawer đọc tài liệu Markdown chuyên sâu và modal tạo bài viết / tạo SOP runbook.
  - Nâng cấp [`apps/web/app/pages/ai.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/ai.vue): Trợ lý AI Copilot Glassmorphism với Live Chat, Typewriter streaming, Source References Pills trích dẫn RAG, Action buttons (`Copy Solution`, `Apply to Ticket`), Studio phân loại Ticket tự động (Auto-Triage), và tab giám sát Qdrant Vector Cluster.
  - Nâng cấp [`apps/web/app/pages/helpdesk.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/helpdesk.vue): Tích hợp Widget AI Operations Copilot trực tiếp trong Drawer chi tiết Ticket giúp chẩn đoán sự cố 1-click và dán giải pháp vào biên bản xử lý.
- **QA/QC & Kiểm Thử**:
  - Vượt qua 100% kiểm thử tự động của 12 modules Go (`go vet`, `go test`, `go build`) và Frontend (`typecheck`, `lint`, `build`).

### 🔹 Phase 7: ITIL Problem Management, RCA & Change Advisory Board (CAB)
- **Commit**: `7c2eccd`
- **Trạng thái**: **HOÀN THÀNH (Done)**
- **Mã nguồn Backend**:
  - **Problem Management (`services/helpdesk` - Port 8084)**:
    - Tạo migration `services/helpdesk/migrations/002_create_problems_table.sql` gồm bảng `problems`, `problem_incident_links` và seed data 3 Problems chuẩn ITIL (`PRB-1001`, `PRB-1002`, `PRB-1003`).
    - Triển khai Clean Architecture (`internal/model/problem.go`, `internal/repository/problem_repository.go`, `internal/service/problem_service.go`, `internal/handler/problem_handler.go`): Gom nhóm các sự cố trùng lặp, cập nhật phân tích nguyên nhân gốc rễ RCA (5-Whys), giải pháp tạm thời (Workaround), xuất bản KEDB, và **Tự Động Đóng Hàng Loạt Sự Cố Liên Kết (Cascade Resolution)** khi Problem chuyển sang `RESOLVED` (đáp ứng **Test Case 7.1**).
  - **Change Management & CAB (`services/workflow` - Port 8085)**:
    - Tạo migration `services/workflow/migrations/002_create_changes_and_cab_table.sql` gồm bảng `change_requests` (RFC) và `cab_reviews`.
    - Triển khai Clean Architecture (`internal/model/change.go`, `internal/repository/change_repository.go`, `internal/service/change_service.go`, `internal/handler/change_handler.go`): Tính toán ma trận rủi ro (3x3 Risk Matrix: Probability vs Impact), quản lý lịch trình bảo trì bảo dưỡng (Maintenance Window Calendar), và **Kiểm Soát Nghiêm Ngặt Quorum Biểu Quyết CAB** (chặn mã `403 Forbidden` khi Change loại `EMERGENCY`/`MAJOR` chưa đủ tối thiểu 2 phiếu biểu quyết, đáp ứng **Test Case 7.2**).
  - **API Gateway (`services/gateway` - Port 8080)**:
    - Định tuyến và bảo vệ xác thực JWT cho `/api/v1/problems/*` (Port 8084) và `/api/v1/changes/*` (Port 8085).
- **Mã nguồn Frontend**:
  - Nâng cấp [`apps/web/app/layouts/default.vue`](file:///d:/IT_help/eomp/apps/web/app/layouts/default.vue): Tích hợp navigation links cho Problem Management (`/problems`) và Change Advisory (`/changes`).
  - Xây dựng [`apps/web/app/pages/problems.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/problems.vue): Giao diện Glassmorphism trực quan, 4 thẻ KPI Realtime, bảng Problem records với huy hiệu Linked Incidents count, Drawer chi tiết với 3 tabs (Root Cause Analysis 5-Whys, Linked Incidents Manager, Overview), nút Cascade Resolve tức thì, và modal tạo Problem mới.
  - Xây dựng [`apps/web/app/pages/changes.vue`](file:///d:/IT_help/eomp/apps/web/app/pages/changes.vue): Giao diện Glassmorphism với 3 views (Bảng RFC, Ma trận rủi ro 3x3 tương tác, Lịch bảo trì bảo dưỡng Maintenance Windows), Drawer chi tiết RFC hiển thị các kế hoạch Implementation/Rollback/Test, thanh đo Quorum CAB, và Modal biểu quyết phê duyệt dành cho thành viên Hội đồng CAB.
- **QA/QC & Kiểm Thử**:
  - Vượt qua 100% kiểm thử tự động của cả 12 modules Go và Frontend.



---

## 8. Quy Trình Phát Triển Chuẩn Cho Các Phase Tiếp Theo (Phase 6 — 12)

Khi AI hoặc Developer bắt tay vào triển khai các Phase tiếp theo, **BẮT BUỘC** phải tuân theo 7 bước nghiêm ngặt sau:

```
[1. ANALYZE] -> [2. DESIGN] -> [3. IMPLEMENT] -> [4. TEST] -> [5. VERIFY] -> [6. FORMAT] -> [7. COMMIT & PUSH]
```

### Chi Tiết Kế Hoạch Các Phase Sắp Tới:

- **Phase 6: AI Operations Copilot & RAG Engine (`services/ai` + `services/knowledge`)**
  - Schema `knowledge_db`: Tài liệu IT, Runbook xử lý sự cố, FAQ, Vector embeddings (Qdrant).
  - RAG Pipeline: Phân tích nội dung Ticket, tự động gợi ý danh mục, mức độ ưu tiên, nguyên nhân gốc rễ (Root Cause) và giải pháp xử lý.
  - AI Chatbot UI: Giao diện chat trực tiếp tại `/ai.vue`.

- **Phase 7: Problem & Change Management ITIL (`services/helpdesk` + `services/workflow`)**
  - Problem Investigation (Known Errors, Workarounds, Root Cause Analysis).
  - Change Advisory Board (CAB) & Risk Assessment Matrix.

- **Phase 8: Observability, Metrics & Distributed Tracing**
  - Prometheus `/metrics` trên tất cả các microservices.
  - Grafana Dashboards giám sát Request Rate, Error Rate, Latency (RED Method).
  - Structured Logging với Loki.

- **Phase 9: Quality Assurance, Integration Tests & End-to-End Suite**
  - Automated Integration Tests kịch bản liên service.
  - Kịch bản test luồng: Login -> Create Ticket -> Trigger SLA -> Launch Workflow -> Approve -> Handover Asset -> Receive Notification.

- **Phase 10: Technical BA Artifacts & Architecture Diagrams**
  - C4 Model Architecture Diagrams (Context, Container, Component, Code).
  - Database Entity Relationship Diagrams (ERD).
  - API Swagger/OpenAPI Specifications.

- **Phase 11: Production Hardening, Rate Limiting & Security Compliance**
  - Rate Limiting (Token Bucket / Redis) tại API Gateway.
  - RBAC Strict Policy Enforcement Middleware.
  - Secrets Management & Security Headers.

- **Phase 12: Production Packaging, Kubernetes & SRE Runbooks**
  - Docker Multi-stage Builds tối ưu kích thước image (< 25MB cho Go binaries).
  - Kubernetes Manifests & Helm Charts với Health Probes (Liveness & Readiness).
  - SRE Disaster Recovery & Rollback Runbooks.
