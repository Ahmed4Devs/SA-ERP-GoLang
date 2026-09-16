-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS purchasing;

-- ============================================================================
-- 1. PURCHASE REQUESTS (Requisitions / طلبات الشراء الداخلية)
-- ============================================================================
CREATE TABLE IF NOT EXISTS purchasing.purchase_requests (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    pr_number               TEXT NOT NULL,
    requested_by_user_id    UUID NOT NULL REFERENCES iam.users(id),
    department_id           UUID REFERENCES organization.departments(id),
    request_date            DATE NOT NULL,
    required_by_date        DATE,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'SUBMITTED', 'APPROVED', 'REJECTED', 'PARTIALLY_ORDERED', 'ORDERED', 'CANCELLED')),
    purpose                 TEXT,
    notes                   TEXT,
    estimated_total         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (estimated_total >= 0),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_purchase_request_number
        UNIQUE (legal_entity_id, pr_number)
);

CREATE INDEX idx_purchase_requests_entity ON purchasing.purchase_requests(legal_entity_id);
CREATE INDEX idx_purchase_requests_status ON purchasing.purchase_requests(status);
CREATE INDEX idx_purchase_requests_date ON purchasing.purchase_requests(request_date);

CREATE TABLE IF NOT EXISTS purchasing.purchase_request_lines (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    purchase_request_id     UUID NOT NULL REFERENCES purchasing.purchase_requests(id) ON DELETE CASCADE,
    line_number             INT NOT NULL,
    item_id                 UUID REFERENCES products.items(id),
    description             TEXT NOT NULL,
    uom                     TEXT NOT NULL DEFAULT 'EA',
    quantity                NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    ordered_quantity        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (ordered_quantity >= 0),
    estimated_unit_cost     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (estimated_unit_cost >= 0),
    estimated_total_cost    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (estimated_total_cost >= 0),
    suggested_supplier_id   UUID REFERENCES payables.suppliers(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_pr_line_number
        UNIQUE (purchase_request_id, line_number)
);

CREATE INDEX idx_pr_lines_item ON purchasing.purchase_request_lines(item_id);

-- ============================================================================
-- 2. REQUEST FOR QUOTATIONS (RFQs / طلبات عروض الأسعار)
-- ============================================================================
CREATE TABLE IF NOT EXISTS purchasing.rfqs (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    rfq_number              TEXT NOT NULL,
    purchase_request_id     UUID REFERENCES purchasing.purchase_requests(id) ON DELETE SET NULL,
    title                   TEXT NOT NULL,
    rfq_date                DATE NOT NULL,
    closing_date            DATE NOT NULL,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'SENT', 'CLOSED', 'CANCELLED')),
    notes                   TEXT,
    created_by_user_id      UUID REFERENCES iam.users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_rfq_number
        UNIQUE (legal_entity_id, rfq_number)
);

CREATE INDEX idx_rfqs_entity ON purchasing.rfqs(legal_entity_id);
CREATE INDEX idx_rfqs_status ON purchasing.rfqs(status);

CREATE TABLE IF NOT EXISTS purchasing.rfq_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    rfq_id              UUID NOT NULL REFERENCES purchasing.rfqs(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    pr_line_id          UUID REFERENCES purchasing.purchase_request_lines(id) ON DELETE SET NULL,
    item_id             UUID REFERENCES products.items(id),
    description         TEXT NOT NULL,
    uom                 TEXT NOT NULL DEFAULT 'EA',
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    target_unit_price   NUMERIC(18, 4) CHECK (target_unit_price IS NULL OR target_unit_price >= 0),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_rfq_line_number
        UNIQUE (rfq_id, line_number)
);

