-- name: ListManagedUsers :many
SELECT * FROM iam.users
WHERE sqlc.arg(search)::text = '' OR strpos(lower(coalesce(username, '') || ' ' || coalesce(email, '') || ' ' || display_name), lower(sqlc.arg(search))) > 0
ORDER BY created_at DESC, id DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: LockManagedUser :one
SELECT * FROM iam.users WHERE id = $1 FOR UPDATE;

-- name: UpdateManagedUserState :one
UPDATE iam.users
SET status = sqlc.arg(status),
    disabled_at = CASE WHEN sqlc.arg(status) = 'DISABLED' THEN now() ELSE NULL END,
    updated_at = now(), version = version + 1
WHERE id = sqlc.arg(id) AND version = sqlc.arg(expected_version)
RETURNING *;

-- name: LockUserAdministration :exec
SELECT pg_advisory_xact_lock(197704, 1);

-- name: IsGlobalAdministrator :one
SELECT EXISTS (
    SELECT 1 FROM authz.user_role_assignments a
    JOIN authz.role_permissions rp ON rp.role_id = a.role_id
    JOIN authz.permissions p ON p.id = rp.permission_id
    WHERE a.user_id = $1 AND a.legal_entity_id IS NULL AND a.branch_id IS NULL AND p.code IN ('*', 'authz.*', 'authz.manage')
)::boolean;

-- name: CountOtherActiveAdministrators :one
SELECT count(DISTINCT u.id) FROM iam.users u
JOIN authz.user_role_assignments a ON a.user_id = u.id
JOIN authz.role_permissions rp ON rp.role_id = a.role_id
JOIN authz.permissions p ON p.id = rp.permission_id
WHERE u.id <> $1 AND u.status = 'ACTIVE'
  AND a.legal_entity_id IS NULL AND a.branch_id IS NULL AND p.code IN ('*', 'authz.*', 'authz.manage');
