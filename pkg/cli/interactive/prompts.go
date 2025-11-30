// Package interactive provides interactive prompt functionality for CLI operations.
//
// The interactive package implements user-friendly prompts using pterm for terminal UI.
// It supports multiple prompt types with validation and dynamic option loading from APIs.
//
// # Prompt Types
//
//   - text: Free-form text input with optional regex validation
//   - select: Single selection from a list of options
//   - confirm: Yes/no confirmation prompt
//   - number: Numeric input with min/max validation
//
// # Example Usage
//
//	prompter := interactive.NewPrompter(nil)
//
//	// Text prompt with validation
//	name, err := prompter.Text(&interactive.TextPromptOptions{
//		Message: "What is your name?",
//		Validation: "^[A-Za-z ]+$",
//		ValidationMessage: "Name must contain only letters and spaces",
//	})
//
//	// Select prompt
//	choice, err := prompter.Select(&interactive.SelectPromptOptions{
//		Message: "Choose an option",
//		Options: []string{"Option A", "Option B", "Option C"},
//	})
//
//	// Number prompt with range
//	count, err := prompter.Number(&interactive.NumberPromptOptions{
//		Message: "How many instances?",
//		Min: 1,
//		Max: 10,
//	})
//
//	// Confirmation
//	confirmed, err := prompter.Confirm(&interactive.ConfirmPromptOptions{
//		Message: "Are you sure?",
//		Default: false,
//	})
//
// The package automatically handles TTY detection and provides graceful fallbacks
// for non-interactive environments.
package interactive

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
)

// Prompter handles interactive user prompts.
type Prompter struct {
	input  io.Reader
	output io.Writer
	// DisableColor disables colored output
	DisableColor bool
	// DisableInteractive disables interactive prompts (for testing)
	DisableInteractive bool
}

// PrompterConfig configures the Prompter.
type PrompterConfig struct {
	Input              io.Reader
	Output             io.Writer
	DisableColor       bool
	DisableInteractive bool
}

// NewPrompter creates a new Prompter with the given configuration.
// If config is nil, uses default configuration (stdin/stdout).
func NewPrompter(config *PrompterConfig) *Prompter {
	if config == nil {
		config = &PrompterConfig{
			Input:  os.Stdin,
			Output: os.Stdout,
		}
	}

	p := &Prompter{
		input:              config.Input,
		output:             config.Output,
		DisableColor:       config.DisableColor,
		DisableInteractive: config.DisableInteractive,
	}

	// Configure pterm
	if config.DisableColor {
		pterm.DisableColor()
	}

	return p
}

// TextPromptOptions configures a text prompt.
type TextPromptOptions struct {
	Message           string
	Default           string
	Validation        string // Regex pattern
	ValidationMessage string
	Required          bool
}

// Text prompts for text input with optional validation.
func (p *Prompter) Text(opts *TextPromptOptions) (string, error) {
	if opts == nil {
		return "", fmt.Errorf("options cannot be nil")
	}

	if p.DisableInteractive {
		if opts.Default != "" {
			return opts.Default, nil
		}
		return "", fmt.Errorf("interactive prompts disabled")
	}

	// Compile validation regex if provided
	var validationRegex *regexp.Regexp
	if opts.Validation != "" {
		var err error
		validationRegex, err = regexp.Compile(opts.Validation)
		if err != nil {
			return "", fmt.Errorf("invalid validation pattern: %w", err)
		}
	}

	for {
		// Build prompt message
		message := opts.Message
		if opts.Default != "" {
			message = fmt.Sprintf("%s (default: %s)", message, opts.Default)
		}

		// Create text input
		result, err := pterm.DefaultInteractiveTextInput.
			WithMultiLine(false).
			Show(message)
		if err != nil {
			return "", fmt.Errorf("failed to read input: %w", err)
		}

		// Trim whitespace
		result = strings.TrimSpace(result)

		// Use default if empty
		if result == "" && opts.Default != "" {
			result = opts.Default
		}

		// Check if required
		if result == "" && opts.Required {
			pterm.Error.Println("This field is required")
			continue
		}

		// If empty and not required, allow it
		if result == "" && !opts.Required {
			return result, nil
		}

		// Validate if regex provided
		if validationRegex != nil {
			if !validationRegex.MatchString(result) {
				errMsg := opts.ValidationMessage
				if errMsg == "" {
					errMsg = fmt.Sprintf("Input does not match required pattern: %s", opts.Validation)
				}
				pterm.Error.Println(errMsg)
				continue
			}
		}

		return result, nil
	}
}

// SelectPromptOptions configures a select prompt.
type SelectPromptOptions struct {
	Message string
	Options []string
	Default string
}

// Select prompts for selection from a list of options.
func (p *Prompter) Select(opts *SelectPromptOptions) (string, error) {
	if opts == nil {
		return "", fmt.Errorf("options cannot be nil")
	}

	if len(opts.Options) == 0 {
		return "", fmt.Errorf("options list cannot be empty")
	}

	if p.DisableInteractive {
		if opts.Default != "" {
			return opts.Default, nil
		}
		// Return first option as fallback
		return opts.Options[0], nil
	}

	// Find default index
	defaultIndex := 0
	if opts.Default != "" {
		for i, opt := range opts.Options {
			if opt == opts.Default {
				defaultIndex = i
				break
			}
		}
	}

	// Create interactive select
	result, err := pterm.DefaultInteractiveSelect.
		WithOptions(opts.Options).
		WithDefaultOption(opts.Options[defaultIndex]).
		Show(opts.Message)
	if err != nil {
		return "", fmt.Errorf("failed to read selection: %w", err)
	}

	return result, nil
}

// ConfirmPromptOptions configures a confirmation prompt.
type ConfirmPromptOptions struct {
	Message string
	Default bool
}

