package openapi

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// Validator validates OpenAPI specs and CLI extensions.
type Validator struct {
	// Strict enables strict validation mode
	Strict bool
	// AllowUnknownExtensions allows unknown x-cli-* extensions
	AllowUnknownExtensions bool
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{
		Strict:                 false,
		AllowUnknownExtensions: true,
	}
}

// ValidationResult contains validation results.
type ValidationResult struct {
	Valid    bool
	Errors   []ValidationError
	Warnings []ValidationWarning
}

// ValidationError represents a validation error.
type ValidationError struct {
	Path    string
	Field   string
	Message string
}

// ValidationWarning represents a validation warning.
type ValidationWarning struct {
	Path    string
	Field   string
	Message string
}

// Validate validates a parsed spec and its CLI extensions.
func (v *Validator) Validate(ctx context.Context, spec *ParsedSpec) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid: true,
	}

	// Validate basic OpenAPI structure
	if err := spec.Spec.Validate(ctx); err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:    "spec",
			Message: fmt.Sprintf("OpenAPI validation failed: %v", err),
		})
		if v.Strict {
			return result, nil
		}
	}

	// Validate global extensions
	if spec.Extensions != nil {
		v.validateExtensions(spec, result)
	}

	// Validate operations
	operations, err := spec.GetOperations()
	if err != nil {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Path:    "operations",
			Message: fmt.Sprintf("Failed to get operations: %v", err),
		})
		return result, nil
	}

	v.validateOperations(operations, result)

	// Check for duplicate command names
	v.checkDuplicateCommands(operations, result)

	return result, nil
}

// validateExtensions validates global CLI extensions.
func (v *Validator) validateExtensions(spec *ParsedSpec, result *ValidationResult) {
	if config := spec.Extensions.Config; config != nil {
		// Validate x-cli-config
		if config.Name == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Path:    "x-cli-config",
				Field:   "name",
				Message: "CLI name not specified",
			})
		}

		if config.Version == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Path:    "x-cli-config",
				Field:   "version",
				Message: "CLI version not specified",
			})
		}

		// Validate color format
		if config.Branding != nil && config.Branding.Colors != nil {
			if config.Branding.Colors.Primary != "" {
				if !isValidHexColor(config.Branding.Colors.Primary) {
					result.Errors = append(result.Errors, ValidationError{
						Path:    "x-cli-config.branding.colors",
						Field:   "primary",
						Message: "Invalid hex color format",
					})
					result.Valid = false
				}
			}
			if config.Branding.Colors.Secondary != "" {
				if !isValidHexColor(config.Branding.Colors.Secondary) {
					result.Errors = append(result.Errors, ValidationError{
						Path:    "x-cli-config.branding.colors",
						Field:   "secondary",
						Message: "Invalid hex color format",
					})
					result.Valid = false
				}
			}
		}

		// Validate auth type
		if config.Auth != nil {
			validAuthTypes := map[string]bool{
				"oauth2": true,
				"apikey": true,
				"basic":  true,
				"bearer": true,
			}
			if config.Auth.Type != "" && !validAuthTypes[config.Auth.Type] {
				result.Errors = append(result.Errors, ValidationError{
					Path:    "x-cli-config.auth",
					Field:   "type",
					Message: fmt.Sprintf("Invalid auth type: %s (must be oauth2, apikey, basic, or bearer)", config.Auth.Type),
				})
				result.Valid = false
			}

			validStorageTypes := map[string]bool{
				"file":    true,
				"keyring": true,
				"memory":  true,
			}
			if config.Auth.Storage != "" && !validStorageTypes[config.Auth.Storage] {
				result.Errors = append(result.Errors, ValidationError{
					Path:    "x-cli-config.auth",
					Field:   "storage",
					Message: fmt.Sprintf("Invalid storage type: %s (must be file, keyring, or memory)", config.Auth.Storage),
				})
				result.Valid = false
			}
		}

		// Validate output formats
		if config.Output != nil {
			validFormats := map[string]bool{
				"table": true,
				"json":  true,
				"yaml":  true,
				"csv":   true,
			}
			if config.Output.DefaultFormat != "" && !validFormats[config.Output.DefaultFormat] {
				result.Errors = append(result.Errors, ValidationError{
					Path:    "x-cli-config.output",
					Field:   "default-format",
					Message: fmt.Sprintf("Invalid default format: %s", config.Output.DefaultFormat),
				})
				result.Valid = false
			}
			for _, format := range config.Output.SupportedFormats {
				if !validFormats[format] {
					result.Errors = append(result.Errors, ValidationError{
						Path:    "x-cli-config.output",
						Field:   "supported-formats",
						Message: fmt.Sprintf("Invalid format: %s", format),
					})
					result.Valid = false
				}
			}
		}
	}

	// Validate changelog entries
	for i, entry := range spec.Extensions.Changelog {
		if entry.Version == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Path:    fmt.Sprintf("x-cli-changelog[%d]", i),
				Field:   "version",
				Message: "Changelog entry missing version",
			})
		}
		if entry.Date == "" {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Path:    fmt.Sprintf("x-cli-changelog[%d]", i),
				Field:   "date",
				Message: "Changelog entry missing date",
			})
		}

		// Validate change types and severities
		for j, change := range entry.Changes {
			validTypes := map[string]bool{
				"added":      true,
				"removed":    true,
				"modified":   true,
				"deprecated": true,
				"security":   true,
			}
			if change.Type != "" && !validTypes[change.Type] {
				result.Errors = append(result.Errors, ValidationError{
					Path:    fmt.Sprintf("x-cli-changelog[%d].changes[%d]", i, j),
					Field:   "type",
					Message: fmt.Sprintf("Invalid change type: %s", change.Type),
				})
				result.Valid = false
			}

			validSeverities := map[string]bool{
				"breaking":  true,
				"dangerous": true,
				"safe":      true,
			}
			if change.Severity != "" && !validSeverities[change.Severity] {
				result.Errors = append(result.Errors, ValidationError{
					Path:    fmt.Sprintf("x-cli-changelog[%d].changes[%d]", i, j),
					Field:   "severity",
					Message: fmt.Sprintf("Invalid severity: %s", change.Severity),
				})
				result.Valid = false
			}
		}
	}
}

