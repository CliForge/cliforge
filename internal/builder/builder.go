// Package builder converts OpenAPI specs to Cobra command trees.
package builder

import (
	"fmt"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

// Builder builds Cobra commands from OpenAPI specifications.
type Builder struct {
	spec       *openapi.ParsedSpec
	config     *BuilderConfig
	commandMap map[string]*cobra.Command
}

// BuilderConfig configures command building behavior.
type BuilderConfig struct {
	// RootName is the name of the root command
	RootName string
	// RootDescription is the description of the root command
	RootDescription string
	// GroupByTags groups commands by OpenAPI tags
	GroupByTags bool
	// FlattenSingleOperations removes unnecessary nesting
	FlattenSingleOperations bool
	// DefaultExecutor is the default command executor function
	DefaultExecutor func(cmd *cobra.Command, args []string) error
}

// NewBuilder creates a new command builder.
func NewBuilder(spec *openapi.ParsedSpec, config *BuilderConfig) *Builder {
	if config == nil {
		config = DefaultBuilderConfig()
	}
	return &Builder{
		spec:       spec,
		config:     config,
		commandMap: make(map[string]*cobra.Command),
	}
}

// DefaultBuilderConfig returns default builder configuration.
func DefaultBuilderConfig() *BuilderConfig {
	return &BuilderConfig{
		RootName:                "cli",
		RootDescription:         "CLI generated from OpenAPI spec",
		GroupByTags:             true,
		FlattenSingleOperations: true,
	}
}

// Build builds the complete command tree from the OpenAPI spec.
func (b *Builder) Build() (*cobra.Command, error) {
	// Create root command
	rootCmd := b.buildRootCommand()

	// Get all operations from spec
	operations, err := b.spec.GetOperations()
	if err != nil {
		return nil, fmt.Errorf("failed to get operations: %w", err)
	}

	// Build command tree based on configuration
	if b.config.GroupByTags {
		if err := b.buildTagGroupedCommands(rootCmd, operations); err != nil {
			return nil, err
		}
	} else {
		if err := b.buildPathGroupedCommands(rootCmd, operations); err != nil {
			return nil, err
		}
	}

	return rootCmd, nil
}

// buildRootCommand creates the root command from spec info.
func (b *Builder) buildRootCommand() *cobra.Command {
	info := b.spec.GetInfo()

	rootCmd := &cobra.Command{
		Use:   b.config.RootName,
		Short: info.Title,
		Long:  info.Description,
	}

	// Apply CLI config extensions if present
	if b.spec.Extensions.Config != nil {
		if b.spec.Extensions.Config.Name != "" {
			rootCmd.Use = b.spec.Extensions.Config.Name
		}
		if b.spec.Extensions.Config.Description != "" {
			rootCmd.Short = b.spec.Extensions.Config.Description
		}
	}

	// Store root command in map
	b.commandMap[""] = rootCmd

	return rootCmd
}

// buildTagGroupedCommands builds commands grouped by OpenAPI tags.
func (b *Builder) buildTagGroupedCommands(rootCmd *cobra.Command, operations []*openapi.Operation) error {
	// Group operations by tag
	tagGroups := make(map[string][]*openapi.Operation)
	untaggedOps := make([]*openapi.Operation, 0)

	for _, op := range operations {
		if len(op.Tags) == 0 {
			untaggedOps = append(untaggedOps, op)
			continue
		}
		// Use first tag for grouping
		tag := op.Tags[0]
		tagGroups[tag] = append(tagGroups[tag], op)
	}

	// Create group commands for each tag
	for tag, ops := range tagGroups {
		groupCmd := b.buildTagGroupCommand(tag)

		// Add operations to group
		for _, op := range ops {
			opCmd, err := b.buildOperationCommand(op)
			if err != nil {
				return fmt.Errorf("failed to build command for operation %s: %w", op.OperationID, err)
			}
			groupCmd.AddCommand(opCmd)
		}

		rootCmd.AddCommand(groupCmd)
	}

	// Add untagged operations directly to root
	for _, op := range untaggedOps {
		opCmd, err := b.buildOperationCommand(op)
		if err != nil {
			return fmt.Errorf("failed to build command for operation %s: %w", op.OperationID, err)
		}
		rootCmd.AddCommand(opCmd)
	}

	return nil
}

// buildTagGroupCommand creates a group command for a tag.
func (b *Builder) buildTagGroupCommand(tag string) *cobra.Command {
	cmdName := toCommandName(tag)

	cmd := &cobra.Command{
		Use:   cmdName,
		Short: fmt.Sprintf("%s operations", tag),
	}

	b.commandMap[tag] = cmd
	return cmd
}

// buildPathGroupedCommands builds commands grouped by path structure.
func (b *Builder) buildPathGroupedCommands(rootCmd *cobra.Command, operations []*openapi.Operation) error {
	for _, op := range operations {
		// Build command hierarchy from path
		parentCmd := b.getOrCreatePathCommand(rootCmd, op.Path)

		// Create operation command
		opCmd, err := b.buildOperationCommand(op)
		if err != nil {
			return fmt.Errorf("failed to build command for operation %s: %w", op.OperationID, err)
		}

		parentCmd.AddCommand(opCmd)
	}

	return nil
}

// getOrCreatePathCommand gets or creates a command for a path segment.
func (b *Builder) getOrCreatePathCommand(rootCmd *cobra.Command, path string) *cobra.Command {
	// Parse path into segments
	segments := parsePathSegments(path)

	if len(segments) == 0 {
		return rootCmd
	}

	// Navigate/create command hierarchy
	currentCmd := rootCmd
	currentPath := ""

	for i, segment := range segments {
		// Skip parameter segments at group level
		if strings.HasPrefix(segment, "{") && i < len(segments)-1 {
			continue
		}

		currentPath = currentPath + "/" + segment

		// Check if command exists
		if existing, ok := b.commandMap[currentPath]; ok {
			currentCmd = existing
			continue
		}

		// Create new group command
		cmdName := toCommandName(segment)
		groupCmd := &cobra.Command{
			Use:   cmdName,
			Short: fmt.Sprintf("Manage %s", segment),
		}

		currentCmd.AddCommand(groupCmd)
		b.commandMap[currentPath] = groupCmd
		currentCmd = groupCmd
	}

	return currentCmd
}

// buildOperationCommand builds a command for a single operation.
func (b *Builder) buildOperationCommand(op *openapi.Operation) (*cobra.Command, error) {
	// Determine command name
	cmdName := b.determineCommandName(op)

	// Create command
	cmd := &cobra.Command{
		Use:   cmdName,
		Short: op.Summary,
		Long:  op.Description,
	}

	// Apply x-cli-command override
	if op.CLICommand != "" {
		cmd.Use = op.CLICommand
	}

	// Apply x-cli-aliases
	if len(op.CLIAliases) > 0 {
		cmd.Aliases = op.CLIAliases
	}

	// Store operation metadata in command annotations
	if cmd.Annotations == nil {
		cmd.Annotations = make(map[string]string)
	}
	cmd.Annotations["operationID"] = op.OperationID
	cmd.Annotations["method"] = op.Method
	cmd.Annotations["path"] = op.Path

	// Set executor if provided
	if b.config.DefaultExecutor != nil {
		cmd.RunE = b.config.DefaultExecutor
	}

	return cmd, nil
}

// determineCommandName determines the command name from an operation.
func (b *Builder) determineCommandName(op *openapi.Operation) string {
	// Priority: x-cli-command > operationID > method
	if op.CLICommand != "" {
		return op.CLICommand
	}

	if op.OperationID != "" {
		return toCommandName(op.OperationID)
	}

	// Fallback to method name
	return strings.ToLower(op.Method)
}

// parsePathSegments splits a path into command segments.
func parsePathSegments(path string) []string {
	// Remove leading/trailing slashes
	path = strings.Trim(path, "/")

	if path == "" {
		return []string{}
	}

	segments := strings.Split(path, "/")

	// Filter out empty segments
	filtered := make([]string, 0, len(segments))
	for _, seg := range segments {
		if seg != "" {
			filtered = append(filtered, seg)
		}
	}

	return filtered
}

// toCommandName converts a string to a valid command name.
func toCommandName(s string) string {
	// Remove special characters and convert to kebab-case
	s = strings.TrimSpace(s)

	// Replace underscores and spaces with hyphens
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ReplaceAll(s, " ", "-")

	// Convert camelCase to kebab-case
	s = camelToKebab(s)

	// Remove curly braces from path parameters
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")

	// Convert to lowercase
	s = strings.ToLower(s)

	return s
}

// camelToKebab converts camelCase to kebab-case.
func camelToKebab(s string) string {
	var result strings.Builder
	prevWasUpper := false
	for i, r := range s {
		isUpper := r >= 'A' && r <= 'Z'
		// Add hyphen before uppercase letter if:
		// - not at start
		// - previous char was not uppercase (avoid "API" becoming "a-p-i")
		// - current char is uppercase
		if i > 0 && isUpper && !prevWasUpper {
			result.WriteRune('-')
		}
		result.WriteRune(r)
		prevWasUpper = isUpper
	}
	return result.String()
}

// GetCommandByPath retrieves a command by its path.
func (b *Builder) GetCommandByPath(path string) (*cobra.Command, bool) {
	cmd, ok := b.commandMap[path]
	return cmd, ok
}

// GetCommandByOperationID retrieves a command by its operation ID.
func (b *Builder) GetCommandByOperationID(rootCmd *cobra.Command, operationID string) *cobra.Command {
	return findCommandByAnnotation(rootCmd, "operationID", operationID)
}

// findCommandByAnnotation recursively finds a command with a specific annotation.
func findCommandByAnnotation(cmd *cobra.Command, key, value string) *cobra.Command {
	if cmd.Annotations != nil {
		if annotationValue, ok := cmd.Annotations[key]; ok && annotationValue == value {
			return cmd
		}
	}

	for _, subCmd := range cmd.Commands() {
		if found := findCommandByAnnotation(subCmd, key, value); found != nil {
			return found
		}
	}

	return nil
}
