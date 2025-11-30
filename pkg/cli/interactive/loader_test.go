package interactive

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
)

// TestNewOptionLoader tests option loader creation.
func TestNewOptionLoader(t *testing.T) {
	tests := []struct {
		name   string
		config *OptionLoaderConfig
		want   bool
	}{
		{
			name:   "nil config uses defaults",
			config: nil,
			want:   true,
		},
		{
			name: "custom config",
			config: &OptionLoaderConfig{
				HTTPClient: &http.Client{},
				BaseURL:    "https://api.example.com",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewOptionLoader(tt.config)
			if (loader != nil) != tt.want {
				t.Errorf("NewOptionLoader() = %v, want %v", loader != nil, tt.want)
			}
			if loader != nil && tt.config != nil {
				if loader.baseURL != tt.config.BaseURL {
					t.Errorf("BaseURL = %v, want %v", loader.baseURL, tt.config.BaseURL)
				}
			}
		})
	}
}

// TestLoadOptions tests loading options from various sources.
func TestLoadOptions(t *testing.T) {
	// Setup mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/regions":
			// Simple array of strings
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`["us-east-1", "us-west-2", "eu-west-1"]`))

		case "/clusters":
			// Array of objects with items wrapper
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{
				"items": [
					{"id": "cluster-1", "name": "Production"},
					{"id": "cluster-2", "name": "Staging"},
					{"id": "cluster-3", "name": "Development"}
				]
			}`))

		case "/environments":
			// Direct array of objects
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[
				{"value": "dev", "display": "Development"},
				{"value": "staging", "display": "Staging"},
				{"value": "prod", "display": "Production"}
			]`))

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	loader := NewOptionLoader(&OptionLoaderConfig{
		BaseURL: server.URL,
	})

	tests := []struct {
		name    string
		source  *openapi.PromptSource
		want    []string
		wantErr bool
	}{
		{
			name:    "nil source",
			source:  nil,
			wantErr: true,
		},
		{
			name: "no endpoint specified",
			source: &openapi.PromptSource{
				Endpoint: "",
			},
			wantErr: true,
		},
		{
			name: "simple array of strings",
			source: &openapi.PromptSource{
				Endpoint: "/regions",
			},
			want:    []string{"us-east-1", "us-west-2", "eu-west-1"},
			wantErr: false,
		},
		{
			name: "array of objects with value field",
			source: &openapi.PromptSource{
				Endpoint:   "/clusters",
				ValueField: "id",
			},
			want:    []string{"cluster-1", "cluster-2", "cluster-3"},
			wantErr: false,
		},
		{
			name: "array of objects with display field",
			source: &openapi.PromptSource{
				Endpoint:     "/clusters",
				DisplayField: "name",
			},
			want:    []string{"Production", "Staging", "Development"},
			wantErr: false,
		},
		{
			name: "array with both value and display (uses value)",
			source: &openapi.PromptSource{
				Endpoint:     "/environments",
				ValueField:   "value",
				DisplayField: "display",
			},
			want:    []string{"dev", "staging", "prod"},
			wantErr: false,
		},
		{
			name: "non-existent endpoint",
			source: &openapi.PromptSource{
				Endpoint: "/nonexistent",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loader.LoadOptions(tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("LoadOptions() got %d items, want %d", len(got), len(tt.want))
					return
				}
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("LoadOptions()[%d] = %v, want %v", i, v, tt.want[i])
					}
				}
			}
		})
	}
}

// TestLoadOptions_AbsoluteURL tests loading from absolute URLs.
func TestLoadOptions_AbsoluteURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`["option1", "option2"]`))
	}))
	defer server.Close()

	loader := NewOptionLoader(&OptionLoaderConfig{
		BaseURL: "https://different-base.com", // Should be ignored for absolute URLs
	})

	source := &openapi.PromptSource{
		Endpoint: server.URL + "/absolute",
	}

	got, err := loader.LoadOptions(source)
	if err != nil {
		t.Errorf("LoadOptions() with absolute URL failed: %v", err)
		return
	}

	if len(got) != 2 {
		t.Errorf("LoadOptions() got %d items, want 2", len(got))
	}
}

