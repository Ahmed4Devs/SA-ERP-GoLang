-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS warehouse;
CREATE SCHEMA IF NOT EXISTS inventory;

-- ============================================================================
-- 1. WAREHOUSE SCHEMA
-- ============================================================================

-- 1.1 Warehouses
CREATE TABLE IF NOT EXISTS warehouse.warehouses (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code            TEXT NOT NULL,
    name_en         TEXT NOT NULL,
    name_ar         TEXT NOT NULL,
    warehouse_type  TEXT NOT NULL DEFAULT 'PHYSICAL' CHECK (warehouse_type IN ('PHYSICAL', 'VIRTUAL', 'TRANSIT', 'CONSIGNMENT')),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    address_line1   TEXT,
    city            TEXT,
    postal_code     TEXT,
    country_code    TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_warehouses_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_warehouses_entity
ON warehouse.warehouses (legal_entity_id, code);

-- 1.2 Bin Locations (Every warehouse MUST have at least one system bin)
CREATE TABLE IF NOT EXISTS warehouse.bin_locations (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    warehouse_id    UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE CASCADE,
    code            TEXT NOT NULL,
    aisle           TEXT,
    rack            TEXT,
    shelf           TEXT,
    bin             TEXT,
    is_default      BOOLEAN NOT NULL DEFAULT false,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_bin_locations_warehouse_code
        UNIQUE (warehouse_id, code)
);

CREATE INDEX IF NOT EXISTS idx_bin_locations_wh
ON warehouse.bin_locations (warehouse_id, code);

-- ============================================================================
-- 2. INVENTORY SCHEMA (IMMUTABLE LEDGER & PROJECTIONS)
-- ============================================================================

-- 2.1 Stock Movements (Source-of-truth Movement Ledger - IMMUTABLE)
CREATE TABLE IF NOT EXISTS inventory.stock_movements (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    movement_number     TEXT NOT NULL,
    movement_type       TEXT NOT NULL CHECK (movement_type IN (
                            'OPENING', 'PURCHASE_RECEIPT', 'SALES_ISSUE',
                            'CUSTOMER_RETURN', 'SUPPLIER_RETURN',
                            'TRANSFER_OUT', 'TRANSFER_IN',
                            'ADJUSTMENT_IN', 'ADJUSTMENT_OUT',
                            'PRODUCTION_IN', 'PRODUCTION_OUT', 'REVERSAL'
                        )),
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE RESTRICT,
    sku_id              UUID REFERENCES products.skus(id) ON DELETE RESTRICT,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE RESTRICT,
    bin_id              UUID NOT NULL REFERENCES warehouse.bin_locations(id) ON DELETE RESTRICT,
    lot_number          TEXT,
    serial_number       TEXT,
    quantity            NUMERIC(18, 4) NOT NULL, -- Positive for IN, Negative for OUT
    uom_id              UUID NOT NULL REFERENCES products.units_of_measure(id) ON DELETE RESTRICT,
    unit_cost           NUMERIC(18, 4) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18, 4) NOT NULL DEFAULT 0,
    source_doc_type     TEXT, -- 'INVOICE', 'PURCHASE_RECEIPT', 'TRANSFER', 'ADJUSTMENT', 'COUNT'
    source_doc_id       UUID,
    source_doc_line_id  UUID,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_by          UUID NOT NULL REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_stock_movements_entity_number
        UNIQUE (legal_entity_id, movement_number)
);

CREATE INDEX IF NOT EXISTS idx_stock_movements_lookup
ON inventory.stock_movements (legal_entity_id, item_id, warehouse_id, occurred_at DESC);

-- Trigger: Enforce Immutability on stock_movements (NO UPDATE or DELETE)
CREATE OR REPLACE FUNCTION inventory.prevent_stock_movement_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Stock movements are immutable audit records and cannot be modified or deleted (Movement ID: %)', OLD.id
        USING ERRCODE = '23506';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_immutable_stock_movements ON inventory.stock_movements;
CREATE TRIGGER trg_immutable_stock_movements
BEFORE UPDATE OR DELETE ON inventory.stock_movements
FOR EACH ROW
EXECUTE FUNCTION inventory.prevent_stock_movement_mutation();

-- 2.2 Stock Balances (Projection, Atomic Updates with Row Locking)
CREATE TABLE IF NOT EXISTS inventory.stock_balances (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE CASCADE,
    bin_id              UUID NOT NULL REFERENCES warehouse.bin_locations(id) ON DELETE CASCADE,
    quantity_on_hand    NUMERIC(18, 4) NOT NULL DEFAULT 0,
    quantity_reserved   NUMERIC(18, 4) NOT NULL DEFAULT 0 CHECK (quantity_reserved >= 0),
    quantity_available  NUMERIC(18, 4) GENERATED ALWAYS AS (quantity_on_hand - quantity_reserved) STORED,
    average_cost        NUMERIC(18, 4) NOT NULL DEFAULT 0 CHECK (average_cost >= 0),
    total_value         NUMERIC(18, 4) NOT NULL DEFAULT 0,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_stock_balances_dim
        UNIQUE (legal_entity_id, item_id, warehouse_id, bin_id)
);

