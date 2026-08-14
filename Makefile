# ==============================
# EOMP — Makefile
# ==============================
# For Linux / macOS / CI environments.
# Windows users: use scripts/dev.ps1

.PHONY: help dev build test lint format docker-up docker-down logs health clean

# Default target
help: ## Show this help
	@echo "EOMP — Enterprise Operations Management Platform"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

# ==============================
# Development
# ==============================

dev: ## Start frontend development server
	cd apps/web && pnpm dev

build: ## Build all services
	@echo "Building Go services..."
	@for dir in services/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "  Building $$dir..."; \
			cd "$$dir" && go build ./cmd/... && cd ../..; \
		fi \
	done
	@echo "Building frontend..."
	cd apps/web && pnpm build

test: ## Run all tests
	@echo "Testing Go services..."
	@for dir in services/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "  Testing $$dir..."; \
			cd "$$dir" && go test ./... && cd ../..; \
		fi \
	done
	@echo "All tests passed."

lint: ## Run linters
	@echo "Linting Go services..."
	@for dir in services/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			echo "  Linting $$dir..."; \
			cd "$$dir" && go vet ./... && cd ../..; \
		fi \
	done
	@echo "Linting frontend..."
	cd apps/web && pnpm lint

format: ## Format code
	@echo "Formatting Go services..."
	@for dir in services/*/; do \
		if [ -f "$$dir/go.mod" ]; then \
			gofmt -w "$$dir"; \
		fi \
	done

# ==============================
# Docker
# ==============================

docker-up: ## Start all infrastructure services
	docker compose up -d

docker-down: ## Stop all infrastructure services
	docker compose down

docker-reset: ## Stop and remove all volumes
	docker compose down -v

logs: ## Show container logs (use SERVICE=name to filter)
ifdef SERVICE
	docker compose logs -f $(SERVICE)
else
	docker compose logs -f
endif

# ==============================
# Health Check
# ==============================

health: ## Check health of all services
	@echo "=== Infrastructure Health ==="
	@echo -n "PostgreSQL: " && docker compose exec -T postgres pg_isready -U eomp 2>/dev/null && echo "OK" || echo "FAIL"
	@echo -n "Redis:      " && docker compose exec -T redis redis-cli ping 2>/dev/null || echo "FAIL"
	@echo -n "RabbitMQ:   " && docker compose exec -T rabbitmq rabbitmq-diagnostics -q ping 2>/dev/null && echo "OK" || echo "FAIL"
	@echo -n "MinIO:      " && curl -sf http://localhost:9000/minio/health/live >/dev/null 2>&1 && echo "OK" || echo "FAIL"
	@echo -n "Qdrant:     " && curl -sf http://localhost:6333/healthz >/dev/null 2>&1 && echo "OK" || echo "FAIL"
	@echo -n "Prometheus: " && curl -sf http://localhost:9090/-/healthy >/dev/null 2>&1 && echo "OK" || echo "FAIL"
	@echo -n "Grafana:    " && curl -sf http://localhost:3001/api/health >/dev/null 2>&1 && echo "OK" || echo "FAIL"
	@echo -n "Loki:       " && curl -sf http://localhost:3100/ready >/dev/null 2>&1 && echo "OK" || echo "FAIL"

# ==============================
# Cleanup
# ==============================

clean: ## Clean build artifacts
	@echo "Cleaning Go binaries..."
	@find services -name "*.exe" -delete 2>/dev/null || true
	@find services -name "*.test" -delete 2>/dev/null || true
	@echo "Cleaning frontend..."
	cd apps/web && rm -rf .nuxt .output dist node_modules 2>/dev/null || true
	@echo "Clean complete."
