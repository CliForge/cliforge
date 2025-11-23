# Built-in Commands & Global Flags Design

## Overview

CliForge-generated CLIs need to support standard CLI conventions while allowing customization to match existing company CLI patterns. This document defines how built-in commands and global flags are configured.

---

## Problem Statement

Different CLI ecosystems have different conventions:

| Tool | Version | Help | Info | Config |
|------|---------|------|------|--------|
| **Docker** | `docker version` + `docker --version` | `docker --help` | `docker info` | `docker config` |
| **kubectl** | `kubectl version` | `kubectl --help` | `kubectl cluster-info` | `kubectl config` |
| **curl** | `curl --version` / `curl -V` | `curl --help` / `curl -h` | N/A | N/A |
| **node** | `node --version` / `node -v` | `node --help` / `node -h` | N/A | N/A |
| **git** | `git version` + `git --version` | `git help` + `git --help` | N/A | `git config` |

**Key observations:**
1. **Hybrid is common**: Many tools support BOTH subcommand AND flag style (e.g., `docker version` AND `docker --version`)
2. **Short + long**: Traditional tools offer both `-v` and `--version`
3. **Ecosystem consistency**: Companies want their generated CLI to match their existing tools

**CliForge must support all these patterns** through configuration.

---

## CLI Styles

### 1. Subcommand Style (Modern CLIs)

```bash
mycli version              # Show version
mycli help                 # Show help
mycli config show          # Show config
mycli config set key value # Set config value
mycli completion bash      # Generate bash completion
mycli update               # Update CLI
```

**Characteristics:**
- Everything is a subcommand
- Organized hierarchically
- Used by: Docker, kubectl, gh, terraform

### 2. Flag Style (Traditional Unix)

```bash
mycli --version            # Show version
mycli -V                   # Show version (short)
mycli --help               # Show help
mycli -h                   # Show help (short)
```

**Characteristics:**
- Everything is a flag
- Flat structure
- Used by: curl, grep, ls, node, python

### 3. Hybrid Style (Best of both)

```bash
mycli version              # Subcommand
mycli --version            # Flag (also works)
mycli -V                   # Short flag (also works)
mycli help command         # Subcommand with argument
mycli command --help       # Per-command help flag
```

**Characteristics:**
- Supports both paradigms
- Maximum flexibility
- Used by: git, npm, docker

---

## Built-in Commands

### Standard Built-in Commands

Every CliForge-generated CLI should support these built-in commands:

| Command | Purpose | Default Style | Configurable |
|---------|---------|---------------|--------------|
| **version** | Show version info | Subcommand + `--version` flag | ✅ Yes |
| **help** | Show help | Subcommand + `--help` flag | ✅ Yes |
| **info** | Show CLI + API info | Subcommand | ✅ Yes |
| **config** | Manage user config | Subcommand group | ✅ Yes |
| **completion** | Shell completion | Subcommand | ✅ Yes |
| **update** | Self-update CLI | Subcommand | ✅ Yes |
| **changelog** | Show changelog | Subcommand | ✅ Yes |
| **deprecations** | Show deprecations | Subcommand | ✅ Yes |
| **cache** | Manage cache | Subcommand group | ✅ Yes |
| **auth** | Manage authentication | Subcommand group | ✅ Yes |

### version

**Displays both CLI binary version AND API version** (similar to `oc version` showing client and server):

**Subcommand style:**
```bash
$ mycli version
Client Version: v1.2.3
Server Version: v2.1.0 (My API)
OpenAPI Spec: v2.1.0
Built: 2025-01-11T12:00:00Z
Go: go1.21.5
```

**Flag style (shorter output):**
```bash
$ mycli --version
Client: v1.2.3  Server: v2.1.0
```

**Key insight**: The CLI binary version changes RARELY (security patches, CLI features), while the API version changes FREQUENTLY (new endpoints, schema changes). Users need to see both.

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    version:
      enabled: true
      style: hybrid  # subcommand, flag, or hybrid
      flags:
        - "--version"
        - "-V"
      show_api_version: true
      show_build_info: true
```

### help

**Subcommand style:**
```bash
$ mycli help
$ mycli help users
$ mycli help users create
```

**Flag style:**
```bash
$ mycli --help
$ mycli users --help
$ mycli users create --help
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    help:
      enabled: true
      style: hybrid
      flags:
        - "--help"
        - "-h"
      show_examples: true
      show_aliases: true
```

### info

**Shows detailed CLI and API information:**
```bash
$ mycli info

