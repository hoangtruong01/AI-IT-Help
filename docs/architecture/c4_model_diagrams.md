# EOMP C4 Model Architecture Blueprints & System Design

Tài liệu này đặc tả toàn bộ kiến trúc hệ thống **EOMP (Enterprise Operations Management Platform)** theo chuẩn mô hình kiến trúc **C4 Model (Context, Container, Component, Code)** tiêu chuẩn quốc tế.

---

## 1. 🌐 Level 1: System Context Diagram

Sơ đồ ngữ cảnh hệ thống định nghĩa ranh giới nền tảng EOMP và mối quan hệ tương tác với người dùng nội bộ và các hệ thống bên ngoài.

```mermaid
C4Context
    title System Context Diagram for EOMP (Enterprise Operations Management Platform)

    Person(employee, "Employee / Requester", "Nhân viên doanh nghiệp gửi yêu cầu CNTT, tra cứu tài sản, xem trạng thái phê duyệt.")
    Person(agent, "IT Operations Specialist", "Kỹ thuật viên IT Support L1/L2 xử lý sự cố, phân bổ thiết bị, thực thi runbook.")
    Person(manager, "Department / IT Manager", "Quản lý phòng ban phê duyệt chi phí, cấp phát thiết bị và xem báo cáo BI.")
    Person(admin, "SRE & System Admin", "Quản trị viên hạ tầng, cấu hình RBAC, giám sát hệ thống và xem nhật ký audit.")

    Enterprise_Boundary(eomp_boundary, "EOMP Platform Boundary") {
        System(eomp, "EOMP Core System", "Nền tảng quản trị vận hành CNTT tập trung: Helpdesk, CMDB, Workflow, AI Copilot, BI Analytics, Observability, Security Audit.")
    }

    System_Ext(mail_server, "Corporate SMTP Server", "Hệ thống email gửi thông báo, mã OTP, phiếu giao việc.")
    System_Ext(vector_db, "Qdrant Vector DB", "Lưu trữ vector embeddings phục vụ RAG và Semantic Search.")
    System_Ext(prometheus_mesh, "Prometheus / Grafana", "Thu thập và trực quan hóa RED Metrics, SLI/SLA và Heatmap hạ tầng.")

    Rel(employee, eomp, "Tạo ticket, xem runbook, theo dõi phê duyệt", "HTTPS/WSS")
    Rel(agent, eomp, "Xử lý ticket, cập nhật CMDB, điều tra lỗi", "HTTPS/WSS")
    Rel(manager, eomp, "Phê duyệt Workflow, xem dashboard SLA/CSAT", "HTTPS/WSS")
    Rel(admin, eomp, "Cấu hình RBAC, xem Audit Trail, SRE Console", "HTTPS/WSS")

    Rel(eomp, mail_server, "Gửi email thông báo sự kiện", "SMTP/TLS")
    Rel(eomp, vector_db, "Tra cứu tài liệu kiến thức tương đồng", "gRPC/HTTP")
    Rel(eomp, prometheus_mesh, "Expose /metrics RED telemetry", "HTTP Pull")
```

---

## 2. 📦 Level 2: Container Diagram

Sơ đồ Container mô tả cấu trúc chi tiết 11 Microservices Go, API Gateway, 8 Cơ sở dữ liệu phân tán PostgreSQL, Message Brokers và Frontend Nuxt 4.

```mermaid
graph TD
    subgraph Client Layer
        Web[Nuxt 4 SPA / SSR Web Client<br/>Port 3000<br/>Vue 3 + TailwindCSS + TypeScript]
    end

    subgraph Edge & Security Boundary
        GW[API Gateway Router<br/>Port 8080 (Golang)<br/>JWT Auth, Strict RBAC, Rate Limiter, Reverse Proxy]
    end

    Web -->|HTTPS REST / WebSocket| GW

    subgraph Core Microservices Layer Golang 1.25
        AuthSvc[auth-service :8081<br/>JWT/RBAC, MFA]
        EmpSvc[employee-service :8082<br/>Org Chart, Departments]
        AssetSvc[asset-service :8083<br/>CMDB, Hardware Lifecycle]
        HelpdeskSvc[helpdesk-service :8084<br/>Incidents, Service Catalog]
        WfSvc[workflow-service :8085<br/>Approval State Machine]
        NotifSvc[notification-service :8086<br/>In-App & Email Alerts]
        KnowSvc[knowledge-service :8087<br/>Runbooks, Articles]
        AISvc[ai-service :8088<br/>RAG Engine, Copilot]
        AuditSvc[audit-service :8089<br/>Immutable SHA-256 Audit Trail]
        ReportSvc[reporting-service :8090<br/>BI Analytics, MTTR/MTTD, SLA]
    end

    GW -->|/api/v1/auth| AuthSvc
    GW -->|/api/v1/employees| EmpSvc
    GW -->|/api/v1/assets| AssetSvc
    GW -->|/api/v1/tickets| HelpdeskSvc
    GW -->|/api/v1/workflows| WfSvc
    GW -->|/api/v1/notifications| NotifSvc
    GW -->|/api/v1/knowledge| KnowSvc
    GW -->|/api/v1/ai| AISvc
    GW -->|/api/v1/audit| AuditSvc
    GW -->|/api/v1/reports| ReportSvc

    subgraph Persistence & Infrastructure Layer
        DB_Auth[(auth_db<br/>Postgres)]
        DB_Emp[(employee_db<br/>Postgres)]
        DB_Asset[(asset_db<br/>Postgres)]
        DB_Helpdesk[(helpdesk_db<br/>Postgres)]
        DB_Wf[(workflow_db<br/>Postgres)]
        DB_Know[(knowledge_db<br/>Postgres)]
        DB_Audit[(audit_db<br/>Postgres)]
        DB_Report[(reporting_db<br/>Postgres)]

        RedisCache[(Redis Cluster :6379<br/>Sessions, Rate Limits, Cache)]
        EventBus[(RabbitMQ :5672<br/>CloudEvents Message Bus)]
        QdrantDB[(Qdrant Vector DB :6333<br/>Vector Embeddings)]
        MinIO[(MinIO Object Storage :9000<br/>Attachments, Exports)]
        Prometheus[(Prometheus :9090<br/>Telemetry Collector)]
        Grafana[(Grafana :3002<br/>SRE Dashboards)]
    end

    AuthSvc --> DB_Auth
    EmpSvc --> DB_Emp
    AssetSvc --> DB_Asset
    HelpdeskSvc --> DB_Helpdesk
    WfSvc --> DB_Wf
    KnowSvc --> DB_Know
    AuditSvc --> DB_Audit
    ReportSvc --> DB_Report

    HelpdeskSvc -.->|Publish Events| EventBus
    WfSvc -.->|Publish Events| EventBus
    EventBus -.->|Consume Events| NotifSvc
    AISvc --> QdrantDB
    KnowSvc --> MinIO
    GW --> RedisCache
    Prometheus -->|Scrape /metrics| GW
    Grafana --> Prometheus
```

