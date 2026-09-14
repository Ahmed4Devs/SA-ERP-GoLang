-- name: CreateOrganization :one
INSERT INTO organization.organizations (
    code,
    name_ar,
    name_en,
    status
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organization.organizations
WHERE id = $1;

-- name: GetOrganizationByCode :one
SELECT * FROM organization.organizations
WHERE code = $1;

-- name: ListOrganizations :many
SELECT * FROM organization.organizations
ORDER BY created_at DESC;

-- name: UpdateOrganization :one
UPDATE organization.organizations
SET
    name_ar = $2,
    name_en = $3,
    status = $4,
    updated_at = now(),
    version = version + 1
WHERE id = $1
RETURNING *;
