-- +goose Up
-- Historical posted amounts are preserved. New invoices retain their tax code.
ALTER TABLE payables.supplier_invoice_lines
    ADD COLUMN tax_code_id UUID REFERENCES tax.tax_codes(id);

-- +goose Down
ALTER TABLE payables.supplier_invoice_lines DROP COLUMN tax_code_id;
