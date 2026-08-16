# EOMP — Tài Liệu Toàn Diện Dành Cho Developer & Intern

> **Enterprise Operations Management Platform (EOMP)**  
> Bản hướng dẫn chi tiết từ A đến Z dành cho Software Engineer / Intern Developer để nắm bắt toàn bộ kiến trúc hệ thống, mã nguồn, hạ tầng, luồng nghiệp vụ và quy trình phát triển.

---

## 📑 Mục Lục

1. [Tổng Quan Dự Án & Bài Toán Nghiệp Vụ](#1-tổng-quan-dự-án--bài-toán-nghiệp-vụ)
2. [Kiến Trúc Tổng Thể Hệ Thống (System Architecture)](#2-kiến-trúc-tổng-thể-hệ-thống-system-architecture)
3. [Bản Đồ Thư Mục & Giải Thích Chi Tiết Cấu Trúc Mã Nguồn](#3-bản-đồ-thư-mục--giải-thích-chi-tiết-cấu-trúc-mã-nguồn)
4. [Bảng Công Nghệ (Technology Stack & Tools)](#4-bảng-công-nghệ-technology-stack--tools)
5. [Hướng Dẫn Cài Đặt Môi Trường Từ Đầu (Local Setup Guide)](#5-hướng-dẫn-cài-đặt-môi-trường-từ-đầu-local-setup-guide)
6. [Ma Trận Cổng & Đường Dẫn Dịch Vụ (Service Port Matrix)](#6-ma-trận-cổng--đường-dẫn-dịch-vụ-service-port-matrix)
7. [Phân Tích Chi Tiết Từng Microservice (Deep-Dive 11 Services)](#7-phân-tích-chi-tiết-từng-microservice-deep-dive-11-services)
8. [Kiến Trúc Frontend (Nuxt 4 / Vue 3 / Tailwind v4)](#8-kiến-trúc-frontend-nuxt-4--vue-3--tailwind-v4)
9. [Tầng Hạ Tầng, Cơ Sở Dữ Liệu & DevOps (Infrastructure Deep-Dive)](#9-tầng-hạ-tầng-cơ-sở-dữ-liệu--devops-infrastructure-deep-dive)
10. [CI/CD Pipeline với Jenkins](#10-cicd-pipeline-với-jenkins)
11. [Quy Chuẩn Code & Best Practices Cho Intern](#11-quy-chuẩn-code--best-practices-cho-intern)
12. [Hướng Dẫn Thực Hành: Xây Dựng Một Tính Năng Mới Từ A-Z](#12-hướng-dẫn-thực-hành-xây-dựng-một-tính-năng-mới-từ-a-z)
13. [Chiến Lược Kiểm Thử (Testing & QA Guide)](#13-chiến-lược-kiểm-thử-testing--qa-guide)
14. [Xử Lý Sự Cố Thường Gặp (Troubleshooting & FAQs)](#14-xử-lý-sự-cố-thường-gặp-troubleshooting--faqs)

---

## 1. Tổng Quan Dự Án & Bài Toán Nghiệp Vụ

### 1.1. EOMP Là Gì?
**EOMP (Enterprise Operations Management Platform)** là nền tảng quản trị vận hành công nghệ thông tin và quy trình doanh nghiệp cấp doanh nghiệp (Enterprise-grade). Hệ thống giải quyết bài toán quản trị tập trung toàn bộ các nghiệp vụ IT nội bộ, từ quản lý nhân sự IT, tài sản phần cứng/phần mềm, hệ thống tiếp nhận và xử lý sự cố (IT Helpdesk), quy trình phê duyệt đa cấp, kho tri thức số hóa, cho tới trợ lý ảo trí tuệ nhân tạo (AI Copilot) hỗ trợ phân loại sự cố tự động.

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    EOMP Enterprise Platform Ecosystem                   │
├───────────────────┬───────────────────┬─────────────────────────────────┤
│ Quản Lý Nhân Sự   │ Quản Lý Tài Sản   │ IT Helpdesk & Xử Lý Sự Cố       │
│ (Employee & Org)  │ (Asset Lifecycle) │ (Ticketing & SLA Escalation)    │
├───────────────────┼───────────────────┼─────────────────────────────────┤
│ Quy Trình Phê Duyệt│ Kho Tri Thức Số   │ AI Operations Copilot           │
│ (Workflow Engine) │ (RAG Knowledge)   │ (Ticket Triage & Chatbot)       │
├───────────────────┼───────────────────┼─────────────────────────────────┤
│ Đa Kênh Thông Báo │ Nhật Ký Tuân Thủ  │ Báo Cáo & Đo Lường Vận Hành     │
│ (Notifications)   │ (Audit Trails)    │ (Reporting & Analytics)         │
└───────────────────┴───────────────────┴─────────────────────────────────┘
```

### 1.2. 10 Phân Hệ Chức Năng Cốt Lõi

| Phân hệ (Module) | Mục tiêu nghiệp vụ | Người dùng chính |
|---|---|---|
| **Employee Management** | Quản lý thông tin nhân viên, phòng ban, cây sơ đồ tổ chức, chức danh và phân quyền. | HR, IT Admin, Manager |
| **Asset Management** | Theo dõi vòng đời tài sản CNTT: Laptop, màn hình, máy chủ, bản quyền phần mềm (License), bảo hành, bàn giao/thu hồi. | IT Asset Manager, Sysadmin |
| **IT Helpdesk & Ticketing** | Cổng tiếp nhận yêu cầu hỗ trợ kỹ thuật, quản trị SLA (cam kết thời gian phản hồi/xử lý), gán kỹ thuật viên, phân loại sự cố. | Nhân viên (End-user), IT Support L1/L2/L3 |
| **Workflow & Approval** | Định nghĩa và thực thi quy trình phê duyệt nghiệp vụ nhiều bước (Cấp phát thiết bị, xin quyền truy cập, nâng cấp server,...). | Team Lead, Trưởng phòng, Ban Giám Đốc |
| **Knowledge Base** | Thư viện tài liệu hướng dẫn, sổ tay IT nội bộ, tích hợp tìm kiếm thông minh dạng ngữ nghĩa (Vector Semantic Search). | Toàn bộ nhân viên, Kỹ thuật viên |
| **Notification Engine** | Trung tâm điều phối thông báo đa kênh: Email (SMTP), Tin nhắn hệ thống (In-App notification), Web Push. | Tự động gửi theo sự kiện hệ thống |
| **AI Operations Assistant** | Trợ lý ảo AI sử dụng RAG (Retrieval-Augmented Generation) & LLM để tự động phân loại ticket, tóm tắt sự cố, gợi ý cách sửa lỗi tức thì. | IT Support, End-user |
| **Reporting & BI** | Thống kê số liệu vận hành, đo lường chỉ số MTTR (Mean Time to Resolve), tỷ lệ đạt SLA, hiệu suất kỹ thuật viên. | IT Director, CIO, Team Lead |
| **Audit Logging** | Ghi nhận nhật ký bất biến (immutable log) mọi hành vi tạo, sửa, xóa, đăng nhập, truy cập dữ liệu nhạy cảm để kiểm toán an ninh. | Compliance Officer, Security Admin |
| **Observability System** | Hệ thống giám sát tài nguyên máy chủ, đo lường chỉ số Prometheus, biểu đồ Grafana, truy vấn tập trung log Loki. | DevOps, SRE, Tech Lead |

### 1.3. Các Vai Trò Trong Hệ Thống (RBAC)
- **Super Admin (`ROLE_ADMIN`)**: Quyền hạn tối cao trên toàn bộ hệ thống, quản lý cấu hình microservices và người dùng.
- **IT Manager (`ROLE_MANAGER`)**: Phê duyệt quy trình cấp phát, xem báo cáo thống kê, phân công kỹ thuật viên.
- **IT Support (`ROLE_AGENT`)**: Tiếp nhận và giải quyết ticket helpdesk, cập nhật trạng thái tài sản, viết bài viết Knowledge Base.
- **Employee (`ROLE_EMPLOYEE`)**: Tạo yêu cầu hỗ trợ, đăng ký mượn/cấp thiết bị, xem bài viết hướng dẫn, chat với AI Copilot.

---

## 2. Kiến Trúc Tổng Thể Hệ Thống (System Architecture)

EOMP áp dụng mô hình kiến trúc **Microservices phân tán** theo nguyên lý **Domain-Driven Design (DDD)** kết hợp **Monorepo**.

### 2.1. Sơ Đồ Kiến Trúc Luồng Dữ Liệu

```
                          ┌───────────────────────────┐
                          │     Người Dùng / Browser  │
                          └─────────────┬─────────────┘
                                        │ HTTPS / HTTP
                                        ▼
                          ┌───────────────────────────┐
                          │   Frontend Web (Nuxt 4)   │
                          │   Vue 3 · Tailwind CSS v4 │
                          │   Port 3000               │
                          └─────────────┬─────────────┘
                                        │ REST API (JSON)
                                        ▼
┌────────────────────────────────────────────────────────────────────────────────────────┐
│                               API GATEWAY SERVICE (:8080)                              │
│         Middleware Stack: CORS · Request Logger · Recoverer · Auth & Rate Limiter      │
└────────┬──────────┬──────────┬──────────┬──────────┬──────────┬──────────┬────────┬────┘
         │          │          │          │          │          │          │        │
         │ :8081    │ :8082    │ :8083    │ :8084    │ :8085    │ :8086    │ :8087  │ :8088
         ▼          ▼          ▼          ▼          ▼          ▼          ▼        ▼
     ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐┌───────┐
     │ Auth  │  │Employ-│  │ Asset │  │ Help- │  │ Work- │  │Notifi-│  │ Know- ││  AI   │
     │Service│  │  ee   │  │Service│  │ desk  │  │ flow  │  │cation │  │ ledge ││Service│
     └───┬───┘  └───┬───┘  └───┬───┘  └───┬───┘  └───┬───┘  └───┬───┘  └───┬───┘└───┬───┘
         │          │          │          │          │          │          │        │
         ▼          ▼          ▼          ▼          ▼          │          ▼        ▼
     ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐  ┌───────┐      │      ┌───────┐┌───────┐
     │auth_db│  │employ-│  │asset_ │  │help-  │  │work-  │      │      │know-  ││Qdrant │
     │       │  │ee_db  │  │db     │  │desk_db│  │flow_db│      │      │ledge  ││Vector │
     └───────┘  └───────┘  └───────┘  └───────┘  └───────┘      │      │_db    ││DB     │
         │          │          │          │          │          │      └───────┘└───────┘
         └──────────┴──────────┴──────────┴──────────┴──────────┴──────────┴────────┤
                                              │                                     │
                                    ┌─────────▼────────┐                            │
                                    │ RabbitMQ Events  │◄───────────────────────────┘
                                    │ (Async EventBus) │
                                    └─────────┬────────┘
                                              │
                      ┌───────────────────────┴───────────────────────┐
                      ▼                                               ▼
              ┌──────────────┐                                ┌──────────────┐
              │Audit Service │(:8089)                         │Reporting Svc │(:8090)
              │ └► audit_db  │                                │ └► MinIO S3  │
              └──────────────┘                                └──────────────┘
```

### 2.2. Các Nguyên Tắc Kiến Trúc Bắt Buộc (Architectural Tenets)

1. **Database-per-Service (Sở hữu dữ liệu độc lập)**:
   - Mỗi microservice sở hữu cơ sở dữ liệu PostgreSQL riêng biệt (ví dụ: `auth_db`, `asset_db`, `helpdesk_db`).
   - **Tuyệt đối cấm**: Service A kết nối trực tiếp vào Database của Service B.
   - Muốn lấy dữ liệu từ service khác, bắt buộc phải gọi qua **REST API / gRPC** hoặc lắng nghe **Domain Event qua RabbitMQ**.

2. **Single Entry Point (Điểm vào duy nhất)**:
   - Frontend không bao giờ gọi trực tiếp vào các microservice nội bộ.
   - Mọi request từ Client đều bắt buộc đi qua **API Gateway (Port 8080)** để thực hiện xác thực, phân quyền, rate limiting, logging correlation ID.

3. **Giao tiếp Đồng Bộ vs Bất Đồng Bộ**:
   - **Đồng bộ (Synchronous)**: Dùng **HTTP REST API** hoặc **gRPC (Protobuf)** khi cần kết quả phản hồi ngay lập tức (ví dụ: Đăng nhập, lấy thông tin cá nhân, kiểm tra trạng thái thiết bị).
   - **Bất đồng bộ (Asynchronous Event-Driven)**: Dùng **RabbitMQ** khi một hành động kích hoạt nhiều tác vụ phụ trợ không cần chặn luồng chính (ví dụ: Khi tạo ticket mới -> bắn event `ticket.created` -> Notification Service gửi email + Audit Service ghi log + AI Service phân tích tự động).

4. **Clean Architecture / Hexagonal Architecture trong từng Go Service**:
   - Mỗi service được cấu trúc chuẩn hóa:
     - `cmd/server/main.go`: Entry point khởi tạo dependency injection và chạy HTTP/gRPC server.
     - `internal/config`: Đọc biến môi trường.
     - `internal/handler`: Nhận request, validate dữ liệu, serialize/deserialize JSON.
     - `internal/service`: Chứa nghiệp vụ kinh doanh cốt lõi (Business Logic).
     - `internal/model`: Định nghĩa struct dữ liệu.
     - `internal/provider` hoặc `internal/repository`: Giao tiếp với database, Redis, S3 hoặc bên thứ ba.

---

## 3. Bản Đồ Thư Mục & Giải Thích Chi Tiết Cấu Trúc Mã Nguồn

Dưới đây là cấu trúc toàn bộ dự án cùng giải thích chi tiết chức năng của từng file/thư mục:

```
d:/IT_help/eomp/
├── .env                         # File chứa biến môi trường thực tế chạy cục bộ (Không commit git)
├── .env.example                 # File mẫu khai báo toàn bộ các biến môi trường cần thiết
├── .gitignore                   # Cấu hình bỏ qua các file rác, binary, node_modules khi push git
├── docker-compose.yml           # Khởi tạo toàn bộ hạ tầng (Postgres, Redis, RabbitMQ, MinIO, Qdrant, Prometheus, Grafana, Loki)
├── go.work                      # Go Multi-Module Workspace quản lý liên kết 12 Go modules
├── Makefile                     # Các câu lệnh quản trị dự án cho Linux / macOS / CI
├── Jenkinsfile                  # Pipeline tự động hóa CI/CD cho Jenkins
├── README.md                    # Tài liệu tổng quan dự án
│
├── apps/                        # Chứa các ứng dụng Client
│   └── web/                     # Ứng dụng Frontend (Nuxt 4 / Vue 3)
│       ├── app/
│       │   ├── app.config.ts    # Cấu hình theme, icon, meta của Nuxt App
│       │   ├── app.vue          # Component gốc của Vue application
│       │   ├── assets/          # CSS, hình ảnh tĩnh, fonts (Tailwind v4 css)
│       │   ├── components/      # UI components tái sử dụng (AppLogo, TemplateMenu,...)
│       │   ├── composables/     # Vue Composables (useAuth, useApi, useToast,...)
│       │   ├── layouts/         # Layout trang (default.vue với Sidebar, Header, Breadcrumbs)
│       │   ├── middleware/      # Nuxt route middleware (auth.ts kiểm tra đăng nhập)
│       │   ├── pages/           # File-based routing (index, helpdesk, assets, employees, ai,...)
│       │   ├── plugins/         # Nuxt plugins (tiện ích mở rộng khởi động cùng app)
│       │   ├── services/        # Client API services gọi về Gateway
│       │   ├── stores/          # Pinia stores quản lý state toàn cục (auth.ts, ui.ts)
│       │   ├── types/           # TypeScript interface & type definitions
│       │   └── utils/           # Helper functions (formatDate, currencyFormat,...)
│       ├── nuxt.config.ts       # Cấu hình chính của Nuxt framework, modules, proxy
│       └── package.json         # Danh sách dependencies của Frontend (pnpm)
│
├── packages/                    # Mã nguồn dùng chung nội bộ (Shared Libraries)
│   ├── proto/                   # Khai báo Protocol Buffers (.proto) cho gRPC
│   │   └── common.proto         # Protobuf chuẩn hóa Pagination, HealthCheck
│   └── shared/                  # Thư viện Go dùng chung cho cả 11 microservices
│       ├── go.mod               # Go module của package shared (`eomp/packages/shared`)
│       └── pkg/
│           ├── config/          # Helper đọc environment variables chuẩn hóa
│           ├── logger/          # Structured JSON Logger chuẩn Go `log/slog`
│           ├── middleware/      # HTTP Middlewares: RequestLogger, Recoverer (chống crash), CORS
│           └── response/        # Chuẩn hóa JSON Response (`response.JSON`, `response.Error`)
│
├── services/                    # 11 Microservices viết bằng Golang
│   ├── ai/                      # AI Assistant, RAG & Ticket Triage (:8088)
│   ├── asset/                   # Quản lý tài sản CNTT & License (:8083)
│   ├── audit/                   # Ghi nhật ký kiểm toán hệ thống (:8089)
│   ├── auth/                    # Xác thực, JWT, cấp quyền RBAC (:8081)
│   ├── employee/                # Nhân sự, phòng ban, sơ đồ tổ chức (:8082)
│   ├── gateway/                 # API Gateway tiếp nhận toàn bộ request (:8080)
│   ├── helpdesk/                # Quản lý sự cố, yêu cầu, SLA (:8084)
│   ├── knowledge/               # Bài viết hướng dẫn & đồng bộ Vector (:8087)
│   ├── notification/            # Gửi Email, WebPush, Thông báo nội bộ (:8086)
│   ├── reporting/               # Báo cáo thống kê, số liệu vận hành (:8090)
│   └── workflow/                # Quy trình phê duyệt đa bước (:8085)
│
├── infrastructure/              # Cấu hình Docker & Script khởi tạo hạ tầng
│   ├── docker/                  # Dockerfiles bổ trợ
│   ├── grafana/                 # Cấu hình datasource & dashboards tự động nạp
│   ├── loki/                    # Cấu hình log retention & storage của Loki
│   ├── minio/                   # Cấu hình khởi tạo buckets lưu file
│   ├── nginx/                   # Reverse proxy configuration
│   ├── postgres/                # Script SQL khởi tạo 7 database độc lập (`01-init-databases.sql`)
│   ├── prometheus/              # Cấu hình scrape endpoints đo lường metric
│   ├── qdrant/                  # Vector DB storage config
│   ├── rabbitmq/                # AMQP queue definitions & plugins
│   └── redis/                   # Cấu hình Redis persistence
│
├── docs/                        # Thư mục chứa toàn bộ tài liệu kỹ thuật của dự án
│   ├── INTERN_DEVELOPER_GUIDE.md# [FILE NÀY] Tài liệu hướng dẫn chi tiết nhất cho intern
│   ├── api.md                   # Đặc tả các API endpoints
│   ├── architecture.md          # Chi tiết thiết kế kiến trúc
│   ├── database.md              # Sơ đồ và quy tắc cơ sở dữ liệu
│   ├── deployment.md            # Hướng dẫn build và deploy
│   ├── development.md           # Quy tắc phát triển và đóng góp code
│   ├── environment.md           # Danh sách toàn bộ biến môi trường
│   ├── setup.md                 # Hướng dẫn cài đặt nhanh
│   └── testing.md               # Chiến lược viết và chạy test
│
└── scripts/                     # Tooling tự động hóa phát triển
    ├── dev.ps1                  # Bộ công cụ PowerShell cho Windows (Build, Test, Lint, Docker)
    └── qa.ps1                   # Bộ script kiểm tra chất lượng tự động toàn diện (QA/QC)
```

---

## 4. Bảng Công Nghệ (Technology Stack & Tools)

| Lớp (Layer) | Công nghệ | Phiên bản | Mục đích sử dụng trong dự án |
|---|---|---|---|
| **Frontend Framework** | **Nuxt 4** (Vue 3) | 4.5.x | SSR/SSG/SPA Hybrid framework, auto-imports, file-based routing |
| **Frontend UI & Styling**| **Tailwind CSS v4** + **Nuxt UI** | v4 / v3 | Design system hiện đại, dark mode, animation, micro-interactions |
| **Frontend State** | **Pinia** | 3.x | Quản lý trạng thái xác thực và dữ liệu toàn cục |
| **Backend Language** | **Golang** | 1.24.x | Ngôn ngữ hiệu năng cao, concurrency tuyệt vời cho 11 microservices |
| **API Gateway** | Go Native HTTP Mux | 1.24 | Reverse proxy, router, correlation ID, rate limit |
| **Relational Database** | **PostgreSQL** | 17 Alpine | Cơ sở dữ liệu quan hệ cho 7 microservices nghiệp vụ |
| **In-Memory Cache** | **Redis** | 7 Alpine | Caching dữ liệu thường truy cập, session store, rate limit token bucket |
| **Message Broker** | **RabbitMQ** | 4 Management | Hàng đợi tin nhắn trao đổi sự kiện bất đồng bộ giữa các services |
| **Object Storage** | **MinIO** | Latest | Lưu trữ file ảnh đính kèm ticket, hợp đồng tài sản, báo cáo PDF |
| **Vector Database** | **Qdrant** | Latest | Lưu trữ vector embeddings phục vụ RAG và Semantic Search |
| **Metrics Monitoring** | **Prometheus** | Latest | Thu thập chỉ số hiệu năng (CPU, Memory, Request Rate, Latency) |
| **Visualization** | **Grafana** | Latest | Hiển thị bảng điều khiển (Dashboards) trực quan giám sát hệ thống |
| **Log Aggregation** | **Grafana Loki** | Latest | Gom toàn bộ log từ các container phục vụ debug và tra cứu |
| **CI/CD Tool** | **Jenkins** | Pipeline as Code | Tự động hóa kiểm tra lint, test, build binary, build Docker image |
| **Package Managers** | **pnpm** (Web) & **Go Modules** | 10.x & 1.24 | Quản lý dependencies tối ưu tốc độ và dung lượng |

---

## 5. Hướng Dẫn Cài Đặt Môi Trường Từ Đầu (Local Setup Guide)

Dành riêng cho intern mới nhận máy tính để thiết lập môi trường phát triển đầy đủ.

### 5.1. Yêu Cầu Cài Đặt Phần Mềm Tiên Quyết (Prerequisites)

1. **Git**: >= 2.40 ([Tải Git](https://git-scm.com/))
2. **Node.js**: >= 20.19 LTS ([Tải Node.js](https://nodejs.org/))
3. **pnpm**: >= 10.x (Cài qua terminal: `npm install -g pnpm`)
4. **Go (Golang)**: >= 1.24.x ([Tải Golang](https://go.dev/dl/))
5. **Docker Desktop**: >= 28.x với Docker Compose >= 2.20 ([Tải Docker Desktop](https://www.docker.com/products/docker-desktop/))
   - *Lưu ý trên Windows*: Bật tính năng WSL 2 (Windows Subsystem for Linux 2).

### 5.2. Các Bước Cài Đặt Từng Bước (Step-by-Step)

#### Bước 1: Clone mã nguồn và cấu hình file môi trường
```bash
# Clone repository về máy
git clone https://github.com/hoangtruong01/AI-IT-Help.git
cd AI-IT-Help

# Tạo file .env từ file mẫu
cp .env.example .env
```

> [!IMPORTANT]
> Mở file `.env` bằng VSCode và kiểm tra các cấu hình mật khẩu mặc định:
> - `POSTGRES_USER=eomp`
> - `POSTGRES_PASSWORD=eomp_dev_password`
> - `RABBITMQ_USER=eomp`
> - `RABBITMQ_PASSWORD=eomp_dev_password`
> - `MINIO_ACCESS_KEY=eomp_minio`
> - `MINIO_SECRET_KEY=eomp_minio_secret`

#### Bước 2: Khởi động toàn bộ Hạ Tầng qua Docker Compose
Mở PowerShell (Windows) hoặc Terminal (Linux/macOS) tại thư mục gốc dự án:

```powershell
# Trên Windows (PowerShell):
.\scripts\dev.ps1 docker-up

# Trên Linux / macOS:
make docker-up
```

#### Bước 3: Kiểm tra sức khỏe toàn bộ hệ thống (Health Check)
Đảm bảo tất cả 8 hạ tầng (Postgres, Redis, RabbitMQ, MinIO, Qdrant, Prometheus, Grafana, Loki) đều hiển thị **OK**:

```powershell
# Trên Windows:
.\scripts\dev.ps1 health

# Trên Linux / macOS:
make health
```

*Kết quả mong đợi:*
```
=== EOMP Infrastructure Health ===
  PostgreSQL    OK
  Redis         OK
  RabbitMQ      OK
  MinIO         OK
  Qdrant        OK
  Prometheus    OK
  Grafana       OK
  Loki          OK
```

#### Bước 4: Cài đặt và Chạy Frontend Web (Nuxt 4)
Mở một cửa sổ Terminal mới:
```bash
cd apps/web
pnpm install
pnpm dev
```
👉 Truy cập giao diện tại: **`http://localhost:3000`**

#### Bước 5: Chạy các Backend Services (Golang)
Mở các cửa sổ Terminal riêng biệt để chạy các service cần dev:

```bash
# 1. Chạy API Gateway (Bắt buộc chạy service này trước)
cd services/gateway
go run ./cmd/server

# 2. Chạy AI Service (Ví dụ muốn dev tính năng AI)
cd services/ai
go run ./cmd/server

# 3. Chạy Auth Service
cd services/auth
go run ./cmd/server
```

---

## 6. Ma Trận Cổng & Đường Dẫn Dịch Vụ (Service Port Matrix)

Intern hãy lưu lại bảng tra cứu này để biết cổng truy cập của từng dịch vụ và công cụ quản trị:

| Dịch Vụ / Công Cụ | Giao Thức / Cổng | URL Truy Cập Trực Tiếp | Tài Khoản Đăng Nhập Mặc Định |
|---|---|---|---|
| **Frontend Web App** | HTTP `3000` | http://localhost:3000 | — |
| **API Gateway** | HTTP `8080` | http://localhost:8080/health | — |
| **Auth Service** | HTTP `8081` | http://localhost:8081/health | — |
| **Employee Service** | HTTP `8082` | http://localhost:8082/health | — |
| **Asset Service** | HTTP `8083` | http://localhost:8083/health | — |
| **Helpdesk Service** | HTTP `8084` | http://localhost:8084/health | — |
| **Workflow Service** | HTTP `8085` | http://localhost:8085/health | — |
| **Notification Service**| HTTP `8086` | http://localhost:8086/health | — |
| **Knowledge Service** | HTTP `8087` | http://localhost:8087/health | — |
| **AI Operations Service**| HTTP `8088` | http://localhost:8088/health | — |
| **Audit Service** | HTTP `8089` | http://localhost:8089/health | — |
| **Reporting Service** | HTTP `8090` | http://localhost:8090/health | — |
| **PostgreSQL Database** | TCP `5432` | `localhost:5432` | User: `eomp` / Pass: `eomp_dev_password` |
| **Redis Cache** | TCP `6379` | `localhost:6379` | Pass: `eomp_dev_password` |
| **RabbitMQ AMQP** | TCP `5672` | `localhost:5672` | User: `eomp` / Pass: `eomp_dev_password` |
| **RabbitMQ Dashboard** | HTTP `15672`| http://localhost:15672 | User: `eomp` / Pass: `eomp_dev_password` |
| **MinIO Web Console** | HTTP `9001` | http://localhost:9001 | User: `eomp_minio` / Pass: `eomp_minio_secret` |
| **MinIO S3 API** | HTTP `9000` | http://localhost:9000 | — |
| **Qdrant Vector DB** | HTTP `6333` | http://localhost:6333/dashboard | — |
| **Prometheus Metrics** | HTTP `9090` | http://localhost:9090 | — |
| **Grafana Dashboards** | HTTP `3002` | http://localhost:3002 | User: `admin` / Pass: `eomp_grafana` |
| **Loki Log Aggregator**| HTTP `3100` | http://localhost:3100 | — |

---

## 7. Phân Tích Chi Tiết Từng Microservice (Deep-Dive 11 Services)

Mỗi microservice trong `services/` được thiết kế theo cấu trúc module độc lập và có mục đích rõ ràng. Dưới đây là phân tích chi tiết từng dịch vụ:

### 7.1. API Gateway (`services/gateway` — Port 8080)
- **Vai trò**: Cổng giao tiếp duy nhất giữa Frontend và Backend.
- **Tính năng chính**:
  - Nhận HTTP request từ Frontend Nuxt 4.
  - Tự động sinh mã `X-Request-ID` (Correlation ID) cho mỗi request để lần vết log xuyên suốt các service.
  - Xử lý CORS (Cross-Origin Resource Sharing).
  - Tự động bắt lỗi Panic (`middleware.Recoverer`) tránh tình trạng sập server.
  - Ghi log chi tiết (Method, Path, Status, Latency) qua Go `slog`.
- **Cấu trúc code**:
  - `cmd/server/main.go`: Khởi tạo HTTP router, cấu hình Graceful Shutdown.
  - `internal/config/config.go`: Đọc `APP_PORT=8080`, `APP_ENV`.
  - `internal/handler/health.go`: Endpoint `/health` kiểm tra trạng thái sống.

### 7.2. Auth Service (`services/auth` — Port 8081)
- **Vai trò**: Quản lý định danh, đăng nhập, cấp phát và thu hồi JWT Token, kiểm tra quyền truy cập (RBAC).
- **Cơ sở dữ liệu**: `auth_db`
- **Thực thể dữ liệu chính (Models)**:
  - `User`: `id`, `email`, `password_hash`, `role`, `status`, `created_at`.
  - `Session`: `id`, `user_id`, `refresh_token`, `expires_at`, `ip_address`.
  - `Permission`: `id`, `code` (ví dụ: `ticket:create`, `asset:delete`), `role`.
- **Endpoints chính**:
  - `POST /api/auth/login`: Xác thực email/password, trả về Access Token (JWT) & Refresh Token.
  - `POST /api/auth/refresh`: Cấp mới Access Token bằng Refresh Token.
  - `POST /api/auth/logout`: Hủy session trong Redis.
  - `GET /api/auth/me`: Lấy thông tin user hiện tại từ Token.

### 7.3. Employee Service (`services/employee` — Port 8082)
- **Vai trò**: Quản lý hồ sơ nhân viên, chức danh, phòng ban và quan hệ cấp trên - cấp dưới (Organizational Structure).
- **Cơ sở dữ liệu**: `employee_db`
- **Thực thể dữ liệu chính (Models)**:
  - `Employee`: `id`, `user_id`, `full_name`, `email`, `phone`, `department_id`, `manager_id`, `avatar_url`, `status`.
  - `Department`: `id`, `name`, `code`, `parent_id`, `lead_id`.
- **Lưu trữ file**: Avatar nhân viên được lưu tại MinIO bucket `employees`.

### 7.4. Asset Service (`services/asset` — Port 8083)
- **Vai trò**: Quản lý toàn bộ vòng đời tài sản CNTT trong doanh nghiệp.
- **Cơ sở dữ liệu**: `asset_db`
- **Thực thể dữ liệu chính (Models)**:
  - `Asset`: `id`, `asset_tag` (mã kiểm kê, ví dụ `AST-2026-001`), `name`, `category` (Laptop, Server, Monitor, License), `serial_number`, `purchase_date`, `warranty_expiry`, `status` (`in_stock`, `assigned`, `maintenance`, `retired`), `assigned_to_employee_id`.
  - `AssetMaintenance`: Lịch sử sửa chữa, bảo trì thiết bị.
- **Lưu trữ file**: Hóa đơn mua hàng, biên bản bàn giao lưu tại MinIO bucket `assets`.

### 7.5. Helpdesk Service (`services/helpdesk` — Port 8084)
- **Vai trò**: Tiếp nhận sự cố, yêu cầu hỗ trợ IT, gán kỹ thuật viên và theo dõi SLA.
- **Cơ sở dữ liệu**: `helpdesk_db`
- **Thực thể dữ liệu chính (Models)**:
  - `Ticket`: `id`, `ticket_code` (ví dụ `TK-1094`), `title`, `description`, `category` (`Hardware`, `Software`, `Network`, `Account`), `priority` (`low`, `medium`, `high`, `critical`), `status` (`open`, `in_progress`, `waiting_user`, `resolved`, `closed`), `creator_id`, `assignee_id`, `sla_due_at`, `resolved_at`.
  - `TicketComment`: Bình luận trao đổi giữa người dùng và kỹ thuật viên.
  - `TicketAttachment`: File đính kèm lưu tại MinIO bucket `tickets`.
- **Event Bus (RabbitMQ)**:
  - Bắn event `ticket.created` khi có sự cố mới.
  - Bắn event `ticket.status_changed` khi trạng thái thay đổi.

### 7.6. Workflow Service (`services/workflow` — Port 8085)
- **Vai trò**: State Machine điều khiển quy trình duyệt nhiều cấp (Multi-step approval).
- **Cơ sở dữ liệu**: `workflow_db`
- **Thực thể dữ liệu chính (Models)**:
  - `WorkflowDefinition`: Định nghĩa quy trình (Tên, danh sách các bước duyệt, điều kiện).
  - `WorkflowInstance`: Một phiên chạy thực tế (ví dụ: Đơn xin cấp laptop mới của nhân viên A).
  - `ApprovalStep`: `step_index`, `approver_id`, `status` (`pending`, `approved`, `rejected`), `comment`.

### 7.7. Notification Service (`services/notification` — Port 8086)
- **Vai trò**: Trung tâm gửi thông báo đa kênh.
- **Cơ chế hoạt động**:
  - Không có database riêng; chủ yếu lắng nghe các message từ RabbitMQ (`amq.topic`).
  - Gửi Email qua SMTP.
  - Gửi In-App notification tới Web qua Server-Sent Events (SSE) hoặc WebSocket.

### 7.8. Knowledge Service (`services/knowledge` — Port 8087)
- **Vai trò**: Quản lý tài liệu kỹ thuật, cẩm nang IT, hướng dẫn xử lý sự cố.
- **Cơ sở dữ liệu**: `knowledge_db` & **Qdrant Vector DB**
- **Cơ chế hoạt động**:
  - Khi một bài viết Knowledge Base được tạo hoặc sửa, service sẽ tách văn bản thành các đoạn (chunks), tạo vector embeddings (768 chiều) và đẩy vào Qdrant collection `knowledge_base` phục vụ tìm kiếm ngữ nghĩa.

### 7.9. AI Operations Service (`services/ai` — Port 8088)
- **Vai trò**: Trợ lý AI thông minh tích hợp LLM & RAG.
- **Cấu trúc chi tiết của AI Service**:
  - `internal/model/ai.go`: Định nghĩa struct `ChatRequest`, `ChatResponse`, `Citation`, `TicketAnalysis`.
  - `internal/prompt/prompt.go`: Chứa System Prompt định hướng cho trợ lý AI.
  - `internal/provider/llm.go`: Interface trừu tượng hóa nhà cung cấp LLM (OpenAI, Gemini, Local Ollama, Mock).
  - `internal/provider/embedding.go`: Interface tạo vector embeddings.
  - `internal/rag/retriever.go`: Cơ chế tìm kiếm tài liệu tương đồng trong Qdrant Vector DB.
  - `internal/service/ai.go`: Điều phối luồng xử lý RAG & Phân loại Ticket.
- **Nguyên tắc an toàn AI**:
  - Cờ `requires_human_review: true` luôn được thiết lập; AI chỉ đóng vai trò khuyến nghị và hỗ trợ, không được tự ý thực thi các thay đổi quan trọng mà không có sự xác nhận của kỹ thuật viên.

### 7.10. Audit Service (`services/audit` — Port 8089)
- **Vai trò**: Lưu vết toàn bộ hoạt động của người dùng và hệ thống nhằm phục vụ kiểm toán an ninh bảo mật.
- **Cơ sở dữ liệu**: `audit_db`
- **Đặc tính**: Dữ liệu Append-only (Chỉ ghi thêm, không bao giờ update hoặc delete).
- **Thực thể dữ liệu chính (Models)**:
  - `AuditLog`: `id`, `timestamp`, `actor_id`, `actor_email`, `action` (`CREATE`, `UPDATE`, `DELETE`, `LOGIN`), `resource_type` (`TICKET`, `ASSET`, `USER`), `resource_id`, `ip_address`, `user_agent`, `old_values` (JSON), `new_values` (JSON).

### 7.11. Reporting Service (`services/reporting` — Port 8090)
- **Vai trò**: Tổng hợp dữ liệu phân tích vận hành, xuất file báo cáo Excel/PDF.
- **Lưu trữ**: File báo cáo định kỳ được lưu tại MinIO bucket `reports`.

---

## 8. Kiến Trúc Frontend (Nuxt 4 / Vue 3 / Tailwind v4)

Frontend được đặt tại thư mục `apps/web/` và xây dựng trên nền tảng **Nuxt 4**, sử dụng **Vue 3 Composition API**, **TypeScript** và **Tailwind CSS v4**.

```
apps/web/
├── app/
│   ├── app.vue                 # Điểm khởi đầu của ứng dụng Vue
│   ├── layouts/
│   │   └── default.vue         # Layout chuẩn có Sidebar navigation, Header, User Menu
│   ├── pages/
│   │   ├── index.vue           # Trang Dashboard tổng quan với Metrics & Charts
│   │   ├── helpdesk.vue        # Trang quản lý Ticket Helpdesk
│   │   ├── assets.vue          # Trang quản trị Tài sản IT
│   │   ├── employees.vue       # Trang danh sách Nhân viên & Phòng ban
│   │   ├── workflows.vue       # Trang quy trình Phê duyệt
│   │   ├── knowledge.vue       # Trang Kho tri thức & Hướng dẫn
│   │   ├── ai.vue              # Trang AI Operations Copilot Chat & Triage
│   │   ├── audit.vue           # Trang Nhật ký kiểm toán an ninh
│   │   └── reports.vue         # Trang Báo cáo & Phân tích vận hành
│   ├── stores/
│   │   └── auth.ts             # Pinia Store quản lý thông tin User đăng nhập
│   └── types/
│       └── index.ts            # Khai báo TypeScript types chuẩn
```

### 8.1. Các Nguyên Tắc Lập Trình Frontend Cho Intern

1. **Sử dụng `<script setup lang="ts">`**: Bắt buộc sử dụng cú pháp Composition API hiện đại của Vue 3.
2. **Quản lý Layout**: Mọi trang trong `pages/` đều khai báo:
   ```vue
   <script setup lang="ts">
   definePageMeta({ layout: 'default' })
   </script>
   ```
3. **Gọi API về Backend**:
   - URL API Backend được cấu hình qua `NUXT_PUBLIC_API_URL` (mặc định `http://localhost:8080`).
   - Sử dụng `useFetch` hoặc `$fetch` tích hợp sẵn của Nuxt để tự động serialize/deserialize dữ liệu.
4. **Giao diện & Thẩm mỹ**:
   - Sử dụng bảng màu tối hiện đại (Dark slate theme: `bg-slate-950`, `bg-slate-900/60`, `border-slate-800`).
   - Kết hợp Glassmorphism (`backdrop-blur-xl`), hiệu ứng Gradient, và Icon chuẩn từ `@nuxt/ui` / Lucide icons (`i-lucide-*`).

---

## 9. Tầng Hạ Tầng, Cơ Sở Dữ Liệu & DevOps (Infrastructure Deep-Dive)

Hạ tầng phát triển được đóng gói hoàn toàn trong file `docker-compose.yml`.

### 9.1. PostgreSQL (Cơ Sở Dữ Liệu Quan Hệ)
- **Container**: `eomp-postgres` (Image: `postgres:17-alpine`)
- **Port**: `5432`
- **Script Khởi Tạo**: Khi container chạy lần đầu, file `infrastructure/postgres/01-init-databases.sql` sẽ tự động tạo **7 database độc lập**:
  1. `auth_db`
  2. `employee_db`
  3. `asset_db`
  4. `helpdesk_db`
  5. `workflow_db`
  6. `knowledge_db`
  7. `audit_db`

### 9.2. Redis (Cache & Session Store)
- **Container**: `eomp-redis` (Image: `redis:7-alpine`)
- **Port**: `6379`
- **Chế độ**: Bật AOF (`--appendonly yes`) để dữ liệu không bị mất khi restart.

### 9.3. RabbitMQ (Message Broker & Event Bus)
- **Container**: `eomp-rabbitmq` (Image: `rabbitmq:4-management-alpine`)
- **Cổng giao tiếp**: `5672` (AMQP Protocol)
- **Cổng giao diện Dashboard**: `15672` (Truy cập `http://localhost:15672` với user: `eomp` / pass: `eomp_dev_password`)

### 9.4. MinIO (S3-Compatible Object Storage)
- **Container**: `eomp-minio` & `eomp-minio-init`
- **Cổng S3 API**: `9000` | **Cổng Console**: `9001`
- **Khởi tạo tự động 5 Buckets**:
  - `employees`: Chứa ảnh avatar và hồ sơ scan.
  - `assets`: Chứa hợp đồng, hóa đơn thiết bị.
  - `tickets`: Chứa ảnh/video đính kèm sự cố kỹ thuật.
  - `knowledge`: Chứa tài liệu PDF hướng dẫn.
  - `reports`: Chứa file xuất báo cáo Excel/PDF.

### 9.5. Qdrant (Vector Database)
- **Container**: `eomp-qdrant` (Image: `qdrant/qdrant:latest`)
- **Port**: `6333` (HTTP REST API) & `6334` (gRPC)
- **Mục đích**: Lưu vector embeddings của bài viết tri thức phục vụ tìm kiếm ngữ nghĩa cho AI Copilot.

### 9.6. Bộ Công Cụ Giám Sát & Quan Sát (Observability Stack)
1. **Prometheus (`:9090`)**: Tự động cào (scrape) các metrics từ tất cả các microservices định kỳ mỗi 15 giây.
2. **Grafana (`:3002`)**: Bảng điều khiển trực quan hóa metrics và logs. Đã nạp sẵn cấu hình datasource Prometheus và Loki trong `infrastructure/grafana/`.
3. **Loki (`:3100`)**: Thu thập toàn bộ log tập trung từ các microservices.

---

## 10. CI/CD Pipeline với Jenkins

File `Jenkinsfile` ở thư mục gốc định nghĩa quy trình tự động hóa kiểm thử và build phần mềm chuẩn CI/CD gồm 6 giai đoạn:

```
┌──────────────┐     ┌──────────────────────┐     ┌──────────────────────┐
│ 1. Checkout  ├────►│ 2. Install Dependency├────►│ 3. Lint & Formatting │
└──────────────┘     └──────────────────────┘     └──────────────────────┘
                                                             │
┌──────────────┐     ┌──────────────────────┐                │
│ 6. Complete  │◄────┤ 5. Build Artifacts   │◄────┤ 4. Run Unit Tests    │
└──────────────┘     └──────────────────────┘     └──────────────────────┘
```

1. **Stage 1: Checkout**: Kéo mã nguồn mới nhất từ Git Repository.
2. **Stage 2: Install Dependencies**:
   - Cài đặt Frontend dependencies bằng `pnpm install --frozen-lockfile`.
   - Đồng bộ Go Workspace bằng `go work sync`.
3. **Stage 3: Lint & Static Analysis**:
   - Chạy `pnpm lint` trên Frontend.
   - Kiểm tra định dạng `gofmt` và chạy `go vet ./...` trên toàn bộ 12 Go modules.
4. **Stage 4: Run Tests**:
   - Chạy kiểm tra kiểu tĩnh `pnpm typecheck` trên Frontend.
   - Chạy toàn bộ Unit Test Go với cờ đua tranh bộ nhớ và đo độ phủ: `go test -v -race -cover ./...`.
5. **Stage 5: Build Artifacts**:
   - Build Production Bundle cho Frontend (`pnpm build`).
   - Biên dịch binary cho toàn bộ Go services vào thư mục `bin/`.
6. **Stage 6: Docker Image Build**: Đóng gói Docker image cho các services (`eomp-<service>:latest`).

---

## 11. Quy Chuẩn Code & Best Practices Cho Intern

### 11.1. Quy Ước Đặt Tên Nhánh Git (Git Branching Strategy)
- `main`: Nhánh production, tuyệt đối không commit trực tiếp.
- `develop`: Nhánh tích hợp chính của đội ngũ dev.
- `feature/<tên-tính-năng>`: Nhánh làm tính năng mới (ví dụ: `feature/ticket-filter-status`).
- `bugfix/<tên-lỗi>`: Nhánh sửa lỗi (ví dụ: `bugfix/asset-export-excel-null`).
- `hotfix/<tên-lỗi-khẩn-cấp>`: Sửa lỗi trực tiếp trên production.

### 11.2. Quy Ước Viết Commit Chuẩn (Conventional Commits)
Mỗi commit message bắt buộc tuân theo định dạng: `<type>(<scope>): <mô tả ngắn>`

Các `type` được chấp nhận:
- `feat`: Thêm tính năng mới (ví dụ: `feat(helpdesk): add SLA due date calculator`).
- `fix`: Sửa lỗi (ví dụ: `fix(auth): resolve JWT expiration timezone bug`).
- `refactor`: Tối ưu hoặc cấu trúc lại code mà không đổi hành vi.
- `test`: Thêm hoặc sửa unit test.
- `docs`: Cập nhật tài liệu.
- `chore`: Việc bảo trì, cấu hình build, dependencies.

### 11.3. Chuẩn Viết Mã Nguồn Golang
1. **Formatting**: Luôn chạy `gofmt -w .` trước khi commit code.
2. **Xử lý lỗi (Error Handling)**:
   - Luôn kiểm tra `if err != nil` ngay sau khi gọi hàm.
   - Không được bỏ qua lỗi bằng `_ = doSomething()` trừ trường hợp đóng reader hoặc encode JSON không critical.
   - Tuyệt đối không gọi `panic()` trong luồng xử lý request HTTP; hãy trả về lỗi dạng JSON chuẩn qua `response.Error(w, http.StatusBadRequest, "thông báo lỗi")`.
3. **Ghi log có cấu trúc (Structured Logging)**:
   - Sử dụng package chuẩn `log/slog`.
   - **Tốt**: `logger.Info("ticket created", slog.String("ticket_id", id), slog.String("creator", email))`
   - **Không tốt**: `fmt.Println("Created ticket " + id)`
4. **Context Propagation**: Luôn truyền `ctx context.Context` vào các hàm gọi Database, Redis, HTTP client để hỗ trợ timeout và hủy request khi client ngắt kết nối.
5. **Dependency Injection**: Khởi tạo struct thông qua hàm Constructor `New<Service>()` và nhận interface thay vì struct cụ thể để dễ viết unit test mock.

### 11.4. Chuẩn Viết Mã Nguồn TypeScript / Vue
1. **Strict Type**: Không sử dụng kiểu `any` bừa bãi; hãy định nghĩa interface rõ ràng trong `types/`.
2. **Reactivity**: Dùng `ref()` cho giá trị nguyên thủy và `reactive()` cho object phức tạp.
3. **Component Reusability**: Tách nhỏ các component nếu dài quá 200 dòng.

---

## 12. Hướng Dẫn Thực Hành: Xây Dựng Một Tính Năng Mới Từ A-Z

Giả sử intern nhận task: *"Viết thêm API xem chi tiết Ticket và hiển thị trên màn hình Helpdesk"*. Hãy làm theo 6 bước chuẩn sau:

### Bước 1: Khai báo Model dữ liệu (Backend)
Mở file `services/helpdesk/internal/model/ticket.go`:
```go
package model

import "time"

type TicketDetail struct {
    ID          string    `json:"id"`
    TicketCode  string    `json:"ticket_code"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    Priority    string    `json:"priority"`
    CreatedAt   time.Time `json:"created_at"`
}
```

### Bước 2: Viết Business Logic trong Service Layer
Mở file `services/helpdesk/internal/service/ticket.go`:
```go
package service

import (
    "context"
    "errors"
    "eomp/services/helpdesk/internal/model"
)

type TicketService interface {
    GetTicketByID(ctx context.Context, id string) (*model.TicketDetail, error)
}

type ticketService struct {
    // Inject repository / db connection ở đây
}

func NewTicketService() TicketService {
    return &ticketService{}
}

func (s *ticketService) GetTicketByID(ctx context.Context, id string) (*model.TicketDetail, error) {
    if id == "" {
        return nil, errors.New("ticket id is required")
    }
    // Giả lập truy vấn dữ liệu
    return &model.TicketDetail{
        ID:         id,
        TicketCode: "TK-1094",
        Title:      "VPN Connection drops every 10 mins",
        Status:     "in_progress",
        Priority:   "high",
    }, nil
}
```

### Bước 3: Viết HTTP Handler tiếp nhận Request
Mở file `services/helpdesk/internal/handler/ticket.go`:
```go
package handler

import (
    "net/http"
    "eomp/packages/shared/pkg/response"
    "eomp/services/helpdesk/internal/service"
)

type TicketHandler struct {
    svc service.TicketService
}

func NewTicketHandler(svc service.TicketService) *TicketHandler {
    return &TicketHandler{svc: svc}
}

func (h *TicketHandler) GetDetail(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    ticket, err := h.svc.GetTicketByID(r.Context(), id)
    if err != nil {
        response.Error(w, http.StatusNotFound, err.Error())
        return
    }
    response.JSON(w, http.StatusOK, ticket)
}
```

### Bước 4: Đăng Ký Tuyến Đường (Route Registration)
Trong `services/helpdesk/cmd/server/main.go`:
```go
ticketSvc := service.NewTicketService()
ticketHandler := handler.NewTicketHandler(ticketSvc)

mux.HandleFunc("GET /api/tickets/{id}", ticketHandler.GetDetail)
```

### Bước 5: Viết Unit Test kiểm thử
Tạo file `services/helpdesk/internal/service/ticket_test.go`:
```go
package service_test

import (
    "context"
    "testing"
    "eomp/services/helpdesk/internal/service"
)

func TestGetTicketByID(t *testing.T) {
    svc := service.NewTicketService()
    
    // Case 1: ID rỗng
    _, err := svc.GetTicketByID(context.Background(), "")
    if err == nil {
        t.Errorf("expected error when id is empty, got nil")
    }

    // Case 2: ID hợp lệ
    ticket, err := svc.GetTicketByID(context.Background(), "mock-123")
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if ticket.ID != "mock-123" {
        t.Errorf("expected ticket ID mock-123, got %s", ticket.ID)
    }
}
```

### Bước 6: Tích Hợp Lên Giao Diện Frontend (Vue 3 / Nuxt 4)
Trong `apps/web/app/pages/helpdesk.vue`:
```vue
<script setup lang="ts">
const config = useRuntimeConfig()
const ticket = ref(null)
const loading = ref(true)

async function fetchTicket(id: string) {
  loading.value = true
  try {
    ticket.value = await $fetch(`${config.public.apiUrl}/api/tickets/${id}`)
  } catch (err) {
    console.error('Lỗi tải ticket:', err)
  } finally {
    loading.value = false
  }
}
</script>
```

---

## 13. Chiến Lược Kiểm Thử (Testing & QA Guide)

Dự án cung cấp các công cụ kiểm thử tự động toàn diện.

### 13.1. Chạy Unit Test Cho Backend Go
```bash
# Chạy test cho một service cụ thể
cd services/ai
go test -v ./...

# Chạy test có đo lường độ phủ code (Coverage)
go test -cover ./...

# Chạy test kiểm tra tranh chấp bộ nhớ (Race Detector)
go test -race ./...
```

### 13.2. Chạy Kiểm Tra Frontend Web
```bash
cd apps/web

# Kiểm tra lỗi cú pháp TypeScript
pnpm typecheck

# Kiểm tra quy chuẩn mã nguồn ESLint
pnpm lint

# Kiểm tra khả năng build production
pnpm build
```

### 13.3. Chạy Toàn Bộ Bộ Kiểm Thử Tự Động QA/QC
Trên Windows PowerShell, chạy script kiểm thử toàn diện:
```powershell
.\scripts\qa.ps1
```

Script sẽ tự động kiểm tra 5 hạng mục:
1. **Frontend QA**: `typecheck`, `lint`, `build`
2. **Backend Go QA**: `go vet`, `go test`, `go build` trên cả 12 modules
3. **Hạ Tầng Health Check**: Postgres, Redis, RabbitMQ, MinIO, Qdrant, Prometheus, Grafana, Loki
4. **Database Verification**: Kiểm tra sự tồn tại của 7 database
5. **Docker & CI**: Kiểm tra cú pháp `docker-compose.yml` và `Jenkinsfile`

---

## 14. Xử Lý Sự Cố Thường Gặp (Troubleshooting & FAQs)

### ❓ Lỗi 1: `port is already allocated` (Trùng cổng 5432, 8080, 3000,...)
- **Nguyên nhân**: Bạn đang chạy một dịch vụ Postgres/Redis cài trực tiếp trên máy hoặc một phiên bản khác đang chiếm dụng cổng.
- **Cách khắc phục**:
  - Tắt các dịch vụ chạy nền của Windows (`services.msc` -> Dừng dịch vụ `postgresql-x64`).
  - Hoặc đổi cổng trong file `.env` (ví dụ `POSTGRES_PORT=5433`).

### ❓ Lỗi 2: `go.work` báo lỗi không nhận module mới
- **Nguyên nhân**: Bạn vừa tạo một thư mục service mới nhưng chưa đăng ký vào Go Workspace.
- **Cách khắc phục**:
  ```bash
  # Đồng bộ lại workspace
  go work sync
  ```

### ❓ Lỗi 3: Frontend báo lỗi CORS khi gọi API
- **Nguyên nhân**: API Gateway chưa cấu hình cho phép Origin hoặc Request Header `Authorization`/`X-Request-ID`.
- **Cách khắc phục**: Kiểm tra middleware `CORS` trong `packages/shared/pkg/middleware/middleware.go`, đảm bảo `Access-Control-Allow-Origin: *` và đầy đủ các headers.

### ❓ Lỗi 4: Docker Compose không tạo được các database riêng biệt
- **Nguyên nhân**: Volume của Postgres đã được khởi tạo từ trước khi có file script SQL.
- **Cách khắc phục**: Reset lại toàn bộ volume của Docker:
  ```powershell
  .\scripts\dev.ps1 docker-reset
  .\scripts\dev.ps1 docker-up
  ```

### ❓ Lỗi 5: MinIO báo lỗi kết nối hoặc không tìm thấy Bucket
- **Nguyên nhân**: Container `minio-init` chưa chạy xong trước khi ứng dụng khởi động.
- **Cách khắc phục**: Truy cập MinIO Console tại `http://localhost:9001` (user: `eomp_minio` / pass: `eomp_minio_secret`) và kiểm tra danh sách buckets `employees`, `assets`, `tickets`, `knowledge`, `reports`. Nếu thiếu, tạo thủ công qua giao diện.

---

## 📞 Hỗ Trợ Kỹ Thuật & Liên Hệ

Nếu gặp bất kỳ khó khăn nào trong quá trình Onboarding hoặc phát triển dự án, intern hãy:
1. Đọc kỹ lại tài liệu này và các file tài liệu chuyên sâu trong thư mục `docs/`.
2. Kiểm tra log của container tương ứng bằng lệnh `.\scripts\dev.ps1 logs [service-name]`.
3. Trao đổi trực tiếp với Mentor hoặc Tech Lead phụ trách dự án.

*Chúc bạn có một kỳ thực tập và làm việc hiệu quả, học hỏi được nhiều kiến thức giá trị cùng dự án EOMP!*
