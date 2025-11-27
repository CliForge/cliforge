package output

import (
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// YAMLFormatter formats output as YAML.
type YAMLFormatter struct{}

// NewYAMLFormatter creates a new YAML formatter.
func NewYAMLFormatter() *YAMLFormatter {
	return &YAMLFormatter{}
}

// Name returns the formatter name.
func (f *YAMLFormatter) Name() string {
	return "yaml"
}

// Supports returns true if the formatter can handle the given data type.
// YAML formatter can handle any data type.
func (f *YAMLFormatter) Supports(data interface{}) bool {
	return true
}

// Format formats the data as YAML and writes it to the writer.
func (f *YAMLFormatter) Format(w io.Writer, data interface{}, _ *FormatConfig) error {
	// Handle nil data
	if data == nil {
		_, err := w.Write([]byte("null\n"))
		return err
	}

	encoder := yaml.NewEncoder(w)
	defer func() { _ = encoder.Close() }()

	// Set indentation (YAML default is 2 spaces)
	encoder.SetIndent(2)

	// Marshal and write the YAML output
	if err := encoder.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return nil
}

// FormatResult formats a Result object as YAML.
func (f *YAMLFormatter) FormatResult(w io.Writer, result *Result, config *FormatConfig) error {
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

// FormatError formats an error as YAML.
func (f *YAMLFormatter) FormatError(w io.Writer, err error, config *FormatConfig) error {
	if config == nil {
		config = NewFormatConfig()
	}

	errorOutput := map[string]interface{}{
		"success": false,
		"error":   err.Error(),
	}

	return f.Format(w, errorOutput, config)
}

// FormatEmpty formats an empty result as YAML.
func (f *YAMLFormatter) FormatEmpty(w io.Writer, message string, config *FormatConfig) error {
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

// FormatCompact formats data as compact YAML (flow style).
func (f *YAMLFormatter) FormatCompact(w io.Writer, data interface{}) error {
	if data == nil {
		_, err := w.Write([]byte("null\n"))
		return err
	}

	// Create a node with flow style
	node := &yaml.Node{}
	if err := node.Encode(data); err != nil {
		return fmt.Errorf("failed to encode YAML node: %w", err)
	}

	// Set flow style for compact output
	setFlowStyle(node)

	encoder := yaml.NewEncoder(w)
	defer func() { _ = encoder.Close() }()

	if err := encoder.Encode(node); err != nil {
		return fmt.Errorf("failed to encode compact YAML: %w", err)
	}

	return nil
}

// setFlowStyle recursively sets flow style on YAML nodes.
func setFlowStyle(node *yaml.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode, yaml.SequenceNode:
		node.Style = yaml.FlowStyle
		for _, child := range node.Content {
			setFlowStyle(child)
		}
	}
}
