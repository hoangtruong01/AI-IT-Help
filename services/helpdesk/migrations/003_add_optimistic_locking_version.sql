-- =============================================================================
-- Migration: 003_add_optimistic_locking_version.sql
-- Database: helpdesk_db
-- =============================================================================

ALTER TABLE tickets ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;
ALTER TABLE problems ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_tickets_version ON tickets(id, version);
CREATE INDEX IF NOT EXISTS idx_problems_version ON problems(id, version);
