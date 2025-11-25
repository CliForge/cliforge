package openapi

import (
	"context"
	"strings"
	"testing"
)

func TestNewChangeDetector(t *testing.T) {
	detector := NewChangeDetector()

	if detector == nil {
		t.Fatal("NewChangeDetector returned nil")
	}
	if !detector.IncludeNonBreaking {
		t.Error("expected IncludeNonBreaking to be true by default")
	}
}

func TestChangeDetector_DetectChanges_VersionChange(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "2.0.0"},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	if len(changes) == 0 {
		t.Error("expected version change to be detected")
	}

	foundVersionChange := false
	for _, change := range changes {
		if change.Path == "info.version" {
			foundVersionChange = true
			if change.Type != ChangeTypeModified {
				t.Errorf("expected change type 'modified', got '%s'", change.Type)
			}
			if change.Severity != ChangeSeveritySafe {
				t.Errorf("expected severity 'safe', got '%s'", change.Severity)
			}
		}
	}
	if !foundVersionChange {
		t.Error("version change not found in changes")
	}
}

func TestChangeDetector_DetectChanges_PathAdded(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundPathAdded := false
	for _, change := range changes {
		if change.Path == "/users" && change.Type == ChangeTypeAdded {
			foundPathAdded = true
			if change.Severity != ChangeSeveritySafe {
				t.Errorf("expected severity 'safe' for added path, got '%s'", change.Severity)
			}
		}
	}
	if !foundPathAdded {
		t.Error("path addition not detected")
	}
}

func TestChangeDetector_DetectChanges_PathRemoved(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundPathRemoved := false
	for _, change := range changes {
		if change.Path == "/users" && change.Type == ChangeTypeRemoved {
			foundPathRemoved = true
			if change.Severity != ChangeSeverityBreaking {
				t.Errorf("expected severity 'breaking' for removed path, got '%s'", change.Severity)
			}
		}
	}
	if !foundPathRemoved {
		t.Error("path removal not detected")
	}
}

func TestChangeDetector_DetectChanges_OperationAdded(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"summary": "List users",
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"summary": "List users",
					"responses": {"200": {"description": "OK"}}
				},
				"post": {
					"operationId": "createUser",
					"summary": "Create user",
					"responses": {"201": {"description": "Created"}}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundOperationAdded := false
	for _, change := range changes {
		if strings.Contains(change.Path, "POST /users") && change.Type == ChangeTypeAdded {
			foundOperationAdded = true
		}
	}
	if !foundOperationAdded {
		t.Error("operation addition not detected")
	}
}

func TestChangeDetector_DetectChanges_ParameterRemoved(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"parameters": [
						{
							"name": "limit",
							"in": "query",
							"schema": {"type": "integer"}
						}
					],
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"parameters": [],
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundParameterRemoved := false
	for _, change := range changes {
		if strings.Contains(change.Path, "parameters.limit") && change.Type == ChangeTypeRemoved {
			foundParameterRemoved = true
			if change.Severity != ChangeSeverityBreaking {
				t.Errorf("expected severity 'breaking' for removed parameter, got '%s'", change.Severity)
			}
		}
	}
	if !foundParameterRemoved {
		t.Error("parameter removal not detected")
	}
}

func TestChangeDetector_DetectChanges_RequiredParameterAdded(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"parameters": [],
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"parameters": [
						{
							"name": "filter",
							"in": "query",
							"required": true,
							"schema": {"type": "string"}
						}
					],
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundRequiredParam := false
	for _, change := range changes {
		if strings.Contains(change.Path, "parameters.filter") && change.Type == ChangeTypeAdded {
			foundRequiredParam = true
			if change.Severity != ChangeSeverityBreaking {
				t.Errorf("expected severity 'breaking' for required parameter, got '%s'", change.Severity)
			}
		}
	}
	if !foundRequiredParam {
		t.Error("required parameter addition not detected")
	}
}

