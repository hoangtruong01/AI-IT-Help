-- Idempotent development-only operational data for notification_db.
INSERT INTO notifications (
    id, recipient_id, recipient_email, title, message, category,
    priority, channel, metadata
) VALUES (
    '90000000-0000-0000-0000-000000000501', 'dev-employee',
    'dev-employee@example.test', 'Development environment ready',
    'Fixture created by the explicit development seed command.',
    'SYSTEM', 'LOW', 'IN_APP', '{"source":"explicit-development-seed"}'::jsonb
)
ON CONFLICT (id) DO NOTHING;
