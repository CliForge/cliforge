package interactive_test

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/CliForge/cliforge/pkg/cli/interactive"
	"github.com/CliForge/cliforge/pkg/openapi"
)

// Example demonstrates the basic usage of all prompt types.
func Example() {
	// Create prompter in non-interactive mode for testing
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	// Text prompt
	name, _ := prompter.Text(&interactive.TextPromptOptions{
		Message: "Enter cluster name",
		Default: "my-cluster",
	})
	fmt.Println("Cluster name:", name)

	// Select prompt
	region, _ := prompter.Select(&interactive.SelectPromptOptions{
		Message: "Choose region",
		Options: []string{"us-east-1", "us-west-2", "eu-west-1"},
		Default: "us-east-1",
	})
	fmt.Println("Region:", region)

	// Confirm prompt
	confirmed, _ := prompter.Confirm(&interactive.ConfirmPromptOptions{
		Message: "Enable multi-AZ?",
		Default: true,
	})
	fmt.Println("Multi-AZ:", confirmed)

	// Number prompt
	min := 1
	max := 10
	def := 3
	replicas, _ := prompter.Number(&interactive.NumberPromptOptions{
		Message: "Number of replicas",
		Min:     &min,
		Max:     &max,
		Default: &def,
	})
	fmt.Println("Replicas:", replicas)

	// Output:
	// Cluster name: my-cluster
	// Region: us-east-1
	// Multi-AZ: true
	// Replicas: 3
}

// Example_validation demonstrates text validation with regex.
func Example_validation() {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	// Valid cluster name
	name, err := prompter.Text(&interactive.TextPromptOptions{
		Message:           "Enter cluster name",
		Default:           "my-cluster-01",
		Validation:        "^[a-z][a-z0-9-]*$",
		ValidationMessage: "Name must start with letter and contain only lowercase letters, numbers, and hyphens",
	})
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Valid name:", name)
	}

	// Output:
	// Valid name: my-cluster-01
}

// Example_numberValidation demonstrates number range validation.
func Example_numberValidation() {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	min := 2
	max := 8
	def := 4

	replicas, err := prompter.Number(&interactive.NumberPromptOptions{
		Message:           "Enter replica count",
		Min:               &min,
		Max:               &max,
		Default:           &def,
		ValidationMessage: "Replica count must be between 2 and 8",
	})
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Replicas:", replicas)
	}

	// Output:
	// Replicas: 4
}

// Example_promptSpec demonstrates using PromptSpec for OpenAPI integration.
func Example_promptSpec() {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	min := 1
	max := 100

	// Define prompts using PromptSpec (as would come from OpenAPI)
	specs := []*interactive.PromptSpec{
		{
			Parameter:         "cluster_name",
			Type:              "text",
			Message:           "Enter cluster name",
			Default:           "my-cluster",
			Validation:        "^[a-z][a-z0-9-]*$",
			ValidationMessage: "Invalid name format",
			Required:          true,
		},
		{
			Parameter: "environment",
			Type:      "select",
			Message:   "Select environment",
			Options:   []string{"development", "staging", "production"},
			Default:   "staging",
		},
		{
			Parameter: "multi_az",
			Type:      "confirm",
			Message:   "Enable multi-AZ?",
			Default:   true,
		},
		{
			Parameter:         "replicas",
			Type:              "number",
			Message:           "Number of replicas",
			Min:               &min,
			Max:               &max,
			Default:           3,
			ValidationMessage: "Must be between 1 and 100",
			Required:          true,
		},
	}

	// Collect all values
	values := make(map[string]interface{})
	for _, spec := range specs {
		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			fmt.Printf("Error on %s: %v\n", spec.Parameter, err)
			continue
		}
		values[spec.Parameter] = result
	}

	// Display collected values
	fmt.Println("Configuration:")
	fmt.Printf("  cluster_name: %v\n", values["cluster_name"])
	fmt.Printf("  environment: %v\n", values["environment"])
	fmt.Printf("  multi_az: %v\n", values["multi_az"])
	fmt.Printf("  replicas: %v\n", values["replicas"])

	// Output:
	// Configuration:
	//   cluster_name: my-cluster
	//   environment: staging
	//   multi_az: true
	//   replicas: 3
}

