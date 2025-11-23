package state

import (
	"os"
	"testing"
	"time"
)

func TestNewHistory(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, err := NewHistory("testcli", 100)
	if err != nil {
		t.Fatalf("Failed to create history: %v", err)
	}

	if h.cliName != "testcli" {
		t.Errorf("Expected cliName to be 'testcli', got %s", h.cliName)
	}

	if h.maxEntries != 100 {
		t.Errorf("Expected maxEntries to be 100, got %d", h.maxEntries)
	}
}

func TestHistoryAdd(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	entry := &HistoryEntry{
		Command:  "mycli describe cluster",
		ExitCode: 0,
	}

	if err := h.Add(entry); err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	if entry.ID != 1 {
		t.Errorf("Expected ID to be 1, got %d", entry.ID)
	}

	if entry.Timestamp.IsZero() {
		t.Error("Expected timestamp to be set")
	}

	if !entry.Success {
		t.Error("Expected entry to be marked as successful")
	}

	// Add another
	entry2 := &HistoryEntry{
		Command:  "mycli create cluster",
		ExitCode: 1,
	}

	h.Add(entry2)

	if entry2.ID != 2 {
		t.Errorf("Expected ID to be 2, got %d", entry2.ID)
	}

	if entry2.Success {
		t.Error("Expected failed command to be marked as unsuccessful")
	}
}

func TestHistoryGet(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	entry := &HistoryEntry{
		Command:  "test command",
		ExitCode: 0,
	}
	h.Add(entry)

	retrieved, err := h.Get(1)
	if err != nil {
		t.Fatalf("Failed to get entry: %v", err)
	}

	if retrieved.Command != "test command" {
		t.Errorf("Expected command to be 'test command', got %s", retrieved.Command)
	}

	_, err = h.Get(999)
	if err == nil {
		t.Error("Expected error when getting non-existent entry")
	}
}

func TestHistoryGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "cmd1", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "cmd2", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "cmd3", ExitCode: 0})

	entries := h.GetAll()
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Verify it's a copy
	entries[0].Command = "modified"
	original, _ := h.Get(1)
	if original.Command == "modified" {
		t.Error("Expected GetAll to return a copy, not original")
	}
}

func TestHistoryGetRecent(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	for i := 1; i <= 5; i++ {
		h.Add(&HistoryEntry{Command: "cmd" + string(rune(i+'0')), ExitCode: 0})
	}

	recent := h.GetRecent(3)
	if len(recent) != 3 {
		t.Errorf("Expected 3 recent entries, got %d", len(recent))
	}

	// Should be last 3 entries
	if recent[0].Command != "cmd3" {
		t.Errorf("Expected first recent to be 'cmd3', got %s", recent[0].Command)
	}

	if recent[2].Command != "cmd5" {
		t.Errorf("Expected last recent to be 'cmd5', got %s", recent[2].Command)
	}
}

func TestHistorySearch(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "mycli create cluster", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "mycli delete cluster", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "mycli list regions", ExitCode: 0})

	results := h.Search("cluster")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'cluster', got %d", len(results))
	}

	results = h.Search("delete")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'delete', got %d", len(results))
	}

	// Case insensitive
	results = h.Search("CLUSTER")
	if len(results) != 2 {
		t.Error("Expected search to be case insensitive")
	}
}

func TestHistoryFilter(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "cmd1", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "cmd2", ExitCode: 1})
	h.Add(&HistoryEntry{Command: "cmd3", ExitCode: 0})

	// Filter successful commands
	successful := h.Filter(func(e *HistoryEntry) bool {
		return e.Success
	})

	if len(successful) != 2 {
		t.Errorf("Expected 2 successful commands, got %d", len(successful))
	}
}

func TestHistoryGetSuccessfulFailed(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "cmd1", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "cmd2", ExitCode: 1})
	h.Add(&HistoryEntry{Command: "cmd3", ExitCode: 0})

	successful := h.GetSuccessful()
	if len(successful) != 2 {
		t.Errorf("Expected 2 successful commands, got %d", len(successful))
	}

	failed := h.GetFailed()
	if len(failed) != 1 {
		t.Errorf("Expected 1 failed command, got %d", len(failed))
	}
}

func TestHistoryGetByContext(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "cmd1", Context: "production", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "cmd2", Context: "staging", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "cmd3", Context: "production", ExitCode: 0})

	prodEntries := h.GetByContext("production")
	if len(prodEntries) != 2 {
		t.Errorf("Expected 2 production entries, got %d", len(prodEntries))
	}
}

func TestHistoryGetSince(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	now := time.Now()
	past := now.Add(-1 * time.Hour)

	h.Add(&HistoryEntry{Command: "old", ExitCode: 0, Timestamp: past})
	time.Sleep(10 * time.Millisecond)
	h.Add(&HistoryEntry{Command: "new", ExitCode: 0, Timestamp: now})

	since := now.Add(-30 * time.Minute)
	recent := h.GetSince(since)

	if len(recent) != 1 {
		t.Errorf("Expected 1 recent entry, got %d", len(recent))
	}

	if recent[0].Command != "new" {
		t.Error("Expected to get the newer entry")
	}
}

