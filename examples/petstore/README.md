# Petstore CLI - Complete CliForge v0.9.0 Example

A comprehensive, working example that demonstrates **ALL** features of CliForge v0.9.0.

## Overview

This example includes:

- **Complete OpenAPI 3.0 Spec** (`petstore-api.yaml`) - Uses every `x-cli-*` extension
- **Full CLI Configuration** (`cli-config.yaml`) - All features enabled and configured
- **Mock API Server** (`mock-server.go`) - Fully functional HTTP server with SSE support
- **Build System** (`build.sh`) - Automated build and setup
- **Interactive Demo** (`demo.sh`) - Guided tour of all features

## Features Demonstrated

### OpenAPI Extensions

| Extension | Feature | Example |
|-----------|---------|---------|
| `x-cli-command` | Command naming | `list pets`, `create pet` |
| `x-cli-aliases` | Alternative names | `ls pets`, `new pet` |
| `x-cli-flags` | CLI flag configuration | `--name`, `--category` |
| `x-cli-interactive` | Interactive prompts | Select, text, confirm inputs |
| `x-cli-preflight` | Pre-execution validation | Capacity check, duplicate check |
| `x-cli-output` | Output formatting | Table, JSON, YAML, CSV |
| `x-cli-async` | Async polling | Order status tracking |
| `x-cli-workflow` | Multi-step operations | Pet adoption workflow |
| `x-cli-plugin` | External tool integration | AWS CLI for S3 backup |
| `x-cli-streaming` | Server-Sent Events | Real-time pet status |
| `x-cli-watch` | Resource monitoring | Watch for changes |
| `x-cli-confirmation` | Destructive operation safety | Delete confirmation |
| `x-cli-deprecation` | Sunset warnings | Deprecated commands |
| `x-cli-context` | Environment switching | dev/staging/prod |
| `x-cli-cache` | Response caching | Reduce API calls |
| `x-cli-hidden` | Internal endpoints | Hide from help |

### Configuration Features

- **Branding**: ASCII art banner, custom colors, themes
- **Authentication**: OAuth2 with auto-refresh, keyring storage
- **Output Formats**: Table, JSON, YAML, CSV, custom templates
- **Rate Limiting**: Request throttling with retry logic
- **Error Handling**: Custom error messages and suggestions
- **Telemetry**: Usage tracking (optional)
- **Logging**: Structured logs with rotation
- **Updates**: Self-update mechanism
- **Secrets**: Detection and masking
- **XDG Compliance**: Standard directory layout

## Quick Start

### Prerequisites

- Go 1.21 or later
- Optional: `jq` for JSON processing

### Run the Example

```bash
# Build everything and start the mock server
./build.sh

# Run the interactive demo
./demo.sh

# Or run in non-interactive mode
./demo.sh --non-interactive
```

### Manual Steps

```bash
# 1. Build and start the mock API server
./build.sh server

# 2. In another terminal, build the CLI generator
./build.sh cli

# 3. Test the mock API
curl http://localhost:8080/openapi.yaml

# 4. Stop the server when done
./build.sh stop

# 5. Clean up
./build.sh clean
```

## File Structure

```
examples/petstore/
├── README.md              # This file
├── petstore-api.yaml      # Complete OpenAPI spec with all x-cli-* extensions
├── cli-config.yaml        # Full CLI configuration
├── mock-server.go         # Mock API server implementation
├── build.sh               # Build and setup script
└── demo.sh                # Interactive demonstration
```

## OpenAPI Spec Highlights

### Complete Extension Usage

The `petstore-api.yaml` demonstrates every extension:

```yaml
# Global configuration
x-cli-config:
  name: petstore
  branding:
    ascii-art: |
      [ASCII art banner]
    colors:
      primary: "#FF6B35"
  features:
    interactive-mode: true
    self-update: true
    context-switching: true

# Per-operation extensions
paths:
  /pets:
    get:
      # Command configuration
      x-cli-command: "list pets"
      x-cli-aliases: ["ls pets", "pets"]

      # Output formatting
      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
              width: 10
            - field: name
              header: NAME
            - field: status
              header: STATUS
              color-map:
                available: green
                pending: yellow

      # Response caching
      x-cli-cache:
        enabled: true
        ttl: 60

    post:
      # Interactive prompts
      x-cli-interactive:
        prompts:
          - parameter: name
            type: text
            message: "What is the pet's name?"
          - parameter: category
            type: select
            choices:
              - value: dog
                label: "Dog 🐕"

      # Pre-flight validation
      x-cli-preflight:
        - name: validate-capacity
          endpoint: "/stores/capacity"
          required: true

  # Workflow orchestration
  /workflow/adopt:
    post:
      x-cli-workflow:
        steps:
          - id: check-availability
            request:
              method: GET
              url: "{base_url}/pets/{args.petId}"
          - id: create-order
            request:
              method: POST
              url: "{base_url}/orders"
            condition: "check-availability.body.status == 'available'"

  # Streaming support
  /pets/{petId}/status/stream:
    get:
      x-cli-streaming:
        enabled: true
        type: sse
        event-types:
          - status-change
          - health-check

  # Plugin integration
  /pets/{petId}/backup:
    post:
      x-cli-plugin:
        type: external
        command: aws
        operations:
          - plugin-call:
              command: "aws"
              args: ["s3", "cp", "-", "s3://{bucket}/pets/{petId}.json"]
```

## CLI Configuration Highlights

The `cli-config.yaml` shows comprehensive configuration:

