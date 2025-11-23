# CliForge

**Dynamic CLI generation from OpenAPI specifications.**

CliForge combines static branded binaries with dynamic spec loading to create professional, self-updating command-line tools from OpenAPI specifications.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go 1.21+](https://img.shields.io/badge/go-1.21+-blue.svg)](https://go.dev/dl/)
[![Status](https://img.shields.io/badge/status-design%20phase-yellow.svg)]()

---

## What is CliForge?

CliForge generates branded CLI tools that load OpenAPI specifications at runtime. This hybrid approach delivers professional, distributable binaries while maintaining the flexibility of dynamic spec loading.

**Key Features:**
- **Hybrid architecture** - Static branded binary + dynamic spec loading
- **Self-updating** - Security patches and config changes delivered seamlessly
- **Zero code generation** - No generated files in your repository
- **OpenAPI-driven** - Your spec is the single source of truth
- **Configuration DSL** - YAML-based configuration for branding and behavior
- **XDG-compliant** - Follows standard directory conventions

---

## Architecture

CliForge uses a tri-level separation model:

**Binary Level:**
- Embedded configuration and branding
- Self-update mechanism
- Runtime engine

**API Level:**
- OpenAPI spec loaded at runtime (cached 5min)
- Paths → Commands
- Operations → Subcommands
- Parameters → Flags

**Behavioral Level:**
- Authentication strategies
- Output formatting
- Rate limiting
- Workflow orchestration

**Update Flow:**
- **Binary changes** (branding, security) → Self-update
- **API changes** (new endpoints) → Spec reload
- **Config changes** (auth, rate limits) → Config update

---

## Configuration

Example `cli-config.yaml`:

```yaml
metadata:
  name: my-api-cli
  version: 1.0.0
  description: My API Command Line Interface

branding:
  colors:
    primary: "#FF6B35"
  ascii_art: |
    ╔════════════════════════╗
    ║   My Amazing API CLI   ║
    ╚════════════════════════╝

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com

updates:
  enabled: true
  update_url: https://releases.example.com/my-api-cli

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: MY_API_KEY
  output:
    default_format: json
    pretty_print: true
```

See [Configuration DSL Reference](docs/configuration-dsl.md) for complete documentation.

---

## Advanced Features

### Workflow Orchestration

Define multi-step API workflows with `x-cli-workflow`:

```yaml
paths:
  /deploy:
    post:
      operationId: deployApp
      x-cli-workflow:
        steps:
          - id: check-readiness
            request:
              method: GET
              url: "{base_url}/readiness"

          - id: create-deployment
            request:
              method: POST
              url: "{base_url}/deployments"
              body:
                app: "{args.app_id}"
            condition: "check-readiness.body.ready == true"

        output:
          transform: |
            {
              "deployment_id": create-deployment.body.id,
              "status": create-deployment.body.status
            }
```

### Change Detection

Automatically notify users of API changes:

```yaml
info:
  x-cli-changelog:
    - date: "2025-01-11"
      version: "2.1.0"
      changes:
        - type: added
          severity: safe
          description: "New analytics endpoint"
        - type: deprecated
          severity: dangerous
          description: "GET /v1/users deprecated"
          migration: "Use GET /v2/users instead"
          sunset: "2025-12-31"
```

---

## Documentation

### User Documentation

- **[Technical Specification](docs/technical-specification.md)** - Complete system design
- **[Configuration DSL](docs/configuration-dsl.md)** - Configuration reference
- **[User Guide](docs/README.md)** - Getting started

### Design Documentation

- **[Architecture](design/)** - Architecture and design documents
- **[ADRs](design/decisions/)** - Architecture Decision Records
- **[Research](design/research/)** - Technology analysis and comparisons

---

## Technology Stack

- **Language:** Go 1.21+
- **CLI Framework:** [spf13/cobra](https://github.com/spf13/cobra)
- **OpenAPI Parser:** [getkin/kin-openapi](https://github.com/getkin/kin-openapi)
- **Expression Engine:** [expr-lang/expr](https://github.com/expr-lang/expr)
- **Configuration:** [spf13/viper](https://github.com/spf13/viper)

---

## Development

```bash
# Clone repository
git clone https://github.com/cliforge/cliforge.git
cd cliforge

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build ./...
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for commit guidelines and development workflow.

---

## Project Status

CliForge is currently in **design phase**.

**Completed:**
- Architecture design and specification
- Configuration DSL design
- OpenAPI extension definitions
- Technical documentation

**In Progress:**
- Core implementation planning
- Build system design

---

## Why CliForge?

Existing tools require tradeoffs:

**Static Generators** (OpenAPI Generator, swagger-codegen):
- Generate branded binaries
- Require regeneration for API changes
- No self-updating capability

**Dynamic Loaders** (Restish, openapi-cli):
- Load specs dynamically
- Cannot be branded or distributed as standalone tools

**CliForge provides both:**
- Branded, distributable binaries
- Dynamic API updates
- Self-updating for security
- Change notifications
- Workflow orchestration

---

## License

MIT License - See [LICENSE](LICENSE) for details.

---

## Links

- **Documentation:** [docs/](docs/)
- **Contributing:** [CONTRIBUTING.md](CONTRIBUTING.md)
