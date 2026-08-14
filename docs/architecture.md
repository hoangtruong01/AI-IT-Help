# EOMP — Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────┐
│                    Frontend (Nuxt 4)                     │
│                   apps/web — Port 3000                   │
└─────────────────────┬───────────────────────────────────┘
                      │ REST API
┌─────────────────────▼───────────────────────────────────┐
│                   API Gateway (Go)                       │
│               services/gateway — Port 8080               │
│   Routing · Auth · Rate Limit · Logging · Correlation   │
└──┬──────┬──────┬──────┬──────┬──────┬──────┬──────┬─────┘
   │      │      │      │      │      │      │      │
   │ gRPC │ gRPC │ gRPC │ gRPC │ gRPC │ gRPC │ gRPC │
   ▼      ▼      ▼      ▼      ▼      ▼      ▼      ▼
┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐┌──────┐
│ Auth ││Employ││Asset ││Help  ││Work  ││Notif ││Knowl │
│      ││ee    ││      ││desk  ││flow  ││      ││edge  │
└──┬───┘└──┬───┘└──┬───┘└──┬───┘└──┬───┘└──┬───┘└──┬───┘
   │       │       │       │       │       │       │
   ▼       ▼       ▼       ▼       ▼       ▼       ▼
┌─────────────────────────────────────────────────────────┐
│                    Infrastructure                        │
│  PostgreSQL · Redis · RabbitMQ · MinIO · Qdrant          │
└─────────────────────────────────────────────────────────┘
```

## Service Responsibilities

| Service | Bounded Context | Port |
|---|---|---|
| gateway | API routing, auth verification, rate limiting | 8080 |
| auth | Authentication, authorization, JWT, RBAC | 8081 |
| employee | Employee profiles, org structure, HR | 8082 |
| asset | IT asset lifecycle, inventory | 8083 |
| helpdesk | Tickets, SLA, assignments | 8084 |
| workflow | Business processes, approvals | 8085 |
| notification | Email, SMS, push, in-app | 8086 |
| knowledge | Documentation, search, KB | 8087 |
| ai | LLM, embeddings, RAG, assistant | 8088 |
| audit | Activity logs, compliance | 8089 |
| reporting | Reports, analytics, dashboards | 8090 |

## Communication Patterns

### External (Frontend → Backend)
- **REST API** via API Gateway
- Frontend never calls microservices directly

### Internal (Service → Service)
- **gRPC** for synchronous calls
- **RabbitMQ** for asynchronous events

### Data Ownership
- Each service owns its own data
- No shared database access
- Cross-service data via API calls or events

## Infrastructure

| Component | Purpose | Port |
|---|---|---|
| PostgreSQL | Primary database | 5432 |
| Redis | Cache, sessions, rate limiting | 6379 |
| RabbitMQ | Message broker, events | 5672 / 15672 |
| MinIO | Object storage (files, images) | 9000 / 9001 |
| Qdrant | Vector database (AI/search) | 6333 / 6334 |
| Prometheus | Metrics collection | 9090 |
| Grafana | Dashboards, visualization | 3001 |
| Loki | Log aggregation | 3100 |
