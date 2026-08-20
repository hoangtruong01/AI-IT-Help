# EOMP — Đặc Tả Chi Tiết Lộ Trình Triển Khai Từ Phase 6 Đến Phase 14

> **Tài liệu Hướng Dẫn & Đặc Tả Kỹ Thuật Đa Vai Trò (Multi-Role Specification Document)**  
> **Dành cho:** Business Analyst (BA), UI/UX Designer (Des), QA/QC Engineer, Manual/Automation Tester & Developers.  
> **Mục tiêu:** Cung cấp bức tranh toàn cảnh, chi tiết nghiệp vụ (User Stories, Acceptance Criteria), hướng dẫn thiết kế giao diện (UI/UX Specs) và kịch bản kiểm thử (Test Matrix) từ **Phase 6 đến Phase 14** nhằm đảm bảo toàn bộ đội ngũ phát triển đồng nhất 100%.

---

## 📑 MỤC LỤC TỔNG QUAN

| Phase | Tên Giai Đoạn (Module Scope) | Dịch Vụ Phụ Trách | Trọng Tâm Nghiệp Vụ & Kỹ Thuật |
|:---:|---|---|---|
| **Phase 6** | **AI Operations Copilot & RAG Engine** | `services/ai`, `services/knowledge` | Trợ lý AI IT Ops, Tìm kiếm ngữ nghĩa Qdrant, RAG Runbooks, Phân loại Ticket tự động |
| **Phase 7** | **ITIL Problem & Change Management (CAB)** | `services/helpdesk`, `services/workflow` | Quản lý vấn đề (RCA, Known Errors, Workaround), Hội đồng CAB, Ma trận rủi ro |
| **Phase 8** | **Enterprise Observability & Tracing** | `services/gateway`, Toàn bộ 11 Services | Prometheus Metrics, Grafana Dashboard (RED Method), Loki Log Aggregation |
| **Phase 9** | **Reporting, BI Analytics & SLA Dashboard** | `services/reporting` | Thống kê MTTR/MTTD, Tỷ lệ vi phạm SLA, Hiệu suất kỹ thuật viên, Xuất báo cáo PDF/Excel |
| **Phase 10** | **Security Hardening, RBAC & Audit Trail** | `services/audit`, `packages/shared` | Nhật ký bất biến (Immutable Audit), Che giấu dữ liệu (Data Masking), Rate Limiting, OWASP |
| **Phase 11** | **QA Automation & End-to-End Test Suite** | Toàn bộ hệ thống (`apps/web`, Backend) | Unit Test, Integration Test kịch bản liên service, E2E Testing, K6 Load Testing |
| **Phase 12** | **Technical BA Artifacts & Architecture Blueprints** | `docs/` | Sơ đồ C4 Model, Sơ đồ luồng BPMN, Data Dictionary, OpenAPI 3.0 Specs |
| **Phase 13** | **Production Packaging, Docker & Kubernetes (K8s)** | `deploy/` | Docker Multi-stage (<25MB), K8s Helm Charts, Ingress Nginx, HPA Auto-scaling |
| **Phase 14** | **SRE Operations, DR Runbooks & Handover** | `docs/sre/`, `scripts/` | Kịch bản phục hồi thảm họa (DR), Giả lập sự cố (Chaos test), Bàn giao Production |

---

---

# PHASE 6: AI OPERATIONS COPILOT & RAG ENGINE

## 1. 🎯 Mục Tiêu Nghiệp Vụ (BA Perspective)
Tích hợp trí tuệ nhân tạo (Generative AI / LLM) kết hợp kiến trúc RAG (Retrieval-Augmented Generation) và cơ sở dữ liệu vector (Qdrant) vào quy trình IT Helpdesk nhằm:
1. **Tự động phân loại (Auto-triage)** danh mục và mức độ ưu tiên của Ticket mới.
2. **Gợi ý giải pháp tức thì (Suggested Resolution / Runbooks)** dựa trên kho tài liệu kỹ thuật của doanh nghiệp.
3. **Trợ lý ảo IT Copilot trực tiếp** hỗ trợ nhân viên giải quyết nhanh sự cố mạng, máy tính, tài khoản.

### User Stories & Acceptance Criteria (Gherkin)
```gherkin
Feature: AI Ticket Resolution & Copilot Assistant
  As an IT Support Agent
  I want the AI Copilot to suggest root causes and standard operating procedures (SOP)
  So that I can reduce Mean Time to Resolve (MTTR) by over 40%.

  Scenario: AI automatically analyzes ticket and suggests action
    Given a new incident ticket "Cannot connect to VPN Staging Server" is created
    When the AI service processes the ticket payload
    Then it should classify category as "Network & Access" and priority as "HIGH"
    And it should retrieve top 3 relevant SOP runbooks from Qdrant vector store
    And display a confidence score (e.g., 94%) on the UI.
```

