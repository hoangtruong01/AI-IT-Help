-- EOMP Reporting & BI Analytics Service schema.
-- Operational telemetry is intentionally not seeded in production migrations.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

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
CREATE INDEX IF NOT EXISTS idx_sla_metrics_date ON sla_metrics_daily(metric_date);

CREATE TABLE IF NOT EXISTS agent_performance (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    agent_id VARCHAR(100) NOT NULL,
    agent_name VARCHAR(150) NOT NULL,
    agent_avatar VARCHAR(255) DEFAULT '',
    job_title VARCHAR(100) DEFAULT 'IT Support Specialist',
    department VARCHAR(100) DEFAULT 'IT Operations',
    tickets_assigned INT NOT NULL DEFAULT 0,
    tickets_resolved INT NOT NULL DEFAULT 0,
    avg_mttr_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    sla_compliance_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    period VARCHAR(50) NOT NULL DEFAULT '',
    metric_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_agent_performance_period ON agent_performance(period);

CREATE TABLE IF NOT EXISTS category_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_name VARCHAR(100) NOT NULL,
    category_code VARCHAR(50) NOT NULL,
    icon VARCHAR(100) DEFAULT 'i-lucide-folder',
    total_count INT NOT NULL DEFAULT 0,
    resolved_count INT NOT NULL DEFAULT 0,
    avg_resolution_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    share_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    period VARCHAR(50) NOT NULL DEFAULT '',
    metric_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS department_sla_metrics (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    department_name VARCHAR(150) NOT NULL,
    department_code VARCHAR(50) NOT NULL,
    total_tickets INT NOT NULL DEFAULT 0,
    within_sla_count INT NOT NULL DEFAULT 0,
    breached_sla_count INT NOT NULL DEFAULT 0,
    sla_compliance_pct NUMERIC(5,2) NOT NULL DEFAULT 0.0,
    avg_mttr_minutes NUMERIC(8,2) NOT NULL DEFAULT 0.0,
    period VARCHAR(50) NOT NULL DEFAULT '',
    metric_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS raw_incident_records (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_number VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    priority VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL,
    requester_name VARCHAR(150) NOT NULL,
    assignee_name VARCHAR(150) NOT NULL,
    assignee_id VARCHAR(100),
    department VARCHAR(100) NOT NULL,
    mttd_minutes INT NOT NULL DEFAULT 0,
    mttr_minutes INT NOT NULL DEFAULT 0,
    sla_status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    resolved_at TIMESTAMP WITH TIME ZONE,
    source_event_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS reporting_processed_events (
    event_id VARCHAR(255) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_category_metrics_date_code ON category_metrics(metric_date, category_code);
CREATE UNIQUE INDEX IF NOT EXISTS uq_department_metrics_date_code ON department_sla_metrics(metric_date, department_code);
CREATE UNIQUE INDEX IF NOT EXISTS uq_agent_metrics_date_agent ON agent_performance(metric_date, agent_id);
CREATE UNIQUE INDEX IF NOT EXISTS uq_raw_incident_ticket ON raw_incident_records(ticket_number);
