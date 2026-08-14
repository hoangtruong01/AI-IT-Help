# EOMP — Environment Variables

This document describes all environment variables used by EOMP.

See `.env.example` for a complete template.

## Application

| Variable | Description | Default |
|---|---|---|
| `APP_ENV` | Environment (`development`, `staging`, `production`) | `development` |
| `APP_PORT` | Gateway port | `8080` |

## PostgreSQL

| Variable | Description | Default |
|---|---|---|
| `POSTGRES_HOST` | Database host | `localhost` |
| `POSTGRES_PORT` | Database port | `5432` |
| `POSTGRES_USER` | Database user | `eomp` |
| `POSTGRES_PASSWORD` | Database password | — |
| `POSTGRES_DB` | Database name | `eomp` |

## Redis

| Variable | Description | Default |
|---|---|---|
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | — |

## RabbitMQ

| Variable | Description | Default |
|---|---|---|
| `RABBITMQ_HOST` | RabbitMQ host | `localhost` |
| `RABBITMQ_PORT` | AMQP port | `5672` |
| `RABBITMQ_USER` | RabbitMQ user | `eomp` |
| `RABBITMQ_PASSWORD` | RabbitMQ password | — |
| `RABBITMQ_MANAGEMENT_PORT` | Management UI port | `15672` |

## MinIO

| Variable | Description | Default |
|---|---|---|
| `MINIO_ENDPOINT` | MinIO endpoint | `localhost:9000` |
| `MINIO_ACCESS_KEY` | Access key | — |
| `MINIO_SECRET_KEY` | Secret key | — |
| `MINIO_CONSOLE_PORT` | Console UI port | `9001` |

## Qdrant

| Variable | Description | Default |
|---|---|---|
| `QDRANT_HOST` | Qdrant host | `localhost` |
| `QDRANT_PORT` | REST API port | `6333` |
| `QDRANT_GRPC_PORT` | gRPC port | `6334` |

## AI

| Variable | Description | Default |
|---|---|---|
| `AI_API_KEY` | LLM provider API key | — |
| `AI_MODEL` | LLM model name | — |
| `EMBEDDING_MODEL` | Embedding model name | — |

## Monitoring

| Variable | Description | Default |
|---|---|---|
| `GRAFANA_PORT` | Grafana UI port | `3001` |
| `GRAFANA_ADMIN_USER` | Grafana admin username | `admin` |
| `GRAFANA_ADMIN_PASSWORD` | Grafana admin password | — |
| `PROMETHEUS_PORT` | Prometheus port | `9090` |
| `LOKI_PORT` | Loki port | `3100` |

## Frontend

| Variable | Description | Default |
|---|---|---|
| `NUXT_PORT` | Nuxt dev server port | `3000` |
| `NUXT_PUBLIC_API_URL` | Public API URL for frontend | `http://localhost:8080` |
