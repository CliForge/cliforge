package builtin

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// GlobalFlags represents all global flags that can be used across commands.
type GlobalFlags struct {
	// Config and profile
	Config  string
	Profile string

	// Output formatting
	Output   string
	Verbose  int
	Quiet    bool
	Debug    bool
	NoColor  bool

	// HTTP client settings
	Timeout time.Duration
	Retry   int
	NoCache bool

	// Automation
	Yes bool

	// Update behavior
	NoUpdateCheck bool
}

// FlagManager manages global flags for a CLI.
type FlagManager struct {
	config *cli.CLIConfig
	flags  *GlobalFlags
}

// NewFlagManager creates a new flag manager.
func NewFlagManager(config *cli.CLIConfig) *FlagManager {
	return &FlagManager{
		config: config,
		flags:  &GlobalFlags{},
	}
}

// AddGlobalFlags adds global flags to a command.
func (fm *FlagManager) AddGlobalFlags(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()

	// Config and profile flags
	fm.addConfigFlags(pf)

	// Output flags
	fm.addOutputFlags(pf)

	// HTTP flags
	fm.addHTTPFlags(pf)

	// Automation flags
	fm.addAutomationFlags(pf)

	// Update flags
	fm.addUpdateFlags(pf)

	// Bind environment variables
	fm.bindEnvVars()
}

// addConfigFlags adds configuration-related flags.
func (fm *FlagManager) addConfigFlags(pf *pflag.FlagSet) {
	if fm.config.Behaviors != nil && fm.config.Behaviors.GlobalFlags != nil {
		// Config flag
		if fm.config.Behaviors.GlobalFlags.Config != nil && fm.config.Behaviors.GlobalFlags.Config.Enabled {
			pf.StringVarP(&fm.flags.Config, "config", "c", "",
				"Path to config file")
		}

		// Profile flag
		if fm.config.Behaviors.GlobalFlags.Profile != nil && fm.config.Behaviors.GlobalFlags.Profile.Enabled {
			pf.StringVar(&fm.flags.Profile, "profile", "default",
				"Configuration profile to use")
		}
	}
}

// addOutputFlags adds output-related flags.
func (fm *FlagManager) addOutputFlags(pf *pflag.FlagSet) {
	defaultOutput := "json"
	if fm.config.Defaults != nil && fm.config.Defaults.Output != nil {
		defaultOutput = fm.config.Defaults.Output.Format
	}

	if fm.config.Behaviors != nil && fm.config.Behaviors.GlobalFlags != nil {
		// Output format flag
		if fm.config.Behaviors.GlobalFlags.Output != nil && fm.config.Behaviors.GlobalFlags.Output.Enabled {
			pf.StringVarP(&fm.flags.Output, "output", "o", defaultOutput,
				"Output format (json|yaml|table|csv|text)")
		}

		// Verbose flag (can be repeated: -v, -vv, -vvv)
		if fm.config.Behaviors.GlobalFlags.Verbose != nil && fm.config.Behaviors.GlobalFlags.Verbose.Enabled {
			pf.CountVarP(&fm.flags.Verbose, "verbose", "v",
				"Verbose output (can be repeated)")
		}

		// Quiet flag
		if fm.config.Behaviors.GlobalFlags.Quiet != nil && fm.config.Behaviors.GlobalFlags.Quiet.Enabled {
			pf.BoolVarP(&fm.flags.Quiet, "quiet", "q", false,
				"Suppress non-error output")
		}

		// Debug flag
		if fm.config.Behaviors.GlobalFlags.Debug != nil && fm.config.Behaviors.GlobalFlags.Debug.Enabled {
			pf.BoolVar(&fm.flags.Debug, "debug", false,
				"Enable debug logging")
		}

		// No color flag
		if fm.config.Behaviors.GlobalFlags.NoColor != nil && fm.config.Behaviors.GlobalFlags.NoColor.Enabled {
			pf.BoolVar(&fm.flags.NoColor, "no-color", false,
				"Disable colored output")
		}
	}
}

// addHTTPFlags adds HTTP client-related flags.
func (fm *FlagManager) addHTTPFlags(pf *pflag.FlagSet) {
	defaultTimeout := 30 * time.Second
	defaultRetry := 3

	if fm.config.Defaults != nil && fm.config.Defaults.HTTP != nil {
		if fm.config.Defaults.HTTP.Timeout != "" {
			if d, err := time.ParseDuration(fm.config.Defaults.HTTP.Timeout); err == nil {
				defaultTimeout = d
			}
		}
	}

	if fm.config.Defaults != nil && fm.config.Defaults.Retry != nil {
		defaultRetry = fm.config.Defaults.Retry.MaxAttempts
	}

	if fm.config.Behaviors != nil && fm.config.Behaviors.GlobalFlags != nil {
		// Timeout flag
		if fm.config.Behaviors.GlobalFlags.Timeout != nil && fm.config.Behaviors.GlobalFlags.Timeout.Enabled {
			pf.DurationVar(&fm.flags.Timeout, "timeout", defaultTimeout,
				"Request timeout (e.g., 30s, 1m)")
		}

		// Retry flag
		if fm.config.Behaviors.GlobalFlags.Retry != nil && fm.config.Behaviors.GlobalFlags.Retry.Enabled {
			pf.IntVar(&fm.flags.Retry, "retry", defaultRetry,
				"Number of retry attempts")
		}

		// No cache flag
		if fm.config.Behaviors.GlobalFlags.NoCache != nil && fm.config.Behaviors.GlobalFlags.NoCache.Enabled {
			pf.BoolVar(&fm.flags.NoCache, "no-cache", false,
				"Disable response caching")
		}
	}
}

