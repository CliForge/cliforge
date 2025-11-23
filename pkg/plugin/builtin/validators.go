package builtin

import (
	"context"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/plugin"
)

// ValidatorsPlugin provides custom validation logic.
type ValidatorsPlugin struct{}

// NewValidatorsPlugin creates a new validators plugin.
func NewValidatorsPlugin() *ValidatorsPlugin {
	return &ValidatorsPlugin{}
}

// Execute performs validation.
func (p *ValidatorsPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
	validator, ok := input.Data["validator"].(string)
	if !ok {
		return nil, fmt.Errorf("validator is required")
	}

	value, ok := input.Data["value"]
	if !ok {
		return nil, fmt.Errorf("value is required")
	}

	startTime := time.Now()

	var valid bool
	var message string
	var err error

	switch validator {
	case "regex":
		valid, message, err = p.validateRegex(value, input.Data)
	case "email":
		valid, message, err = p.validateEmail(value)
	case "url":
		valid, message, err = p.validateURL(value)
	case "ip":
		valid, message, err = p.validateIP(value)
	case "cidr":
		valid, message, err = p.validateCIDR(value)
	case "cluster-name":
		valid, message, err = p.validateClusterName(value)
	case "dns-label":
		valid, message, err = p.validateDNSLabel(value)
	case "length":
		valid, message, err = p.validateLength(value, input.Data)
	case "range":
		valid, message, err = p.validateRange(value, input.Data)
	case "enum":
		valid, message, err = p.validateEnum(value, input.Data)
	case "format":
		valid, message, err = p.validateFormat(value, input.Data)
	default:
		return nil, fmt.Errorf("unknown validator: %s", validator)
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
		Data: map[string]interface{}{
			"valid":   valid,
			"message": message,
		},
		Duration: duration,
	}, nil
}

// Validate checks if the plugin is properly configured.
func (p *ValidatorsPlugin) Validate() error {
	return nil
}

// Describe returns plugin metadata.
func (p *ValidatorsPlugin) Describe() *plugin.PluginInfo {
	manifest := plugin.PluginManifest{
		Name:        "validators",
		Version:     "1.0.0",
		Type:        plugin.PluginTypeBuiltin,
		Description: "Custom validation logic for various data types",
		Author:      "CliForge",
		Permissions: []plugin.Permission{},
	}

	return &plugin.PluginInfo{
		Manifest: manifest,
		Capabilities: []string{
			"regex", "email", "url", "ip", "cidr",
			"cluster-name", "dns-label", "length", "range", "enum", "format",
		},
		Status: plugin.PluginStatusReady,
	}
}

// validateRegex validates against a regex pattern.
func (p *ValidatorsPlugin) validateRegex(value interface{}, data map[string]interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	pattern, ok := data["pattern"].(string)
	if !ok {
		return false, "", fmt.Errorf("pattern is required")
	}

	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false, "", fmt.Errorf("invalid regex pattern: %w", err)
	}

	if !matched {
		return false, fmt.Sprintf("value does not match pattern: %s", pattern), nil
	}

	return true, "valid", nil
}

// validateEmail validates an email address.
func (p *ValidatorsPlugin) validateEmail(value interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	_, err := mail.ParseAddress(str)
	if err != nil {
		return false, "invalid email address", nil
	}

	return true, "valid email address", nil
}

// validateURL validates a URL.
func (p *ValidatorsPlugin) validateURL(value interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	parsedURL, err := url.Parse(str)
	if err != nil {
		return false, "invalid URL", nil
	}

	if parsedURL.Scheme == "" {
		return false, "URL must have a scheme (http, https, etc.)", nil
	}

	if parsedURL.Host == "" {
		return false, "URL must have a host", nil
	}

	return true, "valid URL", nil
}

// validateIP validates an IP address.
func (p *ValidatorsPlugin) validateIP(value interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	ip := net.ParseIP(str)
	if ip == nil {
		return false, "invalid IP address", nil
	}

	return true, "valid IP address", nil
}

// validateCIDR validates a CIDR notation.
func (p *ValidatorsPlugin) validateCIDR(value interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	_, _, err := net.ParseCIDR(str)
	if err != nil {
		return false, "invalid CIDR notation", nil
	}

	return true, "valid CIDR", nil
}

// validateClusterName validates a Kubernetes-style cluster name.
func (p *ValidatorsPlugin) validateClusterName(value interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	// Cluster name rules:
	// - 1-54 characters
	// - Start with lowercase letter
	// - Contain only lowercase letters, numbers, and hyphens
	// - End with letter or number
	pattern := `^[a-z][a-z0-9-]{0,52}[a-z0-9]$`
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false, "", err
	}

	if !matched {
		return false, "invalid cluster name: must be 1-54 characters, start with letter, contain only lowercase letters, numbers, and hyphens", nil
	}

	return true, "valid cluster name", nil
}

