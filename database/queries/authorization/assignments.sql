-- name: AssignUserRole :one
INSERT INTO authz.user_role_assignments (
    user_id,
    role_id,
    legal_entity_id,
    branch_id
) VALUES (
    $1, $2, $3, $4
)
ON CONFLICT (user_id, role_id, legal_entity_id, branch_id) DO UPDATE
SET assigned_at = now()
RETURNING *;

-- name: RevokeUserRole :exec
DELETE FROM authz.user_role_assignments
WHERE user_id = $1 AND role_id = $2
  AND legal_entity_id IS NOT DISTINCT FROM $3
  AND branch_id IS NOT DISTINCT FROM $4;

-- name: ListUserAssignments :many
SELECT
    ura.id,
    ura.user_id,
    ura.role_id,
    r.code AS role_code,
    r.name_ar AS role_name_ar,
    r.name_en AS role_name_en,
    ura.legal_entity_id,
    le.name_ar AS legal_entity_name_ar,
    le.name_en AS legal_entity_name_en,
    ura.branch_id,
    b.name_ar AS branch_name_ar,
    b.name_en AS branch_name_en,
    ura.assigned_at
FROM authz.user_role_assignments ura
JOIN authz.roles r ON r.id = ura.role_id
LEFT JOIN organization.legal_entities le ON le.id = ura.legal_entity_id
LEFT JOIN organization.branches b ON b.id = ura.branch_id
WHERE ura.user_id = $1
ORDER BY ura.assigned_at DESC;

-- name: GetUserPermissionsInScope :many
SELECT DISTINCT p.code
FROM authz.user_role_assignments ura
JOIN authz.role_permissions rp ON rp.role_id = ura.role_id
JOIN authz.permissions p ON p.id = rp.permission_id
WHERE ura.user_id = $1
  AND ura.branch_id IS NULL
  AND (
      ura.legal_entity_id IS NULL -- Global Scope (Super Admin)
      OR ura.legal_entity_id = $2  -- Exact Legal Entity Match
  );

-- name: GetUserPermissionsInBranch :many
SELECT DISTINCT p.code
FROM authz.user_role_assignments ura
JOIN authz.role_permissions rp ON rp.role_id = ura.role_id
JOIN authz.permissions p ON p.id = rp.permission_id
WHERE ura.user_id = sqlc.arg(user_id)
  AND (ura.legal_entity_id IS NULL OR ura.legal_entity_id = sqlc.arg(legal_entity_id))
  AND (ura.branch_id IS NULL OR ura.branch_id = sqlc.arg(branch_id))
  AND EXISTS (SELECT 1 FROM organization.branches b WHERE b.id = sqlc.arg(branch_id) AND b.legal_entity_id = sqlc.arg(legal_entity_id));

-- name: IsBranchInEntity :one
SELECT EXISTS (SELECT 1 FROM organization.branches WHERE id = sqlc.arg(branch_id) AND legal_entity_id = sqlc.arg(legal_entity_id))::boolean;

-- name: WouldRevokeLastAdministrator :one
SELECT (EXISTS (
    SELECT 1 FROM authz.user_role_assignments a
    JOIN iam.users u ON u.id = a.user_id
    JOIN authz.role_permissions rp ON rp.role_id = a.role_id
    JOIN authz.permissions p ON p.id = rp.permission_id
    WHERE a.user_id = sqlc.arg(user_id) AND a.role_id = sqlc.arg(role_id)
      AND a.legal_entity_id IS NULL AND a.branch_id IS NULL AND u.status = 'ACTIVE' AND p.code IN ('*', 'authz.*', 'authz.manage')
) AND NOT EXISTS (
    SELECT 1 FROM authz.user_role_assignments a
    JOIN iam.users u ON u.id = a.user_id
    JOIN authz.role_permissions rp ON rp.role_id = a.role_id
    JOIN authz.permissions p ON p.id = rp.permission_id
    WHERE (a.user_id <> sqlc.arg(user_id) OR a.role_id <> sqlc.arg(role_id))
      AND a.legal_entity_id IS NULL AND a.branch_id IS NULL AND u.status = 'ACTIVE' AND p.code IN ('*', 'authz.*', 'authz.manage')
))::boolean AS blocked;
