# EOMP — Enterprise Operations Management Platform

> A modular, microservices-based platform for enterprise operations management.

## Overview

EOMP is a comprehensive enterprise platform covering:

- **Employee Management** — HR, profiles, organizational structure
- **Asset Management** — IT assets, lifecycle tracking
- **IT Help Desk** — Ticket management, SLA tracking
- **Workflow & Approval** — Business process automation
- **Knowledge Base** — Internal documentation, AI-powered search
- **Notification** — Multi-channel alerts and communications
- **AI Assistant** — LLM-powered help desk and analytics
- **Reporting** — Business intelligence and dashboards
- **Audit Log** — Compliance and activity tracking
- **Monitoring** — System observability

## Tech Stack

| Layer | Technology |
|---|---|
| Frontend | Nuxt 4, Vue 3, TypeScript, Tailwind CSS, Nuxt UI, Pinia |
| Backend | Go, REST API, gRPC, Microservices |
| Message Broker | RabbitMQ |
| Database | PostgreSQL, Redis |
| AI | LLM, Embeddings, RAG, Qdrant |
| Storage | MinIO |
| Monitoring | Prometheus, Grafana, Loki, OpenTelemetry |
| DevOps | Docker, Docker Compose, Jenkins, Nginx |

## Repository Structure

```
eomp/
├── apps/web/              # Nuxt 4 frontend
├── services/              # Go microservices
│   ├── gateway/           # API Gateway
│   ├── auth/              # Authentication & Authorization
│   ├── employee/          # Employee Management
│   ├── asset/             # Asset Management
│   ├── helpdesk/          # IT Help Desk
│   ├── workflow/          # Workflow & Approval
│   ├── notification/      # Notification Service
│   ├── knowledge/         # Knowledge Base
│   ├── ai/                # AI Assistant
│   ├── audit/             # Audit Log
│   └── reporting/         # Reporting
├── packages/              # Shared packages
│   ├── proto/             # gRPC protobuf definitions
│   └── shared/            # Shared Go libraries
├── infrastructure/        # Infrastructure configs
├── deployment/            # Deployment configurations
├── docs/                  # Documentation
└── scripts/               # Development scripts
```

## Quick Start

### Prerequisites

- Node.js >= 20
- pnpm >= 10
- Go >= 1.24
- Docker & Docker Compose
- Git

### Setup

```bash
# 1. Clone the repository
git clone <repository-url>
cd eomp

# 2. Copy environment configuration
cp .env.example .env
# Edit .env with your values

# 3. Start infrastructure services
docker compose up -d

# 4. Start frontend (development)
cd apps/web
pnpm install
pnpm dev

# 5. Start a backend service (example: gateway)
cd services/gateway
go run ./cmd/server
```

### Windows (PowerShell)

```powershell
# Start all infrastructure
.\scripts\dev.ps1 docker-up

# Stop all infrastructure
.\scripts\dev.ps1 docker-down

# Check health
.\scripts\dev.ps1 health

# View logs
.\scripts\dev.ps1 logs
```

### Linux / CI (Makefile)

```bash
make docker-up    # Start infrastructure
make docker-down  # Stop infrastructure
make health       # Check service health
make dev          # Start development
make build        # Build all services
make test         # Run all tests
make lint         # Run linters
```

## Documentation

- [Architecture](docs/architecture.md)
- [Setup Guide](docs/setup.md)
- [Development Guide](docs/development.md)
- [Environment Variables](docs/environment.md)
- [API Reference](docs/api.md)
- [Database](docs/database.md)
- [Testing](docs/testing.md)
- [Deployment](docs/deployment.md)

## License

Proprietary — All rights reserved.
