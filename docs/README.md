# EOMP Documentation Hub

Chào mừng bạn đến với trung tâm tài liệu kỹ thuật của **EOMP (Enterprise Operations Management Platform)**.

Tất cả tài liệu được phân chia theo từng chủ đề chuyên sâu nhằm phục vụ việc nghiên cứu, phát triển, kiểm thử và vận hành hệ thống.

---

## 🌟 Tài Liệu Khuyên Đọc Đầu Tiên Cho Developer / Intern

👉 **[Tài Liệu Hướng Dẫn Toàn Diện Cho Developer & Intern (INTERN_DEVELOPER_GUIDE.md)](INTERN_DEVELOPER_GUIDE.md)**  
*Bản hướng dẫn chi tiết từ A-Z: bài toán kinh doanh, kiến trúc microservices, cấu trúc thư mục, hướng dẫn cài đặt, phân tích 11 microservices, kiến trúc Frontend Nuxt 4, hạ tầng Docker, CI/CD Jenkins, hướng dẫn code tính năng mới, testing và troubleshooting.*

---

## 📚 Danh Mục Tài Liệu Chi Tiết Theo Chủ Đề

| File Tài Liệu | Nội Dung Chính | Đối Tượng |
|---|---|---|
| **[PROJECT_STRUCTURE_AND_CHANGELOG.md](PROJECT_STRUCTURE_AND_CHANGELOG.md)** | **Quy chuẩn kiến trúc, cấu trúc file dự án & nhật ký thay đổi chi tiết** | **Tất cả AI, Tech Lead, Developers** |
| **[INTERN_DEVELOPER_GUIDE.md](INTERN_DEVELOPER_GUIDE.md)** | Cẩm nang toàn diện tổng hợp mọi khía cạnh hệ thống | Developer mới, Intern, Tech Lead |
| **[architecture.md](architecture.md)** | Thiết kế kiến trúc tổng thể, ranh giới domain, giao tiếp liên service | Solution Architect, Developer |
| **[setup.md](setup.md)** | Hướng dẫn cài đặt nhanh môi trường phát triển cục bộ | Tất cả lập trình viên |
| **[environment.md](environment.md)** | Bảng tra cứu toàn bộ các biến môi trường cấu hình trong `.env` | DevOps, Developer |
| **[api.md](api.md)** | Danh mục và đặc tả các REST API endpoints của các services | Frontend & Backend Devs |
| **[database.md](database.md)** | Thiết kế cơ sở dữ liệu phân tán (Postgres 7 DBs, Redis, Qdrant, MinIO) | Database Admin, Backend Dev |
| **[development.md](development.md)** | Quy tắc phân nhánh Git, commit message chuẩn, convention viết code | Tất cả lập trình viên |
| **[testing.md](testing.md)** | Chiến lược kiểm thử tự động (Unit test, Lint, Typecheck, QA Suite) | QC/QA, Developer |
| **[deployment.md](deployment.md)** | Hướng dẫn đóng gói Docker, CI/CD với Jenkins, môi trường staging/prod | DevOps, SRE |

---

## 🚀 Lệnh Nhanh Thường Dùng (Cheatsheet)

### Windows (PowerShell)
```powershell
.\scripts\dev.ps1 help         # Xem danh sách lệnh
.\scripts\dev.ps1 docker-up    # Bật toàn bộ hạ tầng (Postgres, Redis, RabbitMQ,...)
.\scripts\dev.ps1 health       # Kiểm tra sức khỏe các container
.\scripts\dev.ps1 dev          # Chạy Frontend Nuxt 4 (http://localhost:3000)
.\scripts\dev.ps1 build        # Build toàn bộ 11 Go services và Frontend
.\scripts\dev.ps1 test         # Chạy Unit Tests
.\scripts\qa.ps1               # Chạy full bộ QA/QC tự động
```

### Linux / macOS (Makefile)
```bash
make help          # Xem hướng dẫn
make docker-up     # Khởi động hạ tầng
make health        # Kiểm tra trạng thái
make dev           # Chạy frontend
make test          # Chạy test
make build         # Build ứng dụng
```
