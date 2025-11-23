package builtin

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/state"
	"github.com/spf13/cobra"
)

// HistoryOptions configures the history command behavior.
type HistoryOptions struct {
	History *state.History
	Output  io.Writer
}

// NewHistoryCommand creates a new history command.
func NewHistoryCommand(opts *HistoryOptions) *cobra.Command {
	var limit int
	var search string
	var successOnly bool
	var failedOnly bool
	var outputFormat string
	var contextFilter string

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show command history",
		Long: `Display command execution history.

The history command shows previously executed commands along with
their status, execution time, and other metadata.

Examples:
  history                    # Show recent history
  history --limit 50         # Show last 50 commands
  history --search "users"   # Search for commands containing "users"
  history --success-only     # Show only successful commands
  history --failed-only      # Show only failed commands`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistory(opts, limit, search, successOnly, failedOnly, contextFilter, outputFormat)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "Number of entries to show")
	cmd.Flags().StringVar(&search, "search", "", "Search pattern")
	cmd.Flags().BoolVar(&successOnly, "success-only", false, "Show only successful commands")
	cmd.Flags().BoolVar(&failedOnly, "failed-only", false, "Show only failed commands")
	cmd.Flags().StringVar(&contextFilter, "context", "", "Filter by context")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table|json|yaml)")

	// Add subcommands
	cmd.AddCommand(newHistoryClearCommand(opts))
	cmd.AddCommand(newHistoryStatsCommand(opts))

	return cmd
}

// newHistoryClearCommand creates the history clear subcommand.
func newHistoryClearCommand(opts *HistoryOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear command history",
		Long:  "Remove all command history entries.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.History.Clear(); err != nil {
				return err
			}
			if err := opts.History.Save(); err != nil {
				return err
			}
			fmt.Fprintln(opts.Output, "✓ History cleared")
			return nil
		},
	}
}

// newHistoryStatsCommand creates the history stats subcommand.
func newHistoryStatsCommand(opts *HistoryOptions) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show history statistics",
		Long:  "Display statistics about command history.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHistoryStats(opts, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")

	return cmd
}

// runHistory displays command history.
func runHistory(opts *HistoryOptions, limit int, search string, successOnly, failedOnly bool, contextFilter, outputFormat string) error {
	var entries []*state.HistoryEntry

	// Apply filters
	if search != "" {
		entries = opts.History.Search(search)
	} else if successOnly {
		entries = opts.History.GetSuccessful()
	} else if failedOnly {
		entries = opts.History.GetFailed()
	} else if contextFilter != "" {
		entries = opts.History.GetByContext(contextFilter)
	} else {
		entries = opts.History.GetRecent(limit)
	}

	if len(entries) == 0 {
		fmt.Fprintln(opts.Output, "No history entries found")
		return nil
	}

	// Apply limit if we got all entries
	if limit > 0 && len(entries) > limit {
		entries = entries[len(entries)-limit:]
	}

	switch outputFormat {
	case "json":
		return formatHistoryJSON(entries, opts.Output)
	case "yaml":
		return formatHistoryYAML(entries, opts.Output)
	default:
		return formatHistoryTable(entries, opts.Output)
	}
}

// runHistoryStats displays history statistics.
func runHistoryStats(opts *HistoryOptions, outputFormat string) error {
	stats := opts.History.GetStats()

	switch outputFormat {
	case "json":
		encoder := json.NewEncoder(opts.Output)
		encoder.SetIndent("", "  ")
		return encoder.Encode(stats)
	case "yaml":
		return formatStatsYAML(&stats, opts.Output)
	default:
		return formatStatsText(&stats, opts.Output)
	}
}

