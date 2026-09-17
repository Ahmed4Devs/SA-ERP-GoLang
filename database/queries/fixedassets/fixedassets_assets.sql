-- name: CreateFixedAsset :one
INSERT INTO fixedassets.assets (
    id, legal_entity_id, category_id, location_id, asset_number, serial_number, barcode,
    name_en, name_ar, description_en, description_ar,
    acquisition_date, in_service_date, acquisition_cost, salvage_value, useful_life_months,
    depreciation_method, status, current_net_book_value, accumulated_depreciation,
    impairment_total, supplier_id, purchase_invoice_id
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)
RETURNING *;

-- name: GetFixedAssetByID :one
SELECT a.*, c.code as category_code, c.name_en as category_name_en, c.name_ar as category_name_ar,
       c.depreciation_expense_account_id, c.accumulated_depreciation_account_id, c.asset_account_id,
       l.code as location_code, l.name_en as location_name_en, l.name_ar as location_name_ar
FROM fixedassets.assets a
JOIN fixedassets.categories c ON c.id = a.category_id
LEFT JOIN fixedassets.asset_locations l ON l.id = a.location_id
WHERE a.id = $1;

-- name: GetFixedAssetByNumber :one
SELECT a.*, c.code as category_code, c.name_en as category_name_en, c.name_ar as category_name_ar,
       c.depreciation_expense_account_id, c.accumulated_depreciation_account_id, c.asset_account_id,
       l.code as location_code, l.name_en as location_name_en, l.name_ar as location_name_ar
FROM fixedassets.assets a
JOIN fixedassets.categories c ON c.id = a.category_id
LEFT JOIN fixedassets.asset_locations l ON l.id = a.location_id
WHERE a.legal_entity_id = $1 AND a.asset_number = $2;

-- name: ListFixedAssets :many
SELECT a.*, c.code as category_code, c.name_en as category_name_en, c.name_ar as category_name_ar,
       l.code as location_code, l.name_en as location_name_en, l.name_ar as location_name_ar
FROM fixedassets.assets a
JOIN fixedassets.categories c ON c.id = a.category_id
LEFT JOIN fixedassets.asset_locations l ON l.id = a.location_id
WHERE a.legal_entity_id = $1
ORDER BY a.asset_number;

-- name: ListActiveFixedAssets :many
SELECT a.*, c.code as category_code, c.name_en as category_name_en, c.name_ar as category_name_ar,
       c.depreciation_expense_account_id, c.accumulated_depreciation_account_id, c.asset_account_id,
       l.code as location_code, l.name_en as location_name_en, l.name_ar as location_name_ar
FROM fixedassets.assets a
JOIN fixedassets.categories c ON c.id = a.category_id
LEFT JOIN fixedassets.asset_locations l ON l.id = a.location_id
WHERE a.legal_entity_id = $1 AND a.status = 'ACTIVE'
ORDER BY a.asset_number;

-- name: UpdateFixedAssetStatus :one
UPDATE fixedassets.assets
SET status = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateFixedAssetBookValue :one
UPDATE fixedassets.assets
SET current_net_book_value = $2,
    accumulated_depreciation = $3,
    status = CASE 
        WHEN status IN ('DISPOSED', 'RETIRED') THEN status
        WHEN $2 <= salvage_value THEN 'FULLY_DEPRECIATED' 
        ELSE status 
    END,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: UpdateFixedAssetLocation :one
UPDATE fixedassets.assets
SET location_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1
RETURNING *;

-- name: CreateAssetAssignment :one
INSERT INTO fixedassets.asset_assignments (
    id, asset_id, assigned_to_type, employee_id, department_id,
    assigned_date, return_date, notes, is_current
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: ListAssignmentsByAsset :many
SELECT aa.*, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar, e.employee_number
FROM fixedassets.asset_assignments aa
LEFT JOIN hr.employees e ON e.id = aa.employee_id
WHERE aa.asset_id = $1
ORDER BY aa.assigned_date DESC;

-- name: EndAssetAssignment :exec
UPDATE fixedassets.asset_assignments
SET return_date = $2,
    is_current = false
WHERE asset_id = $1 AND is_current = true;
