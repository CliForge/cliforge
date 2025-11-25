# CliForge Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [Unreleased]

---

## [0.9.2] - 2025-11-25

### Added

**GitHub Actions CI/CD Pipeline:**
- Comprehensive CI workflow (tests, linting, builds, docs, coverage)
  - Multi-platform testing (Ubuntu, macOS, Windows)
  - Multi-version Go testing (1.21, 1.22, 1.23)
  - Codecov integration for coverage tracking
  - Artifact uploads for binaries and coverage reports
- Release automation workflow
  - Multi-platform binary builds
  - Checksum generation
  - GitHub Release creation with changelog extraction
  - GitHub Pages documentation deployment
- Pull request checks workflow
  - Conventional Commits validation
  - Large file detection
  - Secret scanning
  - Coverage reporting with PR comments
  - Documentation build verification

**Test Coverage Improvements:**
- pkg/config: 62.3% → **95.0%** (+32.7%)
  - copyConfig: 45.2% → 100.0%
  - MergeDefaults: 35.3% → 100.0%
  - applyDebugOverrides: 51.9% → 92.6%
  - applyUserPreferences: 58.5% → 97.6%
  - validateBehaviors: 38.5% → 96.2%
- Added 40 comprehensive test functions with 109 test cases
- All helper functions now at 83%+ coverage

**Documentation:**
- **Migration Guides** (`docs/migration-guides.md`, 3,736 lines)
  - Migrating from OpenAPI Generator
  - Migrating from Restish
  - Migrating from curl/shell scripts
  - Migrating from AWS CLI patterns
  - Migration decision matrix and success stories

- **Operations Guide** (`docs/operations-guide.md`, 2,639 lines)
  - Production deployment strategies
  - Monitoring and observability
  - Update management
  - Enterprise configuration
  - Troubleshooting in production
  - Security hardening
  - Disaster recovery

- **Security Guide** (`docs/security-guide.md`, 2,755 lines)
  - Security architecture and threat model
  - Credential management
  - Network security (TLS/SSL, mTLS)
  - Configuration security
  - Audit and compliance (SOC2, HIPAA, GDPR)
  - Security best practices for all stakeholders
  - Vulnerability management

- **Tutorial Series** (`docs/tutorials/`, 5,257 lines total)
  - REST API CLI tutorial (1,486 lines) - GitHub API example
  - Cloud management tutorial (1,959 lines) - Infrastructure automation
  - CI/CD integration tutorial (1,812 lines) - Pipeline integration
  - Tutorial index and quick reference (906 lines)

**Infrastructure:**
- mkdocs.yml updated with new documentation sections
- Automated testing enforcing 60%+ coverage requirement
- PR comment automation for coverage reporting

### Improved

**Documentation Quality:**
- Comprehensive migration paths for all major CLI tools
- Production-ready deployment and operations documentation
- Enterprise-grade security documentation
- Step-by-step tutorials for common use cases
- Total new documentation: 14,387 lines

**Developer Experience:**
- Automated CI/CD reducing manual testing effort
- Commit message validation enforcing Conventional Commits
- Automated release process with multi-platform binaries
- Coverage tracking and reporting on every PR

**Code Quality:**
- pkg/config now at enterprise-grade 95% coverage
- Comprehensive edge case testing
- Nil pointer safety verified
- Error handling thoroughly tested

### Project Statistics

**Test Coverage:**
- pkg/config: 95.0%
- pkg/openapi: 84.1%
- internal/builder: 85.0%
- internal/executor: 74.9%
- Average (core packages): ~85%

**Documentation:**
- Total lines added: 14,387
- Migration guides: 3,736 lines
- Operations guide: 2,639 lines
- Security guide: 2,755 lines
- Tutorial series: 5,257 lines

**CI/CD:**
- 3 comprehensive GitHub Actions workflows
- Multi-platform testing (3 OS × 3 Go versions = 9 matrix combinations)
- Automated releases with binary distribution
- Documentation deployment automation

---

