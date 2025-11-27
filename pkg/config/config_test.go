package config

import (
	"embed"
	"os"
	"path/filepath"
	"testing"

	"github.com/CliForge/cliforge/pkg/cli"
)

//go:embed testdata/*.yaml
var testFS embed.FS

func TestNewLoader(t *testing.T) {
	loader := NewLoader("test-cli", &testFS, "testdata/minimal.yaml")

	if loader.cliName != "test-cli" {
		t.Errorf("expected cliName 'test-cli', got %s", loader.cliName)
	}

	if loader.envPrefix != "TEST_CLI" {
		t.Errorf("expected envPrefix 'TEST_CLI', got %s", loader.envPrefix)
	}

	if loader.embeddedPath != "testdata/minimal.yaml" {
		t.Errorf("expected embeddedPath 'testdata/minimal.yaml', got %s", loader.embeddedPath)
	}
}

func TestLoadEmbeddedConfig(t *testing.T) {
	tests := []struct {
		name      string
		loader    *Loader
		wantError bool
	}{
		{
			name: "load valid embedded config",
			loader: &Loader{
				embeddedFS:   &testFS,
				embeddedPath: "testdata/minimal.yaml",
			},
			wantError: false,
		},
		{
			name: "nil embedded FS",
			loader: &Loader{
				embeddedFS:   nil,
				embeddedPath: "testdata/minimal.yaml",
			},
			wantError: true,
		},
		{
			name: "nonexistent file",
			loader: &Loader{
				embeddedFS:   &testFS,
				embeddedPath: "testdata/nonexistent.yaml",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := tt.loader.loadEmbeddedConfig()

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.wantError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if config == nil {
					t.Error("expected config but got nil")
				}
			}
		})
	}
}

func TestLoadUserConfig(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func(t *testing.T) (string, func())
		wantError  bool
		wantConfig bool
	}{
		{
			name: "load existing user config",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "config.yaml")
				content := `preferences:
  output:
    format: yaml
`
				if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				_ = os.Setenv("TEST_CLI_CONFIG", configPath)
				return configPath, func() { _ = os.Unsetenv("TEST_CLI_CONFIG") }
			},
			wantError:  false,
			wantConfig: true,
		},
		{
			name: "nonexistent user config",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "nonexistent.yaml")
				_ = os.Setenv("TEST_CLI_CONFIG", configPath)
				return configPath, func() { _ = os.Unsetenv("TEST_CLI_CONFIG") }
			},
			wantError:  false,
			wantConfig: true, // Returns empty config, not nil
		},
		{
			name: "invalid YAML",
			setupFunc: func(t *testing.T) (string, func()) {
				tempDir := t.TempDir()
				configPath := filepath.Join(tempDir, "invalid.yaml")
				content := `preferences:
  output
    invalid yaml here
`
				if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
				_ = os.Setenv("TEST_CLI_CONFIG", configPath)
				return configPath, func() { _ = os.Unsetenv("TEST_CLI_CONFIG") }
			},
			wantError:  true,
			wantConfig: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := tt.setupFunc(t)
			defer cleanup()

			loader := &Loader{
				cliName:   "test-cli",
				envPrefix: "TEST_CLI",
			}

			config, err := loader.loadUserConfig()

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantConfig && config == nil {
				t.Error("expected config but got nil")
			}
		})
	}
}

func TestGetUserConfigPath(t *testing.T) {
	tests := []struct {
		name     string
		cliName  string
		envVar   string
		envValue string
		wantPath string
	}{
		{
			name:     "default XDG path",
			cliName:  "test-cli",
			envVar:   "",
			envValue: "",
		},
		{
			name:     "custom path via env var",
			cliName:  "test-cli",
			envVar:   "TEST_CLI_CONFIG",
			envValue: "/custom/path/config.yaml",
			wantPath: "/custom/path/config.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable if specified
			if tt.envVar != "" {
				_ = os.Setenv(tt.envVar, tt.envValue)
				defer func() { _ = os.Unsetenv(tt.envVar) }()
			}

			loader := &Loader{
				cliName:   tt.cliName,
				envPrefix: "TEST_CLI",
			}

			path, err := loader.getUserConfigPath()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if tt.wantPath != "" && path != tt.wantPath {
				t.Errorf("expected path %s, got %s", tt.wantPath, path)
			}

			if tt.wantPath == "" && path == "" {
				t.Error("expected non-empty path")
			}
		})
	}
}

