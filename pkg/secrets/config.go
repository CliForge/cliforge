package secrets

import (
	"fmt"
	"os"

	"github.com/CliForge/cliforge/pkg/cli"
)

// Config wraps secrets behavior configuration with additional runtime settings.
type Config struct {
	// Behavior is the core secrets configuration
	Behavior *cli.SecretsBehavior

	// DisableMasking can be set via flag/env to temporarily disable masking (dangerous!)
	DisableMasking bool

	// WarnOnDisable controls whether to show warnings when masking is disabled
	WarnOnDisable bool
}

// NewConfig creates a new secrets configuration with sensible defaults.
func NewConfig() *Config {
	return &Config{
		Behavior:       DefaultSecretsBehavior(),
		DisableMasking: false,
		WarnOnDisable:  true,
	}
}

// NewConfigFromBehavior creates a config from a secrets behavior.
func NewConfigFromBehavior(behavior *cli.SecretsBehavior) *Config {
	if behavior == nil {
		return NewConfig()
	}

	return &Config{
		Behavior:       behavior,
		DisableMasking: false,
		WarnOnDisable:  true,
	}
}

// IsEnabled returns whether secret masking is enabled.
func (c *Config) IsEnabled() bool {
	if c.DisableMasking {
		return false
	}
	return c.Behavior != nil && c.Behavior.Enabled
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.Behavior == nil {
		return nil
	}

	return ValidateSecretsBehavior(c.Behavior)
}

// ApplyEnvironmentOverrides applies environment variable overrides to the config.
// Supports: {PREFIX}_NO_MASK_SECRETS=true to disable masking.
func (c *Config) ApplyEnvironmentOverrides(envPrefix string) {
	// Check for NO_MASK_SECRETS environment variable
	if envVal := os.Getenv(envPrefix + "_NO_MASK_SECRETS"); envVal != "" {
		if envVal == "1" || envVal == "true" || envVal == "TRUE" {
			c.DisableMasking = true

			// Show warning if enabled
			if c.WarnOnDisable {
				ShowMaskingDisabledWarning()
			}
		}
	}
}

// ShowMaskingDisabledWarning displays a prominent warning when masking is disabled.
func ShowMaskingDisabledWarning() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "╔════════════════════════════════════════════════════════════════╗")
	fmt.Fprintln(os.Stderr, "║  ⚠️  WARNING: SECRET MASKING DISABLED                        ║")
	fmt.Fprintln(os.Stderr, "╠════════════════════════════════════════════════════════════════╣")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintln(os.Stderr, "║  Sensitive data may be exposed in output!                     ║")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintln(os.Stderr, "║  This includes:                                                ║")
	fmt.Fprintln(os.Stderr, "║  - API keys and tokens                                         ║")
	fmt.Fprintln(os.Stderr, "║  - Passwords and secrets                                       ║")
	fmt.Fprintln(os.Stderr, "║  - Authentication headers                                      ║")
	fmt.Fprintln(os.Stderr, "║  - Personal information                                        ║")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintln(os.Stderr, "║  ⚠️  DO NOT USE IN PRODUCTION                                ║")
	fmt.Fprintln(os.Stderr, "║  ⚠️  DO NOT SHARE OUTPUT FROM THIS SESSION                   ║")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintln(os.Stderr, "║  To re-enable masking, remove the --no-mask-secrets flag      ║")
	fmt.Fprintln(os.Stderr, "║  or unset the NO_MASK_SECRETS environment variable.           ║")
	fmt.Fprintln(os.Stderr, "║                                                                ║")
	fmt.Fprintln(os.Stderr, "╚════════════════════════════════════════════════════════════════╝")
	fmt.Fprintln(os.Stderr, "")
}

// ConfigBuilder provides a fluent interface for building secrets configurations.
type ConfigBuilder struct {
	config *Config
}

// NewConfigBuilder creates a new ConfigBuilder with default settings.
func NewConfigBuilder() *ConfigBuilder {
	return &ConfigBuilder{
		config: NewConfig(),
	}
}

// WithBehavior sets the secrets behavior.
func (b *ConfigBuilder) WithBehavior(behavior *cli.SecretsBehavior) *ConfigBuilder {
	b.config.Behavior = behavior
	return b
}

// WithEnabled sets whether secret detection is enabled.
func (b *ConfigBuilder) WithEnabled(enabled bool) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.Enabled = enabled
	return b
}

