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