mycli v1.2.3
Command-line interface for My API

API Information:
  Title: My API
  Version: 2.0.0
  Base URL: https://api.example.com/v2
  Spec URL: https://api.example.com/openapi.yaml
  Endpoints: 42

Configuration:
  Config file: ~/.config/mycli/config.yaml
  Cache dir: ~/.cache/mycli
  Auth: API Key (configured)

Status:
  ✓ API reachable
  ✓ Authentication valid
  ✓ Spec cached (age: 2m)
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    info:
      enabled: true
      style: subcommand  # only subcommand, no flag equivalent
      check_api_health: true  # Ping API to verify connectivity
      show_config_paths: true
```

### config

**Manage user configuration:**
```bash
$ mycli config show
$ mycli config get output.format
$ mycli config set output.format json
$ mycli config unset output.format
$ mycli config edit  # Opens $EDITOR
$ mycli config path  # Show config file path
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    config:
      enabled: true
      style: subcommand
      allow_edit: true  # Enable 'config edit'
      subcommands:
        - show
        - get
        - set
        - unset
        - edit
        - path
```

### completion

**Generate shell completion scripts:**
```bash
$ mycli completion bash > /etc/bash_completion.d/mycli
$ mycli completion zsh > ~/.zsh/completion/_mycli
$ mycli completion fish > ~/.config/fish/completions/mycli.fish
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    completion:
      enabled: true
      shells:
        - bash
        - zsh
        - fish
        - powershell
```

### update

**Self-update the CLI:**
```bash
$ mycli update
Checking for updates...
Current version: v1.2.3
Latest version: v1.3.0

Changelog:
- Added new 'users export' command
- Fixed authentication bug
- Improved error messages

Update now? [Y/n]: y
Downloading v1.3.0... done
Installing... done
✓ Updated to v1.3.0
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    update:
      enabled: true
      auto_check: false  # Check on startup (respects user config)
      require_confirmation: true
      show_changelog: true
```

### changelog

**Shows BOTH binary changes (rare) and API changes (frequent):**
```bash
$ mycli changelog

CLI Binary Changelog:
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
v1.2.3 (2025-01-15)
  • Fixed security vulnerability in credential storage
  • Improved error messages for network timeouts
  • Added --region flag support

v1.2.2 (2024-12-01)
  • Initial release

API Changelog (My API):
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
v2.1.0 (2025-01-14) - Current
  • Added new /analytics endpoint
  • DEPRECATED: GET /v1/users (use /v2/users)
  • Breaking: POST /users now requires email field

v2.0.0 (2025-01-01)
  • Major API redesign
  • Pagination now required
```

**Filter by type:**
```bash
$ mycli changelog --binary-only      # Only CLI changes
$ mycli changelog --api-only         # Only API changes
$ mycli changelog --since v2.0.0     # Since specific version
$ mycli changelog --format json      # JSON output
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    changelog:
      enabled: true
      show_binary_changes: true   # Show CLI binary changelog
      show_api_changes: true      # Show API changelog
      binary_changelog:
        source: embedded          # embedded, url, or api
        url: "https://releases.example.com/cli-changelog.yaml"
      api_changelog:
        source: api               # From x-cli-changelog in OpenAPI spec
        url: "https://api.example.com/openapi.yaml"
      default_limit: 10           # Show last 10 versions
```

**Key insight**: Binary changes happen monthly/quarterly (security, CLI features), while API changes happen weekly/daily (new endpoints, parameters). They have different audiences and update frequencies.

### deprecations

**Shows BOTH binary deprecations (rare) and API deprecations (frequent):**
```bash
$ mycli deprecations

Active Deprecations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

CLI Binary Deprecations (0):
  ✓ No deprecated CLI features

API Deprecations (2):

  🔴 CRITICAL - 28 days remaining
  ├─ Operation: GET /v1/users (users list)
  ├─ Sunset: December 31, 2025
  ├─ Replacement: users list-v2
  └─ Docs: https://docs.example.com/migration/v1-to-v2

  ⚠️  WARNING - 89 days remaining
  ├─ Parameter: --filter (on users list)
  ├─ Sunset: February 28, 2026
  ├─ Replacement: --search
  └─ Migration: Replace --filter "x" with --search "x"

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Show details: mycli deprecations show <operation-id>
Scan usage: mycli deprecations scan
```

**Filter by type:**
```bash
$ mycli deprecations --binary-only   # Only CLI deprecations
$ mycli deprecations --api-only      # Only API deprecations
$ mycli deprecations show listUsersV1
$ mycli deprecations scan ./scripts/
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    deprecations:
      enabled: true
      show_binary_deprecations: true  # Show CLI deprecations
      show_api_deprecations: true     # Show API deprecations
      show_by_default: true
      allow_scan: true
      allow_auto_fix: true
