// Package builder provides nested resource handling for OpenAPI-based CLI commands.
//
// This file implements support for the x-cli-parent-resource extension, which enables
// hierarchical resource relationships in CLI commands. When an operation specifies a
// parent resource, the builder automatically adds required flags and handles path
// parameter substitution.
//
// # Example
//
// Given an OpenAPI operation:
//
//	paths:
//	  /api/v1/clusters/{cluster_id}/machine_pools:
//	    post:
//	      operationId: createMachinePool
//	      x-cli-parent-resource: "cluster"
//
// The builder generates:
//
//	mycli create machinepool --cluster my-cluster --name pool1
//
// The --cluster flag is automatically added, marked as required, and used to
// substitute {cluster_id} in the path before making the API request.
package builder

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

// NestedResourceHandler manages nested resource relationships and path parameters.
type NestedResourceHandler struct {
	pathParamPattern *regexp.Regexp
}

// NewNestedResourceHandler creates a new nested resource handler.
func NewNestedResourceHandler() *NestedResourceHandler {
	return &NestedResourceHandler{
		pathParamPattern: regexp.MustCompile(`\{([^}]+)\}`),
	}
}

// AddParentResourceFlag adds a required flag for the parent resource ID.
// For example, if an operation has path "/api/v1/clusters/{cluster_id}/machine_pools"
// and x-cli-parent-resource: "cluster", it adds a --cluster flag.
func (h *NestedResourceHandler) AddParentResourceFlag(cmd *cobra.Command, op *openapi.Operation) error {
	if op.CLIParentRes == "" {
		// No parent resource defined, nothing to do
		return nil
	}

	// Extract path parameters from the operation path
	pathParams := h.extractPathParameters(op.Path)
	if len(pathParams) == 0 {
		// No path parameters, parent resource cannot be resolved
		return fmt.Errorf("operation has x-cli-parent-resource '%s' but no path parameters", op.CLIParentRes)
	}

	// Find the parent resource path parameter
	parentParam := h.findParentParameter(op.CLIParentRes, pathParams)
	if parentParam == "" {
		return fmt.Errorf("parent resource '%s' not found in path parameters: %v", op.CLIParentRes, pathParams)
	}

	// Determine flag name - use parent resource name (e.g., "cluster" -> "--cluster")
	flagName := toFlagName(op.CLIParentRes)

	// Add the flag as a required string flag
	description := fmt.Sprintf("ID or name of the %s", op.CLIParentRes)
	cmd.Flags().String(flagName, "", description)

	// Mark as required
	if err := cmd.MarkFlagRequired(flagName); err != nil {
		return fmt.Errorf("failed to mark parent resource flag as required: %w", err)
	}

	// Store parent resource mapping in annotations for later path substitution
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations["parent-resource"] = op.CLIParentRes
	cmd.Annotations["parent-param"] = parentParam
	cmd.Annotations[fmt.Sprintf("parent:%s", flagName)] = parentParam

	return nil
}

// extractPathParameters extracts all path parameter names from a path template.
// For example, "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}"
// returns ["cluster_id", "machinepool_id"]
func (h *NestedResourceHandler) extractPathParameters(path string) []string {
	matches := h.pathParamPattern.FindAllStringSubmatch(path, -1)
	params := make([]string, 0, len(matches))

	for _, match := range matches {
		if len(match) > 1 {
			params = append(params, match[1])
		}
	}

	return params
}

// findParentParameter finds the path parameter that matches the parent resource name.
// It tries several matching strategies:
// 1. Exact match: "cluster" matches "{cluster}"
// 2. With _id suffix: "cluster" matches "{cluster_id}"
// 3. CamelCase variations: "cluster" matches "{clusterId}"
func (h *NestedResourceHandler) findParentParameter(parentResource string, pathParams []string) string {
	// Normalize parent resource name
	normalized := strings.ToLower(strings.ReplaceAll(parentResource, "-", "_"))

	for _, param := range pathParams {
		paramNorm := strings.ToLower(param)

		// Strategy 1: Exact match
		if paramNorm == normalized {
			return param
		}

		// Strategy 2: Match with _id suffix
		if paramNorm == normalized+"_id" {
			return param
		}

		// Strategy 3: Match with Id suffix (camelCase)
		if paramNorm == normalized+"id" {
			return param
		}

		// Strategy 4: Check if param starts with normalized name
		if strings.HasPrefix(paramNorm, normalized) {
			return param
		}
	}

	return ""
}

// SubstituteParentResource substitutes the parent resource ID into the path template.
// It reads the parent resource flag value and replaces the corresponding path parameter.
func (h *NestedResourceHandler) SubstituteParentResource(cmd *cobra.Command, path string) (string, error) {
	if cmd.Annotations == nil {
		// No parent resource annotations, return path as-is
		return path, nil
	}

	parentResource, hasParent := cmd.Annotations["parent-resource"]
	if !hasParent {
		// No parent resource, return path as-is
		return path, nil
	}

	parentParam, hasParam := cmd.Annotations["parent-param"]
	if !hasParam {
		return "", fmt.Errorf("parent resource '%s' defined but no parent-param annotation", parentResource)
	}

	// Get the flag value
	flagName := toFlagName(parentResource)
	parentID, err := cmd.Flags().GetString(flagName)
	if err != nil {
		return "", fmt.Errorf("failed to get parent resource flag '%s': %w", flagName, err)
	}

	if parentID == "" {
		return "", fmt.Errorf("parent resource flag '%s' is required", flagName)
	}

	// Substitute the path parameter with the actual value
	substituted := strings.ReplaceAll(path, fmt.Sprintf("{%s}", parentParam), parentID)

	return substituted, nil
}

// GetNestedResourceLevel determines the nesting level of a resource.
// For example:
//   - "/api/v1/clusters" -> 0 (top-level resource)
//   - "/api/v1/clusters/{cluster_id}/machine_pools" -> 1 (nested under cluster)
//   - "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}/nodes" -> 2 (nested under machine pool)
func (h *NestedResourceHandler) GetNestedResourceLevel(path string) int {
	pathParams := h.extractPathParameters(path)
	if len(pathParams) == 0 {
		return 0
	}

	// Count unique resource types (not individual IDs)
	// This is a heuristic: count path parameters before the last segment
	segments := parsePathSegments(path)
	level := 0

	for i, segment := range segments {
		// If this segment is a path parameter and not the last segment
		if strings.HasPrefix(segment, "{") && i < len(segments)-1 {
			level++
		}
	}

	return level
}

// ValidateNestedResourceFlags validates that parent resource flags are properly set.
func (h *NestedResourceHandler) ValidateNestedResourceFlags(cmd *cobra.Command) error {
	if cmd.Annotations == nil {
		return nil
	}

	parentResource, hasParent := cmd.Annotations["parent-resource"]
	if !hasParent {
		return nil
	}

	flagName := toFlagName(parentResource)
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		return fmt.Errorf("parent resource flag '%s' not found", flagName)
	}

	if !flag.Changed {
		return fmt.Errorf("required flag --%s not set", flagName)
	}

	value := flag.Value.String()
	if value == "" {
		return fmt.Errorf("parent resource flag --%s cannot be empty", flagName)
	}

	return nil
}
