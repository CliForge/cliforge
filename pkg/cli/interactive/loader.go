package interactive

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/CliForge/cliforge/pkg/openapi"
)

// OptionLoader handles dynamic loading of prompt options from API endpoints.
type OptionLoader struct {
	client  *http.Client
	baseURL string
}

// OptionLoaderConfig configures the OptionLoader.
type OptionLoaderConfig struct {
	HTTPClient *http.Client
	BaseURL    string
}

// NewOptionLoader creates a new OptionLoader for fetching dynamic prompt options.
func NewOptionLoader(config *OptionLoaderConfig) *OptionLoader {
	if config == nil {
		config = &OptionLoaderConfig{}
	}

	client := config.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &OptionLoader{
		client:  client,
		baseURL: config.BaseURL,
	}
}

// LoadOptions loads options from a PromptSource.
// It supports both static values and dynamic endpoint loading.
func (l *OptionLoader) LoadOptions(source *openapi.PromptSource) ([]string, error) {
	if source == nil {
		return nil, fmt.Errorf("source cannot be nil")
	}

	// If endpoint is specified, fetch from API
	if source.Endpoint != "" {
		return l.loadFromEndpoint(source)
	}

	// Otherwise, we'd need static values (not in PromptSource struct currently)
	return nil, fmt.Errorf("no endpoint specified in prompt source")
}

// loadFromEndpoint fetches options from an API endpoint.
func (l *OptionLoader) loadFromEndpoint(source *openapi.PromptSource) ([]string, error) {
	// Build full URL
	url := source.Endpoint
	if l.baseURL != "" && !isAbsoluteURL(url) {
		url = l.baseURL + url
	}

	// Make HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch options from %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from %s", resp.StatusCode, url)
	}

	// Parse response
	var data interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Extract options based on field configuration
	return l.extractOptions(data, source)
}

// extractOptions extracts option values from the response data.
func (l *OptionLoader) extractOptions(data interface{}, source *openapi.PromptSource) ([]string, error) {
	// Handle array response
	items, ok := data.([]interface{})
	if !ok {
		// Try to find items in a wrapper object
		if dataMap, ok := data.(map[string]interface{}); ok {
			// Common patterns: "items", "data", "results"
			for _, key := range []string{"items", "data", "results"} {
				if itemsVal, exists := dataMap[key]; exists {
					if itemsArray, ok := itemsVal.([]interface{}); ok {
						items = itemsArray
						break
					}
				}
			}
		}
		if items == nil {
			return nil, fmt.Errorf("response is not an array and does not contain items/data/results field")
		}
	}

	var options []string
	valueField := source.ValueField
	displayField := source.DisplayField

	// If no fields specified, assume array of strings
	if valueField == "" && displayField == "" {
		for _, item := range items {
			if str, ok := item.(string); ok {
				options = append(options, str)
			}
		}
		return options, nil
	}

	// Extract from objects
	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		// Determine which field to use for the option value
		field := valueField
		if field == "" {
			field = displayField
		}
		if field == "" {
			continue
		}

		// Extract value
		if val, exists := itemMap[field]; exists {
			if strVal, ok := val.(string); ok {
				options = append(options, strVal)
			} else {
				// Convert to string
				options = append(options, fmt.Sprintf("%v", val))
			}
		}
	}

	return options, nil
}

// isAbsoluteURL checks if a URL is absolute (has scheme).
func isAbsoluteURL(url string) bool {
	return len(url) > 7 && (url[:7] == "http://" || url[:8] == "https://")
}

// PromptFromInteractive creates a PromptSpec from an OpenAPI InteractivePrompt.
// It handles loading dynamic options if a source endpoint is specified.
func PromptFromInteractive(prompt *openapi.InteractivePrompt, loader *OptionLoader) (*PromptSpec, error) {
	if prompt == nil {
		return nil, fmt.Errorf("prompt cannot be nil")
	}

	spec := &PromptSpec{
		Parameter:         prompt.Parameter,
		Type:              prompt.Type,
		Message:           prompt.Message,
		Default:           prompt.Default,
		Validation:        prompt.Validation,
		ValidationMessage: prompt.ValidationMessage,
	}

	// Load dynamic options if source is specified
	if prompt.Source != nil && loader != nil {
		options, err := loader.LoadOptions(prompt.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to load options for %s: %w", prompt.Parameter, err)
		}
		spec.Options = options
	}

	// For number prompts, we could extract min/max from validation or schema
	// This would require additional OpenAPI schema parsing
	// For now, these would need to be added separately

	return spec, nil
}

// PromptsFromInteractive converts multiple OpenAPI prompts to PromptSpecs.
func PromptsFromInteractive(interactive *openapi.CLIInteractive, loader *OptionLoader) ([]*PromptSpec, error) {
	if interactive == nil || !interactive.Enabled {
		return nil, nil
	}

	specs := make([]*PromptSpec, 0, len(interactive.Prompts))
	for _, prompt := range interactive.Prompts {
		spec, err := PromptFromInteractive(prompt, loader)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	return specs, nil
}
