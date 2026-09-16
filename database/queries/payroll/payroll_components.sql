-- name: CreateSalaryComponent :one
INSERT INTO payroll.salary_components (
    id, legal_entity_id, code, name_en, name_ar, component_type,
    is_taxable_or_gosi_basis, calculation_type, formula_expression, default_amount, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetSalaryComponentByID :one
SELECT * FROM payroll.salary_components
WHERE id = $1 LIMIT 1;

-- name: ListSalaryComponents :many
SELECT * FROM payroll.salary_components
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: CreateSalaryStructure :one
INSERT INTO payroll.salary_structures (
    id, legal_entity_id, code, name_en, name_ar, currency_code, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetSalaryStructureByID :one
SELECT * FROM payroll.salary_structures
WHERE id = $1 LIMIT 1;

-- name: ListSalaryStructures :many
SELECT * FROM payroll.salary_structures
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: CreateSalaryStructureLine :one
INSERT INTO payroll.salary_structure_lines (
    id, salary_structure_id, component_id, line_number, rate_percentage, fixed_amount
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListSalaryStructureLines :many
SELECT sl.*, sc.code as component_code, sc.name_en as component_name_en, sc.name_ar as component_name_ar, sc.component_type
FROM payroll.salary_structure_lines sl
JOIN payroll.salary_components sc ON sc.id = sl.component_id
WHERE sl.salary_structure_id = $1
ORDER BY sl.line_number ASC;

-- name: CreateEmployeeSalaryAssignment :one
INSERT INTO payroll.employee_salary_assignments (
    id, legal_entity_id, employee_id, salary_structure_id, effective_from, effective_to,
    basic_salary, housing_allowance, transport_allowance, other_allowances, gross_salary
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
) RETURNING *;

-- name: GetActiveEmployeeSalaryAssignment :one
SELECT * FROM payroll.employee_salary_assignments
WHERE employee_id = $1 AND (effective_to IS NULL OR effective_to >= $2)
ORDER BY effective_from DESC
LIMIT 1;
