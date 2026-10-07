\set ON_ERROR_STOP on

-- Run as erp_app against a migrated database after grant-runtime.sql.
DO $$
BEGIN
    IF current_user <> 'erp_app'
       OR current_setting('is_superuser') <> 'off'
       OR pg_has_role(current_user, 'erp_owner', 'MEMBER') THEN
        RAISE EXCEPTION 'The runtime must be erp_app without owner privileges';
    END IF;
    IF NOT has_table_privilege(current_user, 'accounting.journal_entries', 'SELECT,INSERT,UPDATE,DELETE')
       OR NOT has_table_privilege(current_user, 'audit.events', 'SELECT,INSERT')
       OR NOT has_table_privilege(current_user, 'goose_db_version', 'SELECT') THEN
        RAISE EXCEPTION 'Required runtime privileges are missing';
    END IF;
    BEGIN
        EXECUTE 'DELETE FROM audit.events WHERE false';
        RAISE EXCEPTION 'Runtime can delete audit history';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    BEGIN
        EXECUTE 'ALTER TABLE accounting.journal_entries DISABLE TRIGGER ALL';
        RAISE EXCEPTION 'Runtime can disable financial invariants';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    BEGIN
        EXECUTE 'CREATE TABLE accounting.forbidden_runtime_ddl (id integer)';
        RAISE EXCEPTION 'Runtime can create schema objects';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
    BEGIN
        EXECUTE 'DELETE FROM goose_db_version WHERE false';
        RAISE EXCEPTION 'Runtime can mutate migration metadata';
    EXCEPTION WHEN insufficient_privilege THEN NULL;
    END;
END $$;

SELECT count(*) AS reference_currency_count FROM ref.currencies;
