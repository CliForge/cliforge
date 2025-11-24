# Getting Started with CliForge

Welcome to CliForge! This guide will help you create your first API-driven CLI in just a few minutes.

## What is CliForge?

CliForge is a tool that generates professional command-line interfaces from OpenAPI specifications. Unlike traditional code generators, CliForge creates **branded, self-updating binaries** that dynamically load API specifications at runtime.

### Why Use CliForge?

**For API Providers:**
- Distribute a professional CLI to your users without writing CLI code
- Update API endpoints without releasing new CLI versions
- Provide a consistent developer experience across your ecosystem
- Self-updating binaries ensure users always have security patches

**For Developers:**
- Single binary distribution - no dependencies to install
- Automatic command structure from your OpenAPI spec
- Built-in authentication, rate limiting, and output formatting
- Context switching between dev/staging/production environments

### How It Works

```
┌─────────────────┐
│  OpenAPI Spec   │ ──┐
│  (Your API)     │   │
└─────────────────┘   │
                      ├──> CliForge ──> Branded CLI Binary
┌─────────────────┐   │                  (my-api-cli)
│  Config YAML    │ ──┘
│  (Branding)     │
└─────────────────┘
```

At runtime, your CLI:
1. Loads the OpenAPI spec (cached for speed)
2. Builds commands from API operations
3. Executes API calls with proper auth and formatting
4. Self-updates when new versions are available

---

## Quick Start (5 Minutes)

Let's create a simple CLI for a hypothetical petstore API.

### Prerequisites

- Go 1.21 or later installed
- Basic familiarity with REST APIs and OpenAPI
- An OpenAPI specification for your API (or use our example)

### Step 1: Install CliForge

```bash
# From releases (recommended - coming soon)
curl -L https://github.com/cliforge/cliforge/releases/latest/download/cliforge-$(uname -s)-$(uname -m) -o cliforge
chmod +x cliforge
sudo mv cliforge /usr/local/bin/

# Or from source (current method)
git clone https://github.com/cliforge/cliforge.git
cd cliforge
go build -o cliforge ./cmd/cliforge
sudo mv cliforge /usr/local/bin/
```

### Step 2: Initialize Your CLI Project

```bash
# Create a new CLI project
cliforge init my-api-cli

# This creates:
# my-api-cli/
# ├── cli-config.yaml    # CLI configuration
# └── README.md          # Project documentation
```

### Step 3: Configure Your CLI

Edit `cli-config.yaml`:

```yaml
metadata:
  name: my-api-cli
  version: 1.0.0
  description: My API Command Line Interface

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com

branding:
  colors:
    primary: "#FF6B35"
  tagline: "Manage your API from the command line"

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: MY_API_KEY
```

### Step 4: Build Your CLI

```bash
cd my-api-cli
cliforge build --config cli-config.yaml --output dist/

# This generates platform-specific binaries:
# dist/
# ├── my-api-cli-darwin-amd64
# ├── my-api-cli-linux-amd64
# └── my-api-cli-windows-amd64.exe
```

### Step 5: Try Your CLI

```bash
# Set your API key
export MY_API_KEY="your-api-key-here"

# Run your CLI
./dist/my-api-cli-darwin-amd64 --help

# The CLI automatically generates commands from your OpenAPI spec
./dist/my-api-cli-darwin-amd64 users list
./dist/my-api-cli-darwin-amd64 users get --user-id 123
```

**That's it!** You now have a working CLI for your API.

---

## Your First CLI (Step-by-Step)

Let's build a more complete example using the included petstore example.

### Example: Petstore CLI

The CliForge repository includes a complete petstore example that demonstrates all features.

#### 1. Clone and Setup

```bash
# Clone the repository
git clone https://github.com/cliforge/cliforge.git
cd cliforge/examples/petstore

# Review the files
ls -la
# cli-config.yaml     # Full configuration example
# petstore-api.yaml   # OpenAPI spec with all extensions
# mock-server.go      # Mock API server for testing
# build.sh            # Build script
# demo.sh             # Interactive demo
```

#### 2. Start the Mock API Server

```bash
# Build and start the mock server
./build.sh server

# In another terminal, verify it's running
curl http://localhost:8080/openapi.yaml
curl http://localhost:8080/pets
```

The mock server provides:
- Full CRUD operations for pets, stores, orders, and users
- Server-Sent Events for real-time updates
- Pre-flight validation endpoints
- Realistic data and error responses

#### 3. Examine the Configuration

Open `cli-config.yaml` to see comprehensive configuration options:

