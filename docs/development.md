# EOMP — Development Guide

> Development guidelines, conventions, and workflows.

## Branch Strategy

```
main            — Production-ready code
develop         — Integration branch
feature/*       — New features
bugfix/*        — Bug fixes
hotfix/*        — Critical production fixes
```

## Commit Convention

```
feat:      New feature
fix:       Bug fix
refactor:  Code refactoring
test:      Adding or updating tests
docs:      Documentation changes
chore:     Maintenance tasks
ci:        CI/CD changes
```

Examples:
```
feat(auth): initialize authentication service
chore(infra): add postgres redis rabbitmq
feat(web): initialize Nuxt application
```

## Code Style

### Go
- `gofmt` for formatting
- `go vet` for static analysis
- Clean Architecture pattern
- Structured logging (slog)
- Context propagation
- Dependency injection via constructor

### Frontend
- ESLint + Prettier
- TypeScript strict mode
- Composition API
- `<script setup>` syntax

## Adding a New Service

1. Create directory under `services/<name>/`
2. Initialize Go module: `go mod init eomp/services/<name>`
3. Create `cmd/server/main.go` with health endpoint
4. Create `Dockerfile`
5. Add to `docker-compose.yml` (when ready)
6. Update documentation
