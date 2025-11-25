package update

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestChecker_Check(t *testing.T) {
	// Create test server
	release := &ReleaseInfo{
		Version:      "1.2.3",
		URL:          "https://example.com/download",
		Checksum:     "abc123",
		ChecksumAlgo: "sha256",
		Size:         1024,
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(release)
	}))
	defer server.Close()

	tests := []struct {
		name           string
		currentVersion string
		wantStatus     UpdateStatus
		wantErr        bool
	}{
		{
			name:           "newer version available",
			currentVersion: "1.0.0",
			wantStatus:     UpdateStatusAvailable,
			wantErr:        false,
		},
		{
			name:           "up to date",
			currentVersion: "1.2.3",
			wantStatus:     UpdateStatusUpToDate,
			wantErr:        false,
		},
		{
			name:           "current version newer",
			currentVersion: "2.0.0",
			wantStatus:     UpdateStatusUpToDate,
			wantErr:        false,
		},
		{
			name:           "invalid current version",
			currentVersion: "invalid",
			wantStatus:     UpdateStatusFailed,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp directory for state
			tmpDir := t.TempDir()

			config := &UpdateConfig{
				CurrentVersion: tt.currentVersion,
				UpdateURL:      server.URL,
				StateDir:       tmpDir,
				HTTPTimeout:    5 * time.Second,
			}

			checker := NewChecker(config)
			result, err := checker.Check(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("Checker.Check() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if result.Status != tt.wantStatus {
				t.Errorf("Checker.Check() status = %v, want %v", result.Status, tt.wantStatus)
			}
		})
	}
}

func TestChecker_ShouldNotify(t *testing.T) {
	tmpDir := t.TempDir()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		StateDir:       tmpDir,
	}

	checker := NewChecker(config)

	// Create a check result with update available
	currentVersion, _ := ParseVersion("1.0.0")
	latestVersion, _ := ParseVersion("1.2.3")

	result := &CheckResult{
		Status:         UpdateStatusAvailable,
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		CheckedAt:      time.Now(),
	}

	// First check - should notify
	shouldNotify, err := checker.ShouldNotify(result)
	if err != nil {
		t.Fatalf("ShouldNotify() error = %v", err)
	}
	if !shouldNotify {
		t.Error("ShouldNotify() = false, want true (first check)")
	}

	// Skip version
	if err := checker.SkipVersion("1.2.3"); err != nil {
		t.Fatalf("SkipVersion() error = %v", err)
	}

	// Second check - should not notify (version skipped)
	shouldNotify, err = checker.ShouldNotify(result)
	if err != nil {
		t.Fatalf("ShouldNotify() error = %v", err)
	}
	if shouldNotify {
		t.Error("ShouldNotify() = true, want false (version skipped)")
	}
}

func TestChecker_GetLastCheck(t *testing.T) {
	tmpDir := t.TempDir()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		StateDir:       tmpDir,
	}

	checker := NewChecker(config)

	// First check - should return empty info
	info, err := checker.GetLastCheck()
	if err != nil {
		t.Fatalf("GetLastCheck() error = %v", err)
	}
	if !info.CheckedAt.IsZero() {
		t.Error("GetLastCheck() returned non-zero time for first check")
	}

	// Save check info
	testInfo := &LastCheckInfo{
		CheckedAt:     time.Now(),
		LatestVersion: "1.2.3",
	}

	data, _ := json.Marshal(testInfo)
	lastCheckPath := filepath.Join(tmpDir, "last_check.json")
	if err := os.WriteFile(lastCheckPath, data, 0600); err != nil {
		t.Fatalf("Failed to write test data: %v", err)
	}

	// Second check - should return saved info
	info, err = checker.GetLastCheck()
	if err != nil {
		t.Fatalf("GetLastCheck() error = %v", err)
	}
	if info.LatestVersion != "1.2.3" {
		t.Errorf("GetLastCheck() LatestVersion = %s, want 1.2.3", info.LatestVersion)
	}
}

func TestLastCheckInfo_ShouldCheck(t *testing.T) {
	tests := []struct {
		name      string
		checkedAt time.Time
		interval  time.Duration
		want      bool
	}{
		{
			name:      "never checked",
			checkedAt: time.Time{},
			interval:  24 * time.Hour,
			want:      true,
		},
		{
			name:      "checked recently",
			checkedAt: time.Now().Add(-1 * time.Hour),
			interval:  24 * time.Hour,
			want:      false,
		},
		{
			name:      "checked long ago",
			checkedAt: time.Now().Add(-25 * time.Hour),
			interval:  24 * time.Hour,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &LastCheckInfo{
				CheckedAt: tt.checkedAt,
			}

			if got := info.ShouldCheck(tt.interval); got != tt.want {
				t.Errorf("LastCheckInfo.ShouldCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckResult_UpdateAvailable(t *testing.T) {
	tests := []struct {
		name    string
		status  UpdateStatus
		current string
		latest  string
		want    bool
	}{
		{
			name:    "update available",
			status:  UpdateStatusAvailable,
			current: "1.0.0",
			latest:  "1.2.3",
			want:    true,
		},
		{
			name:    "up to date",
			status:  UpdateStatusUpToDate,
			current: "1.2.3",
			latest:  "1.2.3",
			want:    false,
		},
		{
			name:    "wrong status",
			status:  UpdateStatusFailed,
			current: "1.0.0",
			latest:  "1.2.3",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentVersion, _ := ParseVersion(tt.current)
			latestVersion, _ := ParseVersion(tt.latest)

			result := &CheckResult{
				Status:         tt.status,
				CurrentVersion: currentVersion,
				LatestVersion:  latestVersion,
			}

			if got := result.UpdateAvailable(); got != tt.want {
				t.Errorf("CheckResult.UpdateAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
