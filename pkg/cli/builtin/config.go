package builtin

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ConfigOptions configures the config command behavior.
type ConfigOptions struct {
	CLIName      string
	AllowEdit    bool
	Output       io.Writer
	UserConfig   *cli.UserConfig
	SaveFunc     func(*cli.UserConfig) error
}

// NewConfigCommand creates a new config command group.
func NewConfigCommand(opts *ConfigOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage configuration",
		Long: `Manage user configuration settings.

Configuration is stored in the XDG-compliant config directory:
  ` + filepath.Join(xdg.ConfigHome, opts.CLIName, "config.yaml") + `

Available subcommands:
  show   - Display current configuration
  get    - Get a configuration value
  set    - Set a configuration value
  unset  - Unset a configuration value
  edit   - Edit configuration in $EDITOR
  path   - Show configuration file path`,
	}

	// Add subcommands
	cmd.AddCommand(newConfigShowCommand(opts))
	cmd.AddCommand(newConfigGetCommand(opts))
	cmd.AddCommand(newConfigSetCommand(opts))
	cmd.AddCommand(newConfigUnsetCommand(opts))
	cmd.AddCommand(newConfigPathCommand(opts))

	if opts.AllowEdit {
		cmd.AddCommand(newConfigEditCommand(opts))
	}

	return cmd
}

// newConfigShowCommand creates the config show subcommand.
func newConfigShowCommand(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Show current configuration",
		Long:  "Display the current user configuration in YAML format.",
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := yaml.Marshal(opts.UserConfig)
			if err != nil {
				return fmt.Errorf("failed to marshal config: %w", err)
			}

			fmt.Fprint(opts.Output, string(data))
			return nil
		},
	}
}

// newConfigGetCommand creates the config get subcommand.
func newConfigGetCommand(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Get a configuration value",
		Long: `Get a configuration value by key.

Examples:
  config get output.format
  config get defaults.http.timeout
  config get preferences.region`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value, err := getConfigValue(opts.UserConfig, key)
			if err != nil {
				return err
			}

			fmt.Fprintln(opts.Output, value)
			return nil
		},
	}
}

// newConfigSetCommand creates the config set subcommand.
func newConfigSetCommand(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Set a configuration value",
		Long: `Set a configuration value by key.

Examples:
  config set output.format json
  config set defaults.http.timeout 30s
  config set preferences.region us-west-2`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]
			value := args[1]

			if err := setConfigValue(opts.UserConfig, key, value); err != nil {
				return err
			}

			if err := opts.SaveFunc(opts.UserConfig); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(opts.Output, "Set %s = %s\n", key, value)
			return nil
		},
	}
}

// newConfigUnsetCommand creates the config unset subcommand.
func newConfigUnsetCommand(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "unset <key>",
		Short: "Unset a configuration value",
		Long: `Unset a configuration value by key.

Examples:
  config unset output.format
  config unset defaults.http.timeout`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			key := args[0]

			if err := unsetConfigValue(opts.UserConfig, key); err != nil {
				return err
			}

			if err := opts.SaveFunc(opts.UserConfig); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			fmt.Fprintf(opts.Output, "Unset %s\n", key)
			return nil
		},
	}
}

// newConfigEditCommand creates the config edit subcommand.
func newConfigEditCommand(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Edit configuration in $EDITOR",
		Long:  "Open the configuration file in the default editor ($EDITOR).",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(xdg.ConfigHome, opts.CLIName, "config.yaml")

			// Ensure config file exists
			if _, err := os.Stat(configPath); os.IsNotExist(err) {
				// Create empty config
				if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
					return fmt.Errorf("failed to create config directory: %w", err)
				}
				if err := os.WriteFile(configPath, []byte("# User configuration\n"), 0600); err != nil {
					return fmt.Errorf("failed to create config file: %w", err)
				}
			}

			// Get editor
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi" // Default to vi
			}

			// Open editor
			editorCmd := exec.Command(editor, configPath)
			editorCmd.Stdin = os.Stdin
			editorCmd.Stdout = os.Stdout
			editorCmd.Stderr = os.Stderr

			if err := editorCmd.Run(); err != nil {
				return fmt.Errorf("failed to edit config: %w", err)
			}

			fmt.Fprintf(opts.Output, "Configuration updated: %s\n", configPath)
			return nil
		},
	}
}

// newConfigPathCommand creates the config path subcommand.
func newConfigPathCommand(opts *ConfigOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show configuration file path",
		Long:  "Display the path to the configuration file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := filepath.Join(xdg.ConfigHome, opts.CLIName, "config.yaml")
			fmt.Fprintln(opts.Output, configPath)
			return nil
		},
	}
}

// getConfigValue retrieves a value from the config by dot-notation key.
func getConfigValue(config *cli.UserConfig, key string) (string, error) {
	parts := strings.Split(key, ".")

	// Handle preferences
	if parts[0] == "preferences" {
		if config.Preferences == nil {
			return "", fmt.Errorf("preferences not configured")
		}

		if len(parts) == 1 {
			// Return all preferences as YAML
			data, err := yaml.Marshal(config.Preferences)
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(data)), nil
		}

		// For simplicity, serialize to map and navigate
		var prefsMap map[string]interface{}
		data, _ := yaml.Marshal(config.Preferences)
		if err := yaml.Unmarshal(data, &prefsMap); err != nil {
			return "", err
		}

		return getNestedValue(prefsMap, parts[1:])
	}

	return "", fmt.Errorf("key not found: %s", key)
}

