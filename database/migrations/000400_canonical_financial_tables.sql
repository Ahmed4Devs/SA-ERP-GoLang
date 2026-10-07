-- +goose Up
-- +goose StatementBegin
-- New financial records carry their company in every foreign key. Existing
-- journals and movements are never rewritten to populate these new tables.
ALTER TABLE accounting.accounts ADD CONSTRAINT uq_accounts_id_entity UNIQUE (id, legal_entity_id);
ALTER TABLE accounting.fiscal_periods ADD CONSTRAINT uq_fiscal_periods_id_entity UNIQUE (id, legal_entity_id);
ALTER TABLE accounting.journals ADD CONSTRAINT uq_journals_id_entity UNIQUE (id, legal_entity_id);
ALTER TABLE accounting.journal_entries ADD CONSTRAINT uq_journal_entries_id_entity UNIQUE (id, legal_entity_id);
ALTER TABLE accounting.journal_lines ADD CONSTRAINT uq_journal_lines_id_entity UNIQUE (id, legal_entity_id);
ALTER TABLE products.items ADD CONSTRAINT uq_items_id_entity UNIQUE (id, legal_entity_id);
ALTER TABLE warehouse.warehouses ADD CONSTRAINT uq_warehouses_id_entity UNIQUE (id, legal_entity_id);

CREATE TABLE accounting.dimensions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    code TEXT NOT NULL CHECK (btrim(code) <> ''),
    name_en TEXT NOT NULL CHECK (btrim(name_en) <> ''),
    name_ar TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (legal_entity_id, code),
    UNIQUE (id, legal_entity_id)
);
CREATE TABLE accounting.dimension_values (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    dimension_id UUID NOT NULL,
    code TEXT NOT NULL CHECK (btrim(code) <> ''),
    name_en TEXT NOT NULL CHECK (btrim(name_en) <> ''),
    name_ar TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (dimension_id, legal_entity_id) REFERENCES accounting.dimensions(id, legal_entity_id),
    UNIQUE (dimension_id, code),
    UNIQUE (id, dimension_id, legal_entity_id)
);
CREATE TABLE accounting.journal_line_dimensions (
    journal_line_id UUID NOT NULL,
    legal_entity_id UUID NOT NULL,
    dimension_id UUID NOT NULL,
    dimension_value_id UUID NOT NULL,
    PRIMARY KEY (journal_line_id, dimension_id),
    FOREIGN KEY (journal_line_id, legal_entity_id) REFERENCES accounting.journal_lines(id, legal_entity_id),
    FOREIGN KEY (dimension_value_id, dimension_id, legal_entity_id)
        REFERENCES accounting.dimension_values(id, dimension_id, legal_entity_id)
);
CREATE INDEX idx_journal_line_dimensions_value ON accounting.journal_line_dimensions(legal_entity_id, dimension_value_id);

CREATE TABLE accounting.posting_rules (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    source_module TEXT NOT NULL CHECK (btrim(source_module) <> ''),
    code TEXT NOT NULL CHECK (btrim(code) <> ''),
    name_en TEXT NOT NULL CHECK (btrim(name_en) <> ''),
    name_ar TEXT NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (legal_entity_id, source_module, code),
    UNIQUE (id, legal_entity_id)
);
-- Each code identifies one posting role, e.g. AR_CONTROL or REALIZED_FX_GAIN.
-- A new effective version replaces configuration without rewriting history.
CREATE TABLE accounting.posting_rule_versions (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL,
    posting_rule_id UUID NOT NULL,
    version_number INTEGER NOT NULL CHECK (version_number > 0),
    effective_from DATE NOT NULL,
    account_id UUID NOT NULL,
    journal_id UUID,
    created_by UUID NOT NULL REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (posting_rule_id, legal_entity_id) REFERENCES accounting.posting_rules(id, legal_entity_id),
    FOREIGN KEY (account_id, legal_entity_id) REFERENCES accounting.accounts(id, legal_entity_id),
    FOREIGN KEY (journal_id, legal_entity_id) REFERENCES accounting.journals(id, legal_entity_id),
    UNIQUE (posting_rule_id, version_number)
);
CREATE INDEX idx_posting_rule_versions_effective ON accounting.posting_rule_versions(posting_rule_id, effective_from DESC, version_number DESC);

