# Configuration DSL Specification

## Overview

The configuration DSL is a YAML-based declarative language for defining branded CLI tools. It specifies branding, behavior, API endpoints, and features.

**File Name**: `cli-config.yaml` or `<project-name>.yaml`

---

## Table of Contents

1. [Complete Schema](#complete-schema)
2. [Section Reference](#section-reference)
3. [Examples](#examples)
4. [Validation Rules](#validation-rules)
5. [Advanced Configurations](#advanced-configurations)

---

## Complete Schema

```yaml
# ============================================================================
# CLI Configuration DSL - Complete Schema
# ============================================================================

# ----------------------------------------------------------------------------
# Metadata Section
# ----------------------------------------------------------------------------
metadata:
  # Required: CLI binary name (must be valid filename)
  name: string (required)

  # Required: Semantic version
  version: string (required, semver format)

  # Required: Short description
  description: string (required)

  # Optional: Long description (for --help)
  long_description: string

  # Optional: Author information
  author:
    name: string
    email: string
    url: string

  # Optional: License
  license: string

  # Optional: Homepage URL
  homepage: string

  # Optional: Bug report URL
  bugs_url: string

  # Optional: Documentation URL
  docs_url: string

  # Optional: Debug mode (default: false)
  # When true, allows ALL config overrides via debug_override section in user config
  # SECURITY: Should ONLY be true in development/testing builds
  # Production builds should ALWAYS set this to false
  debug: boolean (default: false)

# ----------------------------------------------------------------------------
# Branding Section
# ----------------------------------------------------------------------------
branding:
  # Optional: Color scheme
  colors:
    primary: string (hex color)
    secondary: string (hex color)
    success: string (hex color)
    warning: string (hex color)
    error: string (hex color)
    info: string (hex color)

  # Optional: ASCII art banner (displayed on first run or with --version)
  ascii_art: string (multiline)

  # Optional: Custom prompts
  prompts:
    command: string (default: "$")
    error: string (default: "✗")
    success: string (default: "✓")
    warning: string (default: "⚠")
    info: string (default: "ℹ")


# ----------------------------------------------------------------------------
# API Section
# ----------------------------------------------------------------------------
api:
  # Required: OpenAPI spec URL or path (single source of truth)
  openapi_url: string (required, URL or file path)

  # Required: API base URL (default for all operations)
  # Note: Individual operations can override with absolute URLs via x-cli-workflow
  base_url: string (required, URL)

  # Optional: API version (if not in base_url)
  version: string

  # Optional: Multiple environments
  environments:
    - name: string
      openapi_url: string
      base_url: string
      default: boolean

  # Optional: Default headers (applied to all requests)
  default_headers:
    key: value

  # Optional: User agent
  user_agent: string

  # Optional: Telemetry endpoint URL (where to send usage data)
  telemetry_url: string (URL)

  # SECURITY NOTE: The ENTIRE api section is LOCKED to embedded config
  # Users cannot override any api.* settings (except in debug mode)
  # This prevents pointing CLI to wrong APIs or redirecting telemetry

# ----------------------------------------------------------------------------
# Defaults Section (User-Overridable Settings)
# ----------------------------------------------------------------------------
defaults:
  # HTTP client settings
  http:
    timeout: duration (default: "30s")  # Request timeout

  # Caching settings
  caching:
    enabled: boolean (default: true)  # Enable response caching

  # Pagination settings
  pagination:
    limit: integer (default: 20)  # Default page size (enforced by behaviors.pagination.max_limit)

  # Output settings
  output:
    format: string (default: "json")  # json, yaml, table, csv
    pretty_print: boolean (default: true)
    color: string (default: "auto")  # auto, always, never
    paging: boolean (default: true)  # Use pager for long output

  # Deprecation warning settings
  deprecations:
    always_show: boolean (default: false)  # Show once then cache (except critical/removed - always shown)
    min_severity: string (default: "info")  # info, warning, urgent, critical, removed

  # Retry settings
  retry:
    max_attempts: integer (default: 3)  # Number of retry attempts

  # NOTE: Users can override these defaults via preferences section in their config file
  # Precedence: ENV > Flag > User Config (preferences) > Embedded Config (defaults) > Built-in Default

# ----------------------------------------------------------------------------
# Updates Section
# ----------------------------------------------------------------------------
updates:
  # Required: Enable update checking
  enabled: boolean (default: true)

  # Required: Update server URL (defines the "channel" via URL)
  # Examples:
  #   stable:  https://releases.acme.com/stable/cli
  #   beta:    https://releases.acme.com/beta/cli
  #   nightly: https://releases.acme.com/nightly/cli
  update_url: string (required if enabled)

  # Optional: Check interval
  check_interval: duration (default: "24h")

  # Optional: Public key for signature verification (PEM format)
  public_key: string (multiline PEM)

  # SECURITY NOTE: auto_install is NOT allowed in embedded config
  # It can ONLY be set by users in their local config file
  # See "User Configuration File" section below

  # SECURITY NOTE: The ENTIRE updates section is LOCKED to embedded config
  # Users cannot override update_url, check_interval, or public_key

# ----------------------------------------------------------------------------
# Behaviors Section (LOCKED - Not User-Overridable)
# ----------------------------------------------------------------------------
behaviors:
  # Authentication configuration
  auth:
    # Auth type: none, api_key, oauth2, basic
    type: string (required)

    # API Key auth
    api_key:
      header: string (e.g., "X-API-Key")
      env_var: string (e.g., "MY_API_KEY")

    # OAuth2 auth
    oauth2:
      client_id: string
      client_secret: string
      auth_url: string
      token_url: string
      scopes: [string]
      redirect_url: string

    # Basic auth
    basic:
      username_env: string
      password_env: string

  # Caching configuration (LOCKED)
  # Note: defaults.caching.enabled is user-overridable
  caching:
    # OpenAPI spec cache TTL (LOCKED)
    spec_ttl: duration (default: "5m")

    # Response cache TTL (LOCKED)
    response_ttl: duration (default: "1m")

    # Cache directory (LOCKED)
    directory: string (default: OS-specific)

    # Max cache size (LOCKED)
    max_size: string (e.g., "100MB")

  # Retry logic (LOCKED)
  # Note: defaults.retry.max_attempts is user-overridable
  retry:
    enabled: boolean (default: true)
    initial_delay: duration (default: "1s")
    max_delay: duration (default: "30s")
    backoff_multiplier: float (default: 2.0)
    retry_on_status: [integer] (e.g., [429, 500, 502, 503, 504])

  # Pagination (LOCKED)
  # Note: defaults.pagination.limit is user-overridable
  pagination:
    max_limit: integer  # Maximum page size (prevents API abuse)
    delay: duration (default: "100ms")  # Inter-page delay for auto-pagination

  # Secrets & sensitive data handling
  secrets:
    enabled: boolean (default: true)

    # Masking strategy
    masking:
      style: string (partial|full|hash, default: partial)
      partial_show_chars: integer (default: 6)
      replacement: string (default: "***")

    # Field name patterns (auto-detect secrets)
    field_patterns: [string] (default: ["*password*", "*secret*", "*token*", "*key", "*credential*", "auth*", "*bearer*"])

    # Value patterns (regex for detecting secret-like values)
    value_patterns:
      - name: string
        pattern: string (regex)
        enabled: boolean (default: true)

    # Explicit field paths to mask (JSONPath-like)
    explicit_fields: [string]

    # Headers to mask
    headers: [string] (default: ["Authorization", "X-API-Key", "X-Auth-Token", "Cookie", "Set-Cookie"])

    # Where to apply masking
    mask_in:
      stdout: boolean (default: true)
      stderr: boolean (default: true)
      logs: boolean (default: true)
      audit: boolean (default: false)  # Don't mask in audit (uses hashing instead)
      debug_output: boolean (default: true)

  # Built-in commands configuration
  builtin_commands:
    # version command
    version:
      enabled: boolean (default: true)
      style: string (subcommand|flag|hybrid, default: hybrid)
      flags: [string] (default: ["--version", "-V"])
      show_api_version: boolean (default: true)
      show_build_info: boolean (default: true)

    # help command
    help:
      enabled: boolean (default: true)
      style: string (subcommand|flag|hybrid, default: hybrid)
      flags: [string] (default: ["--help", "-h"])
      show_examples: boolean (default: true)
      show_aliases: boolean (default: true)

    # info command (CLI + API details)
    info:
      enabled: boolean (default: true)
      style: string (default: subcommand)
      check_api_health: boolean (default: true)
      show_config_paths: boolean (default: true)

    # config command
    config:
      enabled: boolean (default: true)
      style: string (default: subcommand)
      command_name: string (default: "config")
      allow_edit: boolean (default: true)
      subcommands: [string] (default: [show, get, set, unset, edit, path])

    # completion command
    completion:
      enabled: boolean (default: true)
      shells: [string] (default: [bash, zsh, fish, powershell])

    # update command (self-update)
    update:
      enabled: boolean (default: false)
      auto_check: boolean (default: false)
      require_confirmation: boolean (default: true)
      show_changelog: boolean (default: true)

    # changelog command
    changelog:
      enabled: boolean (default: true)
      show_binary_changes: boolean (default: true)   # Show CLI binary changelog
      show_api_changes: boolean (default: true)      # Show API changelog
      binary_changelog:
        source: string (embedded|url|api, default: embedded)
        url: string (if source: url)
      api_changelog:
        source: string (embedded|url|api, default: api)  # From x-cli-changelog
        url: string (OpenAPI spec URL)
      default_limit: integer (default: 10)  # Number of versions to show

    # deprecations command
    deprecations:
      enabled: boolean (default: true)
      show_binary_deprecations: boolean (default: true)  # Show CLI deprecations
      show_api_deprecations: boolean (default: true)     # Show API deprecations
      show_by_default: boolean (default: true)
      allow_scan: boolean (default: true)
      allow_auto_fix: boolean (default: false)

    # cache command
    cache:
      enabled: boolean (default: true)
      subcommands: [string] (default: [info, clear])

    # auth command
    auth:
      enabled: boolean (default: true)
      subcommands: [string] (default: [login, logout, status, refresh])

  # Global flags configuration
  global_flags:
    # config flag
    config:
      enabled: boolean (default: true)
      flag: string (default: "--config")
      short: string (default: "-c")
      env_var: string
      description: string

    # profile flag (selects section in config file, like AWS CLI)
    profile:
      enabled: boolean (default: true)
      flag: string (default: "--profile")
      short: string
      env_var: string
      description: string
      default: string (default: "default")

    # region flag (common for cloud/datacenter APIs)
    region:
      enabled: boolean (default: true)
      flag: string (default: "--region")
      short: string (default: "-r")
      env_var: string
      description: string
      allowed_values: [string]  # Optional: restrict to specific regions

    # output flag
    output:
      enabled: boolean (default: true)
      flag: string (default: "--output")
      short: string (default: "-o")
      env_var: string
      description: string
      default: string (default: "json")
      allowed_values: [string]

    # verbose flag
    verbose:
      enabled: boolean (default: true)
      flag: string (default: "--verbose")
      short: string (default: "-v")
      env_var: string
      description: string
      repeatable: boolean (default: true)

    # quiet flag
    quiet:
      enabled: boolean (default: true)
      flag: string (default: "--quiet")
      short: string (default: "-q")
      env_var: string
      description: string
      conflicts_with: [string] (default: [verbose])

    # debug flag
    debug:
      enabled: boolean (default: true)
      flag: string (default: "--debug")
      short: string
      env_var: string
      description: string

    # no-color flag
    no_color:
      enabled: boolean (default: true)
      flag: string (default: "--no-color")
      short: string
      env_var: string (default: "NO_COLOR")
      description: string

    # HTTP client flags
    timeout:
      enabled: boolean (default: true)
      flag: string (default: "--timeout")
      short: string
      env_var: string
      description: string
      default: string (default: "30s")

    retry:
      enabled: boolean (default: true)
      flag: string (default: "--retry")
      short: string
      env_var: string
      description: string
      default: integer (default: 3)

    no_cache:
      enabled: boolean (default: true)
      flag: string (default: "--no-cache")
      short: string
      env_var: string
      description: string

    yes:
      enabled: boolean (default: true)
      flag: string (default: "--yes")
      short: string (default: "-y")
      env_var: string
      description: string (default: "Skip confirmation prompts")
      default: boolean (default: false)

    # Custom global flags
    custom:
      - name: string
        flag: string
        short: string
        env_var: string
        description: string
        type: string (string|int|bool|duration)
        default: any
        required: boolean
        allowed_values: [string]
        sensitive: boolean
        conflicts_with: [string]

# ----------------------------------------------------------------------------
# Features Section
# ----------------------------------------------------------------------------
features:
  # Config file support
  config_file: boolean (default: true)
  config_file_path: string

  # Interactive mode
  interactive_mode: boolean (default: false)

# NOTE: The following features were REMOVED:
# - telemetry: Moved to api.telemetry_url (endpoint) and preferences.telemetry.enabled (user opt-in)
# - shell_completion: Always enabled, not configurable
# - verbose/debug: These are CLI flags (--verbose, --debug), not features
# - offline_mode: Redundant with caching behavior
# - validate_requests/validate_responses: Always validate requests, never validate responses

# NOTE: The following sections were REMOVED from v0.8.0:
#
# - commands: Users should use shell aliases instead of embedding command customization
#   Example: alias mycli-users-ls='mycli users list'
#
# - hooks: Users should wrap the CLI with shell scripts for lifecycle events
#   Example: mycli users list && echo "Success!" || echo "Failed"
#
# - plugins: Too complex for v1.0, security concerns with arbitrary code execution
#   May be reconsidered for future versions with proper sandboxing
```

---

## Section Reference

### Metadata Section

**Purpose**: Identify and describe the CLI tool

#### Required Fields

```yaml
metadata:
  name: my-api-cli
  version: 1.0.0
  description: CLI for My API
```

#### All Fields

```yaml
metadata:
  name: acme-cli
  version: 2.1.0
  description: ACME Corporation API Command Line Interface
  long_description: |
    The ACME CLI provides a powerful command-line interface
    to interact with all ACME Corporation APIs. It supports
    user management, resource provisioning, and analytics.

  author:
    name: ACME DevTools Team
    email: devtools@acme.com
    url: https://github.com/acme/cli

  license: MIT
  homepage: https://acme.com/cli
  bugs_url: https://github.com/acme/cli/issues
  docs_url: https://docs.acme.com/cli
```

---

### Branding Section

**Purpose**: Customize the look and feel

#### Color Schemes

```yaml
branding:
  colors:
    primary: "#FF6B6B"      # Main brand color
    secondary: "#4ECDC4"    # Accent color
    success: "#51CF66"      # Success messages
    warning: "#FFC078"      # Warnings
    error: "#FF6B6B"        # Errors
    info: "#74C0FC"         # Info messages
```

**Color Reference**:
- Use 6-digit hex codes
- Alpha channel not supported
- Must be valid hex color

#### ASCII Art

```yaml
branding:
  ascii_art: |
    ╔═══════════════════════════════════════╗
    ║                                       ║
    ║   █████╗  ██████╗███╗   ███╗███████╗ ║
    ║  ██╔══██╗██╔════╝████╗ ████║██╔════╝ ║
    ║  ███████║██║     ██╔████╔██║█████╗   ║
    ║  ██╔══██║██║     ██║╚██╔╝██║██╔══╝   ║
    ║  ██║  ██║╚██████╗██║ ╚═╝ ██║███████╗ ║
    ║  ╚═╝  ╚═╝ ╚═════╝╚═╝     ╚═╝╚══════╝ ║
    ║                                       ║
    ║     Command Line Interface v2.0       ║
    ║                                       ║
    ╚═══════════════════════════════════════╝
```

**ASCII Art Guidelines**:
- Keep under 80 chars wide
- Use box-drawing characters for clean rendering
- Test on different terminals
- Avoid emoji (not universally supported)

#### Custom Prompts

```yaml
branding:
  prompts:
    command: "acme>"
    error: "✗"
    success: "✓"
    warning: "⚠"
    info: "ℹ"
```

#### Theme

```yaml
branding:
  theme:
    name: auto  # auto, light, dark
    syntax_highlighting: true
```

---

### API Section

**Purpose**: Define API endpoints and connection settings

#### Basic Configuration

```yaml
api:
  openapi_url: https://api.acme.com/openapi.yaml
  base_url: https://api.acme.com
  timeout: 30s
```

#### Multiple Environments

```yaml
api:
  openapi_url: https://api.acme.com/openapi.yaml
  base_url: https://api.acme.com

  environments:
    - name: production
      openapi_url: https://api.acme.com/openapi.yaml
      base_url: https://api.acme.com
      default: true

    - name: staging
      openapi_url: https://staging-api.acme.com/openapi.yaml
      base_url: https://staging-api.acme.com

    - name: development
      openapi_url: https://dev-api.acme.com/openapi.yaml
      base_url: https://dev-api.acme.com

    - name: local
      openapi_url: file://./openapi.yaml
      base_url: http://localhost:8080
```

**Usage**:
```bash
# Use default (production)
acme-cli users list

# Use specific environment
acme-cli --env staging users list

# Switch default environment
acme-cli config set-env staging
```

#### Default Headers

```yaml
api:
  default_headers:
    X-Client-Version: "2.1.0"
    Accept: "application/json"
    User-Agent: "acme-cli/2.1.0"
```

---

### Updates Section

**Purpose**: Configure binary self-update behavior

**IMPORTANT SECURITY BOUNDARY**: The embedded config can only enable update **checking**. Automatic installation must be explicitly opted-in by users in their local config.

#### Basic Configuration (Embedded)

```yaml
updates:
  enabled: true
  update_url: https://releases.acme.com/stable/cli
  check_interval: 24h
```

#### Full Configuration (Embedded)

```yaml
updates:
  enabled: true
  update_url: https://releases.acme.com/stable/cli
  check_interval: 24h

  # Ed25519 public key for signature verification
  public_key: |
    -----BEGIN PUBLIC KEY-----
    MCowBQYDK2VwAyEAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    -----END PUBLIC KEY-----
```

**What This Does**:
- ✅ Checks for updates on startup (non-blocking)
- ✅ Notifies user when updates are available
- ✅ Provides `my-cli update` command to install
- ❌ **NEVER** auto-installs without user permission

**Auto-install requires explicit user opt-in** - see User Configuration section below.

#### Release Channels via URL

Different release channels are handled via different `update_url` values or separate binaries:

**Option 1: URL-based channels** (single binary)
```yaml
# Stable binary
updates:
  update_url: https://releases.acme.com/stable/cli

# Beta binary (different embedded config)
updates:
  update_url: https://releases.acme.com/beta/cli

# User can override in ~/.config/acme-cli/config.yaml:
updates:
  update_url: https://releases.acme.com/nightly/cli  # Switch to nightly
```

**Option 2: Separate binaries** (recommended)
```bash
# Install stable
curl -L https://releases.acme.com/install-stable.sh | sh

# Install beta
curl -L https://releases.acme.com/install-beta.sh | sh

# Each has different update_url embedded
```

**Why no `channel` field?**:
- Simpler: channel is implicit in the `update_url`
- More flexible: Users can point to custom update servers
- Less magic: No special channel-switching logic needed

---

### Behaviors Section

**Purpose**: Configure runtime behavior

#### Authentication

**API Key Authentication**:
```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: ACME_API_KEY
```

**OAuth2 Authentication**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: cli-client-id
      client_secret: ${OAUTH_CLIENT_SECRET}  # From env var
      auth_url: https://auth.acme.com/oauth/authorize
      token_url: https://auth.acme.com/oauth/token
      scopes:
        - api:read
        - api:write
      redirect_url: http://localhost:8080/callback
```

**Basic Authentication**:
```yaml
behaviors:
  auth:
    type: basic
    basic:
      username_env: ACME_USERNAME
      password_env: ACME_PASSWORD
```

**No Authentication**:
```yaml
behaviors:
  auth:
    type: none
```

#### Caching

```yaml
behaviors:
  caching:
    enabled: true
    spec_ttl: 5m        # Cache OpenAPI spec for 5 minutes
    response_ttl: 1m    # Cache API responses for 1 minute
    directory: ~/.cache/acme-cli
    max_size: 100MB
```

**Cache Behavior**:
- `spec_ttl`: How long to cache the OpenAPI specification
- `response_ttl`: How long to cache GET responses (POST/PUT/DELETE never cached)
- Respects HTTP cache headers (`Cache-Control`, `ETag`)
- `--refresh` flag bypasses cache

#### Rate Limiting

```yaml
behaviors:
  rate_limiting:
    enabled: true
    requests_per_minute: 60
    burst: 10
```

**Implementation**: Token bucket algorithm
- Allows `burst` requests immediately
- Refills at `requests_per_minute` rate

#### Retry Logic

```yaml
behaviors:
  retry:
    enabled: true
    max_attempts: 3
    initial_delay: 1s
    max_delay: 30s
    backoff_multiplier: 2.0
    retry_on_status:
      - 429  # Rate limited
      - 500  # Internal server error
      - 502  # Bad gateway
      - 503  # Service unavailable
      - 504  # Gateway timeout
```

**Retry Schedule**:
- Attempt 1: Immediate
- Attempt 2: After `initial_delay` (1s)
- Attempt 3: After `initial_delay * backoff_multiplier` (2s)
- Attempt 4: After previous * multiplier, capped at `max_delay`

#### Output

```yaml
behaviors:
  output:
    default_format: json  # json, yaml, table, csv
    pretty_print: true
    color: auto          # auto, always, never
    paging: true
    pager: less -R
```

**Format Examples**:

**JSON**:
```json
{
  "id": 123,
  "name": "John Doe",
  "email": "john@example.com"
}
```

**YAML**:
```yaml
id: 123
name: John Doe
email: john@example.com
```

**Table**:
```
ID   NAME      EMAIL
123  John Doe  john@example.com
```

**CSV**:
```
id,name,email
123,John Doe,john@example.com
```

#### Pagination

```yaml
behaviors:
  pagination:
    default_limit: 20
    max_limit: 100
    auto_paginate: true  # Automatically fetch all pages
```

**Auto-pagination**:
```bash
# Without auto-paginate: Returns 20 results
acme-cli users list

# With auto-paginate: Fetches all pages automatically
acme-cli users list
```

#### Notifications

```yaml
behaviors:
  notifications:
    show_changelog: true      # Show API changes on startup
    show_deprecations: true   # Warn about deprecated endpoints
    check_interval: 7d        # Check for changes every 7 days
```

---

### Features Section

**Purpose**: Enable/disable features

#### Telemetry

```yaml
features:
  telemetry:
    enabled: false  # Opt-in, not opt-out
    endpoint: https://telemetry.acme.com/events
    anonymous: true  # No personally identifiable information
```

**What's Collected** (if enabled):
- Command usage (e.g., `users list` executed)
- Error frequency (without details)
- Performance metrics
- CLI version, OS, architecture

**Not Collected**:
- API keys, tokens, credentials
- API request/response data
- User data

#### Shell Completion

```yaml
features:
  shell_completion: true
```

**Installation**:
```bash
# Bash
acme-cli completion bash > /etc/bash_completion.d/acme-cli

# Zsh
acme-cli completion zsh > ~/.zsh/completion/_acme-cli

# Fish
acme-cli completion fish > ~/.config/fish/completions/acme-cli.fish
```

#### Config File

```yaml
features:
  config_file: true
  config_file_path: ~/.config/acme-cli/config.yaml
```

**User Config** (overrides defaults):
```yaml
# ~/.config/acme-cli/config.yaml
api:
  base_url: https://staging-api.acme.com

behaviors:
  output:
    default_format: yaml
```

#### Validation

```yaml
features:
  validate_requests: true   # Validate against OpenAPI schema before sending
  validate_responses: false # Validate responses (slower, useful for debugging)
```

---

### Commands Section

**Purpose**: Customize command behavior

#### Global Flags

```yaml
commands:
  global_flags:
    - name: org-id
      short: o
      description: Organization ID
      type: string
      required: false

    - name: verbose
      short: v
      description: Verbose output
      type: bool
      default: false
```

**Usage**:
```bash
acme-cli --org-id acme-123 users list
acme-cli -o acme-123 users list
```

#### Hidden Commands

```yaml
commands:
  hidden:
    - internalDebugEndpoint
    - experimentalFeature
```

#### Custom Aliases

```yaml
commands:
  aliases:
    listUsers:
      - users
      - list-all-users
```

**Usage**:
```bash
acme-cli listUsers          # Original
acme-cli users              # Alias 1
acme-cli list-all-users     # Alias 2
```

#### Custom Commands

```yaml
commands:
  custom:
    - name: deploy
      description: Deploy to production
      script: |
        #!/bin/bash
        echo "Deploying..."
        git push production main
```

---

### Hooks Section

**Purpose**: Execute scripts at lifecycle events

#### Command Hooks

```yaml
hooks:
  pre_command: echo "Executing command..."
  post_command: echo "Command completed"

  operations:
    createUser:
      pre: echo "Creating user..."
      post: echo "User created successfully"

  on_error: |
    echo "Error occurred: $ERROR_MESSAGE"
    echo "Request ID: $REQUEST_ID"
```

**Available Variables**:
- `$COMMAND`: Command being executed
- `$ERROR_MESSAGE`: Error message (on_error only)
- `$REQUEST_ID`: API request ID
- `$STATUS_CODE`: HTTP status code
- `$RESPONSE`: API response body

#### Lifecycle Hooks

```yaml
hooks:
  on_first_run: |
    echo "Welcome to ACME CLI!"
    echo "Run 'acme-cli auth login' to get started"

  on_update: |
    echo "Updated to version $NEW_VERSION"
    echo "Changelog: $CHANGELOG_URL"
```

---

## User Configuration File

**Purpose**: Allow users to override embedded defaults with their own preferences

**IMPORTANT**: The user configuration file is the **only** place certain sensitive settings can be configured, such as `auto_install` for updates.

### File Location

**Follows XDG Base Directory Specification**

**Linux/macOS**:
```
$XDG_CONFIG_HOME/<cli-name>/config.yaml
```
Default (if `XDG_CONFIG_HOME` not set): `~/.config/<cli-name>/config.yaml`

**Windows**:
```
%APPDATA%\<cli-name>\config.yaml
```

**Other File Locations** (XDG-compliant):

**Cache** (OpenAPI specs, HTTP responses):
```
$XDG_CACHE_HOME/<cli-name>/
```
Default: `~/.cache/<cli-name>/`

**Data** (logs, audit logs):
```
$XDG_DATA_HOME/<cli-name>/
```
Default: `~/.local/share/<cli-name>/`

**State** (last update check timestamp, etc.):
```
$XDG_STATE_HOME/<cli-name>/
```
Default: `~/.local/state/<cli-name>/`

**Runtime** (PID files, sockets - if needed):
```
$XDG_RUNTIME_DIR/<cli-name>/
```
Note: `XDG_RUNTIME_DIR` has no default, skip if not set

#### XDG Environment Variables Reference

| XDG Variable | Default | Purpose | CLI Usage |
|--------------|---------|---------|-----------|
| `XDG_CONFIG_HOME` | `~/.config` | User config files | `config.yaml`, user preferences |
| `XDG_CACHE_HOME` | `~/.cache` | Cached data (can be deleted) | OpenAPI specs, HTTP responses |
| `XDG_DATA_HOME` | `~/.local/share` | User data files | Audit logs, telemetry data |
| `XDG_STATE_HOME` | `~/.local/state` | State data | Last update check, session info |
| `XDG_RUNTIME_DIR` | *(no default)* | Runtime files | PID files, sockets (if needed) |

**Implementation Notes**:
- Always check environment variable first, fall back to default
- Create directories with appropriate permissions (700 for config, 755 for cache)
- Respect user's XDG customization
- On macOS, same behavior (XDG spec is cross-platform)
- On Windows, use `%APPDATA%` and `%LOCALAPPDATA%` equivalents

### File Structure

The user config file has **two main sections**:

1. **`preferences`** - Overrides the `defaults` section from embedded config (always active)
2. **`debug_override`** - Overrides ANY embedded config setting (only active when `metadata.debug: true`)

**Basic structure**:
```yaml
# ~/.config/my-cli/config.yaml

# ===================================================================
# Normal user preferences (work in ALL builds)
# ===================================================================
preferences:
  http:
    timeout: 60s              # Override embedded default
    proxy: http://proxy.corp.com:8080
    ca_bundle: /etc/ssl/certs/corp-ca.pem

  output:
    format: yaml              # Prefer YAML over JSON
    color: always             # Always use colors

  pagination:
    limit: 50                 # Increase from default 20

  deprecations:
    always_show: true         # Always show deprecation warnings
    min_severity: warning     # Only show warning+ severity

  retry:
    max_attempts: 5           # More retries for flaky networks

  telemetry:
    enabled: false            # User opts out (defaults to false)

# ===================================================================
# Debug-only overrides (ONLY work when metadata.debug: true)
# ===================================================================
debug_override:
  # Override embedded API configuration (normally locked)
  api:
    base_url: http://localhost:8080       # Point to local test API
    openapi_url: file://./test-openapi.yaml

  # Override embedded metadata (normally locked)
  metadata:
    name: mycli-custom

  # Override embedded branding (normally locked)
  branding:
    colors:
      primary: "#FF0000"      # Red for local testing visibility

  # Override any other embedded settings
  behaviors:
    auth:
      type: none              # Disable auth for local testing
```

**Behavior**:
- **Production build** (`debug: false`): `debug_override` section **IGNORED** entirely
  - CLI shows warning: "⚠️ debug_override section in config ignored (not a debug build)"
- **Debug build** (`debug: true`): `debug_override` section **APPLIED**
  - Overrides embedded config
  - Security warning displayed on EVERY command
  - Shows which overrides are active

### User-Only Settings

These settings can **ONLY** be set in the user config file, never in embedded config:

#### 1. Auto-Install for Updates

```yaml
# ~/.config/my-cli/config.yaml
preferences:
  updates:
    auto_install: true  # ONLY allowed in user config
```

**Why**: Allowing developers to enable `auto_install` in the embedded config would mean:
- Developers could push arbitrary code to user machines without consent
- Users lose control over when/if binaries are updated
- Security vulnerability if update server is compromised

**The user must explicitly opt-in** to auto-updates in their own config file.

#### 2. Telemetry Opt-In

```yaml
# ~/.config/my-cli/config.yaml
preferences:
  telemetry:
    enabled: false  # User controls opt-in (defaults to false)
```

**Why**: The embedded config defines WHERE telemetry is sent (`api.telemetry_url` - locked), but only the user can enable WHETHER to send it.

### Common User Overrides

#### Change Output Preferences

```yaml
# ~/.config/acme-cli/config.yaml
preferences:
  output:
    format: yaml        # Prefer YAML over JSON
    color: always       # Always use colors
    pretty_print: true
    paging: false       # Disable pager
```

#### Increase Timeouts and Retries

```yaml
# ~/.config/acme-cli/config.yaml
preferences:
  http:
    timeout: 120s       # Longer timeout for slow APIs

  retry:
    max_attempts: 5     # More retries for unreliable networks
```

#### Enterprise Proxy and TLS

```yaml
# ~/.config/acme-cli/config.yaml
preferences:
  http:
    proxy: http://proxy.corp.com:8080
    https_proxy: https://proxy.corp.com:8443
    no_proxy: [localhost, "127.0.0.1", ".internal"]

    tls:
      ca_bundle: /etc/ssl/certs/corporate-ca.pem
      # OR for insecure mode (shows warning):
      # insecure_skip_verify: true
```

#### Adjust Pagination

```yaml
# ~/.config/acme-cli/config.yaml
preferences:
  pagination:
    limit: 100          # Increase from default 20 (up to max_limit)
```

#### Disable Caching

```yaml
# ~/.config/acme-cli/config.yaml
preferences:
  caching:
    enabled: false      # Disable all caching
```

#### Deprecation Warning Controls

```yaml
# ~/.config/acme-cli/config.yaml
preferences:
  deprecations:
    always_show: true       # Show warnings every time (not just once)
    min_severity: warning   # Only show warning+ (skip info level)
```

### Debug Mode Overrides

**⚠️ ONLY for development/testing builds** where `metadata.debug: true`

#### Point to Local Test API

```yaml
# ~/.config/mycli-dev/config.yaml
debug_override:
  api:
    base_url: http://localhost:8080
    openapi_url: file://./test-openapi.yaml
```

#### Disable Authentication

```yaml
# ~/.config/mycli-dev/config.yaml
debug_override:
  behaviors:
    auth:
      type: none
```

#### Override Branding for Testing

```yaml
# ~/.config/mycli-dev/config.yaml
debug_override:
  branding:
    colors:
      primary: "#FF0000"    # Bright red to distinguish test build

  metadata:
    name: mycli-test
```

**Security warning displayed**:
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

### Configuration Priority

Values are resolved in this order (highest to lowest):

1. **Environment variable** (e.g., `MYCLI_TIMEOUT=60s`) - HIGHEST
2. **Command-line flag** (e.g., `--timeout 60s`)
3. **User config file**: `~/.config/my-cli/config.yaml` (`preferences` section)
4. **Embedded config**: Built into binary (`defaults` section from `cli-config.yaml`)
5. **Built-in default**: Hardcoded fallback - LOWEST

**Example**:
```bash
# Embedded config (defaults section):
defaults:
  output:
    format: json

# User config (preferences section):
preferences:
  output:
    format: yaml

# Command line:
$ MYCLI_OUTPUT_FORMAT=table mycli --output csv users list

# Result: Uses 'csv' (flag wins over env var, env var wins over user config, user config wins over embedded)
```

**Debug mode priority**:
```bash
# When metadata.debug: true, debug_override section has priority:
# 1. ENV > 2. Flag > 3. User preferences > 4. Debug override > 5. Embedded > 6. Built-in
```

### Validation

The CLI validates the user config file on startup:
- Warns about unknown fields in `preferences` section
- Rejects invalid values
- Falls back to embedded defaults on error
- **Special**: Detects `debug_override` section and either applies (debug build) or warns (production build)

### User Config Commands

CLIs should provide commands to manage user config:

```bash
# View current user configuration (preferences only)
my-cli config show

# View effective configuration (after merging embedded + user)
my-cli config show --effective

# View what debug_override would apply (debug builds only)
my-cli config show --debug-overrides

# Edit user config in $EDITOR
my-cli config edit

# Set a specific preference value
my-cli config set preferences.output.format yaml

# Get a specific preference value
my-cli config get preferences.output.format

# Reset preferences to embedded defaults
my-cli config reset

# Validate user config
my-cli config validate
```

### Example: Opting Into Auto-Updates

**User wants auto-updates**:

```bash
# Option 1: Use config command
$ my-cli config set preferences.updates.auto_install true
✓ Set preferences.updates.auto_install = true
⚠️  WARNING: Auto-install will update binaries without prompting.
   Only enable if you trust the update source.

# Option 2: Edit config file directly
$ my-cli config edit
# Opens ~/.config/my-cli/config.yaml in $EDITOR
# User adds:
# preferences:
#   updates:
#     auto_install: true

# Option 3: Create config manually
$ mkdir -p ~/.config/my-cli
$ cat > ~/.config/my-cli/config.yaml <<EOF
preferences:
  updates:
    auto_install: true
EOF
```

**Result**: Future updates will install automatically.

### Example: Developer Testing with Local API

**Developer working on local API instance**:

```yaml
# $XDG_CONFIG_HOME/mycli-dev/config.yaml
# (default: ~/.config/mycli-dev/config.yaml)

# Normal preferences (always active)
preferences:
  output:
    format: yaml      # Prefer YAML during development

  http:
    timeout: 300s     # Long timeout for debugging

# Debug overrides (only works if binary built with debug: true)
debug_override:
  api:
    base_url: http://localhost:8080
    openapi_url: file://./openapi.yaml

  behaviors:
    auth:
      type: none      # No auth for local testing
```

**Usage**:
```bash
# Production binary (debug: false)
$ mycli users list
⚠️ Warning: debug_override section found in config file but ignored
   This is a production build (debug: false)
[... uses https://api.example.com ...]

# Debug binary (debug: true)
$ mycli-dev users list
🚨 DEBUG MODE - Using overridden api.base_url: http://localhost:8080
[... uses http://localhost:8080 ...]
```

### Example: Custom XDG Locations

**User wants non-standard config location**:

```bash
# Set XDG environment variables
export XDG_CONFIG_HOME="$HOME/.custom-config"
export XDG_CACHE_HOME="$HOME/.custom-cache"
export XDG_DATA_HOME="$HOME/.custom-data"

# CLI will now use:
# - Config: ~/.custom-config/acme-cli/config.yaml
# - Cache:  ~/.custom-cache/acme-cli/
# - Data:   ~/.custom-data/acme-cli/
```

### Security Best Practices

**For CLI Developers**:
- ❌ Never set `auto_install` in embedded config
- ❌ Never set default API keys or tokens in embedded config
- ❌ Never enable `metadata.debug: true` in production binaries
- ✅ Provide clear warnings when users enable auto-updates
- ✅ Document user config file location and structure
- ✅ Validate user config strictly
- ✅ Use separate binary names for debug builds (`mycli` vs `mycli-dev`)
- ✅ Lock all security-critical settings in `api` and `behaviors` sections

**For CLI Users**:
- ⚠️ Only enable `auto_install` if you fully trust the CLI provider
- ⚠️ Never use debug builds in production
- ⚠️ Be cautious with `debug_override` - only for local testing
- ✅ Use environment variables for sensitive credentials
- ✅ Review user config file permissions (`chmod 600`)
- ✅ Understand what each setting does before enabling
- ✅ Keep `preferences` section minimal (only override what you need)

### Schema Subset

**What can be in `preferences` section** (production builds):

✅ **Overridable** (from `defaults` section):
- `preferences.http.timeout` - Request timeout
- `preferences.http.proxy` - HTTP proxy settings (not in embedded, user-only)
- `preferences.http.ca_bundle` - Custom CA certificates (not in embedded, user-only)
- `preferences.caching.enabled` - Enable/disable caching
- `preferences.pagination.limit` - Default page size (up to `behaviors.pagination.max_limit`)
- `preferences.output.*` - Output format, colors, paging, pretty-print
- `preferences.deprecations.*` - Warning display preferences
- `preferences.retry.max_attempts` - Number of retries
- `preferences.telemetry.enabled` - User opt-in for telemetry (not in embedded, user-only)
- `preferences.updates.auto_install` - Auto-update opt-in (not in embedded, user-only)

❌ **NOT Overridable** (locked to embedded config):
- `metadata.*` - CLI name, version, author, etc.
- `branding.*` - Colors, ASCII art, prompts
- `api.*` - **ALL API settings** (base_url, openapi_url, environments, headers, telemetry_url)
- `updates.*` - Update server URL, check interval, public key (except auto_install)
- `behaviors.auth.*` - Auth type and credentials
- `behaviors.retry.*` - Retry delays, backoff, status codes (except max_attempts)
- `behaviors.caching.*` - Cache TTLs, directory, max size (except enabled)
- `behaviors.pagination.*` - Max limit, inter-page delay (except default limit)
- `behaviors.secrets.*` - Secret detection and masking patterns
- `behaviors.builtin_commands.*` - Built-in command configuration
- `behaviors.global_flags.*` - Global flag configuration

**What can be in `debug_override` section** (debug builds only):

⚠️ **Everything** - when `metadata.debug: true`, ALL embedded config can be overridden via `debug_override` section
- Shows security warning on EVERY command
- Production builds ignore this section entirely
- Intended ONLY for development/testing

---

## Examples

### Example 1: Minimal Configuration

```yaml
metadata:
  name: simple-cli
  version: 1.0.0
  description: Simple API CLI

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com

updates:
  enabled: false

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: API_KEY
```

---

### Example 2: Full-Featured SaaS Product

```yaml
metadata:
  name: acme-cli
  version: 2.1.0
  description: ACME Corporation API CLI
  long_description: |
    The official CLI for ACME Corporation APIs.
    Manage users, resources, and analytics from the command line.

  author:
    name: ACME DevTools
    email: devtools@acme.com
    url: https://github.com/acme/cli

  license: Apache-2.0
  homepage: https://acme.com/cli
  docs_url: https://docs.acme.com/cli

branding:
  colors:
    primary: "#FF6B6B"
    secondary: "#4ECDC4"
    success: "#51CF66"
    warning: "#FFC078"
    error: "#FF6B6B"

  ascii_art: |
    ╔═══════════════════════════╗
    ║      ACME CLI v2.1        ║
    ╚═══════════════════════════╝

  theme:
    name: auto
    syntax_highlighting: true

api:
  openapi_url: https://api.acme.com/v2/openapi.yaml
  base_url: https://api.acme.com/v2

  environments:
    - name: production
      openapi_url: https://api.acme.com/v2/openapi.yaml
      base_url: https://api.acme.com/v2
      default: true

    - name: staging
      openapi_url: https://staging.acme.com/v2/openapi.yaml
      base_url: https://staging.acme.com/v2

  default_headers:
    X-Client-Version: "2.1.0"

  telemetry_url: https://telemetry.acme.com/events

# User-overridable defaults
defaults:
  http:
    timeout: 30s

  caching:
    enabled: true

  pagination:
    limit: 20

  output:
    format: json
    pretty_print: true
    color: auto
    paging: true

  deprecations:
    always_show: false
    min_severity: info

  retry:
    max_attempts: 3

updates:
  enabled: true
  update_url: https://releases.acme.com/cli
  check_interval: 24h

  public_key: |
    -----BEGIN PUBLIC KEY-----
    MCowBQYDK2VwAyEAXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    -----END PUBLIC KEY-----

# Locked behaviors (not user-overridable)
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: acme-cli
      auth_url: https://auth.acme.com/oauth/authorize
      token_url: https://auth.acme.com/oauth/token
      scopes:
        - api:read
        - api:write
        - admin
      redirect_url: http://localhost:8080/callback

  caching:
    spec_ttl: 5m
    response_ttl: 1m
    directory: ~/.cache/acme-cli
    max_size: 100MB

  retry:
    enabled: true
    initial_delay: 1s
    max_delay: 30s
    backoff_multiplier: 2.0
    retry_on_status: [429, 500, 502, 503, 504]

  pagination:
    max_limit: 100
    delay: 100ms

features:
  config_file: true
  interactive_mode: false
```

---

### Example 3: Internal Enterprise Tool

```yaml
metadata:
  name: internal-api-cli
  version: 1.0.0
  description: Internal API Management Tool
  author:
    name: Platform Team
    email: platform@company.com

branding:
  colors:
    primary: "#007ACC"
    secondary: "#00C7AC"

  prompts:
    command: "company>"

api:
  openapi_url: https://internal-api.company.internal/openapi.yaml
  base_url: https://internal-api.company.internal

  default_headers:
    X-Internal-Tool: "true"
    X-Team: "platform"

  telemetry_url: https://telemetry.company.internal

defaults:
  http:
    timeout: 60s  # Longer timeout for internal networks

  caching:
    enabled: true

  pagination:
    limit: 50

  output:
    format: table  # Enterprise users prefer tables
    paging: true

  deprecations:
    always_show: false
    min_severity: info

  retry:
    max_attempts: 3

updates:
  enabled: true
  update_url: https://artifacts.company.internal/cli
  check_interval: 12h
  # Note: auto_install can only be set in user config

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: COMPANY_API_KEY

  caching:
    spec_ttl: 30m  # Longer TTL for internal APIs
    response_ttl: 5m
    directory: ~/.cache/internal-api-cli
    max_size: 100MB

  retry:
    enabled: true
    initial_delay: 1s
    max_delay: 30s
    backoff_multiplier: 2.0
    retry_on_status: [429, 500, 502, 503, 504]

  pagination:
    max_limit: 200  # Higher limit for internal tools
    delay: 50ms

features:
  config_file: true
  interactive_mode: false
```

---

## Validation Rules

### Metadata Validation

```yaml
metadata:
  name:
    - required: true
    - pattern: ^[a-z0-9-]+$  # Lowercase, numbers, hyphens only
    - max_length: 50

  version:
    - required: true
    - pattern: ^\d+\.\d+\.\d+$  # Semantic versioning

  description:
    - required: true
    - min_length: 10
    - max_length: 200
```

### API Validation

```yaml
api:
  openapi_url:
    - required: true
    - format: uri OR file_path

  base_url:
    - required: true
    - format: uri
    - protocol: http OR https

  timeout:
    - format: duration
    - min: 1s
    - max: 5m
```

### Colors Validation

```yaml
branding:
  colors:
    primary:
      - pattern: ^#[0-9A-Fa-f]{6}$  # 6-digit hex
```

### Duration Format

Supported units:
- `s` - seconds
- `m` - minutes
- `h` - hours
- `d` - days

Examples:
- `30s` - 30 seconds
- `5m` - 5 minutes
- `24h` - 24 hours
- `7d` - 7 days

---

## Advanced Configurations

### Environment Variable Substitution

```yaml
api:
  base_url: ${API_BASE_URL}  # From environment

behaviors:
  auth:
    oauth2:
      client_secret: ${OAUTH_SECRET}
```

**Usage**:
```bash
export API_BASE_URL=https://api.example.com
export OAUTH_SECRET=secret123
acme-cli users list
```

---

### Conditional Configuration

```yaml
api:
  base_url: |
    {{if eq .Env "production"}}
      https://api.acme.com
    {{else if eq .Env "staging"}}
      https://staging-api.acme.com
    {{else}}
      http://localhost:8080
    {{end}}
```

---

### Includes

```yaml
# cli-config.yaml
metadata:
  name: my-cli
  version: 1.0.0
  description: My CLI

# Include common config
include:
  - common/auth.yaml
  - common/branding.yaml
```

```yaml
# common/auth.yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: shared-client
      auth_url: https://auth.example.com/oauth/authorize
      token_url: https://auth.example.com/oauth/token
```

---

### Profiles

```yaml
profiles:
  - name: developer
    behaviors:
      output:
        default_format: json
      caching:
        spec_ttl: 1m
    features:
      verbose: true

  - name: production
    behaviors:
      output:
        default_format: table
      caching:
        spec_ttl: 30m
    features:
      verbose: false
```

**Usage**:
```bash
acme-cli --profile developer users list
acme-cli --profile production users list
```

---

## Configuration Priority

Configuration values are resolved in this order (highest to lowest priority):

1. **Command-line flags**: `--base-url https://...`
2. **Environment variables**: `$API_BASE_URL`
3. **User config file**: `~/.config/my-cli/config.yaml`
4. **Embedded config**: Built into binary
5. **Defaults**: Hardcoded defaults

**Example**:
```bash
# Embedded config: base_url = https://api.acme.com
# User config: base_url = https://staging-api.acme.com
# Command line: --base-url https://localhost:8080

# Result: Uses https://localhost:8080 (highest priority)
acme-cli --base-url https://localhost:8080 users list
```

---

## Schema Validation

The generator validates configuration using JSON Schema:

```bash
# Validate configuration
cliforge validate cli-config.yaml

# Output:
✓ Configuration is valid
✓ All required fields present
✓ No unknown fields
✓ All values in valid format
```

**Common Errors**:
```
✗ Configuration validation failed:

  Line 12: metadata.version: Does not match pattern ^\d+\.\d+\.\d+$
  Line 34: api.base_url: Must be a valid URL
  Line 56: branding.colors.primary: Must be a hex color (#RRGGBB)
```

---

**End of Configuration DSL Specification**

*Version 0.8.0*
*Last Updated: 2025-01-11*

**Changelog**:
- v0.8.0 (2025-01-11): **Configuration Override Architecture Finalized**
  - **BREAKING**: Completely redesigned configuration override system
  - **New `defaults` section** in embedded config (13 overridable settings)
  - **New `preferences` section** in user config (overrides defaults)
  - **New `debug_override` section** in user config (only works when `metadata.debug: true`)
  - **Added `metadata.debug`** boolean field (enables/disables debug override mode)
  - **Locked entire `api` section** (100% embedded-only, including telemetry_url)
  - **Moved `timeout`** from `api.timeout` to `defaults.http.timeout` (user-overridable)
  - **Added `api.telemetry_url`** (locked, defines WHERE to send telemetry)
  - **Moved telemetry opt-in** to `preferences.telemetry.enabled` (user-only, defaults to false)
  - **Removed `rate_limiting` section** entirely (redundant with retry + server 429)
  - **Added `behaviors.pagination.delay`** (locked, inter-page delay for auto-pagination)
  - **Removed auto_paginate from config** (flag/ENV only: `--auto-page`, `MYCLI_AUTO_PAGE`)
  - **Locked all behaviors** except specific defaults (auth, retry delays, caching TTLs, secrets, etc.)
  - **Removed sections entirely**: `commands`, `hooks`, `plugins` (use shell alternatives)
  - **Removed features**: `telemetry`, `shell_completion`, `validate_*`, `offline_mode`
  - **Updated all examples** to use new `defaults` section
  - **Debug mode security warnings** on EVERY command when `debug: true`
  - **Final count**: 71 locked settings, 13 overridable settings
  - **Clear separation**: `defaults` (overridable) vs `behaviors` (locked)
  - **User config structure**: `preferences` (always active) vs `debug_override` (debug-only)
  - **Configuration priority**: ENV > Flag > User Config (preferences) > Embedded (defaults) > Built-in Default
- v0.7.0 (2025-01-11): **Built-in Commands, Global Flags & Secrets**
  - Added `behaviors.builtin_commands` section for configuring standard CLI commands
  - Added comprehensive `behaviors.global_flags` section for global flag configuration
  - Added `behaviors.secrets` section for sensitive data masking
  - Support for three CLI styles: subcommand, flag, and hybrid
  - Standard built-in commands: version, help, info, config, completion, update, changelog, deprecations, cache, auth
  - Standard global flags: --config, --profile, --region, --output, --verbose, --yes, etc.
  - Added `--yes` / `-y` flag for non-interactive mode (CI/CD)
  - Added `--region` / `-r` flag for cloud/datacenter APIs
  - Removed `--api-key` and `--base-url` from defaults (use custom flags and profiles)
  - Custom global flags support with conflict detection and environment variable mapping
  - Secret detection via `x-cli-secret`, field patterns, and value patterns
  - Masking strategies: partial, full, hash
  - Removed duplicate simple `commands.global_flags` (now in behaviors.global_flags)
  - Added company customization examples (Docker-like, Unix-like, minimal, etc.)
- v0.4.0 (2025-01-11): **Rebranded to CliForge**
  - Changed project name from "Alpha-Omega" to "CliForge"
  - Updated generator CLI name from `alpha-omega-gen` to `cliforge`
- v0.3.0 (2025-11-10): Security and usability improvements
  - **BREAKING SECURITY CHANGE**: Removed `auto_install` from embedded config schema
  - `auto_install` can ONLY be set in user configuration file (security boundary)
  - Removed redundant `channel` field (use `update_url` instead)
  - Added comprehensive "User Configuration File" section
  - Added `config` command examples for managing user preferences
  - Clarified configuration priority (CLI flags > env vars > user config > embedded > defaults)
  - **XDG Base Directory Specification compliance** for all file locations
  - Added XDG environment variables reference table
  - Added examples for custom XDG locations
- v0.1.0 (2025-11-09): Initial specification
