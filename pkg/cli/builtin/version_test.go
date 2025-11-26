package builtin

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestNewVersionCommand(t *testing.T) {
	config := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name:        "testcli",
			Version:     "1.0.0",
			Description: "Test CLI",
		},
		API: cli.API{
			Version: "2.0.0",
		},
	}

	output := &bytes.Buffer{}
	opts := &VersionOptions{
		Config:         config,
		BuildTime:      time.Now(),
		ShowAPIVersion: true,
		OutputFormat:   "text",
		Output:         output,
	}

	cmd := NewVersionCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "version" {
		t.Errorf("expected Use 'version', got %q", cmd.Use)
	}
}

func TestRunVersion_Text(t *testing.T) {
	config := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name:        "testcli",
			Version:     "1.0.0",
			Description: "Test CLI",
		},
		API: cli.API{
			Version: "2.0.0",
		},
	}

	buildTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	output := &bytes.Buffer{}
	opts := &VersionOptions{
		Config:         config,
		BuildTime:      buildTime,
		ShowAPIVersion: true,
		OutputFormat:   "text",
		Output:         output,
	}

	err := runVersion(opts)
	if err != nil {
		t.Fatalf("runVersion failed: %v", err)
	}

	result := output.String()

	// Check that output contains expected strings
	if !strings.Contains(result, "Client Version: 1.0.0") {
		t.Errorf("expected client version in output, got: %s", result)
	}

	if !strings.Contains(result, "Server Version: 2.0.0") {
		t.Errorf("expected server version in output, got: %s", result)
	}

	if !strings.Contains(result, "Built:") {
		t.Errorf("expected build time in output, got: %s", result)
	}
}

func TestRunVersion_JSON(t *testing.T) {
	config := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name:    "testcli",
			Version: "1.0.0",
		},
		API: cli.API{
			Version: "2.0.0",
		},
	}

	output := &bytes.Buffer{}
	opts := &VersionOptions{
		Config:         config,
		BuildTime:      time.Now(),
		ShowAPIVersion: true,
		OutputFormat:   "json",
		Output:         output,
	}

	err := runVersion(opts)
	if err != nil {
		t.Fatalf("runVersion failed: %v", err)
	}

	result := output.String()

	// Check JSON output
	if !strings.Contains(result, `"client_version"`) {
		t.Errorf("expected client_version in JSON output, got: %s", result)
	}

	if !strings.Contains(result, `"server_version"`) {
		t.Errorf("expected server_version in JSON output, got: %s", result)
	}
}

func TestRunVersion_YAML(t *testing.T) {
	config := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name:        "testcli",
			Version:     "1.0.0",
			Description: "Test CLI",
		},
		API: cli.API{
			Version: "2.0.0",
		},
	}

	buildTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	output := &bytes.Buffer{}
	opts := &VersionOptions{
		Config:         config,
		BuildTime:      buildTime,
		ShowAPIVersion: true,
		OutputFormat:   "yaml",
		Output:         output,
	}

	err := runVersion(opts)
	if err != nil {
		t.Fatalf("runVersion failed: %v", err)
	}

	result := output.String()

	// Check YAML output
	if !strings.Contains(result, "client_version: 1.0.0") {
		t.Errorf("expected client_version in YAML output, got: %s", result)
	}

	if !strings.Contains(result, "server_version: 2.0.0") {
		t.Errorf("expected server_version in YAML output, got: %s", result)
	}

	if !strings.Contains(result, "api_title: Test CLI") {
		t.Errorf("expected api_title in YAML output, got: %s", result)
	}

	if !strings.Contains(result, "built:") {
		t.Errorf("expected built in YAML output, got: %s", result)
	}
}

func TestGetVersionShort(t *testing.T) {
	tests := []struct {
		name           string
		clientVersion  string
		serverVersion  string
		expectedOutput string
	}{
		{
			name:           "both versions",
			clientVersion:  "1.0.0",
			serverVersion:  "2.0.0",
			expectedOutput: "Client: 1.0.0  Server: 2.0.0",
		},
		{
			name:           "client only",
			clientVersion:  "1.0.0",
			serverVersion:  "",
			expectedOutput: "v1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &cli.CLIConfig{
				Metadata: cli.Metadata{
					Version: tt.clientVersion,
				},
				API: cli.API{
					Version: tt.serverVersion,
				},
			}

			result := GetVersionShort(config)
			if result != tt.expectedOutput {
				t.Errorf("expected %q, got %q", tt.expectedOutput, result)
			}
		})
	}
}
