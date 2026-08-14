# EOMP — Deployment

> Deployment documentation will be expanded when CI/CD pipeline is configured.

## Development

```bash
# Infrastructure
docker compose up -d

# Frontend
cd apps/web && pnpm dev

# Backend services
cd services/<name> && go run ./cmd/server
```

## Docker Build

```bash
# Build a single service
docker build -t eomp-<service>:latest -f services/<service>/Dockerfile services/<service>

# Build all via compose (when service containers are added)
docker compose build
```

## Environments

| Environment | Purpose |
|---|---|
| `development` | Local development |
| `staging` | Pre-production testing |
| `production` | Live system |

## CI/CD

Jenkins pipeline defined in `Jenkinsfile` (to be created in Step 14).
