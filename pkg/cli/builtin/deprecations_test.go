package builtin

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

func TestNewDeprecationsCommand(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
	}

	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		Config:                 config,
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{}, nil
		},
	}

	cmd := NewDeprecationsCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "deprecations" {
		t.Errorf("expected Use 'deprecations', got %q", cmd.Use)
	}
}

func TestNewDeprecationsCommand_WithScan(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		AllowScan:              true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{}, nil
		},
	}

	cmd := NewDeprecationsCommand(opts)

	// Check that scan subcommand is added
	scanCmd := findSubcommand(cmd, "scan")
	if scanCmd == nil {
		t.Error("expected scan subcommand when AllowScan is true")
	}
}

func findSubcommand(cmd *cobra.Command, name string) *cobra.Command {
	for _, sub := range cmd.Commands() {
		if sub.Use == name || sub.Use == name+" [path]" {
			return sub
		}
	}
	return nil
}

func TestRunDeprecations_NoBinaryNoAPI(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{}, nil
		},
	}

	err := runDeprecations(opts, false, false, "text")
	if err != nil {
		t.Fatalf("runDeprecations failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "No deprecated CLI features") {
		t.Error("expected message about no deprecated CLI features")
	}
	if !strings.Contains(result, "No deprecated API endpoints") {
		t.Error("expected message about no deprecated API endpoints")
	}
}

func TestRunDeprecations_WithBinaryDeprecations(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    false,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{
				{
					Feature:     "--old-flag",
					Replacement: "--new-flag",
					Message:     "Use --new-flag instead",
					Severity:    "warning",
					Sunset:      time.Now().Add(30 * 24 * time.Hour),
				},
			}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{}, nil
		},
	}

	err := runDeprecations(opts, false, false, "text")
	if err != nil {
		t.Fatalf("runDeprecations failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "--old-flag") {
		t.Error("expected old flag in output")
	}
	if !strings.Contains(result, "--new-flag") {
		t.Error("expected replacement flag in output")
	}
}

func TestRunDeprecations_WithAPIDeprecations(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: false,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{
				{
					OperationID: "getOldEndpoint",
					Method:      "GET",
					Path:        "/api/v1/old",
					Message:     "Use /api/v2/new instead",
					Sunset:      time.Now().Add(60 * 24 * time.Hour),
				},
			}, nil
		},
	}

	err := runDeprecations(opts, false, false, "text")
	if err != nil {
		t.Fatalf("runDeprecations failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "GET /api/v1/old") {
		t.Error("expected deprecated endpoint in output")
	}
	if !strings.Contains(result, "getOldEndpoint") {
		t.Error("expected operation ID in output")
	}
}

func TestRunDeprecations_BinaryOnly(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{
				{
					Feature:  "--old-flag",
					Severity: "warning",
					Message:  "Deprecated",
				},
			}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{
				{
					OperationID: "getOld",
					Method:      "GET",
					Path:        "/old",
				},
			}, nil
		},
	}

	err := runDeprecations(opts, true, false, "text")
	if err != nil {
		t.Fatalf("runDeprecations failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "--old-flag") {
		t.Error("expected binary deprecation in output")
	}
	if strings.Contains(result, "getOld") {
		t.Error("did not expect API deprecation in output with --binary-only")
	}
}

func TestRunDeprecations_APIOnly(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{
				{
					Feature:  "--old-flag",
					Severity: "warning",
					Message:  "Deprecated",
				},
			}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{
				{
					OperationID: "getOld",
					Method:      "GET",
					Path:        "/old",
				},
			}, nil
		},
	}

	err := runDeprecations(opts, false, true, "text")
	if err != nil {
		t.Fatalf("runDeprecations failed: %v", err)
	}

	result := output.String()
	if strings.Contains(result, "--old-flag") {
		t.Error("did not expect binary deprecation in output with --api-only")
	}
	if !strings.Contains(result, "getOld") {
		t.Error("expected API deprecation in output")
	}
}

func TestRunDeprecations_JSON(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{}, nil
		},
	}

	err := runDeprecations(opts, false, false, "json")
	if err != nil {
		t.Fatalf("runDeprecations failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"binary_deprecations"`) {
		t.Error("expected binary_deprecations in JSON output")
	}
	if !strings.Contains(result, `"api_deprecations"`) {
		t.Error("expected api_deprecations in JSON output")
	}
}

