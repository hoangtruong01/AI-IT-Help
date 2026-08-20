-- ============================================================
-- EOMP Reporting & BI Analytics Service - Migration 001
-- Tables: sla_metrics_daily, agent_performance, category_metrics, department_sla_metrics, raw_incident_records
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Daily SLA Metrics History
CREATE TABLE IF NOT EXISTS sla_metrics_daily (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    metric_date DATE NOT NULL UNIQUE,
    total_incidents INT NOT NULL DEFAULT 0,
    within_sla_count INT NOT NULL DEFAULT 0,
    breached_sla_count INT NOT NULL DEFAULT 0,
    avg_mttd_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    avg_mttr_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    sla_compliance_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sla_metrics_date ON sla_metrics_daily (metric_date);

-- 2. Agent Performance Scorecard
CREATE TABLE IF NOT EXISTS agent_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id UUID NOT NULL,
    agent_name VARCHAR(150) NOT NULL,
    agent_avatar VARCHAR(255) DEFAULT '',
    job_title VARCHAR(100) DEFAULT 'IT Support Specialist',
    department VARCHAR(100) DEFAULT 'IT Operations',
    tickets_assigned INT NOT NULL DEFAULT 0,
    tickets_resolved INT NOT NULL DEFAULT 0,
    avg_mttr_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    csat_rating NUMERIC(3,2) NOT NULL DEFAULT 5.0, -- Scale 1.0 to 5.0
    sla_compliance_pct NUMERIC(5,2) NOT NULL DEFAULT 100.0,
    period VARCHAR(50) NOT NULL DEFAULT '2026-08',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_agent_performance_period ON agent_performance (period);

-- 3. Category Metrics Breakdown
CREATE TABLE IF NOT EXISTS category_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_name VARCHAR(100) NOT NULL,
    category_code VARCHAR(50) NOT NULL,
    icon VARCHAR(100) DEFAULT 'i-lucide-folder',
    total_count INT NOT NULL DEFAULT 0,
    resolved_count INT NOT NULL DEFAULT 0,
    avg_resolution_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    share_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    period VARCHAR(50) NOT NULL DEFAULT '2026-08',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 4. Department SLA Metrics Breakdown
CREATE TABLE IF NOT EXISTS department_sla_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    department_name VARCHAR(150) NOT NULL,
    department_code VARCHAR(50) NOT NULL,
    total_tickets INT NOT NULL DEFAULT 0,
    within_sla_count INT NOT NULL DEFAULT 0,
    breached_sla_count INT NOT NULL DEFAULT 0,
    sla_compliance_pct NUMERIC(5,2) NOT NULL DEFAULT 100.0,
    avg_mttr_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    period VARCHAR(50) NOT NULL DEFAULT '2026-08',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 5. Raw Incident Records (for large-volume export benchmark & analytics)
CREATE TABLE IF NOT EXISTS raw_incident_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_number VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    priority VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    requester_name VARCHAR(150) NOT NULL,
    assignee_name VARCHAR(150) NOT NULL,
    department VARCHAR(100) NOT NULL,
    mttd_minutes INT NOT NULL DEFAULT 10,
    mttr_minutes INT NOT NULL DEFAULT 45,
    sla_status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE
);

-- ============================================================
-- SEED DATA (Realistic Operations Telemetry)
-- ============================================================

-- Daily SLA Trend (Past 14 Days)
INSERT INTO sla_metrics_daily (metric_date, total_incidents, within_sla_count, breached_sla_count, avg_mttd_minutes, avg_mttr_minutes, sla_compliance_pct)
VALUES 
    (CURRENT_DATE - INTERVAL '13 days', 42, 41, 1, 8.5, 38.2, 97.62),
    (CURRENT_DATE - INTERVAL '12 days', 38, 37, 1, 7.2, 34.0, 97.37),
    (CURRENT_DATE - INTERVAL '11 days', 45, 43, 2, 9.0, 41.5, 95.56),
    (CURRENT_DATE - INTERVAL '10 days', 50, 48, 2, 8.0, 36.8, 96.00),
    (CURRENT_DATE - INTERVAL '9 days', 52, 51, 1, 6.8, 32.4, 98.08),
    (CURRENT_DATE - INTERVAL '8 days', 28, 28, 0, 5.5, 26.0, 100.00),
    (CURRENT_DATE - INTERVAL '7 days', 24, 24, 0, 5.0, 24.5, 100.00),
    (CURRENT_DATE - INTERVAL '6 days', 48, 46, 2, 7.8, 35.0, 95.83),
    (CURRENT_DATE - INTERVAL '5 days', 54, 52, 2, 8.2, 37.5, 96.30),
    (CURRENT_DATE - INTERVAL '4 days', 60, 58, 2, 7.5, 33.2, 96.67),
    (CURRENT_DATE - INTERVAL '3 days', 49, 48, 1, 6.9, 31.0, 97.96),
    (CURRENT_DATE - INTERVAL '2 days', 55, 53, 2, 7.1, 32.8, 96.36),
    (CURRENT_DATE - INTERVAL '1 day', 41, 40, 1, 6.5, 29.4, 97.56),
    (CURRENT_DATE, 35, 34, 1, 6.2, 28.0, 97.14)
