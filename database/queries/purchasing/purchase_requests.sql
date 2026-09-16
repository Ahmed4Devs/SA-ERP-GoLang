-- name: CreatePurchaseRequest :one
INSERT INTO purchasing.purchase_requests (
    id,
    legal_entity_id,
    pr_number,
    requested_by_user_id,
    department_id,
    request_date,
    required_by_date,
    status,
    purpose,
    notes,
    estimated_total,
    currency_code
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: CreatePurchaseRequestLine :one
INSERT INTO purchasing.purchase_request_lines (
    id,
    purchase_request_id,
    line_number,
    item_id,
    description,
    uom,
    quantity,
    estimated_unit_cost,
    estimated_total_cost,
    suggested_supplier_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetPurchaseRequest :one
SELECT * FROM purchasing.purchase_requests
WHERE id = $1;

-- name: ListPurchaseRequests :many
SELECT * FROM purchasing.purchase_requests
WHERE legal_entity_id = $1
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListPurchaseRequestLines :many
SELECT * FROM purchasing.purchase_request_lines
WHERE purchase_request_id = $1
ORDER BY line_number ASC;

-- name: UpdatePurchaseRequestStatus :one
UPDATE purchasing.purchase_requests
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdatePRLineOrderedQty :one
UPDATE purchasing.purchase_request_lines
SET ordered_quantity = ordered_quantity + $2
WHERE id = $1
RETURNING *;
