-- =============================================================================
-- Migration: 001_create_notifications_table.sql
-- Database: notification_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Notifications Table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    recipient_id VARCHAR(100) NOT NULL,
    recipient_email VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'SYSTEM',
    priority VARCHAR(50) NOT NULL DEFAULT 'MEDIUM',
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    read_at TIMESTAMP WITH TIME ZONE,
    channel VARCHAR(50) NOT NULL DEFAULT 'IN_APP',
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient ON notifications(recipient_id);
CREATE INDEX IF NOT EXISTS idx_notifications_is_read ON notifications(is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications(created_at DESC);

-- 2. Notification Templates Table
CREATE TABLE IF NOT EXISTS notification_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type VARCHAR(100) NOT NULL UNIQUE,
    channel VARCHAR(50) NOT NULL DEFAULT 'IN_APP',
    subject_template VARCHAR(255) NOT NULL,
    body_template TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Seed Initial Notification Templates
INSERT INTO notification_templates (event_type, channel, subject_template, body_template)
VALUES
    ('ticket.created', 'IN_APP', 'New Ticket Raised: {{ticket_number}}', 'Ticket "{{title}}" was created by {{requester_name}} with priority {{priority}}.'),
    ('ticket.sla_warning', 'IN_APP', 'SLA Warning: {{ticket_number}} is approaching deadline', 'Ticket "{{title}}" has less than 20% SLA resolution time remaining.'),
    ('approval.requested', 'IN_APP', 'Action Required: Pending Approval for {{instance_number}}', 'You have a pending approval request for "{{title}}".'),
    ('asset.assigned', 'IN_APP', 'Asset Handover Confirmed: {{asset_tag}}', 'Asset {{name}} has been assigned to {{user_name}}.')
ON CONFLICT (event_type) DO NOTHING;

-- Operational notifications are created from runtime events or the explicit development seed command.
