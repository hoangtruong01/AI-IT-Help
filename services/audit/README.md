# EOMP — audit Service

> Audit Service - Immutable Audit Logging, Compliance Tracking & Security Events

## Port

- HTTP: `8089`

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Service health check |
| `GET` | `/api/health` | API Gateway alias health check |

## Structure

``
services/audit/
├── cmd/
│   └── server/
│       ├── main.go       # Server entrypoint with graceful shutdown
│       └── main_test.go  # Unit tests
├── internal/
│   ├── config/           # Configuration management
│   ├── handler/          # HTTP request handlers
│   ├── middleware/       # Service-specific middleware
│   ├── model/            # Domain models (to be added in module phase)
│   ├── repository/       # Data persistence (to be added in module phase)
│   └── service/          # Business logic (to be added in module phase)
├── migrations/           # Database migration files
├── Dockerfile            # Multi-stage production container build
├── go.mod                # Module definitions
└── README.md
``

## Running Locally

``bash
cd services/audit
go run ./cmd/server
``