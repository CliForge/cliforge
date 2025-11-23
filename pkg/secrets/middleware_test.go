package secrets

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestHTTPMiddleware_MaskRequest(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled: true,
		Headers: []string{"Authorization", "X-API-Key"},
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

	middleware := NewHTTPMiddleware(detector, nil)

	req, err := http.NewRequest("GET", "https://api.example.com/users", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer secret_token_12345")
	req.Header.Set("X-API-Key", "my_api_key_67890")
	req.Header.Set("Content-Type", "application/json")

	masked := middleware.MaskRequest(req)

	if masked == nil {
		t.Fatal("MaskRequest should not return nil")
	}

	// Check that method and URL are preserved
	if masked.Method != "GET" {
		t.Errorf("Method = %q, want %q", masked.Method, "GET")
	}
	if masked.URL != "https://api.example.com/users" {
		t.Errorf("URL = %q, want %q", masked.URL, "https://api.example.com/users")
	}

	// Check headers exist
	if len(masked.Header) == 0 {
		t.Fatal("Headers should not be empty")
	}

	// Check that secret headers are masked (note: may have different casing)
	foundAuth := false
	foundAPIKey := false
	for key, values := range masked.Header {
		lowerKey := strings.ToLower(key)
		if lowerKey == "authorization" {
			foundAuth = true
			if len(values) > 0 && values[0] == "Bearer secret_token_12345" {
				t.Error("Authorization header should be masked")
			}
		}
		if lowerKey == "x-api-key" {
			foundAPIKey = true
			if len(values) > 0 && values[0] == "my_api_key_67890" {
				t.Error("X-API-Key header should be masked")
			}
		}
		if lowerKey == "content-type" {
			if len(values) > 0 && values[0] != "application/json" {
				t.Error("Content-Type should not be masked")
			}
		}
	}

	if !foundAuth {
		t.Error("Authorization header should be present")
	}
	if !foundAPIKey {
		t.Error("X-API-Key header should be present")
	}

	// Check method and URL are preserved
	if masked.Method != "GET" {
		t.Errorf("Method = %q, want %q", masked.Method, "GET")
	}
	if masked.URL != "https://api.example.com/users" {
		t.Errorf("URL = %q, want %q", masked.URL, "https://api.example.com/users")
	}
}

func TestHTTPMiddleware_MaskRequestBody(t *testing.T) {
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

	middleware := NewHTTPMiddleware(detector, nil)

	tests := []struct {
		name        string
		body        string
		contentType string
		check       func(t *testing.T, masked []byte)
	}{
		{
			name:        "JSON with secrets",
			body:        `{"username":"john","password":"secret123"}`,
			contentType: "application/json",
			check: func(t *testing.T, masked []byte) {
				if strings.Contains(string(masked), "secret123") {
					t.Error("Password should be masked in JSON body")
				}
			},
		},
		{
			name:        "plain text",
			body:        "username=john&password=secret123",
			contentType: "text/plain",
			check: func(t *testing.T, masked []byte) {
				// Plain text should still be processed through string masking
				maskedStr := string(masked)
				_ = maskedStr // Would be masked by value patterns if they match
			},
		},
		{
			name:        "invalid JSON",
			body:        `{invalid json}`,
			contentType: "application/json",
			check: func(t *testing.T, masked []byte) {
				// Should return original if invalid JSON
				if string(masked) != `{invalid json}` {
					t.Error("Invalid JSON should be returned as-is")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			masked, err := middleware.MaskRequestBody([]byte(tt.body), tt.contentType)
			if err != nil {
				t.Fatalf("MaskRequestBody failed: %v", err)
			}
			tt.check(t, masked)
		})
	}
}

func TestHTTPMiddleware_MaskResponse(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled: true,
		Headers: []string{"Set-Cookie"},
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

	middleware := NewHTTPMiddleware(detector, nil)

	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Set-Cookie":   []string{"session_id=abc123def456"},
		},
	}

	masked, err := middleware.MaskResponse(resp)
	if err != nil {
		t.Fatalf("MaskResponse failed: %v", err)
	}

	// Check that secret headers are masked
	if masked.Header["Set-Cookie"][0] == "session_id=abc123def456" {
		t.Error("Set-Cookie header should be masked")
	}

	// Check that normal headers are not masked
	if masked.Header["Content-Type"][0] != "application/json" {
		t.Error("Content-Type should not be masked")
	}

	// Check status is preserved
	if masked.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want %d", masked.StatusCode, 200)
	}
}