```

**Key insight**: Binary deprecations affect CLI usage patterns (flags, commands), while API deprecations affect endpoints and parameters. Both need to be tracked separately.

### cache

**Manage cache:**
```bash
$ mycli cache info
$ mycli cache clear
$ mycli cache clear --spec-only
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    cache:
      enabled: true
      subcommands:
        - info
        - clear
```

### auth

**Manage authentication:**
```bash
$ mycli auth login
$ mycli auth logout
$ mycli auth status
$ mycli auth refresh
```

**Configuration:**
```yaml
behaviors:
  builtin_commands:
    auth:
      enabled: true
      subcommands:
        - login
        - logout
        - status
        - refresh
```

---

## Global Flags

### Standard Global Flags

These flags should be available for ALL commands (including API-generated commands):

| Flag | Short | Type | Purpose | Default |
|------|-------|------|---------|---------|
| `--config` | `-c` | string | Config file path | OS-specific |
| `--profile` | | string | Profile section in config | `default` |
| `--region` | `-r` | string | Region/datacenter | from config |
| `--output` | `-o` | string | Output format | `json` |
| `--verbose` | `-v` | bool | Verbose output | `false` |
| `--quiet` | `-q` | bool | Quiet mode | `false` |
| `--debug` | | bool | Debug mode | `false` |
| `--no-color` | | bool | Disable colors | `false` |
| `--timeout` | | duration | Request timeout | `30s` |
| `--retry` | | int | Retry attempts | `3` |
| `--no-cache` | | bool | Disable cache | `false` |
| `--no-update-check` | | bool | Skip update check | `false` |
| `--yes` | `-y` | bool | Skip confirmations (non-interactive) | `false` |
| `--help` | `-h` | bool | Show help | N/A |
| `--version` | `-V` | bool | Show version | N/A |

**Removed from defaults:**
- ~~`--api-key`~~ - Auth varies too much (OAuth2, API key, Basic, etc.) - configure per-CLI via custom flags
- ~~`--base-url`~~ - **SECURITY BOUNDARY** - base URL is NEVER user-configurable
  - **Embedded in binary only** - defined in branding DSL config at generation time
  - Cannot be overridden by users (no flag, no config file, no env var)
  - **Rationale**: Prevents users from accidentally/maliciously talking to wrong API
  - **For multi-environment**: Generate separate binaries (`mycli-prod`, `mycli-staging`) or use separate config profiles with different embedded configs
  - **Exception**: Proxy settings and CA certificates ARE configurable (see below)

### Profile Support

The `--profile` flag selects a named section in the config file, similar to AWS CLI:

**Config file structure** (`~/.config/mycli/config.yaml`):
```yaml
# Default profile (used when --profile not specified)
default:
  preferences:
    output:
      format: json
    region: us-east-1
    # Note: Credentials go in environment variables, not config file

# Production profile
production:
  preferences:
    output:
      format: table
      color: always
    region: us-west-2

# Development profile
dev:
  preferences:
    output:
      format: yaml
    region: us-east-1

    # Dev might use different proxy/CA settings
    http:
      proxy: http://localhost:8888  # Local debugging proxy
      tls:
        ca_bundle: /path/to/dev-ca.pem

  # Debug overrides (only work if binary has metadata.debug: true)
  debug_override:
    api:
      base_url: http://localhost:8080  # Point to local API
```

**Usage:**
```bash
# Uses 'default' profile
$ mycli users list

# Uses 'production' profile
$ mycli --profile production users list

