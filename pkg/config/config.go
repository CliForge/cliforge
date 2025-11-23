// Package config handles loading, merging, and validation of CLI configurations.
package config

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/adrg/xdg"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Loader handles loading configurations from various sources.
type Loader struct {
	cliName          string
	embeddedFS       *embed.FS
	embeddedPath     string
	userConfigPath   string
	envPrefix        string
}

// NewLoader creates a new configuration loader.
func NewLoader(cliName string, embeddedFS *embed.FS, embeddedPath string) *Loader {
	return &Loader{
		cliName:      cliName,
		embeddedFS:   embeddedFS,
		embeddedPath: embeddedPath,
		envPrefix:    strings.ToUpper(strings.ReplaceAll(cliName, "-", "_")),
	}
}

// LoadConfig loads and merges configuration from all sources.
// Priority: ENV > Flag > User Config > Debug Override > Embedded > Default
func (l *Loader) LoadConfig() (*cli.LoadedConfig, error) {
	// Load embedded configuration
	embeddedConfig, err := l.loadEmbeddedConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load embedded config: %w", err)
	}

	// Load user configuration
	userConfig, err := l.loadUserConfig()
	if err != nil {
		// User config is optional, log warning but continue
		fmt.Fprintf(os.Stderr, "Warning: failed to load user config: %v\n", err)
		userConfig = &cli.UserConfig{}
	}

	// Create merged configuration
	merged, debugOverrides, err := l.mergeConfigs(embeddedConfig, userConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to merge configs: %w", err)
	}

	// Apply environment variables and flags
	merged, err = l.applyEnvironmentOverrides(merged)
	if err != nil {
		return nil, fmt.Errorf("failed to apply environment overrides: %w", err)
	}

	return &cli.LoadedConfig{
		Final:          merged,
		EmbeddedConfig: embeddedConfig,
		UserConfig:     userConfig,
		DebugOverrides: debugOverrides,
	}, nil
}

// loadEmbeddedConfig loads the configuration embedded in the binary.
func (l *Loader) loadEmbeddedConfig() (*cli.CLIConfig, error) {
	if l.embeddedFS == nil {
		return nil, fmt.Errorf("no embedded filesystem provided")
	}

	data, err := l.embeddedFS.ReadFile(l.embeddedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded config: %w", err)
	}

	var config cli.CLIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse embedded config: %w", err)
	}

	return &config, nil
}

// loadUserConfig loads user-specific configuration from XDG-compliant paths.
func (l *Loader) loadUserConfig() (*cli.UserConfig, error) {
	// Get XDG config path
	configPath, err := l.getUserConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// User config is optional
		return &cli.UserConfig{}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read user config: %w", err)
	}

	var config cli.UserConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse user config: %w", err)
	}

	return &config, nil
}

// getUserConfigPath returns the XDG-compliant user config file path.
func (l *Loader) getUserConfigPath() (string, error) {
	// If user specified custom path via environment
	if customPath := os.Getenv(l.envPrefix + "_CONFIG"); customPath != "" {
		return customPath, nil
	}

	// Use XDG config directory
	configDir := filepath.Join(xdg.ConfigHome, l.cliName)
	return filepath.Join(configDir, "config.yaml"), nil
}

// GetCacheDir returns the XDG-compliant cache directory.
func (l *Loader) GetCacheDir() string {
	return filepath.Join(xdg.CacheHome, l.cliName)
}

// GetDataDir returns the XDG-compliant data directory.
func (l *Loader) GetDataDir() string {
	return filepath.Join(xdg.DataHome, l.cliName)
}

// GetStateDir returns the XDG-compliant state directory.
func (l *Loader) GetStateDir() string {
	return filepath.Join(xdg.StateHome, l.cliName)
}

