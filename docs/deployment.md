# EOMP — Production Deployment & Cloud Orchestration Guide

> **Tài liệu Hướng Dẫn Đóng Gói & Triển Khai Hệ Thống (Production Deployment Guide)**  
> **Áp dụng cho:** DevOps, SRE, SysAdmins, Developers & Infrastructure Engineers.  
> **Phiên bản:** Phase 13 Enterprise Standard

---

## 📑 MỤC LỤC
1. [Tổng Quan Kiến Trúc Đóng Gói & Containerization](#1-tổng-quan-kiến-trúc-đóng-gói--containerization)
2. [Docker Multi-Stage Build (< 25MB Specification)](#2-docker-multi-stage-build--25mb-specification)
3. [Triển Khai Production Docker Compose](#3-triển-khai-production-docker-compose)
4. [Triển Khai Kubernetes Manifests (Native K8s)](#4-triển-khai-kubernetes-manifests-native-k8s)
5. [Triển Khai Bằng Kubernetes Helm Chart](#5-triển-khai-bằng-kubernetes-helm-chart)
6. [Cấu Hình Auto-scaling (HPA) & Ingress Controller](#6-cấu-hình-auto-scaling-hpa--ingress-controller)
7. [Health Probes & Giám Sát Vận Hành (Liveness/Readiness)](#7-health-probes--giám-sát-vận-hành-livenessreadiness)
8. [CLI Helper Scripts & CI/CD Pipeline](#8-cli-helper-scripts--cicd-pipeline)

---

## 1. Tổng Quan Kiến Trúc Đóng Gói & Containerization

Toàn bộ hệ sinh thái EOMP bao gồm **11 Go Microservices**, **1 Nuxt 4 Web Frontend (SSR)**, và **hệ thống cơ sở hạ tầng phân tán** (PostgreSQL 17, Redis 7, RabbitMQ 4, MinIO, Qdrant Vector Store, Prometheus, Grafana, Loki) được đóng gói và cấu hình sẵn sàng cho mọi nền tảng điện toán đám mây (AWS EKS, GCP GKE, Azure AKS, Bare-Metal Kubernetes, Docker Swarm).

```
                        ┌─────────────────────────────────────────┐
                        │      Internet / Enterprise Network      │
                        └───────────────────┬─────────────────────┘
                                            │
                                            ▼
                        ┌─────────────────────────────────────────┐
                        │        Ingress Nginx Controller         │
                        │    (SSL Termination, Rate Limiting)     │
                        └─────────┬─────────────────────┬─────────┘
                                  │                     │
                    /api/v1/*     │                     │  /*
                    ┌─────────────┘                     └─────────────┐
                    ▼                                                 ▼
        ┌───────────────────────┐                         ┌───────────────────────┐
        │   API Gateway :8080   │                         │  Nuxt 4 Web App :3000 │
        │ (JWT Auth, Proxy RED) │                         │  (Node 22 SSR Cluster)│
        └───────────┬───────────┘                         └───────────────────────┘
                    │
    ┌───────────────┼───────────────┬───────────────┬───────────────┐
    ▼               ▼               ▼               ▼               ▼
┌────────┐      ┌────────┐      ┌────────┐      ┌────────┐      ┌────────┐
│  Auth  │      │Employee│      │ Asset  │      │Helpdesk│      │Workflow│ ... (11 Services)
│ :8081  │      │ :8082  │      │ :8083  │      │ :8084  │      │ :8085  │
└───┬────┘      └───┬────┘      └───┬────┘      └───┬────┘      └───┬────┘
    │               │               │               │               │
    └───────────────┴───────┬───────┴───────────────┴───────────────┘
                            ▼
            ┌───────────────────────────────┐
            │   Infrastructure & Storage    │
            │ PostgreSQL, Redis, RabbitMQ,  │
            │  MinIO, Qdrant, Prometheus    │
            └───────────────────────────────┘
```

---

## 2. Docker Multi-Stage Build (< 25MB Specification)

### 2.1 Universal Go Multi-stage Dockerfile
Được thiết kế tại [`deploy/docker/Dockerfile.go-service`](file:///d:/IT_help/eomp/deploy/docker/Dockerfile.go-service):
* **Stage 1 (Builder):** `golang:1.24-alpine` biên dịch tĩnh với cờ `CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w -extldflags '-static'"`.
* **Stage 2 (Runner):** `alpine:3.21` tối giản, tạo non-root user `appuser` (UID: `10001`), tự động copy SQL migrations nếu có.
* **Dung lượng image đầu ra:** Đạt **~18MB - 22MB** (thỏa mãn tiêu chí `< 25MB`).

```bash
# Build 1 service bất kỳ
docker build \
  --build-arg SERVICE_NAME=auth \
  --build-arg PORT=8081 \
  -t eomp/auth:latest \
  -f deploy/docker/Dockerfile.go-service .
```

### 2.2 Frontend Web Nuxt 4 Dockerfile
Được thiết kế tại [`deploy/docker/Dockerfile.web`](file:///d:/IT_help/eomp/deploy/docker/Dockerfile.web):
* **Stage 1 (Builder):** `node:22-alpine` chạy `pnpm build` với Nitro SSR engine.
* **Stage 2 (Runner):** Node runtime tối giản chạy non-root user, kích thước nhẹ và bảo mật cao.

```bash
# Build Frontend Web
docker build -t eomp/web:latest -f deploy/docker/Dockerfile.web .
```

---

## 3. Triển Khai Production Docker Compose

Tệp cấu hình [`deploy/docker-compose.prod.yml`](file:///d:/IT_help/eomp/deploy/docker-compose.prod.yml) cung cấp toàn bộ môi trường Production khép kín:

```bash
# Khởi động toàn bộ cụm dịch vụ Production (11 Services + Web + Nginx + DBs)
docker compose -f deploy/docker-compose.prod.yml up -d

# Xem log thời gian thực
docker compose -f deploy/docker-compose.prod.yml logs -f

# Dừng hệ thống
docker compose -f deploy/docker-compose.prod.yml down
```

---

## 4. Triển Khai Kubernetes Manifests (Native K8s)

Bộ tài nguyên Kubernetes được phân tầng rõ ràng trong [`deploy/kubernetes/manifests/`](file:///d:/IT_help/eomp/deploy/kubernetes/manifests/):

| Thứ Tự Áp Dụng | File Manifest | Trách Nhiệm Kỹ Thuật |
|:---:|---|---|
| `00` | `00-namespace.yaml` | Tạo Namespace cô lập `eomp` |
| `01` | `01-configmaps.yaml` | Cấu hình biến môi trường toàn cục, Gateway và Frontend Web |
| `02` | `02-secrets.yaml` | Quản lý an toàn JWT Secret, Mật khẩu Database, MinIO, RabbitMQ |
| `03` | `03-pvcs.yaml` | Khai báo PersistentVolumeClaims cho Postgres, Redis, RabbitMQ, MinIO, Qdrant, Prometheus, Grafana, Loki |
| `04` | `04-infrastructure.yaml` | Deployments & ClusterIP Services cho tất cả cơ sở dữ liệu và công cụ giám sát |
| `05` | `05-microservices-deployments.yaml` | Deployments cho 11 Go Microservices + Web (kèm Probes, Limits, Non-root UID) |
| `06` | `06-microservices-services.yaml` | ClusterIP Services kết nối nội bộ giữa các microservices |
| `07` | `07-hpa.yaml` | Horizontal Pod Autoscaler tự động co giãn từ 2 đến 10 Pods |
| `08` | `08-ingress.yaml` | Ingress Nginx Controller phân luồng và bảo vệ Edge API |

### Lệnh Triển Khai:
```bash
# Áp dụng toàn bộ Manifests
kubectl apply -f deploy/kubernetes/manifests/

# Kiểm tra trạng thái toàn bộ Pods
kubectl get pods -n eomp

# Kiểm tra trạng thái HPA và Ingress
kubectl get hpa,ingress -n eomp
```

---

## 5. Triển Khai Bằng Kubernetes Helm Chart

Helm Chart tiêu chuẩn Production tại [`deploy/kubernetes/helm/eomp/`](file:///d:/IT_help/eomp/deploy/kubernetes/helm/eomp/):

```bash
# 1. Kiểm tra tính hợp lệ của Chart
helm lint deploy/kubernetes/helm/eomp

# 2. Render thử nghiệm cấu hình Kubernetes
helm template eomp-release deploy/kubernetes/helm/eomp

# 3. Cài đặt hoặc Nâng cấp hệ thống 1-click vào namespace 'eomp'
helm upgrade --install eomp deploy/kubernetes/helm/eomp --namespace eomp --create-namespace

# 4. Gỡ cài đặt khi cần thiết
helm uninstall eomp --namespace eomp
```

---

## 6. Cấu Hình Auto-scaling (HPA) & Ingress Controller

### 6.1 Horizontal Pod Autoscaler (HPA v2)
Các dịch vụ chịu tải cao (`gateway`, `auth`, `helpdesk`, `asset`, `workflow`, `web`) được thiết lập HPA v2:
* **Min Replicas:** 2 Pods
* **Max Replicas:** 8 - 10 Pods
* **Ngưỡng kích hoạt:**
  * CPU Utilization: `> 70%`
  * Memory Utilization: `> 80%`

### 6.2 Nginx Ingress Controller
* Cổng vào duy nhất với hostname `eomp.local`.
* Định tuyến thông minh:
  * `/api/*` -> `gateway-service:8080` (Hỗ trợ SSE/WebSocket stream)
  * `/monitoring/grafana/*` -> `grafana-service:3000`
  * `/*` -> `web-service:3000` (Frontend Nuxt 4)
* Bảo vệ an ninh: CORS Headers, Body Size Limit (50MB), Rate Limit (100 RPS).

---

## 7. Health Probes & Giám Sát Vận Hành (Liveness/Readiness)

Mỗi Pod được giám sát liên tục qua 2 đầu dò sức khỏe:
* **Liveness Probe:** `httpGet: /health` (Port Service) -> Khởi động lại Pod nếu gặp sự cố treo Deadlock.
* **Readiness Probe:** `httpGet: /health` -> Đảm bảo Pod chỉ nhận Traffic khi kết nối DB / Cache đã sẵn sàng.

---

## 8. CLI Helper Scripts & CI/CD Pipeline

EOMP cung cấp CLI tự động hóa giúp kiểm tra và thao tác nhanh chóng:

### Dành cho Windows (PowerShell):
```powershell
# Xem danh mục lệnh
.\scripts\deploy.ps1 help

# Kiểm tra cú pháp toàn bộ manifests
.\scripts\deploy.ps1 validate

# Khởi động Docker Compose Production
.\scripts\deploy.ps1 prod-up

# Áp dụng K8s manifests
.\scripts\deploy.ps1 k8s-apply

# Cài đặt qua Helm
.\scripts\deploy.ps1 helm-install
```

### Dành cho Linux / macOS / CI (Bash):
```bash
chmod +x ./scripts/deploy.sh
./scripts/deploy.sh validate
./scripts/deploy.sh prod-up
./scripts/deploy.sh k8s-apply
./scripts/deploy.sh helm-install
```
