package openapi

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ChangeDetector detects changes between OpenAPI spec versions.
type ChangeDetector struct {
	// IncludeNonBreaking includes non-breaking changes in results
	IncludeNonBreaking bool
}

// NewChangeDetector creates a new ChangeDetector instance.
func NewChangeDetector() *ChangeDetector {
	return &ChangeDetector{
		IncludeNonBreaking: true,
	}
}

// DetectedChange represents a detected change between spec versions.
type DetectedChange struct {
	Type        ChangeType
	Severity    ChangeSeverity
	Path        string
	Description string
	OldValue    interface{}
	NewValue    interface{}
}

// ChangeType represents the type of change.
type ChangeType string

const (
	// ChangeTypeAdded represents a newly added element.
	ChangeTypeAdded ChangeType = "added"
	// ChangeTypeRemoved represents a removed element.
	ChangeTypeRemoved ChangeType = "removed"
	// ChangeTypeModified represents a modified element.
	ChangeTypeModified ChangeType = "modified"
)

// ChangeSeverity represents the severity of a change.
type ChangeSeverity string

const (
	// ChangeSeverityBreaking represents a breaking change.
	ChangeSeverityBreaking ChangeSeverity = "breaking"
	// ChangeSeverityDangerous represents a potentially dangerous change.
	ChangeSeverityDangerous ChangeSeverity = "dangerous"
	// ChangeSeveritySafe represents a safe change.
	ChangeSeveritySafe ChangeSeverity = "safe"
)

// DetectChanges compares two specs and returns detected changes.
func (cd *ChangeDetector) DetectChanges(oldSpec, newSpec *ParsedSpec) ([]*DetectedChange, error) {
	var changes []*DetectedChange

	// Compare versions
	oldInfo := oldSpec.GetInfo()
	newInfo := newSpec.GetInfo()

	if oldInfo.Version != newInfo.Version {
		changes = append(changes, &DetectedChange{
			Type:        ChangeTypeModified,
			Severity:    ChangeSeveritySafe,
			Path:        "info.version",
			Description: "API version changed",
			OldValue:    oldInfo.Version,
			NewValue:    newInfo.Version,
		})
	}

	// Compare paths
	pathChanges := cd.comparePaths(oldSpec.Spec.Paths, newSpec.Spec.Paths)
	changes = append(changes, pathChanges...)

	// Compare operations
	operationChanges := cd.compareOperations(oldSpec.Spec, newSpec.Spec)
	changes = append(changes, operationChanges...)

	// Compare schemas
	schemaChanges := cd.compareSchemas(oldSpec.Spec, newSpec.Spec)
	changes = append(changes, schemaChanges...)

	// Compare security requirements
	securityChanges := cd.compareSecurity(oldSpec.Spec, newSpec.Spec)
	changes = append(changes, securityChanges...)

	// Filter out non-breaking changes if requested
	if !cd.IncludeNonBreaking {
		filtered := make([]*DetectedChange, 0)
		for _, change := range changes {
			if change.Severity != ChangeSeveritySafe {
				filtered = append(filtered, change)
			}
		}
		changes = filtered
	}

	return changes, nil
}

// comparePaths compares paths between two specs.
func (cd *ChangeDetector) comparePaths(oldPaths, newPaths *openapi3.Paths) []*DetectedChange {
	var changes []*DetectedChange

	if oldPaths == nil && newPaths == nil {
		return changes
	}

	oldPathMap := make(map[string]*openapi3.PathItem)
	newPathMap := make(map[string]*openapi3.PathItem)

	if oldPaths != nil {
		oldPathMap = oldPaths.Map()
	}
	if newPaths != nil {
		newPathMap = newPaths.Map()
	}

	// Find removed paths
	for path := range oldPathMap {
		if _, exists := newPathMap[path]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeRemoved,
				Severity:    ChangeSeverityBreaking,
				Path:        path,
				Description: fmt.Sprintf("Path removed: %s", path),
			})
		}
	}

	// Find added paths
	for path := range newPathMap {
		if _, exists := oldPathMap[path]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeAdded,
				Severity:    ChangeSeveritySafe,
				Path:        path,
				Description: fmt.Sprintf("Path added: %s", path),
			})
		}
	}

	return changes
}

// compareOperations compares operations between two specs.
func (cd *ChangeDetector) compareOperations(oldSpec, newSpec *openapi3.T) []*DetectedChange {
	var changes []*DetectedChange

	oldOps := cd.getOperationMap(oldSpec)
	newOps := cd.getOperationMap(newSpec)

	// Find removed operations
	for key, oldOp := range oldOps {
		if _, exists := newOps[key]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeRemoved,
				Severity:    ChangeSeverityBreaking,
				Path:        key,
				Description: fmt.Sprintf("Operation removed: %s", oldOp.Summary),
			})
		}
	}

	// Find added operations
	for key, newOp := range newOps {
		if _, exists := oldOps[key]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeAdded,
				Severity:    ChangeSeveritySafe,
				Path:        key,
				Description: fmt.Sprintf("Operation added: %s", newOp.Summary),
			})
		}
	}

	// Find modified operations
	for key, oldOp := range oldOps {
		if newOp, exists := newOps[key]; exists {
			opChanges := cd.compareOperation(key, oldOp, newOp)
			changes = append(changes, opChanges...)
		}
	}

	return changes
}

