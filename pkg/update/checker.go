package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// Checker handles checking for updates.
type Checker struct {
	config     *UpdateConfig
	httpClient *http.Client
}

// NewChecker creates a new update checker.
func NewChecker(config *UpdateConfig) *Checker {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	return &Checker{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}
}

// Check checks for available updates.
func (c *Checker) Check(ctx context.Context) (*CheckResult, error) {
	// Parse current version
	currentVersion, err := ParseVersion(c.config.CurrentVersion)
	if err != nil {
		return &CheckResult{
			Status:    UpdateStatusFailed,
			Error:     fmt.Errorf("invalid current version: %w", err),
			CheckedAt: time.Now(),
		}, err
	}

	// Fetch release info
	release, err := c.fetchReleaseInfo(ctx)
	if err != nil {
		return &CheckResult{
			Status:         UpdateStatusFailed,
			CurrentVersion: currentVersion,
			Error:          fmt.Errorf("failed to fetch release info: %w", err),
			CheckedAt:      time.Now(),
		}, err
	}

	// Parse latest version
	latestVersion, err := ParseVersion(release.Version)
	if err != nil {
		return &CheckResult{
			Status:         UpdateStatusFailed,
			CurrentVersion: currentVersion,
			Error:          fmt.Errorf("invalid latest version: %w", err),
			CheckedAt:      time.Now(),
		}, err
	}

	// Skip prerelease versions if not allowed
	if latestVersion.IsPrerelease() && !c.config.AllowPrerelease {
		return &CheckResult{
			Status:         UpdateStatusUpToDate,
			CurrentVersion: currentVersion,
			LatestVersion:  latestVersion,
			Release:        release,
			CheckedAt:      time.Now(),
		}, nil
	}

	// Determine status
	status := UpdateStatusUpToDate
	if latestVersion.IsNewer(currentVersion) {
		status = UpdateStatusAvailable
	}

	result := &CheckResult{
		Status:         status,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		Release:        release,
		CheckedAt:      time.Now(),
	}

	// Save last check info
	if err := c.saveLastCheck(result); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to save last check info: %v\n", err)
	}

	return result, nil
}

// fetchReleaseInfo fetches the latest release information from the update server.
func (c *Checker) fetchReleaseInfo(ctx context.Context) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.config.UpdateURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", fmt.Sprintf("CliForge-Update-Checker/%s", c.config.CurrentVersion))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch update info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var release ReleaseInfo
	if err := json.Unmarshal(body, &release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	// Set default checksum algorithm if not specified
	if release.ChecksumAlgo == "" {
		release.ChecksumAlgo = "sha256"
	}

	return &release, nil
}

// GetLastCheck retrieves the last check information.
func (c *Checker) GetLastCheck() (*LastCheckInfo, error) {
	if c.config.StateDir == "" {
		return nil, fmt.Errorf("state directory not configured")
	}

	lastCheckPath := filepath.Join(c.config.StateDir, "last_check.json")

	data, err := os.ReadFile(lastCheckPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &LastCheckInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read last check info: %w", err)
	}

	var info LastCheckInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse last check info: %w", err)
	}

	return &info, nil
}

// saveLastCheck saves the last check information.
func (c *Checker) saveLastCheck(result *CheckResult) error {
	if c.config.StateDir == "" {
		return fmt.Errorf("state directory not configured")
	}

	// Ensure state directory exists
	if err := os.MkdirAll(c.config.StateDir, 0700); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	info := &LastCheckInfo{
		CheckedAt: result.CheckedAt,
	}

	if result.LatestVersion != nil {
		info.LatestVersion = result.LatestVersion.String()
	}

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal last check info: %w", err)
	}

	lastCheckPath := filepath.Join(c.config.StateDir, "last_check.json")
	if err := os.WriteFile(lastCheckPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write last check info: %w", err)
	}

	return nil
}

// SkipVersion marks a version as skipped so the user won't be notified again.
func (c *Checker) SkipVersion(version string) error {
	if c.config.StateDir == "" {
		return fmt.Errorf("state directory not configured")
	}

	info, err := c.GetLastCheck()
	if err != nil {
		return err
	}

	info.UpdateSkipped = true
	info.SkippedVersion = version
	info.SkippedAt = time.Now()

	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal last check info: %w", err)
	}

	lastCheckPath := filepath.Join(c.config.StateDir, "last_check.json")
	if err := os.WriteFile(lastCheckPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write last check info: %w", err)
	}

	return nil
}

// ShouldNotify returns true if the user should be notified about an update.
func (c *Checker) ShouldNotify(result *CheckResult) (bool, error) {
	if !result.UpdateAvailable() {
		return false, nil
	}

	// Check if this version was already skipped
	info, err := c.GetLastCheck()
	if err != nil {
		return true, nil // On error, notify anyway
	}

	if info.UpdateSkipped && info.SkippedVersion == result.LatestVersion.String() {
		return false, nil
	}

	return true, nil
}
