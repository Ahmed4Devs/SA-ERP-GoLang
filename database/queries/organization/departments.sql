-- name: CreateDepartment :one
INSERT INTO organization.departments (
    legal_entity_id,
    parent_id,
    code,
    name_ar,
    name_en,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetDepartmentByID :one
SELECT * FROM organization.departments
WHERE id = $1;

-- name: ListDepartmentsByEntity :many
SELECT * FROM organization.departments
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: UpdateDepartment :one
UPDATE organization.departments
SET
    parent_id = $2,
    name_ar = $3,
    name_en = $4,
    is_active = $5,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;
