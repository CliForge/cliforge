package config

import (
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestValidator_ValidateMetadata(t *testing.T) {
	tests := []struct {
		name      string
		metadata  cli.Metadata
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid metadata",
			metadata: cli.Metadata{
				Name:        "my-cli",
				Version:     "1.0.0",
				Description: "A test CLI tool for testing",
			},
			wantError: false,
		},
		{
			name: "missing name",
			metadata: cli.Metadata{
				Version:     "1.0.0",
				Description: "A test CLI tool",
			},
			wantError: true,
			errorMsg:  "metadata.name",
		},
		{
			name: "invalid name - uppercase",
			metadata: cli.Metadata{
				Name:        "My-CLI",
				Version:     "1.0.0",
				Description: "A test CLI tool",
			},
			wantError: true,
			errorMsg:  "metadata.name",
		},
		{
			name: "invalid version format",
			metadata: cli.Metadata{
				Name:        "my-cli",
				Version:     "1.0",
				Description: "A test CLI tool",
			},
			wantError: true,
			errorMsg:  "metadata.version",
		},
		{
			name: "description too short",
			metadata: cli.Metadata{
				Name:        "my-cli",
				Version:     "1.0.0",
				Description: "Short",
			},
			wantError: true,
			errorMsg:  "metadata.description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateMetadata(&tt.metadata)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
			if tt.wantError && len(v.errors) > 0 {
				found := false
				for _, err := range v.errors {
					if err.Field == tt.errorMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error field %s, got errors: %v", tt.errorMsg, v.errors)
				}
			}
		})
	}
}

func TestValidator_ValidateAPI(t *testing.T) {
	tests := []struct {
		name      string
		api       cli.API
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid API",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
			},
			wantError: false,
		},
		{
			name: "valid file path",
			api: cli.API{
				OpenAPIURL: "./openapi.yaml",
				BaseURL:    "https://api.example.com",
			},
			wantError: false,
		},
		{
			name: "missing openapi_url",
			api: cli.API{
				BaseURL: "https://api.example.com",
			},
			wantError: true,
			errorMsg:  "api.openapi_url",
		},
		{
			name: "invalid base_url",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "not-a-url",
			},
			wantError: true,
			errorMsg:  "api.base_url",
		},
		{
			name: "valid environments",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						Name:       "production",
						OpenAPIURL: "https://api.example.com/openapi.yaml",
						BaseURL:    "https://api.example.com",
						Default:    true,
					},
				},
			},
			wantError: false,
		},
		{
			name: "no default environment",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						Name:       "production",
						OpenAPIURL: "https://api.example.com/openapi.yaml",
						BaseURL:    "https://api.example.com",
						Default:    false,
					},
				},
			},
			wantError: true,
			errorMsg:  "api.environments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateAPI(&tt.api)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
		})
	}
}

func TestValidator_ValidateDefaults(t *testing.T) {
	tests := []struct {
		name      string
		defaults  cli.Defaults
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid defaults",
			defaults: cli.Defaults{
				HTTP: &cli.DefaultsHTTP{
					Timeout: "30s",
				},
				Output: &cli.DefaultsOutput{
					Format: "json",
					Color:  "auto",
				},
			},
			wantError: false,
		},
		{
			name: "invalid timeout",
			defaults: cli.Defaults{
				HTTP: &cli.DefaultsHTTP{
					Timeout: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "defaults.http.timeout",
		},
		{
			name: "invalid output format",
			defaults: cli.Defaults{
				Output: &cli.DefaultsOutput{
					Format: "xml",
				},
			},
			wantError: true,
			errorMsg:  "defaults.output.format",
		},
		{
			name: "invalid color setting",
			defaults: cli.Defaults{
				Output: &cli.DefaultsOutput{
					Color: "sometimes",
				},
			},
			wantError: true,
			errorMsg:  "defaults.output.color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateDefaults(&tt.defaults)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
		})
	}
}

