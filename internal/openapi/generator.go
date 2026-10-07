package openapi

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

var httpMethods = map[string]string{
	"Get": "get", "Post": "post", "Put": "put", "Patch": "patch", "Delete": "delete",
}

type Operation struct {
	Method      string
	Path        string
	Module      string
	Handler     string
	RequestType string
	BodyKind    string
	handlerKey  string
}

type ObjectSchema struct {
	Name     string
	Module   string
	Fields   []SchemaField
	Required []string
}

type SchemaField struct {
	Name string
	Node SchemaNode
}

type SchemaNode struct {
	Type                 string
	Format               string
	Ref                  string
	Items                *SchemaNode
	AdditionalProperties bool
}

func Discover(repositoryRoot string) ([]Operation, error) {
	internalDir := filepath.Join(repositoryRoot, "internal")
	mounts := map[string]string{}
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		module := entry.Name()
		mountFile := filepath.Join(internalDir, module, "module.go")
		prefix, err := discoverMount(mountFile)
		if err != nil {
			return nil, err
		}
		if prefix != "" {
			mounts[module] = prefix
		}
	}

	operations := []Operation{
		{Method: "get", Path: "/", Module: "core"},
		{Method: "get", Path: "/healthz", Module: "core"},
		{Method: "get", Path: "/readyz", Module: "core"},
		{Method: "get", Path: "/metrics", Module: "core"},
	}
	for module, prefix := range mounts {
		routeDir := filepath.Join(internalDir, module, "transport", "http")
		if _, err := os.Stat(routeDir); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		// Routes and their handlers can live in different files of the package.
		// Resolve request metadata after collecting every handler in the module.
		requestTypes := make(map[string]string)
		bodyKinds := make(map[string]string)
		moduleStart := len(operations)
		err := filepath.WalkDir(routeDir, func(filename string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || !strings.HasSuffix(filename, ".go") || strings.HasSuffix(filename, "_test.go") {
				return nil
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
			if err != nil {
				return err
			}
			for handler, requestType := range discoverHandlerRequestTypes(parsed) {
				requestTypes[filepath.Dir(filename)+":"+parsed.Name.Name+":"+handler] = requestType
			}
			for handler, bodyKind := range discoverMultipartHandlers(parsed) {
				bodyKinds[filepath.Dir(filename)+":"+parsed.Name.Name+":"+handler] = bodyKind
			}
			routes, err := discoverRoutes(filename, module, prefix)
			if err != nil {
				return err
			}
			operations = append(operations, routes...)
			return nil
		})
		if err != nil {
			return nil, err
		}
		for i := moduleStart; i < len(operations); i++ {
			operations[i].RequestType = requestTypes[operations[i].handlerKey]
			operations[i].BodyKind = bodyKinds[operations[i].handlerKey]
		}
	}

	sort.Slice(operations, func(i, j int) bool {
		if operations[i].Path == operations[j].Path {
			return operations[i].Method < operations[j].Method
		}
		return operations[i].Path < operations[j].Path
	})
	seen := make(map[string]struct{}, len(operations))
	unique := operations[:0]
	for _, operation := range operations {
		key := operation.Method + " " + operation.Path
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		unique = append(unique, operation)
	}
	return unique, nil
}

func discoverMount(filename string) (string, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	prefix := ""
	ast.Inspect(parsed, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) == 0 {
			return true
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Mount" {
			return true
		}
		if value, ok := stringLiteral(call.Args[0]); ok && strings.HasPrefix(value, "/api/") {
			prefix = value
		}
		return true
	})
	return prefix, nil
}

func discoverRoutes(filename, module, prefix string) ([]Operation, error) {
	parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
	if err != nil {
		return nil, err
	}
	requestTypes := discoverHandlerRequestTypes(parsed)
	multipartHandlers := discoverMultipartHandlers(parsed)
	var routes []Operation
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok || len(call.Args) == 0 {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			method, ok := httpMethods[selector.Sel.Name]
			if !ok {
				return true
			}
			if _, ok := selector.X.(*ast.Ident); !ok {
				return true
			}
			route, ok := stringLiteral(call.Args[0])
			if !ok || !strings.HasPrefix(route, "/") {
				return true
			}
			fullPath := path.Join(prefix, route)
			if route == "/" {
				fullPath = strings.TrimSuffix(prefix, "/")
			}
			handler := handlerName(call.Args)
			key := handler
			if len(call.Args) > 1 && function.Recv != nil {
				if selected, ok := call.Args[1].(*ast.SelectorExpr); ok {
					if receiver, ok := selected.X.(*ast.Ident); ok {
						for _, field := range function.Recv.List {
							for _, name := range field.Names {
								if name.Name == receiver.Name {
									key = localTypeName(field.Type) + "." + handler
								}
							}
						}
					}
				}
			}
			routes = append(routes, Operation{
				Method:      method,
				Path:        fullPath,
				Module:      module,
				Handler:     handler,
				RequestType: requestTypes[key],
				BodyKind:    multipartHandlers[key],
				handlerKey:  filepath.Dir(filename) + ":" + parsed.Name.Name + ":" + key,
			})
			return true
		})
	}
	return routes, nil
}