## [0.9.1] - 2025-11-25

### Added

**Test Coverage Improvements:**
- Comprehensive test suite for pkg/config (31.2% → 62.3%, +31.1%)
- Comprehensive test suite for pkg/openapi (33.1% → 84.1%, +51.0%)
- Comprehensive test suite for internal/builder (48.7% → 85.0%, +36.3%)
- Comprehensive test suite for internal/executor (24.0% → 74.9%, +50.9%)
- End-to-end integration test suite (6 test functions)
  - Binary build verification
  - CLI command testing
  - Completion generation testing
  - Petstore example validation

**Documentation:**
- Quick Reference Guide (`docs/quick-reference.md`)
  - Command cheat sheet
  - Configuration patterns
  - OpenAPI extension quick lookup
  - Common workflows and troubleshooting
- mkdocs.yml navigation updated with quick reference

**Test Infrastructure:**
- Mock HTTP servers for realistic testing
- Table-driven tests for comprehensive coverage
- OpenAPI test fixtures (testdata/)
- Integration test framework

### Fixed

- Config loader NO_COLOR environment variable logic (inverted semantics)
- Test coverage for previously untested functions:
  - NewLoader, LoadConfig, loadEmbeddedConfig, loadUserConfig
  - SaveUserConfig, EnsureConfigDirs, applyEnvironmentOverrides
  - NewChangeDetector, DetectChanges, changelog formatting
  - NewValidator, Validate, validateUpdates, ValidateUserPreferences
  - buildPathGroupedCommands, determineCommandName
  - addRequestBodyFlags, addCustomFlags, ValidateRequiredFlags
  - executeHTTPOperation, buildRequest, handleAsyncOperation
  - NewRuntime, initializeManagers, buildCommandTree

### Test Coverage Summary

| Package | Before | After | Improvement |
|---------|--------|-------|-------------|
| pkg/config | 31.2% | 62.3% | +31.1% |
| pkg/openapi | 33.1% | 84.1% | +51.0% |
| internal/builder | 48.7% | 85.0% | +36.3% |
| internal/executor | 24.0% | 74.9% | +50.9% |

**Overall Project Coverage:** Significantly improved across all core packages

### Documentation Assessment

- Comprehensive documentation review completed
- Grade: B+ (85% complete)
- 16 documentation files reviewed
- Identified gaps and created quick reference guide
- Recommended additional guides for future releases

---

## [0.9.0] - 2025-11-23

### Added

**Core Foundation:**
- Configuration system with multi-source loading (ENV > Flag > User > Embedded)
- OpenAPI 3.x and Swagger 2.0 parser with all 16 x-cli-* extensions
- Spec caching with ETag support and 5-minute TTL

**Advanced Features:**
- Plugin architecture (built-in, binary, WASM-ready) for external tool integration
- Workflow orchestration engine with DAG execution, conditionals, retry, and rollback
- Authentication system (API key, Basic, OAuth2 all flows) with keyring storage
- State management with kubectl-style contexts, command history, and smart defaults
- Output formatting (JSON, YAML, rich tables with pterm)
- Progress indicators (spinners, bars, multi-step displays)
- Streaming support (SSE, WebSocket, polling fallback)
- Secrets detection and masking with multiple strategies
- Deprecation warnings with time-based escalation
- Self-update mechanism with checksum verification

**Built-in Commands:**
- version, help, info, config (get/set/unset/edit/path)
- completion (bash, zsh, fish, PowerShell)
- auth (login/logout/status/refresh)
- context (create/use/list/set/get/delete)
- history, cache, update, changelog, deprecations

**Developer Tools:**
- Generator CLI (cliforge init/build/validate)
- Complete runtime template system
- Command tree builder from OpenAPI specs
- Petstore example demonstrating all features

**Documentation:**
- Complete architecture designs for all ROSA CLI requirements
- Gap analysis with implementation roadmap
- Plugin, workflow, file ops, progress, and state management design docs

