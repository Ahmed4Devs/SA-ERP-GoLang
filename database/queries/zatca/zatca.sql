-- name: GetEGSUnitByID :one
SELECT * FROM zatca.egs_units
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetEGSUnitByCustomID :one
SELECT * FROM zatca.egs_units
WHERE custom_id = $1 AND legal_entity_id = $2;

-- name: ListEGSUnits :many
SELECT * FROM zatca.egs_units
WHERE legal_entity_id = $1
ORDER BY created_at DESC;

-- name: CreateEGSUnit :one
INSERT INTO zatca.egs_units (
    id, legal_entity_id, branch_id, uuid, custom_id, model, serial_number,
    solution_name, vat_number, organization_name, organization_unit_name,
    country_code, invoice_type_mask, status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateEGSUnitStatus :one
UPDATE zatca.egs_units
SET status = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: GetActiveCSID :one
SELECT * FROM zatca.csids
WHERE egs_unit_id = $1 AND status = 'ACTIVE'
ORDER BY created_at DESC
LIMIT 1;

-- name: ListCSIDs :many
SELECT * FROM zatca.csids
WHERE egs_unit_id = $1
ORDER BY created_at DESC;

-- name: CreateCSID :one
INSERT INTO zatca.csids (
    id, egs_unit_id, csid_type, binary_security_token, secret,
    private_key_pem, public_key_pem, certificate_pem,
    request_id, compliance_request_id, expires_at, status, created_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11, $12, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetInvoiceSequence :one
SELECT * FROM zatca.invoice_sequences
WHERE egs_unit_id = $1 AND invoice_type = $2;

-- name: CreateInvoiceSequence :one
INSERT INTO zatca.invoice_sequences (
    id, egs_unit_id, invoice_type, last_icv, last_invoice_hash, updated_at
) VALUES (
    $1, $2, $3, $4, $5, CURRENT_TIMESTAMP
) RETURNING *;

-- name: IncrementInvoiceSequence :one
UPDATE zatca.invoice_sequences
SET last_icv = last_icv + 1,
    last_invoice_hash = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE egs_unit_id = $1 AND invoice_type = $2
RETURNING *;

-- name: GetDocumentByID :one
SELECT * FROM zatca.documents
WHERE id = $1 AND legal_entity_id = $2;

-- name: GetDocumentByUUID :one
SELECT * FROM zatca.documents
WHERE invoice_uuid = $1;

-- name: ListDocuments :many
SELECT * FROM zatca.documents
WHERE legal_entity_id = $1
ORDER BY issue_date DESC, issue_time DESC, created_at DESC;

-- name: CreateDocument :one
INSERT INTO zatca.documents (
    id, legal_entity_id, egs_unit_id, sales_invoice_id,
    invoice_uuid, invoice_number, invoice_type, invoice_subtype, submission_mode,
    icv, pih, issue_date, issue_time, ubl_xml, signed_xml,
    invoice_hash, digital_signature, public_key_hashing, qr_code_data,
    status, created_at, updated_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9,
    $10, $11, $12, $13, $14, $15,
    $16, $17, $18, $19,
    $20, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
) RETURNING *;

-- name: UpdateDocumentSigned :one
UPDATE zatca.documents
SET signed_xml = $3,
    invoice_hash = $4,
    digital_signature = $5,
    public_key_hashing = $6,
    qr_code_data = $7,
    status = $8,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: UpdateDocumentStatus :one
UPDATE zatca.documents
SET status = $3,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1 AND legal_entity_id = $2
RETURNING *;

-- name: CreateValidationResult :one
INSERT INTO zatca.validation_results (
    id, document_id, overall_status, rule_version, validated_at
) VALUES (
    $1, $2, $3, $4, CURRENT_TIMESTAMP
) RETURNING *;

-- name: CreateValidationMessage :one
INSERT INTO zatca.validation_messages (
    id, validation_result_id, message_type, code, category, message_text, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP
) RETURNING *;

-- name: GetValidationResultsByDocument :many
SELECT * FROM zatca.validation_results
WHERE document_id = $1
ORDER BY validated_at DESC;

-- name: GetValidationMessagesByResult :many
SELECT * FROM zatca.validation_messages
WHERE validation_result_id = $1
ORDER BY created_at ASC;

-- name: GetActiveConfiguration :one
SELECT * FROM zatca.configuration_versions
WHERE legal_entity_id = $1 AND is_active = true
ORDER BY effective_from DESC
LIMIT 1;

-- name: CountDocumentsByStatus :one
SELECT 
    COUNT(*)::BIGINT AS total_documents,
    COUNT(*) FILTER (WHERE status = 'CLEARED')::BIGINT AS cleared_count,
    COUNT(*) FILTER (WHERE status = 'REPORTED')::BIGINT AS reported_count,
    COUNT(*) FILTER (WHERE status = 'WARNING')::BIGINT AS warning_count,
    COUNT(*) FILTER (WHERE status = 'REJECTED' OR status = 'FAILED')::BIGINT AS failed_count
FROM zatca.documents
WHERE legal_entity_id = $1;
