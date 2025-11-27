package deprecation

import (
	"fmt"
	"strings"
)

// CLIDeprecation represents a CLI-specific deprecation (flags, commands, etc).
type CLIDeprecation struct {
	Type     CLIDeprecationType
	OldName  string
	NewName  string
	Command  string
	Reason   string
	Sunset   string
	Severity Severity
	AutoFix  bool // Whether auto-mapping is enabled
}

// CLIDeprecationType indicates what CLI element is deprecated.
type CLIDeprecationType string

const (
	CLIDeprecationTypeFlag     CLIDeprecationType = "flag"
	CLIDeprecationTypeCommand  CLIDeprecationType = "command"
	CLIDeprecationTypeBehavior CLIDeprecationType = "behavior"
	CLIDeprecationTypeAlias    CLIDeprecationType = "alias"
)

// FlagMapper handles automatic mapping of deprecated flags to new ones.
type FlagMapper struct {
	deprecations []*CLIDeprecation
	verbose      bool
}

// NewFlagMapper creates a new FlagMapper instance.
func NewFlagMapper() *FlagMapper {
	return &FlagMapper{
		deprecations: []*CLIDeprecation{},
		verbose:      true,
	}
}

// AddDeprecation adds a flag deprecation mapping.
func (fm *FlagMapper) AddDeprecation(dep *CLIDeprecation) {
	fm.deprecations = append(fm.deprecations, dep)
}

// MapFlag checks if a flag is deprecated and returns the new flag name.
// Returns: newFlag, wasDeprecated, warning message
func (fm *FlagMapper) MapFlag(command, flagName string) (string, bool, string) {
	for _, dep := range fm.deprecations {
		if dep.Type == CLIDeprecationTypeFlag &&
			dep.OldName == flagName &&
			(dep.Command == "" || dep.Command == command) {

			warning := fm.formatFlagWarning(dep)
			return dep.NewName, true, warning
		}
	}
	return flagName, false, ""
}

// formatFlagWarning formats a warning message for deprecated flag.
func (fm *FlagMapper) formatFlagWarning(dep *CLIDeprecation) string {
	msg := fmt.Sprintf("Flag '%s' is deprecated", dep.OldName)

	if dep.NewName != "" {
		msg += fmt.Sprintf(", using '%s' instead", dep.NewName)
	}

	if dep.Sunset != "" {
		msg += fmt.Sprintf(" (will be removed on %s)", dep.Sunset)
	}

	if dep.Reason != "" {
		msg += fmt.Sprintf("\n  Reason: %s", dep.Reason)
	}

	return msg
}

// MigrationAssistant helps users migrate from deprecated CLI usage.
type MigrationAssistant struct {
	deprecations []*CLIDeprecation
}

// NewMigrationAssistant creates a new MigrationAssistant.
func NewMigrationAssistant() *MigrationAssistant {
	return &MigrationAssistant{
		deprecations: []*CLIDeprecation{},
	}
}

// AddDeprecation adds a deprecation for migration assistance.
func (ma *MigrationAssistant) AddDeprecation(dep *CLIDeprecation) {
	ma.deprecations = append(ma.deprecations, dep)
}

// SuggestMigration suggests the new command syntax.
func (ma *MigrationAssistant) SuggestMigration(oldCommand string, args []string) *MigrationSuggestion {
	for _, dep := range ma.deprecations {
		if dep.Type == CLIDeprecationTypeCommand && dep.OldName == oldCommand {
			return &MigrationSuggestion{
				OldCommand: oldCommand,
				NewCommand: dep.NewName,
				Args:       args,
				Reason:     dep.Reason,
				AutoFix:    dep.AutoFix,
			}
		}
	}
	return nil
}

// MigrationSuggestion contains suggested migration.
type MigrationSuggestion struct {
	OldCommand string
	NewCommand string
	Args       []string
	Reason     string
	AutoFix    bool
}

