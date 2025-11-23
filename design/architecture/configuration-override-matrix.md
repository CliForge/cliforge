# Configuration Override Matrix

**Version**: 0.8.0 - ✅ FINALIZED

**Purpose**: Documents which embedded configuration settings can be overridden by environment variables, CLI flags, or user config files, and which settings are locked to the embedded configuration for security.

---

## Override Mechanisms

| Mechanism | Priority | Symbol | Example |
|-----------|----------|--------|---------|
| Environment Variable | 1 (HIGHEST) | 🌍 | `MYCLI_OUTPUT_FORMAT=json` |
| Command-line Flag | 2 | 🚩 | `--output json` |
| User Config File | 3 | 📄 | `~/.config/mycli/config.yaml` |
| Embedded Config | 4 | 📦 | Built into binary |
| Built-in Default | 5 (LOWEST) | 🔧 | Hardcoded fallback |

**Precedence**: `ENV > Flag > Config File > Embedded > Default`

---

## 🔓 Debug Mode Override

**CRITICAL SECURITY FEATURE**: The embedded config can include a `debug` boolean that controls override behavior:

```yaml
# In embedded config (cli-config.yaml)
metadata:
  name: mycli
  version: 1.0.0
  debug: false  # Production build - overrides LOCKED

# OR for development/testing builds:
metadata:
  name: mycli-dev
  version: 1.0.0-dev
  debug: true   # Debug build - overrides ALLOWED
```

### Debug Mode Behavior

**When `metadata.debug: false` (PRODUCTION - default)**:
- ❌ Override rules in this document apply strictly
- ❌ Locked sections cannot be overridden (api, metadata, branding, etc.)
- ✅ Only explicitly overridable settings can be changed
- ✅ No warnings displayed

**When `metadata.debug: true` (DEVELOPMENT/TESTING)**:
- ⚠️ **ALL** embedded configuration can be overridden via special `debug_override` section in config file
- ⚠️ **SECURITY WARNING** displayed on EVERY command execution
- ⚠️ Intended ONLY for development, testing, debugging
- ⚠️ Should NEVER be used in production binaries

### User Config File Structure with Debug Overrides

```yaml
# ~/.config/mycli/config.yaml

# ===================================================================
# Normal user preferences (work in ALL builds)
# ===================================================================
behaviors:
  output:
    default_format: yaml  # User prefers YAML
  notifications:
    show_changelog: false  # User doesn't want changelog notifications

# ===================================================================
# Debug-only overrides (ONLY work when metadata.debug: true)
# ===================================================================
debug_override:
  # Override embedded API configuration (normally locked)
  api:
    base_url: http://localhost:8080  # Point to local test API
    openapi_url: file://./test-openapi.yaml  # Use local spec

  # Override embedded metadata (normally locked)
  metadata:
    name: mycli-custom

  # Override embedded branding (normally locked)
  branding:
    colors:
      primary: "#FF0000"  # Red for local testing visibility

  # Override any other embedded settings
  behaviors:
    auth:
      type: none  # Disable auth for local testing
```

**Behavior**:
- **Production build** (`debug: false`): `debug_override` section **IGNORED** entirely
  - CLI shows warning: "⚠️ debug_override section in config ignored (not a debug build)"
  - This prevents accidental misuse

- **Debug build** (`debug: true`): `debug_override` section **APPLIED**
  - Overrides embedded config
  - Security warning displayed on every command
  - Shows which overrides are active:
    ```
    🚨 Debug overrides active:
      - api.base_url: http://localhost:8080 (was: https://api.example.com)
      - api.openapi_url: file://./test-openapi.yaml
      - behaviors.auth.type: none (was: oauth2)
    ```

### Warning Display (Debug Mode)

Every command execution shows:

