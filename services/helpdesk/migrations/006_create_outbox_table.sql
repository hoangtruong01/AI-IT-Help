-- =============================================================================
-- Migration 006: Create Outbox Events Table (Transactional Outbox Pattern)
-- =============================================================================
-- Enforces zero event loss between PostgreSQL and RabbitMQ. All domain events
-- are written in the same ACID transaction as the business mutation, then
-- delivered reliably by the outbox publisher worker.

CREATE TABLE IF NOT EXISTS outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type VARCHAR(100) NOT NULL,
    source VARCHAR(100) NOT NULL DEFAULT 'helpdesk',
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    retry_count INT NOT NULL DEFAULT 0,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_outbox_events_status_created 
ON outbox_events (status, created_at) 
WHERE status = 'PENDING';
