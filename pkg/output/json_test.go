package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestJSONFormatterName(t *testing.T) {
	formatter := NewJSONFormatter()
	if formatter.Name() != "json" {
		t.Errorf("Expected name 'json', got '%s'", formatter.Name())
	}
}

func TestJSONFormatterSupports(t *testing.T) {
	formatter := NewJSONFormatter()

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

func TestJSONFormatterFormat(t *testing.T) {
	formatter := NewJSONFormatter()

	tests := []struct {
		name   string
		data   interface{}
		config *FormatConfig
		check  func(t *testing.T, output string)
	}{
		{
			name:   "simple map",
			data:   map[string]string{"key": "value"},
			config: NewFormatConfig().WithPretty(true),
			check: func(t *testing.T, output string) {
				var result map[string]string
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
				if result["key"] != "value" {
					t.Error("Expected key=value")
				}
			},
		},
		{
			name:   "compact format",
			data:   map[string]string{"key": "value"},
			config: NewFormatConfig().WithCompact(true),
			check: func(t *testing.T, output string) {
				if strings.Contains(output, "\n") {
					t.Error("Expected compact output without newlines")
				}
			},
		},
		{
			name:   "nil data",
			data:   nil,
			config: NewFormatConfig().WithPretty(true),
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
			config: NewFormatConfig().WithPretty(true),
			check: func(t *testing.T, output string) {
				var result map[string]interface{}
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
				}
				if result["count"].(float64) != 42 {
					t.Error("Expected count=42")
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

func TestJSONFormatterFormatResult(t *testing.T) {
	formatter := NewJSONFormatter()

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
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
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
				if err := json.Unmarshal([]byte(output), &result); err != nil {
					t.Errorf("Failed to unmarshal JSON: %v", err)
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

func TestJSONFormatterFormatError(t *testing.T) {
	formatter := NewJSONFormatter()
	var buf bytes.Buffer

	err := &testError{"test error"}
	config := NewFormatConfig()

	if err := formatter.FormatError(&buf, err, config); err != nil {
		t.Errorf("FormatError failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if result["success"].(bool) {
		t.Error("Expected success=false")
	}
	if result["error"].(string) != "test error" {
		t.Error("Expected error='test error'")
	}
}

func TestJSONFormatterFormatEmpty(t *testing.T) {
	formatter := NewJSONFormatter()
	var buf bytes.Buffer

	config := NewFormatConfig()
	if err := formatter.FormatEmpty(&buf, "No results", config); err != nil {
		t.Errorf("FormatEmpty failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if !result["success"].(bool) {
		t.Error("Expected success=true")
	}
	if result["message"].(string) != "No results" {
		t.Error("Expected message='No results'")
	}
}

func TestJSONFormatterSetIndent(t *testing.T) {
	formatter := NewJSONFormatter().SetIndent("    ")
	var buf bytes.Buffer

	data := map[string]string{"key": "value"}
	config := NewFormatConfig().WithPretty(true)

	if err := formatter.Format(&buf, data, config); err != nil {
		t.Errorf("Format failed: %v", err)
	}

	// Check that the output contains 4-space indentation
	if !strings.Contains(buf.String(), "    ") {
		t.Error("Expected 4-space indentation")
	}
}
