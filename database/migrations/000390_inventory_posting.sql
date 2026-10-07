-- +goose Up
ALTER TABLE inventory.stock_adjustments
 ADD COLUMN inventory_account_id UUID REFERENCES accounting.accounts(id),
 ADD COLUMN offset_account_id UUID REFERENCES accounting.accounts(id),
 ADD COLUMN journal_entry_id UUID REFERENCES accounting.journal_entries(id);
ALTER TABLE inventory.stock_counts
 ADD COLUMN inventory_account_id UUID REFERENCES accounting.accounts(id),
 ADD COLUMN offset_account_id UUID REFERENCES accounting.accounts(id),
 ADD COLUMN journal_entry_id UUID REFERENCES accounting.journal_entries(id),
 ADD COLUMN created_by UUID REFERENCES iam.users(id);
ALTER TABLE inventory.stock_count_lines ADD COLUMN unit_cost NUMERIC(24,6) NOT NULL DEFAULT 0 CHECK (unit_cost >= 0);

-- +goose Down
ALTER TABLE inventory.stock_count_lines DROP COLUMN unit_cost;
ALTER TABLE inventory.stock_counts DROP COLUMN inventory_account_id, DROP COLUMN offset_account_id, DROP COLUMN journal_entry_id, DROP COLUMN created_by;
ALTER TABLE inventory.stock_adjustments DROP COLUMN inventory_account_id, DROP COLUMN offset_account_id, DROP COLUMN journal_entry_id;
