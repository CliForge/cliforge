package config

import (
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestMergeDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   *cli.Config
		expected *cli.Config
	}{
		{
			name:   "empty config gets defaults",
			config: &cli.Config{},
			expected: &cli.Config{
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
			config: &cli.Config{
				Defaults: &cli.Defaults{
					HTTP: &cli.DefaultsHTTP{
						Timeout: "60s",
					},
				},
			},
			expected: &cli.Config{
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
		config      *cli.Config
		preferences *cli.UserPreferences
		expected    map[string]interface{}
	}{
		{
			name: "apply HTTP timeout preference",
			config: &cli.Config{
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
			config: &cli.Config{
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
			config: &cli.Config{
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
			config: &cli.Config{
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
		config          *cli.Config
		override        *cli.Config
		expectedChanges map[string]interface{}
	}{
		{
			name: "override API base_url",
			config: &cli.Config{
				API: cli.API{
					BaseURL: "https://api.production.com",
				},
			},
			override: &cli.Config{
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
			config: &cli.Config{
				Behaviors: &cli.Behaviors{
					Auth: &cli.AuthBehavior{
						Type: "oauth2",
					},
				},
			},
			override: &cli.Config{
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
	original := &cli.Config{
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

	copiedConfig := copyConfig(original)

	// Verify copy is not nil
	if copiedConfig == nil {
		t.Fatal("copy is nil")
	}

	// Verify values are copied
	if copiedConfig.Metadata.Name != original.Metadata.Name {
		t.Errorf("Name = %v, want %v", copiedConfig.Metadata.Name, original.Metadata.Name)
	}
	if copiedConfig.API.BaseURL != original.API.BaseURL {
		t.Errorf("BaseURL = %v, want %v", copiedConfig.API.BaseURL, original.API.BaseURL)
	}

	// Verify deep copy (modifying copy doesn't affect original)
	copiedConfig.Metadata.Name = "modified"
	if original.Metadata.Name == "modified" {
		t.Error("original was modified when copy was changed - not a deep copy")
	}

	copiedConfig.Defaults.HTTP.Timeout = "60s"
	if original.Defaults.HTTP.Timeout == "60s" {
		t.Error("original defaults were modified when copy was changed - not a deep copy")
	}
}

func TestMergeConfigsIntegration(t *testing.T) {
	embedded := &cli.Config{
		Metadata: cli.Metadata{
			Name:    "test-cli",
			Version: "1.0.0",
			Debug:   true,
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

	user := &cli.UserConfig{
		Preferences: &cli.UserPreferences{
			HTTP: &cli.PreferencesHTTP{
				Timeout: "60s",
			},
		},
		DebugOverride: &cli.Config{
			API: cli.API{
				BaseURL: "http://localhost:8080",
			},
		},
	}

	loader := &Loader{cliName: "test-cli"}
	merged, overrides, err := loader.mergeConfigs(embedded, user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if merged == nil {
		t.Fatal("expected merged config but got nil")
	}

	// Check debug overrides were applied
	if merged.API.BaseURL != "http://localhost:8080" {
		t.Errorf("expected base URL from debug override, got %s", merged.API.BaseURL)
	}

	if len(overrides) == 0 {
		t.Error("expected debug overrides to be tracked")
	}

	// Check preferences were applied
	if merged.Defaults.HTTP.Timeout != "60s" {
		t.Errorf("expected timeout from preferences, got %s", merged.Defaults.HTTP.Timeout)
	}
}

func TestMergeConfigsNoDebugMode(t *testing.T) {
	embedded := &cli.Config{
		Metadata: cli.Metadata{
			Name:    "test-cli",
			Version: "1.0.0",
			Debug:   false, // Debug disabled
		},
		API: cli.API{
			BaseURL: "https://api.example.com",
		},
	}

	user := &cli.UserConfig{
		DebugOverride: &cli.Config{
			API: cli.API{
				BaseURL: "http://localhost:8080",
			},
		},
	}

	loader := &Loader{cliName: "test-cli"}
	merged, overrides, err := loader.mergeConfigs(embedded, user)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Debug overrides should NOT be applied when debug is false
	if merged.API.BaseURL != "https://api.example.com" {
		t.Error("debug overrides should not be applied when debug mode is disabled")
	}

	if len(overrides) != 0 {
		t.Error("expected no debug overrides when debug mode is disabled")
	}
}

func TestCopyConfig_DeepCopy(t *testing.T) {
	original := &cli.Config{
		Metadata: cli.Metadata{
			Name:    "test-cli",
			Version: "1.0.0",
		},
		API: cli.API{
			BaseURL: "https://api.example.com",
		},
		Defaults: &cli.Defaults{
			HTTP: &cli.DefaultsHTTP{
				Timeout: "30s",
			},
		},
	}

	copied := copyConfig(original)

	// Verify it's a different object
	if copied == original {
		t.Error("copyConfig should return a new object")
	}

	// Modify copy and verify original is unchanged
	copied.API.BaseURL = "https://different.com"
	if original.API.BaseURL != "https://api.example.com" {
		t.Error("modifying copy should not affect original")
	}

	// Verify nested structures are copied
	if copied.Defaults != nil {
		copied.Defaults.HTTP.Timeout = "60s"
		if original.Defaults.HTTP.Timeout != "30s" {
			t.Error("modifying nested copy should not affect original")
		}
	}
}

func TestCopyConfig_NilInput(t *testing.T) {
	result := copyConfig(nil)
	if result != nil {
		t.Error("copyConfig(nil) should return nil")
	}
}

func TestCopyConfig_CompleteStructure(t *testing.T) {
	// Test copying a config with all possible fields populated
	original := &cli.Config{
		Metadata: cli.Metadata{
			Name:        "test-cli",
			Version:     "1.0.0",
			Description: "Test CLI",
		},
		API: cli.API{
			BaseURL:    "https://api.example.com",
			OpenAPIURL: "https://api.example.com/openapi.yaml",
			Environments: []cli.Environment{
				{Name: "prod", BaseURL: "https://prod.example.com", OpenAPIURL: "https://prod.example.com/openapi.yaml"},
				{Name: "dev", BaseURL: "https://dev.example.com", OpenAPIURL: "https://dev.example.com/openapi.yaml"},
			},
			DefaultHeaders: map[string]string{
				"X-Custom-Header": "value",
			},
		},
		Branding: &cli.Branding{
			Colors: &cli.Colors{
				Primary:   "#FF0000",
				Secondary: "#00FF00",
			},
			Prompts: &cli.Prompts{
				Command: "$ ",
			},
			Theme: &cli.Theme{
				Name: "dark",
			},
			ASCIIArt: "ASCII Art",
		},
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
			},
			Deprecations: &cli.DefaultsDeprecations{
				AlwaysShow:  false,
				MinSeverity: "info",
			},
			Retry: &cli.DefaultsRetry{
				MaxAttempts: 3,
			},
		},
		Behaviors: &cli.Behaviors{
			Auth: &cli.AuthBehavior{
				Type: "api_key",
				APIKey: &cli.APIKeyAuth{
					Header: "X-API-Key",
					EnvVar: "API_KEY",
				},
				OAuth2: &cli.OAuth2Auth{
					ClientID: "client-id",
					Scopes:   []string{"read", "write"},
				},
				Basic: &cli.BasicAuth{
					UsernameEnv: "USERNAME",
					PasswordEnv: "PASSWORD",
				},
			},
			Caching: &cli.CachingBehavior{
				SpecTTL:     "5m",
				ResponseTTL: "1m",
			},
			Retry: &cli.RetryBehavior{
				InitialDelay:      "1s",
				MaxDelay:          "30s",
				BackoffMultiplier: 2.0,
				RetryOnStatus:     []int{429, 500, 502},
			},
			Pagination: &cli.PaginationBehavior{
				MaxLimit: 100,
				Delay:    "100ms",
			},
			Secrets: &cli.SecretsBehavior{
				Masking: &cli.SecretsMasking{
					Style:            "partial",
					PartialShowChars: 4,
				},
				FieldPatterns: []string{"password", "token"},
				Headers:       []string{"Authorization"},
				MaskIn: &cli.SecretsMaskIn{
					Stdout: true,
					Logs:   true,
				},
			},
		},
		Updates: &cli.Updates{
			Enabled:   true,
			UpdateURL: "https://updates.example.com",
		},
		Features: &cli.Features{
			ConfigFile: true,
		},
	}

	copied := copyConfig(original)

	// Verify branding is deep copied
	if copied.Branding.Colors.Primary != original.Branding.Colors.Primary {
		t.Error("branding colors not copied correctly")
	}
	copied.Branding.Colors.Primary = "#0000FF"
	if original.Branding.Colors.Primary == "#0000FF" {
		t.Error("modifying copied branding affected original")
	}

	// Verify environments slice is deep copied
	if len(copied.API.Environments) != len(original.API.Environments) {
		t.Error("environments not copied correctly")
	}
	copied.API.Environments[0].Name = "modified"
	if original.API.Environments[0].Name == "modified" {
		t.Error("modifying copied environments affected original")
	}

	// Verify default headers map is deep copied
	if copied.API.DefaultHeaders["X-Custom-Header"] != "value" {
		t.Error("default headers not copied correctly")
	}
	copied.API.DefaultHeaders["X-Custom-Header"] = "modified"
	if original.API.DefaultHeaders["X-Custom-Header"] == "modified" {
		t.Error("modifying copied headers affected original")
	}

	// Verify behaviors auth is deep copied
	if copied.Behaviors.Auth.APIKey.Header != original.Behaviors.Auth.APIKey.Header {
		t.Error("auth api_key not copied correctly")
	}
	copied.Behaviors.Auth.APIKey.Header = "X-Modified"
	if original.Behaviors.Auth.APIKey.Header == "X-Modified" {
		t.Error("modifying copied auth affected original")
	}

	// Verify OAuth2 scopes slice is deep copied
	if len(copied.Behaviors.Auth.OAuth2.Scopes) != 2 {
		t.Error("OAuth2 scopes not copied correctly")
	}
	copied.Behaviors.Auth.OAuth2.Scopes[0] = "modified"
	if original.Behaviors.Auth.OAuth2.Scopes[0] == "modified" {
		t.Error("modifying copied scopes affected original")
	}

	// Verify retry behavior slices are deep copied
	if len(copied.Behaviors.Retry.RetryOnStatus) != 3 {
		t.Error("retry on status not copied correctly")
	}
	copied.Behaviors.Retry.RetryOnStatus[0] = 999
	if original.Behaviors.Retry.RetryOnStatus[0] == 999 {
		t.Error("modifying copied retry status affected original")
	}

	// Verify secrets behavior slices are deep copied
	if len(copied.Behaviors.Secrets.FieldPatterns) != 2 {
		t.Error("field patterns not copied correctly")
	}
	copied.Behaviors.Secrets.FieldPatterns[0] = "modified"
	if original.Behaviors.Secrets.FieldPatterns[0] == "modified" {
		t.Error("modifying copied field patterns affected original")
	}

	// Verify updates is deep copied
	copied.Updates.UpdateURL = "https://modified.com"
	if original.Updates.UpdateURL == "https://modified.com" {
		t.Error("modifying copied updates affected original")
	}

	// Verify features is deep copied
	copied.Features.ConfigFile = false
	if original.Features.ConfigFile == false {
		t.Error("modifying copied features affected original")
	}
}

func TestMergeDefaults_NilConfig(t *testing.T) {
	err := MergeDefaults(nil)
	if err == nil {
		t.Error("expected error for nil config")
	}
	if err.Error() != "config is nil" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestMergeDefaults_BehaviorDefaults(t *testing.T) {
	tests := []struct {
		name     string
		config   *cli.Config
		expected map[string]interface{}
	}{
		{
			name: "apply caching behavior defaults",
			config: &cli.Config{
				Behaviors: &cli.Behaviors{
					Caching: &cli.CachingBehavior{},
				},
			},
			expected: map[string]interface{}{
				"caching.spec_ttl":     "5m",
				"caching.response_ttl": "1m",
				"caching.max_size":     "100MB",
			},
		},
		{
			name: "apply retry behavior defaults",
			config: &cli.Config{
				Behaviors: &cli.Behaviors{
					Retry: &cli.RetryBehavior{},
				},
			},
			expected: map[string]interface{}{
				"retry.initial_delay":       "1s",
				"retry.max_delay":           "30s",
				"retry.backoff_multiplier":  2.0,
				"retry.retry_on_status_len": 5,
			},
		},
		{
			name: "apply pagination behavior defaults",
			config: &cli.Config{
				Behaviors: &cli.Behaviors{
					Pagination: &cli.PaginationBehavior{},
				},
			},
			expected: map[string]interface{}{
				"pagination.delay":     "100ms",
				"pagination.max_limit": 100,
			},
		},
		{
			name: "preserve existing behavior values",
			config: &cli.Config{
				Behaviors: &cli.Behaviors{
					Caching: &cli.CachingBehavior{
						SpecTTL:     "10m",
						ResponseTTL: "2m",
						MaxSize:     "200MB",
					},
				},
			},
			expected: map[string]interface{}{
				"caching.spec_ttl":     "10m",
				"caching.response_ttl": "2m",
				"caching.max_size":     "200MB",
			},
		},
		{
			name: "partial behavior defaults",
			config: &cli.Config{
				Behaviors: &cli.Behaviors{
					Retry: &cli.RetryBehavior{
						InitialDelay: "2s",
					},
				},
			},
			expected: map[string]interface{}{
				"retry.initial_delay":       "2s",
				"retry.max_delay":           "30s",
				"retry.backoff_multiplier":  2.0,
				"retry.retry_on_status_len": 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MergeDefaults(tt.config)
			if err != nil {
				t.Fatalf("MergeDefaults failed: %v", err)
			}

			for key, expected := range tt.expected {
				switch key {
				case "caching.spec_ttl":
					if tt.config.Behaviors.Caching.SpecTTL != expected {
						t.Errorf("SpecTTL = %v, want %v", tt.config.Behaviors.Caching.SpecTTL, expected)
					}
				case "caching.response_ttl":
					if tt.config.Behaviors.Caching.ResponseTTL != expected {
						t.Errorf("ResponseTTL = %v, want %v", tt.config.Behaviors.Caching.ResponseTTL, expected)
					}
				case "caching.max_size":
					if tt.config.Behaviors.Caching.MaxSize != expected {
						t.Errorf("MaxSize = %v, want %v", tt.config.Behaviors.Caching.MaxSize, expected)
					}
				case "retry.initial_delay":
					if tt.config.Behaviors.Retry.InitialDelay != expected {
						t.Errorf("InitialDelay = %v, want %v", tt.config.Behaviors.Retry.InitialDelay, expected)
					}
				case "retry.max_delay":
					if tt.config.Behaviors.Retry.MaxDelay != expected {
						t.Errorf("MaxDelay = %v, want %v", tt.config.Behaviors.Retry.MaxDelay, expected)
					}
				case "retry.backoff_multiplier":
					if tt.config.Behaviors.Retry.BackoffMultiplier != expected {
						t.Errorf("BackoffMultiplier = %v, want %v", tt.config.Behaviors.Retry.BackoffMultiplier, expected)
					}
				case "retry.retry_on_status_len":
					if len(tt.config.Behaviors.Retry.RetryOnStatus) != expected {
						t.Errorf("RetryOnStatus length = %v, want %v", len(tt.config.Behaviors.Retry.RetryOnStatus), expected)
					}
				case "pagination.delay":
					if tt.config.Behaviors.Pagination.Delay != expected {
						t.Errorf("Delay = %v, want %v", tt.config.Behaviors.Pagination.Delay, expected)
					}
				case "pagination.max_limit":
					if tt.config.Behaviors.Pagination.MaxLimit != expected {
						t.Errorf("MaxLimit = %v, want %v", tt.config.Behaviors.Pagination.MaxLimit, expected)
					}
				}
			}
		})
	}
}

func TestMergeDefaults_AllDefaultSections(t *testing.T) {
	config := &cli.Config{
		Defaults: &cli.Defaults{
			HTTP:         &cli.DefaultsHTTP{},
			Pagination:   &cli.DefaultsPagination{Limit: 0},
			Output:       &cli.DefaultsOutput{},
			Deprecations: &cli.DefaultsDeprecations{},
			Retry:        &cli.DefaultsRetry{MaxAttempts: 0},
		},
	}

	err := MergeDefaults(config)
	if err != nil {
		t.Fatalf("MergeDefaults failed: %v", err)
	}

	// Verify HTTP timeout is set when empty
	if config.Defaults.HTTP.Timeout != "30s" {
		t.Errorf("HTTP.Timeout = %v, want 30s", config.Defaults.HTTP.Timeout)
	}

	// Verify pagination limit is set when 0
	if config.Defaults.Pagination.Limit != 20 {
		t.Errorf("Pagination.Limit = %v, want 20", config.Defaults.Pagination.Limit)
	}

	// Verify output defaults are set
	if config.Defaults.Output.Format != "json" {
		t.Errorf("Output.Format = %v, want json", config.Defaults.Output.Format)
	}
	if config.Defaults.Output.Color != "auto" {
		t.Errorf("Output.Color = %v, want auto", config.Defaults.Output.Color)
	}

	// Verify deprecation severity is set
	if config.Defaults.Deprecations.MinSeverity != "info" {
		t.Errorf("Deprecations.MinSeverity = %v, want info", config.Defaults.Deprecations.MinSeverity)
	}

	// Verify retry max attempts is set when 0
	if config.Defaults.Retry.MaxAttempts != 3 {
		t.Errorf("Retry.MaxAttempts = %v, want 3", config.Defaults.Retry.MaxAttempts)
	}
}

func TestApplyDebugOverrides_AllFields(t *testing.T) {
	loader := &Loader{}

	tests := []struct {
		name            string
		config          *cli.Config
		override        *cli.Config
		expectedChanges map[string]interface{}
	}{
		{
			name: "override OpenAPI URL",
			config: &cli.Config{
				API: cli.API{
					OpenAPIURL: "https://api.production.com/openapi.yaml",
				},
			},
			override: &cli.Config{
				API: cli.API{
					OpenAPIURL: "http://localhost:8080/openapi.yaml",
				},
			},
			expectedChanges: map[string]interface{}{
				"api.openapi_url": "http://localhost:8080/openapi.yaml",
			},
		},
		{
			name: "override metadata name",
			config: &cli.Config{
				Metadata: cli.Metadata{
					Name: "production-cli",
				},
			},
			override: &cli.Config{
				Metadata: cli.Metadata{
					Name: "debug-cli",
				},
			},
			expectedChanges: map[string]interface{}{
				"metadata.name": "debug-cli",
			},
		},
		{
			name: "override branding with nil branding",
			config: &cli.Config{
				Metadata: cli.Metadata{
					Name: "test-cli",
				},
			},
			override: &cli.Config{
				Branding: &cli.Branding{
					Colors: &cli.Colors{
						Primary: "#FF0000",
					},
				},
			},
			expectedChanges: map[string]interface{}{
				"branding.colors.primary": "#FF0000",
			},
		},
		{
			name: "override auth with nil behaviors",
			config: &cli.Config{
				Metadata: cli.Metadata{
					Name: "test-cli",
				},
				Behaviors: &cli.Behaviors{
					Auth: &cli.AuthBehavior{
						Type: "oauth2",
					},
				},
			},
			override: &cli.Config{
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
		{
			name: "no changes when values are same",
			config: &cli.Config{
				API: cli.API{
					BaseURL: "https://api.example.com",
				},
			},
			override: &cli.Config{
				API: cli.API{
					BaseURL: "https://api.example.com",
				},
			},
			expectedChanges: map[string]interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, overrides := loader.applyDebugOverrides(tt.config, tt.override)

			// Check that overrides map contains expected changes
			if len(overrides) != len(tt.expectedChanges) {
				t.Errorf("overrides count = %v, want %v", len(overrides), len(tt.expectedChanges))
			}

			for key, expected := range tt.expectedChanges {
				if overrides[key] != expected {
					t.Errorf("override[%s] = %v, want %v", key, overrides[key], expected)
				}
			}

			// Verify result is not nil
			if result == nil {
				t.Fatal("result should not be nil")
			}
		})
	}
}

func TestApplyUserPreferences_InitializeNilDefaults(t *testing.T) {
	loader := &Loader{}

	tests := []struct {
		name        string
		config      *cli.Config
		preferences *cli.UserPreferences
		checkFunc   func(*testing.T, *cli.Config)
	}{
		{
			name:   "initialize nil defaults",
			config: &cli.Config{},
			preferences: &cli.UserPreferences{
				HTTP: &cli.PreferencesHTTP{
					Timeout: "60s",
				},
			},
			checkFunc: func(t *testing.T, cfg *cli.Config) {
				if cfg.Defaults == nil {
					t.Error("Defaults should be initialized")
				}
				if cfg.Defaults.HTTP == nil {
					t.Error("HTTP defaults should be initialized")
				}
				if cfg.Defaults.HTTP.Timeout != "60s" {
					t.Errorf("HTTP.Timeout = %v, want 60s", cfg.Defaults.HTTP.Timeout)
				}
			},
		},
		{
			name: "initialize nil HTTP defaults",
			config: &cli.Config{
				Defaults: &cli.Defaults{},
			},
			preferences: &cli.UserPreferences{
				HTTP: &cli.PreferencesHTTP{
					Timeout: "45s",
				},
			},
			checkFunc: func(t *testing.T, cfg *cli.Config) {
				if cfg.Defaults.HTTP == nil {
					t.Error("HTTP defaults should be initialized")
				}
			},
		},
		{
			name: "apply caching preference",
			config: &cli.Config{
				Defaults: &cli.Defaults{},
			},
			preferences: &cli.UserPreferences{
				Caching: &cli.PreferencesCaching{
					Enabled: false,
				},
			},
			checkFunc: func(t *testing.T, cfg *cli.Config) {
				if cfg.Defaults.Caching == nil {
					t.Error("Caching defaults should be initialized")
				}
				if cfg.Defaults.Caching.Enabled != false {
					t.Error("Caching.Enabled should be false")
				}
			},
		},
		{
			name: "apply all output preferences",
			config: &cli.Config{
				Defaults: &cli.Defaults{},
			},
			preferences: &cli.UserPreferences{
				Output: &cli.PreferencesOutput{
					Format:      "yaml",
					Color:       "never",
					PrettyPrint: false,
					Paging:      false,
				},
			},
			checkFunc: func(t *testing.T, cfg *cli.Config) {
				if cfg.Defaults.Output == nil {
					t.Fatal("Output defaults should be initialized")
				}
				if cfg.Defaults.Output.Format != "yaml" {
					t.Errorf("Output.Format = %v, want yaml", cfg.Defaults.Output.Format)
				}
				if cfg.Defaults.Output.Color != "never" {
					t.Errorf("Output.Color = %v, want never", cfg.Defaults.Output.Color)
				}
				if cfg.Defaults.Output.PrettyPrint != false {
					t.Error("Output.PrettyPrint should be false")
				}
				if cfg.Defaults.Output.Paging != false {
					t.Error("Output.Paging should be false")
				}
			},
		},
		{
			name: "apply deprecations preference",
			config: &cli.Config{
				Defaults: &cli.Defaults{},
			},
			preferences: &cli.UserPreferences{
				Deprecations: &cli.PreferencesDeprecations{
					AlwaysShow:  true,
					MinSeverity: "warning",
				},
			},
			checkFunc: func(t *testing.T, cfg *cli.Config) {
				if cfg.Defaults.Deprecations == nil {
					t.Fatal("Deprecations defaults should be initialized")
				}
				if cfg.Defaults.Deprecations.AlwaysShow != true {
					t.Error("Deprecations.AlwaysShow should be true")
				}
				if cfg.Defaults.Deprecations.MinSeverity != "warning" {
					t.Errorf("Deprecations.MinSeverity = %v, want warning", cfg.Defaults.Deprecations.MinSeverity)
				}
			},
		},
		{
			name: "apply retry preference",
			config: &cli.Config{
				Defaults: &cli.Defaults{},
			},
			preferences: &cli.UserPreferences{
				Retry: &cli.PreferencesRetry{
					MaxAttempts: 5,
				},
			},
			checkFunc: func(t *testing.T, cfg *cli.Config) {
				if cfg.Defaults.Retry == nil {
					t.Fatal("Retry defaults should be initialized")
				}
				if cfg.Defaults.Retry.MaxAttempts != 5 {
					t.Errorf("Retry.MaxAttempts = %v, want 5", cfg.Defaults.Retry.MaxAttempts)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := loader.applyUserPreferences(tt.config, tt.preferences)
			tt.checkFunc(t, result)
		})
	}
}