func functionKey(function *ast.FuncDecl) string {
	if function.Recv != nil && len(function.Recv.List) > 0 {
		return localTypeName(function.Recv.List[0].Type) + "." + function.Name.Name
	}
	return function.Name.Name
}

func discoverMultipartHandlers(file *ast.File) map[string]string {
	result := make(map[string]string)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		ast.Inspect(function.Body, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if ok && (selector.Sel.Name == "FormFile" || selector.Sel.Name == "ParseMultipartForm") {
				result[functionKey(function)] = "multipart/form-data"
			}
			return true
		})
	}
	return result
}

func handlerName(args []ast.Expr) string {
	if len(args) < 2 {
		return ""
	}
	switch expression := args[1].(type) {
	case *ast.SelectorExpr:
		return expression.Sel.Name
	case *ast.Ident:
		return expression.Name
	default:
		return ""
	}
}

func discoverHandlerRequestTypes(file *ast.File) map[string]string {
	result := make(map[string]string)
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Body == nil {
			continue
		}
		variables := make(map[string]ast.Expr)
		ast.Inspect(function.Body, func(node ast.Node) bool {
			switch statement := node.(type) {
			case *ast.DeclStmt:
				general, ok := statement.Decl.(*ast.GenDecl)
				if !ok {
					return true
				}
				for _, spec := range general.Specs {
					value, ok := spec.(*ast.ValueSpec)
					if !ok || value.Type == nil {
						continue
					}
					for _, name := range value.Names {
						variables[name.Name] = value.Type
					}
				}
			case *ast.AssignStmt:
				for index, left := range statement.Lhs {
					name, ok := left.(*ast.Ident)
					if !ok || index >= len(statement.Rhs) {
						continue
					}
					if literal, ok := statement.Rhs[index].(*ast.CompositeLit); ok {
						variables[name.Name] = literal.Type
					}
				}
			case *ast.CallExpr:
				selector, ok := statement.Fun.(*ast.SelectorExpr)
				if !ok || selector.Sel.Name != "Decode" || len(statement.Args) != 1 {
					return true
				}
				unary, ok := statement.Args[0].(*ast.UnaryExpr)
				if !ok || unary.Op != token.AND {
					return true
				}
				name, ok := unary.X.(*ast.Ident)
				if !ok {
					return true
				}
				if typeName := localTypeName(variables[name.Name]); typeName != "" {
					result[functionKey(function)] = typeName
				}
			}
			return true
		})
	}
	return result
}

func localTypeName(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		return typed.Name
	case *ast.SelectorExpr:
		return typed.Sel.Name
	case *ast.StarExpr:
		return localTypeName(typed.X)
	default:
		return ""
	}
}

