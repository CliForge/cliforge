package builtin

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/plugin"
	"gopkg.in/yaml.v3"
)

// TransformersPlugin provides data transformation capabilities.
type TransformersPlugin struct{}

// NewTransformersPlugin creates a new transformers plugin.
func NewTransformersPlugin() *TransformersPlugin {
	return &TransformersPlugin{}
}

// Execute performs data transformation.
func (p *TransformersPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
	transformation, ok := input.Data["transformation"].(string)
	if !ok {
		return nil, fmt.Errorf("transformation is required")
	}

	startTime := time.Now()

	var result map[string]interface{}
	var err error

	switch transformation {
	case "json-to-yaml":
		result, err = p.jsonToYAML(input.Data)
	case "yaml-to-json":
		result, err = p.yamlToJSON(input.Data)
	case "base64-encode":
		result, err = p.base64Encode(input.Data)
	case "base64-decode":
		result, err = p.base64Decode(input.Data)
	case "htpasswd-to-users":
		result, err = p.htpasswdToUsers(input.Data)
	case "users-to-htpasswd":
		result, err = p.usersToHTPasswd(input.Data)
	case "extract-field":
		result, err = p.extractField(input.Data)
	case "merge":
		result, err = p.mergeData(input.Data)
	case "filter":
		result, err = p.filterData(input.Data)
	case "template":
		result, err = p.applyTemplate(input.Data)
	default:
		return nil, fmt.Errorf("unknown transformation: %s", transformation)
	}

	duration := time.Since(startTime)

	if err != nil {
		return &plugin.PluginOutput{
			ExitCode: 1,
			Error:    err.Error(),
			Duration: duration,
		}, nil
	}

	return &plugin.PluginOutput{
		ExitCode: 0,
		Data:     result,
		Duration: duration,
	}, nil
}

// Validate checks if the plugin is properly configured.
func (p *TransformersPlugin) Validate() error {
	return nil
}

// Describe returns plugin metadata.
func (p *TransformersPlugin) Describe() *plugin.PluginInfo {
	manifest := plugin.PluginManifest{
		Name:        "transformers",
		Version:     "1.0.0",
		Type:        plugin.PluginTypeBuiltin,
		Description: "Data transformation utilities",
		Author:      "CliForge",
		Permissions: []plugin.Permission{},
	}

	return &plugin.PluginInfo{
		Manifest: manifest,
		Capabilities: []string{
			"json-to-yaml", "yaml-to-json",
			"base64-encode", "base64-decode",
			"htpasswd-to-users", "users-to-htpasswd",
			"extract-field", "merge", "filter", "template",
		},
		Status: plugin.PluginStatusReady,
	}
}

// jsonToYAML converts JSON to YAML.
func (p *TransformersPlugin) jsonToYAML(data map[string]interface{}) (map[string]interface{}, error) {
	input, ok := data["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is required")
	}

	var jsonData interface{}
	if err := json.Unmarshal([]byte(input), &jsonData); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	yamlBytes, err := yaml.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to YAML: %w", err)
	}

	return map[string]interface{}{
		"output": string(yamlBytes),
		"format": "yaml",
	}, nil
}

// yamlToJSON converts YAML to JSON.
func (p *TransformersPlugin) yamlToJSON(data map[string]interface{}) (map[string]interface{}, error) {
	input, ok := data["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is required")
	}

	var yamlData interface{}
	if err := yaml.Unmarshal([]byte(input), &yamlData); err != nil {
		return nil, fmt.Errorf("invalid YAML: %w", err)
	}

	jsonBytes, err := json.MarshalIndent(yamlData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to convert to JSON: %w", err)
	}

	return map[string]interface{}{
		"output": string(jsonBytes),
		"format": "json",
	}, nil
}

// base64Encode encodes data to base64.
func (p *TransformersPlugin) base64Encode(data map[string]interface{}) (map[string]interface{}, error) {
	input, ok := data["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is required")
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(input))

	return map[string]interface{}{
		"output": encoded,
	}, nil
}

// base64Decode decodes base64 data.
func (p *TransformersPlugin) base64Decode(data map[string]interface{}) (map[string]interface{}, error) {
	input, ok := data["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is required")
	}

	decoded, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return nil, fmt.Errorf("invalid base64: %w", err)
	}

	return map[string]interface{}{
		"output": string(decoded),
	}, nil
}

// htpasswdToUsers converts htpasswd format to user list.
func (p *TransformersPlugin) htpasswdToUsers(data map[string]interface{}) (map[string]interface{}, error) {
	input, ok := data["input"].(string)
	if !ok {
		return nil, fmt.Errorf("input is required")
	}

	lines := strings.Split(input, "\n")
	users := make([]map[string]string, 0)

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid htpasswd format at line %d", i+1)
		}

		users = append(users, map[string]string{
			"username": parts[0],
			"password": parts[1],
		})
	}

	return map[string]interface{}{
		"users": users,
		"count": len(users),
	}, nil
}