// validateOperations validates operation-level extensions.
func (v *Validator) validateOperations(operations []*Operation, result *ValidationResult) {
	for _, op := range operations {
		path := fmt.Sprintf("%s %s", op.Method, op.Path)

		// Validate x-cli-flags
		for i, flag := range op.CLIFlags {
			if flag.Name == "" {
				result.Errors = append(result.Errors, ValidationError{
					Path:    path,
					Field:   fmt.Sprintf("x-cli-flags[%d].name", i),
					Message: "Flag name is required",
				})
				result.Valid = false
			}
			if flag.Source == "" {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Path:    path,
					Field:   fmt.Sprintf("x-cli-flags[%d].source", i),
					Message: "Flag source not specified",
				})
			}
			if flag.Flag != "" && !strings.HasPrefix(flag.Flag, "--") {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Path:    path,
					Field:   fmt.Sprintf("x-cli-flags[%d].flag", i),
					Message: "Flag should start with '--'",
				})
			}

			// Validate flag type
			validTypes := map[string]bool{
				"string":  true,
				"integer": true,
				"boolean": true,
				"array":   true,
				"number":  true,
			}
			if flag.Type != "" && !validTypes[flag.Type] {
				result.Errors = append(result.Errors, ValidationError{
					Path:    path,
					Field:   fmt.Sprintf("x-cli-flags[%d].type", i),
					Message: fmt.Sprintf("Invalid flag type: %s", flag.Type),
				})
				result.Valid = false
			}
		}

		// Validate x-cli-interactive
		if interactive := op.CLIInteractive; interactive != nil {
			for i, prompt := range interactive.Prompts {
				validPromptTypes := map[string]bool{
					"text":     true,
					"select":   true,
					"confirm":  true,
					"number":   true,
					"password": true,
				}
				if prompt.Type != "" && !validPromptTypes[prompt.Type] {
					result.Errors = append(result.Errors, ValidationError{
						Path:    path,
						Field:   fmt.Sprintf("x-cli-interactive.prompts[%d].type", i),
						Message: fmt.Sprintf("Invalid prompt type: %s", prompt.Type),
					})
					result.Valid = false
				}

				if prompt.Type == "select" && prompt.Source == nil {
					result.Warnings = append(result.Warnings, ValidationWarning{
						Path:    path,
						Field:   fmt.Sprintf("x-cli-interactive.prompts[%d].source", i),
						Message: "Select prompt should have a source",
					})
				}
			}
		}

		// Validate x-cli-preflight
		for i, check := range op.CLIPreflight {
			if check.Endpoint == "" {
				result.Errors = append(result.Errors, ValidationError{
					Path:    path,
					Field:   fmt.Sprintf("x-cli-preflight[%d].endpoint", i),
					Message: "Preflight endpoint is required",
				})
				result.Valid = false
			}
			if check.Method == "" {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Path:    path,
					Field:   fmt.Sprintf("x-cli-preflight[%d].method", i),
					Message: "Preflight method not specified, will default to GET",
				})
			}
		}

		// Validate x-cli-async
		if async := op.CLIAsync; async != nil && async.Enabled {
			if async.StatusEndpoint == "" {
				result.Errors = append(result.Errors, ValidationError{
					Path:    path,
					Field:   "x-cli-async.status-endpoint",
					Message: "Async status endpoint is required",
				})
				result.Valid = false
			}
			if len(async.TerminalStates) == 0 {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Path:    path,
					Field:   "x-cli-async.terminal-states",
					Message: "No terminal states defined",
				})
			}
			if async.Polling != nil {
				if async.Polling.Interval <= 0 {
					result.Warnings = append(result.Warnings, ValidationWarning{
						Path:    path,
						Field:   "x-cli-async.polling.interval",
						Message: "Polling interval should be positive",
					})
				}
				if async.Polling.Timeout <= 0 {
					result.Warnings = append(result.Warnings, ValidationWarning{
						Path:    path,
						Field:   "x-cli-async.polling.timeout",
						Message: "Polling timeout should be positive",
					})
				}
			}
		}

		// Validate x-cli-workflow
		if workflow := op.CLIWorkflow; workflow != nil {
			for i, step := range workflow.Steps {
				if step.ID == "" {
					result.Errors = append(result.Errors, ValidationError{
						Path:    path,
						Field:   fmt.Sprintf("x-cli-workflow.steps[%d].id", i),
						Message: "Workflow step ID is required",
					})
					result.Valid = false
				}
				if step.Request == nil {
					result.Errors = append(result.Errors, ValidationError{
						Path:    path,
						Field:   fmt.Sprintf("x-cli-workflow.steps[%d].request", i),
						Message: "Workflow step request is required",
					})
					result.Valid = false
				} else {
					if step.Request.Method == "" {
						result.Errors = append(result.Errors, ValidationError{
							Path:    path,
							Field:   fmt.Sprintf("x-cli-workflow.steps[%d].request.method", i),
							Message: "Request method is required",
						})
						result.Valid = false
					}
					if step.Request.URL == "" {
						result.Errors = append(result.Errors, ValidationError{
							Path:    path,
							Field:   fmt.Sprintf("x-cli-workflow.steps[%d].request.url", i),
							Message: "Request URL is required",
						})
						result.Valid = false
					}
				}

				// Validate foreach/as pairing
				if (step.ForEach != "" && step.As == "") || (step.ForEach == "" && step.As != "") {
					result.Errors = append(result.Errors, ValidationError{
						Path:    path,
						Field:   fmt.Sprintf("x-cli-workflow.steps[%d]", i),
						Message: "foreach and as must be used together",
					})
					result.Valid = false
				}
			}
		}
	}
}

