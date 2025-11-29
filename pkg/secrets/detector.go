// Package secrets provides multi-layer secret detection and masking for CliForge.
// It prevents sensitive data from being logged to stdout/stderr/logfiles.
package secrets

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/CliForge/cliforge/pkg/cli"
)

// Detector detects sensitive data using multiple strategies.
type Detector struct {
	config         *cli.SecretsBehavior
	fieldPatterns  []*regexp.Regexp
	valuePatterns  []*compiledValuePattern
	explicitPaths  []string
	headerPatterns map[string]bool
}

// compiledValuePattern wraps a value pattern with its compiled regex.
type compiledValuePattern struct {
	name    string
	pattern *regexp.Regexp
	enabled bool
}

// DetectionResult contains information about detected secrets.
type DetectionResult struct {
	IsSecret    bool
	DetectedBy  string // "field_pattern", "value_pattern", "explicit_path", "header", "x-cli-secret"
	PatternName string // Name of the pattern that matched
	FieldPath   string // JSONPath-like field path
	MaskedValue string // Pre-masked value for convenience
}

// NewDetector creates a new secret detector from configuration.
func NewDetector(config *cli.SecretsBehavior) (*Detector, error) {
	if config == nil {
		return &Detector{
			config:         &cli.SecretsBehavior{Enabled: false},
			fieldPatterns:  []*regexp.Regexp{},
			valuePatterns:  []*compiledValuePattern{},
			explicitPaths:  []string{},
			headerPatterns: make(map[string]bool),
		}, nil
	}

	d := &Detector{
		config:         config,
		fieldPatterns:  make([]*regexp.Regexp, 0, len(config.FieldPatterns)),
		valuePatterns:  make([]*compiledValuePattern, 0, len(config.ValuePatterns)),
		explicitPaths:  config.ExplicitFields,
		headerPatterns: make(map[string]bool),
	}

	// Compile field patterns (glob-style wildcards)
	for _, pattern := range config.FieldPatterns {
		// Convert glob pattern to regex
		regex, err := globToRegex(pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid field pattern %q: %w", pattern, err)
		}
		d.fieldPatterns = append(d.fieldPatterns, regex)
	}

	// Compile value patterns (regex)
	for _, vp := range config.ValuePatterns {
		regex, err := regexp.Compile(vp.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid value pattern %q: %w", vp.Pattern, err)
		}
		d.valuePatterns = append(d.valuePatterns, &compiledValuePattern{
			name:    vp.Name,
			pattern: regex,
			enabled: vp.Enabled,
		})
	}

	// Build header lookup map (case-insensitive)
	for _, header := range config.Headers {
		d.headerPatterns[strings.ToLower(header)] = true
	}

	return d, nil
}

// IsEnabled returns whether secret detection is enabled.
func (d *Detector) IsEnabled() bool {
	return d.config != nil && d.config.Enabled
}

// IsSecretField checks if a field name indicates a secret using field patterns.
func (d *Detector) IsSecretField(fieldName string) bool {
	if !d.IsEnabled() {
		return false
	}

	lowerField := strings.ToLower(fieldName)

	// Check configured field patterns
	for _, pattern := range d.fieldPatterns {
		if pattern.MatchString(lowerField) {
			return true
		}
	}

	return false
}

// IsSecretValue checks if a value looks like a secret using value patterns.
func (d *Detector) IsSecretValue(value string) (bool, string) {
	if !d.IsEnabled() {
		return false, ""
	}

	// Check configured value patterns
	for _, vp := range d.valuePatterns {
		if !vp.enabled {
			continue
		}

		if vp.pattern.MatchString(value) {
			return true, vp.name
		}
	}

	return false, ""
}

// IsSecretHeader checks if a header should be masked.
func (d *Detector) IsSecretHeader(headerName string) bool {
	if !d.IsEnabled() {
		return false
	}

	return d.headerPatterns[strings.ToLower(headerName)]
}

// DetectInField performs full detection on a field name and value.
func (d *Detector) DetectInField(fieldName string, value interface{}) *DetectionResult {
	result := &DetectionResult{
		IsSecret:  false,
		FieldPath: fieldName,
	}

	if !d.IsEnabled() {
		return result
	}

	// Convert value to string for analysis
	strValue := fmt.Sprint(value)

	// Check 1: Field name pattern matching
	if d.IsSecretField(fieldName) {
		result.IsSecret = true
		result.DetectedBy = "field_pattern"
		result.PatternName = fieldName
		return result
	}

	// Check 2: Value pattern matching
	if isSecret, patternName := d.IsSecretValue(strValue); isSecret {
		result.IsSecret = true
		result.DetectedBy = "value_pattern"
		result.PatternName = patternName
		return result
	}

	// Check 3: Explicit path matching (if supported)
	if d.matchesExplicitPath(fieldName) {
		result.IsSecret = true
		result.DetectedBy = "explicit_path"
		result.PatternName = fieldName
		return result
	}

	return result
}

// MaskJSON recursively masks secrets in JSON-like data structures.
func (d *Detector) MaskJSON(data interface{}) interface{} {
	return d.maskJSONRecursive(data, "")
}

