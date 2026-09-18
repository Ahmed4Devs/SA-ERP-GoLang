package storage

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// LocalFilesystemBlobStore implements BlobStore using the local filesystem.
type LocalFilesystemBlobStore struct {
	rootDir string
}

// NewLocalFilesystemBlobStore creates a new LocalFilesystemBlobStore rooted at the given directory.
func NewLocalFilesystemBlobStore(rootDir string) (*LocalFilesystemBlobStore, error) {
	absPath, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path for storage root: %w", err)
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage root directory %s: %w", absPath, err)
	}

	return &LocalFilesystemBlobStore{
		rootDir: absPath,
	}, nil
}

// RootDir returns the absolute root directory of the store.
func (s *LocalFilesystemBlobStore) RootDir() string {
	return s.rootDir
}

// Check verifies that the persistent storage root is writable. It is used by
// readiness probes so a pod is never sent document traffic with a read-only or
// disconnected volume.
func (s *LocalFilesystemBlobStore) Check() error {
	testFile, err := os.CreateTemp(s.rootDir, ".readiness-*")
	if err != nil {
		return fmt.Errorf("storage root is not writable: %w", err)
	}
	name := testFile.Name()
	if err := testFile.Close(); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("failed to close storage readiness file: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("failed to clean storage readiness file: %w", err)
	}
	return nil
}

// resolvePath checks for directory traversal and returns the absolute filesystem path for the given key.
func (s *LocalFilesystemBlobStore) resolvePath(key string) (string, error) {
	cleanKey := filepath.Clean(key)
	if cleanKey == "." || cleanKey == "/" || strings.HasPrefix(cleanKey, "..") || strings.Contains(cleanKey, "/../") || strings.Contains(cleanKey, `\..\`) {
		return "", fmt.Errorf("%w: path traversal detected: %s", ErrInvalidKey, key)
	}

	target := filepath.Join(s.rootDir, cleanKey)

	// Ensure the target is strictly inside rootDir
	rel, err := filepath.Rel(s.rootDir, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%w: key escapes root directory: %s", ErrInvalidKey, key)
	}

	return target, nil
}

// Put writes an object to the local filesystem atomically with SHA-256 calculation.
func (s *LocalFilesystemBlobStore) Put(ctx context.Context, object Object) (StoredObject, error) {
	if object.Key == "" {
		return StoredObject{}, fmt.Errorf("%w: empty key", ErrInvalidKey)
	}

	targetPath, err := s.resolvePath(object.Key)
	if err != nil {
		return StoredObject{}, err
	}

	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return StoredObject{}, fmt.Errorf("%w: failed to create parent dir: %v", ErrStorageWriteFailed, err)
	}

	// Create a random temporary file in the same directory for atomic rename
	randBytes := make([]byte, 8)
	_, _ = rand.Read(randBytes)
	tempPath := filepath.Join(targetDir, fmt.Sprintf(".tmp_%s_%s", filepath.Base(targetPath), hex.EncodeToString(randBytes)))

	tempFile, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return StoredObject{}, fmt.Errorf("%w: failed to create temp file: %v", ErrStorageWriteFailed, err)
	}

	hasher := sha256.New()
	multiWriter := io.MultiWriter(tempFile, hasher)

	written, err := io.Copy(multiWriter, object.Data)
	if closeErr := tempFile.Close(); closeErr != nil && err == nil {
		err = closeErr
	}

	if err != nil {
		_ = os.Remove(tempPath)
		return StoredObject{}, fmt.Errorf("%w: stream copy failed: %v", ErrStorageWriteFailed, err)
	}

	// Atomic rename to final path
	if err := os.Rename(tempPath, targetPath); err != nil {
		_ = os.Remove(tempPath)
		return StoredObject{}, fmt.Errorf("%w: atomic rename failed: %v", ErrStorageWriteFailed, err)
	}

	sha256Hex := hex.EncodeToString(hasher.Sum(nil))

	contentType := object.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return StoredObject{
		Key:         object.Key,
		Provider:    "local",
		Size:        written,
		ContentType: contentType,
		ETag:        fmt.Sprintf("\"%s\"", sha256Hex),
		SHA256:      sha256Hex,
		CreatedAt:   time.Now().UTC(),
	}, nil
}

// Open opens an object from the local filesystem for reading.
func (s *LocalFilesystemBlobStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	targetPath, err := s.resolvePath(key)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrStorageReadFailed, err)
	}

	return file, nil
}

// Delete removes an object from the local filesystem. If the file does not exist, Delete returns nil.
func (s *LocalFilesystemBlobStore) Delete(ctx context.Context, key string) error {
	targetPath, err := s.resolvePath(key)
	if err != nil {
		return err
	}

	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
