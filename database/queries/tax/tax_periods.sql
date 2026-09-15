-- name: ListTaxPeriodsByEntity :many
SELECT
    id,
    legal_entity_id,
    period_number,
    fiscal_year,
    period_type,
    start_date,
    end_date,
    due_date,
    status,
    created_at,
    updated_at
FROM tax.tax_periods
WHERE legal_entity_id = $1
ORDER BY fiscal_year DESC, period_number DESC;

-- name: GetTaxPeriodByID :one
SELECT
    id,
    legal_entity_id,
    period_number,
    fiscal_year,
    period_type,
    start_date,
    end_date,
    due_date,
    status,
    created_at,
    updated_at
FROM tax.tax_periods
WHERE id = $1;

-- name: GetOpenTaxPeriodForDate :one
SELECT
    id,
    legal_entity_id,
    period_number,
    fiscal_year,
    period_type,
    start_date,
    end_date,
    due_date,
    status,
    created_at,
    updated_at
FROM tax.tax_periods
WHERE legal_entity_id = $1
  AND start_date <= $2
  AND end_date >= $2
LIMIT 1;

-- name: CreateTaxPeriod :one
INSERT INTO tax.tax_periods (
    legal_entity_id,
    period_number,
    fiscal_year,
    period_type,
    start_date,
    end_date,
    due_date,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: UpdateTaxPeriodStatus :one
UPDATE tax.tax_periods
SET
    status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;
