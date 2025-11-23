package secrets

import (
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestNewDetector(t *testing.T) {
	tests := []struct {
		name    string
		config  *cli.SecretsBehavior
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: false,
		},
		{
			name:    "default config",
			config:  DefaultSecretsBehavior(),
			wantErr: false,
		},
		{
			name: "custom field pattern",
			config: &cli.SecretsBehavior{
				Enabled: true,
				FieldPatterns: []string{
					"*custom*",
				},
			},
			wantErr: false,
		},
		{
			name: "invalid value pattern",
			config: &cli.SecretsBehavior{
				Enabled: true,
				ValuePatterns: []cli.ValuePattern{
					{
						Name:    "Invalid",
						Pattern: "[invalid(",
						Enabled: true,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewDetector(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDetector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDetector_IsSecretField(t *testing.T) {
	detector, err := NewDetector(DefaultSecretsBehavior())
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	tests := []struct {
		name      string
		fieldName string
		want      bool
	}{
		{"password field", "password", true},
		{"api_key field", "api_key", true},
		{"secret field", "secret_token", true},
		{"bearer field", "bearer_auth", true},
		{"normal field", "username", false},
		{"email field", "email", false},
		{"case insensitive", "PASSWORD", true},
		{"case insensitive key", "API_KEY", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.IsSecretField(tt.fieldName)
			if got != tt.want {
				t.Errorf("IsSecretField(%q) = %v, want %v", tt.fieldName, got, tt.want)
			}
		})
	}
}

func TestDetector_IsSecretValue(t *testing.T) {
	detector, err := NewDetector(DefaultSecretsBehavior())
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	tests := []struct {
		name        string
		value       string
		wantSecret  bool
		wantPattern string
	}{
		{
			name:        "AWS access key",
			value:       "AKIAIOSFODNN7EXAMPLE",
			wantSecret:  true,
			wantPattern: "AWS Access Key ID",
		},
		{
			name:        "Generic API key",
			value:       "sk_live_abc123def456ghi789jkl012",
			wantSecret:  true,
			wantPattern: "Generic API Key (sk_ prefix)",
		},
		{
			name:        "JWT token",
			value:       "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantSecret:  true,
			wantPattern: "JWT Token",
		},
		{
			name:        "GitHub token",
			value:       "ghp_1234567890abcdefghijklmnopqrstuv",
			wantSecret:  true,
			wantPattern: "GitHub Personal Access Token",
		},
		{
			name:       "normal value",
			value:      "john.doe@example.com",
			wantSecret: false,
		},
		{
			name:       "normal string",
			value:      "hello world",
			wantSecret: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSecret, gotPattern := detector.IsSecretValue(tt.value)
			if gotSecret != tt.wantSecret {
				t.Errorf("IsSecretValue(%q) secret = %v, want %v", tt.value, gotSecret, tt.wantSecret)
			}
			if tt.wantSecret && gotPattern != tt.wantPattern {
				t.Errorf("IsSecretValue(%q) pattern = %q, want %q", tt.value, gotPattern, tt.wantPattern)
			}
		})
	}
}

func TestDetector_IsSecretHeader(t *testing.T) {
	detector, err := NewDetector(DefaultSecretsBehavior())
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	tests := []struct {
		name       string
		headerName string
		want       bool
	}{
		{"Authorization", "Authorization", true},
		{"X-API-Key", "X-API-Key", true},
		{"Cookie", "Cookie", true},
		{"Case insensitive", "authorization", true},
		{"Normal header", "Content-Type", false},
		{"Normal header", "Accept", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detector.IsSecretHeader(tt.headerName)
			if got != tt.want {
				t.Errorf("IsSecretHeader(%q) = %v, want %v", tt.headerName, got, tt.want)
			}
		})
	}
}

func TestDetector_MaskJSON(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled:       true,
		FieldPatterns: []string{"*password*", "*token*", "*key"},
		ValuePatterns: []cli.ValuePattern{
			{
				Name:    "Test API Key",
				Pattern: `sk_test_[a-z0-9]+`,
				Enabled: true,
			},
		},
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

	tests := []struct {
		name  string
		input map[string]interface{}
		check func(t *testing.T, result interface{})
	}{
		{
			name: "mask password field",
			input: map[string]interface{}{
				"username": "john",
				"password": "supersecret123",
			},
			check: func(t *testing.T, result interface{}) {
				m := result.(map[string]interface{})
				if m["username"] != "john" {
					t.Errorf("username should not be masked")
				}
				if m["password"] == "supersecret123" {
					t.Errorf("password should be masked")
				}
			},
		},
		{
			name: "mask api key value",
			input: map[string]interface{}{
				"user_id": "123",
				"api_key": "sk_test_abcdefghijklmnop",
			},
			check: func(t *testing.T, result interface{}) {
				m := result.(map[string]interface{})
				if m["user_id"] != "123" {
					t.Errorf("user_id should not be masked")
				}
				if m["api_key"] == "sk_test_abcdefghijklmnop" {
					t.Errorf("api_key should be masked")
				}
			},
		},
		{
			name: "nested objects",
			input: map[string]interface{}{
				"user": map[string]interface{}{
					"name":     "John",
					"password": "secret123",
				},
			},
			check: func(t *testing.T, result interface{}) {
				m := result.(map[string]interface{})
				user := m["user"].(map[string]interface{})
				if user["name"] != "John" {
					t.Errorf("name should not be masked")
				}
				if user["password"] == "secret123" {
					t.Errorf("password should be masked")
				}
			},
		},
		{
			name: "arrays",
			input: map[string]interface{}{
				"users": []interface{}{
					map[string]interface{}{
						"name":     "Alice",
						"password": "pass123",
					},
					map[string]interface{}{
						"name":     "Bob",
						"password": "pass456",
					},
				},
			},
			check: func(t *testing.T, result interface{}) {
				m := result.(map[string]interface{})
				users := m["users"].([]interface{})
				user1 := users[0].(map[string]interface{})
				user2 := users[1].(map[string]interface{})

				if user1["name"] != "Alice" {
					t.Errorf("Alice name should not be masked")
				}
				if user1["password"] == "pass123" {
					t.Errorf("Alice password should be masked")
				}
				if user2["name"] != "Bob" {
					t.Errorf("Bob name should not be masked")
				}
				if user2["password"] == "pass456" {
					t.Errorf("Bob password should be masked")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.MaskJSON(tt.input)
			tt.check(t, result)
		})
	}
}

func TestDetector_MaskString(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled: true,
		ValuePatterns: []cli.ValuePattern{
			{
				Name:    "Test API Key",
				Pattern: `sk_test_[a-z0-9]+`,
				Enabled: true,
			},
		},
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

	tests := []struct {
		name  string
		input string
		check func(t *testing.T, result string)
	}{
		{
			name:  "mask api key in text",
			input: "Your API key is sk_test_abcdefghijklmnop for testing",
			check: func(t *testing.T, result string) {
				if result == "Your API key is sk_test_abcdefghijklmnop for testing" {
					t.Errorf("API key should be masked in text")
				}
				if result != "Your API key is sk_tes*** for testing" {
					t.Errorf("Unexpected masking: %s", result)
				}
			},
		},
		{
			name:  "no secrets",
			input: "This is a normal message",
			check: func(t *testing.T, result string) {
				if result != "This is a normal message" {
					t.Errorf("Normal text should not be modified")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detector.MaskString(tt.input)
			tt.check(t, result)
		})
	}
}

func TestDetector_MaskHeaders(t *testing.T) {
	detector, err := NewDetector(DefaultSecretsBehavior())
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	headers := map[string][]string{
		"Content-Type":  {"application/json"},
		"Authorization": {"Bearer secret_token_12345"},
		"X-API-Key":     {"my_api_key_67890"},
		"User-Agent":    {"CliForge/1.0"},
	}

	masked := detector.MaskHeaders(headers)

	// Check that normal headers are not masked
	if masked["Content-Type"][0] != "application/json" {
		t.Errorf("Content-Type should not be masked")
	}
	if masked["User-Agent"][0] != "CliForge/1.0" {
		t.Errorf("User-Agent should not be masked")
	}

	// Check that secret headers are masked
	if masked["Authorization"][0] == "Bearer secret_token_12345" {
		t.Errorf("Authorization should be masked")
	}
	if masked["X-API-Key"][0] == "my_api_key_67890" {
		t.Errorf("X-API-Key should be masked")
	}
}

func TestDetector_ShouldMaskInContext(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled: true,
		MaskIn: &cli.SecretsMaskIn{
			Stdout:      true,
			Stderr:      true,
			Logs:        true,
			Audit:       false,
			DebugOutput: true,
		},
	}

	detector, err := NewDetector(config)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	tests := []struct {
		context string
		want    bool
	}{
		{"stdout", true},
		{"stderr", true},
		{"logs", true},
		{"debug", true},
		{"audit", false},
		{"unknown", true}, // Mask by default in unknown contexts
	}

	for _, tt := range tests {
		t.Run(tt.context, func(t *testing.T) {
			got := detector.ShouldMaskInContext(tt.context)
			if got != tt.want {
				t.Errorf("ShouldMaskInContext(%q) = %v, want %v", tt.context, got, tt.want)
			}
		})
	}
}

func TestDetector_DisabledDetector(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled: false,
	}

	detector, err := NewDetector(config)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	// All detection should return false when disabled
	if detector.IsSecretField("password") {
		t.Errorf("Disabled detector should not detect field secrets")
	}

	if secret, _ := detector.IsSecretValue("sk_test_abc123"); secret {
		t.Errorf("Disabled detector should not detect value secrets")
	}

	if detector.IsSecretHeader("Authorization") {
		t.Errorf("Disabled detector should not detect header secrets")
	}

	// Masking should be no-op when disabled
	input := map[string]interface{}{
		"password": "secret123",
	}
	result := detector.MaskJSON(input)
	resultMap := result.(map[string]interface{})
	if resultMap["password"] != "secret123" {
		t.Errorf("Disabled detector should not mask values")
	}
}

func TestGlobToRegex(t *testing.T) {
	tests := []struct {
		pattern   string
		testStr   string
		wantMatch bool
	}{
		{"*password*", "my_password", true},
		{"*password*", "PASSWORD", true}, // case insensitive
		{"*password*", "username", false},
		{"*key", "api_key", true},
		{"*key", "apikey", false},
		{"auth", "auth", true},
		{"auth", "oauth", false},
		{"*token*", "refresh_token", true},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+" vs "+tt.testStr, func(t *testing.T) {
			regex, err := globToRegex(tt.pattern)
			if err != nil {
				t.Fatalf("Failed to compile pattern %q: %v", tt.pattern, err)
			}

			got := regex.MatchString(tt.testStr)
			if got != tt.wantMatch {
				t.Errorf("Pattern %q match %q = %v, want %v", tt.pattern, tt.testStr, got, tt.wantMatch)
			}
		})
	}
}