CREATE TABLE accounting.exchange_rates (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    from_currency CHAR(3) NOT NULL REFERENCES ref.currencies(code),
    to_currency CHAR(3) NOT NULL REFERENCES ref.currencies(code),
    effective_date DATE NOT NULL,
    rate NUMERIC(24,12) NOT NULL CHECK (rate > 0),
    version_number INTEGER NOT NULL DEFAULT 1 CHECK (version_number > 0),
    source TEXT NOT NULL CHECK (btrim(source) <> ''),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (from_currency <> to_currency OR rate = 1),
    UNIQUE (legal_entity_id, from_currency, to_currency, effective_date, version_number)
);
CREATE INDEX idx_exchange_rates_lookup ON accounting.exchange_rates(legal_entity_id, from_currency, to_currency, effective_date DESC, version_number DESC);

CREATE TABLE accounting.opening_balances (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    fiscal_period_id UUID NOT NULL,
    account_id UUID NOT NULL,
    currency_code CHAR(3) NOT NULL REFERENCES ref.currencies(code),
    debit NUMERIC(24,6) NOT NULL DEFAULT 0 CHECK (debit >= 0),
    credit NUMERIC(24,6) NOT NULL DEFAULT 0 CHECK (credit >= 0),
    journal_entry_id UUID,
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','POSTED')),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (fiscal_period_id, legal_entity_id) REFERENCES accounting.fiscal_periods(id, legal_entity_id),
    FOREIGN KEY (account_id, legal_entity_id) REFERENCES accounting.accounts(id, legal_entity_id),
    FOREIGN KEY (journal_entry_id, legal_entity_id) REFERENCES accounting.journal_entries(id, legal_entity_id),
    CHECK ((debit > 0 AND credit = 0) OR (credit > 0 AND debit = 0)),
    CHECK (status <> 'POSTED' OR journal_entry_id IS NOT NULL),
    UNIQUE (legal_entity_id, fiscal_period_id, account_id, currency_code)
);

CREATE TABLE accounting.reconciliations (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    account_id UUID NOT NULL,
    reference TEXT NOT NULL CHECK (btrim(reference) <> ''),
    reconciliation_date DATE NOT NULL,
    currency_code CHAR(3) NOT NULL REFERENCES ref.currencies(code),
    status TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT','COMPLETED')),
    created_by UUID NOT NULL REFERENCES iam.users(id),
    completed_by UUID REFERENCES iam.users(id),
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (account_id, legal_entity_id) REFERENCES accounting.accounts(id, legal_entity_id),
    CHECK (status <> 'COMPLETED' OR (completed_by IS NOT NULL AND completed_by <> created_by AND completed_at IS NOT NULL)),
    UNIQUE (legal_entity_id, reference),
    UNIQUE (id, legal_entity_id)
);
CREATE TABLE accounting.reconciliation_lines (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL,
    reconciliation_id UUID NOT NULL,
    journal_line_id UUID NOT NULL UNIQUE,
    FOREIGN KEY (reconciliation_id, legal_entity_id) REFERENCES accounting.reconciliations(id, legal_entity_id),
    FOREIGN KEY (journal_line_id, legal_entity_id) REFERENCES accounting.journal_lines(id, legal_entity_id)
);
CREATE INDEX idx_reconciliation_lines_parent ON accounting.reconciliation_lines(reconciliation_id);