-- ============================================================================
-- 3. SUPPLIER QUOTES (عروض أسعار الموردين)
-- ============================================================================
CREATE TABLE IF NOT EXISTS purchasing.supplier_quotes (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    rfq_id                      UUID REFERENCES purchasing.rfqs(id) ON DELETE SET NULL,
    supplier_id                 UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE RESTRICT,
    supplier_quote_reference    TEXT NOT NULL,
    quote_date                  DATE NOT NULL,
    valid_until                 DATE NOT NULL,
    currency_code               TEXT NOT NULL DEFAULT 'SAR',
    exchange_rate               NUMERIC(12, 6) NOT NULL DEFAULT 1.000000 CHECK (exchange_rate > 0),
    payment_terms_id            UUID REFERENCES payables.payment_terms(id),
    subtotal                    NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (subtotal >= 0),
    tax_amount                  NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    total_amount                NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_amount >= 0),
    status                      TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'ACCEPTED', 'REJECTED', 'EXPIRED')),
    notes                       TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_quote_ref
        UNIQUE (legal_entity_id, supplier_id, supplier_quote_reference)
);

CREATE INDEX idx_supplier_quotes_rfq ON purchasing.supplier_quotes(rfq_id);
CREATE INDEX idx_supplier_quotes_supplier ON purchasing.supplier_quotes(supplier_id);

CREATE TABLE IF NOT EXISTS purchasing.supplier_quote_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    quote_id            UUID NOT NULL REFERENCES purchasing.supplier_quotes(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    rfq_line_id         UUID REFERENCES purchasing.rfq_lines(id) ON DELETE SET NULL,
    item_id             UUID REFERENCES products.items(id),
    description         TEXT NOT NULL,
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    discount_percent    NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    tax_rate_percent    NUMERIC(5, 2) NOT NULL DEFAULT 15.00 CHECK (tax_rate_percent >= 0),
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    total_amount        NUMERIC(18, 4) NOT NULL CHECK (total_amount >= 0),
    is_awarded          BOOLEAN NOT NULL DEFAULT false,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_supplier_quote_line_number
        UNIQUE (quote_id, line_number)
);

-- ============================================================================
-- 4. PURCHASE ORDERS (أوامر الشراء والتعميد)
-- ============================================================================
CREATE TABLE IF NOT EXISTS purchasing.purchase_orders (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    po_number               TEXT NOT NULL,
    purchase_request_id     UUID REFERENCES purchasing.purchase_requests(id) ON DELETE SET NULL,
    rfq_id                  UUID REFERENCES purchasing.rfqs(id) ON DELETE SET NULL,
    supplier_quote_id       UUID REFERENCES purchasing.supplier_quotes(id) ON DELETE SET NULL,
    supplier_id             UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE RESTRICT,
    order_date              DATE NOT NULL,
    expected_delivery_date  DATE,
    payment_terms_id        UUID REFERENCES payables.payment_terms(id),
    delivery_warehouse_id   UUID REFERENCES warehouse.warehouses(id),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    exchange_rate           NUMERIC(12, 6) NOT NULL DEFAULT 1.000000 CHECK (exchange_rate > 0),
    subtotal                NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (subtotal >= 0),
    discount_amount         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (discount_amount >= 0),
    taxable_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (taxable_amount >= 0),
    tax_amount              NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    total_amount            NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_amount >= 0),
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'ISSUED', 'PARTIALLY_RECEIVED', 'RECEIVED', 'CLOSED', 'CANCELLED')),
    notes                   TEXT,
    created_by_user_id      UUID REFERENCES iam.users(id),
    approved_by_user_id     UUID REFERENCES iam.users(id),
    approved_at             TIMESTAMPTZ,
    issued_at               TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_po_number
        UNIQUE (legal_entity_id, po_number)
);

CREATE INDEX idx_purchase_orders_supplier ON purchasing.purchase_orders(supplier_id);
CREATE INDEX idx_purchase_orders_status ON purchasing.purchase_orders(status);
CREATE INDEX idx_purchase_orders_date ON purchasing.purchase_orders(order_date);

