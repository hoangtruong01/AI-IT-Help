# EOMP — System Operations Manual (Day-2 Ops)

> **Cẩm Nang Vận Hành & Bảo Trì Hệ Thống Định Kỳ (Day-2 Operations Manual)**  
> **Áp dụng cho:** System Administrators, DevOps Engineers, SRE & Platform Engineers.  
> **Phiên bản:** Phase 14 Enterprise Master Standard

---

## 📑 MỤC LỤC
1. [Khởi Động & Dừng Hệ Thống (Startup & Shutdown Procedures)](#1-khởi-động--dừng-hệ-thống-startup--shutdown-procedures)
2. [Quản Lý & Cập Nhật Cơ Sở Dữ Liệu (Database Migration Workflow)](#2-quản-lý--cập-nhật-cơ-sở-dữ-liệu-database-migration-workflow)
3. [Quy Trình Xoay Vòng Khóa Bảo Mật (Secret & Key Rotation)](#3-quy-trình-xoay-vòng-khóa-bảo-mật-secret--key-rotation)
4. [Bảo Trì Nâng Cấp Không Gián Đoạn (Zero-Downtime Rolling Updates)](#4-bảo-trì-nâng-cấp-không-gián-đoạn-zero-downtime-rolling-updates)
5. [Giám Sát & Thu Thập Logs (Observability & Log Analytics)](#5-giám-sát--thu-thập-logs-observability--log-analytics)

---

## 1. Khởi Động & Dừng Hệ Thống (Startup & Shutdown Procedures)

### 1.1 Thứ Tự Khởi Động Chuẩn (Dependency Order):
1. **Lớp 1 (Cơ Sở Dữ Liệu & Hàng Đợi):** PostgreSQL 17 -> Redis 7 -> RabbitMQ 4 -> MinIO -> Qdrant.
2. **Lớp 2 (Hạ Tầng Giám Sát):** Prometheus -> Loki -> Grafana.
3. **Lớp 3 (Microservices Cốt Lõi):** `auth` (:8081) -> `employee` (:8082) -> `audit` (:8089).
4. **Lớp 4 (Microservices Nghiệp Vụ):** `helpdesk`, `asset`, `workflow`, `notification`, `knowledge`, `ai`, `reporting`.
5. **Lớp 5 (Cổng Giao Tiếp & Giao Diện):** `gateway` (:8080) -> `web` (:3000) -> `nginx` (:80, :443).

### 1.2 Lệnh Thực Hiện:
```bash
# Khởi động toàn bộ qua script
./scripts/deploy.sh prod-up

# Dừng an toàn (Graceful Shutdown)
./scripts/deploy.sh prod-down
```

---

## 2. Quản Lý & Cập Nhật Cơ Sở Dữ Liệu (Database Migration Workflow)

Tất cả các microservices của EOMP đều tích hợp sẵn bộ **Auto Migration Runner** (`packages/shared/pkg/database/postgres.go`).

1. Khi tạo bảng mới hoặc sửa đổi cột, tạo file SQL đánh số thứ tự trong thư mục `services/<service_name>/migrations/`:  
   Ví dụ: `003_add_new_feature_table.sql`.
2. Khi container khởi động, engine sẽ tự động kiểm tra bảng `schema_migrations`, thực thi file SQL mới trong một Transaction khép kín và ghi nhận trạng thái.
3. Nếu migration bị lỗi, container sẽ tự động dừng lại kèm thông báo log chi tiết mà không gây sai lệch dữ liệu cũ.

---

## 3. Quy Trình Xoay Vòng Khóa Bảo Mật (Secret & Key Rotation)

Theo tiêu chuẩn bảo mật SOC2 / ISO 27001, các bí mật hệ thống phải được xoay vòng mỗi 90 ngày:

### 3.1 Xoay Vòng JWT Secret Token
```bash
# 1. Cập nhật secret mới trong Kubernetes
kubectl create secret generic eomp-secrets \
  --from-literal=JWT_SECRET="eomp_new_secret_key_$(date +%Y%m%d)" \
  --dry-run=client -o yaml | kubectl apply -f -

# 2. Khởi động lại Gateway và Auth Service tuần tự
kubectl rollout restart deployment/gateway-deployment -n eomp
kubectl rollout restart deployment/auth-deployment -n eomp
```

### 3.2 Xoay Vòng Mật Khẩu Database PostgreSQL
```bash
# 1. Đổi mật khẩu trong Postgres engine
docker compose exec postgres psql -U eomp -c "ALTER USER eomp WITH PASSWORD 'new_secure_password_2026';"

# 2. Cập nhật file .env / Kubernetes Secret và restart các services liên quan
```

---

## 4. Bảo Trì Nâng Cấp Không Gián Đoạn (Zero-Downtime Rolling Updates)

Kubernetes Deployments được cấu hình chiến lược `RollingUpdate` với `maxUnavailable: 0` và `maxSurge: 1`:

```bash
# Cập nhật phiên bản image mới
kubectl set image deployment/helpdesk-deployment helpdesk=eomp/helpdesk:v2.1.0 -n eomp

# Giám sát quá trình chuyển giao Pods
kubectl rollout status deployment/helpdesk-deployment -n eomp

# Hoàn tác ngay lập tức nếu phát hiện lỗi
kubectl rollout undo deployment/helpdesk-deployment -n eomp
```

---

## 5. Giám Sát & Thu Thập Logs (Observability & Log Analytics)

### 5.1 Các Đường Dẫn Giám Sát Trực Tiếp:
* **SRE Web Portal:** `http://localhost/monitoring`
* **Grafana Dashboards:** `http://localhost:3002` (User: `admin` / Pass: `eomp_grafana`)
* **Prometheus Metrics:** `http://localhost:9090`
* **RabbitMQ Management:** `http://localhost:15672` (User: `eomp` / Pass: `eomp_dev_password`)
* **MinIO Console:** `http://localhost:9001` (User: `eomp_minio` / Pass: `eomp_minio_secret`)

### 5.2 Truy Vấn Log Nhanh Qua Loki (LogCLI / Grafana Explore):
```logql
# Tìm tất cả các lỗi ERROR của Helpdesk Service trong 1 giờ qua
{service="helpdesk"} |= "ERROR"

# Thống kê số lượng mã lỗi 500 theo từng endpoint
sum(count_over_time({app="gateway"} |= "status=500" [5m])) by (path)
```
