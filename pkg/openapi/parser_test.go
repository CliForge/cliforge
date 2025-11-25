package openapi

import (
	"context"
	"strings"
	"testing"
)

func TestParser_DetectVersion(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		expected    SpecVersion
		expectError bool
	}{
		{
			name: "OpenAPI 3.0",
			spec: `{
				"openapi": "3.0.0",
				"info": {"title": "Test", "version": "1.0.0"}
			}`,
			expected: SpecVersionOpenAPI3,
		},
		{
			name: "OpenAPI 3.1",
			spec: `{
				"openapi": "3.1.0",
				"info": {"title": "Test", "version": "1.0.0"}
			}`,
			expected: SpecVersionOpenAPI31,
		},
		{
			name: "Swagger 2.0",
			spec: `{
				"swagger": "2.0",
				"info": {"title": "Test", "version": "1.0.0"}
			}`,
			expected: SpecVersionSwagger2,
		},
		{
			name: "Invalid spec",
			spec: `{
				"title": "Test"
			}`,
			expectError: true,
		},
	}

	parser := NewParser()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := parser.detectVersion([]byte(tt.spec))
			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if version != tt.expected {
				t.Errorf("expected version %s, got %s", tt.expected, version)
			}
		})
	}
}

func TestParser_ParseOpenAPI3(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0",
			"x-cli-version": "2024.01.01",
			"x-cli-min-version": "1.0.0"
		},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"summary": "List users",
					"x-cli-command": "list users",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	if parsed.OriginalVersion != SpecVersionOpenAPI3 {
		t.Errorf("expected version OpenAPI3, got %s", parsed.OriginalVersion)
	}

	info := parsed.GetInfo()
	if info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got '%s'", info.Title)
	}
	if info.CLIVersion != "2024.01.01" {
		t.Errorf("expected CLI version '2024.01.01', got '%s'", info.CLIVersion)
	}
	if info.MinCLIVersion != "1.0.0" {
		t.Errorf("expected min CLI version '1.0.0', got '%s'", info.MinCLIVersion)
	}

	operations, err := parsed.GetOperations()
	if err != nil {
		t.Fatalf("failed to get operations: %v", err)
	}
	if len(operations) != 1 {
		t.Errorf("expected 1 operation, got %d", len(operations))
	}

	op := operations[0]
	if op.CLICommand != "list users" {
		t.Errorf("expected CLI command 'list users', got '%s'", op.CLICommand)
	}
}

func TestParser_ParseSwagger2(t *testing.T) {
	spec := `{
		"swagger": "2.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"summary": "List users",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	if parsed.OriginalVersion != SpecVersionSwagger2 {
		t.Errorf("expected version Swagger2, got %s", parsed.OriginalVersion)
	}

	// Should be converted to OpenAPI 3.0
	if parsed.Spec.OpenAPI == "" {
		t.Error("expected spec to be converted to OpenAPI 3.0")
	}
}

func TestParser_HiddenOperations(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"summary": "List users",
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			},
			"/internal/debug": {
				"get": {
					"operationId": "debugEndpoint",
					"summary": "Debug endpoint",
					"x-cli-hidden": true,
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	operations, err := parsed.GetOperations()
	if err != nil {
		t.Fatalf("failed to get operations: %v", err)
	}

	// Should only return non-hidden operation
	if len(operations) != 1 {
		t.Errorf("expected 1 visible operation, got %d", len(operations))
	}
	if operations[0].OperationID != "listUsers" {
		t.Errorf("expected operation 'listUsers', got '%s'", operations[0].OperationID)
	}
}

func TestParser_ParseFile(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	spec, err := parser.ParseFile(ctx, "testdata/simple_openapi3.json")
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	if spec == nil {
		t.Fatal("spec is nil")
	}

	info := spec.GetInfo()
	if info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got '%s'", info.Title)
	}
}

func TestParser_ParseFile_NotFound(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	_, err := parser.ParseFile(ctx, "testdata/nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestParser_ParseReader(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Reader Test API",
			"version": "1.0.0"
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()

	reader := strings.NewReader(spec)
	parsed, err := parser.ParseReader(ctx, reader)
	if err != nil {
		t.Fatalf("failed to parse from reader: %v", err)
	}

	info := parsed.GetInfo()
	if info.Title != "Reader Test API" {
		t.Errorf("expected title 'Reader Test API', got '%s'", info.Title)
	}
}

func TestParser_ParseReader_InvalidJSON(t *testing.T) {
	parser := NewParser()
	ctx := context.Background()

	reader := strings.NewReader("invalid json")
	_, err := parser.ParseReader(ctx, reader)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParser_DisableValidation(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test",
			"version": "1.0.0"
		},
		"paths": {}
	}`

	parser := NewParser()
	parser.DisableValidation = true
	ctx := context.Background()

	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	if parsed == nil {
		t.Fatal("spec is nil")
	}
}

func TestParser_ParseOperationExtensions(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"get": {
					"operationId": "test",
					"summary": "Test operation",
					"x-cli-command": "test-cmd",
					"x-cli-aliases": ["tc", "test"],
					"x-cli-parent-resource": "parent",
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	operations, err := parsed.GetOperations()
	if err != nil {
		t.Fatalf("failed to get operations: %v", err)
	}

	if len(operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(operations))
	}

	op := operations[0]
	if op.CLICommand != "test-cmd" {
		t.Errorf("expected CLI command 'test-cmd', got '%s'", op.CLICommand)
	}
	if len(op.CLIAliases) != 2 {
		t.Errorf("expected 2 aliases, got %d", len(op.CLIAliases))
	}
	if op.CLIParentRes != "parent" {
		t.Errorf("expected parent resource 'parent', got '%s'", op.CLIParentRes)
	}
}

func TestParser_UnsupportedSwaggerVersion(t *testing.T) {
	spec := `{
		"swagger": "1.2",
		"info": {"title": "Test", "version": "1.0.0"}
	}`

	parser := NewParser()
	ctx := context.Background()

	_, err := parser.Parse(ctx, []byte(spec))
	if err == nil {
		t.Error("expected error for unsupported Swagger version")
	}
}

func TestParser_UnsupportedOpenAPIVersion(t *testing.T) {
	spec := `{
		"openapi": "4.0.0",
		"info": {"title": "Test", "version": "1.0.0"}
	}`

	parser := NewParser()
	ctx := context.Background()

	_, err := parser.Parse(ctx, []byte(spec))
	if err == nil {
		t.Error("expected error for unsupported OpenAPI version")
	}
}
