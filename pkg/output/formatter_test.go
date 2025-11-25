package output

import (
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
)

func TestNewFormatConfig(t *testing.T) {
	config := NewFormatConfig()

	if !config.Pretty {
		t.Error("Expected Pretty to be true by default")
	}
	if !config.Colors {
		t.Error("Expected Colors to be true by default")
	}
	if !config.ShowHeaders {
		t.Error("Expected ShowHeaders to be true by default")
	}
	if !config.SortAsc {
		t.Error("Expected SortAsc to be true by default")
	}
	if config.MaxWidth != 120 {
		t.Errorf("Expected MaxWidth to be 120, got %d", config.MaxWidth)
	}
	if config.AdditionalData == nil {
		t.Error("Expected AdditionalData to be initialized")
	}
}

func TestFormatConfigChaining(t *testing.T) {
	config := NewFormatConfig().
		WithPretty(false).
		WithColors(false).
		WithCompact(true).
		WithSorting("name", false).
		WithMaxWidth(80).
		WithData("key", "value")

	if config.Pretty {
		t.Error("Expected Pretty to be false")
	}
	if config.Colors {
		t.Error("Expected Colors to be false")
	}
	if !config.Compact {
		t.Error("Expected Compact to be true")
	}
	if config.SortBy != "name" {
		t.Errorf("Expected SortBy to be 'name', got '%s'", config.SortBy)
	}
	if config.SortAsc {
		t.Error("Expected SortAsc to be false")
	}
	if config.MaxWidth != 80 {
		t.Errorf("Expected MaxWidth to be 80, got %d", config.MaxWidth)
	}
	if config.AdditionalData["key"] != "value" {
		t.Error("Expected AdditionalData to contain key=value")
	}
}

func TestFormatConfigWithOutputConfig(t *testing.T) {
	outputConfig := &openapi.CLIOutput{
		Format:         "table",
		SuccessMessage: "Success!",
	}

	config := NewFormatConfig().WithOutputConfig(outputConfig)

	if config.OutputConfig == nil {
		t.Error("Expected OutputConfig to be set")
	}
	if config.OutputConfig.Format != "table" {
		t.Errorf("Expected format to be 'table', got '%s'", config.OutputConfig.Format)
	}
}

func TestNewResult(t *testing.T) {
	data := map[string]string{"key": "value"}
	result := NewResult(data)

	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.Data == nil {
		t.Error("Expected Data to be set")
	}
	if result.Error != "" {
		t.Error("Expected Error to be empty")
	}
	if result.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestNewErrorResult(t *testing.T) {
	err := &testError{"test error"}
	result := NewErrorResult(err)

	if result.Success {
		t.Error("Expected Success to be false")
	}
	if result.Error != "test error" {
		t.Errorf("Expected Error to be 'test error', got '%s'", result.Error)
	}
	if result.Metadata == nil {
		t.Error("Expected Metadata to be initialized")
	}
}

func TestResultChaining(t *testing.T) {
	result := NewResult(map[string]string{"key": "value"}).
		WithMessage("Test message").
		WithMetadata("meta_key", "meta_value")

	if result.Message != "Test message" {
		t.Errorf("Expected Message to be 'Test message', got '%s'", result.Message)
	}
	if result.Metadata["meta_key"] != "meta_value" {
		t.Error("Expected Metadata to contain meta_key=meta_value")
	}
}

// Test helper types
type testError struct {
	message string
}

func (e *testError) Error() string {
	return e.message
}
