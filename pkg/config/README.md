# Configuration Package

This package implements the configuration system for CliForge v0.9.0, handling loading, merging, and validation of CLI configurations.

## Features

- **Multi-Source Loading**: Loads configuration from embedded files, user config files, environment variables, and flags
- **XDG Compliance**: Uses XDG Base Directory specification for config, cache, data, and state directories
- **Priority Chain**: ENV > Flag > User Config > Debug Override > Embedded > Default
- **Debug Mode**: Supports debug builds with full configuration override capability
- **Validation**: Comprehensive validation of all configuration sections
- **Merging**: Smart merging of configurations respecting locked vs overridable settings

## Configuration Priority

Configuration values are resolved in this order (highest to lowest):

1. **Environment Variables** - e.g., `MYCLI_OUTPUT_FORMAT=json`
2. **Command-line Flags** - e.g., `--output json`
3. **User Config File** - `~/.config/mycli/config.yaml` (preferences section)
4. **Debug Override** - Only active when `metadata.debug: true`
5. **Embedded Config** - Built into binary (defaults section)
6. **Built-in Defaults** - Hardcoded fallbacks

## Override Rules

### Locked Settings (71 total)

These settings are locked to embedded config and cannot be overridden by users:

- `metadata.*` (9 settings) - CLI identity
- `branding.*` (4 settings) - Visual branding
- `api.*` (7 settings) - API endpoints and telemetry
- `behaviors.auth.*` (5 settings) - Authentication
- `behaviors.retry.*` (5 settings) - Retry behavior
- `behaviors.caching.*` (4 settings) - Cache behavior
- `behaviors.pagination.*` (2 settings) - Pagination limits
- `behaviors.secrets.*` (10 settings) - Secret masking
- `behaviors.builtin_commands.*` (10 settings) - Built-in commands
- `behaviors.global_flags.*` (15 settings) - Global flags

### Overridable Settings (13 total)

Users can override these via the `preferences` section:

- `defaults.http.timeout` - Request timeout
- `defaults.caching.enabled` - Enable/disable caching
- `defaults.pagination.limit` - Default page size
- `defaults.output.*` (4 settings) - Output formatting
- `defaults.deprecations.*` (2 settings) - Deprecation warnings
- `defaults.retry.max_attempts` - Retry attempts
- `preferences.http.proxy` - HTTP proxy (user-only)
- `preferences.http.ca_bundle` - CA certificates (user-only)
- `preferences.telemetry.enabled` - Telemetry opt-in (user-only)

## Debug Mode

When `metadata.debug: true`, users can override ANY setting via the `debug_override` section:

```yaml
# User config
preferences:
  output:
    format: yaml

debug_override:
  api:
    base_url: http://localhost:8080
  behaviors:
    auth:
      type: none
```

Debug mode displays a prominent security warning on every command execution.

## Usage

### Loading Configuration

```go
import (
    "embed"
    "github.com/CliForge/cliforge/pkg/config"
)

//go:embed cli-config.yaml
var embeddedFS embed.FS

func main() {
    loader := config.NewLoader("my-cli", &embeddedFS, "cli-config.yaml")

    // Ensure XDG directories exist
    if err := loader.EnsureConfigDirs(); err != nil {
        log.Fatal(err)
    }

    // Load and merge all configs
    loaded, err := loader.LoadConfig()
    if err != nil {
        log.Fatal(err)
    }

    // Show warnings (debug mode, ignored overrides, etc.)
    loader.ShowWarnings(loaded)

    // Use the final merged config
    config := loaded.Final
}
```

### Validating Configuration

```go
import "github.com/CliForge/cliforge/pkg/config"

validator := config.NewValidator()
if err := validator.Validate(cfg); err != nil {
    log.Fatalf("Invalid config: %v", err)
}
```

### Saving User Configuration

```go
userConfig := &cli.UserConfig{
    Preferences: &cli.UserPreferences{
        Output: &cli.PreferencesOutput{
            Format: "yaml",
        },
    },
}

if err := loader.SaveUserConfig(userConfig); err != nil {
    log.Fatal(err)
}
```

## XDG Directories

The configuration system uses XDG-compliant paths:

- **Config**: `$XDG_CONFIG_HOME/mycli/config.yaml` (default: `~/.config/mycli/`)
- **Cache**: `$XDG_CACHE_HOME/mycli/` (default: `~/.cache/mycli/`)
- **Data**: `$XDG_DATA_HOME/mycli/` (default: `~/.local/share/mycli/`)
- **State**: `$XDG_STATE_HOME/mycli/` (default: `~/.local/state/mycli/`)

## Environment Variables

Common environment variable mappings:

- `MYCLI_OUTPUT_FORMAT` → `defaults.output.format`
- `MYCLI_TIMEOUT` → `defaults.http.timeout`
- `NO_COLOR` → `defaults.output.color` (set to "never")
- `MYCLI_NO_CACHE` → `defaults.caching.enabled` (set to false)
- `MYCLI_PAGE_LIMIT` → `defaults.pagination.limit`
- `MYCLI_RETRY` → `defaults.retry.max_attempts`

## Testing

Run tests:

```bash
go test ./pkg/config/...
```

Test fixtures are available in `testdata/config/`:

- `valid-embedded.yaml` - Complete valid embedded config
- `valid-user-preferences.yaml` - Valid user preferences
- `debug-build-with-override.yaml` - Debug build configuration
- `user-debug-override.yaml` - User config with debug overrides
- `invalid-*.yaml` - Various invalid configurations for validation testing
