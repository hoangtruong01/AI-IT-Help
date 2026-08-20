-- ============================================================
-- EOMP Security & Compliance Audit Service - Migration 001
-- Tables: audit_logs, security_events
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Immutable Audit Logs Table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL, -- AUTH_LOGIN_SUCCESS, TICKET_UPDATE, ASSET_DELETE, APPROVAL_DECISION, ROLE_CHANGE, CONFIG_UPDATE
    actor_id UUID,
    actor_name VARCHAR(150) NOT NULL,
    actor_email VARCHAR(150) NOT NULL,
    actor_role VARCHAR(50) NOT NULL, -- ROLE_ADMIN, ROLE_MANAGER, ROLE_AGENT, ROLE_EMPLOYEE
    service_name VARCHAR(50) NOT NULL, -- auth, employee, asset, helpdesk, workflow, etc.
    ip_address VARCHAR(50) NOT NULL,
    user_agent VARCHAR(255) DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'SUCCESS', -- SUCCESS, FORBIDDEN, FAILED
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    checksum_sha256 VARCHAR(64) NOT NULL, -- Immutable Tamper-evident SHA-256 Hash
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_event_type ON audit_logs (event_type);
CREATE INDEX IF NOT EXISTS idx_audit_actor_email ON audit_logs (actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_status ON audit_logs (status);

-- 2. Security Events & Alert Trail
CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_code VARCHAR(100) NOT NULL, -- BRUTE_FORCE_ATTEMPT, RATE_LIMIT_EXCEEDED, PRIVILEGE_ESCALATION_BLOCKED, SENSITIVE_DATA_ACCESS
    severity VARCHAR(50) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    source_ip VARCHAR(50) NOT NULL,
    target_endpoint VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events (severity);

-- ============================================================
-- SEED DATA (Realistic SOC2 / ISO 27001 Audit Records)
-- ============================================================

INSERT INTO audit_logs (id, event_type, actor_id, actor_name, actor_email, actor_role, service_name, ip_address, user_agent, status, resource_type, resource_id, old_values, new_values, checksum_sha256, created_at)
VALUES
    (
        'a0000000-0000-0000-0000-000000000001',
        'AUTH_LOGIN_SUCCESS',
        'u0000000-0000-0000-0000-000000000001',
        'Administrator',
        'admin@eomp.local',
        'ROLE_ADMIN',
        'auth',
        '192.168.1.10',
        'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0',
        'SUCCESS',
        'user_session',
        'sess-88910a',
        '{}'::jsonb,
        '{"mfa_verified": true, "token_scope": "full_admin"}'::jsonb,
        'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
        NOW() - INTERVAL '15 minutes'
    ),
    (
        'a0000000-0000-0000-0000-000000000002',
        'ROLE_CHANGE',
        'u0000000-0000-0000-0000-000000000001',
        'Administrator',
        'admin@eomp.local',
        'ROLE_ADMIN',
        'auth',
        '192.168.1.10',
        'Mozilla/5.0 (Windows NT 10.0; Win64; x64) Chrome/128.0',
        'SUCCESS',
        'user',
        'u0000000-0000-0000-0000-000000000004',
        '{"role": "ROLE_AGENT", "department": "IT Support"}'::jsonb,
        '{"role": "ROLE_MANAGER", "department": "IT Security", "elevated_by": "admin@eomp.local"}'::jsonb,
        '9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08',
        NOW() - INTERVAL '42 minutes'
    ),
    (
        'a0000000-0000-0000-0000-000000000003',
        'ASSET_DELETE',
        'u0000000-0000-0000-0000-000000000003',
        'Marcus Vance',
        'marcus.vance@eomp.local',
        'ROLE_AGENT',
        'asset',
        '192.168.1.45',
        'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)',
        'SUCCESS',
        'asset',
        'AST-00921',
        '{"asset_tag": "AST-00921", "name": "Dell PowerEdge R740 (Decommissioned)", "status": "RETIRED", "serial": "SN-8829-DEL"}'::jsonb,
        '{"status": "DISPOSED", "disposed_notes": "Hard drives shredded according to NIST 800-88"}'::jsonb,
        '5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8',
        NOW() - INTERVAL '2 hours'
    ),
    (
        'a0000000-0000-0000-0000-000000000004',
        'APPROVAL_DECISION',
        'u0000000-0000-0000-0000-000000000002',
        'Sarah Jenkins',
        'sarah.jenkins@eomp.local',
        'ROLE_MANAGER',
        'workflow',
        '192.168.1.18',
        'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
        'SUCCESS',
        'change_request',
        'CHG-2001',
        '{"status": "CAB_REVIEW", "approved_votes": 1}'::jsonb,
        '{"status": "APPROVED", "approved_votes": 2, "quorum": "2/2", "notes": "Production DB migration risk approved"}'::jsonb,
        '4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a',
        NOW() - INTERVAL '3 hours'
    ),
    (
        'a0000000-0000-0000-0000-000000000005',
        'RBAC_ACCESS_DENIED',
        'u0000000-0000-0000-0000-000000000008',
        'Kenji Sato',
        'kenji.sato@eomp.local',
        'ROLE_EMPLOYEE',
        'gateway',
        '192.168.2.110',
        'curl/8.4.0',
        'FORBIDDEN',
        'audit_logs',
        'api/v1/audit/logs',
        '{}'::jsonb,
        '{"attempted_endpoint": "/api/v1/audit/logs", "error": "INSUFFICIENT_PERMISSIONS", "required_roles": ["ROLE_ADMIN", "ROLE_MANAGER"]}'::jsonb,
        'ef2d127de37b942baad06145e54b0c619a1f22327b2ebbcfbec78f5564afe39d',
        NOW() - INTERVAL '5 hours'
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO security_events (event_code, severity, source_ip, target_endpoint, description, is_blocked, created_at)
VALUES
    ('RBAC_VIOLATION_BLOCKED', 'HIGH', '192.168.2.110', '/api/v1/audit/logs', 'Unauthorized employee account attempted to view administrative audit records', TRUE, NOW() - INTERVAL '5 hours'),
    ('RATE_LIMIT_EXCEEDED', 'MEDIUM', '10.0.4.55', '/api/v1/auth/login', 'Exceeded 5 failed login attempts in 1 minute window -> IP blocked for 15 mins', TRUE, NOW() - INTERVAL '6 hours'),
    ('DATA_MASKING_APPLIED', 'LOW', '192.168.1.10', '/api/v1/auth/login', 'Sanitized sensitive passwords and JWT bearer tokens from trace log output', FALSE, NOW() - INTERVAL '8 hours')
ON CONFLICT (id) DO NOTHING;