func TestDebugLogger_LogRequest(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled:       true,
		Headers:       []string{"Authorization"},
		FieldPatterns: []string{"*password*"},
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

	buf := &bytes.Buffer{}
	logger := NewDebugLogger(detector, buf)

	req, err := http.NewRequest("POST", "https://api.example.com/login",
		bytes.NewBufferString(`{"username":"john","password":"secret123"}`))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer secret_token")
	req.Header.Set("Content-Type", "application/json")

	err = logger.LogRequest(req)
	if err != nil {
		t.Fatalf("LogRequest failed: %v", err)
	}

	output := buf.String()

	// Check that request is logged
	if !strings.Contains(output, "POST") {
		t.Error("Method should be logged")
	}
	if !strings.Contains(output, "https://api.example.com/login") {
		t.Error("URL should be logged")
	}

	// Check that secret header is masked
	if strings.Contains(output, "secret_token") {
		t.Error("Authorization token should be masked in logs")
	}

	// Check that password in body is masked
	if strings.Contains(output, "secret123") {
		t.Error("Password should be masked in request body")
	}
}

func TestDebugLogger_LogResponse(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled:       true,
		FieldPatterns: []string{"*token*"},
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

	buf := &bytes.Buffer{}
	logger := NewDebugLogger(detector, buf)

	resp := &http.Response{
		StatusCode: 200,
		Status:     "200 OK",
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		Body: io.NopCloser(bytes.NewBufferString(`{"user_id":"123","access_token":"secret_token_12345"}`)),
	}

	err = logger.LogResponse(resp)
	if err != nil {
		t.Fatalf("LogResponse failed: %v", err)
	}

	output := buf.String()

	// Check that response is logged
	if !strings.Contains(output, "200 OK") {
		t.Error("Status should be logged")
	}

	// Check that token in body is masked
	if strings.Contains(output, "secret_token_12345") {
		t.Error("Access token should be masked in response body")
	}

	// Check that user_id is not masked
	if !strings.Contains(output, "123") {
		t.Error("User ID should not be masked")
	}
}

func TestDebugLogger_Disabled(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled: true,
	}

	detector, err := NewDetector(config)
	if err != nil {
		t.Fatalf("Failed to create detector: %v", err)
	}

	buf := &bytes.Buffer{}
	logger := NewDebugLogger(detector, buf)
	logger.SetEnabled(false)

	req, err := http.NewRequest("GET", "https://api.example.com/users", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	err = logger.LogRequest(req)
	if err != nil {
		t.Fatalf("LogRequest failed: %v", err)
	}

	// When disabled, nothing should be logged
	if buf.Len() > 0 {
		t.Error("Nothing should be logged when logger is disabled")
	}
}

func TestFormatHeaders(t *testing.T) {
	tests := []struct {
		name    string
		headers map[string][]string
		check   func(t *testing.T, result string)
	}{
		{
			name:    "empty headers",
			headers: map[string][]string{},
			check: func(t *testing.T, result string) {
				if result != "{}" {
					t.Errorf("Expected %q, got %q", "{}", result)
				}
			},
		},
		{
			name: "single value",
			headers: map[string][]string{
				"Content-Type": {"application/json"},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Content-Type") {
					t.Error("Result should contain header name")
				}
				if !strings.Contains(result, "application/json") {
					t.Error("Result should contain header value")
				}
			},
		},
		{
			name: "multiple values",
			headers: map[string][]string{
				"Accept": {"application/json", "text/html"},
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Accept") {
					t.Error("Result should contain header name")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatHeaders(tt.headers)
			tt.check(t, result)
		})
	}
}

func TestResponseBodyReader(t *testing.T) {
	config := &cli.SecretsBehavior{
		Enabled:       true,
		FieldPatterns: []string{"*password*"},
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

	body := `{"username":"john","password":"secret123"}`
	bodyReader := io.NopCloser(bytes.NewBufferString(body))

	reader := NewResponseBodyReader(detector, "application/json", bodyReader)

	// Read the body
	result, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read body: %v", err)
	}

	// Check that password is masked
	if strings.Contains(string(result), "secret123") {
		t.Error("Password should be masked in response body")
	}

	// Check that username is not masked
	if !strings.Contains(string(result), "john") {
		t.Error("Username should not be masked")
	}

	// Close the reader
	err = reader.Close()
	if err != nil {
		t.Fatalf("Failed to close reader: %v", err)
	}
}
