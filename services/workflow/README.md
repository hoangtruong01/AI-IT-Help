# EOMP — workflow Service

> Workflow Service - Multi-step Business Processes, Approvals & State Machine

## Port

- HTTP: `8085`

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Service health check |
| `GET` | `/api/health` | API Gateway alias health check |

## Structure

``
services/workflow/
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
cd services/workflow
go run ./cmd/server
``