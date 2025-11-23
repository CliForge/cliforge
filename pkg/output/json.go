package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// JSONFormatter formats output as JSON with optional pretty printing.
type JSONFormatter struct {
	indent string
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{
		indent: "  ",
	}
}

// Name returns the formatter name.
func (f *JSONFormatter) Name() string {
	return "json"
}

// Supports returns true if the formatter can handle the given data type.
// JSON formatter can handle any data type.
func (f *JSONFormatter) Supports(data interface{}) bool {
	return true
}

// Format formats the data as JSON and writes it to the writer.
func (f *JSONFormatter) Format(w io.Writer, data interface{}, config *FormatConfig) error {
	if config == nil {
		config = NewFormatConfig()
	}

	// Handle nil data
	if data == nil {
		if config.Pretty {
			_, err := w.Write([]byte("null\n"))
			return err
		}
		_, err := w.Write([]byte("null"))
		return err
	}

	var encoder *json.Encoder
	var output []byte
	var err error

	if config.Pretty && !config.Compact {
		// Pretty-print with indentation
		output, err = json.MarshalIndent(data, "", f.indent)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		_, err = w.Write(output)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte("\n"))
		return err
	}

	// Compact JSON output
	if config.Compact {
		output, err = json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}
		_, err = w.Write(output)
		return err
	}

	// Default: use encoder for streaming
	encoder = json.NewEncoder(w)
	if config.Pretty {
		encoder.SetIndent("", f.indent)
	}

	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}

// SetIndent sets the indentation string for pretty printing.
func (f *JSONFormatter) SetIndent(indent string) *JSONFormatter {
	f.indent = indent
	return f
}

// FormatResult formats a Result object as JSON.
func (f *JSONFormatter) FormatResult(w io.Writer, result *Result, config *FormatConfig) error {
	if config == nil {
		config = NewFormatConfig()
	}

	// Create a map representation of the result
	output := make(map[string]interface{})
	output["success"] = result.Success

	if result.Success {
		output["data"] = result.Data
	} else {
		output["error"] = result.Error
	}

	if result.Message != "" {
		output["message"] = result.Message
	}

	if len(result.Metadata) > 0 {
		output["metadata"] = result.Metadata
	}

	return f.Format(w, output, config)
}

// FormatError formats an error as JSON.
func (f *JSONFormatter) FormatError(w io.Writer, err error, config *FormatConfig) error {
	if config == nil {
		config = NewFormatConfig()
	}

	errorOutput := map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	}

	return f.Format(w, errorOutput, config)
}

// FormatEmpty formats an empty result as JSON.
func (f *JSONFormatter) FormatEmpty(w io.Writer, message string, config *FormatConfig) error {
	if config == nil {
		config = NewFormatConfig()
	}

	emptyOutput := map[string]interface{}{
		"success": true,
		"data":    nil,
	}

	if message != "" {
		emptyOutput["message"] = message
	}

	return f.Format(w, emptyOutput, config)
}