CREATE TABLE IF NOT EXISTS purchasing.purchase_order_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    purchase_order_id   UUID NOT NULL REFERENCES purchasing.purchase_orders(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    pr_line_id          UUID REFERENCES purchasing.purchase_request_lines(id) ON DELETE SET NULL,
    item_id             UUID NOT NULL REFERENCES products.items(id),
    description         TEXT NOT NULL,
    uom                 TEXT NOT NULL DEFAULT 'EA',
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    received_quantity   NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (received_quantity >= 0),
    billed_quantity     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (billed_quantity >= 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    discount_percent    NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    discount_amount     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (discount_amount >= 0),
    taxable_amount      NUMERIC(18, 4) NOT NULL CHECK (taxable_amount >= 0),
    tax_code            TEXT NOT NULL DEFAULT 'VAT-15',
    tax_rate_percent    NUMERIC(5, 2) NOT NULL DEFAULT 15.00 CHECK (tax_rate_percent >= 0),
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (tax_amount >= 0),
    total_amount        NUMERIC(18, 4) NOT NULL CHECK (total_amount >= 0),
    expense_account_id  UUID REFERENCES accounting.accounts(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_po_line_number
        UNIQUE (purchase_order_id, line_number)
);

CREATE INDEX idx_po_lines_item ON purchasing.purchase_order_lines(item_id);

-- ============================================================================
-- 5. GOODS RECEIPTS & INSPECTION (سندات استلام وفحص البضائع GRN)
-- ============================================================================
CREATE TABLE IF NOT EXISTS purchasing.goods_receipts (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    receipt_number          TEXT NOT NULL,
    purchase_order_id       UUID NOT NULL REFERENCES purchasing.purchase_orders(id) ON DELETE RESTRICT,
    supplier_id             UUID NOT NULL REFERENCES payables.suppliers(id) ON DELETE RESTRICT,
    warehouse_id            UUID NOT NULL REFERENCES warehouse.warehouses(id),
    receipt_date            DATE NOT NULL,
    delivery_note_number    TEXT,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'ACCEPTED', 'REJECTED', 'CANCELLED')),
    journal_entry_id        UUID REFERENCES accounting.journal_entries(id),
    notes                   TEXT,
    received_by_user_id     UUID REFERENCES iam.users(id),
    inspected_by_user_id    UUID REFERENCES iam.users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_receipt_number
        UNIQUE (legal_entity_id, receipt_number)
);

CREATE INDEX idx_goods_receipts_po ON purchasing.goods_receipts(purchase_order_id);
CREATE INDEX idx_goods_receipts_supplier ON purchasing.goods_receipts(supplier_id);
CREATE INDEX idx_goods_receipts_status ON purchasing.goods_receipts(status);

CREATE TABLE IF NOT EXISTS purchasing.goods_receipt_lines (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    goods_receipt_id        UUID NOT NULL REFERENCES purchasing.goods_receipts(id) ON DELETE CASCADE,
    line_number             INT NOT NULL,
    po_line_id              UUID NOT NULL REFERENCES purchasing.purchase_order_lines(id) ON DELETE RESTRICT,
    item_id                 UUID NOT NULL REFERENCES products.items(id),
    warehouse_id            UUID NOT NULL REFERENCES warehouse.warehouses(id),
    bin_id                  UUID REFERENCES warehouse.bin_locations(id),
    received_quantity       NUMERIC(18, 4) NOT NULL CHECK (received_quantity >= 0),
    accepted_quantity       NUMERIC(18, 4) NOT NULL CHECK (accepted_quantity >= 0),
    rejected_quantity       NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (rejected_quantity >= 0),
    unit_cost               NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (unit_cost >= 0),
    total_cost              NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (total_cost >= 0),
    inventory_movement_id   UUID REFERENCES inventory.stock_movements(id),
    rejection_reason        TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_gr_line_number
        UNIQUE (goods_receipt_id, line_number)
);

CREATE INDEX idx_gr_lines_po_line ON purchasing.goods_receipt_lines(po_line_id);
CREATE INDEX idx_gr_lines_item ON purchasing.goods_receipt_lines(item_id);