func TestChangeDetector_DetectChanges_OperationDeprecated(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"deprecated": false,
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "listUsers",
					"deprecated": true,
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundDeprecation := false
	for _, change := range changes {
		if strings.Contains(change.Description, "deprecated") {
			foundDeprecation = true
			if change.Severity != ChangeSeverityDangerous {
				t.Errorf("expected severity 'dangerous' for deprecation, got '%s'", change.Severity)
			}
		}
	}
	if !foundDeprecation {
		t.Error("deprecation not detected")
	}
}

func TestChangeDetector_DetectChanges_SchemaAdded(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"components": {
			"schemas": {}
		},
		"paths": {}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"components": {
			"schemas": {
				"User": {
					"type": "object",
					"properties": {
						"id": {"type": "string"}
					}
				}
			}
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundSchemaAdded := false
	for _, change := range changes {
		if strings.Contains(change.Path, "schemas.User") && change.Type == ChangeTypeAdded {
			foundSchemaAdded = true
		}
	}
	if !foundSchemaAdded {
		t.Error("schema addition not detected")
	}
}

func TestChangeDetector_DetectChanges_SecurityAdded(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"security": [
			{"apiKey": []}
		],
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	detector := NewChangeDetector()
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	foundSecurityAdded := false
	for _, change := range changes {
		if change.Path == "security" && change.Type == ChangeTypeAdded {
			foundSecurityAdded = true
			if change.Severity != ChangeSeverityBreaking {
				t.Errorf("expected severity 'breaking' for security added, got '%s'", change.Severity)
			}
		}
	}
	if !foundSecurityAdded {
		t.Error("security addition not detected")
	}
}

func TestChangeDetector_IncludeNonBreaking(t *testing.T) {
	oldSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {}
	}`

	newSpec := `{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "2.0.0"},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()

	oldParsed, _ := parser.Parse(ctx, []byte(oldSpec))
	newParsed, _ := parser.Parse(ctx, []byte(newSpec))

	// Exclude non-breaking changes
	detector := NewChangeDetector()
	detector.IncludeNonBreaking = false
	changes, err := detector.DetectChanges(oldParsed, newParsed)
	if err != nil {
		t.Fatalf("DetectChanges failed: %v", err)
	}

	// Version change is safe, should be filtered out
	if len(changes) > 0 {
		t.Errorf("expected 0 changes with IncludeNonBreaking=false, got %d", len(changes))
	}
}

func TestIsBreaking(t *testing.T) {
	changes := []*DetectedChange{
		{Type: ChangeTypeAdded, Severity: ChangeSeveritySafe},
		{Type: ChangeTypeModified, Severity: ChangeSeveritySafe},
	}

	if IsBreaking(changes) {
		t.Error("expected IsBreaking to return false for safe changes")
	}

	changes = append(changes, &DetectedChange{
		Type:     ChangeTypeRemoved,
		Severity: ChangeSeverityBreaking,
	})

	if !IsBreaking(changes) {
		t.Error("expected IsBreaking to return true for breaking changes")
	}
}

func TestGroupByType(t *testing.T) {
	changes := []*DetectedChange{
		{Type: ChangeTypeAdded, Path: "/path1"},
		{Type: ChangeTypeAdded, Path: "/path2"},
		{Type: ChangeTypeRemoved, Path: "/path3"},
		{Type: ChangeTypeModified, Path: "/path4"},
	}

	groups := GroupByType(changes)

	if len(groups[ChangeTypeAdded]) != 2 {
		t.Errorf("expected 2 added changes, got %d", len(groups[ChangeTypeAdded]))
	}
	if len(groups[ChangeTypeRemoved]) != 1 {
		t.Errorf("expected 1 removed change, got %d", len(groups[ChangeTypeRemoved]))
	}
	if len(groups[ChangeTypeModified]) != 1 {
		t.Errorf("expected 1 modified change, got %d", len(groups[ChangeTypeModified]))
	}
}

