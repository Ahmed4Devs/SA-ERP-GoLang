-- name: CreateOvertimeRequest :one
INSERT INTO attendance.overtime_requests (
    id,
    legal_entity_id,
    employee_id,
    request_date,
    overtime_date,
    overtime_hours,
    rate_multiplier,
    reason,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetOvertimeRequestByID :one
SELECT * FROM attendance.overtime_requests
WHERE id = $1 AND legal_entity_id = $2;

-- name: ListOvertimeRequests :many
SELECT o.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar
FROM attendance.overtime_requests o
JOIN hr.employees e ON o.employee_id = e.id
WHERE o.legal_entity_id = $1
ORDER BY o.overtime_date DESC;

-- name: UpdateOvertimeStatus :one
UPDATE attendance.overtime_requests
SET
    status = $3,
    approved_by_user_id = $4,
    approved_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: CreateAttendanceAdjustment :one
INSERT INTO attendance.adjustments (
    id,
    attendance_record_id,
    requested_by_employee_id,
    original_clock_in,
    original_clock_out,
    adjusted_clock_in,
    adjusted_clock_out,
    reason,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetAttendanceAdjustmentByID :one
SELECT * FROM attendance.adjustments
WHERE id = $1;

-- name: UpdateAdjustmentStatus :one
UPDATE attendance.adjustments
SET
    status = $2,
    reviewed_by_user_id = $3,
    reviewed_at = CURRENT_TIMESTAMP,
    review_notes = $4
WHERE id = $1
RETURNING *;
