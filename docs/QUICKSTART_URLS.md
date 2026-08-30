# EOMP quickstart URLs

Status baseline: version `0.1.0`, remediation in progress. See [IMPLEMENTATION_STATUS.md](IMPLEMENTATION_STATUS.md) before evaluating release readiness.

## Start local infrastructure

1. Copy `.env.example` to `.env` and replace every development credential before using a shared environment.
2. Start infrastructure with `docker compose up -d`.
3. Start the Go services and Nuxt application using the commands in [setup.md](setup.md).

The repository does not seed fixed application users. To provision the first administrator, set `BOOTSTRAP_ADMIN_EMAIL` and `BOOTSTRAP_ADMIN_PASSWORD` for one controlled auth-service startup, verify the account, and then remove the bootstrap password.

## Application endpoints

| Component | Local URL | Notes |
|---|---|---|
| Web application | <http://localhost:3000> | Nuxt application |
| API gateway | <http://localhost:8080> | Use `/health` for liveness |
| Auth | <http://localhost:8081> | `/health`; `/ready` checks PostgreSQL |
| Employee | <http://localhost:8082> | `/health`; `/ready` checks PostgreSQL |
| Asset | <http://localhost:8083> | `/health`; `/ready` checks PostgreSQL |
| Helpdesk | <http://localhost:8084> | `/health`; `/ready` checks PostgreSQL |
| Workflow | <http://localhost:8085> | `/health`; `/ready` checks PostgreSQL |
| Notification | <http://localhost:8086> | `/health`; `/ready` checks PostgreSQL |
| Knowledge | <http://localhost:8087> | `/health`; `/ready` checks PostgreSQL |
| AI | <http://localhost:8088> | `/health`; mock mode is not production acceptance |
| Audit | <http://localhost:8089> | `/health`; `/ready` checks PostgreSQL |
| Reporting | <http://localhost:8090> | `/health`; `/ready` checks PostgreSQL |

## Infrastructure endpoints

| Component | Local endpoint |
|---|---|
| PostgreSQL | `localhost:5432` |
| Redis | `localhost:6379` |
| RabbitMQ AMQP | `localhost:5672` |
| RabbitMQ management | <http://localhost:15672> |
| MinIO API | <http://localhost:9000> |
| MinIO console | <http://localhost:9001> |
| Qdrant | <http://localhost:6333/dashboard> |
| Prometheus | <http://localhost:9090> |
| Grafana | <http://localhost:3002> |
| Loki | <http://localhost:3100> |

Credentials come from `.env`; this document intentionally does not publish them. The production Compose and Kubernetes configurations require externally supplied secrets and must not reuse local-development values.
