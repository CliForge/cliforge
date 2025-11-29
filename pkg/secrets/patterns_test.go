package secrets

import (
	"regexp"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestDefaultFieldPatterns(t *testing.T) {
	patterns := DefaultFieldPatterns()

	if len(patterns) == 0 {
		t.Error("DefaultFieldPatterns should not be empty")
	}

	// Check for some expected patterns
	expectedPatterns := []string{
		"*password*",
		"*secret*",
		"*token*",
		"*key",
	}

	for _, expected := range expectedPatterns {
		found := false
		for _, pattern := range patterns {
			if pattern == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected pattern %q not found in default patterns", expected)
		}
	}
}

func TestDefaultValuePatterns(t *testing.T) {
	patterns := DefaultValuePatterns()

	if len(patterns) == 0 {
		t.Error("DefaultValuePatterns should not be empty")
	}

	// Test that all patterns compile
	for _, vp := range patterns {
		_, err := regexp.Compile(vp.Pattern)
		if err != nil {
			t.Errorf("Pattern %q (%s) failed to compile: %v", vp.Name, vp.Pattern, err)
		}
	}

	// Test specific patterns against known values
	tests := []struct {
		patternName string
		testValue   string
		shouldMatch bool
	}{
		{"AWS Access Key ID", "AKIAIOSFODNN7EXAMPLE", true},
		{"Generic API Key (sk_ prefix)", "sk_test_abc123def456", true},
		{"JWT Token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0In0.abc123", true},
		{"GitHub Personal Access Token", "ghp_1234567890abcdefghijklmnopqrstuv1234", true},
		{"Stripe API Key", "sk_live_abc123def456ghi789jkl012", true},
		{"Google API Key", "AIzaSyDaGmWKa4JsXZ-HjGw7ISLn_3namBGewQe", true},
	}

	for _, tt := range tests {
		t.Run(tt.patternName+" with "+tt.testValue, func(t *testing.T) {
			// Find the pattern
			var pattern *cli.ValuePattern
			for i, vp := range patterns {
				if vp.Name == tt.patternName {
					pattern = &patterns[i]
					break
				}
			}

			if pattern == nil {
				t.Fatalf("Pattern %q not found", tt.patternName)
			}

			regex, err := regexp.Compile(pattern.Pattern)
			if err != nil {
				t.Fatalf("Pattern %q failed to compile: %v", tt.patternName, err)
			}

			matched := regex.MatchString(tt.testValue)
			if matched != tt.shouldMatch {
				t.Errorf("Pattern %q match %q = %v, want %v", tt.patternName, tt.testValue, matched, tt.shouldMatch)
			}
		})
	}
}

func TestDefaultHeaders(t *testing.T) {
	headers := DefaultHeaders()

	if len(headers) == 0 {
		t.Error("DefaultHeaders should not be empty")
	}

	// Check for some expected headers
	expectedHeaders := []string{
		"Authorization",
		"X-API-Key",
		"Cookie",
		"Set-Cookie",
	}

	for _, expected := range expectedHeaders {
		found := false
		for _, header := range headers {
			if header == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected header %q not found in default headers", expected)
		}
	}
}

func TestDefaultSecretsBehavior(t *testing.T) {
	config := DefaultSecretsBehavior()

	if config == nil {
		t.Fatal("DefaultSecretsBehavior should not return nil")
	}

	if !config.Enabled {
		t.Error("Default config should be enabled")
	}

	if config.Masking == nil {
		t.Fatal("Default config should have masking settings")
	}

	if config.Masking.Style != "partial" {
		t.Errorf("Default masking style = %q, want %q", config.Masking.Style, "partial")
	}

	if config.Masking.PartialShowChars != 6 {
		t.Errorf("Default partial show chars = %d, want %d", config.Masking.PartialShowChars, 6)
	}

	if config.MaskIn == nil {
		t.Fatal("Default config should have mask_in settings")
	}

	if !config.MaskIn.Stdout {
		t.Error("Default should mask stdout")
	}

	if !config.MaskIn.Stderr {
		t.Error("Default should mask stderr")
	}

	if !config.MaskIn.Logs {
		t.Error("Default should mask logs")
	}

	if config.MaskIn.Audit {
		t.Error("Default should not mask audit logs")
	}
}

func TestMergeSecretsBehavior(t *testing.T) {
	tests := []struct {
		name     string
		user     *cli.SecretsBehavior
		default_ *cli.SecretsBehavior
		check    func(t *testing.T, result *cli.SecretsBehavior)
	}{
		{
			name:     "nil user config uses defaults",
			user:     nil,
			default_: DefaultSecretsBehavior(),
			check: func(t *testing.T, result *cli.SecretsBehavior) {
				if result == nil {
					t.Error("Result should not be nil")
				}
				if !result.Enabled {
					t.Error("Should use default enabled state")
				}
			},
		},
		{
			name:     "nil default config uses user",
			user:     DefaultSecretsBehavior(),
			default_: nil,
			check: func(t *testing.T, result *cli.SecretsBehavior) {
				if result == nil {
					t.Error("Result should not be nil")
				}
			},
		},
		{
			name: "user patterns extend defaults",
			user: &cli.SecretsBehavior{
				Enabled:       true,
				FieldPatterns: []string{"*custom*"},
			},
			default_: &cli.SecretsBehavior{
				Enabled:       true,
				FieldPatterns: []string{"*password*"},
			},
			check: func(t *testing.T, result *cli.SecretsBehavior) {
				if len(result.FieldPatterns) != 2 {
					t.Errorf("Expected 2 field patterns, got %d", len(result.FieldPatterns))
				}
			},
		},
		{
			name: "user masking overrides defaults",
			user: &cli.SecretsBehavior{
				Enabled: true,
				Masking: &cli.SecretsMasking{
					Style: "full",
				},
			},
			default_: &cli.SecretsBehavior{
				Enabled: true,
				Masking: &cli.SecretsMasking{
					Style: "partial",
				},
			},
			check: func(t *testing.T, result *cli.SecretsBehavior) {
				if result.Masking.Style != "full" {
					t.Errorf("Expected masking style %q, got %q", "full", result.Masking.Style)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MergeSecretsBehavior(tt.user, tt.default_)
			tt.check(t, result)
		})
	}
}

func TestValidateSecretsBehavior(t *testing.T) {
	tests := []struct {
		name    string
		config  *cli.SecretsBehavior
		wantErr bool
	}{
		{
			name:    "nil config is valid",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "default config is valid",
			config:  DefaultSecretsBehavior(),
			wantErr: false,
		},
		{
			name: "invalid masking style",
			config: &cli.SecretsBehavior{
				Enabled: true,
				Masking: &cli.SecretsMasking{
					Style: "invalid",
				},
			},
			wantErr: true,
		},
		{
			name: "negative partial_show_chars",
			config: &cli.SecretsBehavior{
				Enabled: true,
				Masking: &cli.SecretsMasking{
					Style:            "partial",
					PartialShowChars: -1,
				},
			},
			wantErr: true,
		},
		{
			name: "value pattern missing name",
			config: &cli.SecretsBehavior{
				Enabled: true,
				ValuePatterns: []cli.ValuePattern{
					{
						Name:    "",
						Pattern: "test",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "value pattern missing pattern",
			config: &cli.SecretsBehavior{
				Enabled: true,
				ValuePatterns: []cli.ValuePattern{
					{
						Name:    "Test",
						Pattern: "",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecretsBehavior(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecretsBehavior() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "test.field",
		Message: "test error",
	}

	expected := "test.field: test error"
	if err.Error() != expected {
		t.Errorf("ValidationError.Error() = %q, want %q", err.Error(), expected)
	}
}
