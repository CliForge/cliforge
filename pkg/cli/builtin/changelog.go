package builtin

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

// ChangelogOptions configures the changelog command behavior.
type ChangelogOptions struct {
	Config              *cli.CLIConfig
	ShowBinaryChanges   bool
	ShowAPIChanges      bool
	BinaryChangelogFunc func() ([]ChangelogEntry, error)
	APIChangelogFunc    func() ([]openapi.ChangelogEntry, error)
	Output              io.Writer
}

// ChangelogEntry represents a single changelog entry.
type ChangelogEntry struct {
	Version     string    `json:"version"`
	Date        time.Time `json:"date"`
	Changes     []string  `json:"changes"`
	Breaking    []string  `json:"breaking,omitempty"`
	Deprecated  []string  `json:"deprecated,omitempty"`
	IsCurrent   bool      `json:"is_current,omitempty"`
}

// NewChangelogCommand creates a new changelog command.
func NewChangelogCommand(opts *ChangelogOptions) *cobra.Command {
	var binaryOnly bool
	var apiOnly bool
	var sinceVersion string
	var limit int
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "changelog",
		Short: "Show changelog",
		Long: `Display changelog for both CLI binary and API.

The changelog shows:
- CLI binary changes (new features, bug fixes, improvements)
- API changes (new endpoints, schema changes, deprecations)

Changes are separated because they have different release cycles:
- Binary changes: Monthly/quarterly (security, CLI features)
- API changes: Weekly/daily (new endpoints, parameters)

Examples:
  changelog                     # Show all changes
  changelog --binary-only       # Show only CLI changes
  changelog --api-only          # Show only API changes
  changelog --since v2.0.0      # Show changes since version
  changelog --limit 5           # Show last 5 versions`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runChangelog(opts, binaryOnly, apiOnly, sinceVersion, limit, outputFormat)
		},
	}

	cmd.Flags().BoolVar(&binaryOnly, "binary-only", false, "Show only CLI binary changes")
	cmd.Flags().BoolVar(&apiOnly, "api-only", false, "Show only API changes")
	cmd.Flags().StringVar(&sinceVersion, "since", "", "Show changes since this version")
	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Limit number of versions shown")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")

	return cmd
}

// runChangelog executes the changelog command.
func runChangelog(opts *ChangelogOptions, binaryOnly, apiOnly bool, sinceVersion string, limit int, outputFormat string) error {
	var binaryChanges []ChangelogEntry
	var apiChanges []openapi.ChangelogEntry
	var err error

	// Determine what to show
	showBinary := opts.ShowBinaryChanges && !apiOnly
	showAPI := opts.ShowAPIChanges && !binaryOnly

	// Get binary changelog
	if showBinary {
		binaryChanges, err = opts.BinaryChangelogFunc()
		if err != nil {
			fmt.Fprintf(opts.Output, "Warning: failed to load binary changelog: %v\n", err)
			binaryChanges = []ChangelogEntry{}
		}
	}

	// Get API changelog
	if showAPI {
		apiChanges, err = opts.APIChangelogFunc()
		if err != nil {
			fmt.Fprintf(opts.Output, "Warning: failed to load API changelog: %v\n", err)
			apiChanges = []openapi.ChangelogEntry{}
		}
	}

	// Filter by version if specified
	if sinceVersion != "" {
		binaryChanges = filterChangelogSince(binaryChanges, sinceVersion)
		apiChanges = filterAPIChangelogSince(apiChanges, sinceVersion)
	}

	// Apply limit
	if limit > 0 {
		if len(binaryChanges) > limit {
			binaryChanges = binaryChanges[:limit]
		}
		if len(apiChanges) > limit {
			apiChanges = apiChanges[:limit]
		}
	}

	// Format output
	switch outputFormat {
	case "json":
		return formatChangelogJSON(binaryChanges, apiChanges, opts.Output)
	case "yaml":
		return formatChangelogYAML(binaryChanges, apiChanges, opts.Output)
	default:
		return formatChangelogText(binaryChanges, apiChanges, showBinary, showAPI, opts.Output)
	}
}

