package deprecation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// WarningLevel represents the urgency of a deprecation warning.
type WarningLevel string

const (
	// WarningLevelInfo - shown once per week (> 6 months until sunset)
	WarningLevelInfo WarningLevel = "info"
	// WarningLevelWarning - shown every time (3-6 months until sunset)
	WarningLevelWarning WarningLevel = "warning"
	// WarningLevelUrgent - shown every time, yellow (1-3 months until sunset)
	WarningLevelUrgent WarningLevel = "urgent"
	// WarningLevelCritical - shown every time, red, requires --force (< 1 month)
	WarningLevelCritical WarningLevel = "critical"
	// WarningLevelRemoved - past sunset, operation blocked
	WarningLevelRemoved WarningLevel = "removed"
)

// WarningManager manages deprecation warnings and user preferences.
type WarningManager struct {
	// User preferences
	config *WarningConfig
	// Tracking file path
	trackingPath string
	// Tracking data
	tracking *WarningTracking
}

// WarningConfig contains user preferences for deprecation warnings.
type WarningConfig struct {
	// Enable/disable all deprecation warnings
	Enabled bool `yaml:"enabled"`
	// Minimum severity level to show
	MinSeverity Severity `yaml:"min_severity"`
	// Minimum warning level to show
	MinLevel WarningLevel `yaml:"min_level"`
	// Show warnings in CI environments
	ShowInCI bool `yaml:"show_in_ci"`
	// Suppressed operations (by operation ID)
	SuppressedOperations []string `yaml:"suppressed_operations"`
	// Suppressed paths
	SuppressedPaths []string `yaml:"suppressed_paths"`
	// Info-level cooldown period
	InfoCooldown time.Duration `yaml:"info_cooldown"`
}

// WarningTracking tracks when warnings were last shown.
type WarningTracking struct {
	LastShown map[string]time.Time `yaml:"last_shown"`
}

// NewWarningManager creates a new WarningManager with default config.
func NewWarningManager(appName string) (*WarningManager, error) {
	trackingPath := filepath.Join(xdg.DataHome, appName, "deprecation-tracking.yaml")

	config := &WarningConfig{
		Enabled:              true,
		MinSeverity:          SeverityInfo,
		MinLevel:             WarningLevelInfo,
		ShowInCI:             false,
		SuppressedOperations: []string{},
		SuppressedPaths:      []string{},
		InfoCooldown:         7 * 24 * time.Hour, // 1 week
	}

	tracking := &WarningTracking{
		LastShown: make(map[string]time.Time),
	}

	// Load existing tracking data
	if data, err := os.ReadFile(trackingPath); err == nil {
		yaml.Unmarshal(data, tracking)
	}

	return &WarningManager{
		config:       config,
		trackingPath: trackingPath,
		tracking:     tracking,
	}, nil
}

// SetConfig updates the warning configuration.
func (wm *WarningManager) SetConfig(config *WarningConfig) {
	wm.config = config
}

// ShouldShowWarning determines if a warning should be shown to the user.
func (wm *WarningManager) ShouldShowWarning(info *DeprecationInfo) bool {
	// Check if warnings are globally disabled
	if !wm.config.Enabled {
		return false
	}

	// Check if running in CI and CI warnings are disabled
	if !wm.config.ShowInCI && isCI() {
		return false
	}

	// Check if operation is suppressed
	if wm.isOperationSuppressed(info) {
		return false
	}

	// Check if severity is below minimum
	if !wm.meetsMinSeverity(info.Severity) {
		return false
	}

	// Check if level is below minimum
	if !wm.meetsMinLevel(info.Level) {
		return false
	}

	// For info-level warnings, check cooldown period
	if info.Level == WarningLevelInfo {
		if !wm.shouldShowInfoWarning(info) {
			return false
		}
	}

	return true
}

// MarkShown records that a warning was shown.
func (wm *WarningManager) MarkShown(info *DeprecationInfo) error {
	key := wm.getTrackingKey(info)
	wm.tracking.LastShown[key] = time.Now()
	return wm.saveTracking()
}

// calculateWarningLevel calculates the warning level based on days remaining.
func calculateWarningLevel(daysRemaining int) WarningLevel {
	if daysRemaining < 0 {
		return WarningLevelRemoved
	} else if daysRemaining < 30 {
		return WarningLevelCritical
	} else if daysRemaining < 90 {
		return WarningLevelUrgent
	} else if daysRemaining < 180 {
		return WarningLevelWarning
	} else {
		return WarningLevelInfo
	}
}