// checkDuplicateCommands checks for duplicate CLI command names.
func (v *Validator) checkDuplicateCommands(operations []*Operation, result *ValidationResult) {
	commandMap := make(map[string][]string)

	for _, op := range operations {
		if op.CLICommand != "" {
			path := fmt.Sprintf("%s %s", op.Method, op.Path)
			commandMap[op.CLICommand] = append(commandMap[op.CLICommand], path)
		}

		// Check aliases too
		for _, alias := range op.CLIAliases {
			path := fmt.Sprintf("%s %s (alias)", op.Method, op.Path)
			commandMap[alias] = append(commandMap[alias], path)
		}
	}

	for command, paths := range commandMap {
		if len(paths) > 1 {
			result.Errors = append(result.Errors, ValidationError{
				Path:    "commands",
				Field:   command,
				Message: fmt.Sprintf("Duplicate command '%s' found in: %s", command, strings.Join(paths, ", ")),
			})
			result.Valid = false
		}
	}
}

// isValidHexColor validates hex color format (#RRGGBB or #RGB).
func isValidHexColor(color string) bool {
	match, _ := regexp.MatchString(`^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$`, color)
	return match
}

// Error returns a formatted error string.
func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s.%s: %s", e.Path, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

// String returns a formatted warning string.
func (w ValidationWarning) String() string {
	if w.Field != "" {
		return fmt.Sprintf("%s.%s: %s", w.Path, w.Field, w.Message)
	}
	return fmt.Sprintf("%s: %s", w.Path, w.Message)
}
