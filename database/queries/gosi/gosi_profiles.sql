-- name: CreateInsuranceProfile :one
INSERT INTO gosi.employee_insurance_profiles (
    id, legal_entity_id, employee_id, gosi_number, registration_date,
    nationality_scope, is_active, contribution_basis_override, notes
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetInsuranceProfileByEmployee :one
SELECT * FROM gosi.employee_insurance_profiles
WHERE legal_entity_id = $1 AND employee_id = $2;

-- name: ListInsuranceProfiles :many
SELECT p.*, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar, e.employee_number
FROM gosi.employee_insurance_profiles p
JOIN hr.employees e ON e.id = p.employee_id
WHERE p.legal_entity_id = $1
ORDER BY e.employee_number;

-- name: UpdateInsuranceProfile :one
UPDATE gosi.employee_insurance_profiles
SET gosi_number = COALESCE($2, gosi_number),
    is_active = COALESCE($3, is_active),
    contribution_basis_override = $4,
    deregistration_date = $5,
    notes = COALESCE($6, notes),
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: ListActiveInsuranceProfiles :many
SELECT p.*, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar,
       e.employee_number, e.nationality_country_code
FROM gosi.employee_insurance_profiles p
JOIN hr.employees e ON e.id = p.employee_id
WHERE p.legal_entity_id = $1 AND p.is_active = true
ORDER BY e.employee_number;
