-- =============================================================================
-- Migration: 002_create_problems_table.sql
-- Database: helpdesk_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Problems Table (ITIL Problem Management & KEDB)
CREATE TABLE IF NOT EXISTS problems (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    problem_number VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL DEFAULT 'Infrastructure',
    priority VARCHAR(50) NOT NULL DEFAULT 'HIGH',
    status VARCHAR(50) NOT NULL DEFAULT 'OPEN',
    impact VARCHAR(50) NOT NULL DEFAULT 'HIGH',
    urgency VARCHAR(50) NOT NULL DEFAULT 'HIGH',
    assignee_id VARCHAR(100),
    assignee_name VARCHAR(255),
    root_cause TEXT,
    workaround TEXT,
    resolution TEXT,
    is_known_error BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP WITH TIME ZONE,
    closed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_problems_number ON problems(problem_number);
CREATE INDEX IF NOT EXISTS idx_problems_status ON problems(status);
CREATE INDEX IF NOT EXISTS idx_problems_category ON problems(category);
CREATE INDEX IF NOT EXISTS idx_problems_known_error ON problems(is_known_error);

-- 2. Problem - Incident Links Table (Many-to-One / Many-to-Many Aggregation)
CREATE TABLE IF NOT EXISTS problem_incident_links (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    problem_id UUID NOT NULL REFERENCES problems(id) ON DELETE CASCADE,
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    ticket_number VARCHAR(50) NOT NULL,
    ticket_title VARCHAR(255) NOT NULL,
    linked_by VARCHAR(255) NOT NULL DEFAULT 'IT Problem Manager',
    linked_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_problem_ticket UNIQUE (problem_id, ticket_id)
);

CREATE INDEX IF NOT EXISTS idx_prob_inc_problem ON problem_incident_links(problem_id);
CREATE INDEX IF NOT EXISTS idx_prob_inc_ticket ON problem_incident_links(ticket_id);

-- Operational problems and incident links are loaded only by the explicit development seed command.