# Uses 'dev' profile with override
$ mycli --profile dev --region us-west-2 users list
```

**Environment variable:**
```bash
export MYCLI_PROFILE=production
mycli users list  # Uses production profile
```

**Precedence** (IMPORTANT):
1. **Environment variable**: `MYCLI_PROFILE=production` (HIGHEST)
2. **Command-line flag**: `--profile production`
3. **Config file**: `default` section
4. **Built-in default**: hardcoded fallback (LOWEST)

### Configuration

```yaml
behaviors:
  global_flags:
    # Enable/disable individual flags
    config:
      enabled: true
      flag: "--config"
      short: "-c"
      env_var: "MYCLI_CONFIG"
      description: "Path to config file"

    profile:
      enabled: true
      flag: "--profile"
      short: ""  # No short form
      env_var: "MYCLI_PROFILE"
      description: "Configuration profile to use (matches section in config file)"
      default: "default"

    region:
      enabled: true
      flag: "--region"
      short: "-r"
      env_var: "MYCLI_REGION"
      description: "Region/datacenter to use"
      # Note: Companies can customize allowed values and default per their API
      allowed_values: []  # Empty = no validation, or specify: ["us-east-1", "us-west-2", "eu-west-1"]

    output:
      enabled: true
      flag: "--output"
      short: "-o"
      env_var: "MYCLI_OUTPUT_FORMAT"
      description: "Output format (json|yaml|table|csv)"
      default: "json"
      allowed_values: ["json", "yaml", "table", "csv"]

    verbose:
      enabled: true
      flag: "--verbose"
      short: "-v"
      env_var: "MYCLI_VERBOSE"
      description: "Enable verbose output"
      default: false
      # Can be repeated: -v, -vv, -vvv
      repeatable: true

    quiet:
      enabled: true
      flag: "--quiet"
      short: "-q"
      env_var: "MYCLI_QUIET"
      description: "Suppress non-error output"
      default: false
      conflicts_with: ["verbose"]

    debug:
      enabled: true
      flag: "--debug"
      short: ""
      env_var: "MYCLI_DEBUG"
      description: "Enable debug logging"
      default: false

    no_color:
      enabled: true
      flag: "--no-color"
      short: ""
      env_var: "NO_COLOR"  # Standard env var
      description: "Disable colored output"
      default: false

    # HTTP client flags
    timeout:
      enabled: true
      flag: "--timeout"
      short: ""
      env_var: "MYCLI_TIMEOUT"
      description: "Request timeout (e.g., 30s, 1m)"
      default: "30s"

    retry:
      enabled: true
      flag: "--retry"
      short: ""
      env_var: "MYCLI_RETRY"
      description: "Number of retry attempts"
      default: 3

    no_cache:
      enabled: true
      flag: "--no-cache"
      short: ""
      env_var: "MYCLI_NO_CACHE"
      description: "Disable response caching"
      default: false

    yes:
      enabled: true
      flag: "--yes"
      short: "-y"
      env_var: "MYCLI_YES"
      description: "Skip confirmation prompts (non-interactive mode)"
      default: false
      # Critical for CI/CD automation

    # Custom global flags (defined by CLI developer)
    # Use this for API-specific flags like --api-key, --token, --org, etc.
    custom:
      - name: "api_key"
        flag: "--api-key"
        short: ""
        env_var: "MYCLI_API_KEY"
        description: "API key for authentication"
        sensitive: true  # Don't log, mask in output
        # Only add if your API uses API key auth

      - name: "token"
        flag: "--token"
        short: "-t"
        env_var: "MYCLI_TOKEN"
        description: "Bearer token"
        sensitive: true
        # Only add if your API uses bearer token auth

      - name: "org"
        flag: "--org"
        short: ""
        env_var: "MYCLI_ORG"
        description: "Organization ID"
        required: false
        # Add if your API is multi-tenant
```

---

## Configuration Examples

### Example 1: Modern Docker-like CLI

```yaml
metadata:
  name: acmectl
  version: 1.0.0
  description: ACME Corporation API CLI

behaviors:
  # Hybrid style: support both subcommands and flags
  builtin_commands:
    version:
      enabled: true
      style: hybrid  # Both 'acmectl version' and 'acmectl --version'
      flags: ["--version", "-V"]
      show_api_version: true

    help:
      enabled: true
      style: hybrid
      flags: ["--help", "-h"]

    info:
      enabled: true
      style: subcommand

    config:
      enabled: true
      style: subcommand

  global_flags:
    verbose:
      enabled: true
      repeatable: true  # -v, -vv, -vvv for increasing verbosity

    output:
      enabled: true
      default: "table"  # More user-friendly for interactive use
```

### Example 2: Traditional Unix-style CLI

```yaml
metadata:
  name: acme-api
  version: 1.0.0
  description: ACME API client

behaviors:
  # Flag-only style
  builtin_commands:
    version:
      enabled: true
      style: flag  # Only 'acme-api --version', no subcommand
      flags: ["--version", "-V"]

    help:
      enabled: true
      style: flag
      flags: ["--help", "-h"]

    info:
      enabled: false  # No info subcommand

    config:
      enabled: false  # No config subcommand

  global_flags:
    output:
      enabled: true
      flag: "--output"
      short: "-o"