-- ============================================================================
-- 6. PERMISSIONS SEEDING
-- ============================================================================
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'purchasing.request.create', 'Create Purchase Requisition', 'إنشاء طلب شراء داخلي', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.request.read',   'Read Purchase Requisitions',  'عرض طلبات الشراء الداخلية', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.request.approve','Approve Purchase Requisition', 'اعتماد طلب الشراء الداخلي', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.rfq.create',     'Create RFQ',                   'إنشاء طلب عروض أسعار', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.rfq.read',       'Read RFQ',                     'عرض طلبات عروض الأسعار', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.quote.manage',   'Manage Supplier Quotes',       'إدارة عروض أسعار الموردين', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.order.create',   'Create Purchase Order',        'إنشاء أمر شراء', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.order.read',     'Read Purchase Orders',         'عرض أوامر الشراء', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.order.approve',  'Approve Purchase Order',       'اعتماد أمر الشراء', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.receipt.create', 'Create Goods Receipt',         'إنشاء سند استلام بضائع', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.receipt.read',   'Read Goods Receipts',          'عرض سندات استلام البضائع', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.receipt.inspect','Inspect Goods Receipt',        'فحص واعتماد سند الاستلام', 'PURCHASING', now()),
    (uuidv7(), 'purchasing.match.execute',  'Execute 3-Way Match',          'تنفيذ المطابقة الثلاثية', 'PURCHASING', now())
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- 7. SEED FUNCTION
-- ============================================================================
CREATE OR REPLACE FUNCTION purchasing.seed_default_purchasing_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_admin_id UUID;
    v_supplier_id UUID;
    v_item_id UUID;
    v_warehouse_id UUID;
    v_bin_id UUID;
    v_terms_id UUID;
    v_pr_id UUID;
    v_po_id UUID;
    v_po_line_id UUID;
    v_receipt_id UUID;
    v_uom_id UUID;