```bash
$ mycli-dev users list

╔════════════════════════════════════════════════════════════════╗
║  🚨 DEBUG MODE ENABLED - SECURITY WARNING                     ║
╠════════════════════════════════════════════════════════════════╣
║                                                                ║
║  This is a DEBUG BUILD.                                        ║
║  All embedded configuration can be overridden.                 ║
║                                                                ║
║  ⚠️  DO NOT USE IN PRODUCTION                                 ║
║                                                                ║
║  Build info:                                                   ║
║  - Version: 1.0.0-dev                                          ║
║  - Debug: ENABLED                                              ║
║  - Config overrides: ALLOWED                                   ║
║                                                                ║
║  Active debug_override settings (3):                           ║
║  - api.base_url: http://localhost:8080                         ║
║  - api.openapi_url: file://./test-openapi.yaml                 ║
║  - behaviors.auth.type: none                                   ║
║                                                                ║
╚════════════════════════════════════════════════════════════════╝

[... normal command output ...]
```

**If production build detects `debug_override` section**:

```bash
$ mycli users list

⚠️  Warning: debug_override section found in config file but ignored
    This is a production build (debug: false)
    debug_override section is only active in debug builds
    Location: ~/.config/mycli/config.yaml

[... normal command output ...]
```

### Use Cases for Debug Mode

**✅ Appropriate uses**:
- Local development and testing
- QA/staging environments with test APIs
- Debugging production issues in controlled environments
- Security research and penetration testing (authorized)
- CI/CD pipeline testing with mock APIs

**❌ Inappropriate uses**:
- Production deployments
- Customer-facing binaries
- Public releases
- Untrusted environments

### Implementation Notes

```go
// In CLI runtime
type UserConfig struct {
    // Normal user preferences (always active)
    Behaviors BehaviorsConfig
    Features  FeaturesConfig

    // Debug-only overrides (only active when embedded debug: true)
    DebugOverride *EmbeddedConfig `yaml:"debug_override,omitempty"`
}

func LoadConfig() (*Config, error) {
    // Load embedded config from binary
    embedded := loadEmbeddedConfig()

    // Load user config from file
    user := loadUserConfig()

    // Merge configs based on debug mode
    if embedded.Metadata.Debug {
        // Debug build - apply debug_override section
        if user.DebugOverride != nil {
            // Merge debug overrides into embedded config
            embedded = mergeConfig(embedded, user.DebugOverride)

            // Track which overrides were applied
            activeOverrides = collectOverrides(user.DebugOverride)
        }

        // Show warning on EVERY command
        displayDebugWarning(activeOverrides)
    } else {
        // Production build - ignore debug_override section
        if user.DebugOverride != nil {
            // Warn user that debug_override is being ignored
            warnDebugOverrideIgnored()
        }

        // Strict override enforcement
        enforceOverrideRules()
    }

    // Apply normal user preferences (always applied)
    config = applyUserPreferences(embedded, user)

    return config, nil
}
```

**Binary naming convention**:
- Production: `mycli` (debug: false)
- Development: `mycli-dev` (debug: true)
- Testing: `mycli-test` (debug: true)

This makes it visually obvious which binary is being used.

---

## Configuration Categories

### ❌ NOT Overridable (Locked to Embedded Config)

These settings are **NEVER** user-configurable for security/branding reasons:

#### **metadata section** (100% locked)
- `metadata.name` - Binary name
- `metadata.version` - Binary version
- `metadata.description` - Description
- `metadata.long_description` - Long description
- `metadata.author.*` - Author info
- `metadata.license` - License
- `metadata.homepage` - Homepage URL
- `metadata.bugs_url` - Bug report URL
- `metadata.docs_url` - Documentation URL

**Rationale**: Branding and identity should be consistent across all users.

#### **branding section** (100% locked)
- `branding.colors.*` - Color scheme
- `branding.ascii_art` - ASCII banner
- `branding.prompts.*` - Prompt symbols
- `branding.theme.*` - Theme settings

**Rationale**: Company branding should be consistent.

#### **api section** (100% locked)
- `api.openapi_url` - OpenAPI spec URL
- `api.base_url` - **API base URL (SECURITY BOUNDARY)**
- `api.version` - API version
- `api.environments[]` - Multi-environment config
- `api.default_headers` - Default headers
- `api.user_agent` - User agent string
- `api.telemetry_url` - Telemetry endpoint URL (where to send usage data)

