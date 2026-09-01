-- =============================================================================
-- Migration: 001_create_assets_and_cmdb_table.sql
-- Database: asset_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Assets Table (Hardware, Equipment, Licenses)
CREATE TABLE IF NOT EXISTS assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    asset_tag VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    model VARCHAR(150),
    serial_number VARCHAR(100) UNIQUE,
    purchase_date DATE DEFAULT CURRENT_DATE,
    purchase_cost NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    warranty_expiry DATE,
    current_value NUMERIC(12, 2) NOT NULL DEFAULT 0.00,
    status VARCHAR(50) NOT NULL DEFAULT 'IN_STOCK',
    location VARCHAR(150) NOT NULL DEFAULT 'Headquarters Warehouse',
    assigned_to_user_id VARCHAR(100),
    assigned_to_user_name VARCHAR(255),
    assigned_at TIMESTAMP WITH TIME ZONE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_assets_tag ON assets(asset_tag);
CREATE INDEX IF NOT EXISTS idx_assets_category ON assets(category);
CREATE INDEX IF NOT EXISTS idx_assets_status ON assets(status);
CREATE INDEX IF NOT EXISTS idx_assets_assigned_user ON assets(assigned_to_user_id);

-- 2. Asset Assignment History Table
CREATE TABLE IF NOT EXISTS asset_assignments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    asset_id UUID NOT NULL REFERENCES assets(id) ON DELETE CASCADE,
    user_id VARCHAR(100) NOT NULL,
    user_name VARCHAR(255) NOT NULL,
    department_id VARCHAR(100),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    returned_at TIMESTAMP WITH TIME ZONE,
    condition_on_assign VARCHAR(100) NOT NULL DEFAULT 'EXCELLENT',
    condition_on_return VARCHAR(100),
    notes TEXT
);

CREATE INDEX IF NOT EXISTS idx_asset_assignments_asset ON asset_assignments(asset_id);
CREATE INDEX IF NOT EXISTS idx_asset_assignments_user ON asset_assignments(user_id);

-- 3. CMDB Configuration Items (CI)
CREATE TABLE IF NOT EXISTS configuration_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ci_code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    ci_type VARCHAR(50) NOT NULL,
    environment VARCHAR(50) NOT NULL DEFAULT 'PRODUCTION',
    owner_id VARCHAR(100),
    owner_name VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'OPERATIONAL',
    ip_address VARCHAR(100),
    asset_id UUID REFERENCES assets(id) ON DELETE SET NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ci_code ON configuration_items(ci_code);
CREATE INDEX IF NOT EXISTS idx_ci_type ON configuration_items(ci_type);
CREATE INDEX IF NOT EXISTS idx_ci_environment ON configuration_items(environment);
CREATE INDEX IF NOT EXISTS idx_ci_status ON configuration_items(status);

-- 4. CI Relationships & Dependency Topology
CREATE TABLE IF NOT EXISTS ci_relationships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    parent_ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    child_ci_id UUID NOT NULL REFERENCES configuration_items(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL,
    impact_weight VARCHAR(50) NOT NULL DEFAULT 'HIGH',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(parent_ci_id, child_ci_id, relationship_type)
);

CREATE INDEX IF NOT EXISTS idx_ci_rel_parent ON ci_relationships(parent_ci_id);
CREATE INDEX IF NOT EXISTS idx_ci_rel_child ON ci_relationships(child_ci_id);

-- Operational inventory and topology are loaded only by the explicit development seed command.
