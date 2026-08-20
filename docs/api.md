# EOMP — API Reference

> API specification for all EOMP microservices exposed via the API Gateway (`http://localhost:8080`).

---

## 1. API Gateway Routing Matrix

Base URL: `http://localhost:8080`

| Microservice | Internal Port | Gateway Path Prefix | Auth Filter |
|---|:---:|---|:---:|
| **Auth Service** | `:8081` | `/api/v1/auth/` | Public (Login/Register) / JWT (Me) |
| **Employee Service** | `:8082` | `/api/v1/employees/`, `/api/v1/departments/` | JWT Required |
| **Asset Service** | `:8083` | `/api/v1/assets/`, `/api/v1/cmdb/` | JWT Required |
| **Helpdesk Service** | `:8084` | `/api/v1/tickets/`, `/api/v1/services/`, `/api/v1/problems/` | JWT Required |
| **Workflow Engine** | `:8085` | `/api/v1/workflows/`, `/api/v1/approvals/`, `/api/v1/changes/` | JWT Required |
| **Notification Service** | `:8086` | `/api/v1/notifications/` | JWT Required |
| **Knowledge Base** | `:8087` | `/api/v1/knowledge/` | JWT Required |
| **AI Operations Copilot** | `:8088` | `/api/v1/ai/` | JWT Required |

---

## 2. Knowledge Base Service (`services/knowledge` — Port 8087)

### 2.1. Get Dashboard Statistics
`GET /api/v1/knowledge/stats`

**Response (`200 OK`):**
```json
{
  "total_articles": 6,
  "total_categories": 5,
  "total_runbooks": 4,
  "total_views": 4440
}
```

### 2.2. List Categories
`GET /api/v1/knowledge/categories`

**Response (`200 OK`):**
```json
[
  {
    "id": "c0000000-0000-0000-0000-000000000001",
    "name": "IT Security & Access",
    "code": "sec",
    "icon": "i-lucide-shield-check",
    "description": "MFA tokens, SSO identity, zero-trust security & access management policies."
  }
]
```

### 2.3. List Articles (Paginated & Filterable)
`GET /api/v1/knowledge/articles?category=sec&page=1&page_size=20`