## 2. 🎨 Thiết Kế Giao Diện & Trải Nghiệm (Designer Perspective)
* **Vị trí màn hình:** Trang `/ai` (AI Ops Assistant) và Widget đính kèm trên Drawer chi tiết của `/helpdesk`.
* **Giao diện Chat Glassmorphism:**
  * **Input Bar:** Hộp nhập prompt có nút đính kèm tài liệu/ảnh log lỗi, phím tắt `Ctrl + Enter` để gửi.
  * **Message Bubbles:**
    * *User bubble:* Nền xanh tím đậm `bg-indigo-600/30 border border-indigo-500/40`.
    * *AI bubble:* Nền slate tối `bg-slate-900 border border-slate-800`, icon Robot phát sáng `text-cyan-400 animate-pulse`.
    * *Source References Pill:* Thẻ hiển thị nguồn tài liệu RAG đã trích dẫn (ví dụ: `Doc: WireGuard VPN SOP v2.1 (Score: 0.95)`).
* **Trạng thái UI:**
  * *Streaming State:* Hiệu ứng gõ chữ (Typewriter effect) từng token mượt mà.
  * *Action Buttons:* Nút `Apply Solution to Ticket`, `Copy Code`, `Regenerate`.

## 3. 🧪 Kịch Bản Kiểm Thử (QA/QC & Tester Perspective)
* **Test Case 6.1 (Positive):** Gửi câu hỏi "How to reset user MFA token?" -> Phản hồi đúng quy trình trong tài liệu nội bộ, trích dẫn đúng bài viết Knowledge Base.
* **Test Case 6.2 (Negative / Fallback):** Khi Qdrant Vector Store mất kết nối -> Hệ thống tự động fallback sang In-Memory Knowledge Search hoặc phản hồi cảnh báo rõ ràng mà không crash Gateway.
* **Test Case 6.3 (Performance):** Thời gian nhận Token đầu tiên (Time To First Token - TTFT) phải đạt `< 800ms`.

---

---

# PHASE 7: ITIL PROBLEM & CHANGE MANAGEMENT (CAB)

## 1. 🎯 Mục Tiêu Nghiệp Vụ (BA Perspective)
Triển khai 2 quy trình chuẩn ITIL v4:
* **Problem Management:** Gom nhóm nhiều Incident có cùng nguyên nhân gốc (Root Cause) thành 1 Problem Record, ghi nhận giải pháp tạm thời (Workaround) và Lỗi đã biết (Known Error Database - KEDB).
* **Change Management & CAB:** Quản lý vòng đời Yêu cầu Thay đổi (RFC - Request for Change) với Hội đồng Duyệt Thay Đổi (CAB), đánh giá rủi ro (Risk Matrix) và lên lịch bảo trì (Maintenance Window).

### Quy Tắc Nghiệp Vụ (Business Rules)
* Khi 1 Problem chuyển sang trạng thái `RESOLVED`, tất cả các Ticket Incident liên kết tự động chuyển trạng thái `RESOLVED` và gửi thông báo cho người tạo.
* Mọi Change loại `EMERGENCY` hoặc `MAJOR` bắt buộc phải có tối thiểu 2 phê duyệt từ thành viên CAB trước khi bắt đầu triển khai.

## 2. 🎨 Thiết Kế Giao Diện (Designer Perspective)
* **Màn hình Quản Lý Vấn Đề (`/problems`):**
  * Bảng danh sách Problem: Mã `PRB-101`, Tiêu đề, Số lượng Incident liên kết (Badge số đếm), Tình trạng (`Under Investigation`, `Workaround Found`, `Known Error`, `Closed`).
  * Tab "Root Cause Analysis (RCA)" dạng Markdown Editor chuyên nghiệp.
* **Màn hình Hội Đồng Phê Duyệt Thay Đổi (`/changes`):**
  * **Risk Matrix Visualizer:** Ma trận 3x3 (Khả năng xảy ra vs Mức độ tác động: Low/Medium/High/Critical) hiển thị màu sinh động.
  * **Change Calendar Gantt:** Sơ đồ timeline hiển thị các khoảng thời gian bảo trì (Downtime Windows), tránh xung đột lịch nâng cấp hệ thống.

