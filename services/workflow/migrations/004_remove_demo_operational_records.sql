-- Keep reusable workflow definitions, but remove fictional live executions and RFCs.
DELETE FROM workflow_instances
WHERE (id, instance_number) IN (
    ('e1000000-0000-0000-0000-000000000001'::uuid, 'WFI-1001'),
    ('e1000000-0000-0000-0000-000000000002'::uuid, 'WFI-1002'),
    ('e1000000-0000-0000-0000-000000000003'::uuid, 'WFI-1003')
);

DELETE FROM change_requests
WHERE (id, change_number) IN (
    ('e0000000-0000-0000-0000-000000000001'::uuid, 'CHG-2001'),
    ('e0000000-0000-0000-0000-000000000002'::uuid, 'CHG-2002'),
    ('e0000000-0000-0000-0000-000000000003'::uuid, 'CHG-2003'),
    ('e0000000-0000-0000-0000-000000000004'::uuid, 'CHG-2004')
);
