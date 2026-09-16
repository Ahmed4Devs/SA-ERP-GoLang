-- name: CreatePurchaseOrder :one
INSERT INTO purchasing.purchase_orders (
    id,
    legal_entity_id,
    po_number,
    purchase_request_id,
    rfq_id,
    supplier_quote_id,
    supplier_id,
    order_date,
    expected_delivery_date,
    payment_terms_id,
    delivery_warehouse_id,
    currency_code,
    exchange_rate,
    subtotal,
    discount_amount,
    taxable_amount,
    tax_amount,
    total_amount,
    status,
    notes,
    created_by_user_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
) RETURNING *;

-- name: CreatePurchaseOrderLine :one
INSERT INTO purchasing.purchase_order_lines (
    id,
    purchase_order_id,
    line_number,
    pr_line_id,
    item_id,
    description,
    uom,
    quantity,
    unit_price,
    discount_percent,
    discount_amount,
    taxable_amount,
    tax_code,
    tax_rate_percent,
    tax_amount,
    total_amount,
    expense_account_id
) VALUES (
    COALESCE(sqlc.narg('id'), uuidv7()),
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
) RETURNING *;

-- name: GetPurchaseOrder :one
SELECT * FROM purchasing.purchase_orders
WHERE id = $1;

-- name: GetPurchaseOrderByNumber :one
SELECT * FROM purchasing.purchase_orders
WHERE legal_entity_id = $1 AND po_number = $2;

-- name: ListPurchaseOrders :many
SELECT * FROM purchasing.purchase_orders
WHERE legal_entity_id = $1
  AND (sqlc.narg('supplier_id')::UUID IS NULL OR supplier_id = sqlc.narg('supplier_id'))
  AND (sqlc.narg('status')::TEXT IS NULL OR status = sqlc.narg('status'))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit_count') OFFSET sqlc.arg('offset_count');

-- name: ListPurchaseOrderLines :many
SELECT * FROM purchasing.purchase_order_lines
WHERE purchase_order_id = $1
ORDER BY line_number ASC;

-- name: UpdatePurchaseOrderStatus :one
UPDATE purchasing.purchase_orders
SET status = $2,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: ApprovePurchaseOrder :one
UPDATE purchasing.purchase_orders
SET status = 'APPROVED',
    approved_by_user_id = $2,
    approved_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: IssuePurchaseOrder :one
UPDATE purchasing.purchase_orders
SET status = 'ISSUED',
    issued_at = now(),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdatePOLineReceivedQty :one
UPDATE purchasing.purchase_order_lines
SET received_quantity = received_quantity + $2
WHERE id = $1
RETURNING *;

-- name: UpdatePOLineBilledQty :one
UPDATE purchasing.purchase_order_lines
SET billed_quantity = billed_quantity + $2
WHERE id = $1
RETURNING *;
