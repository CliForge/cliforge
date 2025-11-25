// Package plugin provides a plugin architecture for extending CliForge capabilities.
//
// The plugin system allows CliForge to execute operations that cannot be handled
// by pure OpenAPI specifications, such as:
//   - Executing external CLI tools (AWS CLI, kubectl, etc.)
//   - Performing local file operations and transformations
//   - Implementing custom validation logic
//   - Integrating with authentication providers
//
// Security Model:
//   - All plugins declare required permissions in their manifest
//   - Users must approve permissions before first use
//   - External plugins run in sandboxed environments
//   - Built-in plugins are trusted by default
package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// Plugin is the core interface that all plugins must implement.
type Plugin interface {
	// Execute runs the plugin with the given input and returns output.
	// Context can be used for cancellation and timeouts.
	Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error)

	// Validate checks if the plugin is properly configured and ready to execute.
	// Returns an error if validation fails.
	Validate() error

	// Describe returns metadata about the plugin.
	Describe() *PluginInfo
}

// PluginType represents the type of plugin.
type PluginType string

const (
	// PluginTypeBuiltin is a plugin compiled into the binary.
	PluginTypeBuiltin PluginType = "builtin"

	// PluginTypeBinary is an external executable plugin.
	PluginTypeBinary PluginType = "binary"

	// PluginTypeWASM is a WebAssembly plugin (future).
	PluginTypeWASM PluginType = "wasm"
)