CREATE TABLE inventory.item_costs (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    item_id UUID NOT NULL,
    warehouse_id UUID NOT NULL,
    costing_method TEXT NOT NULL CHECK (costing_method IN ('FIFO','WEIGHTED_AVERAGE','STANDARD')),
    quantity_on_hand NUMERIC(24,6) NOT NULL,
    unit_cost NUMERIC(24,6) NOT NULL CHECK (unit_cost >= 0),
    total_value NUMERIC(24,6) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    FOREIGN KEY (item_id, legal_entity_id) REFERENCES products.items(id, legal_entity_id),
    FOREIGN KEY (warehouse_id, legal_entity_id) REFERENCES warehouse.warehouses(id, legal_entity_id),
    UNIQUE (legal_entity_id, item_id, warehouse_id)
);
CREATE TABLE inventory.valuation_snapshots (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id),
    snapshot_batch_id UUID NOT NULL,
    snapshot_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    item_id UUID NOT NULL,
    warehouse_id UUID NOT NULL,
    costing_method TEXT NOT NULL CHECK (costing_method IN ('FIFO','WEIGHTED_AVERAGE','STANDARD')),
    quantity_on_hand NUMERIC(24,6) NOT NULL,
    unit_cost NUMERIC(24,6) NOT NULL CHECK (unit_cost >= 0),
    total_value NUMERIC(24,6) NOT NULL,
    created_by UUID NOT NULL REFERENCES iam.users(id),
    FOREIGN KEY (item_id, legal_entity_id) REFERENCES products.items(id, legal_entity_id),
    FOREIGN KEY (warehouse_id, legal_entity_id) REFERENCES warehouse.warehouses(id, legal_entity_id),
    UNIQUE (snapshot_batch_id, item_id, warehouse_id)
);
CREATE INDEX idx_valuation_snapshots_entity_date ON inventory.valuation_snapshots(legal_entity_id, snapshot_at DESC);

-- This is a rebuildable warehouse-level projection of all bins. Atomic deltas
-- preserve concurrent movements in different bins without a read/sum/write race.
INSERT INTO inventory.item_costs(legal_entity_id,item_id,warehouse_id,costing_method,quantity_on_hand,unit_cost,total_value)
SELECT sb.legal_entity_id,sb.item_id,sb.warehouse_id,i.costing_method,sum(sb.quantity_on_hand),
    CASE WHEN sum(sb.quantity_on_hand)<>0 THEN greatest(0,round(sum(sb.total_value)/sum(sb.quantity_on_hand),6)) ELSE 0 END,
    sum(sb.total_value)
FROM inventory.stock_balances sb JOIN products.items i ON i.id=sb.item_id
GROUP BY sb.legal_entity_id,sb.item_id,sb.warehouse_id,i.costing_method;
CREATE FUNCTION inventory.apply_item_cost_delta(p_entity UUID,p_item UUID,p_warehouse UUID,p_quantity NUMERIC,p_value NUMERIC) RETURNS void AS $$
BEGIN
    INSERT INTO inventory.item_costs(legal_entity_id,item_id,warehouse_id,costing_method,quantity_on_hand,unit_cost,total_value)
    SELECT p_entity,p_item,p_warehouse,costing_method,p_quantity,
        CASE WHEN p_quantity<>0 THEN greatest(0,round(p_value/p_quantity,6)) ELSE 0 END,p_value
    FROM products.items WHERE id=p_item AND legal_entity_id=p_entity
    ON CONFLICT (legal_entity_id,item_id,warehouse_id) DO UPDATE SET
        quantity_on_hand=inventory.item_costs.quantity_on_hand+p_quantity,
        total_value=inventory.item_costs.total_value+p_value,
        unit_cost=CASE WHEN inventory.item_costs.quantity_on_hand+p_quantity<>0
            THEN greatest(0,round((inventory.item_costs.total_value+p_value)/(inventory.item_costs.quantity_on_hand+p_quantity),6)) ELSE 0 END,
        costing_method=EXCLUDED.costing_method,updated_at=now();
