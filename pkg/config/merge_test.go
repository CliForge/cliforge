package config

import (
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestMergeDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   *cli.CLIConfig
		expected *cli.CLIConfig
	}{
		{
			name:   "empty config gets defaults",
			config: &cli.CLIConfig{},
			expected: &cli.CLIConfig{
				Defaults: &cli.Defaults{
					HTTP: &cli.DefaultsHTTP{
						Timeout: "30s",
					},
					Caching: &cli.DefaultsCaching{
						Enabled: true,
					},
					Pagination: &cli.DefaultsPagination{
						Limit: 20,
					},
					Output: &cli.DefaultsOutput{
						Format:      "json",
						PrettyPrint: true,
						Color:       "auto",
						Paging:      true,
					},
					Deprecations: &cli.DefaultsDeprecations{
						AlwaysShow:  false,
						MinSeverity: "info",
					},
					Retry: &cli.DefaultsRetry{
						MaxAttempts: 3,
					},
				},
			},
		},
		{
			name: "partial config gets remaining defaults",
			config: &cli.CLIConfig{
				Defaults: &cli.Defaults{
					HTTP: &cli.DefaultsHTTP{
						Timeout: "60s",
					},
				},
			},
			expected: &cli.CLIConfig{
				Defaults: &cli.Defaults{
					HTTP: &cli.DefaultsHTTP{
						Timeout: "60s",
					},
					Caching: &cli.DefaultsCaching{
						Enabled: true,
					},
					Pagination: &cli.DefaultsPagination{
						Limit: 20,
					},
					Output: &cli.DefaultsOutput{
						Format:      "json",
						PrettyPrint: true,
						Color:       "auto",
						Paging:      true,
					},
					Deprecations: &cli.DefaultsDeprecations{
						AlwaysShow:  false,
						MinSeverity: "info",
					},
					Retry: &cli.DefaultsRetry{
						MaxAttempts: 3,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergeDefaults(tt.config)
			if err != nil {
				t.Fatalf("MergeDefaults failed: %v", err)
			}

			// Check HTTP timeout
			if tt.config.Defaults.HTTP.Timeout != tt.expected.Defaults.HTTP.Timeout {
				t.Errorf("HTTP.Timeout = %v, want %v", tt.config.Defaults.HTTP.Timeout, tt.expected.Defaults.HTTP.Timeout)
			}

			// Check caching enabled
			if tt.config.Defaults.Caching.Enabled != tt.expected.Defaults.Caching.Enabled {
				t.Errorf("Caching.Enabled = %v, want %v", tt.config.Defaults.Caching.Enabled, tt.expected.Defaults.Caching.Enabled)
			}

			// Check pagination limit
			if tt.config.Defaults.Pagination.Limit != tt.expected.Defaults.Pagination.Limit {
				t.Errorf("Pagination.Limit = %v, want %v", tt.config.Defaults.Pagination.Limit, tt.expected.Defaults.Pagination.Limit)
			}

			// Check output format
			if tt.config.Defaults.Output.Format != tt.expected.Defaults.Output.Format {
				t.Errorf("Output.Format = %v, want %v", tt.config.Defaults.Output.Format, tt.expected.Defaults.Output.Format)
			}
		})
	}
}

func TestApplyUserPreferences(t *testing.T) {
	loader := &Loader{}

	tests := []struct {
		name        string
		config      *cli.CLIConfig
		preferences *cli.UserPreferences
		expected    map[string]interface{}
	}{
		{
			name: "apply HTTP timeout preference",
			config: &cli.CLIConfig{
				Defaults: &cli.Defaults{
					HTTP: &cli.DefaultsHTTP{
						Timeout: "30s",
					},
				},
			},
			preferences: &cli.UserPreferences{
				HTTP: &cli.PreferencesHTTP{
					Timeout: "60s",
				},
			},
			expected: map[string]interface{}{
				"http.timeout": "60s",
			},
		},
		{
			name: "apply output format preference",
			config: &cli.CLIConfig{
				Defaults: &cli.Defaults{
					Output: &cli.DefaultsOutput{
						Format: "json",
					},
				},
			},
			preferences: &cli.UserPreferences{
				Output: &cli.PreferencesOutput{
					Format: "yaml",
				},
			},
			expected: map[string]interface{}{
				"output.format": "yaml",
			},
		},
		{
			name: "apply pagination limit within max_limit",
			config: &cli.CLIConfig{
				Defaults: &cli.Defaults{
					Pagination: &cli.DefaultsPagination{
						Limit: 20,
					},
				},
				Behaviors: &cli.Behaviors{
					Pagination: &cli.PaginationBehavior{
						MaxLimit: 100,
					},
				},
			},
			preferences: &cli.UserPreferences{
				Pagination: &cli.PreferencesPagination{
					Limit: 50,
				},
			},
			expected: map[string]interface{}{
				"pagination.limit": 50,
			},
		},
		{
			name: "reject pagination limit exceeding max_limit",
			config: &cli.CLIConfig{
				Defaults: &cli.Defaults{
					Pagination: &cli.DefaultsPagination{
						Limit: 20,
					},
				},
				Behaviors: &cli.Behaviors{
					Pagination: &cli.PaginationBehavior{
						MaxLimit: 100,
					},
				},
			},
			preferences: &cli.UserPreferences{
				Pagination: &cli.PreferencesPagination{
					Limit: 200, // Exceeds max_limit
				},
			},
			expected: map[string]interface{}{
				"pagination.limit": 20, // Should stay at original
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.applyUserPreferences(tt.config, tt.preferences)

			// Check expected values
			for key, expected := range tt.expected {
				switch key {
				case "http.timeout":
					if result.Defaults.HTTP.Timeout != expected {
						t.Errorf("HTTP.Timeout = %v, want %v", result.Defaults.HTTP.Timeout, expected)
					}
				case "output.format":
					if result.Defaults.Output.Format != expected {
						t.Errorf("Output.Format = %v, want %v", result.Defaults.Output.Format, expected)
					}
				case "pagination.limit":
					if result.Defaults.Pagination.Limit != expected {
						t.Errorf("Pagination.Limit = %v, want %v", result.Defaults.Pagination.Limit, expected)
					}
				}
			}
		})
	}
}

func TestApplyDebugOverrides(t *testing.T) {
	loader := &Loader{}

	tests := []struct {
		name            string
		config          *cli.CLIConfig
		override        *cli.CLIConfig
		expectedChanges map[string]interface{}
	}{
		{
			name: "override API base_url",
			config: &cli.CLIConfig{
				API: cli.API{
					BaseURL: "https://api.production.com",
				},
			},
			override: &cli.CLIConfig{
				API: cli.API{
					BaseURL: "http://localhost:8080",
				},
			},
			expectedChanges: map[string]interface{}{
				"api.base_url": "http://localhost:8080",
			},
		},
		{
			name: "override auth type",
			config: &cli.CLIConfig{
				Behaviors: &cli.Behaviors{
					Auth: &cli.AuthBehavior{
						Type: "oauth2",
					},
				},
			},
			override: &cli.CLIConfig{
				Behaviors: &cli.Behaviors{
					Auth: &cli.AuthBehavior{
						Type: "none",
					},
				},
			},
			expectedChanges: map[string]interface{}{
				"behaviors.auth.type": "none",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, overrides := loader.applyDebugOverrides(tt.config, tt.override)

			// Check that overrides map contains expected changes
			for key, expected := range tt.expectedChanges {
				if overrides[key] != expected {
					t.Errorf("override[%s] = %v, want %v", key, overrides[key], expected)
				}
			}

			// Check that config was actually modified
			if tt.expectedChanges["api.base_url"] != nil && result.API.BaseURL != tt.expectedChanges["api.base_url"] {
				t.Errorf("API.BaseURL = %v, want %v", result.API.BaseURL, tt.expectedChanges["api.base_url"])
			}
			if tt.expectedChanges["behaviors.auth.type"] != nil && result.Behaviors.Auth.Type != tt.expectedChanges["behaviors.auth.type"] {
				t.Errorf("Behaviors.Auth.Type = %v, want %v", result.Behaviors.Auth.Type, tt.expectedChanges["behaviors.auth.type"])
			}
		})
	}
}

func TestCopyConfig(t *testing.T) {
	original := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name:    "test-cli",
			Version: "1.0.0",
		},
		API: cli.API{
			BaseURL:    "https://api.example.com",
			OpenAPIURL: "https://api.example.com/openapi.yaml",
		},
		Defaults: &cli.Defaults{
			HTTP: &cli.DefaultsHTTP{
				Timeout: "30s",
			},
		},
	}

	copy := copyConfig(original)

	// Verify copy is not nil
	if copy == nil {
		t.Fatal("copy is nil")
	}

	// Verify values are copied
	if copy.Metadata.Name != original.Metadata.Name {
		t.Errorf("Name = %v, want %v", copy.Metadata.Name, original.Metadata.Name)
	}
	if copy.API.BaseURL != original.API.BaseURL {
		t.Errorf("BaseURL = %v, want %v", copy.API.BaseURL, original.API.BaseURL)
	}

	// Verify deep copy (modifying copy doesn't affect original)
	copy.Metadata.Name = "modified"
	if original.Metadata.Name == "modified" {
		t.Error("original was modified when copy was changed - not a deep copy")
	}

	copy.Defaults.HTTP.Timeout = "60s"
	if original.Defaults.HTTP.Timeout == "60s" {
		t.Error("original defaults were modified when copy was changed - not a deep copy")
	}
}
