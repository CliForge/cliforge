package output

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
)

// TemplateEngine provides message template rendering with variable interpolation.
// It supports both simple variable substitution (e.g., {name}) and expr expressions.
type TemplateEngine struct {
	// Cache for compiled expr programs
	programCache map[string]*vm.Program
}

// NewTemplateEngine creates a new template engine.
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		programCache: make(map[string]*vm.Program),
	}
}

// Render renders a template string with the given data.
// Supports two template syntaxes:
//  1. Simple variables: {variable_name} or {object.field}
//  2. Expr expressions: {{expression}}
func (t *TemplateEngine) Render(template string, data map[string]interface{}) (string, error) {
	if template == "" {
		return "", nil
	}

	if data == nil {
		data = make(map[string]interface{})
	}

	// First, process expr expressions ({{ }})
	result, err := t.processExpressions(template, data)
	if err != nil {
		return "", err
	}

	// Then, process simple variables ({ })
	result, err = t.processVariables(result, data)
	if err != nil {
		return "", err
	}

	return result, nil
}

// processExpressions processes {{ expr }} style expressions.
func (t *TemplateEngine) processExpressions(template string, data map[string]interface{}) (string, error) {
	// Match {{ expression }}
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)

	var lastErr error
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract expression (remove {{ and }})
		expression := strings.TrimSpace(match[2 : len(match)-2])

		// Evaluate expression
		value, err := t.evaluateExpression(expression, data)
		if err != nil {
			lastErr = err
			return match // Keep original on error
		}

		return fmt.Sprint(value)
	})

	if lastErr != nil {
		return "", fmt.Errorf("failed to evaluate expression: %w", lastErr)
	}

	return result, nil
}

// processVariables processes {variable} style simple variable substitution.
func (t *TemplateEngine) processVariables(template string, data map[string]interface{}) (string, error) {
	// Match { variable } or { object.field }
	re := regexp.MustCompile(`\{([a-zA-Z_][a-zA-Z0-9_\.]*)\}`)

	var lastErr error
	result := re.ReplaceAllStringFunc(template, func(match string) string {
		// Extract variable name (remove { and })
		varPath := strings.TrimSpace(match[1 : len(match)-1])

		// Resolve variable
		value, err := t.resolveVariable(varPath, data)
		if err != nil {
			lastErr = err
			return match // Keep original on error
		}

		return fmt.Sprint(value)
	})

	if lastErr != nil {
		return "", fmt.Errorf("failed to resolve variable: %w", lastErr)
	}

	return result, nil
}

// evaluateExpression evaluates an expr expression.
func (t *TemplateEngine) evaluateExpression(expression string, data map[string]interface{}) (interface{}, error) {
	// Check cache first
	program, ok := t.programCache[expression]
	if !ok {
		// Compile expression
		var err error
		program, err = expr.Compile(expression, expr.Env(data), expr.AllowUndefinedVariables())
		if err != nil {
			return nil, fmt.Errorf("failed to compile expression '%s': %w", expression, err)
		}
		t.programCache[expression] = program
	}

	// Execute program
	result, err := expr.Run(program, data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute expression '%s': %w", expression, err)
	}

	return result, nil
}

// resolveVariable resolves a variable path like "name" or "user.email".
func (t *TemplateEngine) resolveVariable(path string, data map[string]interface{}) (interface{}, error) {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("variable '%s' not found", path)
			}
			current = val
		case map[string]string:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("variable '%s' not found", path)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot access field '%s' on non-map type", part)
		}
	}

	return current, nil
}

// RenderSuccess renders a success message template.
func (t *TemplateEngine) RenderSuccess(template string, data map[string]interface{}) (string, error) {
	if template == "" {
		return "Operation completed successfully", nil
	}
	return t.Render(template, data)
}

// RenderError renders an error message template.
func (t *TemplateEngine) RenderError(template string, data map[string]interface{}) (string, error) {
	if template == "" {
		if errMsg, ok := data["error"].(string); ok {
			return errMsg, nil
		}
		return "Operation failed", nil
	}
	return t.Render(template, data)
}

// ClearCache clears the compiled expression cache.
func (t *TemplateEngine) ClearCache() {
	t.programCache = make(map[string]*vm.Program)
}