// maskJSONRecursive recursively traverses and masks JSON data.
func (d *Detector) maskJSONRecursive(data interface{}, path string) interface{} {
	if !d.IsEnabled() {
		return data
	}

	switch v := data.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(v))
		for key, value := range v {
			fieldPath := buildPath(path, key)

			// Detect if this field should be masked
			detection := d.DetectInField(key, value)

			if detection.IsSecret {
				// Mask the value
				result[key] = MaskValue(fmt.Sprint(value), d.config.Masking)
			} else {
				// Recurse into nested structures
				result[key] = d.maskJSONRecursive(value, fieldPath)
			}
		}
		return result

	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = d.maskJSONRecursive(item, fmt.Sprintf("%s[%d]", path, i))
		}
		return result

	case string:
		// Check if the value itself looks like a secret
		if isSecret, _ := d.IsSecretValue(v); isSecret {
			return MaskValue(v, d.config.Masking)
		}
		return v

	default:
		return v
	}
}

// MaskString masks potential secrets in plain text strings.
// This is useful for masking secrets in log messages or error output.
func (d *Detector) MaskString(text string) string {
	if !d.IsEnabled() {
		return text
	}

	result := text

	// Apply value pattern masking
	for _, vp := range d.valuePatterns {
		if !vp.enabled {
			continue
		}

		// Replace all matches with masked values
		result = vp.pattern.ReplaceAllStringFunc(result, func(match string) string {
			return MaskValue(match, d.config.Masking)
		})
	}

	return result
}

// MaskHeaders masks sensitive HTTP headers.
func (d *Detector) MaskHeaders(headers map[string][]string) map[string][]string {
	if !d.IsEnabled() {
		return headers
	}

	result := make(map[string][]string, len(headers))

	for key, values := range headers {
		if d.IsSecretHeader(key) {
			// Mask all values for this header
			maskedValues := make([]string, len(values))
			for i, v := range values {
				maskedValues[i] = MaskValue(v, d.config.Masking)
			}
			result[key] = maskedValues
		} else {
			result[key] = values
		}
	}

	return result
}

// MaskJSONString masks secrets in a JSON string.
func (d *Detector) MaskJSONString(jsonStr string) (string, error) {
	if !d.IsEnabled() {
		return jsonStr, nil
	}

	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		// If not valid JSON, try masking as plain text
		return d.MaskString(jsonStr), nil
	}

	masked := d.MaskJSON(data)

	maskedBytes, err := json.Marshal(masked)
	if err != nil {
		return "", fmt.Errorf("failed to marshal masked JSON: %w", err)
	}

	return string(maskedBytes), nil
}

// ShouldMaskInContext checks if masking should be applied in a given context.
func (d *Detector) ShouldMaskInContext(context string) bool {
	if !d.IsEnabled() || d.config.MaskIn == nil {
		return false
	}

	switch context {
	case "stdout":
		return d.config.MaskIn.Stdout
	case "stderr":
		return d.config.MaskIn.Stderr
	case "logs":
		return d.config.MaskIn.Logs
	case "audit":
		return d.config.MaskIn.Audit
	case "debug":
		return d.config.MaskIn.DebugOutput
	default:
		return true // Mask by default in unknown contexts
	}
}

// Helper functions

// globToRegex converts a glob-style pattern to a regex.
// Supports: * (any chars), ? (single char), and literal strings.
func globToRegex(pattern string) (*regexp.Regexp, error) {
	// Use filepath.Match compatible patterns but convert to regex
	// for consistent case-insensitive matching

	// Escape special regex characters except * and ?
	escaped := regexp.QuoteMeta(pattern)

	// Convert glob wildcards to regex
	escaped = strings.ReplaceAll(escaped, `\*`, ".*")
	escaped = strings.ReplaceAll(escaped, `\?`, ".")

	// Make it case-insensitive and match the whole string
	regexPattern := "(?i)^" + escaped + "$"

	return regexp.Compile(regexPattern)
}

// matchesExplicitPath checks if a field path matches any explicit path patterns.
func (d *Detector) matchesExplicitPath(path string) bool {
	for _, pattern := range d.explicitPaths {
		// Simple JSONPath-like matching
		// Support: $.field, $.*.field, $.parent.child
		if matchPath(pattern, path) {
			return true
		}
	}
	return false
}

// matchPath performs simple JSONPath-like pattern matching.
func matchPath(pattern, path string) bool {
	// Remove leading $. if present
	pattern = strings.TrimPrefix(pattern, "$.")
	path = strings.TrimPrefix(path, "$.")

	// Convert to lowercase for case-insensitive matching
	pattern = strings.ToLower(pattern)
	path = strings.ToLower(path)

	// Exact match
	if pattern == path {
		return true
	}

	// Wildcard support
	matched, _ := filepath.Match(pattern, path)
	return matched
}

// buildPath constructs a JSONPath-like path string.
func buildPath(parent, key string) string {
	if parent == "" {
		return key
	}
	return parent + "." + key
}
