package openapi

import (
	"context"
	"testing"
)

func TestNewValidator(t *testing.T) {
	validator := NewValidator()

	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}
	if validator.Strict {
		t.Error("expected Strict to be false by default")
	}
	if !validator.AllowUnknownExtensions {
		t.Error("expected AllowUnknownExtensions to be true by default")
	}
}

func TestValidator_Validate_BasicSpec(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_CLIConfig(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"x-cli-config": {
			"name": "myapi",
			"version": "1.0.0",
			"branding": {
				"colors": {
					"primary": "#0066CC",
					"secondary": "#FF6600"
				}
			},
			"auth": {
				"type": "oauth2",
				"storage": "keyring"
			},
			"output": {
				"default-format": "json",
				"supported-formats": ["json", "yaml"]
			}
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_InvalidHexColor(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"x-cli-config": {
			"branding": {
				"colors": {
					"primary": "invalid-color"
				}
			}
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to bad color")
	}

	foundColorError := false
	for _, e := range result.Errors {
		if e.Field == "primary" {
			foundColorError = true
			break
		}
	}
	if !foundColorError {
		t.Error("expected color validation error")
	}
}

func TestValidator_Validate_InvalidAuthType(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"x-cli-config": {
			"auth": {
				"type": "invalid-auth-type"
			}
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to bad auth type")
	}
}

func TestValidator_Validate_InvalidStorageType(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"x-cli-config": {
			"auth": {
				"type": "oauth2",
				"storage": "invalid-storage"
			}
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to bad storage type")
	}
}

func TestValidator_Validate_InvalidOutputFormat(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"x-cli-config": {
			"output": {
				"default-format": "invalid-format"
			}
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to bad output format")
	}
}

func TestValidator_Validate_Changelog(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test",
			"version": "1.0.0",
			"x-cli-changelog": [
				{
					"version": "1.0.0",
					"date": "2024-01-01",
					"changes": [
						{
							"type": "added",
							"severity": "safe",
							"description": "New feature"
						}
					]
				}
			]
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_InvalidChangeType(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test",
			"version": "1.0.0",
			"x-cli-changelog": [
				{
					"version": "1.0.0",
					"date": "2024-01-01",
					"changes": [
						{
							"type": "invalid-type",
							"severity": "safe",
							"description": "Test"
						}
					]
				}
			]
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, err := parser.Parse(ctx, []byte(spec))
	if err != nil {
		t.Fatalf("failed to parse spec: %v", err)
	}

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to bad change type")
	}
}

func TestValidator_Validate_OperationFlags(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-flags": [
						{
							"name": "test-flag",
							"source": "test",
							"flag": "--test",
							"type": "string"
						}
					]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_InvalidFlagType(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-flags": [
						{
							"name": "test-flag",
							"type": "invalid-type"
						}
					]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to bad flag type")
	}
}

func TestValidator_Validate_InteractivePrompts(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-interactive": {
						"enabled": true,
						"prompts": [
							{
								"parameter": "name",
								"type": "text",
								"message": "Enter name:"
							}
						]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_InvalidPromptType(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-interactive": {
						"enabled": true,
						"prompts": [
							{
								"parameter": "name",
								"type": "invalid-type",
								"message": "Test"
							}
						]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to bad prompt type")
	}
}

func TestValidator_Validate_Preflight(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-preflight": [
						{
							"name": "check",
							"endpoint": "/health",
							"method": "GET"
						}
					]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_PreflightMissingEndpoint(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-preflight": [
						{
							"name": "check"
						}
					]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to missing preflight endpoint")
	}
}

func TestValidator_Validate_Async(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-async": {
						"enabled": true,
						"status-endpoint": "/status/{id}",
						"terminal-states": ["completed", "failed"],
						"polling": {
							"interval": 5,
							"timeout": 300
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_AsyncMissingEndpoint(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-async": {
						"enabled": true,
						"terminal-states": ["completed"]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to missing async endpoint")
	}
}

func TestValidator_Validate_Workflow(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-workflow": {
						"steps": [
							{
								"id": "step1",
								"request": {
									"method": "GET",
									"url": "/test"
								}
							}
						]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected valid spec, got errors: %v", result.Errors)
	}
}

func TestValidator_Validate_WorkflowMissingStepID(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "test",
					"responses": {"200": {"description": "OK"}},
					"x-cli-workflow": {
						"steps": [
							{
								"request": {
									"method": "GET",
									"url": "/test"
								}
							}
						]
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to missing workflow step ID")
	}
}

func TestValidator_Validate_DuplicateCommands(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test1": {
				"get": {
					"operationId": "test1",
					"x-cli-command": "duplicate-cmd",
					"responses": {"200": {"description": "OK"}}
				}
			},
			"/test2": {
				"get": {
					"operationId": "test2",
					"x-cli-command": "duplicate-cmd",
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

	validator := NewValidator()
	result, err := validator.Validate(ctx, parsed)
	if err != nil {
		t.Fatalf("Validate returned error: %v", err)
	}

	if result.Valid {
		t.Error("expected invalid spec due to duplicate commands")
	}

	foundDuplicateError := false
	for _, e := range result.Errors {
		if e.Path == "commands" && e.Field == "duplicate-cmd" {
			foundDuplicateError = true
			break
		}
	}
	if !foundDuplicateError {
		t.Error("expected duplicate command error")
	}
}

func TestValidationError_Error(t *testing.T) {
	err := ValidationError{
		Path:    "test.path",
		Field:   "field",
		Message: "test message",
	}

	expected := "test.path.field: test message"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}

	errNoField := ValidationError{
		Path:    "test.path",
		Message: "test message",
	}

	expected = "test.path: test message"
	if errNoField.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, errNoField.Error())
	}
}

func TestValidationWarning_String(t *testing.T) {
	warn := ValidationWarning{
		Path:    "test.path",
		Field:   "field",
		Message: "test message",
	}

	expected := "test.path.field: test message"
	if warn.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, warn.String())
	}
}

func TestIsValidHexColor(t *testing.T) {
	tests := []struct {
		color string
		valid bool
	}{
		{"#FFF", true},
		{"#FFFFFF", true},
		{"#000", true},
		{"#000000", true},
		{"#0066CC", true},
		{"invalid", false},
		{"#GG0000", false},
		{"#FF", false},
		{"#FFFFFFF", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			result := isValidHexColor(tt.color)
			if result != tt.valid {
				t.Errorf("isValidHexColor(%s) = %v, want %v", tt.color, result, tt.valid)
			}
		})
	}
}
