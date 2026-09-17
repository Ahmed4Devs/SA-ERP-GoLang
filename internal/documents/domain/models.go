package domain

import (
	"io"
	"time"

	"github.com/google/uuid"
)

// FileCategory defines the categorization of documents.
type FileCategory string

const (
	CategoryGeneral    FileCategory = "GENERAL"
	CategoryContract   FileCategory = "CONTRACT"
	CategoryInvoice    FileCategory = "INVOICE"
	CategoryReceipt    FileCategory = "RECEIPT"
	CategoryLegal      FileCategory = "LEGAL"
	CategoryIdentity   FileCategory = "IDENTITY"
	CategoryTax        FileCategory = "TAX"
	CategoryAttachment FileCategory = "ATTACHMENT"
	CategoryOther      FileCategory = "OTHER"
)

// File represents a document file in the ERP system.
type File struct {
	ID               uuid.UUID          `json:"id"`
	LegalEntityID    uuid.UUID          `json:"legal_entity_id"`
	Title            string             `json:"title"`
	Description      string             `json:"description,omitempty"`
	Category         FileCategory       `json:"category"`
	FolderPath       string             `json:"folder_path"`
	CurrentVersionID *uuid.UUID         `json:"current_version_id,omitempty"`
	CurrentVersion   *FileVersion       `json:"current_version,omitempty"`
	IsArchived       bool               `json:"is_archived"`
	CreatedByUserID  *uuid.UUID         `json:"created_by_user_id,omitempty"`
	CreatedAt        time.Time          `json:"created_at"`
	UpdatedAt        time.Time          `json:"updated_at"`
	Attachments      []EntityAttachment `json:"attachments,omitempty"`
}

// FileVersion represents a specific binary version of a document.
type FileVersion struct {
	ID               uuid.UUID              `json:"id"`
	FileID           uuid.UUID              `json:"file_id"`
	VersionNumber    int                    `json:"version_number"`
	StorageProvider  string                 `json:"storage_provider"`
	StorageKey       string                 `json:"storage_key"`
	OriginalFilename string                 `json:"original_filename"`
	MimeType         string                 `json:"mime_type"`
	ByteSize         int64                  `json:"byte_size"`
	SHA256Hash       string                 `json:"sha256_hash"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
	UploadedBy       *uuid.UUID             `json:"uploaded_by,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
}

// EntityAttachment represents a polymorphic link between a document and an ERP entity.
type EntityAttachment struct {
	ID         uuid.UUID  `json:"id"`
	FileID     uuid.UUID  `json:"file_id"`
	EntityType string     `json:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id"`
	AttachedBy *uuid.UUID `json:"attached_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	File       *File      `json:"file,omitempty"`
}

// UploadFileParams contains parameters for uploading a new file.
type UploadFileParams struct {
	LegalEntityID    uuid.UUID
	Title            string
	Description      string
	Category         FileCategory
	FolderPath       string
	OriginalFilename string
	ContentType      string
	Content          io.Reader
	UserID           *uuid.UUID
	EntityType       string
	EntityID         *uuid.UUID
	Metadata         map[string]interface{}
}

// UploadVersionParams contains parameters for uploading a new version of an existing file.
type UploadVersionParams struct {
	FileID           uuid.UUID
	LegalEntityID    uuid.UUID
	OriginalFilename string
	ContentType      string
	Content          io.Reader
	UserID           *uuid.UUID
	Metadata         map[string]interface{}
}

// FileFilter contains query filters for listing documents.
type FileFilter struct {
	LegalEntityID   uuid.UUID
	Category        *FileCategory
	FolderPath      *string
	Search          *string
	EntityType      *string
	EntityID        *uuid.UUID
	IncludeArchived bool
	Limit           int
	Offset          int
}

// DocumentSummary contains analytical totals for the document management platform.
type DocumentSummary struct {
	TotalFiles      int64            `json:"total_files"`
	TotalVersions   int64            `json:"total_versions"`
	TotalSizeBytes  int64            `json:"total_size_bytes"`
	CategoriesCount map[string]int64 `json:"categories_count"`
}
