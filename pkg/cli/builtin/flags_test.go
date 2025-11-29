package builtin

import (
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/spf13/cobra"
)

func TestNewFlagManager(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
	}

	fm := NewFlagManager(config)

	if fm == nil {
		t.Fatal("expected flag manager, got nil")
	}

	if fm.config != config {
		t.Error("expected config to be set")
	}

	if fm.flags == nil {
		t.Error("expected flags to be initialized")
	}
}

func TestFlagManager_Validate(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
	}

	fm := NewFlagManager(config)

	tests := []struct {
		name    string
		setup   func(*GlobalFlags)
		wantErr bool
	}{
		{
			name: "valid flags",
			setup: func(f *GlobalFlags) {
				f.Output = "json"
				f.Timeout = 30 * time.Second
				f.Retry = 3
			},
			wantErr: false,
		},
		{
			name: "verbose and quiet conflict",
			setup: func(f *GlobalFlags) {
				f.Verbose = 1
				f.Quiet = true
			},
			wantErr: true,
		},
		{
			name: "invalid output format",
			setup: func(f *GlobalFlags) {
				f.Output = "invalid"
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			setup: func(f *GlobalFlags) {
				f.Timeout = -1 * time.Second
			},
			wantErr: true,
		},
		{
			name: "negative retry",
			setup: func(f *GlobalFlags) {
				f.Retry = -1
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm.flags = &GlobalFlags{}
			tt.setup(fm.flags)

			err := fm.Validate()

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestFlagManager_AddGlobalFlags(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
		Behaviors: &cli.Behaviors{
			GlobalFlags: &cli.GlobalFlags{
				Config: &cli.GlobalFlag{
					Enabled: true,
				},
				Output: &cli.GlobalFlag{
					Enabled: true,
				},
				Verbose: &cli.GlobalFlag{
					Enabled: true,
				},
			},
		},
	}

	fm := NewFlagManager(config)
	cmd := &cobra.Command{}

	fm.AddGlobalFlags(cmd)

	// Check that flags were added
	configFlag := cmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("expected --config flag to be added")
	}

	outputFlag := cmd.PersistentFlags().Lookup("output")
	if outputFlag == nil {
		t.Error("expected --output flag to be added")
	}

	verboseFlag := cmd.PersistentFlags().Lookup("verbose")
	if verboseFlag == nil {
		t.Error("expected --verbose flag to be added")
	}
}

func TestFlagManager_IsVerbose(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
	}

	fm := NewFlagManager(config)

	if fm.IsVerbose() {
		t.Error("expected IsVerbose to be false initially")
	}

	fm.flags.Verbose = 1

	if !fm.IsVerbose() {
		t.Error("expected IsVerbose to be true after setting Verbose")
	}
}

func TestFlagManager_GetVerbosityLevel(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
	}

	fm := NewFlagManager(config)

	if fm.GetVerbosityLevel() != 0 {
		t.Error("expected verbosity level 0 initially")
	}

	fm.flags.Verbose = 3

	if fm.GetVerbosityLevel() != 3 {
		t.Errorf("expected verbosity level 3, got %d", fm.GetVerbosityLevel())
	}
}

func TestFlagManager_ApplyToConfig(t *testing.T) {
	baseConfig := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
	}

	fm := NewFlagManager(baseConfig)

	// Set some flags
	fm.flags.Output = "yaml"
	fm.flags.NoColor = true
	fm.flags.Timeout = 60 * time.Second
	fm.flags.Retry = 5
	fm.flags.NoCache = true

	targetConfig := &cli.Config{}
	fm.ApplyToConfig(targetConfig)

	// Verify output format was applied
	if targetConfig.Defaults == nil || targetConfig.Defaults.Output == nil {
		t.Fatal("expected defaults.output to be set")
	}

	if targetConfig.Defaults.Output.Format != "yaml" {
		t.Errorf("expected output format 'yaml', got %q", targetConfig.Defaults.Output.Format)
	}

	// Verify no-color was applied
	if targetConfig.Defaults.Output.Color != "never" {
		t.Errorf("expected color 'never', got %q", targetConfig.Defaults.Output.Color)
	}

	// Verify HTTP settings
	if targetConfig.Defaults.HTTP == nil {
		t.Fatal("expected defaults.http to be set")
	}

	if targetConfig.Defaults.HTTP.Timeout != "1m0s" {
		t.Errorf("expected timeout '1m0s', got %q", targetConfig.Defaults.HTTP.Timeout)
	}

	// Verify retry
	if targetConfig.Defaults.Retry == nil {
		t.Fatal("expected defaults.retry to be set")
	}

	if targetConfig.Defaults.Retry.MaxAttempts != 5 {
		t.Errorf("expected retry 5, got %d", targetConfig.Defaults.Retry.MaxAttempts)
	}

	// Verify caching
	if targetConfig.Defaults.Caching == nil {
		t.Fatal("expected defaults.caching to be set")
	}

	if targetConfig.Defaults.Caching.Enabled {
		t.Error("expected caching to be disabled")
	}
}

func TestFlagManager_GetOutputFormat(t *testing.T) {
	config := &cli.Config{
		Metadata: cli.Metadata{
			Name: "testcli",
		},
	}

	fm := NewFlagManager(config)

	if fm.GetOutputFormat() != "" {
		t.Error("expected empty output format initially")
	}

	fm.flags.Output = "json"

	if fm.GetOutputFormat() != "json" {
		t.Errorf("expected output format 'json', got %q", fm.GetOutputFormat())
	}
}
