# EOMP — API Reference

> API documentation will be added as services are implemented.

## API Gateway

Base URL: `http://localhost:8080`

### Health Check

```
GET /health
```

Response:
```json
{
  "status": "ok",
  "service": "gateway",
  "version": "0.1.0"
}
```

## Service Health Endpoints

Each service exposes a health endpoint:

| Service | Endpoint |
|---|---|
| Gateway | `GET /health` |
| Auth | `GET /health` |
| Employee | `GET /health` |
| Asset | `GET /health` |
| Helpdesk | `GET /health` |
| Workflow | `GET /health` |
| Notification | `GET /health` |
| Knowledge | `GET /health` |
| AI | `GET /health` |
| Audit | `GET /health` |
| Reporting | `GET /health` |
