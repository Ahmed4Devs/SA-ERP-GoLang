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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ref.countries;
DROP TABLE IF EXISTS ref.currencies;
DROP SCHEMA IF EXISTS ref CASCADE;
-- +goose StatementEnd