func stringLiteral(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func DiscoverSchemas(repositoryRoot string) (map[string]ObjectSchema, error) {
	schemas := make(map[string]ObjectSchema)
	internalDir := filepath.Join(repositoryRoot, "internal")
	entries, err := os.ReadDir(internalDir)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		module := entry.Name()
		schemaDirs := []string{
			filepath.Join(internalDir, module, "domain"),
			filepath.Join(internalDir, module, "transport", "http"),
		}
		for _, schemaDir := range schemaDirs {
			if _, err := os.Stat(schemaDir); err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}
			err := filepath.WalkDir(schemaDir, func(filename string, entry fs.DirEntry, walkErr error) error {
				if walkErr != nil {
					return walkErr
				}
				if entry.IsDir() || !strings.HasSuffix(filename, ".go") || strings.HasSuffix(filename, "_test.go") {
					return nil
				}
				parsed, err := parser.ParseFile(token.NewFileSet(), filename, nil, 0)
				if err != nil {
					return err
				}
				for _, declaration := range parsed.Decls {
					general, ok := declaration.(*ast.GenDecl)
					if !ok || general.Tok != token.TYPE {
						continue
					}
					for _, spec := range general.Specs {
						typeSpec, ok := spec.(*ast.TypeSpec)
						if !ok {
							continue
						}
						structure, ok := typeSpec.Type.(*ast.StructType)
						if !ok {
							continue
						}
						schema := objectSchema(module, typeSpec.Name.Name, structure)
						schemas[schema.Name] = schema
					}
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		}
	}
	for name, schema := range schemas {
		for index := range schema.Fields {
			normalizeSchemaReferences(&schema.Fields[index].Node, schemas)
		}
		schemas[name] = schema
	}
	return schemas, nil
}

func normalizeSchemaReferences(node *SchemaNode, schemas map[string]ObjectSchema) {
	if node == nil {
		return
	}
	if node.Ref != "" {
		name := strings.TrimPrefix(node.Ref, "#/components/schemas/")
		if _, exists := schemas[name]; !exists {
			node.Ref = ""
			node.Type = "string"
		}
	}
	normalizeSchemaReferences(node.Items, schemas)
}

func requestSchemas(operations []Operation, all map[string]ObjectSchema) map[string]ObjectSchema {
	selected := make(map[string]ObjectSchema)
	queue := make([]string, 0)
	for _, operation := range operations {
		if operation.RequestType == "" {
			continue
		}
		name := componentName(operation.Module, operation.RequestType)
		if _, exists := all[name]; exists {
			queue = append(queue, name)
		}
	}
	for len(queue) > 0 {
		name := queue[0]
		queue = queue[1:]
		if _, exists := selected[name]; exists {
			continue
		}
		schema, exists := all[name]
		if !exists {
			continue
		}
		selected[name] = schema
		for _, field := range schema.Fields {
			collectSchemaReferences(field.Node, &queue)
		}
	}
	return selected
}

func collectSchemaReferences(node SchemaNode, queue *[]string) {
	if node.Ref != "" {
		*queue = append(*queue, strings.TrimPrefix(node.Ref, "#/components/schemas/"))
	}
	if node.Items != nil {
		collectSchemaReferences(*node.Items, queue)
	}
}

func objectSchema(module, typeName string, structure *ast.StructType) ObjectSchema {
	schema := ObjectSchema{Name: componentName(module, typeName), Module: module}
	for _, field := range structure.Fields.List {
		if len(field.Names) == 0 {
			continue
		}
		for _, fieldName := range field.Names {
			jsonName, omitEmpty, skip := jsonFieldName(fieldName.Name, field.Tag)
			if skip {
				continue
			}
			schema.Fields = append(schema.Fields, SchemaField{
				Name: jsonName,
				Node: schemaNode(module, field.Type),
			})
			if !omitEmpty {
				if _, pointer := field.Type.(*ast.StarExpr); !pointer {
					schema.Required = append(schema.Required, jsonName)
				}
			}
		}
	}
	return schema
}

func jsonFieldName(fallback string, tag *ast.BasicLit) (name string, omitEmpty, skip bool) {
	name = fallback
	if tag == nil {
		return name, false, false
	}
	value, err := strconv.Unquote(tag.Value)
	if err != nil {
		return name, false, false
	}
	jsonTag := reflect.StructTag(value).Get("json")
	if jsonTag == "" {
		return name, false, false
	}
	parts := strings.Split(jsonTag, ",")
	if parts[0] == "-" {
		return "", false, true
	}
	if parts[0] != "" {
		name = parts[0]
	}
	for _, option := range parts[1:] {
		if option == "omitempty" {
			omitEmpty = true
		}
	}
	return name, omitEmpty, false
}

func schemaNode(module string, expression ast.Expr) SchemaNode {
	switch typed := expression.(type) {
	case *ast.StarExpr:
		return schemaNode(module, typed.X)
	case *ast.ArrayType:
		item := schemaNode(module, typed.Elt)
		if identifier, ok := typed.Elt.(*ast.Ident); ok && identifier.Name == "byte" {
			return SchemaNode{Type: "string", Format: "byte"}
		}
		return SchemaNode{Type: "array", Items: &item}
	case *ast.MapType, *ast.InterfaceType:
		return SchemaNode{Type: "object", AdditionalProperties: true}
	case *ast.StructType:
		return SchemaNode{Type: "object", AdditionalProperties: true}
	case *ast.Ident:
		switch typed.Name {
		case "string":
			return SchemaNode{Type: "string"}
		case "bool":
			return SchemaNode{Type: "boolean"}
		case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64":
			return SchemaNode{Type: "integer", Format: "int64"}
		case "float32", "float64":
			return SchemaNode{Type: "number", Format: "double"}
		case "any", "interface{}":
			return SchemaNode{Type: "object", AdditionalProperties: true}
		default:
			return SchemaNode{Ref: "#/components/schemas/" + componentName(module, typed.Name)}
		}
	case *ast.SelectorExpr:
		packageName := ""
		if identifier, ok := typed.X.(*ast.Ident); ok {
			packageName = identifier.Name
		}
		switch packageName + "." + typed.Sel.Name {
		case "uuid.UUID":
			return SchemaNode{Type: "string", Format: "uuid"}
		case "time.Time":
			return SchemaNode{Type: "string", Format: "date-time"}
		case "decimal.Decimal":
			return SchemaNode{Type: "number", Format: "decimal"}
		case "json.RawMessage":
			return SchemaNode{Type: "object", AdditionalProperties: true}
		default:
			for _, suffix := range []string{"Status", "Type", "Channel", "Category", "Priority", "Method", "Scope"} {
				if strings.HasSuffix(typed.Sel.Name, suffix) {
					return SchemaNode{Type: "string"}
				}
			}
			return SchemaNode{Type: "object", AdditionalProperties: true}
		}
	default:
		return SchemaNode{Type: "object", AdditionalProperties: true}
	}
}

func componentName(module, typeName string) string {
	return module + "_" + typeName
}

func writeParameters(output *bytes.Buffer, operation Operation) {
	pathParameters := pathParameterNames(operation.Path)
	includeTenant := strings.HasPrefix(operation.Path, "/api/v1/") && !isPublic(operation)
	if len(pathParameters) == 0 && !includeTenant {
		return
	}
	output.WriteString("      parameters:\n")
	for _, name := range pathParameters {
		fmt.Fprintf(output, "        - name: %s\n", name)
		output.WriteString("          in: path\n          required: true\n          schema:\n            type: string\n")
		if strings.Contains(strings.ToLower(name), "id") {
			output.WriteString("            format: uuid\n")
		}
	}
	if includeTenant {
		output.WriteString("        - $ref: '#/components/parameters/LegalEntityQuery'\n")
		output.WriteString("        - $ref: '#/components/parameters/LegalEntityHeader'\n")
	}
	if operation.Method == "get" && operation.Path == "/api/v1/authz/assignments" {
		output.WriteString("        - name: user_id\n          in: query\n          required: true\n          schema:\n            type: string\n            format: uuid\n")
	}
	if operation.Method == "get" && operation.Path == "/api/v1/auth/users" {
		output.WriteString("        - name: search\n          in: query\n          schema:\n            type: string\n            maxLength: 200\n")
		output.WriteString("        - name: limit\n          in: query\n          schema:\n            type: integer\n            minimum: 1\n            maximum: 100\n            default: 50\n")
		output.WriteString("        - name: offset\n          in: query\n          schema:\n            type: integer\n            minimum: 0\n            default: 0\n")
	}
}

func pathParameterNames(routePath string) []string {
	var names []string
	for remaining := routePath; ; {
		start := strings.IndexByte(remaining, '{')
		if start < 0 {
			break
		}
		remaining = remaining[start+1:]
		end := strings.IndexByte(remaining, '}')
		if end < 0 {
			break
		}
		if name := remaining[:end]; name != "" {
			names = append(names, name)
		}
		remaining = remaining[end+1:]
	}
	return names
}

func writeRequestBody(output *bytes.Buffer, operation Operation, schemas map[string]ObjectSchema) {
	switch operation.Method {
	case "post", "put", "patch":
	default:
		return
	}
	schemaRef := "#/components/schemas/GenericRequest"
	required := false
	if operation.BodyKind == "multipart/form-data" {
		schemaRef = "#/components/schemas/MultipartUploadRequest"
		required = true
	} else if operation.RequestType != "" {
		candidate := componentName(operation.Module, operation.RequestType)
		if _, exists := schemas[candidate]; exists {
			schemaRef = "#/components/schemas/" + candidate
			required = true
		}
	}
	output.WriteString("      requestBody:\n")
	fmt.Fprintf(output, "        required: %t\n", required)
	output.WriteString("        content:\n")
	contentType := operation.BodyKind
	if contentType == "" {
		contentType = "application/json"
	}
	fmt.Fprintf(output, "          %s:\n", contentType)
	output.WriteString("            schema:\n")
	fmt.Fprintf(output, "              $ref: '%s'\n", schemaRef)
}

func writeResponses(output *bytes.Buffer, operation Operation) {
	output.WriteString("      responses:\n")
	switch operation.Method {
	case "post":
		writeSuccessResponse(output, "200", "Operation completed")
		writeSuccessResponse(output, "201", "Resource created")
	case "delete":
		writeSuccessResponse(output, "200", "Resource deleted")
		output.WriteString("        '204':\n          description: Resource deleted with no response body\n")
	default:
		writeSuccessResponse(output, "200", "Successful response")
	}
	output.WriteString("        '400':\n          $ref: '#/components/responses/BadRequest'\n")
	if !isPublic(operation) {
		output.WriteString("        '401':\n          $ref: '#/components/responses/Unauthorized'\n")
		output.WriteString("        '403':\n          $ref: '#/components/responses/Forbidden'\n")
	}
	if len(pathParameterNames(operation.Path)) > 0 {
		output.WriteString("        '404':\n          $ref: '#/components/responses/NotFound'\n")
	}
	output.WriteString("        '429':\n          $ref: '#/components/responses/RateLimited'\n")
	output.WriteString("        '500':\n          $ref: '#/components/responses/InternalError'\n")
}

func writeSuccessResponse(output *bytes.Buffer, status, description string) {
	fmt.Fprintf(output, "        '%s':\n", status)
	fmt.Fprintf(output, "          description: %s\n", description)
	output.WriteString("          content:\n")
	output.WriteString("            application/json:\n")
	output.WriteString("              schema:\n")
	output.WriteString("                $ref: '#/components/schemas/GenericResponse'\n")
}

func writeSchemas(output *bytes.Buffer, schemas map[string]ObjectSchema) {
	names := make([]string, 0, len(schemas))
	for name := range schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		schema := schemas[name]
		fmt.Fprintf(output, "    %s:\n", name)
		output.WriteString("      type: object\n")
		if len(schema.Required) > 0 {
			output.WriteString("      required:\n")
			for _, required := range schema.Required {
				fmt.Fprintf(output, "        - %s\n", required)
			}
		}
		if len(schema.Fields) == 0 {
			output.WriteString("      additionalProperties: true\n")
			continue
		}
		output.WriteString("      properties:\n")
		for _, field := range schema.Fields {
			fmt.Fprintf(output, "        %s:\n", field.Name)
			writeSchemaNode(output, field.Node, "          ")
		}
	}
}

func writeSchemaNode(output *bytes.Buffer, node SchemaNode, indentation string) {
	if node.Ref != "" {
		fmt.Fprintf(output, "%s$ref: '%s'\n", indentation, node.Ref)
		return
	}
	if node.Type == "" {
		node.Type = "object"
	}
	fmt.Fprintf(output, "%stype: %s\n", indentation, node.Type)
	if node.Format != "" {
		fmt.Fprintf(output, "%sformat: %s\n", indentation, node.Format)
	}
	if node.Type == "array" && node.Items != nil {
		fmt.Fprintf(output, "%sitems:\n", indentation)
		writeSchemaNode(output, *node.Items, indentation+"  ")
	}
	if node.AdditionalProperties {
		fmt.Fprintf(output, "%sadditionalProperties: true\n", indentation)
	}
}

func humanize(identifier string) string {
	if identifier == "" {
		return ""
	}
	var result strings.Builder
	for index, character := range identifier {
		if index > 0 && character >= 'A' && character <= 'Z' {
			result.WriteByte(' ')
		}
		result.WriteRune(character)
	}
	return result.String()
}

func Generate(repositoryRoot string) ([]byte, error) {
	operations, err := Discover(repositoryRoot)
	if err != nil {
		return nil, err
	}
	allSchemas, err := DiscoverSchemas(repositoryRoot)
	if err != nil {
		return nil, err
	}
	schemas := requestSchemas(operations, allSchemas)
	grouped := make(map[string][]Operation)
	for _, operation := range operations {
		grouped[operation.Path] = append(grouped[operation.Path], operation)
	}
	paths := make([]string, 0, len(grouped))
	for routePath := range grouped {
		paths = append(paths, routePath)
	}
	sort.Strings(paths)

	var output bytes.Buffer
	output.WriteString(`# Code generated by go run ./cmd/openapi-gen; DO NOT EDIT.
openapi: 3.1.0
info:
  title: ERP-GO API
  version: 1.0.0
  description: |
    HTTP route contract generated directly from the Chi router declarations.
    Regulatory certification and external-provider acceptance are outside this contract.
servers:
  - url: /
    description: Same-origin deployment
tags:
`)
	modules := make(map[string]struct{})
	for _, operation := range operations {
		modules[operation.Module] = struct{}{}
	}
	moduleNames := make([]string, 0, len(modules))
	for module := range modules {
		moduleNames = append(moduleNames, module)
	}
	sort.Strings(moduleNames)
	for _, module := range moduleNames {
		fmt.Fprintf(&output, "  - name: %s\n", module)
	}
	output.WriteString("paths:\n")
	for _, routePath := range paths {
		fmt.Fprintf(&output, "  %s:\n", routePath)
		for _, operation := range grouped[routePath] {
			fmt.Fprintf(&output, "    %s:\n", operation.Method)
			fmt.Fprintf(&output, "      tags: [%s]\n", operation.Module)
			fmt.Fprintf(&output, "      operationId: %s\n", operationID(operation))
			if operation.Handler != "" {
				fmt.Fprintf(&output, "      summary: %s\n", strconv.Quote(humanize(operation.Handler)))
			}
			if isPublic(operation) {
				output.WriteString("      security: []\n")
			} else if operation.Path == "/metrics" {
				output.WriteString("      security:\n        - MetricsBearer: []\n")
			}
			writeParameters(&output, operation)
			writeRequestBody(&output, operation, schemas)
			writeResponses(&output, operation)
		}
	}
	output.WriteString(`components:
  securitySchemes:
    SessionCookie:
      type: apiKey
      in: cookie
      name: erp_session
    MetricsBearer:
      type: http
      scheme: bearer
  parameters:
    LegalEntityQuery:
      name: legal_entity_id
      in: query
      required: false
      description: Legal-entity scope. If also supplied in the header or body, every value must match.
      schema:
        type: string
        format: uuid
    LegalEntityHeader:
      name: X-Legal-Entity-ID
      in: header
      required: false
      description: Optional legal-entity boundary for tenant-scoped authorization.
      schema:
        type: string
        format: uuid
  responses:
    BadRequest:
      description: Invalid request or conflicting tenant scope
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    Unauthorized:
      description: Authentication is required or the session has expired
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    Forbidden:
      description: The authenticated user lacks permission in the requested scope
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    NotFound:
      description: Resource not found, including resources hidden by tenant isolation
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    RateLimited:
      description: Request rate limit exceeded
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
    InternalError:
      description: Unexpected server error
      content:
        application/json:
          schema:
            $ref: '#/components/schemas/ErrorResponse'
  schemas:
    GenericRequest:
      type: object
      additionalProperties: true
    GenericResponse:
      type: object
      additionalProperties: true
    MultipartUploadRequest:
      type: object
      required: [legal_entity_id, file]
      properties:
        legal_entity_id:
          type: string
          format: uuid
        file:
          type: string
          format: binary
        title:
          type: string
        category:
          type: string
        folder_path:
          type: string
        entity_type:
          type: string
        entity_id:
          type: string
          format: uuid
    ErrorResponse:
      type: object
      required: [error]
      properties:
        error:
          oneOf:
            - type: string
            - type: object
              additionalProperties: true
`)
	writeSchemas(&output, schemas)
	output.WriteString(`
security:
  - SessionCookie: []
`)
	return output.Bytes(), nil
}

func isPublic(operation Operation) bool {
	if operation.Path == "/" || operation.Path == "/healthz" || operation.Path == "/readyz" {
		return true
	}
	return operation.Method == "post" && (operation.Path == "/api/v1/auth/login" || operation.Path == "/api/v1/auth/login/mfa")
}

func operationID(operation Operation) string {
	replacer := strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_", ".", "_")
	suffix := strings.Trim(replacer.Replace(operation.Path), "_")
	if suffix == "" {
		suffix = "root"
	}
	return operation.Module + "_" + operation.Method + "_" + suffix
}