CREATE INDEX IF NOT EXISTS idx_stock_balances_item
ON inventory.stock_balances (legal_entity_id, item_id);

CREATE INDEX IF NOT EXISTS idx_stock_balances_warehouse
ON inventory.stock_balances (legal_entity_id, warehouse_id);

-- 2.3 Stock Reservations (Temporary Hard Hold on Available Stock)
CREATE TABLE IF NOT EXISTS inventory.stock_reservations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE CASCADE,
    bin_id              UUID REFERENCES warehouse.bin_locations(id) ON DELETE SET NULL,
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    reservation_status  TEXT NOT NULL DEFAULT 'ACTIVE' CHECK (reservation_status IN ('ACTIVE', 'RELEASED', 'FULFILLED')),
    demand_source_type  TEXT NOT NULL, -- 'SALES_ORDER', 'TRANSFER_ORDER'
    demand_source_id    UUID NOT NULL,
    expires_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_stock_reservations_demand
ON inventory.stock_reservations (legal_entity_id, demand_source_type, demand_source_id);

-- 2.4 Transfer Orders (Warehouse to Warehouse)
CREATE TABLE IF NOT EXISTS inventory.transfer_orders (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    order_number            TEXT NOT NULL,
    from_warehouse_id       UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE RESTRICT,
    to_warehouse_id         UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE RESTRICT,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN (
                                'DRAFT', 'REQUESTED', 'APPROVED', 'DISPATCHED', 'RECEIVED', 'CANCELLED'
                            )),
    shipping_date           TIMESTAMPTZ,
    expected_delivery_date  TIMESTAMPTZ,
    received_date           TIMESTAMPTZ,
    notes                   TEXT,
    created_by              UUID NOT NULL REFERENCES iam.users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_transfer_orders_number
        UNIQUE (legal_entity_id, order_number),
    CONSTRAINT chk_different_warehouses
        CHECK (from_warehouse_id <> to_warehouse_id)
);

CREATE INDEX IF NOT EXISTS idx_transfer_orders_entity
ON inventory.transfer_orders (legal_entity_id, status);

CREATE TABLE IF NOT EXISTS inventory.transfer_order_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    transfer_order_id   UUID NOT NULL REFERENCES inventory.transfer_orders(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE RESTRICT,
    sku_id              UUID REFERENCES products.skus(id) ON DELETE RESTRICT,
    requested_quantity  NUMERIC(18, 4) NOT NULL CHECK (requested_quantity > 0),
    dispatched_quantity NUMERIC(18, 4) NOT NULL DEFAULT 0,
    received_quantity   NUMERIC(18, 4) NOT NULL DEFAULT 0,
    uom_id              UUID NOT NULL REFERENCES products.units_of_measure(id) ON DELETE RESTRICT,
    notes               TEXT
);

CREATE INDEX IF NOT EXISTS idx_transfer_lines_order
ON inventory.transfer_order_lines (transfer_order_id);

-- 2.5 Stock Adjustments
CREATE TABLE IF NOT EXISTS inventory.stock_adjustments (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    adjustment_number   TEXT NOT NULL,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE RESTRICT,
    adjustment_type     TEXT NOT NULL CHECK (adjustment_type IN ('INCREASE', 'DECREASE', 'SCRAP', 'FOUND', 'DAMAGED', 'CYCLE_COUNT')),
    reason_code         TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'POSTED', 'CANCELLED')),
    adjustment_date     DATE NOT NULL DEFAULT CURRENT_DATE,
    posted_at           TIMESTAMPTZ,
    posted_by           UUID REFERENCES iam.users(id),
    notes               TEXT,
    created_by          UUID NOT NULL REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_stock_adjustments_number
        UNIQUE (legal_entity_id, adjustment_number)
);

CREATE INDEX IF NOT EXISTS idx_stock_adjustments_entity
ON inventory.stock_adjustments (legal_entity_id, warehouse_id, status);

CREATE TABLE IF NOT EXISTS inventory.stock_adjustment_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    adjustment_id       UUID NOT NULL REFERENCES inventory.stock_adjustments(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE RESTRICT,
    bin_id              UUID REFERENCES warehouse.bin_locations(id) ON DELETE SET NULL,
    lot_number          TEXT,
    serial_number       TEXT,
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    unit_cost           NUMERIC(18, 4) NOT NULL DEFAULT 0,
    total_cost          NUMERIC(18, 4) NOT NULL DEFAULT 0,
    reason              TEXT
);

