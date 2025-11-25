package deprecation

import (
	"os"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/openapi"
)

func TestNotifier_ShowUpdateNotification(t *testing.T) {
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	notifier, err := NewNotifier("test-cli")
	if err != nil {
		t.Fatalf("Failed to create notifier: %v", err)
	}

	changelog := []openapi.ChangelogEntry{
		{
			Version: "1.1.0",
			Date:    "2025-01-15",
			Changes: []*openapi.Change{
				{
					Type:        "deprecated",
					Description: "Operation X is deprecated",
					Sunset:      "2025-12-31",
				},
			},
		},
	}

	// First call should show notification
	err = notifier.ShowUpdateNotification("1.1.0", changelog)
	if err != nil {
		t.Errorf("ShowUpdateNotification() error = %v", err)
	}

	// Version should be saved (check actual path from notifier)
	if _, err := os.Stat(notifier.lastVersionFile); os.IsNotExist(err) {
		t.Errorf("Version file was not created at %s", notifier.lastVersionFile)
	}

	// Second call with same version should not show notification
	// (this is implicit in the implementation)
	err = notifier.ShowUpdateNotification("1.1.0", changelog)
	if err != nil {
		t.Errorf("ShowUpdateNotification() error = %v", err)
	}
}

func TestNotifier_GetLastVersion(t *testing.T) {
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	notifier, err := NewNotifier("test-cli-versioning")
	if err != nil {
		t.Fatalf("Failed to create notifier: %v", err)
	}

	// Remember initial version (may or may not be empty)
	initialVersion := notifier.GetLastVersion()

	// Save a version
	notifier.saveVersion("1.0.0")

	// Should be retrievable
	if notifier.GetLastVersion() != "1.0.0" {
		t.Errorf("GetLastVersion() = %v, want 1.0.0", notifier.GetLastVersion())
	}

	// Create new notifier, should load persisted version
	notifier2, err := NewNotifier("test-cli-versioning")
	if err != nil {
		t.Fatalf("Failed to create notifier: %v", err)
	}

	if notifier2.GetLastVersion() != "1.0.0" {
		t.Errorf("Loaded last version = %v, want 1.0.0", notifier2.GetLastVersion())
	}

	// Verify it changed from initial
	if initialVersion == "1.0.0" {
		t.Log("Initial version happened to be 1.0.0, test may not be fully validating")
	}
}

func TestFormatChangelog(t *testing.T) {
	entries := []openapi.ChangelogEntry{
		{
			Version: "1.1.0",
			Date:    "2025-01-15",
			Changes: []*openapi.Change{
				{
					Type:        "deprecated",
					Description: "Operation X is deprecated",
					Sunset:      "2025-12-31",
					Migration:   "Use operation Y instead",
				},
				{
					Type:        "added",
					Description: "New feature Z",
				},
			},
		},
		{
			Version: "1.0.0",
			Date:    "2024-12-01",
			Changes: []*openapi.Change{
				{
					Type:        "breaking",
					Severity:    "breaking",
					Description: "Breaking change A",
				},
			},
		},
	}

	result := FormatChangelog(entries, 0)

	if result == "" {
		t.Error("FormatChangelog() returned empty string")
	}

	// Check for key components
	expectedStrings := []string{
		"Changelog",
		"Version 1.1.0",
		"Version 1.0.0",
		"Operation X is deprecated",
		"New feature Z",
		"Breaking change A",
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("FormatChangelog() missing expected string: %s", expected)
		}
	}
}

func TestFormatChangelog_WithLimit(t *testing.T) {
	entries := []openapi.ChangelogEntry{
		{Version: "1.2.0", Date: "2025-02-01"},
		{Version: "1.1.0", Date: "2025-01-15"},
		{Version: "1.0.0", Date: "2024-12-01"},
	}

	result := FormatChangelog(entries, 2)

	// Should only include first 2 versions
	if contains(result, "1.0.0") {
		t.Error("FormatChangelog() should not include 1.0.0 when limit is 2")
	}
	if !contains(result, "1.2.0") || !contains(result, "1.1.0") {
		t.Error("FormatChangelog() should include 1.2.0 and 1.1.0")
	}
}

func TestFormatChangelog_Empty(t *testing.T) {
	result := FormatChangelog([]openapi.ChangelogEntry{}, 0)

	expected := "No changelog entries available"
	if result != expected {
		t.Errorf("FormatChangelog() = %v, want %v", result, expected)
	}
}

func TestAcknowledgmentTracker_Acknowledge(t *testing.T) {
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	tracker, err := NewAcknowledgmentTracker("test-cli")
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	key := "op:listUsersV1"

	// Initially not acknowledged
	if tracker.IsAcknowledged(key) {
		t.Error("Should not be acknowledged initially")
	}

	// Acknowledge
	err = tracker.Acknowledge(key)
	if err != nil {
		t.Errorf("Acknowledge() error = %v", err)
	}

	// Should now be acknowledged
	if !tracker.IsAcknowledged(key) {
		t.Error("Should be acknowledged after Acknowledge()")
	}

	// Should have timestamp
	timestamp := tracker.GetAcknowledgmentTime(key)
	if timestamp == nil {
		t.Error("Should have acknowledgment timestamp")
	}

	// Check persistence (check actual path from tracker)
	if _, err := os.Stat(tracker.trackingPath); os.IsNotExist(err) {
		t.Errorf("Acknowledgment file was not created at %s", tracker.trackingPath)
	}

	// Load from disk
	tracker2, err := NewAcknowledgmentTracker("test-cli")
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	if !tracker2.IsAcknowledged(key) {
		t.Error("Loaded tracker should have acknowledgment")
	}
}

