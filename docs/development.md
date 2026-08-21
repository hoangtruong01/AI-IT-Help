# EOMP — Developer Guidelines & Engineering Conventions

> **Quy Chuẩn Lập Trình & Quy Trình Phát Triển Phần Mềm (Engineering Standards)**  
> **Áp dụng cho:** Tất cả các kỹ sư phát triển phần mềm trong dự án.

---

## 1. Chiến Lược Phân Nhánh Git (Git Branching Strategy)

```
main            ──► Nhánh Production (Chỉ nhận merge từ develop khi có Release Tag)
develop         ──► Nhánh tích hợp chính (Active Development)
feature/<name>  ──► Nhánh phát triển tính năng mới (Tách từ develop)
bugfix/<name>   ──► Nhánh sửa lỗi kiểm thử (Tách từ develop)
hotfix/<name>   ──► Nhánh vá lỗi khẩn cấp Production (Tách từ main)
```

---

## 2. Quy Chuẩn Commit Messages (Conventional Commits)

Format: `<type>(<scope>): <subject>`

* `feat`: Thêm tính năng mới (VD: `feat(phase-13): kubernetes manifests and helm charts`)
* `fix`: Sửa lỗi (VD: `fix(auth): handle expired refresh token`)
* `refactor`: Tái cấu trúc mã nguồn không thay đổi logic
* `test`: Thêm hoặc sửa test cases
* `docs`: Cập nhật tài liệu
* `chore`: Cập nhật cấu hình build, dependencies

---

## 3. Quy Chuẩn Clean Architecture Cho Go Microservices

Mỗi microservice tuân thủ đúng 4 tầng tách biệt:
1. **`internal/model/`:** Chứa Domain Entities, Enums, DTOs và Filter structs.
2. **`internal/repository/`:** Chứa interface `Repository` và SQL queries trực tiếp tới PostgreSQL.
3. **`internal/service/`:** Chứa Business Logic, State Machine, Domain Event triggers và validation.
4. **`internal/handler/`:** Tiếp nhận HTTP requests, kiểm tra quyền, gọi Service và trả về JSON Envelope qua `packages/shared/pkg/response`.

---

## 4. Quy Chuẩn Frontend Nuxt 4 & Vue 3

* **TypeScript Strict:** 100% các biến và hàm đều có Type rõ ràng trong `apps/web/app/types/index.ts`.
* **Dark Glassmorphism UI:** Sử dụng bảng màu slate tối chuẩn (`bg-slate-950`, `bg-slate-900/80`, `border-slate-800/80`).
* **Micro-animations:** Hover card scale, pulsing status dots, bouncing unread badges.
* **Composables:** Dùng `useApi()` để tự động gắn Bearer Token và xử lý lỗi mạng tập trung.
