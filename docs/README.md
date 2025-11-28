# CliForge Documentation

This directory contains **external user-facing documentation** for CliForge.

## Overview

CliForge is a dynamic CLI generator that creates branded command-line tools from OpenAPI specifications. It enables API owners to provide native CLI experiences without writing and maintaining CLI code.

## Core Documentation

### Getting Started

- **README.md** (root) - Project overview and quick start
- **technical-specification.md** - Complete technical specification (v1.2.0)
- **configuration-dsl.md** - Configuration DSL reference (v0.8.0)

### User Guides

#### For CLI Developers (API Owners)

1. **Configuration DSL** (`configuration-dsl.md`)
   - How to configure your branded CLI
   - Branding, authentication, features
   - User configuration file structure
   - Debug mode and overrides

2. **Built-in Commands** ([builtin-commands-design.md](https://github.com/wgordon17/cliforge/blob/main/design/architecture/builtin-commands-design.md))
   - Standard commands: version, help, info, config
   - Global flags configuration
   - CLI styles: subcommand, flag, hybrid

3. **Deprecation Strategy** ([deprecation-strategy.md](https://github.com/wgordon17/cliforge/blob/main/design/architecture/deprecation-strategy.md))
   - Handling API deprecations
   - CLI deprecations
   - Migration assistance

4. **Secrets Handling** ([secrets-handling-design.md](https://github.com/wgordon17/cliforge/blob/main/design/architecture/secrets-handling-design.md))
   - Sensitive data detection and masking
   - Security best practices

5. **Security Guide** (`security-guide.md`)
   - Comprehensive security architecture
   - Credential management and authentication
   - Network security and TLS configuration
   - Audit and compliance (SOC2, HIPAA, GDPR)
   - Vulnerability management

#### For CLI Users

- **User Configuration** - See "User Configuration File" section in `configuration-dsl.md`
- **XDG Compliance** - File locations, caching, configuration
- **Security Best Practices** - See "Security Best Practices > For CLI Users" in `security-guide.md`

## Advanced Topics

### Technical Specification

**File**: `technical-specification.md` (v1.2.0)

Comprehensive technical design covering:
- Hybrid architecture (static binary + dynamic spec loading)
- Tri-level separation (binary, API, behavioral)
- Self-updating mechanism
- Change detection and notifications
- OpenAPI extensions (`x-cli-*`)
- Security model

### Configuration Override Architecture

**File**: `configuration-dsl.md` (v0.8.0)

Detailed documentation on:
- **Embedded config** - Baked into binary (`defaults` section)
- **User config** - User preferences (`preferences` section)
- **Debug mode** - Development overrides (`debug_override` section)
- **Locked vs overridable settings** - 71 locked, 13 overridable
- **Configuration priority** - ENV > Flag > User Config > Embedded > Default

## Related Documentation

### Internal Documentation

For contributors and maintainers, see:
- [Design documents](https://github.com/wgordon17/cliforge/tree/main/design) - Design and architecture documents
- Research and analysis (internal)
- Brand guidelines (internal)

### Decision Records

Architectural decisions are documented as ADRs in [design/decisions/](https://github.com/wgordon17/cliforge/tree/main/design/decisions):
- ADR-000: Use Architecture Decision Records
- ADR-100: Use Cobra for CLI Framework
- ADR-101: Use expr for Templating Language

## Quick Links

- **Project**: https://github.com/cliforge/cliforge
- **Website**: https://cliforge.com
- **Issues**: https://github.com/cliforge/cliforge/issues

## Version History

- **v0.8.0** (2025-01-11) - Configuration override architecture finalized
- **v0.7.0** (2025-01-11) - Built-in commands, global flags, secrets handling
- **v0.6.0** (2025-01-11) - OpenAPI/Swagger compatibility clarified
- **v0.5.0** (2025-01-11) - Deprecation strategy
- **v0.4.0** (2025-01-11) - Rebranded to CliForge
- **v0.1.0** (2025-11-09) - Initial design

---

*⚒️ Forged with ❤️ by the CliForge team*