BEGIN
    -- 1. Find Admin user
    SELECT id INTO v_admin_id FROM iam.users WHERE username = 'admin' LIMIT 1;
    IF v_admin_id IS NULL THEN
        SELECT id INTO v_admin_id FROM iam.users LIMIT 1;
    END IF;

    -- 2. Find default local supplier
    SELECT id, payment_terms_id INTO v_supplier_id, v_terms_id 
    FROM payables.suppliers 
    WHERE legal_entity_id = p_legal_entity_id AND supplier_number = 'SUP-0001' LIMIT 1;

    -- 3. Find default product item
    SELECT id INTO v_item_id FROM products.items WHERE legal_entity_id = p_legal_entity_id LIMIT 1;

    -- 4. Find default warehouse and bin
    SELECT id INTO v_warehouse_id FROM warehouse.warehouses WHERE legal_entity_id = p_legal_entity_id LIMIT 1;
    IF v_warehouse_id IS NOT NULL THEN
        SELECT id INTO v_bin_id FROM warehouse.bin_locations WHERE warehouse_id = v_warehouse_id LIMIT 1;
    END IF;

    -- 5. Find UOM
    SELECT id INTO v_uom_id FROM products.units_of_measure LIMIT 1;

    -- If prerequisite records exist, seed sample PR, PO, and GRN
    IF v_admin_id IS NOT NULL AND v_supplier_id IS NOT NULL AND v_item_id IS NOT NULL AND v_warehouse_id IS NOT NULL THEN
        
        -- A. Sample Purchase Requisition
        INSERT INTO purchasing.purchase_requests (
            id, legal_entity_id, pr_number, requested_by_user_id, request_date, required_by_date,
            status, purpose, estimated_total, currency_code
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'PR-2026-0001', v_admin_id, CURRENT_DATE, CURRENT_DATE + INTERVAL '14 days',
            'APPROVED', 'مستلزمات ومواد تشغيلية دورية للربع الحالي', 5000.0000, 'SAR'
        ) ON CONFLICT (legal_entity_id, pr_number) DO NOTHING
        RETURNING id INTO v_pr_id;

        IF v_pr_id IS NOT NULL THEN
            INSERT INTO purchasing.purchase_request_lines (
                purchase_request_id, line_number, item_id, description, uom, quantity,
                ordered_quantity, estimated_unit_cost, estimated_total_cost, suggested_supplier_id
            ) VALUES (
                v_pr_id, 1, v_item_id, 'توريد مواد أولية قياسية للمستودع الرئيسي', 'EA', 100.0000,
                100.0000, 50.0000, 5000.0000, v_supplier_id
            ) ON CONFLICT (purchase_request_id, line_number) DO NOTHING;
        END IF;

        -- B. Sample Purchase Order (Approved & Issued)
        INSERT INTO purchasing.purchase_orders (
            id, legal_entity_id, po_number, purchase_request_id, supplier_id, order_date,
            expected_delivery_date, payment_terms_id, delivery_warehouse_id, currency_code,
            subtotal, discount_amount, taxable_amount, tax_amount, total_amount, status,
            notes, created_by_user_id, approved_by_user_id, approved_at, issued_at
        ) VALUES (
            uuidv7(), p_legal_entity_id, 'PO-2026-0001', v_pr_id, v_supplier_id, CURRENT_DATE,
            CURRENT_DATE + INTERVAL '7 days', v_terms_id, v_warehouse_id, 'SAR',
            5000.0000, 0.0000, 5000.0000, 750.0000, 5750.0000, 'ISSUED',
            'أمر شراء معتمد للمواد الأولية - توريد فوري', v_admin_id, v_admin_id, now(), now()
        ) ON CONFLICT (legal_entity_id, po_number) DO NOTHING
        RETURNING id INTO v_po_id;

        IF v_po_id IS NOT NULL THEN
            INSERT INTO purchasing.purchase_order_lines (
                id, purchase_order_id, line_number, item_id, description, uom,
                quantity, received_quantity, billed_quantity, unit_price,
                discount_percent, discount_amount, taxable_amount, tax_code, tax_rate_percent,
                tax_amount, total_amount
            ) VALUES (
                uuidv7(), v_po_id, 1, v_item_id, 'توريد مواد أولية قياسية للمستودع الرئيسي', 'EA',
                100.0000, 100.0000, 0.0000, 50.0000,
                0.00, 0.0000, 5000.0000, 'VAT-15', 15.00,
                750.0000, 5750.0000
            ) ON CONFLICT (purchase_order_id, line_number) DO NOTHING
            RETURNING id INTO v_po_line_id;

            -- C. Sample Goods Receipt (GRN)
            INSERT INTO purchasing.goods_receipts (
                id, legal_entity_id, receipt_number, purchase_order_id, supplier_id,
                warehouse_id, receipt_date, delivery_note_number, status, notes,
                received_by_user_id, inspected_by_user_id
            ) VALUES (
                uuidv7(), p_legal_entity_id, 'GRN-2026-0001', v_po_id, v_supplier_id,
                v_warehouse_id, CURRENT_DATE, 'DN-SUP-88991', 'ACCEPTED', 'تم الاستلام والفحص ومطابقة الكميات والمواصفات',
                v_admin_id, v_admin_id
            ) ON CONFLICT (legal_entity_id, receipt_number) DO NOTHING
            RETURNING id INTO v_receipt_id;

            IF v_receipt_id IS NOT NULL AND v_po_line_id IS NOT NULL THEN
                INSERT INTO purchasing.goods_receipt_lines (
                    goods_receipt_id, line_number, po_line_id, item_id, warehouse_id,
                    bin_id, received_quantity, accepted_quantity, rejected_quantity,
                    unit_cost, total_cost
                ) VALUES (
                    v_receipt_id, 1, v_po_line_id, v_item_id, v_warehouse_id,
                    v_bin_id, 100.0000, 100.0000, 0.0000,
                    50.0000, 5000.0000
                ) ON CONFLICT (goods_receipt_id, line_number) DO NOTHING;
            END IF;
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS purchasing.seed_default_purchasing_data(UUID);
DROP TABLE IF EXISTS purchasing.goods_receipt_lines CASCADE;
DROP TABLE IF EXISTS purchasing.goods_receipts CASCADE;
DROP TABLE IF EXISTS purchasing.purchase_order_lines CASCADE;
DROP TABLE IF EXISTS purchasing.purchase_orders CASCADE;
DROP TABLE IF EXISTS purchasing.supplier_quote_lines CASCADE;
DROP TABLE IF EXISTS purchasing.supplier_quotes CASCADE;
DROP TABLE IF EXISTS purchasing.rfq_lines CASCADE;
DROP TABLE IF EXISTS purchasing.rfqs CASCADE;
DROP TABLE IF EXISTS purchasing.purchase_request_lines CASCADE;
DROP TABLE IF EXISTS purchasing.purchase_requests CASCADE;
DROP SCHEMA IF EXISTS purchasing CASCADE;
-- +goose StatementEnd
