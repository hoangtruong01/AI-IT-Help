-- =============================================================================
-- Migration: 001_create_users_table.sql
-- Database: auth_db
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'ROLE_EMPLOYEE',
    department_id UUID,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_department ON users(department_id);

-- 2. Refresh Tokens Table
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_expires ON refresh_tokens(expires_at);

-- 3. Default Seed Data: Super Admin & Demo Accounts
-- Default password is: Admin@123456 (bcrypt hash: $2a$10$wK1F5N8q.mZ8GkPvh73ZRe7Z8Z/4p6O66a0V8k.C1kY4Z7Yq1s62u or generated on startup)
-- We insert using an ON CONFLICT clause for idempotency.
INSERT INTO users (id, email, password_hash, full_name, role, is_active)
VALUES (
    'a0000000-0000-0000-0000-000000000001',
    'admin@eomp.local',
    '$2a$10$4.qJbL6q05Nq3u9lP3a1AecNlXqV05mQhB54QJ.xP2C8QJ9g8sJgK', -- Admin@123456
    'System Administrator',
    'ROLE_ADMIN',
    TRUE
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (id, email, password_hash, full_name, role, is_active)
VALUES (
    'a0000000-0000-0000-0000-000000000002',
    'manager@eomp.local',
    '$2a$10$4.qJbL6q05Nq3u9lP3a1AecNlXqV05mQhB54QJ.xP2C8QJ9g8sJgK', -- Admin@123456
    'IT Operations Manager',
    'ROLE_MANAGER',
    TRUE
)
ON CONFLICT (email) DO NOTHING;

INSERT INTO users (id, email, password_hash, full_name, role, is_active)
VALUES (
    'a0000000-0000-0000-0000-000000000003',
    'agent@eomp.local',
    '$2a$10$4.qJbL6q05Nq3u9lP3a1AecNlXqV05mQhB54QJ.xP2C8QJ9g8sJgK', -- Admin@123456
    'IT Support L1 Lead',
    'ROLE_AGENT',
    TRUE
)
ON CONFLICT (email) DO NOTHING;