```yaml
# Branding
branding:
  ascii_art: |
    ██████╗ ███████╗████████╗███████╗████████╗ ██████╗ ██████╗ ███████╗
    # ... ASCII art banner
  tagline: "Manage your petstore from the command line"
  colors:
    primary: "#FF6B35"
    success: "#06D6A0"

# API configuration
api:
  openapi_url: http://localhost:8080/openapi.yaml
  base_url: http://localhost:8080
  cache:
    enabled: true
    ttl: 300

# Features
features:
  interactive_mode:
    enabled: true
  workflows:
    enabled: true
  watch:
    enabled: true
```

#### 4. Build the CLI (Future)

*Note: CLI generation is in development. The configuration and OpenAPI spec are ready to use.*

```bash
# When implemented, this will work:
cliforge build --config cli-config.yaml --output dist/

# And you'll be able to run:
./dist/petstore-cli list pets
./dist/petstore-cli create pet --name "Max" --category dog
./dist/petstore-cli watch pet --pet-id 1
```

#### 5. Explore the OpenAPI Extensions

Open `petstore-api.yaml` to see advanced features:

**Command Configuration:**
```yaml
paths:
  /pets:
    get:
      x-cli-command: "list pets"
      x-cli-aliases: ["ls pets", "pets"]
      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
            - field: name
              header: NAME
            - field: status
              header: STATUS
              color-map:
                available: green
                pending: yellow
```

**Interactive Prompts:**
```yaml
post:
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
          - value: cat
            label: "Cat 🐱"
```

**Multi-Step Workflows:**
```yaml
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
```

---

## Basic Configuration

Here are the essential configuration options you'll use most often.

### Minimal Configuration

The absolute minimum to get started:

```yaml
metadata:
  name: my-cli
  version: 1.0.0

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com
```

### Adding Authentication

**API Key:**
```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: MY_API_KEY
```

**OAuth2:**
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: my-cli
      scopes: [read, write]
      flow: authorization_code
      authorization_url: https://auth.example.com/authorize
      token_url: https://auth.example.com/token
      storage:
        primary: keyring  # Secure OS keychain
        fallback: file
```

**Basic Auth:**
```yaml
behaviors:
  auth:
    type: basic
    basic:
      username_env: API_USERNAME
      password_env: API_PASSWORD
```

### Output Formatting

```yaml
behaviors:
  output:
    default_format: table  # table, json, yaml, csv
    pretty_print: true
    color: true

    table:
      style: unicode  # unicode, ascii, minimal
      max_width: auto
```

### Rate Limiting

```yaml
behaviors:
  rate_limit:
    enabled: true
    requests_per_minute: 60
    retry:
      enabled: true
      max_attempts: 3
      backoff: exponential
```

### Context Switching

Support multiple environments (dev, staging, production):

```yaml
contexts:
  default: production

  available:
    development:
      api_url: http://localhost:8080
      auth:
        type: none
      color: yellow

    production:
      api_url: https://api.example.com
      auth:
        type: oauth2
      color: green
```

---

## Running Your Generated CLI

### Basic Commands

Every CliForge CLI includes these built-in commands:

```bash
# View help
my-cli --help
my-cli users --help
my-cli users list --help

# List available commands (from your OpenAPI spec)
my-cli --help

# Check version
my-cli --version

# Check for updates
my-cli --update
```

### Authentication

```bash
# Set API key via environment variable
export MY_API_KEY="your-key-here"
my-cli users list

# Or use a config file
my-cli config set api_key "your-key-here"

# OAuth2 login (opens browser)
my-cli login

# Check auth status
my-cli whoami

# Logout
my-cli logout
```

### Working with Data

```bash
# List resources
my-cli users list
my-cli users list --limit 10
my-cli users list --filter 'status=active'

# Get specific resource
my-cli users get --user-id 123

# Create resources
my-cli users create --name "John Doe" --email "john@example.com"

# Update resources
my-cli users update --user-id 123 --status active

# Delete resources (with confirmation)
my-cli users delete --user-id 123
my-cli users delete --user-id 123 --yes  # Skip confirmation
```

### Output Formats

```bash
# Table (default)
my-cli users list

# JSON
my-cli users list --output json

# YAML
my-cli users list --output yaml

# CSV
my-cli users list --output csv

# Pretty-printed JSON
my-cli users get --user-id 123 --output json | jq
```

### Advanced Features

```bash
# Watch mode (auto-refresh)
my-cli users get --user-id 123 --watch

# Dry run (preview without executing)
my-cli users create --name "Test" --dry-run

# Verbose output
my-cli users list --verbose

# Debug mode
my-cli users list --debug

# Different context
my-cli --context staging users list
```

---

## Next Steps

Now that you have a working CLI, here are some things to explore:

### 1. Customize Your Branding

- Add ASCII art banner
- Configure color scheme
- Set custom error messages
- Add shell aliases

See: [Configuration DSL Reference](configuration-dsl.md)

### 2. Add Advanced Features

**Enable interactive mode:**
```yaml
features:
  interactive_mode:
    enabled: true