### Test Coverage
- 17/17 packages passing all tests
- Average coverage: 51.8%
- High coverage: state (87%), auth (81%), output (76%), secrets (75%)

---

### ⚒️ Rebranding - 2025-01-11

#### Changed
- **Project name**: Alpha-Omega → **CliForge**
- **Tagline**: "Forge CLIs from APIs"
- **Generator binary**: `alpha-omega-gen` → `cliforge`
- **Module path**: `github.com/alpha-omega/poc` → `github.com/cliforge/poc`

#### Added
- Complete branding guide (`BRANDING.md`)
  - Logo concepts and ASCII art
  - Color palette (forge fire theme)
  - Typography guidelines
  - Voice & tone guide
  - Social media assets
- Professional README with quick start
- Comprehensive documentation updates
- CliForge color scheme to POC demo config

#### Updated
- All documentation files with new branding
- Technical specification (v1.2.0)
- Configuration DSL (v1.2.0)
- Templating comparison document
- Patentability analysis
- POC example configurations
- Go module imports

---

### 📋 Deprecation Strategy - 2025-01-11

#### Added
- Comprehensive deprecation handling strategy (`deprecation-strategy.md`)
  - **Two-tier deprecation system**:
    - API deprecations (from OpenAPI spec changes)
    - CLI deprecations (tool behavior changes)
  - **Warning severity levels**: info → warning → urgent → critical → removed
  - **Time-based escalation**: Automatic severity based on days until sunset
  - **OpenAPI extensions**:
    - Standard `deprecated: true` field support
    - Enhanced `x-cli-deprecation` extension with sunset dates, replacement info
    - HTTP `Sunset` header (RFC 8594) detection
  - **User controls**:
    - `--no-deprecation-warnings` flag
    - Configuration options for suppression and thresholds
    - Per-operation suppression
  - **Automation features**:
    - `deprecations` command to view all active deprecations
    - `deprecations scan` to find deprecated usage in scripts
    - Auto-fix capabilities for deprecated commands/flags
    - Automatic flag mapping (old → new)
  - **Migration assistance**:
    - Inline migration suggestions
    - Command equivalence display
    - Documentation links
  - **Implementation details**:
    - Go code examples for detection logic
    - Severity calculation algorithms
    - Warning display formatting
    - Best practices for API owners and CLI users

#### Design Decisions
- **Dynamic advantage**: CliForge's runtime spec loading allows deprecation warnings to update without binary recompilation
- **Gradual rollout**: Phased approach from announcement → warning → opt-in → removal
- **User-friendly**: Clear migration paths, automation support, suppressible warnings

---

### 📐 Swagger/OpenAPI Compatibility - 2025-01-11

#### Clarified
- **Specification support**: CliForge supports both Swagger 2.0 and OpenAPI 3.x
  - Swagger 2.0: Fully supported with auto-conversion to OpenAPI 3.0
  - OpenAPI 3.0: Primary target format, fully supported
  - OpenAPI 3.1: In progress via kin-openapi library updates
- **Terminology**: Documented difference between "Swagger" (legacy spec + tooling) and "OpenAPI" (modern spec)
- **Auto-conversion**: Leverages `openapi2conv` from kin-openapi to normalize Swagger 2.0 to OpenAPI 3.0 internally
- **Maximum compatibility**: Enterprise legacy APIs (Swagger 2.0) and modern APIs (OpenAPI 3.x) both supported

#### Updated
- Technical specification (v1.2.0)
  - Added OpenAPI & Swagger Compatibility section (Appendix B)
  - Documented breaking changes between Swagger 2.0 and OpenAPI 3.0
  - Explained auto-conversion mechanism
  - Added compatibility matrix

---

### 🎛️ Built-in Commands & Global Flags - 2025-01-11 (UPDATED)

