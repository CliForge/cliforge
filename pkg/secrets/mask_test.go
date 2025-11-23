package secrets

import (
	"bytes"
	"strings"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		config *cli.SecretsMasking
		want   string
	}{
		{
			name:  "partial masking default",
			value: "sk_live_abc123def456",
			config: &cli.SecretsMasking{
				Style:            "partial",
				PartialShowChars: 6,
				Replacement:      "***",
			},
			want: "sk_liv***",
		},
		{
			name:  "full masking",
			value: "sk_live_abc123def456",
			config: &cli.SecretsMasking{
				Style:       "full",
				Replacement: "***",
			},
			want: "***",
		},
		{
			name:  "hash masking",
			value: "sk_live_abc123def456",
			config: &cli.SecretsMasking{
				Style: "hash",
			},
			want: "sha256:", // Check prefix only
		},
		{
			name:  "short value partial",
			value: "short",
			config: &cli.SecretsMasking{
				Style:            "partial",
				PartialShowChars: 10,
				Replacement:      "***",
			},
			want: "***",
		},
		{
			name:   "nil config uses default",
			value:  "sk_live_abc123def456",
			config: nil,
			want:   "sk_liv***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaskValue(tt.value, tt.config)

			if tt.config != nil && tt.config.Style == "hash" {
				// For hash, just check the prefix
				if !strings.HasPrefix(got, tt.want) {
					t.Errorf("MaskValue() = %q, want prefix %q", got, tt.want)
				}
			} else {
				if got != tt.want {
					t.Errorf("MaskValue() = %q, want %q", got, tt.want)
				}
			}
		})
	}
}

