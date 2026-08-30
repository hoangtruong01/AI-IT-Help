# EOMP — Environment Variables Specification

> `.env.example` is the current variable-name reference. Values shown here are development examples only. Production Compose and Kubernetes require externally supplied secrets; no example secret is approved for shared or production use.

> **Tài Liệu Cấu Hình Biến Môi Trường Hệ Thống (.env Specification)**  
> **Áp dụng cho:** Tất cả các Microservices, Frontend và Hệ Thống Container.

---

## 1. Global & Core Application

| Biến Môi Trường | Mô Tả Nghiệp Vụ & Kỹ Thuật | Giá Trị Mặc Định |
|---|---|---|
| `APP_ENV` | Môi trường triển khai (`development`, `staging`, `production`) | `development` |
| `JWT_SECRET` | Khóa bí mật dùng ký JWT Access Token HS256 (60m) | Required in production; at least 16 characters |
| `TZ` | Múi giờ chuẩn hệ thống | `UTC` |

---

## 2. Microservices Network URLs (API Gateway Routing)

| Biến Môi Trường | Cấu Hình URL Microservice Mục Tiêu | Giá Trị Mặc Định |
|---|---|---|
| `AUTH_SERVICE_URL` | Địa chỉ Auth & RBAC Service | `http://auth:8081` |
| `EMPLOYEE_SERVICE_URL` | Địa chỉ Employee Directory Service | `http://employee:8082` |
| `ASSET_SERVICE_URL` | Địa chỉ Asset & CMDB Topology Service | `http://asset:8083` |
| `HELPDESK_SERVICE_URL` | Địa chỉ Incident & Problem Service | `http://helpdesk:8084` |
| `WORKFLOW_SERVICE_URL` | Địa chỉ Workflow Engine & CAB Service | `http://workflow:8085` |
| `NOTIFICATION_SERVICE_URL` | Địa chỉ In-App Notification Service | `http://notification:8086` |
| `KNOWLEDGE_SERVICE_URL` | Địa chỉ Knowledge Base & SOP Service | `http://knowledge:8087` |
| `AI_SERVICE_URL` | Địa chỉ AI Ops Copilot & RAG Service | `http://ai:8088` |
| `AUDIT_SERVICE_URL` | Địa chỉ Immutable Audit Trail Service | `http://audit:8089` |
| `REPORTING_SERVICE_URL` | Địa chỉ BI Analytics & SLA Service | `http://reporting:8090` |

---

## 3. Database & Caching Infrastructure

| Biến Môi Trường | Phân Hệ Sử Dụng | Giá Trị Mặc Định |
|---|---|---|
| `POSTGRES_HOST` | PostgreSQL Hostname | `localhost` / `postgres` |
| `POSTGRES_PORT` | PostgreSQL Port | `5432` |
| `POSTGRES_USER` | PostgreSQL Username | `eomp` |
| `POSTGRES_PASSWORD` | PostgreSQL Password | Required in production |
| `REDIS_ADDR` | Redis connection address | `localhost:6379` / `redis:6379` |
| `RABBITMQ_URL` | Optional AMQP URL override; component variables are preferred | No production default |
| `MINIO_ENDPOINT` | S3 MinIO Storage API | `localhost:9000` / `minio:9000` |
| `MINIO_ACCESS_KEY` | MinIO Root User | `eomp_minio` |
| `MINIO_SECRET_KEY` | MinIO Secret Key | Required in production |
| `QDRANT_URL` | Vector Database URL cho RAG AI | `http://localhost:6333` |

---

## 4. Observability & SRE Dashboards

| Biến Môi Trường | Mục Đích Sử Dụng | Giá Trị Mặc Định |
|---|---|---|
| `GRAFANA_ADMIN_USER` | Tài khoản quản trị Grafana | `admin` |
| `GRAFANA_ADMIN_PASSWORD` | Mật khẩu quản trị Grafana | Required in production |
| `PROMETHEUS_PORT` | Cổng thu thập metrics Prometheus | `9090` |
| `LOKI_PORT` | Cổng lưu trữ log tập trung Loki | `3100` |

---

## 5. Frontend Nuxt 4 Configuration

| Biến Môi Trường | Mục Đích Sử Dụng | Giá Trị Mặc Định |
|---|---|---|
| `PORT` / `NITRO_PORT` | Cổng lắng nghe Web Frontend | `3000` |
| `NUXT_PUBLIC_API_BASE_URL` | Base API Gateway URL gọi từ trình duyệt | `/api/v1` |