**Rationale**: The `api` section defines WHAT API to talk to and WHERE to send data. This should be locked to embedded config to prevent users from pointing the CLI to wrong APIs, overriding company-defined headers, bypassing API versioning, or redirecting telemetry to untrusted endpoints. Critical security and consistency boundary.

---

## ✅ FINALIZED Architecture

### New `defaults` Section (Embedded Config)

The embedded config includes a **`defaults`** section with sensible defaults that users can override:

```yaml
# In embedded config (cli-config.yaml)
defaults:
  http:
    timeout: 30s  # Request timeout

  caching:
    enabled: true  # Enable response caching

  pagination:
    limit: 20  # Default page size

  output:
    format: json  # json, yaml, table, csv
    pretty_print: true
    color: auto  # auto, always, never
    paging: true  # Use pager for long output

  deprecations:
    always_show: false  # Show once then cache (except critical/removed)
    min_severity: info  # info, warning, urgent, critical, removed

  retry:
    max_attempts: 3  # Retry attempts for failed requests
```

### User Config: `preferences` Section

Users override defaults in their `preferences` section:

```yaml
# In user config (~/.config/mycli/config.yaml)
preferences:
  http:
    timeout: 60s  # Override embedded default
    proxy: http://proxy.corp.com:8080
    ca_bundle: /etc/ssl/certs/corp-ca.pem

  output:
    format: yaml  # Prefer YAML
    color: always

  telemetry:
    enabled: false  # Opt-out (defaults to false)
```

### Locked Behaviors (Cannot Override - Except Debug Mode)

```yaml
behaviors:
  auth:
    type: oauth2  # LOCKED

  retry:
    enabled: true  # LOCKED
    initial_delay: 1s  # LOCKED
    max_delay: 30s  # LOCKED
    backoff_multiplier: 2.0  # LOCKED
    retry_on_status: [429, 500, 502, 503, 504]  # LOCKED

  caching:
    spec_ttl: 5m  # LOCKED
    response_ttl: 1m  # LOCKED
    directory: ~/.cache/mycli  # LOCKED
    max_size: 100MB  # LOCKED

  pagination:
    max_limit: 100  # LOCKED - prevents API abuse
    delay: 100ms  # LOCKED - inter-page delay for auto-pagination

  secrets:
    enabled: true  # LOCKED
    # ... all secret patterns LOCKED

  builtin_commands:
    # All LOCKED

  global_flags:
    # All LOCKED
```

### Removed Sections (Not Needed)

The following sections have been **REMOVED** from the configuration schema:

- ❌ `behaviors.http` - All settings locked (no user-configurable parts)
- ❌ `behaviors.rate_limiting` - Redundant with retry logic and server-side 429 handling
- ❌ `commands` - Users should use shell aliases instead
- ❌ `hooks` - Users should wrap CLI with shell scripts instead
- ❌ `plugins` - Too complex for v1.0, security concerns
- ❌ `features.offline_mode` - Redundant with caching
- ❌ `features.shell_completion` - Always enabled
- ❌ `features.validate_requests` - Always validate requests
- ❌ `features.validate_responses` - Never validate responses (too slow)
- ❌ `behaviors.pagination.auto_paginate` - Flag/ENV only (`--auto-page`)

---

## Complete Override Matrix

### Locked Settings (Embedded Only)

| Section | Settings | Count |
|---------|----------|-------|
| `metadata.*` | name, version, description, author, license, etc. | ~9 |
| `branding.*` | colors, ascii_art, prompts, theme | ~4 |
| `api.*` | openapi_url, base_url, version, environments, headers, user_agent, telemetry_url | ~7 |
| `behaviors.auth.*` | type, credentials, OAuth2 config | ~5 |
| `behaviors.retry.*` | enabled, delays, backoff, retry_on_status | ~5 |
| `behaviors.caching.*` | spec_ttl, response_ttl, directory, max_size | ~4 |
| `behaviors.pagination.*` | max_limit, delay | ~2 |
| `behaviors.secrets.*` | All secret detection and masking config | ~10 |
| `behaviors.builtin_commands.*` | All built-in command configuration | ~10 |
| `behaviors.global_flags.*` | All global flag configuration | ~15 |

