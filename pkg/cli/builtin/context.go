package builtin

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/state"
	"github.com/spf13/cobra"
)

// ContextOptions configures the context command behavior.
type ContextOptions struct {
	ContextManager *state.ContextManager
	Output         io.Writer
}

// NewContextCommand creates a new context command group.
func NewContextCommand(opts *ContextOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Manage contexts",
		Long: `Manage named contexts for different environments or configurations.

Contexts allow you to switch between different sets of configuration values,
similar to kubectl contexts. This is useful for managing multiple environments
(dev, staging, production) or different API configurations.

Available subcommands:
  list         - List all contexts
  current      - Show current context
  use          - Switch to a context
  create       - Create a new context
  delete       - Delete a context
  set          - Set a field in a context
  get          - Get a field from a context
  rename       - Rename a context
  show         - Show details of a context`,
		Aliases: []string{"ctx"},
	}

	// Add subcommands
	cmd.AddCommand(newContextListCommand(opts))
	cmd.AddCommand(newContextCurrentCommand(opts))
	cmd.AddCommand(newContextUseCommand(opts))
	cmd.AddCommand(newContextCreateCommand(opts))
	cmd.AddCommand(newContextDeleteCommand(opts))
	cmd.AddCommand(newContextSetCommand(opts))
	cmd.AddCommand(newContextGetCommand(opts))
	cmd.AddCommand(newContextRenameCommand(opts))
	cmd.AddCommand(newContextShowCommand(opts))

	return cmd
}

// newContextListCommand creates the context list subcommand.
func newContextListCommand(opts *ContextOptions) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List all contexts",
		Aliases: []string{"ls"},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runContextList(opts, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json|yaml)")

	return cmd
}

// newContextCurrentCommand creates the context current subcommand.
func newContextCurrentCommand(opts *ContextOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "current",
		Short: "Show current context",
		RunE: func(cmd *cobra.Command, args []string) error {
			current := opts.ContextManager.CurrentName()
			if current == "" {
				_, _ = fmt.Fprintln(opts.Output, "No current context")
			} else {
				_, _ = fmt.Fprintln(opts.Output, current)
			}
			return nil
		},
	}
}

// newContextUseCommand creates the context use subcommand.
func newContextUseCommand(opts *ContextOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "use <context-name>",
		Short: "Switch to a context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := opts.ContextManager.SwitchTo(name); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(opts.Output, "Switched to context %q\n", name)
			return nil
		},
	}
}

// newContextCreateCommand creates the context create subcommand.
func newContextCreateCommand(opts *ContextOptions) *cobra.Command {
	var description string

	cmd := &cobra.Command{
		Use:   "create <context-name>",
		Short: "Create a new context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if err := opts.ContextManager.Create(name, description, make(map[string]string)); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(opts.Output, "Created context %q\n", name)
			return nil
		},
	}

	cmd.Flags().StringVarP(&description, "description", "d", "", "Context description")

	return cmd
}

// newContextDeleteCommand creates the context delete subcommand.
func newContextDeleteCommand(opts *ContextOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "delete <context-name>",
		Short:   "Delete a context",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// Check if it's the current context
			if opts.ContextManager.CurrentName() == name {
				return fmt.Errorf("cannot delete current context, switch to another context first")
			}

			if err := opts.ContextManager.Delete(name); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(opts.Output, "Deleted context %q\n", name)
			return nil
		},
	}
}

// newContextSetCommand creates the context set subcommand.
func newContextSetCommand(opts *ContextOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "set <context-name> <key> <value>",
		Short: "Set a field in a context",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			key := args[1]
			value := args[2]

			fields := map[string]string{key: value}
			if err := opts.ContextManager.Update(name, fields); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(opts.Output, "Set %s.%s = %s\n", name, key, value)
			return nil
		},
	}
}

// newContextGetCommand creates the context get subcommand.
func newContextGetCommand(opts *ContextOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <context-name> <key>",
		Short: "Get a field from a context",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			key := args[1]

			contexts, err := opts.ContextManager.List()
			if err != nil {
				return err
			}

			ctx, ok := contexts[name]
			if !ok {
				return fmt.Errorf("context %q not found", name)
			}

			value, exists := ctx.Get(key)
			if !exists {
				return fmt.Errorf("key %q not found in context %q", key, name)
			}

			_, _ = fmt.Fprintln(opts.Output, value)
			return nil
		},
	}
}

// newContextRenameCommand creates the context rename subcommand.
func newContextRenameCommand(opts *ContextOptions) *cobra.Command {
	return &cobra.Command{
		Use:     "rename <old-name> <new-name>",
		Short:   "Rename a context",
		Aliases: []string{"mv"},
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			oldName := args[0]
			newName := args[1]

			if err := opts.ContextManager.Rename(oldName, newName); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(opts.Output, "Renamed context %q to %q\n", oldName, newName)
			return nil
		},
	}
}

