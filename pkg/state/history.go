package state

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
)

// History manages command history storage and retrieval.
type History struct {
	cliName     string
	historyPath string
	entries     []*HistoryEntry
	maxEntries  int
	mu          sync.RWMutex
}

// HistoryEntry represents a single command history entry.
type HistoryEntry struct {
	ID         int       `json:"id"`
	Command    string    `json:"command"`
	Timestamp  time.Time `json:"timestamp"`
	ExitCode   int       `json:"exit_code"`
	DurationMS int64     `json:"duration_ms,omitempty"`
	User       string    `json:"user,omitempty"`
	Context    string    `json:"context,omitempty"`
	WorkingDir string    `json:"working_dir,omitempty"`
	Success    bool      `json:"success"`
}

// HistoryData represents the structure of the history file.
type HistoryData struct {
	History    []*HistoryEntry `json:"history"`
	MaxEntries int             `json:"max_entries"`
	Version    string          `json:"version,omitempty"`
}

const (
	// DefaultMaxHistoryEntries is the default maximum number of history entries.
	DefaultMaxHistoryEntries = 1000

	// HistoryVersion is the current history file format version.
	HistoryVersion = "1.0"
)

// NewHistory creates a new history manager.
func NewHistory(cliName string, maxEntries int) (*History, error) {
	if maxEntries <= 0 {
		maxEntries = DefaultMaxHistoryEntries
	}

	historyPath, err := getHistoryPath(cliName)
	if err != nil {
		return nil, fmt.Errorf("failed to get history path: %w", err)
	}

	h := &History{
		cliName:     cliName,
		historyPath: historyPath,
		entries:     make([]*HistoryEntry, 0),
		maxEntries:  maxEntries,
	}

	// Try to load existing history
	if err := h.Load(); err != nil {
		// If file doesn't exist, that's okay
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load history: %w", err)
		}
	}

	return h, nil
}

// Load loads history from disk.
func (h *History) Load() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	data, err := os.ReadFile(h.historyPath)
	if err != nil {
		return err
	}

	var historyData HistoryData
	if err := json.Unmarshal(data, &historyData); err != nil {
		return fmt.Errorf("failed to parse history file: %w", err)
	}

	h.entries = historyData.History
	if historyData.MaxEntries > 0 {
		h.maxEntries = historyData.MaxEntries
	}

	// Reassign IDs to ensure they're sequential
	for i, entry := range h.entries {
		entry.ID = i + 1
	}

	return nil
}

// Save saves history to disk.
func (h *History) Save() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(h.historyPath), 0700); err != nil {
		return fmt.Errorf("failed to create history directory: %w", err)
	}

	historyData := HistoryData{
		History:    h.entries,
		MaxEntries: h.maxEntries,
		Version:    HistoryVersion,
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(historyData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}

	// Write to file with atomic rename
	tmpPath := h.historyPath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}

	if err := os.Rename(tmpPath, h.historyPath); err != nil {
		os.Remove(tmpPath) // Clean up on error
		return fmt.Errorf("failed to save history file: %w", err)
	}

	return nil
}

// Add adds a new entry to the history.
func (h *History) Add(entry *HistoryEntry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Set ID
	if len(h.entries) > 0 {
		entry.ID = h.entries[len(h.entries)-1].ID + 1
	} else {
		entry.ID = 1
	}

	// Set timestamp if not already set
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}

	// Set success based on exit code if not already set
	entry.Success = entry.ExitCode == 0

	// Add to entries
	h.entries = append(h.entries, entry)

	// Trim if exceeds max entries
	if len(h.entries) > h.maxEntries {
		// Remove oldest entries
		h.entries = h.entries[len(h.entries)-h.maxEntries:]
		// Reassign IDs
		for i, e := range h.entries {
			e.ID = i + 1
		}
	}

	return nil
}

// Get returns a history entry by ID.
func (h *History) Get(id int) (*HistoryEntry, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, entry := range h.entries {
		if entry.ID == id {
			return entry, nil
		}
	}

	return nil, fmt.Errorf("history entry %d not found", id)
}

// GetAll returns all history entries.
func (h *History) GetAll() []*HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Return a deep copy to prevent external modifications
	entries := make([]*HistoryEntry, len(h.entries))
	for i, entry := range h.entries {
		entryCopy := *entry
		entries[i] = &entryCopy
	}
	return entries
}

// GetRecent returns the most recent N entries.
func (h *History) GetRecent(n int) []*HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if n <= 0 || n > len(h.entries) {
		n = len(h.entries)
	}

	// Return last N entries
	start := len(h.entries) - n
	entries := make([]*HistoryEntry, n)
	copy(entries, h.entries[start:])
	return entries
}

// Search searches history entries by pattern.
func (h *History) Search(pattern string) []*HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	pattern = strings.ToLower(pattern)
	matches := make([]*HistoryEntry, 0)

	for _, entry := range h.entries {
		if strings.Contains(strings.ToLower(entry.Command), pattern) {
			matches = append(matches, entry)
		}
	}

	return matches
}