```

**Configure workflows:**
```yaml
features:
  workflows:
    enabled: true
    max_steps: 50
```

**Add plugins:**
```yaml
features:
  plugins:
    enabled: true
    search_paths:
      - "~/.my-cli/plugins"
```

### 3. Enhance Your OpenAPI Spec

Add CliForge extensions to your OpenAPI spec for better CLI UX:

- `x-cli-command` - Custom command names
- `x-cli-aliases` - Command shortcuts
- `x-cli-output` - Custom output formatting
- `x-cli-interactive` - Interactive prompts
- `x-cli-workflow` - Multi-step operations

See: [Technical Specification](technical-specification.md#openapi-extensions)

### 4. Set Up Self-Updates

Configure automatic updates for your users:

```yaml
updates:
  enabled: true
  update_url: https://releases.example.com/my-cli
  channels:
    - stable
    - beta
  current_channel: stable
```

### 5. Add Context Switching

Support multiple environments:

```bash
my-cli context list
my-cli context use staging
my-cli context current
```

### 6. Enable Shell Completion

Generate shell completion scripts:

```bash
my-cli completion bash > /etc/bash_completion.d/my-cli
my-cli completion zsh > ~/.zsh/completions/_my-cli
my-cli completion fish > ~/.config/fish/completions/my-cli.fish
```

---

## Examples and Resources

### Example Projects

1. **Petstore CLI** - Complete example with all features
   - Location: `examples/petstore/`
   - Run: `cd examples/petstore && ./demo.sh`
   - Shows: All OpenAPI extensions, workflows, streaming, plugins

### Documentation

- [Installation Guide](installation.md) - Detailed installation instructions
- [Configuration DSL](configuration-dsl.md) - Complete configuration reference
- [Technical Specification](technical-specification.md) - System architecture and design
- [Contributing Guide](../CONTRIBUTING.md) - How to contribute to CliForge

### Getting Help

- **GitHub Issues**: [github.com/cliforge/cliforge/issues](https://github.com/cliforge/cliforge/issues)
- **Documentation**: [docs/](.)
- **Examples**: [examples/](../examples/)

---

## Common Patterns

### Pattern 1: List-Get-Update Workflow

Most CLIs follow this pattern:

```bash
# 1. List resources to find ID
my-cli users list

# 2. Get details for specific resource
my-cli users get --user-id 123

# 3. Update the resource
my-cli users update --user-id 123 --status active

# 4. Verify the update
my-cli users get --user-id 123
```

### Pattern 2: Create with Interactive Mode

For resources with many fields:

```bash
# Interactive mode prompts for each field
my-cli users create --interactive

# Or provide some fields, prompt for others
my-cli users create --email "user@example.com" --interactive
```

### Pattern 3: Bulk Operations with Workflows

Process multiple resources:

```bash
# CLI automatically handles the workflow:
# 1. Get list of users
# 2. For each user, fetch details
# 3. Combine results
my-cli users export --format json > users-backup.json
```

### Pattern 4: Environment-Specific Operations

Work across environments:

```bash
# Development
my-cli --context dev users create --name "Test User"

# Verify in staging
my-cli --context staging users list

# Promote to production
my-cli --context prod users create --name "Real User"
```

---

## Troubleshooting

### CLI Not Finding OpenAPI Spec

```bash
# Check spec URL is accessible
curl -I https://api.example.com/openapi.yaml

# Force refresh the cached spec
my-cli --refresh users list

# Check cache location
my-cli config show cache_dir
ls -la $(my-cli config show cache_dir)
```

### Authentication Issues

```bash
# Check current auth status
my-cli whoami

# Clear cached credentials
my-cli logout

# Re-authenticate
my-cli login

# Verify credentials are stored
my-cli config show auth
```

### Rate Limiting

```bash
# Check rate limit status
my-cli --verbose users list

# Reduce concurrent requests
my-cli config set rate_limit.requests_per_minute 30
```

### Command Not Found

```bash
# Refresh spec to get latest commands
my-cli --refresh --help

# Check if operation is hidden
my-cli --show-hidden --help

# Verify OpenAPI spec has the operation
curl https://api.example.com/openapi.yaml | grep operationId
```

---

## What's Next?

You now have everything you need to create and use a CliForge CLI. Here's what to explore:

1. **Read the [Installation Guide](installation.md)** for detailed setup instructions
2. **Review the [Configuration DSL](configuration-dsl.md)** for all available options
3. **Study the [Petstore Example](../examples/petstore/)** for real-world patterns
4. **Check the [Technical Specification](technical-specification.md)** for advanced features

Happy CLI building with CliForge!