```

### Example 3: Minimal CLI (Flags disabled)

```yaml
metadata:
  name: simple-api
  version: 1.0.0
  description: Simple API client

behaviors:
  # Subcommand-only style
  builtin_commands:
    version:
      enabled: true
      style: subcommand  # Only 'simple-api version'
      flags: []  # No --version flag

    help:
      enabled: true
      style: subcommand
      flags: []

  global_flags:
    # Minimal flags
    config:
      enabled: false
    verbose:
      enabled: false
    debug:
      enabled: false
```

### Example 4: Company-standard CLI (matching existing tools)

```yaml
metadata:
  name: company-api
  version: 1.0.0
  description: Company API CLI

behaviors:
  # Match company's existing CLI conventions
  builtin_commands:
    version:
      enabled: true
      style: flag
      flags: ["--version"]  # Company doesn't use -V short form
      show_build_info: false  # Keep it simple

    help:
      enabled: true
      style: hybrid
      flags: ["--help"]  # Company doesn't use -h

    config:
      enabled: true
      style: subcommand
      # Company uses 'configure' instead of 'config'
      command_name: "configure"

  global_flags:
    # Company-specific flag naming
    output:
      enabled: true
      flag: "--format"  # Company uses --format instead of --output
      short: "-f"
      default: "json"

    verbose:
      enabled: true
      flag: "--verbose"
      short: ""  # Company doesn't use short flags for verbosity
      repeatable: false

    # Company-specific custom flags
    custom:
      - name: "environment"
        flag: "--env"
        short: "-e"
        env_var: "COMPANY_ENV"
        description: "Environment (dev|staging|prod)"
        default: "dev"
        allowed_values: ["dev", "staging", "prod"]
```

---

## Implementation Notes

### Flag Precedence

When the same option can be specified multiple ways, the precedence is:

1. **Command-line flag** (highest priority)
2. **Environment variable**
3. **User config file** (`~/.config/mycli/config.yaml`)
4. **Embedded config** (built into binary)
5. **Default value** (lowest priority)

Example:
```bash
# If user has in ~/.config/mycli/config.yaml:
preferences:
  output:
    format: yaml

# And runs:
MYCLI_OUTPUT_FORMAT=json mycli users list --output table

# Result: Uses 'table' (command-line flag wins)
```

### Conflict Detection

Some flags conflict with each other:
```yaml
global_flags:
  verbose:
    conflicts_with: ["quiet"]

  quiet:
    conflicts_with: ["verbose", "debug"]
```

If user specifies both:
```bash
$ mycli --verbose --quiet users list
Error: flags --verbose and --quiet are mutually exclusive
```

### Sensitive Flags

Sensitive flags (API keys, tokens) should:
- Not appear in logs
- Not show default values in `--help`
- Be masked in output

```yaml
global_flags:
  api_key:
    sensitive: true
```

---

## Enterprise Features

### Proxy Support

**Critical for corporate environments** - many enterprises require HTTP(S) proxies.

**Standard environment variables** (auto-detected):
```bash
export HTTP_PROXY=http://proxy.corp.com:8080
export HTTPS_PROXY=https://proxy.corp.com:8443
export NO_PROXY=localhost,127.0.0.1,.internal

mycli users list  # Uses proxy automatically
```

**Configuration file:**
```yaml
# ~/.config/mycli/config.yaml
default:
  preferences:
    http:
      proxy: http://proxy.corp.com:8080
      https_proxy: https://proxy.corp.com:8443
      no_proxy: [localhost, "127.0.0.1", ".internal", ".corp.com"]
```

**Priority:**
1. **Environment variables** (HTTP_PROXY, HTTPS_PROXY, NO_PROXY) - HIGHEST
2. **Config file** (`http.proxy`, `http.https_proxy`, `http.no_proxy`)
3. **No proxy** - direct connection (LOWEST)

**Implementation notes:**
- Respect standard `HTTP_PROXY`, `HTTPS_PROXY`, `NO_PROXY` env vars
- Support both uppercase and lowercase variants
- Support proxy authentication: `http://user:pass@proxy.corp.com:8080`
- Respect `NO_PROXY` for internal domains

### CA Certificate Override

**Critical for corporate MITM proxies** - enterprises often use SSL inspection.

**Environment variable:**
```bash
export MYCLI_CA_BUNDLE=/etc/ssl/certs/corporate-ca.pem
mycli users list  # Uses custom CA bundle
```