## 3. 🧪 Kịch Bản Kiểm Thử (QA/QC Perspective)
* **Test Case 7.1:** Tạo Problem từ 3 Incident trùng lặp -> Verify cả 3 Incident hiển thị đúng trong danh sách liên kết của Problem.
* **Test Case 7.2:** Thử kích hoạt Change Request khi chưa đủ số phiếu phê duyệt CAB -> Hệ thống phải chặn và báo lỗi `403 Forbidden: Insufficient CAB Approvals`.

---

---

# PHASE 8: ENTERPRISE OBSERVABILITY & TRACING

## 1. 🎯 Mục Tiêu Nghiệp Vụ (BA Perspective)
Đảm bảo đội ngũ SRE và IT Manager có khả năng giám sát 24/7 toàn diện trạng thái của 11 Microservices, phát hiện nghẽn mạng (Bottlenecks), rò rỉ bộ nhớ (Memory leaks) và lỗi hệ thống trước khi người dùng phát hiện.

## 2. 🎨 Thiết Kế Giao Diện (Designer Perspective)
* **Màn hình Giám Sát Sức Khỏe Hệ Thống (`/monitoring`):**
  * **Service Grid 11 Khối:** Hiển thị 11 services (Gateway, Auth, Employee, Asset, Helpdesk, Workflow, Notification, Knowledge, AI, Audit, Reporting).
  * Mỗi khối hiển thị: Đèn tín hiệu (Xanh/Đỏ), Uptime (vd: 99.98%), CPU %, Memory RAM MB, Response Time p95 (ms).
  * **Live Log Streamer:** Khung hiển thị log dạng terminal đen với text highlight theo màu (`INFO` xanh lá, `WARN` vàng, `ERROR` đỏ tươi).

## 3. 🧪 Kịch Bản Kiểm Thử (QA/QC Perspective)
* **Test Case 8.1:** Gửi request tới endpoint `/metrics` của từng service -> Trả về chuẩn Prometheus text format (bao gồm `http_requests_total`, `http_request_duration_seconds`).
* **Test Case 8.2:** Giả lập service bị gián đoạn -> Dashboard cập nhật trạng thái `OFFLINE` trong vòng tối đa 5 giây.

---

---

# PHASE 9: REPORTING, BI ANALYTICS & SLA DASHBOARD

## 1. 🎯 Mục Tiêu Nghiệp Vụ (BA Perspective)
Cung cấp báo cáo phân tích kinh doanh thời gian thực (BI & Executive Reporting) cho Giám đốc CNTT (CIO/IT Director):
* **Chỉ số Vận Hành Cốt Lõi:**
  * **MTTD (Mean Time to Detect):** Thời gian trung bình phát hiện sự cố.
  * **MTTR (Mean Time to Resolve):** Thời gian trung bình xử lý sự cố.
  * **SLA Compliance Rate:** Tỷ lệ phần trăm sự cố giải quyết đúng hạn cam kết (>95%).
  * **Top Categories & Problem Areas:** Top danh mục phát sinh lỗi nhiều nhất trong tháng.
  * **Agent Performance Scorecard:** Đánh giá năng suất xử lý của từng nhân viên hỗ trợ.

## 2. 🎨 Thiết Kế Giao Diện (Designer Perspective)
* **Màn hình Báo Cáo Phân Tích (`/reports`):**
  * **Bộ lọc Thời Gian:** Nút chọn nhanh (Hôm nay, 7 ngày qua, Tháng này, Quý này, Custom Range).
  * **Biểu đồ Trực Quan (ApexCharts / Chart.js):**
    * Biểu đồ đường (Line chart): Xu hướng phát sinh Incident vs Đã xử lý theo ngày.
    * Biểu đồ tròn (Donut chart): Tỷ lệ phân bổ theo độ ưu tiên (Urgent, High, Medium, Low).
    * Biểu đồ cột chồng (Stacked Bar): SLA Compliance theo từng phòng ban.
  * **Nút Xuất Dữ Liệu:** Xuất ra file Excel (.xlsx) hoặc báo cáo tổng hợp PDF chuyên nghiệp kèm logo EOMP.

## 3. 🧪 Kịch Bản Kiểm Thử (QA/QC Perspective)
* **Test Case 9.1:** Xuất báo cáo PDF với 10,000 bản ghi -> File PDF tải về trong vòng `< 3 giây`, đầy đủ biểu đồ và số liệu chính xác.
* **Test Case 9.2:** Lọc báo cáo theo ngày không có dữ liệu -> Hiển thị Empty State đẹp mắt, không bị lỗi `NaN` hoặc crash trang.

---

---

# PHASE 10: SECURITY HARDENING, RBAC & AUDIT TRAIL