// Format formats the migration suggestion.
func (ms *MigrationSuggestion) Format() string {
	var sb strings.Builder

	sb.WriteString("\nMigration suggestion:\n")
	sb.WriteString(fmt.Sprintf("  Old: %s %s\n", ms.OldCommand, strings.Join(ms.Args, " ")))
	sb.WriteString(fmt.Sprintf("  New: %s %s\n", ms.NewCommand, strings.Join(ms.Args, " ")))

	if ms.Reason != "" {
		sb.WriteString(fmt.Sprintf("\n  Reason: %s\n", ms.Reason))
	}

	if ms.AutoFix {
		sb.WriteString("\nRun it now? [y/N]: ")
	}

	return sb.String()
}

// GetNewCommandLine returns the new command line.
func (ms *MigrationSuggestion) GetNewCommandLine() string {
	return fmt.Sprintf("%s %s", ms.NewCommand, strings.Join(ms.Args, " "))
}

// ScriptScanner scans scripts for deprecated CLI usage.
type ScriptScanner struct {
	cliName      string
	deprecations []*CLIDeprecation
}

// NewScriptScanner creates a new ScriptScanner.
func NewScriptScanner(cliName string) *ScriptScanner {
	return &ScriptScanner{
		cliName:      cliName,
		deprecations: []*CLIDeprecation{},
	}
}

// AddDeprecation adds a deprecation pattern to scan for.
func (ss *ScriptScanner) AddDeprecation(dep *CLIDeprecation) {
	ss.deprecations = append(ss.deprecations, dep)
}

// ScanLine scans a single line for deprecated usage.
func (ss *ScriptScanner) ScanLine(line string, lineNum int) []*ScriptIssue {
	var issues []*ScriptIssue

	// Check if line contains CLI command
	if !strings.Contains(line, ss.cliName) {
		return issues
	}

	// Check for deprecated commands
	for _, dep := range ss.deprecations {
		if dep.Type == CLIDeprecationTypeCommand {
			if strings.Contains(line, dep.OldName) {
				issues = append(issues, &ScriptIssue{
					Line:        lineNum,
					Content:     line,
					Deprecation: dep,
					Suggestion:  strings.ReplaceAll(line, dep.OldName, dep.NewName),
				})
			}
		}
	}

	// Check for deprecated flags
	for _, dep := range ss.deprecations {
		if dep.Type == CLIDeprecationTypeFlag {
			if strings.Contains(line, dep.OldName) {
				issues = append(issues, &ScriptIssue{
					Line:        lineNum,
					Content:     line,
					Deprecation: dep,
					Suggestion:  strings.ReplaceAll(line, dep.OldName, dep.NewName),
				})
			}
		}
	}

	return issues
}

// ScriptIssue represents a deprecated usage found in a script.
type ScriptIssue struct {
	Line        int
	Content     string
	Deprecation *CLIDeprecation
	Suggestion  string
}

// Format formats the script issue for display.
func (si *ScriptIssue) Format() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("  Line %d:\n", si.Line))
	sb.WriteString(fmt.Sprintf("    Found: %s\n", strings.TrimSpace(si.Content)))
	sb.WriteString(fmt.Sprintf("    Issue: %s '%s' is deprecated\n", si.Deprecation.Type, si.Deprecation.OldName))

	if si.Suggestion != "" {
		sb.WriteString(fmt.Sprintf("    Fix:   %s\n", strings.TrimSpace(si.Suggestion)))
	}

	return sb.String()
}

// AutoFixer automatically fixes deprecated usage in scripts.
type AutoFixer struct {
	dryRun bool
}

// NewAutoFixer creates a new AutoFixer.
func NewAutoFixer(dryRun bool) *AutoFixer {
	return &AutoFixer{
		dryRun: dryRun,
	}
}

// FixLine fixes deprecated usage in a single line.
func (af *AutoFixer) FixLine(line string, issues []*ScriptIssue) string {
	fixed := line

	for _, issue := range issues {
		if issue.Deprecation.AutoFix && issue.Suggestion != "" {
			fixed = issue.Suggestion
		}
	}

	return fixed
}

// IsDryRun returns whether auto-fixer is in dry-run mode.
func (af *AutoFixer) IsDryRun() bool {
	return af.dryRun
}

// CLIDeprecationConfig represents CLI deprecation configuration.
type CLIDeprecationConfig struct {
	Deprecations []*CLIDeprecation `yaml:"cli_deprecations"`
}