func TestRunDeprecations_YAML(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return []BinaryDeprecation{
				{
					Feature:  "old-cmd",
					Severity: "warning",
					Message:  "Use new-cmd",
				},
			}, nil
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{
				{
					OperationID: "oldOp",
					Method:      "GET",
					Path:        "/old",
				},
			}, nil
		},
	}

	err := runDeprecations(opts, false, false, "yaml")
	if err != nil {
		t.Fatalf("runDeprecations failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "binary_deprecations:") {
		t.Error("expected binary_deprecations in YAML output")
	}
	if !strings.Contains(result, "api_deprecations:") {
		t.Error("expected api_deprecations in YAML output")
	}
}

func TestRunDeprecations_ErrorHandling(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		ShowBinaryDeprecations: true,
		ShowAPIDeprecations:    true,
		Output:                 output,
		BinaryDeprecationsFunc: func() ([]BinaryDeprecation, error) {
			return nil, errors.New("binary error")
		},
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return nil, errors.New("api error")
		},
	}

	err := runDeprecations(opts, false, false, "text")
	// Should not error, just warn
	if err != nil {
		t.Fatalf("runDeprecations should handle errors gracefully: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Warning: failed to load") {
		t.Error("expected warning about failed loading")
	}
}

func TestRunDeprecationsShow(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		Output: output,
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{
				{
					OperationID: "getUsers",
					Method:      "GET",
					Path:        "/api/users",
					Message:     "Use /v2/users instead",
					Sunset:      time.Now().Add(30 * 24 * time.Hour),
				},
			}, nil
		},
	}

	err := runDeprecationsShow(opts, "getUsers")
	if err != nil {
		t.Fatalf("runDeprecationsShow failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "getUsers") {
		t.Error("expected operation ID in output")
	}
	if !strings.Contains(result, "GET /api/users") {
		t.Error("expected operation details in output")
	}
}

func TestRunDeprecationsShow_NotFound(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		Output: output,
		APIDeprecationsFunc: func() ([]openapi.Deprecation, error) {
			return []openapi.Deprecation{}, nil
		},
	}

	err := runDeprecationsShow(opts, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent deprecation")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestRunDeprecationsScan(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &DeprecationsOptions{
		Output: output,
	}

	err := runDeprecationsScan(opts, "/path/to/code")
	if err != nil {
		t.Fatalf("runDeprecationsScan failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Scanning /path/to/code") {
		t.Error("expected scanning message")
	}
	if !strings.Contains(result, "not yet implemented") {
		t.Error("expected not implemented message")
	}
}

func TestPrintBinaryDeprecation_Warning(t *testing.T) {
	dep := &BinaryDeprecation{
		Feature:     "--verbose",
		Replacement: "-v",
		Message:     "Use -v instead",
		Severity:    "warning",
		Sunset:      time.Now().Add(60 * 24 * time.Hour),
	}

	output := &bytes.Buffer{}
	printBinaryDeprecation(dep, output)

	result := output.String()
	if !strings.Contains(result, "WARNING") {
		t.Error("expected WARNING in output")
	}
	if !strings.Contains(result, "--verbose") {
		t.Error("expected feature name in output")
	}
	if !strings.Contains(result, "-v") {
		t.Error("expected replacement in output")
	}
}

func TestPrintBinaryDeprecation_Critical(t *testing.T) {
	dep := &BinaryDeprecation{
		Feature:  "--dangerous",
		Message:  "Removed in next version",
		Severity: "critical",
		Sunset:   time.Now().Add(10 * 24 * time.Hour),
	}

	output := &bytes.Buffer{}
	printBinaryDeprecation(dep, output)

	result := output.String()
	if !strings.Contains(result, "CRITICAL") {
		t.Error("expected CRITICAL in output")
	}
}

