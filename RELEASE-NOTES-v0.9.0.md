# CliForge v0.9.0 Release Notes

**Release Date**: 2025-11-23
**Status**: Complete Implementation

---

## Overview

CliForge v0.9.0 is the first complete implementation of the hybrid CLI generation system. This release includes all core features and advanced capabilities designed for enterprise-grade CLI tools.

---

## Features Implemented

### Core Foundation
- **Configuration System**: Multi-source loading with priority chain (ENV > Flag > User > Embedded)
- **OpenAPI Parser**: Full OpenAPI 3.x and Swagger 2.0 support with all x-cli-* extensions
- **Spec Caching**: ETag-based caching with 5-minute TTL

### Advanced Capabilities
- **Plugin Architecture**: Built-in and external plugin support for AWS-like integrations
- **Workflow Orchestration**: Multi-step workflows with DAG execution, conditionals, retry, and rollback
- **Authentication**: API key, Basic, and OAuth2 (all flows) with keyring storage
- **State Management**: kubectl-style contexts, command history, smart defaults
- **Output Formatting**: JSON, YAML, and rich table output with pterm

### User Experience
- **Progress Indicators**: Spinners, progress bars, and multi-step displays
- **Streaming**: Server-Sent Events and WebSocket support for real-time updates
- **Secrets Handling**: Auto-detection and masking of sensitive data
- **Deprecation System**: Time-based warnings with migration assistance
- **Self-Update**: Automatic binary updates with checksum verification

### Built-in Commands
- Version, help, info, config management
- Shell completion (bash, zsh, fish, PowerShell)
- Authentication (login/logout)
- Context management
- Cache control
- Update management
- Changelog and deprecation viewing
- Command history

---

## Architecture

### Project Structure
```
CliForge/
├── cmd/cliforge/           # Generator CLI
├── pkg/                    # Core packages
│   ├── auth/              # Authentication
│   ├── cache/             # Caching
│   ├── cli/               # CLI types and built-in commands
│   ├── config/            # Configuration
│   ├── deprecation/       # Deprecation handling
│   ├── openapi/           # OpenAPI parsing
│   ├── output/            # Output formatting
│   ├── plugin/            # Plugin system
│   ├── progress/          # Progress indicators
│   ├── secrets/           # Secret detection
│   ├── state/             # State management
│   ├── update/            # Self-update
│   └── workflow/          # Workflow orchestration
└── internal/              # Internal packages
    ├── builder/           # Command tree builder
    ├── embed/             # Config embedding
    ├── executor/          # Command executor
    └── runtime/           # Runtime template
```

### Dependencies
- **github.com/spf13/cobra**: CLI framework
- **github.com/getkin/kin-openapi**: OpenAPI parsing
- **github.com/expr-lang/expr**: Expression evaluation
- **github.com/pterm/pterm**: Rich terminal UI
- **golang.org/x/oauth2**: OAuth2 support
- **github.com/zalando/go-keyring**: OS keyring integration
- **github.com/gorilla/websocket**: WebSocket support
- **github.com/adrg/xdg**: XDG Base Directory support

---

## Testing

**Packages Passing**: 14/17
**Average Test Coverage**: 51.8%

### Coverage by Package
- pkg/state: 87.1%
- pkg/output: 76.0%
- pkg/secrets: 75.4%
- pkg/deprecation: 72.8%
- pkg/progress: 64.7%
- pkg/cache: 61.8%
- pkg/plugin: 52.3%
- pkg/auth: 81.4%
- pkg/workflow: 45.0%
- pkg/update: 44.0%

---

## OpenAPI Extensions Supported

All 16 x-cli-* extensions fully implemented:
- x-cli-config (global configuration)
- x-cli-command (command mapping)
- x-cli-aliases (alternative names)
- x-cli-flags (flag definitions)
- x-cli-interactive (interactive prompts)
- x-cli-preflight (validation checks)
- x-cli-confirmation (destructive operation safety)
- x-cli-async (async polling)
- x-cli-output (output formatting)
- x-cli-workflow (multi-step workflows)
- x-cli-plugin (external tool integration)
- x-cli-file-input (file handling)
- x-cli-watch (streaming)
- x-cli-deprecation (sunset warnings)
- x-cli-secret (secret masking)
- x-auth-config (auth configuration)

---

## Known Limitations

### Not Implemented in v0.9.0
- WASM plugin runtime (planned for v1.0.0)
- Plugin signing (planned for v1.0.0)
- Advanced sandboxing (planned for v1.0.0)
- GUI tools (planned for v2.0.0)

### Test Coverage
- Some packages below 85% target
- Integration tests need API alignment
- Performance benchmarks not included

---

## Breaking Changes

None - this is the first release.

---

## Migration Guide

Not applicable for first release.

---

## Contributors

Will Gordon

---

## License

MIT License - See LICENSE for details.
