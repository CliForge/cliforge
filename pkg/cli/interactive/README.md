# Interactive Prompts Package

Provides interactive prompt functionality for CLI operations using pterm for cross-platform terminal UI.

## Features

- **Multiple prompt types**: text, select, confirm, number, password
- **Validation**: Regex validation for text, min/max for numbers
- **Dynamic options**: Load select options from API endpoints
- **OpenAPI integration**: Direct support for `x-cli-interactive` extension
- **Non-interactive mode**: Graceful fallback for CI/CD environments
- **Cross-platform**: Uses pterm (no CGO dependencies)

## Prompt Types

### Text Prompt

Free-form text input with optional regex validation.

```go
prompter := interactive.NewPrompter(nil)

result, err := prompter.Text(&interactive.TextPromptOptions{
    Message:           "Enter cluster name",
    Default:           "my-cluster",
    Validation:        "^[a-z][a-z0-9-]*$",
    ValidationMessage: "Name must start with letter and contain only lowercase letters, numbers, and hyphens",
    Required:          true,
})
```

### Select Prompt

Single selection from a list of options.

```go
result, err := prompter.Select(&interactive.SelectPromptOptions{
    Message: "Choose deployment environment",
    Options: []string{"development", "staging", "production"},
    Default: "staging",
})
```

### Confirm Prompt

Yes/no confirmation.

```go
confirmed, err := prompter.Confirm(&interactive.ConfirmPromptOptions{
    Message: "Delete this cluster?",
    Default: false,
})
```

### Number Prompt

Numeric input with optional min/max validation.

```go
min := 1
max := 10
def := 3

replicas, err := prompter.Number(&interactive.NumberPromptOptions{
    Message: "How many replicas?",
    Min:     &min,
    Max:     &max,
    Default: &def,
})
```

## Dynamic Option Loading

Load select options dynamically from API endpoints:

```go
loader := interactive.NewOptionLoader(&interactive.OptionLoaderConfig{
    BaseURL: "https://api.example.com",
})

source := &openapi.PromptSource{
    Endpoint:     "/api/v1/regions",
    ValueField:   "id",
    DisplayField: "name",
}

options, err := loader.LoadOptions(source)
if err != nil {
    // Handle error
}

result, err := prompter.Select(&interactive.SelectPromptOptions{
    Message: "Choose a region",
    Options: options,
})
```

### Supported Response Formats

The loader handles multiple API response formats:

**Simple array of strings:**
```json
["us-east-1", "us-west-2", "eu-west-1"]
```

**Array of objects:**
```json
[
    {"id": "us-east-1", "name": "US East (N. Virginia)"},
    {"id": "us-west-2", "name": "US West (Oregon)"}
]
```

**Wrapped in common fields:**
```json
{
    "items": [...],
    "data": [...],
    "results": [...]
}
```

## OpenAPI Integration

Convert OpenAPI `x-cli-interactive` prompts to PromptSpec:

```go
loader := interactive.NewOptionLoader(&interactive.OptionLoaderConfig{
    BaseURL: "https://api.example.com",
})

// Convert single prompt
spec, err := interactive.PromptFromInteractive(openAPIPrompt, loader)
if err != nil {
    // Handle error
}

result, err := prompter.PromptFromSpec(spec)

// Convert multiple prompts
specs, err := interactive.PromptsFromInteractive(cliInteractive, loader)
if err != nil {
    // Handle error
}

values := make(map[string]interface{})
for _, spec := range specs {
    result, err := prompter.PromptFromSpec(spec)
    if err != nil {
        // Handle error
    }
    values[spec.Parameter] = result
}
```

## Non-Interactive Mode

For CI/CD or automated environments, disable interactive prompts:

```go
prompter := interactive.NewPrompter(&interactive.PrompterConfig{
    DisableInteractive: true,
})

// Will use default values or return error if required
result, err := prompter.Text(&interactive.TextPromptOptions{
    Message: "Enter name",
    Default: "default-name",
})
// result = "default-name"
```

## Complete Example

```go
package main

import (
    "fmt"
    "github.com/CliForge/cliforge/pkg/cli/interactive"
    "github.com/CliForge/cliforge/pkg/openapi"
)

func main() {
    // Setup
    prompter := interactive.NewPrompter(nil)
    loader := interactive.NewOptionLoader(&interactive.OptionLoaderConfig{
        BaseURL: "https://api.example.com",
    })

    // Simulate OpenAPI spec (would normally be parsed)
    cliInteractive := &openapi.CLIInteractive{
        Enabled: true,
        Prompts: []*openapi.InteractivePrompt{
            {
                Parameter:         "cluster_name",
                Type:              "text",
                Message:           "Enter cluster name",
                Validation:        "^[a-z][a-z0-9-]*$",
                ValidationMessage: "Invalid cluster name format",
            },
            {
                Parameter: "region",
                Type:      "select",
                Message:   "Select region",
                Source: &openapi.PromptSource{
                    Endpoint:   "/api/v1/regions",
                    ValueField: "id",
                },
            },
            {
                Parameter: "multi_az",
                Type:      "confirm",
                Message:   "Enable multi-AZ deployment?",
                Default:   true,
            },
        },
    }

    // Convert to PromptSpecs
    specs, err := interactive.PromptsFromInteractive(cliInteractive, loader)
    if err != nil {
        panic(err)
    }

    // Collect values
    values := make(map[string]interface{})
    for _, spec := range specs {
        result, err := prompter.PromptFromSpec(spec)
        if err != nil {
            panic(err)
        }
        values[spec.Parameter] = result
    }

    fmt.Printf("Collected configuration: %+v\n", values)
}
```

## Testing

All prompt types support non-interactive mode for testing:

```go
func TestMyCommand(t *testing.T) {
    prompter := interactive.NewPrompter(&interactive.PrompterConfig{
        Input:              strings.NewReader(""),
        Output:             &bytes.Buffer{},
        DisableInteractive: true,
    })

    // Will use default value
    result, err := prompter.Text(&interactive.TextPromptOptions{
        Message: "Enter name",
        Default: "test-name",
    })

    if result != "test-name" {
        t.Errorf("expected 'test-name', got %q", result)
    }
}
```

## Architecture

```
interactive/
├── prompts.go      - Core prompt implementations
├── loader.go       - Dynamic option loading from APIs
├── prompts_test.go - Prompt tests
└── loader_test.go  - Loader tests
```

### Key Types

- **Prompter**: Main interface for displaying prompts
- **OptionLoader**: Loads dynamic options from API endpoints
- **PromptSpec**: Unified prompt specification (from OpenAPI)
- **TextPromptOptions**: Configuration for text prompts
- **SelectPromptOptions**: Configuration for select prompts
- **ConfirmPromptOptions**: Configuration for confirm prompts
- **NumberPromptOptions**: Configuration for number prompts

## Dependencies

- `github.com/pterm/pterm` - Terminal UI (cross-platform, no CGO)
- `github.com/CliForge/cliforge/pkg/openapi` - OpenAPI extension types

## Performance

- Prompts are instantaneous (no startup overhead)
- Option loading caches HTTP client
- All validation is performed locally (no API calls for validation)

## Error Handling

All prompt methods return errors for:
- Invalid configuration (nil options, empty select list)
- Network failures (when loading dynamic options)
- Validation failures in non-interactive mode
- Invalid regex patterns

In interactive mode, validation errors are displayed to the user with the option to retry.

## Future Enhancements

Planned features:
- Multi-select prompts
- Password masking for sensitive input
- Auto-completion support
- Fuzzy search for large select lists
- Custom validators beyond regex
