# EOMP — Master Testing Strategy & QA Automation Suite

> Coverage above 85%, Playwright coverage and deployed load-test results are targets, not current verified results. See `docs/IMPLEMENTATION_STATUS.md`.

> **Tài Liệu Chiến Lược Kiểm Thử Tự Động & Tiêu Chuẩn Chất Lượng (QA/QC Suite)**  
> **Áp dụng cho:** QA Engineers, Automation Testers, Developers & SRE.  
> **Mục tiêu:** Đạt độ bao phủ Code Coverage > 85% và 100% Pass trên pipeline CI/CD.

---

## 📑 MỤC LỤC
1. [Mô Hình Kim Tự Tháp Kiểm Thử (Testing Pyramid)](#1-mô-hình-kim-tự-tháp-kiểm-thử-testing-pyramid)
2. [Kiểm Thử Đơn Vị & Tĩnh (Unit Tests & Static Analysis)](#2-kiểm-thử-đơn-vị--tĩnh-unit-tests--static-analysis)
3. [Kiểm Thử Luồng Nghiệp Vụ Liên Service (Cross-Service E2E Lifecycle)](#3-kiểm-thử-luồng-nghiệp-vụ-liên-service-cross-service-e2e-lifecycle)
4. [Kiểm Thử Tải & Hiệu Năng (K6 Load & Stress Testing 500 VUs)](#4-kiểm-thử-tải--hiệu-năng-k6-load--stress-testing-500-vus)
5. [Tự Động Hóa Pipeline QA CI/CD (`scripts/qa.ps1`)](#5-tự-động-hóa-pipeline-qa-cicd-scriptsqaps1)

---

## 1. Mô Hình Kim Tự Tháp Kiểm Thử (Testing Pyramid)

```
                       ▲
                      / \
                     /   \     End-to-End Tests (Playwright / Cross-Service E2E)
                    / E2E \    Luồng 7 bước hoàn chỉnh từ Login -> Handover
                   /───────\
                  /         \   K6 Load & Stress Testing
                 / Perf & K6 \  500 Concurrent VUs, p95 < 200ms
                /─────────────\
               /               \ Integration & API Tests
              / Integration/API \ HTTP Handlers, DB Repositories, Probes
             /───────────────────\
            /                     \ Unit Tests & Static Verification
           /   Go Unit + Vitest    \ go test, go vet, vue-tsc, ESLint
          └─────────────────────────┘
```

---

## 2. Kiểm Thử Đơn Vị & Tĩnh (Unit Tests & Static Analysis)

### 2.1 Backend Go Services
```bash
# Chạy Unit Tests trên toàn bộ các services
go test -v -race -cover ./packages/shared/... ./services/...

# Chạy kiểm tra lỗi cú pháp và cảnh báo tĩnh
go vet ./packages/shared/... ./services/...

# Định dạng code chuẩn
gofmt -w packages/ services/
```

### 2.2 Frontend Nuxt 4 Web App
```bash
cd apps/web

# Kiểm tra TypeScript Strict Typing (0 errors required)
pnpm typecheck

# Kiểm tra cú pháp và quy chuẩn Vue/Tailwind
pnpm lint

# Chạy Unit Tests tính toán KPI & Data Masking
pnpm test
```

---

## 3. Mô phỏng luồng nghiệp vụ liên service (in-memory simulation)

> Tên thư mục `tests/e2e` là tên legacy. Các test hiện tại dùng struct, mock và `httptest`; chúng không deploy service thật và không phải deployed E2E.

Tệp kiểm thử `tests/e2e/e2e_lifecycle_test.go` mô phỏng kịch bản 7 bước:

1. **`Auth Service`:** Cấp phát JWT Bearer Token với vai trò tương ứng (Employee, Manager, IT Agent).
2. **`Helpdesk Service`:** Khởi tạo Incident Ticket từ Service Catalog (`TK-2026-8801`).
3. **`SLA Engine`:** Tự động tính toán hạn phản hồi và hạn xử lý cam kết trong 4 giờ.
4. **`Workflow Engine`:** Kích hoạt quy trình phê duyệt đa cấp -> Manager thực hiện `APPROVED`.
5. **`Notification Service`:** Lắng nghe CloudEvent realtime `eomp.workflow.approved` gửi tới IT Agent.
6. **`Asset & CMDB Service`:** Gán thiết bị từ kho CMDB (`AST-MBP-9901`) sang trạng thái `IN_USE`.
7. **`Audit Service`:** Xác thực chuỗi **HMAC-SHA256**, liên kết predecessor, endpoint integrity và che giấu trường nhạy cảm (`********`).

```bash
# Chạy bộ unit/in-memory simulation
go test -v ./tests/e2e/...
```

---

## 4. Kiểm Thử Tải & Hiệu Năng (K6 Load & Stress Testing 500 VUs)

Bộ công cụ K6 tại [`infrastructure/k6/`](file:///d:/IT_help/eomp/infrastructure/k6/):

* **Load Test (`load_test.js`):** Mô phỏng **500 Concurrent VUs** trong 3 phút trên các endpoints trọng yếu (Gateway, Auth, Tickets, Assets, Reports, Audit).
  - Tiêu chuẩn nghiệm thu: p95 Response Time `< 200ms`, Error Rate `< 1%`.
* **Stress Test (`stress_test.js`):** Đẩy tải đột biến lên **800 VUs** để kiểm tra tính năng tự động co giãn của HPA và cơ chế Rate Limiting chống sập hệ thống.

```bash
# Chạy K6 Load Test
k6 run infrastructure/k6/load_test.js
```

---

## 5. Tự Động Hóa Pipeline QA CI/CD (`scripts/qa.ps1`)

EOMP cung cấp CLI tự động hóa chạy toàn bộ 6 tầng kiểm định:

```powershell
# Chạy toàn bộ test suite QA 1-click
.\scripts\qa.ps1
```

```
=== Báo Cáo 6 Tầng Kiểm Định QA/QC ===
[1/6] Frontend Quality Gate (Typecheck, Lint, Build) ...... [PASS]
[2/6] Backend Code Quality Gate (12 Go Modules go vet) .... [PASS]
[3/6] Backend Unit & Coverage Test Suite .................. [PASS]
[4/6] Cross-Service E2E Lifecycle Integration Test ........ [PASS]
[5/6] Security & RBAC 403/429 Chokepoint Gate ............. [PASS]
[6/6] SRE Probes & Kubernetes Manifests Validation ........ [PASS]
------------------------------------------------------------------
RESULT: 100% PASS (Unit & In-Memory Simulation Suite)
LƯU Ý: Xem docs/IMPLEMENTATION_STATUS.md và Gate A-D cho Production Readiness Evidence
```