**Total locked**: ~71 settings

### Overridable Settings (defaults → preferences)

| Embedded (`defaults`) | User Override (`preferences`) | ENV | Flag |
|----------------------|------------------------------|-----|------|
| `defaults.http.timeout` | `preferences.http.timeout` | `MYCLI_TIMEOUT` | `--timeout` |
| `defaults.caching.enabled` | `preferences.caching.enabled` | `MYCLI_NO_CACHE` | `--no-cache` |
| `defaults.pagination.limit` | `preferences.pagination.limit` | `MYCLI_PAGE_LIMIT` | `--limit` |
| `defaults.output.format` | `preferences.output.format` | `MYCLI_OUTPUT` | `--output`, `-o` |
| `defaults.output.pretty_print` | `preferences.output.pretty_print` | `MYCLI_PRETTY_PRINT` | `--pretty`, `--no-pretty` |
| `defaults.output.color` | `preferences.output.color` | `NO_COLOR` | `--color`, `--no-color` |
| `defaults.output.paging` | `preferences.output.paging` | `MYCLI_PAGING` | `--paging`, `--no-paging` |
| `defaults.deprecations.always_show` | `preferences.deprecations.always_show` | `MYCLI_DEPRECATIONS_ALWAYS_SHOW` | `--deprecations-always-show` |
| `defaults.deprecations.min_severity` | `preferences.deprecations.min_severity` | `MYCLI_DEPRECATIONS_MIN_SEVERITY` | `--deprecations-min-severity` |
| `defaults.retry.max_attempts` | `preferences.retry.max_attempts` | `MYCLI_RETRY` | `--retry` |
| *(not in embedded)* | `preferences.http.proxy` | `HTTP_PROXY`, `HTTPS_PROXY` | `--proxy` |
| *(not in embedded)* | `preferences.http.ca_bundle` | `MYCLI_CA_BUNDLE` | `--ca-bundle` |
| *(not in embedded)* | `preferences.telemetry.enabled` | `MYCLI_TELEMETRY` | `--telemetry` |

**Total overridable**: 13 settings

### Flag-Only Options (No Config)

| Feature | ENV | Flag | Notes |
|---------|-----|------|-------|
| Auto-pagination | `MYCLI_AUTO_PAGE` | `--auto-page` | Not in config - prevents API abuse |

---

## ⚠️ DEPRECATED - Old Proposed Sections (For Reference Only)

The following sections show the OLD proposal before finalization. **IGNORE THESE.**

---

### **updates section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `updates.enabled` | ❌ | ❌ | ✅ | Enable update checks | ⚠️ REVIEW |
| `updates.update_url` | ❌ | ❌ | ✅ | Update server URL | ⚠️ REVIEW |
| `updates.check_interval` | ❌ | ❌ | ✅ | How often to check | ⚠️ REVIEW |
| `updates.public_key` | ❌ | ❌ | ✅ | Signature verification key | ⚠️ REVIEW |
| `updates.auto_install` | ❌ | ❌ | ✅ User-only | **Can ONLY be set in user config** | ✅ USER-ONLY |

**Questions**:
- Should users be able to disable update checks entirely?
- Should users be able to change `update_url` to point to different update servers?
- Should signature verification public key be overridable?

---

### **behaviors.auth section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `behaviors.auth.type` | ❌ | ❌ | ✅ | Auth type (none, api_key, oauth2, basic) | ⚠️ REVIEW |
| `behaviors.auth.api_key.header` | ❌ | ❌ | ✅ | Header name | ⚠️ REVIEW |
| `behaviors.auth.api_key.env_var` | ❌ | ❌ | ✅ | Which ENV var to read | ⚠️ REVIEW |
| `behaviors.auth.oauth2.*` | ❌ | ❌ | ✅ | OAuth2 configuration | ⚠️ REVIEW |
| `behaviors.auth.basic.*` | ❌ | ❌ | ✅ | Basic auth configuration | ⚠️ REVIEW |

**Questions**:
- Should auth mechanism type be user-configurable?
- Should OAuth2 client IDs, URLs, etc. be overridable?