END;
$$ LANGUAGE plpgsql;
CREATE FUNCTION inventory.project_item_costs() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP='UPDATE' AND OLD.legal_entity_id=NEW.legal_entity_id AND OLD.item_id=NEW.item_id AND OLD.warehouse_id=NEW.warehouse_id THEN
        PERFORM inventory.apply_item_cost_delta(NEW.legal_entity_id,NEW.item_id,NEW.warehouse_id,NEW.quantity_on_hand-OLD.quantity_on_hand,NEW.total_value-OLD.total_value);
    ELSE
        IF TG_OP<>'INSERT' THEN PERFORM inventory.apply_item_cost_delta(OLD.legal_entity_id,OLD.item_id,OLD.warehouse_id,-OLD.quantity_on_hand,-OLD.total_value); END IF;
        IF TG_OP<>'DELETE' THEN PERFORM inventory.apply_item_cost_delta(NEW.legal_entity_id,NEW.item_id,NEW.warehouse_id,NEW.quantity_on_hand,NEW.total_value); END IF;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trg_stock_balances_item_costs AFTER INSERT OR UPDATE OR DELETE ON inventory.stock_balances
FOR EACH ROW EXECUTE FUNCTION inventory.project_item_costs();

CREATE FUNCTION accounting.prevent_reference_history_mutation() RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION '% is append-only; create a new version instead', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trg_posting_rule_versions_immutable BEFORE UPDATE OR DELETE ON accounting.posting_rule_versions
FOR EACH ROW EXECUTE FUNCTION accounting.prevent_reference_history_mutation();
CREATE TRIGGER trg_exchange_rates_immutable BEFORE UPDATE OR DELETE ON accounting.exchange_rates
FOR EACH ROW EXECUTE FUNCTION accounting.prevent_reference_history_mutation();
CREATE TRIGGER trg_valuation_snapshots_immutable BEFORE UPDATE OR DELETE ON inventory.valuation_snapshots
FOR EACH ROW EXECUTE FUNCTION accounting.prevent_reference_history_mutation();

CREATE FUNCTION accounting.protect_journal_line_dimensions() RETURNS TRIGGER AS $$
DECLARE
    v_ids UUID[];
    v_entry RECORD;
BEGIN
    IF TG_OP = 'INSERT' THEN v_ids := ARRAY[NEW.journal_line_id];
    ELSIF TG_OP = 'DELETE' THEN v_ids := ARRAY[OLD.journal_line_id];
    ELSE v_ids := ARRAY[OLD.journal_line_id, NEW.journal_line_id]; END IF;
    FOR v_entry IN SELECT je.id, je.status FROM accounting.journal_entries je
        WHERE je.id IN (SELECT journal_entry_id FROM accounting.journal_lines WHERE id = ANY(v_ids))
        ORDER BY je.id FOR UPDATE
    LOOP
        IF v_entry.status IN ('POSTED','REVERSED') THEN
            RAISE EXCEPTION 'Cannot change dimensions of posted or reversed journal %', v_entry.id;
        END IF;
    END LOOP;
    IF TG_OP = 'DELETE' THEN RETURN OLD; END IF;
    IF NOT EXISTS (SELECT 1 FROM accounting.dimension_values dv JOIN accounting.dimensions d ON d.id=dv.dimension_id
        WHERE dv.id=NEW.dimension_value_id AND dv.is_active AND d.is_active) THEN
        RAISE EXCEPTION 'Journal dimension and value must be active';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trg_journal_line_dimensions_immutable BEFORE INSERT OR UPDATE OR DELETE ON accounting.journal_line_dimensions
FOR EACH ROW EXECUTE FUNCTION accounting.protect_journal_line_dimensions();