// Confirm prompts for yes/no confirmation.
func (p *Prompter) Confirm(opts *ConfirmPromptOptions) (bool, error) {
	if opts == nil {
		return false, fmt.Errorf("options cannot be nil")
	}

	if p.DisableInteractive {
		return opts.Default, nil
	}

	// Create interactive confirm
	result, err := pterm.DefaultInteractiveConfirm.
		WithDefaultValue(opts.Default).
		Show(opts.Message)
	if err != nil {
		return false, fmt.Errorf("failed to read confirmation: %w", err)
	}

	return result, nil
}

// NumberPromptOptions configures a number prompt.
type NumberPromptOptions struct {
	Message           string
	Default           *int
	Min               *int
	Max               *int
	ValidationMessage string
	Required          bool
}

// Number prompts for numeric input with optional min/max validation.
func (p *Prompter) Number(opts *NumberPromptOptions) (int, error) {
	if opts == nil {
		return 0, fmt.Errorf("options cannot be nil")
	}

	if p.DisableInteractive {
		if opts.Default != nil {
			return *opts.Default, nil
		}
		return 0, fmt.Errorf("interactive prompts disabled")
	}

	for {
		// Build prompt message
		message := opts.Message
		if opts.Default != nil {
			message = fmt.Sprintf("%s (default: %d)", message, *opts.Default)
		}
		if opts.Min != nil && opts.Max != nil {
			message = fmt.Sprintf("%s [%d-%d]", message, *opts.Min, *opts.Max)
		} else if opts.Min != nil {
			message = fmt.Sprintf("%s (min: %d)", message, *opts.Min)
		} else if opts.Max != nil {
			message = fmt.Sprintf("%s (max: %d)", message, *opts.Max)
		}

		// Get text input
		result, err := pterm.DefaultInteractiveTextInput.
			WithMultiLine(false).
			Show(message)
		if err != nil {
			return 0, fmt.Errorf("failed to read input: %w", err)
		}

		// Trim whitespace
		result = strings.TrimSpace(result)

		// Use default if empty
		if result == "" && opts.Default != nil {
			return *opts.Default, nil
		}

		// Check if required
		if result == "" && opts.Required {
			pterm.Error.Println("This field is required")
			continue
		}

		// If empty and not required, return 0
		if result == "" && !opts.Required {
			return 0, nil
		}

		// Parse number
		num, err := strconv.Atoi(result)
		if err != nil {
			pterm.Error.Println("Please enter a valid number")
			continue
		}

		// Validate range
		if opts.Min != nil && num < *opts.Min {
			errMsg := opts.ValidationMessage
			if errMsg == "" {
				errMsg = fmt.Sprintf("Number must be at least %d", *opts.Min)
			}
			pterm.Error.Println(errMsg)
			continue
		}

		if opts.Max != nil && num > *opts.Max {
			errMsg := opts.ValidationMessage
			if errMsg == "" {
				errMsg = fmt.Sprintf("Number must be at most %d", *opts.Max)
			}
			pterm.Error.Println(errMsg)
			continue
		}

		return num, nil
	}
}

// PromptFromSpec prompts the user based on an OpenAPI interactive prompt specification.
// This is a convenience function that maps OpenAPI prompt types to the appropriate prompt method.
func (p *Prompter) PromptFromSpec(spec *PromptSpec) (interface{}, error) {
	if spec == nil {
		return nil, fmt.Errorf("spec cannot be nil")
	}

	switch spec.Type {
	case "text", "password":
		opts := &TextPromptOptions{
			Message:           spec.Message,
			Validation:        spec.Validation,
			ValidationMessage: spec.ValidationMessage,
			Required:          spec.Required,
		}
		if spec.Default != nil {
			if defaultStr, ok := spec.Default.(string); ok {
				opts.Default = defaultStr
			}
		}
		return p.Text(opts)

	case "select":
		if len(spec.Options) == 0 {
			return nil, fmt.Errorf("select prompt requires options")
		}
		opts := &SelectPromptOptions{
			Message: spec.Message,
			Options: spec.Options,
		}
		if spec.Default != nil {
			if defaultStr, ok := spec.Default.(string); ok {
				opts.Default = defaultStr
			}
		}
		return p.Select(opts)

	case "confirm":
		opts := &ConfirmPromptOptions{
			Message: spec.Message,
		}
		if spec.Default != nil {
			if defaultBool, ok := spec.Default.(bool); ok {
				opts.Default = defaultBool
			}
		}
		return p.Confirm(opts)

	case "number":
		opts := &NumberPromptOptions{
			Message:           spec.Message,
			ValidationMessage: spec.ValidationMessage,
			Required:          spec.Required,
		}
		if spec.Default != nil {
			if defaultNum, ok := spec.Default.(int); ok {
				opts.Default = &defaultNum
			} else if defaultFloat, ok := spec.Default.(float64); ok {
				num := int(defaultFloat)
				opts.Default = &num
			}
		}
		if spec.Min != nil {
			opts.Min = spec.Min
		}
		if spec.Max != nil {
			opts.Max = spec.Max
		}
		return p.Number(opts)

	default:
		return nil, fmt.Errorf("unsupported prompt type: %s", spec.Type)
	}
}

// PromptSpec represents a prompt specification.
// This is designed to work with the x-cli-interactive OpenAPI extension.
type PromptSpec struct {
	Parameter         string
	Type              string      // text, select, confirm, number, password
	Message           string
	Default           interface{} // type varies by prompt type
	Validation        string      // regex for text prompts
	ValidationMessage string
	Options           []string // for select prompts
	Min               *int     // for number prompts
	Max               *int     // for number prompts
	Required          bool
}