// formatChangelogText formats changelog as human-readable text.
func formatChangelogText(binaryChanges []ChangelogEntry, apiChanges []openapi.ChangelogEntry, showBinary, showAPI bool, w io.Writer) error {
	if showBinary && len(binaryChanges) > 0 {
		fmt.Fprintln(w, "CLI Binary Changelog:")
		fmt.Fprintln(w, strings.Repeat("━", 80))

		for _, entry := range binaryChanges {
			current := ""
			if entry.IsCurrent {
				current = " - Current"
			}
			fmt.Fprintf(w, "%s (%s)%s\n", entry.Version, entry.Date.Format("2006-01-02"), current)

			if len(entry.Breaking) > 0 {
				fmt.Fprintln(w, "  Breaking Changes:")
				for _, change := range entry.Breaking {
					fmt.Fprintf(w, "    ⚠️  %s\n", change)
				}
			}

			if len(entry.Changes) > 0 {
				for _, change := range entry.Changes {
					fmt.Fprintf(w, "  • %s\n", change)
				}
			}

			if len(entry.Deprecated) > 0 {
				fmt.Fprintln(w, "  Deprecated:")
				for _, dep := range entry.Deprecated {
					fmt.Fprintf(w, "    🔔 %s\n", dep)
				}
			}

			fmt.Fprintln(w)
		}
	}

	if showAPI && len(apiChanges) > 0 {
		if showBinary {
			fmt.Fprintln(w, "\n")
		}

		fmt.Fprintln(w, "API Changelog:")
		fmt.Fprintln(w, strings.Repeat("━", 80))

		for _, entry := range apiChanges {
			current := ""
			if entry.IsCurrent {
				current = " - Current"
			}
			fmt.Fprintf(w, "%s (%s)%s\n", entry.Version, entry.Date.Format("2006-01-02"), current)

			if len(entry.Breaking) > 0 {
				fmt.Fprintln(w, "  Breaking Changes:")
				for _, change := range entry.Breaking {
					fmt.Fprintf(w, "    ⚠️  %s\n", change)
				}
			}

			if len(entry.Added) > 0 {
				fmt.Fprintln(w, "  Added:")
				for _, add := range entry.Added {
					fmt.Fprintf(w, "    + %s\n", add)
				}
			}

			if len(entry.Changed) > 0 {
				fmt.Fprintln(w, "  Changed:")
				for _, change := range entry.Changed {
					fmt.Fprintf(w, "    ~ %s\n", change)
				}
			}

			if len(entry.Deprecated) > 0 {
				fmt.Fprintln(w, "  Deprecated:")
				for _, dep := range entry.Deprecated {
					fmt.Fprintf(w, "    🔔 %s\n", dep)
				}
			}

			if len(entry.Removed) > 0 {
				fmt.Fprintln(w, "  Removed:")
				for _, rem := range entry.Removed {
					fmt.Fprintf(w, "    - %s\n", rem)
				}
			}

			fmt.Fprintln(w)
		}
	}

	if len(binaryChanges) == 0 && len(apiChanges) == 0 {
		fmt.Fprintln(w, "No changelog entries found")
	}

	return nil
}

// formatChangelogJSON formats changelog as JSON.
func formatChangelogJSON(binaryChanges []ChangelogEntry, apiChanges []openapi.ChangelogEntry, w io.Writer) error {
	output := map[string]interface{}{
		"binary_changelog": binaryChanges,
		"api_changelog":    apiChanges,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatChangelogYAML formats changelog as YAML.
func formatChangelogYAML(binaryChanges []ChangelogEntry, apiChanges []openapi.ChangelogEntry, w io.Writer) error {
	if len(binaryChanges) > 0 {
		fmt.Fprintln(w, "binary_changelog:")
		for _, entry := range binaryChanges {
			fmt.Fprintf(w, "  - version: %s\n", entry.Version)
			fmt.Fprintf(w, "    date: %s\n", entry.Date.Format("2006-01-02"))
			if len(entry.Changes) > 0 {
				fmt.Fprintln(w, "    changes:")
				for _, change := range entry.Changes {
					fmt.Fprintf(w, "      - %s\n", change)
				}
			}
		}
	}

	if len(apiChanges) > 0 {
		fmt.Fprintln(w, "api_changelog:")
		for _, entry := range apiChanges {
			fmt.Fprintf(w, "  - version: %s\n", entry.Version)
			fmt.Fprintf(w, "    date: %s\n", entry.Date.Format("2006-01-02"))
			if len(entry.Added) > 0 {
				fmt.Fprintln(w, "    added:")
				for _, add := range entry.Added {
					fmt.Fprintf(w, "      - %s\n", add)
				}
			}
		}
	}

	return nil
}

// filterChangelogSince filters changelog entries since a specific version.
func filterChangelogSince(entries []ChangelogEntry, version string) []ChangelogEntry {
	filtered := []ChangelogEntry{}
	found := false

	for _, entry := range entries {
		if entry.Version == version {
			found = true
		}
		if found {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// filterAPIChangelogSince filters API changelog entries since a specific version.
func filterAPIChangelogSince(entries []openapi.ChangelogEntry, version string) []openapi.ChangelogEntry {
	filtered := []openapi.ChangelogEntry{}
	found := false

	for _, entry := range entries {
		if entry.Version == version {
			found = true
		}
		if found {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

// DefaultBinaryChangelogFunc returns a default binary changelog.
func DefaultBinaryChangelogFunc(config *cli.CLIConfig) func() ([]ChangelogEntry, error) {
	return func() ([]ChangelogEntry, error) {
		// This would typically be embedded in the binary or fetched from a URL
		return []ChangelogEntry{
			{
				Version:   config.Metadata.Version,
				Date:      time.Now(),
				Changes:   []string{"Current version"},
				IsCurrent: true,
			},
		}, nil
	}
}

// DefaultAPIChangelogFunc returns a default API changelog.
func DefaultAPIChangelogFunc(config *cli.CLIConfig) func() ([]openapi.ChangelogEntry, error) {
	return func() ([]openapi.ChangelogEntry, error) {
		// This would typically be parsed from the OpenAPI spec x-changelog extension
		return []openapi.ChangelogEntry{}, nil
	}
}
