package storage_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"erp/internal/platform/storage"
)

func TestLocalFilesystemBlobStore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "erp_storage_test_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	store, err := storage.NewLocalFilesystemBlobStore(tempDir)
	if err != nil {
		t.Fatalf("failed to initialize blob store: %v", err)
	}
	if err := store.Check(); err != nil {
		t.Fatalf("expected writable store readiness: %v", err)
	}

	ctx := context.Background()
	content := []byte("Hello Saudi ERP Secure Storage Engine - 2026!")

	// 1. Put Object
	obj := storage.Object{
		Key:         "documents/2026/test_file.txt",
		ContentType: "text/plain",
		Data:        bytes.NewReader(content),
		Size:        int64(len(content)),
	}

	stored, err := store.Put(ctx, obj)
	if err != nil {
		t.Fatalf("failed to put object: %v", err)
	}

	if stored.Size != int64(len(content)) {
		t.Errorf("expected size %d, got %d", len(content), stored.Size)
	}
	if stored.SHA256 == "" {
		t.Errorf("expected non-empty SHA-256 hash")
	}
	if stored.Provider != "local" {
		t.Errorf("expected provider 'local', got '%s'", stored.Provider)
	}

	// 2. Open & Read Object
	reader, err := store.Open(ctx, "documents/2026/test_file.txt")
	if err != nil {
		t.Fatalf("failed to open stored object: %v", err)
	}

	readData, err := io.ReadAll(reader)
	_ = reader.Close()
	if err != nil {
		t.Fatalf("failed to read data: %v", err)
	}
	if !bytes.Equal(readData, content) {
		t.Errorf("read content does not match original: %s vs %s", string(readData), string(content))
	}

	// 3. Path Traversal Protection
	maliciousKeys := []string{
		"../escape.txt",
		"documents/../../etc/passwd",
		"..\\windows\\system32\\cmd.exe",
	}
	for _, key := range maliciousKeys {
		_, err := store.Put(ctx, storage.Object{
			Key:  key,
			Data: bytes.NewReader([]byte("danger")),
		})
		if err == nil {
			t.Errorf("expected error for path traversal key %q, got nil", key)
		}
	}

	// 4. Delete Object
	if err := store.Delete(ctx, "documents/2026/test_file.txt"); err != nil {
		t.Fatalf("failed to delete object: %v", err)
	}

	// 5. Open Deleted Object returns ErrNotFound
	_, err = store.Open(ctx, "documents/2026/test_file.txt")
	if err != storage.ErrNotFound {
		t.Errorf("expected ErrNotFound after deletion, got: %v", err)
	}

	// 6. Delete Non-existent Object is idempotent
	if err := store.Delete(ctx, "non_existent.txt"); err != nil {
		t.Errorf("expected nil deleting non-existent object, got: %v", err)
	}
}
