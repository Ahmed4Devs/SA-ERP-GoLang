-- name: CreateStockReservation :one
INSERT INTO inventory.stock_reservations (
    legal_entity_id,
    item_id,
    warehouse_id,
    bin_id,
    quantity,
    reservation_status,
    demand_source_type,
    demand_source_id,
    expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ReleaseStockReservation :one
UPDATE inventory.stock_reservations
SET
    reservation_status = 'RELEASED',
    updated_at = now()
WHERE id = $1 AND reservation_status = 'ACTIVE'
RETURNING *;

-- name: FulfillStockReservation :one
UPDATE inventory.stock_reservations
SET
    reservation_status = 'FULFILLED',
    updated_at = now()
WHERE id = $1 AND reservation_status = 'ACTIVE'
RETURNING *;

-- name: GetReservationByID :one
SELECT
    id,
    legal_entity_id,
    item_id,
    warehouse_id,
    bin_id,
    quantity,
    reservation_status,
    demand_source_type,
    demand_source_id,
    expires_at,
    created_at,
    updated_at
FROM inventory.stock_reservations
WHERE id = $1;

-- name: ListActiveReservationsByItem :many
SELECT
    id,
    legal_entity_id,
    item_id,
    warehouse_id,
    bin_id,
    quantity,
    reservation_status,
    demand_source_type,
    demand_source_id,
    expires_at,
    created_at,
    updated_at
FROM inventory.stock_reservations
WHERE legal_entity_id = $1 AND item_id = $2 AND reservation_status = 'ACTIVE'
ORDER BY created_at ASC;
