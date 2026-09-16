-- name: CreateContract :one
INSERT INTO hr.contracts (
    id, legal_entity_id, employee_id, contract_number, contract_type,
    start_date, end_date, probation_days, notice_period_days,
    working_hours_per_week, work_location, status, notes,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13,
    now(), now()
)
RETURNING *;

-- name: GetContractByID :one
SELECT * FROM hr.contracts
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetContractByNumber :one
SELECT * FROM hr.contracts
WHERE legal_entity_id = $1 AND contract_number = $2;

-- name: GetContractByEmployeeID :one
SELECT * FROM hr.contracts
WHERE employee_id = $1 AND legal_entity_id = $2
ORDER BY created_at DESC
LIMIT 1;

-- name: ListContracts :many
SELECT * FROM hr.contracts
WHERE legal_entity_id = $1
ORDER BY created_at DESC;

-- name: UpdateContractStatus :one
UPDATE hr.contracts
SET status = $3,
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: CreateContractVersion :one
INSERT INTO hr.contract_versions (
    id, contract_id, version_number, effective_from, effective_to,
    basic_salary, housing_allowance, transport_allowance, other_allowances,
    gross_salary, change_reason, approved_by_user_id, created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, now()
)
RETURNING *;

-- name: ListContractVersions :many
SELECT * FROM hr.contract_versions
WHERE contract_id = $1
ORDER BY version_number DESC;

-- name: GetLatestContractVersion :one
SELECT * FROM hr.contract_versions
WHERE contract_id = $1
ORDER BY version_number DESC
LIMIT 1;