// TestIsAbsoluteURL tests URL type detection.
func TestIsAbsoluteURL(t *testing.T) {
	tests := []struct {
		url  string
		want bool
	}{
		{"http://example.com", true},
		{"https://example.com", true},
		{"https://api.example.com/path", true},
		{"/relative/path", false},
		{"relative/path", false},
		{"", false},
		{"ftp://example.com", false}, // Not http/https
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			if got := isAbsoluteURL(tt.url); got != tt.want {
				t.Errorf("isAbsoluteURL(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

// TestPromptFromInteractive tests converting OpenAPI prompts to PromptSpec.
func TestPromptFromInteractive(t *testing.T) {
	// Setup mock HTTP server for dynamic options
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`["option1", "option2", "option3"]`))
	}))
	defer server.Close()

	loader := NewOptionLoader(&OptionLoaderConfig{
		BaseURL: server.URL,
	})

	tests := []struct {
		name    string
		prompt  *openapi.InteractivePrompt
		loader  *OptionLoader
		wantErr bool
		check   func(*testing.T, *PromptSpec)
	}{
		{
			name:    "nil prompt",
			prompt:  nil,
			loader:  loader,
			wantErr: true,
		},
		{
			name: "text prompt without source",
			prompt: &openapi.InteractivePrompt{
				Parameter:         "name",
				Type:              "text",
				Message:           "Enter name",
				Default:           "John",
				Validation:        "^[A-Za-z]+$",
				ValidationMessage: "Only letters allowed",
			},
			loader:  loader,
			wantErr: false,
			check: func(t *testing.T, spec *PromptSpec) {
				if spec.Parameter != "name" {
					t.Errorf("Parameter = %v, want name", spec.Parameter)
				}
				if spec.Type != "text" {
					t.Errorf("Type = %v, want text", spec.Type)
				}
				if spec.Message != "Enter name" {
					t.Errorf("Message = %v, want 'Enter name'", spec.Message)
				}
				if spec.Validation != "^[A-Za-z]+$" {
					t.Errorf("Validation = %v, want regex", spec.Validation)
				}
			},
		},
		{
			name: "select prompt with source",
			prompt: &openapi.InteractivePrompt{
				Parameter: "region",
				Type:      "select",
				Message:   "Choose region",
				Source: &openapi.PromptSource{
					Endpoint: "/options",
				},
			},
			loader:  loader,
			wantErr: false,
			check: func(t *testing.T, spec *PromptSpec) {
				if spec.Type != "select" {
					t.Errorf("Type = %v, want select", spec.Type)
				}
				if len(spec.Options) != 3 {
					t.Errorf("Options length = %d, want 3", len(spec.Options))
				}
				if len(spec.Options) > 0 && spec.Options[0] != "option1" {
					t.Errorf("Options[0] = %v, want option1", spec.Options[0])
				}
			},
		},
		{
			name: "confirm prompt",
			prompt: &openapi.InteractivePrompt{
				Parameter: "confirm",
				Type:      "confirm",
				Message:   "Are you sure?",
				Default:   true,
			},
			loader:  loader,
			wantErr: false,
			check: func(t *testing.T, spec *PromptSpec) {
				if spec.Type != "confirm" {
					t.Errorf("Type = %v, want confirm", spec.Type)
				}
				if spec.Default != true {
					t.Errorf("Default = %v, want true", spec.Default)
				}
			},
		},
		{
			name: "number prompt",
			prompt: &openapi.InteractivePrompt{
				Parameter: "count",
				Type:      "number",
				Message:   "How many?",
				Default:   5,
			},
			loader:  loader,
			wantErr: false,
			check: func(t *testing.T, spec *PromptSpec) {
				if spec.Type != "number" {
					t.Errorf("Type = %v, want number", spec.Type)
				}
				if spec.Default != 5 {
					t.Errorf("Default = %v, want 5", spec.Default)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PromptFromInteractive(tt.prompt, tt.loader)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptFromInteractive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.check != nil {
				tt.check(t, got)
			}
		})
	}
}

// TestPromptsFromInteractive tests converting multiple prompts.
func TestPromptsFromInteractive(t *testing.T) {
	loader := NewOptionLoader(nil)

	tests := []struct {
		name        string
		interactive *openapi.CLIInteractive
		wantCount   int
		wantErr     bool
	}{
		{
			name:        "nil interactive",
			interactive: nil,
			wantCount:   0,
			wantErr:     false,
		},
		{
			name: "disabled interactive",
			interactive: &openapi.CLIInteractive{
				Enabled: false,
				Prompts: []*openapi.InteractivePrompt{
					{Parameter: "test", Type: "text", Message: "Test"},
				},
			},
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "multiple prompts",
			interactive: &openapi.CLIInteractive{
				Enabled: true,
				Prompts: []*openapi.InteractivePrompt{
					{Parameter: "name", Type: "text", Message: "Name?"},
					{Parameter: "confirm", Type: "confirm", Message: "Sure?"},
					{Parameter: "count", Type: "number", Message: "How many?"},
				},
			},
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "empty prompts list",
			interactive: &openapi.CLIInteractive{
				Enabled: true,
				Prompts: []*openapi.InteractivePrompt{},
			},
			wantCount: 0,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PromptsFromInteractive(tt.interactive, loader)
			if (err != nil) != tt.wantErr {
				t.Errorf("PromptsFromInteractive() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.wantCount {
				t.Errorf("PromptsFromInteractive() returned %d prompts, want %d", len(got), tt.wantCount)
			}
		})
	}
}

