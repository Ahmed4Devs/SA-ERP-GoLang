-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS organization;

-- 1. Organizations (Top-level company umbrella)
CREATE TABLE IF NOT EXISTS organization.organizations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    code                TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT ck_organizations_status
        CHECK (status IN ('ACTIVE', 'INACTIVE')),
    CONSTRAINT uq_organizations_code
        UNIQUE (code)
);

CREATE INDEX IF NOT EXISTS idx_organizations_status ON organization.organizations (status);

-- 2. Legal Entities (Independent financial & corporate entities with KSA Commercial Registration)
CREATE TABLE IF NOT EXISTS organization.legal_entities (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    organization_id     UUID NOT NULL REFERENCES organization.organizations(id) ON DELETE RESTRICT,
    code                TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    cr_number           TEXT,
    base_currency       TEXT NOT NULL DEFAULT 'SAR',
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_legal_entities_org_code
        UNIQUE (organization_id, code)
);

CREATE INDEX IF NOT EXISTS idx_legal_entities_org_id ON organization.legal_entities (organization_id);
CREATE INDEX IF NOT EXISTS idx_legal_entities_active ON organization.legal_entities (is_active);

-- 3. Tax Registrations (ZATCA VAT & Tax Numbers for Legal Entities)
CREATE TABLE IF NOT EXISTS organization.tax_registrations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    tax_type            TEXT NOT NULL DEFAULT 'VAT',
    registration_number TEXT NOT NULL,
    effective_from      DATE NOT NULL,
    effective_to        DATE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_tax_registrations_entity_type
        UNIQUE (legal_entity_id, tax_type)
);

CREATE INDEX IF NOT EXISTS idx_tax_registrations_entity_id ON organization.tax_registrations (legal_entity_id);

-- 4. Branches (Physical or administrative locations per Legal Entity)
CREATE TABLE IF NOT EXISTS organization.branches (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE RESTRICT,
    code                TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_branches_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_branches_entity_id ON organization.branches (legal_entity_id);

-- 5. Business Units
CREATE TABLE IF NOT EXISTS organization.business_units (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE RESTRICT,
    code                TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_business_units_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_business_units_entity_id ON organization.business_units (legal_entity_id);

-- 6. Departments (Hierarchical organizational departments)
CREATE TABLE IF NOT EXISTS organization.departments (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE RESTRICT,
    parent_id           UUID REFERENCES organization.departments(id) ON DELETE RESTRICT,
    code                TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_departments_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_departments_entity_id ON organization.departments (legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_departments_parent_id ON organization.departments (parent_id);

-- 7. Cost Centers (Hierarchical financial analytical dimensions)
CREATE TABLE IF NOT EXISTS organization.cost_centers (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE RESTRICT,
    parent_id           UUID REFERENCES organization.cost_centers(id) ON DELETE RESTRICT,
    code                TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT true,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT uq_cost_centers_entity_code
        UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_cost_centers_entity_id ON organization.cost_centers (legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_cost_centers_parent_id ON organization.cost_centers (parent_id);

-- 8. Addresses (KSA National Address specification)
CREATE TABLE IF NOT EXISTS organization.addresses (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    branch_id           UUID REFERENCES organization.branches(id) ON DELETE CASCADE,
    address_type        TEXT NOT NULL DEFAULT 'NATIONAL_ADDRESS',
    building_number     TEXT NOT NULL,
    street_name_ar      TEXT NOT NULL,
    street_name_en      TEXT NOT NULL,
    district_ar         TEXT NOT NULL,
    district_en         TEXT NOT NULL,
    city_ar             TEXT NOT NULL,
    city_en             TEXT NOT NULL,
    postal_code         TEXT NOT NULL,
    additional_number   TEXT,
    country_code        TEXT NOT NULL DEFAULT 'SA',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT ck_addresses_scope
        CHECK (legal_entity_id IS NOT NULL OR branch_id IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS idx_addresses_entity_id ON organization.addresses (legal_entity_id);
CREATE INDEX IF NOT EXISTS idx_addresses_branch_id ON organization.addresses (branch_id);

-- 9. Number Sequences (Concurrency-safe sequence engine)
CREATE TABLE IF NOT EXISTS organization.number_sequences (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE RESTRICT,
    sequence_code       TEXT NOT NULL,
    prefix              TEXT NOT NULL DEFAULT '',
    suffix              TEXT NOT NULL DEFAULT '',
    next_value          BIGINT NOT NULL DEFAULT 1,
    step                INT NOT NULL DEFAULT 1,
    padding_length      INT NOT NULL DEFAULT 6,
    reset_interval      TEXT NOT NULL DEFAULT 'YEARLY',
    last_reset_year     INT NOT NULL DEFAULT EXTRACT(YEAR FROM CURRENT_DATE),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    version             BIGINT NOT NULL DEFAULT 1,

    CONSTRAINT ck_number_sequences_reset_interval
        CHECK (reset_interval IN ('NEVER', 'YEARLY', 'MONTHLY')),
    CONSTRAINT uq_number_sequences_entity_code
        UNIQUE (legal_entity_id, sequence_code)
);

CREATE INDEX IF NOT EXISTS idx_number_sequences_entity_code ON organization.number_sequences (legal_entity_id, sequence_code);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS organization.number_sequences;
DROP TABLE IF EXISTS organization.addresses;
DROP TABLE IF EXISTS organization.cost_centers;
DROP TABLE IF EXISTS organization.departments;
DROP TABLE IF EXISTS organization.business_units;
DROP TABLE IF EXISTS organization.branches;
DROP TABLE IF EXISTS organization.tax_registrations;
DROP TABLE IF EXISTS organization.legal_entities;
DROP TABLE IF EXISTS organization.organizations;
DROP SCHEMA IF EXISTS organization CASCADE;
-- +goose StatementEnd
