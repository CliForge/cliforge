// Package deprecation provides deprecation detection and management for CliForge.
package deprecation

import (
	"fmt"
	"net/http"
	"time"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/getkin/kin-openapi/openapi3"
)

// Detector detects deprecated operations from OpenAPI specifications.
type Detector struct {
	// Enable detection from various sources
	CheckDeprecatedField bool // Standard OpenAPI deprecated field
	CheckExtension       bool // x-cli-deprecation extension
	CheckSunsetHeader    bool // Sunset header from HTTP responses
}

// NewDetector creates a new Detector with default settings.
func NewDetector() *Detector {
	return &Detector{
		CheckDeprecatedField: true,
		CheckExtension:       true,
		CheckSunsetHeader:    true,
	}
}

// DeprecationType indicates what is deprecated.
type DeprecationType string

const (
	DeprecationTypeOperation DeprecationType = "operation"
	DeprecationTypeParameter DeprecationType = "parameter"
	DeprecationTypeSchema    DeprecationType = "schema"
	DeprecationTypeFlag      DeprecationType = "flag"
	DeprecationTypeCommand   DeprecationType = "command"
)

// DeprecationInfo contains complete deprecation information.
type DeprecationInfo struct {
	// Type of deprecation
	Type DeprecationType

	// Basic info
	Deprecated  bool
	OperationID string
	Path        string
	Method      string
	Name        string // Parameter/schema/flag name

	// Sunset information
	Sunset        *time.Time
	DaysRemaining int

	// Replacement information
	Replacement *Replacement

	// Documentation
	Reason      string
	DocsURL     string
	Migration   string
	Severity    Severity
	Level       WarningLevel // Calculated based on days remaining

	// Breaking changes
	BreakingChanges []string

	// Source of detection
	DetectedFrom string // "deprecated-field", "x-cli-deprecation", "sunset-header"
}

// Replacement provides information about what to use instead.
type Replacement struct {
	OperationID string
	Path        string
	Command     string
	Migration   string
	Example     string
}

// Severity indicates the impact of the deprecation.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityBreaking Severity = "breaking"
)

// DetectOperation checks if an operation is deprecated.
func (d *Detector) DetectOperation(op *openapi3.Operation, operationID, method, path string) *DeprecationInfo {
	if op == nil {
		return nil
	}

	info := &DeprecationInfo{
		Type:        DeprecationTypeOperation,
		OperationID: operationID,
		Path:        path,
		Method:      method,
	}

	// Check standard deprecated field
	if d.CheckDeprecatedField && op.Deprecated {
		info.Deprecated = true
		info.DetectedFrom = "deprecated-field"

		// Extract deprecation details from description
		if op.Description != "" {
			info.Migration = extractMigrationFromDescription(op.Description)
		}
	}

	// Check x-cli-deprecation extension
	if d.CheckExtension {
		if ext, ok := op.Extensions["x-cli-deprecation"]; ok {
			info.Deprecated = true
			if info.DetectedFrom == "" {
				info.DetectedFrom = "x-cli-deprecation"
			}
			d.parseDeprecationExtension(ext, info)
		}
	}

	if !info.Deprecated {
		return nil
	}

	// Calculate days remaining if sunset is set
	if info.Sunset != nil {
		info.DaysRemaining = calculateDaysRemaining(*info.Sunset)
		info.Level = calculateWarningLevel(info.DaysRemaining)
	} else {
		// No sunset specified, default to warning level
		info.Level = WarningLevelWarning
	}

	// Set default severity if not specified
	if info.Severity == "" {
		info.Severity = SeverityWarning
	}

	return info
}

// DetectParameter checks if a parameter is deprecated.
func (d *Detector) DetectParameter(param *openapi3.Parameter, operationID, method, path string) *DeprecationInfo {
	if param == nil {
		return nil
	}

	info := &DeprecationInfo{
		Type:        DeprecationTypeParameter,
		OperationID: operationID,
		Path:        path,
		Method:      method,
		Name:        param.Name,
	}

	// Check standard deprecated field
	if d.CheckDeprecatedField && param.Deprecated {
		info.Deprecated = true
		info.DetectedFrom = "deprecated-field"

		// Extract migration from description
		if param.Description != "" {
			info.Migration = extractMigrationFromDescription(param.Description)
		}
	}

	// Check x-cli-deprecation extension
	if d.CheckExtension {
		if ext, ok := param.Extensions["x-cli-deprecation"]; ok {
			info.Deprecated = true
			if info.DetectedFrom == "" {
				info.DetectedFrom = "x-cli-deprecation"
			}
			d.parseDeprecationExtension(ext, info)
		}
	}

	if !info.Deprecated {
		return nil
	}

	// Calculate warning level
	if info.Sunset != nil {
		info.DaysRemaining = calculateDaysRemaining(*info.Sunset)
		info.Level = calculateWarningLevel(info.DaysRemaining)
	} else {
		info.Level = WarningLevelWarning
	}

	if info.Severity == "" {
		info.Severity = SeverityWarning
	}

	return info
}

