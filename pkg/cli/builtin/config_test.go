package builtin

import (
	"bytes"
	"strings"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestConfigGetValue(t *testing.T) {
	config := &cli.UserConfig{
		Preferences: &cli.UserPreferences{
			Output: &cli.PreferencesOutput{
				Format: "json",
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		expected string
		wantErr  bool
	}{
		{
			name:     "nested value",
			key:      "preferences.output.format",
			expected: "json",
			wantErr:  false,
		},
		{
			name:    "non-existent key",
			key:     "preferences.nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getConfigValue(config, tt.key)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestConfigSetValue(t *testing.T) {
	config := &cli.UserConfig{
		Preferences: &cli.UserPreferences{},
	}

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "set nested value",
			key:   "preferences.output.format",
			value: "yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setConfigValue(config, tt.key, tt.value)
			if err != nil {
				t.Fatalf("setConfigValue failed: %v", err)
			}

			// Verify value was set
			result, err := getConfigValue(config, tt.key)
			if err != nil {
				t.Fatalf("getConfigValue failed: %v", err)
			}

			if !strings.Contains(result, tt.value) {
				t.Errorf("expected value containing %q, got %q", tt.value, result)
			}
		})
	}
}

func TestConfigUnsetValue(t *testing.T) {
	config := &cli.UserConfig{
		Preferences: &cli.UserPreferences{
			Output: &cli.PreferencesOutput{
				Format: "json",
			},
		},
	}

	err := unsetConfigValue(config, "preferences.output.format")
	if err != nil {
		t.Fatalf("unsetConfigValue failed: %v", err)
	}

	// Verify value was unset
	_, err = getConfigValue(config, "preferences.output.format")
	if err == nil {
		t.Error("expected error after unsetting value, got nil")
	}
}

func TestListConfigKeys(t *testing.T) {
	config := &cli.UserConfig{
		Preferences: &cli.UserPreferences{
			Output: &cli.PreferencesOutput{
				Format: "json",
				Color:  "always",
			},
		},
	}

	keys := ListConfigKeys(config)

	if len(keys) == 0 {
		t.Error("expected config keys, got empty list")
	}

	// Check for expected keys
	hasOutputFormat := false
	for _, key := range keys {
		if strings.Contains(key, "output.format") {
			hasOutputFormat = true
			break
		}
	}

	if !hasOutputFormat {
		t.Error("expected to find 'output.format' key")
	}
}

func TestNewConfigCommand(t *testing.T) {
	config := &cli.UserConfig{}
	output := &bytes.Buffer{}

	opts := &ConfigOptions{
		CLIName:    "testcli",
		AllowEdit:  true,
		Output:     output,
		UserConfig: config,
		SaveFunc: func(c *cli.UserConfig) error {
			return nil
		},
	}

	cmd := NewConfigCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "config" {
		t.Errorf("expected Use 'config', got %q", cmd.Use)
	}

	// Check subcommands exist
	expectedSubcommands := []string{"show", "get", "set", "unset", "path", "edit"}
	for _, expected := range expectedSubcommands {
		found := false
		for _, subCmd := range cmd.Commands() {
			if subCmd.Name() == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", expected)
		}
	}
}