// Example_openAPIIntegration demonstrates converting OpenAPI prompts.
func Example_openAPIIntegration() {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	// No option loader needed for this example (no dynamic options)
	loader := interactive.NewOptionLoader(nil)

	// Simulate OpenAPI x-cli-interactive extension
	cliInteractive := &openapi.CLIInteractive{
		Enabled: true,
		Prompts: []*openapi.InteractivePrompt{
			{
				Parameter:         "name",
				Type:              "text",
				Message:           "Resource name",
				Default:           "my-resource",
				Validation:        "^[a-z][a-z0-9-]*$",
				ValidationMessage: "Invalid name format",
			},
			{
				Parameter: "confirm",
				Type:      "confirm",
				Message:   "Proceed with creation?",
				Default:   true,
			},
		},
	}

	// Convert to PromptSpecs
	specs, err := interactive.PromptsFromInteractive(cliInteractive, loader)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Execute prompts
	values := make(map[string]interface{})
	for _, spec := range specs {
		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			fmt.Printf("Error on %s: %v\n", spec.Parameter, err)
			continue
		}
		values[spec.Parameter] = result
	}

	fmt.Printf("name: %v\n", values["name"])
	fmt.Printf("confirm: %v\n", values["confirm"])

	// Output:
	// name: my-resource
	// confirm: true
}

// Example_requiredFields demonstrates required field handling.
func Example_requiredFields() {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	// Required field with default succeeds
	name, err := prompter.Text(&interactive.TextPromptOptions{
		Message:  "Enter name",
		Default:  "default-name",
		Required: true,
	})
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Name:", name)
	}

	// Required field without default fails in non-interactive mode
	_, err = prompter.Text(&interactive.TextPromptOptions{
		Message:  "Enter password",
		Required: true,
	})
	if err != nil {
		fmt.Println("Password error: interactive prompts disabled")
	}

	// Output:
	// Name: default-name
	// Password error: interactive prompts disabled
}

// Example_customConfiguration demonstrates custom prompter configuration.
func Example_customConfiguration() {
	// Custom output buffer for testing
	output := &bytes.Buffer{}

	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		Input:              strings.NewReader(""),
		Output:             output,
		DisableColor:       true,
		DisableInteractive: true,
	})

	result, _ := prompter.Text(&interactive.TextPromptOptions{
		Message: "Enter value",
		Default: "test-value",
	})

	fmt.Println("Result:", result)
	fmt.Println("Color disabled:", prompter.DisableColor)

	// Output:
	// Result: test-value
	// Color disabled: true
}

// Example_allPromptTypes demonstrates all supported prompt types.
func Example_allPromptTypes() {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	min := 1
	max := 10

	specs := []*interactive.PromptSpec{
		{
			Type:    "text",
			Message: "Text prompt",
			Default: "text-value",
		},
		{
			Type:    "password",
			Message: "Password prompt",
			Default: "secret",
		},
		{
			Type:    "select",
			Message: "Select prompt",
			Options: []string{"A", "B", "C"},
			Default: "B",
		},
		{
			Type:    "confirm",
			Message: "Confirm prompt",
			Default: false,
		},
		{
			Type:    "number",
			Message: "Number prompt",
			Default: 5,
			Min:     &min,
			Max:     &max,
		},
	}

	for _, spec := range specs {
		result, _ := prompter.PromptFromSpec(spec)
		fmt.Printf("%s: %v\n", spec.Type, result)
	}

	// Output:
	// text: text-value
	// password: secret
	// select: B
	// confirm: false
	// number: 5
}