#### Added
- **Comprehensive built-in commands system** (`builtin-commands-design.md`)
  - **Standard built-in commands**: version, help, info, config, completion, update, changelog, deprecations, cache, auth
  - **Three CLI styles supported**:
    - Subcommand style (`mycli version`)
    - Flag style (`mycli --version`)
    - Hybrid style (both work)
  - **Configurable command behavior**:
    - Enable/disable individual commands
    - Customize command names (e.g., `config` → `configure`)
    - Control subcommands and features
    - Choose between subcommand vs flag vs hybrid

- **Standard global flags**:
  - `--config`, `-c` - Config file path
  - `--profile` - Profile section selector (like AWS CLI profiles)
  - `--region`, `-r` - Region/datacenter (common cloud pattern)
  - `--output`, `-o` - Output format (json|yaml|table|csv)
  - `--verbose`, `-v` - Verbose output (repeatable: -v, -vv, -vvv)
  - `--quiet`, `-q` - Quiet mode
  - `--debug` - Debug logging
  - `--no-color` - Disable colors (respects NO_COLOR env var)
  - `--timeout` - Request timeout
  - `--retry` - Retry attempts
  - `--no-cache` - Disable caching
  - Custom flags support (for API-specific needs like --api-key, --token, --org)

- **Flag configuration features**:
  - Customizable flag names and short forms
  - Environment variable mapping
  - Default values and allowed values
  - Conflict detection (e.g., --verbose vs --quiet)
  - Sensitive flag masking (API keys, tokens)
  - Repeatable flags (verbosity levels)
  - Flag precedence: CLI flag > env var > config file > default

- **Company customization examples**:
  - Docker-like hybrid CLI
  - Traditional Unix flag-only CLI
  - Minimal subcommand-only CLI
  - Company-standard CLI matching existing tools

#### Changed
- **Version display**: Shows BOTH binary version AND API version (like `oc version`)
  - Binary version changes rarely (security, CLI features)
  - API version changes frequently (endpoints, parameters)
  - Users see: `Client: v1.2.3  Server: v2.1.0`

- **Changelog separation**: Binary changes vs API changes tracked separately
  - `mycli changelog --binary-only` - CLI feature changes
  - `mycli changelog --api-only` - API endpoint changes
  - Different update frequencies and audiences

- **Deprecation tracking**: Binary deprecations vs API deprecations separated
  - CLI deprecations: flag/command changes (rare)
  - API deprecations: endpoint/parameter changes (frequent)

- **Removed from default global flags** (security/design concerns):
  - ❌ `--api-key` - Auth varies (OAuth2, API key, Basic, etc.) - use custom flags
  - ❌ `--base-url` - Security risk allowing override - breaks expectations

- **Added to default global flags**:
  - ✅ `--region` - Common cloud/datacenter pattern

- **Profile support**: `--profile` maps to YAML config sections (like AWS CLI)
  - Config file has `default`, `production`, `dev` sections
  - Each profile can override settings (including `api.base_url` for different environments)
  - Precedence: CLI flag > env var > profile config > default

- **Added `--yes` / `-y` flag** for non-interactive mode
  - Skip confirmation prompts
  - Critical for CI/CD pipelines
  - Can be disabled per-CLI if not needed

- **Secrets & sensitive data handling** (`secrets-handling-design.md`)
  - **Multi-layer detection**:
    - `x-cli-secret` OpenAPI extension (explicit marking)
    - Field name patterns (`*password*`, `*token*`, `*key`)
    - Value patterns (regex: AWS keys, JWTs, API keys)
  - **Masking strategies**:
    - Partial: `sk_live_abc***` (default)
    - Full: `***`
    - Hash: `sha256:a3f2...` (for audit logs)
  - **Applies to**: stdout, stderr, logs, debug output
  - **User controls**: Can customize patterns, masking style, or disable (with warnings)

- **Framework decision documented** (`cobra-framework-decision.md`)
  - Chose `spf13/cobra` over alternatives (urfave/cli, kong, mitchellh/cli)
  - Justification: industry-proven, perfect for API-driven CLIs
  - Used by kubectl, docker, gh (GitHub CLI)

