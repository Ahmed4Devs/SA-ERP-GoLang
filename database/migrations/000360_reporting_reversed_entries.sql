-- +goose Up
-- Reversal adds a compensating entry; the original remains part of the ledger.
CREATE OR REPLACE VIEW reporting.trial_balance AS
SELECT
    a.legal_entity_id,
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(jl.base_debit), 0)::NUMERIC(18, 4) AS total_debit,
    COALESCE(SUM(jl.base_credit), 0)::NUMERIC(18, 4) AS total_credit,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(jl.base_debit), 0) - COALESCE(SUM(jl.base_credit), 0))
        ELSE (COALESCE(SUM(jl.base_credit), 0) - COALESCE(SUM(jl.base_debit), 0))
    END::NUMERIC(18, 4) AS net_balance
FROM accounting.accounts a
LEFT JOIN (
    SELECT jl.* FROM accounting.journal_lines jl
    JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
    WHERE je.status IN ('POSTED', 'REVERSED')
) jl ON jl.account_id = a.id AND jl.legal_entity_id = a.legal_entity_id
GROUP BY a.legal_entity_id, a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance;

-- 2. General Ledger View
CREATE OR REPLACE VIEW reporting.general_ledger AS
SELECT
    je.legal_entity_id,
    je.id AS entry_id,
    je.entry_number,
    je.accounting_date,
    je.source_module,
    je.reference,
    jl.id AS line_id,
    jl.line_number,
    jl.account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    jl.base_debit,
    jl.base_credit,
    jl.description
FROM accounting.journal_entries je
JOIN accounting.journal_lines jl ON jl.journal_entry_id = je.id
JOIN accounting.accounts a ON a.id = jl.account_id
WHERE je.status IN ('POSTED', 'REVERSED');


-- +goose Down
CREATE OR REPLACE VIEW reporting.trial_balance AS
SELECT
    a.legal_entity_id,
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(jl.base_debit), 0)::NUMERIC(18, 4) AS total_debit,
    COALESCE(SUM(jl.base_credit), 0)::NUMERIC(18, 4) AS total_credit,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(jl.base_debit), 0) - COALESCE(SUM(jl.base_credit), 0))
        ELSE (COALESCE(SUM(jl.base_credit), 0) - COALESCE(SUM(jl.base_debit), 0))
    END::NUMERIC(18, 4) AS net_balance
FROM accounting.accounts a
LEFT JOIN accounting.journal_lines jl ON jl.account_id = a.id
LEFT JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id AND je.status = 'POSTED'
GROUP BY a.legal_entity_id, a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance;

-- 2. General Ledger View
CREATE OR REPLACE VIEW reporting.general_ledger AS
SELECT
    je.legal_entity_id,
    je.id AS entry_id,
    je.entry_number,
    je.accounting_date,
    je.source_module,
    je.reference,
    jl.id AS line_id,
    jl.line_number,
    jl.account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    jl.base_debit,
    jl.base_credit,
    jl.description
FROM accounting.journal_entries je
JOIN accounting.journal_lines jl ON jl.journal_entry_id = je.id
JOIN accounting.accounts a ON a.id = jl.account_id
WHERE je.status = 'POSTED';
