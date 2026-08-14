# ==============================
# Build Stage
# ==============================
FROM golang:1.24-alpine AS builder

WORKDIR /workspace

# Copy shared packages & service code
COPY packages/shared packages/shared
COPY services/reporting services/reporting
COPY go.work .

WORKDIR /workspace/services/reporting

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/server ./cmd/server

# ==============================
# Runtime Stage
# ==============================
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata curl

WORKDIR /app
COPY --from=builder /bin/server /app/server

EXPOSE 8090

HEALTHCHECK --interval=10s --timeout=3s --retries=3 \
  CMD curl -f http://localhost:8090/health || exit 1

ENTRYPOINT ["/app/server"]