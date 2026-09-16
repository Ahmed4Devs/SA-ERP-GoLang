-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS sales;

-- 1. Sales Quotations (عروض الأسعار)
CREATE TABLE IF NOT EXISTS sales.quotations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    quotation_number    TEXT NOT NULL,
    customer_id         UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE RESTRICT,
    customer_name_en    TEXT NOT NULL,
    customer_name_ar    TEXT NOT NULL,
    quote_date          DATE NOT NULL,
    valid_until         DATE NOT NULL,
    currency_code       TEXT NOT NULL DEFAULT 'SAR',
    exchange_rate       NUMERIC(12, 6) NOT NULL DEFAULT 1.000000,
    subtotal            NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    discount_amount     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    taxable_amount      NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_amount        NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    status              TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'SENT', 'ACCEPTED', 'REJECTED', 'EXPIRED', 'CONVERTED')),
    converted_order_id  UUID,
    notes               TEXT,
    created_by_user_id  UUID REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_quotation_number
        UNIQUE (legal_entity_id, quotation_number)
);

CREATE INDEX idx_quotations_customer ON sales.quotations(customer_id);
CREATE INDEX idx_quotations_status ON sales.quotations(status);

-- 2. Quotation Lines
CREATE TABLE IF NOT EXISTS sales.quotation_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    quotation_id        UUID NOT NULL REFERENCES sales.quotations(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    item_id             UUID NOT NULL REFERENCES products.items(id),
    description_en      TEXT NOT NULL,
    description_ar      TEXT NOT NULL,
    uom                 TEXT NOT NULL DEFAULT 'EA',
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    discount_percent    NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    discount_amount     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    taxable_amount      NUMERIC(18, 4) NOT NULL,
    tax_code            TEXT NOT NULL DEFAULT 'VAT-15',
    tax_rate_percent    NUMERIC(5, 2) NOT NULL DEFAULT 15.00,
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_amount        NUMERIC(18, 4) NOT NULL,

    CONSTRAINT uq_quotation_line
        UNIQUE (quotation_id, line_number)
);

-- 3. Sales Orders (أوامر البيع والتعميد)
CREATE TABLE IF NOT EXISTS sales.orders (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    order_number            TEXT NOT NULL,
    customer_id             UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE RESTRICT,
    quotation_id            UUID REFERENCES sales.quotations(id) ON DELETE SET NULL,
    customer_po_number      TEXT,
    order_date              DATE NOT NULL,
    expected_delivery_date  DATE,
    payment_terms_id        UUID REFERENCES receivables.payment_terms(id),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    exchange_rate           NUMERIC(12, 6) NOT NULL DEFAULT 1.000000,
    subtotal                NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    discount_amount         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    taxable_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    tax_amount              NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_amount            NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'CONFIRMED', 'PARTIALLY_DELIVERED', 'DELIVERED', 'PARTIALLY_INVOICED', 'INVOICED', 'CANCELLED')),
    notes                   TEXT,
    created_by_user_id      UUID REFERENCES iam.users(id),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_order_number
        UNIQUE (legal_entity_id, order_number)
);

CREATE INDEX idx_orders_customer ON sales.orders(customer_id);
CREATE INDEX idx_orders_status ON sales.orders(status);

-- 4. Sales Order Lines
CREATE TABLE IF NOT EXISTS sales.order_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    order_id            UUID NOT NULL REFERENCES sales.orders(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    item_id             UUID NOT NULL REFERENCES products.items(id),
    description_en      TEXT NOT NULL,
    description_ar      TEXT NOT NULL,
    uom                 TEXT NOT NULL DEFAULT 'EA',
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    delivered_quantity  NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (delivered_quantity >= 0),
    invoiced_quantity   NUMERIC(18, 4) NOT NULL DEFAULT 0.0000 CHECK (invoiced_quantity >= 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    discount_percent    NUMERIC(5, 2) NOT NULL DEFAULT 0.00 CHECK (discount_percent >= 0 AND discount_percent <= 100),
    discount_amount     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    taxable_amount      NUMERIC(18, 4) NOT NULL,
    tax_code            TEXT NOT NULL DEFAULT 'VAT-15',
    tax_rate_percent    NUMERIC(5, 2) NOT NULL DEFAULT 15.00,
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_amount        NUMERIC(18, 4) NOT NULL,

    CONSTRAINT uq_order_line
        UNIQUE (order_id, line_number)
);

-- 5. Goods Deliveries / Shipments (مذكرات التسليم وصرف المخزون)
CREATE TABLE IF NOT EXISTS sales.deliveries (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    delivery_number     TEXT NOT NULL,
    order_id            UUID REFERENCES sales.orders(id) ON DELETE SET NULL,
    customer_id         UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE RESTRICT,
    warehouse_id        UUID NOT NULL REFERENCES warehouse.warehouses(id),
    delivery_date       DATE NOT NULL,
    status              TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'SHIPPED', 'DELIVERED', 'CANCELLED')),
    shipping_reference  TEXT,
    notes               TEXT,
    created_by_user_id  UUID REFERENCES iam.users(id),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_delivery_number
        UNIQUE (legal_entity_id, delivery_number)
);

