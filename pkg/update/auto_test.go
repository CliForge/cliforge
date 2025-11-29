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

func TestNewAutoUpdater(t *testing.T) {
	tests := []struct {
		name   string
		config *UpdateConfig
	}{
		{
			name:   "with config",
			config: DefaultUpdateConfig(),
		},
		{
			name:   "with nil config",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			au := NewAutoUpdater(tt.config)
			if au == nil {
				t.Error("NewAutoUpdater() returned nil")
			}
			if au.config == nil {
				t.Error("AutoUpdater config is nil")
			}
			if au.checker == nil {
				t.Error("AutoUpdater checker is nil")
			}
			if au.downloader == nil {
				t.Error("AutoUpdater downloader is nil")
			}
			if au.installer == nil {
				t.Error("AutoUpdater installer is nil")
			}
		})
	}
}

func TestAutoUpdater_CheckAndNotify(t *testing.T) {
	testData := []byte("binary")

	tests := []struct {
		name           string
		releaseVersion string
		currentVersion string
		lastCheck      *LastCheckInfo
		setupServer    func() *httptest.Server
		expectCheck    bool
	}{
		{
			name:           "update available - should check",
			releaseVersion: "2.0.0",
			currentVersion: "1.0.0",
			lastCheck:      nil,
			expectCheck:    true,
		},
		{
			name:           "checked recently - should skip",
			releaseVersion: "2.0.0",
			currentVersion: "1.0.0",
			lastCheck: &LastCheckInfo{
				CheckedAt:     time.Now().Add(-1 * time.Hour),
				LatestVersion: "2.0.0",
			},
			expectCheck: false,
		},
		{
			name:           "checked long ago - should check",
			releaseVersion: "2.0.0",
			currentVersion: "1.0.0",
			lastCheck: &LastCheckInfo{
				CheckedAt:     time.Now().Add(-25 * time.Hour),
				LatestVersion: "2.0.0",
			},
			expectCheck: true,
		},
		{
			name:           "up to date",
			releaseVersion: "1.0.0",
			currentVersion: "1.0.0",
			lastCheck:      nil,
			expectCheck:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup mock HTTP server FIRST
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				release := &ReleaseInfo{
					Version:      tt.releaseVersion,
					URL:          server.URL + "/binary", // Use mock server
					Checksum:     "abc123",
					ChecksumAlgo: "sha256",
					Size:         int64(len(testData)),
				}
				_ = json.NewEncoder(w).Encode(release)
			}))
			defer server.Close()

			// Setup state dir
			stateDir := filepath.Join(tmpDir, "state")
			if err := os.MkdirAll(stateDir, 0700); err != nil {
				t.Fatalf("Failed to create state dir: %v", err)
			}

			// Save last check if provided
			if tt.lastCheck != nil {
				data, err := json.Marshal(tt.lastCheck)
				if err != nil {
					t.Fatalf("Failed to marshal last check: %v", err)
				}
				lastCheckPath := filepath.Join(stateDir, "last_check.json")
				if err := os.WriteFile(lastCheckPath, data, 0600); err != nil {
					t.Fatalf("Failed to write last check: %v", err)
				}
			}

			config := &UpdateConfig{
				CurrentVersion: tt.currentVersion,
				UpdateURL:      server.URL,
				StateDir:       stateDir,
				CacheDir:       filepath.Join(tmpDir, "cache"),
				HTTPTimeout:    5 * time.Second,
				CheckInterval:  24 * time.Hour,
			}

			au := NewAutoUpdater(config)

			// Run CheckAndNotify
			err := au.CheckAndNotify(context.Background())
			if err != nil {
				t.Errorf("CheckAndNotify() error = %v", err)
			}
		})
	}
}

func TestAutoUpdater_CheckAndNotify_ErrorHandling(t *testing.T) {
	tmpDir := t.TempDir()

	// Server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		UpdateURL:      server.URL,
		StateDir:       filepath.Join(tmpDir, "state"),
		CacheDir:       filepath.Join(tmpDir, "cache"),
		HTTPTimeout:    5 * time.Second,
		CheckInterval:  24 * time.Hour,
	}

	au := NewAutoUpdater(config)

	// Should not return error even if check fails
	err := au.CheckAndNotify(context.Background())
	if err != nil {
		t.Errorf("CheckAndNotify() should not error on check failure, got: %v", err)
	}
}

func TestAutoUpdater_Update(t *testing.T) {
	testData := []byte("binary")

	tests := []struct {
		name              string
		releaseVersion    string
		currentVersion    string
		requireConfirm    bool
		wantUpdateAttempt bool
		wantErr           bool
	}{
		{
			name:              "update available - no confirmation",
			releaseVersion:    "2.0.0",
			currentVersion:    "1.0.0",
			requireConfirm:    false,
			wantUpdateAttempt: true,
			wantErr:           true, // Will fail at download/install stage
		},
		{
			name:              "up to date",
			releaseVersion:    "1.0.0",
			currentVersion:    "1.0.0",
			requireConfirm:    false,
			wantUpdateAttempt: false,
			wantErr:           false,
		},
		{
			name:              "current version newer",
			releaseVersion:    "1.0.0",
			currentVersion:    "2.0.0",
			requireConfirm:    false,
			wantUpdateAttempt: false,
			wantErr:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup binary download server (returns actual binary data)
			binaryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write(testData)
			}))
			defer binaryServer.Close()

			// Setup release check server (returns release info)
			releaseServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				release := &ReleaseInfo{
					Version:      tt.releaseVersion,
					URL:          binaryServer.URL + "/binary", // Point to binary server
					Checksum:     "abc123",
					ChecksumAlgo: "sha256",
					Size:         int64(len(testData)),
				}
				_ = json.NewEncoder(w).Encode(release)
			}))
			defer releaseServer.Close()

			config := &UpdateConfig{
				CurrentVersion:      tt.currentVersion,
				UpdateURL:           releaseServer.URL,
				StateDir:            filepath.Join(tmpDir, "state"),
				CacheDir:            filepath.Join(tmpDir, "cache"),
				HTTPTimeout:         5 * time.Second,
				RequireConfirmation: tt.requireConfirm,
			}

			au := NewAutoUpdater(config)

			err := au.Update(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAutoUpdater_Update_CheckError(t *testing.T) {
	tmpDir := t.TempDir()

	// Server that returns error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		UpdateURL:      server.URL,
		StateDir:       filepath.Join(tmpDir, "state"),
		CacheDir:       filepath.Join(tmpDir, "cache"),
		HTTPTimeout:    5 * time.Second,
	}

	au := NewAutoUpdater(config)

	err := au.Update(context.Background())
	if err == nil {
		t.Error("Update() should return error when check fails")
	}
}

