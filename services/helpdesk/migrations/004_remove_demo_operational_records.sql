-- Remove fictional incidents/problems from historical development migrations.
DELETE FROM problems
WHERE (id, problem_number) IN (
    ('b0000000-0000-0000-0000-000000000001'::uuid, 'PRB-1001'),
    ('b0000000-0000-0000-0000-000000000002'::uuid, 'PRB-1002'),
    ('b0000000-0000-0000-0000-000000000003'::uuid, 'PRB-1003')
);

DELETE FROM tickets
WHERE (id, ticket_number) IN (
    ('e2000000-0000-0000-0000-000000000001'::uuid, 'TK-1094'),
    ('e2000000-0000-0000-0000-000000000002'::uuid, 'TK-1093'),
    ('e2000000-0000-0000-0000-000000000003'::uuid, 'TK-1092'),
    ('e2000000-0000-0000-0000-000000000004'::uuid, 'TK-1091')
);
