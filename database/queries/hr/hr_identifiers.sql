-- name: CreateEmployeeIdentifier :one
INSERT INTO hr.employee_identifiers (
    id, employee_id, identifier_type, identifier_value,
    identifier_value_masked, issue_date, expiry_date,
    issue_place, is_primary, is_verified, created_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10, now()
)
RETURNING *;

-- name: ListEmployeeIdentifiers :many
SELECT * FROM hr.employee_identifiers
WHERE employee_id = $1
ORDER BY is_primary DESC, created_at ASC;

-- name: CreateEmployeeContact :one
INSERT INTO hr.employee_contacts (
    id, employee_id, work_email, personal_email, mobile_phone,
    work_phone, extension, building_no, street_name, district,
    city, postal_code, additional_no, country_code, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9, $10,
    $11, $12, $13, $14, now(), now()
)
RETURNING *;

-- name: GetEmployeeContact :one
SELECT * FROM hr.employee_contacts
WHERE employee_id = $1;
