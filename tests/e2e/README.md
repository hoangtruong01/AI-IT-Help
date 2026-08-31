# In-memory simulation test suite

Despite the legacy directory name, the Go tests in this directory are unit and in-memory simulation tests. They exercise domain flows, middleware, concurrency primitives and mocked integrations without deploying EOMP or connecting the real service, PostgreSQL, RabbitMQ, Redis, Qdrant or Kubernetes stack.

Passing `go test ./...` here is useful regression evidence, but it is not deployed end-to-end, network-isolation, disaster-recovery, load or production-readiness evidence. Current acceptance status is maintained only in `docs/IMPLEMENTATION_STATUS.md`.