ON CONFLICT (metric_date) DO NOTHING;

-- Agent Performance Scorecard
INSERT INTO agent_performance (agent_id, agent_name, agent_avatar, job_title, department, tickets_assigned, tickets_resolved, avg_mttr_minutes, csat_rating, sla_compliance_pct, period)
VALUES
    ('u0000000-0000-0000-0000-000000000003', 'Marcus Vance', 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=150', 'Senior IT Ops Specialist', 'IT Support L2', 68, 66, 28.4, 4.92, 98.5, '2026-08'),
    ('u0000000-0000-0000-0000-000000000004', 'Sarah Jenkins', 'https://images.unsplash.com/photo-1580489944761-15a19d654956?w=150', 'Cybersecurity & IAM Engineer', 'IT Security', 54, 53, 31.2, 4.88, 98.1, '2026-08'),
    ('u0000000-0000-0000-0000-000000000005', 'David Kim', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=150', 'Systems & Cloud Administrator', 'DevOps & Infra', 48, 46, 35.8, 4.79, 95.8, '2026-08'),
    ('u0000000-0000-0000-0000-000000000006', 'Elena Rostova', 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=150', 'IT Support Specialist L1', 'IT Helpdesk L1', 72, 70, 24.6, 4.95, 97.2, '2026-08'),
    ('u0000000-0000-0000-0000-000000000007', 'Alex Chen', 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=150', 'Network Operations Specialist', 'Network Engineering', 42, 40, 42.0, 4.70, 95.2, '2026-08')
ON CONFLICT DO NOTHING;

-- Category Breakdown
INSERT INTO category_metrics (category_name, category_code, icon, total_count, resolved_count, avg_resolution_minutes, share_pct, period)
VALUES
    ('Network & Connectivity', 'network', 'i-lucide-wifi', 184, 178, 38.5, 32.5, '2026-08'),
    ('Account & Access (SSO/MFA)', 'security', 'i-lucide-shield-check', 142, 140, 22.0, 25.1, '2026-08'),
    ('Hardware & Peripherals', 'hardware', 'i-lucide-laptop', 115, 110, 45.2, 20.3, '2026-08'),
    ('Software & Applications', 'software', 'i-lucide-app-window', 82, 79, 34.0, 14.5, '2026-08'),
    ('Email & Collaboration', 'collaboration', 'i-lucide-mail', 43, 42, 18.4, 7.6, '2026-08')
ON CONFLICT DO NOTHING;

-- Department SLA Breakdown
INSERT INTO department_sla_metrics (department_name, department_code, total_tickets, within_sla_count, breached_sla_count, sla_compliance_pct, avg_mttr_minutes, period)
VALUES
    ('Software Engineering', 'ENG', 185, 181, 4, 97.84, 31.4, '2026-08'),
    ('Sales & Business Development', 'SALES', 124, 120, 4, 96.77, 28.5, '2026-08'),
    ('Human Resources', 'HR', 86, 84, 2, 97.67, 25.0, '2026-08'),
    ('Finance & Accounting', 'FIN', 92, 89, 3, 96.74, 33.2, '2026-08'),
    ('Marketing & Operations', 'MKT', 79, 78, 1, 98.73, 27.8, '2026-08')
ON CONFLICT DO NOTHING;

-- Seed Raw Incidents
INSERT INTO raw_incident_records (ticket_number, title, category, priority, status, requester_name, assignee_name, department, mttd_minutes, mttr_minutes, sla_status, created_at, resolved_at)
VALUES
    ('TK-1001', 'Cannot connect to Staging VPN Gateway', 'Network & Connectivity', 'HIGH', 'RESOLVED', 'Alex Morgan', 'Marcus Vance', 'Software Engineering', 5, 28, 'WITHIN_SLA', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '28 minutes'),
    ('TK-1002', 'Request MFA Token Reset in Okta', 'Account & Access (SSO/MFA)', 'MEDIUM', 'RESOLVED', 'Lisa Ray', 'Sarah Jenkins', 'Sales & Business Development', 4, 18, 'WITHIN_SLA', NOW() - INTERVAL '2 days', NOW() - INTERVAL '2 days' + INTERVAL '18 minutes'),
    ('TK-1003', 'MacBook Pro M3 External Monitor Flickering', 'Hardware & Peripherals', 'LOW', 'RESOLVED', 'Kenji Sato', 'Elena Rostova', 'Human Resources', 12, 42, 'WITHIN_SLA', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '42 minutes'),
    ('TK-1004', 'IntelliJ IDEA License Renewal Required', 'Software & Applications', 'MEDIUM', 'RESOLVED', 'Emily Davis', 'David Kim', 'Software Engineering', 8, 30, 'WITHIN_SLA', NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day' + INTERVAL '30 minutes'),
    ('TK-1005', 'Core Database Connection Pool Spike Alert', 'Network & Connectivity', 'URGENT', 'RESOLVED', 'Monitoring Alert', 'Alex Chen', 'DevOps & Infra', 2, 65, 'BREACHED', NOW() - INTERVAL '3 days', NOW() - INTERVAL '3 days' + INTERVAL '65 minutes')
ON CONFLICT DO NOTHING;
