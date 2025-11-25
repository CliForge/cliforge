package openapi

import (
	"context"
	"testing"
)

func TestParseCLIConfig(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"x-cli-config": {
			"name": "myapi-cli",
			"version": "1.0.0",
			"description": "My API CLI",
			"branding": {
				"ascii-art": "ASCII Art Here",
				"colors": {
					"primary": "#0066CC",
					"secondary": "#FF6600"
				}
			},
			"auth": {
				"type": "oauth2",
				"storage": "keyring",
				"auto-refresh": true
			},
			"output": {
				"default-format": "table",
				"supported-formats": ["table", "json", "yaml"]
			},
			"features": {
				"interactive-mode": true,
				"auto-complete": true,
				"self-update": true,
				"telemetry": false
			},
			"cache": {
				"enabled": true,
				"ttl": 300,
				"location": "~/.cache/myapi-cli"
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

	config := parsed.Extensions.Config
	if config == nil {
		t.Fatal("x-cli-config not parsed")
	}

	if config.Name != "myapi-cli" {
		t.Errorf("expected name 'myapi-cli', got '%s'", config.Name)
	}
	if config.Version != "1.0.0" {
		t.Errorf("expected version '1.0.0', got '%s'", config.Version)
	}

	if config.Branding == nil {
		t.Fatal("branding not parsed")
	}
	if config.Branding.Colors.Primary != "#0066CC" {
		t.Errorf("expected primary color '#0066CC', got '%s'", config.Branding.Colors.Primary)
	}

	if config.Auth == nil {
		t.Fatal("auth not parsed")
	}
	if config.Auth.Type != "oauth2" {
		t.Errorf("expected auth type 'oauth2', got '%s'", config.Auth.Type)
	}
	if !config.Auth.AutoRefresh {
		t.Error("expected auto-refresh to be true")
	}

	if config.Output == nil {
		t.Fatal("output not parsed")
	}
	if config.Output.DefaultFormat != "table" {
		t.Errorf("expected default format 'table', got '%s'", config.Output.DefaultFormat)
	}
	if len(config.Output.SupportedFormats) != 3 {
		t.Errorf("expected 3 supported formats, got %d", len(config.Output.SupportedFormats))
	}

	if config.Features == nil {
		t.Fatal("features not parsed")
	}
	if !config.Features.InteractiveMode {
		t.Error("expected interactive-mode to be true")
	}
	if config.Features.Telemetry {
		t.Error("expected telemetry to be false")
	}

	if config.Cache == nil {
		t.Fatal("cache not parsed")
	}
	if !config.Cache.Enabled {
		t.Error("expected cache to be enabled")
	}
	if config.Cache.TTL != 300 {
		t.Errorf("expected TTL 300, got %d", config.Cache.TTL)
	}
}

func TestParseCLIFlags(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/clusters": {
				"post": {
					"operationId": "createCluster",
					"responses": {
						"201": {
							"description": "Created"
						}
					},
					"x-cli-flags": [
						{
							"name": "cluster-name",
							"source": "name",
							"flag": "--cluster-name",
							"aliases": ["-n"],
							"required": true,
							"type": "string",
							"description": "Cluster name"
						},
						{
							"name": "multi-az",
							"source": "multi_az",
							"flag": "--multi-az",
							"type": "boolean",
							"default": false,
							"description": "Multi-AZ deployment"
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

	operations, err := parsed.GetOperations()
	if err != nil {
		t.Fatalf("failed to get operations: %v", err)
	}

	if len(operations) != 1 {
		t.Fatalf("expected 1 operation, got %d", len(operations))
	}

	op := operations[0]
	if len(op.CLIFlags) != 2 {
		t.Fatalf("expected 2 flags, got %d", len(op.CLIFlags))
	}

	flag1 := op.CLIFlags[0]
	if flag1.Name != "cluster-name" {
		t.Errorf("expected flag name 'cluster-name', got '%s'", flag1.Name)
	}
	if flag1.Source != "name" {
		t.Errorf("expected source 'name', got '%s'", flag1.Source)
	}
	if flag1.Flag != "--cluster-name" {
		t.Errorf("expected flag '--cluster-name', got '%s'", flag1.Flag)
	}
	if !flag1.Required {
		t.Error("expected required to be true")
	}
	if len(flag1.Aliases) != 1 || flag1.Aliases[0] != "-n" {
		t.Errorf("expected alias '-n', got %v", flag1.Aliases)
	}

	flag2 := op.CLIFlags[1]
	if flag2.Type != "boolean" {
		t.Errorf("expected type 'boolean', got '%s'", flag2.Type)
	}
	if flag2.Default != false {
		t.Errorf("expected default false, got %v", flag2.Default)
	}
}

func TestParseCLIInteractive(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/clusters": {
				"post": {
					"operationId": "createCluster",
					"responses": {
						"201": {
							"description": "Created"
						}
					},
					"x-cli-interactive": {
						"enabled": true,
						"prompts": [
							{
								"parameter": "name",
								"type": "text",
								"message": "Cluster name:",
								"validation": "^[a-z][a-z0-9-]{0,53}$",
								"validation-message": "Must be lowercase alphanumeric"
							},
							{
								"parameter": "region",
								"type": "select",
								"message": "Select region:",
								"source": {
									"endpoint": "/api/v1/regions",
									"value-field": "id",
									"display-field": "display_name"
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

	operations, err := parsed.GetOperations()
	if err != nil {
		t.Fatalf("failed to get operations: %v", err)
	}

	op := operations[0]
	if op.CLIInteractive == nil {
		t.Fatal("x-cli-interactive not parsed")
	}

	if !op.CLIInteractive.Enabled {
		t.Error("expected enabled to be true")
	}

	if len(op.CLIInteractive.Prompts) != 2 {
		t.Fatalf("expected 2 prompts, got %d", len(op.CLIInteractive.Prompts))
	}

	prompt1 := op.CLIInteractive.Prompts[0]
	if prompt1.Type != "text" {
		t.Errorf("expected type 'text', got '%s'", prompt1.Type)
	}
	if prompt1.Validation != "^[a-z][a-z0-9-]{0,53}$" {
		t.Errorf("expected validation pattern, got '%s'", prompt1.Validation)
	}

	prompt2 := op.CLIInteractive.Prompts[1]
	if prompt2.Type != "select" {
		t.Errorf("expected type 'select', got '%s'", prompt2.Type)
	}
	if prompt2.Source == nil {
		t.Fatal("expected source to be set")
	}
	if prompt2.Source.Endpoint != "/api/v1/regions" {
		t.Errorf("expected endpoint '/api/v1/regions', got '%s'", prompt2.Source.Endpoint)
	}
}

func TestParseCLIAsync(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/clusters/{id}": {
				"parameters": [
					{
						"name": "id",
						"in": "path",
						"required": true,
						"schema": {
							"type": "string"
						}
					}
				],
				"delete": {
					"operationId": "deleteCluster",
					"responses": {
						"202": {
							"description": "Accepted"
						}
					},
					"x-cli-async": {
						"enabled": true,
						"status-field": "state",
						"status-endpoint": "/api/v1/clusters/{id}",
						"terminal-states": ["deleted", "error"],
						"polling": {
							"interval": 30,
							"timeout": 3600,
							"backoff": {
								"enabled": true,
								"multiplier": 1.5,
								"max-interval": 300
							}
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

	op := operations[0]
	if op.CLIAsync == nil {
		t.Fatal("x-cli-async not parsed")
	}

	if !op.CLIAsync.Enabled {
		t.Error("expected enabled to be true")
	}
	if op.CLIAsync.StatusField != "state" {
		t.Errorf("expected status-field 'state', got '%s'", op.CLIAsync.StatusField)
	}
	if len(op.CLIAsync.TerminalStates) != 2 {
		t.Errorf("expected 2 terminal states, got %d", len(op.CLIAsync.TerminalStates))
	}

	if op.CLIAsync.Polling == nil {
		t.Fatal("polling not parsed")
	}
	if op.CLIAsync.Polling.Interval != 30 {
		t.Errorf("expected interval 30, got %d", op.CLIAsync.Polling.Interval)
	}

	if op.CLIAsync.Polling.Backoff == nil {
		t.Fatal("backoff not parsed")
	}
	if op.CLIAsync.Polling.Backoff.Multiplier != 1.5 {
		t.Errorf("expected multiplier 1.5, got %f", op.CLIAsync.Polling.Backoff.Multiplier)
	}
}

func TestParseCLIWorkflow(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/deploy": {
				"post": {
					"operationId": "deploy",
					"responses": {
						"201": {
							"description": "Created"
						}
					},
					"x-cli-workflow": {
						"steps": [
							{
								"id": "check-readiness",
								"request": {
									"method": "GET",
									"url": "{base_url}/api/readiness"
								}
							},
							{
								"id": "create-deployment",
								"request": {
									"method": "POST",
									"url": "{base_url}/api/deployments",
									"body": {
										"app_id": "{args.app_id}"
									}
								},
								"condition": "check-readiness.body.ready == true"
							}
						],
						"output": {
							"format": "json",
							"transform": "{\"deployment_id\": create-deployment.body.id}"
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

	op := operations[0]
	if op.CLIWorkflow == nil {
		t.Fatal("x-cli-workflow not parsed")
	}

	if len(op.CLIWorkflow.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(op.CLIWorkflow.Steps))
	}

	step1 := op.CLIWorkflow.Steps[0]
	if step1.ID != "check-readiness" {
		t.Errorf("expected id 'check-readiness', got '%s'", step1.ID)
	}
	if step1.Request.Method != "GET" {
		t.Errorf("expected method 'GET', got '%s'", step1.Request.Method)
	}

	step2 := op.CLIWorkflow.Steps[1]
	if step2.Condition != "check-readiness.body.ready == true" {
		t.Errorf("expected condition, got '%s'", step2.Condition)
	}

	if op.CLIWorkflow.Output == nil {
		t.Fatal("output not parsed")
	}
	if op.CLIWorkflow.Output.Format != "json" {
		t.Errorf("expected format 'json', got '%s'", op.CLIWorkflow.Output.Format)
	}
}

func TestParseChangelog(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0",
			"x-cli-changelog": [
				{
					"date": "2024-11-09",
					"version": "2.1.0",
					"changes": [
						{
							"type": "added",
							"severity": "safe",
							"description": "New analytics endpoint",
							"path": "/analytics"
						},
						{
							"type": "deprecated",
							"severity": "dangerous",
							"description": "GET /v1/users is deprecated",
							"path": "/v1/users",
							"migration": "Use GET /v2/users instead",
							"sunset": "2025-12-31"
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

	changelog := parsed.Extensions.Changelog
	if len(changelog) != 1 {
		t.Fatalf("expected 1 changelog entry, got %d", len(changelog))
	}

	entry := changelog[0]
	if entry.Version != "2.1.0" {
		t.Errorf("expected version '2.1.0', got '%s'", entry.Version)
	}
	if entry.Date != "2024-11-09" {
		t.Errorf("expected date '2024-11-09', got '%s'", entry.Date)
	}

	if len(entry.Changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(entry.Changes))
	}

	change1 := entry.Changes[0]
	if change1.Type != "added" {
		t.Errorf("expected type 'added', got '%s'", change1.Type)
	}
	if change1.Severity != "safe" {
		t.Errorf("expected severity 'safe', got '%s'", change1.Severity)
	}

	change2 := entry.Changes[1]
	if change2.Type != "deprecated" {
		t.Errorf("expected type 'deprecated', got '%s'", change2.Type)
	}
	if change2.Migration != "Use GET /v2/users instead" {
		t.Errorf("expected migration message, got '%s'", change2.Migration)
	}
}