func TestPartialMask(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		showChars   int
		replacement string
		want        string
	}{
		{
			name:        "normal case",
			value:       "sk_live_abc123",
			showChars:   6,
			replacement: "***",
			want:        "sk_liv***",
		},
		{
			name:        "zero show chars",
			value:       "secret",
			showChars:   0,
			replacement: "***",
			want:        "***",
		},
		{
			name:        "show chars longer than value",
			value:       "abc",
			showChars:   10,
			replacement: "***",
			want:        "***",
		},
		{
			name:        "custom replacement",
			value:       "password123",
			showChars:   4,
			replacement: "[REDACTED]",
			want:        "pass[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := partialMask(tt.value, tt.showChars, tt.replacement)
			if got != tt.want {
				t.Errorf("partialMask() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFullMask(t *testing.T) {
	tests := []struct {
		name        string
		replacement string
		want        string
	}{
		{
			name:        "default replacement",
			replacement: "***",
			want:        "***",
		},
		{
			name:        "custom replacement",
			replacement: "[REDACTED]",
			want:        "[REDACTED]",
		},
		{
			name:        "empty replacement uses default",
			replacement: "",
			want:        "***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fullMask(tt.replacement)
			if got != tt.want {
				t.Errorf("fullMask() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHashMask(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "hash api key",
			value: "sk_live_abc123",
		},
		{
			name:  "hash password",
			value: "password123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hashMask(tt.value)

			// Check format
			if !strings.HasPrefix(got, "sha256:") {
				t.Errorf("hashMask() = %q, should start with 'sha256:'", got)
			}

			// Check length (sha256: + 16 hex chars = 23 chars)
			if len(got) != 23 {
				t.Errorf("hashMask() length = %d, want 23", len(got))
			}

			// Hash should be deterministic
			got2 := hashMask(tt.value)
			if got != got2 {
				t.Errorf("hashMask() not deterministic: %q != %q", got, got2)
			}

			// Different values should produce different hashes
			different := hashMask(tt.value + "x")
			if got == different {
				t.Errorf("hashMask() same for different values")
			}
		})
	}
}

func TestMaskingWriter(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled: true,
		ValuePatterns: []cli.ValuePattern{
			{
				Name:    "Test Key",
				Pattern: `sk_test_[a-z0-9]+`,
				Enabled: true,
			},
		},
		Masking: &cli.SecretsMasking{
			Style:            "partial",
			PartialShowChars: 6,
			Replacement:      "***",
		},
		MaskIn: &cli.SecretsMaskIn{
			Stdout: true,
		},
	}

	detector, err := NewDetector(config)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	buf := &bytes.Buffer{}
	writer := NewMaskingWriter(detector, "stdout", buf)

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, output string)
	}{
		{
			name:  "mask secret in output",
			input: "Your API key is sk_test_abc123def456",
			check: func(t *testing.T, output string) {
				if strings.Contains(output, "sk_test_abc123def456") {
					t.Error("Secret should be masked in output")
				}
				if !strings.Contains(output, "sk_tes***") {
					t.Errorf("Expected masked value not found in output: %q", output)
				}
			},
		},
		{
			name:  "no secrets in output",
			input: "This is a normal message",
			check: func(t *testing.T, output string) {
				if output != "This is a normal message" {
					t.Errorf("Normal message should not be modified: %q", output)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			_, err := writer.Write([]byte(tt.input))
			if err != nil {
				t.Fatalf("Write failed: %v", err)
			}

			output := buf.String()
			tt.check(t, output)
		})
	}
}

func TestMaskFormatter_FormatTable(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled:       true,
		FieldPatterns: []string{"*password*", "*token*"},
		Masking: &cli.SecretsMasking{
			Style:            "partial",
			PartialShowChars: 6,
			Replacement:      "***",
		},
	}

	detector, err := NewDetector(config)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	formatter := NewMaskFormatter(detector)

	rows := []map[string]interface{}{
		{
			"name":     "Alice",
			"password": "alice_secret_123",
			"email":    "alice@example.com",
		},
		{
			"name":     "Bob",
			"password": "bob_secret_456",
			"email":    "bob@example.com",
		},
	}

	masked := formatter.FormatTable(rows)

	// Check that passwords are masked
	if masked[0]["password"] == "alice_secret_123" {
		t.Error("Alice's password should be masked")
	}
	if masked[1]["password"] == "bob_secret_456" {
		t.Error("Bob's password should be masked")
	}

	// Check that other fields are not masked
	if masked[0]["name"] != "Alice" {
		t.Error("Alice's name should not be masked")
	}
	if masked[1]["email"] != "bob@example.com" {
		t.Error("Bob's email should not be masked")
	}
}

func TestMaskStrategy(t *testing.T) {
	tests := []struct {
		name     string
		strategy MaskStrategy
		value    string
		check    func(t *testing.T, result string)
	}{
		{
			name:     "partial strategy",
			strategy: NewPartialMaskStrategy(6, "***"),
			value:    "sk_live_abc123",
			check: func(t *testing.T, result string) {
				if result != "sk_liv***" {
					t.Errorf("Expected %q, got %q", "sk_liv***", result)
				}
			},
		},
		{
			name:     "full strategy",
			strategy: NewFullMaskStrategy("***"),
			value:    "sk_live_abc123",
			check: func(t *testing.T, result string) {
				if result != "***" {
					t.Errorf("Expected %q, got %q", "***", result)
				}
			},
		},
		{
			name:     "hash strategy",
			strategy: NewHashMaskStrategy(),
			value:    "sk_live_abc123",
			check: func(t *testing.T, result string) {
				if !strings.HasPrefix(result, "sha256:") {
					t.Errorf("Expected sha256 hash, got %q", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.strategy.Mask(tt.value)
			tt.check(t, result)

			// Check that strategy name is set
			if tt.strategy.Name() == "" {
				t.Error("Strategy name should not be empty")
			}
		})
	}
}

func TestCreateMaskStrategy(t *testing.T) {
	tests := []struct {
		name         string
		config       *cli.SecretsMasking
		expectedType string
	}{
		{
			name: "partial strategy",
			config: &cli.SecretsMasking{
				Style:            "partial",
				PartialShowChars: 6,
				Replacement:      "***",
			},
			expectedType: "partial",
		},
		{
			name: "full strategy",
			config: &cli.SecretsMasking{
				Style:       "full",
				Replacement: "***",
			},
			expectedType: "full",
		},
		{
			name: "hash strategy",
			config: &cli.SecretsMasking{
				Style: "hash",
			},
			expectedType: "hash",
		},
		{
			name:         "nil config defaults to partial",
			config:       nil,
			expectedType: "partial",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy := CreateMaskStrategy(tt.config)
			if strategy == nil {
				t.Fatal("CreateMaskStrategy should not return nil")
			}

			if strategy.Name() != tt.expectedType {
				t.Errorf("Expected strategy type %q, got %q", tt.expectedType, strategy.Name())
			}
		})
	}
}

func TestStructurePreservingMask(t *testing.T) {
	config := &cli.SecretsMasking{
		Style:            "partial",
		PartialShowChars: 6,
		Replacement:      "***",
	}

	tests := []struct {
		name  string
		value string
		check func(t *testing.T, got string)
	}{
		{
			name:  "underscore separated",
			value: "sk_live_abc123def456",
			check: func(t *testing.T, got string) {
				// Should preserve separator and mask the rest
				if !strings.Contains(got, "***") {
					t.Error("Should contain masked portion")
				}
				if got == "sk_live_abc123def456" {
					t.Error("Should be masked")
				}
			},
		},
		{
			name:  "hyphen separated",
			value: "key-live-abc123def456",
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "***") {
					t.Error("Should contain masked portion")
				}
				if got == "key-live-abc123def456" {
					t.Error("Should be masked")
				}
			},
		},
		{
			name:  "dot separated",
			value: "key.live.abc123def456",
			check: func(t *testing.T, got string) {
				if !strings.Contains(got, "***") {
					t.Error("Should contain masked portion")
				}
				if got == "key.live.abc123def456" {
					t.Error("Should be masked")
				}
			},
		},
		{
			name:  "no separator",
			value: "plainkey12345",
			check: func(t *testing.T, got string) {
				if got != "plaink***" {
					t.Errorf("Expected %q, got %q", "plaink***", got)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StructurePreservingMask(tt.value, config)
			tt.check(t, got)
		})
	}
}
