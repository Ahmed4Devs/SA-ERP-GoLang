-- name: ListJournalsByLegalEntity :many
SELECT id, legal_entity_id, code, name_en, name_ar, journal_type, is_active, created_at, updated_at, version
FROM accounting.journals
WHERE legal_entity_id = $1
ORDER BY code ASC;

-- name: GetJournalByID :one
SELECT id, legal_entity_id, code, name_en, name_ar, journal_type, is_active, created_at, updated_at, version
FROM accounting.journals
WHERE id = $1;

-- name: GetJournalByCode :one
SELECT id, legal_entity_id, code, name_en, name_ar, journal_type, is_active, created_at, updated_at, version
FROM accounting.journals
WHERE legal_entity_id = $1 AND code = $2;

-- name: CreateJournal :one
INSERT INTO accounting.journals (
    legal_entity_id, code, name_en, name_ar, journal_type, is_active
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING id, legal_entity_id, code, name_en, name_ar, journal_type, is_active, created_at, updated_at, version;
