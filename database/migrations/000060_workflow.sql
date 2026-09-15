-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS workflow;

-- 1. Workflow Definitions
CREATE TABLE IF NOT EXISTS workflow.definitions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    code                TEXT NOT NULL UNIQUE,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    module              TEXT NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflow_definitions_code ON workflow.definitions (code);
CREATE INDEX IF NOT EXISTS idx_workflow_definitions_module ON workflow.definitions (module);

-- 2. Versioned Workflow Definitions
CREATE TABLE IF NOT EXISTS workflow.definition_versions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    definition_id       UUID NOT NULL REFERENCES workflow.definitions(id) ON DELETE CASCADE,
    version             INTEGER NOT NULL DEFAULT 1,
    status              TEXT NOT NULL DEFAULT 'DRAFT',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at        TIMESTAMPTZ,

    CONSTRAINT ck_workflow_version_status
        CHECK (status IN ('DRAFT', 'PUBLISHED', 'ARCHIVED')),
    CONSTRAINT uq_workflow_def_version
        UNIQUE (definition_id, version)
);

CREATE INDEX IF NOT EXISTS idx_workflow_versions_lookup
ON workflow.definition_versions (definition_id, status);

-- 3. Workflow Steps
CREATE TABLE IF NOT EXISTS workflow.steps (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    version_id          UUID NOT NULL REFERENCES workflow.definition_versions(id) ON DELETE CASCADE,
    step_order          INTEGER NOT NULL,
    code                TEXT NOT NULL,
    name_ar             TEXT NOT NULL,
    name_en             TEXT NOT NULL,
    approval_type       TEXT NOT NULL DEFAULT 'ANY',
    timeout_hours       INTEGER DEFAULT 24,

    CONSTRAINT ck_workflow_step_approval_type
        CHECK (approval_type IN ('ANY', 'ALL')),
    CONSTRAINT uq_workflow_step_order
        UNIQUE (version_id, step_order)
);

CREATE INDEX IF NOT EXISTS idx_workflow_steps_version
ON workflow.steps (version_id, step_order ASC);

-- 4. Step Conditions (Thresholds, Legal Entity filter, etc.)
CREATE TABLE IF NOT EXISTS workflow.step_conditions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    step_id             UUID NOT NULL REFERENCES workflow.steps(id) ON DELETE CASCADE,
    field_name          TEXT NOT NULL,
    operator            TEXT NOT NULL,
    field_value         TEXT NOT NULL,

    CONSTRAINT ck_step_condition_operator
        CHECK (operator IN ('GT', 'GTE', 'LT', 'LTE', 'EQ', 'NEQ'))
);

CREATE INDEX IF NOT EXISTS idx_step_conditions_step
ON workflow.step_conditions (step_id);

-- 5. Step Approvers (Roles or Users)
CREATE TABLE IF NOT EXISTS workflow.step_approvers (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    step_id             UUID NOT NULL REFERENCES workflow.steps(id) ON DELETE CASCADE,
    approver_type       TEXT NOT NULL,
    role_id             UUID REFERENCES authz.roles(id) ON DELETE CASCADE,
    user_id             UUID REFERENCES iam.users(id) ON DELETE CASCADE,

    CONSTRAINT ck_step_approver_type
        CHECK (approver_type IN ('ROLE', 'USER'))
);

CREATE INDEX IF NOT EXISTS idx_step_approvers_step
ON workflow.step_approvers (step_id);

-- 6. Workflow Instances (Executing instance for a document)
CREATE TABLE IF NOT EXISTS workflow.instances (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    definition_id       UUID NOT NULL REFERENCES workflow.definitions(id),
    version_id          UUID NOT NULL REFERENCES workflow.definition_versions(id),
    legal_entity_id     UUID NOT NULL REFERENCES organization.legal_entities(id),
    resource_type       TEXT NOT NULL,
    resource_id         TEXT NOT NULL,
    initiator_id        UUID NOT NULL REFERENCES iam.users(id),
    status              TEXT NOT NULL DEFAULT 'PENDING',
    context_data        JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,

    CONSTRAINT ck_workflow_instance_status
        CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED', 'CANCELLED'))
);

