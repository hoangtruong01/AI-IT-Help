-- =============================================================================
-- Migration: 003_add_optimistic_locking_version.sql
-- Database: workflow_db
-- =============================================================================

ALTER TABLE workflow_instances ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;
ALTER TABLE change_requests ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_workflow_instances_version ON workflow_instances(id, version);
CREATE INDEX IF NOT EXISTS idx_change_requests_version ON change_requests(id, version);
