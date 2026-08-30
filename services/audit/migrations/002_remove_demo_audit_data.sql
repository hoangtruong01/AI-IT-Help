-- Remove records inserted solely as historical UI/demo fixtures.
DELETE FROM audit_logs
WHERE id IN (
    'a0000000-0000-0000-0000-000000000001',
    'a0000000-0000-0000-0000-000000000002',
    'a0000000-0000-0000-0000-000000000003',
    'a0000000-0000-0000-0000-000000000004',
    'a0000000-0000-0000-0000-000000000005'
);

DELETE FROM security_events
WHERE (event_code, source_ip, description) IN (
    ('RBAC_VIOLATION_BLOCKED', '192.168.2.110', 'Unauthorized employee account attempted to view administrative audit records'),
    ('RATE_LIMIT_EXCEEDED', '10.0.4.55', 'Exceeded 5 failed login attempts in 1 minute window -> IP blocked for 15 mins'),
    ('DATA_MASKING_APPLIED', '192.168.1.10', 'Sanitized sensitive passwords and JWT bearer tokens from trace log output')
);