CREATE INDEX IF NOT EXISTS idx_workflow_instances_resource
ON workflow.instances (resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_workflow_instances_entity
ON workflow.instances (legal_entity_id, status);
CREATE INDEX IF NOT EXISTS idx_workflow_instances_initiator
ON workflow.instances (initiator_id);

-- 7. Instance Steps Progression
CREATE TABLE IF NOT EXISTS workflow.instance_steps (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    instance_id         UUID NOT NULL REFERENCES workflow.instances(id) ON DELETE CASCADE,
    step_id             UUID NOT NULL REFERENCES workflow.steps(id),
    step_order          INTEGER NOT NULL,
    status              TEXT NOT NULL DEFAULT 'PENDING',
    started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at        TIMESTAMPTZ,

    CONSTRAINT ck_instance_step_status
        CHECK (status IN ('PENDING', 'IN_PROGRESS', 'APPROVED', 'REJECTED', 'SKIPPED')),
    CONSTRAINT uq_instance_step
        UNIQUE (instance_id, step_id)
);

CREATE INDEX IF NOT EXISTS idx_instance_steps_instance
ON workflow.instance_steps (instance_id, step_order ASC);

-- 8. Approval Requests (Assigned tasks for users/delegates)
CREATE TABLE IF NOT EXISTS workflow.approval_requests (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    instance_step_id    UUID NOT NULL REFERENCES workflow.instance_steps(id) ON DELETE CASCADE,
    assigned_user_id    UUID NOT NULL REFERENCES iam.users(id),
    status              TEXT NOT NULL DEFAULT 'PENDING',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT ck_approval_request_status
        CHECK (status IN ('PENDING', 'DECIDED', 'CANCELLED', 'DELEGATED'))
);

CREATE INDEX IF NOT EXISTS idx_approval_requests_user
ON workflow.approval_requests (assigned_user_id, status);

-- 9. Approval Decisions (Immutable audit log of decisions)
CREATE TABLE IF NOT EXISTS workflow.approval_decisions (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    request_id          UUID NOT NULL REFERENCES workflow.approval_requests(id) ON DELETE CASCADE,
    instance_id         UUID NOT NULL REFERENCES workflow.instances(id) ON DELETE CASCADE,
    actor_id            UUID NOT NULL REFERENCES iam.users(id),
    decision            TEXT NOT NULL,
    notes               TEXT,
    decided_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT ck_approval_decision
        CHECK (decision IN ('APPROVE', 'REJECT'))
);

CREATE INDEX IF NOT EXISTS idx_approval_decisions_instance
ON workflow.approval_decisions (instance_id, decided_at DESC);

-- 10. Delegations (Temporary proxy authorization)
CREATE TABLE IF NOT EXISTS workflow.delegations (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    delegator_id        UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    delegate_id         UUID NOT NULL REFERENCES iam.users(id) ON DELETE CASCADE,
    legal_entity_id     UUID REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    start_date          TIMESTAMPTZ NOT NULL,
    end_date            TIMESTAMPTZ NOT NULL,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    reason              TEXT,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT ck_delegation_dates
        CHECK (end_date > start_date),
    CONSTRAINT ck_no_self_delegation
        CHECK (delegator_id != delegate_id)
);

CREATE INDEX IF NOT EXISTS idx_workflow_delegations_active
ON workflow.delegations (delegator_id, is_active, start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_workflow_delegations_delegate
ON workflow.delegations (delegate_id, is_active);

-- 11. Seed Core Workflow Definitions & Published Versions
DO $$
DECLARE
    v_def_sales UUID := uuidv7();
    v_ver_sales UUID := uuidv7();
    v_step1_sales UUID := uuidv7();
    v_step2_sales UUID := uuidv7();

    v_def_journal UUID := uuidv7();
    v_ver_journal UUID := uuidv7();
    v_step1_journal UUID := uuidv7();
    v_step2_journal UUID := uuidv7();

    v_role_sales UUID;
    v_role_cfo UUID;
    v_role_accountant UUID;
    v_role_admin UUID;
BEGIN
    SELECT id INTO v_role_sales FROM authz.roles WHERE code = 'SALES_OFFICER';
    SELECT id INTO v_role_cfo FROM authz.roles WHERE code = 'CHIEF_ACCOUNTANT';
    SELECT id INTO v_role_accountant FROM authz.roles WHERE code = 'CHIEF_ACCOUNTANT';
    SELECT id INTO v_role_admin FROM authz.roles WHERE code = 'SUPER_ADMIN';

    -- A. Sales Invoice Approval Workflow
    INSERT INTO workflow.definitions (id, code, name_ar, name_en, module)
    VALUES (v_def_sales, 'SALES_INVOICE_APPROVAL', 'اعتماد فواتير المبيعات', 'Sales Invoice Approval', 'SALES')
    ON CONFLICT (code) DO NOTHING;

    INSERT INTO workflow.definition_versions (id, definition_id, version, status, published_at)
    VALUES (v_ver_sales, v_def_sales, 1, 'PUBLISHED', now())
    ON CONFLICT DO NOTHING;

    -- Step 1: Initial Sales Review
    INSERT INTO workflow.steps (id, version_id, step_order, code, name_ar, name_en, approval_type)
    VALUES (v_step1_sales, v_ver_sales, 1, 'SALES_REVIEW', 'مراجعة المبيعات المبدئية', 'Sales Review', 'ANY')
    ON CONFLICT DO NOTHING;

    IF v_role_sales IS NOT NULL THEN
        INSERT INTO workflow.step_approvers (step_id, approver_type, role_id)
        VALUES (v_step1_sales, 'ROLE', v_role_sales);
    END IF;

    -- Step 2: High-Value Financial Approval (Conditional: amount > 50,000 SAR)
    INSERT INTO workflow.steps (id, version_id, step_order, code, name_ar, name_en, approval_type)
    VALUES (v_step2_sales, v_ver_sales, 2, 'FINANCE_APPROVAL', 'اعتماد الإدارة المالية للمبالغ الكبرى', 'Finance High-Value Approval', 'ANY')
    ON CONFLICT DO NOTHING;

    INSERT INTO workflow.step_conditions (step_id, field_name, operator, field_value)
    VALUES (v_step2_sales, 'amount', 'GT', '50000.00');

    IF v_role_cfo IS NOT NULL THEN
        INSERT INTO workflow.step_approvers (step_id, approver_type, role_id)
        VALUES (v_step2_sales, 'ROLE', v_role_cfo);
    END IF;

    -- B. Journal Entry Posting Approval Workflow
    INSERT INTO workflow.definitions (id, code, name_ar, name_en, module)
    VALUES (v_def_journal, 'JOURNAL_ENTRY_APPROVAL', 'اعتماد وترحيل القيود اليومية', 'Journal Entry Approval', 'ACCOUNTING')
    ON CONFLICT (code) DO NOTHING;

    INSERT INTO workflow.definition_versions (id, definition_id, version, status, published_at)
    VALUES (v_ver_journal, v_def_journal, 1, 'PUBLISHED', now())
    ON CONFLICT DO NOTHING;

    INSERT INTO workflow.steps (id, version_id, step_order, code, name_ar, name_en, approval_type)
    VALUES (v_step1_journal, v_ver_journal, 1, 'ACCOUNTANT_REVIEW', 'تدقيق محاسب أول', 'Senior Accountant Review', 'ANY')
    ON CONFLICT DO NOTHING;

    IF v_role_accountant IS NOT NULL THEN
        INSERT INTO workflow.step_approvers (step_id, approver_type, role_id)
        VALUES (v_step1_journal, 'ROLE', v_role_accountant);
    END IF;

    INSERT INTO workflow.steps (id, version_id, step_order, code, name_ar, name_en, approval_type)
    VALUES (v_step2_journal, v_ver_journal, 2, 'CFO_FINAL_APPROVAL', 'الاعتماد المالي النهائي (CFO)', 'CFO Final Approval', 'ANY')
    ON CONFLICT DO NOTHING;

    INSERT INTO workflow.step_conditions (step_id, field_name, operator, field_value)
    VALUES (v_step2_journal, 'amount', 'GT', '100000.00');

    IF v_role_admin IS NOT NULL THEN
        INSERT INTO workflow.step_approvers (step_id, approver_type, role_id)
        VALUES (v_step2_journal, 'ROLE', v_role_admin);
    END IF;
END $$;

-- 12. Register Workflow Permissions in authz
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('workflow.requests.view', 'عرض طلبات الاعتماد في صندوق الوارد', 'View Workflow Approval Requests', 'WORKFLOW'),
    ('workflow.requests.decide', 'اتخاذ قرار الاعتماد أو الرفض للطلبات', 'Decide Workflow Approval Requests', 'WORKFLOW'),
    ('workflow.definitions.manage', 'إدارة وتصميم مسارات العمل والإصدارات', 'Manage Workflow Definitions', 'WORKFLOW'),
    ('workflow.delegations.manage', 'إدارة وتفعيل تفويض الصلاحيات', 'Manage Approval Delegations', 'WORKFLOW')
ON CONFLICT (code) DO NOTHING;

-- Grant to SUPER_ADMIN
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN' AND p.module = 'WORKFLOW'
ON CONFLICT DO NOTHING;

-- Grant inbox & decision to CHIEF_ACCOUNTANT
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
JOIN authz.permissions p ON p.code IN ('workflow.requests.view', 'workflow.requests.decide', 'workflow.delegations.manage')
WHERE r.code = 'CHIEF_ACCOUNTANT'
ON CONFLICT DO NOTHING;

-- Grant inbox & decision to SALES_OFFICER
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
JOIN authz.permissions p ON p.code IN ('workflow.requests.view', 'workflow.requests.decide')
WHERE r.code = 'SALES_OFFICER'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM authz.role_permissions
WHERE permission_id IN (SELECT id FROM authz.permissions WHERE module = 'WORKFLOW');

DELETE FROM authz.permissions WHERE module = 'WORKFLOW';

DROP TABLE IF EXISTS workflow.delegations;
DROP TABLE IF EXISTS workflow.approval_decisions;
DROP TABLE IF EXISTS workflow.approval_requests;
DROP TABLE IF EXISTS workflow.instance_steps;
DROP TABLE IF EXISTS workflow.instances;
DROP TABLE IF EXISTS workflow.step_approvers;
DROP TABLE IF EXISTS workflow.step_conditions;
DROP TABLE IF EXISTS workflow.steps;
DROP TABLE IF EXISTS workflow.definition_versions;
DROP TABLE IF EXISTS workflow.definitions;
DROP SCHEMA IF EXISTS workflow CASCADE;
-- +goose StatementEnd
