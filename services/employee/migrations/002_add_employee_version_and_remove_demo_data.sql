-- Add optimistic-lock metadata to databases created before versioned writes.
ALTER TABLE employees
    ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- Remove only the exact development fixtures previously shipped by migration 001.
-- Matching both ID and email avoids deleting a real profile that merely reused an ID.
DELETE FROM employees
WHERE (id, email) IN (
    ('e0000000-0000-0000-0000-000000000001'::uuid, 'admin@eomp.local'),
    ('e0000000-0000-0000-0000-000000000002'::uuid, 'manager@eomp.local'),
    ('e0000000-0000-0000-0000-000000000003'::uuid, 'agent@eomp.local'),
    ('e0000000-0000-0000-0000-000000000004'::uuid, 'emily.davis@eomp.local')
);