// usersToHTPasswd converts user list to htpasswd format.
func (p *TransformersPlugin) usersToHTPasswd(data map[string]interface{}) (map[string]interface{}, error) {
	usersInterface, ok := data["users"]
	if !ok {
		return nil, fmt.Errorf("users is required")
	}

	users, ok := usersInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("users must be an array")
	}

	lines := make([]string, 0, len(users))
	for i, userInterface := range users {
		user, ok := userInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid user format at index %d", i)
		}

		username, ok := user["username"].(string)
		if !ok {
			return nil, fmt.Errorf("username is required for user at index %d", i)
		}

		password, ok := user["password"].(string)
		if !ok {
			return nil, fmt.Errorf("password is required for user at index %d", i)
		}

		lines = append(lines, fmt.Sprintf("%s:%s", username, password))
	}

	output := strings.Join(lines, "\n")

	return map[string]interface{}{
		"output": output,
		"count":  len(lines),
	}, nil
}

// extractField extracts a field from structured data.
func (p *TransformersPlugin) extractField(data map[string]interface{}) (map[string]interface{}, error) {
	inputInterface, ok := data["input"]
	if !ok {
		return nil, fmt.Errorf("input is required")
	}

	field, ok := data["field"].(string)
	if !ok {
		return nil, fmt.Errorf("field is required")
	}

	// Parse input if it's a string
	var inputData interface{}
	if inputStr, ok := inputInterface.(string); ok {
		// Try JSON first
		if err := json.Unmarshal([]byte(inputStr), &inputData); err != nil {
			// Try YAML
			if err := yaml.Unmarshal([]byte(inputStr), &inputData); err != nil {
				return nil, fmt.Errorf("input must be valid JSON or YAML")
			}
		}
	} else {
		inputData = inputInterface
	}

	// Extract field using dot notation
	value, err := extractFieldValue(inputData, field)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"value": value,
		"field": field,
	}, nil
}

// mergeData merges multiple data objects.
func (p *TransformersPlugin) mergeData(data map[string]interface{}) (map[string]interface{}, error) {
	sourcesInterface, ok := data["sources"]
	if !ok {
		return nil, fmt.Errorf("sources is required")
	}

	sources, ok := sourcesInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("sources must be an array")
	}

	result := make(map[string]interface{})

	for _, sourceInterface := range sources {
		source, ok := sourceInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("each source must be an object")
		}

		// Merge source into result
		for k, v := range source {
			result[k] = v
		}
	}

	return map[string]interface{}{
		"merged": result,
	}, nil
}

// filterData filters data based on criteria.
func (p *TransformersPlugin) filterData(data map[string]interface{}) (map[string]interface{}, error) {
	inputInterface, ok := data["input"]
	if !ok {
		return nil, fmt.Errorf("input is required")
	}

	criteriaInterface, ok := data["criteria"]
	if !ok {
		return nil, fmt.Errorf("criteria is required")
	}

	criteria, ok := criteriaInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("criteria must be an object")
	}

	// If input is an array, filter items
	if inputArray, ok := inputInterface.([]interface{}); ok {
		filtered := make([]interface{}, 0)
		for _, item := range inputArray {
			if matchesCriteria(item, criteria) {
				filtered = append(filtered, item)
			}
		}
		return map[string]interface{}{
			"filtered": filtered,
			"count":    len(filtered),
		}, nil
	}

	// If input is an object, filter fields
	if inputMap, ok := inputInterface.(map[string]interface{}); ok {
		filtered := make(map[string]interface{})
		for k, v := range inputMap {
			if matchesCriteria(v, criteria) {
				filtered[k] = v
			}
		}
		return map[string]interface{}{
			"filtered": filtered,
			"count":    len(filtered),
		}, nil
	}

	return nil, fmt.Errorf("input must be an array or object")
}

// applyTemplate applies a simple string template.
func (p *TransformersPlugin) applyTemplate(data map[string]interface{}) (map[string]interface{}, error) {
	template, ok := data["template"].(string)
	if !ok {
		return nil, fmt.Errorf("template is required")
	}

	valuesInterface, ok := data["values"]
	if !ok {
		return nil, fmt.Errorf("values is required")
	}

	values, ok := valuesInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("values must be an object")
	}

	// Simple template replacement (supports {{key}} syntax)
	result := template
	for k, v := range values {
		placeholder := fmt.Sprintf("{{%s}}", k)
		replacement := fmt.Sprintf("%v", v)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return map[string]interface{}{
		"output": result,
	}, nil
}

// extractFieldValue extracts a field value using dot notation.
func extractFieldValue(data interface{}, field string) (interface{}, error) {
	parts := strings.Split(field, ".")

	current := data
	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field not found: %s", field)
			}
		case map[interface{}]interface{}:
			// Handle YAML-style maps
			var ok bool
			current, ok = v[part]
			if !ok {
				return nil, fmt.Errorf("field not found: %s", field)
			}
		default:
			return nil, fmt.Errorf("cannot access field %s in non-object", part)
		}
	}

	return current, nil
}

// matchesCriteria checks if data matches criteria.
func matchesCriteria(data interface{}, criteria map[string]interface{}) bool {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return false
	}

	for key, expectedValue := range criteria {
		actualValue, exists := dataMap[key]
		if !exists {
			return false
		}

		if fmt.Sprintf("%v", actualValue) != fmt.Sprintf("%v", expectedValue) {
			return false
		}
	}

	return true
}
