package builtin

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestNewInfoCommand(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name:    "testcli",
			Version: "1.0.0",
		},
		API: cli.API{
			BaseURL:    "https://api.example.com",
			OpenAPIURL: "https://api.example.com/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config: config,
		Output: output,
	}

	cmd := NewInfoCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "info" {
		t.Errorf("expected Use 'info', got %q", cmd.Use)
	}
}

func TestBuildCLIInfo(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name:        "testcli",
			Version:     "1.0.0",
			Description: "Test CLI",
			Homepage:    "https://example.com",
			DocsURL:     "https://example.com/docs",
		},
		API: cli.API{
			Version:    "2.0.0",
			BaseURL:    "https://api.example.com",
			OpenAPIURL: "https://api.example.com/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config: config,
		Output: output,
	}

	info := buildCLIInfo(opts)

	if info.CLI.Name != "testcli" {
		t.Errorf("expected CLI name 'testcli', got %q", info.CLI.Name)
	}

	if info.CLI.Version != "1.0.0" {
		t.Errorf("expected CLI version '1.0.0', got %q", info.CLI.Version)
	}

	if info.API.BaseURL != "https://api.example.com" {
		t.Errorf("expected API base URL 'https://api.example.com', got %q", info.API.BaseURL)
	}

	if info.Config.Auth != "Not configured" {
		t.Errorf("expected auth 'Not configured', got %q", info.Config.Auth)
	}
}

func TestCheckAPIHealth_Success(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
		API: cli.API{
			BaseURL:    server.URL,
			OpenAPIURL: server.URL + "/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config:     config,
		Output:     output,
		HTTPClient: server.Client(),
	}

	status := checkAPIHealth(opts)

	if !status.APIReachable {
		t.Error("expected API to be reachable")
	}

	if status.LastChecked.IsZero() {
		t.Error("expected LastChecked to be set")
	}
}

func TestCheckAPIHealth_Failure(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
		API: cli.API{
			BaseURL:    "https://nonexistent.invalid",
			OpenAPIURL: "https://nonexistent.invalid/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config:     config,
		Output:     output,
		HTTPClient: &http.Client{Timeout: 1 * time.Second},
	}

	status := checkAPIHealth(opts)

	if status.APIReachable {
		t.Error("expected API to be unreachable")
	}
}

func TestCheckAPIHealth_ServerError(t *testing.T) {
	// Create test server that returns 500
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
		API: cli.API{
			BaseURL:    server.URL,
			OpenAPIURL: server.URL + "/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config:     config,
		Output:     output,
		HTTPClient: server.Client(),
	}

	status := checkAPIHealth(opts)

	if status.APIReachable {
		t.Error("expected API to be unreachable with 500 status")
	}
}