// getOperationMap creates a map of operations keyed by "METHOD /path".
func (cd *ChangeDetector) getOperationMap(spec *openapi3.T) map[string]*openapi3.Operation {
	ops := make(map[string]*openapi3.Operation)

	if spec.Paths == nil {
		return ops
	}

	for path, pathItem := range spec.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			if operation != nil {
				key := fmt.Sprintf("%s %s", strings.ToUpper(method), path)
				ops[key] = operation
			}
		}
	}

	return ops
}

// compareOperation compares two operations.
func (cd *ChangeDetector) compareOperation(key string, oldOp, newOp *openapi3.Operation) []*DetectedChange {
	var changes []*DetectedChange

	// Compare parameters
	oldParams := cd.getParameterMap(oldOp.Parameters)
	newParams := cd.getParameterMap(newOp.Parameters)

	// Removed parameters
	for name := range oldParams {
		if _, exists := newParams[name]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeRemoved,
				Severity:    ChangeSeverityBreaking,
				Path:        fmt.Sprintf("%s.parameters.%s", key, name),
				Description: fmt.Sprintf("Parameter removed: %s", name),
			})
		}
	}

	// Added parameters
	for name, param := range newParams {
		if _, exists := oldParams[name]; !exists {
			severity := ChangeSeveritySafe
			if param.Value != nil && param.Value.Required {
				severity = ChangeSeverityBreaking
			}
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeAdded,
				Severity:    severity,
				Path:        fmt.Sprintf("%s.parameters.%s", key, name),
				Description: fmt.Sprintf("Parameter added: %s (required: %v)", name, param.Value != nil && param.Value.Required),
			})
		}
	}

	// Compare request body
	if oldOp.RequestBody != nil && newOp.RequestBody == nil {
		changes = append(changes, &DetectedChange{
			Type:        ChangeTypeRemoved,
			Severity:    ChangeSeverityBreaking,
			Path:        fmt.Sprintf("%s.requestBody", key),
			Description: "Request body removed",
		})
	} else if oldOp.RequestBody == nil && newOp.RequestBody != nil {
		severity := ChangeSeveritySafe
		if newOp.RequestBody.Value != nil && newOp.RequestBody.Value.Required {
			severity = ChangeSeverityDangerous
		}
		changes = append(changes, &DetectedChange{
			Type:        ChangeTypeAdded,
			Severity:    severity,
			Path:        fmt.Sprintf("%s.requestBody", key),
			Description: fmt.Sprintf("Request body added (required: %v)", newOp.RequestBody.Value != nil && newOp.RequestBody.Value.Required),
		})
	}

	// Compare responses
	oldResponses := cd.getResponseMap(oldOp.Responses)
	newResponses := cd.getResponseMap(newOp.Responses)

	for status := range oldResponses {
		if _, exists := newResponses[status]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeRemoved,
				Severity:    ChangeSeverityDangerous,
				Path:        fmt.Sprintf("%s.responses.%s", key, status),
				Description: fmt.Sprintf("Response %s removed", status),
			})
		}
	}

	for status := range newResponses {
		if _, exists := oldResponses[status]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeAdded,
				Severity:    ChangeSeveritySafe,
				Path:        fmt.Sprintf("%s.responses.%s", key, status),
				Description: fmt.Sprintf("Response %s added", status),
			})
		}
	}

	// Check for deprecation
	if newOp.Deprecated && !oldOp.Deprecated {
		changes = append(changes, &DetectedChange{
			Type:        ChangeTypeModified,
			Severity:    ChangeSeverityDangerous,
			Path:        key,
			Description: "Operation marked as deprecated",
		})
	}

	return changes
}

// getParameterMap creates a map of parameters by name.
func (cd *ChangeDetector) getParameterMap(params openapi3.Parameters) map[string]*openapi3.ParameterRef {
	paramMap := make(map[string]*openapi3.ParameterRef)
	for _, paramRef := range params {
		if paramRef.Value != nil {
			paramMap[paramRef.Value.Name] = paramRef
		}
	}
	return paramMap
}

// getResponseMap creates a map of responses by status code.
func (cd *ChangeDetector) getResponseMap(responses *openapi3.Responses) map[string]*openapi3.ResponseRef {
	respMap := make(map[string]*openapi3.ResponseRef)
	if responses == nil {
		return respMap
	}
	for status, respRef := range responses.Map() {
		respMap[status] = respRef
	}
	return respMap
}

