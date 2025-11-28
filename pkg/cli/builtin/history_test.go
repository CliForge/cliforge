package builtin

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/state"
)

func TestNewHistoryCommand(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	output := &bytes.Buffer{}

	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	cmd := NewHistoryCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "history" {
		t.Errorf("expected Use 'history', got %q", cmd.Use)
	}

	// Check that subcommands are added
	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected subcommands to be added")
	}
}

func TestHistoryClear(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "command1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "command2", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	cmd := newHistoryClearCommand(opts)
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "History cleared") {
		t.Errorf("expected clear message, got: %s", result)
	}

	// Verify history was cleared
	recent := history.GetRecent(10)
	if len(recent) != 0 {
		t.Error("history was not cleared")
	}
}

func TestRunHistory_Empty(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	output := &bytes.Buffer{}

	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "", false, false, "", "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "No history entries found") {
		t.Errorf("expected 'No history entries found', got: %s", result)
	}
}

func TestRunHistory_Table(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "command1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "command2", ExitCode: 1, DurationMS: 200, Success: false, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "command3", ExitCode: 0, DurationMS: 150, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "", false, false, "", "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()

	// Check header
	if !strings.Contains(result, "COMMAND") {
		t.Error("expected header in output")
	}
	if !strings.Contains(result, "STATUS") {
		t.Error("expected STATUS column in output")
	}

	// Check commands
	if !strings.Contains(result, "command1") {
		t.Error("expected command1 in output")
	}
	if !strings.Contains(result, "command2") {
		t.Error("expected command2 in output")
	}
}

func TestRunHistory_JSON(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "test-command", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "", false, false, "", "json")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"command"`) {
		t.Error("expected JSON output with command field")
	}
	if !strings.Contains(result, "test-command") {
		t.Error("expected command in JSON output")
	}
}

func TestRunHistory_YAML(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "test-command", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "", false, false, "", "yaml")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "command: test-command") {
		t.Error("expected YAML output with command field")
	}
	if !strings.Contains(result, "success:") {
		t.Error("expected success field in YAML output")
	}
}

func TestRunHistory_WithLimit(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	for i := 1; i <= 50; i++ {
		_ = history.Add(&state.HistoryEntry{Command: "command", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	}

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 5, "", false, false, "", "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	lines := strings.Split(strings.TrimSpace(result), "\n")
	// Header (2 lines) + 5 entries = 7 lines
	if len(lines) > 10 {
		t.Errorf("expected limited output, got %d lines", len(lines))
	}
}

func TestRunHistory_SuccessOnly(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "success1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "failed1", ExitCode: 1, DurationMS: 100, Success: false, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "success2", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "", true, false, "", "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "success1") {
		t.Error("expected successful commands in output")
	}
	if strings.Contains(result, "failed1") {
		t.Error("did not expect failed commands in output")
	}
}

func TestRunHistory_FailedOnly(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "success1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "failed1", ExitCode: 1, DurationMS: 100, Success: false, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "failed2", ExitCode: 1, DurationMS: 100, Success: false, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "", false, true, "", "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "failed1") {
		t.Error("expected failed commands in output")
	}
	if strings.Contains(result, "success1") {
		t.Error("did not expect successful commands in output")
	}
}

func TestRunHistory_Search(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "git status", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "git commit", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "ls -la", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "git", false, false, "", "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "git status") {
		t.Error("expected git commands in search results")
	}
	if strings.Contains(result, "ls -la") {
		t.Error("did not expect non-matching commands in search results")
	}
}

func TestRunHistory_ContextFilter(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)

	// Create entries with different contexts
	entry1 := &state.HistoryEntry{
		ID:        1,
		Command:   "command1",
		Context:   "dev",
		Timestamp: time.Now(),
		Success:   true,
	}
	entry2 := &state.HistoryEntry{
		ID:        2,
		Command:   "command2",
		Context:   "prod",
		Timestamp: time.Now(),
		Success:   true,
	}

	_ = history.Add(entry1)
	_ = history.Add(entry2)

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistory(opts, 20, "", false, false, "dev", "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "command1") {
		t.Error("expected dev context commands in output")
	}
}

func TestRunHistoryStats_Text(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "cmd1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "cmd2", ExitCode: 1, DurationMS: 200, Success: false, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "cmd3", ExitCode: 0, DurationMS: 150, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistoryStats(opts, "text")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "History Statistics") {
		t.Error("expected statistics header")
	}
	if !strings.Contains(result, "Total commands:") {
		t.Error("expected total commands")
	}
	if !strings.Contains(result, "Successful:") {
		t.Error("expected successful count")
	}
	if !strings.Contains(result, "Failed:") {
		t.Error("expected failed count")
	}
}

func TestRunHistoryStats_JSON(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "cmd1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistoryStats(opts, "json")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"total_commands"`) {
		t.Error("expected JSON stats output")
	}
}

func TestRunHistoryStats_YAML(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "cmd1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	err := runHistoryStats(opts, "yaml")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "total_commands:") {
		t.Error("expected YAML stats output")
	}
}

func TestFormatHistoryTable(t *testing.T) {
	entries := []*state.HistoryEntry{
		{
			ID:         1,
			Command:    "test-command",
			Timestamp:  time.Now(),
			Success:    true,
			DurationMS: 150,
		},
		{
			ID:         2,
			Command:    "another-command",
			Timestamp:  time.Now(),
			Success:    false,
			DurationMS: 250,
		},
	}

	output := &bytes.Buffer{}
	err := formatHistoryTable(entries, output)
	if err != nil {
		t.Fatalf("formatHistoryTable failed: %v", err)
	}

	result := output.String()

	// Check header
	if !strings.Contains(result, "COMMAND") {
		t.Error("expected COMMAND header")
	}
	if !strings.Contains(result, "STATUS") {
		t.Error("expected STATUS header")
	}

	// Check entries
	if !strings.Contains(result, "test-command") {
		t.Error("expected first command in output")
	}
	if !strings.Contains(result, "another-command") {
		t.Error("expected second command in output")
	}
}

func TestFormatHistoryJSON(t *testing.T) {
	entries := []*state.HistoryEntry{
		{
			ID:        1,
			Command:   "test",
			Timestamp: time.Now(),
			Success:   true,
		},
	}

	output := &bytes.Buffer{}
	err := formatHistoryJSON(entries, output)
	if err != nil {
		t.Fatalf("formatHistoryJSON failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"command"`) {
		t.Error("expected command field in JSON")
	}
}

func TestFormatHistoryYAML(t *testing.T) {
	entries := []*state.HistoryEntry{
		{
			ID:         1,
			Command:    "test",
			Timestamp:  time.Now(),
			Success:    true,
			DurationMS: 100,
			Context:    "dev",
			User:       "testuser",
		},
	}

	output := &bytes.Buffer{}
	err := formatHistoryYAML(entries, output)
	if err != nil {
		t.Fatalf("formatHistoryYAML failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "command: test") {
		t.Error("expected command field in YAML")
	}
	if !strings.Contains(result, "success: true") {
		t.Error("expected success field in YAML")
	}
	if !strings.Contains(result, "duration_ms: 100") {
		t.Error("expected duration_ms field in YAML")
	}
	if !strings.Contains(result, "context: dev") {
		t.Error("expected context field in YAML")
	}
}

func TestFormatStatsText(t *testing.T) {
	stats := &state.HistoryStats{
		TotalCommands:      100,
		SuccessfulCommands: 90,
		FailedCommands:     10,
		AverageDurationMS:  150,
		FirstCommand:       time.Now().Add(-24 * time.Hour),
		LastCommand:        time.Now(),
	}

	output := &bytes.Buffer{}
	err := formatStatsText(stats, output)
	if err != nil {
		t.Fatalf("formatStatsText failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Total commands: 100") {
		t.Error("expected total commands")
	}
	if !strings.Contains(result, "Successful: 90") {
		t.Error("expected successful commands")
	}
	if !strings.Contains(result, "Success rate: 90.0%") {
		t.Error("expected success rate")
	}
}

func TestFormatStatsYAML(t *testing.T) {
	stats := &state.HistoryStats{
		TotalCommands:      50,
		SuccessfulCommands: 45,
		FailedCommands:     5,
		AverageDurationMS:  200,
	}

	output := &bytes.Buffer{}
	err := formatStatsYAML(stats, output)
	if err != nil {
		t.Fatalf("formatStatsYAML failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "total_commands: 50") {
		t.Error("expected total_commands in YAML")
	}
	if !strings.Contains(result, "successful_commands: 45") {
		t.Error("expected successful_commands in YAML")
	}
}

func TestFormatMilliseconds(t *testing.T) {
	tests := []struct {
		name     string
		ms       int64
		expected string
	}{
		{"under 1 second", 500, "500ms"},
		{"1 second", 1000, "1.0s"},
		{"under 1 minute", 30000, "30.0s"},
		{"1 minute", 60000, "1.0m"},
		{"over 1 minute", 90000, "1.5m"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMilliseconds(tt.ms)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHistoryStatsCommand(t *testing.T) {
	history, _ := state.NewHistory("testcli", 100)
	_ = history.Add(&state.HistoryEntry{Command: "cmd1", ExitCode: 0, DurationMS: 100, Success: true, Timestamp: time.Now()})
	_ = history.Add(&state.HistoryEntry{Command: "cmd2", ExitCode: 0, DurationMS: 150, Success: true, Timestamp: time.Now()})

	output := &bytes.Buffer{}
	opts := &HistoryOptions{
		History: history,
		Output:  output,
	}

	cmd := newHistoryStatsCommand(opts)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("stats command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Total commands:") {
		t.Error("expected statistics in output")
	}
}
