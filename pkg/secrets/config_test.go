package secrets

import (
	"os"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()

	if config == nil {
		t.Fatal("NewConfig should not return nil")
	}

	if config.Behavior == nil {
		t.Fatal("Default config should have behavior set")
	}

	if !config.Behavior.Enabled {
		t.Error("Default config should be enabled")
	}

	if config.DisableMasking {
		t.Error("Masking should not be disabled by default")
	}

	if !config.WarnOnDisable {
		t.Error("Should warn on disable by default")
	}
}

func TestNewConfigFromBehavior(t *testing.T) {
	tests := []struct {
		name     string
		behavior *cli.SecretsBehavior
		check    func(t *testing.T, config *Config)
	}{
		{
			name:     "nil behavior uses defaults",
			behavior: nil,
			check: func(t *testing.T, config *Config) {
				if config == nil {
					t.Error("Config should not be nil")
				}
				if config.Behavior == nil {
					t.Error("Should have default behavior")
				}
			},
		},
		{
			name: "custom behavior",
			behavior: &cli.SecretsBehavior{
				Enabled: false,
			},
			check: func(t *testing.T, config *Config) {
				if config.Behavior.Enabled {
					t.Error("Should use provided behavior")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NewConfigFromBehavior(tt.behavior)
			tt.check(t, config)
		})
	}
}

func TestConfig_IsEnabled(t *testing.T) {
	tests := []struct {
		name           string
		config         *Config
		disableMasking bool
		want           bool
	}{
		{
			name: "enabled by default",
			config: &Config{
				Behavior: &cli.SecretsBehavior{
					Enabled: true,
				},
				DisableMasking: false,
			},
			want: true,
		},
		{
			name: "disabled by flag",
			config: &Config{
				Behavior: &cli.SecretsBehavior{
					Enabled: true,
				},
				DisableMasking: true,
			},
			want: false,
		},
		{
			name: "disabled in config",
			config: &Config{
				Behavior: &cli.SecretsBehavior{
					Enabled: false,
				},
				DisableMasking: false,
			},
			want: false,
		},
		{
			name: "nil behavior",
			config: &Config{
				Behavior:       nil,
				DisableMasking: false,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.IsEnabled()
			if got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Behavior: DefaultSecretsBehavior(),
			},
			wantErr: false,
		},
		{
			name: "invalid masking style",
			config: &Config{
				Behavior: &cli.SecretsBehavior{
					Enabled: true,
					Masking: &cli.SecretsMasking{
						Style: "invalid",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "nil behavior",
			config: &Config{
				Behavior: nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_ApplyEnvironmentOverrides(t *testing.T) {
	// Save and restore environment
	oldEnv := os.Getenv("TEST_NO_MASK_SECRETS")
	defer func() {
		if oldEnv != "" {
			_ = os.Setenv("TEST_NO_MASK_SECRETS", oldEnv)
		} else {
			_ = os.Unsetenv("TEST_NO_MASK_SECRETS")
		}
	}()

	tests := []struct {
		name         string
		envValue     string
		wantDisabled bool
	}{
		{
			name:         "env not set",
			envValue:     "",
			wantDisabled: false,
		},
		{
			name:         "env set to 1",
			envValue:     "1",
			wantDisabled: true,
		},
		{
			name:         "env set to true",
			envValue:     "true",
			wantDisabled: true,
		},
		{
			name:         "env set to TRUE",
			envValue:     "TRUE",
			wantDisabled: true,
		},
		{
			name:         "env set to false",
			envValue:     "false",
			wantDisabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				_ = os.Setenv("TEST_NO_MASK_SECRETS", tt.envValue)
			} else {
				_ = os.Unsetenv("TEST_NO_MASK_SECRETS")
			}

			config := NewConfig()
			config.WarnOnDisable = false // Disable warnings for tests

			config.ApplyEnvironmentOverrides("TEST")

			if config.DisableMasking != tt.wantDisabled {
				t.Errorf("DisableMasking = %v, want %v", config.DisableMasking, tt.wantDisabled)
			}
		})
	}
}

func TestConfigBuilder(t *testing.T) {
	t.Run("basic build", func(t *testing.T) {
		config, err := NewConfigBuilder().
			WithEnabled(true).
			WithMaskingStyle("full").
			WithPartialShowChars(8).
			WithReplacement("[REDACTED]").
			Build()

		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		if !config.Behavior.Enabled {
			t.Error("Should be enabled")
		}

		if config.Behavior.Masking.Style != "full" {
			t.Errorf("Style = %q, want %q", config.Behavior.Masking.Style, "full")
		}

		if config.Behavior.Masking.PartialShowChars != 8 {
			t.Errorf("PartialShowChars = %d, want %d", config.Behavior.Masking.PartialShowChars, 8)
		}

		if config.Behavior.Masking.Replacement != "[REDACTED]" {
			t.Errorf("Replacement = %q, want %q", config.Behavior.Masking.Replacement, "[REDACTED]")
		}
	})

	t.Run("field patterns", func(t *testing.T) {
		config, err := NewConfigBuilder().
			WithFieldPatterns([]string{"*custom*"}).
			AddFieldPattern("*another*").
			Build()

		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		if len(config.Behavior.FieldPatterns) != 2 {
			t.Errorf("Expected 2 field patterns, got %d", len(config.Behavior.FieldPatterns))
		}
	})

	t.Run("value patterns", func(t *testing.T) {
		pattern := cli.ValuePattern{
			Name:    "Test Pattern",
			Pattern: "test_[0-9]+",
			Enabled: true,
		}

		config, err := NewConfigBuilder().
			WithValuePatterns([]cli.ValuePattern{pattern}).
			AddValuePattern(cli.ValuePattern{
				Name:    "Another Pattern",
				Pattern: "another_[a-z]+",
				Enabled: true,
			}).
			Build()

		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		if len(config.Behavior.ValuePatterns) != 2 {
			t.Errorf("Expected 2 value patterns, got %d", len(config.Behavior.ValuePatterns))
		}
	})

	t.Run("headers", func(t *testing.T) {
		config, err := NewConfigBuilder().
			WithHeaders([]string{"X-Custom-Header"}).
			AddHeader("X-Another-Header").
			Build()

		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		if len(config.Behavior.Headers) != 2 {
			t.Errorf("Expected 2 headers, got %d", len(config.Behavior.Headers))
		}
	})

	t.Run("explicit fields", func(t *testing.T) {
		config, err := NewConfigBuilder().
			WithExplicitFields([]string{"$.user.password"}).
			AddExplicitField("$.credentials.token").
			Build()

		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		if len(config.Behavior.ExplicitFields) != 2 {
			t.Errorf("Expected 2 explicit fields, got %d", len(config.Behavior.ExplicitFields))
		}
	})

	t.Run("mask_in settings", func(t *testing.T) {
		maskIn := &cli.SecretsMaskIn{
			Stdout:      false,
			Stderr:      true,
			Logs:        true,
			Audit:       false,
			DebugOutput: true,
		}

		config, err := NewConfigBuilder().
			WithMaskIn(maskIn).
			Build()

		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		if config.Behavior.MaskIn.Stdout {
			t.Error("Stdout masking should be disabled")
		}

		if !config.Behavior.MaskIn.Stderr {
			t.Error("Stderr masking should be enabled")
		}
	})

	t.Run("disable masking and warnings", func(t *testing.T) {
		config, err := NewConfigBuilder().
			WithDisableMasking(true).
			WithWarnOnDisable(false).
			Build()

		if err != nil {
			t.Fatalf("Build() failed: %v", err)
		}

		if !config.DisableMasking {
			t.Error("Masking should be disabled")
		}

		if config.WarnOnDisable {
			t.Error("Warnings should be disabled")
		}
	})

	t.Run("invalid config fails validation", func(t *testing.T) {
		_, err := NewConfigBuilder().
			WithMaskingStyle("invalid").
			Build()

		if err == nil {
			t.Error("Expected validation error for invalid masking style")
		}
	})

	t.Run("MustBuild panics on error", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid config")
			}
		}()

		NewConfigBuilder().
			WithMaskingStyle("invalid").
			MustBuild()
	})
}

func TestLoadConfigFromCLIConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *cli.Config
		check  func(t *testing.T, result *Config)
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			check: func(t *testing.T, result *Config) {
				if result == nil {
					t.Error("Result should not be nil")
				}
				if result.Behavior == nil {
					t.Error("Should have default behavior")
				}
			},
		},
		{
			name: "config without behaviors",
			config: &cli.Config{
				Behaviors: nil,
			},
			check: func(t *testing.T, result *Config) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "config with secrets behavior",
			config: &cli.Config{
				Behaviors: &cli.Behaviors{
					Secrets: &cli.SecretsBehavior{
						Enabled: false,
					},
				},
			},
			check: func(t *testing.T, result *Config) {
				if result.Behavior.Enabled {
					t.Error("Should use provided behavior")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LoadConfigFromCLIConfig(tt.config)
			tt.check(t, result)
		})
	}
}

func TestConfig_GetMaskingStrategy(t *testing.T) {
	tests := []struct {
		name         string
		config       *Config
		expectedType string
	}{
		{
			name: "partial strategy",
			config: &Config{
				Behavior: &cli.SecretsBehavior{
					Masking: &cli.SecretsMasking{
						Style: "partial",
					},
				},
			},
			expectedType: "partial",
		},
		{
			name: "full strategy",
			config: &Config{
				Behavior: &cli.SecretsBehavior{
					Masking: &cli.SecretsMasking{
						Style: "full",
					},
				},
			},
			expectedType: "full",
		},
		{
			name: "hash strategy",
			config: &Config{
				Behavior: &cli.SecretsBehavior{
					Masking: &cli.SecretsMasking{
						Style: "hash",
					},
				},
			},
			expectedType: "hash",
		},
		{
			name: "nil behavior defaults to partial",
			config: &Config{
				Behavior: nil,
			},
			expectedType: "partial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := tt.config.GetMaskingStrategy()
			if strategy == nil {
				t.Fatal("GetMaskingStrategy should not return nil")
			}

			if strategy.Name() != tt.expectedType {
				t.Errorf("Expected strategy type %q, got %q", tt.expectedType, strategy.Name())
			}
		})
	}
}
