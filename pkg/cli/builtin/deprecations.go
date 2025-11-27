package builtin

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

// DeprecationsOptions configures the deprecations command behavior.
type DeprecationsOptions struct {
	Config                 *cli.Config
	ShowBinaryDeprecations bool
	ShowAPIDeprecations    bool
	AllowScan              bool
	BinaryDeprecationsFunc func() ([]BinaryDeprecation, error)
	APIDeprecationsFunc    func() ([]openapi.Deprecation, error)
	Output                 io.Writer
}

// BinaryDeprecation represents a deprecated CLI feature.
type BinaryDeprecation struct {
	Feature     string    `json:"feature"`
	Replacement string    `json:"replacement,omitempty"`
	Sunset      time.Time `json:"sunset,omitempty"`
	Message     string    `json:"message"`
	DocsURL     string    `json:"docs_url,omitempty"`
	Severity    string    `json:"severity"` // warning, critical
}

// NewDeprecationsCommand creates a new deprecations command.
func NewDeprecationsCommand(opts *DeprecationsOptions) *cobra.Command {
	var binaryOnly bool
	var apiOnly bool
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "deprecations",
		Short: "Show deprecations",
		Long: `Display deprecation notices for both CLI features and API endpoints.

Deprecations show:
- CLI binary deprecations (deprecated flags, commands)
- API deprecations (deprecated endpoints, parameters)

Severity levels:
- WARNING: Deprecated, will be removed in future
- CRITICAL: Will be removed soon, update immediately

Examples:
  deprecations              # Show all deprecations
  deprecations --binary-only # Show only CLI deprecations
  deprecations --api-only    # Show only API deprecations`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeprecations(opts, binaryOnly, apiOnly, outputFormat)
		},
	}

	cmd.Flags().BoolVar(&binaryOnly, "binary-only", false, "Show only CLI binary deprecations")
	cmd.Flags().BoolVar(&apiOnly, "api-only", false, "Show only API deprecations")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")

	// Add subcommands
	if opts.AllowScan {
		cmd.AddCommand(newDeprecationsScanCommand(opts))
	}
	cmd.AddCommand(newDeprecationsShowCommand(opts))

	return cmd
}

// newDeprecationsShowCommand creates the deprecations show subcommand.
func newDeprecationsShowCommand(opts *DeprecationsOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "show <operation-id>",
		Short: "Show details of a specific deprecation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			operationID := args[0]
			return runDeprecationsShow(opts, operationID)
		},
	}
}

// newDeprecationsScanCommand creates the deprecations scan subcommand.
func newDeprecationsScanCommand(opts *DeprecationsOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "scan [path]",
		Short: "Scan code for deprecated API usage",
		Long:  "Scan scripts or code for usage of deprecated API endpoints and parameters.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			return runDeprecationsScan(opts, path)
		},
	}
}

// runDeprecations displays deprecation notices.
func runDeprecations(opts *DeprecationsOptions, binaryOnly, apiOnly bool, outputFormat string) error {
	var binaryDeps []BinaryDeprecation
	var apiDeps []openapi.Deprecation
	var err error

	// Determine what to show
	showBinary := opts.ShowBinaryDeprecations && !apiOnly
	showAPI := opts.ShowAPIDeprecations && !binaryOnly

	// Get binary deprecations
	if showBinary {
		binaryDeps, err = opts.BinaryDeprecationsFunc()
		if err != nil {
			_, _ = fmt.Fprintf(opts.Output, "Warning: failed to load binary deprecations: %v\n", err)
			binaryDeps = []BinaryDeprecation{}
		}
	}

	// Get API deprecations
	if showAPI {
		apiDeps, err = opts.APIDeprecationsFunc()
		if err != nil {
			_, _ = fmt.Fprintf(opts.Output, "Warning: failed to load API deprecations: %v\n", err)
			apiDeps = []openapi.Deprecation{}
		}
	}

	// Format output
	switch outputFormat {
	case "json":
		return formatDeprecationsJSON(binaryDeps, apiDeps, opts.Output)
	case "yaml":
		return formatDeprecationsYAML(binaryDeps, apiDeps, opts.Output)
	default:
		return formatDeprecationsText(binaryDeps, apiDeps, showBinary, showAPI, opts.Output)
	}
}

// runDeprecationsShow shows details of a specific deprecation.
func runDeprecationsShow(opts *DeprecationsOptions, operationID string) error {
	apiDeps, err := opts.APIDeprecationsFunc()
	if err != nil {
		return err
	}

	// Find the deprecation
	for _, dep := range apiDeps {
		if dep.OperationID == operationID {
			return formatDeprecationDetail(&dep, opts.Output)
		}
	}

	return fmt.Errorf("deprecation %q not found", operationID)
}