- **Clarified base_url configuration**:
  - `api.base_url` is in EMBEDDED config (branding DSL)
  - NOT available as global flag (security)
  - Users CAN override via profiles (staging, dev, local environments)
  - Company controls valid base URLs

#### Updated
- Configuration DSL (v1.3.0)
  - Added `behaviors.builtin_commands` section
  - Added `behaviors.global_flags` section with comprehensive flag configuration
  - Removed `--api-key` and `--base-url` from default flags (security/flexibility)
  - Added `--region` to default flags
  - Updated `changelog` and `deprecations` config for binary/API separation
  - Removed duplicate simple `commands.global_flags` (now in behaviors)
  - Added examples for different CLI styles
- `builtin-commands-design.md`
  - Updated version display to show client + server versions
  - Updated changelog to separate binary vs API changes
  - Updated deprecations to track binary vs API deprecations
  - Added profile support documentation
  - Removed --api-key and --base-url from defaults
  - Added --region as standard flag

#### Design Decisions
- **Flexibility first**: Support all major CLI conventions (POSIX, GNU, modern)
- **Convention over configuration**: Sensible defaults (hybrid style, standard flags)
- **Company fit**: Allow matching existing CLI patterns without fighting the framework
- **Discoverability**: --help and completion make features findable
- **Cobra integration**: Maps cleanly to spf13/cobra persistent flags and subcommands

---

## [0.8.0] - 2025-01-11

### 🔒 Configuration Override Architecture Finalized - 2025-01-11

#### Added
- **New `defaults` section** in embedded configuration (`configuration-dsl.md`)
  - 13 user-overridable settings (http.timeout, caching.enabled, pagination.limit, output.*, deprecations.*, retry.max_attempts)
  - Clear separation from locked `behaviors` section
  - Self-documenting structure

- **New `preferences` section** in user configuration (`~/.config/<cli-name>/config.yaml`)
  - Overrides `defaults` from embedded config
  - Always active (production and debug builds)
  - Contains runtime preferences and enterprise settings (proxy, CA bundle, telemetry opt-in)

- **New `debug_override` section** in user configuration
  - **ONLY works when `metadata.debug: true`** in embedded config
  - Allows overriding ANY embedded config setting for development/testing
  - Shows security warning on EVERY command when active
  - Production builds ignore this section entirely

- **New `metadata.debug` boolean field** in embedded configuration
  - `false` (default): Production mode - strict override rules, `debug_override` ignored
  - `true`: Debug mode - `debug_override` section active, security warnings displayed
  - Enables separate binary builds: `mycli` (production) vs `mycli-dev` (debug)

- **New `api.telemetry_url` field** (locked, embedded-only)
  - Defines WHERE telemetry is sent
  - User controls WHETHER to send via `preferences.telemetry.enabled` (defaults to false)

- **New `behaviors.pagination.delay` field** (locked, default: 100ms)
  - Inter-page delay for auto-pagination to prevent API spam
  - Replaces removed `rate_limiting` section

- **Comprehensive override matrix documentation** (`configuration-override-matrix.md`)
  - 71 locked settings (metadata, branding, api, behaviors.*)
  - 13 overridable settings (defaults → preferences)
  - Complete debug mode documentation with examples
  - Security best practices for CLI developers and users

#### Changed
- **BREAKING**: Completely redesigned configuration override system
  - Old approach: Users could override many behavior settings directly
  - New approach: Clear separation between `defaults` (overridable) and `behaviors` (locked)

- **Moved `timeout`** from `api.timeout` to `defaults.http.timeout`
  - Rationale: `timeout` is a runtime preference, not an API definition
  - `api.*` section now 100% locked for security

- **Locked entire `api` section** (100% embedded-only)
  - Includes: openapi_url, base_url, environments, default_headers, user_agent, telemetry_url
  - Prevents users from pointing CLI to wrong APIs or redirecting telemetry
  - Critical security boundary

- **Moved telemetry configuration**:
  - Endpoint: `features.telemetry.endpoint` → `api.telemetry_url` (locked)
  - Opt-in: `features.telemetry.enabled` → `preferences.telemetry.enabled` (user-only, defaults to false)

