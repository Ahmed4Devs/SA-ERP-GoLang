-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS attendance;

-- 1. SHIFTS
CREATE TABLE IF NOT EXISTS attendance.shifts (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    code                        TEXT NOT NULL,
    name_en                     TEXT NOT NULL,
    name_ar                     TEXT NOT NULL,
    start_time                  TIME NOT NULL,
    end_time                    TIME NOT NULL,
    grace_period_minutes        INT NOT NULL DEFAULT 15,
    break_duration_minutes      INT NOT NULL DEFAULT 60,
    half_day_threshold_minutes  INT NOT NULL DEFAULT 240,
    is_night_shift              BOOLEAN NOT NULL DEFAULT FALSE,
    work_days                   JSONB NOT NULL DEFAULT '["SUNDAY","MONDAY","TUESDAY","WEDNESDAY","THURSDAY"]'::jsonb,
    is_active                   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_attendance_shifts_code UNIQUE (legal_entity_id, code)
);

CREATE INDEX IF NOT EXISTS idx_attendance_shifts_entity ON attendance.shifts(legal_entity_id);

-- 2. SHIFT ASSIGNMENTS
CREATE TABLE IF NOT EXISTS attendance.shift_assignments (
    id              UUID PRIMARY KEY DEFAULT uuidv7(),
    employee_id     UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    shift_id        UUID NOT NULL REFERENCES attendance.shifts(id) ON DELETE RESTRICT,
    valid_from      DATE NOT NULL,
    valid_to        DATE,
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attendance_assignments_emp ON attendance.shift_assignments(employee_id, is_active);

-- 3. ATTENDANCE RECORDS (DAILY TIMESHEET)
CREATE TABLE IF NOT EXISTS attendance.records (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    employee_id                 UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    shift_id                    UUID REFERENCES attendance.shifts(id) ON DELETE SET NULL,
    work_date                   DATE NOT NULL,
    clock_in                    TIMESTAMPTZ,
    clock_out                   TIMESTAMPTZ,
    check_in_method             TEXT NOT NULL DEFAULT 'MANUAL' CHECK (check_in_method IN ('MANUAL', 'BIOMETRIC', 'MOBILE_GPS', 'SYSTEM')),
    check_out_method            TEXT CHECK (check_out_method IN ('MANUAL', 'BIOMETRIC', 'MOBILE_GPS', 'SYSTEM')),
    expected_work_hours         NUMERIC(5,2) NOT NULL DEFAULT 8.00,
    actual_work_hours           NUMERIC(5,2) NOT NULL DEFAULT 0.00,
    late_minutes                INT NOT NULL DEFAULT 0,
    early_departure_minutes     INT NOT NULL DEFAULT 0,
    overtime_minutes            INT NOT NULL DEFAULT 0,
    status                      TEXT NOT NULL DEFAULT 'PRESENT' CHECK (status IN ('PRESENT', 'ABSENT', 'HALF_DAY', 'ON_LEAVE', 'REST_DAY', 'HOLIDAY')),
    notes                       TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_attendance_records_emp_date UNIQUE (employee_id, work_date)
);

CREATE INDEX IF NOT EXISTS idx_attendance_records_entity_date ON attendance.records(legal_entity_id, work_date);
CREATE INDEX IF NOT EXISTS idx_attendance_records_emp ON attendance.records(employee_id, work_date);

-- 4. ATTENDANCE ADJUSTMENTS
CREATE TABLE IF NOT EXISTS attendance.adjustments (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    attendance_record_id        UUID NOT NULL REFERENCES attendance.records(id) ON DELETE CASCADE,
    requested_by_employee_id    UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    original_clock_in           TIMESTAMPTZ,
    original_clock_out          TIMESTAMPTZ,
    adjusted_clock_in           TIMESTAMPTZ NOT NULL,
    adjusted_clock_out          TIMESTAMPTZ NOT NULL,
    reason                      TEXT NOT NULL,
    status                      TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    reviewed_by_user_id         UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    reviewed_at                 TIMESTAMPTZ,
    review_notes                TEXT,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attendance_adjustments_rec ON attendance.adjustments(attendance_record_id);

-- 5. OVERTIME REQUESTS (SAUDI LABOR LAW ARTICLE 107 - 1.5X RATE)
CREATE TABLE IF NOT EXISTS attendance.overtime_requests (
    id                          UUID PRIMARY KEY DEFAULT uuidv7(),
    legal_entity_id             UUID NOT NULL REFERENCES organization.legal_entities(id) ON DELETE CASCADE,
    employee_id                 UUID NOT NULL REFERENCES hr.employees(id) ON DELETE CASCADE,
    request_date                DATE NOT NULL,
    overtime_date               DATE NOT NULL,
    overtime_hours              NUMERIC(5,2) NOT NULL,
    rate_multiplier             NUMERIC(3,2) NOT NULL DEFAULT 1.50,
    reason                      TEXT NOT NULL,
    status                      TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    approved_by_user_id         UUID REFERENCES iam.users(id) ON DELETE SET NULL,
    approved_at                 TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attendance_overtime_entity ON attendance.overtime_requests(legal_entity_id, overtime_date);

-- 6. PERMISSIONS
INSERT INTO authz.permissions (id, code, name_en, name_ar, module, created_at)
VALUES
    (uuidv7(), 'attendance.shift.read',      'View Attendance Shifts',    'استعراض الورديات وجداول الدوام', 'ATTENDANCE', now()),
    (uuidv7(), 'attendance.shift.manage',    'Manage Attendance Shifts',  'إدارة الورديات وتعيين الموظفين',  'ATTENDANCE', now()),
    (uuidv7(), 'attendance.record.read',     'View Attendance Records',   'استعراض سجلات الحضور والانصراف',  'ATTENDANCE', now()),
    (uuidv7(), 'attendance.record.clock',    'Clock In and Out',          'تسجيل الحضور والانصراف',          'ATTENDANCE', now()),
    (uuidv7(), 'attendance.adjust.manage',   'Manage Adjustments',        'إدارة واعتماد تعديلات الحضور',   'ATTENDANCE', now()),
    (uuidv7(), 'attendance.overtime.manage', 'Manage Overtime Requests',  'إدارة واعتماد العمل الإضافي',     'ATTENDANCE', now())
ON CONFLICT (code) DO NOTHING;

-- 7. SEED DEFAULT ATTENDANCE DATA FUNCTION
CREATE OR REPLACE FUNCTION attendance.seed_default_attendance_data(p_legal_entity_id UUID)
RETURNS VOID AS $$
DECLARE
    v_shift_std_id UUID;
    v_shift_flex_id UUID;
    v_emp_id UUID;
BEGIN
    -- 1. Standard Shift (8 hrs: 08:00 - 17:00, 1 hr break, 15 min grace)
    INSERT INTO attendance.shifts (
        id, legal_entity_id, code, name_en, name_ar, start_time, end_time,
        grace_period_minutes, break_duration_minutes, half_day_threshold_minutes, is_night_shift
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'SHIFT-STD', 'Standard General Shift (08:00 - 17:00)', 'الوردية الصباحية العامة (08:00 - 17:00)',
        '08:00:00'::time, '17:00:00'::time, 15, 60, 240, FALSE
    )
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en
    RETURNING id INTO v_shift_std_id;

    -- 2. Flexible Shift (09:00 - 18:00)
    INSERT INTO attendance.shifts (
        id, legal_entity_id, code, name_en, name_ar, start_time, end_time,
        grace_period_minutes, break_duration_minutes, half_day_threshold_minutes, is_night_shift
    ) VALUES (
        uuidv7(), p_legal_entity_id, 'SHIFT-FLEX', 'Flexible Tech Shift (09:00 - 18:00)', 'الوردية المرنة لقطاع التقنية (09:00 - 18:00)',
        '09:00:00'::time, '18:00:00'::time, 30, 60, 240, FALSE
    )
    ON CONFLICT (legal_entity_id, code) DO UPDATE SET name_en = EXCLUDED.name_en
    RETURNING id INTO v_shift_flex_id;

    -- 3. Assign active employees to standard shift if not assigned
    FOR v_emp_id IN (SELECT id FROM hr.employees WHERE legal_entity_id = p_legal_entity_id AND employment_status = 'ACTIVE') LOOP
        IF NOT EXISTS (SELECT 1 FROM attendance.shift_assignments WHERE employee_id = v_emp_id AND is_active = TRUE) THEN
            INSERT INTO attendance.shift_assignments (
                id, employee_id, shift_id, valid_from, is_active
            ) VALUES (
                uuidv7(), v_emp_id, v_shift_std_id, CURRENT_DATE - INTERVAL '30 days', TRUE
            );
        END IF;
    END LOOP;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS attendance.seed_default_attendance_data(UUID);
DROP TABLE IF EXISTS attendance.overtime_requests CASCADE;
DROP TABLE IF EXISTS attendance.adjustments CASCADE;
DROP TABLE IF EXISTS attendance.records CASCADE;
DROP TABLE IF EXISTS attendance.shift_assignments CASCADE;
DROP TABLE IF EXISTS attendance.shifts CASCADE;
DROP SCHEMA IF EXISTS attendance CASCADE;
-- +goose StatementEnd
