-- +goose Up
-- +goose StatementBegin

-- Journal lines must never target header accounts or cross legal-entity
-- boundaries. Some domain services write journal lines directly, so this
-- invariant belongs in the database as well as in the posting engine.
CREATE OR REPLACE FUNCTION accounting.validate_journal_line_account()
RETURNS TRIGGER AS $$
DECLARE
    v_account_entity_id UUID;
    v_entry_entity_id UUID;
    v_is_active BOOLEAN;
    v_allow_posting BOOLEAN;
BEGIN
    SELECT legal_entity_id, is_active, allow_posting
      INTO v_account_entity_id, v_is_active, v_allow_posting
    FROM accounting.accounts
    WHERE id = NEW.account_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Journal line account % does not exist', NEW.account_id;
    END IF;

    SELECT legal_entity_id
      INTO v_entry_entity_id
    FROM accounting.journal_entries
    WHERE id = NEW.journal_entry_id;

    IF NOT FOUND THEN
        RAISE EXCEPTION 'Journal entry % does not exist', NEW.journal_entry_id;
    END IF;

    IF NEW.legal_entity_id <> v_entry_entity_id
       OR NEW.legal_entity_id <> v_account_entity_id THEN
        RAISE EXCEPTION 'Journal entry, line, and account must belong to the same legal entity';
    END IF;

    IF NOT v_is_active OR NOT v_allow_posting THEN
        RAISE EXCEPTION 'Account % is inactive or does not allow posting', NEW.account_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_validate_journal_line_account ON accounting.journal_lines;
CREATE TRIGGER trg_validate_journal_line_account
BEFORE INSERT OR UPDATE OF journal_entry_id, legal_entity_id, account_id
ON accounting.journal_lines
FOR EACH ROW
EXECUTE FUNCTION accounting.validate_journal_line_account();

-- Upgrade supplier defaults created by older releases from the OPEX header
-- account (520000) to a posting-level expense account (523000).
UPDATE payables.suppliers supplier
SET default_expense_account_id = posting_account.id,
    updated_at = now()
FROM accounting.accounts old_account
JOIN accounting.accounts posting_account
  ON posting_account.legal_entity_id = old_account.legal_entity_id
 AND posting_account.code = '523000'
 AND posting_account.is_active
 AND posting_account.allow_posting
WHERE supplier.default_expense_account_id = old_account.id
  AND old_account.code = '520000';

-- Repair the installed seed function for databases upgraded from an older
-- schema. Fresh databases already receive this definition from migration 140.
DO $$
DECLARE
    function_definition TEXT;
BEGIN
    SELECT pg_get_functiondef('payables.seed_default_payables_data(uuid)'::regprocedure)
    INTO function_definition;

    function_definition := replace(
        function_definition,
        'code = ''520000'' LIMIT 1',
        'code = ''523000'' AND is_active AND allow_posting LIMIT 1'
    );

    EXECUTE function_definition;
END;
$$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_validate_journal_line_account ON accounting.journal_lines;
DROP FUNCTION IF EXISTS accounting.validate_journal_line_account();
-- +goose StatementEnd
