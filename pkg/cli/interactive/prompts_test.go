package interactive

import (
	"bytes"
	"strings"
	"testing"
)

// TestNewPrompter tests prompter creation.
func TestNewPrompter(t *testing.T) {
	tests := []struct {
		name   string
		config *PrompterConfig
		want   bool // whether prompter should be created
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			want:   true,
		},
		{
			name: "custom config",
			config: &PrompterConfig{
				Input:              strings.NewReader("test\n"),
				Output:             &bytes.Buffer{},
				DisableColor:       true,
				DisableInteractive: true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompter(tt.config)
			if (p != nil) != tt.want {
				t.Errorf("NewPrompter() = %v, want %v", p != nil, tt.want)
			}
			if p != nil && tt.config != nil {
				if p.DisableColor != tt.config.DisableColor {
					t.Errorf("DisableColor = %v, want %v", p.DisableColor, tt.config.DisableColor)
				}
				if p.DisableInteractive != tt.config.DisableInteractive {
					t.Errorf("DisableInteractive = %v, want %v", p.DisableInteractive, tt.config.DisableInteractive)
				}
			}
		})
	}
}

// TestTextPrompt tests text prompts with validation.
func TestTextPrompt(t *testing.T) {
	tests := []struct {
		name      string
		opts      *TextPromptOptions
		wantErr   bool
		wantValue string // expected when DisableInteractive=true
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
		},
		{
			name: "with default (interactive disabled)",
			opts: &TextPromptOptions{
				Message: "Enter name",
				Default: "John",
			},
			wantErr:   false,
			wantValue: "John",
		},
		{
			name: "required without default (interactive disabled)",
			opts: &TextPromptOptions{
				Message:  "Enter name",
				Required: true,
			},
			wantErr: true,
		},
		{
			name: "with validation pattern",
			opts: &TextPromptOptions{
				Message:           "Enter email",
				Validation:        "^[a-z]+@[a-z]+\\.[a-z]+$",
				ValidationMessage: "Invalid email format",
				Default:           "user@example.com",
			},
			wantErr:   false,
			wantValue: "user@example.com",
		},
		{
			name: "invalid validation regex",
			opts: &TextPromptOptions{
				Message:    "Enter text",
				Validation: "[invalid(regex",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompter(&PrompterConfig{
				Input:              strings.NewReader(""),
				Output:             &bytes.Buffer{},
				DisableInteractive: true,
			})

			got, err := p.Text(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Text() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("Text() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

// TestSelectPrompt tests select prompts.
func TestSelectPrompt(t *testing.T) {
	tests := []struct {
		name      string
		opts      *SelectPromptOptions
		wantErr   bool
		wantValue string
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
		},
		{
			name: "empty options list",
			opts: &SelectPromptOptions{
				Message: "Choose",
				Options: []string{},
			},
			wantErr: true,
		},
		{
			name: "with default",
			opts: &SelectPromptOptions{
				Message: "Choose color",
				Options: []string{"Red", "Green", "Blue"},
				Default: "Green",
			},
			wantErr:   false,
			wantValue: "Green",
		},
		{
			name: "without default uses first",
			opts: &SelectPromptOptions{
				Message: "Choose size",
				Options: []string{"Small", "Medium", "Large"},
			},
			wantErr:   false,
			wantValue: "Small",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompter(&PrompterConfig{
				Input:              strings.NewReader(""),
				Output:             &bytes.Buffer{},
				DisableInteractive: true,
			})

			got, err := p.Select(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Select() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("Select() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

// TestConfirmPrompt tests confirmation prompts.
func TestConfirmPrompt(t *testing.T) {
	tests := []struct {
		name      string
		opts      *ConfirmPromptOptions
		wantErr   bool
		wantValue bool
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
		},
		{
			name: "default true",
			opts: &ConfirmPromptOptions{
				Message: "Continue?",
				Default: true,
			},
			wantErr:   false,
			wantValue: true,
		},
		{
			name: "default false",
			opts: &ConfirmPromptOptions{
				Message: "Delete?",
				Default: false,
			},
			wantErr:   false,
			wantValue: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompter(&PrompterConfig{
				Input:              strings.NewReader(""),
				Output:             &bytes.Buffer{},
				DisableInteractive: true,
			})

			got, err := p.Confirm(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Confirm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("Confirm() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

// TestNumberPrompt tests numeric prompts with validation.
func TestNumberPrompt(t *testing.T) {
	min1 := 1
	max10 := 10
	default5 := 5

	tests := []struct {
		name      string
		opts      *NumberPromptOptions
		wantErr   bool
		wantValue int
	}{
		{
			name:    "nil options",
			opts:    nil,
			wantErr: true,
		},
		{
			name: "with default",
			opts: &NumberPromptOptions{
				Message: "Enter count",
				Default: &default5,
			},
			wantErr:   false,
			wantValue: 5,
		},
		{
			name: "required without default (interactive disabled)",
			opts: &NumberPromptOptions{
				Message:  "Enter number",
				Required: true,
			},
			wantErr: true,
		},
		{
			name: "with min constraint",
			opts: &NumberPromptOptions{
				Message: "Enter count",
				Min:     &min1,
				Default: &default5,
			},
			wantErr:   false,
			wantValue: 5,
		},
		{
			name: "with max constraint",
			opts: &NumberPromptOptions{
				Message: "Enter count",
				Max:     &max10,
				Default: &default5,
			},
			wantErr:   false,
			wantValue: 5,
		},
		{
			name: "with min and max range",
			opts: &NumberPromptOptions{
				Message: "Enter count",
				Min:     &min1,
				Max:     &max10,
				Default: &default5,
			},
			wantErr:   false,
			wantValue: 5,
		},
		{
			name: "with validation message",
			opts: &NumberPromptOptions{
				Message:           "Enter age",
				Min:               &min1,
				ValidationMessage: "Age must be positive",
				Default:           &default5,
			},
			wantErr:   false,
			wantValue: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompter(&PrompterConfig{
				Input:              strings.NewReader(""),
				Output:             &bytes.Buffer{},
				DisableInteractive: true,
			})

			got, err := p.Number(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Number() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("Number() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

// TestPromptFromSpec tests the PromptFromSpec convenience function.
func TestPromptFromSpec(t *testing.T) {
	min1 := 1
	max10 := 10

	tests := []struct {
		name      string
		spec      *PromptSpec
		wantErr   bool
		wantValue interface{}
	}{
		{
			name:    "nil spec",
			spec:    nil,
			wantErr: true,
		},
		{
			name: "text prompt",
			spec: &PromptSpec{
				Parameter: "name",
				Type:      "text",
				Message:   "Enter name",
				Default:   "John",
			},
			wantErr:   false,
			wantValue: "John",
		},
		{
			name: "password prompt (same as text)",
			spec: &PromptSpec{
				Parameter: "password",
				Type:      "password",
				Message:   "Enter password",
				Default:   "secret",
			},
			wantErr:   false,
			wantValue: "secret",
		},
		{
			name: "select prompt",
			spec: &PromptSpec{
				Parameter: "region",
				Type:      "select",
				Message:   "Choose region",
				Options:   []string{"us-east-1", "us-west-2", "eu-west-1"},
				Default:   "us-west-2",
			},
			wantErr:   false,
			wantValue: "us-west-2",
		},
		{
			name: "select prompt without options",
			spec: &PromptSpec{
				Parameter: "region",
				Type:      "select",
				Message:   "Choose region",
				Options:   []string{},
			},
			wantErr: true,
		},
		{
			name: "confirm prompt",
			spec: &PromptSpec{
				Parameter: "confirm",
				Type:      "confirm",
				Message:   "Continue?",
				Default:   true,
			},
			wantErr:   false,
			wantValue: true,
		},
		{
			name: "number prompt with int default",
			spec: &PromptSpec{
				Parameter: "count",
				Type:      "number",
				Message:   "How many?",
				Default:   5,
				Min:       &min1,
				Max:       &max10,
			},
			wantErr:   false,
			wantValue: 5,
		},
		{
			name: "number prompt with float64 default (from JSON)",
			spec: &PromptSpec{
				Parameter: "count",
				Type:      "number",
				Message:   "How many?",
				Default:   float64(7),
				Min:       &min1,
				Max:       &max10,
			},
			wantErr:   false,
			wantValue: 7,
		},
		{
			name: "unsupported prompt type",
			spec: &PromptSpec{
				Parameter: "test",
				Type:      "invalid",
				Message:   "Test",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewPrompter(&PrompterConfig{
				Input:              strings.NewReader(""),
				Output:             &bytes.Buffer{},
				DisableInteractive: true,
			})

			got, err := p.PromptFromSpec(tt.spec)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptFromSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantValue {
				t.Errorf("PromptFromSpec() = %v, want %v", got, tt.wantValue)
			}
		})
	}
}

// TestPromptSpec_AllTypes demonstrates usage of all prompt types.
func TestPromptSpec_AllTypes(t *testing.T) {
	min1 := 1
	max100 := 100

	// This test demonstrates the complete API for each prompt type
	prompter := NewPrompter(&PrompterConfig{
		Input:              strings.NewReader(""),
		Output:             &bytes.Buffer{},
		DisableInteractive: true,
	})

	t.Run("text prompt with validation", func(t *testing.T) {
		spec := &PromptSpec{
			Parameter:         "email",
			Type:              "text",
			Message:           "Enter your email address",
			Default:           "user@example.com",
			Validation:        "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
			ValidationMessage: "Please enter a valid email address",
			Required:          true,
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("text prompt failed: %v", err)
		}

		email, ok := result.(string)
		if !ok {
			t.Error("expected string result")
		}
		if email != "user@example.com" {
			t.Errorf("got %q, want %q", email, "user@example.com")
		}
	})

	t.Run("select prompt with options", func(t *testing.T) {
		spec := &PromptSpec{
			Parameter: "environment",
			Type:      "select",
			Message:   "Select target environment",
			Options:   []string{"development", "staging", "production"},
			Default:   "staging",
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("select prompt failed: %v", err)
		}

		env, ok := result.(string)
		if !ok {
			t.Error("expected string result")
		}
		if env != "staging" {
			t.Errorf("got %q, want %q", env, "staging")
		}
	})

	t.Run("confirm prompt", func(t *testing.T) {
		spec := &PromptSpec{
			Parameter: "destructive",
			Type:      "confirm",
			Message:   "Are you sure you want to delete this resource?",
			Default:   false,
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("confirm prompt failed: %v", err)
		}

		confirmed, ok := result.(bool)
		if !ok {
			t.Error("expected bool result")
		}
		if confirmed != false {
			t.Errorf("got %v, want %v", confirmed, false)
		}
	})

	t.Run("number prompt with range", func(t *testing.T) {
		spec := &PromptSpec{
			Parameter:         "replicas",
			Type:              "number",
			Message:           "How many replicas?",
			Default:           3,
			Min:               &min1,
			Max:               &max100,
			ValidationMessage: "Replicas must be between 1 and 100",
			Required:          true,
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("number prompt failed: %v", err)
		}

		replicas, ok := result.(int)
		if !ok {
			t.Error("expected int result")
		}
		if replicas != 3 {
			t.Errorf("got %d, want %d", replicas, 3)
		}
	})

	t.Run("password prompt (text variant)", func(t *testing.T) {
		spec := &PromptSpec{
			Parameter: "api_token",
			Type:      "password",
			Message:   "Enter API token",
			Default:   "default-token",
			Required:  true,
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("password prompt failed: %v", err)
		}

		token, ok := result.(string)
		if !ok {
			t.Error("expected string result")
		}
		if token != "default-token" {
			t.Errorf("got %q, want %q", token, "default-token")
		}
	})
}

// TestPromptSpec_ValidationScenarios tests various validation scenarios.
func TestPromptSpec_ValidationScenarios(t *testing.T) {
	prompter := NewPrompter(&PrompterConfig{
		Input:              strings.NewReader(""),
		Output:             &bytes.Buffer{},
		DisableInteractive: true,
	})

	t.Run("text validation with valid default", func(t *testing.T) {
		spec := &PromptSpec{
			Type:              "text",
			Message:           "Enter cluster name",
			Default:           "my-cluster-01",
			Validation:        "^[a-z][a-z0-9-]*$",
			ValidationMessage: "Name must start with letter and contain only lowercase letters, numbers, and hyphens",
		}

		_, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("should accept valid default: %v", err)
		}
	})

	t.Run("number validation within range", func(t *testing.T) {
		min := 2
		max := 8
		spec := &PromptSpec{
			Type:              "number",
			Message:           "Enter replica count",
			Default:           4,
			Min:               &min,
			Max:               &max,
			ValidationMessage: "Replica count must be between 2 and 8",
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("should accept valid number: %v", err)
		}

		num, ok := result.(int)
		if !ok || num != 4 {
			t.Errorf("expected 4, got %v", result)
		}
	})

	t.Run("required field with default succeeds", func(t *testing.T) {
		spec := &PromptSpec{
			Type:     "text",
			Message:  "Enter project name",
			Default:  "my-project",
			Required: true,
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("required field with default should succeed: %v", err)
		}

		if result != "my-project" {
			t.Errorf("expected 'my-project', got %q", result)
		}
	})
}

// TestPromptSpec_EdgeCases tests edge cases and error handling.
func TestPromptSpec_EdgeCases(t *testing.T) {
	prompter := NewPrompter(&PrompterConfig{
		Input:              strings.NewReader(""),
		Output:             &bytes.Buffer{},
		DisableInteractive: true,
	})

	t.Run("select with single option", func(t *testing.T) {
		spec := &PromptSpec{
			Type:    "select",
			Message: "Choose the only option",
			Options: []string{"only-option"},
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("single option select should work: %v", err)
		}

		if result != "only-option" {
			t.Errorf("expected 'only-option', got %q", result)
		}
	})

	t.Run("number with only min constraint", func(t *testing.T) {
		min := 0
		spec := &PromptSpec{
			Type:    "number",
			Message: "Enter a non-negative number",
			Default: 42,
			Min:     &min,
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("number with only min should work: %v", err)
		}

		if result != 42 {
			t.Errorf("expected 42, got %v", result)
		}
	})

	t.Run("number with only max constraint", func(t *testing.T) {
		max := 100
		spec := &PromptSpec{
			Type:    "number",
			Message: "Enter a number up to 100",
			Default: 50,
			Max:     &max,
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("number with only max should work: %v", err)
		}

		if result != 50 {
			t.Errorf("expected 50, got %v", result)
		}
	})

	t.Run("confirm with no default", func(t *testing.T) {
		spec := &PromptSpec{
			Type:    "confirm",
			Message: "Proceed?",
			// No default - should use false
		}

		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("confirm without default should work: %v", err)
		}

		// When DisableInteractive and no default, uses the Default field (false)
		confirmed, ok := result.(bool)
		if !ok {
			t.Error("expected bool result")
		}
		if confirmed != false {
			t.Errorf("expected false, got %v", confirmed)
		}
	})
}

// Example_textPrompt demonstrates text prompt usage.
func Example_textPrompt() {
	prompter := NewPrompter(&PrompterConfig{
		DisableInteractive: true, // For example purposes
	})

	// Simple text prompt with default
	result, err := prompter.Text(&TextPromptOptions{
		Message: "What is your name?",
		Default: "Alice",
	})
	if err != nil {
		panic(err)
	}
	println("Hello, " + result)

	// Text prompt with validation
	email, err := prompter.Text(&TextPromptOptions{
		Message:           "Enter your email",
		Validation:        "^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$",
		ValidationMessage: "Please enter a valid email address",
		Required:          true,
		Default:           "user@example.com",
	})
	if err != nil {
		panic(err)
	}
	println("Email: " + email)
}

// Example_selectPrompt demonstrates select prompt usage.
func Example_selectPrompt() {
	prompter := NewPrompter(&PrompterConfig{
		DisableInteractive: true,
	})

	result, err := prompter.Select(&SelectPromptOptions{
		Message: "Choose your deployment environment",
		Options: []string{"development", "staging", "production"},
		Default: "staging",
	})
	if err != nil {
		panic(err)
	}
	println("Deploying to: " + result)
}

// Example_confirmPrompt demonstrates confirm prompt usage.
func Example_confirmPrompt() {
	prompter := NewPrompter(&PrompterConfig{
		DisableInteractive: true,
	})

	confirmed, err := prompter.Confirm(&ConfirmPromptOptions{
		Message: "Delete this cluster?",
		Default: false,
	})
	if err != nil {
		panic(err)
	}
	if confirmed {
		println("Deleting cluster...")
	} else {
		println("Cancelled")
	}
}

// Example_numberPrompt demonstrates number prompt usage.
func Example_numberPrompt() {
	prompter := NewPrompter(&PrompterConfig{
		DisableInteractive: true,
	})

	min := 1
	max := 10
	def := 3

	replicas, err := prompter.Number(&NumberPromptOptions{
		Message: "How many replicas?",
		Min:     &min,
		Max:     &max,
		Default: &def,
	})
	if err != nil {
		panic(err)
	}
	println("Creating", replicas, "replicas")
}

// Example_promptFromSpec demonstrates using OpenAPI spec-based prompts.
func Example_promptFromSpec() {
	prompter := NewPrompter(&PrompterConfig{
		DisableInteractive: true,
	})

	min := 1
	max := 100

	// This mimics how prompts would be defined in OpenAPI x-cli-interactive extension
	specs := []*PromptSpec{
		{
			Parameter:         "cluster_name",
			Type:              "text",
			Message:           "Enter cluster name",
			Default:           "my-cluster",
			Validation:        "^[a-z][a-z0-9-]*$",
			ValidationMessage: "Name must start with letter and contain only lowercase letters, numbers, and hyphens",
		},
		{
			Parameter: "region",
			Type:      "select",
			Message:   "Select region",
			Options:   []string{"us-east-1", "us-west-2", "eu-west-1"},
			Default:   "us-east-1",
		},
		{
			Parameter: "multi_az",
			Type:      "confirm",
			Message:   "Enable multi-AZ?",
			Default:   true,
		},
		{
			Parameter: "replicas",
			Type:      "number",
			Message:   "Number of replicas",
			Min:       &min,
			Max:       &max,
			Default:   3,
		},
	}

	// Collect all prompt values
	values := make(map[string]interface{})
	for _, spec := range specs {
		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			panic(err)
		}
		values[spec.Parameter] = result
	}

	println("Creating cluster with configuration:")
	for param, value := range values {
		println(" ", param, "=", value)
	}
}
