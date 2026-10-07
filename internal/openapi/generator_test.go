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
		[]byte("/api/v1/auth/users:"),
		[]byte("/api/v1/auth/users/{id}/password:"),
		[]byte("/api/v1/authz/assignments:"),
		[]byte("/api/v1/authz/assignments/revoke:"),
		[]byte("$ref: '#/components/schemas/iam_CreateUserRequest'"),
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

func TestHandlerSchemasAcrossFilesKeepReceiverScope(t *testing.T) {
	root := t.TempDir()
	moduleDir := filepath.Join(root, "internal", "example")
	routeDir := filepath.Join(moduleDir, "transport", "http")
	if err := os.MkdirAll(routeDir, 0755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		filepath.Join(moduleDir, "module.go"): `package example
func Mount(r Router) { r.Mount("/api/v1/example", handler) }`,
		filepath.Join(routeDir, "routes.go"): `package http
func (h *UserHandler) Routes(r Router) { r.Post("/users", h.Create); r.Post("/upload", h.Upload) }
func (h *RoleHandler) Routes(r Router) { r.Post("/roles", h.Create) }`,
		filepath.Join(routeDir, "users.go"): `package http
type CreateUserRequest struct { Username string ` + "`json:\"username\"`" + ` }
func (h *UserHandler) Create(w Writer, r Request) { var body CreateUserRequest; decoder.Decode(&body) }
func (h *UserHandler) Upload(w Writer, r Request) { r.FormFile("file") }`,
		filepath.Join(routeDir, "roles.go"): `package http
type CreateRoleRequest struct { Code string ` + "`json:\"code\"`" + ` }
func (h *RoleHandler) Create(w Writer, r Request) { var body CreateRoleRequest; decoder.Decode(&body) }`,
		filepath.Join(routeDir, "ignored_test.go"): `package http
func (h *UserHandler) Create(w Writer, r Request) { var body IncorrectTestRequest; decoder.Decode(&body) }`,
	}
	for filename, contents := range files {
		if err := os.WriteFile(filename, []byte(contents), 0600); err != nil {
			t.Fatal(err)
		}
	}
	operations, err := Discover(root)
	if err != nil {
		t.Fatal(err)
	}
	byPath := make(map[string]Operation)
	for _, operation := range operations {
		byPath[operation.Path] = operation
	}
	if actual := byPath["/api/v1/example/users"].RequestType; actual != "CreateUserRequest" {
		t.Errorf("user request type = %q, want CreateUserRequest", actual)
	}
	if actual := byPath["/api/v1/example/roles"].RequestType; actual != "CreateRoleRequest" {
		t.Errorf("role request type = %q, want CreateRoleRequest", actual)
	}
	if actual := byPath["/api/v1/example/upload"].BodyKind; actual != "multipart/form-data" {
		t.Errorf("upload body kind = %q, want multipart/form-data", actual)
	}
	contract, err := Generate(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"example_CreateUserRequest", "example_CreateRoleRequest", "MultipartUploadRequest"} {
		if !bytes.Contains(contract, []byte("$ref: '#/components/schemas/"+name+"'")) {
			t.Errorf("missing request schema reference for %s", name)
		}
	}
}
