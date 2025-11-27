# OpenAPI Extensions Reference

**CliForge v0.9.0**

This document provides comprehensive reference documentation for all CliForge OpenAPI extensions (`x-cli-*` and `x-auth-*`). These extensions transform standard OpenAPI specifications into rich CLI definitions, enabling features like interactive prompts, async workflows, progress tracking, and advanced CLI patterns.

## Table of Contents

1. [Overview](#overview)
2. [Global Extensions](#global-extensions)
   - [x-cli-config](#x-cli-config)
3. [Security Extensions](#security-extensions)
   - [x-auth-config](#x-auth-config)
4. [Command Extensions](#command-extensions)
   - [x-cli-command](#x-cli-command)
   - [x-cli-aliases](#x-cli-aliases)
5. [Parameter Extensions](#parameter-extensions)
   - [x-cli-flags](#x-cli-flags)
6. [Interactive Extensions](#interactive-extensions)
   - [x-cli-interactive](#x-cli-interactive)
7. [Validation Extensions](#validation-extensions)
   - [x-cli-preflight](#x-cli-preflight)
   - [x-cli-confirmation](#x-cli-confirmation)
8. [Async Extensions](#async-extensions)
   - [x-cli-async](#x-cli-async)
9. [Output Extensions](#output-extensions)
   - [x-cli-output](#x-cli-output)
10. [Workflow Extensions](#workflow-extensions)
    - [x-cli-workflow](#x-cli-workflow)
11. [Plugin Extensions](#plugin-extensions)
    - [x-cli-plugin](#x-cli-plugin)
12. [File Extensions](#file-extensions)
    - [x-cli-file-input](#x-cli-file-input)
13. [Monitoring Extensions](#monitoring-extensions)
    - [x-cli-watch](#x-cli-watch)
14. [Deprecation Extensions](#deprecation-extensions)
    - [x-cli-deprecation](#x-cli-deprecation)
15. [Secret Extensions](#secret-extensions)
    - [x-cli-secret](#x-cli-secret)
16. [Context Extensions](#context-extensions)
    - [x-cli-context](#x-cli-context)
17. [Migration Guide](#migration-guide)
18. [Extension Patterns](#extension-patterns)

---

## Overview

CliForge extends OpenAPI 3.0 with custom extensions that describe CLI-specific behavior. Each extension is prefixed with `x-cli-` or `x-auth-` and can be added to various parts of the OpenAPI specification.

### Extension Locations

| Extension | Location | Purpose |
|-----------|----------|---------|
| `x-cli-config` | Root | Global CLI configuration |
| `x-auth-config` | SecurityScheme | OAuth2/Auth configuration |
| `x-cli-command` | Operation | Command name mapping |
| `x-cli-aliases` | Operation | Command aliases |
| `x-cli-flags` | Operation | Request body to CLI flags |
| `x-cli-interactive` | Operation | Interactive prompts |
| `x-cli-preflight` | Operation | Pre-execution checks |
| `x-cli-confirmation` | Operation | User confirmation |
| `x-cli-async` | Operation | Async operation handling |
| `x-cli-output` | Operation | Output formatting |
| `x-cli-workflow` | Operation | Multi-step workflows |
| `x-cli-plugin` | Operation | External plugin calls |
| `x-cli-file-input` | Parameter | File upload handling |
| `x-cli-watch` | Operation | Real-time monitoring |
| `x-cli-deprecation` | Operation | Deprecation warnings |
| `x-cli-secret` | Parameter | Secret/sensitive data |
| `x-cli-context` | Operation | Environment contexts |

### Extension Principles

1. **Backward Compatible**: Extensions don't break standard OpenAPI parsers
2. **Progressive Enhancement**: CLIs work with minimal extensions, better with more
3. **Declarative**: Describe behavior, not implementation
4. **Composable**: Extensions work together seamlessly
5. **Validated**: Extensions are validated against schemas

---

## Global Extensions

### x-cli-config

**Location**: Root level of OpenAPI document
**Type**: Object
**Purpose**: Defines global CLI configuration, branding, and feature flags

#### Schema

```yaml
x-cli-config:
  name: string                      # CLI binary name (required)
  version: string                   # CLI version (required)
  description: string               # CLI description

  branding:
    ascii-art: string              # ASCII art for splash screen
    colors:
      primary: string              # Primary color (hex)
      secondary: string            # Secondary color (hex)
      success: string              # Success color (hex)
      warning: string              # Warning color (hex)
      error: string                # Error color (hex)

  auth:
    type: string                   # oauth2, apikey, basic
    storage: string                # file, keyring, memory
    auto-refresh: boolean          # Auto-refresh tokens

  output:
    default-format: string         # table, json, yaml, csv
    supported-formats: [string]    # Available formats

  features:
    interactive-mode: boolean      # Enable interactive prompts
    auto-complete: boolean         # Shell auto-completion
    self-update: boolean           # Binary self-update
    telemetry: boolean             # Usage telemetry
    context-switching: boolean     # Multi-environment support

  cache:
    enabled: boolean               # Enable response caching
    ttl: integer                   # Cache TTL in seconds
    location: string               # Cache directory path

  plugins:
    enabled: boolean               # Enable plugin system
    search-paths: [string]         # Plugin search directories
```

#### Simple Example

```yaml
openapi: 3.0.3
info:
  title: My API
  version: 1.0.0

x-cli-config:
  name: "myapi"
  version: "1.0.0"
  description: "My API CLI tool"

  output:
    default-format: table
    supported-formats: [table, json, yaml]

  features:
    interactive-mode: true
    auto-complete: true
```

#### Advanced Example

```yaml
x-cli-config:
  name: "petstore"
  version: "1.0.0"
  description: "Petstore CLI - Complete CliForge Example"

  branding:
    ascii-art: |
      ██████╗ ███████╗████████╗███████╗████████╗ ██████╗ ██████╗ ███████╗
      ██╔══██╗██╔════╝╚══██╔══╝██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗██╔════╝
      ██████╔╝█████╗     ██║   ███████╗   ██║   ██║   ██║██████╔╝█████╗
      ██╔═══╝ ██╔══╝     ██║   ╚════██║   ██║   ██║   ██║██╔══██╗██╔══╝
      ██║     ███████╗   ██║   ███████║   ██║   ╚██████╔╝██║  ██║███████╗
      ╚═╝     ╚══════╝   ╚═╝   ╚══════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝
    colors:
      primary: "#FF6B35"
      secondary: "#004E89"
      success: "#06D6A0"
      warning: "#F4A261"
      error: "#E76F51"

  auth:
    type: oauth2
    storage: keyring
    auto-refresh: true

  output:
    default-format: table
    supported-formats: [table, json, yaml, csv, text]

  features:
    interactive-mode: true
    auto-complete: true
    self-update: true
    telemetry: false
    context-switching: true

  cache:
    enabled: true
    ttl: 300
    location: "~/.cache/petstore"

  plugins:
    enabled: true
    search-paths:
      - "~/.petstore/plugins"
      - "/usr/local/lib/petstore/plugins"
```

#### Best Practices

1. **Name**: Use lowercase, alphanumeric CLI binary name
2. **Colors**: Use hex colors for consistent branding
3. **Storage**: Prefer `keyring` for security-sensitive CLIs
4. **Formats**: Always support `json` for scriptability
5. **Cache**: Set reasonable TTL based on data volatility
6. **Plugins**: Enable only if you need external tool integration

#### Common Pitfalls

- ❌ Using spaces or special chars in CLI name
- ❌ Not providing default output format
- ❌ Setting cache TTL too high for dynamic data
- ❌ Enabling telemetry without user consent
- ✅ Keep branding colors accessible (good contrast)
- ✅ Document cache location in user docs

---

## Security Extensions

### x-auth-config

**Location**: `components.securitySchemes.<scheme>`
**Type**: Object
**Purpose**: Extends OAuth2 and other auth schemes with CLI-specific configuration

#### Schema

```yaml
x-auth-config:
  flows:
    authorizationCode:
      authorizationUrl: string
      tokenUrl: string
      refreshUrl: string
      scopes:
        <scope>: string              # Scope description
    clientCredentials:
      tokenUrl: string
      scopes: {}
    implicit:
      authorizationUrl: string
      scopes: {}
    password:
      tokenUrl: string
      scopes: {}

  token-storage:
    - type: string                   # file, keyring
      path: string                   # Path for file storage
      permissions: string            # File permissions (octal)
      service: string                # Service name for keyring
      keyring-backends:
        darwin: string               # macOS backend (keychain)
        linux: string                # Linux backend (secret-service)
        windows: string              # Windows backend (wincred)
```

#### Simple Example

```yaml
components:
  securitySchemes:
    OAuth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://auth.example.com/authorize
          tokenUrl: https://auth.example.com/token
          scopes:
            read: Read access
            write: Write access

      x-auth-config:
        flows:
          authorizationCode:
            refreshUrl: https://auth.example.com/token
        token-storage:
          - type: keyring
            service: "myapi-cli"
```

#### Advanced Example

```yaml
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      description: OAuth2 Bearer token

      x-auth-config:
        flows:
          authorizationCode:
            authorizationUrl: https://auth.petstore.example.com/authorize
            tokenUrl: https://auth.petstore.example.com/token
            refreshUrl: https://auth.petstore.example.com/token
            scopes:
              read:pets: Read pet information
              write:pets: Modify pet information
              admin: Administrative access

        # Multi-backend token storage with fallback
        token-storage:
          # Primary: Use OS keyring
          - type: keyring
            service: "petstore-cli"
            keyring-backends:
              darwin: keychain           # macOS Keychain
              linux: secret-service      # GNOME Keyring / KWallet
              windows: wincred           # Windows Credential Manager

          # Fallback: Encrypted file
          - type: file
            path: "~/.petstore/credentials.json"
            permissions: "0600"
```

#### Best Practices

1. **Security**: Always prefer `keyring` over `file` storage
2. **Permissions**: Set restrictive file permissions (0600)
3. **Refresh**: Always provide `refreshUrl` for long-lived sessions
4. **Scopes**: Document each scope clearly
5. **Fallback**: Provide file storage fallback for environments without keyring
6. **Service Name**: Use unique service names to avoid keyring conflicts

#### Common Pitfalls

- ❌ Storing tokens in world-readable files
- ❌ Not providing refresh URL for OAuth2
- ❌ Using same service name across multiple CLIs
- ❌ Missing keyring backends for target platforms
- ✅ Test token storage on all target platforms
- ✅ Document manual token rotation procedures

---

## Command Extensions

### x-cli-command

**Location**: Operation object
**Type**: String
**Purpose**: Maps an OpenAPI operation to a CLI command

#### Schema

```yaml
x-cli-command: string    # Command name (supports subcommands with spaces)
```

#### Simple Example

```yaml
paths:
  /pets:
    get:
      operationId: listPets
      x-cli-command: "list pets"
```

Generates: `myapi list pets`

#### Advanced Example

```yaml
paths:
  /clusters/{clusterId}/machine-pools:
    get:
      operationId: listMachinePools
      x-cli-command: "list machinepools"
      summary: List machine pools for a cluster

    post:
      operationId: createMachinePool
      x-cli-command: "create machinepool"
      summary: Create a new machine pool
```

Generates:
- `myapi list machinepools --cluster <id>`
- `myapi create machinepool --cluster <id>`

#### Best Practices

1. **Naming**: Use verb-noun pattern (list pets, create cluster)
2. **Consistency**: Use consistent verbs across resources
3. **Subcommands**: Use spaces to create command hierarchy
4. **Uniqueness**: Each command must be unique
5. **Discoverability**: Use clear, descriptive names

**Standard Verbs**:
- `list` - List multiple resources
- `get` / `describe` - Get single resource details
- `create` / `add` - Create new resource
- `update` / `edit` - Update existing resource
- `delete` / `remove` - Delete resource

#### Common Pitfalls

- ❌ Inconsistent verb usage (list vs get-all)
- ❌ Overly complex command hierarchies (more than 3 levels)
- ❌ Duplicate command names
- ❌ Using underscores instead of spaces for subcommands
- ✅ Test command auto-completion
- ✅ Document all available commands

---

### x-cli-aliases

**Location**: Operation object
**Type**: Array of strings
**Purpose**: Provides alternative command names

#### Schema

```yaml
x-cli-aliases: [string]    # Alternative command names
```

#### Simple Example

```yaml
paths:
  /pets:
    get:
      x-cli-command: "list pets"
      x-cli-aliases: ["ls pets", "pets"]
```

Generates:
- `myapi list pets`
- `myapi ls pets`
- `myapi pets`

#### Advanced Example

```yaml
paths:
  /clusters/{clusterId}:
    delete:
      x-cli-command: "delete cluster"
      x-cli-aliases: ["remove cluster", "rm cluster", "destroy cluster"]
```

#### Best Practices

1. **Common Shortcuts**: Provide short aliases for frequent commands (ls, rm)
2. **User Familiarity**: Include aliases from similar tools (kubectl, aws)
3. **Documentation**: Document primary command, list aliases in help
4. **Limit Aliases**: Don't exceed 3-4 aliases per command
5. **No Conflicts**: Ensure aliases don't conflict with other commands

#### Common Pitfalls

- ❌ Too many aliases causing confusion
- ❌ Aliases that conflict with primary commands
- ❌ Ambiguous short forms (rm could be remove or rename)
- ✅ Test alias completion
- ✅ Show primary command in help text

---

## Parameter Extensions

### x-cli-flags

**Location**: Operation object
**Type**: Array of objects
**Purpose**: Maps request body fields to CLI flags

#### Schema

```yaml
x-cli-flags:
  - name: string                   # Flag name (kebab-case)
    source: string                 # JSON path in request body
    flag: string                   # Full flag (--flag-name)
    aliases: [string]              # Short flags ([-f, -n])
    required: boolean              # Is flag required
    type: string                   # string, integer, boolean, float, array
    enum: [string]                 # Allowed values
    default: any                   # Default value
    description: string            # Help text
    validation: string             # Regex or expression
    validation-message: string     # Error message
```

#### Simple Example

```yaml
post:
  operationId: createPet
  requestBody:
    content:
      application/json:
        schema:
          type: object
          properties:
            name:
              type: string
            age:
              type: integer

  x-cli-flags:
    - name: name
      source: name
      flag: "--name"
      aliases: ["-n"]
      required: true
      type: string
      description: "Pet name"

    - name: age
      source: age
      flag: "--age"
      type: integer
      description: "Pet age in years"
```

Usage: `myapi create pet --name Fluffy --age 3`

#### Advanced Example

```yaml
post:
  operationId: createCluster

  x-cli-flags:
    - name: cluster-name
      source: name
      flag: "--cluster-name"
      aliases: ["-n"]
      required: true
      type: string
      description: "Cluster name"
      validation: "^[a-z][a-z0-9-]{0,53}$"
      validation-message: "Name must be lowercase alphanumeric with hyphens, 1-54 chars"

    - name: region
      source: region.id
      flag: "--region"
      aliases: ["-r"]
      required: true
      type: string
      enum: [us-east-1, us-west-2, eu-west-1]
      description: "AWS region"

    - name: multi-az
      source: multi_az
      flag: "--multi-az"
      type: boolean
      default: false
      description: "Deploy across multiple availability zones"

    - name: node-count
      source: nodes.count
      flag: "--nodes"
      type: integer
      default: 3
      description: "Number of worker nodes"
      validation: "value >= 1 && value <= 100"
      validation-message: "Node count must be 1-100"

    - name: tags
      source: tags
      flag: "--tag"
      aliases: ["-t"]
      type: array
      description: "Resource tags (key=value)"
```

Usage:
```bash
myapi create cluster \
  --cluster-name my-prod-cluster \
  --region us-east-1 \
  --multi-az \
  --nodes 5 \
  --tag env=prod \
  --tag team=platform
```

#### Best Practices

1. **Naming**: Use kebab-case for flag names
2. **Short Flags**: Single letter for common flags (-n, -r, -f)
3. **Required**: Mark truly required flags, use defaults otherwise
4. **Validation**: Validate early with clear error messages
5. **Defaults**: Provide sensible defaults for optional flags
6. **Arrays**: Use repeatable flags for arrays (--tag, --tag)
7. **Nested**: Use dot notation for nested fields (region.id)

#### Flag Types

| Type | CLI Input | Parsed Value |
|------|-----------|--------------|
| `string` | `--name foo` | `"foo"` |
| `integer` | `--count 5` | `5` |
| `float` | `--price 99.99` | `99.99` |
| `boolean` | `--multi-az` | `true` |
| `array` | `--tag a --tag b` | `["a", "b"]` |

#### Common Pitfalls

- ❌ Using underscores in flag names (use kebab-case)
- ❌ Short flags that conflict (-h for help, -v for version)
- ❌ Required flags with defaults (contradictory)
- ❌ Complex regex without validation messages
- ❌ Flags for read-only fields
- ✅ Test flag parsing with edge cases
- ✅ Document all flags in help text
- ✅ Validate flag combinations

---

## Interactive Extensions

### x-cli-interactive

**Location**: Operation object
**Type**: Object
**Purpose**: Defines interactive prompts for parameters

#### Schema

```yaml
x-cli-interactive:
  enabled: boolean                 # Enable interactive mode
  prompts:
    - parameter: string            # Request body field (JSON path)
      type: string                 # text, select, confirm, number, password
      message: string              # Prompt message
      default: any                 # Default value
      validation: string           # Validation regex/expression
      validation-message: string   # Error message

      # For type=select
      choices:
        - value: string
          label: string

      # For dynamic choices
      source:
        endpoint: string           # API endpoint
        value-field: string        # Field for value
        display-field: string      # Field for display
        filter: string             # Filter expression

      # For type=number
      min: number
      max: number
      format: string               # currency, percentage
```

#### Simple Example

```yaml
post:
  operationId: createPet

  x-cli-interactive:
    enabled: true
    prompts:
      - parameter: name
        type: text
        message: "What is the pet's name?"
        validation: "^[a-zA-Z0-9 ]{1,50}$"
        validation-message: "Name must be 1-50 alphanumeric characters"

      - parameter: category
        type: select
        message: "Select pet category:"
        choices:
          - value: dog
            label: "Dog"
          - value: cat
            label: "Cat"
          - value: bird
            label: "Bird"
```

Interactive session:
```
$ myapi create pet
? What is the pet's name? Fluffy
? Select pet category: Cat
✓ Pet 'Fluffy' created successfully
```

#### Advanced Example

```yaml
post:
  operationId: createCluster

  x-cli-interactive:
    enabled: true
    prompts:
      # Text input with validation
      - parameter: name
        type: text
        message: "Cluster name:"
        validation: "^[a-z][a-z0-9-]{0,53}$"
        validation-message: "Must be lowercase alphanumeric with hyphens"

      # Dynamic select from API
      - parameter: region
        type: select
        message: "Select region:"
        source:
          endpoint: "/api/v1/regions"
          value-field: "id"
          display-field: "display_name"
          filter: "enabled == true"

      # Number with range
      - parameter: nodes.count
        type: number
        message: "How many worker nodes?"
        default: 3
        min: 1
        max: 100

      # Confirmation
      - parameter: multi_az
        type: confirm
        message: "Deploy to multiple availability zones?"
        default: false

      # Password (masked)
      - parameter: admin_password
        type: password
        message: "Admin password:"
        validation: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).{8,}$"
        validation-message: "Password must be 8+ chars with upper, lower, and digit"

      # Currency input
      - parameter: budget
        type: number
        message: "Monthly budget:"
        default: 1000.00
        format: currency
```

#### Prompt Types

| Type | Input | Use Case |
|------|-------|----------|
| `text` | Free text | Names, descriptions |
| `select` | Choose from list | Enum values, regions |
| `confirm` | Yes/No | Boolean flags |
| `number` | Numeric input | Counts, prices |
| `password` | Masked input | Secrets, passwords |

#### Best Practices

1. **Defaults**: Provide sensible defaults for optional fields
2. **Validation**: Validate early with clear messages
3. **Dynamic Data**: Fetch select options from API when appropriate
4. **Skip**: Allow `--non-interactive` flag to skip prompts
5. **Order**: Prompt in logical order (required first)
6. **Confirmation**: Use confirm for dangerous operations
7. **Progress**: Show progress when fetching dynamic options

#### Common Pitfalls

- ❌ Too many prompts (use required flags instead)
- ❌ Prompts for read-only fields
- ❌ No validation on text inputs
- ❌ Dynamic selects without loading indicators
- ❌ Password prompts without strength requirements
- ✅ Test with invalid inputs
- ✅ Support both interactive and flag modes
- ✅ Document non-interactive usage

---

## Validation Extensions

### x-cli-preflight

**Location**: Operation object
**Type**: Array of objects
**Purpose**: Defines pre-execution validation checks

#### Schema

```yaml
x-cli-preflight:
  - name: string                   # Check name
    description: string            # Status message
    endpoint: string               # API endpoint to call
    method: string                 # HTTP method
    required: boolean              # Fail if check fails
    skip-flag: string              # Flag to skip check
    validation:
      condition: string            # Expression to evaluate
      error-message: string        # Message on failure
```

#### Simple Example

```yaml
post:
  operationId: createCluster

  x-cli-preflight:
    - name: verify-quota
      description: "Checking available quota..."
      endpoint: "/api/v1/quota"
      method: GET
      required: true
      validation:
        condition: "response.body.available > 0"
        error-message: "No quota available"
```

Output:
```
$ myapi create cluster --name test
Checking available quota... ✓
Creating cluster...
```

#### Advanced Example

```yaml
post:
  operationId: createCluster

  x-cli-preflight:
    # Required: Verify credentials
    - name: verify-credentials
      description: "Verifying AWS credentials..."
      endpoint: "/api/v1/aws/credentials/verify"
      method: POST
      required: true
      validation:
        condition: "response.status == 200"
        error-message: "Invalid AWS credentials. Run 'myapi configure aws' first."

    # Required: Check quota
    - name: check-quota
      description: "Checking AWS quotas..."
      endpoint: "/api/v1/aws/quotas/verify"
      method: POST
      required: true
      validation:
        condition: "response.body.clusters_available > 0"
        error-message: "Cluster quota exceeded. Current: {response.body.clusters_used}/{response.body.clusters_limit}"

    # Optional: Check reserved IPs
    - name: check-ips
      description: "Checking available IP addresses..."
      endpoint: "/api/v1/network/ips/available"
      method: GET
      required: false
      skip-flag: "--skip-ip-check"
      validation:
        condition: "response.body.available >= 10"
        error-message: "Insufficient IP addresses ({response.body.available} < 10)"

    # Optional: Validate name uniqueness
    - name: validate-name
      description: "Checking cluster name availability..."
      endpoint: "/api/v1/clusters?name={args.name}"
      method: GET
      required: false
      validation:
        condition: "len(response.body) == 0"
        error-message: "Cluster name '{args.name}' already in use"
```

Output:
```
$ myapi create cluster --name prod-cluster
Verifying AWS credentials... ✓
Checking AWS quotas... ✓
Checking available IP addresses... ⚠ (skipped via --skip-ip-check)
Checking cluster name availability... ✓
Creating cluster...
```

#### Best Practices

1. **Fast**: Keep checks under 2 seconds each
2. **Required**: Only fail on critical checks
3. **Messages**: Provide actionable error messages
4. **Skip**: Allow skipping optional checks
5. **Order**: Run cheap checks first
6. **Idempotent**: Checks should be side-effect free
7. **Retry**: Implement retries for network checks

#### Common Pitfalls

- ❌ Slow preflight checks (> 5 seconds)
- ❌ Checks that modify state
- ❌ Generic error messages
- ❌ All checks required (users can't proceed)
- ❌ No skip flags for optional checks
- ✅ Test with failing checks
- ✅ Document required setup (credentials, etc.)
- ✅ Show progress for long checks

---

### x-cli-confirmation

**Location**: Operation object
**Type**: Object
**Purpose**: Require user confirmation before execution

#### Schema

```yaml
x-cli-confirmation:
  enabled: boolean                 # Enable confirmation
  message: string                  # Confirmation prompt (supports templates)
  flag: string                     # Flag to skip prompt (--yes)
  aliases: [string]                # Flag aliases ([-y])
```

#### Simple Example

```yaml
delete:
  operationId: deletePet

  x-cli-confirmation:
    enabled: true
    message: "Are you sure you want to delete this pet?"
    flag: "--yes"
```

Output:
```
$ myapi delete pet --pet-id 123
? Are you sure you want to delete this pet? (y/N) y
Pet deleted successfully
```

#### Advanced Example

```yaml
delete:
  operationId: deleteCluster

  x-cli-confirmation:
    enabled: true
    message: |
      ⚠️  WARNING: This will permanently delete cluster '{cluster_id}'.

      This will destroy:
      - All worker nodes
      - All persistent volumes
      - All network resources

      This action CANNOT be undone.

      Type the cluster name to confirm: {cluster_id}
    flag: "--yes"
    aliases: ["-y"]
```

Output:
```
$ myapi delete cluster --cluster-id prod-cluster-01
⚠️  WARNING: This will permanently delete cluster 'prod-cluster-01'.

This will destroy:
- All worker nodes
- All persistent volumes
- All network resources

This action CANNOT be undone.

Type the cluster name to confirm: prod-cluster-01
> prod-cluster-01
Deleting cluster...
```

#### Message Templates

Templates support variable substitution:

```yaml
message: "Delete {resource_type} '{resource_name}'?"
```

Variables available:
- Path parameters: `{cluster_id}`, `{pet_id}`
- Flag values: `{args.name}`, `{args.region}`
- Response data: `{response.body.name}`

#### Best Practices

1. **Destructive Only**: Only for DELETE or destructive operations
2. **Clear**: Explain what will be deleted/changed
3. **Templates**: Use variable substitution for clarity
4. **Skip Flag**: Always provide --yes for automation
5. **Type to Confirm**: For critical resources, require typing name
6. **Warning**: Use visual warnings (⚠️, colors)

#### Common Pitfalls

- ❌ Confirmation on non-destructive operations
- ❌ Generic "Are you sure?" messages
- ❌ No --yes flag for CI/CD automation
- ❌ Confirmation after side effects occur
- ✅ Test with --yes flag
- ✅ Document automation usage
- ✅ Use colors for warnings

---

## Async Extensions

### x-cli-async

**Location**: Operation object
**Type**: Object
**Purpose**: Configure async operation handling and status polling

#### Schema

```yaml
x-cli-async:
  enabled: boolean                 # Enable async handling
  status-field: string             # Field in response with status
  status-endpoint: string          # Endpoint to poll (supports templates)
  terminal-states: [string]        # States indicating completion
  polling:
    interval: integer              # Seconds between polls
    timeout: integer               # Total timeout in seconds
    backoff:
      enabled: boolean             # Enable exponential backoff
      multiplier: number           # Backoff multiplier
      max-interval: integer        # Max interval in seconds
  progress:
    show: boolean                  # Show progress indicator
    format: string                 # Progress message template
```

#### Simple Example

```yaml
post:
  operationId: createCluster
  responses:
    '201':
      description: Cluster creation initiated

  x-cli-async:
    enabled: true
    status-field: "state"
    status-endpoint: "/api/v1/clusters/{id}"
    terminal-states: [ready, error]
    polling:
      interval: 30
      timeout: 3600
```

Output:
```
$ myapi create cluster --name test
Creating cluster...
Status: pending (0:30)
Status: installing (1:00)
Status: configuring (1:30)
Status: ready ✓ (2:15)
```

#### Advanced Example

```yaml
delete:
  operationId: deleteCluster
  responses:
    '202':
      description: Deletion initiated

  x-cli-async:
    enabled: true
    status-field: "state"
    status-endpoint: "/api/v1/clusters/{cluster_id}"
    terminal-states: [deleted, error, not_found]

    polling:
      interval: 10                 # Start with 10s
      timeout: 3600                # 1 hour max
      backoff:
        enabled: true              # Exponential backoff
        multiplier: 1.5            # 10s, 15s, 22s, 33s...
        max-interval: 60           # Cap at 60s

    progress:
      show: true
      format: "Deleting cluster... {state} ({elapsed}s elapsed, {remaining}s remaining)"
```

Output:
```
$ myapi delete cluster --cluster-id prod-01
Deleting cluster... draining (10s elapsed, ~590s remaining)
Deleting cluster... terminating (25s elapsed, ~575s remaining)
Deleting cluster... cleaning (47s elapsed, ~553s remaining)
Deleting cluster... deleted ✓ (85s elapsed)
```

#### Advanced: Multi-Stage Operations

```yaml
post:
  operationId: upgradeCluster

  x-cli-async:
    enabled: true
    status-field: "upgrade.phase"
    status-endpoint: "/api/v1/clusters/{id}/upgrade-status"
    terminal-states: [completed, failed, rolled_back]

    polling:
      interval: 15
      timeout: 7200              # 2 hours
      backoff:
        enabled: true
        multiplier: 1.2
        max-interval: 120

    progress:
      show: true
      format: |
        Upgrade Phase: {upgrade.phase}
        Progress: {upgrade.progress}% ({upgrade.current_node}/{upgrade.total_nodes} nodes)
        Current: {upgrade.current_version} → {upgrade.target_version}
        Elapsed: {elapsed}s
```

#### Best Practices

1. **Intervals**: Start with reasonable intervals (10-30s)
2. **Backoff**: Use exponential backoff for long operations
3. **Timeouts**: Set realistic timeouts based on operation
4. **Terminal States**: Include all possible end states (success, error)
5. **Progress**: Show meaningful progress indicators
6. **Errors**: Handle terminal error states gracefully
7. **Background**: Offer option to run in background

#### Common Pitfalls

- ❌ Polling too frequently (causes API throttling)
- ❌ No timeout (infinite polling)
- ❌ Missing error terminal states
- ❌ Generic progress messages
- ❌ No backoff for long operations
- ✅ Test with slow operations
- ✅ Handle network failures gracefully
- ✅ Allow cancellation (Ctrl+C)

---

## Output Extensions

### x-cli-output

**Location**: Operation object
**Type**: Object
**Purpose**: Configure output formatting and display

#### Schema

```yaml
x-cli-output:
  format: string                   # Override default format
  success-message: string          # Message on success (template)
  error-message: string            # Message on error (template)
  template: string                 # Custom output template

  table:
    columns:
      - field: string              # JSON path to field
        header: string             # Column header
        width: integer             # Fixed width
        align: string              # left, right, center
        transform: string          # uppercase, lowercase, title
        format: string             # Format string (%.2f)
        prefix: string             # Prefix ($, #)
        suffix: string             # Suffix (%, GB)
        color-map:                 # State-based colors
          <value>: string          # Color name
    sort-by: string                # Default sort field
    sort-order: string             # asc, desc

  json:
    pretty: boolean                # Pretty print
    color: boolean                 # Syntax highlighting

  csv:
    delimiter: string              # Field delimiter
    header: boolean                # Include header row
```

#### Simple Example

```yaml
get:
  operationId: listPets

  x-cli-output:
    table:
      columns:
        - field: id
          header: ID
          width: 10
        - field: name
          header: NAME
          width: 20
        - field: status
          header: STATUS
          width: 12
```

Output:
```
$ myapi list pets
ID          NAME                  STATUS
1           Fluffy                available
2           Rex                   adopted
3           Tweety                pending
```

#### Advanced Example

```yaml
get:
  operationId: listClusters

  x-cli-output:
    table:
      columns:
        - field: id
          header: ID
          width: 25
          align: left

        - field: name
          header: NAME
          width: 30

        - field: state
          header: STATE
          width: 15
          transform: uppercase
          color-map:
            ready: green
            pending: yellow
            error: red
            installing: blue

        - field: nodes.count
          header: NODES
          width: 8
          align: right

        - field: region
          header: REGION
          width: 15

        - field: created_at
          header: AGE
          width: 12
          transform: relative_time

        - field: cost.monthly
          header: COST/MO
          width: 12
          align: right
          prefix: "$"
          format: "%.2f"

      sort-by: created_at
      sort-order: desc

    json:
      pretty: true
      color: true

    csv:
      delimiter: ","
      header: true
```

Output (table):
```
$ myapi list clusters
ID                        NAME                            STATE           NODES    REGION          AGE           COST/MO
cluster-abc123            prod-cluster-01                 READY               5    us-east-1       2d            $450.00
cluster-def456            staging-cluster                 INSTALLING          3    us-west-2       1h            $270.00
cluster-ghi789            dev-cluster                     READY               2    eu-west-1       5d            $180.00
```

Output (json):
```
$ myapi list clusters --output json
[
  {
    "id": "cluster-abc123",
    "name": "prod-cluster-01",
    "state": "ready",
    "nodes": {
      "count": 5
    },
    "region": "us-east-1",
    "created_at": "2024-11-21T10:00:00Z",
    "cost": {
      "monthly": 450.00
    }
  }
]
```

#### Template Output

```yaml
get:
  operationId: getPet

  x-cli-output:
    format: text
    template: |
      Pet Details:
      ------------
      ID:       {{.id}}
      Name:     {{.name}}
      Category: {{.category.name}}
      Status:   {{.status}}
      Age:      {{.age}} years
      Price:    ${{.price}}
      Created:  {{.createdAt}}
```

Output:
```
$ myapi get pet --pet-id 123
Pet Details:
------------
ID:       123
Name:     Fluffy
Category: cat
Status:   available
Age:      3 years
Price:    $150.00
Created:  2024-11-01T10:00:00Z
```

#### Success/Error Messages

```yaml
post:
  operationId: createPet

  x-cli-output:
    success-message: "Pet '{name}' created successfully with ID {id}"
    error-message: "Failed to create pet: {error.message}"
```

Output:
```
$ myapi create pet --name Fluffy
Pet 'Fluffy' created successfully with ID 123
```

#### Best Practices

1. **Tables**: Use for list operations with multiple columns
2. **JSON**: Always support JSON for scripting
3. **Templates**: Use for single-resource display
4. **Colors**: Use color-map for status fields
5. **Alignment**: Right-align numbers, left-align text
6. **Width**: Set reasonable column widths
7. **Messages**: Use templates in success/error messages

#### Column Transforms

| Transform | Example Input | Output |
|-----------|--------------|--------|
| `uppercase` | `ready` | `READY` |
| `lowercase` | `READY` | `ready` |
| `title` | `cluster ready` | `Cluster Ready` |
| `relative_time` | `2024-11-21T10:00:00Z` | `2d ago` |

#### Common Pitfalls

- ❌ No JSON output (breaks scripting)
- ❌ Fixed-width tables (breaks with long values)
- ❌ Too many columns (unreadable)
- ❌ No color in status columns
- ❌ Generic success messages
- ✅ Test with empty results
- ✅ Test with very long values
- ✅ Support --output flag

---

## Workflow Extensions

### x-cli-workflow

**Location**: Operation object
**Type**: Object
**Purpose**: Define multi-step workflows with dependencies

#### Schema

```yaml
x-cli-workflow:
  description: string              # Workflow description

  steps:
    - id: string                   # Step ID (for references)
      description: string          # Status message

      # API call
      request:
        method: string             # HTTP method
        url: string                # URL (supports templates)
        headers: {}                # Request headers
        body: {}                   # Request body
        query: {}                  # Query parameters

      # Plugin call
      plugin-call:
        command: string            # External command
        args: [string]             # Command arguments
        stdin: string              # Stdin content
        env: {}                    # Environment variables

      condition: string            # Execute if condition true
      output-var: string           # Store response in variable

      validation:
        condition: string          # Validation expression
        error-message: string      # Error message

      # Polling
      polling:
        interval: integer
        timeout: integer
        terminal-condition: string

      # Looping
      foreach: string              # Iterate over array
      as: string                   # Variable name

  rollback:
    enabled: boolean               # Enable rollback on failure
    steps: [...]                   # Rollback steps

  output:
    format: string                 # Output format
    transform: string              # Transform expression
```

#### Simple Example

```yaml
post:
  operationId: deployApp

  x-cli-workflow:
    description: "Application deployment workflow"

    steps:
      - id: build
        description: "Building application..."
        request:
          method: POST
          url: "{base_url}/builds"
          body:
            app_id: "{args.app_id}"
        output-var: build

      - id: deploy
        description: "Deploying to production..."
        request:
          method: POST
          url: "{base_url}/deployments"
          body:
            build_id: "{build.body.id}"
        condition: "build.status == 201"
        output-var: deployment

      - id: verify
        description: "Verifying deployment..."
        request:
          method: GET
          url: "{base_url}/deployments/{deployment.body.id}/health"
        condition: "deployment.status == 201"
```

#### Advanced Example: Pet Adoption Workflow

```yaml
post:
  operationId: adoptPet

  x-cli-workflow:
    description: "Complete pet adoption workflow with validation"

    steps:
      # Step 1: Verify pet availability
      - id: check-pet
        description: "Checking pet availability..."
        request:
          method: GET
          url: "{base_url}/pets/{args.pet_id}"
        output-var: pet
        validation:
          condition: "pet.body.status == 'available'"
          error-message: "Pet is not available (status: {pet.body.status})"

      # Step 2: Verify user eligibility
      - id: check-user
        description: "Verifying user eligibility..."
        request:
          method: GET
          url: "{base_url}/users/{args.user_id}"
        output-var: user
        validation:
          condition: "user.status == 200 && user.body.verified == true"
          error-message: "User not found or not verified"

      # Step 3: Run background check (plugin)
      - id: background-check
        description: "Running background check..."
        plugin-call:
          command: "check-service"
          args:
            - "--user-id"
            - "{args.user_id}"
            - "--type"
            - "adoption"
          env:
            CHECK_API_KEY: "{env.CHECK_API_KEY}"
        condition: "check-pet.valid && check-user.valid"
        output-var: bgcheck

      # Step 4: Create adoption order
      - id: create-order
        description: "Creating adoption order..."
        request:
          method: POST
          url: "{base_url}/orders"
          body:
            pet_id: "{args.pet_id}"
            user_id: "{args.user_id}"
            type: "adoption"
            background_check_id: "{bgcheck.stdout}"
        condition: "bgcheck.exit_code == 0"
        output-var: order

      # Step 5: Update pet status
      - id: update-pet
        description: "Updating pet status to adopted..."
        request:
          method: PUT
          url: "{base_url}/pets/{args.pet_id}"
          body:
            status: "adopted"
            owner_id: "{args.user_id}"
        condition: "create-order.status == 201"

      # Step 6: Send notifications
      - id: notify-user
        description: "Sending confirmation email..."
        request:
          method: POST
          url: "{base_url}/notifications"
          body:
            user_id: "{args.user_id}"
            template: "adoption_confirmed"
            data:
              pet_name: "{pet.body.name}"
              order_id: "{order.body.id}"
        condition: "update-pet.status == 200"

      # Step 7: Wait for payment
      - id: wait-payment
        description: "Waiting for payment confirmation..."
        request:
          method: GET
          url: "{base_url}/orders/{order.body.id}"
        polling:
          interval: 10
          timeout: 300
          terminal-condition: "response.body.payment_status in ['paid', 'failed']"
        output-var: final_order

    # Rollback on failure
    rollback:
      enabled: true
      steps:
        - description: "Reverting pet status..."
          request:
            method: PUT
            url: "{base_url}/pets/{args.pet_id}"
            body:
              status: "available"

        - description: "Canceling order..."
          request:
            method: DELETE
            url: "{base_url}/orders/{order.body.id}"

    # Transform output
    output:
      format: json
      transform: |
        {
          "adoption_id": order.body.id,
          "pet": {
            "id": pet.body.id,
            "name": pet.body.name,
            "category": pet.body.category.name
          },
          "adopter": {
            "id": user.body.id,
            "name": user.body.firstName + " " + user.body.lastName,
            "email": user.body.email
          },
          "payment_status": final_order.body.payment_status,
          "message": final_order.body.payment_status == "paid"
            ? "Adoption completed successfully!"
            : "Adoption failed - payment not received"
        }
```

#### Workflow Variables

Variables available in templates:
- `{args.name}` - CLI arguments
- `{env.VAR}` - Environment variables
- `{step_id.body.field}` - Response bodies
- `{step_id.status}` - HTTP status codes
- `{step_id.headers.name}` - Response headers
- `{step_id.valid}` - Validation result (boolean)
- `{step_id.exit_code}` - Plugin exit code
- `{step_id.stdout}` - Plugin stdout

#### Best Practices

1. **Validation**: Validate early, fail fast
2. **Idempotent**: Design steps to be retryable
3. **Rollback**: Always implement rollback for state changes
4. **Progress**: Show clear progress messages
5. **Conditions**: Use conditions for conditional execution
6. **Variables**: Store responses for later steps
7. **Timeout**: Set realistic polling timeouts

#### Common Pitfalls

- ❌ Missing rollback steps
- ❌ Non-idempotent operations
- ❌ Circular dependencies between steps
- ❌ Missing validation conditions
- ❌ Generic error messages
- ✅ Test failure scenarios
- ✅ Test rollback logic
- ✅ Handle partial failures

---

## Plugin Extensions

### x-cli-plugin

**Location**: Operation object
**Type**: Object
**Purpose**: Integrate external CLI tools (like AWS CLI, kubectl)

#### Schema

```yaml
x-cli-plugin:
  type: string                     # external, builtin, wasm
  command: string                  # External command name
  executable: string               # Path to executable
  required: boolean                # Is plugin required
  min-version: string              # Minimum version required
  install-hint: string             # Installation instructions

  operations:
    - description: string          # Step description

      # API call
      api-call:
        endpoint: string           # API endpoint
        method: string             # HTTP method
        output-var: string         # Store in variable

      # Plugin call
      plugin-call:
        command: string            # Plugin command
        args: [string]             # Arguments (supports templates)
        stdin: string              # Stdin content
        env: {}                    # Environment variables
        capture-output: boolean    # Capture stdout
```

#### Simple Example

```yaml
post:
  operationId: backupDatabase

  x-cli-plugin:
    type: external
    command: aws
    required: true
    min-version: "2.0.0"
    install-hint: "Install AWS CLI: https://aws.amazon.com/cli/"

    operations:
      - description: "Uploading to S3..."
        plugin-call:
          command: aws
          args:
            - "s3"
            - "cp"
            - "/tmp/backup.sql"
            - "s3://{args.bucket}/backups/{args.filename}"
          env:
            AWS_REGION: "{args.region}"
```

Usage:
```
$ myapi backup database --bucket my-backups --filename backup.sql
✓ Checking aws CLI... (v2.13.0)
Uploading to S3... ✓
Backup completed successfully
```

#### Advanced Example

```yaml
post:
  operationId: backupPetData

  x-cli-plugin:
    type: external
    command: aws
    required: true
    min-version: "2.0.0"
    install-hint: |
      AWS CLI not found. Install it:
      macOS: brew install awscli
      Linux: https://aws.amazon.com/cli/

    operations:
      # Step 1: Fetch data from API
      - description: "Fetching pet data..."
        api-call:
          endpoint: "/pets/{petId}"
          method: GET
          output-var: pet_data

      # Step 2: Write to temp file
      - description: "Preparing backup file..."
        plugin-call:
          command: "sh"
          args:
            - "-c"
            - "echo '{vars.pet_data}' > /tmp/pet-{petId}.json"

      # Step 3: Upload to S3
      - description: "Uploading to S3..."
        plugin-call:
          command: "aws"
          args:
            - "s3"
            - "cp"
            - "/tmp/pet-{petId}.json"
            - "s3://{args.bucket}/pets/{petId}.json"
            - "--metadata"
            - "pet-id={petId},backup-date={date}"
          env:
            AWS_REGION: "us-east-1"
            AWS_PROFILE: "{env.AWS_PROFILE}"
          capture-output: true

      # Step 4: Verify backup
      - description: "Verifying backup..."
        plugin-call:
          command: "aws"
          args:
            - "s3"
            - "ls"
            - "s3://{args.bucket}/pets/{petId}.json"
          capture-output: true

      # Step 5: Clean up temp file
      - description: "Cleaning up..."
        plugin-call:
          command: "rm"
          args:
            - "/tmp/pet-{petId}.json"
```

#### Plugin Types

| Type | Description | Example |
|------|-------------|---------|
| `external` | External CLI tool | `aws`, `kubectl`, `git` |
| `builtin` | Built-in plugin | Internal functions |
| `wasm` | WebAssembly plugin | Sandboxed plugins |

#### Best Practices

1. **Version Check**: Always specify min-version
2. **Install Hints**: Provide OS-specific install instructions
3. **Required**: Mark as required or provide fallback
4. **Error Handling**: Capture and display plugin errors
5. **Environment**: Pass necessary env vars
6. **Security**: Validate plugin inputs
7. **Capture**: Capture output when needed for verification

#### Common Pitfalls

- ❌ No version requirement (breaks with old versions)
- ❌ Generic install hints
- ❌ Not checking if plugin is installed
- ❌ Hardcoding paths (use PATH lookup)
- ❌ Not handling plugin failures
- ✅ Test with plugin not installed
- ✅ Test with old plugin versions
- ✅ Document required plugins

---

## File Extensions

### x-cli-file-input

**Location**: Parameter object
**Type**: Object
**Purpose**: Handle file uploads with validation

#### Schema

```yaml
x-cli-file-input:
  parameter: string                # Parameter name
  accepts: [string]                # Accepted file extensions
  max-size: integer                # Max size in bytes
  description: string              # Help text
  encoding: string                 # base64, multipart
  multiple: boolean                # Accept multiple files
```

#### Simple Example

```yaml
post:
  operationId: uploadPetPhoto
  parameters:
    - name: photo
      in: query
      schema:
        type: string
        format: binary
      x-cli-file-input:
        accepts: [".jpg", ".png", ".gif"]
        max-size: 5242880            # 5 MB
        description: "Pet photo (JPG, PNG, or GIF)"
```

Usage:
```
$ myapi upload pet-photo --pet-id 123 --photo ~/fluffy.jpg
Uploading pet-photo... ✓ (1.2 MB)
Photo uploaded successfully
```

#### Advanced Example

```yaml
post:
  operationId: importPets
  parameters:
    - name: file
      in: query
      schema:
        type: string
      x-cli-file-input:
        accepts: [".csv", ".json", ".xlsx"]
        max-size: 104857600          # 100 MB
        description: "Pet data file"
        encoding: multipart

        validation:
          # Validate CSV headers
          csv-headers: ["name", "category", "age", "price"]
          # Validate JSON schema
          json-schema:
            type: array
            items:
              type: object
              required: [name, category]

    - name: photos
      in: query
      schema:
        type: array
        items:
          type: string
          format: binary
      x-cli-file-input:
        accepts: [".jpg", ".png"]
        max-size: 10485760           # 10 MB per file
        description: "Pet photos"
        multiple: true
        encoding: multipart
```

Usage:
```
$ myapi import pets --file pets.csv --photos photo1.jpg --photos photo2.jpg
Validating pets.csv... ✓
Uploading pets.csv... ✓ (2.5 MB)
Uploading photo1.jpg... ✓ (1.2 MB)
Uploading photo2.jpg... ✓ (1.5 MB)
Imported 150 pets successfully
```

#### Best Practices

1. **Validation**: Validate file type and size before upload
2. **Extensions**: Be specific about accepted formats
3. **Size Limits**: Set reasonable max sizes
4. **Progress**: Show upload progress for large files
5. **Multiple**: Support multiple files when needed
6. **Errors**: Provide clear error messages
7. **Encoding**: Use multipart for binary files

#### Common Pitfalls

- ❌ No file size limits (OOM errors)
- ❌ No file type validation (security risk)
- ❌ No progress indicator for large files
- ❌ Not handling file not found errors
- ❌ Hardcoding file paths
- ✅ Test with various file sizes
- ✅ Test with wrong file types
- ✅ Handle network failures gracefully

---

## Monitoring Extensions

### x-cli-watch

**Location**: Operation object
**Type**: Object
**Purpose**: Real-time monitoring via SSE, WebSocket, or polling

#### Schema

```yaml
x-cli-watch:
  enabled: boolean                 # Enable watch mode
  type: string                     # sse, websocket, polling
  endpoint: string                 # Watch endpoint
  events: [string]                 # Event types to watch
  interval: integer                # Polling interval (for type=polling)

  format:
    template: string               # Output template
    colors:                        # Event type colors
      <event-type>: string

  exit-on:
    - event: string                # Event type
      condition: string            # Exit condition
      message: string              # Exit message

  alert-on-change:
    - field: string                # Field to watch
      message: string              # Alert message template

  reconnect:
    enabled: boolean               # Auto-reconnect
    max-retries: integer           # Max reconnection attempts
    backoff: string                # linear, exponential
```

#### Simple Example (Polling)

```yaml
get:
  operationId: getPet

  x-cli-watch:
    enabled: true
    type: polling
    interval: 5
    fields: [status, updatedAt]
    alert-on-change:
      - field: status
        message: "Status changed: {old_value} → {new_value}"
```

Usage:
```
$ myapi get pet --pet-id 123 --watch
[10:00:00] Status: available
[10:00:05] Status: available
[10:00:10] Status changed: available → pending
[10:00:15] Status: pending
^C
```

#### Advanced Example (SSE)

```yaml
get:
  operationId: streamPetStatus

  x-cli-watch:
    enabled: true
    type: sse
    endpoint: "/pets/{petId}/status/stream"
    events:
      - status-change
      - location-update
      - health-check
      - alert

    format:
      template: "[{timestamp}] {event}: {data.message}"
      colors:
        status-change: yellow
        location-update: blue
        health-check: green
        alert: red

    exit-on:
      - event: status-change
        condition: "data.status == 'adopted'"
        message: "Pet has been adopted!"

      - event: alert
        condition: "data.severity == 'critical'"
        message: "Critical alert received"

    reconnect:
      enabled: true
      max-retries: 10
      backoff: exponential
```

Output:
```
$ myapi watch pet --pet-id 123
[10:00:00] health-check: Pet is healthy
[10:00:15] location-update: Moved to kennel B-12
[10:00:30] health-check: Pet is healthy
[10:00:45] status-change: Status changed to pending
[10:01:00] location-update: Moved to adoption center
[10:01:15] status-change: Status changed to adopted
Pet has been adopted!
```

#### Advanced Example (WebSocket)

```yaml
get:
  operationId: monitorCluster

  x-cli-watch:
    enabled: true
    type: websocket
    endpoint: "wss://api.example.com/clusters/{cluster_id}/watch"
    events:
      - node-status
      - pod-event
      - warning
      - error

    format:
      template: |
        [{timestamp}] {event}
        Node: {data.node_id}
        Status: {data.status}
        Message: {data.message}
      colors:
        node-status: blue
        pod-event: cyan
        warning: yellow
        error: red

    alert-on-change:
      - field: data.health_status
        message: "⚠️  Health status changed: {old_value} → {new_value}"

    exit-on:
      - event: error
        condition: "data.fatal == true"
        message: "Fatal error occurred: {data.message}"

    reconnect:
      enabled: true
      max-retries: 5
      backoff: exponential
```

#### Best Practices

1. **Type**: Use SSE for server-sent events, polling for simple cases
2. **Intervals**: Don't poll too frequently (respect rate limits)
3. **Exit**: Provide exit conditions to prevent infinite watching
4. **Reconnect**: Always enable auto-reconnect
5. **Colors**: Use colors to distinguish event types
6. **Alerts**: Alert on important changes
7. **Cancellation**: Support Ctrl+C gracefully

#### Common Pitfalls

- ❌ Polling too frequently (< 1s)
- ❌ No exit conditions (infinite watching)
- ❌ No reconnection logic
- ❌ Not handling connection drops
- ❌ Cluttered output (too much info)
- ✅ Test with connection failures
- ✅ Test exit conditions
- ✅ Respect API rate limits

---

## Deprecation Extensions

### x-cli-deprecation

**Location**: Operation object
**Type**: Object
**Purpose**: Mark operations as deprecated with migration paths

#### Schema

```yaml
x-cli-deprecation:
  deprecated: boolean              # Is deprecated
  sunset-date: string              # Removal date (ISO 8601)
  alternative: string              # Replacement command
  message: string                  # Deprecation message
  docs-url: string                 # Documentation URL
  level: string                    # warning, error
```

#### Simple Example

```yaml
delete:
  operationId: deletePet

  x-cli-deprecation:
    deprecated: true
    sunset-date: "2025-12-31"
    alternative: "archive pet"
    message: "Direct deletion is deprecated. Use 'archive pet' instead."
```

Output:
```
$ myapi delete pet --pet-id 123
⚠️  DEPRECATED: This command will be removed on 2025-12-31
    Use 'myapi archive pet' instead
    Direct deletion is deprecated. Use 'archive pet' instead.

? Continue anyway? (y/N)
```

#### Advanced Example

```yaml
paths:
  /v1/clusters:
    get:
      operationId: listClustersV1
      x-cli-command: "list clusters v1"

      x-cli-deprecation:
        deprecated: true
        sunset-date: "2025-06-01"
        alternative: "list clusters"
        level: warning
        message: |
          API v1 is deprecated and will be removed on June 1, 2025.

          Migration steps:
          1. Update to v2 API: 'myapi list clusters'
          2. Review breaking changes in docs
          3. Update automation scripts
        docs-url: "https://docs.example.com/migration/v1-to-v2"

  /v2/clusters:
    get:
      operationId: listClustersV2
      x-cli-command: "list clusters"
      summary: List clusters (v2 API)
```

Output:
```
$ myapi list clusters v1
╔══════════════════════════════════════════════════════════════════╗
║ ⚠️  DEPRECATION WARNING                                          ║
╚══════════════════════════════════════════════════════════════════╝

This command is deprecated and will be removed on 2025-06-01

API v1 is deprecated and will be removed on June 1, 2025.

Migration steps:
1. Update to v2 API: 'myapi list clusters'
2. Review breaking changes in docs
3. Update automation scripts

Documentation: https://docs.example.com/migration/v1-to-v2
Alternative: myapi list clusters

? Continue anyway? (y/N)
```

#### Deprecation Levels

| Level | Behavior | Exit Code |
|-------|----------|-----------|
| `warning` | Show warning, continue | 0 |
| `error` | Show error, require --force | 1 (without --force) |

#### Best Practices

1. **Notice**: Provide 6+ months deprecation period
2. **Alternative**: Always provide replacement command
3. **Migration**: Include migration steps
4. **Docs**: Link to detailed migration guide
5. **Level**: Start with warning, escalate to error near sunset
6. **Automation**: Support --force flag for automation
7. **Visibility**: Log deprecation usage for tracking

#### Common Pitfalls

- ❌ Short deprecation periods (< 3 months)
- ❌ No migration path provided
- ❌ Breaking automation (no --force flag)
- ❌ Generic deprecation messages
- ❌ No sunset date
- ✅ Announce deprecations in release notes
- ✅ Track deprecation usage
- ✅ Test migration path

---

## Secret Extensions

### x-cli-secret

**Location**: Parameter object
**Type**: Object
**Purpose**: Handle sensitive data securely

#### Schema

```yaml
x-cli-secret:
  parameter: string                # Parameter name
  masked: boolean                  # Mask in logs/output
  storage: string                  # env, keyring, prompt, file
  env-var: string                  # Environment variable name
  keyring-key: string              # Keyring key name
  prompt-message: string           # Prompt for value
  validation: string               # Validation expression
```

#### Simple Example

```yaml
post:
  operationId: createAPIKey
  parameters:
    - name: password
      in: query
      schema:
        type: string
      x-cli-secret:
        masked: true
        storage: prompt
        prompt-message: "Enter password:"
```

Usage:
```
$ myapi create-api-key
Enter password: ********
API key created successfully
```

#### Advanced Example

```yaml
post:
  operationId: configureDatabase

  x-cli-flags:
    - name: db-password
      source: database.password
      flag: "--db-password"
      type: string
      required: true

      x-cli-secret:
        masked: true
        storage: keyring
        env-var: "DB_PASSWORD"
        keyring-key: "myapi/database/password"
        prompt-message: "Database password:"
        validation: "^(?=.*[a-z])(?=.*[A-Z])(?=.*\\d).{12,}$"
        validation-message: "Password must be 12+ chars with upper, lower, and digit"

    - name: api-key
      source: api.key
      flag: "--api-key"
      type: string
      required: true

      x-cli-secret:
        masked: true
        storage: env
        env-var: "MYAPI_KEY"
        prompt-message: "API key:"
```

Secret resolution order:
1. Command-line flag (if provided)
2. Environment variable
3. Keyring
4. Prompt user

Usage:
```
$ export DB_PASSWORD=MySecurePass123
$ export MYAPI_KEY=sk-abc123...
$ myapi configure database
✓ Database password loaded from environment
✓ API key loaded from environment
Database configured successfully
```

Or interactive:
```
$ myapi configure database
Database password: ********
API key: ********
✓ Storing database password in keyring
Database configured successfully
```

#### Storage Types

| Type | Description | Security | Use Case |
|------|-------------|----------|----------|
| `env` | Environment variable | Medium | CI/CD, containers |
| `keyring` | OS keyring | High | Local development |
| `prompt` | Interactive prompt | High | One-time operations |
| `file` | Encrypted file | Medium | Legacy systems |

#### Best Practices

1. **Masking**: Always mask secrets in logs and output
2. **Storage**: Prefer keyring > env > prompt > file
3. **Validation**: Validate secret format/strength
4. **Fallback**: Support multiple storage methods
5. **Rotation**: Support secret rotation
6. **Audit**: Log secret access (but not values)
7. **Environment**: Never log env vars

#### Common Pitfalls

- ❌ Secrets in command history
- ❌ Secrets in error messages
- ❌ Secrets in debug logs
- ❌ World-readable secret files
- ❌ No validation (weak passwords)
- ✅ Test secret masking
- ✅ Document secret management
- ✅ Audit secret access

---

## Context Extensions

### x-cli-context

**Location**: Operation object or root level
**Type**: Object
**Purpose**: Multi-environment configuration (dev, staging, prod)

#### Schema

```yaml
x-cli-context:
  enabled: boolean                 # Enable context switching
  default: string                  # Default context name

  contexts:
    <context-name>:
      base-url: string             # API base URL
      auth:
        type: string               # Auth type
        config: {}                 # Auth configuration
      env: {}                      # Environment variables
      timeout: integer             # Request timeout

  switch-command: string           # Command to switch context
  current-command: string          # Command to show current
  storage: string                  # file, memory
  storage-path: string             # Context file path
```

#### Simple Example

```yaml
x-cli-config:
  name: myapi
  features:
    context-switching: true

x-cli-context:
  enabled: true
  default: production

  contexts:
    development:
      base-url: "http://localhost:8080"
      auth:
        type: none

    production:
      base-url: "https://api.example.com"
      auth:
        type: oauth2
```

Usage:
```
$ myapi context list
* production (https://api.example.com)
  development (http://localhost:8080)

$ myapi context use development
Switched to context: development

$ myapi context current
development (http://localhost:8080)
```

#### Advanced Example

```yaml
x-cli-context:
  enabled: true
  default: production
  switch-command: "use context"
  current-command: "current context"
  storage: file
  storage-path: "~/.myapi/contexts.yaml"

  contexts:
    development:
      base-url: "http://localhost:8080"
      auth:
        type: none
      env:
        LOG_LEVEL: "debug"
        CACHE_ENABLED: "false"
      timeout: 30

    staging:
      base-url: "https://staging-api.example.com"
      auth:
        type: oauth2
        config:
          client-id: "staging-client"
          scopes: ["read:all", "write:all"]
      env:
        LOG_LEVEL: "info"
        CACHE_ENABLED: "true"
      timeout: 60

    production:
      base-url: "https://api.example.com"
      auth:
        type: oauth2
        config:
          client-id: "prod-client"
          scopes: ["read:all", "write:all", "admin"]
      env:
        LOG_LEVEL: "warn"
        CACHE_ENABLED: "true"
        RATE_LIMIT: "1000"
      timeout: 60

    dr-site:
      base-url: "https://dr-api.example.com"
      auth:
        type: oauth2
        config:
          client-id: "dr-client"
          scopes: ["read:all"]
      env:
        LOG_LEVEL: "info"
        READ_ONLY: "true"
      timeout: 90
```

Usage:
```
$ myapi use context staging
Switched to context: staging (https://staging-api.example.com)

$ myapi current context
staging
  URL: https://staging-api.example.com
  Auth: OAuth2 (staging-client)
  Timeout: 60s
  Environment:
    LOG_LEVEL=info
    CACHE_ENABLED=true

$ myapi list pets
# Uses staging context

$ myapi use context production
Switched to context: production (https://api.example.com)

$ myapi list pets
# Uses production context
```

#### Context File Format

Stored at `~/.myapi/contexts.yaml`:

```yaml
current-context: production
contexts:
  development:
    base-url: http://localhost:8080
    auth:
      type: none
  production:
    base-url: https://api.example.com
    auth:
      type: oauth2
      token: <encrypted>
```

#### Best Practices

1. **Default**: Always provide sensible default context
2. **Safety**: Require confirmation for production context
3. **Indication**: Show current context in prompt/output
4. **Isolation**: Isolate auth tokens per context
5. **Validation**: Validate context configuration
6. **Commands**: Provide list, use, current commands
7. **Storage**: Store securely with proper permissions

#### Common Pitfalls

- ❌ No visual indication of current context
- ❌ Shared auth tokens across contexts
- ❌ Default to production (dangerous)
- ❌ No context validation
- ❌ World-readable context file
- ✅ Test context switching
- ✅ Document context setup
- ✅ Require confirmation for prod

---

## Migration Guide

### From Standard OpenAPI to CliForge

#### Step 1: Add Global Config

```yaml
# Before (standard OpenAPI)
openapi: 3.0.3
info:
  title: My API
  version: 1.0.0

# After (CliForge)
openapi: 3.0.3
info:
  title: My API
  version: 1.0.0

x-cli-config:
  name: myapi
  version: 1.0.0
  output:
    default-format: table
    supported-formats: [table, json, yaml]
  features:
    interactive-mode: true
```

#### Step 2: Add Command Names

```yaml
# Before
paths:
  /pets:
    get:
      operationId: listPets

# After
paths:
  /pets:
    get:
      operationId: listPets
      x-cli-command: "list pets"
      x-cli-aliases: ["ls pets", "pets"]
```

#### Step 3: Add CLI Flags

```yaml
# Before
post:
  operationId: createPet
  requestBody:
    content:
      application/json:
        schema:
          type: object
          properties:
            name:
              type: string

# After
post:
  operationId: createPet
  x-cli-command: "create pet"

  x-cli-flags:
    - name: name
      source: name
      flag: "--name"
      aliases: ["-n"]
      required: true
      type: string
      description: "Pet name"
```

#### Step 4: Add Interactive Mode

```yaml
post:
  operationId: createPet
  x-cli-command: "create pet"

  x-cli-flags: [...]

  # Add interactive prompts
  x-cli-interactive:
    enabled: true
    prompts:
      - parameter: name
        type: text
        message: "What is the pet's name?"
```

#### Step 5: Add Output Configuration

```yaml
get:
  operationId: listPets
  x-cli-command: "list pets"

  # Add table output
  x-cli-output:
    table:
      columns:
        - field: id
          header: ID
          width: 10
        - field: name
          header: NAME
          width: 20
        - field: status
          header: STATUS
          width: 12
```

### Progressive Enhancement Strategy

1. **Minimal**: Add `x-cli-config` and `x-cli-command` (basic CLI works)
2. **Improved**: Add `x-cli-flags` (better UX than JSON bodies)
3. **Interactive**: Add `x-cli-interactive` (user-friendly)
4. **Advanced**: Add async, workflows, plugins (power features)

---

## Extension Patterns

### Pattern: CRUD Operations

```yaml
paths:
  /resources:
    get:
      x-cli-command: "list resources"
      x-cli-output:
        table:
          columns: [...]

    post:
      x-cli-command: "create resource"
      x-cli-flags: [...]
      x-cli-interactive:
        enabled: true
        prompts: [...]

  /resources/{id}:
    get:
      x-cli-command: "get resource"

    put:
      x-cli-command: "update resource"
      x-cli-flags: [...]

    delete:
      x-cli-command: "delete resource"
      x-cli-confirmation:
        enabled: true
        message: "Delete resource {id}?"
```

### Pattern: Long-Running Operations

```yaml
post:
  x-cli-command: "create cluster"
  x-cli-flags: [...]

  # Preflight checks
  x-cli-preflight:
    - name: verify-quota
      description: "Checking quota..."
      endpoint: "/quota"
      method: GET
      required: true

  # Async handling
  x-cli-async:
    enabled: true
    status-field: "state"
    status-endpoint: "/clusters/{id}"
    terminal-states: [ready, error]
    polling:
      interval: 30
      timeout: 3600
      backoff:
        enabled: true
        multiplier: 1.5
        max-interval: 120
```

### Pattern: Multi-Step Workflows

```yaml
post:
  x-cli-command: "deploy app"

  x-cli-workflow:
    steps:
      - id: build
        description: "Building..."
        request:
          method: POST
          url: "/builds"

      - id: test
        description: "Testing..."
        request:
          method: POST
          url: "/tests"
        condition: "build.status == 201"

      - id: deploy
        description: "Deploying..."
        request:
          method: POST
          url: "/deployments"
        condition: "test.status == 200 && test.body.passed == true"

    rollback:
      enabled: true
      steps: [...]
```

### Pattern: External Tool Integration

```yaml
post:
  x-cli-command: "backup to s3"

  x-cli-plugin:
    type: external
    command: aws
    required: true

    operations:
      - description: "Fetching data..."
        api-call:
          endpoint: "/data"
          method: GET
          output-var: data

      - description: "Uploading to S3..."
        plugin-call:
          command: aws
          args: ["s3", "cp", "-", "s3://{bucket}/{key}"]
          stdin: "{vars.data}"
```

### Pattern: Real-Time Monitoring

```yaml
get:
  x-cli-command: "watch resource"

  x-cli-watch:
    enabled: true
    type: sse
    endpoint: "/resources/{id}/stream"
    events: [update, alert, error]

    exit-on:
      - event: error
        condition: "data.fatal == true"
        message: "Fatal error occurred"

    format:
      template: "[{timestamp}] {event}: {data.message}"
      colors:
        update: blue
        alert: yellow
        error: red
```

---

## Summary

CliForge provides 16 OpenAPI extensions that transform API specifications into rich CLI definitions:

1. **x-cli-config** - Global configuration and branding
2. **x-auth-config** - OAuth2 and authentication
3. **x-cli-command** - Command name mapping
4. **x-cli-aliases** - Command aliases
5. **x-cli-flags** - Request body to CLI flags
6. **x-cli-interactive** - Interactive prompts
7. **x-cli-preflight** - Pre-execution checks
8. **x-cli-confirmation** - User confirmation
9. **x-cli-async** - Async operation handling
10. **x-cli-output** - Output formatting
11. **x-cli-workflow** - Multi-step workflows
12. **x-cli-plugin** - External plugin integration
13. **x-cli-file-input** - File upload handling
14. **x-cli-watch** - Real-time monitoring
15. **x-cli-deprecation** - Deprecation warnings
16. **x-cli-secret** - Secret/sensitive data handling
17. **x-cli-context** - Multi-environment support

### Next Steps

1. Review [technical-specification.md](./technical-specification.md) for implementation details
2. See [petstore-api.yaml](https://github.com/wrale/alpha-omega/tree/main/examples/petstore/petstore-api.yaml) for complete examples
3. Explore [configuration-dsl.md](./configuration-dsl.md) for advanced configuration
4. Start with minimal extensions, add progressively
5. Test generated CLIs with real users

### References

- OpenAPI 3.0 Specification: https://spec.openapis.org/oas/v3.0.3
- CliForge Repository: https://github.com/CliForge/cliforge
- Example Specifications: https://github.com/wrale/alpha-omega/tree/main/examples
