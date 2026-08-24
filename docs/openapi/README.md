# EOMP OpenAPI 3.0 (Swagger) Specification Hub

Tài liệu này hướng dẫn cách tra cứu và xem giao diện tương tác API Swagger UI của nền tảng **EOMP**.

---

## 1. 📄 Vị Trí File Đặc Tả API
- File OpenAPI 3.0 chính: [`docs/openapi/eomp-openapi-spec.yaml`](file:///d:/IT_help/eomp/docs/openapi/eomp-openapi-spec.yaml)

---

## 2. 🚀 Cách Xem Trực Quan Bằng Swagger UI

### Cách 1: Sử dụng VS Code / IDE Extension
- Cài đặt extension **OpenAPI (Swagger) Editor** hoặc **Swagger Viewer**.
- Mở file `eomp-openapi-spec.yaml` và bấm tổ hợp phím `Shift + Alt + E` (hoặc `Preview Swagger`).

### Cách 2: Sử dụng Docker Swagger UI
```bash
docker run -p 8088:8080 -e SWAGGER_JSON=/spec/eomp-openapi-spec.yaml -v "${PWD}/docs/openapi:/spec" swaggerapi/swagger-ui
```
Sau đó truy cập: `http://localhost:8088`

---

## 3. 🛡️ Các Nhóm API Đã Được Chuẩn Hóa
1. **Authentication & Identity (`/api/v1/auth/*`)**
2. **Employee & Department Catalog (`/api/v1/employees/*`, `/api/v1/departments/*`)**
3. **Asset & CMDB Dependency Topology (`/api/v1/assets/*`, `/api/v1/cmdb/*`)**
4. **Helpdesk & Incident Lifecycle (`/api/v1/tickets/*`, `/api/v1/problems/*`, `/api/v1/changes/*`)**
5. **Workflow & Multi-Level Approvals (`/api/v1/workflows/*`, `/api/v1/approvals/*`)**
6. **Notification Center (`/api/v1/notifications/*`)**
7. **Knowledge Base & Runbooks (`/api/v1/knowledge/*`)**
8. **AI Operations Copilot (`/api/v1/ai/*`)**
9. **Observability & SRE Health Mesh (`/api/v1/monitoring/*`, `/metrics`)**
10. **Executive BI & SLA Analytics (`/api/v1/reports/*`)**
11. **Security & Immutable Audit Trail (`/api/v1/audit/*`)**