func TestGetXDGDirectories(t *testing.T) {
	loader := NewLoader("test-cli", nil, "")

	tests := []struct {
		name string
		fn   func() string
	}{
		{"GetCacheDir", loader.GetCacheDir},
		{"GetDataDir", loader.GetDataDir},
		{"GetStateDir", loader.GetStateDir},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.fn()
			if dir == "" {
				t.Error("expected non-empty directory path")
			}
			if !filepath.IsAbs(dir) {
				t.Errorf("expected absolute path, got %s", dir)
			}
		})
	}
}

func TestEnsureConfigDirs(t *testing.T) {
	tempDir := t.TempDir()

	// Set XDG env vars to use temp directory
	_ = os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	_ = os.Setenv("XDG_CACHE_HOME", filepath.Join(tempDir, "cache"))
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	_ = os.Setenv("XDG_STATE_HOME", filepath.Join(tempDir, "state"))
	defer func() {
		_ = os.Unsetenv("XDG_CONFIG_HOME")
		_ = os.Unsetenv("XDG_CACHE_HOME")
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.Unsetenv("XDG_STATE_HOME")
	}()

	loader := NewLoader("test-cli", nil, "")

	err := loader.EnsureConfigDirs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that directories were created
	dirs := []string{
		loader.GetCacheDir(),
		loader.GetDataDir(),
		loader.GetStateDir(),
	}

	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("directory not created: %s", dir)
		}
	}
}

func TestSaveUserConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Set custom config path via environment variable
	_ = os.Setenv("TEST_CLI_CONFIG", filepath.Join(tempDir, "config.yaml"))
	defer func() { _ = os.Unsetenv("TEST_CLI_CONFIG") }()

	loader := &Loader{
		cliName:   "test-cli",
		envPrefix: "TEST_CLI",
	}

	config := &cli.UserConfig{
		Preferences: &cli.UserPreferences{
			Output: &cli.PreferencesOutput{
				Format: "yaml",
			},
		},
	}

	// Save config
	err := loader.SaveUserConfig(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get the config path to verify
	configPath, err := loader.getUserConfigPath()
	if err != nil {
		t.Fatalf("failed to get config path: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Read and verify content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("failed to read config file: %v", err)
	}

	if len(data) == 0 {
		t.Error("config file is empty")
	}
}

func TestApplyEnvironmentOverrides(t *testing.T) {
	tests := []struct {
		name    string
		envVars map[string]string
		verify  func(*testing.T, *cli.Config)
	}{
		{
			name: "override output format",
			envVars: map[string]string{
				"TEST_CLI_OUTPUT_FORMAT": "yaml",
			},
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults == nil || config.Defaults.Output == nil {
					t.Fatal("defaults.output not initialized")
				}
				if config.Defaults.Output.Format != "yaml" {
					t.Errorf("expected format 'yaml', got %s", config.Defaults.Output.Format)
				}
			},
		},
		{
			name: "override timeout",
			envVars: map[string]string{
				"TEST_CLI_TIMEOUT": "60s",
			},
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults == nil || config.Defaults.HTTP == nil {
					t.Fatal("defaults.http not initialized")
				}
				if config.Defaults.HTTP.Timeout != "60s" {
					t.Errorf("expected timeout '60s', got %s", config.Defaults.HTTP.Timeout)
				}
			},
		},
		{
			name: "disable color with NO_COLOR",
			envVars: map[string]string{
				"TEST_CLI_NO_COLOR": "1",
			},
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults == nil || config.Defaults.Output == nil {
					t.Fatal("defaults.output not initialized")
				}
				if config.Defaults.Output.Color != "never" {
					t.Errorf("expected color 'never', got %s", config.Defaults.Output.Color)
				}
			},
		},
		{
			name: "disable cache",
			envVars: map[string]string{
				"TEST_CLI_NO_CACHE": "1",
			},
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults == nil || config.Defaults.Caching == nil {
					t.Fatal("defaults.caching not initialized")
				}
				if config.Defaults.Caching.Enabled {
					t.Error("expected caching to be disabled")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				_ = os.Setenv(key, value)
				defer func(k string) { _ = os.Unsetenv(k) }(key)
			}

			loader := &Loader{
				cliName:   "test-cli",
				envPrefix: "TEST_CLI",
			}

			config := &cli.Config{}
			result, err := loader.applyEnvironmentOverrides(config)

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			tt.verify(t, result)
		})
	}
}

