# EOMP — Testing

> Testing strategy and guidelines.

## Go Services

```bash
# Run all tests for a service
cd services/<name>
go test ./...

# Run with verbose output
go test -v ./...

# Run with coverage
go test -cover ./...

# Run specific test
go test -run TestFunctionName ./...
```

## Frontend

```bash
cd apps/web

# Type check
pnpm typecheck

# Lint
pnpm lint

# Unit tests (when configured)
pnpm test
```

## Testing Categories

| Category | Tool | Scope |
|---|---|---|
| Unit tests (Go) | `go test` | Per service |
| Static analysis (Go) | `go vet` | Per service |
| Lint (Go) | `golangci-lint` | Per service |
| Type check (Frontend) | `vue-tsc` | Frontend |
| Lint (Frontend) | ESLint | Frontend |
| Integration tests | TBD | Cross-service |
| E2E tests | TBD | Full system |