// DetectFromSpec scans an entire spec for deprecations.
func (d *Detector) DetectFromSpec(spec *openapi.ParsedSpec) ([]*DeprecationInfo, error) {
	var deprecations []*DeprecationInfo

	if spec.Spec.Paths == nil {
		return deprecations, nil
	}

	// Scan all operations
	for path, pathItem := range spec.Spec.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}

			operationID := operation.OperationID
			if operationID == "" {
				operationID = fmt.Sprintf("%s_%s", method, path)
			}

			// Check operation itself
			if info := d.DetectOperation(operation, operationID, method, path); info != nil {
				deprecations = append(deprecations, info)
			}

			// Check parameters
			for _, paramRef := range operation.Parameters {
				if paramRef.Value != nil {
					if info := d.DetectParameter(paramRef.Value, operationID, method, path); info != nil {
						deprecations = append(deprecations, info)
					}
				}
			}
		}
	}

	// Check for deprecations from changelog
	if spec.Extensions != nil && len(spec.Extensions.Changelog) > 0 {
		changelogDeps := d.detectFromChangelog(spec.Extensions.Changelog)
		deprecations = append(deprecations, changelogDeps...)
	}

	return deprecations, nil
}

// detectFromChangelog extracts deprecation info from changelog entries.
func (d *Detector) detectFromChangelog(entries []openapi.ChangelogEntry) []*DeprecationInfo {
	var deprecations []*DeprecationInfo

	for _, entry := range entries {
		for _, change := range entry.Changes {
			if change.Type == "deprecated" {
				info := &DeprecationInfo{
					Deprecated:   true,
					Path:         change.Path,
					Migration:    change.Migration,
					DetectedFrom: "x-cli-changelog",
				}

				// Parse sunset date
				if change.Sunset != "" {
					if t, err := time.Parse("2006-01-02", change.Sunset); err == nil {
						info.Sunset = &t
						info.DaysRemaining = calculateDaysRemaining(t)
						info.Level = calculateWarningLevel(info.DaysRemaining)
					}
				}

				// Set severity
				switch change.Severity {
				case "breaking":
					info.Severity = SeverityBreaking
				case "dangerous":
					info.Severity = SeverityWarning
				default:
					info.Severity = SeverityInfo
				}

				deprecations = append(deprecations, info)
			}
		}
	}

	return deprecations
}

// parseDeprecationExtension parses the x-cli-deprecation extension.
func (d *Detector) parseDeprecationExtension(ext interface{}, info *DeprecationInfo) {
	extMap, ok := ext.(map[string]interface{})
	if !ok {
		return
	}

	// Parse sunset date
	if sunset, ok := extMap["sunset"].(string); ok {
		// Try multiple date formats
		formats := []string{
			"2006-01-02",
			time.RFC3339,
			"2006-01-02T15:04:05Z07:00",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, sunset); err == nil {
				info.Sunset = &t
				break
			}
		}
	}

	// Parse replacement
	if replData, ok := extMap["replacement"].(map[string]interface{}); ok {
		repl := &Replacement{}
		if opID, ok := replData["operation"].(string); ok {
			repl.OperationID = opID
		}
		if path, ok := replData["path"].(string); ok {
			repl.Path = path
		}
		if cmd, ok := replData["command"].(string); ok {
			repl.Command = cmd
		}
		if migration, ok := replData["migration"].(string); ok {
			repl.Migration = migration
		}
		if example, ok := replData["example"].(string); ok {
			repl.Example = example
		}
		info.Replacement = repl
	}

	// Parse reason
	if reason, ok := extMap["reason"].(string); ok {
		info.Reason = reason
	}

	// Parse docs URL
	if docsURL, ok := extMap["docs_url"].(string); ok {
		info.DocsURL = docsURL
	}

	// Parse severity
	if severity, ok := extMap["severity"].(string); ok {
		switch severity {
		case "info":
			info.Severity = SeverityInfo
		case "warning":
			info.Severity = SeverityWarning
		case "breaking":
			info.Severity = SeverityBreaking
		}
	}

	// Parse migration
	if migration, ok := extMap["migration"].(string); ok {
		info.Migration = migration
	}

	// Parse breaking changes
	if changes, ok := extMap["breaking_changes"].([]interface{}); ok {
		for _, change := range changes {
			if str, ok := change.(string); ok {
				info.BreakingChanges = append(info.BreakingChanges, str)
			}
		}
	}
}

