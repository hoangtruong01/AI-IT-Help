# EOMP — Disaster Recovery (DR) Plan & Business Continuity Strategy

> **Tài liệu Kế Hoạch Phục Hồi Thảm Họa & Duy Trì Hoạt Động Doanh Nghiệp (DR & BCP)**  
> **Áp dụng cho:** Site Reliability Engineers (SRE), Cloud Architects, DevOps & SysAdmins.  
> **Tiêu chuẩn cam kết:** RPO < 5 phút | RTO < 15 phút  
> **Phiên bản:** Phase 14 Enterprise Master Standard

---

## 📑 MỤC LỤC
1. [Mục Tiêu & Chỉ Số Cam Kết (RPO / RTO Metrics)](#1-mục-tiêu--chỉ-số-cam-kết-rpo--rto-metrics)
2. [Ma Trận Phân Loại Thảm Họa (Disaster Classification Matrix)](#2-ma-trận-phân-loại-thảm-họa-disaster-classification-matrix)
3. [Chiến Lược Sao Lưu 8 Cơ Sở Dữ Liệu (Multi-DB Backup Strategy)](#3-chiến-lược-sao-lưu-8-cơ-sở-dữ-liệu-multi-db-backup-strategy)
4. [Quy Trình Phục Hồi Dữ Liệu Từng Thời Điểm (Point-in-Time Recovery - PITR)](#4-quy-trình-phục-hồi-dữ-liệu-từng-thời-điểm-point-in-time-recovery---pitr)
5. [Quy Trình Chuyển Vùng Sự Cố (Cross-AZ / Multi-Region Failover)](#5-quy-trình-chuyển-vùng-sự-cố-cross-az--multi-region-failover)
6. [Kế Hoạch Kiểm Tra Định Kỳ & Diễn Tập Phục Hồi (DR Drills Schedule)](#6-kế-hoạch-kiểm-tra-định-kỳ--diễn-tập-phục-hồi-dr-drills-schedule)

---

## 1. Mục Tiêu & Chỉ Số Cam Kết (RPO / RTO Metrics)

Hệ sinh thái EOMP cam kết các chỉ số khôi phục thảm họa ở cấp độ doanh nghiệp (Tier-1 Enterprise Critical Service):

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                            TIMELINE XẢY RA SỰ CỐ                            │
├───────────────────────────────┬─────────────────────────────────────────────┤
│         ◄── RPO ──►           │               ◄────── RTO ──────►           │
│   (Mất tối đa < 5 phút data)  │        (Hệ thống hoạt động trở lại < 15m)   │
├───────────────────────────────┼─────────────────────────────────────────────┤
│  Bản sao lưu / WAL gần nhất   │  Sự cố xảy ra              Hệ thống Online  │
│          [ 10:55 ]            │   [ 11:00 ]                   [ 11:15 ]     │
└───────────────────────────────┴─────────────────────────────────────────────┘
```

* **Recovery Point Objective (RPO) `< 5 phút`:** Lượng dữ liệu tối đa chấp nhận mất mát khi có sự cố thảm họa. Đạt được nhờ cơ chế **Continuous WAL Archiving** (Write-Ahead Logging) ghi nhận tức thì các transaction vào MinIO/S3 mỗi 60 giây.
* **Recovery Time Objective (RTO) `< 15 phút`:** Thời gian tối đa để khôi phục toàn bộ 11 Go microservices, Nuxt 4 Web Frontend, và cơ sở dữ liệu trở lại trạng thái sẵn sàng phục vụ người dùng.

---

## 2. Ma Trận Phân Loại Thảm Họa (Disaster Classification Matrix)

| Cấp Độ Thảm Họa | Phạm Vi Ảnh Hưởng | Kịch Bản Điển Hình | Biện Pháp Khắc Phục Tương Ứng | RTO Cam Kết |
|:---:|---|---|---|:---:|
| **Level 1 (Data Corruption)** | 1 Database cục bộ | Nhân viên vô tình drop bảng hoặc lỗi ghi dữ liệu | Point-in-Time Recovery (PITR) từ WAL Log gần nhất | `< 10 phút` |
| **Level 2 (Node Failure)** | 1 hoặc nhiều K8s Pods / Worker Node | Node K8s bị Kernel Panic hoặc cạn RAM (OOM) | Kubernetes Auto-healing, HPA co giãn, Pod rescheduling | `< 2 phút` |
| **Level 3 (Storage Failure)** | Toàn bộ Persistent Volumes | Ổ cứng SAN/EBS bị lỗi vật lý hoặc hỏng filesystem | Khôi phục Volume Snapshot + Đồng bộ WAL Log từ MinIO/S3 | `< 12 phút` |
| **Level 4 (Datacenter Down)** | Toàn bộ Datacenter / Cloud AZ | Mất điện diện rộng hoặc đứt cáp quang biển | Chuyển hướng DNS Ingress sang Standby Cluster (Cross-AZ Failover) | `< 15 phút` |

---

## 3. Chiến Lược Sao Lưu 8 Cơ Sở Dữ Liệu (Multi-DB Backup Strategy)

EOMP quản lý 8 cơ sở dữ liệu PostgreSQL phân tán cô lập (`auth_db`, `employee_db`, `asset_db`, `helpdesk_db`, `workflow_db`, `notification_db`, `knowledge_db`, `reporting_db`, `audit_db`).

```
                              ┌─────────────────────────┐
                              │  PostgreSQL Database    │
                              └────────────┬────────────┘
                                           │
                    ┌──────────────────────┴──────────────────────┐
                    │                                             │
                    ▼                                             ▼
        ┌───────────────────────┐                     ┌───────────────────────┐
        │ Daily Full Backup     │                     │ Continuous WAL Stream │
        │ (pg_dump / pg_basebackup)                   │ (archive_command)     │
        │ Tần suất: 00:00 UTC   │                     │ Tần suất: Mỗi 60 giây │
        └───────────┬───────────┘                     └───────────┬───────────┘
                    │                                             │
                    └──────────────────────┬──────────────────────┘
                                           ▼
                              ┌─────────────────────────┐
                              │ MinIO / AWS S3 Storage  │
                              │ (Encrypted & Immutable) │
                              └─────────────────────────┘
```

1. **Daily Full Backup:** Tự động xuất file dump nén `.sql.gz` cho cả 8 DBs vào lúc `00:00 UTC` mỗi ngày, mã hóa AES-256, lưu trữ tối thiểu 30 ngày.
2. **Continuous WAL Archiving:** Các tệp Write-Ahead Log được đẩy liên tục lên MinIO Bucket `local/backups/wal/` mỗi 60 giây.
3. **Qdrant Vector Snapshots:** Tự động tạo snapshot của các bộ vector tài liệu RAG mỗi 6 giờ.

---

## 4. Quy Trình Phục Hồi Dữ Liệu Từng Thời Điểm (Point-in-Time Recovery - PITR)

Khi xảy ra sự cố sai lệch dữ liệu vào thời điểm $T$, quản trị viên SRE thực hiện khôi phục chính xác về thời điểm $T - 1\text{ phút}$:

```bash
# Bước 1: Khởi tạo container Postgres ở chế độ Recovery
# Bước 2: Tải bản Full Backup gần nhất trước thời điểm T
gunzip -c /backups/eomp_full_backup_latest.sql.gz | psql -U eomp -d eomp

# Bước 3: Cấu hình recovery_target_time trong postgresql.conf:
# restore_command = 'mc cp local/backups/wal/%f %p'
# recovery_target_time = '2026-08-21 15:30:00 UTC'
# recovery_target_action = 'promote'

# Bước 4: Khởi động lại dịch vụ và xác minh tính toàn vẹn Checksum của bảng audit_logs
```

---

## 5. Quy Trình Chuyển Vùng Sự Cố (Cross-AZ / Multi-Region Failover)

Khi toàn bộ cụm Primary Cluster tại Region A gặp thảm họa không thể phục hồi:

```
[Người Dùng] ──► [Cloudflare / Route53 DNS (Health Check Fail)]
                               │
               ┌───────────────┴───────────────┐
               ▼ (Chuyển tiếp sau 60s)         ▼
      [Region A - OFFLINE]            [Region B - Secondary Standby]
      (Cụm máy chủ gặp sự cố)         ├── Ingress Gateway
                                      ├── 11 Go Microservices (HPA: 2-10)
                                      ├── Nuxt 4 Web Frontend
                                      └── PostgreSQL Standby (Đã promote thành Primary)
```

### Các Bước Thực Hiện Failover:
1. **Phát Hiện Sự Cố:** Prometheus & Uptime Probe báo động `Cluster_Down_Alarm` sau 3 lần ping thất bại liên tiếp (15 giây).
2. **Promote Standby Database:** Kích hoạt Standby DB tại Region B thành Primary Database (`SELECT pg_promote();`).
3. **Chuyển Tuyến DNS:** Cập nhật bản ghi DNS `eomp.local` hoặc Public IP sang Ingress Controller của Region B qua Cloudflare API.
4. **Kiểm Tra Khởi Động:** Chạy tự động `scripts/deploy.ps1 validate` và health check `/health` trên toàn bộ 11 microservices.
5. **Thông Báo Hoàn Tất:** Gửi cảnh báo xác nhận khôi phục thành công tới kênh Slack `#sre-alerts` và IT Director.

---

## 6. Kế Hoạch Kiểm Tra Định Kỳ & Diễn Tập Phục Hồi (DR Drills Schedule)

* **Hàng Tuần:** Tự động kiểm tra tính toàn vẹn của tệp backup bằng script `scripts/backup_restore.ps1 test-restore`.
* **Hàng Tháng:** Diễn tập giả lập rớt 1 Database ngẫu nhiên trong môi trường Staging (Chaos Drill).
* **Hàng Quý:** Diễn tập toàn diện kịch bản chuyển vùng Cross-Region Failover với sự tham gia của toàn bộ đội ngũ SRE và QA.