## 1. 🎯 Mục Tiêu Nghiệp Vụ (BA Perspective)
Xây dựng lớp phòng thủ bảo mật doanh nghiệp chuẩn SOC2 / ISO 27001:
* **Kiểm Soát Truy Cập Chặt Chẽ (Strict RBAC):** 4 vai trò rõ rệt (`ROLE_ADMIN`, `ROLE_MANAGER`, `ROLE_AGENT`, `ROLE_USER`).
* **Audit Trail Bất Biến (Immutable Audit Logs):** Ghi nhận mọi hành vi tạo, sửa, xóa, duyệt, xem dữ liệu nhạy cảm kèm IP, User-Agent, Timestamp.
* **Bảo Vệ Dữ Liệu Nhạy Cảm (Data Masking):** Tự động ẩn số thẻ tín dụng, mật khẩu, API token trong logs và giao diện.
* **Rate Limiting:** Chống tấn công Brute-force và DoS tại API Gateway (100 req/min/IP).

## 2. 🎨 Thiết Kế Giao Diện (Designer Perspective)
* **Màn hình Nhật Ký Kiểm Toán (`/audit`):**
  * Bảng tra cứu log chi tiết: Thời gian, Người thực hiện, Hành động (`LOGIN`, `TICKET_UPDATE`, `ASSET_DELETE`, `APPROVAL`), Địa chỉ IP, Trạng thái (`SUCCESS`, `FORBIDDEN`).
  * Nút "Xem Chi Tiết Diffs": Hiển thị so sánh trước và sau khi thay đổi (Old Value vs New Value dạng code diffs).

## 3. 🧪 Kịch Bản Kiểm Thử (QA/QC Perspective)
* **Test Case 10.1:** Tài khoản User thường cố gọi API Admin `/api/v1/audit/logs` -> Gateway trả về `403 Forbidden`.
* **Test Case 10.2:** Đăng nhập sai quá 5 lần liên tiếp -> Khóa IP tạm thời trong 15 phút kèm mã lỗi `429 Too Many Requests`.

---

---

# PHASE 11: QA AUTOMATION & END-TO-END TEST SUITE

## 1. 🎯 Mục Tiêu Nghiệp Vụ (BA Perspective)
Xây dựng mạng lưới kiểm thử tự động toàn diện nhằm ngăn chặn triệt để hiện tượng hồi quy lỗi (Regression bugs) trong mọi đợt phát hành:
* **Độ bao phủ Unit Test (Code Coverage):** Đạt trên **80%** toàn bộ Backend Go và Frontend Vue.
* **Kiểm Thử Luồng Nghiệp Vụ Hoàn Chỉnh (E2E Flow):**
  1. *User Login* -> Tạo Ticket yêu cầu cấp Laptop.
  2. *SLA Engine* tính toán hạn xử lý.
  3. *Workflow Engine* kích hoạt quy trình phê duyệt gửi tới Manager.
  4. *Manager Login* duyệt yêu cầu -> Notification gửi tới IT Agent.
  5. *Agent* bàn giao tài sản từ kho CMDB -> Ticket hoàn thành (`RESOLVED`).

## 2. 🛠️ Công Nghệ & Kịch Bản Kiểm Thử (QA/QC Perspective)
* **Công cụ:**
  * Backend: `go test -v -cover ./...`
  * Frontend Unit/Component: `vitest` + `@vue/test-utils`
  * End-to-End Test: `Playwright`
  * Tải & Hiệu năng: `k6` (Mô phỏng 500 VUs đồng thời)
* **Tiêu Chí Nghiệm Thu QA:**
  * 100% test cases trong pipeline CI/CD phải vượt qua (Pass) trước khi merge vào nhánh `main`.

---

---

# PHASE 12: TECHNICAL BA ARTIFACTS & ARCHITECTURE BLUEPRINTS

## 1. 🎯 Mục Tiêu & Sản Phẩm Bàn Giao (BA & Designer Perspective)
Soạn thảo bộ hồ sơ tài liệu kiến trúc & nghiệp vụ tiêu chuẩn quốc tế:
1. **C4 Model Diagrams:**
   * Level 1: System Context Diagram.
   * Level 2: Container Diagram (11 Microservices + 5 DBs + Nuxt 4 + Nginx).
   * Level 3: Component Diagram (Clean Architecture layer trong từng service).
2. **Data Dictionary & ERD Master:**
   * Bản đồ thực thể quan hệ đầy đủ của 7 cơ sở dữ liệu phân tán.
3. **OpenAPI 3.0 (Swagger) Hub:**
   * Toàn bộ tài liệu API tương tác trực tiếp (Interactive API Documentation) tại cổng `:8080/swagger/`.

---

---

# PHASE 13: PRODUCTION PACKAGING, DOCKER & KUBERNETES

