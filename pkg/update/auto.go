package update

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/pterm/pterm"
)

// AutoUpdater handles automatic update checking and installation.
type AutoUpdater struct {
	config     *UpdateConfig
	checker    *Checker
	downloader *Downloader
	installer  *Installer
}

// NewAutoUpdater creates a new auto-updater.
func NewAutoUpdater(config *UpdateConfig) *AutoUpdater {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	return &AutoUpdater{
		config:     config,
		checker:    NewChecker(config),
		downloader: NewDownloader(config),
		installer:  NewInstaller(config),
	}
}

// CheckAndNotify checks for updates and notifies the user if available.
// This is meant to be called on CLI startup.
func (au *AutoUpdater) CheckAndNotify(ctx context.Context) error {
	// Check if we should check for updates
	lastCheck, err := au.checker.GetLastCheck()
	if err != nil {
		// Non-fatal, just log
		fmt.Fprintf(os.Stderr, "Warning: failed to get last check info: %v\n", err)
		lastCheck = &LastCheckInfo{}
	}

	if !lastCheck.ShouldCheck(au.config.CheckInterval) {
		return nil // Too soon to check again
	}

	// Perform check in background with timeout
	checkCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := au.checker.Check(checkCtx)
	if err != nil {
		// Don't fail CLI startup due to update check error
		return nil
	}

	// Check if we should notify
	shouldNotify, err := au.checker.ShouldNotify(result)
	if err != nil || !shouldNotify {
		return nil
	}

	// Notify user
	au.notify(result)

	return nil
}

// notify displays an update notification to the user.
func (au *AutoUpdater) notify(result *CheckResult) {
	if !result.UpdateAvailable() {
		return
	}

	// Create notification box
	info := pterm.Info.WithShowLineNumber(false)

	message := fmt.Sprintf("A new version of %s is available: %s (current: %s)",
		"CliForge",
		result.LatestVersion.String(),
		result.CurrentVersion.String())

	if result.Release != nil && result.Release.Critical {
		message += "\n⚠️  This is a critical update and is strongly recommended."
	}

	message += fmt.Sprintf("\n\nRun '%s update' to update now.", os.Args[0])

	info.Println(message)
}

