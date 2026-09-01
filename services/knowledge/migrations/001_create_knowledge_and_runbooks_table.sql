-- =============================================================================
-- Migration: 001_create_knowledge_and_runbooks_table.sql
-- Module: Knowledge Service (knowledge_db)
-- Description: Knowledge base articles, categories, SOP runbooks, and vector embeddings
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Knowledge Categories Table
CREATE TABLE IF NOT EXISTS knowledge_categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    icon VARCHAR(50) NOT NULL DEFAULT 'i-lucide-folder',
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 2. Knowledge Base Articles Table
CREATE TABLE IF NOT EXISTS knowledge_articles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    category_id UUID NOT NULL REFERENCES knowledge_categories(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    summary TEXT NOT NULL,
    content TEXT NOT NULL,
    tags TEXT[] NOT NULL DEFAULT '{}',
    author_id VARCHAR(100) NOT NULL DEFAULT 'system',
    author_name VARCHAR(150) NOT NULL DEFAULT 'IT System Admin',
    department_id VARCHAR(100),
    view_count INT NOT NULL DEFAULT 0,
    helpful_count INT NOT NULL DEFAULT 0,
    is_published BOOLEAN NOT NULL DEFAULT TRUE,
    version INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 3. Standard Operating Procedures (SOP) Runbooks Table
CREATE TABLE IF NOT EXISTS runbooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    category VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    prerequisites TEXT NOT NULL,
    steps_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    rollback_steps TEXT NOT NULL,
    author_name VARCHAR(150) NOT NULL DEFAULT 'IT Security Team',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 4. Document Embeddings / Chunks Reference Table (Sync with Qdrant)
CREATE TABLE IF NOT EXISTS document_embeddings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    article_id UUID NOT NULL REFERENCES knowledge_articles(id) ON DELETE CASCADE,
    chunk_index INT NOT NULL DEFAULT 0,
    chunk_text TEXT NOT NULL,
    embedding_id VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 5. Fulltext & Performance Indexes
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_category ON knowledge_articles(category_id);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_published ON knowledge_articles(is_published);
CREATE INDEX IF NOT EXISTS idx_knowledge_articles_created_at ON knowledge_articles(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_runbooks_category ON runbooks(category);
CREATE INDEX IF NOT EXISTS idx_runbooks_code ON runbooks(code);
CREATE INDEX IF NOT EXISTS idx_document_embeddings_article ON document_embeddings(article_id);

-- =============================================================================
-- REFERENCE DATA: Knowledge categories
-- =============================================================================

INSERT INTO knowledge_categories (id, name, code, icon, description)
VALUES 
    ('c0000000-0000-0000-0000-000000000001', 'IT Security & Access', 'sec', 'i-lucide-shield-check', 'MFA tokens, SSO identity, zero-trust security & access management policies.'),
    ('c0000000-0000-0000-0000-000000000002', 'Network & Connectivity', 'net', 'i-lucide-network', 'WireGuard VPN, GlobalProtect, DNS routing, firewall rules & Wi-Fi setup.'),
    ('c0000000-0000-0000-0000-000000000003', 'Hardware & Equipment', 'hw', 'i-lucide-laptop', 'Laptop standard baseline, monitor setup, warranty claims & hardware replacement.'),
    ('c0000000-0000-0000-0000-000000000004', 'DevOps & Cloud Infrastructure', 'devops', 'i-lucide-server', 'PostgreSQL database access, Kubernetes clusters, MinIO storage & staging secrets.'),
    ('c0000000-0000-0000-0000-000000000005', 'Software & Productivity Apps', 'soft', 'i-lucide-layers', 'Operating systems, Slack, Microsoft 365, Docker Desktop & developer tooling.')
ON CONFLICT (code) DO NOTHING;

-- Operational articles and runbooks are loaded only by the explicit development seed command.
