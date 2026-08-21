# EOMP — Developer & Operations Setup Guide

> **Hướng Dẫn Cài Đặt Môi Trường Phát Triển & Khởi Chạy Nền Tảng (Setup Guide)**  
> **Áp dụng cho:** Tất cả lập trình viên, QA/QC và DevOps mới tiếp cận dự án.

---

## 1. Yêu Cầu Môi Trường Cài Đặt (Prerequisites)

| Công Cụ / Runtime | Phiên Bản Tối Thiểu | Ghi Chú Cần Thiết |
|---|:---:|---|
| **Node.js** | `>= 20.19.x` (LTS) | Chạy Nuxt 4 SSR Web App |
| **pnpm** | `>= 10.x` | Quản lý package frontend siêu tốc |
| **Golang** | `>= 1.24.x` | Biên dịch 11 microservices Go |
| **Docker & Compose**| `>= 28.x` / `Compose v2.30+`| Chạy container cơ sở hạ tầng & production |
| **Kubernetes (kubectl)**| `>= 1.28.x` | Quản lý cụm container K8s (nếu deploy K8s) |
| **Helm** | `>= 3.12.x` | Quản lý Helm package manager (nếu deploy Helm) |
| **Git** | `>= 2.40.x` | Quản lý mã nguồn monorepo |

---

## 2. Các Bước Cài Đặt & Khởi Động Nhanh (Step-by-Step)

### Bước 1: Clone Repository & Cấu Hình Biến Môi Trường
```bash
git clone https://github.com/hoangtruong01/AI-IT-Help.git
cd AI-IT-Help

# Tạo file cấu hình từ template
cp .env.example .env
```

### Bước 2: Khởi Động Hạ Tầng Cơ Sở Dữ Liệu & Giám Sát

* **Trên Windows (PowerShell):**
  ```powershell
  .\scripts\dev.ps1 docker-up
  .\scripts\dev.ps1 health
  ```
* **Trên Linux / macOS (Makefile):**
  ```bash
  make docker-up
  make health
  ```

### Bước 3: Khởi Động Frontend Nuxt 4 Web Portal
```bash
cd apps/web
pnpm install
pnpm dev
```
👉 Truy cập giao diện tại: **`http://localhost:3000`**

### Bước 4: Khởi Động Backend Microservices
Có 2 cách khởi chạy Backend:
1. **Chạy Local từng service để Debug:**
   ```bash
   cd services/gateway && go run ./cmd/server
   ```
2. **Khởi động Full Cụm Production (11 Services + Web + Nginx):**
   ```powershell
   .\scripts\deploy.ps1 prod-up
   ```

---

## 3. Danh Mục Cổng Truy Cập & Dịch Vụ Mặc Định

| Phân Hệ | Cổng / URL | Tài Khoản Đăng Nhập Mặc Định |
|---|---|---|
| **Frontend Web App** | `http://localhost:3000` | `admin@eomp.local` / `password123` |
| **API Gateway** | `http://localhost:8080` | Header `Authorization: Bearer <JWT>` |
| **Grafana SRE Dashboard** | `http://localhost:3002` | `admin` / `eomp_grafana` |
| **Prometheus Metrics** | `http://localhost:9090` | *(Truy cập trực tiếp)* |
| **RabbitMQ Management** | `http://localhost:15672` | `eomp` / `eomp_dev_password` |
| **MinIO Console** | `http://localhost:9001` | `eomp_minio` / `eomp_minio_secret` |
| **Qdrant Vector DB** | `http://localhost:6333` | REST API vector search |