CREATE INDEX IF NOT EXISTS idx_stock_adjustment_lines_adj
ON inventory.stock_adjustment_lines (adjustment_id);

-- 2.6 Physical Stock Counts
CREATE TABLE IF NOT EXISTS inventory.stock_counts (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    count_number        TEXT NOT NULL,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id) ON DELETE RESTRICT,
    count_type          TEXT NOT NULL DEFAULT 'FULL' CHECK (count_type IN ('FULL', 'CYCLE', 'SPOT')),
    status              TEXT NOT NULL DEFAULT 'PLANNED' CHECK (status IN ('PLANNED', 'IN_PROGRESS', 'COMPLETED', 'POSTED', 'CANCELLED')),
    count_date          DATE NOT NULL DEFAULT CURRENT_DATE,
    reconciled_at       TIMESTAMPTZ,
    reconciled_by       UUID REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_stock_counts_number
        UNIQUE (legal_entity_id, count_number)
);

CREATE TABLE IF NOT EXISTS inventory.stock_count_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    count_id            UUID NOT NULL REFERENCES inventory.stock_counts(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE RESTRICT,
    bin_id              UUID REFERENCES warehouse.bin_locations(id) ON DELETE SET NULL,
    system_quantity     NUMERIC(18, 4) NOT NULL DEFAULT 0,
    counted_quantity    NUMERIC(18, 4) NOT NULL DEFAULT 0,
    variance            NUMERIC(18, 4) GENERATED ALWAYS AS (counted_quantity - system_quantity) STORED,
    notes               TEXT
);

-- 2.7 Lots & Serial Numbers
CREATE TABLE IF NOT EXISTS inventory.lots (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    lot_number          TEXT NOT NULL,
    manufacture_date    DATE,
    expiry_date         DATE,
    is_blocked          BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_lots_item_lot
        UNIQUE (legal_entity_id, item_id, lot_number)
);

CREATE TABLE IF NOT EXISTS inventory.serial_numbers (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    item_id             UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    serial_number       TEXT NOT NULL,
    warehouse_id        UUID REFERENCES warehouse.warehouses(id) ON DELETE SET NULL,
    bin_id              UUID REFERENCES warehouse.bin_locations(id) ON DELETE SET NULL,
    status              TEXT NOT NULL DEFAULT 'IN_STOCK' CHECK (status IN ('IN_STOCK', 'RESERVED', 'ISSUED', 'RETURNED', 'SCRAPPED')),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_serials_item_serial
        UNIQUE (legal_entity_id, item_id, serial_number)
);

-- ============================================================================
-- 3. PERMISSIONS & ROLES
-- ============================================================================
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('warehouse.view', 'عرض المستودعات وأماكن التخزين', 'View Warehouses & Bins', 'INVENTORY'),
    ('warehouse.manage', 'إدارة وتعديل المستودعات ومواقع التخزين', 'Manage Warehouses & Bin Locations', 'INVENTORY'),
    ('inventory.view', 'عرض أرصدة وحركات المخزون', 'View Stock Balances & Movements', 'INVENTORY'),
    ('inventory.manage', 'إدارة عمليات المخزون الشاملة', 'Manage Full Inventory Operations', 'INVENTORY'),
    ('inventory.post_movement', 'ترحيل حركات الاستلام والصرف المباشر', 'Post Direct Stock Movements', 'INVENTORY'),
    ('inventory.adjust', 'إنشاء واعتماد التسويات المخزنية', 'Create & Post Stock Adjustments', 'INVENTORY'),
    ('inventory.count', 'إجراء واعتماد الجرد الفعلي للمخازن', 'Execute & Post Stock Counts', 'INVENTORY'),
    ('inventory.transfer', 'إدارة أوامر النقل والتحويل بين المستودعات', 'Manage Inter-Warehouse Transfers', 'INVENTORY')
ON CONFLICT (code) DO NOTHING;

-- Grant permissions to SUPER_ADMIN role
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN' AND (p.code LIKE 'warehouse.%' OR p.code LIKE 'inventory.%')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- 4. SEED FUNCTION: Default Warehouse Structure & Opening Balance
-- ============================================================================
CREATE OR REPLACE FUNCTION inventory.seed_default_inventory_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_wh_main_id     UUID;
    v_wh_transit_id  UUID;
    v_bin_main_id    UUID;
    v_bin_transit_id UUID;
    v_item_lap_id    UUID;
    v_uom_pcs_id     UUID;
    v_admin_id       UUID;
