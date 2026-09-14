-- name: CreateCostCenter :one
INSERT INTO organization.cost_centers (
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

-- name: GetCostCenterByID :one
SELECT * FROM organization.cost_centers
WHERE id = $1;

-- name: ListCostCentersByEntity :many
SELECT * FROM organization.cost_centers
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: UpdateCostCenter :one
UPDATE organization.cost_centers
SET
    parent_id = $2,
    name_ar = $3,
    name_en = $4,
    is_active = $5,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;