CREATE FUNCTION accounting.validate_opening_balance() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP <> 'INSERT' AND OLD.status='POSTED' THEN
        RAISE EXCEPTION 'Posted opening balances are immutable';
    END IF;
    IF TG_OP='DELETE' THEN RETURN OLD; END IF;
    IF NEW.status='POSTED' AND NOT EXISTS (
        SELECT 1 FROM accounting.journal_entries je JOIN accounting.journal_lines jl ON jl.journal_entry_id=je.id
        WHERE je.id=NEW.journal_entry_id AND je.legal_entity_id=NEW.legal_entity_id
            AND je.fiscal_period_id=NEW.fiscal_period_id AND je.status IN ('POSTED','REVERSED')
            AND jl.account_id=NEW.account_id AND jl.currency_code=NEW.currency_code
        GROUP BY je.id HAVING sum(jl.transaction_debit)=NEW.debit AND sum(jl.transaction_credit)=NEW.credit
    ) THEN
        RAISE EXCEPTION 'Opening balance must match a posted journal in the same fiscal period';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trg_opening_balance_history BEFORE INSERT OR UPDATE OR DELETE ON accounting.opening_balances
FOR EACH ROW EXECUTE FUNCTION accounting.validate_opening_balance();

CREATE FUNCTION accounting.validate_reconciliation() RETURNS TRIGGER AS $$
DECLARE
    v_count BIGINT;
    v_balance NUMERIC;
BEGIN
    IF TG_OP <> 'INSERT' AND OLD.status='COMPLETED' THEN RAISE EXCEPTION 'Completed reconciliation is immutable'; END IF;
    IF TG_OP='DELETE' THEN RETURN OLD; END IF;
    IF NEW.status='COMPLETED' THEN
        SELECT count(*), sum(jl.transaction_debit-jl.transaction_credit) INTO v_count,v_balance
        FROM accounting.reconciliation_lines rl JOIN accounting.journal_lines jl ON jl.id=rl.journal_line_id
        JOIN accounting.journal_entries je ON je.id=jl.journal_entry_id
        WHERE rl.reconciliation_id=NEW.id AND jl.account_id=NEW.account_id AND jl.currency_code=NEW.currency_code
            AND je.status IN ('POSTED','REVERSED') AND je.accounting_date<=NEW.reconciliation_date;
        IF v_count < 2 OR v_balance <> 0 OR v_count <> (SELECT count(*) FROM accounting.reconciliation_lines WHERE reconciliation_id=NEW.id) THEN
            RAISE EXCEPTION 'Reconciliation requires balanced posted lines of its account and currency on or before its date';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trg_reconciliation_history BEFORE INSERT OR UPDATE OR DELETE ON accounting.reconciliations
FOR EACH ROW EXECUTE FUNCTION accounting.validate_reconciliation();
CREATE FUNCTION accounting.protect_reconciliation_lines() RETURNS TRIGGER AS $$
DECLARE v_parent RECORD; v_ids UUID[];
BEGIN
    IF TG_OP='INSERT' THEN v_ids:=ARRAY[NEW.reconciliation_id];
    ELSIF TG_OP='DELETE' THEN v_ids:=ARRAY[OLD.reconciliation_id];
    ELSE v_ids:=ARRAY[OLD.reconciliation_id,NEW.reconciliation_id]; END IF;
    FOR v_parent IN SELECT id,status FROM accounting.reconciliations WHERE id=ANY(v_ids) ORDER BY id FOR UPDATE LOOP
        IF v_parent.status='COMPLETED' THEN RAISE EXCEPTION 'Completed reconciliation lines are immutable'; END IF;
    END LOOP;
    IF TG_OP='DELETE' THEN RETURN OLD; END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
CREATE TRIGGER trg_reconciliation_lines_history BEFORE INSERT OR UPDATE OR DELETE ON accounting.reconciliation_lines
FOR EACH ROW EXECUTE FUNCTION accounting.protect_reconciliation_lines();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'Financial history migrations are forward-only; restore a verified backup to roll back'; END $$;
-- +goose StatementEnd