CREATE INDEX idx_deliveries_order ON sales.deliveries(order_id);
CREATE INDEX idx_deliveries_customer ON sales.deliveries(customer_id);

-- 6. Delivery Lines
CREATE TABLE IF NOT EXISTS sales.delivery_lines (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    delivery_id             UUID NOT NULL REFERENCES sales.deliveries(id) ON DELETE CASCADE,
    line_number             INT NOT NULL,
    order_line_id           UUID REFERENCES sales.order_lines(id) ON DELETE SET NULL,
    item_id                 UUID NOT NULL REFERENCES products.items(id),
    uom                     TEXT NOT NULL DEFAULT 'EA',
    quantity                NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    warehouse_id            UUID NOT NULL REFERENCES warehouse.warehouses(id),
    bin_id                  UUID REFERENCES warehouse.bin_locations(id),
    inventory_movement_id   UUID REFERENCES inventory.stock_movements(id),
    unit_cost               NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_cost              NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,

    CONSTRAINT uq_delivery_line
        UNIQUE (delivery_id, line_number)
);

-- 7. Sales Invoices (فواتير المبيعات الضريبية والمبسطة المتوافقة مع ZATCA)
CREATE TABLE IF NOT EXISTS sales.invoices (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    invoice_number          TEXT NOT NULL,
    invoice_type            TEXT NOT NULL CHECK (invoice_type IN ('STANDARD', 'SIMPLIFIED')),
    order_id                UUID REFERENCES sales.orders(id) ON DELETE SET NULL,
    delivery_id             UUID REFERENCES sales.deliveries(id) ON DELETE SET NULL,
    customer_id             UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE RESTRICT,
    customer_name_ar        TEXT NOT NULL,
    customer_name_en        TEXT NOT NULL,
    customer_tax_id         TEXT,
    customer_cr_number      TEXT,
    customer_address_json   JSONB,
    issue_date              DATE NOT NULL,
    due_date                DATE NOT NULL,
    payment_terms_id        UUID REFERENCES receivables.payment_terms(id),
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    exchange_rate           NUMERIC(12, 6) NOT NULL DEFAULT 1.000000,
    subtotal                NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    discount_amount         NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    taxable_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    tax_amount              NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_amount            NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'APPROVED', 'POSTED', 'CANCELLED')),
    journal_entry_id        UUID REFERENCES accounting.journal_entries(id),
    open_item_id            UUID REFERENCES receivables.open_items(id),
    qr_code_data            TEXT,
    notes                   TEXT,
    created_by_user_id      UUID REFERENCES iam.users(id),
    posted_at               TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_invoice_number
        UNIQUE (legal_entity_id, invoice_number)
);

CREATE INDEX idx_invoices_customer ON sales.invoices(customer_id);
CREATE INDEX idx_invoices_status ON sales.invoices(status);
CREATE INDEX idx_invoices_issue_date ON sales.invoices(issue_date);

-- 8. Sales Invoice Lines (Preserving Exact Historical Snapshots per Section 40.14)
CREATE TABLE IF NOT EXISTS sales.invoice_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    invoice_id          UUID NOT NULL REFERENCES sales.invoices(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    item_id             UUID NOT NULL REFERENCES products.items(id),
    description         TEXT NOT NULL,
    uom                 TEXT NOT NULL DEFAULT 'EA',
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    gross_amount        NUMERIC(18, 4) NOT NULL,
    discount_amount     NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    taxable_amount      NUMERIC(18, 4) NOT NULL,
    tax_code            TEXT NOT NULL DEFAULT 'VAT-15',
    tax_rate_percent    NUMERIC(5, 2) NOT NULL DEFAULT 15.00,
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    net_amount          NUMERIC(18, 4) NOT NULL,
    exemption_reason    TEXT,

    CONSTRAINT uq_invoice_line
        UNIQUE (invoice_id, line_number)
);

