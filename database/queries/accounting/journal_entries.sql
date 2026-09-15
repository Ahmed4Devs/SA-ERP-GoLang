-- name: CreateJournalEntry :one
INSERT INTO accounting.journal_entries (
    legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
    status, source_module, source_type, source_id, reference, description,
    currency_code, exchange_rate, total_debit, total_credit, created_by
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10, $11,
    $12, $13, $14, $15, $16
) RETURNING id, legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
            status, source_module, source_type, source_id, reference, description,
            currency_code, exchange_rate, total_debit, total_credit,
            created_by, approved_by, posted_by, created_at, approved_at, posted_at,
            reversed_at, reversal_of_id, version;

-- name: GetJournalEntryByID :one
SELECT id, legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
       status, source_module, source_type, source_id, reference, description,
       currency_code, exchange_rate, total_debit, total_credit,
       created_by, approved_by, posted_by, created_at, approved_at, posted_at,
       reversed_at, reversal_of_id, version
FROM accounting.journal_entries
WHERE id = $1;

-- name: GetJournalEntryByNumber :one
SELECT id, legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
       status, source_module, source_type, source_id, reference, description,
       currency_code, exchange_rate, total_debit, total_credit,
       created_by, approved_by, posted_by, created_at, approved_at, posted_at,
       reversed_at, reversal_of_id, version
FROM accounting.journal_entries
WHERE legal_entity_id = $1 AND entry_number = $2;

-- name: ListJournalEntriesByLegalEntity :many
SELECT id, legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
       status, source_module, source_type, source_id, reference, description,
       currency_code, exchange_rate, total_debit, total_credit,
       created_by, approved_by, posted_by, created_at, approved_at, posted_at,
       reversed_at, reversal_of_id, version
FROM accounting.journal_entries
WHERE legal_entity_id = $1
ORDER BY accounting_date DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountJournalEntriesByLegalEntity :one
SELECT COUNT(*)
FROM accounting.journal_entries
WHERE legal_entity_id = $1;

-- name: UpdateJournalEntryStatus :one
UPDATE accounting.journal_entries
SET status = $2,
    updated_at = now(),
    version = version + 1
WHERE id = $1 AND status = 'DRAFT'
RETURNING id, legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
            status, source_module, source_type, source_id, reference, description,
            currency_code, exchange_rate, total_debit, total_credit,
            created_by, approved_by, posted_by, created_at, approved_at, posted_at,
            reversed_at, reversal_of_id, version;

-- name: PostJournalEntry :one
UPDATE accounting.journal_entries
SET status = 'POSTED',
    posted_by = $2,
    posted_at = now(),
    version = version + 1
WHERE id = $1 AND status IN ('DRAFT', 'SUBMITTED', 'APPROVED')
RETURNING id, legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
            status, source_module, source_type, source_id, reference, description,
            currency_code, exchange_rate, total_debit, total_credit,
            created_by, approved_by, posted_by, created_at, approved_at, posted_at,
            reversed_at, reversal_of_id, version;

-- name: MarkJournalEntryReversed :one
UPDATE accounting.journal_entries
SET status = 'REVERSED',
    reversed_at = now(),
    reversal_of_id = $2,
    version = version + 1
WHERE id = $1 AND status = 'POSTED'
RETURNING id, legal_entity_id, journal_id, fiscal_period_id, entry_number, accounting_date,
            status, source_module, source_type, source_id, reference, description,
            currency_code, exchange_rate, total_debit, total_credit,
            created_by, approved_by, posted_by, created_at, approved_at, posted_at,
            reversed_at, reversal_of_id, version;

-- name: CreateJournalLine :one
INSERT INTO accounting.journal_lines (
    journal_entry_id, legal_entity_id, line_number, account_id, description,
    currency_code, exchange_rate, transaction_debit, transaction_credit,
    base_debit, base_credit, branch_id, department_id, cost_center_id
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13, $14
) RETURNING id, journal_entry_id, legal_entity_id, line_number, account_id, description,
            currency_code, exchange_rate, transaction_debit, transaction_credit,
            base_debit, base_credit, branch_id, department_id, cost_center_id;

-- name: ListJournalLinesByEntryID :many
SELECT l.id, l.journal_entry_id, l.legal_entity_id, l.line_number, l.account_id, l.description,
       l.currency_code, l.exchange_rate, l.transaction_debit, l.transaction_credit,
       l.base_debit, l.base_credit, l.branch_id, l.department_id, l.cost_center_id,
       a.code as account_code, a.name_en as account_name_en, a.name_ar as account_name_ar,
       a.normal_balance as account_normal_balance, a.account_type
FROM accounting.journal_lines l
JOIN accounting.accounts a ON a.id = l.account_id
WHERE l.journal_entry_id = $1
ORDER BY l.line_number ASC;

-- name: DeleteJournalLinesByEntryID :exec
DELETE FROM accounting.journal_lines
WHERE journal_entry_id = $1;