// Update performs a full update (check, download, install).
func (au *AutoUpdater) Update(ctx context.Context) error {
	// Check for updates
	result, err := au.checker.Check(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !result.UpdateAvailable() {
		pterm.Success.Println("You are already running the latest version!")
		return nil
	}

	// Show update information
	au.showUpdateInfo(result)

	// Ask for confirmation if required
	if au.config.RequireConfirmation {
		if !au.confirm(result) {
			pterm.Info.Println("Update cancelled.")
			return nil
		}
	}

	// Download update
	pterm.Info.Println("Downloading update...")
	downloadPath, err := au.downloader.DownloadWithProgress(ctx, result.Release)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer os.Remove(downloadPath)

	// Install update
	pterm.Info.Println("Installing update...")
	if err := au.installer.Install(downloadPath); err != nil {
		return fmt.Errorf("failed to install update: %w", err)
	}

	pterm.Success.Printf("Successfully updated to version %s!\n", result.LatestVersion.String())
	pterm.Info.Println("Please restart the CLI to use the new version.")

	return nil
}

// showUpdateInfo displays information about the available update.
func (au *AutoUpdater) showUpdateInfo(result *CheckResult) {
	pterm.DefaultHeader.WithFullWidth().Println("Update Available")

	data := pterm.TableData{
		{"Current Version", result.CurrentVersion.String()},
		{"Latest Version", result.LatestVersion.String()},
	}

	if result.Release != nil {
		if !result.Release.ReleaseDate.IsZero() {
			data = append(data, []string{"Release Date", result.Release.ReleaseDate.Format("2006-01-02")})
		}
		if result.Release.Size > 0 {
			data = append(data, []string{"Download Size", formatBytes(result.Release.Size)})
		}
		if result.Release.Critical {
			data = append(data, []string{"Priority", "CRITICAL"})
		}
	}

	pterm.DefaultTable.WithHasHeader(false).WithData(data).Render()

	// Show changelog if available
	if result.Release != nil && result.Release.Changelog != "" {
		pterm.Println()
		pterm.DefaultHeader.WithFullWidth().Println("Changelog")
		pterm.Println(result.Release.Changelog)
	}

	pterm.Println()
}

// confirm asks the user for confirmation to proceed with the update.
func (au *AutoUpdater) confirm(result *CheckResult) bool {
	message := fmt.Sprintf("Do you want to update to version %s?", result.LatestVersion.String())

	confirmed, err := pterm.DefaultInteractiveConfirm.Show(message)
	if err != nil {
		// If interactive confirmation fails, default to no
		return false
	}

	return confirmed
}

// UpdateAndRestart performs a full update and restarts the CLI.
func (au *AutoUpdater) UpdateAndRestart(ctx context.Context, args []string) error {
	// Check for updates
	result, err := au.checker.Check(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !result.UpdateAvailable() {
		pterm.Success.Println("You are already running the latest version!")
		return nil
	}

	// Show update information
	au.showUpdateInfo(result)

	// Ask for confirmation if required
	if au.config.RequireConfirmation {
		if !au.confirm(result) {
			pterm.Info.Println("Update cancelled.")
			return nil
		}
	}

	// Download update
	pterm.Info.Println("Downloading update...")
	downloadPath, err := au.downloader.DownloadWithProgress(ctx, result.Release)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}

	// Install and restart
	pterm.Info.Println("Installing update and restarting...")
	if err := au.installer.InstallAndRestart(downloadPath, args); err != nil {
		return fmt.Errorf("failed to install and restart: %w", err)
	}

	// This point should never be reached as InstallAndRestart exits the process
	return nil
}

// SkipVersion skips the current available version.
func (au *AutoUpdater) SkipVersion(ctx context.Context) error {
	result, err := au.checker.Check(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	if !result.UpdateAvailable() {
		pterm.Info.Println("No updates available to skip.")
		return nil
	}

	if err := au.checker.SkipVersion(result.LatestVersion.String()); err != nil {
		return fmt.Errorf("failed to skip version: %w", err)
	}

	pterm.Success.Printf("Version %s will be skipped.\n", result.LatestVersion.String())
	return nil
}

// Status shows the current update status.
func (au *AutoUpdater) Status(ctx context.Context) error {
	pterm.DefaultHeader.WithFullWidth().Println("Update Status")

	// Check for updates
	result, err := au.checker.Check(ctx)
	if err != nil {
		return fmt.Errorf("failed to check for updates: %w", err)
	}

	data := pterm.TableData{
		{"Current Version", result.CurrentVersion.String()},
	}

	if result.UpdateAvailable() {
		data = append(data, []string{"Status", pterm.Yellow("Update Available")})
		data = append(data, []string{"Latest Version", result.LatestVersion.String()})

		if result.Release != nil && result.Release.Critical {
			data = append(data, []string{"Priority", pterm.Red("CRITICAL")})
		}
	} else {
		data = append(data, []string{"Status", pterm.Green("Up to date")})
	}

	// Show last check info
	lastCheck, err := au.checker.GetLastCheck()
	if err == nil && !lastCheck.CheckedAt.IsZero() {
		data = append(data, []string{"Last Checked", lastCheck.CheckedAt.Format("2006-01-02 15:04:05")})
	}

	pterm.DefaultTable.WithHasHeader(false).WithData(data).Render()

	if result.UpdateAvailable() {
		pterm.Println()
		pterm.Info.Printf("Run '%s update' to update now.\n", os.Args[0])
	}

	return nil
}

// CleanupCache removes old downloaded files.
func (au *AutoUpdater) CleanupCache() error {
	if err := au.downloader.CleanupOldDownloads(); err != nil {
		return fmt.Errorf("failed to cleanup cache: %w", err)
	}

	pterm.Success.Println("Cache cleaned successfully.")
	return nil
}