// DetectFromSunsetHeader checks for Sunset header in HTTP response (RFC 8594).
func (d *Detector) DetectFromSunsetHeader(resp *http.Response, operationID, method, path string) *DeprecationInfo {
	if !d.CheckSunsetHeader || resp == nil {
		return nil
	}

	sunsetHeader := resp.Header.Get("Sunset")
	deprecationHeader := resp.Header.Get("Deprecation")

	if sunsetHeader == "" && deprecationHeader == "" {
		return nil
	}

	info := &DeprecationInfo{
		Type:         DeprecationTypeOperation,
		Deprecated:   true,
		OperationID:  operationID,
		Path:         path,
		Method:       method,
		DetectedFrom: "sunset-header",
	}

	// Parse Sunset header (HTTP-date format)
	if sunsetHeader != "" {
		if t, err := http.ParseTime(sunsetHeader); err == nil {
			info.Sunset = &t
			info.DaysRemaining = calculateDaysRemaining(t)
			info.Level = calculateWarningLevel(info.DaysRemaining)
		}
	}

	// Check for Link header with successor-version
	linkHeader := resp.Header.Get("Link")
	if linkHeader != "" {
		successor := extractSuccessorFromLink(linkHeader)
		if successor != "" {
			info.Replacement = &Replacement{
				Path: successor,
			}
		}
	}

	if info.Level == "" {
		info.Level = WarningLevelWarning
	}
	if info.Severity == "" {
		info.Severity = SeverityWarning
	}

	return info
}

// calculateDaysRemaining calculates days until sunset.
func calculateDaysRemaining(sunset time.Time) int {
	duration := time.Until(sunset)
	days := int(duration.Hours() / 24)
	return days
}

// extractMigrationFromDescription attempts to extract migration instructions
// from a description field.
func extractMigrationFromDescription(desc string) string {
	// Look for common patterns like "Use ... instead"
	// This is a simple implementation - could be enhanced with regex
	if len(desc) > 0 {
		return desc
	}
	return ""
}

// extractSuccessorFromLink parses Link header for successor-version.
func extractSuccessorFromLink(linkHeader string) string {
	// Simple implementation - parse Link header
	// Format: </v2/users>; rel="successor-version"
	// A full implementation would use proper Link header parsing
	if len(linkHeader) > 0 {
		// Basic extraction - this could be enhanced
		return ""
	}
	return ""
}

// IsDeprecated checks if any deprecations are present.
func IsDeprecated(deprecations []*DeprecationInfo) bool {
	return len(deprecations) > 0
}

// FilterByType filters deprecations by type.
func FilterByType(deprecations []*DeprecationInfo, depType DeprecationType) []*DeprecationInfo {
	var filtered []*DeprecationInfo
	for _, dep := range deprecations {
		if dep.Type == depType {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

// FilterBySeverity filters deprecations by severity.
func FilterBySeverity(deprecations []*DeprecationInfo, severity Severity) []*DeprecationInfo {
	var filtered []*DeprecationInfo
	for _, dep := range deprecations {
		if dep.Severity == severity {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

// FilterByLevel filters deprecations by warning level.
func FilterByLevel(deprecations []*DeprecationInfo, level WarningLevel) []*DeprecationInfo {
	var filtered []*DeprecationInfo
	for _, dep := range deprecations {
		if dep.Level == level {
			filtered = append(filtered, dep)
		}
	}
	return filtered
}

// GetMostUrgent returns the most urgent deprecation (by warning level).
func GetMostUrgent(deprecations []*DeprecationInfo) *DeprecationInfo {
	if len(deprecations) == 0 {
		return nil
	}

	mostUrgent := deprecations[0]
	for _, dep := range deprecations[1:] {
		if compareWarningLevels(dep.Level, mostUrgent.Level) > 0 {
			mostUrgent = dep
		}
	}

	return mostUrgent
}