func TestHistoryClear(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "cmd1", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "cmd2", ExitCode: 0})

	if h.Count() != 2 {
		t.Error("Expected 2 entries before clear")
	}

	h.Clear()

	if h.Count() != 0 {
		t.Error("Expected 0 entries after clear")
	}
}

func TestHistoryMaxEntries(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	// Use unique CLI name for this test
	h, _ := NewHistory("testcli-maxentries", 5)

	// Add more than max
	for i := 1; i <= 10; i++ {
		h.Add(&HistoryEntry{Command: "cmd" + string(rune(i+'0')), ExitCode: 0})
	}

	if h.Count() != 5 {
		t.Errorf("Expected count to be limited to 5, got %d", h.Count())
	}

	// Oldest entries should be removed
	entries := h.GetAll()
	if entries[0].Command != "cmd6" {
		t.Errorf("Expected oldest remaining entry to be 'cmd6', got %s", entries[0].Command)
	}
}

func TestHistorySetMaxEntries(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	// Add 10 entries
	for i := 1; i <= 10; i++ {
		h.Add(&HistoryEntry{Command: "cmd", ExitCode: 0})
	}

	// Reduce max
	h.SetMaxEntries(5)

	if h.Count() != 5 {
		t.Errorf("Expected count to be 5 after reducing max, got %d", h.Count())
	}
}

func TestHistorySaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	// Use unique CLI name for this test
	uniqueCLI := "testcli-saveload"

	// Create and save
	h1, _ := NewHistory(uniqueCLI, 100)
	h1.Clear() // Clear any previously loaded data
	h1.Add(&HistoryEntry{Command: "cmd1", ExitCode: 0})
	h1.Add(&HistoryEntry{Command: "cmd2", ExitCode: 1})

	if err := h1.Save(); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Load in new instance
	h2, err := NewHistory(uniqueCLI, 100)
	if err != nil {
		t.Fatalf("Failed to create new history: %v", err)
	}

	if h2.Count() != 2 {
		t.Errorf("Expected 2 entries after load, got %d", h2.Count())
	}

	entry, _ := h2.Get(1)
	if entry.Command != "cmd1" {
		t.Error("Expected loaded entry to match saved entry")
	}
}

func TestHistoryRecordCommand(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	duration := 500 * time.Millisecond
	err := h.RecordCommand("mycli test", 0, duration, "production")
	if err != nil {
		t.Fatalf("Failed to record command: %v", err)
	}

	entries := h.GetAll()
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Command != "mycli test" {
		t.Errorf("Expected command to be 'mycli test', got %s", entry.Command)
	}

	if entry.DurationMS != 500 {
		t.Errorf("Expected duration to be 500ms, got %d", entry.DurationMS)
	}

	if entry.Context != "production" {
		t.Errorf("Expected context to be 'production', got %s", entry.Context)
	}

	if entry.User == "" {
		t.Error("Expected user to be set")
	}
}

func TestHistoryGetStats(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "cmd1", ExitCode: 0, DurationMS: 100})
	h.Add(&HistoryEntry{Command: "cmd2", ExitCode: 1, DurationMS: 200})
	h.Add(&HistoryEntry{Command: "cmd3", ExitCode: 0, DurationMS: 300})

	stats := h.GetStats()

	if stats.TotalCommands != 3 {
		t.Errorf("Expected total commands to be 3, got %d", stats.TotalCommands)
	}

	if stats.SuccessfulCommands != 2 {
		t.Errorf("Expected 2 successful commands, got %d", stats.SuccessfulCommands)
	}

	if stats.FailedCommands != 1 {
		t.Errorf("Expected 1 failed command, got %d", stats.FailedCommands)
	}

	expectedAvg := int64(200) // (100 + 200 + 300) / 3
	if stats.AverageDurationMS != expectedAvg {
		t.Errorf("Expected average duration to be %d, got %d", expectedAvg, stats.AverageDurationMS)
	}
}

func TestHistoryGetMostUsedCommands(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	h, _ := NewHistory("testcli", 100)
	h.Clear() // Ensure clean state for test

	h.Add(&HistoryEntry{Command: "mycli describe cluster", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "mycli list clusters", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "mycli describe cluster", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "mycli list clusters", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "mycli describe cluster", ExitCode: 0})
	h.Add(&HistoryEntry{Command: "mycli create cluster", ExitCode: 0})

	freqs := h.GetMostUsedCommands(0)

	if len(freqs) != 1 {
		t.Errorf("Expected 1 unique base command, got %d", len(freqs))
	}

	if freqs[0].Command != "mycli" {
		t.Errorf("Expected most used to be 'mycli', got %s", freqs[0].Command)
	}

	if freqs[0].Count != 6 {
		t.Errorf("Expected count to be 6, got %d", freqs[0].Count)
	}
}
