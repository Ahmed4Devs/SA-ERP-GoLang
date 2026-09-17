-- name: CreateContributionRule :one
INSERT INTO gosi.contribution_rules (
    id, legal_entity_id, code, name_en, name_ar, description_en, description_ar, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListContributionRules :many
SELECT * FROM gosi.contribution_rules
WHERE legal_entity_id = $1
ORDER BY code;

-- name: GetContributionRuleByCode :one
SELECT * FROM gosi.contribution_rules
WHERE legal_entity_id = $1 AND code = $2;

-- name: CreateRuleVersion :one
INSERT INTO gosi.contribution_rule_versions (
    id, rule_id, legal_entity_id, version_number, effective_from, effective_to,
    nationality_scope, employee_rate, employer_rate, minimum_basis, maximum_basis,
    calculation_config, legal_reference, notes, is_active
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
RETURNING *;

-- name: GetActiveRuleVersions :many
SELECT rv.*, cr.code as rule_code, cr.name_en as rule_name_en, cr.name_ar as rule_name_ar
FROM gosi.contribution_rule_versions rv
JOIN gosi.contribution_rules cr ON cr.id = rv.rule_id
WHERE rv.legal_entity_id = $1
  AND rv.is_active = true
  AND rv.effective_from <= $2::date
  AND (rv.effective_to IS NULL OR rv.effective_to >= $2::date)
ORDER BY cr.code;

-- name: GetRuleVersionByID :one
SELECT * FROM gosi.contribution_rule_versions WHERE id = $1;

-- name: ListRuleVersions :many
SELECT * FROM gosi.contribution_rule_versions
WHERE rule_id = $1
ORDER BY version_number DESC;
