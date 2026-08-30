CREATE TABLE IF NOT EXISTS notification_reads (
    notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
    recipient_id VARCHAR(100) NOT NULL,
    read_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (notification_id, recipient_id)
);

CREATE INDEX IF NOT EXISTS idx_notification_reads_recipient
    ON notification_reads(recipient_id, read_at DESC);

-- Preserve read state for recipient-specific records created before read receipts.
INSERT INTO notification_reads (notification_id, recipient_id, read_at)
SELECT id, recipient_id, COALESCE(read_at, CURRENT_TIMESTAMP)
FROM notifications
WHERE is_read = TRUE AND recipient_id <> 'all'
ON CONFLICT (notification_id, recipient_id) DO NOTHING;

-- Remove only the historical development fixtures.
DELETE FROM notifications
WHERE (id, recipient_email) IN (
    ('f0000000-0000-0000-0000-000000000001'::uuid, 'admin@eomp.local'),
    ('f0000000-0000-0000-0000-000000000002'::uuid, 'admin@eomp.local'),
    ('f0000000-0000-0000-0000-000000000003'::uuid, 'manager@eomp.local'),
    ('f0000000-0000-0000-0000-000000000004'::uuid, 'agent@eomp.local')
);
