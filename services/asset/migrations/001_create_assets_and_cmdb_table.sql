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

-- 5. Seed Initial Asset Inventory
INSERT INTO assets (
    id, asset_tag, name, category, model, serial_number,
    purchase_date, purchase_cost, warranty_expiry, current_value,
    status, location, assigned_to_user_id, assigned_to_user_name, assigned_at, notes
) VALUES
    (
        'b0000000-0000-0000-0000-000000000001',
        'AST-1001',
        'MacBook Pro 16" M3 Max',
        'LAPTOP',
        'Apple Silicon M3 Max (36GB RAM / 1TB SSD)',
        'C02G81YQMD6R',
        '2025-01-15',
        3499.00,
        '2028-01-15',
        3150.00,
        'IN_USE',
        'Engineering Dept (Desk 4A)',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        CURRENT_TIMESTAMP - INTERVAL '30 days',
        'Primary development workstation for Principal Architect.'
    ),
    (
        'b0000000-0000-0000-0000-000000000002',
        'AST-1002',
        'Dell XPS 15 9530',
        'LAPTOP',
        'Intel Core i9-13900H / 32GB RAM / RTX 4070',
        'DL-99382109',
        '2025-02-10',
        2299.00,
        '2028-02-10',
        2050.00,
        'IN_USE',
        'IT Operations Office',
        'e0000000-0000-0000-0000-000000000002',
        'David Tran (IT Manager)',
        CURRENT_TIMESTAMP - INTERVAL '20 days',
        'Operations management workstation.'
    ),
    (
        'b0000000-0000-0000-0000-000000000003',
        'AST-1003',
        'Dell PowerEdge R750 Rack Server',
        'SERVER',
        'Dual Intel Xeon Gold 6330 / 256GB ECC RAM / 8x 3.84TB NVMe',
        'SV-R750-00492',
        '2024-06-01',
        12500.00,
        '2029-06-01',
        10200.00,
        'IN_USE',
        'Data Center Rack 04 (IDC Tan Thuan)',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        CURRENT_TIMESTAMP - INTERVAL '60 days',
        'Production Kubernetes Node 01.'
    ),
    (
        'b0000000-0000-0000-0000-000000000004',
        'AST-1004',
        'Cisco Catalyst 9300 48-Port PoE Switch',
        'NETWORK',
        'C9300-48P-A 10G Uplink',
        'CSCO-9300-4819',
        '2024-03-15',
        4800.00,
        '2029-03-15',
        4100.00,
        'IN_USE',
        'Server Room Floor 4',
        'e0000000-0000-0000-0000-000000000003',
        'Alex Nguyen (IT Support)',
        CURRENT_TIMESTAMP - INTERVAL '90 days',
        'Core distribution switch for Building A.'
    ),
    (
        'b0000000-0000-0000-0000-000000000005',
        'AST-1005',
        'Dell UltraSharp 27" 4K USB-C Monitor',
        'MONITOR',
        'U2723QE IPS Black',
        'CN-0M3819-742',
        '2025-03-01',
        599.00,
        '2028-03-01',
        540.00,
        'IN_STOCK',
        'Headquarters Warehouse',
        NULL,
        NULL,
        NULL,
        'Ready for deployment to new hires.'
    )
ON CONFLICT (asset_tag) DO NOTHING;

-- 6. Seed Initial CMDB Configuration Items
INSERT INTO configuration_items (
    id, ci_code, name, ci_type, environment, owner_id, owner_name, status, ip_address, asset_id, description
) VALUES
    (
        'f0000000-0000-0000-0000-000000000001',
        'CI-APP-WEB',
        'EOMP Web Portal Frontend',
        'APPLICATION',
        'PRODUCTION',
        'e0000000-0000-0000-0000-000000000004',
        'Emily Davis',
        'OPERATIONAL',
        '10.0.1.50',
        NULL,
        'Nuxt 4 / Vue 3 Operations Dashboard Single Page Application.'
    ),
    (
        'f0000000-0000-0000-0000-000000000002',
        'CI-API-GATEWAY',
        'EOMP API Gateway Service',
        'API_SERVICE',
        'PRODUCTION',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        'OPERATIONAL',
        '10.0.1.10',
        NULL,
        'Go Reverse Proxy Router, JWT Auth Verifier, and Rate Limiter.'
    ),
    (
        'f0000000-0000-0000-0000-000000000003',
        'CI-SRV-PROD-K8S',
        'IDC Kubernetes Worker Node 01',
        'SERVER',
        'PRODUCTION',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        'OPERATIONAL',
        '192.168.10.101',
        'b0000000-0000-0000-0000-000000000003',
        'Dell PowerEdge R750 Bare Metal Host running Microservices containers.'
    ),
    (
        'f0000000-0000-0000-0000-000000000004',
        'CI-DB-POSTGRES',
        'PostgreSQL 17 Primary Database Cluster',
        'DATABASE',
        'PRODUCTION',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        'OPERATIONAL',
        '10.0.2.10',
        NULL,
        'Primary PostgreSQL cluster hosting 7 isolated microservice databases.'
    ),
    (
        'f0000000-0000-0000-0000-000000000005',
        'CI-MQ-RABBIT',
        'RabbitMQ Event Broker Cluster',
        'CLOUD_RESOURCE',
        'PRODUCTION',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        'OPERATIONAL',
        '10.0.2.20',
        NULL,
        'AMQP Event bus for asynchronous messaging and outbox publisher.'
    )
ON CONFLICT (ci_code) DO NOTHING;

-- 7. Seed Initial CMDB Relationships (Topology Dependency Map)
INSERT INTO ci_relationships (parent_ci_id, child_ci_id, relationship_type, impact_weight)
VALUES
    ('f0000000-0000-0000-0000-000000000001', 'f0000000-0000-0000-0000-000000000002', 'DEPENDS_ON', 'CRITICAL'),
    ('f0000000-0000-0000-0000-000000000002', 'f0000000-0000-0000-0000-000000000003', 'RUNS_ON', 'CRITICAL'),
    ('f0000000-0000-0000-0000-000000000002', 'f0000000-0000-0000-0000-000000000004', 'CONNECTS_TO', 'CRITICAL'),
    ('f0000000-0000-0000-0000-000000000002', 'f0000000-0000-0000-0000-000000000005', 'CONNECTS_TO', 'HIGH')
ON CONFLICT DO NOTHING;
