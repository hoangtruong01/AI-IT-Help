# EOMP — Incident Response Playbook & On-Call Escalation Matrix

> **Cẩm Nang Ứng Phó Khẩn Cấp & Ma Trận Xử Lý Sự Cố Hệ Thống (SRE Incident Playbook)**  
> **Áp dụng cho:** Incident Commanders, SRE Engineers, DevOps, Tech Leads & Developers.  
> **Mục tiêu:** Giảm thiểu tối đa MTTR (Mean Time to Resolve) và chuẩn hóa quy trình điều phối sự cố.

---

## 📑 MỤC LỤC
1. [Phân Cấp Mức Độ Nghiêm Trọng (Severity Levels SEV 1-4)](#1-phân-cấp-mức-độ-nghiêm-trọng-severity-levels-sev-1-4)
2. [Quy Trình 5 Bước Ứng Phó Khẩn Cấp (Incident Lifecycle)](#2-quy-trình-5-bước-ứng-phó-khẩn-cấp-incident-lifecycle)
3. [Phân Công Vai Trò Trong Phòng Tác Chiến (War Room Roles)](#3-phân-công-vai-trò-trong-phòng-tác-chiến-war-room-roles)
4. [Sổ Tay Xử Lý Nhanh Các Sự Cố Điển Hình (Quick Diagnostic Cheatsheet)](#4-sổ-tay-xử-lý-nhanh-các-sự-cố-điển-hình-quick-diagnostic-cheatsheet)
5. [Mẫu Biên Bản Phân Tích Nguyên Nhân Gốc Rễ (Post-Mortem & RCA 5-Whys)](#5-mẫu-biên-bản-phân-tích-nguyên-nhân-gốc-rễ-post-mortem--rca-5-whys)

---

## 1. Phân Cấp Mức Độ Nghiêm Trọng (Severity Levels SEV 1-4)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                       INCIDENT SEVERITY HIERARCHY                           │
├─────────┬──────────────────────┬──────────────────────┬─────────────────────┤
│  Level  │       Tên Gọi        │   Mức Độ Tác Động    │   Thời Gian Phản Hồi│
├─────────┼──────────────────────┼──────────────────────┼─────────────────────┤
│  SEV-1  │ CRITICAL OUTAGE      │ Toàn bộ nền tảng sập │ < 5 phút (24/7/365) │
│  SEV-2  │ MAJOR DEGRADATION    │ Module chính bị lỗi  │ < 15 phút           │
│  SEV-3  │ MINOR ISSUE          │ Suy giảm hiệu năng nhẹ│ < 2 giờ             │
│  SEV-4  │ LOW / COSMETIC       │ Lỗi giao diện nhỏ    │ < 24 giờ            │
└─────────┴──────────────────────┴──────────────────────┴─────────────────────┘
```

---

## 2. Quy Trình 5 Bước Ứng Phó Khẩn Cấp (Incident Lifecycle)

```
  [1. Triage]  ──►  [2. Mobilize]  ──►  [3. Mitigate]  ──►  [4. Resolve]  ──►  [5. Learn]
(Phát hiện qua      (Mở War Room,        (Khôi phục       (Verify hệ thống      (Post-mortem,
  Prometheus)         phân vai)          tạm thời)         ổn định 100%)       RCA 5-Whys)
```

1. **Bước 1: Triage (Phát hiện & Phân loại):** Cảnh báo từ Prometheus Alertmanager bắn vào kênh Slack `#sre-alerts` hoặc PagerDuty gọi điện On-call Engineer. Xác định Sev level trong 3 phút.
2. **Bước 2: Mobilize (Thiết lập Phòng tác chiến):** Đối với SEV-1/SEV-2, mở ngay cuộc gọi Google Meet / Zoom khẩn cấp (War Room). Khóa quyền deploy tạm thời.
3. **Bước 3: Mitigate (Hạ nhiệt sự cố):** Ưu tiên khôi phục dịch vụ cho người dùng (Rollback phiên bản, mở Circuit Breaker, tăng Replicas HPA, kích hoạt Fallback mode).
4. **Bước 4: Resolve (Khắc phục triệt để):** Sửa lỗi mã nguồn hoặc database, triển khai hotfix và theo dõi p95 Latency & Error Rate trở về mức an toàn (< 1%).
5. **Bước 5: Learn (Đúc rút kinh nghiệm):** Soạn thảo tài liệu Post-Mortem và tạo Action Items trong vòng 48 giờ.

---

## 3. Phân Công Vai Trò Trong Phòng Tác Chiến (War Room Roles)

| Vai Trò | Người Đảm Nhiệm | Trách Nhiệm Chính |
|---|---|---|
| **Incident Commander (IC)** | Lead SRE / Tech Lead | Quyết định hướng giải quyết cao nhất, điều phối nhân sự, không trực tiếp gõ lệnh debug |
| **Operations Lead (OL)** | Senior Backend / DevOps | Trực tiếp kiểm tra logs, trace metrics, chạy script rollback / restart containers |
| **Communications Lead (CL)**| Product Manager / BA | Cập nhật thông báo trạng thái (Status Page) cho khách hàng và ban giám đốc mỗi 15 phút |
| **Scribe** | QA Engineer | Ghi nhận từng mốc thời gian (Timeline) và các thao tác đã thực hiện vào biên bản |

---

## 4. Sổ Tay Xử Lý Nhanh Các Sự Cố Điển Hình (Quick Diagnostic Cheatsheet)

### Kịch Bản A: Database Connection Pool Cạn Kiệt (`Too many clients`)
```bash
# Kiểm tra số lượng kết nối đang mở
docker compose exec postgres psql -U eomp -d eomp -c "SELECT count(*), state FROM pg_stat_activity GROUP BY state;"

# Hủy các query bị treo quá 30 giây
docker compose exec postgres psql -U eomp -d eomp -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE state = 'active' AND query_start < now() - interval '30 seconds';"
```

### Kịch Bản B: RabbitMQ Queue Backlog Tăng Đột Biến
```bash
# Kiểm tra danh sách Queue và số message tồn đọng
docker compose exec rabbitmq rabbitmqctl list_queues name messages consumers

# Xóa message lỗi chuyển vào Dead-Letter Queue (DLQ)
# Kích hoạt thêm worker consumer pods qua kubectl scale
kubectl scale deployment notification-deployment --replicas=5 -n eomp
```

### Kịch Bản C: Microservice bị OOMKilled (Out Of Memory)
```bash
# Xem log lý do crash
kubectl describe pod -l app=helpdesk -n eomp

# Tạm thời nâng giới hạn RAM limit trong Config
kubectl set resources deployment helpdesk-deployment --limits=memory=512Mi -n eomp
```

---

## 5. Mẫu Biên Bản Phân Tích Nguyên Nhân Gốc Rễ (Post-Mortem & RCA 5-Whys)

```markdown
# 📋 Incident Post-Mortem Report

- **Mã Sự Cố:** INC-2026-0821-SEV1
- **Thời Gian Bắt Đầu:** 2026-08-21 14:10 UTC
- **Thời Gian Khắc Phục:** 2026-08-21 14:22 UTC (MTTR: 12 phút)
- **Chỉ Huy Sự Cố (IC):** Senior SRE Lead
- **Dịch Vụ Bị Ảnh Hưởng:** `services/gateway`, `services/helpdesk`

### 1. Tóm Tắt Sự Cố
Người dùng không thể tạo mới Ticket sự cố qua Web Frontend, nhận mã lỗi 504 Gateway Timeout.

### 2. Phân Tích 5-Whys (Root Cause Analysis)
1. *Tại sao Gateway trả về mã 504?* -> Vì Helpdesk Service không phản hồi request trong 30 giây.
2. *Tại sao Helpdesk Service không phản hồi?* -> Vì tất cả 25 connection trong Postgres DB Pool đều bận.
3. *Tại sao DB connection pool bị chiếm dụng?* -> Do một câu truy vấn tìm kiếm Full-Text Search thiếu Index.
4. *Tại sao câu truy vấn thiếu Index lại lọt vào Production?* -> Do migration script chưa được kiểm thử tải trước khi release.
5. *Tại sao kiểm thử tải không phát hiện?* -> Thiếu bước chạy K6 benchmark trong pipeline CI/CD đối với endpoint mới.

### 3. Biện Pháp Khắc Phục Ngăn Chặn Tái Diễn (Action Items)
- [x] Thêm Index `idx_tickets_search` vào bảng `tickets` (Hoàn thành ngay).
- [ ] Tích hợp K6 Automated Performance Gate vào Jenkinsfile trước khi merge vào `main` (Hạn chót: 3 ngày).
- [ ] Thiết lập Prometheus Alert cảnh báo khi DB Connection Pool sử dụng vượt quá 80% (Hạn chót: 1 ngày).
```
