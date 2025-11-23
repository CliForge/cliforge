package output

import (
	"io"

	"github.com/CliForge/cliforge/pkg/openapi"
)

// Formatter is the interface that all output formatters must implement.
// It provides a unified way to format and display CLI command output in various formats.
type Formatter interface {
	// Format formats the given data according to the formatter's rules
	// and writes the output to the provided writer.
	Format(w io.Writer, data interface{}, config *FormatConfig) error

	// Name returns the name of the formatter (e.g., "json", "yaml", "table").
	Name() string

	// Supports returns true if the formatter can handle the given data type.
	Supports(data interface{}) bool
}

// FormatConfig contains configuration options for formatting output.
type FormatConfig struct {
	// Output configuration from OpenAPI spec
	OutputConfig *openapi.CLIOutput

	// Pretty enables pretty-printing (for JSON/YAML)
	Pretty bool

	// Colors enables colored output
	Colors bool

	// Compact reduces whitespace in output
	Compact bool

	// ShowHeaders controls header display (for tables)
	ShowHeaders bool

	// SortBy specifies the field to sort by (for tables)
	SortBy string

	// SortAsc controls sort direction
	SortAsc bool

	// MaxWidth is the maximum width for output (for tables)
	MaxWidth int

	// AdditionalData contains extra data for template interpolation
	AdditionalData map[string]interface{}
}

// NewFormatConfig creates a new FormatConfig with sensible defaults.
func NewFormatConfig() *FormatConfig {
	return &FormatConfig{
		Pretty:      true,
		Colors:      true,
		ShowHeaders: true,
		SortAsc:     true,
		MaxWidth:    120,
		AdditionalData: make(map[string]interface{}),
	}
}

// WithOutputConfig sets the OpenAPI output configuration.
func (c *FormatConfig) WithOutputConfig(cfg *openapi.CLIOutput) *FormatConfig {
	c.OutputConfig = cfg
	return c
}

// WithPretty sets the pretty-printing option.
func (c *FormatConfig) WithPretty(pretty bool) *FormatConfig {
	c.Pretty = pretty
	return c
}

// WithColors sets the colors option.
func (c *FormatConfig) WithColors(colors bool) *FormatConfig {
	c.Colors = colors
	return c
}

// WithCompact sets the compact option.
func (c *FormatConfig) WithCompact(compact bool) *FormatConfig {
	c.Compact = compact
	return c
}

// WithSorting sets the sorting options.
func (c *FormatConfig) WithSorting(field string, asc bool) *FormatConfig {
	c.SortBy = field
	c.SortAsc = asc
	return c
}

// WithMaxWidth sets the maximum width for output.
func (c *FormatConfig) WithMaxWidth(width int) *FormatConfig {
	c.MaxWidth = width
	return c
}

// WithData adds additional data for template interpolation.
func (c *FormatConfig) WithData(key string, value interface{}) *FormatConfig {
	c.AdditionalData[key] = value
	return c
}

// Result represents a formatted output result with success/error information.
type Result struct {
	// Success indicates whether the operation was successful
	Success bool

	// Data is the actual result data
	Data interface{}

	// Error is the error message if Success is false
	Error string

	// Message is an optional message to display
	Message string

	// Metadata contains additional information about the result
	Metadata map[string]interface{}
}

// NewResult creates a successful result.
func NewResult(data interface{}) *Result {
	return &Result{
		Success:  true,
		Data:     data,
		Metadata: make(map[string]interface{}),
	}
}

// NewErrorResult creates an error result.
func NewErrorResult(err error) *Result {
	return &Result{
		Success:  false,
		Error:    err.Error(),
		Metadata: make(map[string]interface{}),
	}
}

// WithMessage sets the result message.
func (r *Result) WithMessage(msg string) *Result {
	r.Message = msg
	return r
}

// WithMetadata adds metadata to the result.
func (r *Result) WithMetadata(key string, value interface{}) *Result {
	r.Metadata[key] = value
	return r
}
