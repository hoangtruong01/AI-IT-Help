ALTER TABLE knowledge_articles
    ADD COLUMN IF NOT EXISTS department_id VARCHAR(100);

CREATE INDEX IF NOT EXISTS idx_knowledge_articles_department
    ON knowledge_articles(department_id);
