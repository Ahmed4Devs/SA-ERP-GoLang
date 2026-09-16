-- name: CreateEmployee :one
INSERT INTO hr.employees (
    id, legal_entity_id, employee_number, user_id,
    first_name_en, last_name_en, first_name_ar, last_name_ar,
    gender, marital_status, date_of_birth, nationality_country_code,
    is_saudi_national, hire_date, probation_end_date, termination_date,
    employment_status, primary_position_id, department_id, branch_id,
    manager_employee_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10, $11, $12,
    $13, $14, $15, $16,
    $17, $18, $19, $20,
    $21, now(), now()
)
RETURNING *;

-- name: GetEmployeeByID :one
SELECT * FROM hr.employees
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetEmployeeByNumber :one
SELECT * FROM hr.employees
WHERE legal_entity_id = $1 AND employee_number = $2;

-- name: ListEmployees :many
SELECT * FROM hr.employees
WHERE legal_entity_id = $1
ORDER BY employee_number ASC;

-- name: UpdateEmployeeStatus :one
UPDATE hr.employees
SET employment_status = $3,
    termination_date = COALESCE($4, termination_date),
    updated_at = now()
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: CountEmployeesByNationality :many
SELECT is_saudi_national, COUNT(*) as count
FROM hr.employees
WHERE legal_entity_id = $1 AND employment_status != 'TERMINATED'
GROUP BY is_saudi_national;
