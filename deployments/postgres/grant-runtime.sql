\set ON_ERROR_STOP on

-- Execute as erp_migrator (which assumes erp_owner) after every schema upgrade.
-- Runtime credentials cannot change schema, disable triggers, or rewrite audit.
SELECT format('GRANT USAGE ON SCHEMA %I TO erp_app, erp_readonly', nspname)
FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname <> 'information_schema' \gexec

SELECT format('GRANT SELECT ON ALL TABLES IN SCHEMA %I TO erp_readonly', nspname)
FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname <> 'information_schema' \gexec

SELECT format('GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %I TO erp_app', nspname)
FROM pg_namespace
WHERE nspname NOT LIKE 'pg_%' AND nspname NOT IN ('information_schema', 'audit', 'ref', 'public') \gexec

GRANT SELECT, INSERT ON ALL TABLES IN SCHEMA audit TO erp_app;
REVOKE UPDATE, DELETE, TRUNCATE, REFERENCES, TRIGGER ON ALL TABLES IN SCHEMA audit FROM erp_app;
GRANT SELECT ON ALL TABLES IN SCHEMA ref TO erp_app;
GRANT SELECT ON public.goose_db_version TO erp_app;

SELECT format('GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %I TO erp_app', nspname)
FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname NOT IN ('information_schema', 'public', 'ref') \gexec
SELECT format('GRANT SELECT ON ALL SEQUENCES IN SCHEMA %I TO erp_readonly', nspname)
FROM pg_namespace WHERE nspname NOT LIKE 'pg_%' AND nspname <> 'information_schema' \gexec
