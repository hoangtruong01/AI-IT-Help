# EOMP — Enterprise Documentation Hub

> **Trung Tâm Tài Liệu Kỹ Thuật & Kiến Trúc Nền Tảng EOMP (Single Source of Truth)**  
> **Áp dụng cho:** Architects, Tech Leads, Full-stack Developers, QA/QC Engineers, DevOps & SRE.  
> **Trạng thái:** 14/14 Phases Hoàn Thành Trọn Vẹn (Master Release v1.0.0)

---

## 🌟 Tài Liệu Khuyên Đọc Đầu Tiên (Recommended Entry Points)

1. 📖 **[Developer & Intern Comprehensive Guide (INTERN_DEVELOPER_GUIDE.md)](INTERN_DEVELOPER_GUIDE.md)**  
   *Cẩm nang toàn diện từ A-Z: Nghiệp vụ kinh doanh, kiến trúc 11 microservices, Clean Architecture, Frontend Nuxt 4, hạ tầng Docker/K8s, CI/CD Jenkins, hướng dẫn code tính năng mới, testing và troubleshooting.*
2. 📋 **[Project Structure, Implementation Standards & Daily Changelog (PROJECT_STRUCTURE_AND_CHANGELOG.md)](PROJECT_STRUCTURE_AND_CHANGELOG.md)**  
   *Quy chuẩn monorepo, response envelope contracts, port matrix, và nhật ký chi tiết các thay đổi từ Phase 1 đến Phase 14.*
3. 🎯 **[Phase 6 to 14 Multi-Role Roadmap Specification (PHASE_6_TO_14_ROADMAP_SPECIFICATION.md)](PHASE_6_TO_14_ROADMAP_SPECIFICATION.md)**  
   *Đặc tả kỹ thuật đa vai trò dành cho BA, UI/UX Designer, QA/QC, Testers & Developers.*

---

## 📚 Danh Mục Tài Liệu Theo Phân Hệ Chuyên Môn

### 1. 🏛️ Thiết Kế Kiến Trúc & Dữ Liệu (Architecture & Data Hub)
* **[architecture.md](architecture.md)** — Tổng quan kiến trúc hệ thống, ranh giới domain và phương thức giao tiếp REST/EventBus.
* **[architecture/c4_model_diagrams.md](architecture/c4_model_diagrams.md)** — Bộ sơ đồ thiết kế kiến trúc **C4 Model** (Context, Container, Component, Dynamic Lifecycle Sequence).
* **[database.md](database.md)** — Chiến lược cơ sở dữ liệu phân tán và công cụ quản lý migration tự động.
* **[architecture/database_erd_and_data_dictionary.md](architecture/database_erd_and_data_dictionary.md)** — Sơ đồ thực thể quan hệ Master ERD và Từ điển dữ liệu chi tiết của 8 cơ sở dữ liệu phân tán.
* **[environment.md](environment.md)** — Bảng tra cứu toàn bộ các biến môi trường cấu hình trong hệ thống (`.env`).

### 2. 🔌 Giao Tiếp & Đặc Tả API (API & Interface Hub)
* **[api.md](api.md)** — Danh mục và đặc tả chi tiết các REST API endpoints của 11 microservices qua API Gateway.
* **[openapi/eomp-openapi-spec.yaml](openapi/eomp-openapi-spec.yaml)** — Tệp đặc tả chuẩn **OpenAPI 3.0 (Swagger)** của toàn bộ hệ thống.
* **[openapi/README.md](openapi/README.md)** — Hướng dẫn tra cứu và tích hợp tài liệu Swagger UI trực quan.

### 3. 🧪 Phát Triển & Kiểm Thử Tự Động (Engineering & Quality Assurance)
* **[setup.md](setup.md)** — Hướng dẫn cài đặt nhanh môi trường phát triển cục bộ (Local Development).
* **[development.md](development.md)** — Quy tắc phân nhánh Git, commit message chuẩn, convention viết code Go & Vue.
* **[testing.md](testing.md)** — Chiến lược kiểm thử tự động toàn diện: Unit test, Linting, Typecheck, Cross-Service E2E Lifecycle và K6 Load Testing.

### 4. 🚢 Đóng Gói, Triển Khai & Vận Hành SRE (Deployment, SRE & DR)
* **[deployment.md](deployment.md)** — Hướng dẫn đóng gói Docker Multi-stage (<25MB), Production Compose, Kubernetes Native Manifests, và Helm Charts.
* **[sre/disaster_recovery_plan.md](sre/disaster_recovery_plan.md)** — Kế hoạch phục hồi thảm họa chi tiết cam kết **RPO < 5 phút** và **RTO < 15 phút**, PITR và Cross-AZ Failover.
* **[sre/incident_response_playbook.md](sre/incident_response_playbook.md)** — Cẩm nang ứng phó khẩn cấp SEV 1-4, quy trình War Room và mẫu Post-Mortem 5-Whys.
* **[sre/operations_manual.md](sre/operations_manual.md)** — Sổ tay vận hành Day-2, nâng cấp DB migration, xoay vòng Secrets và Zero-downtime updates.
* **[sre/chaos_engineering_runbook.md](sre/chaos_engineering_runbook.md)** — Hướng dẫn thực thi các kịch bản kiểm thử giả lập sự cố (Postgres Outage, Queue Jam, Pod crash).
* **[sre/project_handover_acceptance.md](sre/project_handover_acceptance.md)** — Biên bản nghiệm thu kỹ thuật toàn diện tổng kết 100% hoàn thành 14 Phases.

---

## 🛠️ Danh Mục Lệnh Automation CLI

```powershell
# Dành cho Windows (PowerShell)
.\scripts\dev.ps1 help            # Lệnh dành cho Developer (docker-up, build, test, format)
.\scripts\qa.ps1                  # Chạy toàn bộ 6 tầng QA/QC kiểm thử tự động
.\scripts\deploy.ps1 validate     # Kiểm tra cú pháp Kubernetes manifests & Helm templates
.\scripts\deploy.ps1 prod-up      # Khởi động cụm Production Docker Compose
.\scripts\deploy.ps1 k8s-apply    # Triển khai vào Kubernetes cluster
.\scripts\deploy.ps1 helm-install # Cài đặt bằng Helm Chart vào namespace 'eomp'
.\scripts\chaos.ps1 run-all-chaos # Thực thi kịch bản kiểm thử giả lập sự cố (Chaos test)
.\scripts\backup_restore.ps1      # Tự động sao lưu và khôi phục 8 databases
```

```bash
# Dành cho Linux / macOS / CI (Bash & Makefile)
make docker-up                    # Khởi động hạ tầng
make test                         # Chạy kiểm thử unit tests
./scripts/deploy.sh prod-up       # Chạy Production Compose
./scripts/deploy.sh k8s-apply     # Triển khai K8s
./scripts/chaos.sh run-all-chaos  # Chạy Chaos drill
```
