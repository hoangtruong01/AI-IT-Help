-- =============================================================================
-- Migration: 001_create_workflows_and_approvals_table.sql
-- Database: workflow_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Workflow Definitions Table
CREATE TABLE IF NOT EXISTS workflow_definitions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100) NOT NULL DEFAULT 'ITSM',
    trigger_type VARCHAR(50) NOT NULL DEFAULT 'SERVICE_REQUEST',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    steps_config JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_wf_def_code ON workflow_definitions(code);

-- 2. Workflow Instances Table
CREATE TABLE IF NOT EXISTS workflow_instances (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_number VARCHAR(50) NOT NULL UNIQUE,
    definition_id UUID NOT NULL REFERENCES workflow_definitions(id) ON DELETE CASCADE,
    definition_name VARCHAR(255) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    requester_id VARCHAR(100) NOT NULL,
    requester_name VARCHAR(255) NOT NULL,
    requester_email VARCHAR(255) NOT NULL,
    current_step_name VARCHAR(100) NOT NULL DEFAULT 'Step 1',
    status VARCHAR(50) NOT NULL DEFAULT 'RUNNING',
    context_data JSONB DEFAULT '{}'::jsonb,
    started_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_wf_inst_number ON workflow_instances(instance_number);
CREATE INDEX IF NOT EXISTS idx_wf_inst_status ON workflow_instances(status);
CREATE INDEX IF NOT EXISTS idx_wf_inst_requester ON workflow_instances(requester_id);

-- 3. Workflow Steps Table
CREATE TABLE IF NOT EXISTS workflow_steps (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_order INT NOT NULL DEFAULT 1,
    step_name VARCHAR(100) NOT NULL,
    step_type VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    assigned_to VARCHAR(255),
    started_at TIMESTAMP WITH TIME ZONE,
    completed_at TIMESTAMP WITH TIME ZONE,
    output_data JSONB
);

CREATE INDEX IF NOT EXISTS idx_wf_steps_inst ON workflow_steps(instance_id);

-- 4. Approval Requests Table
CREATE TABLE IF NOT EXISTS approval_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    step_id UUID REFERENCES workflow_steps(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    approver_id VARCHAR(100) NOT NULL,
    approver_name VARCHAR(255) NOT NULL,
    approver_role VARCHAR(50) NOT NULL DEFAULT 'ROLE_MANAGER',
    approval_level INT NOT NULL DEFAULT 1,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    decision_notes TEXT,
    decided_at TIMESTAMP WITH TIME ZONE,
    sla_deadline TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_approvals_inst ON approval_requests(instance_id);
CREATE INDEX IF NOT EXISTS idx_approvals_approver ON approval_requests(approver_id);
CREATE INDEX IF NOT EXISTS idx_approvals_status ON approval_requests(status);

-- 5. Workflow Logs & Audit Table
CREATE TABLE IF NOT EXISTS workflow_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    instance_id UUID NOT NULL REFERENCES workflow_instances(id) ON DELETE CASCADE,
    actor_id VARCHAR(100) NOT NULL,
    actor_name VARCHAR(255) NOT NULL,
    action VARCHAR(100) NOT NULL,
    message TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_wf_logs_inst ON workflow_logs(instance_id);

-- 6. Seed Workflow Definitions
INSERT INTO workflow_definitions (id, code, name, description, category, trigger_type, is_active, steps_config)
VALUES
    (
        'd0000000-0000-0000-0000-000000000001',
        'WF-LAPTOP-PROVISION',
        'New Employee Hardware Provisioning Workflow',
        'Multi-stage approval and fulfillment workflow for MacBook Pro / Workstation setup for onboarding.',
        'HARDWARE',
        'SERVICE_REQUEST',
        TRUE,
        '[{"order": 1, "name": "Department Manager Approval", "type": "APPROVAL", "role": "ROLE_MANAGER"}, {"order": 2, "name": "IT Asset Stock Allocation", "type": "AUTOMATED_ACTION"}, {"order": 3, "name": "IT Support Device Prep & Handover", "type": "APPROVAL", "role": "ROLE_AGENT"}]'::jsonb
    ),
    (
        'd0000000-0000-0000-0000-000000000002',
        'WF-VPN-ACCESS',
        'Remote VPC & Zero-Trust VPN Access Approval',
        'Requires Engineering Lead approval and automated WireGuard/OpenVPN certificate generation.',
        'SECURITY',
        'SERVICE_REQUEST',
        TRUE,
        '[{"order": 1, "name": "Direct Manager Approval", "type": "APPROVAL", "role": "ROLE_MANAGER"}, {"order": 2, "name": "Auto Certificate Generation", "type": "AUTOMATED_ACTION"}, {"order": 3, "name": "Security Audit Log", "type": "NOTIFICATION"}]'::jsonb
    ),
    (
        'd0000000-0000-0000-0000-000000000003',
        'WF-PROD-CHANGE',
        'Production Database & Infrastructure Change CAB',
        'Change Advisory Board (CAB) review for major cluster upgrades, schema migrations, and downtime windows.',
        'INFRASTRUCTURE',
        'MANUAL',
        TRUE,
        '[{"order": 1, "name": "IT Director Review", "type": "APPROVAL", "role": "ROLE_ADMIN"}, {"order": 2, "name": "Pre-flight Backup Verification", "type": "AUTOMATED_ACTION"}, {"order": 3, "name": "Maintenance Execution Window", "type": "APPROVAL", "role": "ROLE_ADMIN"}]'::jsonb
    )
ON CONFLICT (code) DO NOTHING;

-- 7. Seed Active Running Workflow Instances & Pending Approvals
INSERT INTO workflow_instances (
    id, instance_number, definition_id, definition_name, entity_type, entity_id,
    title, requester_id, requester_name, requester_email, current_step_name, status, started_at
) VALUES
    (
        'w0000000-0000-0000-0000-000000000001',
        'WFI-1001',
        'd0000000-0000-0000-0000-000000000001',
        'New Employee Hardware Provisioning Workflow',
        'SERVICE_REQUEST',
        's0000000-0000-0000-0000-000000000005',
        'MacBook Pro 16" Provisioning for Senior Backend Engineer',
        'e0000000-0000-0000-0000-000000000004',
        'Emily Davis',
        'emily.davis@eomp.local',
        'Department Manager Approval',
        'WAITING_APPROVAL',
        CURRENT_TIMESTAMP - INTERVAL '2 hours'
    ),
    (
        'w0000000-0000-0000-0000-000000000002',
        'WFI-1002',
        'd0000000-0000-0000-0000-000000000002',
        'Remote VPC & Zero-Trust VPN Access Approval',
        'SERVICE_REQUEST',
        's0000000-0000-0000-0000-000000000001',
        'Production Staging VPN Access Request',
        'e0000000-0000-0000-0000-000000000003',
        'Alex Nguyen',
        'agent@eomp.local',
        'Direct Manager Approval',
        'WAITING_APPROVAL',
        CURRENT_TIMESTAMP - INTERVAL '30 minutes'
    ),
    (
        'w0000000-0000-0000-0000-000000000003',
        'WFI-1003',
        'd0000000-0000-0000-0000-000000000003',
        'Production Database & Infrastructure Change CAB',
        'CHANGE',
        'CHG-2026-001',
        'PostgreSQL 17.2 Minor Security Patch & Vacuum Tuning',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        'admin@eomp.local',
        'IT Director Review',
        'WAITING_APPROVAL',
        CURRENT_TIMESTAMP - INTERVAL '1 hour'
    )
ON CONFLICT (instance_number) DO NOTHING;

-- 8. Seed Pending Approval Requests
INSERT INTO approval_requests (
    id, instance_id, title, approver_id, approver_name, approver_role, approval_level, status, sla_deadline
) VALUES
    (
        'a0000000-0000-0000-0000-000000000001',
        'w0000000-0000-0000-0000-000000000001',
        'Approve MacBook Pro 16" M3 Max Allocation ($3,499.00)',
        'e0000000-0000-0000-0000-000000000002',
        'David Tran (IT Manager)',
        'ROLE_MANAGER',
        1,
        'PENDING',
        CURRENT_TIMESTAMP + INTERVAL '6 hours'
    ),
    (
        'a0000000-0000-0000-0000-000000000002',
        'w0000000-0000-0000-0000-000000000002',
        'Approve WireGuard VPC Tunnel Certificate Issue',
        'e0000000-0000-0000-0000-000000000002',
        'David Tran (IT Manager)',
        'ROLE_MANAGER',
        1,
        'PENDING',
        CURRENT_TIMESTAMP + INTERVAL '4 hours'
    ),
    (
        'a0000000-0000-0000-0000-000000000003',
        'w0000000-0000-0000-0000-000000000003',
        'CAB Sign-off: PostgreSQL Security Patch Window',
        'e0000000-0000-0000-0000-000000000001',
        'System Administrator',
        'ROLE_ADMIN',
        1,
        'PENDING',
        CURRENT_TIMESTAMP + INTERVAL '2 hours'
    )
ON CONFLICT DO NOTHING;

-- 9. Seed Initial Audit Logs
INSERT INTO workflow_logs (instance_id, actor_id, actor_name, action, message)
VALUES
    ('w0000000-0000-0000-0000-000000000001', 'e0000000-0000-0000-0000-000000000004', 'Emily Davis', 'WORKFLOW_STARTED', 'Initiated Hardware Provisioning workflow instance WFI-1001.'),
    ('w0000000-0000-0000-0000-000000000001', 'system', 'Workflow Engine', 'APPROVAL_REQUESTED', 'Dispatched Level 1 approval request to IT Manager David Tran.'),
    ('w0000000-0000-0000-0000-000000000002', 'e0000000-0000-0000-0000-000000000003', 'Alex Nguyen', 'WORKFLOW_STARTED', 'Initiated VPN Access Approval workflow instance WFI-1002.'),
    ('w0000000-0000-0000-0000-000000000003', 'e0000000-0000-0000-0000-000000000001', 'System Administrator', 'WORKFLOW_STARTED', 'Initiated CAB Review workflow instance WFI-1003.')
ON CONFLICT DO NOTHING;
