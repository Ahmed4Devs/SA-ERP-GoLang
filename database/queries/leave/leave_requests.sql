-- name: CreateLeaveRequest :one
INSERT INTO leave.requests (
    id,
    legal_entity_id,
    employee_id,
    leave_type_id,
    request_number,
    start_date,
    end_date,
    total_days,
    return_date,
    replacement_employee_id,
    reason,
    attachment_url,
    status
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
)
RETURNING *;

-- name: GetLeaveRequestByID :one
SELECT * FROM leave.requests
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetLeaveRequestByNumber :one
SELECT * FROM leave.requests
WHERE legal_entity_id = $1 AND request_number = $2;

-- name: ListLeaveRequests :many
SELECT r.*, e.employee_number, e.first_name_en, e.last_name_en, e.first_name_ar, e.last_name_ar,
       t.code AS leave_type_code, t.name_en AS leave_type_name_en, t.name_ar AS leave_type_name_ar
FROM leave.requests r
JOIN hr.employees e ON r.employee_id = e.id
JOIN leave.types t ON r.leave_type_id = t.id
WHERE r.legal_entity_id = $1
ORDER BY r.start_date DESC;

-- name: ListEmployeeLeaveRequests :many
SELECT r.*, t.code AS leave_type_code, t.name_en AS leave_type_name_en, t.name_ar AS leave_type_name_ar
FROM leave.requests r
JOIN leave.types t ON r.leave_type_id = t.id
WHERE r.employee_id = $1
ORDER BY r.start_date DESC;

-- name: UpdateLeaveRequestStatus :one
UPDATE leave.requests
SET
    status = $3,
    reviewed_by_user_id = $4,
    reviewed_at = CURRENT_TIMESTAMP,
    review_notes = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;
