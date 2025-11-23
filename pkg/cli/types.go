// Package cli defines the core types for CliForge CLI generation system.
package cli

import "time"

// CLIConfig represents the complete CLI configuration as defined in cli-config.yaml.
// This is the root configuration structure that encompasses all aspects of a CLI tool.
type CLIConfig struct {
	Metadata  Metadata   `yaml:"metadata" json:"metadata"`
	Branding  *Branding  `yaml:"branding,omitempty" json:"branding,omitempty"`
	API       API        `yaml:"api" json:"api"`
	Defaults  *Defaults  `yaml:"defaults,omitempty" json:"defaults,omitempty"`
	Updates   *Updates   `yaml:"updates,omitempty" json:"updates,omitempty"`
	Behaviors *Behaviors `yaml:"behaviors,omitempty" json:"behaviors,omitempty"`
	Features  *Features  `yaml:"features,omitempty" json:"features,omitempty"`
}

// Metadata contains identifying information about the CLI tool.
type Metadata struct {
	Name            string  `yaml:"name" json:"name"`
	Version         string  `yaml:"version" json:"version"`
	Description     string  `yaml:"description" json:"description"`
	LongDescription string  `yaml:"long_description,omitempty" json:"long_description,omitempty"`
	Author          *Author `yaml:"author,omitempty" json:"author,omitempty"`
	License         string  `yaml:"license,omitempty" json:"license,omitempty"`
	Homepage        string  `yaml:"homepage,omitempty" json:"homepage,omitempty"`
	BugsURL         string  `yaml:"bugs_url,omitempty" json:"bugs_url,omitempty"`
	DocsURL         string  `yaml:"docs_url,omitempty" json:"docs_url,omitempty"`
	Debug           bool    `yaml:"debug,omitempty" json:"debug,omitempty"`
}

// Author represents the author information.
type Author struct {
	Name  string `yaml:"name,omitempty" json:"name,omitempty"`
	Email string `yaml:"email,omitempty" json:"email,omitempty"`
	URL   string `yaml:"url,omitempty" json:"url,omitempty"`
}

// Branding defines the visual appearance and style of the CLI.
type Branding struct {
	Colors   *Colors  `yaml:"colors,omitempty" json:"colors,omitempty"`
	ASCIIArt string   `yaml:"ascii_art,omitempty" json:"ascii_art,omitempty"`
	Prompts  *Prompts `yaml:"prompts,omitempty" json:"prompts,omitempty"`
	Theme    *Theme   `yaml:"theme,omitempty" json:"theme,omitempty"`
}

// Colors defines the color scheme for CLI output.
type Colors struct {
	Primary   string `yaml:"primary,omitempty" json:"primary,omitempty"`
	Secondary string `yaml:"secondary,omitempty" json:"secondary,omitempty"`
	Success   string `yaml:"success,omitempty" json:"success,omitempty"`
	Warning   string `yaml:"warning,omitempty" json:"warning,omitempty"`
	Error     string `yaml:"error,omitempty" json:"error,omitempty"`
	Info      string `yaml:"info,omitempty" json:"info,omitempty"`
}

// Prompts defines custom prompt symbols.
type Prompts struct {
	Command string `yaml:"command,omitempty" json:"command,omitempty"`
	Error   string `yaml:"error,omitempty" json:"error,omitempty"`
	Success string `yaml:"success,omitempty" json:"success,omitempty"`
	Warning string `yaml:"warning,omitempty" json:"warning,omitempty"`
	Info    string `yaml:"info,omitempty" json:"info,omitempty"`
}

// Theme defines the overall theme settings.
type Theme struct {
	Name               string `yaml:"name,omitempty" json:"name,omitempty"` // auto, light, dark
	SyntaxHighlighting bool   `yaml:"syntax_highlighting,omitempty" json:"syntax_highlighting,omitempty"`
}

