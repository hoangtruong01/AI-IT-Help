-- EOMP audit service base schema. Production data is never seeded by migrations.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    audit_sequence BIGSERIAL NOT NULL UNIQUE,
    event_type VARCHAR(100) NOT NULL,
    actor_id VARCHAR(100),
    actor_name VARCHAR(150) NOT NULL,
    actor_email VARCHAR(150) NOT NULL,
    actor_role VARCHAR(50) NOT NULL,
    service_name VARCHAR(50) NOT NULL,
    ip_address VARCHAR(50) NOT NULL,
    user_agent VARCHAR(255) DEFAULT '',
    status VARCHAR(50) NOT NULL DEFAULT 'SUCCESS',
    resource_type VARCHAR(100) NOT NULL,
    resource_id VARCHAR(100) NOT NULL,
    old_values JSONB,
    new_values JSONB,
    checksum_sha256 VARCHAR(64) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_event_type ON audit_logs (event_type);
CREATE INDEX IF NOT EXISTS idx_audit_actor_email ON audit_logs (actor_email);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON audit_logs (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_status ON audit_logs (status);

CREATE TABLE IF NOT EXISTS security_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_code VARCHAR(100) NOT NULL,
    severity VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    source_ip VARCHAR(50) NOT NULL,
    target_endpoint VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    is_blocked BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_events_severity
    ON security_events (severity);
