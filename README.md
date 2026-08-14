# EOMP — Enterprise Operations Management Platform

[![CI Pipeline](https://img.shields.io/badge/CI-Jenkins-blue.svg)](Jenkinsfile)
[![Nuxt 4](https://img.shields.io/badge/Frontend-Nuxt%204.5-00DC82.svg?logo=nuxt.js)](apps/web)
[![Go 1.24](https://img.shields.io/badge/Backend-Go%201.24-00ADD8.svg?logo=go)](services)
[![Docker](https://img.shields.io/badge/Infra-Docker%20Compose-2496ED.svg?logo=docker)](docker-compose.yml)
[![License](https://img.shields.io/badge/License-Proprietary-lightgrey.svg)](#)

> An enterprise-grade, modular, microservices-based operations management platform.

---

## Overview

EOMP provides an integrated operations foundation encompassing:

- **Employee Management** — Organizational hierarchy, profiles & departments
- **Asset Management** — IT asset tracking, hardware lifecycle & license management
- **IT Help Desk** — Ticketing system, SLA management & incident resolution
- **Workflow & Approval** — Multi-step business processes & approval state machines
- **Knowledge Base** — Documentation, manuals & Qdrant vector-powered search
- **Notification Service** — Multi-channel communications (Email, Web, Push)
- **AI Operations Assistant** — LLM-powered ticket triage & RAG retrieval
- **Reporting & BI** — Operations analytics, metrics & audit logs
- **Audit Logging** — Immutable compliance and security event trail
- **Observability** — Prometheus metrics, Grafana dashboards & Loki log aggregation

---

## Architecture & Technology Stack

```
┌─────────────────────────────────────────────────────────────┐
│                      Frontend (Nuxt 4)                       │
│        Vue 3 · TypeScript · Nuxt UI · Tailwind CSS v4        │
│          Pinia · TanStack Query · VueUse · Port 3000         │
└──────────────────────────────┬──────────────────────────────┘
                               │ REST API
┌──────────────────────────────▼──────────────────────────────┐
│                      API Gateway (Go)                        │
│            Routing · Rate Limit · Auth · Port 8080           │
└──┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬───┘
   │      │      │      │      │      │      │      │      │
   ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼
┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
│ Auth ││Employ││Asset ││Help  ││Work  ││Notif ││Knowl ││  AI  │
│:8081 ││ee:8082││:8083 ││desk  ││flow  ││:8086 ││edge  ││:8088 │
└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘└──────┘
   │       │       │       │       │       │       │       │
   ▼       ▼       ▼       ▼       ▼       ▼       ▼       ▼
┌─────────────────────────────────────────────────────────────┐
│                     Infrastructure                           │
│  PostgreSQL (7 DBs) · Redis · RabbitMQ · MinIO · Qdrant     │
│             Prometheus · Grafana · Loki                      │
└─────────────────────────────────────────────────────────────┘
```

---

## Quick Start (Developer Setup)

### 1. Prerequisites
- **Node.js**: >= 20.x (`v20.19+`)
- **pnpm**: >= 10.x
- **Go**: >= 1.24.x
- **Docker & Docker Compose**: >= 28.x / 2.39+
- **Git**: >= 2.40+

### 2. Setup Step-by-Step

```bash
# 1. Clone the repository
git clone https://github.com/hoangtruong01/AI-IT-Help.git
cd AI-IT-Help

# 2. Configure Environment Variables
cp .env.example .env
# Edit .env with your secrets (NEVER commit .env)

# 3. Start Infrastructure (Postgres, Redis, RabbitMQ, MinIO, Qdrant, Prometheus, Grafana, Loki)
# On Windows PowerShell:
.\scripts\dev.ps1 docker-up

# On Linux/macOS:
make docker-up

# 4. Verify Infrastructure Health
# On Windows PowerShell:
.\scripts\dev.ps1 health

# On Linux/macOS:
make health

# 5. Start Frontend Dev Server (http://localhost:3000)
cd apps/web
pnpm install
pnpm dev

# 6. Start a Backend Service (e.g., API Gateway on port 8080)
cd services/gateway
go run ./cmd/server
```

---

## Service Port Matrix

| Service | Protocol / Port | Local URL |
|---|---|---|
| **Frontend (Nuxt 4)** | HTTP `3000` | http://localhost:3000 |
| **API Gateway** | HTTP `8080` | http://localhost:8080 |
| **Auth Service** | HTTP `8081` | http://localhost:8081 |
| **Employee Service** | HTTP `8082` | http://localhost:8082 |
| **Asset Service** | HTTP `8083` | http://localhost:8083 |
| **Helpdesk Service** | HTTP `8084` | http://localhost:8084 |
| **Workflow Service** | HTTP `8085` | http://localhost:8085 |
| **Notification Service** | HTTP `8086` | http://localhost:8086 |
| **Knowledge Service** | HTTP `8087` | http://localhost:8087 |
| **AI Service** | HTTP `8088` | http://localhost:8088 |
| **Audit Service** | HTTP `8089` | http://localhost:8089 |
| **Reporting Service** | HTTP `8090` | http://localhost:8090 |
| **PostgreSQL** | TCP `5432` | `localhost:5432` |
| **Redis** | TCP `6379` | `localhost:6379` |
| **RabbitMQ AMQP** | TCP `5672` | `localhost:5672` |
| **RabbitMQ Management** | HTTP `15672` | http://localhost:15672 |
| **MinIO API / Console** | HTTP `9000` / `9001` | http://localhost:9001 |
| **Qdrant Vector DB** | HTTP `6333` / gRPC `6334` | http://localhost:6333 |
| **Prometheus** | HTTP `9090` | http://localhost:9090 |
| **Grafana** | HTTP `3002` | http://localhost:3002 |
| **Loki** | HTTP `3100` | `http://localhost:3100` |

---

## Development Commands

### Windows (PowerShell)
```powershell
.\scripts\dev.ps1 help         # Show all available commands
.\scripts\dev.ps1 dev          # Start Nuxt frontend dev server
.\scripts\dev.ps1 build        # Build all 11 Go services and frontend
.\scripts\dev.ps1 test         # Run unit tests across all services
.\scripts\dev.ps1 lint         # Run ESLint & go vet
.\scripts\dev.ps1 format       # Format all Go code (gofmt)
.\scripts\dev.ps1 docker-up    # Start all infrastructure
.\scripts\dev.ps1 docker-down  # Stop all infrastructure
.\scripts\dev.ps1 health       # Check health of all infrastructure
.\scripts\qa.ps1               # Run full automated QA/QC test suite
```

### Linux / CI (Makefile)
```bash
make help          # Show help
make dev           # Start frontend
make build         # Build all services
make test          # Run all unit tests
make lint          # Run linters
make docker-up     # Start containers
make docker-down   # Stop containers
make health        # Check health
```

---

## Documentation Links

- [Architecture Overview](docs/architecture.md)
- [Developer Setup Guide](docs/setup.md)
- [Development Guidelines](docs/development.md)
- [Environment Variables](docs/environment.md)
- [API Reference](docs/api.md)
- [Database Schema & Ownership](docs/database.md)
- [Testing Strategy](docs/testing.md)
- [CI/CD & Deployment](docs/deployment.md)

---

## License

Proprietary © Enterprise Operations Management Platform. All rights reserved.
