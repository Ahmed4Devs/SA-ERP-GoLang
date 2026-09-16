-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS products;

-- 1. Standard Units of Measure (UOM)
CREATE TABLE IF NOT EXISTS products.units_of_measure (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    code            TEXT NOT NULL UNIQUE,
    name_en         TEXT NOT NULL,
    name_ar         TEXT NOT NULL,
    symbol_en       TEXT,
    symbol_ar       TEXT,
    uom_type        TEXT NOT NULL CHECK (uom_type IN ('COUNT', 'WEIGHT', 'LENGTH', 'VOLUME', 'AREA', 'TIME')),
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed System Standard Units
INSERT INTO products.units_of_measure (code, name_en, name_ar, symbol_en, symbol_ar, uom_type) VALUES
    ('PCS', 'Piece / Unit', 'حبة / قطعة', 'pcs', 'قطعة', 'COUNT'),
    ('BOX', 'Box', 'صندوق / كرتون', 'box', 'كرتون', 'COUNT'),
    ('PACK', 'Pack', 'طرد / عبوة', 'pk', 'عبوة', 'COUNT'),
    ('SET', 'Set', 'طقم / مجموعة', 'set', 'طقم', 'COUNT'),
    ('PALLET', 'Pallet', 'طبلية', 'plt', 'طبلية', 'COUNT'),
    ('KG', 'Kilogram', 'كيلوجرام', 'kg', 'كجم', 'WEIGHT'),
    ('G', 'Gram', 'جرام', 'g', 'جم', 'WEIGHT'),
    ('TON', 'Metric Ton', 'طن متري', 't', 'طن', 'WEIGHT'),
    ('M', 'Meter', 'متر', 'm', 'م', 'LENGTH'),
    ('CM', 'Centimeter', 'سنتيمتر', 'cm', 'سم', 'LENGTH'),
    ('MM', 'Millimeter', 'مليمتر', 'mm', 'ملم', 'LENGTH'),
    ('LTR', 'Liter', 'لتر', 'L', 'لتر', 'VOLUME'),
    ('ML', 'Milliliter', 'مليلتر', 'mL', 'مل', 'VOLUME'),
    ('M2', 'Square Meter', 'متر مربع', 'm²', 'م²', 'AREA'),
    ('HOUR', 'Hour', 'ساعة عمل', 'hr', 'ساعة', 'TIME'),
    ('DAY', 'Day', 'يوم عمل', 'day', 'يوم', 'TIME'),
    ('MONTH', 'Month', 'شهر عمل', 'mo', 'شهر', 'TIME')
ON CONFLICT (code) DO NOTHING;

-- 2. General UOM Conversions (Universal Physical Conversions)
CREATE TABLE IF NOT EXISTS products.uom_conversions (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    from_uom_id     UUID NOT NULL REFERENCES products.units_of_measure(id) ON DELETE CASCADE,
    to_uom_id       UUID NOT NULL REFERENCES products.units_of_measure(id) ON DELETE CASCADE,
    conversion_rate NUMERIC(18, 6) NOT NULL CHECK (conversion_rate > 0),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_uom_conversions
        UNIQUE (from_uom_id, to_uom_id)
);

-- Seed standard physical conversions
INSERT INTO products.uom_conversions (from_uom_id, to_uom_id, conversion_rate)
SELECT f.id, t.id, 1000.000000
FROM products.units_of_measure f, products.units_of_measure t
WHERE f.code = 'KG' AND t.code = 'G'
ON CONFLICT DO NOTHING;

INSERT INTO products.uom_conversions (from_uom_id, to_uom_id, conversion_rate)
SELECT f.id, t.id, 0.001000
FROM products.units_of_measure f, products.units_of_measure t
WHERE f.code = 'G' AND t.code = 'KG'
ON CONFLICT DO NOTHING;

INSERT INTO products.uom_conversions (from_uom_id, to_uom_id, conversion_rate)
SELECT f.id, t.id, 100.000000
FROM products.units_of_measure f, products.units_of_measure t
WHERE f.code = 'M' AND t.code = 'CM'
ON CONFLICT DO NOTHING;

INSERT INTO products.uom_conversions (from_uom_id, to_uom_id, conversion_rate)
SELECT f.id, t.id, 1000.000000
FROM products.units_of_measure f, products.units_of_measure t
WHERE f.code = 'LTR' AND t.code = 'ML'
ON CONFLICT DO NOTHING;

-- 3. Product Categories Tree
CREATE TABLE IF NOT EXISTS products.categories (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    parent_id       UUID REFERENCES products.categories(id) ON DELETE RESTRICT,
    code            TEXT NOT NULL,
    name_en         TEXT NOT NULL,
    name_ar         TEXT NOT NULL,
    description     TEXT,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_categories_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_categories_entity
ON products.categories (legal_entity_id, is_active);

-- 4. Product Items (Canonical Master)
CREATE TABLE IF NOT EXISTS products.items (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    category_id         UUID REFERENCES products.categories(id) ON DELETE SET NULL,
    code                TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    description_en      TEXT,
    description_ar      TEXT,
    item_type           TEXT NOT NULL CHECK (item_type IN ('STOCK', 'SERVICE', 'NON_STOCK', 'ASSET')),
    tracking_type       TEXT NOT NULL DEFAULT 'NONE' CHECK (tracking_type IN ('NONE', 'LOT', 'SERIAL')),
    base_uom_id         UUID NOT NULL REFERENCES products.units_of_measure(id) ON DELETE RESTRICT,
    purchase_uom_id     UUID REFERENCES products.units_of_measure(id) ON DELETE SET NULL,
    sales_uom_id        UUID REFERENCES products.units_of_measure(id) ON DELETE SET NULL,
    default_tax_code_id UUID REFERENCES tax.tax_codes(id) ON DELETE SET NULL,
    cost_price          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    list_price          NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    is_sales_enabled    BOOLEAN NOT NULL DEFAULT true,
    is_purchase_enabled BOOLEAN NOT NULL DEFAULT true,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_items_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_items_entity_type
ON products.items (legal_entity_id, item_type, is_active);

CREATE INDEX IF NOT EXISTS idx_items_category
ON products.items (category_id);

-- 5. Product Item Variants
CREATE TABLE IF NOT EXISTS products.item_variants (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    item_id         UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    variant_code    TEXT NOT NULL,
    name_en         TEXT NOT NULL,
    name_ar         TEXT NOT NULL,
    attributes      JSONB NOT NULL DEFAULT '{}'::jsonb,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_item_variants_code
        UNIQUE (item_id, variant_code)
);

CREATE INDEX IF NOT EXISTS idx_item_variants_item
ON products.item_variants (item_id, is_active);

-- 6. Product SKUs (Stock Keeping Units)
CREATE TABLE IF NOT EXISTS products.skus (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    item_id         UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    variant_id      UUID REFERENCES products.item_variants(id) ON DELETE CASCADE,
    sku_code        TEXT NOT NULL,
    cost_price      NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    selling_price   NUMERIC(18, 4) NOT NULL DEFAULT 0.0000,
    is_active       BOOLEAN NOT NULL DEFAULT true,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_skus_entity_code
        UNIQUE (legal_entity_id, sku_code)
);

CREATE INDEX IF NOT EXISTS idx_skus_entity_code
ON products.skus (legal_entity_id, sku_code);

CREATE INDEX IF NOT EXISTS idx_skus_item
ON products.skus (item_id);

-- 7. Product Barcodes
CREATE TABLE IF NOT EXISTS products.barcodes (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    item_id         UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    sku_id          UUID REFERENCES products.skus(id) ON DELETE CASCADE,
    uom_id          UUID REFERENCES products.units_of_measure(id) ON DELETE SET NULL,
    barcode         TEXT NOT NULL,
    barcode_format  TEXT NOT NULL DEFAULT 'EAN13' CHECK (barcode_format IN ('EAN13', 'EAN8', 'UPCA', 'CODE128', 'QR', 'CUSTOM')),
    is_primary      BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_barcodes_entity_barcode
        UNIQUE (legal_entity_id, barcode)
);

CREATE INDEX IF NOT EXISTS idx_barcodes_lookup
ON products.barcodes (legal_entity_id, barcode);

-- 8. Item UOM Conversions (Specific Packaging Rates for Item)
CREATE TABLE IF NOT EXISTS products.item_uoms (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    item_id         UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    uom_id          UUID NOT NULL REFERENCES products.units_of_measure(id) ON DELETE RESTRICT,
    conversion_rate NUMERIC(18, 6) NOT NULL CHECK (conversion_rate > 0),
    is_base_uom     BOOLEAN NOT NULL DEFAULT false,
    is_sales_uom    BOOLEAN NOT NULL DEFAULT false,
    is_purchase_uom BOOLEAN NOT NULL DEFAULT false,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_item_uoms_item_uom
        UNIQUE (item_id, uom_id)
);

CREATE INDEX IF NOT EXISTS idx_item_uoms_item
ON products.item_uoms (item_id);

-- 9. Item Tax Profiles (ZATCA Direct Mapping)
CREATE TABLE IF NOT EXISTS products.item_tax_profiles (
    id                      UUID PRIMARY KEY DEFAULT uuidv7(),
    item_id                 UUID NOT NULL REFERENCES products.items(id) ON DELETE CASCADE,
    tax_code_id             UUID NOT NULL REFERENCES tax.tax_codes(id) ON DELETE RESTRICT,
    exemption_reason_code   TEXT REFERENCES tax.exemption_reasons(code),
    zero_rate_reason_code   TEXT REFERENCES tax.zero_rate_reasons(code),
    is_active               BOOLEAN NOT NULL DEFAULT true,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_item_tax_profiles_item
        UNIQUE (item_id)
);

-- 10. Seed Permissions
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('products.view', 'عرض المنتجات والكتالوج', 'View Products & Catalog', 'PRODUCTS'),
    ('products.manage', 'إدارة الأصناف والوحدات والباركود', 'Manage Items, Units & Barcodes', 'PRODUCTS'),
    ('products.manage_prices', 'تعديل أسعار التكلفة وقوائم البيع', 'Manage Cost & Selling Prices', 'PRODUCTS'),
    ('products.scan_barcode', 'البحث والمسح بالباركود', 'Scan & Search Barcodes', 'PRODUCTS')
ON CONFLICT (code) DO NOTHING;

-- Grant permissions to SUPER_ADMIN role
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN' AND p.code LIKE 'products.%'
ON CONFLICT DO NOTHING;

-- 11. Helper Function: Seed Default Products for a Legal Entity
CREATE OR REPLACE FUNCTION products.seed_default_products_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_uom_pcs_id UUID;
    v_uom_box_id UUID;
    v_uom_hr_id  UUID;
    v_cat_it_id  UUID;
    v_cat_serv_id UUID;
    v_tax_std_id UUID;
    v_item_lap_id UUID;
    v_item_srv_id UUID;
    v_var_lap1_id UUID;
    v_sku_lap1_id UUID;
    v_sku_srv_id  UUID;
BEGIN
    -- A. Units
    SELECT id INTO v_uom_pcs_id FROM products.units_of_measure WHERE code = 'PCS' LIMIT 1;
    SELECT id INTO v_uom_box_id FROM products.units_of_measure WHERE code = 'BOX' LIMIT 1;
    SELECT id INTO v_uom_hr_id  FROM products.units_of_measure WHERE code = 'HOUR' LIMIT 1;

    -- B. Default Tax Code (VAT-STD-15)
    SELECT id INTO v_tax_std_id FROM tax.tax_codes WHERE legal_entity_id = p_legal_entity_id AND code = 'VAT-STD-15' LIMIT 1;

    -- C. Standard Categories
    INSERT INTO products.categories (legal_entity_id, code, name_en, name_ar, description)
    VALUES (p_legal_entity_id, 'CAT-IT', 'IT & Computing Hardware', 'أجهزة وتقنيات الحاسب الآلي', 'Laptops, desktops and accessories')
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_cat_it_id;

    INSERT INTO products.categories (legal_entity_id, code, name_en, name_ar, description)
    VALUES (p_legal_entity_id, 'CAT-SERV', 'Professional & Consulting Services', 'الخدمات الاستشارية والمهنية', 'Engineering, software development and consulting')
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_cat_serv_id;

    -- D. Stock Item: Enterprise Business Laptop
    INSERT INTO products.items (
        legal_entity_id, category_id, code, name_en, name_ar,
        description_en, description_ar, item_type, tracking_type,
        base_uom_id, purchase_uom_id, sales_uom_id, default_tax_code_id,
        cost_price, list_price
    ) VALUES (
        p_legal_entity_id, v_cat_it_id, 'IT-LAPTOP-01', 'Enterprise Business Laptop 16-inch', 'جهاز حاسب محمول للأعمال 16 بوصة',
        'High-performance enterprise laptop with 32GB RAM and 1TB SSD', 'حاسب محمول عالي الأداء مع ذاكرة 32 جيجابايت وقرص 1 تيرابايت',
        'STOCK', 'SERIAL', v_uom_pcs_id, v_uom_box_id, v_uom_pcs_id, v_tax_std_id,
        3800.0000, 4800.0000
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET list_price = EXCLUDED.list_price
    RETURNING id INTO v_item_lap_id;

    -- Item UOM Box (1 Box = 5 Pcs)
    INSERT INTO products.item_uoms (item_id, uom_id, conversion_rate, is_base_uom, is_sales_uom, is_purchase_uom)
    VALUES
        (v_item_lap_id, v_uom_pcs_id, 1.000000, true, true, false),
        (v_item_lap_id, v_uom_box_id, 5.000000, false, false, true)
    ON CONFLICT (item_id, uom_id) DO NOTHING;

    -- Item Variant
    INSERT INTO products.item_variants (item_id, variant_code, name_en, name_ar, attributes)
    VALUES (v_item_lap_id, 'LAP-16-SLV', 'Laptop 16-inch Silver', 'لابتوب 16 بوصة فضي', '{"color": "Silver", "ram": "32GB"}'::jsonb)
    ON CONFLICT (item_id, variant_code) DO UPDATE SET name_ar = EXCLUDED.name_ar
    RETURNING id INTO v_var_lap1_id;

    -- SKU
    INSERT INTO products.skus (legal_entity_id, item_id, variant_id, sku_code, cost_price, selling_price)
    VALUES (p_legal_entity_id, v_item_lap_id, v_var_lap1_id, 'SKU-LAP-16-SLV', 3800.0000, 4800.0000)
    ON CONFLICT (legal_entity_id, sku_code) DO NOTHING
    RETURNING id INTO v_sku_lap1_id;

    -- Barcode
    INSERT INTO products.barcodes (legal_entity_id, item_id, sku_id, uom_id, barcode, barcode_format, is_primary)
    VALUES (p_legal_entity_id, v_item_lap_id, v_sku_lap1_id, v_uom_pcs_id, '6281000123456', 'EAN13', true)
    ON CONFLICT (legal_entity_id, barcode) DO NOTHING;

    -- Item Tax Profile
    IF v_tax_std_id IS NOT NULL THEN
        INSERT INTO products.item_tax_profiles (item_id, tax_code_id)
        VALUES (v_item_lap_id, v_tax_std_id)
        ON CONFLICT (item_id) DO NOTHING;
    END IF;

    -- E. Service Item: Software Architecture & Advisory Service
    INSERT INTO products.items (
        legal_entity_id, category_id, code, name_en, name_ar,
        description_en, description_ar, item_type, tracking_type,
        base_uom_id, purchase_uom_id, sales_uom_id, default_tax_code_id,
        cost_price, list_price
    ) VALUES (
        p_legal_entity_id, v_cat_serv_id, 'SRV-ARCH-01', 'Cloud Software Advisory Service', 'خدمات الاستشارات المعمارية السحابية',
        'Senior software engineering and architecture advisory service per hour', 'استشارات هندسية ومعمارية متقدمة بالساعة',
        'SERVICE', 'NONE', v_uom_hr_id, v_uom_hr_id, v_uom_hr_id, v_tax_std_id,
        250.0000, 500.0000
    ) ON CONFLICT (legal_entity_id, code) DO UPDATE SET list_price = EXCLUDED.list_price
    RETURNING id INTO v_item_srv_id;

    -- Service SKU
    INSERT INTO products.skus (legal_entity_id, item_id, sku_code, cost_price, selling_price)
    VALUES (p_legal_entity_id, v_item_srv_id, 'SKU-SRV-ARCH-01', 250.0000, 500.0000)
    ON CONFLICT (legal_entity_id, sku_code) DO NOTHING
    RETURNING id INTO v_sku_srv_id;

    -- Service Tax Profile
    IF v_tax_std_id IS NOT NULL THEN
        INSERT INTO products.item_tax_profiles (item_id, tax_code_id)
        VALUES (v_item_srv_id, v_tax_std_id)
        ON CONFLICT (item_id) DO NOTHING;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Automatically seed default product data for existing legal entities
DO $$
DECLARE
    le RECORD;
BEGIN
    FOR le IN SELECT id FROM organization.legal_entities WHERE is_active = true LOOP
        PERFORM products.seed_default_products_data(le.id);
    END LOOP;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products.item_tax_profiles;
DROP TABLE IF EXISTS products.item_uoms;
DROP TABLE IF EXISTS products.barcodes;
DROP TABLE IF EXISTS products.skus;
DROP TABLE IF EXISTS products.item_variants;
DROP TABLE IF EXISTS products.items;
DROP TABLE IF EXISTS products.categories;
DROP TABLE IF EXISTS products.uom_conversions;
DROP TABLE IF EXISTS products.units_of_measure;

DELETE FROM authz.role_permissions WHERE permission_id IN (
    SELECT id FROM authz.permissions WHERE code LIKE 'products.%'
);
DELETE FROM authz.permissions WHERE code LIKE 'products.%';

DROP SCHEMA IF EXISTS products CASCADE;
-- +goose StatementEnd
