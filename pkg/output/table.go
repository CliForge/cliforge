package output

import (
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/pterm/pterm"
)

// TableFormatter formats output as a table using pterm.
type TableFormatter struct {
	// Default table style
	style *pterm.TablePrinter
}

// NewTableFormatter creates a new table formatter.
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{
		style: &pterm.DefaultTable,
	}
}

// Name returns the formatter name.
func (f *TableFormatter) Name() string {
	return "table"
}

// Supports returns true if the formatter can handle the given data type.
// Table formatter supports slices, arrays, and maps.
func (f *TableFormatter) Supports(data interface{}) bool {
	if data == nil {
		return false
	}

	v := reflect.ValueOf(data)
	kind := v.Kind()

	// Support slices and arrays
	if kind == reflect.Slice || kind == reflect.Array {
		return v.Len() > 0
	}

	// Support maps
	if kind == reflect.Map {
		return v.Len() > 0
	}

	// Support single structs (will be shown as key-value table)
	if kind == reflect.Struct {
		return true
	}

	// Support pointers to supported types
	if kind == reflect.Ptr {
		return f.Supports(v.Elem().Interface())
	}

	return false
}

// Format formats the data as a table and writes it to the writer.
func (f *TableFormatter) Format(w io.Writer, data interface{}, config *FormatConfig) error {
	if config == nil {
		config = NewFormatConfig()
	}

	if data == nil {
		return fmt.Errorf("cannot format nil data as table")
	}

	v := reflect.ValueOf(data)

	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return fmt.Errorf("cannot format nil pointer as table")
		}
		v = v.Elem()
	}

	var tableData [][]string
	var err error

	switch v.Kind() {
	case reflect.Slice, reflect.Array:
		tableData, err = f.formatSlice(v, config)
	case reflect.Map:
		tableData, err = f.formatMap(v, config)
	case reflect.Struct:
		tableData, err = f.formatStruct(v, config)
	default:
		return fmt.Errorf("unsupported data type for table formatting: %s", v.Kind())
	}

	if err != nil {
		return err
	}

	// Apply sorting if configured
	if config.SortBy != "" && len(tableData) > 1 {
		tableData = f.sortTableData(tableData, config)
	}

	// Create the table
	table := pterm.DefaultTable.WithHasHeader(config.ShowHeaders)

	// Configure colors
	if config.Colors {
		table = table.WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold))
		table = table.WithRowSeparator("-")
	} else {
		// Disable colors
		pterm.DisableColor()
		defer func() {
			pterm.EnableColor()
		}()
	}

	// Render the table
	rendered, err := table.WithData(tableData).Srender()
	if err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	_, err = w.Write([]byte(rendered))
	return err
}

// formatSlice formats a slice or array as a table.
func (f *TableFormatter) formatSlice(v reflect.Value, config *FormatConfig) ([][]string, error) {
	if v.Len() == 0 {
		return nil, fmt.Errorf("empty slice")
	}

	// Get columns configuration
	var columns []*openapi.TableColumn
	if config.OutputConfig != nil && config.OutputConfig.Table != nil {
		columns = config.OutputConfig.Table.Columns
	}

	// If no columns configured, auto-detect from first element
	if len(columns) == 0 {
		columns = f.autoDetectColumns(v.Index(0))
	}

	// Build table data
	tableData := make([][]string, 0, v.Len()+1)

	// Add header row if configured
	if config.ShowHeaders {
		headers := make([]string, len(columns))
		for i, col := range columns {
			headers[i] = col.Header
			if headers[i] == "" {
				headers[i] = col.Field
			}
		}
		tableData = append(tableData, headers)
	}

	// Add data rows
	for i := 0; i < v.Len(); i++ {
		row := make([]string, len(columns))
		elem := v.Index(i)

		for j, col := range columns {
			value := f.extractField(elem, col.Field)
			row[j] = f.transformValue(value, col.Transform, config)

			// Apply width constraint if specified
			if col.Width > 0 && len(row[j]) > col.Width {
				row[j] = row[j][:col.Width-3] + "..."
			}
		}

		tableData = append(tableData, row)
	}

	return tableData, nil
}

// formatMap formats a map as a two-column key-value table.
func (f *TableFormatter) formatMap(v reflect.Value, config *FormatConfig) ([][]string, error) {
	if v.Len() == 0 {
		return nil, fmt.Errorf("empty map")
	}

	tableData := make([][]string, 0, v.Len()+1)

	// Add header
	if config.ShowHeaders {
		tableData = append(tableData, []string{"KEY", "VALUE"})
	}

	// Get and sort keys
	keys := v.MapKeys()
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprint(keys[i].Interface()) < fmt.Sprint(keys[j].Interface())
	})

	// Add rows
	for _, key := range keys {
		value := v.MapIndex(key)
		tableData = append(tableData, []string{
			fmt.Sprint(key.Interface()),
			f.formatValue(value.Interface()),
		})
	}

	return tableData, nil
}