---

### **behaviors.caching section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `behaviors.caching.enabled` | ❌ | ✅ `--no-cache` | ✅ | Enable caching | ⚠️ REVIEW |
| `behaviors.caching.spec_ttl` | ❌ | ❌ | ✅ | Spec cache TTL | ⚠️ REVIEW |
| `behaviors.caching.response_ttl` | ❌ | ❌ | ✅ | Response cache TTL | ⚠️ REVIEW |
| `behaviors.caching.directory` | ❌ | ❌ | ✅ | Cache directory | ⚠️ REVIEW |
| `behaviors.caching.max_size` | ❌ | ❌ | ✅ | Max cache size | ⚠️ REVIEW |

**Questions**:
- Should cache TTLs be user-configurable?
- Should cache directory location be overridable?

---

### **behaviors.retry section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `behaviors.retry.enabled` | ❌ | ❌ | ✅ | Enable retry | ⚠️ REVIEW |
| `behaviors.retry.max_attempts` | ✅ `{CLI}_RETRY` | ✅ `--retry` | ✅ | Max retry attempts | ⚠️ REVIEW |
| `behaviors.retry.initial_delay` | ❌ | ❌ | ✅ | Initial delay | ⚠️ REVIEW |
| `behaviors.retry.max_delay` | ❌ | ❌ | ✅ | Max delay | ⚠️ REVIEW |
| `behaviors.retry.backoff_multiplier` | ❌ | ❌ | ✅ | Backoff multiplier | ⚠️ REVIEW |
| `behaviors.retry.retry_on_status[]` | ❌ | ❌ | ✅ | Which status codes to retry | ⚠️ REVIEW |

**Questions**:
- Should retry logic be user-configurable?
- Should users be able to disable retry entirely?

---

### **behaviors.output section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `behaviors.output.default_format` | ✅ `{CLI}_OUTPUT_FORMAT` | ✅ `--output`/`-o` | ✅ | json, yaml, table, csv | ✅ OVERRIDE OK |
| `behaviors.output.pretty_print` | ❌ | ❌ | ✅ | Pretty print output | ✅ OVERRIDE OK |
| `behaviors.output.color` | ✅ `NO_COLOR` | ✅ `--no-color` | ✅ | auto, always, never | ✅ OVERRIDE OK |
| `behaviors.output.paging` | ❌ | ❌ | ✅ | Enable paging | ✅ OVERRIDE OK |
| `behaviors.output.pager` | ✅ `PAGER` | ❌ | ✅ | Pager command | ✅ OVERRIDE OK |

**Note**: Output preferences are clearly user-specific and should be overridable.

---

### **behaviors.pagination section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `behaviors.pagination.default_limit` | ❌ | ❌ | ✅ | Default page size | ⚠️ REVIEW |
| `behaviors.pagination.max_limit` | ❌ | ❌ | ✅ | Max page size | ⚠️ REVIEW |
| `behaviors.pagination.auto_paginate` | ❌ | ❌ | ✅ | Auto-fetch all pages | ⚠️ REVIEW |

**Questions**:
- Should users be able to change pagination defaults?
- Should max_limit be enforced by embedded config only?

---

### **behaviors.notifications section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `behaviors.notifications.show_changelog` | ❌ | ❌ | ✅ | Show changelog on update | ✅ OVERRIDE OK |
| `behaviors.notifications.show_deprecations` | ❌ | ✅ `--no-deprecation-warnings` | ✅ | Show deprecation warnings | ✅ OVERRIDE OK |
| `behaviors.notifications.check_interval` | ❌ | ❌ | ✅ | Check interval | ✅ OVERRIDE OK |

**Note**: Notification preferences are clearly user-specific.

---