func TestValidator_ValidateBehaviors(t *testing.T) {
	tests := []struct {
		name      string
		behaviors cli.Behaviors
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid api_key auth",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "api_key",
					APIKey: &cli.APIKeyAuth{
						Header: "X-API-Key",
						EnvVar: "API_KEY",
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid auth type",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "custom",
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.type",
		},
		{
			name: "api_key auth missing config",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "api_key",
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.api_key",
		},
		{
			name: "valid oauth2 auth",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "oauth2",
					OAuth2: &cli.OAuth2Auth{
						ClientID: "client-id",
						AuthURL:  "https://auth.example.com/authorize",
						TokenURL: "https://auth.example.com/token",
					},
				},
			},
			wantError: false,
		},
		{
			name: "oauth2 missing token_url",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "oauth2",
					OAuth2: &cli.OAuth2Auth{
						ClientID: "client-id",
						AuthURL:  "https://auth.example.com/authorize",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.oauth2.token_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateBehaviors(&tt.behaviors)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	validConfig := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name:        "test-cli",
			Version:     "1.0.0",
			Description: "Test CLI for validation",
		},
		API: cli.API{
			OpenAPIURL: "https://api.example.com/openapi.yaml",
			BaseURL:    "https://api.example.com",
		},
	}

	invalidConfig := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name: "", // Missing required field
		},
		API: cli.API{
			BaseURL: "https://api.example.com",
			// Missing required openapi_url
		},
	}

	t.Run("valid config", func(t *testing.T) {
		v := NewValidator()
		err := v.Validate(validConfig)
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("invalid config", func(t *testing.T) {
		v := NewValidator()
		err := v.Validate(invalidConfig)
		if err == nil {
			t.Error("expected validation error but got none")
		}
	})
}

func TestValidator_ValidateUpdates(t *testing.T) {
	tests := []struct {
		name      string
		updates   *cli.Updates
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid updates config",
			updates: &cli.Updates{
				Enabled:       true,
				UpdateURL:     "https://releases.example.com/cli",
				CheckInterval: "24h",
			},
			wantError: false,
		},
		{
			name: "updates disabled",
			updates: &cli.Updates{
				Enabled: false,
			},
			wantError: false,
		},
		{
			name: "missing update URL when enabled",
			updates: &cli.Updates{
				Enabled: true,
			},
			wantError: true,
			errorMsg:  "updates.update_url",
		},
		{
			name: "invalid update URL",
			updates: &cli.Updates{
				Enabled:   true,
				UpdateURL: "not-a-url",
			},
			wantError: true,
			errorMsg:  "updates.update_url",
		},
		{
			name: "invalid check interval",
			updates: &cli.Updates{
				Enabled:       true,
				UpdateURL:     "https://releases.example.com/cli",
				CheckInterval: "not-a-duration",
			},
			wantError: true,
			errorMsg:  "updates.check_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateUpdates(tt.updates)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
			if tt.wantError && len(v.errors) > 0 {
				found := false
				for _, err := range v.errors {
					if err.Field == tt.errorMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error field %s, got errors: %v", tt.errorMsg, v.errors)
				}
			}
		})
	}
}