## 1. 🎯 Mục Tiêu Kỹ Thuật (DevOps & SRE Perspective)
Đóng gói toàn bộ hệ thống sẵn sàng triển khai trên môi trường Production Cloud (AWS / GCP / Bare Metal):
* **Docker Multi-stage Builds:** Image của mỗi Go binary có dung lượng siêu nhẹ (`< 25MB`) trên nền `scratch` hoặc `alpine`, đảm bảo không chứa mã nguồn thừa.
* **Kubernetes Manifests & Helm Chart (`deploy/kubernetes/`):**
  * Cấu hình Deployment, Service, ConfigMap, Secret, PersistentVolumeClaim (PVC).
  * **Health Probes:** Liveness Probe (`/health`) & Readiness Probe (`/api/health`).
  * **HPA (Horizontal Pod Autoscaler):** Tự động scale từ 2 lên 10 pods khi CPU vượt quá 70%.

---

---

# PHASE 14: SRE OPERATIONS, DISASTER RECOVERY & HANDOVER

## 1. 🎯 Mục Tiêu Vận Hành (SRE & QA Perspective)
Hoàn thiện quy trình vận hành tin cậy và bàn giao nền tảng:
1. **Disaster Recovery (DR) Plan:**
   * RPO (Recovery Point Objective) `< 5 phút`.
   * RTO (Recovery Time Objective) `< 15 phút`.
2. **Kịch Bản Giả Lập Sự Cố (Chaos Engineering Runbooks):**
   * Giả lập đứt kết nối Database Postgres -> Tự động kích hoạt cơ chế retry và alert qua Slack/Email.
   * Giả lập nghẽn hàng đợi RabbitMQ -> Đẩy message lỗi vào Dead-Letter Queue (DLQ).
3. **Biên Bản Nghiệm Thu & Bàn Giao Toàn Diện:**
   * Bộ tài liệu hướng dẫn vận hành (Operations Manual), Tài liệu bảo trì (Maintenance Runbooks) và Ký duyệt nghiệm thu kỹ thuật.

---

## 🚀 BẢNG TỔNG HỢP TIẾN ĐỘ TOÀN BỘ 14 PHASES (EOMP MASTER ROADMAP)

| Phase | Phạm Vi Triển Khai | Trạng Thái |
|:---:|---|:---:|
| **Phase 0** | Repository Audit, Codebase Clean-up & Architecture Strategy | **HOÀN THÀNH (Done)** |
| **Phase 1** | Business Foundation, Auth/RBAC, Employee Service & Nuxt 4 Integration | **HOÀN THÀNH (Commit `efa44d6`)** |
| **Phase 2** | Incident Management, Service Catalog & SLA Engine | **HOÀN THÀNH (Commit `adde813`)** |
| **Phase 3** | Asset Management & CMDB Infrastructure Dependency Topology | **HOÀN THÀNH (Commit `7b4428a`)** |
| **Phase 4** | Workflow State Machine Engine, Multi-level Approvals & Audit Logs | **HOÀN THÀNH (Commit `e482adc`)** |
| **Phase 5** | Event-Driven Architecture, CloudEvents EventBus & Notification Service | **HOÀN THÀNH (Commit `e67d9fc`)** |
| **Phase 6** | AI Operations Copilot, Qdrant Vector Search & RAG Knowledge Engine | **HOÀN THÀNH (Commit `d6afacc`)** |
| **Phase 7** | ITIL Problem Management, RCA & Change Advisory Board (CAB) | **HOÀN THÀNH (Commit `7c2eccd`)** |
| **Phase 8** | Enterprise Observability (Prometheus, Grafana RED, Loki Logs) | **HOÀN THÀNH (Done)** |
| **Phase 9** | Business Intelligence Reporting, MTTR/MTTD & SLA Dashboard | **HOÀN THÀNH (Done)** |
| **Phase 10** | Enterprise Security Hardening, Strict RBAC & Immutable Audit Trail | **HOÀN THÀNH (Done)** |
| **Phase 11** | QA Automation Suite (Unit, Integration, Playwright E2E, K6 Load) | **READY TO START (Tiếp theo)** |
| **Phase 12** | Technical BA Artifacts, C4 Model Blueprints & OpenAPI Spec Hub | **SẴN SÀNG TRIỂN KHAI** |
| **Phase 13** | Production Packaging, Docker Multi-stage & Kubernetes Helm Charts | **SẴN SÀNG TRIỂN KHAI** |
| **Phase 14** | SRE Disaster Recovery Runbooks, Chaos Testing & Project Handover | **SẴN SÀNG TRIỂN KHAI** |
