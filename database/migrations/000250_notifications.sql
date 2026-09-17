-- +goose Up
-- Notifications & ERP Messaging Schema (Phase 28)
-- Compliant with Master Specification Section 40.29, Section 56, and Section 99

CREATE SCHEMA IF NOT EXISTS notifications;

-- 1. Notifications Master (سجل التنبيهات والإشعارات)
CREATE TABLE IF NOT EXISTS notifications.notifications (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    recipient_user_id UUID REFERENCES iam.users(id) ON DELETE CASCADE,
    recipient_email VARCHAR(255),
    channel VARCHAR(50) NOT NULL DEFAULT 'IN_APP' CHECK (channel IN ('IN_APP', 'EMAIL', 'SMS')),
    title_ar TEXT NOT NULL,
    title_en TEXT NOT NULL,
    body_ar TEXT NOT NULL,
    body_en TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'SYSTEM' CHECK (category IN ('WORKFLOW', 'ZATCA', 'FINANCE', 'INVENTORY', 'PAYROLL', 'SYSTEM')),
    priority VARCHAR(20) NOT NULL DEFAULT 'NORMAL' CHECK (priority IN ('LOW', 'NORMAL', 'HIGH', 'URGENT')),
    action_url TEXT,
    entity_type VARCHAR(100),
    entity_id UUID,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'SENT', 'FAILED', 'READ')),
    error_message TEXT,
    sent_at TIMESTAMPTZ,
    read_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_status ON notifications.notifications(recipient_user_id, status);
CREATE INDEX IF NOT EXISTS idx_notifications_le_cat ON notifications.notifications(legal_entity_id, category);
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications.notifications(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_entity ON notifications.notifications(entity_type, entity_id);

-- 2. Notification Templates (قوالب الإشعارات المعيارية)
CREATE TABLE IF NOT EXISTS notifications.templates (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    code VARCHAR(100) UNIQUE NOT NULL,
    title_ar TEXT NOT NULL,
    title_en TEXT NOT NULL,
    body_ar TEXT NOT NULL,
    body_en TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'SYSTEM',
    channel VARCHAR(50) NOT NULL DEFAULT 'IN_APP',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notification_templates_code ON notifications.templates(code);

-- 3. Permissions Seeding
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'notifications.read',   'Read Notifications',   'استعراض التنبيهات والإشعارات',  'NOTIFICATIONS', CURRENT_TIMESTAMP),
    (uuidv7(), 'notifications.send',   'Send Notifications',   'إرسال التنبيهات والرسائل',      'NOTIFICATIONS', CURRENT_TIMESTAMP),
    (uuidv7(), 'notifications.manage', 'Manage Notifications', 'إدارة قوالب وإعدادات التنبيهات', 'NOTIFICATIONS', CURRENT_TIMESTAMP)
ON CONFLICT (code) DO NOTHING;

-- 4. Seed Default Notification Templates & Initial Welcome Alert
-- +goose StatementBegin
CREATE OR REPLACE PROCEDURE notifications.seed_default_notifications_data(p_legal_entity_id UUID)
LANGUAGE plpgsql
AS $$
DECLARE
    v_admin_id UUID;
BEGIN
    -- Seed standard system templates
    INSERT INTO notifications.templates (id, code, title_ar, title_en, body_ar, body_en, category, channel, is_active, created_at)
    VALUES
        (uuidv7(), 'WORKFLOW_APPROVAL_REQUEST',
         'طلب اعتماد مسار عمل جديد', 'New Workflow Approval Request',
         'يوجد مستند جديد رقم {{doc_no}} يتطلب موافقتك واعتمادك.', 'A new document {{doc_no}} is waiting for your review and approval.',
         'WORKFLOW', 'IN_APP', true, CURRENT_TIMESTAMP),

        (uuidv7(), 'ZATCA_INVOICE_CLEARED',
         'تم اعتماد الفاتورة من هيئة الزكاة (ZATCA Cleared)', 'Invoice Cleared by ZATCA',
         'تم اعتماد الفاتورة الضريبية رقم {{invoice_no}} بنجاح وإصدار الختم الرقمي.', 'Tax invoice {{invoice_no}} has been cleared successfully by ZATCA with clearance stamp.',
         'ZATCA', 'IN_APP', true, CURRENT_TIMESTAMP),

        (uuidv7(), 'BUDGET_THRESHOLD_WARNING',
         'تحذير تجاوز سقف الميزانية التقديرية', 'Budget Threshold Warning',
         'تجاوز بند الميزانية {{budget_line}} نسبة 90% من السقف المالي المعتمد.', 'Budget line {{budget_line}} has exceeded 90% of its approved financial threshold.',
         'FINANCE', 'IN_APP', true, CURRENT_TIMESTAMP),

        (uuidv7(), 'PAYROLL_PROCESSED',
         'تم احتساب مسير الرواتب الشهري', 'Monthly Payroll Processed',
         'تم احتساب واعتماد مسير رواتب شهر {{month}} وتوليد ملف حماية الأجور WPS.', 'Monthly payroll for {{month}} has been processed and WPS file generated.',
         'PAYROLL', 'IN_APP', true, CURRENT_TIMESTAMP)
    ON CONFLICT (code) DO NOTHING;

    -- Look up system admin to receive initial welcome notifications
    SELECT id INTO v_admin_id FROM iam.users WHERE username = 'admin' LIMIT 1;

    IF v_admin_id IS NOT NULL THEN
        -- Check if welcome notification exists
        IF NOT EXISTS (SELECT 1 FROM notifications.notifications WHERE legal_entity_id = p_legal_entity_id AND recipient_user_id = v_admin_id) THEN
            INSERT INTO notifications.notifications (
                id, legal_entity_id, recipient_user_id, recipient_email, channel,
                title_ar, title_en, body_ar, body_en, category, priority,
                action_url, status, sent_at, created_at
            ) VALUES (
                uuidv7(), p_legal_entity_id, v_admin_id, 'admin@erp.sa', 'IN_APP',
                'مرحباً بك في منصة تخطيط الموارد السعودية', 'Welcome to Saudi Enterprise ERP Platform',
                'تم تفعيل وإعداد كافة وحدات النظام بنجاح وجاهزية الفوترة الإلكترونية والتأمينات الاجتماعية.',
                'All enterprise modules, ZATCA Phase 2 e-invoicing, and GOSI compliance engines are operational.',
                'SYSTEM', 'NORMAL', '/dashboard', 'SENT', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
            );
        END IF;
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP PROCEDURE IF EXISTS notifications.seed_default_notifications_data(UUID);
DROP TABLE IF EXISTS notifications.templates CASCADE;
DROP TABLE IF EXISTS notifications.notifications CASCADE;
DROP SCHEMA IF EXISTS notifications CASCADE;
-- +goose StatementEnd