func TestPrintAPIDeprecation(t *testing.T) {
	dep := &openapi.Deprecation{
		OperationID: "listItems",
		Method:      "GET",
		Path:        "/api/items",
		Replacement: "/api/v2/items",
		Message:     "Use v2 endpoint",
		Sunset:      time.Now().Add(45 * 24 * time.Hour),
	}

	output := &bytes.Buffer{}
	printAPIDeprecation(dep, output)

	result := output.String()
	if !strings.Contains(result, "GET /api/items") {
		t.Error("expected endpoint in output")
	}
	if !strings.Contains(result, "listItems") {
		t.Error("expected operation ID in output")
	}
	if !strings.Contains(result, "/api/v2/items") {
		t.Error("expected replacement in output")
	}
}

func TestFormatDeprecationDetail(t *testing.T) {
	dep := &openapi.Deprecation{
		OperationID: "getOldData",
		Method:      "POST",
		Path:        "/api/old/data",
		Message:     "This endpoint is deprecated",
		Replacement: "/api/new/data",
		DocsURL:     "https://docs.example.com/migration",
		Sunset:      time.Now().Add(30 * 24 * time.Hour),
	}

	output := &bytes.Buffer{}
	err := formatDeprecationDetail(dep, output)
	if err != nil {
		t.Fatalf("formatDeprecationDetail failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Deprecation Details") {
		t.Error("expected header")
	}
	if !strings.Contains(result, "POST /api/old/data") {
		t.Error("expected operation")
	}
	if !strings.Contains(result, "getOldData") {
		t.Error("expected operation ID")
	}
	if !strings.Contains(result, "This endpoint is deprecated") {
		t.Error("expected message")
	}
	if !strings.Contains(result, "/api/new/data") {
		t.Error("expected replacement")
	}
	if !strings.Contains(result, "https://docs.example.com/migration") {
		t.Error("expected docs URL")
	}
}

func TestFormatDeprecationsJSON(t *testing.T) {
	binaryDeps := []BinaryDeprecation{
		{
			Feature:  "--old",
			Severity: "warning",
			Message:  "Deprecated",
		},
	}
	apiDeps := []openapi.Deprecation{
		{
			OperationID: "oldOp",
			Method:      "GET",
			Path:        "/old",
		},
	}

	output := &bytes.Buffer{}
	err := formatDeprecationsJSON(binaryDeps, apiDeps, output)
	if err != nil {
		t.Fatalf("formatDeprecationsJSON failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"binary_deprecations"`) {
		t.Error("expected binary_deprecations in JSON")
	}
	if !strings.Contains(result, `"api_deprecations"`) {
		t.Error("expected api_deprecations in JSON")
	}
}

func TestFormatDeprecationsYAML(t *testing.T) {
	binaryDeps := []BinaryDeprecation{
		{
			Feature:     "--old",
			Replacement: "--new",
			Severity:    "warning",
			Message:     "Use --new",
			Sunset:      time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}
	apiDeps := []openapi.Deprecation{
		{
			OperationID: "oldOp",
			Method:      "GET",
			Path:        "/old",
			Replacement: "/new",
			Sunset:      time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
		},
	}

	output := &bytes.Buffer{}
	err := formatDeprecationsYAML(binaryDeps, apiDeps, output)
	if err != nil {
		t.Fatalf("formatDeprecationsYAML failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "binary_deprecations:") {
		t.Error("expected binary_deprecations in YAML")
	}
	if !strings.Contains(result, "api_deprecations:") {
		t.Error("expected api_deprecations in YAML")
	}
	if !strings.Contains(result, "feature: --old") {
		t.Error("expected feature in YAML")
	}
	if !strings.Contains(result, "operation_id: oldOp") {
		t.Error("expected operation_id in YAML")
	}
}

func TestDefaultBinaryDeprecationsFunc(t *testing.T) {
	fn := DefaultBinaryDeprecationsFunc()
	deps, err := fn()

	if err != nil {
		t.Fatalf("DefaultBinaryDeprecationsFunc failed: %v", err)
	}

	if len(deps) != 0 {
		t.Error("expected empty deprecations list")
	}
}

func TestDefaultAPIDeprecationsFunc(t *testing.T) {
	fn := DefaultAPIDeprecationsFunc()
	deps, err := fn()

	if err != nil {
		t.Fatalf("DefaultAPIDeprecationsFunc failed: %v", err)
	}

	if len(deps) != 0 {
		t.Error("expected empty deprecations list")
	}
}
