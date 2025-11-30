package interactive_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli/interactive"
	"github.com/CliForge/cliforge/pkg/openapi"
)

// TestIntegration_CompleteWorkflow tests the complete workflow from OpenAPI to prompts.
func TestIntegration_CompleteWorkflow(t *testing.T) {
	// Setup mock API server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/regions":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"items": [
					{"id": "us-east-1", "name": "US East (N. Virginia)"},
					{"id": "us-west-2", "name": "US West (Oregon)"},
					{"id": "eu-west-1", "name": "EU (Ireland)"}
				]
			}`))
		case "/api/v1/instance-types":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`["t3.micro", "t3.small", "t3.medium"]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	// Create loader and prompter
	loader := interactive.NewOptionLoader(&interactive.OptionLoaderConfig{
		BaseURL: server.URL,
	})

	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	// Simulate OpenAPI x-cli-interactive extension
	cliInteractive := &openapi.CLIInteractive{
		Enabled: true,
		Prompts: []*openapi.InteractivePrompt{
			{
				Parameter:         "cluster_name",
				Type:              "text",
				Message:           "Enter cluster name",
				Default:           "test-cluster",
				Validation:        "^[a-z][a-z0-9-]*$",
				ValidationMessage: "Name must start with letter and contain only lowercase letters, numbers, and hyphens",
			},
			{
				Parameter: "region",
				Type:      "select",
				Message:   "Select AWS region",
				Source: &openapi.PromptSource{
					Endpoint:   "/api/v1/regions",
					ValueField: "id",
				},
			},
			{
				Parameter: "instance_type",
				Type:      "select",
				Message:   "Select instance type",
				Source: &openapi.PromptSource{
					Endpoint: "/api/v1/instance-types",
				},
			},
			{
				Parameter: "multi_az",
				Type:      "confirm",
				Message:   "Enable multi-AZ deployment?",
				Default:   true,
			},
			{
				Parameter: "node_count",
				Type:      "number",
				Message:   "Number of nodes",
				Default:   3,
			},
		},
	}

	// Convert to PromptSpecs
	specs, err := interactive.PromptsFromInteractive(cliInteractive, loader)
	if err != nil {
		t.Fatalf("PromptsFromInteractive failed: %v", err)
	}

	if len(specs) != 5 {
		t.Fatalf("expected 5 specs, got %d", len(specs))
	}

	// Verify each spec was properly created
	expectedParams := []string{"cluster_name", "region", "instance_type", "multi_az", "node_count"}
	for i, spec := range specs {
		if spec.Parameter != expectedParams[i] {
			t.Errorf("spec[%d] parameter = %s, want %s", i, spec.Parameter, expectedParams[i])
		}
	}

	// Verify dynamic options were loaded
	regionSpec := specs[1]
	if len(regionSpec.Options) != 3 {
		t.Errorf("region spec options = %d, want 3", len(regionSpec.Options))
	}
	if regionSpec.Options[0] != "us-east-1" {
		t.Errorf("region spec options[0] = %s, want us-east-1", regionSpec.Options[0])
	}

	instanceSpec := specs[2]
	if len(instanceSpec.Options) != 3 {
		t.Errorf("instance_type spec options = %d, want 3", len(instanceSpec.Options))
	}
	if instanceSpec.Options[0] != "t3.micro" {
		t.Errorf("instance_type spec options[0] = %s, want t3.micro", instanceSpec.Options[0])
	}

	// Execute prompts and collect values
	values := make(map[string]interface{})
	for _, spec := range specs {
		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Fatalf("PromptFromSpec(%s) failed: %v", spec.Parameter, err)
		}
		values[spec.Parameter] = result
	}

	// Verify collected values
	expectedValues := map[string]interface{}{
		"cluster_name":  "test-cluster",
		"region":        "us-east-1",
		"instance_type": "t3.micro",
		"multi_az":      true,
		"node_count":    3,
	}

	for param, expected := range expectedValues {
		if values[param] != expected {
			t.Errorf("values[%s] = %v, want %v", param, values[param], expected)
		}
	}
}

// TestIntegration_ErrorHandling tests error scenarios.
func TestIntegration_ErrorHandling(t *testing.T) {
	// Server that returns errors
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	loader := interactive.NewOptionLoader(&interactive.OptionLoaderConfig{
		BaseURL: server.URL,
	})

	// Prompt that requires loading options from failing endpoint
	prompt := &openapi.InteractivePrompt{
		Parameter: "region",
		Type:      "select",
		Message:   "Select region",
		Source: &openapi.PromptSource{
			Endpoint: "/api/v1/regions",
		},
	}

	// Should fail to load options
	_, err := interactive.PromptFromInteractive(prompt, loader)
	if err == nil {
		t.Error("expected error when endpoint returns 500, got nil")
	}
}

// TestIntegration_StaticPrompts tests prompts without dynamic loading.
func TestIntegration_StaticPrompts(t *testing.T) {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	loader := interactive.NewOptionLoader(nil)

	// Prompts without dynamic sources
	cliInteractive := &openapi.CLIInteractive{
		Enabled: true,
		Prompts: []*openapi.InteractivePrompt{
			{
				Parameter: "name",
				Type:      "text",
				Message:   "Enter name",
				Default:   "default-name",
			},
			{
				Parameter: "confirm",
				Type:      "confirm",
				Message:   "Proceed?",
				Default:   true,
			},
		},
	}

	specs, err := interactive.PromptsFromInteractive(cliInteractive, loader)
	if err != nil {
		t.Fatalf("PromptsFromInteractive failed: %v", err)
	}

	// Execute all prompts
	for _, spec := range specs {
		_, err := prompter.PromptFromSpec(spec)
		if err != nil {
			t.Errorf("PromptFromSpec(%s) failed: %v", spec.Parameter, err)
		}
	}
}

// TestIntegration_DisabledInteractive tests disabled interactive mode.
func TestIntegration_DisabledInteractive(t *testing.T) {
	loader := interactive.NewOptionLoader(nil)

	cliInteractive := &openapi.CLIInteractive{
		Enabled: false, // Disabled
		Prompts: []*openapi.InteractivePrompt{
			{
				Parameter: "test",
				Type:      "text",
				Message:   "Test",
			},
		},
	}

	specs, err := interactive.PromptsFromInteractive(cliInteractive, loader)
	if err != nil {
		t.Fatalf("PromptsFromInteractive failed: %v", err)
	}

	// Should return nil/empty when disabled
	if len(specs) > 0 {
		t.Errorf("expected nil/empty specs when disabled, got %d specs", len(specs))
	}
}

// BenchmarkPromptExecution benchmarks prompt execution.
func BenchmarkPromptExecution(b *testing.B) {
	prompter := interactive.NewPrompter(&interactive.PrompterConfig{
		DisableInteractive: true,
	})

	spec := &interactive.PromptSpec{
		Parameter: "test",
		Type:      "text",
		Message:   "Test",
		Default:   "value",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := prompter.PromptFromSpec(spec)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDynamicOptionLoading benchmarks loading options from API.
func BenchmarkDynamicOptionLoading(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`["option1", "option2", "option3"]`))
	}))
	defer server.Close()

	loader := interactive.NewOptionLoader(&interactive.OptionLoaderConfig{
		BaseURL: server.URL,
	})

	source := &openapi.PromptSource{
		Endpoint: "/options",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadOptions(source)
		if err != nil {
			b.Fatal(err)
		}
	}
}