BEGIN
    -- 1. Get Admin User
    SELECT id INTO v_admin_id FROM iam.users WHERE username = 'admin' LIMIT 1;
    IF v_admin_id IS NULL THEN
        SELECT id INTO v_admin_id FROM iam.users ORDER BY created_at ASC LIMIT 1;
    END IF;

    -- 2. Main Central Warehouse (WH-MAIN)
    INSERT INTO warehouse.warehouses (
        legal_entity_id, code, name_en, name_ar, warehouse_type, is_active, city, country_code
    ) VALUES (
        p_legal_entity_id, 'WH-MAIN', 'Riyadh Central Logistics Hub', 'المستودع الرئيسي - الرياض', 'PHYSICAL', true, 'Riyadh', 'SA'
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_wh_main_id;

    -- Default Bin for Main Warehouse
    INSERT INTO warehouse.bin_locations (
        warehouse_id, code, aisle, rack, shelf, bin, is_default, is_active
    ) VALUES (
        v_wh_main_id, 'DEFAULT', 'A1', 'R1', 'S1', 'B1', true, true
    ) ON CONFLICT (warehouse_id, code) DO UPDATE SET is_default = true
    RETURNING id INTO v_bin_main_id;

    -- 3. In-Transit Warehouse (WH-TRANSIT)
    INSERT INTO warehouse.warehouses (
        legal_entity_id, code, name_en, name_ar, warehouse_type, is_active, city, country_code
    ) VALUES (
        p_legal_entity_id, 'WH-TRANSIT', 'Goods in Transit Hub', 'مستودع البضاعة في الطريق', 'TRANSIT', true, 'Riyadh', 'SA'
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_wh_transit_id;

    -- Default Bin for Transit Warehouse
    INSERT INTO warehouse.bin_locations (
        warehouse_id, code, aisle, rack, shelf, bin, is_default, is_active
    ) VALUES (
        v_wh_transit_id, 'STAGE', 'T1', 'R0', 'S0', 'B0', true, true
    ) ON CONFLICT (warehouse_id, code) DO UPDATE SET is_default = true
    RETURNING id INTO v_bin_transit_id;

    -- 4. Initial Opening Balance for Sample Laptop if present
    SELECT id, base_uom_id INTO v_item_lap_id, v_uom_pcs_id
    FROM products.items
    WHERE legal_entity_id = p_legal_entity_id AND code = 'IT-LAPTOP-01'
    LIMIT 1;

    IF v_item_lap_id IS NOT NULL AND v_admin_id IS NOT NULL THEN
        -- Insert Initial Opening Stock Movement if not already posted
        INSERT INTO inventory.stock_movements (
            legal_entity_id, movement_number, movement_type, item_id, warehouse_id, bin_id,
            quantity, uom_id, unit_cost, total_cost, source_doc_type, created_by
        ) VALUES (
            p_legal_entity_id, 'MOV-INIT-000001', 'OPENING', v_item_lap_id, v_wh_main_id, v_bin_main_id,
            25.0000, v_uom_pcs_id, 3800.0000, 95000.0000, 'OPENING_BALANCE', v_admin_id
        ) ON CONFLICT (legal_entity_id, movement_number) DO NOTHING;

        -- Seed Projected Balance
        INSERT INTO inventory.stock_balances (
            legal_entity_id, item_id, warehouse_id, bin_id,
            quantity_on_hand, quantity_reserved, average_cost, total_value
        ) VALUES (
            p_legal_entity_id, v_item_lap_id, v_wh_main_id, v_bin_main_id,
            25.0000, 0.0000, 3800.0000, 95000.0000
        ) ON CONFLICT (legal_entity_id, item_id, warehouse_id, bin_id)
        DO UPDATE SET
            quantity_on_hand = EXCLUDED.quantity_on_hand,
            average_cost = EXCLUDED.average_cost,
            total_value = EXCLUDED.total_value;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS inventory.seed_default_inventory_data(UUID);
DROP TABLE IF EXISTS inventory.serial_numbers CASCADE;
DROP TABLE IF EXISTS inventory.lots CASCADE;
DROP TABLE IF EXISTS inventory.stock_count_lines CASCADE;
DROP TABLE IF EXISTS inventory.stock_counts CASCADE;
DROP TABLE IF EXISTS inventory.stock_adjustment_lines CASCADE;
DROP TABLE IF EXISTS inventory.stock_adjustments CASCADE;
DROP TABLE IF EXISTS inventory.transfer_order_lines CASCADE;
DROP TABLE IF EXISTS inventory.transfer_orders CASCADE;
DROP TABLE IF EXISTS inventory.stock_reservations CASCADE;
DROP TABLE IF EXISTS inventory.stock_balances CASCADE;
DROP TABLE IF EXISTS inventory.stock_movements CASCADE;
DROP TABLE IF EXISTS warehouse.bin_locations CASCADE;
DROP TABLE IF EXISTS warehouse.warehouses CASCADE;
DROP SCHEMA IF EXISTS inventory CASCADE;
DROP SCHEMA IF EXISTS warehouse CASCADE;
-- +goose StatementEnd
