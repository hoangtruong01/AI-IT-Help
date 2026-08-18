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

-- 4. Seed Initial Realtime Notifications
INSERT INTO notifications (
    id, recipient_id, recipient_email, title, message, category, priority, is_read, channel
) VALUES
    (
        'n0000000-0000-0000-0000-000000000001',
        'e0000000-0000-0000-0000-000000000001',
        'admin@eomp.local',
        'CAB Approval Required: PostgreSQL 17.2 Patch',
        'Instance WFI-1003 is awaiting your sign-off as IT Director.',
        'APPROVAL',
        'HIGH',
        FALSE,
        'IN_APP'
    ),
    (
        'n0000000-0000-0000-0000-000000000002',
        'e0000000-0000-0000-0000-000000000001',
        'admin@eomp.local',
        'Urgent Incident Raised: VPN Connection Failure',
        'Ticket TK-1094 raised by Emily Davis is currently in progress (Priority: URGENT).',
        'INCIDENT',
        'URGENT',
        FALSE,
        'IN_APP'
    ),
    (
        'n0000000-0000-0000-0000-000000000003',
        'e0000000-0000-0000-0000-000000000002',
        'manager@eomp.local',
        'Approval Request: MacBook Pro 16" Allocation',
        'Emily Davis requested MacBook Pro 16" M3 Max provisioning (Instance WFI-1001).',
        'APPROVAL',
        'MEDIUM',
        FALSE,
        'IN_APP'
    ),
    (
        'n0000000-0000-0000-0000-000000000004',
        'e0000000-0000-0000-0000-000000000003',
        'agent@eomp.local',
        'New Ticket Assigned: TK-1092',
        'Cannot access PostgreSQL Staging Cluster has been assigned to your support queue.',
        'INCIDENT',
        'HIGH',
        TRUE,
        'IN_APP'
    )
ON CONFLICT DO NOTHING;
