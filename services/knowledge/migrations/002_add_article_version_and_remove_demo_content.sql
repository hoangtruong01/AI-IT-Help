ALTER TABLE knowledge_articles
    ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Remove the fictional operational content that older installations received.
-- Categories are retained because they are harmless taxonomy/reference data.
DELETE FROM knowledge_articles
WHERE (id, slug) IN (
    ('a0000000-0000-0000-0000-000000000001'::uuid, 'how-to-reset-user-mfa-tokens'),
    ('a0000000-0000-0000-0000-000000000002'::uuid, 'vpn-troubleshooting-guide'),
    ('a0000000-0000-0000-0000-000000000003'::uuid, 'standard-laptop-setup-baseline'),
    ('a0000000-0000-0000-0000-000000000004'::uuid, 'postgres-connection-pool-recovery'),
    ('a0000000-0000-0000-0000-000000000005'::uuid, 'hardware-replacement-warranty-workflow'),
    ('a0000000-0000-0000-0000-000000000006'::uuid, 'dns-timeout-subnet-routing')
);

DELETE FROM runbooks
WHERE (id, code) IN (
    ('b0000000-0000-0000-0000-000000000001'::uuid, 'RB-SEC-02'),
    ('b0000000-0000-0000-0000-000000000002'::uuid, 'RB-NET-01'),
    ('b0000000-0000-0000-0000-000000000003'::uuid, 'RB-DB-03'),
    ('b0000000-0000-0000-0000-000000000004'::uuid, 'RB-HW-04')
);
