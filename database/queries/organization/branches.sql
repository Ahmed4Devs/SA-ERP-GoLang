-- name: CreateBranch :one
INSERT INTO organization.branches (
    legal_entity_id,
    code,
    name_ar,
    name_en,
    is_active
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetBranchByID :one
SELECT * FROM organization.branches
WHERE id = $1;

-- name: GetBranchByEntityAndCode :one
SELECT * FROM organization.branches
WHERE legal_entity_id = $1 AND code = $2;

-- name: ListBranchesByEntity :many
SELECT * FROM organization.branches
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: UpdateBranch :one
UPDATE organization.branches
SET
    name_ar = $2,
    name_en = $3,
    is_active = $4,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;
