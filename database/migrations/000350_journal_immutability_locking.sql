-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION accounting.prevent_posted_journal_mutation()
RETURNS TRIGGER AS $$
DECLARE
    v_entry RECORD;
    v_ids UUID[];
BEGIN
    IF TG_TABLE_NAME = 'journal_entries' THEN
        IF TG_OP = 'DELETE' THEN
            IF OLD.status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot DELETE posted or reversed journal entry (ID: %)', OLD.id;
            END IF;
            RETURN OLD;
        ELSIF TG_OP = 'UPDATE' THEN
            IF OLD.status IN ('POSTED', 'REVERSED') THEN
                -- A reversal may add metadata only; it cannot rewrite money,
                -- source, dates, dimensions, identity, or original actors.
                IF NOT (OLD.status = 'POSTED' AND NEW.status = 'REVERSED'
                        AND NEW.reversal_of_id IS NOT NULL AND NEW.reversed_at IS NOT NULL
                        AND NEW.version = OLD.version + 1)
                   OR (to_jsonb(NEW) - ARRAY['status','reversed_at','reversal_of_id','version'])
                      IS DISTINCT FROM
                      (to_jsonb(OLD) - ARRAY['status','reversed_at','reversal_of_id','version']) THEN
                    RAISE EXCEPTION 'Cannot UPDATE posted or reversed journal entry (ID: %)', OLD.id;
                END IF;
            END IF;
            RETURN NEW;
        END IF;
    ELSIF TG_TABLE_NAME = 'journal_lines' THEN
        IF TG_OP = 'INSERT' THEN
            v_ids := ARRAY[NEW.journal_entry_id];
        ELSIF TG_OP = 'DELETE' THEN
            v_ids := ARRAY[OLD.journal_entry_id];
        ELSE
            v_ids := ARRAY[OLD.journal_entry_id, NEW.journal_entry_id];
        END IF;
        -- A status read alone races posting. Lock both parents on reassignment,
        -- in stable order, and keep the locks until this transaction commits.
        FOR v_entry IN
            SELECT id, status FROM accounting.journal_entries
            WHERE id = ANY(v_ids) ORDER BY id FOR UPDATE
        LOOP
            IF v_entry.status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot % journal lines for posted or reversed journal entry (ID: %)', TG_OP, v_entry.id;
            END IF;
        END LOOP;
        IF TG_OP = 'DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION accounting.prevent_posted_journal_mutation()
RETURNS TRIGGER AS $$
DECLARE
    v_status TEXT;
BEGIN
    IF TG_TABLE_NAME = 'journal_entries' THEN
        IF TG_OP = 'DELETE' THEN
            IF OLD.status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot DELETE posted or reversed journal entry (ID: %)', OLD.id;
            END IF;
            RETURN OLD;
        ELSIF TG_OP = 'UPDATE' THEN
            IF OLD.status IN ('POSTED', 'REVERSED') THEN
                -- Allow ONLY transitioning from POSTED to REVERSED when reversal metadata is recorded
                IF NOT (OLD.status = 'POSTED' AND NEW.status = 'REVERSED' AND NEW.reversal_of_id IS NOT NULL) THEN
                    RAISE EXCEPTION 'Cannot UPDATE posted or reversed journal entry (ID: %)', OLD.id;
                END IF;
            END IF;
            RETURN NEW;
        END IF;
    ELSIF TG_TABLE_NAME = 'journal_lines' THEN
        IF TG_OP IN ('DELETE', 'UPDATE') THEN
            SELECT status INTO v_status FROM accounting.journal_entries WHERE id = OLD.journal_entry_id;
            IF v_status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot % journal lines for posted or reversed journal entry (ID: %)', TG_OP, OLD.journal_entry_id;
            END IF;
            IF TG_OP = 'DELETE' THEN RETURN OLD; ELSE RETURN NEW; END IF;
        ELSIF TG_OP = 'INSERT' THEN
            SELECT status INTO v_status FROM accounting.journal_entries WHERE id = NEW.journal_entry_id;
            IF v_status IN ('POSTED', 'REVERSED') THEN
                RAISE EXCEPTION 'Cannot INSERT journal lines into posted or reversed journal entry (ID: %)', NEW.journal_entry_id;
            END IF;
            RETURN NEW;
        END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd
