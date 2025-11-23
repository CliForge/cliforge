package deprecation

import (
	"testing"
)

func TestFlagMapper_MapFlag(t *testing.T) {
	mapper := NewFlagMapper()

	// Add test deprecations
	mapper.AddDeprecation(&CLIDeprecation{
		Type:    CLIDeprecationTypeFlag,
		OldName: "--filter",
		NewName: "--search",
		Command: "users list",
		Sunset:  "2025-12-31",
	})

	mapper.AddDeprecation(&CLIDeprecation{
		Type:    CLIDeprecationTypeFlag,
		OldName: "--old-flag",
		NewName: "--new-flag",
		Command: "",
	})

	tests := []struct {
		name          string
		command       string
		flagName      string
		wantNewFlag   string
		wantMapped    bool
		wantWarning   bool
	}{
		{
			name:        "map deprecated flag for specific command",
			command:     "users list",
			flagName:    "--filter",
			wantNewFlag: "--search",
			wantMapped:  true,
			wantWarning: true,
		},
		{
			name:        "map deprecated flag for any command",
			command:     "posts list",
			flagName:    "--old-flag",
			wantNewFlag: "--new-flag",
			wantMapped:  true,
			wantWarning: true,
		},
		{
			name:        "non-deprecated flag unchanged",
			command:     "users list",
			flagName:    "--limit",
			wantNewFlag: "--limit",
			wantMapped:  false,
			wantWarning: false,
		},
		{
			name:        "deprecated flag wrong command",
			command:     "posts list",
			flagName:    "--filter",
			wantNewFlag: "--filter",
			wantMapped:  false,
			wantWarning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newFlag, mapped, warning := mapper.MapFlag(tt.command, tt.flagName)

			if newFlag != tt.wantNewFlag {
				t.Errorf("MapFlag() newFlag = %v, want %v", newFlag, tt.wantNewFlag)
			}
			if mapped != tt.wantMapped {
				t.Errorf("MapFlag() mapped = %v, want %v", mapped, tt.wantMapped)
			}
			if tt.wantWarning && warning == "" {
				t.Errorf("MapFlag() should return warning")
			}
			if !tt.wantWarning && warning != "" {
				t.Errorf("MapFlag() should not return warning, got: %v", warning)
			}
		})
	}
}

func TestMigrationAssistant_SuggestMigration(t *testing.T) {
	assistant := NewMigrationAssistant()

	assistant.AddDeprecation(&CLIDeprecation{
		Type:    CLIDeprecationTypeCommand,
		OldName: "users ls",
		NewName: "users list",
		Reason:  "Standardizing command names",
	})

	tests := []struct {
		name       string
		oldCommand string
		args       []string
		wantNil    bool
	}{
		{
			name:       "deprecated command found",
			oldCommand: "users ls",
			args:       []string{"--limit", "10"},
			wantNil:    false,
		},
		{
			name:       "non-deprecated command",
			oldCommand: "users list",
			args:       []string{"--limit", "10"},
			wantNil:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := assistant.SuggestMigration(tt.oldCommand, tt.args)

			if tt.wantNil && suggestion != nil {
				t.Errorf("SuggestMigration() should return nil, got %v", suggestion)
			}
			if !tt.wantNil && suggestion == nil {
				t.Errorf("SuggestMigration() should return suggestion")
			}

			if suggestion != nil {
				if suggestion.OldCommand != tt.oldCommand {
					t.Errorf("OldCommand = %v, want %v", suggestion.OldCommand, tt.oldCommand)
				}
				if len(suggestion.Args) != len(tt.args) {
					t.Errorf("Args length = %v, want %v", len(suggestion.Args), len(tt.args))
				}
			}
		})
	}
}

func TestMigrationSuggestion_Format(t *testing.T) {
	suggestion := &MigrationSuggestion{
		OldCommand: "users ls",
		NewCommand: "users list",
		Args:       []string{"--limit", "10"},
		Reason:     "Standardizing command names",
		AutoFix:    true,
	}

	result := suggestion.Format()

	if result == "" {
		t.Error("Format() returned empty string")
	}

	// Check for key components
	expectedStrings := []string{
		"Migration suggestion",
		suggestion.OldCommand,
		suggestion.NewCommand,
		suggestion.Reason,
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("Format() missing expected string: %s", expected)
		}
	}
}

func TestMigrationSuggestion_GetNewCommandLine(t *testing.T) {
	suggestion := &MigrationSuggestion{
		NewCommand: "users list",
		Args:       []string{"--limit", "10", "--format", "json"},
	}

	result := suggestion.GetNewCommandLine()

	expected := "users list --limit 10 --format json"
	if result != expected {
		t.Errorf("GetNewCommandLine() = %v, want %v", result, expected)
	}
}

func TestScriptScanner_ScanLine(t *testing.T) {
	scanner := NewScriptScanner("mycli")

	scanner.AddDeprecation(&CLIDeprecation{
		Type:    CLIDeprecationTypeCommand,
		OldName: "users ls",
		NewName: "users list",
	})

	scanner.AddDeprecation(&CLIDeprecation{
		Type:    CLIDeprecationTypeFlag,
		OldName: "--filter",
		NewName: "--search",
	})

	tests := []struct {
		name       string
		line       string
		lineNum    int
		wantIssues int
	}{
		{
			name:       "deprecated command found",
			line:       "mycli users ls --limit 10",
			lineNum:    1,
			wantIssues: 1,
		},
		{
			name:       "deprecated flag found",
			line:       "mycli users list --filter 'name=john'",
			lineNum:    2,
			wantIssues: 1,
		},
		{
			name:       "multiple deprecations",
			line:       "mycli users ls --filter 'name=john'",
			lineNum:    3,
			wantIssues: 2,
		},
		{
			name:       "no deprecations",
			line:       "mycli users list --limit 10",
			lineNum:    4,
			wantIssues: 0,
		},
		{
			name:       "non-CLI line",
			line:       "echo 'hello world'",
			lineNum:    5,
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := scanner.ScanLine(tt.line, tt.lineNum)

			if len(issues) != tt.wantIssues {
				t.Errorf("ScanLine() found %d issues, want %d", len(issues), tt.wantIssues)
			}

			for _, issue := range issues {
				if issue.Line != tt.lineNum {
					t.Errorf("Issue line = %d, want %d", issue.Line, tt.lineNum)
				}
				if issue.Suggestion == "" {
					t.Error("Issue should have suggestion")
				}
			}
		})
	}
}

func TestScriptIssue_Format(t *testing.T) {
	issue := &ScriptIssue{
		Line:    42,
		Content: "mycli users ls --filter 'name=john'",
		Deprecation: &CLIDeprecation{
			Type:    CLIDeprecationTypeCommand,
			OldName: "users ls",
			NewName: "users list",
		},
		Suggestion: "mycli users list --filter 'name=john'",
	}

	result := issue.Format()

	if result == "" {
		t.Error("Format() returned empty string")
	}

	// Check for key components
	expectedStrings := []string{
		"Line 42",
		"deprecated",
		issue.Suggestion,
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("Format() missing expected string: %s", expected)
		}
	}
}

