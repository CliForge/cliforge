# CliForge Configuration Guide

**Version:** 0.9.0
**Last Updated:** 2025-01-23

## Table of Contents

1. [Introduction](#introduction)
2. [Configuration File Structure](#configuration-file-structure)
3. [Configuration Locations](#configuration-locations)
4. [Configuration Sections](#configuration-sections)
5. [Configuration Priority Chain](#configuration-priority-chain)
6. [Profiles for Multiple Environments](#profiles-for-multiple-environments)
7. [Debug Mode](#debug-mode)
8. [Common Configuration Patterns](#common-configuration-patterns)
9. [Configuration Validation](#configuration-validation)
10. [Best Practices](#best-practices)

---

## Introduction

CliForge uses a sophisticated configuration system that allows CLI developers to embed default configurations in the binary while giving users the flexibility to override specific settings. This guide explains how to work with CliForge configurations effectively.

### Key Concepts

- **Embedded Configuration**: Built into the CLI binary during generation
- **User Configuration**: Optional file in the user's config directory
- **Environment Variables**: Runtime overrides via environment
- **Command-Line Flags**: Highest priority overrides
- **Configuration Priority**: ENV > Flag > User > Embedded > Default

---

## Configuration File Structure

CliForge uses YAML for configuration files. There are two main configuration file types:

### 1. Embedded Configuration (`cli-config.yaml`)

This file is used during CLI generation and embedded into the binary. It defines the CLI's default behavior and branding.

**Location**: Provided to `cliforge generate` command

**Key Sections**:
- `metadata` - CLI identity and version information
- `branding` - Colors, ASCII art, themes
- `api` - OpenAPI spec location and API endpoints
- `defaults` - User-overridable default settings
- `updates` - Self-update configuration
- `behaviors` - Locked runtime behaviors
- `features` - Feature toggles

### 2. User Configuration (`config.yaml`)

Optional file created by users to override defaults.

**Location**: `$XDG_CONFIG_HOME/<cli-name>/config.yaml`
- Linux/macOS: `~/.config/<cli-name>/config.yaml`
- Windows: `%APPDATA%\<cli-name>\config.yaml`

**Key Sections**:
- `preferences` - Overrides embedded defaults (always active)
- `debug_override` - Overrides any embedded setting (debug builds only)

---

## Configuration Locations

CliForge follows the XDG Base Directory Specification for all file locations.

### XDG Directory Structure

| Directory Type | Environment Variable | Default Location | Purpose | CLI Usage |
|----------------|---------------------|------------------|---------|-----------|
| Config | `XDG_CONFIG_HOME` | `~/.config` | User configuration | `config.yaml`, profiles |
| Cache | `XDG_CACHE_HOME` | `~/.cache` | Cached data | OpenAPI specs, HTTP responses |
| Data | `XDG_DATA_HOME` | `~/.local/share` | Persistent data | Logs, audit trails |
| State | `XDG_STATE_HOME` | `~/.local/state` | State information | Last update check, session data |
| Runtime | `XDG_RUNTIME_DIR` | *(no default)* | Runtime files | PID files, sockets |

### Example Paths for `petstore` CLI

```bash
# Config
~/.config/petstore/config.yaml

# Cache
~/.cache/petstore/openapi-spec.yaml
~/.cache/petstore/responses/

# Data
~/.local/share/petstore/petstore.log
~/.local/share/petstore/audit.log

# State
~/.local/state/petstore/last-update-check
~/.local/state/petstore/session.json
```

### Custom XDG Paths

Users can customize directory locations using environment variables:

```bash
export XDG_CONFIG_HOME="$HOME/.custom-config"
export XDG_CACHE_HOME="$HOME/.custom-cache"
export XDG_DATA_HOME="$HOME/.custom-data"
export XDG_STATE_HOME="$HOME/.custom-state"
```

---

## Configuration Sections

### Metadata Section

Defines the CLI's identity and version information.

**Location**: Embedded config only (not overridable)

```yaml
metadata:
  name: petstore                    # CLI binary name (required)
  version: 1.0.0                    # Semantic version (required)
  description: Petstore CLI         # Short description (required)
  long_description: |               # Extended description (optional)
    Complete CLI for managing the Petstore API.
    Supports pets, users, and store operations.

  author:                           # Author information (optional)
    name: Petstore Team
    email: team@petstore.example.com
    url: https://github.com/petstore/cli

  license: MIT                      # License identifier (optional)
  homepage: https://petstore.example.com/cli
  bugs_url: https://github.com/petstore/cli/issues
  docs_url: https://docs.petstore.example.com

  debug: false                      # Debug mode (default: false)
                                    # SECURITY: Only true in dev builds
```

**Debug Mode**: When `debug: true`, allows `debug_override` section in user config to override ANY embedded setting. Should NEVER be true in production builds.

---

### Branding Section

Customizes the look and feel of the CLI.

**Location**: Embedded config only (not overridable except in debug mode)

```yaml
branding:
  # Color scheme (hex colors)
  colors:
    primary: "#FF6B35"      # Main brand color
    secondary: "#004E89"    # Accent color
    success: "#06D6A0"      # Success messages
    warning: "#F4A261"      # Warnings
    error: "#E76F51"        # Errors
    info: "#2A9D8F"         # Info messages

  # ASCII art banner
  ascii_art: |
    ╔═══════════════════════════════════════╗
    ║         PETSTORE CLI v1.0             ║
    ║  Manage your petstore from the CLI    ║
    ╚═══════════════════════════════════════╝

  # Custom prompts
  prompts:
    command: "petstore>"    # Command prompt
    error: "✗"              # Error symbol
    success: "✓"            # Success symbol
    warning: "⚠"            # Warning symbol
    info: "ℹ"               # Info symbol

  # Theme settings
  theme:
    name: auto              # auto, light, dark
    syntax_highlighting: true
```

**Color Guidelines**:
- Use 6-digit hex codes (#RRGGBB)
- Test on both light and dark terminals
- Ensure accessibility (sufficient contrast)
- Alpha channel not supported

---

### API Section

Defines API endpoints and connection settings.

**Location**: Embedded config only (LOCKED - cannot be overridden except in debug mode)

**Security Note**: The entire `api` section is locked to prevent users from redirecting the CLI to malicious endpoints or telemetry servers.

```yaml
api:
  # OpenAPI spec location (required)
  openapi_url: https://api.petstore.example.com/openapi.yaml

  # Default API base URL (required)
  base_url: https://api.petstore.example.com

  # API version (optional)
  version: v1

  # Default headers for all requests (optional)
  default_headers:
    X-Client-Version: "1.0.0"
    Accept: "application/json"

  # User agent (optional)
  user_agent: "petstore-cli/1.0.0"

  # Telemetry endpoint (optional)
  telemetry_url: https://telemetry.petstore.example.com/events

  # Multiple environments (optional)
  environments:
    - name: production
      openapi_url: https://api.petstore.example.com/openapi.yaml
      base_url: https://api.petstore.example.com
      default: true

    - name: staging
      openapi_url: https://staging-api.petstore.example.com/openapi.yaml
      base_url: https://staging-api.petstore.example.com

    - name: development
      openapi_url: https://dev-api.petstore.example.com/openapi.yaml
      base_url: https://dev-api.petstore.example.com
```

**Using Environments**:

```bash
# Use default environment (production)
petstore pets list

# Use specific environment
petstore --env staging pets list

# Switch default environment in user config
petstore config set preferences.api.default_environment staging
```

---

### Defaults Section

User-overridable default settings. These can be customized in the user's `config.yaml` under the `preferences` section.

**Location**: Embedded config (provides defaults), User config (overrides via `preferences`)

```yaml
defaults:
  # HTTP client settings
  http:
    timeout: 30s                    # Request timeout (duration)

  # Caching settings
  caching:
    enabled: true                   # Enable response caching

  # Pagination settings
  pagination:
    limit: 20                       # Default page size

  # Output settings
  output:
    format: json                    # json, yaml, table, csv
    pretty_print: true              # Pretty-print output
    color: auto                     # auto, always, never
    paging: true                    # Use pager for long output

  # Deprecation warnings
  deprecations:
    always_show: false              # Show warnings once per version
    min_severity: info              # info, warning, urgent, critical, removed

  # Retry settings
  retry:
    max_attempts: 3                 # Number of retry attempts
```

**User Override Example**:

```yaml
# ~/.config/petstore/config.yaml
preferences:
  output:
    format: yaml                    # Prefer YAML over JSON
    color: always                   # Always use colors

  http:
    timeout: 60s                    # Longer timeout

  pagination:
    limit: 50                       # More items per page
```

---

### Updates Section

Configures self-update behavior.

**Location**: Embedded config only (LOCKED except `auto_install`)

**Security Note**: Users cannot override update URL or public key. Only `auto_install` can be set in user config.

```yaml
updates:
  enabled: true                     # Enable update checking
  update_url: https://releases.petstore.example.com/cli
  check_interval: 24h               # Check for updates every 24 hours

  # Public key for signature verification (Ed25519)
  public_key: |
    -----BEGIN PUBLIC KEY-----
    MCowBQYDK2VwAyEAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    -----END PUBLIC KEY-----
```

**User-Only Setting** (`auto_install`):

```yaml
# ~/.config/petstore/config.yaml
preferences:
  updates:
    auto_install: true              # Opt-in to automatic updates
                                    # Can ONLY be set in user config
```

**Release Channels**: Different channels are handled via different binaries or update URLs:

```yaml
# Stable binary
updates:
  update_url: https://releases.petstore.example.com/stable/cli

# Beta binary
updates:
  update_url: https://releases.petstore.example.com/beta/cli
```

---

### Behaviors Section

Locked runtime behaviors that cannot be overridden by users (except in debug mode).

**Location**: Embedded config only (LOCKED)

#### Authentication Behavior

```yaml
behaviors:
  auth:
    type: oauth2                    # none, api_key, oauth2, basic

    # API Key authentication
    api_key:
      header: X-API-Key
      env_var: PETSTORE_API_KEY

    # OAuth2 authentication
    oauth2:
      client_id: petstore-cli
      client_secret: ${OAUTH_CLIENT_SECRET}
      auth_url: https://auth.petstore.example.com/authorize
      token_url: https://auth.petstore.example.com/token
      scopes: [read:pets, write:pets]
      redirect_url: http://localhost:8085/callback

    # Basic authentication
    basic:
      username_env: PETSTORE_USERNAME
      password_env: PETSTORE_PASSWORD
```

#### Caching Behavior

```yaml
behaviors:
  caching:
    spec_ttl: 5m                    # OpenAPI spec cache TTL
    response_ttl: 1m                # API response cache TTL
    directory: ~/.cache/petstore    # Cache directory
    max_size: 100MB                 # Maximum cache size
```

#### Retry Behavior

```yaml
behaviors:
  retry:
    enabled: true
    initial_delay: 1s               # First retry delay
    max_delay: 30s                  # Maximum retry delay
    backoff_multiplier: 2.0         # Exponential backoff multiplier
    retry_on_status: [429, 500, 502, 503, 504]
```

#### Pagination Behavior

```yaml
behaviors:
  pagination:
    max_limit: 100                  # Maximum page size (prevents abuse)
    delay: 100ms                    # Inter-page delay for auto-pagination
```

#### Secrets Behavior

```yaml
behaviors:
  secrets:
    enabled: true

    masking:
      style: partial                # partial, full, hash
      partial_show_chars: 6         # Characters to show in partial mode
      replacement: "***"            # Replacement string

    field_patterns:                 # Auto-detect secret fields
      - "*password*"
      - "*secret*"
      - "*token*"
      - "*key"
      - "*credential*"

    headers:                        # Headers to mask
      - Authorization
      - X-API-Key
      - Cookie
```

---

### Features Section

Feature toggles for optional functionality.

**Location**: Embedded config

```yaml
features:
  config_file: true                 # Enable user config file
  config_file_path: ~/.config/petstore/config.yaml

  interactive_mode: false           # Enable interactive prompts
```

---

## Configuration Priority Chain

Configuration values are resolved in priority order (highest to lowest):

```
1. Environment Variables        (Highest - Runtime overrides)
   ↓
2. Command-Line Flags           (Explicit user intent)
   ↓
3. User Config (preferences)    (User's saved preferences)
   ↓
4. Debug Override              (Only if debug: true in metadata)
   ↓
5. Embedded Config (defaults)   (Developer-provided defaults)
   ↓
6. Built-in Defaults           (Lowest - Hardcoded fallbacks)
```

### Example Priority Resolution

**Embedded config**:
```yaml
defaults:
  output:
    format: json
  http:
    timeout: 30s
```

**User config** (`~/.config/petstore/config.yaml`):
```yaml
preferences:
  output:
    format: yaml
```

**Command execution**:
```bash
# Uses YAML (user config)
petstore pets list

# Uses table (flag overrides user config)
petstore pets list --output table

# Uses CSV (env var overrides everything)
PETSTORE_OUTPUT_FORMAT=csv petstore pets list
```

### Environment Variable Mapping

CliForge automatically maps environment variables to config paths:

| Environment Variable | Config Path | Description |
|---------------------|-------------|-------------|
| `MYCLI_OUTPUT_FORMAT` | `defaults.output.format` | Output format |
| `MYCLI_TIMEOUT` | `defaults.http.timeout` | HTTP timeout |
| `MYCLI_NO_COLOR` | `defaults.output.color` | Disable colors |
| `MYCLI_PRETTY_PRINT` | `defaults.output.pretty_print` | Pretty printing |
| `MYCLI_PAGING` | `defaults.output.paging` | Enable paging |
| `MYCLI_PAGE_LIMIT` | `defaults.pagination.limit` | Page size |
| `MYCLI_RETRY` | `defaults.retry.max_attempts` | Retry attempts |
| `MYCLI_NO_CACHE` | `defaults.caching.enabled` | Disable cache |

**Naming Convention**: `<CLI_NAME>_<CONFIG_PATH>` (uppercase, underscores)

**Example**:
```bash
export PETSTORE_OUTPUT_FORMAT=yaml
export PETSTORE_TIMEOUT=60s
export PETSTORE_PAGE_LIMIT=50

petstore pets list
```

---

## Profiles for Multiple Environments

CliForge doesn't have built-in profile support, but you can achieve multi-environment setups using:

### Method 1: API Environments (Built-in)

Use the `api.environments` section in embedded config:

```yaml
api:
  environments:
    - name: production
      base_url: https://api.petstore.example.com
      default: true

    - name: staging
      base_url: https://staging-api.petstore.example.com
```

**Usage**:
```bash
petstore --env staging pets list
```

### Method 2: Multiple Config Files

Create separate config files for different environments:

```bash
# ~/.config/petstore/production.yaml
preferences:
  http:
    timeout: 30s

# ~/.config/petstore/staging.yaml
preferences:
  http:
    timeout: 60s

# Use with --config flag
petstore --config ~/.config/petstore/staging.yaml pets list
```

### Method 3: Shell Aliases

Create shell aliases for different profiles:

```bash
# ~/.bashrc or ~/.zshrc
alias petstore-prod='petstore --env production'
alias petstore-staging='petstore --env staging'
alias petstore-dev='petstore --env development'

# Usage
petstore-staging pets list
```

### Method 4: Environment-Specific Variables

Use shell environment files:

```bash
# ~/.petstore-production.env
export PETSTORE_ENV=production
export PETSTORE_TIMEOUT=30s

# ~/.petstore-staging.env
export PETSTORE_ENV=staging
export PETSTORE_TIMEOUT=60s

# Usage
source ~/.petstore-staging.env
petstore pets list
```

---

## Debug Mode

Debug mode allows developers to override ANY embedded configuration setting for testing purposes.

**SECURITY WARNING**: Debug mode should NEVER be enabled in production binaries.

### Enabling Debug Mode

**Embedded config**:
```yaml
metadata:
  debug: true                       # ONLY in development builds
```

### Using Debug Overrides

When `debug: true`, users can add a `debug_override` section to their config:

```yaml
# ~/.config/petstore-dev/config.yaml

# Normal preferences (always active)
preferences:
  output:
    format: yaml

# Debug overrides (only active when metadata.debug: true)
debug_override:
  api:
    base_url: http://localhost:8080           # Point to local API
    openapi_url: file://./test-openapi.yaml   # Use local spec

  behaviors:
    auth:
      type: none                              # Disable auth for testing

  branding:
    colors:
      primary: "#FF0000"                      # Red for visibility
```

### Debug Warning Display

When debug mode is active with overrides, a prominent warning is shown on EVERY command:

```
╔════════════════════════════════════════════════════════════════╗
║  🚨 DEBUG MODE ENABLED - SECURITY WARNING                     ║
╠════════════════════════════════════════════════════════════════╣
║  This is a DEBUG BUILD.                                        ║
║  All embedded configuration can be overridden.                 ║
║                                                                ║
║  ⚠️  DO NOT USE IN PRODUCTION                                 ║
║                                                                ║
║  Active debug_override settings (3):                           ║
║  - api.base_url: http://localhost:8080                         ║
║  - api.openapi_url: file://./test-openapi.yaml                 ║
║  - behaviors.auth.type: none                                   ║
╚════════════════════════════════════════════════════════════════╝
```

### Production Build Behavior

When `debug: false` (production build), the `debug_override` section is completely ignored:

```
⚠️ Warning: debug_override section found in config file but ignored
   This is a production build (debug: false)
   debug_override section is only active in debug builds
   Location: ~/.config/petstore/config.yaml
```

---

## Common Configuration Patterns

### Pattern 1: Enterprise Proxy Setup

For corporate environments with proxy servers:

```yaml
# ~/.config/petstore/config.yaml
preferences:
  http:
    proxy: http://proxy.corp.com:8080
    https_proxy: https://proxy.corp.com:8443
    no_proxy: [localhost, "127.0.0.1", ".internal"]

    tls:
      ca_bundle: /etc/ssl/certs/corporate-ca.pem
      # OR for testing (shows warning):
      # insecure_skip_verify: true
```

### Pattern 2: Developer-Friendly Defaults

For developers who prefer verbose output:

```yaml
# ~/.config/petstore/config.yaml
preferences:
  output:
    format: yaml
    pretty_print: true
    color: always

  http:
    timeout: 120s                   # Longer timeout for debugging

  deprecations:
    always_show: true               # See all warnings
```

### Pattern 3: CI/CD Optimized

For automated environments:

```yaml
# ~/.config/petstore/config.yaml
preferences:
  output:
    format: json
    color: never
    paging: false

  http:
    timeout: 300s                   # Long timeout for slow networks

  retry:
    max_attempts: 5                 # More retries for reliability

  deprecations:
    always_show: false
    min_severity: critical          # Only critical warnings
```

### Pattern 4: Minimal Configuration

For users who want defaults:

```yaml
# ~/.config/petstore/config.yaml
preferences:
  output:
    format: table                   # Human-readable tables
```

### Pattern 5: Local Development

For testing against local API:

```yaml
# ~/.config/petstore-dev/config.yaml (debug build only)
debug_override:
  api:
    base_url: http://localhost:8080
    openapi_url: file://./openapi.yaml

  behaviors:
    auth:
      type: none
    caching:
      spec_ttl: 1s                  # Short TTL for rapid iteration
```

---

## Configuration Validation

CliForge validates configuration files on load.

### Built-in Validation

```bash
# Validate configuration
petstore config validate

# Output on success:
✓ Configuration is valid
✓ All required fields present
✓ No unknown fields
✓ All values in valid format

# Output on error:
✗ Configuration validation failed:

  Line 12: defaults.http.timeout: Invalid duration format
  Line 34: defaults.output.format: Invalid value 'xml' (allowed: json, yaml, table, csv)
  Line 56: preferences.unknown_field: Unknown configuration field
```

### Validation Rules

**Metadata Validation**:
- `name`: Required, lowercase, alphanumeric with hyphens, max 50 chars
- `version`: Required, semantic version format (X.Y.Z)
- `description`: Required, 10-200 characters

**API Validation**:
- `openapi_url`: Required, valid URL or file path
- `base_url`: Required, valid HTTP/HTTPS URL

**Colors Validation**:
- Must be 6-digit hex codes (#RRGGBB)
- Case-insensitive

**Duration Validation**:
- Format: `<number><unit>` (e.g., `30s`, `5m`, `24h`, `7d`)
- Units: `s` (seconds), `m` (minutes), `h` (hours), `d` (days)

### Common Validation Errors

**Error: Invalid duration**
```yaml
# Wrong
defaults:
  http:
    timeout: 30                     # Missing unit

# Correct
defaults:
  http:
    timeout: 30s
```

**Error: Invalid color**
```yaml
# Wrong
branding:
  colors:
    primary: "red"                  # Named colors not supported

# Correct
branding:
  colors:
    primary: "#FF0000"
```

**Error: Unknown field**
```yaml
# Wrong
preferences:
  custom_setting: true              # Not a valid preference

# Check documentation for valid fields
```

---

## Best Practices

### For CLI Developers

1. **Never enable debug mode in production**
   ```yaml
   metadata:
     debug: false                   # Production builds
   ```

2. **Lock security-critical settings**
   - Keep all `api.*` settings in embedded config
   - Don't allow user overrides of auth endpoints
   - Don't allow telemetry URL overrides

3. **Provide sensible defaults**
   ```yaml
   defaults:
     http:
       timeout: 30s                 # Balance usability and performance
     output:
       format: json                 # Standard format
       color: auto                  # Respect terminal capabilities
   ```

4. **Document environment variables**
   - List all supported env vars in CLI help
   - Follow naming convention: `MYCLI_<CONFIG_PATH>`

5. **Use semantic versioning**
   ```yaml
   metadata:
     version: 1.2.3                 # Major.Minor.Patch
   ```

6. **Validate thoroughly**
   - Validate embedded config during CLI generation
   - Validate user config on load
   - Provide helpful error messages

7. **Separate debug and production binaries**
   ```bash
   # Production: mycli
   # Debug: mycli-dev
   ```

### For CLI Users

1. **Keep user config minimal**
   - Only override what you need
   - Let defaults handle the rest

2. **Use environment variables for sensitive data**
   ```bash
   export PETSTORE_API_KEY=secret123
   # Don't put secrets in config.yaml
   ```

3. **Secure your config file**
   ```bash
   chmod 600 ~/.config/petstore/config.yaml
   ```

4. **Use profiles via shell aliases**
   ```bash
   alias petstore-prod='petstore --env production'
   alias petstore-staging='petstore --env staging'
   ```

5. **Test configuration changes**
   ```bash
   # Validate before using
   petstore config validate

   # View effective config
   petstore config show --effective
   ```

6. **Understand the priority chain**
   - Flags override everything
   - Env vars override config files
   - User config overrides embedded config

7. **Never use debug builds in production**
   - Debug overrides compromise security
   - Only use for local development

8. **Review update settings carefully**
   ```yaml
   preferences:
     updates:
       auto_install: true           # Only if you trust the provider
   ```

### Security Best Practices

1. **For Developers**:
   - Never embed API keys or secrets in config
   - Use environment variables for credentials
   - Sign binary updates with Ed25519 keys
   - Lock all security-critical settings
   - Disable debug mode in production

2. **For Users**:
   - Review auto-update settings before enabling
   - Protect config files with appropriate permissions
   - Use environment variables for credentials
   - Verify binary signatures after updates
   - Never share config files with secrets

---

## Example: Complete User Configuration

Here's a comprehensive example user configuration:

```yaml
# ~/.config/petstore/config.yaml

# User preferences (always active)
preferences:
  # Output preferences
  output:
    format: yaml                    # Prefer YAML
    color: always                   # Always use colors
    pretty_print: true
    paging: true

  # HTTP settings
  http:
    timeout: 60s                    # Longer timeout
    proxy: http://proxy.corp.com:8080
    tls:
      ca_bundle: /etc/ssl/certs/corp-ca.pem

  # Pagination
  pagination:
    limit: 50                       # More items per page

  # Retry behavior
  retry:
    max_attempts: 5                 # More retries

  # Deprecation warnings
  deprecations:
    always_show: true
    min_severity: warning

  # Caching
  caching:
    enabled: true

  # Telemetry (user opt-in)
  telemetry:
    enabled: false                  # Opt out

  # Updates (user opt-in)
  updates:
    auto_install: false             # Manual updates only

# Debug overrides (only active in debug builds)
debug_override:
  api:
    base_url: http://localhost:8080
    openapi_url: file://./openapi.yaml

  behaviors:
    auth:
      type: none
```

---

## Troubleshooting

### Config File Not Found

```
Warning: failed to load user config: file not found
```

**Solution**: Config file is optional. Create one if needed:

```bash
mkdir -p ~/.config/petstore
petstore config edit
```

### Invalid Configuration

```
✗ Configuration validation failed:
  Line 5: Invalid duration format
```

**Solution**: Fix the syntax error and validate:

```bash
petstore config validate
```

### Debug Override Ignored

```
⚠️ Warning: debug_override section found but ignored
```

**Solution**: This is expected in production builds. Use a debug build for testing.

### Permissions Error

```
Error: failed to write config file: permission denied
```

**Solution**: Check file permissions:

```bash
chmod 600 ~/.config/petstore/config.yaml
```

### Environment Variable Not Working

**Solution**: Check variable name matches convention:

```bash
# Wrong
export OUTPUT_FORMAT=yaml

# Correct
export PETSTORE_OUTPUT_FORMAT=yaml
```

---

## Summary

CliForge's configuration system provides:

- **Embedded defaults** for out-of-the-box functionality
- **User preferences** for personalization
- **Environment variables** for runtime overrides
- **Command-line flags** for explicit control
- **Debug mode** for development and testing
- **XDG compliance** for proper file organization
- **Security boundaries** to prevent misuse

By understanding the configuration priority chain and using the right tools for each scenario, you can effectively customize your CLI experience while maintaining security and stability.

For authentication-specific configuration, see the [Authentication Guide](./user-guide-authentication.md).
