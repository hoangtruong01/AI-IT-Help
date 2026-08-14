-- =============================================================================
-- EOMP Microservices Databases Initialization
-- =============================================================================
-- Each bounded context microservice owns its dedicated database.
-- Cross-service access to these databases is strictly prohibited by architecture rules.

SELECT 'CREATE DATABASE auth_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth_db')\gexec

SELECT 'CREATE DATABASE employee_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'employee_db')\gexec

SELECT 'CREATE DATABASE asset_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'asset_db')\gexec

SELECT 'CREATE DATABASE helpdesk_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'helpdesk_db')\gexec

SELECT 'CREATE DATABASE workflow_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'workflow_db')\gexec

SELECT 'CREATE DATABASE knowledge_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'knowledge_db')\gexec

SELECT 'CREATE DATABASE audit_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'audit_db')\gexec
