-- Idempotent development-only operational data for helpdesk_db.
INSERT INTO tickets (
    id, ticket_number, title, description, category, priority, status,
    requester_id, requester_name, requester_email, department_id,
    sla_response_deadline, sla_resolution_deadline, sla_status
) VALUES (
    '90000000-0000-0000-0000-000000000201', 'DEV-TK-0001',
    'Development VPN connectivity check',
    'Fixture created by the explicit development seed command.',
    'Network & Access', 'MEDIUM', 'OPEN', 'dev-employee',
    'Development Employee', 'dev-employee@example.test', 'dev-department',
    CURRENT_TIMESTAMP + INTERVAL '4 hours',
    CURRENT_TIMESTAMP + INTERVAL '8 hours', 'WITHIN_SLA'
)
ON CONFLICT (ticket_number) DO NOTHING;

INSERT INTO problems (
    id, problem_number, title, description, category, priority, status,
    impact, urgency, assignee_id, assignee_name, is_known_error
) VALUES (
    '90000000-0000-0000-0000-000000000202', 'DEV-PRB-0001',
    'Development recurring VPN issue',
    'Fixture created by the explicit development seed command.',
    'Network & Access', 'MEDIUM', 'OPEN', 'MEDIUM', 'MEDIUM',
    'dev-agent', 'Development Agent', FALSE
)
ON CONFLICT (problem_number) DO NOTHING;

INSERT INTO problem_incident_links (
    id, problem_id, ticket_id, ticket_number, ticket_title, linked_by
)
SELECT
    '90000000-0000-0000-0000-000000000203', p.id, t.id,
    t.ticket_number, t.title, 'Development Seed'
FROM problems p
JOIN tickets t ON t.ticket_number = 'DEV-TK-0001'
WHERE p.problem_number = 'DEV-PRB-0001'
ON CONFLICT (problem_id, ticket_id) DO NOTHING;