// compareWarningLevels compares two warning levels.
// Returns: 1 if a > b, -1 if a < b, 0 if equal
func compareWarningLevels(a, b WarningLevel) int {
	levels := map[WarningLevel]int{
		WarningLevelInfo:     1,
		WarningLevelWarning:  2,
		WarningLevelUrgent:   3,
		WarningLevelCritical: 4,
		WarningLevelRemoved:  5,
	}

	aVal := levels[a]
	bVal := levels[b]

	if aVal > bVal {
		return 1
	} else if aVal < bVal {
		return -1
	}
	return 0
}

// isOperationSuppressed checks if an operation is in the suppression list.
func (wm *WarningManager) isOperationSuppressed(info *DeprecationInfo) bool {
	for _, suppressed := range wm.config.SuppressedOperations {
		if info.OperationID == suppressed {
			return true
		}
	}

	for _, suppressedPath := range wm.config.SuppressedPaths {
		if info.Path == suppressedPath {
			return true
		}
	}

	return false
}

// meetsMinSeverity checks if severity meets minimum threshold.
func (wm *WarningManager) meetsMinSeverity(severity Severity) bool {
	severityLevels := map[Severity]int{
		SeverityInfo:     1,
		SeverityWarning:  2,
		SeverityBreaking: 3,
	}

	return severityLevels[severity] >= severityLevels[wm.config.MinSeverity]
}

// meetsMinLevel checks if warning level meets minimum threshold.
func (wm *WarningManager) meetsMinLevel(level WarningLevel) bool {
	return compareWarningLevels(level, wm.config.MinLevel) >= 0
}

// shouldShowInfoWarning checks if an info-level warning should be shown
// based on cooldown period.
func (wm *WarningManager) shouldShowInfoWarning(info *DeprecationInfo) bool {
	key := wm.getTrackingKey(info)
	lastShown, exists := wm.tracking.LastShown[key]

	if !exists {
		return true // Never shown before
	}

	// Check if cooldown period has elapsed
	elapsed := time.Since(lastShown)
	return elapsed >= wm.config.InfoCooldown
}

// getTrackingKey generates a unique key for tracking warnings.
func (wm *WarningManager) getTrackingKey(info *DeprecationInfo) string {
	if info.OperationID != "" {
		return fmt.Sprintf("op:%s", info.OperationID)
	}
	if info.Path != "" && info.Name != "" {
		return fmt.Sprintf("param:%s:%s", info.Path, info.Name)
	}
	if info.Path != "" {
		return fmt.Sprintf("path:%s", info.Path)
	}
	return "unknown"
}

// saveTracking persists tracking data to disk.
func (wm *WarningManager) saveTracking() error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(wm.trackingPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create tracking directory: %w", err)
	}

	data, err := yaml.Marshal(wm.tracking)
	if err != nil {
		return fmt.Errorf("failed to marshal tracking data: %w", err)
	}

	if err := os.WriteFile(wm.trackingPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tracking file: %w", err)
	}

	return nil
}

// isCI checks if running in a CI environment.
func isCI() bool {
	ciEnvVars := []string{
		"CI",
		"CONTINUOUS_INTEGRATION",
		"GITHUB_ACTIONS",
		"GITLAB_CI",
		"CIRCLECI",
		"TRAVIS",
		"JENKINS_URL",
	}

	for _, envVar := range ciEnvVars {
		if os.Getenv(envVar) != "" {
			return true
		}
	}

	return false
}