// applyEnvironmentOverrides applies environment variable overrides.
func (l *Loader) applyEnvironmentOverrides(config *cli.CLIConfig) (*cli.CLIConfig, error) {
	v := viper.New()
	v.SetEnvPrefix(l.envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Map common environment variables to config paths
	envMappings := map[string]string{
		"OUTPUT_FORMAT":              "defaults.output.format",
		"TIMEOUT":                    "defaults.http.timeout",
		"NO_COLOR":                   "defaults.output.color",
		"PRETTY_PRINT":               "defaults.output.pretty_print",
		"PAGING":                     "defaults.output.paging",
		"PAGE_LIMIT":                 "defaults.pagination.limit",
		"RETRY":                      "defaults.retry.max_attempts",
		"NO_CACHE":                   "defaults.caching.enabled",
		"DEPRECATIONS_ALWAYS_SHOW":   "defaults.deprecations.always_show",
		"DEPRECATIONS_MIN_SEVERITY":  "defaults.deprecations.min_severity",
	}

	// Apply environment variable overrides
	for envKey, configPath := range envMappings {
		fullEnvKey := l.envPrefix + "_" + envKey
		if val := os.Getenv(fullEnvKey); val != "" {
			// Parse the value and set in config
			if err := l.setConfigValue(config, configPath, val); err != nil {
				return nil, fmt.Errorf("failed to apply env var %s: %w", fullEnvKey, err)
			}
		}
	}

	return config, nil
}

// setConfigValue sets a value in the config using a dot-notation path.
func (l *Loader) setConfigValue(config *cli.CLIConfig, path, value string) error {
	// Handle specific known paths
	switch {
	case strings.HasPrefix(path, "defaults.output.format"):
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.Output == nil {
			config.Defaults.Output = &cli.DefaultsOutput{}
		}
		config.Defaults.Output.Format = value

	case strings.HasPrefix(path, "defaults.http.timeout"):
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.HTTP == nil {
			config.Defaults.HTTP = &cli.DefaultsHTTP{}
		}
		config.Defaults.HTTP.Timeout = value

	case strings.HasPrefix(path, "defaults.output.color"):
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.Output == nil {
			config.Defaults.Output = &cli.DefaultsOutput{}
		}
		if value == "1" || value == "true" {
			config.Defaults.Output.Color = "always"
		} else {
			config.Defaults.Output.Color = "never"
		}

	case strings.HasPrefix(path, "defaults.caching.enabled"):
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.Caching == nil {
			config.Defaults.Caching = &cli.DefaultsCaching{}
		}
		config.Defaults.Caching.Enabled = value != "1" && value != "true"

	default:
		return fmt.Errorf("unknown config path: %s", path)
	}

	return nil
}

// EnsureConfigDirs creates all necessary XDG directories.
func (l *Loader) EnsureConfigDirs() error {
	dirs := []string{
		filepath.Join(xdg.ConfigHome, l.cliName),
		filepath.Join(xdg.CacheHome, l.cliName),
		filepath.Join(xdg.DataHome, l.cliName),
		filepath.Join(xdg.StateHome, l.cliName),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// SaveUserConfig saves user configuration to the XDG config directory.
func (l *Loader) SaveUserConfig(config *cli.UserConfig) error {
	configPath, err := l.getUserConfigPath()
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// ShowWarnings displays configuration warnings (e.g., debug_override in production).
func (l *Loader) ShowWarnings(loaded *cli.LoadedConfig) {
	// Warn if debug_override is present but debug mode is disabled
	if loaded.UserConfig.DebugOverride != nil && !loaded.EmbeddedConfig.Metadata.Debug {
		fmt.Fprintf(os.Stderr, "\n⚠️  Warning: debug_override section found in config file but ignored\n")
		fmt.Fprintf(os.Stderr, "   This is a production build (debug: false)\n")
		fmt.Fprintf(os.Stderr, "   debug_override section is only active in debug builds\n")

		configPath, _ := l.getUserConfigPath()
		fmt.Fprintf(os.Stderr, "   Location: %s\n\n", configPath)
	}

	// Show debug warnings if in debug mode
	if loaded.EmbeddedConfig.Metadata.Debug && len(loaded.DebugOverrides) > 0 {
		l.showDebugWarning(loaded)
	}
}

// showDebugWarning displays a prominent warning when debug mode is active.
func (l *Loader) showDebugWarning(loaded *cli.LoadedConfig) {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "╔════════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "║  🚨 DEBUG MODE ENABLED - SECURITY WARNING                     ║")
	fmt.Fprintln(os.Stderr, "╠════════════════════════════════════════════════════════════════╣")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintln(os.Stderr, "║  This is a DEBUG BUILD.                                        ║")
	fmt.Fprintln(os.Stderr, "║  All embedded configuration can be overridden.                 ║")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintln(os.Stderr, "║  ⚠️  DO NOT USE IN PRODUCTION                                 ║")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintf(os.Stderr, "║  Build info:                                                   ║\n")
	fmt.Fprintf(os.Stderr, "║  - Version: %-50s║\n", loaded.EmbeddedConfig.Metadata.Version)
	fmt.Fprintf(os.Stderr, "║  - Debug: ENABLED                                              ║\n")
	fmt.Fprintf(os.Stderr, "║  - Config overrides: ALLOWED                                   ║\n")
	fmt.Fprintln(os.Stderr, "║                                                                ║")

	if len(loaded.DebugOverrides) > 0 {
		fmt.Fprintf(os.Stderr, "║  Active debug_override settings (%d):                         ║\n", len(loaded.DebugOverrides))
		count := 0
		for key := range loaded.DebugOverrides {
			if count < 3 {
				fmt.Fprintf(os.Stderr, "║  - %-60s║\n", key)
				count++
			}
		}
		if len(loaded.DebugOverrides) > 3 {
			fmt.Fprintf(os.Stderr, "║  ... and %d more                                              ║\n", len(loaded.DebugOverrides)-3)
		}
		fmt.Fprintln(os.Stderr, "║                                                                ║")
	}

	fmt.Fprintln(os.Stderr, "╚════════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")
}
