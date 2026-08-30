# EOMP — Enterprise Operations Management Platform

[![CI Pipeline](https://img.shields.io/badge/CI-Jenkins%20Multi--Stage-blue.svg)](Jenkinsfile)
[![Nuxt 4](https://img.shields.io/badge/Frontend-Nuxt%204.5%20%7C%20Vue%203%20%7C%20Tailwind-00DC82.svg?logo=nuxt.js)](apps/web)
[![Go 1.25](https://img.shields.io/badge/Backend-11%20Go%20Microservices-00ADD8.svg?logo=go)](services)
[![Docker](https://img.shields.io/badge/Docker-Multi--Stage%20%3C25MB-2496ED.svg?logo=docker)](deploy/docker)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-Helm%20Chart%20%26%20HPA-326CE5.svg?logo=kubernetes)](deploy/kubernetes)
[![SRE & DR](https://img.shields.io/badge/SRE%20%26%20DR-Awaiting%20Drill-orange.svg)](docs/IMPLEMENTATION_STATUS.md)
[![Status](https://img.shields.io/badge/Status-Remediation%20in%20Progress-orange.svg)](docs/IMPLEMENTATION_STATUS.md)

> **Enterprise-grade, modular, event-driven microservices operations management platform designed for high-scale organizations.**

---

## 📑 Table of Contents
1. [Platform Overview & Core Modules](#-platform-overview--core-modules)
2. [Master Architecture (C4 Model & RED Method)](#-master-architecture-c4-model--red-method)
3. [Service & Network Port Matrix](#-service--network-port-matrix)
4. [Quick Start (Local & Production Setup)](#-quick-start-local--production-setup)
5. [Development & Operations CLI](#-development--operations-cli)
6. [Master Roadmap](#-master-roadmap)
7. [Engineering Documentation Hub](#-engineering-documentation-hub)

---

## 🌟 Platform Overview & Core Modules

EOMP provides an end-to-end, integrated operations ecosystem structured across 11 Golang microservices and a modern Nuxt 4 SSR Web Application:

1. **Authentication & RBAC (`services/auth` - :8081)**: Multi-role access control (`admin`, `manager`, `agent`, `employee`), JWT HS256 authentication, Bcrypt password hashing, and token refresh.
2. **Employee Directory (`services/employee` - :8082)**: Organizational structure, department hierarchy, search & employee profile management.
3. **Asset Management & CMDB (`services/asset` - :8083)**: IT hardware lifecycle, stock allocation/return, and CMDB Infrastructure Dependency Topology graph.
4. **IT Helpdesk & SLA Engine (`services/helpdesk` - :8084)**: Incident ticketing, dynamic SLA deadline calculations (`URGENT` 15m/2h, `HIGH` 30m/4h), and ITIL Problem Management with RCA 5-Whys.
5. **Workflow & Approval Engine (`services/workflow` - :8085)**: State machine workflow execution, multi-level approvals, and Change Advisory Board (CAB) voting.
6. **Notification Service (`services/notification` - :8086)**: CloudEvents v1.0 event bus consumer, in-app notification center, and email delivery.
7. **Knowledge Base & SOP Runbooks (`services/knowledge` - :8087)**: IT standard operating procedures, documentation management, and vector embeddings.
8. **AI Operations Copilot (`services/ai` - :8088)**: LLM ticket auto-triage, semantic search with Qdrant vector store, and RAG root-cause assistant.
9. **Audit Trail & Compliance (`services/audit` - :8089)**: Append-only audit logging with a chained **HMAC-SHA256 proof**, integrity verification and automated data masking (`********`).
10. **Reporting & BI Analytics (`services/reporting` - :8090)**: MTTR/MTTD, SLA compliance, agent performance scorecards and bounded report export. Performance targets require deployed-environment evidence.
11. **API Gateway (`services/gateway` - :8080)**: Reverse proxy routing, rate limiting (100 req/min/IP), correlation ID injection, and SRE monitoring aggregation.
12. **Frontend Web App (`apps/web` - :3000)**: Nuxt 4, Vue 3, Tailwind CSS, Dark Glassmorphism aesthetic, realtime live widgets & SSE streaming.

---

## 🏛️ Master Architecture (C4 Model & RED Method)

```
                            ┌──────────────────────────────────────────┐
                            │        Nuxt 4 Web Frontend (:3000)       │
                            │   Vue 3 · Tailwind CSS · Pinia · SSR    │
                            └────────────────────┬─────────────────────┘
                                                 │ REST / SSE
                            ┌────────────────────▼─────────────────────┐
                            │         API Gateway Go (:8080)           │
                            │  Reverse Proxy · JWT · Rate Limit · RED  │
                            └───────┬────────────┬─────────────┬───────┘
                                    │            │             │
        ┌───────────────────────────┼────────────┴─────────────┼───────────────────────────┐
        ▼                           ▼                          ▼                           ▼
 ┌──────────────┐            ┌──────────────┐           ┌──────────────┐            ┌──────────────┐
 │     Auth     │            │   Employee   │           │    Asset     │            │   Helpdesk   │
 │    :8081     │            │    :8082     │           │    :8083     │            │    :8084     │
 └──────┬───────┘            └──────┬───────┘           └──────┬───────┘            └──────┬───────┘
        │                           │                          │                           │
        ▼                           ▼                          ▼                           ▼
 ┌──────────────┐            ┌──────────────┐           ┌──────────────┐            ┌──────────────┐
 │   Workflow   │            │ Notification │           │  Knowledge   │            │  AI Copilot  │
 │    :8085     │            │    :8086     │           │    :8087     │            │    :8088     │
 └──────┬───────┘            └──────┬───────┘           └──────┬───────┘            └──────┬───────┘
        │                           │                          │                           │
        ▼                           ▼                          ▼                           ▼
 ┌──────────────┐            ┌──────────────┐           ┌──────────────────────────────────────────┐
 │ Audit Trail  │            │ Reporting/BI │           │          CloudEvents EventBus            │
 │    :8089     │            │    :8090     │           │       RabbitMQ 4 · Asynchronous          │
 └──────┬───────┘            └──────┬───────┘           └──────────────────────────────────────────┘
        │                           │
        └───────────────────────────┴──────────────────────────┐
                                                               ▼
 ┌─────────────────────────────────────────────────────────────────────────────────────────────┐
 │                                   Distributed Infrastructure                                │
 │   PostgreSQL 17 (9 Isolated DBs) · Redis 7 · MinIO S3 · Qdrant Vector Store                 │
 │   Prometheus Metrics · Grafana Dashboards (:3002) · Loki Log Aggregator                     │
 └─────────────────────────────────────────────────────────────────────────────────────────────┘
```

---

## 🔌 Service & Network Port Matrix

| Service / Infrastructure | Port | Protocol | Data Store / Role |
|---|:---:|:---:|---|
| **Frontend Web App** | `:3000` | HTTP | Nuxt 4 SSR UI Portal |
| **API Gateway** | `:8080` | HTTP | Stateless Reverse Proxy & SRE Aggregator |
| **Auth Service** | `:8081` | HTTP | `auth_db` (PostgreSQL) |
| **Employee Service** | `:8082` | HTTP | `employee_db` (PostgreSQL) |
| **Asset & CMDB** | `:8083` | HTTP | `asset_db` (PostgreSQL) |
| **Helpdesk & SLA** | `:8084` | HTTP | `helpdesk_db` (PostgreSQL) |
| **Workflow Engine** | `:8085` | HTTP | `workflow_db` (PostgreSQL) |
| **Notification** | `:8086` | HTTP | `notification_db` + RabbitMQ |
| **Knowledge Base** | `:8087` | HTTP | `knowledge_db` + Qdrant |
| **AI Copilot** | `:8088` | HTTP | RAG SmartRetriever (Stateless) |
| **Audit Trail** | `:8089` | HTTP | `audit_db` (chained HMAC-SHA256 proof) |
| **Reporting & BI** | `:8090` | HTTP | `reporting_db` (PostgreSQL) |
| **PostgreSQL Engine** | `:5432` | TCP | 9 Isolated DBs Pool |
| **Redis Cache** | `:6379` | TCP | Session & Token Bucket Store |
| **RabbitMQ AMQP / UI** | `:5672` / `:15672` | TCP/HTTP | CloudEvents Message Broker |
| **MinIO Storage / UI** | `:9000` / `:9001` | HTTP | S3 Object Storage |
| **Qdrant Vector DB** | `:6333` / `:6334` | HTTP/gRPC | Vector Embeddings Store |
| **Prometheus** | `:9090` | HTTP | SRE RED Method Time-Series |
| **Grafana** | `:3002` | HTTP | Executive & Operational Dashboards |
| **Loki** | `:3100` | HTTP | Structured Log Aggregation |

---

## 🚀 Quick Start (Local & Production Setup)

### 1. Development Mode

```bash
# 1. Clone the repository
git clone https://github.com/hoangtruong01/AI-IT-Help.git
cd AI-IT-Help

# 2. Setup Environment Variables
cp .env.example .env

# 3. Start Infrastructure via PowerShell CLI (Windows)
.\scripts\dev.ps1 docker-up
.\scripts\dev.ps1 health

# Or via Makefile (Linux/macOS)
make docker-up
make health

# 4. Start Frontend Web Server (http://localhost:3000)
cd apps/web && pnpm install && pnpm dev
```

### 2. Production Docker & Kubernetes Deployment

```bash
# Production Docker Compose (All 11 Services + Web + Nginx + DBs)
.\scripts\deploy.ps1 prod-up

# Deploy to Kubernetes via Native Manifests
.\scripts\deploy.ps1 k8s-apply

# Deploy to Kubernetes via Production Helm Chart
.\scripts\deploy.ps1 helm-install
```

---

## 🛠️ Development & Operations CLI

EOMP provides automated CLI toolkits for developer, testing, deployment and SRE workflows:

| Script | Purpose | Key Commands |
|---|---|---|
| [`scripts/dev.ps1`](scripts/dev.ps1) / `Makefile` | Local developer workflow | `docker-up`, `dev`, `build`, `test`, `lint`, `health` |
| [`scripts/qa.ps1`](scripts/qa.ps1) | Automated 6-tier QA/QC verification | `.\scripts\qa.ps1` (E2E, unit tests, linters, probes) |
| [`scripts/deploy.ps1`](scripts/deploy.ps1) / [`deploy.sh`](scripts/deploy.sh) | Production deployment orchestration | `validate`, `prod-up`, `k8s-apply`, `helm-install` |
| [`scripts/chaos.ps1`](scripts/chaos.ps1) / [`chaos.sh`](scripts/chaos.sh) | SRE Chaos Engineering simulations | `simulate-db-down`, `simulate-rabbit-jam`, `run-all-chaos` |
| [`scripts/backup_restore.ps1`](scripts/backup_restore.ps1) | Multi-DB Backup & Disaster Recovery | `backup`, `list`, `test-restore` |

---

## 🚀 Master Roadmap

| Phase | Module Scope | Status |
|:---:|---|:---:|
| **Phase 0–9** | Core architecture and business modules | **Implemented; integration acceptance pending** |
| **Phase 10** | Security hardening | **In progress; see current blockers** |
| **Phase 11** | QA automation | **Partial: unit/static checks exist; browser/load evidence pending** |
| **Phase 12** | Architecture and OpenAPI documentation | **101/101 operation parity; domain schemas in progress** |
| **Phase 13** | Docker and Kubernetes packaging | **Implemented; runtime validation pending** |
| **Phase 14** | SRE, disaster recovery and handover | **Pending measured DR drill and owner acceptance** |

The evidence-backed status and release blockers are maintained in [docs/IMPLEMENTATION_STATUS.md](docs/IMPLEMENTATION_STATUS.md).

---

## 📚 Engineering Documentation Hub

Detailed specifications and architectural artifacts:

- **Core Guides:**
  - 📖 **[Developer & Intern Comprehensive Guide](docs/INTERN_DEVELOPER_GUIDE.md)** *(A-Z onboarding)*
  - 📋 **[Project Structure & Daily Changelog](docs/PROJECT_STRUCTURE_AND_CHANGELOG.md)** *(Single Source of Truth)*
  - 🎯 **[Phase 6 to 14 Multi-Role Specification](docs/PHASE_6_TO_14_ROADMAP_SPECIFICATION.md)**
- **Architecture & API:**
  - 🏛️ **[C4 Architecture Diagrams (Levels 1-4)](docs/architecture/c4_model_diagrams.md)**
  - 🗄️ **[Master ERD & Data Dictionary](docs/architecture/database_erd_and_data_dictionary.md)**
  - 🔌 **[OpenAPI 3.0 Specification Hub](docs/openapi/eomp-openapi-spec.yaml)**
- **Operations & SRE:**
  - 🚢 **[Production Deployment & Kubernetes Guide](docs/deployment.md)**
  - 🛡️ **[Disaster Recovery Plan (RPO<5m, RTO<15m)](docs/sre/disaster_recovery_plan.md)**
  - 🚨 **[Incident Response Playbook (SEV 1-4)](docs/sre/incident_response_playbook.md)**
  - 📖 **[SRE Operations Manual (Day-2 Ops)](docs/sre/operations_manual.md)**
  - 🔥 **[Chaos Engineering Runbook](docs/sre/chaos_engineering_runbook.md)**
  - 📜 **[Master Platform Handover Certificate](docs/sre/project_handover_acceptance.md)**

---

## 📄 License

Proprietary © 2026 Enterprise Operations Management Platform. All rights reserved.