// PrecompileTemplate precompiles a template's expr expressions.
func (t *TemplateEngine) PrecompileTemplate(template string, env map[string]interface{}) error {
	// Find all expr expressions
	re := regexp.MustCompile(`\{\{([^}]+)\}\}`)
	matches := re.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) > 1 {
			expression := strings.TrimSpace(match[1])

			// Compile and cache
			program, err := expr.Compile(expression, expr.Env(env), expr.AllowUndefinedVariables())
			if err != nil {
				return fmt.Errorf("failed to precompile expression '%s': %w", expression, err)
			}
			t.programCache[expression] = program
		}
	}

	return nil
}

// MessageTemplate represents a reusable message template.
type MessageTemplate struct {
	Name        string
	Template    string
	Description string
	engine      *TemplateEngine
}

// NewMessageTemplate creates a new message template.
func NewMessageTemplate(name, template, description string) *MessageTemplate {
	return &MessageTemplate{
		Name:        name,
		Template:    template,
		Description: description,
		engine:      NewTemplateEngine(),
	}
}

// Render renders the template with the given data.
func (m *MessageTemplate) Render(data map[string]interface{}) (string, error) {
	return m.engine.Render(m.Template, data)
}

// Precompile precompiles the template expressions.
func (m *MessageTemplate) Precompile(env map[string]interface{}) error {
	return m.engine.PrecompileTemplate(m.Template, env)
}

// TemplateLibrary manages a collection of message templates.
type TemplateLibrary struct {
	templates map[string]*MessageTemplate
}

// NewTemplateLibrary creates a new template library.
func NewTemplateLibrary() *TemplateLibrary {
	return &TemplateLibrary{
		templates: make(map[string]*MessageTemplate),
	}
}

// Add adds a template to the library.
func (l *TemplateLibrary) Add(template *MessageTemplate) {
	l.templates[template.Name] = template
}

// Get retrieves a template by name.
func (l *TemplateLibrary) Get(name string) (*MessageTemplate, error) {
	template, ok := l.templates[name]
	if !ok {
		return nil, fmt.Errorf("template '%s' not found", name)
	}
	return template, nil
}

// Render renders a template by name with the given data.
func (l *TemplateLibrary) Render(name string, data map[string]interface{}) (string, error) {
	template, err := l.Get(name)
	if err != nil {
		return "", err
	}
	return template.Render(data)
}

// List returns all template names.
func (l *TemplateLibrary) List() []string {
	names := make([]string, 0, len(l.templates))
	for name := range l.templates {
		names = append(names, name)
	}
	return names
}

// DefaultTemplates provides commonly used message templates.
var DefaultTemplates = func() *TemplateLibrary {
	lib := NewTemplateLibrary()

	lib.Add(NewMessageTemplate(
		"created",
		"{resource_type} '{name}' created successfully",
		"Resource creation success message",
	))

	lib.Add(NewMessageTemplate(
		"updated",
		"{resource_type} '{name}' updated successfully",
		"Resource update success message",
	))

	lib.Add(NewMessageTemplate(
		"deleted",
		"{resource_type} '{name}' deleted successfully",
		"Resource deletion success message",
	))

	lib.Add(NewMessageTemplate(
		"not_found",
		"{resource_type} '{name}' not found",
		"Resource not found error message",
	))

	lib.Add(NewMessageTemplate(
		"operation_pending",
		"{operation} is in progress for '{name}'",
		"Async operation pending message",
	))

	lib.Add(NewMessageTemplate(
		"operation_complete",
		"{operation} completed successfully for '{name}'",
		"Async operation complete message",
	))

	lib.Add(NewMessageTemplate(
		"operation_failed",
		"{operation} failed for '{name}': {error}",
		"Async operation failed message",
	))

	return lib
}()

// Helper functions for common template patterns

// RenderCreated renders a resource creation message.
func RenderCreated(resourceType, name string) string {
	msg, _ := DefaultTemplates.Render("created", map[string]interface{}{
		"resource_type": resourceType,
		"name":          name,
	})
	return msg
}

// RenderUpdated renders a resource update message.
func RenderUpdated(resourceType, name string) string {
	msg, _ := DefaultTemplates.Render("updated", map[string]interface{}{
		"resource_type": resourceType,
		"name":          name,
	})
	return msg
}

// RenderDeleted renders a resource deletion message.
func RenderDeleted(resourceType, name string) string {
	msg, _ := DefaultTemplates.Render("deleted", map[string]interface{}{
		"resource_type": resourceType,
		"name":          name,
	})
	return msg
}

// RenderNotFound renders a resource not found message.
func RenderNotFound(resourceType, name string) string {
	msg, _ := DefaultTemplates.Render("not_found", map[string]interface{}{
		"resource_type": resourceType,
		"name":          name,
	})
	return msg
}