func TestCheckSpecCache(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
		API: cli.API{
			OpenAPIURL: "https://api.example.com/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config: config,
		Output: output,
	}

	status := checkSpecCache(opts)

	if status.LastChecked.IsZero() {
		t.Error("expected LastChecked to be set")
	}

	// SpecCached should be false for non-existent cache
	if status.SpecCached {
		t.Error("expected SpecCached to be false")
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
		expected string
	}{
		{"seconds", 30 * time.Second, "30s"},
		{"minutes", 5 * time.Minute, "5m"},
		{"hours", 3 * time.Hour, "3h"},
		{"days", 2 * 24 * time.Hour, "2d"},
		{"almost minute", 59 * time.Second, "59s"},
		{"just over hour", 61 * time.Minute, "1h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.duration)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestFormatInfoText(t *testing.T) {
	info := &CLIInfo{
		CLI: CLIDetails{
			Name:        "testcli",
			Version:     "1.0.0",
			Description: "Test CLI for testing",
		},
		API: APIDetails{
			Title:   "Test API",
			Version: "2.0.0",
			BaseURL: "https://api.example.com",
			SpecURL: "https://api.example.com/openapi.json",
		},
		Config: ConfigDetails{
			ConfigFile: "/home/user/.config/testcli/config.yaml",
			CacheDir:   "/home/user/.cache/testcli",
			DataDir:    "/home/user/.local/share/testcli",
			StateDir:   "/home/user/.local/state/testcli",
			Auth:       "Configured",
		},
		Status: StatusDetails{
			APIReachable: true,
			SpecCached:   true,
			SpecAge:      "2h",
			LastChecked:  time.Now(),
		},
	}

	output := &bytes.Buffer{}
	err := formatInfoText(info, output)
	if err != nil {
		t.Fatalf("formatInfoText failed: %v", err)
	}

	result := output.String()

	// Verify key information is present
	if !strings.Contains(result, "testcli") {
		t.Errorf("expected CLI name in output, got: %s", result)
	}

	if !strings.Contains(result, "1.0.0") {
		t.Errorf("expected CLI version in output, got: %s", result)
	}

	if !strings.Contains(result, "API reachable") {
		t.Errorf("expected API status in output, got: %s", result)
	}

	if !strings.Contains(result, "Spec cached") {
		t.Errorf("expected spec cache status in output, got: %s", result)
	}
}

func TestFormatInfoJSON(t *testing.T) {
	info := &CLIInfo{
		CLI: CLIDetails{
			Name:    "testcli",
			Version: "1.0.0",
		},
		API: APIDetails{
			BaseURL: "https://api.example.com",
		},
		Config: ConfigDetails{
			ConfigFile: "/config/file",
		},
	}

	output := &bytes.Buffer{}
	err := formatInfoJSON(info, output)
	if err != nil {
		t.Fatalf("formatInfoJSON failed: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, `"name": "testcli"`) {
		t.Errorf("expected name in JSON output, got: %s", result)
	}

	if !strings.Contains(result, `"base_url": "https://api.example.com"`) {
		t.Errorf("expected base_url in JSON output, got: %s", result)
	}
}

func TestFormatInfoYAML(t *testing.T) {
	info := &CLIInfo{
		CLI: CLIDetails{
			Name:    "testcli",
			Version: "1.0.0",
		},
		API: APIDetails{
			BaseURL: "https://api.example.com",
		},
		Config: ConfigDetails{
			ConfigFile: "/config/file",
		},
	}

	output := &bytes.Buffer{}
	err := formatInfoYAML(info, output)
	if err != nil {
		t.Fatalf("formatInfoYAML failed: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, "name: testcli") {
		t.Errorf("expected name in YAML output, got: %s", result)
	}

	if !strings.Contains(result, "base_url: https://api.example.com") {
		t.Errorf("expected base_url in YAML output, got: %s", result)
	}
}

func TestRunInfo_WithHealthCheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &cli.Config{
		Metadata: cli.Metadata{
			Name:    "testcli",
			Version: "1.0.0",
		},
		API: cli.API{
			BaseURL:    server.URL,
			OpenAPIURL: server.URL + "/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config:         config,
		CheckAPIHealth: true,
		Output:         output,
		HTTPClient:     server.Client(),
	}

	err := runInfo(opts)
	if err != nil {
		t.Fatalf("runInfo failed: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, "testcli") {
		t.Errorf("expected CLI name in output, got: %s", result)
	}
}

func TestRunInfo_JSONFormat(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name:    "testcli",
			Version: "1.0.0",
		},
		API: cli.API{
			BaseURL:    "https://api.example.com",
			OpenAPIURL: "https://api.example.com/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config:       config,
		OutputFormat: "json",
		Output:       output,
	}

	err := runInfo(opts)
	if err != nil {
		t.Fatalf("runInfo failed: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, `"name"`) {
		t.Errorf("expected JSON output, got: %s", result)
	}
}

func TestRunInfo_YAMLFormat(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name:    "testcli",
			Version: "1.0.0",
		},
		API: cli.API{
			BaseURL:    "https://api.example.com",
			OpenAPIURL: "https://api.example.com/openapi.json",
		},
	}

	output := &bytes.Buffer{}
	opts := &InfoOptions{
		Config:       config,
		OutputFormat: "yaml",
		Output:       output,
	}

	err := runInfo(opts)
	if err != nil {
		t.Fatalf("runInfo failed: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, "name: testcli") {
		t.Errorf("expected YAML output, got: %s", result)
	}
}
