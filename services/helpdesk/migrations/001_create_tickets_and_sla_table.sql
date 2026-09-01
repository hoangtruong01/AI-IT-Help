-- =============================================================================
-- Migration: 001_create_tickets_and_sla_table.sql
-- Database: helpdesk_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Service Categories Table
CREATE TABLE IF NOT EXISTS service_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL UNIQUE,
    icon VARCHAR(100) DEFAULT 'i-lucide-folder',
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Service Catalog Items Table
CREATE TABLE IF NOT EXISTS service_catalog_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID REFERENCES service_categories(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    default_priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    sla_response_minutes INT NOT NULL DEFAULT 240,
    sla_resolution_minutes INT NOT NULL DEFAULT 480,
    requires_approval BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Tickets (Incidents & Service Requests)
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_number VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    service_item_id UUID REFERENCES service_catalog_items(id) ON DELETE SET NULL,
    category VARCHAR(100) NOT NULL,
    priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    requester_id VARCHAR(100) NOT NULL,
    requester_name VARCHAR(255) NOT NULL,
    requester_email VARCHAR(255) NOT NULL,
    assignee_id VARCHAR(100),
    assignee_name VARCHAR(255),
    department_id VARCHAR(100),
    affected_ci_id VARCHAR(100),
    sla_response_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    sla_resolution_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    responded_at TIMESTAMP WITH TIME ZONE,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE,
    sla_status VARCHAR(50) NOT NULL DEFAULT 'WITHIN_SLA',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tickets_number ON tickets(ticket_number);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_priority ON tickets(priority);
CREATE INDEX IF NOT EXISTS idx_tickets_requester ON tickets(requester_id);
CREATE INDEX IF NOT EXISTS idx_tickets_assignee ON tickets(assignee_id);
CREATE INDEX IF NOT EXISTS idx_tickets_sla_resolution ON tickets(sla_resolution_deadline);

-- 4. Ticket Comments Table
CREATE TABLE IF NOT EXISTS ticket_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    author_id VARCHAR(100) NOT NULL,
    author_name VARCHAR(255) NOT NULL,
    author_role VARCHAR(50) NOT NULL,
    content TEXT NOT NULL,
    is_internal BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_comments_ticket ON ticket_comments(ticket_id);

-- 5. Ticket Audit Timeline Table
CREATE TABLE IF NOT EXISTS ticket_timeline (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    actor_id VARCHAR(100) NOT NULL,
    actor_name VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    old_value VARCHAR(255),
    new_value VARCHAR(255),
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_timeline_ticket ON ticket_timeline(ticket_id);

-- 6. Seed Initial Service Catalog & Items
INSERT INTO service_categories (id, name, icon, description)
VALUES
    ('c0000000-0000-0000-0000-000000000001', 'Network & Access', 'i-lucide-wifi', 'VPN, WiFi, Firewalls & Network issues'),
    ('c0000000-0000-0000-0000-000000000002', 'Hardware & Equipment', 'i-lucide-laptop', 'Laptop, Monitors, Peripherals & Repairs'),
    ('c0000000-0000-0000-0000-000000000003', 'Software & Licenses', 'i-lucide-app-window', 'OS, Microsoft 365, IDE & Tool licenses'),
    ('c0000000-0000-0000-0000-000000000004', 'Cloud & DevOps', 'i-lucide-server', 'Cloud access, CI/CD, Database & Clusters'),
    ('c0000000-0000-0000-0000-000000000005', 'Workplace IT', 'i-lucide-printer', 'Printers, Meeting rooms, TV & Projectors')
ON CONFLICT (name) DO NOTHING;

INSERT INTO service_catalog_items (id, category_id, name, code, description, default_priority, sla_response_minutes, sla_resolution_minutes, requires_approval)
VALUES
    ('f1000000-0000-0000-0000-000000000001', 'c0000000-0000-0000-0000-000000000001', 'VPN Connection Failure', 'NET_VPN_FAIL', 'Remote VPN connection drops or handshake failure', 'URGENT', 15, 120, FALSE),
    ('f1000000-0000-0000-0000-000000000002', 'c0000000-0000-0000-0000-000000000002', 'Request Dual Monitor & Dock', 'HW_DUAL_MON', 'Request secondary display setup for workstation', 'MEDIUM', 240, 480, TRUE),
    ('f1000000-0000-0000-0000-000000000003', 'c0000000-0000-0000-0000-000000000004', 'PostgreSQL Staging Access Issue', 'DEV_DB_ACCESS', 'Access denied to database cluster or staging replication', 'HIGH', 30, 240, FALSE),
    ('f1000000-0000-0000-0000-000000000004', 'c0000000-0000-0000-0000-000000000003', 'Microsoft 365 License & 2FA Reset', 'SW_M365_2FA', 'MFA token reset or Office 365 subscription issue', 'MEDIUM', 240, 480, FALSE),
    ('f1000000-0000-0000-0000-000000000005', 'c0000000-0000-0000-0000-000000000002', 'New Employee Laptop Provisioning', 'HW_PROVISION_LAPTOP', 'Setup MacBook Pro / ThinkPad for onboarding team member', 'HIGH', 30, 240, TRUE),
    ('f1000000-0000-0000-0000-000000000006', 'c0000000-0000-0000-0000-000000000005', 'Office Printer Offline', 'WP_PRINTER_OFFLINE', 'Floor printer paper jam or offline from print server', 'LOW', 480, 1440, FALSE)
ON CONFLICT (code) DO NOTHING;

-- Operational tickets are loaded only by the explicit development seed command.