-- 9. Sales Credit Notes (إشعارات الدائن والمرتجعات)
CREATE TABLE IF NOT EXISTS sales.credit_notes (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id         UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    credit_note_number      TEXT NOT NULL,
    original_invoice_id     UUID NOT NULL REFERENCES sales.invoices(id),
    customer_id             UUID NOT NULL REFERENCES receivables.customers(id) ON DELETE RESTRICT,
    issue_date              DATE NOT NULL,
    reason_code             TEXT NOT NULL CHECK (reason_code IN ('RETURN', 'DISCOUNT', 'CORRECTION', 'OTHER')),
    reason_details          TEXT,
    currency_code           TEXT NOT NULL DEFAULT 'SAR',
    exchange_rate           NUMERIC(12, 6) NOT NULL DEFAULT 1.000000,
    subtotal                NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    tax_amount              NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_amount            NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    status                  TEXT NOT NULL DEFAULT 'DRAFT' CHECK (status IN ('DRAFT', 'POSTED', 'CANCELLED')),
    journal_entry_id        UUID REFERENCES accounting.journal_entries(id),
    open_item_id            UUID REFERENCES receivables.open_items(id),
    created_by_user_id      UUID REFERENCES iam.users(id),
    posted_at               TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_credit_note_number
        UNIQUE (legal_entity_id, credit_note_number)
);

CREATE INDEX idx_credit_notes_invoice ON sales.credit_notes(original_invoice_id);
CREATE INDEX idx_credit_notes_customer ON sales.credit_notes(customer_id);

-- 10. Credit Note Lines
CREATE TABLE IF NOT EXISTS sales.credit_note_lines (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    credit_note_id      UUID NOT NULL REFERENCES sales.credit_notes(id) ON DELETE CASCADE,
    line_number         INT NOT NULL,
    item_id             UUID NOT NULL REFERENCES products.items(id),
    description         TEXT NOT NULL,
    quantity            NUMERIC(18, 4) NOT NULL CHECK (quantity > 0),
    unit_price          NUMERIC(18, 4) NOT NULL CHECK (unit_price >= 0),
    taxable_amount      NUMERIC(18, 4) NOT NULL,
    tax_code            TEXT NOT NULL DEFAULT 'VAT-15',
    tax_rate_percent    NUMERIC(5, 2) NOT NULL DEFAULT 15.00,
    tax_amount          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    total_amount        NUMERIC(18, 4) NOT NULL,

    CONSTRAINT uq_credit_note_line
        UNIQUE (credit_note_id, line_number)
);

-- 11. Authz Permissions for Sales Module
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at) VALUES
    (uuidv7(), 'sales.quotation.create', 'Create Quotation', 'إنشاء عرض أسعار', 'SALES', now()),
    (uuidv7(), 'sales.quotation.read', 'Read Quotations', 'استعراض عروض الأسعار', 'SALES', now()),
    (uuidv7(), 'sales.quotation.update', 'Update Quotation', 'تعديل عرض أسعار', 'SALES', now()),
    (uuidv7(), 'sales.quotation.delete', 'Delete Quotation', 'حذف عرض أسعار', 'SALES', now()),
    (uuidv7(), 'sales.order.create', 'Create Sales Order', 'إنشاء أمر بيع', 'SALES', now()),
    (uuidv7(), 'sales.order.read', 'Read Sales Orders', 'استعراض أوامر البيع', 'SALES', now()),
    (uuidv7(), 'sales.order.update', 'Update Sales Order', 'تعديل أمر بيع', 'SALES', now()),
    (uuidv7(), 'sales.order.approve', 'Approve Sales Order', 'اعتماد أمر بيع', 'SALES', now()),
    (uuidv7(), 'sales.delivery.create', 'Create Delivery Note', 'إنشاء مذكرة تسليم', 'SALES', now()),
    (uuidv7(), 'sales.delivery.read', 'Read Delivery Notes', 'استعراض مذكرات التسليم', 'SALES', now()),
    (uuidv7(), 'sales.delivery.post', 'Post Delivery Note', 'ترحيل وصرف مذكرة تسليم', 'SALES', now()),
    (uuidv7(), 'sales.invoice.create', 'Create Sales Invoice', 'إنشاء فاتورة مبيعات', 'SALES', now()),
    (uuidv7(), 'sales.invoice.read', 'Read Sales Invoices', 'استعراض فواتير المبيعات', 'SALES', now()),
    (uuidv7(), 'sales.invoice.update', 'Update Sales Invoice', 'تعديل فاتورة مبيعات', 'SALES', now()),
    (uuidv7(), 'sales.invoice.approve', 'Approve Sales Invoice', 'اعتماد فاتورة مبيعات', 'SALES', now()),
    (uuidv7(), 'sales.invoice.post', 'Post Sales Invoice', 'ترحيل واعتماد فاتورة مبيعات', 'SALES', now()),
    (uuidv7(), 'sales.credit_note.create', 'Create Credit Note', 'إنشاء إشعار دائن', 'SALES', now()),
    (uuidv7(), 'sales.credit_note.read', 'Read Credit Notes', 'استعراض إشعارات الدائن', 'SALES', now()),
    (uuidv7(), 'sales.credit_note.post', 'Post Credit Note', 'ترحيل إشعار دائن', 'SALES', now())