// PluginManifest describes a plugin's metadata and requirements.
type PluginManifest struct {
	// Name is the unique identifier for the plugin.
	Name string `yaml:"name" json:"name"`

	// Version is the semantic version of the plugin.
	Version string `yaml:"version" json:"version"`

	// Type indicates the plugin type (builtin, binary, wasm).
	Type PluginType `yaml:"type" json:"type"`

	// Description provides a human-readable description.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Author is the plugin author or organization.
	Author string `yaml:"author,omitempty" json:"author,omitempty"`

	// Executable is the path to the executable (for binary plugins).
	Executable string `yaml:"executable,omitempty" json:"executable,omitempty"`

	// Entrypoint is the entry function name (for WASM plugins).
	Entrypoint string `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`

	// Permissions lists required permissions.
	Permissions []Permission `yaml:"permissions" json:"permissions"`

	// Metadata contains additional plugin-specific metadata.
	Metadata map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// Permission represents a capability that a plugin requires.
type Permission struct {
	// Type is the permission type (execute, read:env, write:file, etc.).
	Type PermissionType `yaml:"type" json:"type"`

	// Resource specifies the resource pattern (e.g., "aws" for execute:aws).
	Resource string `yaml:"resource,omitempty" json:"resource,omitempty"`

	// Description explains why this permission is needed.
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// PermissionType defines the types of permissions a plugin can request.
type PermissionType string

const (
	// PermissionExecute allows executing external commands.
	PermissionExecute PermissionType = "execute"

	// PermissionReadEnv allows reading environment variables.
	PermissionReadEnv PermissionType = "read:env"

	// PermissionWriteEnv allows writing environment variables.
	PermissionWriteEnv PermissionType = "write:env"

	// PermissionReadFile allows reading files.
	PermissionReadFile PermissionType = "read:file"

	// PermissionWriteFile allows writing files.
	PermissionWriteFile PermissionType = "write:file"

	// PermissionNetwork allows making network requests.
	PermissionNetwork PermissionType = "network"

	// PermissionCredential allows accessing stored credentials.
	PermissionCredential PermissionType = "credential"
)

// String returns the full permission string (type:resource).
func (p Permission) String() string {
	if p.Resource != "" {
		return fmt.Sprintf("%s:%s", p.Type, p.Resource)
	}
	return string(p.Type)
}

// PluginInfo provides metadata about a plugin.
type PluginInfo struct {
	// Manifest is the plugin's manifest.
	Manifest PluginManifest `json:"manifest"`

	// Capabilities lists what the plugin can do.
	Capabilities []string `json:"capabilities,omitempty"`

	// Status indicates the plugin's current state.
	Status PluginStatus `json:"status"`

	// LastUsed is the timestamp when the plugin was last executed.
	LastUsed *time.Time `json:"last_used,omitempty"`
}

// PluginStatus represents the operational status of a plugin.
type PluginStatus string

const (
	// PluginStatusReady indicates the plugin is ready to execute.
	PluginStatusReady PluginStatus = "ready"

	// PluginStatusDisabled indicates the plugin is disabled.
	PluginStatusDisabled PluginStatus = "disabled"

	// PluginStatusError indicates the plugin has an error.
	PluginStatusError PluginStatus = "error"

	// PluginStatusPendingApproval indicates awaiting user permission approval.
	PluginStatusPendingApproval PluginStatus = "pending_approval"
)

// PluginInput represents input data passed to a plugin.
type PluginInput struct {
	// Command is the plugin command to execute.
	Command string `json:"command,omitempty"`

	// Args are command-line arguments.
	Args []string `json:"args,omitempty"`

	// Env contains environment variables.
	Env map[string]string `json:"env,omitempty"`

	// Data contains arbitrary input data.
	Data map[string]interface{} `json:"data,omitempty"`

	// Files contains file paths for file operations.
	Files map[string]string `json:"files,omitempty"`

	// Stdin is input data for stdin.
	Stdin string `json:"stdin,omitempty"`

	// Timeout is the execution timeout.
	Timeout time.Duration `json:"timeout,omitempty"`

	// WorkingDir is the working directory for execution.
	WorkingDir string `json:"working_dir,omitempty"`
}

// PluginOutput represents output from a plugin execution.
type PluginOutput struct {
	// Stdout is the standard output.
	Stdout string `json:"stdout,omitempty"`

	// Stderr is the standard error output.
	Stderr string `json:"stderr,omitempty"`

	// ExitCode is the exit code (for command execution).
	ExitCode int `json:"exit_code"`

	// Data contains structured output data.
	Data map[string]interface{} `json:"data,omitempty"`

	// Error contains error information if execution failed.
	Error string `json:"error,omitempty"`

	// Duration is the execution duration.
	Duration time.Duration `json:"duration,omitempty"`

	// Metadata contains additional execution metadata.
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Success returns true if the plugin execution succeeded.
func (o *PluginOutput) Success() bool {
	return o.ExitCode == 0 && o.Error == ""
}

// GetString retrieves a string value from output data.
func (o *PluginOutput) GetString(key string) (string, bool) {
	if o.Data == nil {
		return "", false
	}
	val, exists := o.Data[key]
	if !exists {
		return "", false
	}
	str, ok := val.(string)
	return str, ok
}

// GetInt retrieves an integer value from output data.
func (o *PluginOutput) GetInt(key string) (int, bool) {
	if o.Data == nil {
		return 0, false
	}
	val, exists := o.Data[key]
	if !exists {
		return 0, false
	}

	// Handle different numeric types
	switch v := val.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}

// GetBool retrieves a boolean value from output data.
func (o *PluginOutput) GetBool(key string) (bool, bool) {
	if o.Data == nil {
		return false, false
	}
	val, exists := o.Data[key]
	if !exists {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// GetMap retrieves a map value from output data.
func (o *PluginOutput) GetMap(key string) (map[string]interface{}, bool) {
	if o.Data == nil {
		return nil, false
	}
	val, exists := o.Data[key]
	if !exists {
		return nil, false
	}
	m, ok := val.(map[string]interface{})
	return m, ok
}

// MarshalJSON customizes JSON marshaling for PluginOutput.
func (o *PluginOutput) MarshalJSON() ([]byte, error) {
	type Alias PluginOutput
	return json.Marshal(&struct {
		Duration string `json:"duration,omitempty"`
		*Alias
	}{
		Duration: o.Duration.String(),
		Alias:    (*Alias)(o),
	})
}

// PluginError represents a plugin execution error.
type PluginError struct {
	// PluginName is the name of the plugin that failed.
	PluginName string

	// Message is the error message.
	Message string

	// Cause is the underlying error.
	Cause error

	// Recoverable indicates if the error is recoverable.
	Recoverable bool

	// Suggestion provides a hint to resolve the error.
	Suggestion string
}

// Error implements the error interface.
func (e *PluginError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("plugin '%s': %s: %v", e.PluginName, e.Message, e.Cause)
	}
	return fmt.Sprintf("plugin '%s': %s", e.PluginName, e.Message)
}

// Unwrap returns the underlying error.
func (e *PluginError) Unwrap() error {
	return e.Cause
}

// NewPluginError creates a new plugin error.
func NewPluginError(pluginName, message string, cause error) *PluginError {
	return &PluginError{
		PluginName:  pluginName,
		Message:     message,
		Cause:       cause,
		Recoverable: false,
	}
}

// WithSuggestion adds a suggestion to the error.
func (e *PluginError) WithSuggestion(suggestion string) *PluginError {
	e.Suggestion = suggestion
	return e
}

// AsRecoverable marks the error as recoverable.
func (e *PluginError) AsRecoverable() *PluginError {
	e.Recoverable = true
	return e
}

// PluginContext contains contextual information for plugin execution.
type PluginContext struct {
	// CLIName is the name of the CLI application.
	CLIName string

	// CLIVersion is the version of the CLI application.
	CLIVersion string

	// ConfigDir is the configuration directory path.
	ConfigDir string

	// CacheDir is the cache directory path.
	CacheDir string

	// DataDir is the data directory path.
	DataDir string

	// Debug indicates if debug mode is enabled.
	Debug bool

	// DryRun indicates if this is a dry run (no side effects).
	DryRun bool

	// UserData contains arbitrary user-provided context data.
	UserData map[string]interface{}
}

// ValidationError represents a plugin validation error.
type ValidationError struct {
	Plugin  string
	Field   string
	Message string
}

// Error implements the error interface.
func (e *ValidationError) Error() string {
	return fmt.Sprintf("plugin '%s': validation failed for '%s': %s", e.Plugin, e.Field, e.Message)
}
