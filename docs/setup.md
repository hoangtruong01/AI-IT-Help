# EOMP — Setup Guide

## Prerequisites

| Tool | Version | Required |
|---|---|---|
| Node.js | >= 20 | Yes |
| pnpm | >= 10 | Yes |
| Go | >= 1.24 | Yes |
| Docker | >= 28 | Yes |
| Docker Compose | >= 2.20 | Yes |
| Git | >= 2.40 | Yes |

## Step-by-Step Setup

### 1. Clone & Configure

```bash
git clone <repository-url>
cd eomp
cp .env.example .env
```

Edit `.env` with your local values. At minimum, set passwords for:
- `POSTGRES_PASSWORD`
- `REDIS_PASSWORD`
- `RABBITMQ_PASSWORD`
- `MINIO_SECRET_KEY`
- `GRAFANA_ADMIN_PASSWORD`

### 2. Start Infrastructure

**Windows:**
```powershell
.\scripts\dev.ps1 docker-up
```

**Linux/macOS:**
```bash
make docker-up
```

### 3. Verify Infrastructure

**Windows:**
```powershell
.\scripts\dev.ps1 health
```

**Linux/macOS:**
```bash
make health
```

All services should show `OK`.

### 4. Frontend Setup

```bash
cd apps/web
pnpm install
pnpm dev
```

Open http://localhost:3000

### 5. Backend Service (Example)

```bash
cd services/gateway
go run ./cmd/server
```

## Ports Reference

| Service | Port | URL |
|---|---|---|
| Frontend | 3000 | http://localhost:3000 |
| API Gateway | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | — |
| Redis | 6379 | — |
| RabbitMQ | 5672 | — |
| RabbitMQ UI | 15672 | http://localhost:15672 |
| MinIO API | 9000 | — |
| MinIO Console | 9001 | http://localhost:9001 |
| Qdrant | 6333 | http://localhost:6333 |
| Prometheus | 9090 | http://localhost:9090 |
| Grafana | 3002 | http://localhost:3002 |
| Loki | 3100 | — |