ON CONFLICT (code) DO NOTHING;

-- Grant Sales Permissions to SUPER_ADMIN
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.name_en = 'Super Administrator' AND p.module = 'SALES'
ON CONFLICT DO NOTHING;

-- 12. Helper Seed Function for Sales
CREATE OR REPLACE FUNCTION sales.seed_default_sales_data(p_legal_entity_id UUID)
RETURNS void AS $$
DECLARE
    v_cust_id UUID;
    v_item_id UUID;
    v_wh_id UUID;
    v_quote_id UUID;
    v_order_id UUID;
    v_inv_id UUID;
    v_user_id UUID;
BEGIN
    SELECT id INTO v_cust_id FROM receivables.customers WHERE legal_entity_id = p_legal_entity_id LIMIT 1;
    SELECT id INTO v_item_id FROM products.items WHERE legal_entity_id = p_legal_entity_id LIMIT 1;
    SELECT id INTO v_wh_id FROM warehouse.warehouses WHERE legal_entity_id = p_legal_entity_id LIMIT 1;
    SELECT id INTO v_user_id FROM iam.users WHERE username = 'admin' LIMIT 1;

    IF v_cust_id IS NOT NULL AND v_item_id IS NOT NULL THEN
        -- Seed initial quotation
        INSERT INTO sales.quotations (
            legal_entity_id, quotation_number, customer_id, customer_name_en, customer_name_ar,
            quote_date, valid_until, subtotal, taxable_amount, tax_amount, total_amount, status, created_by_user_id
        ) VALUES (
            p_legal_entity_id, 'QT-2026-0001', v_cust_id, 'Al-Arabiya Modern Contracting Co.', 'شركة العربية الحديثة للمقاولات العامة',
            CURRENT_DATE, CURRENT_DATE + INTERVAL '30 days', 10000.0000, 10000.0000, 1500.0000, 11500.0000, 'ACCEPTED', v_user_id
        ) ON CONFLICT (legal_entity_id, quotation_number) DO NOTHING
        RETURNING id INTO v_quote_id;

        IF v_quote_id IS NOT NULL THEN
            INSERT INTO sales.quotation_lines (
                quotation_id, line_number, item_id, description_en, description_ar,
                quantity, unit_price, taxable_amount, tax_amount, total_amount
            ) VALUES (
                v_quote_id, 1, v_item_id, 'Sample Item Supply', 'توريد صنف قياسي للتجربة',
                10.0000, 1000.0000, 10000.0000, 1500.0000, 11500.0000
            ) ON CONFLICT DO NOTHING;
        END IF;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Seed default sales for existing legal entities
DO $$
DECLARE
    le RECORD;
BEGIN
    FOR le IN SELECT id FROM organization.legal_entities LOOP
        PERFORM sales.seed_default_sales_data(le.id);
    END LOOP;
END $$;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS sales.seed_default_sales_data(UUID);
DROP TABLE IF EXISTS sales.credit_note_lines CASCADE;
DROP TABLE IF EXISTS sales.credit_notes CASCADE;
DROP TABLE IF EXISTS sales.invoice_lines CASCADE;
DROP TABLE IF EXISTS sales.invoices CASCADE;
DROP TABLE IF EXISTS sales.delivery_lines CASCADE;
DROP TABLE IF EXISTS sales.deliveries CASCADE;
DROP TABLE IF EXISTS sales.order_lines CASCADE;
DROP TABLE IF EXISTS sales.orders CASCADE;
DROP TABLE IF EXISTS sales.quotation_lines CASCADE;
DROP TABLE IF EXISTS sales.quotations CASCADE;
DROP SCHEMA IF EXISTS sales CASCADE;
-- +goose StatementEnd