// API defines the API endpoints and connection settings.
type API struct {
	OpenAPIURL     string              `yaml:"openapi_url" json:"openapi_url"`
	BaseURL        string              `yaml:"base_url" json:"base_url"`
	Version        string              `yaml:"version,omitempty" json:"version,omitempty"`
	Environments   []Environment       `yaml:"environments,omitempty" json:"environments,omitempty"`
	DefaultHeaders map[string]string   `yaml:"default_headers,omitempty" json:"default_headers,omitempty"`
	UserAgent      string              `yaml:"user_agent,omitempty" json:"user_agent,omitempty"`
	TelemetryURL   string              `yaml:"telemetry_url,omitempty" json:"telemetry_url,omitempty"`
}

// Environment represents a multi-environment configuration.
type Environment struct {
	Name        string `yaml:"name" json:"name"`
	OpenAPIURL  string `yaml:"openapi_url" json:"openapi_url"`
	BaseURL     string `yaml:"base_url" json:"base_url"`
	Default     bool   `yaml:"default,omitempty" json:"default,omitempty"`
}

// Defaults contains user-overridable default settings.
type Defaults struct {
	HTTP         *DefaultsHTTP         `yaml:"http,omitempty" json:"http,omitempty"`
	Caching      *DefaultsCaching      `yaml:"caching,omitempty" json:"caching,omitempty"`
	Pagination   *DefaultsPagination   `yaml:"pagination,omitempty" json:"pagination,omitempty"`
	Output       *DefaultsOutput       `yaml:"output,omitempty" json:"output,omitempty"`
	Deprecations *DefaultsDeprecations `yaml:"deprecations,omitempty" json:"deprecations,omitempty"`
	Retry        *DefaultsRetry        `yaml:"retry,omitempty" json:"retry,omitempty"`
}

// DefaultsHTTP contains HTTP client defaults.
type DefaultsHTTP struct {
	Timeout string `yaml:"timeout,omitempty" json:"timeout,omitempty"` // duration string
}

