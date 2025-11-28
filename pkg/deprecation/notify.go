package deprecation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// Notifier manages deprecation notifications and changelog display.
type Notifier struct {
	appName         string
	lastVersionFile string
	lastVersion     string
}

// NewNotifier creates a new Notifier instance.
func NewNotifier(appName string) (*Notifier, error) {
	lastVersionFile := filepath.Join(xdg.DataHome, appName, "last-version.txt")

	// Load last version
	lastVersion := ""
	if data, err := os.ReadFile(lastVersionFile); err == nil {
		lastVersion = strings.TrimSpace(string(data))
	}

	return &Notifier{
		appName:         appName,
		lastVersionFile: lastVersionFile,
		lastVersion:     lastVersion,
	}, nil
}

// ShowUpdateNotification shows a notification when CLI is updated.
func (n *Notifier) ShowUpdateNotification(currentVersion string, changelog []openapi.ChangelogEntry) error {
	// Check if version has changed
	if n.lastVersion == currentVersion {
		return nil // No update
	}

	// Display update banner
	fmt.Println(n.formatUpdateBanner(n.lastVersion, currentVersion, changelog))

	// Save current version
	if err := n.saveVersion(currentVersion); err != nil {
		return fmt.Errorf("failed to save version: %w", err)
	}

	return nil
}

// formatUpdateBanner formats an update notification banner.
func (n *Notifier) formatUpdateBanner(oldVersion, newVersion string, changelog []openapi.ChangelogEntry) string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔════════════════════════════════════════════════════════╗\n")
	if oldVersion != "" {
		sb.WriteString(fmt.Sprintf("║  📢 CLI Update: v%s → v%s", oldVersion, newVersion))
	} else {
		sb.WriteString(fmt.Sprintf("║  📢 CLI Version: v%s", newVersion))
	}

	// Pad to align with border
	padding := 56 - len(fmt.Sprintf("  📢 CLI Update: v%s → v%s", oldVersion, newVersion))
	if padding > 0 {
		sb.WriteString(strings.Repeat(" ", padding))
	}
	sb.WriteString("║\n")
	sb.WriteString("╠════════════════════════════════════════════════════════╣\n")
	sb.WriteString("║                                                        ║\n")

	// Extract deprecations from latest changelog
	deprecations := extractDeprecationsFromChangelog(changelog)

	if len(deprecations) > 0 {
		sb.WriteString("║  Deprecations in this release:                        ║\n")
		sb.WriteString("║                                                        ║\n")

		for i, dep := range deprecations {
			if i >= 3 { // Limit to 3 deprecations in banner
				sb.WriteString("║  ... and more                                          ║\n")
				break
			}

			icon := getWarningIcon(WarningLevelWarning)
			line := fmt.Sprintf("║  %s  %s", icon, dep.Description)
			if len(line) < 56 {
				line += strings.Repeat(" ", 56-len(line))
			} else if len(line) > 56 {
				line = line[:53] + "..."
			}
			sb.WriteString(line + "║\n")
		}
		sb.WriteString("║                                                        ║\n")
	}

	sb.WriteString("║  Run 'deprecations' for details                        ║\n")
	sb.WriteString("║  Run 'changelog' for full release notes               ║\n")
	sb.WriteString("║                                                        ║\n")
	sb.WriteString("╚════════════════════════════════════════════════════════╝\n")
	sb.WriteString("\n")

	return sb.String()
}

// extractDeprecationsFromChangelog extracts deprecation changes from changelog.
func extractDeprecationsFromChangelog(changelog []openapi.ChangelogEntry) []*openapi.Change {
	var deprecations []*openapi.Change

	for _, entry := range changelog {
		for _, change := range entry.Changes {
			if change.Type == "deprecated" {
				deprecations = append(deprecations, change)
			}
		}
	}

	return deprecations
}

