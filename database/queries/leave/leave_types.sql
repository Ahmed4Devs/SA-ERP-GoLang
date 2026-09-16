-- name: CreateLeaveType :one
INSERT INTO leave.types (
    id,
    legal_entity_id,
    code,
    name_en,
    name_ar,
    is_paid,
    requires_attachment,
    deducts_from_annual,
    max_days_per_year,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetLeaveTypeByID :one
SELECT * FROM leave.types
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetLeaveTypeByCode :one
SELECT * FROM leave.types
WHERE legal_entity_id = $1 AND code = $2;

-- name: ListLeaveTypes :many
SELECT * FROM leave.types
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: CreateLeavePolicy :one
INSERT INTO leave.policies (
    id,
    legal_entity_id,
    leave_type_id,
    accrual_frequency,
    accrual_rate_per_period,
    max_carryover_days,
    probation_restricted,
    encashment_allowed,
    min_service_days_required,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetLeavePolicyByType :one
SELECT * FROM leave.policies
WHERE legal_entity_id = $1 AND leave_type_id = $2;

-- name: ListLeavePolicies :many
SELECT p.*, t.code AS leave_type_code, t.name_en AS leave_type_name_en, t.name_ar AS leave_type_name_ar
FROM leave.policies p
JOIN leave.types t ON p.leave_type_id = t.id
WHERE p.legal_entity_id = $1
ORDER BY t.code ASC;