```yaml
metadata:
  name: petstore
  version: 1.0.0

branding:
  colors:
    primary: "#FF6B35"
    success: "#06D6A0"
  ascii_art: |
    [Petstore banner]

api:
  openapi_url: http://localhost:8080/openapi.yaml
  cache:
    enabled: true
    ttl: 300

behaviors:
  auth:
    type: oauth2
    storage:
      primary: keyring
      fallback: file
    auto_refresh: true

  output:
    default_format: table
    supported_formats: [table, json, yaml, csv]

  rate_limit:
    requests_per_minute: 60
    retry:
      enabled: true
      max_attempts: 3

features:
  interactive_mode:
    enabled: true
  workflows:
    enabled: true
  plugins:
    enabled: true
  watch:
    enabled: true

contexts:
  default: production
  available:
    development:
      api_url: http://localhost:8080
    production:
      api_url: https://api.petstore.example.com
```

## Mock API Server

The `mock-server.go` implements:

- Full CRUD operations for pets, stores, orders, users
- Server-Sent Events (SSE) for streaming
- Realistic data models and validation
- Async order processing simulation
- Context and region endpoints
- Pre-flight check endpoints

### API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/openapi.yaml` | OpenAPI specification |
| GET | `/pets` | List pets with filtering |
| POST | `/pets` | Create a pet |
| GET | `/pets/{id}` | Get pet details |
| PUT | `/pets/{id}` | Update a pet |
| DELETE | `/pets/{id}` | Delete a pet |
| GET | `/pets/{id}/status/stream` | Stream pet status (SSE) |
| GET | `/stores` | List stores |
| GET | `/stores/capacity` | Get capacity info |
| POST | `/orders` | Create order |
| GET | `/orders/{id}` | Get order status |
| GET | `/users` | List users |
| GET | `/users/{id}` | Get user details |
| GET | `/contexts` | List contexts |
| GET | `/regions` | List regions |

## Example Commands

Once the CLI is generated (future implementation), these commands will work:

```bash
# List operations
petstore-cli list pets
petstore-cli list pets --status available --category dog
petstore-cli list pets --output json

# Create operations
petstore-cli create pet --name "Max" --category dog --age 3
petstore-cli create pet --interactive

# Get operations
petstore-cli get pet --pet-id 1
petstore-cli get pet --pet-id 1 --output yaml

# Update operations
petstore-cli update pet --pet-id 1 --status adopted

# Delete operations
petstore-cli delete pet --pet-id 5 --yes

# Watch mode
petstore-cli get pet --pet-id 1 --watch
petstore-cli watch pet --pet-id 1

# Workflows
petstore-cli adopt pet --pet-id 1 --user-id 1

# Context switching
petstore-cli list contexts
petstore-cli use context production
petstore-cli current context

# Plugin operations
petstore-cli backup pet --pet-id 1 --bucket my-backups

# Advanced
petstore-cli create pet --name "Test" --dry-run
petstore-cli list pets --verbose
```

## Testing the Mock Server

```bash
# Start the server
./build.sh server

# Test endpoints
curl http://localhost:8080/openapi.yaml
curl http://localhost:8080/pets
curl http://localhost:8080/pets/1
curl -X POST http://localhost:8080/pets \
  -H "Content-Type: application/json" \
  -d '{"name":"Fluffy","category":{"name":"cat"},"age":2}'

# Test SSE streaming
curl -N http://localhost:8080/pets/1/status/stream

# Stop the server
./build.sh stop
```

## Extension Reference

### x-cli-command

Defines the CLI command name and structure.

```yaml
x-cli-command: "list pets"  # Creates: petstore-cli list pets
```

### x-cli-interactive

Enables interactive prompts for parameters.

```yaml
x-cli-interactive:
  enabled: true
  prompts:
    - parameter: name
      type: text
      message: "Pet name?"
      validation: "^[a-zA-Z0-9 ]{1,50}$"
```

### x-cli-workflow

Defines multi-step workflows.

```yaml
x-cli-workflow:
  steps:
    - id: step1
      request:
        method: GET
        url: "{base_url}/resource"
    - id: step2
      condition: "step1.body.ready == true"
      request:
        method: POST
        url: "{base_url}/action"
```

### x-cli-output

Customizes output formatting.

```yaml
x-cli-output:
  table:
    columns:
      - field: id
        header: ID
        width: 10
      - field: status
        header: STATUS
        color-map:
          available: green
          pending: yellow
```

### x-cli-async

Configures async polling.

```yaml
x-cli-async:
  enabled: true
  status-field: "status"
  status-endpoint: "/orders/{id}"
  terminal-states: [delivered, cancelled]
  polling:
    interval: 30
    timeout: 3600
```

## Development Notes

This example is designed to be:

1. **Self-contained**: All files in one directory
2. **Fully functional**: Mock server works out of the box
3. **Comprehensive**: Demonstrates every feature
4. **Educational**: Well-commented and documented
5. **Realistic**: Uses patterns from real-world CLIs (AWS, Heroku, etc.)

## Implementation Status

- [x] OpenAPI spec with all extensions
- [x] CLI configuration
- [x] Mock API server
- [x] Build scripts
- [x] Demo scripts
- [ ] CliForge CLI generator (future implementation)
- [ ] Generated CLI binary (future implementation)

## Related Documentation

- [CliForge Technical Specification](../../docs/technical-specification.md)
- [Configuration DSL Reference](../../docs/configuration-dsl.md)
- [Contributing Guidelines](../../CONTRIBUTING.md)

## License

MIT License - Part of the CliForge project.
