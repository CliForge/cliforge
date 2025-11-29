package output

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
)

// Manager manages output formatting and provides high-level formatting methods.
type Manager struct {
	formatters     map[string]Formatter
	defaultFormat  string
	config         *FormatConfig
	templateEngine *TemplateEngine
}

// NewManager creates a new output manager with default formatters.
func NewManager() *Manager {
	m := &Manager{
		formatters:     make(map[string]Formatter),
		defaultFormat:  "json",
		config:         NewFormatConfig(),
		templateEngine: NewTemplateEngine(),
	}

	// Register default formatters
	m.RegisterFormatter(NewJSONFormatter())
	m.RegisterFormatter(NewYAMLFormatter())
	m.RegisterFormatter(NewTableFormatter())

	return m
}

// RegisterFormatter registers a new formatter.
func (m *Manager) RegisterFormatter(formatter Formatter) {
	m.formatters[formatter.Name()] = formatter
}

// GetFormatter returns a formatter by name.
func (m *Manager) GetFormatter(name string) (Formatter, error) {
	formatter, ok := m.formatters[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("formatter '%s' not found", name)
	}
	return formatter, nil
}

// SetDefaultFormat sets the default output format.
func (m *Manager) SetDefaultFormat(format string) {
	m.defaultFormat = format
}

// SetConfig sets the format configuration.
func (m *Manager) SetConfig(config *FormatConfig) {
	m.config = config
}

// GetConfig returns the current format configuration.
func (m *Manager) GetConfig() *FormatConfig {
	return m.config
}

// Format formats data using the specified format.
func (m *Manager) Format(w io.Writer, data interface{}, format string) error {
	if format == "" {
		format = m.defaultFormat
	}

	formatter, err := m.GetFormatter(format)
	if err != nil {
		return err
	}

	// Check if formatter supports the data type
	if !formatter.Supports(data) {
		return fmt.Errorf("formatter '%s' does not support data type %T", format, data)
	}

	return formatter.Format(w, data, m.config)
}

// FormatWithConfig formats data using the specified format and config.
func (m *Manager) FormatWithConfig(w io.Writer, data interface{}, format string, config *FormatConfig) error {
	if format == "" {
		format = m.defaultFormat
	}

	formatter, err := m.GetFormatter(format)
	if err != nil {
		return err
	}

	// Check if formatter supports the data type
	if !formatter.Supports(data) {
		return fmt.Errorf("formatter '%s' does not support data type %T", format, data)
	}

	return formatter.Format(w, data, config)
}

// FormatResult formats a Result object.
func (m *Manager) FormatResult(w io.Writer, result *Result, format string) error {
	if format == "" {
		format = m.defaultFormat
	}

	formatter, err := m.GetFormatter(format)
	if err != nil {
		return err
	}

	// Try to use specialized FormatResult method if available
	switch f := formatter.(type) {
	case interface {
		FormatResult(io.Writer, *Result, *FormatConfig) error
	}:
		return f.FormatResult(w, result, m.config)
	default:
		// Fallback to regular Format
		if result.Success {
			return formatter.Format(w, result.Data, m.config)
		}
		return m.FormatError(w, fmt.Errorf("%s", result.Error), format)
	}
}

// FormatError formats an error.
func (m *Manager) FormatError(w io.Writer, err error, format string) error {
	if format == "" {
		format = m.defaultFormat
	}

	if err == nil {
		return nil
	}

	formatter, fmtErr := m.GetFormatter(format)
	if fmtErr != nil {
		return fmtErr
	}

	// Try to use specialized FormatError method if available
	switch f := formatter.(type) {
	case interface {
		FormatError(io.Writer, error, *FormatConfig) error
	}:
		return f.FormatError(w, err, m.config)
	default:
		// Fallback to regular Format
		errorData := map[string]interface{}{
			"error": err.Error(),
		}
		return formatter.Format(w, errorData, m.config)
	}
}

// FormatEmpty formats an empty result with an optional message.
func (m *Manager) FormatEmpty(w io.Writer, message string, format string) error {
	if format == "" {
		format = m.defaultFormat
	}

	formatter, err := m.GetFormatter(format)
	if err != nil {
		return err
	}

	// Try to use specialized FormatEmpty method if available
	switch f := formatter.(type) {
	case interface {
		FormatEmpty(io.Writer, string, *FormatConfig) error
	}:
		return f.FormatEmpty(w, message, m.config)
	default:
		// Fallback to regular Format
		if message != "" {
			_, err := w.Write([]byte(message + "\n"))
			return err
		}
		return nil
	}
}

// FormatSuccess formats a success message with optional data.
func (m *Manager) FormatSuccess(w io.Writer, message string, data interface{}, format string) error {
	if format == "" {
		format = m.defaultFormat
	}

	result := NewResult(data).WithMessage(message)
	return m.FormatResult(w, result, format)
}