// addAutomationFlags adds automation-related flags.
func (fm *FlagManager) addAutomationFlags(pf *pflag.FlagSet) {
	if fm.config.Behaviors != nil && fm.config.Behaviors.GlobalFlags != nil {
		// Yes flag (skip confirmations)
		if fm.config.Behaviors.GlobalFlags.Yes != nil && fm.config.Behaviors.GlobalFlags.Yes.Enabled {
			pf.BoolVarP(&fm.flags.Yes, "yes", "y", false,
				"Skip confirmation prompts (non-interactive mode)")
		}
	}
}

// addUpdateFlags adds update-related flags.
func (fm *FlagManager) addUpdateFlags(pf *pflag.FlagSet) {
	if fm.config.Behaviors != nil && fm.config.Behaviors.GlobalFlags != nil {
		// No update check flag
		if fm.config.Behaviors.GlobalFlags.NoUpdateCheck != nil && fm.config.Behaviors.GlobalFlags.NoUpdateCheck.Enabled {
			pf.BoolVar(&fm.flags.NoUpdateCheck, "no-update-check", false,
				"Skip automatic update check")
		}
	}
}

// bindEnvVars binds environment variables to flags.
func (fm *FlagManager) bindEnvVars() {
	envPrefix := strings.ToUpper(strings.ReplaceAll(fm.config.Metadata.Name, "-", "_"))

	// Create environment variable mappings
	envMappings := map[string]*string{
		envPrefix + "_CONFIG":  &fm.flags.Config,
		envPrefix + "_PROFILE": &fm.flags.Profile,
		envPrefix + "_OUTPUT":  &fm.flags.Output,
	}

	// Apply environment variables if they exist
	for envVar, target := range envMappings {
		if val := os.Getenv(envVar); val != "" {
			*target = val
		}
	}

	// Handle NO_COLOR standard environment variable
	if os.Getenv("NO_COLOR") != "" {
		fm.flags.NoColor = true
	}
}

// GetFlags returns the global flags.
func (fm *FlagManager) GetFlags() *GlobalFlags {
	return fm.flags
}

// Validate validates the global flags.
func (fm *FlagManager) Validate() error {
	// Check for conflicting flags
	if fm.flags.Verbose > 0 && fm.flags.Quiet {
		return fmt.Errorf("--verbose and --quiet are mutually exclusive")
	}

	// Validate output format
	if fm.flags.Output != "" {
		validFormats := map[string]bool{
			"json":  true,
			"yaml":  true,
			"table": true,
			"csv":   true,
			"text":  true,
		}

		if !validFormats[fm.flags.Output] {
			return fmt.Errorf("invalid output format: %s (valid: json, yaml, table, csv, text)", fm.flags.Output)
		}
	}

	// Validate timeout
	if fm.flags.Timeout < 0 {
		return fmt.Errorf("timeout must be positive")
	}

	// Validate retry
	if fm.flags.Retry < 0 {
		return fmt.Errorf("retry count must be non-negative")
	}

	return nil
}

// ApplyToConfig applies global flags to the loaded configuration.
func (fm *FlagManager) ApplyToConfig(config *cli.CLIConfig) {
	// Apply output format
	if fm.flags.Output != "" {
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.Output == nil {
			config.Defaults.Output = &cli.DefaultsOutput{}
		}
		config.Defaults.Output.Format = fm.flags.Output
	}

	// Apply no-color flag
	if fm.flags.NoColor {
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.Output == nil {
			config.Defaults.Output = &cli.DefaultsOutput{}
		}
		config.Defaults.Output.Color = "never"
	}

	// Apply HTTP settings
	if fm.flags.Timeout > 0 {
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.HTTP == nil {
			config.Defaults.HTTP = &cli.DefaultsHTTP{}
		}
		config.Defaults.HTTP.Timeout = fm.flags.Timeout.String()
	}

	if fm.flags.Retry >= 0 {
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.Retry == nil {
			config.Defaults.Retry = &cli.DefaultsRetry{}
		}
		config.Defaults.Retry.MaxAttempts = fm.flags.Retry
	}

	// Apply caching
	if fm.flags.NoCache {
		if config.Defaults == nil {
			config.Defaults = &cli.Defaults{}
		}
		if config.Defaults.Caching == nil {
			config.Defaults.Caching = &cli.DefaultsCaching{}
		}
		config.Defaults.Caching.Enabled = false
	}
}

// IsVerbose returns true if verbose mode is enabled.
func (fm *FlagManager) IsVerbose() bool {
	return fm.flags.Verbose > 0
}

// GetVerbosityLevel returns the verbosity level (0, 1, 2, 3, ...).
func (fm *FlagManager) GetVerbosityLevel() int {
	return fm.flags.Verbose
}

// IsQuiet returns true if quiet mode is enabled.
func (fm *FlagManager) IsQuiet() bool {
	return fm.flags.Quiet
}

// IsDebug returns true if debug mode is enabled.
func (fm *FlagManager) IsDebug() bool {
	return fm.flags.Debug
}

// ShouldSkipConfirmation returns true if confirmations should be skipped.
func (fm *FlagManager) ShouldSkipConfirmation() bool {
	return fm.flags.Yes
}

// GetOutputFormat returns the output format.
func (fm *FlagManager) GetOutputFormat() string {
	return fm.flags.Output
}