**Config file:**
```yaml
# ~/.config/mycli/config.yaml
default:
  preferences:
    http:
      tls:
        ca_bundle: /etc/ssl/certs/corporate-ca.pem
        # OR multiple CA files
        ca_bundle:
          - /etc/ssl/certs/corporate-ca.pem
          - /usr/local/share/ca-certificates/internal-ca.crt
```

**System CA bundle locations** (auto-detected fallbacks):
- Linux: `/etc/ssl/certs/ca-certificates.crt`, `/etc/pki/tls/certs/ca-bundle.crt`
- macOS: System keychain + `/etc/ssl/cert.pem`
- Windows: System certificate store

**Insecure mode** (disable TLS verification - dangerous!):
```bash
export MYCLI_TLS_INSECURE=true  # NOT RECOMMENDED!
mycli users list
```

```yaml
# Config file (NOT RECOMMENDED!)
default:
  preferences:
    http:
      tls:
        insecure_skip_verify: true  # Shows warning on every command
```

**Priority:**
1. **Environment variable**: `MYCLI_CA_BUNDLE` (HIGHEST)
2. **Config file**: `http.tls.ca_bundle`
3. **System CA bundle**: OS-specific default (LOWEST)

**Security warnings:**
- Setting `insecure_skip_verify: true` shows a warning on EVERY command execution
- Warning includes: "⚠️  TLS verification disabled - connection is NOT secure"
- Audit logs record when insecure mode is used

### Configuration Priority (Complete)

**For ALL settings**, the precedence is:

```
1. Environment variable (HIGHEST)
   ↓
2. Command-line flag
   ↓
3. User config file (~/.config/mycli/config.yaml - preferences section)
   ↓
4. Embedded config (in binary, defaults section from cli-config.yaml)
   ↓
5. Built-in default (LOWEST)
```

**Example with `--output` flag:**
```bash
# Embedded config (in binary):
defaults:
  output:
    format: json

# User config file:
default:
  preferences:
    output:
      format: yaml

# Command execution:
MYCLI_OUTPUT_FORMAT=table mycli --output csv users list

# Result: Uses 'csv' (flag wins over env var)
# Priority: csv (flag) > table (env) > yaml (user pref) > json (embedded default)
```

**Special case - `api.base_url`:**
```
1. Embedded config (in binary) - ONLY SOURCE
   ↓
2. CANNOT be overridden by env var, flag, or user config
```

---

## Cobra Integration

CliForge uses Cobra for CLI framework. Built-in commands map to:

```go
// Root command
rootCmd := &cobra.Command{
    Use:   "mycli",
    Short: "ACME API CLI",
}

// Built-in subcommands
versionCmd := &cobra.Command{Use: "version", Run: showVersion}
helpCmd := &cobra.Command{Use: "help", Run: showHelp}
infoCmd := &cobra.Command{Use: "info", Run: showInfo}
configCmd := &cobra.Command{Use: "config", Short: "Manage configuration"}
// ... etc

// Global persistent flags
rootCmd.PersistentFlags().StringP("output", "o", "json", "Output format")
rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Verbose output")
rootCmd.PersistentFlags().String("api-key", "", "API key")
// ... etc

// Hybrid support: version as both subcommand and flag
rootCmd.PersistentFlags().BoolP("version", "V", false, "Show version")
rootCmd.ParseFlags(os.Args[1:])
if versionFlag {
    showVersion()
    os.Exit(0)
}
```

---

## Summary

**Key Design Principles:**

1. **Flexibility**: Support subcommand style, flag style, or hybrid
2. **Convention over configuration**: Sensible defaults that match modern CLI patterns
3. **Customization**: Companies can match their existing CLI conventions
4. **Consistency**: Standard built-in commands across all CliForge CLIs
5. **Discoverability**: `--help` and completion make features discoverable

**Default behavior (if not configured):**
- Hybrid style for version/help (both subcommand and flags work)
- Standard global flags (--output, --verbose, --config, etc.)
- All built-in commands enabled

**Customization power:**
- Disable any built-in command
- Change command names (`config` → `configure`)
- Change flag names (`--output` → `--format`)
- Choose subcommand-only, flag-only, or hybrid
- Add custom global flags

This gives CliForge maximum flexibility while maintaining sensible defaults.

---

**Version**: 0.7.0
**Last Updated**: 2025-01-11
**Project**: CliForge - Forge CLIs from APIs
