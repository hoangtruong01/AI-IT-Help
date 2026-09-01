-- =============================================================================
-- Migration: 002_create_changes_and_cab_table.sql
-- Database: workflow_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Change Requests Table (ITIL RFC - Request for Change)
CREATE TABLE IF NOT EXISTS change_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    change_number VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    change_type VARCHAR(50) NOT NULL DEFAULT 'NORMAL', -- STANDARD, NORMAL, EMERGENCY, MAJOR
    category VARCHAR(100) NOT NULL DEFAULT 'Infrastructure',
    priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    risk_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, CRITICAL
    impact_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    probability_level VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT', -- DRAFT, SUBMITTED, CAB_REVIEW, APPROVED, REJECTED, SCHEDULED, IMPLEMENTING, COMPLETED, FAILED, CANCELLED
    requester_id VARCHAR(100) NOT NULL,
    requester_name VARCHAR(255) NOT NULL,
    requester_email VARCHAR(255) NOT NULL,
    assigned_to_id VARCHAR(100),
    assigned_to_name VARCHAR(255),
    reason_for_change TEXT NOT NULL,
    implementation_plan TEXT NOT NULL,
    rollback_plan TEXT NOT NULL,
    test_plan TEXT NOT NULL,
    scheduled_start_time TIMESTAMP WITH TIME ZONE,
    scheduled_end_time TIMESTAMP WITH TIME ZONE,
    actual_start_time TIMESTAMP WITH TIME ZONE,
    actual_end_time TIMESTAMP WITH TIME ZONE,
    downtime_required BOOLEAN NOT NULL DEFAULT FALSE,
    downtime_minutes INT NOT NULL DEFAULT 0,
    cab_required_count INT NOT NULL DEFAULT 2, -- Default 2 for EMERGENCY/MAJOR, 1 for NORMAL, 0 for STANDARD
    cab_approved_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_changes_number ON change_requests(change_number);
CREATE INDEX IF NOT EXISTS idx_changes_status ON change_requests(status);
CREATE INDEX IF NOT EXISTS idx_changes_type ON change_requests(change_type);
CREATE INDEX IF NOT EXISTS idx_changes_risk ON change_requests(risk_level);
CREATE INDEX IF NOT EXISTS idx_changes_scheduled ON change_requests(scheduled_start_time, scheduled_end_time);

-- 2. CAB Reviews & Voting Table
CREATE TABLE IF NOT EXISTS cab_reviews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    change_id UUID NOT NULL REFERENCES change_requests(id) ON DELETE CASCADE,
    reviewer_id VARCHAR(100) NOT NULL,
    reviewer_name VARCHAR(255) NOT NULL,
    reviewer_role VARCHAR(100) NOT NULL DEFAULT 'CAB Member',
    vote VARCHAR(50) NOT NULL, -- APPROVED, REJECTED, ABSTAIN
    comments TEXT,
    reviewed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_change_reviewer UNIQUE (change_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_cab_reviews_change ON cab_reviews(change_id);
CREATE INDEX IF NOT EXISTS idx_cab_reviews_reviewer ON cab_reviews(reviewer_id);

-- Operational change requests and CAB reviews are loaded only by the explicit development seed command.
