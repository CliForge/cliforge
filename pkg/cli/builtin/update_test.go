package builtin

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

func TestNewUpdateCommand(t *testing.T) {
	config := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Name:    "testcli",
			Version: "1.0.0",
		},
	}

	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		Config:         config,
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	cmd := NewUpdateCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "update" {
		t.Errorf("expected Use 'update', got %q", cmd.Use)
	}
}

func TestRunUpdate_NoUpdateAvailable(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.0.0",
				UpdateAvailable: false,
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "already running the latest version") {
		t.Errorf("expected 'already running the latest version', got: %s", result)
	}
}

func TestRunUpdate_UpdateAvailable(t *testing.T) {
	updatePerformed := false
	output := &bytes.Buffer{}

	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.1.0",
				UpdateAvailable: true,
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			updatePerformed = true
			if version != "1.1.0" {
				t.Errorf("expected version '1.1.0', got %q", version)
			}
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	if !updatePerformed {
		t.Error("expected update to be performed")
	}

	result := output.String()
	if !strings.Contains(result, "Updated to 1.1.0") {
		t.Errorf("expected update success message, got: %s", result)
	}
}

func TestRunUpdate_WithChangelog(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		ShowChangelog:  true,
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.1.0",
				UpdateAvailable: true,
				Changelog: []string{
					"Added new feature",
					"Fixed bug in config",
				},
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Changelog:") {
		t.Error("expected changelog header")
	}
	if !strings.Contains(result, "Added new feature") {
		t.Error("expected changelog entry")
	}
	if !strings.Contains(result, "Fixed bug in config") {
		t.Error("expected changelog entry")
	}
}

func TestRunUpdate_WithReleaseNotes(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.1.0",
				UpdateAvailable: true,
				ReleaseNotes:    "This is a major release with important fixes.",
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "This is a major release") {
		t.Error("expected release notes in output")
	}
}

func TestRunUpdate_CheckFailed(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return nil, errors.New("network error")
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "failed to check for updates") {
		t.Errorf("expected check error, got: %v", err)
	}
}

func TestRunUpdate_UpdateFailed(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.1.0",
				UpdateAvailable: true,
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return errors.New("download failed")
		},
	}

	err := runUpdate(opts, false, true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "update failed") {
		t.Errorf("expected update error, got: %v", err)
	}
}

func TestRunUpdate_ForceUpdate(t *testing.T) {
	updatePerformed := false
	output := &bytes.Buffer{}

	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.0.0",
				UpdateAvailable: false,
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			updatePerformed = true
			return nil
		},
	}

	// Without force, should not update
	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	if updatePerformed {
		t.Error("did not expect update to be performed without force")
	}

	// With force, should update
	output.Reset()
	err = runUpdate(opts, true, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	if !updatePerformed {
		t.Error("expected update to be performed with force")
	}
}

func TestRunUpdate_VersionDisplay(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.2.0",
				UpdateAvailable: true,
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Current version: 1.0.0") {
		t.Error("expected current version in output")
	}
	if !strings.Contains(result, "Latest version: 1.2.0") {
		t.Error("expected latest version in output")
	}
}

func TestDefaultCheckUpdateFunc(t *testing.T) {
	config := &cli.CLIConfig{
		Metadata: cli.Metadata{
			Version: "1.0.0",
		},
	}

	checkFunc := DefaultCheckUpdateFunc(config, "1.0.0")
	info, err := checkFunc(context.Background())

	if err != nil {
		t.Fatalf("checkFunc failed: %v", err)
	}

	if info.CurrentVersion != "1.0.0" {
		t.Errorf("expected current version '1.0.0', got %q", info.CurrentVersion)
	}

	if info.UpdateAvailable {
		t.Error("expected no update available")
	}
}

func TestDefaultPerformUpdateFunc(t *testing.T) {
	performFunc := DefaultPerformUpdateFunc()
	err := performFunc(context.Background(), "1.1.0")

	if err == nil {
		t.Fatal("expected error from default implementation")
	}

	if !strings.Contains(err.Error(), "not implemented") {
		t.Errorf("expected 'not implemented' error, got: %v", err)
	}
}

func TestUpdateInfo_Structure(t *testing.T) {
	info := &UpdateInfo{
		CurrentVersion:  "1.0.0",
		LatestVersion:   "1.1.0",
		UpdateAvailable: true,
		DownloadURL:     "https://example.com/download",
		Changelog:       []string{"Feature A", "Fix B"},
		ReleaseNotes:    "Important release",
	}

	if info.CurrentVersion != "1.0.0" {
		t.Error("CurrentVersion not set correctly")
	}
	if info.LatestVersion != "1.1.0" {
		t.Error("LatestVersion not set correctly")
	}
	if !info.UpdateAvailable {
		t.Error("UpdateAvailable should be true")
	}
	if len(info.Changelog) != 2 {
		t.Error("Changelog not set correctly")
	}
}

func TestRunUpdate_EmptyChangelog(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		ShowChangelog:  true,
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "1.1.0",
				UpdateAvailable: true,
				Changelog:       []string{},
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	result := output.String()
	// Should not show "Changelog:" if there are no changelog entries
	changelogCount := strings.Count(result, "Changelog:")
	if changelogCount > 0 {
		t.Error("did not expect Changelog section for empty changelog")
	}
}

func TestRunUpdate_Downloading(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &UpdateOptions{
		CurrentVersion: "1.0.0",
		Output:         output,
		CheckUpdateFunc: func(ctx context.Context) (*UpdateInfo, error) {
			return &UpdateInfo{
				CurrentVersion:  "1.0.0",
				LatestVersion:   "2.0.0",
				UpdateAvailable: true,
			}, nil
		},
		PerformUpdateFunc: func(ctx context.Context, version string) error {
			return nil
		},
	}

	err := runUpdate(opts, false, true)
	if err != nil {
		t.Fatalf("runUpdate failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Downloading 2.0.0") {
		t.Error("expected downloading message")
	}
	if !strings.Contains(result, "Installing...") {
		t.Error("expected installing message")
	}
}