### **behaviors.secrets section**

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| `behaviors.secrets.enabled` | ❌ | ❌ | ✅ | Enable secret masking | ⚠️ REVIEW |
| `behaviors.secrets.masking.style` | ❌ | ❌ | ✅ | partial, full, hash | ⚠️ REVIEW |
| `behaviors.secrets.masking.partial_show_chars` | ❌ | ❌ | ✅ | How many chars to show | ⚠️ REVIEW |
| `behaviors.secrets.masking.replacement` | ❌ | ❌ | ✅ | Replacement string | ⚠️ REVIEW |
| `behaviors.secrets.field_patterns[]` | ❌ | ❌ | ✅ | Field name patterns | ⚠️ REVIEW |
| `behaviors.secrets.value_patterns[]` | ❌ | ❌ | ✅ | Value regex patterns | ⚠️ REVIEW |
| `behaviors.secrets.explicit_fields[]` | ❌ | ❌ | ✅ | Explicit field paths | ⚠️ REVIEW |
| `behaviors.secrets.headers[]` | ❌ | ❌ | ✅ | Headers to mask | ⚠️ REVIEW |
| `behaviors.secrets.mask_in.*` | ❌ | ❌ | ✅ | Where to apply masking | ⚠️ REVIEW |

**Questions**:
- Should users be able to disable secret masking? (Security risk)
- Should secret detection patterns be user-configurable?

---

### **behaviors.builtin_commands section**

| Setting Category | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|-----------------|--------|---------|-----------|----------|------------|
| `version.enabled` | ❌ | ❌ | ✅ | Enable version command | ⚠️ REVIEW |
| `version.style` | ❌ | ❌ | ✅ | subcommand, flag, hybrid | ⚠️ REVIEW |
| `version.flags[]` | ❌ | ❌ | ✅ | Which flags to support | ⚠️ REVIEW |
| `version.show_api_version` | ❌ | ❌ | ✅ | Show API version | ⚠️ REVIEW |
| `version.show_build_info` | ❌ | ❌ | ✅ | Show build info | ⚠️ REVIEW |
| *(similar for all builtin commands)* | ❌ | ❌ | ✅ | help, info, config, completion, update, changelog, deprecations, cache, auth | ⚠️ REVIEW |

**Questions**:
- Should users be able to disable built-in commands?
- Should command behavior be user-configurable?

---

### **behaviors.global_flags section** (configuration metadata)

| Setting Category | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|-----------------|--------|---------|-----------|----------|------------|
| `config.enabled` | ❌ | ❌ | ✅ | Enable --config flag | ⚠️ REVIEW |
| `config.flag` | ❌ | ❌ | ✅ | Flag name (--config) | ⚠️ REVIEW |
| `config.short` | ❌ | ❌ | ✅ | Short form (-c) | ⚠️ REVIEW |
| `config.env_var` | ❌ | ❌ | ✅ | ENV var name | ⚠️ REVIEW |
| `config.description` | ❌ | ❌ | ✅ | Help text | ⚠️ REVIEW |
| *(similar for all global flags)* | ❌ | ❌ | ✅ | profile, region, output, verbose, quiet, debug, no_color, timeout, retry, no_cache, yes | ⚠️ REVIEW |

**Questions**:
- Should global flag metadata (names, descriptions) be user-configurable?
- Or should this be locked to embedded config for consistency?

**Note**: The *values* of these flags at runtime are definitely user-controllable (see next section).

---

### **Runtime Global Flag Values** (clearly overridable)

| Flag | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|------|--------|---------|-----------|----------|------------|
| Config file path | ✅ `{CLI}_CONFIG` | ✅ `--config`/`-c` | ❌ | Bootstrap setting | ✅ OVERRIDE OK |
| Profile | ✅ `{CLI}_PROFILE` | ✅ `--profile` | ✅ | Which profile to use | ✅ OVERRIDE OK |
| Region | ✅ `{CLI}_REGION` | ✅ `--region`/`-r` | ✅ | Region/datacenter | ✅ OVERRIDE OK |
| Output format | ✅ `{CLI}_OUTPUT_FORMAT` | ✅ `--output`/`-o` | ✅ | json, yaml, table, csv | ✅ OVERRIDE OK |
| Verbose | ✅ `{CLI}_VERBOSE` | ✅ `--verbose`/`-v` | ✅ | Verbose output | ✅ OVERRIDE OK |
| Quiet | ✅ `{CLI}_QUIET` | ✅ `--quiet`/`-q` | ✅ | Quiet mode | ✅ OVERRIDE OK |
| Debug | ✅ `{CLI}_DEBUG` | ✅ `--debug` | ✅ | Debug logging | ✅ OVERRIDE OK |
| No color | ✅ `NO_COLOR` | ✅ `--no-color` | ✅ | Disable colors | ✅ OVERRIDE OK |
| Timeout | ✅ `{CLI}_TIMEOUT` | ✅ `--timeout` | ✅ | Request timeout | ✅ OVERRIDE OK |
| Retry | ✅ `{CLI}_RETRY` | ✅ `--retry` | ✅ | Retry attempts | ✅ OVERRIDE OK |
| No cache | ✅ `{CLI}_NO_CACHE` | ✅ `--no-cache` | ✅ | Disable cache | ✅ OVERRIDE OK |
| Yes (non-interactive) | ✅ `{CLI}_YES` | ✅ `--yes`/`-y` | ✅ | Skip confirmations | ✅ OVERRIDE OK |