**Response (`200 OK`):**
```json
{
  "data": [
    {
      "id": "a0000000-0000-0000-0000-000000000001",
      "category_id": "c0000000-0000-0000-0000-000000000001",
      "category_name": "IT Security & Access",
      "category_code": "sec",
      "title": "How to Reset User MFA and Okta Verify Tokens",
      "slug": "how-to-reset-user-mfa-tokens",
      "summary": "Official standard operating procedure for IT Support Agents to verify employee identity and securely reset multi-factor authentication tokens in Okta / Keycloak.",
      "content": "## Step 1: Verify Identity...",
      "tags": ["MFA", "Okta", "Security", "SOP"],
      "author_name": "Sarah Jenkins (IT Security Lead)",
      "view_count": 1240,
      "is_published": true
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

### 2.4. Semantic & Fulltext Search
`GET /api/v1/knowledge/search?q=VPN`

**Response (`200 OK`):**
```json
{
  "query": "VPN",
  "total": 2,
  "results": [
    {
      "id": "a0000000-0000-0000-0000-000000000002",
      "type": "article",
      "title": "Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide",
      "snippet": "Comprehensive resolution guide for remote engineers experiencing VPN disconnects...",
      "category": "Network & Connectivity",
      "score": 0.98,
      "tags": ["VPN", "WireGuard", "Network"],
      "slug_or_code": "vpn-troubleshooting-guide"
    },
    {
      "id": "r0000000-0000-0000-0000-000000000002",
      "type": "runbook",
      "title": "RB-NET-01: Emergency VPN Tunnel Failover SOP",
      "snippet": "Failover procedure when primary WireGuard VPN gateway server exhibits packet loss...",
      "category": "Network",
      "score": 0.95,
      "tags": ["SOP", "Runbook"],
      "slug_or_code": "RB-NET-01"
    }
  ]
}
```

### 2.5. List SOP Runbooks
`GET /api/v1/knowledge/runbooks`

**Response (`200 OK`):**
```json
{
  "data": [
    {
      "id": "r0000000-0000-0000-0000-000000000001",
      "code": "RB-SEC-02",
      "title": "User MFA Token Reset and Identity Verification SOP",
      "category": "IT Security",
      "description": "Standardized operational procedure for identity authentication and Okta/Keycloak multi-factor token reissue.",
      "prerequisites": "1. Active IT Support Agent credentials.\n2. Manager written approval.",
      "steps_json": [
        {
          "step": 1,
          "action": "Verify Employee Identity via secondary communication channel",
          "command": "Check employee record in /employees directory",
          "expected": "Active status confirmed"
        }
      ],
      "rollback_steps": "If user fails verification, lock account for 30 minutes and alert SOC.",
      "author_name": "Sarah Jenkins (IT Security Lead)",
      "is_active": true
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20,
  "total_pages": 1
}
```

---

## 3. AI Operations Copilot Service (`services/ai` — Port 8088)

### 3.1. Interactive Copilot Chat (RAG-Grounded)
`POST /api/v1/ai/chat`

**Request:**
```json
{
  "session_id": "session-123",
  "messages": [
    {
      "role": "user",
      "content": "How to reset user MFA token?"
    }
  ]
}
```

**Response (`200 OK`):**
```json
{
  "answer": "### 🛡️ Standard Operating Procedure: User MFA & Okta Token Reset\n\n1. **Mandatory Identity Verification**...\n2. **Okta Admin Reset**...\n3. **Re-enrollment**...",
  "citations": [
    {
      "article_id": "a0000000-0000-0000-0000-000000000001",
      "title": "How to Reset User MFA and Okta Verify Tokens",
      "score": 0.96,
      "category": "IT Security",
      "type": "article"
    },
    {
      "article_id": "r0000000-0000-0000-0000-000000000001",
      "title": "RB-SEC-02: User MFA Token Reset and Identity Verification SOP",
      "score": 0.95,
      "category": "IT Security",
      "type": "runbook"
    }
  ],
  "confidence": 0.955,
  "tokens_used": 145,
  "fallback_mode": false
}
```

### 3.2. Helpdesk Ticket Auto-Triage & Diagnostics
`POST /api/v1/ai/analyze-ticket`

**Request:**
```json
{
  "title": "Cannot connect to VPN Staging Server",
  "description": "Remote engineer is unable to maintain WireGuard connection to internal staging subnet. Handshake times out every 10 minutes."
}
```

**Response (`200 OK`):**
```json
{
  "ticket_id": "AI-TK-1094",
  "suggested_category": "Network & Access",
  "priority": "HIGH",
  "summary": "VPN tunnel connection timeout to Staging Server cluster.",
  "root_cause": "WireGuard handshake packet drop or MTU mismatch on Gateway subnet.",
  "suggested_resolution": "1. Instruct user to flush DNS cache and verify client MTU is 1380.\n2. Verify upstream VPN gateway status on 10.8.0.1.\n3. Refer to Runbook RB-NET-01 if gateway failover is required.",
  "confidence": 0.94,
  "citations": [
    {
      "article_id": "a0000000-0000-0000-0000-000000000002",
      "title": "Corporate WireGuard & GlobalProtect VPN Troubleshooting Guide",
      "score": 0.95,
      "category": "Network & Access",
      "type": "article"
    },
    {
      "article_id": "r0000000-0000-0000-0000-000000000002",
      "title": "RB-NET-01: Emergency VPN Tunnel Failover SOP",
      "score": 0.94,
      "category": "Network & Access",
      "type": "runbook"
    }
  ],
  "requires_human_review": true,
  "created_at": "2026-08-19T10:00:00Z"
}
```

---

## 4. Problem Management API (`services/helpdesk` — Port 8084)

### 4.1. Get Problem Statistics
`GET /api/v1/problems/stats`

**Response (`200 OK`):**
```json
{
  "total_problems": 3,
  "under_investigation": 1,
  "known_errors": 2,
  "resolved_problems": 1,
  "total_linked_tickets": 4
}
```

### 4.2. List Problems
`GET /api/v1/problems?category=Network&status=KNOWN_ERROR&page=1&page_size=20`

### 4.3. Get Problem by ID with Linked Incidents
`GET /api/v1/problems/{id}`

### 4.4. Create ITIL Problem Record
`POST /api/v1/problems`

### 4.5. Update Root Cause Analysis (RCA) & Workaround (KEDB)
`PATCH /api/v1/problems/{id}/rca`

### 4.6. Link Incident Ticket to Problem
`POST /api/v1/problems/{id}/link-incident`

### 4.7. Cascade Resolution to All Linked Incidents (Test Case 7.1)
`PATCH /api/v1/problems/{id}/status`
- When payload `{"status": "RESOLVED", "resolution": "..."}` is sent, the system automatically resolves all linked incident tickets and writes audit timeline records.

---

## 5. Change Advisory Board (CAB) & RFC Lifecycle (`services/workflow` — Port 8085)

### 5.1. Get Change Management Dashboard KPIs
`GET /api/v1/changes/stats`

### 5.2. Get Maintenance Window Schedule Calendar
`GET /api/v1/changes/calendar?start=2026-08-01T00:00:00Z&end=2026-09-01T00:00:00Z`

### 5.3. Create Request for Change (RFC) with 3x3 Risk Matrix
`POST /api/v1/changes`

### 5.4. Submit CAB Member Vote
`POST /api/v1/changes/{id}/cab-vote`
- Payload: `{"reviewer_id": "u1", "reviewer_name": "Sarah Jenkins", "reviewer_role": "Security Lead", "vote": "APPROVED", "comments": "Approved"}`

### 5.5. Transition Change Lifecycle State (Enforces Quorum - Test Case 7.2)
`PATCH /api/v1/changes/{id}/status`
- If change is `EMERGENCY` or `MAJOR` and `cab_approved_count < 2`, the API rejects with `403 Forbidden` (`insufficient CAB approvals`).

---

## 6. Enterprise Observability & SRE Health Mesh (`services/gateway` — Port 8080)

### 6.1. Prometheus 2.0 Text Exposition Format (Test Case 8.1)
`GET /metrics`
- Scrapes RED metrics across microservices: `http_requests_total`, `http_request_duration_seconds`, `service_uptime_seconds`, `service_memory_bytes`, `service_goroutines_count`.
- Content-Type: `text/plain; version=0.0.4; charset=utf-8`

### 6.2. Cluster Overview RED Metrics
`GET /api/v1/monitoring/overview`

**Response (`200 OK`):**
```json
{
  "total_services": 11,
  "online_services": 11,
  "degraded_services": 0,
  "offline_services": 0,
  "cluster_health_pct": 100.0,
  "total_requests_per_min": 1420,
  "avg_latency_p95_ms": 12.4,
  "error_rate_pct": 0.02
}
```

### 6.3. List Microservice Health Matrix
`GET /api/v1/monitoring/services`

### 6.4. Active Health Probe (Outage Detection < 5s - Test Case 8.2)
`POST /api/v1/monitoring/probe/{id}`
- Actively tests target microservice `/health` endpoint and updates status immediately.

### 6.5. SRE Live Log Streamer (Tail -f)
`GET /api/v1/monitoring/logs?service=helpdesk&level=INFO&limit=50&search=cascade`


