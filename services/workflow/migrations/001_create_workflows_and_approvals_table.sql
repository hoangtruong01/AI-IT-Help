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
    department_id VARCHAR(100),
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
CREATE INDEX IF NOT EXISTS idx_wf_inst_department ON workflow_instances(department_id);

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

-- Operational workflow instances, approvals, and logs are loaded only by the explicit development seed command.