**Note**: These runtime values are clearly user preferences and should be fully overridable.

---

### **Enterprise Settings** (from environment/config file)

| Setting | 🌍 ENV | 🚩 Flag | 📄 Config | Proposed | Should Be? |
|---------|--------|---------|-----------|----------|------------|
| HTTP proxy | ✅ `HTTP_PROXY` / `http_proxy` | ❌ | ✅ | HTTP proxy URL | ✅ OVERRIDE OK |
| HTTPS proxy | ✅ `HTTPS_PROXY` / `https_proxy` | ❌ | ✅ | HTTPS proxy URL | ✅ OVERRIDE OK |
| No proxy | ✅ `NO_PROXY` / `no_proxy` | ❌ | ✅ | Bypass proxy for domains | ✅ OVERRIDE OK |
| CA bundle | ✅ `{CLI}_CA_BUNDLE` | ❌ | ✅ | Custom CA certificates | ✅ OVERRIDE OK |
| TLS insecure | ✅ `{CLI}_TLS_INSECURE` | ❌ | ✅ | Skip TLS verification | ✅ OVERRIDE OK (with warnings) |

**Note**: Enterprise proxy/TLS settings are clearly environment-specific and should be overridable.

---

## Summary Statistics (FINALIZED)

**Locked settings** (embedded only):
- `metadata.*` (9 settings) - Branding/identity
- `branding.*` (4 settings) - Company branding
- `api.*` (7 settings) - **ENTIRE api section locked** (including telemetry_url)
- `behaviors.auth.*` (5 settings) - Auth configuration
- `behaviors.retry.*` (5 settings) - Retry behavior (except max_attempts)
- `behaviors.caching.*` (4 settings) - Cache behavior (except enabled)
- `behaviors.pagination.*` (2 settings) - Pagination limits (max_limit, delay)
- `behaviors.secrets.*` (10 settings) - Secret masking
- `behaviors.builtin_commands.*` (10 settings) - Built-in commands
- `behaviors.global_flags.*` (15 settings) - Global flags

**Total locked**: ~71 settings (when `metadata.debug: false`)

**Overridable settings** (`defaults` → `preferences`):
- `defaults.http.timeout` → `preferences.http.timeout`
- `defaults.caching.enabled` → `preferences.caching.enabled`
- `defaults.pagination.limit` → `preferences.pagination.limit`
- `defaults.output.*` (4 settings) → `preferences.output.*`
- `defaults.deprecations.*` (2 settings) → `preferences.deprecations.*`
- `defaults.retry.max_attempts` → `preferences.retry.max_attempts`
- Enterprise: `preferences.http.proxy`, `preferences.http.ca_bundle`
- User-only: `preferences.telemetry.enabled`

**Total overridable**: 13 settings

**Removed sections**: commands, hooks, plugins, rate_limiting, validate_*, offline_mode, completion

---

## 🔓 Debug Mode Exception

**When `metadata.debug: true`**:
- ⚠️ **ALL ~115 settings** become overridable via **`debug_override`** section in user config
- 🚨 Security warning displayed on EVERY command execution
- 🛠️ Intended ONLY for development/testing builds
- ❌ Should NEVER be enabled in production binaries

