-- Security-sensitive user lifecycle events. These rows are written in the same
-- transaction as the account/password change they describe.
CREATE TABLE IF NOT EXISTS security_audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    actor_id UUID REFERENCES users(id) ON DELETE SET NULL,
    actor_email VARCHAR(255) NOT NULL,
    action VARCHAR(64) NOT NULL,
    target_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_security_audit_actor ON security_audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_security_audit_target ON security_audit_logs(target_user_id);
CREATE INDEX IF NOT EXISTS idx_security_audit_created ON security_audit_logs(created_at);