- **Updated all examples** to use new `defaults` section
  - Example 1: Minimal Configuration
  - Example 2: Full-Featured SaaS Product
  - Example 3: Internal Enterprise Tool

- **Updated user configuration file documentation**
  - Clear `preferences` vs `debug_override` structure
  - Security warnings and best practices
  - Debug mode examples with warning display
  - Configuration priority with new architecture

- **Updated profile configuration examples** (`builtin-commands-design.md`)
  - All profiles now use `preferences` section
  - Added debug mode examples

- **Configuration priority updated**:
  - ENV > Flag > User Config (`preferences`) > Embedded (`defaults`) > Built-in Default

#### Removed
- **`behaviors.rate_limiting` section** removed entirely
  - Redundant with retry logic and server-side 429 handling
  - Replaced with `behaviors.pagination.delay` for auto-pagination throttling

- **auto_paginate from configuration** (now flag/ENV only)
  - Cannot be enabled by default in config (prevents API abuse)
  - Must be explicit: `--auto-page` or `MYCLI_AUTO_PAGE=1`

- **Sections removed entirely**: `commands`, `hooks`, `plugins`
  - Users should use shell aliases instead of embedded command customization
  - Users should wrap CLI with shell scripts for lifecycle events
  - Plugins too complex for v1.0, security concerns

- **Features removed**: `telemetry`, `shell_completion`, `validate_requests`, `validate_responses`, `offline_mode`
  - `telemetry`: Split into `api.telemetry_url` (locked) and `preferences.telemetry.enabled` (user-only)
  - `shell_completion`: Always enabled, not configurable
  - `validate_requests`/`validate_responses`: Always validate requests, never responses (performance)
  - `offline_mode`: Redundant with caching behavior

#### Security
- **Debug mode security**:
  - Warning displayed on EVERY command when `metadata.debug: true`
  - Clear distinction between production and debug binaries
  - `debug_override` section ignored in production builds with warning

- **Locked settings count**: 71 settings (metadata, branding, api, behaviors.*)
- **Overridable settings count**: 13 settings (defaults → preferences)

#### Updated Documentation
- `configuration-dsl.md` → v0.8.0
  - Complete schema rewrite with `defaults`, `preferences`, `debug_override`
  - Updated all examples
  - New User Configuration File section
  - Updated changelog

- `configuration-override-matrix.md`
  - Finalized override matrix (71 locked, 13 overridable)
  - Complete debug mode documentation
  - Security best practices

- `builtin-commands-design.md`
  - Updated profile configuration examples
  - Updated configuration priority section
  - Added `preferences` section to all config examples

---

## [0.7.0] - 2025-01-11

### See: Built-in Commands & Global Flags section above

---

## [0.6.0] - 2025-01-11

### See: Swagger/OpenAPI Compatibility section above

---

## [0.5.0] - 2025-01-11

### See: Deprecation Strategy section above

---

## [0.4.0] - 2025-01-11

### See: Rebranding section above

---

## [0.3.0] - 2025-11-10

### See: Security & XDG Compliance section above

---

## [0.2.0] - 2025-11-10

### Added - expr Expression Language

#### Changed
- **`x-cli-workflow` redesign**: Now uses expr expression language
  - Removed redundant `output` field from workflow steps
  - Steps referenced directly by `id` (e.g., `get-objects.body.field`)
  - Added `foreach` iteration support at YAML level
  - Documented expr syntax for conditions and transformations

#### Added
- Comprehensive expr function reference
  - Array functions: `filter()`, `map()`, `all()`, `any()`, `len()`
  - String functions: `trim()`, `upper()`, `lower()`, `split()`, `join()`
  - Type conversion: `int()`, `float()`, `string()`
  - Math functions: `abs()`, `ceil()`, `floor()`, `round()`
  - Optional chaining: `user.email?.verified`
  - Nil coalescing: `user.email ?? "none"`