// runDeprecationsScan scans code for deprecated usage.
func runDeprecationsScan(opts *DeprecationsOptions, path string) error {
	_, _ = fmt.Fprintf(opts.Output, "Scanning %s for deprecated API usage...\n", path)
	_, _ = fmt.Fprintln(opts.Output, "⚠️  Scan functionality not yet implemented")
	return nil
}

// formatDeprecationsText formats deprecations as human-readable text.
func formatDeprecationsText(binaryDeps []BinaryDeprecation, apiDeps []openapi.Deprecation, showBinary, showAPI bool, w io.Writer) error {
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Active Deprecations")
	_, _ = fmt.Fprintln(w, strings.Repeat("━", 80))
	_, _ = fmt.Fprintln(w)

	if showBinary {
		if len(binaryDeps) == 0 {
			_, _ = fmt.Fprintln(w, "CLI Binary Deprecations (0):")
			_, _ = fmt.Fprintln(w, "  ✓ No deprecated CLI features")
			_, _ = fmt.Fprintln(w)
		} else {
			_, _ = fmt.Fprintf(w, "CLI Binary Deprecations (%d):\n", len(binaryDeps))
			_, _ = fmt.Fprintln(w)

			for _, dep := range binaryDeps {
				printBinaryDeprecation(&dep, w)
			}
		}
	}

	if showAPI {
		if len(apiDeps) == 0 {
			_, _ = fmt.Fprintln(w, "API Deprecations (0):")
			_, _ = fmt.Fprintln(w, "  ✓ No deprecated API endpoints")
			_, _ = fmt.Fprintln(w)
		} else {
			_, _ = fmt.Fprintf(w, "API Deprecations (%d):\n", len(apiDeps))
			_, _ = fmt.Fprintln(w)

			// Sort by days remaining
			sort.Slice(apiDeps, func(i, j int) bool {
				if apiDeps[i].Sunset.IsZero() {
					return false
				}
				if apiDeps[j].Sunset.IsZero() {
					return true
				}
				return apiDeps[i].Sunset.Before(apiDeps[j].Sunset)
			})

			for _, dep := range apiDeps {
				printAPIDeprecation(&dep, w)
			}
		}
	}

	_, _ = fmt.Fprintln(w, strings.Repeat("━", 80))
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Show details: deprecations show <operation-id>")

	return nil
}

// printBinaryDeprecation prints a binary deprecation notice.
func printBinaryDeprecation(dep *BinaryDeprecation, w io.Writer) {
	icon := "⚠️  WARNING"
	if dep.Severity == "critical" {
		icon = "🔴 CRITICAL"
	}

	_, _ = fmt.Fprintf(w, "  %s", icon)
	if !dep.Sunset.IsZero() {
		daysRemaining := int(time.Until(dep.Sunset).Hours() / 24)
		_, _ = fmt.Fprintf(w, " - %d days remaining", daysRemaining)
	}
	_, _ = fmt.Fprintln(w)

	_, _ = fmt.Fprintf(w, "  ├─ Feature: %s\n", dep.Feature)
	if dep.Replacement != "" {
		_, _ = fmt.Fprintf(w, "  ├─ Replacement: %s\n", dep.Replacement)
	}
	if !dep.Sunset.IsZero() {
		_, _ = fmt.Fprintf(w, "  ├─ Sunset: %s\n", dep.Sunset.Format("January 2, 2006"))
	}
	if dep.DocsURL != "" {
		_, _ = fmt.Fprintf(w, "  └─ Docs: %s\n", dep.DocsURL)
	} else {
		_, _ = fmt.Fprintf(w, "  └─ Message: %s\n", dep.Message)
	}
	_, _ = fmt.Fprintln(w)
}

// printAPIDeprecation prints an API deprecation notice.
func printAPIDeprecation(dep *openapi.Deprecation, w io.Writer) {
	icon := "⚠️  WARNING"
	if !dep.Sunset.IsZero() {
		daysRemaining := int(time.Until(dep.Sunset).Hours() / 24)
		if daysRemaining < 30 {
			icon = "🔴 CRITICAL"
		}
		_, _ = fmt.Fprintf(w, "  %s - %d days remaining\n", icon, daysRemaining)
	} else {
		_, _ = fmt.Fprintf(w, "  %s\n", icon)
	}

	_, _ = fmt.Fprintf(w, "  ├─ Operation: %s %s", dep.Method, dep.Path)
	if dep.OperationID != "" {
		_, _ = fmt.Fprintf(w, " (%s)", dep.OperationID)
	}
	_, _ = fmt.Fprintln(w)

	if !dep.Sunset.IsZero() {
		_, _ = fmt.Fprintf(w, "  ├─ Sunset: %s\n", dep.Sunset.Format("January 2, 2006"))
	}

	if dep.Replacement != "" {
		_, _ = fmt.Fprintf(w, "  ├─ Replacement: %s\n", dep.Replacement)
	}

	if dep.DocsURL != "" {
		_, _ = fmt.Fprintf(w, "  └─ Docs: %s\n", dep.DocsURL)
	} else if dep.Message != "" {
		_, _ = fmt.Fprintf(w, "  └─ Message: %s\n", dep.Message)
	}

	_, _ = fmt.Fprintln(w)
}

