package openapi

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"go.yaml.in/yaml/v3"
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
		[]byte("$ref: '#/components/schemas/accounting_CreateAccountRequest'"),
		[]byte("$ref: '#/components/parameters/LegalEntityQuery'"),
		[]byte("$ref: '#/components/responses/Forbidden'"),
		[]byte("format: uuid"),
		[]byte("multipart/form-data:"),
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
	if bytes.Contains(contract, []byte("        default:\n")) {
		t.Fatal("generated contract still contains response skeletons")
	}
	var parsed map[string]any
	if err := yaml.Unmarshal(contract, &parsed); err != nil {
		t.Fatalf("generated contract is not valid YAML: %v", err)
	}

	refPattern := regexp.MustCompile(`\$ref: '#/components/(schemas|responses|parameters)/([^']+)'`)
	for _, match := range refPattern.FindAllSubmatch(contract, -1) {
		if !bytes.Contains(contract, []byte("    "+string(match[2])+":\n")) {
			t.Errorf("generated contract contains unresolved component reference %q", match[0])
		}
	}

	onDisk, err := os.ReadFile(filepath.Join(repositoryRoot, "api", "openapi", "openapi.yaml"))
	if err != nil {
		t.Fatalf("read generated OpenAPI file: %v", err)
	}
	if !bytes.Equal(contract, onDisk) {
		t.Fatal("api/openapi/openapi.yaml is stale; run go run ./cmd/openapi-gen")
	}
}
