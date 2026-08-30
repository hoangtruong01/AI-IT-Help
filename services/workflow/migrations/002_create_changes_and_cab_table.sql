-- =============================================================================
-- Migration: 002_create_changes_and_cab_table.sql
-- Database: workflow_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Change Requests Table (ITIL RFC - Request for Change)
CREATE TABLE IF NOT EXISTS change_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    change_number VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    change_type VARCHAR(50) NOT NULL DEFAULT 'NORMAL', -- STANDARD, NORMAL, EMERGENCY, MAJOR
    category VARCHAR(100) NOT NULL DEFAULT 'Infrastructure',
    priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    risk_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    impact_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    probability_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, CAB_REVIEW, APPROVED, REJECTED, SCHEDULED, IMPLEMENTING, COMPLETED, FAILED, CANCELLED
    requester_id VARCHAR(100) NOT NULL,
    requester_name VARCHAR(255) NOT NULL,
    requester_email VARCHAR(255) NOT NULL,
    assigned_to_id VARCHAR(100),
    assigned_to_name VARCHAR(255),
    reason_for_change TEXT NOT NULL,
    implementation_plan TEXT NOT NULL,
    rollback_plan TEXT NOT NULL,
    test_plan TEXT NOT NULL,
    scheduled_start_time TIMESTAMP WITH TIME ZONE,
    scheduled_end_time TIMESTAMP WITH TIME ZONE,
    actual_start_time TIMESTAMP WITH TIME ZONE,
    actual_end_time TIMESTAMP WITH TIME ZONE,
    downtime_required BOOLEAN NOT NULL DEFAULT FALSE,
    downtime_minutes INT NOT NULL DEFAULT 0,
    cab_required_count INT NOT NULL DEFAULT 2, -- Default 2 for EMERGENCY/MAJOR, 1 for NORMAL, 0 for STANDARD
    cab_approved_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_changes_number ON change_requests(change_number);
CREATE INDEX IF NOT EXISTS idx_changes_status ON change_requests(status);
CREATE INDEX IF NOT EXISTS idx_changes_type ON change_requests(change_type);
CREATE INDEX IF NOT EXISTS idx_changes_risk ON change_requests(risk_level);
CREATE INDEX IF NOT EXISTS idx_changes_scheduled ON change_requests(scheduled_start_time, scheduled_end_time);