// FormatChangelog formats changelog entries for display.
func FormatChangelog(entries []openapi.ChangelogEntry, limit int) string {
	if len(entries) == 0 {
		return "No changelog entries available"
	}

	var sb strings.Builder

	sb.WriteString("\nChangelog\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	count := 0
	for _, entry := range entries {
		if limit > 0 && count >= limit {
			break
		}
		count++

		sb.WriteString(fmt.Sprintf("## Version %s (%s)\n\n", entry.Version, entry.Date))

		// Group changes by type
		byType := make(map[string][]*openapi.Change)
		for _, change := range entry.Changes {
			byType[change.Type] = append(byType[change.Type], change)
		}

		// Display in order: breaking, deprecated, removed, modified, added
		order := []string{"breaking", "deprecated", "removed", "modified", "added", "security"}

		for _, changeType := range order {
			if changes, ok := byType[changeType]; ok && len(changes) > 0 {
				sb.WriteString(formatChangeSection(changeType, changes))
			}
		}

		sb.WriteString("\n")
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return sb.String()
}

// formatChangeSection formats a section of changes by type.
func formatChangeSection(changeType string, changes []*openapi.Change) string {
	var sb strings.Builder

	// Section header
	switch changeType {
	case "breaking":
		sb.WriteString("### 🔴 BREAKING CHANGES\n\n")
	case "deprecated":
		sb.WriteString("### ⚠️  DEPRECATIONS\n\n")
	case "removed":
		sb.WriteString("### ❌ REMOVED\n\n")
	case "modified":
		sb.WriteString("### 📝 MODIFIED\n\n")
	case "added":
		sb.WriteString("### ✨ ADDED\n\n")
	case "security":
		sb.WriteString("### 🔒 SECURITY\n\n")
	default:
		sb.WriteString(fmt.Sprintf("### %s\n\n", strings.ToUpper(changeType)))
	}

	// List changes
	for _, change := range changes {
		sb.WriteString(fmt.Sprintf("- %s", change.Description))

		if change.Path != "" {
			sb.WriteString(fmt.Sprintf(" (%s)", change.Path))
		}

		sb.WriteString("\n")

		if change.Migration != "" {
			sb.WriteString(fmt.Sprintf("  Migration: %s\n", change.Migration))
		}

		if change.Sunset != "" {
			sb.WriteString(fmt.Sprintf("  Sunset: %s\n", change.Sunset))
		}
	}

	sb.WriteString("\n")

	return sb.String()
}

// AcknowledgmentTracker tracks user acknowledgments of deprecations.
type AcknowledgmentTracker struct {
	trackingPath string
	data         *AcknowledgmentData
}

// AcknowledgmentData stores acknowledgment tracking data.
type AcknowledgmentData struct {
	Acknowledged map[string]time.Time `yaml:"acknowledged"`
}

// NewAcknowledgmentTracker creates a new AcknowledgmentTracker.
func NewAcknowledgmentTracker(appName string) (*AcknowledgmentTracker, error) {
	trackingPath := filepath.Join(xdg.DataHome, appName, "acknowledgments.yaml")

	data := &AcknowledgmentData{
		Acknowledged: make(map[string]time.Time),
	}

	// Load existing data
	if fileData, err := os.ReadFile(trackingPath); err == nil {
		_ = yaml.Unmarshal(fileData, data)
	}

	return &AcknowledgmentTracker{
		trackingPath: trackingPath,
		data:         data,
	}, nil
}

// IsAcknowledged checks if a deprecation has been acknowledged.
func (at *AcknowledgmentTracker) IsAcknowledged(key string) bool {
	_, exists := at.data.Acknowledged[key]
	return exists
}

// Acknowledge marks a deprecation as acknowledged.
func (at *AcknowledgmentTracker) Acknowledge(key string) error {
	at.data.Acknowledged[key] = time.Now()
	return at.save()
}

// GetAcknowledgmentTime returns when a deprecation was acknowledged.
func (at *AcknowledgmentTracker) GetAcknowledgmentTime(key string) *time.Time {
	if t, exists := at.data.Acknowledged[key]; exists {
		return &t
	}
	return nil
}

// ClearAcknowledgment removes an acknowledgment.
func (at *AcknowledgmentTracker) ClearAcknowledgment(key string) error {
	delete(at.data.Acknowledged, key)
	return at.save()
}

// save persists acknowledgment data to disk.
func (at *AcknowledgmentTracker) save() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(at.trackingPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tracking directory: %w", err)
	}

	data, err := yaml.Marshal(at.data)
	if err != nil {
		return fmt.Errorf("failed to marshal acknowledgment data: %w", err)
	}

	if err := os.WriteFile(at.trackingPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write acknowledgment file: %w", err)
	}

	return nil
}

// saveVersion saves the current version to disk.
func (n *Notifier) saveVersion(version string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(n.lastVersionFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create version directory: %w", err)
	}

	if err := os.WriteFile(n.lastVersionFile, []byte(version), 0644); err != nil {
		return fmt.Errorf("failed to write version file: %w", err)
	}

	n.lastVersion = version
	return nil
}

// GetLastVersion returns the last recorded version.
func (n *Notifier) GetLastVersion() string {
	return n.lastVersion
}

// ShowDeprecationNotice displays a deprecation notice.
func ShowDeprecationNotice(info *DeprecationInfo, manager *WarningManager) {
	if !manager.ShouldShowWarning(info) {
		return
	}

	// Display the warning
	fmt.Fprint(os.Stderr, FormatWarning(info))

	// Mark as shown
	_ = manager.MarkShown(info)
}

// ShowShortNotice displays a short deprecation notice.
func ShowShortNotice(info *DeprecationInfo, manager *WarningManager) {
	if !manager.ShouldShowWarning(info) {
		return
	}

	// Display short warning
	fmt.Fprintln(os.Stderr, FormatShortWarning(info))

	// Mark as shown
	_ = manager.MarkShown(info)
}

// ShowBlockedNotice displays a notice for blocked operations.
func ShowBlockedNotice(info *DeprecationInfo) error {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("❌ OPERATION BLOCKED\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	if info.OperationID != "" {
		sb.WriteString(fmt.Sprintf("  Operation: %s\n", info.OperationID))
	}

	if info.Sunset != nil {
		sb.WriteString(fmt.Sprintf("  Sunset Date: %s\n", info.Sunset.Format("January 2, 2006")))
		if info.DaysRemaining < 0 {
			sb.WriteString(fmt.Sprintf("  Status: Removed %d days ago\n", -info.DaysRemaining))
		} else {
			sb.WriteString(fmt.Sprintf("  Status: Sunset in %d days\n", info.DaysRemaining))
		}
	}

	sb.WriteString("\n  This operation is no longer available.\n")

	if info.Replacement != nil {
		sb.WriteString("\n  Replacement:\n")
		if info.Replacement.Command != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", info.Replacement.Command))
		}
		if info.Replacement.Migration != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", info.Replacement.Migration))
		}
	}

	if info.DocsURL != "" {
		sb.WriteString(fmt.Sprintf("\n  For help: %s\n", info.DocsURL))
	}

	sb.WriteString("\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	fmt.Fprint(os.Stderr, sb.String())

	return fmt.Errorf("operation blocked: %s", info.OperationID)
}