// Filter filters history entries by criteria.
func (h *History) Filter(fn func(*HistoryEntry) bool) []*HistoryEntry {
	h.mu.RLock()
	defer h.mu.RUnlock()

	matches := make([]*HistoryEntry, 0)

	for _, entry := range h.entries {
		if fn(entry) {
			matches = append(matches, entry)
		}
	}

	return matches
}

// GetByContext returns entries for a specific context.
func (h *History) GetByContext(context string) []*HistoryEntry {
	return h.Filter(func(e *HistoryEntry) bool {
		return e.Context == context
	})
}

// GetSuccessful returns only successful commands.
func (h *History) GetSuccessful() []*HistoryEntry {
	return h.Filter(func(e *HistoryEntry) bool {
		return e.Success
	})
}

// GetFailed returns only failed commands.
func (h *History) GetFailed() []*HistoryEntry {
	return h.Filter(func(e *HistoryEntry) bool {
		return !e.Success
	})
}

// GetSince returns entries since a specific time.
func (h *History) GetSince(since time.Time) []*HistoryEntry {
	return h.Filter(func(e *HistoryEntry) bool {
		return e.Timestamp.After(since)
	})
}

// Clear clears all history entries.
func (h *History) Clear() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.entries = make([]*HistoryEntry, 0)
	return nil
}

// Count returns the number of history entries.
func (h *History) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.entries)
}

// GetPath returns the path to the history file.
func (h *History) GetPath() string {
	return h.historyPath
}

// SetMaxEntries sets the maximum number of entries.
func (h *History) SetMaxEntries(max int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.maxEntries = max

	// Trim if needed
	if len(h.entries) > h.maxEntries {
		h.entries = h.entries[len(h.entries)-h.maxEntries:]
		// Reassign IDs
		for i, e := range h.entries {
			e.ID = i + 1
		}
	}
}

// GetMaxEntries returns the maximum number of entries.
func (h *History) GetMaxEntries() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.maxEntries
}

// GetStats returns statistics about the history.
func (h *History) GetStats() HistoryStats {
	h.mu.RLock()
	defer h.mu.RUnlock()

	stats := HistoryStats{
		TotalCommands: len(h.entries),
	}

	if len(h.entries) == 0 {
		return stats
	}

	var totalDuration int64
	for _, entry := range h.entries {
		if entry.Success {
			stats.SuccessfulCommands++
		} else {
			stats.FailedCommands++
		}
		totalDuration += entry.DurationMS

		if stats.FirstCommand.IsZero() || entry.Timestamp.Before(stats.FirstCommand) {
			stats.FirstCommand = entry.Timestamp
		}
		if stats.LastCommand.IsZero() || entry.Timestamp.After(stats.LastCommand) {
			stats.LastCommand = entry.Timestamp
		}
	}

	if stats.TotalCommands > 0 {
		stats.AverageDurationMS = totalDuration / int64(stats.TotalCommands)
	}

	return stats
}

// HistoryStats represents statistics about command history.
type HistoryStats struct {
	TotalCommands      int       `json:"total_commands"`
	SuccessfulCommands int       `json:"successful_commands"`
	FailedCommands     int       `json:"failed_commands"`
	AverageDurationMS  int64     `json:"average_duration_ms"`
	FirstCommand       time.Time `json:"first_command"`
	LastCommand        time.Time `json:"last_command"`
}

// getHistoryPath returns the path to the history file.
func getHistoryPath(cliName string) (string, error) {
	// Use XDG state directory
	stateDir := filepath.Join(xdg.StateHome, cliName)
	return filepath.Join(stateDir, "history.json"), nil
}

// RecordCommand is a helper function to record a command execution.
func (h *History) RecordCommand(command string, exitCode int, duration time.Duration, context string) error {
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}

	workingDir, _ := os.Getwd()

	entry := &HistoryEntry{
		Command:    command,
		ExitCode:   exitCode,
		DurationMS: duration.Milliseconds(),
		User:       username,
		Context:    context,
		WorkingDir: workingDir,
	}

	if err := h.Add(entry); err != nil {
		return err
	}

	return h.Save()
}

// GetMostUsedCommands returns the most frequently used command patterns.
func (h *History) GetMostUsedCommands(limit int) []CommandFrequency {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// Count command frequencies
	freqMap := make(map[string]int)
	for _, entry := range h.entries {
		// Extract base command (first word)
		parts := strings.Fields(entry.Command)
		if len(parts) > 0 {
			baseCmd := parts[0]
			freqMap[baseCmd]++
		}
	}

	// Convert to slice and sort
	frequencies := make([]CommandFrequency, 0, len(freqMap))
	for cmd, count := range freqMap {
		frequencies = append(frequencies, CommandFrequency{
			Command: cmd,
			Count:   count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(frequencies)-1; i++ {
		for j := i + 1; j < len(frequencies); j++ {
			if frequencies[j].Count > frequencies[i].Count {
				frequencies[i], frequencies[j] = frequencies[j], frequencies[i]
			}
		}
	}

	// Limit results
	if limit > 0 && limit < len(frequencies) {
		frequencies = frequencies[:limit]
	}

	return frequencies
}

// CommandFrequency represents command usage frequency.
type CommandFrequency struct {
	Command string `json:"command"`
	Count   int    `json:"count"`
}