func TestAutoUpdater_SkipVersion(t *testing.T) {
	testData := []byte("binary")

	tests := []struct {
		name           string
		releaseVersion string
		currentVersion string
		wantErr        bool
	}{
		{
			name:           "update available",
			releaseVersion: "2.0.0",
			currentVersion: "1.0.0",
			wantErr:        false,
		},
		{
			name:           "no update available",
			releaseVersion: "1.0.0",
			currentVersion: "1.0.0",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup mock HTTP server FIRST
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				release := &ReleaseInfo{
					Version:      tt.releaseVersion,
					URL:          server.URL + "/binary", // Use mock server
					Checksum:     "abc123",
					ChecksumAlgo: "sha256",
					Size:         int64(len(testData)),
				}
				_ = json.NewEncoder(w).Encode(release)
			}))
			defer server.Close()

			config := &UpdateConfig{
				CurrentVersion: tt.currentVersion,
				UpdateURL:      server.URL,
				StateDir:       filepath.Join(tmpDir, "state"),
				CacheDir:       filepath.Join(tmpDir, "cache"),
				HTTPTimeout:    5 * time.Second,
			}

			au := NewAutoUpdater(config)

			err := au.SkipVersion(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("SkipVersion() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify version was skipped if update was available
			if !tt.wantErr && tt.releaseVersion != tt.currentVersion {
				info, err := au.checker.GetLastCheck()
				if err != nil {
					t.Fatalf("GetLastCheck() error = %v", err)
				}
				if !info.UpdateSkipped {
					t.Error("UpdateSkipped should be true")
				}
				if info.SkippedVersion != tt.releaseVersion {
					t.Errorf("SkippedVersion = %s, want %s", info.SkippedVersion, tt.releaseVersion)
				}
			}
		})
	}
}

func TestAutoUpdater_SkipVersion_CheckError(t *testing.T) {
	tmpDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		UpdateURL:      server.URL,
		StateDir:       filepath.Join(tmpDir, "state"),
		CacheDir:       filepath.Join(tmpDir, "cache"),
		HTTPTimeout:    5 * time.Second,
	}

	au := NewAutoUpdater(config)

	err := au.SkipVersion(context.Background())
	if err == nil {
		t.Error("SkipVersion() should return error when check fails")
	}
}

func TestAutoUpdater_Status(t *testing.T) {
	testData := []byte("binary")

	tests := []struct {
		name           string
		releaseVersion string
		currentVersion string
		wantErr        bool
	}{
		{
			name:           "update available",
			releaseVersion: "2.0.0",
			currentVersion: "1.0.0",
			wantErr:        false,
		},
		{
			name:           "up to date",
			releaseVersion: "1.0.0",
			currentVersion: "1.0.0",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Setup mock HTTP server FIRST
			var server *httptest.Server
			server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				release := &ReleaseInfo{
					Version:      tt.releaseVersion,
					URL:          server.URL + "/binary", // Use mock server
					Checksum:     "abc123",
					ChecksumAlgo: "sha256",
					Size:         int64(len(testData)),
					Critical:     true,
				}
				_ = json.NewEncoder(w).Encode(release)
			}))
			defer server.Close()

			config := &UpdateConfig{
				CurrentVersion: tt.currentVersion,
				UpdateURL:      server.URL,
				StateDir:       filepath.Join(tmpDir, "state"),
				CacheDir:       filepath.Join(tmpDir, "cache"),
				HTTPTimeout:    5 * time.Second,
			}

			au := NewAutoUpdater(config)

			err := au.Status(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Status() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAutoUpdater_Status_CheckError(t *testing.T) {
	tmpDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		UpdateURL:      server.URL,
		StateDir:       filepath.Join(tmpDir, "state"),
		CacheDir:       filepath.Join(tmpDir, "cache"),
		HTTPTimeout:    5 * time.Second,
	}

	au := NewAutoUpdater(config)

	err := au.Status(context.Background())
	if err == nil {
		t.Error("Status() should return error when check fails")
	}
}

func TestAutoUpdater_CleanupCache(t *testing.T) {
	tmpDir := t.TempDir()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		StateDir:       filepath.Join(tmpDir, "state"),
		CacheDir:       filepath.Join(tmpDir, "cache"),
		HTTPTimeout:    5 * time.Second,
	}

	// Create cache directory with old files
	if err := os.MkdirAll(config.CacheDir, 0700); err != nil {
		t.Fatalf("Failed to create cache dir: %v", err)
	}

	oldFile := filepath.Join(config.CacheDir, "old.tmp")
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}

	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set file time: %v", err)
	}

	au := NewAutoUpdater(config)

	err := au.CleanupCache()
	if err != nil {
		t.Errorf("CleanupCache() error = %v", err)
	}

	// Verify old file was removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been removed")
	}
}
