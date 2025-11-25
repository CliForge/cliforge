package builtin

import (
	"context"
	"fmt"
	"io"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/spf13/cobra"
)

// UpdateOptions configures the update command behavior.
type UpdateOptions struct {
	Config             *cli.CLIConfig
	CurrentVersion     string
	RequireConfirm     bool
	ShowChangelog      bool
	Output             io.Writer
	CheckUpdateFunc    func(ctx context.Context) (*UpdateInfo, error)
	PerformUpdateFunc  func(ctx context.Context, version string) error
}

// UpdateInfo contains information about available updates.
type UpdateInfo struct {
	CurrentVersion string   `json:"current_version"`
	LatestVersion  string   `json:"latest_version"`
	UpdateAvailable bool    `json:"update_available"`
	DownloadURL    string   `json:"download_url,omitempty"`
	Changelog      []string `json:"changelog,omitempty"`
	ReleaseNotes   string   `json:"release_notes,omitempty"`
}

// NewUpdateCommand creates a new update command.
func NewUpdateCommand(opts *UpdateOptions) *cobra.Command {
	var force bool
	var skipConfirm bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the CLI to the latest version",
		Long: `Check for and install updates to the CLI binary.

The update command will:
1. Check for the latest version
2. Display the changelog
3. Prompt for confirmation (unless --yes is specified)
4. Download and install the update

Examples:
  update              # Check and install updates
  update --yes        # Skip confirmation prompt
  update --force      # Force update even if already latest`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(opts, force, skipConfirm)
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "Force update even if already latest")
	cmd.Flags().BoolVarP(&skipConfirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

// runUpdate executes the update command.
func runUpdate(opts *UpdateOptions, force, skipConfirm bool) error {
	ctx := context.Background()

	fmt.Fprintln(opts.Output, "Checking for updates...")

	// Check for updates
	info, err := opts.CheckUpdateFunc(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	fmt.Fprintf(opts.Output, "Current version: %s\n", info.CurrentVersion)
	fmt.Fprintf(opts.Output, "Latest version: %s\n", info.LatestVersion)
	fmt.Fprintln(opts.Output)

	// Check if update is needed
	if !info.UpdateAvailable && !force {
		fmt.Fprintln(opts.Output, "✓ You are already running the latest version")
		return nil
	}

	// Show changelog if enabled
	if opts.ShowChangelog && len(info.Changelog) > 0 {
		fmt.Fprintln(opts.Output, "Changelog:")
		for _, change := range info.Changelog {
			fmt.Fprintf(opts.Output, "  • %s\n", change)
		}
		fmt.Fprintln(opts.Output)
	}

	// Show release notes if available
	if info.ReleaseNotes != "" {
		fmt.Fprintln(opts.Output, info.ReleaseNotes)
		fmt.Fprintln(opts.Output)
	}

	// Confirm update
	if opts.RequireConfirm && !skipConfirm {
		fmt.Fprint(opts.Output, "Update now? [Y/n]: ")
		var response string
		fmt.Scanln(&response)
		if response != "" && response != "Y" && response != "y" {
			fmt.Fprintln(opts.Output, "Update cancelled")
			return nil
		}
	}

	// Perform update
	fmt.Fprintf(opts.Output, "Downloading %s...\n", info.LatestVersion)

	if err := opts.PerformUpdateFunc(ctx, info.LatestVersion); err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	fmt.Fprintln(opts.Output, "Installing...")
	fmt.Fprintf(opts.Output, "✓ Updated to %s\n", info.LatestVersion)

	return nil
}

// DefaultCheckUpdateFunc is a default implementation for checking updates.
func DefaultCheckUpdateFunc(config *cli.CLIConfig, currentVersion string) func(ctx context.Context) (*UpdateInfo, error) {
	return func(ctx context.Context) (*UpdateInfo, error) {
		// This is a placeholder implementation
		// In a real implementation, this would:
		// 1. Fetch latest release info from GitHub/release server
		// 2. Parse version numbers
		// 3. Compare versions
		// 4. Fetch changelog

		return &UpdateInfo{
			CurrentVersion:  currentVersion,
			LatestVersion:   currentVersion,
			UpdateAvailable: false,
			Changelog:       []string{},
		}, nil
	}
}

// DefaultPerformUpdateFunc is a default implementation for performing updates.
func DefaultPerformUpdateFunc() func(ctx context.Context, version string) error {
	return func(ctx context.Context, version string) error {
		// This is a placeholder implementation
		// In a real implementation, this would:
		// 1. Download the new binary
		// 2. Verify checksum/signature
		// 3. Replace the current binary
		// 4. Set permissions

		return fmt.Errorf("self-update not implemented for this build")
	}
}