// ShowCriticalNotice displays a critical deprecation notice requiring force flag.
func ShowCriticalNotice(info *DeprecationInfo) error {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("🔴 CRITICAL DEPRECATION\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	if info.Sunset != nil {
		sb.WriteString(fmt.Sprintf("  Removal imminent: %d days remaining\n", info.DaysRemaining))
		sb.WriteString(fmt.Sprintf("  Sunset Date: %s\n", info.Sunset.Format("January 2, 2006")))
	}

	sb.WriteString("\n  This operation will stop working soon!\n")

	if info.Replacement != nil && info.Replacement.Command != "" {
		sb.WriteString(fmt.Sprintf("\n  Use instead: %s\n", info.Replacement.Command))
	}

	sb.WriteString("\n  To proceed anyway, use: --force\n")

	if info.DocsURL != "" {
		sb.WriteString(fmt.Sprintf("\n  Documentation: %s\n", info.DocsURL))
	}

	sb.WriteString("\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	fmt.Fprint(os.Stderr, sb.String())

	return fmt.Errorf("operation requires --force flag due to critical deprecation")
}

// FormatDeprecationList formats a list of active deprecations.
func FormatDeprecationList(deprecations []*DeprecationInfo) string {
	if len(deprecations) == 0 {
		return "No active deprecations"
	}

	var sb strings.Builder

	sb.WriteString("\nActive Deprecations\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Group by type
	byType := make(map[DeprecationType][]*DeprecationInfo)
	for _, dep := range deprecations {
		byType[dep.Type] = append(byType[dep.Type], dep)
	}

	// Display API deprecations
	if apiDeps, ok := byType[DeprecationTypeOperation]; ok && len(apiDeps) > 0 {
		sb.WriteString(fmt.Sprintf("API Deprecations (%d):\n\n", len(apiDeps)))
		for _, dep := range apiDeps {
			sb.WriteString(formatDeprecationSummary(dep))
			sb.WriteString("\n")
		}
	}

	// Display parameter deprecations
	if paramDeps, ok := byType[DeprecationTypeParameter]; ok && len(paramDeps) > 0 {
		sb.WriteString(fmt.Sprintf("Parameter Deprecations (%d):\n\n", len(paramDeps)))
		for _, dep := range paramDeps {
			sb.WriteString(formatDeprecationSummary(dep))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\nRun 'deprecations show <id>' for details\n")

	return sb.String()
}

// formatDeprecationSummary formats a summary of a single deprecation.
func formatDeprecationSummary(dep *DeprecationInfo) string {
	icon := getWarningIcon(dep.Level)
	level := strings.ToUpper(string(dep.Level))

	var sb strings.Builder

	if dep.Sunset != nil {
		sb.WriteString(fmt.Sprintf("  %s %s - %d days remaining\n", icon, level, dep.DaysRemaining))
	} else {
		sb.WriteString(fmt.Sprintf("  %s %s\n", icon, level))
	}

	if dep.OperationID != "" {
		sb.WriteString(fmt.Sprintf("  ├─ Operation: %s %s (%s)\n", dep.Method, dep.Path, dep.OperationID))
	} else if dep.Name != "" {
		sb.WriteString(fmt.Sprintf("  ├─ %s: %s\n", dep.Type, dep.Name))
	}

	if dep.Sunset != nil {
		sb.WriteString(fmt.Sprintf("  ├─ Sunset: %s\n", dep.Sunset.Format("January 2, 2006")))
	}

	if dep.Replacement != nil && dep.Replacement.Command != "" {
		sb.WriteString(fmt.Sprintf("  ├─ Replacement: %s\n", dep.Replacement.Command))
	}

	if dep.DocsURL != "" {
		sb.WriteString(fmt.Sprintf("  └─ Docs: %s\n", dep.DocsURL))
	}

	return sb.String()
}
