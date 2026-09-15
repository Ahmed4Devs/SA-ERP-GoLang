-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS authz;

-- 1. Roles Definition
CREATE TABLE IF NOT EXISTS authz.roles (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    code                TEXT NOT NULL UNIQUE,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    description         TEXT,
    is_system           BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_authz_roles_code ON authz.roles (code);

-- 2. Granular Permissions Definition
CREATE TABLE IF NOT EXISTS authz.permissions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    code                TEXT NOT NULL UNIQUE,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    module              TEXT NOT NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_authz_permissions_code ON authz.permissions (code);
CREATE INDEX IF NOT EXISTS idx_authz_permissions_module ON authz.permissions (module);

-- 3. Role Permissions Mapping
CREATE TABLE IF NOT EXISTS authz.role_permissions (
    role_id             UUID NOT NULL REFERENCES authz.roles(id) ON DELETE CASCADE,
    permission_id       UUID NOT NULL REFERENCES authz.permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE INDEX IF NOT EXISTS idx_authz_role_permissions_role ON authz.role_permissions (role_id);

-- 4. User Role Assignments with Legal Entity Scope
CREATE TABLE IF NOT EXISTS authz.user_role_assignments (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    role_id             UUID NOT NULL REFERENCES authz.roles(id) ON DELETE CASCADE,
    legal_entity_id     UUID REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    branch_id           UUID REFERENCES organization.branches(id) ON DELETE CASCADE,
    assigned_at         TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_user_role_scope
        UNIQUE NULLS NOT DISTINCT (user_id, role_id, legal_entity_id, branch_id)
);

CREATE INDEX IF NOT EXISTS idx_authz_assignments_lookup
ON authz.user_role_assignments (user_id, legal_entity_id);

-- 5. Seed Core Permissions
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('org.organizations.view', 'عرض المنظمات القابضة', 'View Organizations', 'ORGANIZATION'),
    ('org.organizations.manage', 'إدارة المنظمات القابضة', 'Manage Organizations', 'ORGANIZATION'),
    ('org.legal_entities.view', 'عرض الكيانات القانونية', 'View Legal Entities', 'ORGANIZATION'),
    ('org.legal_entities.manage', 'إدارة الكيانات القانونية', 'Manage Legal Entities', 'ORGANIZATION'),
    ('org.branches.view', 'عرض الفروع', 'View Branches', 'ORGANIZATION'),
    ('org.branches.manage', 'إدارة الفروع', 'Manage Branches', 'ORGANIZATION'),
    ('org.departments.view', 'عرض الأقسام', 'View Departments', 'ORGANIZATION'),
    ('org.cost_centers.view', 'عرض مراكز التكلفة', 'View Cost Centers', 'ORGANIZATION'),
    ('org.number_sequences.generate', 'توليد أرقام الوثائق التسلسلية', 'Generate Document Numbers', 'ORGANIZATION'),

    ('iam.users.view', 'عرض المستخدمين', 'View Users', 'IAM'),
    ('iam.users.manage', 'إدارة المستخدمين', 'Manage Users', 'IAM'),
    ('iam.roles.view', 'عرض الأدوار والصلاحيات', 'View Roles & Permissions', 'IAM'),
    ('iam.roles.manage', 'إدارة الأدوار والصلاحيات', 'Manage Roles & Permissions', 'IAM'),

    ('gl.journals.view', 'عرض القيود اليومية', 'View Journal Entries', 'ACCOUNTING'),
    ('gl.journals.post', 'ترحيل القيود اليومية', 'Post Journal Entries', 'ACCOUNTING'),
    ('gl.accounts.view', 'عرض دليل الحسابات', 'View Chart of Accounts', 'ACCOUNTING'),

    ('sales.invoices.view', 'عرض فواتير المبيعات', 'View Sales Invoices', 'SALES'),
    ('sales.invoices.create', 'إنشاء فواتير المبيعات', 'Create Sales Invoices', 'SALES'),

    ('zatca.invoices.sign', 'توقيع وإرسال الفواتير الإلكترونية', 'Sign & Transmit ZATCA Invoices', 'ZATCA')
ON CONFLICT (code) DO NOTHING;

-- 6. Seed System Roles
INSERT INTO authz.roles (code, name_ar, name_en, description, is_system) VALUES
    ('SUPER_ADMIN', 'مدير النظام الشامل', 'Super Administrator', 'Full unrestricted enterprise control across all legal entities', TRUE),
    ('ORG_ADMIN', 'مدير الكيان القانوني', 'Legal Entity Administrator', 'Administrative control scoped to assigned legal entity', TRUE),
    ('CHIEF_ACCOUNTANT', 'رئيس الحسابات', 'Chief Accountant', 'Financial management, journal posting, and reporting', TRUE),
    ('SALES_OFFICER', 'مسؤول المبيعات', 'Sales Officer', 'Sales invoice generation and customer operations', TRUE),
    ('AUDITOR', 'مدقق مالي', 'Financial Auditor', 'Read-only financial and operational auditing', TRUE)
ON CONFLICT (code) DO NOTHING;

-- 7. Grant All Permissions to SUPER_ADMIN
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN'
ON CONFLICT DO NOTHING;

-- 8. Grant Scoped Permissions to CHIEF_ACCOUNTANT
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
JOIN authz.permissions p ON p.code IN (
    'org.legal_entities.view', 'org.branches.view', 'org.cost_centers.view',
    'gl.journals.view', 'gl.journals.post', 'gl.accounts.view',
    'org.number_sequences.generate'
)
WHERE r.code = 'CHIEF_ACCOUNTANT'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS authz.user_role_assignments;
DROP TABLE IF EXISTS authz.role_permissions;
DROP TABLE IF EXISTS authz.permissions;
DROP TABLE IF EXISTS authz.roles;
DROP SCHEMA IF EXISTS authz CASCADE;
-- +goose StatementEnd
