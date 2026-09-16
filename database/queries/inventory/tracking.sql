-- name: GetLot :one
SELECT
    id,
    legal_entity_id,
    item_id,
    lot_number,
    manufacture_date,
    expiry_date,
    is_blocked,
    created_at
FROM inventory.lots
WHERE legal_entity_id = $1 AND item_id = $2 AND lot_number = $3;

-- name: CreateLot :one
INSERT INTO inventory.lots (
    legal_entity_id,
    item_id,
    lot_number,
    manufacture_date,
    expiry_date,
    is_blocked
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (legal_entity_id, item_id, lot_number)
DO UPDATE SET
    manufacture_date = EXCLUDED.manufacture_date,
    expiry_date = EXCLUDED.expiry_date,
    is_blocked = EXCLUDED.is_blocked
RETURNING *;

-- name: ListLotsByItem :many
SELECT
    id,
    legal_entity_id,
    item_id,
    lot_number,
    manufacture_date,
    expiry_date,
    is_blocked,
    created_at
FROM inventory.lots
WHERE legal_entity_id = $1 AND item_id = $2
ORDER BY expiry_date ASC NULLS LAST;

-- name: GetSerialNumber :one
SELECT
    id,
    legal_entity_id,
    item_id,
    serial_number,
    warehouse_id,
    bin_id,
    status,
    created_at,
    updated_at
FROM inventory.serial_numbers
WHERE legal_entity_id = $1 AND item_id = $2 AND serial_number = $3;

-- name: CreateSerialNumber :one
INSERT INTO inventory.serial_numbers (
    legal_entity_id,
    item_id,
    serial_number,
    warehouse_id,
    bin_id,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6
)
ON CONFLICT (legal_entity_id, item_id, serial_number)
DO UPDATE SET
    warehouse_id = EXCLUDED.warehouse_id,
    bin_id = EXCLUDED.bin_id,
    status = EXCLUDED.status,
    updated_at = now()
RETURNING *;

-- name: UpdateSerialNumberStatus :one
UPDATE inventory.serial_numbers
SET
    status = $2,
    warehouse_id = $3,
    bin_id = $4,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ListSerialNumbersByItem :many
SELECT
    id,
    legal_entity_id,
    item_id,
    serial_number,
    warehouse_id,
    bin_id,
    status,
    created_at,
    updated_at
FROM inventory.serial_numbers
WHERE legal_entity_id = $1 AND item_id = $2
ORDER BY serial_number ASC;
