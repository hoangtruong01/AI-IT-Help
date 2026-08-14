# EOMP — AI Service

> AI Operations Assistant, Vector Search, Ticket Triage & RAG Layer

## Port

- HTTP: `8088`

## Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/health` | Service health check |
| `GET` | `/api/health` | Gateway health check alias |
| `POST` | `/api/ai/chat` | Conversational AI Assistant with RAG citations |
| `POST` | `/api/ai/analyze-ticket` | Automated ticket classification and recommendation |

## Architecture & Layers

```
services/ai/
├── cmd/server/           # Service entrypoint & graceful shutdown
├── internal/
│   ├── config/           # AI model, embedding & Qdrant configuration
│   ├── handler/          # HTTP controllers (Chat, Triage, Health)
│   ├── model/            # Chat messages, citations, analysis models
│   ├── prompt/           # System prompts & specialized templates
│   ├── provider/         # LLM & Embedding provider abstractions (Mock / OpenAI / Gemini / Ollama)
│   ├── rag/              # Vector database retrieval layer (Qdrant)
│   └── service/          # Business logic orchestrator
├── migrations/           # DB schema migrations (if needed)
├── Dockerfile            # Multi-stage production container build
├── go.mod                # Module definition
└── README.md
```

## AI Safety Rules (Section 35)

As an enterprise operations platform, the AI assistant acts strictly as an **advisory, analyzer, and recommendation engine**.
It is explicitly prohibited from autonomously:
- Deleting entities
- Disabling user accounts or revoking credentials
- Retiring assets
- Approving workflow requests
- Triggering production deployments

All AI responses on ticket classification include `requires_human_review: true`.