---

## 3. 🧩 Level 3: Component Diagram (Clean Architecture)

Tất cả các microservices trong hệ thống EOMP đều tuân thủ kiến trúc **Clean Architecture (Onion Architecture)** phân tầng nghiêm ngặt:

```mermaid
graph TD
    subgraph External Drivers
        HTTPReq[HTTP / REST Client Request]
        BrokerReq[RabbitMQ Event Consumer]
    end

    subgraph Presentation / Interface Layer
        Middleware[Middlewares Stack<br/>• RequestLogger<br/>• Recoverer<br/>• HTTPMetricsMiddleware<br/>• RequireRoles<br/>• ExtractGatewayHeaders]
        Handler[REST Handlers<br/>• Decode JSON Payload<br/>• Validate Input DTO<br/>• Invoke Service<br/>• Encode Response Envelope]
    end

    subgraph Application / Business Logic Layer
        Service[Domain Service Implementation<br/>• Business Rules & State Machine<br/>• Transaction Coordination<br/>• Data Masking Engine<br/>• Cryptographic Hash Calculation]
        ServiceContract[Service Interface]
    end

    subgraph Persistence / Data Access Layer
        RepoContract[Repository Interface]
        PostgresRepo[PostgreSQL Repository<br/>• Parameterized SQL Queries<br/>• Connection Pooling<br/>• In-Memory Mock Fallback]
    end

    subgraph Enterprise Core
        Model[Domain Entities & Value Objects]
    end

    HTTPReq --> Middleware
    Middleware --> Handler
    Handler --> ServiceContract
    ServiceContract --> Service
    Service --> Model
    Service --> RepoContract
    RepoContract --> PostgresRepo
    PostgresRepo --> Model
```

---

## 4. 🔄 Level 4: Enterprise Cross-Service Dynamic Lifecycle Diagram

Sơ đồ tuần tự thể hiện sự phối hợp nhịp nhàng giữa 7 microservices khi thực thi một quy trình kinh doanh hoàn chỉnh:

```mermaid
sequenceDiagram
    autonumber
    actor User as Employee (Kenji)
    actor Mgr as Manager (Sarah)
    actor Agent as IT Specialist (Marcus)
    participant GW as API Gateway (:8080)
    participant Auth as auth-service (:8081)
    participant HD as helpdesk-service (:8084)
    participant WF as workflow-service (:8085)
    participant Notif as notification-service (:8086)
    participant Asset as asset-service (:8083)
    participant Audit as audit-service (:8089)

    User->>GW: POST /api/v1/auth/login (Credentials)
    GW->>Auth: Validate Password & Issue JWT Token
    Auth-->>User: JWT Token (Role: ROLE_EMPLOYEE)

    User->>GW: POST /api/v1/tickets (Request Laptop M3 Max, Priority: HIGH)
    GW->>HD: Create Ticket TK-2026-8801 & Compute SLA (4 Hours)
    HD->>WF: Initiate Approval Workflow WF-9901-HW
    HD-->>User: Ticket Created (Status: OPEN, Pending Approval)

    Mgr->>GW: POST /api/v1/workflows/WF-9901-HW/approve (Role: ROLE_MANAGER)
    GW->>WF: Validate Manager Permission & Update Status to APPROVED
    WF->>Notif: Publish CloudEvent 'eomp.workflow.approved'
    Notif-->>Agent: Push In-App & Email Alert ("Dispatch Hardware Asset")

    Agent->>GW: POST /api/v1/assets/AST-MBP-9901/assign (Assign to Kenji)
    GW->>Asset: Update CMDB Status AVAILABLE -> IN_USE
    Asset-->>Agent: Asset Assigned Successfully

    Agent->>GW: PUT /api/v1/tickets/TK-2026-8801/resolve (Resolution Notes)
    GW->>HD: Mark Ticket RESOLVED (Within SLA: TRUE)
    HD->>Audit: POST /api/v1/audit/logs (Log Action ASSET_ASSIGNMENT_COMPLETED)
    Audit->>Audit: Mask Passwords/Tokens & Compute SHA-256 Checksum Proof
    Audit-->>HD: Audit Log Sealed

    User->>GW: POST /api/v1/tickets/TK-2026-8801/csat (Rating: 5.0 Stars)
    GW->>HD: Record Customer Satisfaction
```