// formatStruct formats a struct as a two-column key-value table.
func (f *TableFormatter) formatStruct(v reflect.Value, config *FormatConfig) ([][]string, error) {
	t := v.Type()
	tableData := make([][]string, 0, t.NumField()+1)

	// Add header
	if config.ShowHeaders {
		tableData = append(tableData, []string{"FIELD", "VALUE"})
	}

	// Add rows for each field
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get field name (use json tag if available)
		fieldName := field.Name
		if jsonTag := field.Tag.Get("json"); jsonTag != "" {
			parts := strings.Split(jsonTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				fieldName = parts[0]
			}
		}

		tableData = append(tableData, []string{
			fieldName,
			f.formatValue(value.Interface()),
		})
	}

	return tableData, nil
}

// autoDetectColumns automatically detects columns from a value.
func (f *TableFormatter) autoDetectColumns(v reflect.Value) []*openapi.TableColumn {
	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	var columns []*openapi.TableColumn

	if v.Kind() == reflect.Struct {
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported fields
			if !field.IsExported() {
				continue
			}

			// Get field name (use json tag if available)
			fieldName := field.Name
			header := field.Name
			if jsonTag := field.Tag.Get("json"); jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] != "" && parts[0] != "-" {
					fieldName = parts[0]
					header = strings.ToUpper(fieldName)
				}
			}

			columns = append(columns, &openapi.TableColumn{
				Field:  fieldName,
				Header: header,
			})
		}
	} else if v.Kind() == reflect.Map {
		// For maps, create dynamic columns from keys
		for _, key := range v.MapKeys() {
			keyStr := fmt.Sprint(key.Interface())
			columns = append(columns, &openapi.TableColumn{
				Field:  keyStr,
				Header: strings.ToUpper(keyStr),
			})
		}
	}

	return columns
}

// extractField extracts a field value from a reflect.Value.
func (f *TableFormatter) extractField(v reflect.Value, field string) interface{} {
	// Handle pointer
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil
		}
		v = v.Elem()
	}

	// Handle map
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			if fmt.Sprint(key.Interface()) == field {
				return v.MapIndex(key).Interface()
			}
		}
		return nil
	}

	// Handle struct
	if v.Kind() == reflect.Struct {
		// Try direct field access
		if field := v.FieldByName(field); field.IsValid() {
			return field.Interface()
		}

		// Try json tag lookup
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if jsonTag := f.Tag.Get("json"); jsonTag != "" {
				parts := strings.Split(jsonTag, ",")
				if parts[0] == field {
					return v.Field(i).Interface()
				}
			}
		}
	}

	return nil
}

// transformValue applies a transformation to a value.
func (f *TableFormatter) transformValue(value interface{}, transform string, config *FormatConfig) string {
	str := f.formatValue(value)

	switch strings.ToLower(transform) {
	case "uppercase", "upper":
		return strings.ToUpper(str)
	case "lowercase", "lower":
		return strings.ToLower(str)
	case "title":
		// Simple title case - capitalize first letter of each word
		words := strings.Fields(str)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + word[1:]
			}
		}
		return strings.Join(words, " ")
	case "trim":
		return strings.TrimSpace(str)
	default:
		return str
	}
}

// formatValue formats a value as a string.
func (f *TableFormatter) formatValue(value interface{}) string {
	if value == nil {
		return ""
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		value = v.Elem().Interface()
	}

	switch val := value.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return fmt.Sprint(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// sortTableData sorts table data by the specified column.
func (f *TableFormatter) sortTableData(data [][]string, config *FormatConfig) [][]string {
	if len(data) <= 1 {
		return data
	}

	// Find column index
	colIndex := -1
	if config.ShowHeaders {
		for i, header := range data[0] {
			if strings.EqualFold(header, config.SortBy) {
				colIndex = i
				break
			}
		}
	}

	if colIndex == -1 {
		return data
	}

	// Separate header and data rows
	var header []string
	var rows [][]string
	if config.ShowHeaders {
		header = data[0]
		rows = data[1:]
	} else {
		rows = data
	}

	// Sort rows
	sort.Slice(rows, func(i, j int) bool {
		if colIndex >= len(rows[i]) || colIndex >= len(rows[j]) {
			return false
		}

		if config.SortAsc {
			return rows[i][colIndex] < rows[j][colIndex]
		}
		return rows[i][colIndex] > rows[j][colIndex]
	})

	// Reconstruct table
	if config.ShowHeaders {
		return append([][]string{header}, rows...)
	}
	return rows
}

// FormatResult formats a Result object as a table.
func (f *TableFormatter) FormatResult(w io.Writer, result *Result, config *FormatConfig) error {
	if config == nil {
		config = NewFormatConfig()
	}

	if !result.Success {
		// For errors, show as key-value table
		errorData := map[string]interface{}{
			"success": false,
			"error":   result.Error,
		}
		return f.Format(w, errorData, config)
	}

	return f.Format(w, result.Data, config)
}

// FormatError formats an error as a table.
func (f *TableFormatter) FormatError(w io.Writer, err error, config *FormatConfig) error {
	errorData := map[string]interface{}{
		"error": err.Error(),
	}
	return f.Format(w, errorData, config)
}

// FormatEmpty formats an empty result message.
func (f *TableFormatter) FormatEmpty(w io.Writer, message string, config *FormatConfig) error {
	if message == "" {
		message = "No results found"
	}
	_, err := w.Write([]byte(message + "\n"))
	return err
}
