-- name: GetTrialBalanceReport :many
SELECT
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(l.base_debit), 0)::NUMERIC(18, 4) AS total_debit,
    COALESCE(SUM(l.base_credit), 0)::NUMERIC(18, 4) AS total_credit,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(l.base_debit), 0) - COALESCE(SUM(l.base_credit), 0))
        ELSE (COALESCE(SUM(l.base_credit), 0) - COALESCE(SUM(l.base_debit), 0))
    END::NUMERIC(18, 4) AS net_balance
FROM accounting.accounts a
LEFT JOIN (
    SELECT jl.account_id, jl.base_debit, jl.base_credit
    FROM accounting.journal_lines jl
    JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
    WHERE je.legal_entity_id = $1
      AND je.status = 'POSTED'
      AND ($2::DATE IS NULL OR je.accounting_date >= $2::DATE)
      AND ($3::DATE IS NULL OR je.accounting_date <= $3::DATE)
) l ON l.account_id = a.id
WHERE a.legal_entity_id = $1 AND a.allow_posting = true
GROUP BY a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance
ORDER BY a.code ASC;

-- name: GetGeneralLedgerReport :many
SELECT
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
WHERE je.legal_entity_id = $1
  AND je.status = 'POSTED'
  AND ($2::UUID IS NULL OR a.id = $2::UUID)
  AND ($3::DATE IS NULL OR je.accounting_date >= $3::DATE)
  AND ($4::DATE IS NULL OR je.accounting_date <= $4::DATE)
ORDER BY je.accounting_date ASC, je.entry_number ASC, jl.line_number ASC;

-- name: GetBalanceSheetAccounts :many
SELECT
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(l.base_debit), 0) - COALESCE(SUM(l.base_credit), 0))
        ELSE (COALESCE(SUM(l.base_credit), 0) - COALESCE(SUM(l.base_debit), 0))
    END::NUMERIC(18, 4) AS balance
FROM accounting.accounts a
LEFT JOIN (
    SELECT jl.account_id, jl.base_debit, jl.base_credit
    FROM accounting.journal_lines jl
    JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
    WHERE je.legal_entity_id = $1
      AND je.status = 'POSTED'
      AND ($2::DATE IS NULL OR je.accounting_date <= $2::DATE)
) l ON l.account_id = a.id
WHERE a.legal_entity_id = $1
  AND a.allow_posting = true
  AND a.account_type IN ('ASSET', 'LIABILITY', 'EQUITY')
GROUP BY a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance
ORDER BY a.code ASC;

-- name: GetIncomeStatementAccounts :many
SELECT
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(l.base_debit), 0) - COALESCE(SUM(l.base_credit), 0))
        ELSE (COALESCE(SUM(l.base_credit), 0) - COALESCE(SUM(l.base_debit), 0))
    END::NUMERIC(18, 4) AS period_balance
FROM accounting.accounts a
LEFT JOIN (
    SELECT jl.account_id, jl.base_debit, jl.base_credit
    FROM accounting.journal_lines jl
    JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
    WHERE je.legal_entity_id = $1
      AND je.status = 'POSTED'
      AND ($2::DATE IS NULL OR je.accounting_date >= $2::DATE)
      AND ($3::DATE IS NULL OR je.accounting_date <= $3::DATE)
) l ON l.account_id = a.id
WHERE a.legal_entity_id = $1
  AND a.allow_posting = true
  AND a.account_type IN ('REVENUE', 'EXPENSE')
GROUP BY a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance
ORDER BY a.code ASC;
