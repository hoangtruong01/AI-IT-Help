-- =============================================================================
-- Migration: 002_add_optimistic_locking_version.sql
-- Database: asset_db
-- =============================================================================

ALTER TABLE assets ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;
ALTER TABLE configuration_items ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

CREATE INDEX IF NOT EXISTS idx_assets_version ON assets(id, version);
CREATE INDEX IF NOT EXISTS idx_configuration_items_version ON configuration_items(id, version);