func TestAutoFixer_FixLine(t *testing.T) {
	fixer := NewAutoFixer(false)

	tests := []struct {
		name    string
		line    string
		issues  []*ScriptIssue
		want    string
	}{
		{
			name: "fix with auto-fix enabled",
			line: "mycli users ls",
			issues: []*ScriptIssue{
				{
					Deprecation: &CLIDeprecation{AutoFix: true},
					Suggestion:  "mycli users list",
				},
			},
			want: "mycli users list",
		},
		{
			name: "no fix when auto-fix disabled",
			line: "mycli users ls",
			issues: []*ScriptIssue{
				{
					Deprecation: &CLIDeprecation{AutoFix: false},
					Suggestion:  "mycli users list",
				},
			},
			want: "mycli users ls",
		},
		{
			name:   "no issues",
			line:   "mycli users list",
			issues: []*ScriptIssue{},
			want:   "mycli users list",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fixer.FixLine(tt.line, tt.issues)
			if got != tt.want {
				t.Errorf("FixLine() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCLIDeprecations(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		want    int
		wantErr bool
	}{
		{
			name: "valid deprecations",
			data: []interface{}{
				map[string]interface{}{
					"type":        "flag",
					"old_flag":    "--filter",
					"new_flag":    "--search",
					"sunset":      "2025-12-31",
				},
				map[string]interface{}{
					"type":        "command",
					"old_command": "users ls",
					"new_command": "users list",
				},
			},
			want:    2,
			wantErr: false,
		},
		{
			name:    "invalid data type",
			data:    "not an array",
			want:    0,
			wantErr: true,
		},
		{
			name: "empty array",
			data: []interface{}{},
			want: 0,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCLIDeprecations(tt.data)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCLIDeprecations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != tt.want {
				t.Errorf("ParseCLIDeprecations() returned %d items, want %d", len(got), tt.want)
			}

			// Validate parsed deprecations
			for _, dep := range got {
				if dep.Type == "" {
					t.Error("Deprecation type should not be empty")
				}
				if dep.OldName == "" {
					t.Error("Old name should not be empty")
				}
			}
		})
	}
}

func TestFormatCLIDeprecationList(t *testing.T) {
	deprecations := []*CLIDeprecation{
		{
			Type:     CLIDeprecationTypeFlag,
			OldName:  "--filter",
			NewName:  "--search",
			Command:  "users list",
			Reason:   "Aligning with API v2",
			Sunset:   "2025-12-31",
			Severity: SeverityWarning,
		},
		{
			Type:     CLIDeprecationTypeCommand,
			OldName:  "users ls",
			NewName:  "users list",
			Reason:   "Standardizing commands",
			Sunset:   "2026-01-31",
			Severity: SeverityInfo,
		},
	}

	result := FormatCLIDeprecationList(deprecations)

	if result == "" {
		t.Error("FormatCLIDeprecationList() returned empty string")
	}

	// Check for key components
	expectedStrings := []string{
		"CLI Deprecations",
		"--filter",
		"--search",
		"users ls",
		"users list",
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("FormatCLIDeprecationList() missing expected string: %s", expected)
		}
	}
}

func TestFormatCLIDeprecationList_Empty(t *testing.T) {
	result := FormatCLIDeprecationList([]*CLIDeprecation{})

	expected := "No CLI deprecations"
	if result != expected {
		t.Errorf("FormatCLIDeprecationList() = %v, want %v", result, expected)
	}
}

func TestValidateCLIDeprecation(t *testing.T) {
	tests := []struct {
		name    string
		dep     *CLIDeprecation
		wantErr bool
	}{
		{
			name: "valid flag deprecation",
			dep: &CLIDeprecation{
				Type:    CLIDeprecationTypeFlag,
				OldName: "--filter",
				NewName: "--search",
			},
			wantErr: false,
		},
		{
			name: "valid behavior deprecation (no new name)",
			dep: &CLIDeprecation{
				Type:    CLIDeprecationTypeBehavior,
				OldName: "auto-confirm",
			},
			wantErr: false,
		},
		{
			name: "missing type",
			dep: &CLIDeprecation{
				OldName: "--filter",
				NewName: "--search",
			},
			wantErr: true,
		},
		{
			name: "missing old name",
			dep: &CLIDeprecation{
				Type:    CLIDeprecationTypeFlag,
				NewName: "--search",
			},
			wantErr: true,
		},
		{
			name: "flag missing new name",
			dep: &CLIDeprecation{
				Type:    CLIDeprecationTypeFlag,
				OldName: "--filter",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCLIDeprecation(tt.dep)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCLIDeprecation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
