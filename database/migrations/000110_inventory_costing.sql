-- +goose Up
-- SQL in this section is executed when the migration is applied.

-- 1. Add costing fields to products.items
ALTER TABLE products.items
ADD COLUMN IF NOT EXISTS costing_method TEXT NOT NULL DEFAULT 'WEIGHTED_AVERAGE'
    CHECK (costing_method IN ('WEIGHTED_AVERAGE', 'FIFO', 'STANDARD')),
ADD COLUMN IF NOT EXISTS standard_cost NUMERIC(18, 4) NOT NULL DEFAULT 0.0000;

-- 2. FIFO Cost Layers Table (Tracks discrete inventory lots / receipt costs)
CREATE TABLE IF NOT EXISTS inventory.cost_layers (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE RESTRICT,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE RESTRICT,
    source_movement_id  UUID NOT NULL REFERENCES inventory.stock_movements(id) ON DELETE RESTRICT,
    original_quantity   NUMERIC(18, 4) NOT NULL CHECK (original_quantity > 0),
    remaining_quantity  NUMERIC(18, 4) NOT NULL CHECK (remaining_quantity >= 0),
    unit_cost           NUMERIC(18, 4) NOT NULL CHECK (unit_cost >= 0),
    total_cost          NUMERIC(18, 4) NOT NULL CHECK (total_cost >= 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_cost_layers_active
ON inventory.cost_layers (legal_entity_id, item_id, warehouse_id, created_at ASC)
WHERE remaining_quantity > 0;

CREATE INDEX IF NOT EXISTS idx_cost_layers_movement
ON inventory.cost_layers (source_movement_id);

-- 3. Cost Layer Consumptions (Tracks which FIFO layers were consumed by each issue movement)
CREATE TABLE IF NOT EXISTS inventory.cost_layer_consumptions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    cost_layer_id       UUID NOT NULL REFERENCES inventory.cost_layers(id) ON DELETE RESTRICT,
    issue_movement_id   UUID NOT NULL REFERENCES inventory.stock_movements(id) ON DELETE RESTRICT,
    quantity_consumed   NUMERIC(18, 4) NOT NULL CHECK (quantity_consumed > 0),
    unit_cost           NUMERIC(18, 4) NOT NULL CHECK (unit_cost >= 0),
    total_cost          NUMERIC(18, 4) NOT NULL CHECK (total_cost >= 0),
    consumed_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_layer_consumptions_issue
ON inventory.cost_layer_consumptions (issue_movement_id);

CREATE INDEX IF NOT EXISTS idx_layer_consumptions_layer
ON inventory.cost_layer_consumptions (cost_layer_id);

-- 4. Inventory Revaluations Table
CREATE TABLE IF NOT EXISTS inventory.inventory_revaluations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    revaluation_number  TEXT NOT NULL,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE RESTRICT,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE RESTRICT,
    quantity_on_hand    NUMERIC(18, 4) NOT NULL CHECK (quantity_on_hand > 0),
    old_unit_cost       NUMERIC(18, 4) NOT NULL CHECK (old_unit_cost >= 0),
    new_unit_cost       NUMERIC(18, 4) NOT NULL CHECK (new_unit_cost >= 0),
    old_total_value     NUMERIC(18, 4) NOT NULL,
    new_total_value     NUMERIC(18, 4) NOT NULL,
    variance_amount     NUMERIC(18, 4) NOT NULL,
    reason              TEXT NOT NULL,
    journal_entry_id    UUID REFERENCES accounting.journal_entries(id) ON DELETE SET NULL,
    created_by          UUID NOT NULL REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_inventory_revaluation_number
        UNIQUE (legal_entity_id, revaluation_number)
);

CREATE INDEX IF NOT EXISTS idx_inventory_revaluations_entity
ON inventory.inventory_revaluations (legal_entity_id, warehouse_id, item_id);

-- Permissions for Costing
INSERT INTO authz.permissions (id, code, name_en, name_ar, module)
VALUES
    (uuidv7(), 'inventory:costing:read', 'View Inventory Costing & Valuation', 'عرض تقييم وتكاليف المخزون', 'inventory'),
    (uuidv7(), 'inventory:revaluation:create', 'Create Inventory Revaluation', 'إنشاء إعادة تقييم للمخزون', 'inventory')
ON CONFLICT (code) DO NOTHING;

-- Grant permissions to Admin role
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code IN ('admin', 'general_accountant', 'finance_manager', 'inventory_manager')
  AND p.code IN ('inventory:costing:read', 'inventory:revaluation:create')
ON CONFLICT DO NOTHING;

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.

DROP TABLE IF EXISTS inventory.inventory_revaluations CASCADE;
DROP TABLE IF EXISTS inventory.cost_layer_consumptions CASCADE;
DROP TABLE IF EXISTS inventory.cost_layers CASCADE;

ALTER TABLE products.items
DROP COLUMN IF EXISTS costing_method,
DROP COLUMN IF EXISTS standard_cost;
