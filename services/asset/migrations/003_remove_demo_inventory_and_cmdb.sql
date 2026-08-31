-- Remove only the fixed portfolio inventory and fictional production topology.
-- Relationship rows and assignment history are removed by the declared cascades.
DELETE FROM configuration_items
WHERE (id, ci_code) IN (
    ('f0000000-0000-0000-0000-000000000001'::uuid, 'CI-APP-WEB'),
    ('f0000000-0000-0000-0000-000000000002'::uuid, 'CI-API-GATEWAY'),
    ('f0000000-0000-0000-0000-000000000003'::uuid, 'CI-SRV-PROD-K8S'),
    ('f0000000-0000-0000-0000-000000000004'::uuid, 'CI-DB-POSTGRES'),
    ('f0000000-0000-0000-0000-000000000005'::uuid, 'CI-MQ-RABBIT')
);

DELETE FROM assets
WHERE (id, asset_tag) IN (
    ('b0000000-0000-0000-0000-000000000001'::uuid, 'AST-1001'),
    ('b0000000-0000-0000-0000-000000000002'::uuid, 'AST-1002'),
    ('b0000000-0000-0000-0000-000000000003'::uuid, 'AST-1003'),
    ('b0000000-0000-0000-0000-000000000004'::uuid, 'AST-1004'),
    ('b0000000-0000-0000-0000-000000000005'::uuid, 'AST-1005')
);