// validateDNSLabel validates a DNS label (RFC 1123).
func (p *ValidatorsPlugin) validateDNSLabel(value interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	// DNS label rules (RFC 1123):
	// - 1-63 characters
	// - Start with alphanumeric
	// - Contain only alphanumeric and hyphens
	// - End with alphanumeric
	pattern := `^[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$`
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return false, "", err
	}

	if !matched {
		return false, "invalid DNS label: must be 1-63 characters, alphanumeric and hyphens only", nil
	}

	return true, "valid DNS label", nil
}

// validateLength validates string length.
func (p *ValidatorsPlugin) validateLength(value interface{}, data map[string]interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	length := len(str)

	if minLen, ok := data["min"]; ok {
		min := int(minLen.(float64))
		if length < min {
			return false, fmt.Sprintf("length must be at least %d characters", min), nil
		}
	}

	if maxLen, ok := data["max"]; ok {
		max := int(maxLen.(float64))
		if length > max {
			return false, fmt.Sprintf("length must be at most %d characters", max), nil
		}
	}

	return true, "valid length", nil
}

// validateRange validates numeric range.
func (p *ValidatorsPlugin) validateRange(value interface{}, data map[string]interface{}) (bool, string, error) {
	var num float64

	switch v := value.(type) {
	case float64:
		num = v
	case int:
		num = float64(v)
	case string:
		var err error
		num, err = strconv.ParseFloat(v, 64)
		if err != nil {
			return false, "value must be a number", nil
		}
	default:
		return false, "value must be a number", nil
	}

	if minVal, ok := data["min"]; ok {
		min := minVal.(float64)
		if num < min {
			return false, fmt.Sprintf("value must be at least %v", min), nil
		}
	}

	if maxVal, ok := data["max"]; ok {
		max := maxVal.(float64)
		if num > max {
			return false, fmt.Sprintf("value must be at most %v", max), nil
		}
	}

	return true, "valid range", nil
}

// validateEnum validates against a set of allowed values.
func (p *ValidatorsPlugin) validateEnum(value interface{}, data map[string]interface{}) (bool, string, error) {
	allowedInterface, ok := data["allowed"]
	if !ok {
		return false, "", fmt.Errorf("allowed values are required")
	}

	allowed, ok := allowedInterface.([]interface{})
	if !ok {
		return false, "", fmt.Errorf("allowed must be an array")
	}

	valueStr := fmt.Sprintf("%v", value)

	for _, allowedVal := range allowed {
		if fmt.Sprintf("%v", allowedVal) == valueStr {
			return true, "valid enum value", nil
		}
	}

	allowedStrs := make([]string, len(allowed))
	for i, v := range allowed {
		allowedStrs[i] = fmt.Sprintf("%v", v)
	}

	return false, fmt.Sprintf("value must be one of: %s", strings.Join(allowedStrs, ", ")), nil
}

// validateFormat validates against common formats.
func (p *ValidatorsPlugin) validateFormat(value interface{}, data map[string]interface{}) (bool, string, error) {
	str, ok := value.(string)
	if !ok {
		return false, "value must be a string", nil
	}

	format, ok := data["format"].(string)
	if !ok {
		return false, "", fmt.Errorf("format is required")
	}

	switch format {
	case "uuid":
		return p.validateUUID(str)
	case "date":
		return p.validateDate(str)
	case "time":
		return p.validateTime(str)
	case "datetime":
		return p.validateDateTime(str)
	case "semver":
		return p.validateSemVer(str)
	default:
		return false, "", fmt.Errorf("unsupported format: %s", format)
	}
}

// validateUUID validates a UUID.
func (p *ValidatorsPlugin) validateUUID(str string) (bool, string, error) {
	pattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(pattern, strings.ToLower(str))
	if !matched {
		return false, "invalid UUID format", nil
	}
	return true, "valid UUID", nil
}

// validateDate validates a date in YYYY-MM-DD format.
func (p *ValidatorsPlugin) validateDate(str string) (bool, string, error) {
	_, err := time.Parse("2006-01-02", str)
	if err != nil {
		return false, "invalid date format (expected YYYY-MM-DD)", nil
	}
	return true, "valid date", nil
}

// validateTime validates a time in HH:MM:SS format.
func (p *ValidatorsPlugin) validateTime(str string) (bool, string, error) {
	_, err := time.Parse("15:04:05", str)
	if err != nil {
		return false, "invalid time format (expected HH:MM:SS)", nil
	}
	return true, "valid time", nil
}

// validateDateTime validates a datetime in RFC3339 format.
func (p *ValidatorsPlugin) validateDateTime(str string) (bool, string, error) {
	_, err := time.Parse(time.RFC3339, str)
	if err != nil {
		return false, "invalid datetime format (expected RFC3339)", nil
	}
	return true, "valid datetime", nil
}

// validateSemVer validates semantic versioning.
func (p *ValidatorsPlugin) validateSemVer(str string) (bool, string, error) {
	pattern := `^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`
	matched, _ := regexp.MatchString(pattern, str)
	if !matched {
		return false, "invalid semantic version", nil
	}
	return true, "valid semantic version", nil
}
