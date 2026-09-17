package ports

import (
	"context"
	"io"

	"erp/internal/documents/domain"
	"github.com/google/uuid"
)

// Repository defines data access operations for documents, versions, and attachments.
type Repository interface {
	CreateFile(ctx context.Context, file *domain.File, version *domain.FileVersion) (*domain.File, error)
	GetFileByID(ctx context.Context, legalEntityID, fileID uuid.UUID) (*domain.File, error)
	ListFiles(ctx context.Context, filter domain.FileFilter) ([]domain.File, int, error)
	UpdateFile(ctx context.Context, file *domain.File) error
	AddVersion(ctx context.Context, version *domain.FileVersion) (*domain.FileVersion, error)
	GetVersionByID(ctx context.Context, fileID, versionID uuid.UUID) (*domain.FileVersion, error)
	ListVersions(ctx context.Context, fileID uuid.UUID) ([]domain.FileVersion, error)
	ArchiveFile(ctx context.Context, legalEntityID, fileID uuid.UUID) error
	DeleteFile(ctx context.Context, legalEntityID, fileID uuid.UUID) error
	CreateAttachment(ctx context.Context, attachment *domain.EntityAttachment) error
	DeleteAttachment(ctx context.Context, fileID uuid.UUID, entityType string, entityID uuid.UUID) error
	ListAttachmentsByEntity(ctx context.Context, entityType string, entityID uuid.UUID) ([]domain.EntityAttachment, error)
	ListAttachmentsByFile(ctx context.Context, fileID uuid.UUID) ([]domain.EntityAttachment, error)
	GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.DocumentSummary, error)
}

// Service defines high-level document orchestration workflows.
type Service interface {
	UploadFile(ctx context.Context, params domain.UploadFileParams) (*domain.File, error)
	UploadVersion(ctx context.Context, params domain.UploadVersionParams) (*domain.FileVersion, error)
	GetFile(ctx context.Context, legalEntityID, fileID uuid.UUID) (*domain.File, error)
	ListFiles(ctx context.Context, filter domain.FileFilter) ([]domain.File, int, error)
	DownloadFile(ctx context.Context, legalEntityID, fileID uuid.UUID, versionID *uuid.UUID) (io.ReadCloser, *domain.FileVersion, error)
	ArchiveFile(ctx context.Context, legalEntityID, fileID uuid.UUID) error
	AttachToEntity(ctx context.Context, fileID uuid.UUID, entityType string, entityID uuid.UUID, userID *uuid.UUID) error
	DetachFromEntity(ctx context.Context, fileID uuid.UUID, entityType string, entityID uuid.UUID) error
	GetEntityAttachments(ctx context.Context, entityType string, entityID uuid.UUID) ([]domain.EntityAttachment, error)
	GetSummary(ctx context.Context, legalEntityID uuid.UUID) (*domain.DocumentSummary, error)
}