**Binary separation**:
- `mycli` (production) → `debug: false` → strict override rules, `debug_override` ignored
- `mycli-dev` (development) → `debug: true` → `debug_override` active + warnings

**Config file structure**:
```yaml
# Normal preferences (always work)
behaviors:
  output:
    default_format: yaml

# Debug-only overrides (only work when debug: true)
debug_override:
  api:
    base_url: http://localhost:8080  # Override embedded base_url
```

---

## ✅ All Questions SETTLED

1. **✅ SETTLED: API configuration is 100% locked to embedded config**
   - All of `api.*` section is embedded-only (security boundary)
   - Includes `api.telemetry_url` (prevents redirecting telemetry)
   - `timeout` moved to `defaults.http.timeout` (overridable)

2. **✅ SETTLED: Debug mode allows override escape hatch**
   - `metadata.debug: true` → ALL config overridable via `debug_override` section
   - `metadata.debug: false` → strict override rules (production builds)
   - Security warning displayed on EVERY command when debug enabled
   - Production builds ignore `debug_override` section with warning

3. **✅ SETTLED: Behaviors are 100% locked to embedded config**
   - Auth mechanisms (type, credentials, OAuth2 config) - LOCKED
   - Retry logic (delays, backoff, retry_on_status) - LOCKED (except max_attempts → defaults)
   - Caching policies (TTLs, directory, max_size) - LOCKED (except enabled → defaults)
   - Pagination (max_limit, delay) - LOCKED (except default_limit → defaults)
   - Secret masking (all patterns) - LOCKED
   - Built-in commands (all config) - LOCKED
   - Global flags (all config) - LOCKED

4. **✅ SETTLED: Rate limiting removed entirely**
   - Redundant with retry logic and server-side 429 handling
   - Auto-pagination uses simple hardcoded delay (100ms) in `behaviors.pagination.delay`

5. **✅ SETTLED: Updates configuration is 100% locked**
   - `updates.update_url`, `updates.check_interval`, `updates.public_key` - LOCKED
   - `updates.auto_install` removed from embedded config (user config only, not in spec yet)

6. **✅ SETTLED: Removed sections**
   - ❌ `commands` - Users should use shell aliases
   - ❌ `hooks` - Users should wrap CLI with scripts
   - ❌ `plugins` - Too complex, security concerns
   - ❌ `features.offline_mode` - Redundant with caching
   - ❌ `features.shell_completion` - Always enabled
   - ❌ `features.validate_*` - Always validate requests, never responses

7. **✅ SETTLED: Final architecture**
   - **Embedded config**: `defaults` section (13 overridable settings)
   - **User config**: `preferences` section (overrides defaults)
   - **User config**: `debug_override` section (only works in debug builds)
   - Clear separation: `defaults` (overridable) vs `behaviors` (locked)

---

## Final Decision: Option A (Minimal Override) ✅

**Chosen approach**:
- Lock ALL behaviors, API config, metadata, branding to embedded config
- Only allow override of:
  - ✅ Runtime preferences (output format, colors, paging, timeout)
  - ✅ Enterprise settings (proxy, CA bundle)
  - ✅ User-specific settings (telemetry opt-out, deprecation display, cache enabled)
  - ✅ Pagination defaults (limit, up to max_limit)
  - ✅ Retry max attempts (for CI/CD flexibility)
- Debug mode provides escape hatch for development/testing

---

**Version**: 0.8.0
**Last Updated**: 2025-01-11
**Status**: ✅ FINALIZED - Configuration override architecture complete
**Project**: CliForge - Forge CLIs from APIs

---

## Implementation Status

✅ **Completed**:
1. Updated `configuration-dsl.md` to v0.8.0 with finalized `defaults` section
2. Updated `configuration-dsl.md` to remove eliminated sections
3. Updated `builtin-commands-design.md` with final decisions
4. Updated `CHANGELOG.md` with v0.8.0 changes

📋 **Next Steps**:
- Begin POC implementation in Go
- Test configuration override behavior
- Validate security boundaries (debug mode, locked sections)
