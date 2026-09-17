package application

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"erp/internal/documents/domain"
	"erp/internal/documents/ports"
	"erp/internal/platform/storage"
	"github.com/google/uuid"
)

// DocumentService coordinates document management and persistent blob storage.
type DocumentService struct {
	repo      ports.Repository
	blobStore storage.BlobStore
}

var _ ports.Service = (*DocumentService)(nil)

// NewDocumentService initializes a new DocumentService.
func NewDocumentService(repo ports.Repository, blobStore storage.BlobStore) *DocumentService {
	return &DocumentService{
		repo:      repo,
		blobStore: blobStore,
	}
}

// UploadFile uploads a new file, persists its binary content into the BlobStore, and creates records.
func (s *DocumentService) UploadFile(ctx context.Context, params domain.UploadFileParams) (*domain.File, error) {
	if params.LegalEntityID == uuid.Nil {
		return nil, fmt.Errorf("legal_entity_id is required")
	}
	if strings.TrimSpace(params.Title) == "" {
		return nil, domain.ErrEmptyTitle
	}
	if strings.TrimSpace(params.OriginalFilename) == "" {
		return nil, domain.ErrEmptyFilename
	}
	if params.Content == nil {
		return nil, fmt.Errorf("file content stream is nil")
	}

	category := params.Category
	if category == "" {
		category = domain.CategoryGeneral
	}

	folder := params.FolderPath
	if folder == "" {
		folder = "/"
	}

	// Format safe unique storage key: tenants/<le_id>/<category>/<short_uuid>_<safe_filename>
	safeFilename := filepath.Base(params.OriginalFilename)
	safeFilename = strings.ReplaceAll(safeFilename, " ", "_")
	storageKey := fmt.Sprintf("tenants/%s/%s/%s_%s",
		params.LegalEntityID.String(),
		strings.ToLower(string(category)),
		uuid.New().String()[:8],
		safeFilename,
	)

	// Stream file to blob store
	storedObj, err := s.blobStore.Put(ctx, storage.Object{
		Key:         storageKey,
		ContentType: params.ContentType,
		Data:        params.Content,
		Metadata: map[string]string{
			"original_filename": params.OriginalFilename,
			"legal_entity_id":   params.LegalEntityID.String(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrStorageFailed, err)
	}

	// Prepare domain entities
	fileID := uuid.New()
	versionID := uuid.New()

	file := &domain.File{
		ID:              fileID,
		LegalEntityID:   params.LegalEntityID,
		Title:           params.Title,
		Description:     params.Description,
		Category:        category,
		FolderPath:      folder,
		IsArchived:      false,
		CreatedByUserID: params.UserID,
	}

	version := &domain.FileVersion{
		ID:               versionID,
		FileID:           fileID,
		VersionNumber:    1,
		StorageProvider:  storedObj.Provider,
		StorageKey:       storedObj.Key,
		OriginalFilename: params.OriginalFilename,
		MimeType:         storedObj.ContentType,
		ByteSize:         storedObj.Size,
		SHA256Hash:       storedObj.SHA256,
		Metadata:         params.Metadata,
		UploadedBy:       params.UserID,
	}

	createdFile, err := s.repo.CreateFile(ctx, file, version)
	if err != nil {
		// Clean up stored file if DB write failed
		_ = s.blobStore.Delete(ctx, storageKey)
		return nil, err
	}

	// If entity link was requested upon upload, attach immediately
	if params.EntityType != "" && params.EntityID != nil && *params.EntityID != uuid.Nil {
		_ = s.repo.CreateAttachment(ctx, &domain.EntityAttachment{
			ID:         uuid.New(),
			FileID:     createdFile.ID,
			EntityType: params.EntityType,
			EntityID:   *params.EntityID,
			AttachedBy: params.UserID,
		})
	}

	return createdFile, nil
}

// UploadVersion adds a new version to an existing document.
func (s *DocumentService) UploadVersion(ctx context.Context, params domain.UploadVersionParams) (*domain.FileVersion, error) {
	if params.FileID == uuid.Nil {
		return nil, fmt.Errorf("file_id is required")
	}
	if strings.TrimSpace(params.OriginalFilename) == "" {
		return nil, domain.ErrEmptyFilename
	}
	if params.Content == nil {
		return nil, fmt.Errorf("file content stream is nil")
	}

	file, err := s.repo.GetFileByID(ctx, params.LegalEntityID, params.FileID)
	if err != nil {
		return nil, err
	}
	if file.IsArchived {
		return nil, domain.ErrFileArchived
	}

	safeFilename := filepath.Base(params.OriginalFilename)
	safeFilename = strings.ReplaceAll(safeFilename, " ", "_")
	storageKey := fmt.Sprintf("tenants/%s/%s/v_%s_%s",
		params.LegalEntityID.String(),
		strings.ToLower(string(file.Category)),
		uuid.New().String()[:8],
		safeFilename,
	)

	storedObj, err := s.blobStore.Put(ctx, storage.Object{
		Key:         storageKey,
		ContentType: params.ContentType,
		Data:        params.Content,
		Metadata: map[string]string{
			"file_id":           params.FileID.String(),
			"original_filename": params.OriginalFilename,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrStorageFailed, err)
	}

	version := &domain.FileVersion{
		ID:               uuid.New(),
		FileID:           params.FileID,
		StorageProvider:  storedObj.Provider,
		StorageKey:       storedObj.Key,
		OriginalFilename: params.OriginalFilename,
		MimeType:         storedObj.ContentType,
		ByteSize:         storedObj.Size,
		SHA256Hash:       storedObj.SHA256,
		Metadata:         params.Metadata,
		UploadedBy:       params.UserID,
	}

	createdVer, err := s.repo.AddVersion(ctx, version)
	if err != nil {
		_ = s.blobStore.Delete(ctx, storageKey)
		return nil, err
	}

	return createdVer, nil
}

// GetFile retrieves a file and its current version.
func (s *DocumentService) GetFile(ctx context.Context, legalEntityID, fileID uuid.UUID) (*domain.File, error) {
	return s.repo.GetFileByID(ctx, legalEntityID, fileID)
}

// ListFiles lists files matching filters.
func (s *DocumentService) ListFiles(ctx context.Context, filter domain.FileFilter) ([]domain.File, int, error) {
	return s.repo.ListFiles(ctx, filter)
}

// DownloadFile opens a stream for downloading the file content.
func (s *DocumentService) DownloadFile(ctx context.Context, legalEntityID, fileID uuid.UUID, versionID *uuid.UUID) (io.ReadCloser, *domain.FileVersion, error) {
	file, err := s.repo.GetFileByID(ctx, legalEntityID, fileID)
	if err != nil {
		return nil, nil, err
	}

	var targetVer *domain.FileVersion
	if versionID != nil && *versionID != uuid.Nil {
		ver, err := s.repo.GetVersionByID(ctx, fileID, *versionID)
		if err != nil {
			return nil, nil, err
		}
		targetVer = ver
	} else {
		if file.CurrentVersion == nil {
			return nil, nil, domain.ErrVersionNotFound
		}
		targetVer = file.CurrentVersion
	}

	reader, err := s.blobStore.Open(ctx, targetVer.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %v", domain.ErrStorageFailed, err)
	}

	return reader, targetVer, nil
}

// ArchiveFile archives a document file.
func (s *DocumentService) ArchiveFile(ctx context.Context, legalEntityID, fileID uuid.UUID) error {
	return s.repo.ArchiveFile(ctx, legalEntityID, fileID)
}

// AttachToEntity links a file to an entity.
func (s *DocumentService) AttachToEntity(ctx context.Context, fileID uuid.UUID, entityType string, entityID uuid.UUID, userID *uuid.UUID) error {
	if strings.TrimSpace(entityType) == "" {
		return domain.ErrInvalidAttachmentType
	}
	if entityID == uuid.Nil {
		return fmt.Errorf("entity_id is required")
	}

	return s.repo.CreateAttachment(ctx, &domain.EntityAttachment{
		ID:         uuid.New(),
		FileID:     fileID,
		EntityType: entityType,
		EntityID:   entityID,
		AttachedBy: userID,
	})
}

// DetachFromEntity removes a link between a file and an entity.
func (s *DocumentService) DetachFromEntity(ctx context.Context, fileID uuid.UUID, entityType string, entityID uuid.UUID) error {
	return s.repo.DeleteAttachment(ctx, fileID, entityType, entityID)
}

// GetEntityAttachments returns all attachments for an entity.
func (s *DocumentService) GetEntityAttachments(ctx context.Context, entityType string, entityID uuid.UUID) ([]domain.EntityAttachment, error) {
	return s.repo.ListAttachmentsByEntity(ctx, entityType, entityID)
}

// GetSummary returns storage statistics for a legal entity.
func (s *DocumentService) GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.DocumentSummary, error) {
	return s.repo.GetSummary(ctx, legalEntityID)
}
