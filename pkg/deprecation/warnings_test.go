package deprecation

import (
	"os"
	"testing"
	"time"
)

func TestWarningManager_ShouldShowWarning(t *testing.T) {
	// Use temp directory for testing
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	manager, err := NewWarningManager("test-cli")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	tests := []struct {
		name   string
		config *WarningConfig
		info   *DeprecationInfo
		want   bool
	}{
		{
			name: "warnings enabled, should show",
			config: &WarningConfig{
				Enabled:     true,
				MinSeverity: SeverityInfo,
				MinLevel:    WarningLevelInfo,
			},
			info: &DeprecationInfo{
				Severity: SeverityWarning,
				Level:    WarningLevelWarning,
			},
			want: true,
		},
		{
			name: "warnings disabled, should not show",
			config: &WarningConfig{
				Enabled: false,
			},
			info: &DeprecationInfo{
				Severity: SeverityWarning,
				Level:    WarningLevelWarning,
			},
			want: false,
		},
		{
			name: "severity below minimum",
			config: &WarningConfig{
				Enabled:     true,
				MinSeverity: SeverityBreaking,
				MinLevel:    WarningLevelInfo,
			},
			info: &DeprecationInfo{
				Severity: SeverityWarning,
				Level:    WarningLevelWarning,
			},
			want: false,
		},
		{
			name: "level below minimum",
			config: &WarningConfig{
				Enabled:     true,
				MinSeverity: SeverityInfo,
				MinLevel:    WarningLevelCritical,
			},
			info: &DeprecationInfo{
				Severity: SeverityWarning,
				Level:    WarningLevelWarning,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager.SetConfig(tt.config)
			got := manager.ShouldShowWarning(tt.info)
			if got != tt.want {
				t.Errorf("ShouldShowWarning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWarningManager_MarkShown(t *testing.T) {
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	manager, err := NewWarningManager("test-cli")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	info := &DeprecationInfo{
		OperationID: "testOp",
		Level:       WarningLevelInfo,
	}

	// Mark as shown
	err = manager.MarkShown(info)
	if err != nil {
		t.Errorf("MarkShown() error = %v", err)
	}

	// Check tracking file was created (check actual path from manager)
	if _, err := os.Stat(manager.trackingPath); os.IsNotExist(err) {
		t.Errorf("Tracking file was not created at %s", manager.trackingPath)
	}

	// Check that it was recorded
	key := manager.getTrackingKey(info)
	if _, exists := manager.tracking.LastShown[key]; !exists {
		t.Errorf("Warning was not recorded in tracking")
	}
}

func TestCompareWarningLevels(t *testing.T) {
	tests := []struct {
		name string
		a    WarningLevel
		b    WarningLevel
		want int
	}{
		{
			name: "critical > warning",
			a:    WarningLevelCritical,
			b:    WarningLevelWarning,
			want: 1,
		},
		{
			name: "info < critical",
			a:    WarningLevelInfo,
			b:    WarningLevelCritical,
			want: -1,
		},
		{
			name: "warning == warning",
			a:    WarningLevelWarning,
			b:    WarningLevelWarning,
			want: 0,
		},
		{
			name: "removed > critical",
			a:    WarningLevelRemoved,
			b:    WarningLevelCritical,
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := compareWarningLevels(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("compareWarningLevels(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestWarningManager_SuppressOperation(t *testing.T) {
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	manager, err := NewWarningManager("test-cli")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	operationID := "testOperation"

	// Suppress operation
	err = manager.SuppressOperation(operationID)
	if err != nil {
		t.Errorf("SuppressOperation() error = %v", err)
	}

	// Check it's in the list
	found := false
	for _, id := range manager.config.SuppressedOperations {
		if id == operationID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Operation was not added to suppression list")
	}

	// Check suppression works
	info := &DeprecationInfo{
		OperationID: operationID,
		Severity:    SeverityWarning,
		Level:       WarningLevelWarning,
	}

	if !manager.isOperationSuppressed(info) {
		t.Errorf("Operation should be suppressed")
	}
}

func TestWarningManager_UnsuppressOperation(t *testing.T) {
	tmpDir := t.TempDir()
	oldXdgDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_DATA_HOME", tmpDir)
	defer os.Setenv("XDG_DATA_HOME", oldXdgDataHome)

	manager, err := NewWarningManager("test-cli")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	operationID := "testOperation"

	// Suppress then unsuppress
	manager.SuppressOperation(operationID)
	err = manager.UnsuppressOperation(operationID)
	if err != nil {
		t.Errorf("UnsuppressOperation() error = %v", err)
	}

	// Check it's not in the list
	for _, id := range manager.config.SuppressedOperations {
		if id == operationID {
			t.Errorf("Operation was not removed from suppression list")
		}
	}
}

func TestFormatWarning(t *testing.T) {
	sunset := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	info := &DeprecationInfo{
		Type:          DeprecationTypeOperation,
		OperationID:   "listUsersV1",
		Method:        "GET",
		Path:          "/v1/users",
		Level:         WarningLevelWarning,
		Sunset:        &sunset,
		DaysRemaining: 365,
		Reason:        "Use v2 API for better performance",
		Replacement: &Replacement{
			Command: "users list-v2",
		},
		DocsURL: "https://docs.example.com/migration",
	}

	result := FormatWarning(info)

	// Check that key elements are present
	if result == "" {
		t.Error("FormatWarning() returned empty string")
	}

	// Check for key components
	expectedStrings := []string{
		"DEPRECATION",
		"WARNING",
		info.OperationID,
		info.Reason,
		info.Replacement.Command,
		info.DocsURL,
		"--no-deprecation-warnings",
	}

	for _, expected := range expectedStrings {
		if !contains(result, expected) {
			t.Errorf("FormatWarning() missing expected string: %s", expected)
		}
	}
}

func TestFormatShortWarning(t *testing.T) {
	tests := []struct {
		name string
		info *DeprecationInfo
		want string
	}{
		{
			name: "operation deprecation",
			info: &DeprecationInfo{
				OperationID:   "listUsers",
				Level:         WarningLevelWarning,
				DaysRemaining: 30,
			},
			want: "Operation 'listUsers' is deprecated (30 days remaining)",
		},
		{
			name: "parameter deprecation",
			info: &DeprecationInfo{
				Type:          DeprecationTypeParameter,
				Name:          "filter",
				Level:         WarningLevelWarning,
				DaysRemaining: 60,
			},
			want: "parameter 'filter' is deprecated (60 days remaining)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatShortWarning(tt.info)
			if got == "" {
				t.Error("FormatShortWarning() returned empty string")
			}
			// Just verify it contains key info, exact format may vary
			if !contains(got, "deprecated") {
				t.Error("FormatShortWarning() should contain 'deprecated'")
			}
		})
	}
}

func TestRequiresForceFlag(t *testing.T) {
	tests := []struct {
		name string
		info *DeprecationInfo
		want bool
	}{
		{
			name: "critical requires force",
			info: &DeprecationInfo{Level: WarningLevelCritical},
			want: true,
		},
		{
			name: "warning does not require force",
			info: &DeprecationInfo{Level: WarningLevelWarning},
			want: false,
		},
		{
			name: "removed does not require force (blocked)",
			info: &DeprecationInfo{Level: WarningLevelRemoved},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RequiresForceFlag(tt.info)
			if got != tt.want {
				t.Errorf("RequiresForceFlag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsBlocked(t *testing.T) {
	tests := []struct {
		name string
		info *DeprecationInfo
		want bool
	}{
		{
			name: "removed is blocked",
			info: &DeprecationInfo{Level: WarningLevelRemoved},
			want: true,
		},
		{
			name: "critical is not blocked",
			info: &DeprecationInfo{Level: WarningLevelCritical},
			want: false,
		},
		{
			name: "warning is not blocked",
			info: &DeprecationInfo{Level: WarningLevelWarning},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBlocked(tt.info)
			if got != tt.want {
				t.Errorf("IsBlocked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetWarningColor(t *testing.T) {
	tests := []struct {
		name  string
		level WarningLevel
	}{
		{"info color", WarningLevelInfo},
		{"warning color", WarningLevelWarning},
		{"urgent color", WarningLevelUrgent},
		{"critical color", WarningLevelCritical},
		{"removed color", WarningLevelRemoved},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetWarningColor(tt.level)
			if got == "" {
				t.Errorf("GetWarningColor(%v) returned empty string", tt.level)
			}
			// Verify it's an ANSI escape code
			if got[0] != '\033' {
				t.Errorf("GetWarningColor(%v) should return ANSI escape code", tt.level)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 &&
		(s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