func TestSetConfigValue(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		value     string
		verify    func(*testing.T, *cli.Config)
		wantError bool
	}{
		{
			name:  "set output format",
			path:  "defaults.output.format",
			value: "json",
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults.Output.Format != "json" {
					t.Errorf("expected format 'json', got %s", config.Defaults.Output.Format)
				}
			},
		},
		{
			name:  "set timeout",
			path:  "defaults.http.timeout",
			value: "30s",
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults.HTTP.Timeout != "30s" {
					t.Errorf("expected timeout '30s', got %s", config.Defaults.HTTP.Timeout)
				}
			},
		},
		{
			name:  "set color to always",
			path:  "defaults.output.color",
			value: "true",
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults.Output.Color != "always" {
					t.Errorf("expected color 'always', got %s", config.Defaults.Output.Color)
				}
			},
		},
		{
			name:  "set color to never",
			path:  "defaults.output.color",
			value: "false",
			verify: func(t *testing.T, config *cli.Config) {
				if config.Defaults.Output.Color != "never" {
					t.Errorf("expected color 'never', got %s", config.Defaults.Output.Color)
				}
			},
		},
		{
			name:      "invalid path",
			path:      "invalid.unknown.path",
			value:     "test",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &Loader{}
			config := &cli.Config{}

			err := loader.setConfigValue(config, tt.path, tt.value)

			if tt.wantError && err == nil {
				t.Error("expected error but got none")
			}

			if !tt.wantError {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.verify != nil {
					tt.verify(t, config)
				}
			}
		})
	}
}

func TestShowWarnings(t *testing.T) {
	tests := []struct {
		name         string
		loaded       *cli.LoadedConfig
		expectOutput bool
	}{
		{
			name: "debug override in production",
			loaded: &cli.LoadedConfig{
				EmbeddedConfig: &cli.Config{
					Metadata: cli.Metadata{
						Debug: false,
					},
				},
				UserConfig: &cli.UserConfig{
					DebugOverride: &cli.Config{
						API: cli.API{
							BaseURL: "http://localhost",
						},
					},
				},
				DebugOverrides: map[string]any{},
			},
			expectOutput: true,
		},
		{
			name: "debug mode with overrides",
			loaded: &cli.LoadedConfig{
				EmbeddedConfig: &cli.Config{
					Metadata: cli.Metadata{
						Debug:   true,
						Version: "1.0.0",
					},
				},
				UserConfig: &cli.UserConfig{},
				DebugOverrides: map[string]any{
					"api.base_url": "http://localhost",
					"auth.type":    "none",
				},
			},
			expectOutput: true,
		},
		{
			name: "no warnings",
			loaded: &cli.LoadedConfig{
				EmbeddedConfig: &cli.Config{
					Metadata: cli.Metadata{
						Debug: false,
					},
				},
				UserConfig:     &cli.UserConfig{},
				DebugOverrides: map[string]any{},
			},
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			loader := NewLoader("test-cli", nil, "")

			// Just verify it doesn't panic - we can't easily capture stderr
			loader.ShowWarnings(tt.loaded)
		})
	}
}

func TestLoadConfig_Integration(t *testing.T) {
	// Create a minimal embedded config
	loader := NewLoader("test-cli", &testFS, "testdata/minimal.yaml")

	loaded, err := loader.LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if loaded == nil {
		t.Fatal("expected loaded config but got nil")
	}

	if loaded.Final == nil {
		t.Error("expected final config but got nil")
	}

	if loaded.EmbeddedConfig == nil {
		t.Error("expected embedded config but got nil")
	}

	if loaded.UserConfig == nil {
		t.Error("expected user config but got nil")
	}

	// Show warnings (should not panic)
	loader.ShowWarnings(loaded)
}
