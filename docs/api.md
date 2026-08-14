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

## AI Service Endpoints

Base URL: `http://localhost:8088`

### AI Chat (Assistant)

```
POST /api/ai/chat
```

Request:
```json
{
  "session_id": "session-123",
  "messages": [
    {
      "role": "user",
      "content": "How do I request a new laptop?"
    }
  ]
}
```

Response:
```json
{
  "answer": "To request a new laptop, navigate to the Asset Management module...",
  "citations": [
    {
      "article_id": "kb-001",
      "title": "IT Asset Request Workflow",
      "score": 0.95
    }
  ],
  "confidence": 0.95,
  "tokens_used": 120
}
```

### AI Ticket Triage & Classification

```
POST /api/ai/analyze-ticket
```

Request:
```json
{
  "title": "VPN connection drops every 10 minutes",
  "description": "User is unable to maintain remote connection to internal services."
}
```

Response:
```json
{
  "ticket_id": "mock-ticket-001",
  "suggested_category": "Network / VPN",
  "priority": "medium",
  "summary": "Issue regarding: VPN connection drops every 10 minutes",
  "suggested_resolution": "Check VPN client version and verify MTU settings.",
  "requires_human_review": true,
  "created_at": "2026-08-14T10:45:00Z"
}
```
