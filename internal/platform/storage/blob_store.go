package storage

import (
	"context"
	"errors"
	"io"
	"time"
)

var (
	// ErrNotFound is returned when a storage key does not exist.
	ErrNotFound = errors.New("storage object not found")
	// ErrInvalidKey is returned when a key is malformed or attempts path traversal.
	ErrInvalidKey = errors.New("invalid storage key")
	// ErrStorageWriteFailed is returned when an object cannot be written.
	ErrStorageWriteFailed = errors.New("failed to write storage object")
	// ErrStorageReadFailed is returned when an object cannot be opened or read.
	ErrStorageReadFailed = errors.New("failed to read storage object")
)

// Object represents an object to be stored in the BlobStore.
type Object struct {
	Key         string
	ContentType string
	Data        io.Reader
	Size        int64
	Metadata    map[string]string
}

// StoredObject represents the metadata of an object stored in the BlobStore.
type StoredObject struct {
	Key         string
	Provider    string
	Size        int64
	ContentType string
	ETag        string
	SHA256      string
	CreatedAt   time.Time
}

// BlobStore represents the canonical interface for persistent blob storage (Section 33).
type BlobStore interface {
	Put(
		ctx context.Context,
		object Object,
	) (StoredObject, error)

	Open(
		ctx context.Context,
		key string,
	) (io.ReadCloser, error)

	Delete(
		ctx context.Context,
		key string,
	) error
}