// getNestedValue retrieves a nested value from a map.
func getNestedValue(data interface{}, keys []string) (string, error) {
	if len(keys) == 0 {
		return fmt.Sprintf("%v", data), nil
	}

	switch v := data.(type) {
	case map[string]interface{}:
		if val, ok := v[keys[0]]; ok {
			if len(keys) == 1 {
				return fmt.Sprintf("%v", val), nil
			}
			return getNestedValue(val, keys[1:])
		}
	case map[interface{}]interface{}:
		if val, ok := v[keys[0]]; ok {
			if len(keys) == 1 {
				return fmt.Sprintf("%v", val), nil
			}
			return getNestedValue(val, keys[1:])
		}
	}

	return "", fmt.Errorf("key not found: %s", keys[0])
}

// setConfigValue sets a value in the config by dot-notation key.
func setConfigValue(config *cli.UserConfig, key, value string) error {
	parts := strings.Split(key, ".")

	// Handle preferences
	if parts[0] == "preferences" {
		if config.Preferences == nil {
			config.Preferences = &cli.UserPreferences{}
		}

		if len(parts) == 1 {
			return fmt.Errorf("cannot set entire preferences section, use a specific key")
		}

		// Convert to map, modify, then convert back
		var prefsMap map[string]interface{}
		data, _ := yaml.Marshal(config.Preferences)
		if err := yaml.Unmarshal(data, &prefsMap); err != nil {
			prefsMap = make(map[string]interface{})
		}

		if err := setNestedValue(prefsMap, parts[1:], value); err != nil {
			return err
		}

		// Convert back to UserPreferences
		data, _ = yaml.Marshal(prefsMap)
		return yaml.Unmarshal(data, config.Preferences)
	}

	return fmt.Errorf("invalid key: %s", key)
}

// setNestedValue sets a nested value in a map.
func setNestedValue(data map[string]interface{}, keys []string, value string) error {
	if len(keys) == 1 {
		// Try to parse value as appropriate type
		if value == "true" {
			data[keys[0]] = true
		} else if value == "false" {
			data[keys[0]] = false
		} else {
			data[keys[0]] = value
		}
		return nil
	}

	// Navigate to the parent
	if _, ok := data[keys[0]]; !ok {
		data[keys[0]] = make(map[string]interface{})
	}

	nested, ok := data[keys[0]].(map[string]interface{})
	if !ok {
		nested = make(map[string]interface{})
		data[keys[0]] = nested
	}

	return setNestedValue(nested, keys[1:], value)
}

// unsetConfigValue removes a value from the config by dot-notation key.
func unsetConfigValue(config *cli.UserConfig, key string) error {
	parts := strings.Split(key, ".")

	// Handle preferences
	if parts[0] == "preferences" {
		if config.Preferences == nil {
			return nil // Already unset
		}

		if len(parts) == 1 {
			config.Preferences = nil
			return nil
		}

		// Convert to map, modify, then convert back
		var prefsMap map[string]interface{}
		data, _ := yaml.Marshal(config.Preferences)
		if err := yaml.Unmarshal(data, &prefsMap); err != nil {
			return nil // Already empty
		}

		if err := unsetNestedValue(prefsMap, parts[1:]); err != nil {
			return err
		}

		// Convert back to UserPreferences
		data, _ = yaml.Marshal(prefsMap)
		return yaml.Unmarshal(data, config.Preferences)
	}

	return fmt.Errorf("invalid key: %s", key)
}

// unsetNestedValue removes a nested value from a map.
func unsetNestedValue(data map[string]interface{}, keys []string) error {
	if len(keys) == 1 {
		delete(data, keys[0])
		return nil
	}

	nested, ok := data[keys[0]].(map[string]interface{})
	if !ok {
		return nil // Already unset
	}

	return unsetNestedValue(nested, keys[1:])
}

// ListConfigKeys returns all available config keys.
func ListConfigKeys(config *cli.UserConfig) []string {
	keys := []string{}

	if config.Preferences != nil {
		collectKeys("preferences", config.Preferences, &keys)
	}

	sort.Strings(keys)
	return keys
}

// collectKeys recursively collects all keys from a nested map.
func collectKeys(prefix string, data interface{}, keys *[]string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			fullKey := prefix + "." + key
			*keys = append(*keys, fullKey)

			if nested, ok := value.(map[string]interface{}); ok {
				collectKeys(fullKey, nested, keys)
			}
		}
	case map[interface{}]interface{}:
		for key, value := range v {
			fullKey := prefix + "." + fmt.Sprint(key)
			*keys = append(*keys, fullKey)

			if nested, ok := value.(map[interface{}]interface{}); ok {
				collectKeys(fullKey, nested, keys)
			}
		}
	}
}