// compareSchemas compares schemas between two specs.
func (cd *ChangeDetector) compareSchemas(oldSpec, newSpec *openapi3.T) []*DetectedChange {
	var changes []*DetectedChange

	if oldSpec.Components == nil || newSpec.Components == nil {
		return changes
	}

	oldSchemas := oldSpec.Components.Schemas
	newSchemas := newSpec.Components.Schemas

	if oldSchemas == nil && newSchemas == nil {
		return changes
	}

	// Find removed schemas
	for name := range oldSchemas {
		if _, exists := newSchemas[name]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeRemoved,
				Severity:    ChangeSeverityBreaking,
				Path:        fmt.Sprintf("components.schemas.%s", name),
				Description: fmt.Sprintf("Schema removed: %s", name),
			})
		}
	}

	// Find added schemas
	for name := range newSchemas {
		if _, exists := oldSchemas[name]; !exists {
			changes = append(changes, &DetectedChange{
				Type:        ChangeTypeAdded,
				Severity:    ChangeSeveritySafe,
				Path:        fmt.Sprintf("components.schemas.%s", name),
				Description: fmt.Sprintf("Schema added: %s", name),
			})
		}
	}

	return changes
}

// compareSecurity compares security requirements.
func (cd *ChangeDetector) compareSecurity(oldSpec, newSpec *openapi3.T) []*DetectedChange {
	var changes []*DetectedChange

	oldSecLen := len(oldSpec.Security)
	newSecLen := len(newSpec.Security)

	if oldSecLen == 0 && newSecLen > 0 {
		changes = append(changes, &DetectedChange{
			Type:        ChangeTypeAdded,
			Severity:    ChangeSeverityBreaking,
			Path:        "security",
			Description: "Security requirements added",
		})
	} else if oldSecLen > 0 && newSecLen == 0 {
		changes = append(changes, &DetectedChange{
			Type:        ChangeTypeRemoved,
			Severity:    ChangeSeverityDangerous,
			Path:        "security",
			Description: "Security requirements removed",
		})
	}

	return changes
}

// IsBreaking returns true if any changes are breaking.
// It iterates through all detected changes and returns true if any have breaking severity.
func IsBreaking(changes []*DetectedChange) bool {
	for _, change := range changes {
		if change.Severity == ChangeSeverityBreaking {
			return true
		}
	}
	return false
}

// GroupByType groups changes by their type (added, removed, modified).
// It returns a map where keys are change types and values are slices of changes.
func GroupByType(changes []*DetectedChange) map[ChangeType][]*DetectedChange {
	groups := make(map[ChangeType][]*DetectedChange)
	for _, change := range changes {
		groups[change.Type] = append(groups[change.Type], change)
	}
	return groups
}

// GroupBySeverity groups changes by their severity (breaking, dangerous, safe).
// It returns a map where keys are severity levels and values are slices of changes.
func GroupBySeverity(changes []*DetectedChange) map[ChangeSeverity][]*DetectedChange {
	groups := make(map[ChangeSeverity][]*DetectedChange)
	for _, change := range changes {
		groups[change.Severity] = append(groups[change.Severity], change)
	}
	return groups
}

// FormatChangelog formats detected changes into a human-readable changelog string.
// It groups changes by severity and formats them as breaking changes, warnings, and new features.
func FormatChangelog(changes []*DetectedChange) string {
	if len(changes) == 0 {
		return "No changes detected"
	}

	var sb strings.Builder

	// Group by severity
	groups := GroupBySeverity(changes)

	// Breaking changes first
	if breaking, ok := groups[ChangeSeverityBreaking]; ok && len(breaking) > 0 {
		sb.WriteString("BREAKING CHANGES:\n")
		for _, change := range breaking {
			sb.WriteString(fmt.Sprintf("  - [%s] %s\n", change.Type, change.Description))
		}
		sb.WriteString("\n")
	}

	// Dangerous changes
	if dangerous, ok := groups[ChangeSeverityDangerous]; ok && len(dangerous) > 0 {
		sb.WriteString("DEPRECATIONS & WARNINGS:\n")
		for _, change := range dangerous {
			sb.WriteString(fmt.Sprintf("  - [%s] %s\n", change.Type, change.Description))
		}
		sb.WriteString("\n")
	}

	// Safe changes
	if safe, ok := groups[ChangeSeveritySafe]; ok && len(safe) > 0 {
		sb.WriteString("NEW FEATURES & IMPROVEMENTS:\n")
		for _, change := range safe {
			if change.Type == ChangeTypeAdded {
				sb.WriteString(fmt.Sprintf("  - %s\n", change.Description))
			}
		}
	}

	return sb.String()
}

// GetChangelog extracts changelog entries from a spec and returns them sorted by version.
// It returns changelog entries in descending order (newest first).
func GetChangelog(spec *ParsedSpec) []ChangelogEntry {
	if spec.Extensions == nil {
		return nil
	}

	// Sort by version (descending)
	entries := make([]ChangelogEntry, len(spec.Extensions.Changelog))
	copy(entries, spec.Extensions.Changelog)

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Version > entries[j].Version
	})

	return entries
}

// GetLatestChanges returns changes since a specific version.
// It filters changelog entries to include only versions newer than sinceVersion.
func GetLatestChanges(spec *ParsedSpec, sinceVersion string) []ChangelogEntry {
	allEntries := GetChangelog(spec)
	var result []ChangelogEntry

	for _, entry := range allEntries {
		if entry.Version > sinceVersion {
			result = append(result, entry)
		} else {
			break
		}
	}

	return result
}
