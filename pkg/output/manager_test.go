package output

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()

	if manager == nil {
		t.Error("Expected manager to be created")
	}
	if manager.formatters == nil {
		t.Error("Expected formatters map to be initialized")
	}
	if manager.defaultFormat != "json" {
		t.Errorf("Expected default format 'json', got '%s'", manager.defaultFormat)
	}
	if manager.config == nil {
		t.Error("Expected config to be initialized")
	}
	if manager.templateEngine == nil {
		t.Error("Expected templateEngine to be initialized")
	}

	// Check default formatters are registered
	formats := manager.GetSupportedFormats()
	expectedFormats := []string{"json", "yaml", "table"}
	for _, expected := range expectedFormats {
		found := false
		for _, format := range formats {
			if format == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected format '%s' to be registered", expected)
		}
	}
}

func TestManagerRegisterFormatter(t *testing.T) {
	manager := NewManager()

	// Create a custom formatter
	customFormatter := &mockFormatter{name: "custom"}
	manager.RegisterFormatter(customFormatter)

	formatter, err := manager.GetFormatter("custom")
	if err != nil {
		t.Errorf("GetFormatter failed: %v", err)
	}
	if formatter.Name() != "custom" {
		t.Error("Expected to retrieve registered formatter")
	}
}

func TestManagerGetFormatter(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name        string
		format      string
		shouldError bool
	}{
		{"json", "json", false},
		{"yaml", "yaml", false},
		{"table", "table", false},
		{"nonexistent", "nonexistent", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := manager.GetFormatter(tt.format)
			if tt.shouldError && err == nil {
				t.Error("Expected error for non-existent formatter")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestManagerSetDefaultFormat(t *testing.T) {
	manager := NewManager()
	manager.SetDefaultFormat("yaml")

	if manager.defaultFormat != "yaml" {
		t.Errorf("Expected default format 'yaml', got '%s'", manager.defaultFormat)
	}
}

func TestManagerFormat(t *testing.T) {
	manager := NewManager()

	data := map[string]string{"key": "value"}
	var buf bytes.Buffer

	if err := manager.Format(&buf, data, "json"); err != nil {
		t.Errorf("Format failed: %v", err)
	}

	if !strings.Contains(buf.String(), "key") {
		t.Error("Expected output to contain 'key'")
	}
}

func TestManagerFormatWithConfig(t *testing.T) {
	manager := NewManager()

	data := map[string]string{"key": "value"}
	config := NewFormatConfig().WithPretty(true)
	var buf bytes.Buffer

	if err := manager.FormatWithConfig(&buf, data, "json", config); err != nil {
		t.Errorf("FormatWithConfig failed: %v", err)
	}

	if !strings.Contains(buf.String(), "key") {
		t.Error("Expected output to contain 'key'")
	}
}

func TestManagerFormatResult(t *testing.T) {
	manager := NewManager()

	result := NewResult(map[string]string{"key": "value"}).
		WithMessage("Success!")

	var buf bytes.Buffer
	if err := manager.FormatResult(&buf, result, "json"); err != nil {
		t.Errorf("FormatResult failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "success") {
		t.Error("Expected output to contain 'success'")
	}
}

func TestManagerFormatError(t *testing.T) {
	manager := NewManager()

	testErr := &testError{"test error"}
	var buf bytes.Buffer

	if err := manager.FormatError(&buf, testErr, "json"); err != nil {
		t.Errorf("FormatError failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "test error") {
		t.Error("Expected output to contain error message")
	}
}

func TestManagerFormatEmpty(t *testing.T) {
	manager := NewManager()

	var buf bytes.Buffer
	if err := manager.FormatEmpty(&buf, "No results", "json"); err != nil {
		t.Errorf("FormatEmpty failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "No results") {
		t.Error("Expected output to contain message")
	}
}

func TestManagerFormatSuccess(t *testing.T) {
	manager := NewManager()

	data := map[string]string{"id": "123"}
	var buf bytes.Buffer

	if err := manager.FormatSuccess(&buf, "Created successfully", data, "json"); err != nil {
		t.Errorf("FormatSuccess failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Created successfully") {
		t.Error("Expected output to contain success message")
	}
}

func TestManagerSelectFormat(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name           string
		explicit       string
		outputConfig   *openapi.CLIOutput
		globalConfig   *openapi.OutputSettings
		expectedFormat string
	}{
		{
			name:           "explicit format",
			explicit:       "yaml",
			outputConfig:   &openapi.CLIOutput{Format: "table"},
			globalConfig:   &openapi.OutputSettings{DefaultFormat: "json"},
			expectedFormat: "yaml",
		},
		{
			name:           "output config",
			explicit:       "",
			outputConfig:   &openapi.CLIOutput{Format: "table"},
			globalConfig:   &openapi.OutputSettings{DefaultFormat: "json"},
			expectedFormat: "table",
		},
		{
			name:           "global config",
			explicit:       "",
			outputConfig:   nil,
			globalConfig:   &openapi.OutputSettings{DefaultFormat: "yaml"},
			expectedFormat: "yaml",
		},
		{
			name:           "default",
			explicit:       "",
			outputConfig:   nil,
			globalConfig:   nil,
			expectedFormat: "json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format := manager.SelectFormat(tt.explicit, tt.outputConfig, tt.globalConfig)
			if format != tt.expectedFormat {
				t.Errorf("Expected format '%s', got '%s'", tt.expectedFormat, format)
			}
		})
	}
}

func TestManagerApplyOutputRules(t *testing.T) {
	manager := NewManager()

	outputConfig := &openapi.CLIOutput{
		Format:         "table",
		SuccessMessage: "Done!",
	}

	config := manager.ApplyOutputRules(outputConfig)

	if config.OutputConfig == nil {
		t.Error("Expected OutputConfig to be set")
	}
	if config.OutputConfig.Format != "table" {
		t.Error("Expected format to be 'table'")
	}
}

func TestManagerIsFormatSupported(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		format   string
		expected bool
	}{
		{"json", true},
		{"yaml", true},
		{"table", true},
		{"xml", false},
		{"csv", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			if manager.IsFormatSupported(tt.format) != tt.expected {
				t.Errorf("Expected IsFormatSupported('%s') to be %v", tt.format, tt.expected)
			}
		})
	}
}

func TestManagerGetSupportedFormats(t *testing.T) {
	manager := NewManager()
	formats := manager.GetSupportedFormats()

	if len(formats) < 3 {
		t.Error("Expected at least 3 supported formats")
	}

	// Check for default formats
	hasJSON := false
	hasYAML := false
	hasTable := false

	for _, format := range formats {
		switch format {
		case "json":
			hasJSON = true
		case "yaml":
			hasYAML = true
		case "table":
			hasTable = true
		}
	}

	if !hasJSON || !hasYAML || !hasTable {
		t.Error("Expected default formats to be supported")
	}
}

func TestManagerValidateFormat(t *testing.T) {
	manager := NewManager()

	tests := []struct {
		name        string
		format      string
		data        interface{}
		shouldError bool
	}{
		{"json with map", "json", map[string]string{"key": "value"}, false},
		{"yaml with slice", "yaml", []string{"a", "b"}, false},
		{"table with slice", "table", []string{"a", "b"}, false},
		{"table with nil", "table", nil, true},
		{"nonexistent format", "xml", map[string]string{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := manager.ValidateFormat(tt.format, tt.data)
			if tt.shouldError && err == nil {
				t.Error("Expected validation error")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

func TestManagerRenderMessage(t *testing.T) {
	manager := NewManager()

	template := "Hello {name}"
	data := map[string]interface{}{"name": "World"}

	result, err := manager.RenderMessage(template, data)
	if err != nil {
		t.Errorf("RenderMessage failed: %v", err)
	}

	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", result)
	}
}

func TestManagerRenderSuccessMessage(t *testing.T) {
	manager := NewManager()

	outputConfig := &openapi.CLIOutput{
		SuccessMessage: "Cluster {name} created successfully",
	}
	data := map[string]interface{}{"name": "test-cluster"}

	result, err := manager.RenderSuccessMessage(outputConfig, data)
	if err != nil {
		t.Errorf("RenderSuccessMessage failed: %v", err)
	}

	if !strings.Contains(result, "test-cluster") {
		t.Errorf("Expected result to contain 'test-cluster', got '%s'", result)
	}
}

func TestManagerRenderErrorMessage(t *testing.T) {
	manager := NewManager()

	outputConfig := &openapi.CLIOutput{
		ErrorMessage: "Failed to create {resource}: {error}",
	}
	data := map[string]interface{}{
		"resource": "cluster",
		"error":    "quota exceeded",
	}

	result, err := manager.RenderErrorMessage(outputConfig, data)
	if err != nil {
		t.Errorf("RenderErrorMessage failed: %v", err)
	}

	if !strings.Contains(result, "quota exceeded") {
		t.Errorf("Expected result to contain 'quota exceeded', got '%s'", result)
	}
}

func TestManagerFormatWithTemplate(t *testing.T) {
	manager := NewManager()

	data := map[string]string{"id": "123"}
	template := "Resource {id} created"
	templateData := map[string]interface{}{"id": "123"}

	var buf bytes.Buffer
	if err := manager.FormatWithTemplate(&buf, data, template, templateData, "json"); err != nil {
		t.Errorf("FormatWithTemplate failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Resource 123 created") {
		t.Log("Output:", output)
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// Test that package-level functions work
	data := map[string]string{"test": "value"}
	var buf bytes.Buffer

	if err := Format(&buf, data, "json"); err != nil {
		t.Errorf("Format failed: %v", err)
	}

	if !strings.Contains(buf.String(), "test") {
		t.Error("Expected output to contain 'test'")
	}
}

// Mock formatter for testing
type mockFormatter struct {
	name string
}

func (m *mockFormatter) Name() string {
	return m.name
}

func (m *mockFormatter) Supports(data interface{}) bool {
	return true
}

func (m *mockFormatter) Format(w io.Writer, data interface{}, config *FormatConfig) error {
	return nil
}
