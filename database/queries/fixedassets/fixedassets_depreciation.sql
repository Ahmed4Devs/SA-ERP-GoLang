-- name: CreateDepreciationSchedule :one
INSERT INTO fixedassets.depreciation_schedules (
    id, asset_id, book_id, period_year, period_month,
    opening_net_book_value, depreciation_amount, closing_net_book_value, accumulated_depreciation,
    is_posted, depreciation_run_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (asset_id, book_id, period_year, period_month)
DO UPDATE SET
    opening_net_book_value = EXCLUDED.opening_net_book_value,
    depreciation_amount = EXCLUDED.depreciation_amount,
    closing_net_book_value = EXCLUDED.closing_net_book_value,
    accumulated_depreciation = EXCLUDED.accumulated_depreciation
RETURNING *;

-- name: ListDepreciationSchedulesByAsset :many
SELECT * FROM fixedassets.depreciation_schedules
WHERE asset_id = $1 AND book_id = $2
ORDER BY period_year, period_month;

-- name: MarkSchedulePosted :exec
UPDATE fixedassets.depreciation_schedules
SET is_posted = true,
    depreciation_run_id = $3
WHERE asset_id = $1 AND period_year = $4 AND period_month = $5 AND book_id = $2;

-- name: CreateDepreciationRun :one
INSERT INTO fixedassets.depreciation_runs (
    id, legal_entity_id, book_id, run_number, period_year, period_month,
    run_date, total_assets_processed, total_depreciation_amount, status,
    journal_entry_id, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (legal_entity_id, book_id, period_year, period_month)
DO UPDATE SET
    total_assets_processed = EXCLUDED.total_assets_processed,
    total_depreciation_amount = EXCLUDED.total_depreciation_amount,
    status = EXCLUDED.status,
    updated_at = CURRENT_TIMESTAMP
RETURNING *;

-- name: GetDepreciationRunByID :one
SELECT r.*, b.code as book_code, b.name_en as book_name_en, b.name_ar as book_name_ar
FROM fixedassets.depreciation_runs r
JOIN fixedassets.depreciation_books b ON b.id = r.book_id
WHERE r.id = $1;

-- name: ListDepreciationRuns :many
SELECT r.*, b.code as book_code, b.name_en as book_name_en, b.name_ar as book_name_ar
FROM fixedassets.depreciation_runs r
JOIN fixedassets.depreciation_books b ON b.id = r.book_id
WHERE r.legal_entity_id = $1
ORDER BY r.period_year DESC, r.period_month DESC;

-- name: UpdateDepreciationRunStatus :one
UPDATE fixedassets.depreciation_runs
SET status = @status::VARCHAR,
    journal_entry_id = COALESCE(@journal_entry_id, journal_entry_id),
    approved_by = COALESCE(@approved_by, approved_by),
    updated_at = CURRENT_TIMESTAMP
WHERE id = @id
RETURNING *;

-- name: CreateDepreciationLine :one
INSERT INTO fixedassets.depreciation_lines (
    id, run_id, asset_id, asset_number, cost_basis,
    depreciation_amount, accumulated_before, accumulated_after, net_book_value_after,
    expense_account_id, accumulated_account_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (run_id, asset_id)
DO UPDATE SET
    depreciation_amount = EXCLUDED.depreciation_amount,
    accumulated_after = EXCLUDED.accumulated_after,
    net_book_value_after = EXCLUDED.net_book_value_after
RETURNING *;

-- name: ListDepreciationLinesByRun :many
SELECT l.*, a.name_en as asset_name_en, a.name_ar as asset_name_ar
FROM fixedassets.depreciation_lines l
JOIN fixedassets.assets a ON a.id = l.asset_id
WHERE l.run_id = $1
ORDER BY l.asset_number;

-- name: CreateAssetTransfer :one
INSERT INTO fixedassets.transfers (
    id, legal_entity_id, asset_id, transfer_number, transfer_date,
    from_location_id, to_location_id, from_employee_id, to_employee_id,
    reason, status, transferred_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: ListAssetTransfers :many
SELECT t.*,
       fl.code as from_location_code, fl.name_en as from_location_name_en,
       tl.code as to_location_code, tl.name_en as to_location_name_en,
       fe.first_name_en as from_emp_first_en, fe.last_name_en as from_emp_last_en,
       te.first_name_en as to_emp_first_en, te.last_name_en as to_emp_last_en
FROM fixedassets.transfers t
LEFT JOIN fixedassets.asset_locations fl ON fl.id = t.from_location_id
JOIN fixedassets.asset_locations tl ON tl.id = t.to_location_id
LEFT JOIN hr.employees fe ON fe.id = t.from_employee_id
LEFT JOIN hr.employees te ON te.id = t.to_employee_id
WHERE t.legal_entity_id = $1
ORDER BY t.transfer_date DESC;

-- name: CreateAssetImpairment :one
INSERT INTO fixedassets.impairments (
    id, legal_entity_id, asset_id, impairment_number, impairment_date,
    carrying_amount_before, recoverable_amount, impairment_loss,
    reason, journal_entry_id, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: CreateAssetDisposal :one
INSERT INTO fixedassets.disposals (
    id, legal_entity_id, asset_id, disposal_number, disposal_date,
    disposal_type, proceeds_amount, cost_at_disposal,
    accumulated_depreciation_at_disposal, net_book_value_at_disposal, gain_or_loss_amount,
    journal_entry_id, customer_id, reason, status, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
RETURNING *;

-- name: GetAssetDisposalByID :one
SELECT * FROM fixedassets.disposals WHERE id = $1;
