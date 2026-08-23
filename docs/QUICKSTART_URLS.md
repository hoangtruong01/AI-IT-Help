# 🚀 EOMP — QUICKSTART DASHBOARD & SERVICE URLS

> **Environment:** Development & Local Sandbox  
> **Platform Version:** 2.0.0 Enterprise Master Edition  

---

## 🌐 Web & Application Dashboards

| Service / Interface | Local URL | Default Credentials | Description |
|---|---|---|---|
| **Web Portal (Nuxt 4)** | [http://localhost:3000](http://localhost:3000) | `admin@eomp.local` / `Admin@123456` | Single Page Application (13 operational modules) |
| **API Gateway Ingress** | [http://localhost:8080](http://localhost:8080) | — | Central Reverse Proxy & Auth Filter |
| **RabbitMQ Management** | [http://localhost:15672](http://localhost:15672) | `eomp` / `eomp_dev_password` | AMQP 4 Message Broker & Exchanges |
| **MinIO Object Console** | [http://localhost:9001](http://localhost:9001) | `eomp_minio` / `eomp_minio_secret` | S3-Compatible Object Storage Buckets |
| **Grafana Dashboards** | [http://localhost:3002](http://localhost:3002) | `admin` / `eomp_grafana` | RED Metrics, SLA Trends & System Monitoring |
| **Prometheus Metrics** | [http://localhost:9090](http://localhost:9090) | — | Metrics Scraping Engine & Alerts |
| **Qdrant Vector DB** | [http://localhost:6333/dashboard](http://localhost:6333/dashboard) | — | Vector Database Web UI & Collections |
| **Loki Log Viewer** | [http://localhost:3100](http://localhost:3100) | — | High-performance log aggregation |

---

## ⚡ Microservices Direct Ports (Internal Dev / Probes)

| Service Name | Port | Direct Health Endpoint | Dedicated Database |
|---|---|---|---|
| **API Gateway** | `:8080` | `http://localhost:8080/health` | — |
| **Auth & Identity** | `:8081` | `http://localhost:8081/health` | `auth_db` |
| **Employee & Organization** | `:8082` | `http://localhost:8082/health` | `employee_db` |
| **Asset & CMDB** | `:8083` | `http://localhost:8083/health` | `asset_db` |
| **Helpdesk & Incident** | `:8084` | `http://localhost:8084/health` | `helpdesk_db` |
| **Workflow & Changes** | `:8085` | `http://localhost:8085/health` | `workflow_db` |
| **Notification Center** | `:8086` | `http://localhost:8086/health` | `notification_db` |
| **Knowledge Base** | `:8087` | `http://localhost:8087/health` | `knowledge_db` |
| **AI Copilot** | `:8088` | `http://localhost:8088/health` | Qdrant (`:6333`) |
| **Audit & Security** | `:8089` | `http://localhost:8089/health` | `audit_db` |
| **Reporting & BI** | `:8090` | `http://localhost:8090/health` | `reporting_db` |

---

## 👥 Demo Pre-Seeded Accounts

| Email | Password | Role | Description |
|---|---|---|---|
| `admin@eomp.local` | `Admin@123456` | `ROLE_ADMIN` | System Administrator / IT Director |
| `manager@eomp.local` | `Admin@123456` | `ROLE_MANAGER` | IT Operations & Approvals Manager |
| `agent@eomp.local` | `Admin@123456` | `ROLE_AGENT` | L1 / L2 IT Support Specialist |
| `emily.davis@eomp.local` | `Admin@123456` | `ROLE_EMPLOYEE` | Standard Employee / Engineer |