// TestExtractOptions tests option extraction from different response formats.
func TestExtractOptions(t *testing.T) {
	loader := NewOptionLoader(nil)

	tests := []struct {
		name    string
		data    interface{}
		source  *openapi.PromptSource
		want    []string
		wantErr bool
	}{
		{
			name: "array of strings",
			data: []interface{}{"a", "b", "c"},
			source: &openapi.PromptSource{
				Endpoint: "/test",
			},
			want:    []string{"a", "b", "c"},
			wantErr: false,
		},
		{
			name: "wrapped in items",
			data: map[string]interface{}{
				"items": []interface{}{
					map[string]interface{}{"id": "1", "name": "One"},
					map[string]interface{}{"id": "2", "name": "Two"},
				},
			},
			source: &openapi.PromptSource{
				Endpoint:   "/test",
				ValueField: "id",
			},
			want:    []string{"1", "2"},
			wantErr: false,
		},
		{
			name: "wrapped in data",
			data: map[string]interface{}{
				"data": []interface{}{"x", "y", "z"},
			},
			source: &openapi.PromptSource{
				Endpoint: "/test",
			},
			want:    []string{"x", "y", "z"},
			wantErr: false,
		},
		{
			name: "wrapped in results",
			data: map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{"value": "v1"},
					map[string]interface{}{"value": "v2"},
				},
			},
			source: &openapi.PromptSource{
				Endpoint:   "/test",
				ValueField: "value",
			},
			want:    []string{"v1", "v2"},
			wantErr: false,
		},
		{
			name: "invalid format (not array or wrapper)",
			data: map[string]interface{}{
				"other": "value",
			},
			source: &openapi.PromptSource{
				Endpoint: "/test",
			},
			wantErr: true,
		},
		{
			name: "numeric values converted to strings",
			data: []interface{}{
				map[string]interface{}{"count": 1},
				map[string]interface{}{"count": 2},
				map[string]interface{}{"count": 3},
			},
			source: &openapi.PromptSource{
				Endpoint:   "/test",
				ValueField: "count",
			},
			want:    []string{"1", "2", "3"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := loader.extractOptions(tt.data, tt.source)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(got) != len(tt.want) {
					t.Errorf("extractOptions() got %d items, want %d", len(got), len(tt.want))
					return
				}
				for i, v := range got {
					if v != tt.want[i] {
						t.Errorf("extractOptions()[%d] = %v, want %v", i, v, tt.want[i])
					}
				}
			}
		})
	}
}

// Example_optionLoader demonstrates loading dynamic options from an API.
func Example_optionLoader() {
	// This would typically connect to a real API
	loader := NewOptionLoader(&OptionLoaderConfig{
		BaseURL: "https://api.example.com",
	})

	// Define a prompt source that loads from an endpoint
	source := &openapi.PromptSource{
		Endpoint:     "/api/v1/regions",
		ValueField:   "id",
		DisplayField: "name",
	}

	// Load options dynamically
	options, err := loader.LoadOptions(source)
	if err != nil {
		// Handle error (endpoint not reachable in example)
		return
	}

	// Use options in a select prompt
	prompter := NewPrompter(&PrompterConfig{
		DisableInteractive: true,
	})

	result, err := prompter.Select(&SelectPromptOptions{
		Message: "Choose a region",
		Options: options,
	})
	if err != nil {
		panic(err)
	}

	println("Selected region:", result)
}

// Example_fullWorkflow demonstrates the complete workflow from OpenAPI spec to prompts.
func Example_fullWorkflow() {
	// Setup loader for dynamic options
	loader := NewOptionLoader(&OptionLoaderConfig{
		BaseURL: "https://api.example.com",
	})

	// Simulate OpenAPI x-cli-interactive extension
	interactive := &openapi.CLIInteractive{
		Enabled: true,
		Prompts: []*openapi.InteractivePrompt{
			{
				Parameter: "cluster_name",
				Type:      "text",
				Message:   "Enter cluster name",
				Validation: "^[a-z][a-z0-9-]*$",
				ValidationMessage: "Name must start with letter and contain only lowercase letters, numbers, and hyphens",
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
		},
	}

	// Convert to PromptSpecs
	specs, err := PromptsFromInteractive(interactive, loader)
	if err != nil {
		// Handle error (endpoint not reachable in example)
		return
	}

	// Create prompter and collect values
	prompter := NewPrompter(&PrompterConfig{
		DisableInteractive: true,
	})

	values := make(map[string]interface{})
	for _, spec := range specs {
		result, err := prompter.PromptFromSpec(spec)
		if err != nil {
			panic(err)
		}
		values[spec.Parameter] = result
	}

	println("Collected values:", values)
}
