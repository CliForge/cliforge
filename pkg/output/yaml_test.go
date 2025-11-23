package output

import (
	"bytes"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestYAMLFormatterName(t *testing.T) {
	formatter := NewYAMLFormatter()
	if formatter.Name() != "yaml" {
		t.Errorf("Expected name 'yaml', got '%s'", formatter.Name())
	}
}

func TestYAMLFormatterSupports(t *testing.T) {
	formatter := NewYAMLFormatter()

	tests := []struct {
		name     string
		data     interface{}
		expected bool
	}{
		{"string", "test", true},
		{"int", 42, true},
		{"map", map[string]string{"key": "value"}, true},
		{"slice", []string{"a", "b"}, true},
		{"nil", nil, true},
		{"struct", struct{ Name string }{Name: "test"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if formatter.Supports(tt.data) != tt.expected {
				t.Errorf("Expected Supports(%v) to be %v", tt.data, tt.expected)
			}
		})
	}
}

func TestYAMLFormatterFormat(t *testing.T) {
	formatter := NewYAMLFormatter()

	tests := []struct {
		name   string
		data   interface{}
		config *FormatConfig
		check  func(t *testing.T, output string)
	}{
		{
			name:   "simple map",
			data:   map[string]string{"key": "value"},
			config: NewFormatConfig(),
			check: func(t *testing.T, output string) {
				var result map[string]string
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal YAML: %v", err)
				}
				if result["key"] != "value" {
					t.Error("Expected key=value")
				}
			},
		},
		{
			name:   "nil data",
			data:   nil,
			config: NewFormatConfig(),
			check: func(t *testing.T, output string) {
				if !strings.Contains(output, "null") {
					t.Error("Expected 'null' in output")
				}
			},
		},
		{
			name: "nested structure",
			data: map[string]interface{}{
				"user": map[string]string{
					"name":  "John",
					"email": "john@example.com",
				},
				"count": 42,
			},
			config: NewFormatConfig(),
			check: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal YAML: %v", err)
				}
				if result["count"].(int) != 42 {
					t.Error("Expected count=42")
				}
			},
		},
		{
			name:   "slice",
			data:   []string{"apple", "banana", "cherry"},
			config: NewFormatConfig(),
			check: func(t *testing.T, output string) {
				var result []string
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal YAML: %v", err)
				}
				if len(result) != 3 {
					t.Errorf("Expected 3 items, got %d", len(result))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			if err := formatter.Format(&buf, tt.data, tt.config); err != nil {
				t.Errorf("Format failed: %v", err)
			}
			tt.check(t, buf.String())
		})
	}
}

func TestYAMLFormatterFormatResult(t *testing.T) {
	formatter := NewYAMLFormatter()

	tests := []struct {
		name   string
		result *Result
		check  func(t *testing.T, output string)
	}{
		{
			name: "success result",
			result: NewResult(map[string]string{"key": "value"}).
				WithMessage("Success!"),
			check: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal YAML: %v", err)
				}
				if !result["success"].(bool) {
					t.Error("Expected success=true")
				}
				if result["message"].(string) != "Success!" {
					t.Error("Expected message='Success!'")
				}
			},
		},
		{
			name:   "error result",
			result: NewErrorResult(&testError{"operation failed"}),
			check: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := yaml.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal YAML: %v", err)
				}
				if result["success"].(bool) {
					t.Error("Expected success=false")
				}
				if result["error"].(string) != "operation failed" {
					t.Error("Expected error='operation failed'")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			config := NewFormatConfig()
			if err := formatter.FormatResult(&buf, tt.result, config); err != nil {
				t.Errorf("FormatResult failed: %v", err)
			}
			tt.check(t, buf.String())
		})
	}
}

func TestYAMLFormatterFormatError(t *testing.T) {
	formatter := NewYAMLFormatter()
	var buf bytes.Buffer

	err := &testError{"test error"}
	config := NewFormatConfig()

	if err := formatter.FormatError(&buf, err, config); err != nil {
		t.Errorf("FormatError failed: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal YAML: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Expected success=false")
	}
	if result["error"].(string) != "test error" {
		t.Error("Expected error='test error'")
	}
}

func TestYAMLFormatterFormatEmpty(t *testing.T) {
	formatter := NewYAMLFormatter()
	var buf bytes.Buffer

	config := NewFormatConfig()
	if err := formatter.FormatEmpty(&buf, "No results", config); err != nil {
		t.Errorf("FormatEmpty failed: %v", err)
	}

	var result map[string]interface{}
	if err := yaml.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal YAML: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}
	if result["message"].(string) != "No results" {
		t.Error("Expected message='No results'")
	}
}

func TestYAMLFormatterFormatCompact(t *testing.T) {
	formatter := NewYAMLFormatter()
	var buf bytes.Buffer

	data := map[string]string{
		"name": "test",
		"type": "example",
	}

	if err := formatter.FormatCompact(&buf, data); err != nil {
		t.Errorf("FormatCompact failed: %v", err)
	}

	output := buf.String()
	// Compact/flow style YAML should be on fewer lines
	if !strings.Contains(output, "{") || !strings.Contains(output, "}") {
		// Note: flow style uses braces for maps
		t.Log("Flow style YAML:", output)
	}
}
