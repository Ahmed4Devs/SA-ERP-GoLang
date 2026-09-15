-- name: ListFiscalYearsByLegalEntity :many
SELECT id, legal_entity_id, code, name_en, name_ar, start_date, end_date, status, created_at, updated_at, version
FROM accounting.fiscal_years
WHERE legal_entity_id = $1
ORDER BY start_date DESC;

-- name: GetFiscalYearByID :one
SELECT id, legal_entity_id, code, name_en, name_ar, start_date, end_date, status, created_at, updated_at, version
FROM accounting.fiscal_years
WHERE id = $1;

-- name: GetFiscalYearByDate :one
SELECT id, legal_entity_id, code, name_en, name_ar, start_date, end_date, status, created_at, updated_at, version
FROM accounting.fiscal_years
WHERE legal_entity_id = $1
  AND $2::DATE >= start_date AND $2::DATE <= end_date
LIMIT 1;

-- name: ListFiscalPeriodsByYear :many
SELECT id, fiscal_year_id, legal_entity_id, period_number, period_name_en, period_name_ar, quarter, start_date, end_date, status, created_at, updated_at, version
FROM accounting.fiscal_periods
WHERE fiscal_year_id = $1
ORDER BY period_number ASC;

-- name: GetFiscalPeriodByID :one
SELECT id, fiscal_year_id, legal_entity_id, period_number, period_name_en, period_name_ar, quarter, start_date, end_date, status, created_at, updated_at, version
FROM accounting.fiscal_periods
WHERE id = $1;

-- name: GetFiscalPeriodByDate :one
SELECT id, fiscal_year_id, legal_entity_id, period_number, period_name_en, period_name_ar, quarter, start_date, end_date, status, created_at, updated_at, version
FROM accounting.fiscal_periods
WHERE legal_entity_id = $1
  AND $2::DATE >= start_date AND $2::DATE <= end_date
LIMIT 1;

-- name: UpdateFiscalPeriodStatus :one
UPDATE accounting.fiscal_periods
SET status = $2,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING id, fiscal_year_id, legal_entity_id, period_number, period_name_en, period_name_ar, quarter, start_date, end_date, status, created_at, updated_at, version;
