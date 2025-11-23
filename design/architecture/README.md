# Architecture Documentation

This directory contains detailed architecture and design documents for CliForge.

## Documents

### Core Design Documents

#### Built-in Commands System

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

#### Secrets Handling

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

#### Deprecation Strategy

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

#### Configuration Override Matrix

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

### Advanced Features (Post-Gap Analysis)

#### Plugin Architecture

**File**: `plugin-architecture.md` (v1.0.0)

Plugin system for extending CliForge beyond pure HTTP APIs:

- **Plugin Types**: Built-in, external binary, WebAssembly
- **Security Model**: Permissions system, sandboxing, user approval
- **Use Cases**: AWS CLI integration, file operations, custom validators
- **Integration**: Seamless workflow integration via `x-cli-plugin`

**Key Concepts**:
- Optional (simple CLIs don't need plugins)
- Capability-based permissions
- Plugin manifest and lifecycle
- Built-in plugins: exec, file-ops, validators, transformers

#### Workflow Orchestration

**File**: `workflow-orchestration.md` (v1.0.0)

Multi-step workflow execution engine:

- **Step Types**: api-call, plugin, conditional, loop, wait, parallel
- **Dependencies**: Explicit and implicit dependency resolution
- **Control Flow**: Conditionals, retries, rollback on failure
- **State Management**: Execution state persistence and resume capability

**Key Concepts**:
- Declarative YAML workflows in `x-cli-workflow`
- DAG-based execution with parallelization
- Expression language (expr) for conditions and transformations
- Rollback support for failed operations

#### File Operations

**File**: `file-operations.md` (v1.0.0)

File input handling framework:

- **Supported Formats**: PEM, htpasswd, JSON, YAML, plain text
- **Operations**: Read, parse, validate, transform
- **Integration**: `x-cli-file-input` extension, file-ops plugin
- **Security**: Access control, size limits, sensitive data masking

**Key Concepts**:
- File flags (`--from-file`, `--ca-file`)
- Format detection and validation
- Transform to API-compatible formats
- Security-first file access

#### Progress & Streaming

**File**: `progress-and-streaming.md` (v1.0.0)

Real-time feedback and streaming:

- **Progress Indicators**: Spinners, progress bars, multi-step display
- **Streaming Protocols**: Server-Sent Events (SSE), WebSocket, polling fallback
- **Watch Mode**: `--watch` flag for real-time updates
- **UX Library**: pterm for rich terminal UI

**Key Concepts**:
- Visual progress for long operations
- Log streaming from APIs
- Graceful degradation (streaming → polling)
- Step-by-step workflow progress

#### State Management & Context

**File**: `state-management.md` (v1.0.0)

State persistence and context system:

- **Context System**: Current cluster/resource selection (kubectl-style)
- **Smart Defaults**: Recent values, usage-based ordering
- **Command History**: Searchable history with replay
- **Workflow State**: Resume failed workflows

**Key Concepts**:
- XDG-compliant state storage
- Named contexts for multiple environments
- Workflow state persistence
- Recent items for autocomplete

#### Gap Analysis

**File**: `gap-analysis-rosa-requirements.md` (v1.0.0)

Comprehensive analysis of ROSA CLI requirements vs CliForge design:

- **Critical Gaps**: 6 architectural blockers identified
- **Important Gaps**: 6 feature parity issues
- **Well-Covered**: 8 existing features
- **Recommendations**: Phased implementation plan (v0.8-v1.0)

**Key Insights**:
- Plugin architecture essential for AWS integration
- Workflow orchestration needs completion
- Current design solid for auth, output, deprecation
- 40% of requirements well-covered, 60% needs work

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