- Future enhancement: `x-cli-functions` (Phase 2)
  - Dynamic function registration from OpenAPI specs
  - API owners can define custom functions
  - Runtime compilation with expr

- Deep research document: `templating-language-comparison.md`
  - Compared expr, jsonnet, gomplate
  - Scored on 13 criteria (weighted)
  - **expr won** (4.68/5) for safety, performance, familiar syntax

#### Updated
- Technology stack: Added expr-lang/expr
- All workflow examples to use expr syntax
- Technical specification to v1.1.0

---

### Added - Security & XDG Compliance

#### Changed - User Configuration Security
- **BREAKING SECURITY**: Removed `auto_install` from embedded config schema
  - Can ONLY be set in user configuration file
  - Prevents developers from pushing automatic binary updates
  - Added security warnings when users enable it

- **Removed redundant `channel` field**
  - Channel now implicit in `update_url`
  - Example: `https://releases.acme.com/stable/cli` vs `/beta/cli`
  - Simpler, more flexible

#### Added - User Configuration File
- Comprehensive "User Configuration File" section (200+ lines)
  - XDG Base Directory Specification compliance
  - File locations: `$XDG_CONFIG_HOME`, `$XDG_CACHE_HOME`, etc.
  - Configuration priority clearly defined
  - User config commands: `config show`, `config edit`, `config set`
  - Security best practices
  - Example configurations

- XDG environment variables reference table
  - What goes where (config, cache, data, state)
  - Default locations for Linux/macOS/Windows
  - Go implementation examples

#### Updated
- Configuration DSL to v1.1.0
- All file location references to be XDG-compliant

---

## [0.1.0] - 2025-11-09

### Added - Initial Design

#### Core Architecture
- Hybrid CLI generation system
  - Static branded binary generation
  - Dynamic OpenAPI spec loading at runtime
  - Self-updating binary capability
  - Change notification system

#### Tri-Level Separation
- **Binary-level**: Branding, URLs, security patches
- **API-level**: Endpoints, operations, schemas
- **Behavioral**: Auth flows, rate limits, caching

#### Documentation
- Technical specification (`technical-specification.md`)
- Configuration DSL specification (`configuration-dsl.md`)
- Patentability analysis (`api-driven-cli-research.md`)
- Project structure and implementation plan

#### OpenAPI Extensions
- `x-cli-version` - API version tracking
- `x-cli-changelog` - User-facing change documentation
- `x-cli-aliases` - Command aliases
- `x-cli-examples` - Usage examples
- `x-cli-hidden` - Hide operations from CLI
- `x-cli-auth` - Per-operation auth requirements
- `x-cli-workflow` - Multi-step API orchestration

#### Technology Stack
- **Language**: Go 1.21+
- **CLI Framework**: spf13/cobra
- **Config**: spf13/viper
- **OpenAPI**: getkin/kin-openapi
- **HTTP**: net/http (stdlib)

#### Security Features
- Ed25519 binary signature verification
- TLS certificate pinning for updates
- OS-specific secure credential storage
- Audit logging

---

## Development Milestones

### ✅ Completed
- [x] Initial architecture design
- [x] Technical specification
- [x] Configuration DSL design
- [x] OpenAPI extensions design
- [x] Templating language research (expr selected)
- [x] Security boundary definition (user config)
- [x] XDG compliance
- [x] Project rebranding to CliForge
- [x] Comprehensive documentation
- [x] Deprecation strategy design

### 🚧 In Progress
- [ ] Proof-of-concept implementation

### 📋 Planned
- [ ] Core generator CLI (`cliforge`)
- [ ] Runtime engine with expr integration
- [ ] Workflow orchestration
- [ ] Self-update mechanism
- [ ] Change detection system
- [ ] Production release

---

## Links

- **Repository**: https://github.com/cliforge/cliforge (coming soon)
- **Documentation**: https://docs.cliforge.com (coming soon)
- **Website**: https://cliforge.com

---

*⚒️ Forged with ❤️ by the CliForge team*
