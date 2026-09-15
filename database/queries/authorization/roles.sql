-- name: GetRoleByCode :one
SELECT * FROM authz.roles
WHERE code = $1;

-- name: GetRoleByID :one
SELECT * FROM authz.roles
WHERE id = $1;

-- name: ListRoles :many
SELECT * FROM authz.roles
ORDER BY created_at ASC;

-- name: ListPermissions :many
SELECT * FROM authz.permissions
ORDER BY module ASC, code ASC;

-- name: ListPermissionsForRole :many
SELECT p.* FROM authz.permissions p
JOIN authz.role_permissions rp ON rp.permission_id = p.id
WHERE rp.role_id = $1
ORDER BY p.code ASC;

-- name: CreateRole :one
INSERT INTO authz.roles (
    code,
    name_ar,
    name_en,
    description,
    is_system
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: AssignPermissionToRole :exec
INSERT INTO authz.role_permissions (
    role_id,
    permission_id
) VALUES (
    $1, $2
)
ON CONFLICT DO NOTHING;