// formatHistoryTable formats history as a table.
func formatHistoryTable(entries []*state.HistoryEntry, w io.Writer) error {
	// Print header
	fmt.Fprintf(w, "%-5s %-40s %-8s %-10s %s\n", "ID", "COMMAND", "STATUS", "DURATION", "TIMESTAMP")
	fmt.Fprintln(w, strings.Repeat("-", 100))

	// Print entries
	for _, entry := range entries {
		status := "✓"
		if !entry.Success {
			status = "✗"
		}

		command := entry.Command
		if len(command) > 40 {
			command = command[:37] + "..."
		}

		duration := formatMilliseconds(entry.DurationMS)
		timestamp := entry.Timestamp.Format("15:04:05")

		fmt.Fprintf(w, "%-5d %-40s %-8s %-10s %s\n",
			entry.ID, command, status, duration, timestamp)
	}

	return nil
}

// formatHistoryJSON formats history as JSON.
func formatHistoryJSON(entries []*state.HistoryEntry, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(entries)
}

// formatHistoryYAML formats history as YAML.
func formatHistoryYAML(entries []*state.HistoryEntry, w io.Writer) error {
	for i, entry := range entries {
		if i > 0 {
			fmt.Fprintln(w, "---")
		}
		fmt.Fprintf(w, "id: %d\n", entry.ID)
		fmt.Fprintf(w, "command: %s\n", entry.Command)
		fmt.Fprintf(w, "timestamp: %s\n", entry.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(w, "exit_code: %d\n", entry.ExitCode)
		fmt.Fprintf(w, "success: %t\n", entry.Success)
		if entry.DurationMS > 0 {
			fmt.Fprintf(w, "duration_ms: %d\n", entry.DurationMS)
		}
		if entry.Context != "" {
			fmt.Fprintf(w, "context: %s\n", entry.Context)
		}
		if entry.User != "" {
			fmt.Fprintf(w, "user: %s\n", entry.User)
		}
	}
	return nil
}

// formatStatsText formats stats as human-readable text.
func formatStatsText(stats *state.HistoryStats, w io.Writer) error {
	fmt.Fprintln(w, "History Statistics:")
	fmt.Fprintln(w)
	fmt.Fprintf(w, "  Total commands: %d\n", stats.TotalCommands)
	fmt.Fprintf(w, "  Successful: %d\n", stats.SuccessfulCommands)
	fmt.Fprintf(w, "  Failed: %d\n", stats.FailedCommands)

	if stats.TotalCommands > 0 {
		successRate := float64(stats.SuccessfulCommands) / float64(stats.TotalCommands) * 100
		fmt.Fprintf(w, "  Success rate: %.1f%%\n", successRate)
	}

	if stats.AverageDurationMS > 0 {
		fmt.Fprintf(w, "  Average duration: %s\n", formatMilliseconds(stats.AverageDurationMS))
	}

	fmt.Fprintln(w)

	if !stats.FirstCommand.IsZero() {
		fmt.Fprintf(w, "  First command: %s\n", stats.FirstCommand.Format(time.RFC3339))
	}
	if !stats.LastCommand.IsZero() {
		fmt.Fprintf(w, "  Last command: %s\n", stats.LastCommand.Format(time.RFC3339))
	}

	return nil
}

// formatStatsYAML formats stats as YAML.
func formatStatsYAML(stats *state.HistoryStats, w io.Writer) error {
	fmt.Fprintf(w, "total_commands: %d\n", stats.TotalCommands)
	fmt.Fprintf(w, "successful_commands: %d\n", stats.SuccessfulCommands)
	fmt.Fprintf(w, "failed_commands: %d\n", stats.FailedCommands)
	fmt.Fprintf(w, "average_duration_ms: %d\n", stats.AverageDurationMS)
	if !stats.FirstCommand.IsZero() {
		fmt.Fprintf(w, "first_command: %s\n", stats.FirstCommand.Format(time.RFC3339))
	}
	if !stats.LastCommand.IsZero() {
		fmt.Fprintf(w, "last_command: %s\n", stats.LastCommand.Format(time.RFC3339))
	}
	return nil
}

// formatMilliseconds formats milliseconds in a human-readable way.
func formatMilliseconds(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	} else if ms < 60000 {
		return fmt.Sprintf("%.1fs", float64(ms)/1000)
	}
	return fmt.Sprintf("%.1fm", float64(ms)/60000)
}