// newContextShowCommand creates the context show subcommand.
func newContextShowCommand(opts *ContextOptions) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show <context-name>",
		Short: "Show details of a context",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			return runContextShow(opts, name, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "yaml", "Output format (yaml|json)")

	return cmd
}

// runContextList lists all contexts.
func runContextList(opts *ContextOptions, outputFormat string) error {
	contexts, err := opts.ContextManager.List()
	if err != nil {
		return err
	}

	if len(contexts) == 0 {
		_, _ = fmt.Fprintln(opts.Output, "No contexts found")
		return nil
	}

	currentName := opts.ContextManager.CurrentName()

	switch outputFormat {
	case "json":
		return formatContextsJSON(contexts, currentName, opts.Output)
	case "yaml":
		return formatContextsYAML(contexts, currentName, opts.Output)
	default:
		return formatContextsTable(contexts, currentName, opts.Output)
	}
}

// runContextShow shows details of a context.
func runContextShow(opts *ContextOptions, name, outputFormat string) error {
	contexts, err := opts.ContextManager.List()
	if err != nil {
		return err
	}

	ctx, ok := contexts[name]
	if !ok {
		return fmt.Errorf("context %q not found", name)
	}

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(opts.Output)
		encoder.SetIndent("", "  ")
		return encoder.Encode(ctx)
	default: // yaml
		_, _ = fmt.Fprintf(opts.Output, "name: %s\n", ctx.Name)
		if ctx.Description != "" {
			_, _ = fmt.Fprintf(opts.Output, "description: %s\n", ctx.Description)
		}
		_, _ = fmt.Fprintf(opts.Output, "created_at: %s\n", ctx.CreatedAt.Format(time.RFC3339))
		if !ctx.LastUsed.IsZero() {
			_, _ = fmt.Fprintf(opts.Output, "last_used: %s\n", ctx.LastUsed.Format(time.RFC3339))
		}
		if ctx.UseCount > 0 {
			_, _ = fmt.Fprintf(opts.Output, "use_count: %d\n", ctx.UseCount)
		}

		if len(ctx.Fields) > 0 {
			_, _ = fmt.Fprintln(opts.Output, "fields:")
			// Sort keys for consistent output
			keys := make([]string, 0, len(ctx.Fields))
			for k := range ctx.Fields {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				_, _ = fmt.Fprintf(opts.Output, "  %s: %s\n", k, ctx.Fields[k])
			}
		}
	}

	return nil
}

// formatContextsTable formats contexts as a table.
func formatContextsTable(contexts map[string]*state.Context, currentName string, w io.Writer) error {
	// Sort context names
	names := make([]string, 0, len(contexts))
	for name := range contexts {
		names = append(names, name)
	}
	sort.Strings(names)

	// Print header
	_, _ = fmt.Fprintf(w, "%-20s %-10s %-30s %s\n", "NAME", "CURRENT", "DESCRIPTION", "FIELDS")
	_, _ = fmt.Fprintln(w, strings.Repeat("-", 80))

	// Print contexts
	for _, name := range names {
		ctx := contexts[name]
		current := ""
		if name == currentName {
			current = "*"
		}

		desc := ctx.Description
		if len(desc) > 30 {
			desc = desc[:27] + "..."
		}

		fieldCount := fmt.Sprintf("%d fields", len(ctx.Fields))

		_, _ = fmt.Fprintf(w, "%-20s %-10s %-30s %s\n", name, current, desc, fieldCount)
	}

	return nil
}

// formatContextsJSON formats contexts as JSON.
func formatContextsJSON(contexts map[string]*state.Context, currentName string, w io.Writer) error {
	output := map[string]interface{}{
		"current":  currentName,
		"contexts": contexts,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatContextsYAML formats contexts as YAML.
func formatContextsYAML(contexts map[string]*state.Context, currentName string, w io.Writer) error {
	_, _ = fmt.Fprintf(w, "current: %s\n", currentName)
	_, _ = fmt.Fprintln(w, "contexts:")

	// Sort context names
	names := make([]string, 0, len(contexts))
	for name := range contexts {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		ctx := contexts[name]
		_, _ = fmt.Fprintf(w, "  %s:\n", name)
		if ctx.Description != "" {
			_, _ = fmt.Fprintf(w, "    description: %s\n", ctx.Description)
		}
		_, _ = fmt.Fprintf(w, "    created_at: %s\n", ctx.CreatedAt.Format(time.RFC3339))
		if !ctx.LastUsed.IsZero() {
			_, _ = fmt.Fprintf(w, "    last_used: %s\n", ctx.LastUsed.Format(time.RFC3339))
		}
		if len(ctx.Fields) > 0 {
			_, _ = fmt.Fprintln(w, "    fields:")
			// Sort keys
			keys := make([]string, 0, len(ctx.Fields))
			for k := range ctx.Fields {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				_, _ = fmt.Fprintf(w, "      %s: %s\n", k, ctx.Fields[k])
			}
		}
	}

	return nil
}
