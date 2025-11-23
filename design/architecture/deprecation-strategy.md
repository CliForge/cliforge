# CliForge Deprecation Strategy

## Overview

CliForge handles **two types of deprecations**:
1. **API Deprecations** - The underlying API deprecates endpoints/parameters
2. **CLI Deprecations** - The CLI tool itself changes behavior/commands/flags

This document defines how each is handled to ensure smooth transitions while maintaining backward compatibility.

---

## Table of Contents

1. [API Deprecations](#api-deprecations)
2. [CLI Deprecations](#cli-deprecations)
3. [Communication Strategy](#communication-strategy)
4. [Implementation Details](#implementation-details)
5. [Best Practices](#best-practices)

---

## API Deprecations

### Detection Methods

#### 1. OpenAPI `deprecated` Field

Standard OpenAPI way to mark deprecations:

```yaml
paths:
  /v1/users:
    get:
      operationId: listUsersV1
      deprecated: true
      summary: List users (deprecated)
      description: |
        ⚠️ DEPRECATED: This endpoint will be removed on 2025-12-31.
        Use GET /v2/users instead.
```

**Parameters can also be deprecated**:

```yaml
parameters:
  - name: filter
    in: query
    deprecated: true
    description: |
      ⚠️ DEPRECATED: Use 'search' parameter instead.
      Will be removed in API v3.0.
    schema:
      type: string
```

#### 2. `x-cli-deprecation` Extension (Enhanced)

For more detailed deprecation info:

```yaml
paths:
  /v1/users:
    get:
      operationId: listUsersV1
      deprecated: true
      x-cli-deprecation:
        # When it will be removed
        sunset: "2025-12-31"

        # Alternative to use instead
        replacement:
          operation: listUsersV2
          path: /v2/users
          migration: |
            Replace:
              my-cli users list --filter "name=john"
            With:
              my-cli users list --search "name:john"

        # Reason for deprecation
        reason: "v1 API has performance issues and lacks pagination"

        # Severity level
        severity: warning  # info, warning, breaking

        # Link to migration guide
        docs_url: "https://docs.example.com/migration/v1-to-v2"
```

#### 3. HTTP `Sunset` Header (RFC 8594)

API can send sunset information in response headers:

```http
HTTP/1.1 200 OK
Sunset: Sat, 31 Dec 2025 23:59:59 GMT
Deprecation: true
Link: </v2/users>; rel="successor-version"
```

**CliForge detects and displays these headers.**

---

### CLI Behavior for API Deprecations

#### Warning Display

**When user executes deprecated command**:

```bash
$ my-cli users list

⚠️  DEPRECATION WARNING
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Command: users list
  Status: DEPRECATED
  Sunset: December 31, 2025

  This endpoint will be removed in 45 days.

  Migration:
    Use 'my-cli users list-v2' instead

    Old: my-cli users list --filter "name=john"
    New: my-cli users list-v2 --search "name:john"

  Docs: https://docs.example.com/migration/v1-to-v2

  To suppress this warning: --no-deprecation-warnings
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

[... normal command output follows ...]
```

#### Warning Levels

**Based on time until sunset**:

| Time Remaining | Level | Display |
|----------------|-------|---------|
| > 6 months | `info` | ℹ️ Informational (shown once per week) |
| 3-6 months | `warning` | ⚠️ Warning (shown every time) |
| 1-3 months | `urgent` | 🚨 Urgent (shown every time, yellow) |
| < 1 month | `critical` | 🔴 Critical (shown every time, red, requires `--force`) |
| Past sunset | `removed` | ❌ Blocked (returns error, suggest alternative) |

#### Suppression Options

Users can suppress warnings:

```bash
# Suppress for single command
my-cli users list --no-deprecation-warnings

# Suppress globally (in user config)
my-cli config set warnings.deprecations false

# Suppress for specific operations
my-cli config set warnings.suppress listUsersV1
```

#### Migration Assistance

**Show migration command**:

```bash
# CliForge can suggest the new command
$ my-cli users list --filter "name=john" --show-migration

Migration suggestion:
  my-cli users list-v2 --search "name:john"

Run it now? [y/N]: _
```

---

### Tracking Deprecation Usage

**Audit log** (if enabled):

```yaml
# ~/.local/share/my-cli/audit.log
2025-11-11T15:30:00Z DEPRECATION operation=listUsersV1 sunset=2025-12-31 days_remaining=45
```

**Usage stats** (if telemetry enabled):

Send anonymous stats to help API owners understand migration progress:

```json
{
  "event": "deprecated_operation_used",
  "operation": "listUsersV1",
  "days_until_sunset": 45,
  "cli_version": "1.2.0"
}
```

---

## CLI Deprecations

### When CLI Behavior Changes

Sometimes the **CLI itself** needs to deprecate features:
- Flag names change
- Command structure reorganizes
- Output format changes
- Default behavior changes

### CLI Deprecation Metadata

Stored in embedded config or update metadata:

```yaml
# In cli-config.yaml or update metadata
cli_deprecations:
  - type: flag
    command: users list
    old_flag: --filter
    new_flag: --search
    sunset: "2025-12-31"
    reason: "Aligning with API v2 parameter names"
    migration: "Replace --filter with --search"

  - type: command
    old_command: users ls
    new_command: users list
    sunset: "2026-01-31"
    reason: "Standardizing command names"

  - type: behavior
    command: users create
    change: "Now requires --confirm flag for safety"
    sunset: "2025-11-30"
    severity: breaking
```

### Automatic Flag Mapping

**CliForge can auto-map deprecated flags**:

```bash
$ my-cli users list --filter "name=john"

⚠️  Flag '--filter' is deprecated
    Using '--search' instead
    This mapping will be removed on 2025-12-31

[... command executes with --search ...]
```

**Implementation**:

```go
// In CLI runtime
func (r *Runtime) mapDeprecatedFlags(cmd *cobra.Command, args []string) {
    for _, deprecation := range r.config.CLIDeprecations {
        if deprecation.Type == "flag" && cmd.Name() == deprecation.Command {
            if cmd.Flags().Lookup(deprecation.OldFlag) != nil {
                // Auto-map to new flag
                oldValue := cmd.Flags().Lookup(deprecation.OldFlag).Value.String()
                cmd.Flags().Set(deprecation.NewFlag, oldValue)

                // Show warning
                showDeprecationWarning(deprecation)
            }
        }
    }
}
```

### Command Aliases During Transition

**Note**: As of CliForge v0.8.0, the `commands:` section was removed from the configuration DSL. Users should use **shell aliases** for custom command shortcuts.

**Deprecated OpenAPI operations still work** (via operation ID mapping):

```yaml
# In OpenAPI spec
paths:
  /users:
    get:
      operationId: listUsersV1
      deprecated: true
      x-cli-deprecation:
        sunset_date: "2026-01-31"
        replacement: "listUsersV2"
        message: "Use the v2 endpoint for better performance"
```

**Shell alias approach** (user's ~/.bashrc or ~/.zshrc):

```bash
# Deprecated shorthand (user maintains their own aliases)
alias mycli-users-ls='mycli users list'

# With deprecation warning (user can add custom warnings)
mycli-users-ls() {
  echo "⚠️  'mycli-users-ls' is deprecated, use 'mycli users list' instead" >&2
  mycli users list "$@"
}
```

When the deprecated operation is called directly:

```bash
$ my-cli users list-v1

⚠️  Operation 'listUsersV1' is deprecated
    Use 'my-cli users list' (listUsersV2) instead
    This endpoint will be removed on 2026-01-31
    Reason: Use the v2 endpoint for better performance

[... command executes ...]
```

---

## Communication Strategy

### 1. In-CLI Notifications

**On first run after update**:

```bash
$ my-cli users list

╔════════════════════════════════════════════════════════╗
║  📢 CLI Update: v1.2.0 → v1.3.0                        ║
╠════════════════════════════════════════════════════════╣
║                                                        ║
║  Deprecations in this release:                        ║
║                                                        ║
║  ⚠️  Flag --filter → --search                         ║
║     Affected: users list, posts list                  ║
║     Sunset: December 31, 2025                         ║
║                                                        ║
║  ℹ️  Command 'users ls' → 'users list'                ║
║     Sunset: January 31, 2026                          ║
║                                                        ║
║  Run 'my-cli deprecations' for details                ║
║  Run 'my-cli changelog' for full release notes        ║
║                                                        ║
╚════════════════════════════════════════════════════════╝
```

### 2. Dedicated `deprecations` Command

```bash
$ my-cli deprecations

Active Deprecations
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

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

CLI Deprecations (1):

  ℹ️  INFO - 120 days remaining
  ├─ Alias: users ls → users list
  ├─ Sunset: March 31, 2026
  └─ Impact: Alias will no longer work

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Show details: my-cli deprecations show <operation-id>
Scan usage: my-cli deprecations scan
```

### 3. Proactive Scanning

**Scan scripts for deprecated usage**:

```bash
$ my-cli deprecations scan ./scripts/

Scanning for deprecated CLI usage...

Found 3 deprecations in your scripts:

  ./scripts/deploy.sh:12
  ├─ Command: my-cli users ls --filter "role=admin"
  ├─ Issue: Command 'users ls' deprecated
  ├─ Issue: Flag '--filter' deprecated
  └─ Suggested fix:
      my-cli users list --search "role:admin"

  ./scripts/backup.sh:45
  ├─ Command: my-cli posts create --no-publish
  ├─ Issue: Flag '--no-publish' deprecated
  └─ Suggested fix:
      my-cli posts create --draft

  ./ci/pipeline.yml:89
  ├─ Command: my-cli deploy --env prod
  ├─ Issue: Command uses deprecated 'deploy' operation
  └─ Suggested fix:
      my-cli deployments create --environment production

Apply fixes automatically? [y/N]: _
```

### 4. Email/Newsletter (for SaaS)

**API owners can configure notification emails**:

```yaml
# In cli-config.yaml
notifications:
  deprecation_emails:
    enabled: true
    endpoint: https://api.example.com/cli/deprecations
    # CLI sends usage stats, API owner sends targeted emails
```

---

## Implementation Details

### OpenAPI Extension Schema

**Complete `x-cli-deprecation` schema**:

```yaml
x-cli-deprecation:
  type: object
  properties:
    sunset:
      type: string
      format: date
      description: ISO 8601 date when feature will be removed
      example: "2025-12-31"

    replacement:
      type: object
      properties:
        operation:
          type: string
          description: Operation ID of replacement
        path:
          type: string
          description: API path of replacement
        command:
          type: string
          description: CLI command to use instead
        migration:
          type: string
          description: Migration instructions
      example:
        operation: listUsersV2
        path: /v2/users
        command: my-cli users list-v2
        migration: "Replace --filter with --search parameter"

    reason:
      type: string
      description: Why this is being deprecated
      example: "Performance issues, lacks pagination"

    severity:
      type: string
      enum: [info, warning, breaking]
      default: warning
      description: Severity of deprecation

    docs_url:
      type: string
      format: uri
      description: Link to migration guide
      example: "https://docs.example.com/migration/v1-to-v2"

    breaking_changes:
      type: array
      items:
        type: string
      description: List of breaking changes
      example:
        - "Response format changed from array to object"
        - "Pagination now required"
```

### Deprecation Detection Logic

```go
package deprecation

import (
    "time"
    "github.com/getkin/kin-openapi/openapi3"
)

type DeprecationInfo struct {
    Type        DeprecationType  // Operation, Parameter, Schema
    Deprecated  bool
    Sunset      *time.Time
    Replacement *Replacement
    Reason      string
    Severity    Severity
    DocsURL     string
}

type DeprecationType string
const (
    DeprecationOperation DeprecationType = "operation"
    DeprecationParameter DeprecationType = "parameter"
    DeprecationSchema    DeprecationType = "schema"
)

type Severity string
const (
    SeverityInfo     Severity = "info"
    SeverityWarning  Severity = "warning"
    SeverityUrgent   Severity = "urgent"
    SeverityCritical Severity = "critical"
    SeverityRemoved  Severity = "removed"
)

type Replacement struct {
    OperationID string
    Path        string
    Command     string
    Migration   string
}

// DetectDeprecation checks if an operation is deprecated
func DetectDeprecation(op *openapi3.Operation) *DeprecationInfo {
    if op.Deprecated {
        info := &DeprecationInfo{
            Type:       DeprecationOperation,
            Deprecated: true,
        }

        // Check for x-cli-deprecation extension
        if ext, ok := op.Extensions["x-cli-deprecation"]; ok {
            parseDeprecationExtension(ext, info)
        }

        // Calculate severity based on sunset date
        if info.Sunset != nil {
            info.Severity = calculateSeverity(*info.Sunset)
        }

        return info
    }

    return nil
}

// calculateSeverity based on time remaining
func calculateSeverity(sunset time.Time) Severity {
    daysRemaining := int(time.Until(sunset).Hours() / 24)

    if daysRemaining < 0 {
        return SeverityRemoved
    } else if daysRemaining < 30 {
        return SeverityCritical
    } else if daysRemaining < 90 {
        return SeverityUrgent
    } else if daysRemaining < 180 {
        return SeverityWarning
    } else {
        return SeverityInfo
    }
}

// ShowDeprecationWarning displays warning to user
func ShowDeprecationWarning(info *DeprecationInfo) {
    if shouldShowWarning(info) {
        displayWarning(info)
        recordUsage(info)
    }
}

// shouldShowWarning based on severity and user preferences
func shouldShowWarning(info *DeprecationInfo) bool {
    // Check user preferences
    if config.Get("warnings.deprecations") == false {
        return false
    }

    // Info level: show once per week
    if info.Severity == SeverityInfo {
        return !wasShownRecently(info, 7*24*time.Hour)
    }

    // Warning and above: show every time
    return true
}
```

### Warning Display Implementation

```go
package ui

import "fmt"

func DisplayDeprecationWarning(info *DeprecationInfo) {
    icon := getIcon(info.Severity)
    color := getColor(info.Severity)

    fmt.Printf("\n%s DEPRECATION %s\n", icon, strings.ToUpper(string(info.Severity)))
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

    if info.Sunset != nil {
        daysRemaining := int(time.Until(*info.Sunset).Hours() / 24)
        fmt.Printf("  This feature will be removed in %d days (%s)\n",
            daysRemaining, info.Sunset.Format("January 2, 2006"))
    }

    if info.Reason != "" {
        fmt.Printf("  Reason: %s\n", info.Reason)
    }

    if info.Replacement != nil {
        fmt.Printf("\n  Migration:\n")
        if info.Replacement.Command != "" {
            fmt.Printf("    Use: %s\n", info.Replacement.Command)
        }
        if info.Replacement.Migration != "" {
            fmt.Printf("    %s\n", info.Replacement.Migration)
        }
    }

    if info.DocsURL != "" {
        fmt.Printf("\n  Docs: %s\n", info.DocsURL)
    }

    fmt.Println("\n  To suppress: --no-deprecation-warnings")
    fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
    fmt.Println()
}
```

---

## Best Practices

### For API Owners

#### 1. Give Adequate Notice

**Minimum timelines**:
- **Minor changes** (parameter renamed): 3 months
- **Major changes** (endpoint restructured): 6 months
- **Breaking changes** (removed entirely): 12 months

#### 2. Use Semantic Versioning for APIs

```
/v1/users  → /v2/users
```

Keep v1 running alongside v2 during transition period.

#### 3. Document Migration Paths

Always provide:
- ✅ What's changing
- ✅ Why it's changing
- ✅ How to migrate
- ✅ Timeline for removal
- ✅ Link to migration guide

#### 4. Gradual Rollout

**Phase 1: Announce**
- Add `deprecated: true` to spec
- Show info warnings

**Phase 2: Warn**
- Increase warning severity
- Send email notifications

**Phase 3: Require Opt-In**
- Require `--force` flag for critical deprecations
- Block new usage, allow existing

**Phase 4: Remove**
- Return 410 Gone
- Suggest alternative in error message

### For CLI Tool Developers (CliForge Users)

#### 1. Monitor Deprecations

Add to CI/CD:

```bash
# In CI pipeline
my-cli deprecations scan ./scripts/ --fail-on-critical
```

#### 2. Subscribe to Changelogs

```bash
my-cli config set notifications.show_changelog true
```

#### 3. Test with Latest CLI

Regularly update CLI to see deprecation warnings before they become critical.

#### 4. Automate Migration

Use `deprecations scan` to find and fix issues:

```bash
my-cli deprecations scan ./scripts/ --auto-fix
```

---

## User Configuration Options

```yaml
# ~/.config/my-cli/config.yaml

warnings:
  # Global deprecation warning toggle
  deprecations: true

  # Suppress specific operations
  suppress:
    - listUsersV1
    - getPostsOld

  # Severity threshold (only show this level and above)
  min_severity: warning  # info, warning, urgent, critical

  # Show warnings in CI environments
  show_in_ci: false

  # Rate limiting for info-level warnings
  info_cooldown: 168h  # Show once per week

notifications:
  # Show changelog on CLI update
  show_changelog: true

  # Check for deprecations on startup
  check_on_startup: true

  # Email notifications (if API supports it)
  email_updates: true
```

---

## Example: Complete Deprecation Flow

### API Owner's Perspective

**1. Deprecate endpoint in OpenAPI spec**:

```yaml
paths:
  /v1/users:
    get:
      operationId: listUsersV1
      deprecated: true
      x-cli-deprecation:
        sunset: "2026-06-30"
        replacement:
          operation: listUsersV2
          path: /v2/users
          command: my-cli users list
          migration: "Use /v2/users endpoint. Changes: pagination now required, filter syntax changed to search."
        reason: "v1 API lacks pagination and has performance issues"
        severity: warning
        docs_url: "https://docs.example.com/api/migration/v1-to-v2"

  /v2/users:
    get:
      operationId: listUsersV2
      summary: List users (v2)
      # ... new endpoint definition
```

**2. Update changelog**:

```yaml
info:
  x-cli-changelog:
    - date: "2025-11-11"
      version: "2.0.0"
      changes:
        - type: deprecated
          severity: warning
          description: "GET /v1/users is deprecated"
          path: "/v1/users"
          migration: "Use GET /v2/users instead"
          sunset: "2026-06-30"
```

### User's Experience

**Month 1-3: Info warnings**

```bash
$ my-cli users list

ℹ️  This command uses a deprecated endpoint
   It will be removed on June 30, 2026 (231 days remaining)
   Use 'my-cli users list' (v2) instead

[... command output ...]
```

**Month 4-6: Regular warnings**

```bash
$ my-cli users list

⚠️  DEPRECATION WARNING
   This endpoint will be removed in 89 days (June 30, 2026)
   Migration: Use 'my-cli users list' (v2 API)
   Docs: https://docs.example.com/api/migration/v1-to-v2

[... command output ...]
```

**Month 7-9: Urgent warnings**

```bash
$ my-cli users list

🚨 URGENT DEPRECATION
   This endpoint will be removed in 28 days!
   You MUST migrate to v2 API before June 30, 2026

   Migration command:
     my-cli users list  # Already uses v2!

   Need help? https://docs.example.com/api/migration/v1-to-v2

[... command output ...]
```

**Month 10+: Critical (requires confirmation)**

```bash
$ my-cli users list

🔴 CRITICAL: Endpoint removal imminent (5 days remaining)
   This command will stop working on June 30, 2026

   To proceed, use: my-cli users list --force
   To migrate now: my-cli users list  # Uses v2

Error: Command blocked due to imminent deprecation
```

**After sunset: Removed**

```bash
$ my-cli users list

❌ This endpoint has been removed (as of June 30, 2026)

   Replacement:
     The v2 API is already the default!
     Just run: my-cli users list

   For help: https://docs.example.com/api/migration/v1-to-v2

Error: Operation no longer available
```

---

## Summary

### Key Principles

1. **Be Explicit** - Clearly communicate what's deprecated and why
2. **Be Gradual** - Give adequate time and escalating warnings
3. **Be Helpful** - Provide migration paths and automation
4. **Be Respectful** - Don't break existing scripts suddenly
5. **Be Trackable** - Allow users to audit and fix deprecations

### CliForge's Advantage

Unlike static CLI generators, CliForge can:
- ✅ Detect deprecations dynamically from OpenAPI spec
- ✅ Update warnings without binary recompilation
- ✅ Provide real-time sunset countdowns
- ✅ Auto-map deprecated flags to new ones
- ✅ Scan user scripts for deprecated usage
- ✅ Show migration suggestions inline

This makes deprecation management **seamless** for both API owners and CLI users.

---

**Version**: 0.7.0
**Last Updated**: 2025-01-11
**Project**: CliForge - Forge CLIs from APIs
