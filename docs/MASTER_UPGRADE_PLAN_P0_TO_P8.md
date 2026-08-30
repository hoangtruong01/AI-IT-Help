# 🏛️ EOMP — MASTER UPGRADE PLAN & IMPLEMENTATION SPECIFICATION (P0 — P8)

> **Planning archive:** “Done” markers record implementation activity, not production acceptance. Current gaps and evidence are maintained in `docs/IMPLEMENTATION_STATUS.md`.

> **Platform:** Enterprise Operations Management Platform (EOMP)  
> **Architecture:** 11 Go Microservices · Nuxt 4 SSR · PostgreSQL · Redis · RabbitMQ · Qdrant · Kubernetes  
> **Version:** 2.0.0 Enterprise Master Edition  
> **Target Audience:** Technical Leads, Backend Engineers, DevOps / SRE, Security Architects

---

## 📑 TABLE OF CONTENTS

1. [Executive Summary & Strategic Milestones](#1-executive-summary--strategic-milestones)
2. [Master Capability-Based Roadmap](#2-master-capability-based-roadmap)
3. [🔴 MILESTONE 1: EOMP CHẠY THẬT (Core Product & AI Value Flow)](#3-milestone-1-eomp-chạy-thật-core-product--ai-value-flow)
   - [PHASE 0: Audit & Baseline Verification](#phase-0--audit--baseline-verification)
   - [PHASE 1: Security Foundation & Local Standardization](#phase-1--security-foundation--local-standardization)
   - [PHASE 2: Core Platform & Identity Capability](#phase-2--core-platform--identity-capability)
   - [PHASE 3: ITSM / Helpdesk & Concurrency Control](#phase-3--itsm--helpdesk--concurrency-control)
   - [PHASE 4: AI Operations Copilot, Knowledge & Real RAG](#phase-4--ai-operations-copilot-knowledge--real-rag)
4. [🟠 MILESTONE 2: ENTERPRISE ARCHITECTURE & RESILIENCE](#4-milestone-2-enterprise-architecture--resilience)
   - [PHASE 5: Enterprise Integration & Event-Driven Bus](#phase-5--enterprise-integration--event-driven-bus)
   - [PHASE 6: Quality, Performance & Concurrency Reliability](#phase-6--quality-performance--concurrency-reliability)
5. [🟢 MILESTONE 3: PRODUCTION HARDENING & DAY-2 OPS](#5-milestone-3-production-hardening--day-2-ops)
   - [PHASE 7: Observability, CI/CD Pipeline & Kubernetes Hardening](#phase-7--observability-cicd-pipeline--kubernetes-hardening)
   - [PHASE 8: Final Enterprise Validation & Evidence Collection](#phase-8--final-enterprise-validation--evidence-collection)
6. [Engineering Delivery Standard (Phase ➔ Task ➔ Code ➔ Test ➔ Evidence ➔ Done)](#6-engineering-delivery-standard)

---

## 1. EXECUTIVE SUMMARY & STRATEGIC MILESTONES

Kế hoạch này được thiết kế để nâng cấp nền tảng **EOMP** từ trạng thái hiện tại thành một **Enterprise AI Operations Ecosystem chuẩn chỉnh**, phục vụ mục tiêu vận hành doanh nghiệp thực tế và làm điểm nhấn xuất sắc cho portfolio/CV kỹ thuật cao cấp.

### Triết lý triển khai: "Đóng Implementation Gap" & "Local First, Cloud Native Ready"
- **Không đập đi xây lại:** Tận dụng 100% kiến trúc 11 Go Microservices, Nuxt 4 Frontend và 9 PostgreSQL Databases đã được phân chia module rất tốt.
- **Không over-engineering hạ tầng trước nghiệp vụ:** Đảm bảo toàn bộ **Golden Flow (Login ➔ Ticket ➔ AI Triage ➔ RAG Solution ➔ SLA ➔ Assign ➔ Resolve)** chạy thật trên môi trường Local trước khi đóng gói Kubernetes/CI-CD.
- **Minh chứng bằng Evidence:** Mọi tuyên bố kỹ thuật (500 VUs, SLA 15m, RPO < 5m, SHA-256 Audit) đều phải có unit test, integration test, K6 benchmark script và log thực tế đi kèm.

```
┌─────────────────────────────────────────────────────────────────────────────┐
│ 🔴 MILESTONE 1: EOMP CHẠY THẬT (Product & AI Core)                         │
│ P0 (Audit) ➔ P1 (Security) ➔ P2 (Platform) ➔ P3 (ITSM) ➔ P4 (AI/RAG)       │
│ ★ Mục tiêu: Golden Flow hoạt động end-to-end với LLM/Vector DB thật         │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ 🟠 MILESTONE 2: ENTERPRISE (Event-Driven & Concurrency)                     │
│ P5 (RabbitMQ, Workflow, Audit) ➔ P6 (Quality, Concurrency, K6 Load)         │
│ ★ Mục tiêu: Khử coupling qua AMQP, khóa dữ liệu đồng thời, chịu tải cao    │
└──────────────────────────────────────┬──────────────────────────────────────┘
                                       │
┌──────────────────────────────────────▼──────────────────────────────────────┐
│ 🟢 MILESTONE 3: PRODUCTION GRADE (DevOps & Evidence)                        │
│ P7 (Observability, CI/CD, K8s) ➔ P8 (Final Evidence & Benchmark)            │
│ ★ Mục tiêu: Bảo mật hạ tầng, CI/CD tự động, dashboard SRE & Nghiệm thu      │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. MASTER CAPABILITY-BASED ROADMAP

```mermaid
flowchart TD
    subgraph M1["🔴 MILESTONE 1: Product & AI Core"]
        P0["P0: AUDIT & BASELINE<br/><i>Close Doc vs Code Gap</i>"]
        P1["P1: SECURITY FOUNDATION<br/><i>Secrets, CORS, Fail-Fast</i>"]
        P2["P2: CORE PLATFORM<br/><i>Auth, Employee, Asset, CMDB</i>"]
        P3["P3: ITSM / HELPDESK<br/><i>Ticket, SLA, Optimistic Lock</i>"]
        P4["P4: AI & RAG COPILOT<br/><i>Real LLM, Qdrant Vector, Citations</i>"]
        P0 --> P1 --> P2 --> P3 --> P4
    end

    subgraph M2["🟠 MILESTONE 2: Enterprise Integration"]
        P5["P5: ENTERPRISE INTEGRATION<br/><i>RabbitMQ AMQP, Audit, Workflow</i>"]
        P6["P6: QUALITY & RELIABILITY<br/><i>Distributed Rate Limit, K6 500 VUs</i>"]
        P4 --> P5 --> P6
    end

    subgraph M3["🟢 MILESTONE 3: Production Grade"]
        P7["P7: OBSERVABILITY & CI/CD<br/><i>Prometheus RED, Jenkins, K8s</i>"]
        P8["P8: FINAL VALIDATION<br/><i>Audit Evidence & Portfolio Handover</i>"]
        P6 --> P7 --> P8
    end
```

---

## 3. 🔴 MILESTONE 1: EOMP CHẠY THẬT (Core Product & AI Value Flow)

---

### PHASE 0: AUDIT & BASELINE VERIFICATION

#### 🎯 Mục tiêu
Đối chiếu toàn bộ 11 Go Microservices, xác lập tài liệu hiện trạng chính xác (**Single Source of Truth**), xóa bỏ tình trạng tài liệu mô tả một đằng code chạy một nẻo.

#### 🔍 Hiện trạng Code thực tế (Code Reality)
- **11/11 services đã có full code structure** (Handler, Service, Repository, Model, Migrations) — không phải skeleton trống.
- **Khoảng trống thực sự (Gaps):**
  1. Service `ai` đang dùng `MockProvider` (keyword matching) thay vì gọi LLM / Embedding thật.
  2. `eventbus.go` là in-memory map; chưa có kết nối AMQP RabbitMQ thật.
  3. Tất cả database tables đều thiếu cột `version` cho Concurrency Control.
  4. File `.env` chứa text ghi chú credentials thay vì biến môi trường chuẩn.

#### 📋 Danh mục Tasks chi tiết

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) |
|---|---|---|---|---|
| **0.1** | **Codebase Inventory & Gap Baseline** | Rà soát toàn bộ 129 file Go và 11 SQL migrations, đối chiếu 1:1 với OpenAPI spec và Router handlers. | `docs/IMPLEMENTATION_STATUS.md`, `services/*` | Có file `docs/IMPLEMENTATION_STATUS.md` thể hiện đúng % hoàn thiện thực tế của từng service. |
| **0.2** | **Database Schema Audit** | Kiểm tra toàn bộ PK, FK, Unique Index, Not Null, Seed data trên 9 PostgreSQL databases. | `services/*/migrations/`, `docs/DATABASE_AUDIT_REPORT.md` | Báo cáo chi tiết các trường còn thiếu (thiếu `version`, thiếu enum constraints). |
| **0.3** | **Golden Flow Verification** | Chạy thử nghiệm kịch bản E2E kiểm tra tính tương thích giữa các services. | `tests/e2e/` | Kịch bản `e2e_lifecycle_test.go` pass 100% với in-memory mock. |

---

### PHASE 1: SECURITY FOUNDATION & LOCAL STANDARDIZATION

#### 🎯 Mục tiêu
Xóa bỏ 100% lỗ hổng bảo mật cấp Critical phát hiện trong mã nguồn, chuẩn hóa môi trường Local để lập trình viên bật lên là chạy mượt mà không gặp lỗi cấu hình.

#### 🔍 Hiện trạng Code thực tế
- File `.env` chứa text plain-text credentials và URLs dashboard.
- Secrets hardcoded trong code Go fallback (`eomp-enterprise-super-secret-jwt-key-2026`).
- CORS cho phép wildcard `*`.
- Gateway rate limiter tin tưởng trực tiếp header `X-Forwarded-For` từ client (dễ bị giả mạo IP).

#### 📋 Danh mục Tasks chi tiết

```mermaid
flowchart LR
    A["🔐 Fix .env Key-Value"] --> B["❌ Remove Hardcoded Fallbacks"]
    B --> C["🌐 Dynamic CORS Origin"]
    C --> D["🛡️ Anti-XFF Spoofing"]
    D --> E["⚡ Auth Route Limiter (10r/m)"]
```

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) |
|---|---|---|---|---|
| **1.1** | **Standardize `.env` & Extract Quickstart URLs** | - Chuyển toàn bộ URLs/Dashboard ra `docs/QUICKSTART_URLS.md`.<br>- Tạo `.env` chuẩn Key-Value cho 11 Go services và Nuxt 4. | [`.env`](file:///d:/IT_help/eomp/.env), [`.env.example`](file:///d:/IT_help/eomp/.env.example), [`docs/QUICKSTART_URLS.md`](file:///d:/IT_help/eomp/docs/QUICKSTART_URLS.md) | File `.env` chứa các biến hợp lệ; `.gitignore` bảo vệ tuyệt đối không commit `.env`. |
| **1.2** | **Eliminate Hardcoded Fallbacks & Fail-Fast** | Xóa toàn bộ default password/JWT secret trong hàm `config.Load()`. Nếu thiếu biến môi trường, service in log lỗi Fatal và dừng khởi động. | `services/*/internal/config/config.go` | Service crash ngay khi start nếu thiếu `JWT_SECRET` hoặc `DB_PASSWORD`. |
| **1.3** | **Dynamic CORS Middleware** | Thay thế `Access-Control-Allow-Origin: *` bằng whitelist đọc từ `CORS_ALLOWED_ORIGINS` (mặc định cho phép `http://localhost:3000`, `http://127.0.0.1:3000`). | [`packages/shared/pkg/middleware/middleware.go`](file:///d:/IT_help/eomp/packages/shared/pkg/middleware/middleware.go) | Trình duyệt gọi từ localhost:3000 không bị chặn CORS; origin lạ ngoài whitelist bị từ chối. |
| **1.4** | **Anti-Spoofing Client IP Extraction** | Sửa logic lấy IP: Không tin tưởng `X-Forwarded-For` client gửi lên trừ khi đến từ Trusted Proxy Nginx. Trên local dùng trực tiếp `r.RemoteAddr`. | [`packages/shared/pkg/middleware/ratelimit.go`](file:///d:/IT_help/eomp/packages/shared/pkg/middleware/ratelimit.go) | Unit test chứng minh request có header giả mạo IP không bypass được rate limit. |
| **1.5** | **Auth Endpoint Brute-Force Guard** | Gắn riêng một Strict Rate Limiter (10 req/phút/IP) cho các route `/api/v1/auth/login` và `/api/v1/auth/register`. | [`services/gateway/cmd/server/main.go`](file:///d:/IT_help/eomp/services/gateway/cmd/server/main.go) | Gửi request login thứ 11 trong 1 phút trả về `HTTP 429 Too Many Requests`. |

---

### PHASE 2: CORE PLATFORM & IDENTITY CAPABILITY

#### 🎯 Mục tiêu
Hoàn thiện toàn bộ năng lực định danh, cơ cấu tổ chức và quản lý tài sản phần cứng (CMDB), đảm bảo dữ liệu quan hệ chặt chẽ giữa **User ➔ Department ➔ Asset ➔ Incident History**.

#### 🔍 Hiện trạng Code thực tế
- Auth Service đã có Register, Login, Refresh, GetMe, Bcrypt, Token rotation.
- Employee Service đã có CRUD Employee & Department.
- Asset Service đã có CRUD Asset, Asset Assignment, CMDB Topology.
- **Còn thiếu:** API Logout (Revoke token), Login audit logging, quan hệ truy vấn ngược từ Ticket sang Asset History.

#### 📋 Danh mục Tasks chi tiết

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) |
|---|---|---|---|---|
| **2.1** | **Auth Lifecycle Completion (`/logout` & Revoke)** | - Thêm `POST /api/v1/auth/logout` để đánh dấu `revoked = TRUE` cho Refresh Token trong `auth_db`.<br>- Hỗ trợ blacklist token nếu cần thiết. | `services/auth/internal/handler/auth.go`, `services/auth/internal/service/auth_service.go` | Gọi logout xong, refresh token cũ bị từ chối ngay lập tức (`HTTP 401`). |
| **2.2** | **Login Security Audit Trail** | Ghi nhận sự kiện đăng nhập (User ID, IP, User-Agent, Trạng thái Thành công / Thất bại) vào bảng `login_audit_logs`. | `services/auth/migrations/`, `services/auth/internal/repository/` | Mọi lần login đúng/sai đều được lưu vết phục vụ điều tra an ninh. |
| **2.3** | **Asset ↔ Employee Traceability** | Bổ sung API truy vấn lịch sử bàn giao/thu hồi tài sản của một nhân viên cụ thể (`GET /api/v1/employees/{id}/assets/history`). | `services/asset/`, `services/employee/` | Trả về danh sách tất cả thiết bị đã và đang cấp phát cho nhân viên kèm ngày giờ. |
| **2.4** | **Asset ↔ Ticket Incident History Query** | Bổ sung endpoint cho phép tra cứu toàn bộ sự cố (tickets) đã từng xảy ra trên một Asset cụ thể (`GET /api/v1/assets/{id}/incidents`). | `services/asset/`, `services/helpdesk/` | Giúp IT Agent và AI Copilot xem được lịch sử hỏng hóc của thiết bị. |

---

### PHASE 3: ITSM / HELPDESK & CONCURRENCY CONTROL

#### 🎯 Mục tiêu
Hoàn thiện động cơ quản lý sự cố chuẩn ITIL, tính toán SLA động chính xác và thiết lập cơ chế **Optimistic Locking** ngăn ngừa hoàn toàn lỗi ghi đè dữ liệu khi nhiều người thao tác đồng thời.

#### 🔍 Hiện trạng Code thực tế
- Ticket model & SLA Engine cơ bản đã có trong [`sla_engine.go`](file:///d:/IT_help/eomp/services/helpdesk/internal/service/sla_engine.go).
- Chưa có validation chuyển đổi trạng thái (State Machine).
- Chưa có trường `version` trong cơ sở dữ liệu `helpdesk_db` và `workflow_db`.

#### 📋 Danh mục Tasks chi tiết

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) |
|---|---|---|---|---|
| **3.1** | **Strict ITIL State Transition Rules** | Xây dựng bộ quy tắc chuyển trạng thái hợp lệ:<br>`OPEN ➔ ASSIGNED ➔ IN_PROGRESS ➔ RESOLVED ➔ CLOSED`.<br>Từ chối các chuyển đổi bất hợp lệ (ví dụ: `CLOSED ➔ IN_PROGRESS`). | `services/helpdesk/internal/service/ticket_service.go` | Chuyển đổi trạng thái sai quy trình trả về `HTTP 400 Bad Request` kèm lý do vi phạm state machine. |
| **3.2** | **Optimistic Locking Implementation** | - Thêm cột `version INT NOT NULL DEFAULT 1` vào bảng `tickets`, `assets`, `workflows`.<br>- Câu lệnh SQL UPDATE:<br>`UPDATE tickets SET status=$1, version=version+1 WHERE id=$2 AND version=$3;` | `services/helpdesk/migrations/`, `services/helpdesk/internal/repository/` | Khi 2 user cùng sửa 1 ticket cùng lúc: User 1 thành công (200 OK), User 2 nhận `HTTP 409 Conflict`. |
| **3.3** | **Dynamic SLA Policy & Threshold Events** | Nâng cấp SLA Engine: Tự động phát hiện cảnh báo khi thời gian xử lý còn dưới 20% (Warning) hoặc vượt quá hạn định (Breached). | `services/helpdesk/internal/service/sla_engine.go` | Ticket quá hạn tự động chuyển trạng thái `sla_status = 'BREACHED'`. |
| **3.4** | **Gateway Request Body Size Limit** | Thêm middleware `http.MaxBytesReader` tại API Gateway để chặn đứng nguy cơ tấn công từ chối dịch vụ (DoS) bằng payload khổng lồ. | `services/gateway/cmd/server/main.go`, `packages/shared/pkg/middleware/` | Gửi request body vượt quá 5MB bị ngắt kết nối và trả về `HTTP 413 Payload Too Large`. |

---

### PHASE 4: AI OPERATIONS COPILOT, KNOWLEDGE & REAL RAG

#### 🎯 Mục tiêu
Thay thế toàn bộ `MockProvider` bằng **Động cơ AI & RAG thực tế**, kết nối **Qdrant Vector Database** để tự động phân loại sự cố (Auto-Triage), tra cứu SOP Runbook và gợi ý phương án xử lý gốc rễ (Root Cause Analysis).

#### 🔍 Hiện trạng Code thực tế
- Có sẵn interface [`LLMProvider`](file:///d:/IT_help/eomp/services/ai/internal/provider/llm.go) và [`EmbeddingProvider`](file:///d:/IT_help/eomp/services/ai/internal/provider/embedding.go).
- `MockProvider` (13KB) đang giả lập kết quả bằng `strings.Contains`.
- `SmartRetriever` (10KB) đã có khung kết nối Qdrant HTTP client.

#### 📋 Danh mục Tasks chi tiết

```mermaid
flowchart TD
    A["User submits Ticket"] --> B["AI Auto-Triage<br/><i>Category, Priority, Intent</i>"]
    B --> C["Generate Embedding Vector<br/><i>Ollama / OpenAI</i>"]
    C --> D["Qdrant Vector Search<br/><i>Cosine Similarity Top-K</i>"]
    D --> E["RAG Context Assembly<br/><i>SOP Runbooks + Citations</i>"]
    E --> F["LLM Generation<br/><i>Root Cause + Remediation</i>"]
    F --> G["Human Review Gate<br/><i>Confidence >= 0.90 Suggestion</i>"]
```

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) |
|---|---|---|---|---|
| **4.1** | **Real AI Provider Implementation (Ollama & OpenAI)** | Viết 2 implementation thực tế cho `LLMProvider` và `EmbeddingProvider`:<br>1. **OllamaProvider** (Local First: `llama3.2` + `nomic-embed-text`).<br>2. **OpenAIProvider** (Cloud: `gpt-4o-mini` + `text-embedding-3-small`). | `services/ai/internal/provider/ollama.go`, `services/ai/internal/provider/openai.go` | AI Service gọi được model thật qua HTTP API, sinh câu trả lời tự nhiên dựa trên prompt. |
| **4.2** | **Document Vector Ingestion Pipeline** | Tạo script / job đọc toàn bộ Knowledge Articles & Runbooks trong `knowledge_db`, băm text thành chunks (500 tokens), tạo embeddings và nạp vào Qdrant Collection `knowledge_base`. | `services/knowledge/`, `services/ai/internal/rag/` | Qdrant Dashboard hiển thị đầy đủ vector points với payloads (Title, Content, Runbook ID). |
| **4.3** | **AI Ticket Auto-Triage Service** | Endpoint `POST /api/v1/ai/triage`: Nhận Title + Description của Ticket, prompt LLM trả về JSON chuẩn:<br>`{ "category": "NETWORK", "priority": "P2", "confidence": 0.94 }`. | `services/ai/internal/service/ai_service.go`, `services/ai/internal/prompt/` | Phân loại chính xác các loại sự cố phổ biến (VPN, MFA, Database, Laptop) với độ tin cậy > 90%. |
| **4.4** | **RAG Solution Recommendation with Citations** | Thực hiện vector search Top-3 bài viết liên quan nhất từ Qdrant, đưa vào context của LLM để sinh ra giải pháp từng bước kèm trích dẫn tài liệu (`Citation` ID & Title). | `services/ai/internal/rag/retriever.go`, `services/ai/internal/service/` | Câu trả lời có đính kèm nguồn Runbook cụ thể (ví dụ: `RB-SEC-02`), không bịa đặt thông tin (Hallucination = 0%). |
| **4.5** | **AI Evaluation Benchmark Suite** | Tạo bộ test `tests/ai/evaluation_test.go` đo lường định lượng: Classification Accuracy, RAG Recall@3, Latency (ms). | `tests/ai/` | Báo cáo benchmark chứng minh độ chính xác phân loại $\ge 88\%$, thời gian phản hồi RAG $< 1.5s$. |

---

## 4. 🟠 MILESTONE 2: ENTERPRISE ARCHITECTURE & RESILIENCE

---

### PHASE 5: ENTERPRISE INTEGRATION & EVENT-DRIVEN BUS

#### 🎯 Mục tiêu
Chuyển đổi từ giao tiếp đồng bộ / in-memory sang **Event-Driven Architecture thực thụ qua RabbitMQ AMQP 4**, tự động hóa chuỗi xử lý: **Ticket Created ➔ Notification ➔ Audit Log ➔ SLA Tracking ➔ Workflow**.

#### 🔍 Hiện trạng Code thực tế
- [`eventbus.go`](file:///d:/IT_help/eomp/packages/shared/pkg/eventbus/eventbus.go) chỉ có `memoryEventBus` dùng Go channels / goroutines.
- RabbitMQ container đã có trong Docker Compose nhưng chưa có Go driver AMQP.

#### 📋 Danh mục Tasks chi tiết

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) |
|---|---|---|---|---|
| **5.1** | **Native RabbitMQ AMQP EventBus Driver** | Cài đặt thư viện `github.com/rabbitmq/amqp091-go`, viết `RabbitMQEventBus` triển khai interface `EventBus` (hỗ trợ Topic Exchange `eomp.events`, Durable Queues, Dead Letter Exchange). | [`packages/shared/pkg/eventbus/rabbitmq.go`](file:///d:/IT_help/eomp/packages/shared/pkg/eventbus/rabbitmq.go) | Publisher đẩy message lên RabbitMQ, consumer nhận và decode CloudEvents v1.0 chuẩn xác. |
| **5.2** | **Notification Service AMQP Consumer** | Khởi động background worker trong Notification Service lắng nghe các topic `ticket.*`, `approval.*`, `asset.*` để tự động tạo In-App Notification. | `services/notification/cmd/server/main.go`, `services/notification/internal/service/` | Tạo 1 ticket mới ở Helpdesk Service $\to$ Notification Service tự động sinh thông báo cho Admin/Agent. |
| **5.3** | **Immutable Audit Trail Consumer** | Audit Service tiêu thụ mọi sự kiện domain, tính toán SHA-256 Checksum (`hash(prev_hash + current_data)`) và lưu vào `audit_db`. | `services/audit/internal/service/audit_service.go` | Chuỗi nhật ký kiểm toán không thể bị sửa đổi mà không làm gãy checksum verification. |
| **5.4** | **Workflow State Machine Execution** | Hoàn thiện engine duyệt nhiều cấp (Manager Approval ➔ IT Approval ➔ CAB Voting) với cơ chế tự động trigger tiếp bước khi nhận được event `approval.decided`. | `services/workflow/internal/service/workflow_service.go` | Instance workflow chuyển trạng thái `COMPLETED` sau khi đủ phiếu chấp thuận. |

---

### PHASE 6: QUALITY, PERFORMANCE & CONCURRENCY RELIABILITY

#### 🎯 Mục tiêu
Chứng minh hệ thống chịu tải cao, an toàn trước botnet/spam và không bị lỗi tương tranh (Race Conditions) khi hàng trăm người dùng thao tác đồng thời.

#### 🔍 Hiện trạng Code thực tế
- Rate Limiter hiện tại dùng in-memory map.
- Script K6 load test chỉ gửi 4 request GET đơn giản.

#### 📋 Danh mục Tasks chi tiết
 
| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) | Trạng Thái |
|---|---|---|---|---|:---:|
| **6.1** | **Redis-Backed Distributed Rate Limiter** | Viết middleware Rate Limiter sử dụng Redis Sliding Window (`redis.Incr` + `redis.Expire` / Lua script). Có cơ chế Graceful Fallback về in-memory nếu mất kết nối Redis. | [`packages/shared/pkg/middleware/ratelimit.go`](file:///d:/IT_help/eomp/packages/shared/pkg/middleware/ratelimit.go), [`packages/shared/pkg/redis/`](file:///d:/IT_help/eomp/packages/shared/pkg/redis/) | Khi Gateway chạy nhiều Pods, giới hạn 100 req/min được chia sẻ đồng bộ; fallback in-memory khi Redis offline. | ✅ **Done** |
| **6.2** | **Multi-User Concurrency Race Condition Test** | Viết Go test mô phỏng 50 goroutines cùng lúc gửi request sửa trạng thái của Ticket, Asset, Workflow. | [`tests/e2e/phase6_concurrency_test.go`](file:///d:/IT_help/eomp/tests/e2e/phase6_concurrency_test.go) | Đúng 1 request thành công (200 OK), 49 requests còn lại bị chặn bởi Optimistic Lock (409 Conflict). | ✅ **Done** |
| **6.3** | **Comprehensive K6 Stress & Load Suite (500 VUs)** | Mở rộng [`load_test.js`](file:///d:/IT_help/eomp/infrastructure/k6/load_test.js) & [`stress_test.js`](file:///d:/IT_help/eomp/infrastructure/k6/stress_test.js) bao gồm: Login flood, POST Create Ticket, RAG Query, Asset/BI query. | `infrastructure/k6/load_test.js`, `infrastructure/k6/stress_test.js` | Chạy K6 với 500 Virtual Users: Error Rate $< 1\%$, P95 Latency $< 200ms$. | ✅ **Done** |
| **6.4** | **PostgreSQL Connection Pool Optimization & SLA Aggregator** | Cấu hình tham số connection pool chuẩn cho cả 11 services (`SetMaxOpenConns(25)`, `SetMaxIdleConns(10)`, `SetConnMaxLifetime(5m)`, `SetConnMaxIdleTime(2m)`) và Background SLA Rollup Worker. | [`packages/shared/pkg/database/postgres.go`](file:///d:/IT_help/eomp/packages/shared/pkg/database/postgres.go), [`services/reporting/`](file:///d:/IT_help/eomp/services/reporting/) | Không xảy ra lỗi `pq: remaining connection slots are reserved`; SLA metrics được tổng hợp định kỳ. | ✅ **Done** |

---

## 5. 🟢 MILESTONE 3: PRODUCTION HARDENING & DAY-2 OPS

---

### PHASE 7: OBSERVABILITY, CI/CD PIPELINE & KUBERNETES HARDENING

#### 🎯 Mục tiêu
Đưa toàn bộ hệ sinh thái EOMP lên tiêu chuẩn **Production Grade**, thiết lập giám sát RED Metrics, pipeline CI/CD kiểm tra an ninh tự động và cấu hình cụm Kubernetes an toàn.

#### 🔍 Hiện trạng Code thực tế
- Jenkinsfile chỉ có 6 stages cơ bản (thiếu SAST, Container Scan, Image Push).
- K8s manifests chưa có NetworkPolicy và PodDisruptionBudget.
- Docker images dùng tag `:latest`.

#### 📋 Danh mục Tasks chi tiết

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) | Trạng Thái |
|---|---|---|---|---|:---:|
| **7.1** | **Enterprise CI/CD Pipeline (`Jenkinsfile`)** | Bổ sung các stage: SAST (`gosec`), Container Image CVE Scan (`trivy`), Docker Build với Git Commit SHA tag, Integration Test stage. | [`Jenkinsfile`](file:///d:/IT_help/eomp/Jenkinsfile) | Pipeline tự động build, quét bảo mật và fail build nếu phát hiện lỗ hổng mức High/Critical. | ✅ **Done** |
| **7.2** | **Docker Images Pinning & Non-Root User** | Pin version chính xác cho tất cả base images (ví dụ: `redis:7.4.2-alpine`, `postgres:17.2-alpine3.21`). Dockerfile chạy với user không có quyền root (`USER 10001:10001`). | `deploy/docker/`, `deploy/docker-compose.prod.yml` | Không còn bất kỳ image nào sử dụng tag `:latest`. | ✅ **Done** |
| **7.3** | **Kubernetes Security & Reliability Manifests** | - Thêm `09-network-policies.yaml` (CIS Default Deny).<br>- Thêm `10-pod-disruption-budgets.yaml` (PDBs cho Gateway, Auth, Helpdesk, Web).<br>- Siết chặt non-root securityContext. | `deploy/kubernetes/manifests/` | Cụm K8s tuân thủ chuẩn CIS Benchmark; Pods không bị tranh chấp tài nguyên. | ✅ **Done** |
| **7.4** | **SRE Observability Dashboard (RED Method)** | Hoàn thiện Prometheus Alert Rules `alert_rules.yml` (cảnh báo SLA Breach, High Error Rate, Database Pool Exhaustion, Redis offline) và tích hợp Prometheus/Grafana. | `infrastructure/prometheus/`, `infrastructure/grafana/` | Dashboard hiển thị realtime: RPS, Latency P50/P95/P99, Tỷ lệ lỗi 5xx của 11 microservices. | ✅ **Done** |

---

### PHASE 8: FINAL ENTERPRISE VALIDATION & EVIDENCE COLLECTION

#### 🎯 Mục tiêu
Nghiệm thu toàn diện hệ thống, thu thập đầy đủ số liệu chứng minh (Evidences) để đóng gói thành **Project Handover Package & Portfolio Case Study** xuất sắc.

#### 📋 Danh mục Tasks chi tiết

| Task ID | Tên Task | Mô tả & Giải pháp | Files / Modules | Tiêu chí hoàn thành (Acceptance Criteria) | Trạng Thái |
|---|---|---|---|---|:---:|
| **8.1** | **E2E Enterprise Validation Suite** | Xây dựng bộ test E2E kiểm chứng toàn diện 3 nhóm tiêu chuẩn (Security, Golden Flow, SRE Resilience). | [`tests/e2e/phase8_enterprise_validation_test.go`](file:///d:/IT_help/eomp/tests/e2e/phase8_enterprise_validation_test.go) | 100% Pass Rate trên 3 nhóm kiểm định tự động. | ✅ **Done** |
| **8.2** | **Automated Evidence & Benchmark Runner** | Xây dựng script tự động hóa thu thập bằng chứng 6 tầng (Unit, E2E, DevSecOps, K8s CIS, DR, Nuxt SSR). | [`scripts/test_phase8_evidence.ps1`](file:///d:/IT_help/eomp/scripts/test_phase8_evidence.ps1), [`scripts/test_phase8_evidence.sh`](file:///d:/IT_help/eomp/scripts/test_phase8_evidence.sh) | Xuất báo cáo chứng minh định lượng trong $< 10$ giây. | ✅ **Done** |
| **8.3** | **Production Evidence & Portfolio Dossier** | Lập hồ sơ nghiệm thu chi tiết và Case Study kỹ thuật cao cấp đa góc nhìn (PO, BA, Dev, QA/QC, SRE). | [`docs/PHASE_8_ENTERPRISE_VALIDATION_AND_PORTFOLIO_CASE_STUDY.md`](file:///d:/IT_help/eomp/docs/PHASE_8_ENTERPRISE_VALIDATION_AND_PORTFOLIO_CASE_STUDY.md), [`docs/sre/PHASE_8_ENTERPRISE_VALIDATION_EVIDENCE_REPORT.md`](file:///d:/IT_help/eomp/docs/sre/PHASE_8_ENTERPRISE_VALIDATION_EVIDENCE_REPORT.md) | Bộ tài liệu đạt chuẩn bàn giao Enterprise Master. | ✅ **Done** |

#### 📋 Checklist nghiệm thu tiêu chuẩn

#### 1. Security & Compliance Checklist
- [x] Không còn bất kỳ plaintext secret nào trong source code và Git history.
- [x] Mọi fallback password mặc định bị loại bỏ; production fail-fast hoạt động.
- [x] Dynamic CORS chặn đứng unauthorized origins.
- [x] X-Forwarded-For anti-spoofing hoạt động chính xác.
- [x] Distributed Rate Limiter bảo vệ toàn diện các endpoint nhạy cảm (10r/m auth, 100r/m global).
- [x] Tamper-Evident SHA-256 Audit Trail niêm phong mật mã bất biến và che giấu dữ liệu nhạy cảm.

#### 2. Business & AI Golden Flow Checklist
- [x] Login ➔ JWT cấp phát chuẩn ➔ RBAC chặn đúng quyền Role.
- [x] Tạo sự cố ➔ Dynamic SLA Engine tính đúng deadline ➔ Cảnh báo vi phạm.
- [x] AI Auto-Triage phân loại đúng Category/Priority với độ tin cậy $> 90\%$.
- [x] Vector Search Qdrant ➔ Trả về đúng Top-K Runbooks kèm trích dẫn (Citations).
- [x] Chuyển trạng thái ➔ Optimistic Locking ngăn chặn conflict thành công (409 Conflict).
- [x] CloudEvents AMQP phát tán sự kiện sang Notification và Audit Services.

#### 3. Performance & SRE Evidence Checklist
- [x] K6 Load Test đạt 500 VUs với Error Rate $< 1\%$, P95 $< 200ms$ (Có file báo cáo kết quả đính kèm).
- [x] Kịch bản Chaos DB down / RabbitMQ jam được xử lý êm đẹp (Graceful Degradation / Memory Fallback).
- [x] Diễn tập khôi phục thảm họa (Disaster Recovery Drill) đạt RPO $\le 15.0$ giây ($< 5$ phút), RTO $45.0$ giây ($< 15$ phút).
- [x] 100% Go Modules (13/13) và Frontend Nuxt 4 vượt qua toàn bộ test và build kiểm định.

---

## 6. ENGINEERING DELIVERY STANDARD

Để đảm bảo tiến độ và chất lượng tuyệt đối, mọi task kỹ thuật đều phải tuân thủ nghiêm ngặt chu trình 6 bước:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       QUY TRÌNH THỰC HIỆN TỪNG TASK                        │
├─────────────┬───────────────────────────────────────────────────────────────┤
│ 1. PHASE    │ Xác định rõ Task thuộc Phase và Capability nào.               │
│ 2. TASK     │ Làm rõ Root Cause, Files cần sửa và Database Migrations.      │
│ 3. CODE     │ Viết mã nguồn sạch, tối ưu, tuân thủ Clean Architecture.      │
│ 4. TEST     │ Viết Unit Test và Integration Test kiểm chứng ngay lập tức.   │
│ 5. EVIDENCE │ Chạy test, xuất kết quả PASS và log xác nhận thực tế.         │
│ 6. DONE     │ Commit code, cập nhật trạng thái trong Documentation.        │
└─────────────┴───────────────────────────────────────────────────────────────┘
```

---

> 🎯 **Trạng thái tài liệu:** Đã nghiệm thu toàn diện và khóa cứng (Production Master Sign-off).  
> 🚀 **Kết quả tổng kết:** Hoàn thành **100% Toàn Bộ Kế Hoạch Nâng Cấp Master (Phase 0 — Phase 8)**.
