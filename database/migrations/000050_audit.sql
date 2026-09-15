-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS audit;

-- 1. Canonical Audit Events (Append-only, Cryptographically hash-chained)
CREATE TABLE IF NOT EXISTS audit.events (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    legal_entity_id     UUID REFERENCES organization.legal_entities(id) ON DELETE SET NULL,
    action              TEXT NOT NULL,
    resource_type       TEXT NOT NULL,
    resource_id         TEXT NOT NULL,
    payload             JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address          TEXT,
    user_agent          TEXT,
    previous_hash       TEXT,
    record_hash         TEXT NOT NULL,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_events_occurred_at ON audit.events (occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_entity ON audit.events (legal_entity_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_user ON audit.events (user_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_events_resource ON audit.events (resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_events_action ON audit.events (action);

-- 2. Canonical Entity State Changes (Before & After Diff)
CREATE TABLE IF NOT EXISTS audit.entity_changes (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    event_id            UUID NOT NULL REFERENCES audit.events(id) ON DELETE RESTRICT,
    table_name          TEXT NOT NULL,
    record_id           TEXT NOT NULL,
    before_state        JSONB,
    after_state         JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_entity_changes_event ON audit.entity_changes (event_id);
CREATE INDEX IF NOT EXISTS idx_audit_entity_changes_record ON audit.entity_changes (table_name, record_id);

-- 3. Canonical Security Events (Authentication, Authorization, Abuse prevention)
CREATE TABLE IF NOT EXISTS audit.security_events (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    event_type          TEXT NOT NULL,
    status              TEXT NOT NULL,
    details             JSONB NOT NULL DEFAULT '{}'::jsonb,
    ip_address          TEXT,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_sec_events_type ON audit.security_events (event_type, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_sec_events_user ON audit.security_events (user_id, occurred_at DESC);

-- 4. Canonical Access Events (HTTP Access Logs with Latency & Status)
CREATE TABLE IF NOT EXISTS audit.access_events (
    id                  UUID PRIMARY KEY DEFAULT uuidv7(),
    user_id             UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    resource_path       TEXT NOT NULL,
    method              TEXT NOT NULL,
    status_code         INTEGER NOT NULL,
    ip_address          TEXT,
    latency_ms          INTEGER NOT NULL DEFAULT 0,
    occurred_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_audit_acc_events_path ON audit.access_events (resource_path, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_acc_events_user ON audit.access_events (user_id, occurred_at DESC);

-- 5. Strict Append-Only Trigger: Block all UPDATE and DELETE operations
CREATE OR REPLACE FUNCTION audit.prevent_audit_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Audit records are strictly append-only. UPDATE and DELETE operations are forbidden on %.%', TG_TABLE_SCHEMA, TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_no_update_delete_events ON audit.events;
CREATE TRIGGER trg_no_update_delete_events
BEFORE UPDATE OR DELETE ON audit.events
FOR EACH ROW EXECUTE FUNCTION audit.prevent_audit_mutation();

DROP TRIGGER IF EXISTS trg_no_update_delete_entity_changes ON audit.entity_changes;
CREATE TRIGGER trg_no_update_delete_entity_changes
BEFORE UPDATE OR DELETE ON audit.entity_changes
FOR EACH ROW EXECUTE FUNCTION audit.prevent_audit_mutation();

DROP TRIGGER IF EXISTS trg_no_update_delete_security_events ON audit.security_events;
CREATE TRIGGER trg_no_update_delete_security_events
BEFORE UPDATE OR DELETE ON audit.security_events
FOR EACH ROW EXECUTE FUNCTION audit.prevent_audit_mutation();

DROP TRIGGER IF EXISTS trg_no_update_delete_access_events ON audit.access_events;
CREATE TRIGGER trg_no_update_delete_access_events
BEFORE UPDATE OR DELETE ON audit.access_events
FOR EACH ROW EXECUTE FUNCTION audit.prevent_audit_mutation();

-- 6. Register Audit Permissions in authz
INSERT INTO authz.permissions (code, name_ar, name_en, module) VALUES
    ('audit.events.view', 'عرض سجلات التدقيق المالي والأمني', 'View Financial & Security Audit Logs', 'AUDIT'),
    ('audit.integrity.verify', 'التحقق من سلامة السلسلة التشفيرية لسجل التدقيق', 'Verify Audit Cryptographic Integrity', 'AUDIT'),
    ('audit.security_events.view', 'عرض سجل الأحداث الأمنية والمخاطر', 'View Security Events & Alerts', 'AUDIT')
ON CONFLICT (code) DO NOTHING;

-- Grant to SUPER_ADMIN
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
CROSS JOIN authz.permissions p
WHERE r.code = 'SUPER_ADMIN' AND p.module = 'AUDIT'
ON CONFLICT DO NOTHING;

-- Grant to AUDITOR
INSERT INTO authz.role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM authz.roles r
JOIN authz.permissions p ON p.code IN ('audit.events.view', 'audit.integrity.verify', 'audit.security_events.view')
WHERE r.code = 'AUDITOR'
ON CONFLICT DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM authz.role_permissions
WHERE permission_id IN (SELECT id FROM authz.permissions WHERE module = 'AUDIT');

DELETE FROM authz.permissions WHERE module = 'AUDIT';

DROP TRIGGER IF EXISTS trg_no_update_delete_access_events ON audit.access_events;
DROP TRIGGER IF EXISTS trg_no_update_delete_security_events ON audit.security_events;
DROP TRIGGER IF EXISTS trg_no_update_delete_entity_changes ON audit.entity_changes;
DROP TRIGGER IF EXISTS trg_no_update_delete_events ON audit.events;

DROP FUNCTION IF EXISTS audit.prevent_audit_mutation();

DROP TABLE IF EXISTS audit.access_events;
DROP TABLE IF EXISTS audit.security_events;
DROP TABLE IF EXISTS audit.entity_changes;
DROP TABLE IF EXISTS audit.events;
DROP SCHEMA IF EXISTS audit CASCADE;
-- +goose StatementEnd
