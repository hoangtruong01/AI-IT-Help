# EOMP — Chaos Engineering Runbook & Fault Injection Scenarios

> **Cẩm Nang Kiểm Thử Giả Lập Sự Cố & Đánh Giá Độ Kiên Cường Hệ Thống (Resilience Testing)**  
> **Áp dụng cho:** SRE, Chaos Engineers, QA/QC Automation & Backend Leads.  
> **Mục tiêu:** Chứng minh hệ thống tự động phục hồi (Self-healing), không bị sập dây chuyền (Cascading Failure), và đảm bảo Zero-Downtime.

---

## 📑 MỤC LỤC
1. [Nguyên Tắc Kiểm Thử Chaos (Principles of Chaos Engineering)](#1-nguyên-tắc-kiểm-thử-chaos-principles-of-chaos-engineering)
2. [Kịch Bản 1: Giả Lập Mất Kết Nối PostgreSQL (Database Outage)](#2-kịch-bản-1-giả-lập-mất-kết-nối-postgresql-database-outage)
3. [Kịch Bản 2: Giả Lập Tắc Nghẽn Hàng Đợi RabbitMQ (Queue Backpressure & DLQ)](#3-kịch-bản-2-giả-lập-tắc-nghẽn-hàng-đợi-rabbitmq-queue-backpressure--dlq)
4. [Kịch Bản 3: Giả Lập Đột Tử Pods Liên Tục (Pod CrashLoop & Eviction)](#4-kịch-bản-3-giả-lập-đột-tử-pods-liên-tục-pod-crashloop--eviction)
5. [Kịch Bản 4: Giả Lập Nghẽn Mạng & Độ Trễ Cao (Network Latency Injection)](#5-kịch-bản-4-giả-lập-nghẽn-mạng--độ-trễ-cao-network-latency-injection)
6. [Tự Động Hóa Thực Thi Kịch Bản Chaos (Automated Chaos CLI)](#6-tự-động-hóa-thực-thi-kịch-bản-chaos-automated-chaos-cli)

---

## 1. Nguyên Tắc Kiểm Thử Chaos (Principles of Chaos Engineering)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       CHAOS EXPERIMENT LIFECYCLE                            │
├─────────────────────────────────────────────────────────────────────────────┤
│ 1. Xác định Trạng thái Ổn định (Steady State: Error Rate < 1%, p95 < 200ms) │
│ 2. Đưa ra Giả thuyết (Hypothesis: Nếu DB tạm ngắt, Service sẽ Retry an toàn)│
│ 3. Tiêm Nhiễm Sự Cố (Inject Failure: Dừng container hoặc drop gói tin)       │
│ 4. Đo lường & Xác minh (Verify: UI không crash, Alert bắn đúng thời gian)   │
│ 5. Khôi phục & Đúc rút (Restore & Fix Weaknesses)                           │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 2. Kịch Bản 1: Giả Lập Mất Kết Nối PostgreSQL (Database Outage)

* **Giả thuyết:** Khi PostgreSQL bị tạm dừng, các services chuyển sang trạng thái chờ với Exponential Backoff Retry. Gateway trả về mã `503 Service Unavailable` có cấu trúc mà không bị Crash Panic. Khi DB khởi động lại, toàn bộ services tự động kết nối lại thành công mà không cần restart container.
* **Lệnh tiêm lỗi:**
  ```bash
  # Tạm dừng container Postgres
  docker pause eomp-postgres || docker pause eomp-prod-postgres
  ```
* **Kỳ vọng nghiệm thu:**
  - Health Probe `/health` chuyển sang trạng thái `UNHEALTHY` trong vòng 5 giây.
  - SRE Dashboard tại `/monitoring` hiển thị đèn đỏ cảnh báo.
  - Không có goroutine nào bị leak (Memory ổn định).
* **Lệnh khôi phục:**
  ```bash
  docker unpause eomp-postgres || docker unpause eomp-prod-postgres
  ```

---

## 3. Kịch Bản 2: Giả Lập Tắc Nghẽn Hàng Đợi RabbitMQ (Queue Backpressure & DLQ)

* **Giả thuyết:** Khi RabbitMQ bị dừng đột ngột hoặc consumer bị tắc nghẽn, các Event CloudEvents được lưu đệm trong In-Memory Buffer hoặc gửi sang Dead-Letter Queue (DLQ) mà không làm mất mát dữ liệu nghiệp vụ (Zero Message Loss).
* **Lệnh tiêm lỗi:**
  ```bash
  docker stop eomp-rabbitmq || docker stop eomp-prod-rabbitmq
  ```
* **Kỳ vọng nghiệm thu:**
  - Ticket và Asset vẫn được tạo thành công trong Database.
  - Notification Service tự động reconnect với RabbitMQ sau khi khởi động lại và tiêu thụ hết hàng đợi dồn ứ.

---

## 4. Kịch Bản 3: Giả Lập Đột Tử Pods Liên Tục (Pod CrashLoop & Eviction)

* **Giả thuyết:** Khi một Pod của `helpdesk` hoặc `gateway` bị `kill -9`, Kubernetes Ingress Controller tự động điều hướng 100% traffic sang Pod replica còn lại mà người dùng không gặp bất kỳ lỗi gián đoạn nào (Zero Downtime).
* **Lệnh tiêm lỗi:**
  ```bash
  kubectl delete pod -l app=helpdesk -n eomp --now
  ```
* **Kỳ vọng nghiệm thu:**
  - K6 Load Test duy trì Success Rate `> 99.9%`.
  - Kubernetes ReplicaSet tự động sinh Pod thay thế trong `< 3 giây`.

---

## 5. Kịch Bản 4: Giả Lập Nghẽn Mạng & Độ Trễ Cao (Network Latency Injection)

* **Giả thuyết:** Khi mạng giữa AI Service và Knowledge Base bị trễ thêm 1500ms, RAG SmartRetriever tự động kích hoạt **Fallback In-Memory Catalog** đảm bảo giao diện phản hồi trong `< 800ms`.
* **Lệnh tiêm lỗi:**
  ```bash
  # Thêm 2000ms latency vào card mạng container
  docker compose exec ai tc qdisc add dev eth0 root netem delay 2000ms
  ```
* **Kỳ vọng nghiệm thu:**
  - AI Copilot vẫn phản hồi câu trả lời hợp lệ từ Local Catalog.

---

## 6. Tự Động Hóa Thực Thi Kịch Bản Chaos (Automated Chaos CLI)

Sử dụng bộ công cụ được tích hợp sẵn:
```powershell
# Chạy mô phỏng kiểm thử Chaos trên Windows
.\scripts\chaos.ps1 simulate-db-down
.\scripts\chaos.ps1 restore-db
.\scripts\chaos.ps1 run-all-chaos
```

```bash
# Chạy trên Linux / macOS / CI
./scripts/chaos.sh run-all-chaos
```
