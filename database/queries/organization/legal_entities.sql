-- name: CreateLegalEntity :one
INSERT INTO organization.legal_entities (
    organization_id,
    code,
    name_ar,
    name_en,
    cr_number,
    base_currency,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetLegalEntityByID :one
SELECT * FROM organization.legal_entities
WHERE id = $1;

-- name: GetLegalEntityByOrgAndCode :one
SELECT * FROM organization.legal_entities
WHERE organization_id = $1 AND code = $2;

-- name: ListLegalEntitiesByOrg :many
SELECT * FROM organization.legal_entities
WHERE organization_id = $1
ORDER BY code ASC;

-- name: ListAllActiveLegalEntities :many
SELECT * FROM organization.legal_entities
WHERE is_active = true
ORDER BY code ASC;

-- name: UpdateLegalEntity :one
UPDATE organization.legal_entities
SET
    name_ar = $2,
    name_en = $3,
    cr_number = $4,
    base_currency = $5,
    is_active = $6,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;
