-- =============================================================================
-- Migration: 002_create_login_audit_logs_table.sql
-- Database: auth_db
-- =============================================================================

CREATE TABLE IF NOT EXISTS login_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    email VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45) NOT NULL,
    user_agent TEXT,
    status VARCHAR(20) NOT NULL, -- 'SUCCESS', 'FAILED', 'LOCKED'
    failure_reason VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_login_audit_logs_user_id ON login_audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_login_audit_logs_email ON login_audit_logs(email);
CREATE INDEX IF NOT EXISTS idx_login_audit_logs_created_at ON login_audit_logs(created_at);
