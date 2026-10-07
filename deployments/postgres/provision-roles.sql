\set ON_ERROR_STOP on
\getenv migrator_password ERP_MIGRATOR_PASSWORD
\getenv app_password ERP_APP_PASSWORD
\getenv backup_password ERP_BACKUP_PASSWORD

-- Run as the installation administrator. Passwords are read from the process
-- environment by psql, never interpolated into shell command arguments.
SELECT 'CREATE ROLE erp_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'erp_owner') \gexec
SELECT 'CREATE ROLE erp_migrator LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'erp_migrator') \gexec
SELECT 'CREATE ROLE erp_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'erp_app') \gexec
SELECT 'CREATE ROLE erp_readonly NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'erp_readonly') \gexec
SELECT 'CREATE ROLE erp_backup LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS'
WHERE NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'erp_backup') \gexec

ALTER ROLE erp_owner NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
ALTER ROLE erp_readonly NOLOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
ALTER ROLE erp_migrator LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD :'migrator_password';
ALTER ROLE erp_app LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD :'app_password';
ALTER ROLE erp_backup LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS PASSWORD :'backup_password';
GRANT erp_owner TO erp_migrator;
GRANT erp_readonly TO erp_backup;

DO $$
BEGIN
    IF pg_has_role('erp_app', 'erp_owner', 'MEMBER')
       OR pg_has_role('erp_backup', 'erp_owner', 'MEMBER') THEN
        RAISE EXCEPTION 'Runtime and backup identities must not belong to erp_owner; review existing role memberships';
    END IF;
END $$;

SELECT format('ALTER DATABASE %I OWNER TO erp_owner', current_database()) \gexec
SELECT format('REVOKE CREATE ON DATABASE %I FROM PUBLIC', current_database()) \gexec
SELECT format('GRANT CONNECT ON DATABASE %I TO erp_migrator, erp_app, erp_readonly', current_database()) \gexec
SELECT format('ALTER ROLE erp_migrator IN DATABASE %I SET role TO erp_owner', current_database()) \gexec
REVOKE CREATE ON SCHEMA public FROM PUBLIC;
