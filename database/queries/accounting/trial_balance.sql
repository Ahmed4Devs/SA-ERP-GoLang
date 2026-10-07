-- name: GetTrialBalance :many
SELECT
    a.id AS account_id,
    a.code AS account_code,
    a.name_en AS account_name_en,
    a.name_ar AS account_name_ar,
    a.account_type,
    a.normal_balance,
    COALESCE(SUM(l.base_debit), 0)::NUMERIC(24,6) AS total_debit,
    COALESCE(SUM(l.base_credit), 0)::NUMERIC(24,6) AS total_credit,
    CASE
        WHEN a.normal_balance = 'DEBIT' THEN (COALESCE(SUM(l.base_debit), 0) - COALESCE(SUM(l.base_credit), 0))
        ELSE (COALESCE(SUM(l.base_credit), 0) - COALESCE(SUM(l.base_debit), 0))
    END::NUMERIC(24,6) AS net_balance
FROM accounting.accounts a
LEFT JOIN (
    SELECT jl.account_id, jl.base_debit, jl.base_credit
    FROM accounting.journal_lines jl
    JOIN accounting.journal_entries je ON je.id = jl.journal_entry_id
    WHERE je.legal_entity_id = $1
      AND je.status IN ('POSTED', 'REVERSED')
      AND ($2::DATE IS NULL OR je.accounting_date <= $2::DATE)
) l ON l.account_id = a.id
WHERE a.legal_entity_id = $1
GROUP BY a.id, a.code, a.name_en, a.name_ar, a.account_type, a.normal_balance
ORDER BY a.code ASC;

-- name: GetTrialBalanceSummary :one
SELECT
    COALESCE(SUM(l.base_debit), 0)::NUMERIC(24,6) AS sum_debit,
    COALESCE(SUM(l.base_credit), 0)::NUMERIC(24,6) AS sum_credit,
    (COALESCE(SUM(l.base_debit), 0) = COALESCE(SUM(l.base_credit), 0)) AS is_balanced
FROM accounting.journal_lines l
JOIN accounting.journal_entries je ON je.id = l.journal_entry_id
WHERE je.legal_entity_id = $1
  AND je.status IN ('POSTED', 'REVERSED')
  AND ($2::DATE IS NULL OR je.accounting_date <= $2::DATE);
