-- Upgrade audit records from an unkeyed checksum to an ordered HMAC-SHA256 chain.
-- Existing records remain labelled as legacy because their original payload cannot
-- be authenticated with the newly provisioned key.

ALTER TABLE audit_logs
    ADD COLUMN IF NOT EXISTS audit_sequence BIGSERIAL,
    ADD COLUMN IF NOT EXISTS previous_checksum VARCHAR(64) NOT NULL
        DEFAULT repeat('0', 64),
    ADD COLUMN IF NOT EXISTS checksum_algorithm VARCHAR(32) NOT NULL
        DEFAULT 'SHA-256-LEGACY';

ALTER TABLE audit_logs
    ALTER COLUMN actor_id TYPE VARCHAR(100) USING actor_id::text;

ALTER TABLE audit_logs
    ALTER COLUMN checksum_algorithm SET DEFAULT 'HMAC-SHA256';

CREATE OR REPLACE FUNCTION reject_audit_log_mutation()
RETURNS trigger
LANGUAGE plpgsql
AS $$
BEGIN
    RAISE EXCEPTION 'audit_logs is append-only; % is prohibited', TG_OP
        USING ERRCODE = '55000';
END;
$$;

DROP TRIGGER IF EXISTS audit_logs_append_only ON audit_logs;
CREATE TRIGGER audit_logs_append_only
BEFORE UPDATE OR DELETE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION reject_audit_log_mutation();

CREATE INDEX IF NOT EXISTS idx_audit_checksum_algorithm
    ON audit_logs (checksum_algorithm);

CREATE UNIQUE INDEX IF NOT EXISTS idx_audit_sequence
    ON audit_logs (audit_sequence);