// FormatWarning formats a deprecation warning for display.
func FormatWarning(info *DeprecationInfo) string {
	var sb strings.Builder

	// Header with icon and level
	icon := getWarningIcon(info.Level)
	level := strings.ToUpper(string(info.Level))

	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s DEPRECATION %s\n", icon, level))
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	// Type and name
	if info.OperationID != "" {
		sb.WriteString(fmt.Sprintf("  Operation: %s %s (%s)\n", info.Method, info.Path, info.OperationID))
	} else if info.Name != "" {
		sb.WriteString(fmt.Sprintf("  %s: %s\n", info.Type, info.Name))
	}

	// Sunset information
	if info.Sunset != nil {
		if info.DaysRemaining >= 0 {
			sb.WriteString(fmt.Sprintf("  Sunset: %s (%d days remaining)\n",
				info.Sunset.Format("January 2, 2006"), info.DaysRemaining))
		} else {
			sb.WriteString(fmt.Sprintf("  Sunset: %s (EXPIRED %d days ago)\n",
				info.Sunset.Format("January 2, 2006"), -info.DaysRemaining))
		}
	}

	// Reason
	if info.Reason != "" {
		sb.WriteString(fmt.Sprintf("\n  Reason: %s\n", info.Reason))
	}

	// Replacement/Migration
	if info.Replacement != nil {
		sb.WriteString("\n  Migration:\n")
		if info.Replacement.Command != "" {
			sb.WriteString(fmt.Sprintf("    Use: %s\n", info.Replacement.Command))
		}
		if info.Replacement.Migration != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", info.Replacement.Migration))
		}
		if info.Replacement.Example != "" {
			sb.WriteString(fmt.Sprintf("\n    Example:\n    %s\n", info.Replacement.Example))
		}
	} else if info.Migration != "" {
		sb.WriteString(fmt.Sprintf("\n  Migration:\n    %s\n", info.Migration))
	}

	// Breaking changes
	if len(info.BreakingChanges) > 0 {
		sb.WriteString("\n  Breaking Changes:\n")
		for _, change := range info.BreakingChanges {
			sb.WriteString(fmt.Sprintf("    - %s\n", change))
		}
	}

	// Documentation
	if info.DocsURL != "" {
		sb.WriteString(fmt.Sprintf("\n  Docs: %s\n", info.DocsURL))
	}

	// Suppression hint
	sb.WriteString("\n  To suppress: --no-deprecation-warnings\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	sb.WriteString("\n")

	return sb.String()
}

// FormatShortWarning formats a concise deprecation warning.
func FormatShortWarning(info *DeprecationInfo) string {
	icon := getWarningIcon(info.Level)
	msg := fmt.Sprintf("%s ", icon)

	if info.OperationID != "" {
		msg += fmt.Sprintf("Operation '%s' is deprecated", info.OperationID)
	} else if info.Name != "" {
		msg += fmt.Sprintf("%s '%s' is deprecated", info.Type, info.Name)
	} else {
		msg += "This feature is deprecated"
	}

	if info.Sunset != nil && info.DaysRemaining >= 0 {
		msg += fmt.Sprintf(" (%d days remaining)", info.DaysRemaining)
	}

	if info.Replacement != nil && info.Replacement.Command != "" {
		msg += fmt.Sprintf(" - use '%s' instead", info.Replacement.Command)
	}

	return msg
}

// getWarningIcon returns the appropriate icon for a warning level.
func getWarningIcon(level WarningLevel) string {
	switch level {
	case WarningLevelInfo:
		return "ℹ️"
	case WarningLevelWarning:
		return "⚠️"
	case WarningLevelUrgent:
		return "🚨"
	case WarningLevelCritical:
		return "🔴"
	case WarningLevelRemoved:
		return "❌"
	default:
		return "⚠️"
	}
}

// GetWarningColor returns ANSI color code for a warning level.
func GetWarningColor(level WarningLevel) string {
	switch level {
	case WarningLevelInfo:
		return "\033[36m" // Cyan
	case WarningLevelWarning:
		return "\033[33m" // Yellow
	case WarningLevelUrgent:
		return "\033[93m" // Bright yellow
	case WarningLevelCritical:
		return "\033[91m" // Bright red
	case WarningLevelRemoved:
		return "\033[31m" // Red
	default:
		return "\033[0m" // Reset
	}
}

// ResetColor returns ANSI reset code.
func ResetColor() string {
	return "\033[0m"
}

// RequiresForceFlag checks if deprecation requires --force flag to proceed.
func RequiresForceFlag(info *DeprecationInfo) bool {
	return info.Level == WarningLevelCritical
}

// IsBlocked checks if operation is completely blocked.
func IsBlocked(info *DeprecationInfo) bool {
	return info.Level == WarningLevelRemoved
}

// SuppressOperation adds an operation to the suppression list.
func (wm *WarningManager) SuppressOperation(operationID string) error {
	// Check if already suppressed
	for _, id := range wm.config.SuppressedOperations {
		if id == operationID {
			return nil // Already suppressed
		}
	}

	wm.config.SuppressedOperations = append(wm.config.SuppressedOperations, operationID)
	return nil
}

// UnsuppressOperation removes an operation from the suppression list.
func (wm *WarningManager) UnsuppressOperation(operationID string) error {
	filtered := []string{}
	for _, id := range wm.config.SuppressedOperations {
		if id != operationID {
			filtered = append(filtered, id)
		}
	}
	wm.config.SuppressedOperations = filtered
	return nil
}

// GetSuppressedOperations returns list of suppressed operations.
func (wm *WarningManager) GetSuppressedOperations() []string {
	return wm.config.SuppressedOperations
}