// ParseCLIDeprecations parses CLI deprecation configuration from raw data.
// It accepts an array of deprecation entries and returns a slice of CLIDeprecation objects.
func ParseCLIDeprecations(data interface{}) ([]*CLIDeprecation, error) {
	depList, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("cli_deprecations must be an array")
	}

	var deprecations []*CLIDeprecation

	for _, depData := range depList {
		depMap, ok := depData.(map[string]interface{})
		if !ok {
			continue
		}

		dep := &CLIDeprecation{
			AutoFix: true, // Default to enabled
		}

		if depType, ok := depMap["type"].(string); ok {
			dep.Type = CLIDeprecationType(depType)
		}

		if oldName, ok := depMap["old_name"].(string); ok {
			dep.OldName = oldName
		}
		if oldFlag, ok := depMap["old_flag"].(string); ok {
			dep.OldName = oldFlag
		}
		if oldCmd, ok := depMap["old_command"].(string); ok {
			dep.OldName = oldCmd
		}

		if newName, ok := depMap["new_name"].(string); ok {
			dep.NewName = newName
		}
		if newFlag, ok := depMap["new_flag"].(string); ok {
			dep.NewName = newFlag
		}
		if newCmd, ok := depMap["new_command"].(string); ok {
			dep.NewName = newCmd
		}

		if command, ok := depMap["command"].(string); ok {
			dep.Command = command
		}

		if reason, ok := depMap["reason"].(string); ok {
			dep.Reason = reason
		}

		if sunset, ok := depMap["sunset"].(string); ok {
			dep.Sunset = sunset
		}

		if severity, ok := depMap["severity"].(string); ok {
			switch severity {
			case "info":
				dep.Severity = SeverityInfo
			case "warning":
				dep.Severity = SeverityWarning
			case "breaking":
				dep.Severity = SeverityBreaking
			}
		}

		if autoFix, ok := depMap["auto_fix"].(bool); ok {
			dep.AutoFix = autoFix
		}

		deprecations = append(deprecations, dep)
	}

	return deprecations, nil
}

// FormatCLIDeprecationList formats a list of CLI deprecations into a human-readable string.
// It groups deprecations by type and displays them with icons, reasons, and sunset dates.
func FormatCLIDeprecationList(deprecations []*CLIDeprecation) string {
	if len(deprecations) == 0 {
		return "No CLI deprecations"
	}

	var sb strings.Builder

	sb.WriteString("CLI Deprecations:\n")
	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	// Group by type
	byType := make(map[CLIDeprecationType][]*CLIDeprecation)
	for _, dep := range deprecations {
		byType[dep.Type] = append(byType[dep.Type], dep)
	}

	// Display each type
	for depType, deps := range byType {
		sb.WriteString(fmt.Sprintf("%s Deprecations (%d):\n\n", strings.Title(string(depType)), len(deps)))

		for _, dep := range deps {
			icon := getWarningIcon(WarningLevelWarning)
			sb.WriteString(fmt.Sprintf("  %s %s → %s\n", icon, dep.OldName, dep.NewName))

			if dep.Command != "" {
				sb.WriteString(fmt.Sprintf("     Command: %s\n", dep.Command))
			}
			if dep.Reason != "" {
				sb.WriteString(fmt.Sprintf("     Reason: %s\n", dep.Reason))
			}
			if dep.Sunset != "" {
				sb.WriteString(fmt.Sprintf("     Sunset: %s\n", dep.Sunset))
			}

			sb.WriteString("\n")
		}
	}

	sb.WriteString("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

	return sb.String()
}

// ValidateCLIDeprecation validates a CLI deprecation entry to ensure it has required fields.
// It returns an error if the deprecation type, old name, or new name (where applicable) is missing.
func ValidateCLIDeprecation(dep *CLIDeprecation) error {
	if dep.Type == "" {
		return fmt.Errorf("deprecation type is required")
	}

	if dep.OldName == "" {
		return fmt.Errorf("old name is required")
	}

	// For most types, new name is required
	if dep.Type != CLIDeprecationTypeBehavior && dep.NewName == "" {
		return fmt.Errorf("new name is required for type %s", dep.Type)
	}

	return nil
}
