-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE SCHEMA IF NOT EXISTS ref;

CREATE TABLE IF NOT EXISTS ref.currencies (
    code                CHAR(3) PRIMARY KEY,
    name_en             TEXT NOT NULL,
    name_ar             TEXT,
    decimal_places      SMALLINT NOT NULL DEFAULT 2,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT ck_currency_code_upper
        CHECK (code = upper(code)),

    CONSTRAINT ck_currency_decimal_places
        CHECK (decimal_places BETWEEN 0 AND 6)
);

CREATE TABLE IF NOT EXISTS ref.countries (
    code                CHAR(2) PRIMARY KEY,
    name_en             TEXT NOT NULL,
    name_ar             TEXT,
    phone_prefix        TEXT,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,

    CONSTRAINT ck_country_code_upper
        CHECK (code = upper(code))
);

-- Canonical reference data required by downstream foreign keys and defaults.
-- Keep these rows idempotent because this migration is also used to bootstrap
-- disposable development and integration-test databases.
INSERT INTO ref.currencies (code, name_en, name_ar, decimal_places, is_active) VALUES
    ('SAR', 'Saudi Riyal', 'الريال السعودي', 2, TRUE),
    ('USD', 'US Dollar', 'الدولار الأمريكي', 2, TRUE),
    ('EUR', 'Euro', 'اليورو', 2, TRUE),
    ('AED', 'UAE Dirham', 'الدرهم الإماراتي', 2, TRUE)
ON CONFLICT (code) DO UPDATE SET
    name_en = EXCLUDED.name_en,
    name_ar = EXCLUDED.name_ar,
    decimal_places = EXCLUDED.decimal_places,
    is_active = EXCLUDED.is_active;

INSERT INTO ref.countries (code, name_en, name_ar, phone_prefix, is_active) VALUES
    ('SA', 'Saudi Arabia', 'المملكة العربية السعودية', '+966', TRUE),
    ('AE', 'United Arab Emirates', 'الإمارات العربية المتحدة', '+971', TRUE),
    ('BH', 'Bahrain', 'البحرين', '+973', TRUE),
    ('KW', 'Kuwait', 'الكويت', '+965', TRUE),
    ('OM', 'Oman', 'عُمان', '+968', TRUE),
    ('QA', 'Qatar', 'قطر', '+974', TRUE)
ON CONFLICT (code) DO UPDATE SET
    name_en = EXCLUDED.name_en,
    name_ar = EXCLUDED.name_ar,
    phone_prefix = EXCLUDED.phone_prefix,
    is_active = EXCLUDED.is_active;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ref.countries;
DROP TABLE IF EXISTS ref.currencies;
DROP SCHEMA IF EXISTS ref CASCADE;
-- +goose StatementEnd
