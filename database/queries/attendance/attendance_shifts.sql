-- name: CreateShift :one
INSERT INTO attendance.shifts (
    id,
    legal_entity_id,
    code,
    name_en,
    name_ar,
    start_time,
    end_time,
    grace_period_minutes,
    break_duration_minutes,
    half_day_threshold_minutes,
    is_night_shift,
    work_days,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetShiftByID :one
SELECT * FROM attendance.shifts
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetShiftByCode :one
SELECT * FROM attendance.shifts
WHERE legal_entity_id = $1 AND code = $2;

-- name: ListShifts :many
SELECT * FROM attendance.shifts
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: CreateShiftAssignment :one
INSERT INTO attendance.shift_assignments (
    id,
    employee_id,
    shift_id,
    valid_from,
    valid_to,
    is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetActiveShiftAssignment :one
SELECT sa.*, s.name_en AS shift_name_en, s.name_ar AS shift_name_ar, s.start_time, s.end_time, s.grace_period_minutes
FROM attendance.shift_assignments sa
JOIN attendance.shifts s ON sa.shift_id = s.id
WHERE sa.employee_id = $1 AND sa.is_active = TRUE
LIMIT 1;

-- name: ListShiftAssignments :many
SELECT sa.*, s.code AS shift_code, s.name_en AS shift_name_en, s.name_ar AS shift_name_ar
FROM attendance.shift_assignments sa
JOIN attendance.shifts s ON sa.shift_id = s.id
WHERE sa.employee_id = $1
ORDER BY sa.valid_from DESC;
