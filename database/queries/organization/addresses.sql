-- name: CreateAddress :one
INSERT INTO organization.addresses (
    legal_entity_id,
    branch_id,
    address_type,
    building_number,
    street_name_ar,
    street_name_en,
    district_ar,
    district_en,
    city_ar,
    city_en,
    postal_code,
    additional_number,
    country_code
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetAddressByID :one
SELECT * FROM organization.addresses
WHERE id = $1;

-- name: ListAddressesByLegalEntity :many
SELECT * FROM organization.addresses
WHERE legal_entity_id = $1
ORDER BY created_at ASC;

-- name: ListAddressesByBranch :many
SELECT * FROM organization.addresses
WHERE branch_id = $1
ORDER BY created_at ASC;
