# Architecture Documentation

This directory contains detailed architecture and design documents for CliForge.

## Documents

### Built-in Commands System

**File**: `builtin-commands-design.md` (v0.7.0)

Comprehensive design for standard CLI commands and global flags:

- **CLI Styles**: Subcommand, flag, and hybrid styles
- **Standard Commands**: version, help, info, config, completion, update, changelog, deprecations, cache, auth
- **Global Flags**: --config, --profile, --region, --output, --verbose, etc.
- **Customization**: Configure command names, flags, behavior per CLI
- **Examples**: Docker-like, Unix-like, minimal, company-standard CLIs

**Key Concepts**:
- Version display shows BOTH binary and API versions
- Changelog separates binary changes from API changes
- Profile support (like AWS CLI profiles)
- Enterprise features: proxy support, CA certificate override

### Secrets Handling

**File**: `secrets-handling-design.md` (v0.7.0)

Sensitive data detection and masking system:

- **Multi-layer Detection**: `x-cli-secret` extension, field name patterns, value patterns
- **Masking Strategies**: Partial (`sk_live_abc***`), full (`***`), hash (`sha256:a3f2...`)
- **Applies To**: stdout, stderr, logs, debug output
- **User Controls**: Customize patterns, masking style, or disable (with warnings)

**Key Concepts**:
- Security-first design
- Pattern-based auto-detection
- Explicit marking via OpenAPI extension
- Configurable per-CLI

### Deprecation Strategy

**File**: `deprecation-strategy.md` (v0.7.0)

Comprehensive deprecation handling for APIs and CLIs:

- **Two-tier System**: API deprecations (from OpenAPI spec) and CLI deprecations (tool behavior)
- **Warning Levels**: info → warning → urgent → critical → removed
- **Time-based Escalation**: Automatic severity based on days until sunset
- **User Controls**: Suppression options, per-operation controls

**Key Concepts**:
- OpenAPI `deprecated: true` field support
- `x-cli-deprecation` extension with sunset dates, replacement info
- HTTP `Sunset` header (RFC 8594) detection
- Migration assistance and auto-fix capabilities

### Configuration Override Matrix

**File**: `configuration-override-matrix.md` (v0.8.0)

Detailed rules for configuration overrides:

- **71 Locked Settings**: Cannot be overridden by users (security boundary)
- **13 Overridable Settings**: User preferences in `preferences` section
- **Debug Mode**: Complete override capability when `metadata.debug: true`
- **Priority Chain**: ENV > Flag > User Config > Embedded > Default

**Key Concepts**:
- Security boundaries (api.*, behaviors.*, updates.*)
- User overridable defaults (output, caching, pagination, etc.)
- Debug mode for development/testing
- Clear separation: `defaults` (overridable) vs `behaviors` (locked)

## Document Relationships

```
configuration-override-matrix.md ─── Defines rules for ──→ User Configuration
        │                                                           │
        └── Referenced by ────────────────────────┐                │
                                                   │                │
builtin-commands-design.md ────── Uses ───────────┴→ Global Flags
        │
        └── Includes ──→ Profile Support ──→ Different environments
                              │
                              └── References ──→ configuration-dsl.md

secrets-handling-design.md ─── Integrated into ──→ Output formatting
                                                    Debug output
                                                    Audit logs

deprecation-strategy.md ────── Separates ──→ Binary deprecations
                                              API deprecations
                                              Migration paths
```

## Related Documentation

- **External Docs**: `../../docs/` - User-facing documentation
- **ADRs**: `../decisions/` - Architectural decisions
- **Research**: `../../research/` - Research and analysis

---

*⚒️ Forged with ❤️ by the CliForge team*
