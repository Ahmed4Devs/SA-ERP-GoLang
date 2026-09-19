-- name: CreateContributionCalculation :one
INSERT INTO gosi.contribution_calculations (
    id, legal_entity_id, employee_id, profile_id, calculation_year, calculation_month,
    contribution_basis, total_employee_amount, total_employer_amount, total_amount,
    calculation_snapshot, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
ON CONFLICT (legal_entity_id, employee_id, calculation_year, calculation_month)
DO UPDATE SET
    contribution_basis = EXCLUDED.contribution_basis,
    total_employee_amount = EXCLUDED.total_employee_amount,
    total_employer_amount = EXCLUDED.total_employer_amount,
    total_amount = EXCLUDED.total_amount,
    calculation_snapshot = EXCLUDED.calculation_snapshot,
    status = EXCLUDED.status
RETURNING *;

-- name: CreateGosiCalculationLine :one
INSERT INTO gosi.contribution_calculation_lines (
    id, calculation_id, rule_version_id, rule_code, rule_name_en, rule_name_ar,
    basis_amount, employee_rate, employer_rate, employee_amount, employer_amount, formula_note
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
RETURNING *;

-- name: DeleteGosiCalculationLines :exec
DELETE FROM gosi.contribution_calculation_lines
WHERE calculation_id = $1;

-- name: GetGosiContributionCalculation :one
SELECT * FROM gosi.contribution_calculations
WHERE id = $1 AND legal_entity_id = $2;

-- name: ListGosiCalculationsByPeriod :many
SELECT c.*, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar, e.employee_number
FROM gosi.contribution_calculations c
JOIN hr.employees e ON e.id = c.employee_id
WHERE c.legal_entity_id = $1 AND c.calculation_year = $2 AND c.calculation_month = $3
ORDER BY e.employee_number;

-- name: LockGosiCalculationsByPeriod :many
SELECT *
FROM gosi.contribution_calculations
WHERE legal_entity_id = $1 AND calculation_year = $2 AND calculation_month = $3
ORDER BY employee_id
FOR UPDATE;

-- name: ListGosiCalculationLines :many
SELECT * FROM gosi.contribution_calculation_lines
WHERE calculation_id = $1
ORDER BY rule_code;

-- name: CreateExportBatch :one
INSERT INTO gosi.export_batches (
    id, legal_entity_id, batch_number, export_year, export_month,
    employee_count, total_employee_amount, total_employer_amount, total_amount,
    file_content, file_format, status, created_by
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetExportBatchByID :one
SELECT * FROM gosi.export_batches WHERE id = $1;

-- name: ListExportBatches :many
SELECT * FROM gosi.export_batches
WHERE legal_entity_id = $1
ORDER BY created_at DESC;

-- name: UpdateExportBatchStatus :one
UPDATE gosi.export_batches
SET status = @status::VARCHAR,
    submitted_at = CASE WHEN @status::VARCHAR = 'SUBMITTED' THEN CURRENT_TIMESTAMP ELSE submitted_at END,
    confirmed_at = CASE WHEN @status::VARCHAR = 'CONFIRMED' THEN CURRENT_TIMESTAMP ELSE confirmed_at END
WHERE id = @id
RETURNING *;

-- name: CreateIntegrationAttempt :one
INSERT INTO gosi.integration_attempts (
    id, legal_entity_id, batch_id, attempt_number, adapter_type,
    request_payload, response_payload, status, error_message
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListIntegrationAttempts :many
SELECT * FROM gosi.integration_attempts
WHERE batch_id = $1
ORDER BY attempt_number;

-- name: LinkCalculationsToExportBatch :execrows
UPDATE gosi.contribution_calculations
SET export_batch_id = $1, status = 'EXPORTED'
WHERE legal_entity_id = $2 AND calculation_year = $3 AND calculation_month = $4 AND status = 'CALCULATED';
