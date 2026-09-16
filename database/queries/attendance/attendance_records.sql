-- name: CreateAttendanceRecord :one
INSERT INTO attendance.records (
    id,
    legal_entity_id,
    employee_id,
    shift_id,
    work_date,
    clock_in,
    clock_out,
    check_in_method,
    check_out_method,
    expected_work_hours,
    actual_work_hours,
    late_minutes,
    early_departure_minutes,
    overtime_minutes,
    status,
    notes
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
)
RETURNING *;

-- name: GetAttendanceRecordByID :one
SELECT * FROM attendance.records
WHERE id = $1;

-- name: GetAttendanceRecordByEmployeeAndDate :one
SELECT * FROM attendance.records
WHERE employee_id = $1 AND work_date = $2;

-- name: UpdateAttendanceRecordClockOut :one
UPDATE attendance.records
SET
    clock_out = $3,
    check_out_method = $4,
    actual_work_hours = $5,
    early_departure_minutes = $6,
    overtime_minutes = $7,
    status = $8,
    notes = $9,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: ListAttendanceRecordsByDate :many
SELECT r.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar
FROM attendance.records r
JOIN hr.employees e ON r.employee_id = e.id
WHERE r.legal_entity_id = $1 AND r.work_date = $2
ORDER BY e.employee_number ASC;

-- name: ListEmployeeAttendanceRecords :many
SELECT * FROM attendance.records
WHERE employee_id = $1 AND work_date >= $2 AND work_date <= $3
ORDER BY work_date DESC;