func TestValidator_ValidateUserPreferences(t *testing.T) {
	tests := []struct {
		name      string
		prefs     *cli.UserPreferences
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid preferences",
			prefs: &cli.UserPreferences{
				HTTP: &cli.PreferencesHTTP{
					Timeout:    "30s",
					Proxy:      "http://proxy.example.com:8080",
					HTTPSProxy: "https://proxy.example.com:8080",
				},
				Output: &cli.PreferencesOutput{
					Format: "json",
					Color:  "auto",
				},
			},
			wantError: false,
		},
		{
			name: "invalid timeout",
			prefs: &cli.UserPreferences{
				HTTP: &cli.PreferencesHTTP{
					Timeout: "not-a-duration",
				},
			},
			wantError: true,
			errorMsg:  "preferences.http.timeout",
		},
		{
			name: "invalid proxy URL",
			prefs: &cli.UserPreferences{
				HTTP: &cli.PreferencesHTTP{
					Proxy: "not-a-url",
				},
			},
			wantError: true,
			errorMsg:  "preferences.http.proxy",
		},
		{
			name: "invalid output format",
			prefs: &cli.UserPreferences{
				Output: &cli.PreferencesOutput{
					Format: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "preferences.output.format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			err := v.ValidateUserPreferences(tt.prefs)

			if tt.wantError && err == nil {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
			if tt.wantError && err != nil {
				ve, ok := err.(ValidationErrors)
				if !ok {
					t.Fatalf("expected ValidationErrors, got %T", err)
				}
				found := false
				for _, e := range ve {
					if e.Field == tt.errorMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error field %s, got errors: %v", tt.errorMsg, ve)
				}
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	err := &ValidationError{
		Field:   "test.field",
		Message: "test message",
	}
	expected := "test.field: test message"
	if err.Error() != expected {
		t.Errorf("Error() = %v, want %v", err.Error(), expected)
	}
}

func TestValidationErrors_Error(t *testing.T) {
	tests := []struct {
		name     string
		errors   ValidationErrors
		expected string
	}{
		{
			name:     "empty errors",
			errors:   ValidationErrors{},
			expected: "",
		},
		{
			name: "single error",
			errors: ValidationErrors{
				{Field: "field1", Message: "message1"},
			},
			expected: "validation failed:\n  - field1: message1",
		},
		{
			name: "multiple errors",
			errors: ValidationErrors{
				{Field: "field1", Message: "message1"},
				{Field: "field2", Message: "message2"},
			},
			expected: "validation failed:\n  - field1: message1\n  - field2: message2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.errors.Error()
			if result != tt.expected {
				t.Errorf("Error() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestValidator_ValidateBehaviors_Advanced(t *testing.T) {
	tests := []struct {
		name      string
		behaviors cli.Behaviors
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid caching behavior",
			behaviors: cli.Behaviors{
				Caching: &cli.CachingBehavior{
					SpecTTL:     "5m",
					ResponseTTL: "1m",
				},
			},
			wantError: false,
		},
		{
			name: "invalid caching spec_ttl",
			behaviors: cli.Behaviors{
				Caching: &cli.CachingBehavior{
					SpecTTL: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "behaviors.caching.spec_ttl",
		},
		{
			name: "invalid caching response_ttl",
			behaviors: cli.Behaviors{
				Caching: &cli.CachingBehavior{
					ResponseTTL: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "behaviors.caching.response_ttl",
		},
		{
			name: "valid retry behavior",
			behaviors: cli.Behaviors{
				Retry: &cli.RetryBehavior{
					InitialDelay:      "1s",
					MaxDelay:          "30s",
					BackoffMultiplier: 2.0,
				},
			},
			wantError: false,
		},
		{
			name: "invalid retry initial_delay",
			behaviors: cli.Behaviors{
				Retry: &cli.RetryBehavior{
					InitialDelay: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "behaviors.retry.initial_delay",
		},
		{
			name: "invalid retry max_delay",
			behaviors: cli.Behaviors{
				Retry: &cli.RetryBehavior{
					MaxDelay: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "behaviors.retry.max_delay",
		},
		{
			name: "invalid retry backoff_multiplier",
			behaviors: cli.Behaviors{
				Retry: &cli.RetryBehavior{
					BackoffMultiplier: 0.5,
				},
			},
			wantError: true,
			errorMsg:  "behaviors.retry.backoff_multiplier",
		},
		{
			name: "valid pagination behavior",
			behaviors: cli.Behaviors{
				Pagination: &cli.PaginationBehavior{
					MaxLimit: 100,
					Delay:    "100ms",
				},
			},
			wantError: false,
		},
		{
			name: "invalid pagination max_limit",
			behaviors: cli.Behaviors{
				Pagination: &cli.PaginationBehavior{
					MaxLimit: 0,
				},
			},
			wantError: true,
			errorMsg:  "behaviors.pagination.max_limit",
		},
		{
			name: "invalid pagination delay",
			behaviors: cli.Behaviors{
				Pagination: &cli.PaginationBehavior{
					Delay: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "behaviors.pagination.delay",
		},
		{
			name: "valid secrets masking",
			behaviors: cli.Behaviors{
				Secrets: &cli.SecretsBehavior{
					Masking: &cli.SecretsMasking{
						Style:            "partial",
						PartialShowChars: 4,
					},
				},
			},
			wantError: false,
		},
		{
			name: "invalid secrets masking style",
			behaviors: cli.Behaviors{
				Secrets: &cli.SecretsBehavior{
					Masking: &cli.SecretsMasking{
						Style: "invalid",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.secrets.masking.style",
		},
		{
			name: "invalid secrets partial_show_chars",
			behaviors: cli.Behaviors{
				Secrets: &cli.SecretsBehavior{
					Masking: &cli.SecretsMasking{
						PartialShowChars: -1,
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.secrets.masking.partial_show_chars",
		},
		{
			name: "valid basic auth",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "basic",
					Basic: &cli.BasicAuth{
						UsernameEnv: "USERNAME",
						PasswordEnv: "PASSWORD",
					},
				},
			},
			wantError: false,
		},
		{
			name: "basic auth missing username_env",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "basic",
					Basic: &cli.BasicAuth{
						PasswordEnv: "PASSWORD",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.basic.username_env",
		},
		{
			name: "basic auth missing password_env",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "basic",
					Basic: &cli.BasicAuth{
						UsernameEnv: "USERNAME",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.basic.password_env",
		},
		{
			name: "api_key auth missing header",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "api_key",
					APIKey: &cli.APIKeyAuth{
						EnvVar: "API_KEY",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.api_key.header",
		},
		{
			name: "api_key auth missing env_var",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "api_key",
					APIKey: &cli.APIKeyAuth{
						Header: "X-API-Key",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.api_key.env_var",
		},
		{
			name: "oauth2 missing client_id",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "oauth2",
					OAuth2: &cli.OAuth2Auth{
						AuthURL:  "https://auth.example.com/authorize",
						TokenURL: "https://auth.example.com/token",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.oauth2.client_id",
		},
		{
			name: "oauth2 missing auth_url",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "oauth2",
					OAuth2: &cli.OAuth2Auth{
						ClientID: "client-id",
						TokenURL: "https://auth.example.com/token",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.oauth2.auth_url",
		},
		{
			name: "oauth2 invalid auth_url",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "oauth2",
					OAuth2: &cli.OAuth2Auth{
						ClientID: "client-id",
						AuthURL:  "not-a-url",
						TokenURL: "https://auth.example.com/token",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.oauth2.auth_url",
		},
		{
			name: "oauth2 invalid token_url",
			behaviors: cli.Behaviors{
				Auth: &cli.AuthBehavior{
					Type: "oauth2",
					OAuth2: &cli.OAuth2Auth{
						ClientID: "client-id",
						AuthURL:  "https://auth.example.com/authorize",
						TokenURL: "not-a-url",
					},
				},
			},
			wantError: true,
			errorMsg:  "behaviors.auth.oauth2.token_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateBehaviors(&tt.behaviors)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
			if tt.wantError && len(v.errors) > 0 {
				found := false
				for _, err := range v.errors {
					if err.Field == tt.errorMsg {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected error field %s, got errors: %v", tt.errorMsg, v.errors)
				}
			}
		})
	}
}

func TestValidator_ValidateDefaults_Advanced(t *testing.T) {
	tests := []struct {
		name      string
		defaults  cli.Defaults
		wantError bool
		errorMsg  string
	}{
		{
			name: "negative pagination limit",
			defaults: cli.Defaults{
				Pagination: &cli.DefaultsPagination{
					Limit: -1,
				},
			},
			wantError: true,
			errorMsg:  "defaults.pagination.limit",
		},
		{
			name: "negative retry max_attempts",
			defaults: cli.Defaults{
				Retry: &cli.DefaultsRetry{
					MaxAttempts: -1,
				},
			},
			wantError: true,
			errorMsg:  "defaults.retry.max_attempts",
		},
		{
			name: "retry max_attempts exceeds limit",
			defaults: cli.Defaults{
				Retry: &cli.DefaultsRetry{
					MaxAttempts: 11,
				},
			},
			wantError: true,
			errorMsg:  "defaults.retry.max_attempts",
		},
		{
			name: "invalid deprecations min_severity",
			defaults: cli.Defaults{
				Deprecations: &cli.DefaultsDeprecations{
					MinSeverity: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "defaults.deprecations.min_severity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateDefaults(&tt.defaults)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
		})
	}
}

func TestValidator_ValidateMetadata_Advanced(t *testing.T) {
	tests := []struct {
		name      string
		metadata  cli.Metadata
		wantError bool
		errorMsg  string
	}{
		{
			name: "name too long",
			metadata: cli.Metadata{
				Name:        "this-is-a-very-long-name-that-exceeds-the-maximum-allowed-length-of-fifty-characters",
				Version:     "1.0.0",
				Description: "A test CLI tool",
			},
			wantError: true,
			errorMsg:  "metadata.name",
		},
		{
			name: "description too long",
			metadata: cli.Metadata{
				Name:        "my-cli",
				Version:     "1.0.0",
				Description: "This is a very long description that exceeds the maximum allowed length of two hundred characters. It should trigger a validation error because it is too long. This description is intentionally verbose to test the validation logic properly.",
			},
			wantError: true,
			errorMsg:  "metadata.description",
		},
		{
			name: "invalid homepage URL",
			metadata: cli.Metadata{
				Name:        "my-cli",
				Version:     "1.0.0",
				Description: "A test CLI tool",
				Homepage:    "not-a-url",
			},
			wantError: true,
			errorMsg:  "metadata.homepage",
		},
		{
			name: "invalid bugs_url",
			metadata: cli.Metadata{
				Name:        "my-cli",
				Version:     "1.0.0",
				Description: "A test CLI tool",
				BugsURL:     "not-a-url",
			},
			wantError: true,
			errorMsg:  "metadata.bugs_url",
		},
		{
			name: "invalid docs_url",
			metadata: cli.Metadata{
				Name:        "my-cli",
				Version:     "1.0.0",
				Description: "A test CLI tool",
				DocsURL:     "not-a-url",
			},
			wantError: true,
			errorMsg:  "metadata.docs_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateMetadata(&tt.metadata)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
		})
	}
}

func TestValidator_ValidateAPI_Advanced(t *testing.T) {
	tests := []struct {
		name      string
		api       cli.API
		wantError bool
		errorMsg  string
	}{
		{
			name: "invalid telemetry_url",
			api: cli.API{
				OpenAPIURL:   "https://api.example.com/openapi.yaml",
				BaseURL:      "https://api.example.com",
				TelemetryURL: "not-a-url",
			},
			wantError: true,
			errorMsg:  "api.telemetry_url",
		},
		{
			name: "multiple default environments",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						Name:       "prod",
						OpenAPIURL: "https://api.example.com/openapi.yaml",
						BaseURL:    "https://api.example.com",
						Default:    true,
					},
					{
						Name:       "staging",
						OpenAPIURL: "https://api.example.com/openapi.yaml",
						BaseURL:    "https://staging.example.com",
						Default:    true,
					},
				},
			},
			wantError: true,
			errorMsg:  "api.environments",
		},
		{
			name: "environment missing name",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						OpenAPIURL: "https://api.example.com/openapi.yaml",
						BaseURL:    "https://api.example.com",
					},
				},
			},
			wantError: true,
			errorMsg:  "api.environments[0].name",
		},
		{
			name: "environment missing openapi_url",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						Name:    "prod",
						BaseURL: "https://api.example.com",
					},
				},
			},
			wantError: true,
			errorMsg:  "api.environments[0].openapi_url",
		},
		{
			name: "environment invalid openapi_url",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						Name:       "prod",
						OpenAPIURL: "not-a-url-or-path",
						BaseURL:    "https://api.example.com",
					},
				},
			},
			wantError: true,
			errorMsg:  "api.environments[0].openapi_url",
		},
		{
			name: "environment missing base_url",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						Name:       "prod",
						OpenAPIURL: "https://api.example.com/openapi.yaml",
					},
				},
			},
			wantError: true,
			errorMsg:  "api.environments[0].base_url",
		},
		{
			name: "environment invalid base_url",
			api: cli.API{
				OpenAPIURL: "https://api.example.com/openapi.yaml",
				BaseURL:    "https://api.example.com",
				Environments: []cli.Environment{
					{
						Name:       "prod",
						OpenAPIURL: "https://api.example.com/openapi.yaml",
						BaseURL:    "not-a-url",
					},
				},
			},
			wantError: true,
			errorMsg:  "api.environments[0].base_url",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.validateAPI(&tt.api)

			if tt.wantError && len(v.errors) == 0 {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && len(v.errors) > 0 {
				t.Errorf("unexpected validation errors: %v", v.errors)
			}
		})
	}
}

func TestValidator_ValidateUserPreferences_Advanced(t *testing.T) {
	tests := []struct {
		name      string
		prefs     *cli.UserPreferences
		wantError bool
		errorMsg  string
	}{
		{
			name: "invalid https_proxy",
			prefs: &cli.UserPreferences{
				HTTP: &cli.PreferencesHTTP{
					HTTPSProxy: "not-a-url",
				},
			},
			wantError: true,
			errorMsg:  "preferences.http.https_proxy",
		},
		{
			name: "negative pagination limit",
			prefs: &cli.UserPreferences{
				Pagination: &cli.PreferencesPagination{
					Limit: -1,
				},
			},
			wantError: true,
			errorMsg:  "preferences.pagination.limit",
		},
		{
			name: "invalid output color",
			prefs: &cli.UserPreferences{
				Output: &cli.PreferencesOutput{
					Color: "sometimes",
				},
			},
			wantError: true,
			errorMsg:  "preferences.output.color",
		},
		{
			name: "invalid deprecations min_severity",
			prefs: &cli.UserPreferences{
				Deprecations: &cli.PreferencesDeprecations{
					MinSeverity: "invalid",
				},
			},
			wantError: true,
			errorMsg:  "preferences.deprecations.min_severity",
		},
		{
			name: "negative retry max_attempts",
			prefs: &cli.UserPreferences{
				Retry: &cli.PreferencesRetry{
					MaxAttempts: -1,
				},
			},
			wantError: true,
			errorMsg:  "preferences.retry.max_attempts",
		},
		{
			name: "retry max_attempts exceeds limit",
			prefs: &cli.UserPreferences{
				Retry: &cli.PreferencesRetry{
					MaxAttempts: 11,
				},
			},
			wantError: true,
			errorMsg:  "preferences.retry.max_attempts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			err := v.ValidateUserPreferences(tt.prefs)

			if tt.wantError && err == nil {
				t.Error("expected validation error but got none")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected validation error: %v", err)
			}
		})
	}
}

func TestValidator_isValidURL(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name  string
		url   string
		valid bool
	}{
		{"valid https URL", "https://example.com", true},
		{"valid http URL", "http://example.com", true},
		{"valid URL with path", "https://example.com/path", true},
		{"valid URL with port", "https://example.com:8080", true},
		{"empty string", "", false},
		{"no scheme", "example.com", false},
		{"no host", "https://", false},
		{"invalid URL", "not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := v.isValidURL(tt.url)
			if result != tt.valid {
				t.Errorf("isValidURL(%q) = %v, want %v", tt.url, result, tt.valid)
			}
		})
	}
}
