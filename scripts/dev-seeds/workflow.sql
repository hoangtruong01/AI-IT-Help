-- Idempotent development-only operational data for workflow_db.
INSERT INTO workflow_instances (
    id, instance_number, definition_id, definition_name, entity_type,
    entity_id, title, requester_id, requester_name, requester_email,
    department_id, current_step_name, status, context_data
) VALUES (
    '90000000-0000-0000-0000-000000000301', 'DEV-WFI-0001',
    'd0000000-0000-0000-0000-000000000001',
    'New Employee Hardware Provisioning Workflow', 'SERVICE_REQUEST',
    'DEV-TK-0001', 'Development laptop approval', 'dev-employee',
    'Development Employee', 'dev-employee@example.test', 'dev-department',
    'Department Manager Approval', 'WAITING_APPROVAL',
    '{"source":"explicit-development-seed"}'::jsonb
)
ON CONFLICT (instance_number) DO NOTHING;

INSERT INTO approval_requests (
    id, instance_id, title, approver_id, approver_name, approver_role,
    approval_level, status, sla_deadline
)
SELECT
    '90000000-0000-0000-0000-000000000302', wi.id,
    'Approve development laptop request', 'ROLE_MANAGER',
    'Development Manager Pool', 'ROLE_MANAGER', 1, 'PENDING',
    CURRENT_TIMESTAMP + INTERVAL '1 day'
FROM workflow_instances wi
WHERE wi.instance_number = 'DEV-WFI-0001'
ON CONFLICT (id) DO NOTHING;