// DefaultsCaching contains caching defaults.
type DefaultsCaching struct {
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// DefaultsPagination contains pagination defaults.
type DefaultsPagination struct {
	Limit int `yaml:"limit,omitempty" json:"limit,omitempty"`
}

// DefaultsOutput contains output formatting defaults.
type DefaultsOutput struct {
	Format      string `yaml:"format,omitempty" json:"format,omitempty"`           // json, yaml, table, csv
	PrettyPrint bool   `yaml:"pretty_print,omitempty" json:"pretty_print,omitempty"`
	Color       string `yaml:"color,omitempty" json:"color,omitempty"`             // auto, always, never
	Paging      bool   `yaml:"paging,omitempty" json:"paging,omitempty"`
}

// DefaultsDeprecations contains deprecation warning defaults.
type DefaultsDeprecations struct {
	AlwaysShow  bool   `yaml:"always_show,omitempty" json:"always_show,omitempty"`
	MinSeverity string `yaml:"min_severity,omitempty" json:"min_severity,omitempty"` // info, warning, urgent, critical, removed
}

// DefaultsRetry contains retry defaults.
type DefaultsRetry struct {
	MaxAttempts int `yaml:"max_attempts,omitempty" json:"max_attempts,omitempty"`
}

// Updates defines the self-update configuration.
type Updates struct {
	Enabled       bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	UpdateURL     string `yaml:"update_url,omitempty" json:"update_url,omitempty"`
	CheckInterval string `yaml:"check_interval,omitempty" json:"check_interval,omitempty"` // duration string
	PublicKey     string `yaml:"public_key,omitempty" json:"public_key,omitempty"`         // PEM format
}

// Behaviors defines locked runtime behaviors.
type Behaviors struct {
	Auth            *AuthBehavior            `yaml:"auth,omitempty" json:"auth,omitempty"`
	Caching         *CachingBehavior         `yaml:"caching,omitempty" json:"caching,omitempty"`
	Retry           *RetryBehavior           `yaml:"retry,omitempty" json:"retry,omitempty"`
	Pagination      *PaginationBehavior      `yaml:"pagination,omitempty" json:"pagination,omitempty"`
	Secrets         *SecretsBehavior         `yaml:"secrets,omitempty" json:"secrets,omitempty"`
	BuiltinCommands *BuiltinCommands         `yaml:"builtin_commands,omitempty" json:"builtin_commands,omitempty"`
	GlobalFlags     *GlobalFlags             `yaml:"global_flags,omitempty" json:"global_flags,omitempty"`
}

// AuthBehavior defines authentication behavior.
type AuthBehavior struct {
	Type   string        `yaml:"type" json:"type"` // none, api_key, oauth2, basic
	APIKey *APIKeyAuth   `yaml:"api_key,omitempty" json:"api_key,omitempty"`
	OAuth2 *OAuth2Auth   `yaml:"oauth2,omitempty" json:"oauth2,omitempty"`
	Basic  *BasicAuth    `yaml:"basic,omitempty" json:"basic,omitempty"`
}

// APIKeyAuth defines API key authentication.
type APIKeyAuth struct {
	Header string `yaml:"header" json:"header"`
	EnvVar string `yaml:"env_var" json:"env_var"`
}

// OAuth2Auth defines OAuth2 authentication.
type OAuth2Auth struct {
	ClientID     string   `yaml:"client_id" json:"client_id"`
	ClientSecret string   `yaml:"client_secret,omitempty" json:"client_secret,omitempty"`
	AuthURL      string   `yaml:"auth_url" json:"auth_url"`
	TokenURL     string   `yaml:"token_url" json:"token_url"`
	Scopes       []string `yaml:"scopes,omitempty" json:"scopes,omitempty"`
	RedirectURL  string   `yaml:"redirect_url,omitempty" json:"redirect_url,omitempty"`
}

// BasicAuth defines basic authentication.
type BasicAuth struct {
	UsernameEnv string `yaml:"username_env" json:"username_env"`
	PasswordEnv string `yaml:"password_env" json:"password_env"`
}

// CachingBehavior defines caching behavior.
type CachingBehavior struct {
	SpecTTL     string `yaml:"spec_ttl,omitempty" json:"spec_ttl,omitempty"`         // duration string
	ResponseTTL string `yaml:"response_ttl,omitempty" json:"response_ttl,omitempty"` // duration string
	Directory   string `yaml:"directory,omitempty" json:"directory,omitempty"`
	MaxSize     string `yaml:"max_size,omitempty" json:"max_size,omitempty"` // e.g., "100MB"
}

// RetryBehavior defines retry behavior.
type RetryBehavior struct {
	Enabled            bool    `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	InitialDelay       string  `yaml:"initial_delay,omitempty" json:"initial_delay,omitempty"` // duration string
	MaxDelay           string  `yaml:"max_delay,omitempty" json:"max_delay,omitempty"`         // duration string
	BackoffMultiplier  float64 `yaml:"backoff_multiplier,omitempty" json:"backoff_multiplier,omitempty"`
	RetryOnStatus      []int   `yaml:"retry_on_status,omitempty" json:"retry_on_status,omitempty"`
}

// PaginationBehavior defines pagination behavior.
type PaginationBehavior struct {
	MaxLimit int    `yaml:"max_limit,omitempty" json:"max_limit,omitempty"`
	Delay    string `yaml:"delay,omitempty" json:"delay,omitempty"` // duration string for inter-page delay
}

// SecretsBehavior defines secrets handling behavior.
type SecretsBehavior struct {
	Enabled        bool             `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Masking        *SecretsMasking  `yaml:"masking,omitempty" json:"masking,omitempty"`
	FieldPatterns  []string         `yaml:"field_patterns,omitempty" json:"field_patterns,omitempty"`
	ValuePatterns  []ValuePattern   `yaml:"value_patterns,omitempty" json:"value_patterns,omitempty"`
	ExplicitFields []string         `yaml:"explicit_fields,omitempty" json:"explicit_fields,omitempty"`
	Headers        []string         `yaml:"headers,omitempty" json:"headers,omitempty"`
	MaskIn         *SecretsMaskIn   `yaml:"mask_in,omitempty" json:"mask_in,omitempty"`
}

// SecretsMasking defines masking strategy.
type SecretsMasking struct {
	Style            string `yaml:"style,omitempty" json:"style,omitempty"`               // partial, full, hash
	PartialShowChars int    `yaml:"partial_show_chars,omitempty" json:"partial_show_chars,omitempty"`
	Replacement      string `yaml:"replacement,omitempty" json:"replacement,omitempty"`
}

// ValuePattern defines a regex pattern for secret detection.
type ValuePattern struct {
	Name    string `yaml:"name" json:"name"`
	Pattern string `yaml:"pattern" json:"pattern"` // regex
	Enabled bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// SecretsMaskIn defines where to apply masking.
type SecretsMaskIn struct {
	Stdout      bool `yaml:"stdout,omitempty" json:"stdout,omitempty"`
	Stderr      bool `yaml:"stderr,omitempty" json:"stderr,omitempty"`
	Logs        bool `yaml:"logs,omitempty" json:"logs,omitempty"`
	Audit       bool `yaml:"audit,omitempty" json:"audit,omitempty"`
	DebugOutput bool `yaml:"debug_output,omitempty" json:"debug_output,omitempty"`
}

// BuiltinCommands defines configuration for built-in commands.
type BuiltinCommands struct {
	Version      *BuiltinCommand      `yaml:"version,omitempty" json:"version,omitempty"`
	Help         *BuiltinCommand      `yaml:"help,omitempty" json:"help,omitempty"`
	Info         *BuiltinCommand      `yaml:"info,omitempty" json:"info,omitempty"`
	Config       *ConfigCommand       `yaml:"config,omitempty" json:"config,omitempty"`
	Completion   *CompletionCommand   `yaml:"completion,omitempty" json:"completion,omitempty"`
	Update       *UpdateCommand       `yaml:"update,omitempty" json:"update,omitempty"`
	Changelog    *ChangelogCommand    `yaml:"changelog,omitempty" json:"changelog,omitempty"`
	Deprecations *DeprecationsCommand `yaml:"deprecations,omitempty" json:"deprecations,omitempty"`
	Cache        *CacheCommand        `yaml:"cache,omitempty" json:"cache,omitempty"`
	Auth         *AuthCommand         `yaml:"auth,omitempty" json:"auth,omitempty"`
}

// BuiltinCommand defines a basic built-in command.
type BuiltinCommand struct {
	Enabled bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Style   string   `yaml:"style,omitempty" json:"style,omitempty"` // subcommand, flag, hybrid
	Flags   []string `yaml:"flags,omitempty" json:"flags,omitempty"`
}

// ConfigCommand defines the config command.
type ConfigCommand struct {
	Enabled     bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Style       string   `yaml:"style,omitempty" json:"style,omitempty"`
	CommandName string   `yaml:"command_name,omitempty" json:"command_name,omitempty"`
	AllowEdit   bool     `yaml:"allow_edit,omitempty" json:"allow_edit,omitempty"`
	Subcommands []string `yaml:"subcommands,omitempty" json:"subcommands,omitempty"`
}

// CompletionCommand defines the completion command.
type CompletionCommand struct {
	Enabled bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Shells  []string `yaml:"shells,omitempty" json:"shells,omitempty"`
}

// UpdateCommand defines the update command.
type UpdateCommand struct {
	Enabled             bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	AutoCheck           bool `yaml:"auto_check,omitempty" json:"auto_check,omitempty"`
	RequireConfirmation bool `yaml:"require_confirmation,omitempty" json:"require_confirmation,omitempty"`
	ShowChangelog       bool `yaml:"show_changelog,omitempty" json:"show_changelog,omitempty"`
}

// ChangelogCommand defines the changelog command.
type ChangelogCommand struct {
	Enabled             bool               `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	ShowBinaryChanges   bool               `yaml:"show_binary_changes,omitempty" json:"show_binary_changes,omitempty"`
	ShowAPIChanges      bool               `yaml:"show_api_changes,omitempty" json:"show_api_changes,omitempty"`
	BinaryChangelog     *ChangelogSource   `yaml:"binary_changelog,omitempty" json:"binary_changelog,omitempty"`
	APIChangelog        *ChangelogSource   `yaml:"api_changelog,omitempty" json:"api_changelog,omitempty"`
	DefaultLimit        int                `yaml:"default_limit,omitempty" json:"default_limit,omitempty"`
}

// ChangelogSource defines where to fetch changelog data.
type ChangelogSource struct {
	Source string `yaml:"source,omitempty" json:"source,omitempty"` // embedded, url, api
	URL    string `yaml:"url,omitempty" json:"url,omitempty"`
}

// DeprecationsCommand defines the deprecations command.
type DeprecationsCommand struct {
	Enabled                bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	ShowBinaryDeprecations bool `yaml:"show_binary_deprecations,omitempty" json:"show_binary_deprecations,omitempty"`
	ShowAPIDeprecations    bool `yaml:"show_api_deprecations,omitempty" json:"show_api_deprecations,omitempty"`
	ShowByDefault          bool `yaml:"show_by_default,omitempty" json:"show_by_default,omitempty"`
	AllowScan              bool `yaml:"allow_scan,omitempty" json:"allow_scan,omitempty"`
	AllowAutoFix           bool `yaml:"allow_auto_fix,omitempty" json:"allow_auto_fix,omitempty"`
}

// CacheCommand defines the cache command.
type CacheCommand struct {
	Enabled     bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Subcommands []string `yaml:"subcommands,omitempty" json:"subcommands,omitempty"`
}

// AuthCommand defines the auth command.
type AuthCommand struct {
	Enabled     bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Subcommands []string `yaml:"subcommands,omitempty" json:"subcommands,omitempty"`
}

// GlobalFlags defines global flag configuration.
type GlobalFlags struct {
	Config   *GlobalFlag   `yaml:"config,omitempty" json:"config,omitempty"`
	Profile  *GlobalFlag   `yaml:"profile,omitempty" json:"profile,omitempty"`
	Region   *GlobalFlag   `yaml:"region,omitempty" json:"region,omitempty"`
	Output   *GlobalFlag   `yaml:"output,omitempty" json:"output,omitempty"`
	Verbose  *GlobalFlag   `yaml:"verbose,omitempty" json:"verbose,omitempty"`
	Quiet    *GlobalFlag   `yaml:"quiet,omitempty" json:"quiet,omitempty"`
	Debug    *GlobalFlag   `yaml:"debug,omitempty" json:"debug,omitempty"`
	NoColor  *GlobalFlag   `yaml:"no_color,omitempty" json:"no_color,omitempty"`
	Timeout  *GlobalFlag   `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Retry    *GlobalFlag   `yaml:"retry,omitempty" json:"retry,omitempty"`
	NoCache  *GlobalFlag   `yaml:"no_cache,omitempty" json:"no_cache,omitempty"`
	Yes      *GlobalFlag   `yaml:"yes,omitempty" json:"yes,omitempty"`
	Custom   []CustomFlag  `yaml:"custom,omitempty" json:"custom,omitempty"`
}

// GlobalFlag defines a global flag configuration.
type GlobalFlag struct {
	Enabled       bool     `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	Flag          string   `yaml:"flag,omitempty" json:"flag,omitempty"`
	Short         string   `yaml:"short,omitempty" json:"short,omitempty"`
	EnvVar        string   `yaml:"env_var,omitempty" json:"env_var,omitempty"`
	Description   string   `yaml:"description,omitempty" json:"description,omitempty"`
	Default       any      `yaml:"default,omitempty" json:"default,omitempty"`
	AllowedValues []string `yaml:"allowed_values,omitempty" json:"allowed_values,omitempty"`
	Repeatable    bool     `yaml:"repeatable,omitempty" json:"repeatable,omitempty"`
	ConflictsWith []string `yaml:"conflicts_with,omitempty" json:"conflicts_with,omitempty"`
}

// CustomFlag defines a custom global flag.
type CustomFlag struct {
	Name          string   `yaml:"name" json:"name"`
	Flag          string   `yaml:"flag" json:"flag"`
	Short         string   `yaml:"short,omitempty" json:"short,omitempty"`
	EnvVar        string   `yaml:"env_var,omitempty" json:"env_var,omitempty"`
	Description   string   `yaml:"description,omitempty" json:"description,omitempty"`
	Type          string   `yaml:"type,omitempty" json:"type,omitempty"` // string, int, bool, duration
	Default       any      `yaml:"default,omitempty" json:"default,omitempty"`
	Required      bool     `yaml:"required,omitempty" json:"required,omitempty"`
	AllowedValues []string `yaml:"allowed_values,omitempty" json:"allowed_values,omitempty"`
	Sensitive     bool     `yaml:"sensitive,omitempty" json:"sensitive,omitempty"`
	ConflictsWith []string `yaml:"conflicts_with,omitempty" json:"conflicts_with,omitempty"`
}

// Features defines optional feature flags.
type Features struct {
	ConfigFile       bool   `yaml:"config_file,omitempty" json:"config_file,omitempty"`
	ConfigFilePath   string `yaml:"config_file_path,omitempty" json:"config_file_path,omitempty"`
	InteractiveMode  bool   `yaml:"interactive_mode,omitempty" json:"interactive_mode,omitempty"`
}

// UserConfig represents user-specific configuration overrides.
type UserConfig struct {
	Preferences   *UserPreferences `yaml:"preferences,omitempty" json:"preferences,omitempty"`
	DebugOverride *CLIConfig       `yaml:"debug_override,omitempty" json:"debug_override,omitempty"`
}

// UserPreferences represents user preferences that can override defaults.
type UserPreferences struct {
	HTTP         *PreferencesHTTP         `yaml:"http,omitempty" json:"http,omitempty"`
	Caching      *PreferencesCaching      `yaml:"caching,omitempty" json:"caching,omitempty"`
	Pagination   *PreferencesPagination   `yaml:"pagination,omitempty" json:"pagination,omitempty"`
	Output       *PreferencesOutput       `yaml:"output,omitempty" json:"output,omitempty"`
	Deprecations *PreferencesDeprecations `yaml:"deprecations,omitempty" json:"deprecations,omitempty"`
	Retry        *PreferencesRetry        `yaml:"retry,omitempty" json:"retry,omitempty"`
	Telemetry    *PreferencesTelemetry    `yaml:"telemetry,omitempty" json:"telemetry,omitempty"`
	Updates      *PreferencesUpdates      `yaml:"updates,omitempty" json:"updates,omitempty"`
}

// PreferencesHTTP contains HTTP preferences.
type PreferencesHTTP struct {
	Timeout     string              `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Proxy       string              `yaml:"proxy,omitempty" json:"proxy,omitempty"`
	HTTPSProxy  string              `yaml:"https_proxy,omitempty" json:"https_proxy,omitempty"`
	NoProxy     []string            `yaml:"no_proxy,omitempty" json:"no_proxy,omitempty"`
	CABundle    string              `yaml:"ca_bundle,omitempty" json:"ca_bundle,omitempty"`
	TLS         *PreferencesTLS     `yaml:"tls,omitempty" json:"tls,omitempty"`
}

// PreferencesTLS contains TLS preferences.
type PreferencesTLS struct {
	CABundle           string `yaml:"ca_bundle,omitempty" json:"ca_bundle,omitempty"`
	InsecureSkipVerify bool   `yaml:"insecure_skip_verify,omitempty" json:"insecure_skip_verify,omitempty"`
}

// PreferencesCaching contains caching preferences.
type PreferencesCaching struct {
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// PreferencesPagination contains pagination preferences.
type PreferencesPagination struct {
	Limit int `yaml:"limit,omitempty" json:"limit,omitempty"`
}

// PreferencesOutput contains output preferences.
type PreferencesOutput struct {
	Format      string `yaml:"format,omitempty" json:"format,omitempty"`
	Color       string `yaml:"color,omitempty" json:"color,omitempty"`
	PrettyPrint bool   `yaml:"pretty_print,omitempty" json:"pretty_print,omitempty"`
	Paging      bool   `yaml:"paging,omitempty" json:"paging,omitempty"`
}

// PreferencesDeprecations contains deprecation preferences.
type PreferencesDeprecations struct {
	AlwaysShow  bool   `yaml:"always_show,omitempty" json:"always_show,omitempty"`
	MinSeverity string `yaml:"min_severity,omitempty" json:"min_severity,omitempty"`
}

// PreferencesRetry contains retry preferences.
type PreferencesRetry struct {
	MaxAttempts int `yaml:"max_attempts,omitempty" json:"max_attempts,omitempty"`
}

// PreferencesTelemetry contains telemetry preferences (user-only).
type PreferencesTelemetry struct {
	Enabled bool `yaml:"enabled,omitempty" json:"enabled,omitempty"`
}

// PreferencesUpdates contains update preferences (user-only).
type PreferencesUpdates struct {
	AutoInstall bool `yaml:"auto_install,omitempty" json:"auto_install,omitempty"`
}

// OpenAPIExtensions represents custom OpenAPI extensions for CLI generation.
type OpenAPIExtensions struct {
	CLIVersion       string            `yaml:"x-cli-version,omitempty" json:"x-cli-version,omitempty"`
	CLIMinVersion    string            `yaml:"x-cli-min-version,omitempty" json:"x-cli-min-version,omitempty"`
	CLIChangelog     []ChangelogEntry  `yaml:"x-cli-changelog,omitempty" json:"x-cli-changelog,omitempty"`
	CLIAliases       []string          `yaml:"x-cli-aliases,omitempty" json:"x-cli-aliases,omitempty"`
	CLIExamples      []Example         `yaml:"x-cli-examples,omitempty" json:"x-cli-examples,omitempty"`
	CLIHidden        bool              `yaml:"x-cli-hidden,omitempty" json:"x-cli-hidden,omitempty"`
	CLIAuth          *CLIAuth          `yaml:"x-cli-auth,omitempty" json:"x-cli-auth,omitempty"`
	CLIWorkflow      *Workflow         `yaml:"x-cli-workflow,omitempty" json:"x-cli-workflow,omitempty"`
	CLIDeprecation   *Deprecation      `yaml:"x-cli-deprecation,omitempty" json:"x-cli-deprecation,omitempty"`
	CLISecret        bool              `yaml:"x-cli-secret,omitempty" json:"x-cli-secret,omitempty"`
}

// ChangelogEntry represents a single changelog entry.
type ChangelogEntry struct {
	Date    string           `yaml:"date" json:"date"`
	Version string           `yaml:"version" json:"version"`
	Changes []ChangelogChange `yaml:"changes" json:"changes"`
}

// ChangelogChange represents a single change within a changelog entry.
type ChangelogChange struct {
	Type        string `yaml:"type" json:"type"`               // added, removed, modified, deprecated, security
	Severity    string `yaml:"severity" json:"severity"`       // breaking, dangerous, safe
	Description string `yaml:"description" json:"description"`
	Path        string `yaml:"path,omitempty" json:"path,omitempty"`
	Migration   string `yaml:"migration,omitempty" json:"migration,omitempty"`
	Sunset      string `yaml:"sunset,omitempty" json:"sunset,omitempty"` // date string
}

// Example represents a CLI usage example.
type Example struct {
	Description string `yaml:"description" json:"description"`
	Command     string `yaml:"command" json:"command"`
}

// CLIAuth represents authentication requirements for an operation.
type CLIAuth struct {
	Required bool     `yaml:"required,omitempty" json:"required,omitempty"`
	Scopes   []string `yaml:"scopes,omitempty" json:"scopes,omitempty"`
}

// Workflow represents a multi-step API workflow.
type Workflow struct {
	Steps  []WorkflowStep  `yaml:"steps" json:"steps"`
	Output *WorkflowOutput `yaml:"output,omitempty" json:"output,omitempty"`
}

// WorkflowStep represents a single step in a workflow.
type WorkflowStep struct {
	ID        string           `yaml:"id" json:"id"`
	Request   *WorkflowRequest `yaml:"request,omitempty" json:"request,omitempty"`
	Foreach   string           `yaml:"foreach,omitempty" json:"foreach,omitempty"` // expr expression
	As        string           `yaml:"as,omitempty" json:"as,omitempty"`           // loop variable name
	Condition string           `yaml:"condition,omitempty" json:"condition,omitempty"` // expr expression
}

// WorkflowRequest represents an HTTP request within a workflow step.
type WorkflowRequest struct {
	Method  string            `yaml:"method" json:"method"`
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	Body    any               `yaml:"body,omitempty" json:"body,omitempty"`
	Query   map[string]string `yaml:"query,omitempty" json:"query,omitempty"`
}

// WorkflowOutput defines how to transform and display workflow results.
type WorkflowOutput struct {
	Format    string `yaml:"format,omitempty" json:"format,omitempty"`       // json, yaml, table, csv
	Transform string `yaml:"transform,omitempty" json:"transform,omitempty"` // expr expression
}

// Deprecation represents deprecation information for an operation.
type Deprecation struct {
	Severity    string    `yaml:"severity" json:"severity"`                             // info, warning, urgent, critical, removed
	Sunset      string    `yaml:"sunset,omitempty" json:"sunset,omitempty"`             // ISO 8601 date
	Message     string    `yaml:"message" json:"message"`
	Replacement *string   `yaml:"replacement,omitempty" json:"replacement,omitempty"`   // alternative operation
	Migration   *string   `yaml:"migration,omitempty" json:"migration,omitempty"`       // migration guide
	Links       []string  `yaml:"links,omitempty" json:"links,omitempty"`
}

// PluginManifest represents a plugin's manifest file.
type PluginManifest struct {
	Name        string          `yaml:"name" json:"name"`
	Version     string          `yaml:"version" json:"version"`
	Type        string          `yaml:"type" json:"type"` // builtin, binary, wasm
	Binary      string          `yaml:"binary,omitempty" json:"binary,omitempty"`
	Permissions []string        `yaml:"permissions,omitempty" json:"permissions,omitempty"`
	Commands    []PluginCommand `yaml:"commands,omitempty" json:"commands,omitempty"`
}

// PluginCommand represents a command provided by a plugin.
type PluginCommand struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Aliases     []string `yaml:"aliases,omitempty" json:"aliases,omitempty"`
}

// ConfigPriority represents the priority of configuration sources.
type ConfigPriority int

const (
	PriorityDefault ConfigPriority = iota
	PriorityEmbedded
	PriorityUserConfig
	PriorityDebugOverride
	PriorityFlag
	PriorityEnv
)

// LoadedConfig represents a fully loaded and merged configuration.
type LoadedConfig struct {
	Final              *CLIConfig
	EmbeddedConfig     *CLIConfig
	UserConfig         *UserConfig
	DebugOverrides     map[string]any
	EffectiveTimestamp time.Time
}