// WithMaskingStyle sets the masking style.
func (b *ConfigBuilder) WithMaskingStyle(style string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	if b.config.Behavior.Masking == nil {
		b.config.Behavior.Masking = &cli.SecretsMasking{}
	}
	b.config.Behavior.Masking.Style = style
	return b
}

// WithPartialShowChars sets the number of characters to show in partial masking.
func (b *ConfigBuilder) WithPartialShowChars(chars int) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	if b.config.Behavior.Masking == nil {
		b.config.Behavior.Masking = &cli.SecretsMasking{}
	}
	b.config.Behavior.Masking.PartialShowChars = chars
	return b
}

// WithReplacement sets the replacement string for masking.
func (b *ConfigBuilder) WithReplacement(replacement string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	if b.config.Behavior.Masking == nil {
		b.config.Behavior.Masking = &cli.SecretsMasking{}
	}
	b.config.Behavior.Masking.Replacement = replacement
	return b
}

// WithFieldPatterns sets the field name patterns.
func (b *ConfigBuilder) WithFieldPatterns(patterns []string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.FieldPatterns = patterns
	return b
}

// AddFieldPattern adds a field name pattern.
func (b *ConfigBuilder) AddFieldPattern(pattern string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.FieldPatterns = append(b.config.Behavior.FieldPatterns, pattern)
	return b
}

// WithValuePatterns sets the value patterns.
func (b *ConfigBuilder) WithValuePatterns(patterns []cli.ValuePattern) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.ValuePatterns = patterns
	return b
}

// AddValuePattern adds a value pattern.
func (b *ConfigBuilder) AddValuePattern(pattern cli.ValuePattern) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.ValuePatterns = append(b.config.Behavior.ValuePatterns, pattern)
	return b
}

// WithHeaders sets the headers to mask.
func (b *ConfigBuilder) WithHeaders(headers []string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.Headers = headers
	return b
}

// AddHeader adds a header to mask.
func (b *ConfigBuilder) AddHeader(header string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.Headers = append(b.config.Behavior.Headers, header)
	return b
}

// WithExplicitFields sets the explicit fields to mask.
func (b *ConfigBuilder) WithExplicitFields(fields []string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.ExplicitFields = fields
	return b
}

// AddExplicitField adds an explicit field path to mask.
func (b *ConfigBuilder) AddExplicitField(field string) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.ExplicitFields = append(b.config.Behavior.ExplicitFields, field)
	return b
}

// WithMaskIn sets where to apply masking.
func (b *ConfigBuilder) WithMaskIn(maskIn *cli.SecretsMaskIn) *ConfigBuilder {
	if b.config.Behavior == nil {
		b.config.Behavior = DefaultSecretsBehavior()
	}
	b.config.Behavior.MaskIn = maskIn
	return b
}

// WithDisableMasking sets whether to disable masking (for debugging).
func (b *ConfigBuilder) WithDisableMasking(disable bool) *ConfigBuilder {
	b.config.DisableMasking = disable
	return b
}

// WithWarnOnDisable sets whether to warn when masking is disabled.
func (b *ConfigBuilder) WithWarnOnDisable(warn bool) *ConfigBuilder {
	b.config.WarnOnDisable = warn
	return b
}

// Build builds and validates the configuration.
func (b *ConfigBuilder) Build() (*Config, error) {
	if err := b.config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid secrets configuration: %w", err)
	}
	return b.config, nil
}

// MustBuild builds the configuration and panics on error.
func (b *ConfigBuilder) MustBuild() *Config {
	config, err := b.Build()
	if err != nil {
		panic(err)
	}
	return config
}

// LoadConfigFromCLIConfig loads secrets config from a CLI configuration.
func LoadConfigFromCLIConfig(cliConfig *cli.CLIConfig) *Config {
	if cliConfig == nil || cliConfig.Behaviors == nil || cliConfig.Behaviors.Secrets == nil {
		return NewConfig()
	}

	return NewConfigFromBehavior(cliConfig.Behaviors.Secrets)
}

// GetMaskingStrategy returns the masking strategy from config.
func (c *Config) GetMaskingStrategy() MaskStrategy {
	if c.Behavior == nil || c.Behavior.Masking == nil {
		return NewPartialMaskStrategy(6, "***")
	}

	return CreateMaskStrategy(c.Behavior.Masking)
}