func TestGroupBySeverity(t *testing.T) {
	changes := []*DetectedChange{
		{Severity: ChangeSeveritySafe, Path: "/path1"},
		{Severity: ChangeSeveritySafe, Path: "/path2"},
		{Severity: ChangeSeverityBreaking, Path: "/path3"},
		{Severity: ChangeSeverityDangerous, Path: "/path4"},
	}

	groups := GroupBySeverity(changes)

	if len(groups[ChangeSeveritySafe]) != 2 {
		t.Errorf("expected 2 safe changes, got %d", len(groups[ChangeSeveritySafe]))
	}
	if len(groups[ChangeSeverityBreaking]) != 1 {
		t.Errorf("expected 1 breaking change, got %d", len(groups[ChangeSeverityBreaking]))
	}
	if len(groups[ChangeSeverityDangerous]) != 1 {
		t.Errorf("expected 1 dangerous change, got %d", len(groups[ChangeSeverityDangerous]))
	}
}

func TestFormatChangelog(t *testing.T) {
	changes := []*DetectedChange{
		{
			Type:        ChangeTypeRemoved,
			Severity:    ChangeSeverityBreaking,
			Path:        "/users",
			Description: "Path removed",
		},
		{
			Type:        ChangeTypeModified,
			Severity:    ChangeSeverityDangerous,
			Path:        "/posts",
			Description: "Operation deprecated",
		},
		{
			Type:        ChangeTypeAdded,
			Severity:    ChangeSeveritySafe,
			Path:        "/comments",
			Description: "New endpoint added",
		},
	}

	formatted := FormatChangelog(changes)

	if !strings.Contains(formatted, "BREAKING CHANGES") {
		t.Error("expected formatted changelog to contain 'BREAKING CHANGES'")
	}
	if !strings.Contains(formatted, "DEPRECATIONS & WARNINGS") {
		t.Error("expected formatted changelog to contain 'DEPRECATIONS & WARNINGS'")
	}
	if !strings.Contains(formatted, "NEW FEATURES & IMPROVEMENTS") {
		t.Error("expected formatted changelog to contain 'NEW FEATURES & IMPROVEMENTS'")
	}
	if !strings.Contains(formatted, "Path removed") {
		t.Error("expected formatted changelog to contain change description")
	}
}

func TestFormatChangelog_Empty(t *testing.T) {
	formatted := FormatChangelog([]*DetectedChange{})

	if formatted != "No changes detected" {
		t.Errorf("expected 'No changes detected', got '%s'", formatted)
	}
}

func TestGetChangelog(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test",
			"version": "1.0.0",
			"x-cli-changelog": [
				{
					"version": "2.0.0",
					"date": "2024-02-01",
					"changes": []
				},
				{
					"version": "1.5.0",
					"date": "2024-01-15",
					"changes": []
				},
				{
					"version": "1.0.0",
					"date": "2024-01-01",
					"changes": []
				}
			]
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, _ := parser.Parse(ctx, []byte(spec))

	entries := GetChangelog(parsed)

	if len(entries) != 3 {
		t.Errorf("expected 3 changelog entries, got %d", len(entries))
	}

	// Should be sorted in descending order
	if entries[0].Version != "2.0.0" {
		t.Errorf("expected first entry to be version 2.0.0, got %s", entries[0].Version)
	}
	if entries[2].Version != "1.0.0" {
		t.Errorf("expected last entry to be version 1.0.0, got %s", entries[2].Version)
	}
}

func TestGetLatestChanges(t *testing.T) {
	spec := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test",
			"version": "1.0.0",
			"x-cli-changelog": [
				{
					"version": "2.0.0",
					"date": "2024-02-01",
					"changes": []
				},
				{
					"version": "1.5.0",
					"date": "2024-01-15",
					"changes": []
				},
				{
					"version": "1.0.0",
					"date": "2024-01-01",
					"changes": []
				}
			]
		},
		"paths": {}
	}`

	parser := NewParser()
	ctx := context.Background()
	parsed, _ := parser.Parse(ctx, []byte(spec))

	latestChanges := GetLatestChanges(parsed, "1.0.0")

	if len(latestChanges) != 2 {
		t.Errorf("expected 2 changes since 1.0.0, got %d", len(latestChanges))
	}

	if latestChanges[0].Version != "2.0.0" {
		t.Errorf("expected first change to be 2.0.0, got %s", latestChanges[0].Version)
	}
}