// SelectFormat selects the appropriate format based on configuration and flags.
// Priority: explicit format > output config > global config > default
func (m *Manager) SelectFormat(explicitFormat string, outputConfig *openapi.CLIOutput, globalConfig *openapi.OutputSettings) string {
	// 1. Explicit format flag takes highest priority
	if explicitFormat != "" {
		return explicitFormat
	}

	// 2. Check operation-level output configuration
	if outputConfig != nil && outputConfig.Format != "" {
		return outputConfig.Format
	}

	// 3. Check global output settings
	if globalConfig != nil && globalConfig.DefaultFormat != "" {
		return globalConfig.DefaultFormat
	}

	// 4. Use manager default
	return m.defaultFormat
}

// ApplyOutputRules applies output configuration rules from OpenAPI spec.
func (m *Manager) ApplyOutputRules(config *openapi.CLIOutput) *FormatConfig {
	formatConfig := m.config
	if formatConfig == nil {
		formatConfig = NewFormatConfig()
	}

	if config != nil {
		formatConfig = formatConfig.WithOutputConfig(config)
	}

	return formatConfig
}

// IsFormatSupported checks if a format is supported.
func (m *Manager) IsFormatSupported(format string) bool {
	_, ok := m.formatters[strings.ToLower(format)]
	return ok
}

// GetSupportedFormats returns a list of all supported format names.
func (m *Manager) GetSupportedFormats() []string {
	formats := make([]string, 0, len(m.formatters))
	for name := range m.formatters {
		formats = append(formats, name)
	}
	return formats
}

// ValidateFormat validates if the format is supported and can handle the data.
func (m *Manager) ValidateFormat(format string, data interface{}) error {
	if format == "" {
		format = m.defaultFormat
	}

	formatter, err := m.GetFormatter(format)
	if err != nil {
		return err
	}

	if !formatter.Supports(data) {
		return fmt.Errorf("format '%s' does not support data type %T", format, data)
	}

	return nil
}

// Print is a convenience method to format and print to stdout.
func (m *Manager) Print(data interface{}, format string) error {
	return m.Format(os.Stdout, data, format)
}

// PrintResult is a convenience method to format and print a Result to stdout.
func (m *Manager) PrintResult(result *Result, format string) error {
	return m.FormatResult(os.Stdout, result, format)
}

// PrintError is a convenience method to format and print an error to stderr.
func (m *Manager) PrintError(err error, format string) error {
	return m.FormatError(os.Stderr, err, format)
}

// PrintSuccess is a convenience method to format and print a success message to stdout.
func (m *Manager) PrintSuccess(message string, data interface{}, format string) error {
	return m.FormatSuccess(os.Stdout, message, data, format)
}

// RenderMessage renders a message template with the given data.
func (m *Manager) RenderMessage(template string, data map[string]interface{}) (string, error) {
	return m.templateEngine.Render(template, data)
}

// RenderSuccessMessage renders a success message from output config.
func (m *Manager) RenderSuccessMessage(outputConfig *openapi.CLIOutput, data map[string]interface{}) (string, error) {
	if outputConfig == nil || outputConfig.SuccessMessage == "" {
		return "", nil
	}
	return m.templateEngine.Render(outputConfig.SuccessMessage, data)
}

// RenderErrorMessage renders an error message from output config.
func (m *Manager) RenderErrorMessage(outputConfig *openapi.CLIOutput, data map[string]interface{}) (string, error) {
	if outputConfig == nil || outputConfig.ErrorMessage == "" {
		return "", nil
	}
	return m.templateEngine.Render(outputConfig.ErrorMessage, data)
}

// FormatWithTemplate formats data and includes a rendered message template.
func (m *Manager) FormatWithTemplate(w io.Writer, data interface{}, messageTemplate string, templateData map[string]interface{}, format string) error {
	if format == "" {
		format = m.defaultFormat
	}

	// Render message template if provided
	var message string
	if messageTemplate != "" {
		var err error
		message, err = m.templateEngine.Render(messageTemplate, templateData)
		if err != nil {
			return fmt.Errorf("failed to render message template: %w", err)
		}
	}

	// Create result with message
	result := NewResult(data)
	if message != "" {
		result = result.WithMessage(message)
	}

	return m.FormatResult(w, result, format)
}

// DefaultManager is the global default output manager.
var DefaultManager = NewManager()

// Package-level convenience functions

// Format formats data using the default manager.
func Format(w io.Writer, data interface{}, format string) error {
	return DefaultManager.Format(w, data, format)
}

// FormatResult formats a Result using the default manager.
func FormatResult(w io.Writer, result *Result, format string) error {
	return DefaultManager.FormatResult(w, result, format)
}

// Print formats and prints data to stdout using the default manager.
func Print(data interface{}, format string) error {
	return DefaultManager.Print(data, format)
}

// PrintResult formats and prints a Result to stdout using the default manager.
func PrintResult(result *Result, format string) error {
	return DefaultManager.PrintResult(result, format)
}

// PrintError formats and prints an error to stderr using the default manager.
func PrintError(err error, format string) error {
	return DefaultManager.PrintError(err, format)
}

// PrintSuccess formats and prints a success message to stdout using the default manager.
func PrintSuccess(message string, data interface{}, format string) error {
	return DefaultManager.PrintSuccess(message, data, format)
}