func TestAcknowledgmentTracker_ClearAcknowledgment(t *testing.T) {
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	tracker, err := NewAcknowledgmentTracker("test-cli")
	if err != nil {
		t.Fatalf("Failed to create tracker: %v", err)
	}

	key := "op:listUsersV1"

	// Acknowledge then clear
	tracker.Acknowledge(key)
	err = tracker.ClearAcknowledgment(key)
	if err != nil {
		t.Errorf("ClearAcknowledgment() error = %v", err)
	}

	// Should not be acknowledged
	if tracker.IsAcknowledged(key) {
		t.Error("Should not be acknowledged after ClearAcknowledgment()")
	}

	// Timestamp should be nil
	if tracker.GetAcknowledgmentTime(key) != nil {
		t.Error("Acknowledgment time should be nil after clearing")
	}
}

func TestFormatDeprecationList(t *testing.T) {
	sunset := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	deprecations := []*DeprecationInfo{
		{
			Type:          DeprecationTypeOperation,
			OperationID:   "listUsersV1",
			Method:        "GET",
			Path:          "/v1/users",
			Level:         WarningLevelCritical,
			Sunset:        &sunset,
			DaysRemaining: 30,
			Replacement: &Replacement{
				Command: "users list-v2",
			},
			DocsURL: "https://docs.example.com",
		},
		{
			Type:  DeprecationTypeParameter,
			Name:  "filter",
			Level: WarningLevelWarning,
		},
	}

	result := FormatDeprecationList(deprecations)

	if result == "" {
		t.Error("FormatDeprecationList() returned empty string")
	}

	// Check for key components
	expectedStrings := []string{
		"Active Deprecations",
		"API Deprecations",
		"Parameter Deprecations",
		"listUsersV1",
		"filter",
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("FormatDeprecationList() missing expected string: %s", expected)
		}
	}
}

func TestFormatDeprecationList_Empty(t *testing.T) {
	result := FormatDeprecationList([]*DeprecationInfo{})

	expected := "No active deprecations"
	if result != expected {
		t.Errorf("FormatDeprecationList() = %v, want %v", result, expected)
	}
}

func TestExtractDeprecationsFromChangelog(t *testing.T) {
	changelog := []openapi.ChangelogEntry{
		{
			Version: "1.1.0",
			Changes: []*openapi.Change{
				{
					Type:        "deprecated",
					Description: "Operation X deprecated",
				},
				{
					Type:        "added",
					Description: "Feature Y added",
				},
				{
					Type:        "deprecated",
					Description: "Parameter Z deprecated",
				},
			},
		},
	}

	result := extractDeprecationsFromChangelog(changelog)

	if len(result) != 2 {
		t.Errorf("extractDeprecationsFromChangelog() returned %d items, want 2", len(result))
	}

	// Verify all returned changes are deprecations
	for _, change := range result {
		if change.Type != "deprecated" {
			t.Errorf("extractDeprecationsFromChangelog() returned non-deprecated change: %v", change.Type)
		}
	}
}

func TestFormatDeprecationSummary(t *testing.T) {
	sunset := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	info := &DeprecationInfo{
		Type:          DeprecationTypeOperation,
		OperationID:   "listUsersV1",
		Method:        "GET",
		Path:          "/v1/users",
		Level:         WarningLevelCritical,
		Sunset:        &sunset,
		DaysRemaining: 30,
		Replacement: &Replacement{
			Command: "users list-v2",
		},
		DocsURL: "https://docs.example.com",
	}

	result := formatDeprecationSummary(info)

	if result == "" {
		t.Error("formatDeprecationSummary() returned empty string")
	}

	// Check for key components
	expectedStrings := []string{
		"CRITICAL",
		"30 days",
		"listUsersV1",
		"users list-v2",
		"https://docs.example.com",
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("formatDeprecationSummary() missing expected string: %s", expected)
		}
	}
}

func TestShowBlockedNotice(t *testing.T) {
	sunset := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	info := &DeprecationInfo{
		OperationID:   "listUsersV1",
		Level:         WarningLevelRemoved,
		Sunset:        &sunset,
		DaysRemaining: -10,
		Replacement: &Replacement{
			Command:   "users list-v2",
			Migration: "Use the v2 API endpoint",
		},
		DocsURL: "https://docs.example.com",
	}

	err := ShowBlockedNotice(info)

	if err == nil {
		t.Error("ShowBlockedNotice() should return error")
	}

	if !contains(err.Error(), "blocked") {
		t.Errorf("Error should mention 'blocked': %v", err)
	}
}

func TestShowCriticalNotice(t *testing.T) {
	sunset := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)

	info := &DeprecationInfo{
		OperationID:   "listUsersV1",
		Level:         WarningLevelCritical,
		Sunset:        &sunset,
		DaysRemaining: 5,
		Replacement: &Replacement{
			Command: "users list-v2",
		},
		DocsURL: "https://docs.example.com",
	}

	err := ShowCriticalNotice(info)

	if err == nil {
		t.Error("ShowCriticalNotice() should return error")
	}

	if !contains(err.Error(), "--force") {
		t.Errorf("Error should mention '--force': %v", err)
	}
}
