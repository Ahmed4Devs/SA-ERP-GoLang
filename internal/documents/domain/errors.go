package domain

import "errors"

var (
	ErrFileNotFound          = errors.New("document file not found")
	ErrVersionNotFound       = errors.New("document version not found")
	ErrAttachmentNotFound    = errors.New("entity attachment not found")
	ErrInvalidCategory       = errors.New("invalid document category")
	ErrEmptyFilename         = errors.New("filename cannot be empty")
	ErrEmptyTitle            = errors.New("document title cannot be empty")
	ErrStorageFailed         = errors.New("underlying blob storage operation failed")
	ErrFileArchived          = errors.New("cannot modify archived document")
	ErrDuplicateAttachment   = errors.New("document is already attached to this entity")
	ErrInvalidAttachmentType = errors.New("invalid entity attachment type")
)
