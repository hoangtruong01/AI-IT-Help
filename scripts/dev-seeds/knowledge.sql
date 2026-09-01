-- Idempotent development-only operational data for knowledge_db.
INSERT INTO knowledge_articles (
    id, category_id, title, slug, summary, content, tags,
    author_id, author_name, department_id, is_published
) VALUES (
    '90000000-0000-0000-0000-000000000401',
    'c0000000-0000-0000-0000-000000000002',
    'Development VPN troubleshooting article', 'dev-vpn-troubleshooting',
    'Development-only knowledge fixture.',
    'Verify the local VPN profile and development gateway.',
    ARRAY['development', 'vpn'], 'dev-agent', 'Development Agent',
    'dev-department', TRUE
)
ON CONFLICT (slug) DO NOTHING;

INSERT INTO runbooks (
    id, code, title, category, description, prerequisites, steps_json,
    rollback_steps, author_name, is_active
) VALUES (
    '90000000-0000-0000-0000-000000000402', 'DEV-RB-0001',
    'Development service recovery runbook', 'Development',
    'Development-only runbook fixture.', ARRAY['Local development access'],
    '[{"order":1,"action":"Check the local service health endpoint"}]'::jsonb,
    ARRAY['Stop the local test process'], 'Development Agent', TRUE
)
ON CONFLICT (code) DO NOTHING;
