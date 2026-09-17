package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"erp/internal/documents/domain"
	"erp/internal/documents/ports"
	"erp/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// DocumentRepository implements ports.Repository using PostgreSQL.
type DocumentRepository struct {
	db *database.DB
}

var _ ports.Repository = (*DocumentRepository)(nil)

// NewDocumentRepository creates a new DocumentRepository instance.
func NewDocumentRepository(db *database.DB) *DocumentRepository {
	return &DocumentRepository{db: db}
}

// CreateFile atomically creates a file record and its initial version.
func (r *DocumentRepository) CreateFile(ctx context.Context, file *domain.File, version *domain.FileVersion) (*domain.File, error) {
	if file.ID == uuid.Nil {
		file.ID = uuid.New()
	}
	if version.ID == uuid.Nil {
		version.ID = uuid.New()
	}
	version.FileID = file.ID
	version.VersionNumber = 1

	now := time.Now().UTC()
	if file.CreatedAt.IsZero() {
		file.CreatedAt = now
	}
	file.UpdatedAt = now
	if version.CreatedAt.IsZero() {
		version.CreatedAt = now
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert file record (initially with current_version_id = NULL)
	fileQuery := `
		INSERT INTO documents.files (
			id, legal_entity_id, title, description, category, folder_path,
			current_version_id, is_archived, created_by_user_id, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, NULL, $7, $8, $9, $10)
	`
	_, err = tx.Exec(ctx, fileQuery,
		file.ID,
		file.LegalEntityID,
		file.Title,
		file.Description,
		string(file.Category),
		file.FolderPath,
		file.IsArchived,
		file.CreatedByUserID,
		file.CreatedAt,
		file.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document file: %w", err)
	}

	// 2. Insert version record
	metaJSON, err := json.Marshal(version.Metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	verQuery := `
		INSERT INTO documents.file_versions (
			id, file_id, version_number, storage_provider, storage_key,
			original_filename, mime_type, byte_size, sha256_hash, metadata,
			uploaded_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.Exec(ctx, verQuery,
		version.ID,
		version.FileID,
		version.VersionNumber,
		version.StorageProvider,
		version.StorageKey,
		version.OriginalFilename,
		version.MimeType,
		version.ByteSize,
		version.SHA256Hash,
		metaJSON,
		version.UploadedBy,
		version.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert document version: %w", err)
	}

	// 3. Update file record with current_version_id
	updateQuery := `UPDATE documents.files SET current_version_id = $1 WHERE id = $2`
	_, err = tx.Exec(ctx, updateQuery, version.ID, file.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to update current_version_id: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit document creation: %w", err)
	}

	file.CurrentVersionID = &version.ID
	file.CurrentVersion = version
	return file, nil
}

// GetFileByID retrieves a single file and its active version by ID.
func (r *DocumentRepository) GetFileByID(ctx context.Context, legalEntityID, fileID uuid.UUID) (*domain.File, error) {
	query := `
		SELECT 
			f.id, f.legal_entity_id, f.title, f.description, f.category, f.folder_path,
			f.current_version_id, f.is_archived, f.created_by_user_id, f.created_at, f.updated_at,
			v.id, v.version_number, v.storage_provider, v.storage_key, v.original_filename,
			v.mime_type, v.byte_size, v.sha256_hash, v.metadata, v.uploaded_by, v.created_at
		FROM documents.files f
		LEFT JOIN documents.file_versions v ON f.current_version_id = v.id
		WHERE f.legal_entity_id = $1 AND f.id = $2
	`

	var file domain.File
	var cat string
	var desc *string
	var currVerID *uuid.UUID
	var vID *uuid.UUID
	var vNum *int
	var vProvider, vKey, vFilename, vMime, vHash *string
	var vSize *int64
	var vMeta []byte
	var vUploader *uuid.UUID
	var vCreated *time.Time

	err := r.db.Pool.QueryRow(ctx, query, legalEntityID, fileID).Scan(
		&file.ID,
		&file.LegalEntityID,
		&file.Title,
		&desc,
		&cat,
		&file.FolderPath,
		&currVerID,
		&file.IsArchived,
		&file.CreatedByUserID,
		&file.CreatedAt,
		&file.UpdatedAt,
		&vID,
		&vNum,
		&vProvider,
		&vKey,
		&vFilename,
		&vMime,
		&vSize,
		&vHash,
		&vMeta,
		&vUploader,
		&vCreated,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFileNotFound
		}
		return nil, fmt.Errorf("failed to get document file: %w", err)
	}

	if desc != nil {
		file.Description = *desc
	}
	file.Category = domain.FileCategory(cat)
	file.CurrentVersionID = currVerID

	if vID != nil && *vID != uuid.Nil {
		ver := domain.FileVersion{
			ID:               *vID,
			FileID:           file.ID,
			VersionNumber:    *vNum,
			StorageProvider:  *vProvider,
			StorageKey:       *vKey,
			OriginalFilename: *vFilename,
			MimeType:         *vMime,
			ByteSize:         *vSize,
			SHA256Hash:       *vHash,
			UploadedBy:       vUploader,
			CreatedAt:        *vCreated,
		}
		if len(vMeta) > 0 {
			_ = json.Unmarshal(vMeta, &ver.Metadata)
		}
		file.CurrentVersion = &ver
	}

	// Fetch attachments for this file
	attachments, err := r.ListAttachmentsByFile(ctx, file.ID)
	if err == nil {
		file.Attachments = attachments
	}

	return &file, nil
}

// ListFiles returns filtered files for a legal entity.
func (r *DocumentRepository) ListFiles(ctx context.Context, filter domain.FileFilter) ([]domain.File, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, fmt.Sprintf("f.legal_entity_id = $%d", argIdx))
	args = append(args, filter.LegalEntityID)
	argIdx++

	if !filter.IncludeArchived {
		conditions = append(conditions, "f.is_archived = false")
	}

	if filter.Category != nil && *filter.Category != "" {
		conditions = append(conditions, fmt.Sprintf("f.category = $%d", argIdx))
		args = append(args, string(*filter.Category))
		argIdx++
	}

	if filter.FolderPath != nil && *filter.FolderPath != "" {
		conditions = append(conditions, fmt.Sprintf("f.folder_path = $%d", argIdx))
		args = append(args, *filter.FolderPath)
		argIdx++
	}

	if filter.Search != nil && strings.TrimSpace(*filter.Search) != "" {
		searchTerm := "%" + strings.TrimSpace(*filter.Search) + "%"
		conditions = append(conditions, fmt.Sprintf("(f.title ILIKE $%d OR f.description ILIKE $%d OR v.original_filename ILIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, searchTerm)
		argIdx++
	}

	joinEntity := ""
	if filter.EntityType != nil && *filter.EntityType != "" && filter.EntityID != nil {
		joinEntity = fmt.Sprintf("JOIN documents.entity_attachments ea ON ea.file_id = f.id AND ea.entity_type = $%d AND ea.entity_id = $%d", argIdx, argIdx+1)
		args = append(args, *filter.EntityType, *filter.EntityID)
		argIdx += 2
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total matching
	countQuery := fmt.Sprintf(`
		SELECT COUNT(DISTINCT f.id)
		FROM documents.files f
		LEFT JOIN documents.file_versions v ON f.current_version_id = v.id
		%s
		WHERE %s
	`, joinEntity, whereClause)

	var totalCount int
	if err := r.db.Pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count files: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	selectQuery := fmt.Sprintf(`
		SELECT 
			f.id, f.legal_entity_id, f.title, f.description, f.category, f.folder_path,
			f.current_version_id, f.is_archived, f.created_by_user_id, f.created_at, f.updated_at,
			v.id, v.version_number, v.storage_provider, v.storage_key, v.original_filename,
			v.mime_type, v.byte_size, v.sha256_hash, v.metadata, v.uploaded_by, v.created_at
		FROM documents.files f
		LEFT JOIN documents.file_versions v ON f.current_version_id = v.id
		%s
		WHERE %s
		ORDER BY f.updated_at DESC
		LIMIT $%d OFFSET $%d
	`, joinEntity, whereClause, argIdx, argIdx+1)

	args = append(args, limit, offset)

	rows, err := r.db.Pool.Query(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var files []domain.File
	for rows.Next() {
		var file domain.File
		var cat string
		var desc *string
		var currVerID *uuid.UUID
		var vID *uuid.UUID
		var vNum *int
		var vProvider, vKey, vFilename, vMime, vHash *string
		var vSize *int64
		var vMeta []byte
		var vUploader *uuid.UUID
		var vCreated *time.Time

		err := rows.Scan(
			&file.ID,
			&file.LegalEntityID,
			&file.Title,
			&desc,
			&cat,
			&file.FolderPath,
			&currVerID,
			&file.IsArchived,
			&file.CreatedByUserID,
			&file.CreatedAt,
			&file.UpdatedAt,
			&vID,
			&vNum,
			&vProvider,
			&vKey,
			&vFilename,
			&vMime,
			&vSize,
			&vHash,
			&vMeta,
			&vUploader,
			&vCreated,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan file row: %w", err)
		}

		if desc != nil {
			file.Description = *desc
		}
		file.Category = domain.FileCategory(cat)
		file.CurrentVersionID = currVerID

		if vID != nil && *vID != uuid.Nil {
			ver := domain.FileVersion{
				ID:               *vID,
				FileID:           file.ID,
				VersionNumber:    *vNum,
				StorageProvider:  *vProvider,
				StorageKey:       *vKey,
				OriginalFilename: *vFilename,
				MimeType:         *vMime,
				ByteSize:         *vSize,
				SHA256Hash:       *vHash,
				UploadedBy:       vUploader,
				CreatedAt:        *vCreated,
			}
			if len(vMeta) > 0 {
				_ = json.Unmarshal(vMeta, &ver.Metadata)
			}
			file.CurrentVersion = &ver
		}

		files = append(files, file)
	}

	return files, totalCount, nil
}

// UpdateFile updates mutable fields of a file.
func (r *DocumentRepository) UpdateFile(ctx context.Context, file *domain.File) error {
	query := `
		UPDATE documents.files
		SET title = $1, description = $2, category = $3, folder_path = $4, updated_at = $5
		WHERE legal_entity_id = $6 AND id = $7
	`
	file.UpdatedAt = time.Now().UTC()
	tag, err := r.db.Pool.Exec(ctx, query,
		file.Title,
		file.Description,
		string(file.Category),
		file.FolderPath,
		file.UpdatedAt,
		file.LegalEntityID,
		file.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update document file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFileNotFound
	}
	return nil
}

// AddVersion records a new version for a file and updates the file's current version pointer.
func (r *DocumentRepository) AddVersion(ctx context.Context, version *domain.FileVersion) (*domain.FileVersion, error) {
	if version.ID == uuid.Nil {
		version.ID = uuid.New()
	}
	now := time.Now().UTC()
	if version.CreatedAt.IsZero() {
		version.CreatedAt = now
	}

	tx, err := r.db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Calculate next version number
	var nextVer int
	nextVerQuery := `SELECT COALESCE(MAX(version_number), 0) + 1 FROM documents.file_versions WHERE file_id = $1`
	if err := tx.QueryRow(ctx, nextVerQuery, version.FileID).Scan(&nextVer); err != nil {
		return nil, fmt.Errorf("failed to compute next version number: %w", err)
	}
	version.VersionNumber = nextVer

	metaJSON, err := json.Marshal(version.Metadata)
	if err != nil {
		metaJSON = []byte("{}")
	}

	insertQuery := `
		INSERT INTO documents.file_versions (
			id, file_id, version_number, storage_provider, storage_key,
			original_filename, mime_type, byte_size, sha256_hash, metadata,
			uploaded_by, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err = tx.Exec(ctx, insertQuery,
		version.ID,
		version.FileID,
		version.VersionNumber,
		version.StorageProvider,
		version.StorageKey,
		version.OriginalFilename,
		version.MimeType,
		version.ByteSize,
		version.SHA256Hash,
		metaJSON,
		version.UploadedBy,
		version.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert new version: %w", err)
	}

	// Update file's current_version_id and updated_at
	updateFileQuery := `UPDATE documents.files SET current_version_id = $1, updated_at = $2 WHERE id = $3`
	_, err = tx.Exec(ctx, updateFileQuery, version.ID, now, version.FileID)
	if err != nil {
		return nil, fmt.Errorf("failed to update current_version_id: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit version transaction: %w", err)
	}

	return version, nil
}

// GetVersionByID retrieves a specific version of a file.
func (r *DocumentRepository) GetVersionByID(ctx context.Context, fileID, versionID uuid.UUID) (*domain.FileVersion, error) {
	query := `
		SELECT 
			id, file_id, version_number, storage_provider, storage_key,
			original_filename, mime_type, byte_size, sha256_hash, metadata,
			uploaded_by, created_at
		FROM documents.file_versions
		WHERE file_id = $1 AND id = $2
	`

	var ver domain.FileVersion
	var meta []byte
	err := r.db.Pool.QueryRow(ctx, query, fileID, versionID).Scan(
		&ver.ID,
		&ver.FileID,
		&ver.VersionNumber,
		&ver.StorageProvider,
		&ver.StorageKey,
		&ver.OriginalFilename,
		&ver.MimeType,
		&ver.ByteSize,
		&ver.SHA256Hash,
		&meta,
		&ver.UploadedBy,
		&ver.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVersionNotFound
		}
		return nil, fmt.Errorf("failed to get file version: %w", err)
	}

	if len(meta) > 0 {
		_ = json.Unmarshal(meta, &ver.Metadata)
	}
	return &ver, nil
}

// ListVersions returns all versions of a file ordered newest first.
func (r *DocumentRepository) ListVersions(ctx context.Context, fileID uuid.UUID) ([]domain.FileVersion, error) {
	query := `
		SELECT 
			id, file_id, version_number, storage_provider, storage_key,
			original_filename, mime_type, byte_size, sha256_hash, metadata,
			uploaded_by, created_at
		FROM documents.file_versions
		WHERE file_id = $1
		ORDER BY version_number DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query file versions: %w", err)
	}
	defer rows.Close()

	var versions []domain.FileVersion
	for rows.Next() {
		var ver domain.FileVersion
		var meta []byte
		err := rows.Scan(
			&ver.ID,
			&ver.FileID,
			&ver.VersionNumber,
			&ver.StorageProvider,
			&ver.StorageKey,
			&ver.OriginalFilename,
			&ver.MimeType,
			&ver.ByteSize,
			&ver.SHA256Hash,
			&meta,
			&ver.UploadedBy,
			&ver.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan version: %w", err)
		}
		if len(meta) > 0 {
			_ = json.Unmarshal(meta, &ver.Metadata)
		}
		versions = append(versions, ver)
	}

	return versions, nil
}

// ArchiveFile marks a file as archived (soft delete).
func (r *DocumentRepository) ArchiveFile(ctx context.Context, legalEntityID, fileID uuid.UUID) error {
	query := `
		UPDATE documents.files
		SET is_archived = true, updated_at = CURRENT_TIMESTAMP
		WHERE legal_entity_id = $1 AND id = $2
	`
	tag, err := r.db.Pool.Exec(ctx, query, legalEntityID, fileID)
	if err != nil {
		return fmt.Errorf("failed to archive document file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFileNotFound
	}
	return nil
}

// DeleteFile permanently removes a document file and its cascade associations.
func (r *DocumentRepository) DeleteFile(ctx context.Context, legalEntityID, fileID uuid.UUID) error {
	query := `DELETE FROM documents.files WHERE legal_entity_id = $1 AND id = $2`
	tag, err := r.db.Pool.Exec(ctx, query, legalEntityID, fileID)
	if err != nil {
		return fmt.Errorf("failed to delete document file: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrFileNotFound
	}
	return nil
}

// CreateAttachment attaches a file to an ERP entity.
func (r *DocumentRepository) CreateAttachment(ctx context.Context, attachment *domain.EntityAttachment) error {
	if attachment.ID == uuid.Nil {
		attachment.ID = uuid.New()
	}
	if attachment.CreatedAt.IsZero() {
		attachment.CreatedAt = time.Now().UTC()
	}

	query := `
		INSERT INTO documents.entity_attachments (id, file_id, entity_type, entity_id, attached_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (file_id, entity_type, entity_id) DO NOTHING
	`
	_, err := r.db.Pool.Exec(ctx, query,
		attachment.ID,
		attachment.FileID,
		attachment.EntityType,
		attachment.EntityID,
		attachment.AttachedBy,
		attachment.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to attach document to entity: %w", err)
	}
	return nil
}

// DeleteAttachment unlinks a file from an entity.
func (r *DocumentRepository) DeleteAttachment(ctx context.Context, fileID uuid.UUID, entityType string, entityID uuid.UUID) error {
	query := `DELETE FROM documents.entity_attachments WHERE file_id = $1 AND entity_type = $2 AND entity_id = $3`
	tag, err := r.db.Pool.Exec(ctx, query, fileID, entityType, entityID)
	if err != nil {
		return fmt.Errorf("failed to delete entity attachment: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrAttachmentNotFound
	}
	return nil
}

// ListAttachmentsByEntity returns all attachments for a specific entity.
func (r *DocumentRepository) ListAttachmentsByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]domain.EntityAttachment, error) {
	query := `
		SELECT 
			ea.id, ea.file_id, ea.entity_type, ea.entity_id, ea.attached_by, ea.created_at,
			f.id, f.legal_entity_id, f.title, f.description, f.category, f.folder_path,
			f.current_version_id, f.is_archived, f.created_by_user_id, f.created_at, f.updated_at,
			v.id, v.version_number, v.storage_provider, v.storage_key, v.original_filename,
			v.mime_type, v.byte_size, v.sha256_hash, v.metadata, v.uploaded_by, v.created_at
		FROM documents.entity_attachments ea
		JOIN documents.files f ON ea.file_id = f.id
		LEFT JOIN documents.file_versions v ON f.current_version_id = v.id
		WHERE ea.entity_type = $1 AND ea.entity_id = $2
		ORDER BY ea.created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, entityType, entityID)
	if err != nil {
		return nil, fmt.Errorf("failed to query entity attachments: %w", err)
	}
	defer rows.Close()

	var attachments []domain.EntityAttachment
	for rows.Next() {
		var att domain.EntityAttachment
		var file domain.File
		var cat string
		var desc *string
		var currVerID *uuid.UUID
		var vID *uuid.UUID
		var vNum *int
		var vProvider, vKey, vFilename, vMime, vHash *string
		var vSize *int64
		var vMeta []byte
		var vUploader *uuid.UUID
		var vCreated *time.Time

		err := rows.Scan(
			&att.ID,
			&att.FileID,
			&att.EntityType,
			&att.EntityID,
			&att.AttachedBy,
			&att.CreatedAt,
			&file.ID,
			&file.LegalEntityID,
			&file.Title,
			&desc,
			&cat,
			&file.FolderPath,
			&currVerID,
			&file.IsArchived,
			&file.CreatedByUserID,
			&file.CreatedAt,
			&file.UpdatedAt,
			&vID,
			&vNum,
			&vProvider,
			&vKey,
			&vFilename,
			&vMime,
			&vSize,
			&vHash,
			&vMeta,
			&vUploader,
			&vCreated,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment row: %w", err)
		}

		if desc != nil {
			file.Description = *desc
		}
		file.Category = domain.FileCategory(cat)
		file.CurrentVersionID = currVerID

		if vID != nil && *vID != uuid.Nil {
			ver := domain.FileVersion{
				ID:               *vID,
				FileID:           file.ID,
				VersionNumber:    *vNum,
				StorageProvider:  *vProvider,
				StorageKey:       *vKey,
				OriginalFilename: *vFilename,
				MimeType:         *vMime,
				ByteSize:         *vSize,
				SHA256Hash:       *vHash,
				UploadedBy:       vUploader,
				CreatedAt:        *vCreated,
			}
			if len(vMeta) > 0 {
				_ = json.Unmarshal(vMeta, &ver.Metadata)
			}
			file.CurrentVersion = &ver
		}

		att.File = &file
		attachments = append(attachments, att)
	}

	return attachments, nil
}

// ListAttachmentsByFile returns all entity links for a given file.
func (r *DocumentRepository) ListAttachmentsByFile(ctx context.Context, fileID uuid.UUID) ([]domain.EntityAttachment, error) {
	query := `
		SELECT id, file_id, entity_type, entity_id, attached_by, created_at
		FROM documents.entity_attachments
		WHERE file_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to query attachments by file: %w", err)
	}
	defer rows.Close()

	var attachments []domain.EntityAttachment
	for rows.Next() {
		var att domain.EntityAttachment
		err := rows.Scan(
			&att.ID,
			&att.FileID,
			&att.EntityType,
			&att.EntityID,
			&att.AttachedBy,
			&att.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, att)
	}
	return attachments, nil
}

// GetSummary calculates high-level document analytics for a legal entity.
func (r *DocumentRepository) GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.DocumentSummary, error) {
	query := `
		SELECT 
			COUNT(DISTINCT f.id) as total_files,
			COUNT(v.id) as total_versions,
			COALESCE(SUM(v.byte_size), 0) as total_size_bytes
		FROM documents.files f
		LEFT JOIN documents.file_versions v ON f.id = v.file_id
		WHERE f.legal_entity_id = $1 AND f.is_archived = false
	`

	summary := &domain.DocumentSummary{
		CategoriesCount: make(map[string]int64),
	}

	err := r.db.Pool.QueryRow(ctx, query, legalEntityID).Scan(
		&summary.TotalFiles,
		&summary.TotalVersions,
		&summary.TotalSizeBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate document summary: %w", err)
	}

	catQuery := `
		SELECT category, COUNT(id)
		FROM documents.files
		WHERE legal_entity_id = $1 AND is_archived = false
		GROUP BY category
	`
	rows, err := r.db.Pool.Query(ctx, catQuery, legalEntityID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate category summary: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var cat string
		var count int64
		if err := rows.Scan(&cat, &count); err == nil {
			summary.CategoriesCount[cat] = count
		}
	}

	return summary, nil
}