-- 2. CAB Reviews & Voting Table
CREATE TABLE IF NOT EXISTS cab_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    change_id UUID NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    reviewer_id VARCHAR(100) NOT NULL,
    reviewer_name VARCHAR(255) NOT NULL,
    reviewer_role VARCHAR(100) NOT NULL DEFAULT 'CAB Member',
    vote VARCHAR(50) NOT NULL, -- APPROVED, REJECTED, ABSTAIN
    comments TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_change_reviewer UNIQUE (change_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_cab_reviews_change ON cab_reviews(change_id);
CREATE INDEX IF NOT EXISTS idx_cab_reviews_reviewer ON cab_reviews(reviewer_id);

-- 3. Seed Initial ITIL Change Requests
INSERT INTO change_requests (
    id, change_number, title, description, change_type, category,
    priority, risk_level, impact_level, probability_level, status,
    requester_id, requester_name, requester_email, assigned_to_id, assigned_to_name,
    reason_for_change, implementation_plan, rollback_plan, test_plan,
    scheduled_start_time, scheduled_end_time, downtime_required, downtime_minutes,
    cab_required_count, cab_approved_count, created_at, updated_at
) VALUES
(
    'e0000000-0000-0000-0000-000000000001',
    'CHG-2001',
    'Kubernetes Production Ingress Gateway Upgrade to Traefik v3.1',
    'Upgrade production Kubernetes Ingress controllers across all worker nodes to support HTTP/3 QUIC protocol and enhanced TLS 1.3 cipher security.',
    'MAJOR',
    'DevOps & Infrastructure',
    'HIGH',
    'HIGH',
    'HIGH',
    'MEDIUM',
    'CAB_REVIEW',
    'u2',
    'Alex Rivera (Network Architect)',
    'alex.rivera@eomp.local',
    'u2',
    'Alex Rivera (Network Architect)',
    'Deprecation of Traefik v2.x upstream support and requirement for zero-trust mTLS header validation.',
    '1. Apply Helm values v3.1 on Staging\n2. Perform blue-green ingress cutover on Prod cluster\n3. Verify HTTP 200 on health probes.',
    'Helm rollback to revision 42 (Traefik v2.10) within 3 minutes if error rate exceeds 0.5%.',
    'Automated synthetic test suite checking /health endpoints across all 11 microservices.',
    CURRENT_TIMESTAMP + INTERVAL '2 days',
    CURRENT_TIMESTAMP + INTERVAL '2 days' + INTERVAL '2 hours',
    FALSE,
    0,
    2,
    1,
    CURRENT_TIMESTAMP - INTERVAL '2 days',
    CURRENT_TIMESTAMP - INTERVAL '3 hours'
),
(
    'e0000000-0000-0000-0000-000000000002',
    'CHG-2002',
    'PostgreSQL 16 Analytical Partial Index Creation for SLA Auditing',
    'Add CONCURRENTLY partial indexes on `tickets(sla_resolution_deadline)` where `status != RESOLVED` to accelerate BI metrics calculation.',
    'NORMAL',
    'Database & Storage',
    'MEDIUM',
    'LOW',
    'LOW',
    'LOW',
    'APPROVED',
    'u4',
    'Marcus Vance (Lead DBA)',
    'marcus.vance@eomp.local',
    'u4',
    'Marcus Vance (Lead DBA)',
    'Speed up SLA breach warning worker batch query from 1,200ms down to 18ms.',
    'Execute `CREATE INDEX CONCURRENTLY idx_tickets_active_sla ON tickets(sla_resolution_deadline) WHERE status != RESOLVED;`',
    'Execute `DROP INDEX CONCURRENTLY IF EXISTS idx_tickets_active_sla;`',
    'Verify `EXPLAIN ANALYZE` execution plan reflects Index Scan.',
    CURRENT_TIMESTAMP + INTERVAL '1 day',
    CURRENT_TIMESTAMP + INTERVAL '1 day' + INTERVAL '30 minutes',
    FALSE,
    0,
    1,
    1,
    CURRENT_TIMESTAMP - INTERVAL '1 day',
    CURRENT_TIMESTAMP - INTERVAL '5 hours'
),
(
    'e0000000-0000-0000-0000-000000000003',
    'CHG-2003',
    'Emergency Zero-Day CVE-2026-8819 Patching on Core Edge Routers',
    'Urgent vendor microcode update to mitigate unauthenticated buffer overflow vulnerability on border gateway BGP routers.',
    'EMERGENCY',
    'Network & Access',
    'CRITICAL',
    'CRITICAL',
    'CRITICAL',
    'LOW',
    'CAB_REVIEW',
    'u1',
    'Sarah Jenkins (IT Security Lead)',
    'sarah.jenkins@eomp.local',
    'u2',
    'Alex Rivera (Network Architect)',
    'Critical security advisory requiring remediation within 24 hours per SOC2 compliance policy.',
    '1. Isolate primary router 01 via VRRP failover\n2. Flash firmware 15.4.1-P2\n3. Warm reload\n4. Repeat for redundant node 02.',
    'Fallback to dual-homed LTE backup WAN interface if VRRP failover fails.',
    'BGP route flapping tests and traceroute latency verification.',
    CURRENT_TIMESTAMP + INTERVAL '6 hours',
    CURRENT_TIMESTAMP + INTERVAL '7 hours',
    TRUE,
    15,
    2,
    1,
    CURRENT_TIMESTAMP - INTERVAL '4 hours',
    CURRENT_TIMESTAMP - INTERVAL '1 hour'
),
(
    'e0000000-0000-0000-0000-000000000004',
    'CHG-2004',
    'Standard Bi-Weekly Corporate CA Root Certificate Renewal',
    'Routine automated certificate rotation for internal developer staging subdomains (*.stage.eomp.local).',
    'STANDARD',
    'IT Security & Access',
    'LOW',
    'LOW',
    'LOW',
    'LOW',
    'SCHEDULED',
    'u1',
    'Sarah Jenkins (IT Security Lead)',
    'sarah.jenkins@eomp.local',
    'u1',
    'Sarah Jenkins (IT Security Lead)',
    'Prevent certificate expiration warnings for remote developers.',
    'Cert-manager automated renewal via ACME / Vault internal issuer.',
    'Revert to backup wildcard cert secret.',
    'curl -Iv https://api.stage.eomp.local/health',
    CURRENT_TIMESTAMP + INTERVAL '4 days',
    CURRENT_TIMESTAMP + INTERVAL '4 days' + INTERVAL '10 minutes',
    FALSE,
    0,
    0,
    0,
    CURRENT_TIMESTAMP - INTERVAL '1 day',
    CURRENT_TIMESTAMP
)
ON CONFLICT (id) DO NOTHING;

-- 4. Seed CAB Reviews
INSERT INTO cab_reviews (id, change_id, reviewer_id, reviewer_name, reviewer_role, vote, comments, reviewed_at)
VALUES
(
    'b0000000-0000-0000-0000-000000000001',
    'e0000000-0000-0000-0000-000000000001',
    'u1',
    'Sarah Jenkins (IT Security Lead)',
    'Security Officer',
    'APPROVED',
    'Security evaluation complete. Zero-trust mTLS header validation passes compliance.',
    CURRENT_TIMESTAMP - INTERVAL '1 day'
),
(
    'b0000000-0000-0000-0000-000000000002',
    'e0000000-0000-0000-0000-000000000002',
    'u2',
    'Alex Rivera (Network Architect)',
    'Infrastructure Architect',
    'APPROVED',
    'Approved. Minimal CPU lock impact on PostgreSQL primary node.',
    CURRENT_TIMESTAMP - INTERVAL '10 hours'
),
(
    'b0000000-0000-0000-0000-000000000003',
    'e0000000-0000-0000-0000-000000000003',
    'u1',
    'Sarah Jenkins (IT Security Lead)',
    'Security Officer',
    'APPROVED',
    'Emergency authorization granted. Security patch must proceed within maintenance window.',
    CURRENT_TIMESTAMP - INTERVAL '2 hours'
)
ON CONFLICT (id) DO NOTHING;
