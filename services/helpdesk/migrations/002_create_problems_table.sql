-- =============================================================================
-- Migration: 002_create_problems_table.sql
-- Database: helpdesk_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Problems Table (ITIL Problem Management & KEDB)
CREATE TABLE IF NOT EXISTS problems (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    problem_number VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'Infrastructure',
    priority VARCHAR(50) NOT NULL DEFAULT 'HIGH',
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    impact VARCHAR(50) NOT NULL DEFAULT 'HIGH',
    urgency VARCHAR(50) NOT NULL DEFAULT 'HIGH',
    assignee_id VARCHAR(100),
    assignee_name VARCHAR(255),
    root_cause TEXT,
    workaround TEXT,
    resolution TEXT,
    is_known_error BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_problems_number ON problems(problem_number);
CREATE INDEX IF NOT EXISTS idx_problems_status ON problems(status);
CREATE INDEX IF NOT EXISTS idx_problems_category ON problems(category);
CREATE INDEX IF NOT EXISTS idx_problems_known_error ON problems(is_known_error);

-- 2. Problem - Incident Links Table (Many-to-One / Many-to-Many Aggregation)
CREATE TABLE IF NOT EXISTS problem_incident_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    ticket_number VARCHAR(50) NOT NULL,
    ticket_title VARCHAR(255) NOT NULL,
    linked_by VARCHAR(255) NOT NULL DEFAULT 'IT Problem Manager',
    linked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_problem_ticket UNIQUE (problem_id, ticket_id)
);

CREATE INDEX IF NOT EXISTS idx_prob_inc_problem ON problem_incident_links(problem_id);
CREATE INDEX IF NOT EXISTS idx_prob_inc_ticket ON problem_incident_links(ticket_id);

-- 3. Seed Initial ITIL Problems & Linked Incidents
INSERT INTO problems (
    id, problem_number, title, description, category, priority, status,
    impact, urgency, assignee_id, assignee_name, root_cause, workaround,
    resolution, is_known_error, created_at, updated_at
) VALUES
(
    'b0000000-0000-0000-0000-000000000001',
    'PRB-1001',
    'Intermittent WireGuard VPN Gateway Handshake Drops under High Concurrency',
    'Multiple remote software engineers report sporadic VPN tunnel disconnections every 15-20 minutes when active peer count exceeds 250 connections on Gateway 10.8.0.1.',
    'Network & Access',
    'CRITICAL',
    'KNOWN_ERROR',
    'HIGH',
    'HIGH',
    'u2',
    'Alex Rivera (Network Architect)',
    '### 5-Whys Root Cause Analysis\n1. Why did tunnels drop? Packet loss on UDP port 51820.\n2. Why was there packet loss? Linux kernel UDP buffer overrun.\n3. Why overrun? `net.core.rmem_max` was set to default 212KB.\n4. Why default? Provisioning template missed sysctl high-load network tuning.\n5. Root Cause: Sysctl socket memory buffers inadequate for >200 concurrent WireGuard peers.',
    'Execute `sysctl -w net.core.rmem_max=26214400 net.core.wmem_max=26214400` on primary VPN gateway host and restart wg-quick service.',
    'Permanent fix applied via Ansible configuration baseline playbook v2.4 across all production VPN nodes.',
    TRUE,
    CURRENT_TIMESTAMP - INTERVAL '3 days',
    CURRENT_TIMESTAMP - INTERVAL '1 hour'
),
(
    'b0000000-0000-0000-0000-000000000002',
    'PRB-1002',
    'PostgreSQL 16 Connection Pool Exhaustion on Reporting Analytics Query Burst',
    'Application services experience connection timeouts when automated BI reporting jobs trigger unbounded sequential joins during peak business hours.',
    'DevOps & Infrastructure',
    'HIGH',
    'UNDER_INVESTIGATION',
    'HIGH',
    'MEDIUM',
    'u4',
    'Marcus Vance (Lead Database Administrator)',
    '### Investigation Status\nPgBouncer connection pool configured with `max_client_conn = 100` and `pool_mode = session`. Long-running ETL worker queries hold persistent idle-in-transaction locks.',
    'Scale PgBouncer replica pool to transaction mode on port 6432 for read-only analytical workloads.',
    NULL,
    FALSE,
    CURRENT_TIMESTAMP - INTERVAL '1 day',
    CURRENT_TIMESTAMP
),
(
    'b0000000-0000-0000-0000-000000000003',
    'PRB-1003',
    'Okta MFA WebAuthn Security Key Registration Desynchronization',
    'Hardware FIDO2 YubiKey registration encounters error 400 when users switch between multiple corporate browser profiles.',
    'IT Security & Access',
    'MEDIUM',
    'WORKAROUND_FOUND',
    'MEDIUM',
    'LOW',
    'u1',
    'Sarah Jenkins (IT Security Lead)',
    'RP ID (Relying Party Identifier) origin mismatch when accessing staging vs production SSO portals.',
    'Enforce unified subdomain `id.eomp.local` with strict WebAuthn origin validation.',
    NULL,
    TRUE,
    CURRENT_TIMESTAMP - INTERVAL '5 days',
    CURRENT_TIMESTAMP - INTERVAL '12 hours'
)
ON CONFLICT (id) DO NOTHING;

-- 4. Seed Problem Incident Links
INSERT INTO problem_incident_links (id, problem_id, ticket_id, ticket_number, ticket_title, linked_by, linked_at)
SELECT
    'l0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000001',
    t.id,
    t.ticket_number,
    t.title,
    'Alex Rivera (Network Architect)',
    CURRENT_TIMESTAMP - INTERVAL '2 days'
FROM tickets t
WHERE t.ticket_number = 'INC-1001'
ON CONFLICT (id) DO NOTHING;
