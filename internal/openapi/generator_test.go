package openapi

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestGeneratedContractCoversKnownRoutes(t *testing.T) {
	repositoryRoot := filepath.Join("..", "..")
	contract, err := Generate(repositoryRoot)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	for _, expected := range [][]byte{
		[]byte("/api/v1/auth/login/mfa:"),
		[]byte("/api/v1/accounting/journal-entries/{id}/post:"),
		[]byte("/api/v1/fixedassets/assets:"),
		[]byte("/api/v1/workflow/instances:"),
	} {
		if !bytes.Contains(contract, expected) {
			t.Fatalf("generated contract does not contain %q", expected)
		}
	}

	operations, err := Discover(repositoryRoot)
	if err != nil {
		t.Fatalf("Discover() error = %v", err)
	}
	if len(operations) < 300 {
		t.Fatalf("discovered only %d operations, expected at least 300", len(operations))
	}
}
