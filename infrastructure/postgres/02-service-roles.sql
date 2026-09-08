-- =============================================================================
-- EOMP Per-Service PostgreSQL Least Privilege Credentials & Database Isolation
-- =============================================================================
-- This script enforces the Security Least Privilege boundary (DEVOPS-01).
-- Each microservice receives its dedicated DB role. Default connection from
-- PUBLIC is revoked, preventing any service credential from accessing another
-- service's bounded context database.

-- 1. Create dedicated database user roles with login permissions
DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'auth_svc') THEN
    CREATE ROLE auth_svc WITH LOGIN PASSWORD 'auth_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'employee_svc') THEN
    CREATE ROLE employee_svc WITH LOGIN PASSWORD 'employee_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'asset_svc') THEN
    CREATE ROLE asset_svc WITH LOGIN PASSWORD 'asset_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'helpdesk_svc') THEN
    CREATE ROLE helpdesk_svc WITH LOGIN PASSWORD 'helpdesk_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'workflow_svc') THEN
    CREATE ROLE workflow_svc WITH LOGIN PASSWORD 'workflow_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'knowledge_svc') THEN
    CREATE ROLE knowledge_svc WITH LOGIN PASSWORD 'knowledge_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'notification_svc') THEN
    CREATE ROLE notification_svc WITH LOGIN PASSWORD 'notification_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'audit_svc') THEN
    CREATE ROLE audit_svc WITH LOGIN PASSWORD 'audit_svc_secret';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'reporting_svc') THEN
    CREATE ROLE reporting_svc WITH LOGIN PASSWORD 'reporting_svc_pass';
  END IF;
END $$;

-- 2. Revoke default connection privileges from PUBLIC to enforce isolation
REVOKE CONNECT ON DATABASE auth_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE employee_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE asset_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE helpdesk_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE workflow_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE knowledge_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE notification_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE audit_db FROM PUBLIC;
REVOKE CONNECT ON DATABASE reporting_db FROM PUBLIC;

-- 3. Grant access strictly to the dedicated service role
GRANT ALL PRIVILEGES ON DATABASE auth_db TO auth_svc;
GRANT ALL PRIVILEGES ON DATABASE employee_db TO employee_svc;
GRANT ALL PRIVILEGES ON DATABASE asset_db TO asset_svc;
GRANT ALL PRIVILEGES ON DATABASE helpdesk_db TO helpdesk_svc;
GRANT ALL PRIVILEGES ON DATABASE workflow_db TO workflow_svc;
GRANT ALL PRIVILEGES ON DATABASE knowledge_db TO knowledge_svc;
GRANT ALL PRIVILEGES ON DATABASE notification_db TO notification_svc;
GRANT ALL PRIVILEGES ON DATABASE audit_db TO audit_svc;
GRANT ALL PRIVILEGES ON DATABASE reporting_db TO reporting_svc;