// formatDeprecationDetail formats detailed deprecation information.
func formatDeprecationDetail(dep *openapi.Deprecation, w io.Writer) error {
	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, "Deprecation Details")
	_, _ = fmt.Fprintln(w, strings.Repeat("━", 80))
	_, _ = fmt.Fprintln(w)

	_, _ = fmt.Fprintf(w, "Operation: %s %s\n", dep.Method, dep.Path)
	if dep.OperationID != "" {
		_, _ = fmt.Fprintf(w, "Operation ID: %s\n", dep.OperationID)
	}
	_, _ = fmt.Fprintln(w)

	if !dep.Sunset.IsZero() {
		daysRemaining := int(time.Until(dep.Sunset).Hours() / 24)
		_, _ = fmt.Fprintf(w, "Sunset Date: %s (%d days remaining)\n", dep.Sunset.Format("January 2, 2006"), daysRemaining)
	}

	if dep.Message != "" {
		_, _ = fmt.Fprintln(w, "\nMessage:")
		_, _ = fmt.Fprintf(w, "%s\n", dep.Message)
	}

	if dep.Replacement != "" {
		_, _ = fmt.Fprintln(w, "\nReplacement:")
		_, _ = fmt.Fprintf(w, "%s\n", dep.Replacement)
	}

	if dep.DocsURL != "" {
		_, _ = fmt.Fprintln(w, "\nDocumentation:")
		_, _ = fmt.Fprintf(w, "%s\n", dep.DocsURL)
	}

	_, _ = fmt.Fprintln(w)

	return nil
}

// formatDeprecationsJSON formats deprecations as JSON.
func formatDeprecationsJSON(binaryDeps []BinaryDeprecation, apiDeps []openapi.Deprecation, w io.Writer) error {
	output := map[string]interface{}{
		"binary_deprecations": binaryDeps,
		"api_deprecations":    apiDeps,
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(output)
}

// formatDeprecationsYAML formats deprecations as YAML.
func formatDeprecationsYAML(binaryDeps []BinaryDeprecation, apiDeps []openapi.Deprecation, w io.Writer) error {
	if len(binaryDeps) > 0 {
		_, _ = fmt.Fprintln(w, "binary_deprecations:")
		for _, dep := range binaryDeps {
			_, _ = fmt.Fprintf(w, "  - feature: %s\n", dep.Feature)
			if dep.Replacement != "" {
				_, _ = fmt.Fprintf(w, "    replacement: %s\n", dep.Replacement)
			}
			if !dep.Sunset.IsZero() {
				_, _ = fmt.Fprintf(w, "    sunset: %s\n", dep.Sunset.Format("2006-01-02"))
			}
			_, _ = fmt.Fprintf(w, "    severity: %s\n", dep.Severity)
			_, _ = fmt.Fprintf(w, "    message: %s\n", dep.Message)
		}
	}

	if len(apiDeps) > 0 {
		_, _ = fmt.Fprintln(w, "api_deprecations:")
		for _, dep := range apiDeps {
			_, _ = fmt.Fprintf(w, "  - operation_id: %s\n", dep.OperationID)
			_, _ = fmt.Fprintf(w, "    method: %s\n", dep.Method)
			_, _ = fmt.Fprintf(w, "    path: %s\n", dep.Path)
			if !dep.Sunset.IsZero() {
				_, _ = fmt.Fprintf(w, "    sunset: %s\n", dep.Sunset.Format("2006-01-02"))
			}
			if dep.Replacement != "" {
				_, _ = fmt.Fprintf(w, "    replacement: %s\n", dep.Replacement)
			}
		}
	}

	return nil
}

// DefaultBinaryDeprecationsFunc returns default binary deprecations.
func DefaultBinaryDeprecationsFunc() func() ([]BinaryDeprecation, error) {
	return func() ([]BinaryDeprecation, error) {
		// No binary deprecations by default
		return []BinaryDeprecation{}, nil
	}
}

// DefaultAPIDeprecationsFunc returns default API deprecations.
func DefaultAPIDeprecationsFunc() func() ([]openapi.Deprecation, error) {
	return func() ([]openapi.Deprecation, error) {
		// This would be parsed from OpenAPI spec x-deprecation extensions
		return []openapi.Deprecation{}, nil
	}
}
