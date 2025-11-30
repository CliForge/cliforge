# Building an Enterprise CLI with CliForge: A Complete Tutorial

## Table of Contents

1. [Introduction](#introduction)
2. [Architecture Overview](#architecture-overview)
3. [Designing the OpenAPI Spec](#designing-the-openapi-spec)
4. [x-cli-* Extensions Deep Dive](#x-cli-extensions-deep-dive)
5. [OAuth2 Authentication Patterns](#oauth2-authentication-patterns)
6. [Interactive Command Patterns](#interactive-command-patterns)
7. [Async Operations and Polling](#async-operations-and-polling)
8. [Nested Resource Commands](#nested-resource-commands)
9. [Output Formatting](#output-formatting)
10. [Testing Strategy](#testing-strategy)
11. [Best Practices](#best-practices)
12. [Case Study: ROSA CLI](#case-study-rosa-cli)

---

## Introduction

### Why Build CLIs from OpenAPI Specs?

Modern enterprises face a common challenge: maintaining consistency between their REST APIs and command-line interfaces. Traditional approaches require:

- **Duplicate code**: Writing both API clients and CLI commands manually
- **Sync overhead**: Keeping CLI flags in sync with API parameters
- **Documentation drift**: Maintaining separate docs for API and CLI
- **Testing burden**: Testing both API and CLI independently

CliForge solves this by making your OpenAPI specification the **single source of truth** for both your API and CLI.

### The Challenge of Enterprise CLIs

Enterprise CLIs like AWS CLI, Kubernetes `kubectl`, and Red Hat's `rosa` CLI share common requirements:

1. **Complex authentication** - OAuth2, SSO, token refresh, keyring storage
2. **Nested resources** - Clusters → Machine Pools → Nodes
3. **Interactive workflows** - Multi-step wizards with validation
4. **Async operations** - Long-running operations with status polling
5. **Multiple output formats** - Tables for humans, JSON/YAML for scripts
6. **Pre-flight validation** - Check permissions/quotas before operations
7. **Confirmation prompts** - Safety for destructive operations
8. **Contextual help** - Dynamic command discovery and autocomplete

Building all this from scratch takes months. Maintaining it takes years.

### How CliForge Solves It

CliForge generates production-ready CLIs from OpenAPI specs using custom `x-cli-*` extensions:

```yaml
paths:
  /api/v1/clusters:
    post:
      x-cli-command: "create cluster"
      x-cli-interactive:
        enabled: true
        prompts:
          - parameter: name
            type: text
            message: "Cluster name:"
      x-cli-preflight:
        - name: verify-credentials
          description: "Verifying AWS credentials..."
      x-cli-output:
        success-message: "Cluster '{name}' is being created"
```

This single specification generates:
- CLI command structure
- Flag parsing and validation
- Interactive prompts
- Pre-flight checks
- Output formatting
- Authentication flow
- Status polling

**The result**: Build in days what would normally take months.

---

## Architecture Overview

### The Single Source of Truth

```
┌─────────────────────────────────────────────────────────┐
│                   OpenAPI Specification                  │
│                  (with x-cli-* extensions)               │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                   CliForge Generator                     │
│  • Parses OpenAPI spec                                   │
│  • Extracts x-cli-* extensions                          │
│  • Generates command tree                               │
│  • Creates CLI binary                                    │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                   Generated CLI Binary                   │
│  ┌─────────────────┐  ┌─────────────────┐              │
│  │  Built-in       │  │  Generated      │              │
│  │  Commands       │  │  Commands       │              │
│  │                 │  │                 │              │
│  │  • login        │  │  • create       │              │
│  │  • logout       │  │  • list         │              │
│  │  • whoami       │  │  • describe     │              │
│  │  • token        │  │  • delete       │              │
│  │  • version      │  │  • edit         │              │
│  │  • completion   │  │  (from spec)    │              │
│  └─────────────────┘  └─────────────────┘              │
└─────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────┐
│                     Your REST API                        │
└─────────────────────────────────────────────────────────┘
```

### Component Diagram

```
┌─────────────────────────────────────────────────────────────┐
│                         CLI Layer                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Command    │  │   Flag       │  │   Help       │     │
│  │   Router     │  │   Parser     │  │   Generator  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                      Business Logic                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Interactive │  │   Workflow   │  │   Async      │     │
│  │  Prompts     │  │   Engine     │  │   Polling    │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            │
┌─────────────────────────────────────────────────────────────┐
│                      Platform Layer                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  OAuth2      │  │   Storage    │  │   Output     │     │
│  │  Auth        │  │   (Keyring)  │  │   Formatter  │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
```

### Generated vs Built-in Commands

**Built-in Commands** (provided by CliForge):
- `login`, `logout`, `whoami`, `token` - OAuth2 authentication
- `version` - CLI version information
- `completion` - Shell autocomplete scripts
- `config get/set` - Configuration management

**Generated Commands** (from your OpenAPI spec):
- All CRUD operations defined in your spec
- Nested resource commands
- Custom workflows
- API-specific operations

This separation ensures authentication and core functionality works consistently across all CLIs while letting you focus on your domain-specific operations.

---

## Designing the OpenAPI Spec

### Path Structure for CLI Commands

The way you structure API paths directly affects CLI command ergonomics. Here are proven patterns:

#### Pattern 1: Flat Resource Collections

```yaml
paths:
  /api/v1/clusters:
    get:
      x-cli-command: "list clusters"
    post:
      x-cli-command: "create cluster"

  /api/v1/clusters/{cluster_id}:
    get:
      x-cli-command: "describe cluster"
    patch:
      x-cli-command: "edit cluster"
    delete:
      x-cli-command: "delete cluster"
```

**Generates:**
```bash
mycli list clusters
mycli create cluster --name prod-cluster
mycli describe cluster abc-123
mycli edit cluster abc-123 --replicas 5
mycli delete cluster abc-123
```

#### Pattern 2: Nested Resources

```yaml
paths:
  /api/v1/clusters/{cluster_id}/machine_pools:
    get:
      x-cli-command: "list machinepools"
      x-cli-parent-resource: "cluster"
    post:
      x-cli-command: "create machinepool"
      x-cli-parent-resource: "cluster"

  /api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}:
    get:
      x-cli-command: "describe machinepool"
      x-cli-parent-resource: "cluster"
    delete:
      x-cli-command: "delete machinepool"
      x-cli-parent-resource: "cluster"
```

**Generates:**
```bash
mycli list machinepools --cluster prod-cluster
mycli create machinepool --cluster prod-cluster --name workers
mycli describe machinepool workers --cluster prod-cluster
mycli delete machinepool workers --cluster prod-cluster
```

The `x-cli-parent-resource` extension tells CliForge that this command operates on a subresource and needs the parent resource specified.

#### Pattern 3: Action Endpoints

```yaml
paths:
  /api/v1/clusters/{cluster_id}/upgrade:
    post:
      x-cli-command: "upgrade cluster"
      x-cli-flags:
        - name: version
          source: version
          flag: "--version"
          required: true
        - name: schedule-date
          source: schedule_date
          flag: "--schedule-date"
          description: "Schedule upgrade for specific date (YYYY-MM-DD)"
```

**Generates:**
```bash
mycli upgrade cluster prod-cluster --version 4.14.0
mycli upgrade cluster prod-cluster --version 4.14.0 --schedule-date 2025-12-15
```

### Using Tags for Command Grouping

OpenAPI tags become command categories in help text:

```yaml
tags:
  - name: Clusters
    description: Cluster lifecycle management
  - name: Machine Pools
    description: Worker node pool management
  - name: Identity Providers
    description: Authentication provider configuration

paths:
  /api/v1/clusters:
    post:
      tags:
        - Clusters
      x-cli-command: "create cluster"
```

**Generated help output:**
```
$ mycli --help

Available Commands:
  Clusters:
    create cluster     Create a new cluster
    list clusters      List all clusters
    describe cluster   Show cluster details

  Machine Pools:
    create machinepool  Add worker pool to cluster
    list machinepools   List cluster worker pools

  Identity Providers:
    create idp         Add identity provider
    list idps          List cluster identity providers
```

### Request Body to CLI Flags Mapping

CliForge automatically converts request body schemas to CLI flags. You can customize this with `x-cli-flags`:

```yaml
paths:
  /api/v1/clusters:
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                  description: Cluster name
                region:
                  type: string
                  description: AWS region
                multi_az:
                  type: boolean
                  description: Deploy to multiple availability zones
                compute:
                  type: object
                  properties:
                    replicas:
                      type: integer
                      description: Number of worker nodes
                    machine_type:
                      type: string
                      description: EC2 instance type

      x-cli-flags:
        - name: cluster-name
          source: name
          flag: "--cluster-name"
          aliases: ["-n"]
          required: true
          description: "Name of the cluster"
        - name: region
          source: region
          flag: "--region"
          required: true
          description: "AWS region"
        - name: multi-az
          source: multi_az
          flag: "--multi-az"
          type: boolean
          description: "Deploy to multiple availability zones"
        - name: replicas
          source: compute.replicas
          flag: "--replicas"
          type: integer
          default: 2
          description: "Number of worker nodes (2-100)"
        - name: machine-type
          source: compute.machine_type
          flag: "--machine-type"
          default: "m5.xlarge"
          description: "EC2 instance type for worker nodes"
```

**Key concepts:**
- `source`: JSON path in request body (supports nested paths with dot notation)
- `flag`: CLI flag name (use kebab-case)
- `aliases`: Short flags (e.g., `-n` for `--name`)
- `type`: Flag type (string, integer, boolean, array)
- `default`: Default value if flag not provided
- `required`: Whether flag is mandatory

**Generated command:**
```bash
mycli create cluster \
  --cluster-name prod-cluster \
  --region us-east-1 \
  --multi-az \
  --replicas 4 \
  --machine-type m5.2xlarge
```

### Validation in the Spec

Use JSON Schema validation keywords to ensure data quality:

```yaml
schema:
  type: object
  properties:
    name:
      type: string
      pattern: "^[a-z][a-z0-9-]{0,53}[a-z0-9]$"
      minLength: 2
      maxLength: 54
      description: "Lowercase alphanumeric with hyphens, 2-54 characters"

    replicas:
      type: integer
      minimum: 2
      maximum: 100
      description: "Number of worker nodes"

    region:
      type: string
      enum: ["us-east-1", "us-west-2", "eu-west-1"]
      description: "AWS region"
```

CliForge automatically validates input against these constraints and provides helpful error messages:

```bash
$ mycli create cluster --cluster-name INVALID_NAME
Error: Invalid value for --cluster-name: "INVALID_NAME"
  Must match pattern: ^[a-z][a-z0-9-]{0,53}[a-z0-9]$
  Must be lowercase alphanumeric with hyphens, 2-54 characters

$ mycli create cluster --cluster-name test --replicas 150
Error: Invalid value for --replicas: 150
  Must be between 2 and 100
```

---

## x-cli-* Extensions Deep Dive

### x-cli-interactive: Dynamic Prompts

Interactive mode transforms CLIs from command-line tools into user-friendly wizards. CliForge's `x-cli-interactive` extension enables rich prompts with validation:

```yaml
post:
  operationId: createCluster
  x-cli-command: "create cluster"

  x-cli-interactive:
    enabled: true
    prompts:
      # Text input with validation
      - parameter: name
        type: text
        message: "Cluster name:"
        validation: "^[a-z][a-z0-9-]{0,53}[a-z0-9]$"
        validation-message: "Name must be lowercase alphanumeric with hyphens (2-54 chars)"

      # Select from static options
      - parameter: region
        type: select
        message: "Select region:"
        source:
          endpoint: "/api/v1/regions"
          value-field: "id"
          display-field: "display_name"
          filter: "enabled == true"

      # Number input with range
      - parameter: replicas
        type: number
        message: "Number of worker nodes:"
        default: 2
        validation: "value >= 2 && value <= 100"
        validation-message: "Must be between 2 and 100"

      # Boolean confirmation
      - parameter: multi_az
        type: confirm
        message: "Deploy to multiple availability zones?"
        default: false

      # Password (hidden input)
      - parameter: admin_password
        type: password
        message: "Admin password:"
        validation: "len(value) >= 12"
        validation-message: "Password must be at least 12 characters"
```

**Prompt types:**
- `text` - Free-form text input
- `select` - Choose from a list
- `confirm` - Yes/no question
- `number` - Numeric input
- `password` - Hidden input (for secrets)

**Dynamic option loading:**

The `source.endpoint` feature is powerful - it fetches options from your API in real-time:

```yaml
- parameter: cluster_version
  type: select
  message: "Select OpenShift version:"
  source:
    endpoint: "/api/v1/versions"
    value-field: "id"           # Field to use as value
    display-field: "display_name"  # Field to show user
    filter: "enabled == true && channel_group == 'stable'"
```

If the API returns:
```json
[
  {"id": "4.13.0", "display_name": "4.13.0 (stable)", "enabled": true, "channel_group": "stable"},
  {"id": "4.14.0", "display_name": "4.14.0 (stable)", "enabled": true, "channel_group": "stable"},
  {"id": "4.15.0", "display_name": "4.15.0 (candidate)", "enabled": false, "channel_group": "candidate"}
]
```

The user sees:
```
? Select OpenShift version:
  4.13.0 (stable)
❯ 4.14.0 (stable)
```

**Interactive vs Non-Interactive:**

Users can run commands in either mode:

```bash
# Interactive mode - prompts for all inputs
$ mycli create cluster
? Cluster name: prod-cluster
? Select region: us-east-1
? Number of worker nodes: 4
? Deploy to multiple availability zones? Yes
Creating cluster...

# Non-interactive mode - all flags provided
$ mycli create cluster \
  --cluster-name prod-cluster \
  --region us-east-1 \
  --replicas 4 \
  --multi-az
Creating cluster...
```

The CLI automatically detects if input is a TTY and enables/disables interactive mode accordingly. This makes commands work in both interactive terminals and CI/CD pipelines.

### x-cli-preflight: Validation Before Execution

Pre-flight checks validate prerequisites before executing expensive operations:

```yaml
post:
  operationId: createCluster
  x-cli-command: "create cluster"

  x-cli-preflight:
    - name: verify-credentials
      description: "Verifying AWS credentials..."
      endpoint: "/api/v1/aws/credentials/verify"
      method: POST
      required: true
      skip-flag: "--skip-credential-check"

    - name: check-quota
      description: "Checking AWS service quotas..."
      endpoint: "/api/v1/aws/quotas/verify"
      method: POST
      required: false
      skip-flag: "--skip-quota-check"

    - name: validate-network
      description: "Validating VPC and subnet configuration..."
      endpoint: "/api/v1/aws/network/verify"
      method: POST
      required: true
```

**Runtime behavior:**

```bash
$ mycli create cluster --cluster-name prod
✓ Verifying AWS credentials...
✓ Checking AWS service quotas...
✗ Validating VPC and subnet configuration...
  Error: Subnet subnet-123 does not have enough available IPs (need 64, have 12)

Cluster creation aborted due to validation failures.
Run with --skip-quota-check to skip optional checks.
```

**Key features:**
- `required: true` - Failure blocks operation
- `required: false` - Warning only, operation continues
- `skip-flag` - Allow users to skip specific checks
- Executed in parallel for speed
- Progress indicators for each check

### x-cli-confirmation: Safety for Destructive Ops

Prevent accidental data loss with confirmation prompts:

```yaml
delete:
  operationId: deleteCluster
  x-cli-command: "delete cluster"

  x-cli-confirmation:
    enabled: true
    message: "Are you sure you want to delete cluster '{cluster_id}'? This action cannot be undone."
    flag: "--yes"
```

**Runtime behavior:**

```bash
$ mycli delete cluster prod-cluster
? Are you sure you want to delete cluster 'prod-cluster'? This action cannot be undone. (y/N) y
Deleting cluster...

# Skip prompt with flag (for scripts)
$ mycli delete cluster prod-cluster --yes
Deleting cluster...
```

**Template variables:**

Confirmation messages support variable substitution from path/query parameters:

```yaml
x-cli-confirmation:
  message: "Delete cluster '{cluster_id}' in region '{region}'? All data will be lost."
```

### x-cli-parent-resource: Resource Hierarchy

Nested resources are common in REST APIs (clusters → machine pools → nodes). The `x-cli-parent-resource` extension creates intuitive CLI hierarchies:

```yaml
paths:
  /api/v1/clusters/{cluster_id}/machine_pools:
    post:
      x-cli-command: "create machinepool"
      x-cli-parent-resource: "cluster"

      x-cli-flags:
        - name: machinepool-name
          source: name
          flag: "--name"
          required: true
        - name: replicas
          source: replicas
          flag: "--replicas"
          required: true
        - name: instance-type
          source: instance_type
          flag: "--instance-type"
          required: true
```

**Generated command:**

```bash
mycli create machinepool \
  --cluster prod-cluster \
  --name workers \
  --replicas 3 \
  --instance-type m5.xlarge
```

**Path parameter substitution:**

CliForge automatically:
1. Detects `{cluster_id}` in the path
2. Adds `--cluster` flag via `x-cli-parent-resource`
3. Looks up the cluster ID from the name
4. Substitutes it into the API path

**Multiple levels of nesting:**

```yaml
# /api/v1/clusters/{cluster_id}/machine_pools/{pool_id}/nodes
paths:
  /api/v1/clusters/{cluster_id}/machine_pools/{pool_id}/nodes/{node_id}:
    get:
      x-cli-command: "describe node"
      x-cli-parent-resource: "cluster"
      x-cli-parent-resource-2: "machinepool"
```

**Generates:**
```bash
mycli describe node node-abc \
  --cluster prod-cluster \
  --machinepool workers
```

### Real Examples from ROSA Spec

Let's examine a complete example from the ROSA-like CLI implementation:

```yaml
paths:
  /api/v1/clusters:
    post:
      operationId: createCluster
      summary: Create a new cluster
      tags:
        - Clusters

      # CLI command definition
      x-cli-command: "create cluster"

      # CLI flags
      x-cli-flags:
        - name: cluster-name
          source: name
          flag: "--cluster-name"
          aliases: ["-n"]
          required: true
          description: "Name of the cluster"

        - name: region
          source: region
          flag: "--region"
          required: true
          description: "AWS region"

        - name: multi-az
          source: multi_az
          flag: "--multi-az"
          type: boolean
          description: "Deploy to multiple availability zones"

        - name: version
          source: version
          flag: "--version"
          description: "OpenShift version"

        - name: compute-machine-type
          source: compute.machine_type
          flag: "--compute-machine-type"
          default: "m5.xlarge"
          description: "Instance type for worker nodes"

        - name: compute-replicas
          source: compute.replicas
          flag: "--compute-replicas"
          type: integer
          default: 2
          description: "Number of worker nodes"

      # Interactive mode
      x-cli-interactive:
        enabled: true
        prompts:
          - parameter: name
            type: text
            message: "Cluster name:"
            validation: "^[a-z][a-z0-9-]{0,53}[a-z0-9]$"
            validation-message: "Name must be lowercase alphanumeric with hyphens"

          - parameter: region
            type: select
            message: "Select AWS region:"
            source:
              endpoint: "/api/v1/regions"
              value-field: "id"
              display-field: "display_name"

          - parameter: version
            type: select
            message: "Select OpenShift version:"
            source:
              endpoint: "/api/v1/versions"
              value-field: "id"
              display-field: "display_name"
              filter: "enabled == true"

          - parameter: multi_az
            type: confirm
            message: "Deploy to multiple availability zones?"
            default: false

          - parameter: compute.replicas
            type: number
            message: "Number of worker nodes:"
            default: 2
            validation: "value >= 2 && value <= 100"
            validation-message: "Must be between 2 and 100"

      # Pre-flight checks
      x-cli-preflight:
        - name: verify-credentials
          description: "Verifying AWS credentials..."
          endpoint: "/api/v1/aws/credentials/verify"
          method: POST
          required: true

        - name: check-quota
          description: "Checking AWS service quotas..."
          endpoint: "/api/v1/aws/quotas/verify"
          method: POST
          required: false
          skip-flag: "--skip-quota-check"

      # Output configuration
      x-cli-output:
        success-message: "Cluster '{name}' is being created (ID: {id})"
        watch-status: true
        status-field: "state"
        status-endpoint: "/api/v1/clusters/{id}"
        terminal-states: ["ready", "error"]
        polling:
          interval: 30
          timeout: 3600

      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ClusterCreateRequest'

      responses:
        '201':
          description: Cluster creation initiated
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Cluster'
        '400':
          description: Invalid request
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'
```

This single operation definition generates a full-featured CLI command with:
- Flag parsing and validation
- Interactive wizard
- Pre-flight validation
- Async operation polling
- Progress indicators
- Error handling

---

## OAuth2 Authentication Patterns

### Authorization Code Flow with PKCE

CliForge implements OAuth2 authorization code flow with PKCE (Proof Key for Code Exchange) for maximum security:

```yaml
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      x-auth-config:
        flows:
          authorizationCode:
            authorizationUrl: https://sso.example.com/auth
            tokenUrl: https://sso.example.com/token
            refreshUrl: https://sso.example.com/token
            scopes:
              openid: OpenID Connect scope
              api:clusters:read: Read cluster information
              api:clusters:write: Create and modify clusters
              api:clusters:delete: Delete clusters

        token-storage:
          - type: file
            path: "~/.mycli/config.json"
          - type: keyring
            service: "mycli"
            keyring-backends:
              darwin: keychain
              linux: secret-service
              windows: wincred
```

**How it works:**

1. **User initiates login:**
   ```bash
   $ mycli login
   ```

2. **CLI generates PKCE challenge:**
   ```go
   // CliForge generates cryptographically secure verifier
   verifier := generateCodeVerifier()  // Random 43-128 char string
   challenge := sha256(verifier)       // SHA-256 hash
   ```

3. **Browser opens to authorization URL:**
   ```
   https://sso.example.com/auth?
     client_id=mycli
     &response_type=code
     &redirect_uri=http://localhost:8080/callback
     &scope=openid+api:clusters:read+api:clusters:write
     &code_challenge=E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM
     &code_challenge_method=S256
   ```

4. **User authenticates and grants permission**

5. **Authorization server redirects to localhost with code:**
   ```
   http://localhost:8080/callback?code=abc123
   ```

6. **CLI exchanges code for token:**
   ```
   POST https://sso.example.com/token
   Content-Type: application/x-www-form-urlencoded

   grant_type=authorization_code
   &code=abc123
   &redirect_uri=http://localhost:8080/callback
   &client_id=mycli
   &code_verifier=<original_verifier>
   ```

7. **Server validates and returns token:**
   ```json
   {
     "access_token": "eyJ...",
     "refresh_token": "def456",
     "expires_in": 3600,
     "token_type": "Bearer"
   }
   ```

8. **CLI stores token securely**

**Implementation in CliForge:**

```go
// From pkg/auth/oauth2.go

func (o *OAuth2Auth) authenticateAuthorizationCode(ctx context.Context) (*Token, error) {
    // Generate PKCE verifier and challenge
    verifier, err := generateCodeVerifier()
    if err != nil {
        return nil, err
    }

    challenge := generateCodeChallenge(verifier)
    o.pkceVerifier = verifier

    // Start local server for callback
    callbackChan := make(chan string, 1)
    server := o.startCallbackServer(callbackChan)
    defer server.Shutdown(ctx)

    // Build authorization URL
    authURL := o.oauth2Config.AuthCodeURL("state",
        oauth2.SetAuthURLParam("code_challenge", challenge),
        oauth2.SetAuthURLParam("code_challenge_method", "S256"),
    )

    // Open browser
    fmt.Printf("Opening browser to: %s\n", authURL)
    if err := openBrowser(authURL); err != nil {
        fmt.Printf("Please open this URL manually: %s\n", authURL)
    }

    // Wait for callback
    code := <-callbackChan

    // Exchange code for token
    token, err := o.oauth2Config.Exchange(ctx, code,
        oauth2.SetAuthURLParam("code_verifier", verifier),
    )
    if err != nil {
        return nil, err
    }

    return &Token{
        AccessToken:  token.AccessToken,
        RefreshToken: token.RefreshToken,
        ExpiresAt:    token.Expiry,
    }, nil
}
```

### Token Storage: File + Keyring

CliForge supports dual storage strategy for tokens:

**1. File Storage** - Fast, simple, cross-platform
```json
// ~/.mycli/config.json
{
  "access_token": "eyJ...",
  "refresh_token": "def456",
  "expires_at": "2025-12-01T12:00:00Z"
}
```

**2. Keyring Storage** - Secure, encrypted by OS
- **macOS**: Keychain
- **Linux**: Secret Service (GNOME Keyring, KWallet)
- **Windows**: Windows Credential Manager

**Storage strategy:**

```yaml
token-storage:
  - type: file
    path: "~/.mycli/config.json"
  - type: keyring
    service: "mycli"
```

**Implementation:**

```go
// From pkg/auth/storage/storage.go

type MultiStorage struct {
    storages []Storage
}

func (m *MultiStorage) Store(token *Token) error {
    var lastErr error
    for _, storage := range m.storages {
        if err := storage.Store(token); err != nil {
            lastErr = err
            continue
        }
    }
    return lastErr
}

func (m *MultiStorage) Retrieve() (*Token, error) {
    // Try each storage in order
    for _, storage := range m.storages {
        token, err := storage.Retrieve()
        if err == nil && token != nil {
            return token, nil
        }
    }
    return nil, ErrTokenNotFound
}
```

**Keyring implementation:**

```go
// From pkg/auth/storage/keyring.go

func (k *KeyringStorage) Store(token *Token) error {
    item := keyring.Item{
        Key:         k.key,
        Data:        tokenJSON,
        Label:       fmt.Sprintf("%s token", k.service),
        Description: fmt.Sprintf("Authentication token for %s", k.service),
    }

    return k.ring.Set(item)
}

func (k *KeyringStorage) Retrieve() (*Token, error) {
    item, err := k.ring.Get(k.key)
    if err != nil {
        return nil, err
    }

    var token Token
    if err := json.Unmarshal(item.Data, &token); err != nil {
        return nil, err
    }

    return &token, nil
}
```

### Automatic Token Refresh

CliForge automatically refreshes expired tokens before API calls:

```go
// From pkg/auth/manager.go

func (m *Manager) EnsureValidToken(ctx context.Context) error {
    token, err := m.storage.Retrieve()
    if err != nil {
        return fmt.Errorf("not authenticated - run 'login' first")
    }

    // Check if token is expired or will expire soon
    if time.Until(token.ExpiresAt) < 5*time.Minute {
        // Refresh token
        newToken, err := m.authenticator.RefreshToken(ctx, token)
        if err != nil {
            return fmt.Errorf("token refresh failed: %w", err)
        }

        // Store new token
        if err := m.storage.Store(newToken); err != nil {
            return err
        }

        token = newToken
    }

    // Update HTTP client with current token
    m.httpClient.Transport = &oauth2.Transport{
        Source: oauth2.StaticTokenSource(&oauth2.Token{
            AccessToken: token.AccessToken,
        }),
    }

    return nil
}
```

**Refresh flow:**

```go
func (o *OAuth2Auth) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
    // Create token source from refresh token
    tok := &oauth2.Token{
        RefreshToken: token.RefreshToken,
    }

    tokenSource := o.oauth2Config.TokenSource(ctx, tok)

    // Get new token
    newToken, err := tokenSource.Token()
    if err != nil {
        return nil, fmt.Errorf("refresh failed: %w", err)
    }

    return &Token{
        AccessToken:  newToken.AccessToken,
        RefreshToken: newToken.RefreshToken,
        ExpiresAt:    newToken.Expiry,
    }, nil
}
```

### Built-in Auth Commands

CliForge automatically generates authentication commands:

**login:**
```bash
$ mycli login
Opening browser to: https://sso.example.com/auth...
✓ Authentication successful
Token stored in keyring
```

**logout:**
```bash
$ mycli logout
✓ Logged out successfully
Tokens cleared from storage
```

**whoami:**
```bash
$ mycli whoami
Logged in as: john.doe@example.com
Organization: ACME Corp
Scopes: openid, api:clusters:read, api:clusters:write
Token expires: 2025-12-01 12:00:00 (in 23 hours)
```

**token:**
```bash
# Display access token
$ mycli token
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...

# Refresh and display
$ mycli token --refresh
Refreshing token...
✓ Token refreshed
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

---

## Token-Based Authentication (v0.11.0)

### The Token Resolution Problem

Enterprise CLIs like ROSA need to support multiple token sources for different workflows:
- **CI/CD pipelines**: Environment variables for automation
- **Interactive use**: Browser OAuth flows
- **Offline tokens**: Long-lived tokens from web consoles
- **Headless servers**: Device code flows

CliForge v0.11.0 introduces the TokenResolver pattern to handle this complexity.

### Token Resolution Chain

The TokenResolver checks multiple sources in priority order:

```go
// pkg/auth/resolver.go
resolver := auth.NewTokenResolver(
    auth.WithFlagToken(flagValue),              // 1. --token flag (highest)
    auth.WithEnvVars("ROSA_TOKEN", "OCM_TOKEN"), // 2-3. Environment variables
    auth.WithStorage(storage),                  // 4. Config file
    auth.WithPromptFunc(promptFunc),            // 5. Interactive prompt (lowest)
)

token, source, err := resolver.Resolve(ctx)
// Returns: (token string, source indicator, error)
```

**Priority order matches ROSA CLI:**
1. Command-line flag (`--token`)
2. `ROSA_TOKEN` environment variable
3. `OCM_TOKEN` environment variable (OCM compatibility)
4. Saved token from config file
5. Interactive prompt (last resort)

### JWT Token Type Detection

CliForge automatically detects token types by parsing the JWT `typ` claim:

```go
// pkg/auth/jwt.go
tokenType, err := auth.DetectTokenType(tokenString)

// Returns one of:
// - TokenTypeAccess:  typ="" or typ="Bearer" → access token
// - TokenTypeRefresh: typ="Refresh" → refresh token
// - TokenTypeOffline: typ="Offline" → offline token (ROSA)
// - TokenTypeUnknown: unrecognized type
```

**How it works:**

```go
func DetectTokenType(tokenString string) (TokenType, error) {
    claims, err := ParseJWT(tokenString)
    if err != nil {
        return TokenTypeUnknown, err
    }

    typ := strings.ToLower(claims.Type)
    switch typ {
    case "bearer":
        return TokenTypeBearer, nil
    case "":
        return TokenTypeAccess, nil
    case "refresh":
        return TokenTypeRefresh, nil
    case "offline":
        return TokenTypeOffline, nil
    default:
        return TokenTypeUnknown, nil
    }
}
```

**Token routing based on type:**

```go
if tokenType == TokenTypeOffline || tokenType == TokenTypeRefresh {
    // Use as refresh token to get access token
    newToken, err := oauth2Config.Exchange(ctx, refreshToken)
} else {
    // Use directly as access token
    httpClient.Transport = &oauth2.Transport{
        Source: oauth2.StaticTokenSource(token),
    }
}
```

### Username Extraction from Tokens

Extract user identity from JWT claims for display:

```go
// pkg/auth/jwt.go
username, err := auth.ExtractUsername(tokenString)
// Returns: preferred_username or username claim
```

**Implementation:**

```go
func ExtractUsername(tokenString string) (string, error) {
    claims, err := ParseJWT(tokenString)
    if err != nil {
        return "", err
    }

    // Try preferred_username first (OIDC standard)
    if claims.PreferredUsername != "" {
        return claims.PreferredUsername, nil
    }

    // Fallback to username
    if claims.Username != "" {
        return claims.Username, nil
    }

    return "", fmt.Errorf("no username found in token claims")
}
```

**Used in whoami command:**

```bash
$ rosa whoami
Logged in: Yes
Username: john.doe@company.com
Token expires: 2025-12-01T12:00:00Z
```

### JWE (Encrypted Token) Detection

Detect encrypted tokens (JWE) for FedRAMP/GovCloud environments:

```go
// pkg/auth/jwt.go
isEncrypted := auth.IsEncryptedToken(tokenString)
```

**How it works:**

```go
func IsEncryptedToken(tokenString string) bool {
    // JWE has 5 parts: header.encryptedKey.iv.ciphertext.tag
    // JWT has 3 parts: header.payload.signature
    parts := strings.Split(tokenString, ".")
    if len(parts) != 5 {
        return false
    }

    // Verify JWE header structure
    decoded, err := base64.StdEncoding.DecodeString(parts[0] + "==")
    if err != nil {
        return false
    }

    var header JWETokenHeader
    err = json.Unmarshal(decoded, &header)
    if err != nil {
        return false
    }

    // Check for encryption algorithm and content type
    return header.Encryption != "" && header.ContentType == "JWT"
}
```

### Complete Login Flow with Token Resolution

**Implementation in ROSA-like CLI:**

```go
// examples/rosa-like/cmd/rosa/builtin.go
func runLoginEnhanced(ctx context.Context, cmd *cobra.Command, authMgr *auth.Manager,
    outputMgr *output.Manager, token string, useAuthCode bool, useDeviceCode bool) error {

    // Step 1: Create resolver with all token sources
    storage, _ := authMgr.GetStorage("default")
    resolver := auth.NewTokenResolver(
        auth.WithFlagToken(token),
        auth.WithEnvVars("ROSA_TOKEN", "OCM_TOKEN"),
        auth.WithStorage(storage),
        auth.WithPromptFunc(promptForToken),
    )

    // Step 2: Try to resolve existing token
    if !useAuthCode && !useDeviceCode {
        resolvedToken, source, err := resolver.Resolve(ctx)
        if err != nil {
            return fmt.Errorf("failed to resolve token: %w", err)
        }

        if resolvedToken != "" {
            fmt.Fprintf(cmd.OutOrStdout(), "Using token from %s\n", source)

            // Step 3: Detect token type
            tokenType, err := auth.DetectTokenType(resolvedToken)
            if err != nil {
                return fmt.Errorf("invalid token format: %w", err)
            }
            fmt.Fprintf(cmd.OutOrStdout(), "Token type: %s\n", tokenType)

            // Step 4: Authenticate with detected token
            oauth2Auth, _ := authMgr.GetAuthenticator("default").(*auth.OAuth2Auth)
            oauth2Auth.SetToken(resolvedToken)

            authToken, err := authMgr.Authenticate(ctx, "default")
            if err != nil {
                return fmt.Errorf("authentication failed: %w", err)
            }

            fmt.Fprintln(cmd.OutOrStdout(), "✓ Login successful")
            return nil
        }
    }

    // Step 5: No token found, perform interactive auth
    return performInteractiveAuth(ctx, cmd, authMgr, outputMgr, useAuthCode, useDeviceCode)
}
```

### Usage Examples

**Example 1: CI/CD Pipeline (Environment Variable)**

```yaml
# GitHub Actions
- name: Authenticate with ROSA
  env:
    ROSA_TOKEN: ${{ secrets.ROSA_OFFLINE_TOKEN }}
  run: |
    rosa login
    # Token automatically resolved from ROSA_TOKEN
```

**Example 2: Interactive (Command-Line Flag)**

```bash
# Get token from console.redhat.com/openshift/token
rosa login --token=eyJhbGciOiJSUzI1NiIs...

# Output:
# Using token from flag
# Token type: offline
# ✓ Login successful
```

**Example 3: Saved Token (Auto-Resume)**

```bash
# First login
rosa login --token=$OFFLINE_TOKEN
# Token saved to ~/.config/rosa/auth.json

# Later session
rosa whoami
# Logged in: Yes
# Username: john.doe@company.com
# (Token auto-loaded from file)
```

**Example 4: Token from Environment with Fallback**

```bash
# Set multiple token sources (resolver picks first available)
export ROSA_TOKEN=$PRIMARY_TOKEN
export OCM_TOKEN=$FALLBACK_TOKEN

rosa login
# Uses ROSA_TOKEN (higher priority)
```

### Configuration in OpenAPI Spec

Enable token-based authentication in your OpenAPI spec:

```yaml
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      x-auth-config:
        flows:
          authorizationCode:
            authorizationUrl: https://sso.redhat.com/auth
            tokenUrl: https://sso.redhat.com/token
          # NEW: Token injection flow
          token:
            enabled: true
            env-vars:
              - ROSA_TOKEN
              - OCM_TOKEN
            auto-detect-type: true

        token-storage:
          - type: file
            path: "~/.rosa/config.json"
          - type: keyring
            service: "rosa-cli"
```

### Benefits

**1. ROSA Compatibility**
- Matches ROSA CLI token resolution order exactly
- Supports `ROSA_TOKEN` and `OCM_TOKEN` environment variables
- JWT type detection compatible with Red Hat SSO tokens

**2. Flexibility**
- One login method for all environments (dev, staging, prod)
- Works in CI/CD, containers, and interactive terminals
- No code changes needed for different deployment scenarios

**3. Security**
- Tokens stored in system keyring (encrypted)
- Auto-refresh prevents credential exposure
- Type detection prevents token misuse (access vs refresh)

**4. Developer Experience**
- Auto-resume from saved tokens
- Clear error messages when tokens invalid/expired
- Username display for identity confirmation

### Troubleshooting

**Problem: Token not recognized**

```bash
rosa login --token=$TOKEN
# Error: invalid token format: failed to parse JWT
```

**Solution:** Verify token is a valid JWT:

```bash
# Check token format (should have 3 parts for JWT, 5 for JWE)
echo $TOKEN | tr '.' '\n' | wc -l

# Parse token claims (use jwt.io or jwt-cli)
jwt decode $TOKEN
```

**Problem: Token type mismatch**

```bash
rosa login --token=$ACCESS_TOKEN
# Warning: Access token will expire in 1 hour
# Consider using offline token for longer sessions
```

**Solution:** Get offline token from console:

```bash
# Visit: https://console.redhat.com/openshift/token/rosa
# Copy offline token (valid for 30 days)
export ROSA_TOKEN=$OFFLINE_TOKEN
rosa login
```

**Problem: Environment variable not working**

```bash
export ROSA_TOKEN=$TOKEN
rosa login
# Still prompting for token
```

**Solution:** Check variable name and shell:

```bash
# Verify variable is set
echo $ROSA_TOKEN

# Export in same shell as rosa command
export ROSA_TOKEN=$TOKEN && rosa login

# Or use flag instead
rosa login --token=$TOKEN
```

---

## Interactive Command Patterns

### Dynamic Option Loading from APIs

One of CliForge's most powerful features is loading options from your API in real-time:

```yaml
x-cli-interactive:
  enabled: true
  prompts:
    - parameter: region
      type: select
      message: "Select AWS region:"
      source:
        endpoint: "/api/v1/regions"
        value-field: "id"
        display-field: "display_name"
        filter: "enabled == true && supports_rosa == true"
```

**API response:**
```json
[
  {
    "id": "us-east-1",
    "display_name": "US East (N. Virginia)",
    "enabled": true,
    "supports_rosa": true
  },
  {
    "id": "us-west-2",
    "display_name": "US West (Oregon)",
    "enabled": true,
    "supports_rosa": true
  },
  {
    "id": "ap-south-1",
    "display_name": "Asia Pacific (Mumbai)",
    "enabled": true,
    "supports_rosa": false
  }
]
```

**User sees:**
```
? Select AWS region:
❯ US East (N. Virginia)
  US West (Oregon)
```

**Implementation:**

```go
// From internal/interactive/prompt.go

func (p *Prompter) SelectFromAPI(config PromptConfig) (string, error) {
    // Fetch options from API
    resp, err := p.httpClient.Get(config.Source.Endpoint)
    if err != nil {
        return "", fmt.Errorf("failed to fetch options: %w", err)
    }
    defer resp.Body.Close()

    var items []map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
        return "", err
    }

    // Apply filter if specified
    if config.Source.Filter != "" {
        items = p.filterItems(items, config.Source.Filter)
    }

    // Build display options
    options := make([]string, 0, len(items))
    valueMap := make(map[string]string)

    for _, item := range items {
        value := item[config.Source.ValueField].(string)
        display := item[config.Source.DisplayField].(string)

        options = append(options, display)
        valueMap[display] = value
    }

    // Show selection prompt
    selected, err := p.select(config.Message, options)
    if err != nil {
        return "", err
    }

    return valueMap[selected], nil
}
```

### Validation: Regex and Ranges

**Regex validation:**

```yaml
- parameter: name
  type: text
  message: "Cluster name:"
  validation: "^[a-z][a-z0-9-]{0,53}[a-z0-9]$"
  validation-message: "Name must be lowercase alphanumeric with hyphens (2-54 chars)"
```

**Runtime:**
```
? Cluster name: PROD-Cluster
✗ Name must be lowercase alphanumeric with hyphens (2-54 chars)
? Cluster name: prod-cluster
✓
```

**Range validation:**

```yaml
- parameter: replicas
  type: number
  message: "Number of worker nodes:"
  validation: "value >= 2 && value <= 100"
  validation-message: "Must be between 2 and 100"
```

**Runtime:**
```
? Number of worker nodes: 150
✗ Must be between 2 and 100
? Number of worker nodes: 4
✓
```

**Expression-based validation:**

```yaml
- parameter: machine_type
  type: select
  message: "Select instance type:"
  validation: "!startsWith(value, 't2.') || confirm('t2 instances not recommended for production. Continue?')"
```

### Multi-step Wizards

Combine prompts into a complete workflow:

```yaml
x-cli-interactive:
  enabled: true
  prompts:
    # Step 1: Basic info
    - parameter: name
      type: text
      message: "Cluster name:"
      validation: "^[a-z][a-z0-9-]{0,53}[a-z0-9]$"

    # Step 2: Infrastructure
    - parameter: region
      type: select
      message: "Select AWS region:"
      source:
        endpoint: "/api/v1/regions"
        value-field: "id"
        display-field: "display_name"

    - parameter: multi_az
      type: confirm
      message: "Deploy to multiple availability zones?"
      default: false

    # Step 3: Compute
    - parameter: compute.machine_type
      type: select
      message: "Worker node instance type:"
      source:
        endpoint: "/api/v1/machine-types"
        value-field: "id"
        display-field: "name"
        filter: "category == 'general-purpose'"

    - parameter: compute.replicas
      type: number
      message: "Number of worker nodes:"
      default: 2
      validation: "value >= 2 && value <= 100"

    # Step 4: Networking
    - parameter: network.vpc_cidr
      type: text
      message: "VPC CIDR block:"
      default: "10.0.0.0/16"
      validation: "^\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}/\\d{1,2}$"

    # Step 5: Confirmation
    - parameter: confirm_create
      type: confirm
      message: "Create cluster with these settings?"
      default: true
```

**User experience:**

```
Creating a new cluster...

? Cluster name: prod-cluster
? Select AWS region: US East (N. Virginia)
? Deploy to multiple availability zones? Yes
? Worker node instance type: m5.xlarge (4 vCPUs, 16 GiB RAM)
? Number of worker nodes: 4
? VPC CIDR block: 10.0.0.0/16

Summary:
  Name:          prod-cluster
  Region:        us-east-1 (Multi-AZ)
  Worker nodes:  4 × m5.xlarge
  Network:       10.0.0.0/16

? Create cluster with these settings? Yes

✓ Verifying AWS credentials...
✓ Checking AWS service quotas...
Creating cluster 'prod-cluster'...
```

### Non-Interactive Mode for CI/CD

All interactive commands work in CI/CD by providing flags:

```bash
# Interactive (local development)
$ mycli create cluster
? Cluster name: ...

# Non-interactive (CI/CD)
$ mycli create cluster \
  --cluster-name prod-cluster \
  --region us-east-1 \
  --multi-az \
  --compute-machine-type m5.xlarge \
  --compute-replicas 4 \
  --network-vpc-cidr 10.0.0.0/16 \
  --yes
```

**Auto-detection:**

```go
// CliForge automatically detects TTY
func shouldUseInteractive() bool {
    // Check if stdin is a terminal
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return false
    }

    // Check if flags were provided
    if flagsProvided() {
        return false
    }

    // Check environment variable
    if os.Getenv("CI") != "" {
        return false
    }

    return true
}
```

---

## Async Operations and Polling

### Long-Running Operations

Enterprise operations like cluster creation take minutes to hours. CliForge handles this with `x-cli-async`:

```yaml
post:
  operationId: createCluster
  x-cli-command: "create cluster"

  responses:
    '201':
      description: Cluster creation initiated
      content:
        application/json:
          schema:
            type: object
            properties:
              id:
                type: string
              name:
                type: string
              state:
                type: string
                enum: [pending, installing, ready, error]

  x-cli-async:
    enabled: true
    status-field: "state"
    status-endpoint: "/api/v1/clusters/{id}"
    terminal-states: ["ready", "error"]
    polling:
      interval: 30
      timeout: 3600
      backoff:
        enabled: true
        multiplier: 1.5
        max-interval: 300
```

**How it works:**

1. **Initiate operation:**
   ```bash
   $ mycli create cluster --cluster-name prod
   ```

2. **CLI makes POST request:**
   ```
   POST /api/v1/clusters
   {
     "name": "prod",
     "region": "us-east-1"
   }
   ```

3. **API returns 201 Created:**
   ```json
   {
     "id": "cluster-abc123",
     "name": "prod",
     "state": "pending"
   }
   ```

4. **CLI polls status endpoint:**
   ```
   GET /api/v1/clusters/cluster-abc123

   [30s later]
   {"id": "cluster-abc123", "state": "installing", "progress": 15}

   [30s later]
   {"id": "cluster-abc123", "state": "installing", "progress": 45}

   [30s later]
   {"id": "cluster-abc123", "state": "ready", "progress": 100}
   ```

5. **CLI detects terminal state and exits**

### Status Polling Patterns

**Fixed interval:**
```yaml
polling:
  interval: 30        # Poll every 30 seconds
  timeout: 3600       # Give up after 1 hour
```

**Exponential backoff:**
```yaml
polling:
  interval: 10        # Start with 10 seconds
  timeout: 3600
  backoff:
    enabled: true
    multiplier: 1.5   # Increase by 50% each time
    max-interval: 300 # Cap at 5 minutes
```

**Polling sequence:**
- Poll 1: 10s
- Poll 2: 15s (10 × 1.5)
- Poll 3: 22s (15 × 1.5)
- Poll 4: 33s (22 × 1.5)
- ...
- Poll N: 300s (capped)

**Implementation:**

```go
// From internal/executor/async.go

func (e *AsyncExecutor) PollUntilComplete(config AsyncConfig, resourceID string) error {
    interval := time.Duration(config.Polling.Interval) * time.Second
    timeout := time.Duration(config.Polling.Timeout) * time.Second

    startTime := time.Now()
    attempt := 0

    for {
        // Check timeout
        if time.Since(startTime) > timeout {
            return fmt.Errorf("operation timed out after %v", timeout)
        }

        // Poll status
        status, err := e.getStatus(config.StatusEndpoint, resourceID)
        if err != nil {
            return err
        }

        // Check if terminal state
        if slices.Contains(config.TerminalStates, status.State) {
            if status.State == "error" {
                return fmt.Errorf("operation failed: %s", status.Message)
            }
            return nil // Success
        }

        // Update progress indicator
        e.updateProgress(status)

        // Calculate next interval
        nextInterval := interval
        if config.Polling.Backoff.Enabled {
            multiplier := config.Polling.Backoff.Multiplier
            maxInterval := time.Duration(config.Polling.Backoff.MaxInterval) * time.Second

            nextInterval = time.Duration(float64(interval) * math.Pow(multiplier, float64(attempt)))
            if nextInterval > maxInterval {
                nextInterval = maxInterval
            }
        }

        time.Sleep(nextInterval)
        attempt++
    }
}
```

### Progress Indicators

**Spinner for operations without progress:**

```go
spinner := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
spinner.Suffix = " Creating cluster..."
spinner.Start()
defer spinner.Stop()
```

**Output:**
```
⠋ Creating cluster... (15s)
```

**Progress bar for operations with percentage:**

```go
bar := progressbar.NewOptions(100,
    progressbar.OptionSetDescription("Installing cluster"),
    progressbar.OptionShowCount(),
    progressbar.OptionShowIts(),
)

for status := range statusChan {
    bar.Set(status.Progress)
}
```

**Output:**
```
Installing cluster:  45% |████████████░░░░░░░░░░░░| (45/100) [2m15s:2m45s]
```

**Detailed status messages:**

```
Creating cluster 'prod-cluster'...

✓ Validating configuration
✓ Creating AWS resources
⠸ Installing OpenShift components (15 minutes remaining)
  • Bootstrapping control plane...
  • Configuring networking...
  • Deploying operators...
```

### Timeout Handling

**Graceful timeout:**

```yaml
polling:
  interval: 30
  timeout: 3600   # 1 hour
```

**User experience:**

```
Creating cluster 'prod-cluster'...

✓ Validating configuration
✓ Creating AWS resources
⠸ Installing OpenShift components

⚠ Operation is taking longer than expected (45 minutes)
  You can:
  • Continue waiting (timeout in 15 minutes)
  • Cancel and check status later with: mycli describe cluster prod-cluster

? Continue waiting? (Y/n)
```

**Cancel and resume:**

```bash
# Cancel with Ctrl+C
^C
Operation cancelled. Cluster creation is still in progress.

Check status with:
  mycli describe cluster prod-cluster

Watch status with:
  mycli describe cluster prod-cluster --watch

# Later...
$ mycli describe cluster prod-cluster --watch
Cluster: prod-cluster
State: installing (65% complete)
⠸ Installing OpenShift components (estimated 5 minutes remaining)
```

---

## Nested Resource Commands

### Parent Resource Flags

Nested resources require parent context. CliForge handles this elegantly:

```yaml
paths:
  /api/v1/clusters/{cluster_id}/machine_pools:
    post:
      operationId: createMachinePool
      x-cli-command: "create machinepool"
      x-cli-parent-resource: "cluster"

      parameters:
        - $ref: '#/components/parameters/ClusterId'

      x-cli-flags:
        - name: name
          source: name
          flag: "--name"
          required: true
        - name: replicas
          source: replicas
          flag: "--replicas"
          required: true
        - name: instance-type
          source: instance_type
          flag: "--instance-type"
          required: true

components:
  parameters:
    ClusterId:
      name: cluster_id
      in: path
      required: true
      schema:
        type: string
      x-cli-flag: "--cluster"
      x-cli-aliases: ["-c"]
```

**Generated command:**

```bash
mycli create machinepool \
  --cluster prod-cluster \
  --name workers \
  --replicas 3 \
  --instance-type m5.xlarge
```

**Key points:**
- `x-cli-parent-resource: "cluster"` creates `--cluster` flag
- Flag is required for nested commands
- Can use cluster name or ID

### Path Parameter Substitution

CliForge automatically resolves resource names to IDs:

**User provides name:**
```bash
mycli create machinepool --cluster prod-cluster --name workers
```

**CliForge resolves:**
1. Looks up cluster by name "prod-cluster"
2. Finds ID "cluster-abc123"
3. Substitutes into path: `/api/v1/clusters/cluster-abc123/machine_pools`
4. Makes API call

**Implementation:**

```go
// From internal/executor/resource.go

func (e *Executor) ResolveResourceID(resourceType, identifier string) (string, error) {
    // Try as ID first (exact match)
    resource, err := e.getByID(resourceType, identifier)
    if err == nil {
        return resource.ID, nil
    }

    // Try as name
    resources, err := e.listByName(resourceType, identifier)
    if err != nil {
        return "", err
    }

    if len(resources) == 0 {
        return "", fmt.Errorf("%s '%s' not found", resourceType, identifier)
    }

    if len(resources) > 1 {
        return "", fmt.Errorf("multiple %ss found with name '%s', please use ID",
            resourceType, identifier)
    }

    return resources[0].ID, nil
}
```

**Ambiguity handling:**

```bash
$ mycli create machinepool --cluster test --name workers
Error: multiple clusters found with name 'test', please use ID

$ mycli list clusters -o yaml | grep -A 2 "name: test"
- id: cluster-abc123
  name: test
  region: us-east-1
- id: cluster-def456
  name: test
  region: us-west-2

$ mycli create machinepool --cluster cluster-abc123 --name workers
Creating machine pool...
```

### Resource Hierarchy

**Example hierarchy:**

```
Cluster
├── Machine Pools
│   ├── Machine Pool
│   └── Machine Pool
├── Identity Providers
│   ├── GitHub IDP
│   ├── GitLab IDP
│   └── LDAP IDP
├── Add-ons
│   ├── Monitoring
│   └── Logging
└── Ingresses
    ├── Default Ingress
    └── Custom Ingress
```

**OpenAPI structure:**

```yaml
paths:
  # Clusters
  /api/v1/clusters/{cluster_id}:
    get:
      x-cli-command: "describe cluster"

  # Machine Pools (level 2)
  /api/v1/clusters/{cluster_id}/machine_pools:
    get:
      x-cli-command: "list machinepools"
      x-cli-parent-resource: "cluster"
    post:
      x-cli-command: "create machinepool"
      x-cli-parent-resource: "cluster"

  /api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}:
    get:
      x-cli-command: "describe machinepool"
      x-cli-parent-resource: "cluster"
    delete:
      x-cli-command: "delete machinepool"
      x-cli-parent-resource: "cluster"

  # Identity Providers (level 2)
  /api/v1/clusters/{cluster_id}/identity_providers:
    post:
      x-cli-command: "create idp"
      x-cli-parent-resource: "cluster"

  # Add-ons (level 2)
  /api/v1/clusters/{cluster_id}/addons:
    post:
      x-cli-command: "install addon"
      x-cli-parent-resource: "cluster"
```

**Generated commands:**

```bash
# Cluster operations
mycli list clusters
mycli describe cluster prod-cluster

# Machine pool operations (requires --cluster)
mycli list machinepools --cluster prod-cluster
mycli create machinepool --cluster prod-cluster --name workers
mycli describe machinepool workers --cluster prod-cluster
mycli delete machinepool workers --cluster prod-cluster

# IDP operations (requires --cluster)
mycli create idp --cluster prod-cluster --type github --name github-sso

# Add-on operations (requires --cluster)
mycli install addon --cluster prod-cluster --name monitoring
```

**Deep nesting (3+ levels):**

```yaml
# Nodes within machine pools within clusters
paths:
  /api/v1/clusters/{cluster_id}/machine_pools/{pool_id}/nodes/{node_id}:
    get:
      x-cli-command: "describe node"
      x-cli-parent-resource: "cluster"
      x-cli-parent-resource-2: "machinepool"
```

**Generated:**

```bash
mycli describe node node-abc \
  --cluster prod-cluster \
  --machinepool workers
```

---

## Output Formatting

### Table Output for Humans

Table format is the default for list commands:

```bash
$ mycli list clusters

ID             NAME            REGION      STATE    AGE
cluster-abc    prod-cluster    us-east-1   ready    2d
cluster-def    dev-cluster     us-west-2   ready    5h
cluster-ghi    staging         eu-west-1   error    1d
```

**Configuration:**

```yaml
get:
  operationId: listClusters
  x-cli-command: "list clusters"
  x-cli-output:
    table:
      columns:
        - field: id
          header: ID
          width: 14
        - field: name
          header: NAME
          width: 16
        - field: region
          header: REGION
          width: 12
        - field: state
          header: STATE
          transform: uppercase
        - field: created_at
          header: AGE
          transform: age
```

**Transforms:**

- `uppercase` - Convert to uppercase
- `lowercase` - Convert to lowercase
- `age` - Convert timestamp to human-readable age (2d, 5h)
- `date` - Format timestamp as date
- `datetime` - Format timestamp as date and time
- `json` - Stringify complex objects
- `truncate:<n>` - Truncate to n characters

**Implementation:**

```go
// From pkg/output/table.go

func (t *TableFormatter) Format(data interface{}, config OutputConfig) error {
    items := extractItems(data)

    // Build table
    table := tablewriter.NewWriter(os.Stdout)
    table.SetAutoWrapText(false)
    table.SetAutoFormatHeaders(true)
    table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
    table.SetAlignment(tablewriter.ALIGN_LEFT)

    // Set headers
    headers := make([]string, len(config.Table.Columns))
    for i, col := range config.Table.Columns {
        headers[i] = col.Header
    }
    table.SetHeader(headers)

    // Add rows
    for _, item := range items {
        row := make([]string, len(config.Table.Columns))
        for i, col := range config.Table.Columns {
            value := extractField(item, col.Field)
            value = applyTransform(value, col.Transform)
            row[i] = fmt.Sprint(value)
        }
        table.Append(row)
    }

    table.Render()
    return nil
}

func applyTransform(value interface{}, transform string) interface{} {
    switch transform {
    case "uppercase":
        return strings.ToUpper(fmt.Sprint(value))
    case "lowercase":
        return strings.ToLower(fmt.Sprint(value))
    case "age":
        t, _ := time.Parse(time.RFC3339, fmt.Sprint(value))
        return humanizeDuration(time.Since(t))
    // ... other transforms
    }
    return value
}
```

### JSON/YAML for Scripts

**JSON output:**

```bash
$ mycli list clusters -o json
[
  {
    "id": "cluster-abc",
    "name": "prod-cluster",
    "region": "us-east-1",
    "state": "ready",
    "created_at": "2025-11-27T10:00:00Z"
  },
  {
    "id": "cluster-def",
    "name": "dev-cluster",
    "region": "us-west-2",
    "state": "ready",
    "created_at": "2025-11-29T05:00:00Z"
  }
]
```

**YAML output:**

```bash
$ mycli list clusters -o yaml
- id: cluster-abc
  name: prod-cluster
  region: us-east-1
  state: ready
  created_at: 2025-11-27T10:00:00Z
- id: cluster-def
  name: dev-cluster
  region: us-west-2
  state: ready
  created_at: 2025-11-29T05:00:00Z
```

**Scripting with jq:**

```bash
# Get all cluster IDs
$ mycli list clusters -o json | jq -r '.[].id'
cluster-abc
cluster-def

# Count ready clusters
$ mycli list clusters -o json | jq '[.[] | select(.state == "ready")] | length'
2

# Get clusters in us-east-1
$ mycli list clusters -o json | jq '.[] | select(.region == "us-east-1")'
```

### Colored Output

CliForge supports colored output for better readability:

```bash
$ mycli list clusters

ID             NAME            REGION      STATE       AGE
cluster-abc    prod-cluster    us-east-1   ready       2d   # green
cluster-def    dev-cluster     us-west-2   ready       5h   # green
cluster-ghi    staging         eu-west-1   error       1d   # red
cluster-jkl    test            eu-west-1   pending     2m   # yellow
```

**Configuration:**

```yaml
x-cli-output:
  table:
    columns:
      - field: state
        header: STATE
        transform: uppercase
        colors:
          ready: green
          installing: yellow
          pending: yellow
          error: red
          deleting: red
```

**Implementation:**

```go
// From pkg/output/colors.go

import "github.com/fatih/color"

var colorMap = map[string]*color.Color{
    "green":  color.New(color.FgGreen),
    "red":    color.New(color.FgRed),
    "yellow": color.New(color.FgYellow),
    "blue":   color.New(color.FgBlue),
    "gray":   color.New(color.FgHiBlack),
}

func colorize(value string, colorName string) string {
    if c, ok := colorMap[colorName]; ok {
        return c.Sprint(value)
    }
    return value
}
```

**Disable colors:**

```bash
# Environment variable
$ NO_COLOR=1 mycli list clusters

# Flag
$ mycli list clusters --no-color

# CI/CD auto-detection
$ mycli list clusters | tee output.txt  # Colors disabled automatically
```

### Pagination

**Server-side pagination:**

```yaml
get:
  operationId: listClusters
  x-cli-command: "list clusters"

  parameters:
    - name: page
      in: query
      schema:
        type: integer
        default: 1
    - name: size
      in: query
      schema:
        type: integer
        default: 100

  x-cli-pagination:
    enabled: true
    page-param: page
    size-param: size
    total-field: total
    items-field: items
```

**Generated flags:**

```bash
mycli list clusters --page 2 --size 50
```

**Client-side pagination (interactive):**

```bash
$ mycli list clusters

ID             NAME            REGION      STATE    AGE
cluster-001    prod-1          us-east-1   ready    2d
cluster-002    prod-2          us-east-1   ready    2d
...
cluster-025    prod-25         us-east-1   ready    1d

Showing 25 of 150 clusters
[n]ext page  [p]revious page  [q]uit

> n

cluster-026    prod-26         us-east-1   ready    1d
...
```

---

## Testing Strategy

### Mock API Servers

CliForge CLIs should be tested against mock APIs, not production:

**Using httptest (Go):**

```go
// mock-server/main.go

package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"
)

func main() {
    mux := http.NewServeMux()

    // Mock cluster listing
    mux.HandleFunc("/api/v1/clusters", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodGet {
            w.WriteHeader(http.StatusMethodNotAllowed)
            return
        }

        clusters := []map[string]interface{}{
            {
                "id":         "cluster-abc",
                "name":       "prod-cluster",
                "region":     "us-east-1",
                "state":      "ready",
                "created_at": "2025-11-27T10:00:00Z",
            },
        }

        json.NewEncoder(w).Encode(clusters)
    })

    // Mock cluster creation
    mux.HandleFunc("/api/v1/clusters", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            return
        }

        var req map[string]interface{}
        json.NewDecoder(r.Body).Decode(&req)

        w.WriteHeader(http.StatusCreated)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "id":    "cluster-new",
            "name":  req["name"],
            "state": "pending",
        })
    })

    server := httptest.NewServer(mux)
    defer server.Close()

    println("Mock server running at:", server.URL)
    select {} // Keep running
}
```

**Using the mock server:**

```bash
# Start mock server
$ go run mock-server/main.go
Mock server running at: http://127.0.0.1:8080

# Configure CLI to use mock
$ export API_URL=http://127.0.0.1:8080
$ mycli list clusters
```

### Integration Tests

**Test cluster lifecycle:**

```go
// tests/cluster_test.go

func TestClusterLifecycle(t *testing.T) {
    // Setup
    mockServer := startMockServer()
    defer mockServer.Close()

    cli := newTestCLI(mockServer.URL)

    // Test: List clusters (should be empty)
    out := cli.Run("list", "clusters")
    assert.Contains(t, out, "No clusters found")

    // Test: Create cluster
    out = cli.Run("create", "cluster",
        "--cluster-name", "test-cluster",
        "--region", "us-east-1",
        "--multi-az",
    )
    assert.Contains(t, out, "Cluster 'test-cluster' is being created")

    // Test: List clusters (should show new cluster)
    out = cli.Run("list", "clusters")
    assert.Contains(t, out, "test-cluster")

    // Test: Describe cluster
    out = cli.Run("describe", "cluster", "test-cluster")
    assert.Contains(t, out, "Name: test-cluster")
    assert.Contains(t, out, "Region: us-east-1")
    assert.Contains(t, out, "Multi-AZ: true")

    // Test: Delete cluster
    out = cli.Run("delete", "cluster", "test-cluster", "--yes")
    assert.Contains(t, out, "Deleting cluster")

    // Test: List clusters (should be empty again)
    out = cli.Run("list", "clusters")
    assert.Contains(t, out, "No clusters found")
}
```

**Test interactive mode:**

```go
func TestInteractiveClusterCreation(t *testing.T) {
    mockServer := startMockServer()
    defer mockServer.Close()

    cli := newTestCLI(mockServer.URL)

    // Simulate user input
    input := strings.NewReader("test-cluster\n1\nn\n2\n")
    cli.SetInput(input)

    // Run interactive command
    out := cli.Run("create", "cluster")

    assert.Contains(t, out, "Cluster name:")
    assert.Contains(t, out, "Select region:")
    assert.Contains(t, out, "Deploy to multiple availability zones?")
    assert.Contains(t, out, "Number of worker nodes:")
    assert.Contains(t, out, "Cluster 'test-cluster' is being created")
}
```

**Test error handling:**

```go
func TestErrorHandling(t *testing.T) {
    tests := []struct {
        name          string
        args          []string
        expectedError string
    }{
        {
            name: "invalid cluster name",
            args: []string{"create", "cluster", "--cluster-name", "INVALID"},
            expectedError: "Name must be lowercase alphanumeric",
        },
        {
            name: "missing required flag",
            args: []string{"create", "cluster"},
            expectedError: "required flag --cluster-name not provided",
        },
        {
            name: "cluster not found",
            args: []string{"describe", "cluster", "nonexistent"},
            expectedError: "cluster 'nonexistent' not found",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cli := newTestCLI(mockServer.URL)
            out := cli.RunExpectError(tt.args...)
            assert.Contains(t, out, tt.expectedError)
        })
    }
}
```

### CI/CD Integration

**GitHub Actions workflow:**

```yaml
# .github/workflows/test.yml

name: Test CLI

on:
  push:
    branches: [main]
  pull_request:

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Start mock API server
        run: |
          go run mock-server/main.go &
          sleep 2
          curl http://localhost:8080/health

      - name: Build CLI
        run: make build

      - name: Run integration tests
        env:
          API_URL: http://localhost:8080
        run: |
          ./rosa list clusters
          ./rosa create cluster --cluster-name test --region us-east-1 --yes
          ./rosa describe cluster test
          ./rosa delete cluster test --yes

      - name: Run Go tests
        run: go test -v ./...
```

**Testing in Docker:**

```dockerfile
# Dockerfile.test

FROM golang:1.21

WORKDIR /app

# Copy source
COPY . .

# Build CLI
RUN make build

# Start mock server in background
RUN go run mock-server/main.go &

# Run tests
CMD ["make", "test"]
```

```bash
# Run tests in Docker
$ docker build -f Dockerfile.test -t rosa-cli-test .
$ docker run --rm rosa-cli-test
```

---

## Best Practices

### API Design for CLIs

**1. Use consistent resource naming:**

```yaml
# Good
/api/v1/clusters
/api/v1/clusters/{id}
/api/v1/clusters/{id}/machine_pools

# Avoid
/api/v1/get-clusters        # Verb in path
/api/v1/cluster/{id}        # Inconsistent singular/plural
/api/v1/clusters/{id}/pools # Inconsistent naming
```

**2. Return complete objects:**

```yaml
# Good - POST returns full object
post:
  responses:
    '201':
      content:
        application/json:
          schema:
            type: object
            properties:
              id: string
              name: string
              state: string
              created_at: string

# Avoid - POST returns only ID
post:
  responses:
    '201':
      content:
        application/json:
          schema:
            type: object
            properties:
              id: string
```

This allows the CLI to display full details immediately without a second API call.

**3. Use meaningful error codes:**

```yaml
responses:
  '400':
    description: Invalid request
    content:
      application/json:
        schema:
          type: object
          properties:
            code: string  # e.g., "INVALID_NAME"
            message: string
            details: object

x-cli-error-mapping:
  INVALID_NAME: "Invalid cluster name: {message}"
  QUOTA_EXCEEDED: "AWS quota exceeded: {message}"
  UNAUTHORIZED: "Not authenticated. Run 'mycli login' first."
```

**4. Support idempotency:**

```yaml
post:
  parameters:
    - name: Idempotency-Key
      in: header
      schema:
        type: string
```

CliForge can generate idempotency keys automatically for safe retries.

**5. Provide list filters:**

```yaml
get:
  parameters:
    - name: state
      in: query
      schema:
        type: string
        enum: [pending, ready, error]
    - name: region
      in: query
      schema:
        type: string
```

Generates:
```bash
mycli list clusters --state ready --region us-east-1
```

### Error Messages

**Good error messages are:**

1. **Specific** - Tell user exactly what's wrong
2. **Actionable** - Suggest how to fix it
3. **Contextual** - Include relevant details

**Bad:**
```
Error: invalid input
```

**Good:**
```
Error: Invalid cluster name "PROD_CLUSTER"
  Cluster names must be lowercase alphanumeric with hyphens
  Example: prod-cluster, my-app-1
```

**Implementation:**

```yaml
x-cli-error-mapping:
  INVALID_NAME: |
    Invalid cluster name "{value}"
      Cluster names must be lowercase alphanumeric with hyphens
      Example: prod-cluster, my-app-1

  QUOTA_EXCEEDED: |
    AWS quota exceeded: {message}
      You can:
        • Request quota increase: aws service-quotas request-service-quota-increase
        • Delete unused resources: mycli list clusters --state unused
        • Use a different region: mycli list regions

  UNAUTHORIZED: |
    Not authenticated
      Run 'mycli login' to authenticate
```

### Help Text

**Auto-generated from OpenAPI:**

```yaml
post:
  summary: Create a new cluster
  description: |
    Creates a new OpenShift cluster on AWS.

    This operation is asynchronous - the cluster will be in 'pending' state initially
    and transition to 'ready' when installation completes (typically 30-40 minutes).

    Prerequisites:
      • Valid AWS credentials configured
      • Sufficient AWS service quotas
      • VPC and subnets configured (or use --create-vpc)

    Examples:
      # Create a basic cluster
      mycli create cluster --cluster-name prod --region us-east-1

      # Create a multi-AZ cluster with custom compute
      mycli create cluster --cluster-name prod --region us-east-1 \
        --multi-az --compute-replicas 4 --compute-machine-type m5.2xlarge

  x-cli-flags:
    - name: cluster-name
      description: |
        Name of the cluster (2-54 characters)
        Must be lowercase alphanumeric with hyphens
        Must be unique within your organization
```

**Generated help:**

```bash
$ mycli create cluster --help

Create a new OpenShift cluster on AWS

This operation is asynchronous - the cluster will be in 'pending' state initially
and transition to 'ready' when installation completes (typically 30-40 minutes).

Prerequisites:
  • Valid AWS credentials configured
  • Sufficient AWS service quotas
  • VPC and subnets configured (or use --create-vpc)

Usage:
  mycli create cluster [flags]

Flags:
  -n, --cluster-name string        Name of the cluster (2-54 characters)
                                   Must be lowercase alphanumeric with hyphens
                                   Must be unique within your organization
      --region string              AWS region (required)
      --multi-az                   Deploy to multiple availability zones
      --compute-replicas int       Number of worker nodes (default: 2)
      --compute-machine-type string Instance type for workers (default: m5.xlarge)

Examples:
  # Create a basic cluster
  mycli create cluster --cluster-name prod --region us-east-1

  # Create a multi-AZ cluster with custom compute
  mycli create cluster --cluster-name prod --region us-east-1 \
    --multi-az --compute-replicas 4 --compute-machine-type m5.2xlarge
```

### Backward Compatibility

**Version your OpenAPI spec:**

```yaml
info:
  version: 2.1.0

x-cli-config:
  version: 2.1.0
  min-api-version: 2.0.0  # Minimum compatible API version
```

**Deprecate old flags gracefully:**

```yaml
x-cli-flags:
  - name: compute-replicas
    source: compute.replicas
    flag: "--compute-replicas"
    aliases: ["--replicas"]  # Old flag still works
    deprecated: true
    deprecated-message: "Use --compute-replicas instead"
```

**User sees:**

```bash
$ mycli create cluster --replicas 4
⚠ Warning: Flag --replicas is deprecated, use --compute-replicas instead

Creating cluster...
```

**Breaking changes checklist:**

1. Bump major version (1.x.x → 2.0.0)
2. Document breaking changes in CHANGELOG
3. Provide migration guide
4. Support old flags with warnings for at least one minor version
5. Use `x-cli-since` to mark new features:

```yaml
x-cli-flags:
  - name: enable-fips
    flag: "--enable-fips"
    x-cli-since: "2.1.0"  # New in 2.1.0
```

---

## Case Study: ROSA CLI

### What We Implemented

The ROSA-like CLI demonstrates CliForge's capabilities through a real-world example:

**20+ commands across 5 categories:**

1. **Authentication** (4 commands)
   - `login` - OAuth2 authorization code flow with PKCE
   - `logout` - Clear stored credentials
   - `whoami` - Display authenticated user
   - `token` - Show/refresh access token

2. **Cluster Management** (6 commands)
   - `list clusters` - Table/JSON/YAML output with pagination
   - `create cluster` - Interactive wizard or flag-based
   - `describe cluster` - Detailed cluster information
   - `edit cluster` - Update cluster configuration
   - `delete cluster` - Async deletion with confirmation
   - `upgrade cluster` - Schedule OpenShift upgrades

3. **Nested Resources** (8 commands)
   - `create/list/describe/delete machinepool`
   - `create/list/delete idp` (identity providers)
   - `install/list addon`

4. **Utilities** (4 commands)
   - `list versions` - Available OpenShift versions
   - `list regions` - Available AWS regions
   - `verify permissions` - Pre-flight AWS permission check
   - `verify quota` - Pre-flight AWS quota check

**Key features demonstrated:**

- ✅ OAuth2 authentication with token refresh
- ✅ File + keyring token storage
- ✅ Interactive mode with API-driven select prompts
- ✅ Pre-flight validation checks
- ✅ Async operations with status polling
- ✅ Nested resource hierarchies
- ✅ Multiple output formats
- ✅ Confirmation prompts for destructive ops
- ✅ Error mapping with helpful messages

**OpenAPI spec stats:**

- 1,241 lines of YAML
- 30+ paths/operations
- 15+ x-cli-* extensions used
- 10+ schema definitions
- Comprehensive error responses

### Lessons Learned

**1. Interactive mode is critical for adoption**

Users love interactive prompts. The ability to run `rosa create cluster` without memorizing 15+ flags significantly lowers the barrier to entry:

```bash
# Before (intimidating)
$ rosa create cluster --cluster-name prod --region us-east-1 --multi-az \
    --compute-replicas 4 --compute-machine-type m5.xlarge \
    --version 4.14.0 --network-vpc-cidr 10.0.0.0/16 \
    --network-subnet-ids subnet-123,subnet-456 ...

# After (friendly)
$ rosa create cluster
? Cluster name: prod
? Select region: US East (N. Virginia)
? OpenShift version: 4.14.0
? Deploy to multiple AZs? Yes
? Number of worker nodes: 4
Creating cluster...
```

**Lesson**: Always provide interactive mode for complex operations.

**2. Dynamic option loading is a game-changer**

Loading options from the API (regions, versions, machine types) ensures the CLI is always up-to-date:

```yaml
- parameter: region
  type: select
  source:
    endpoint: "/api/v1/regions"
    value-field: "id"
    display-field: "display_name"
```

This eliminates hardcoded lists that go stale.

**Lesson**: Use API-driven selects for any data that changes (versions, regions, instance types).

**3. Pre-flight checks save time and frustration**

Running validation checks before expensive operations catches errors early:

```bash
$ rosa create cluster --cluster-name prod
✓ Verifying AWS credentials...
✗ Checking AWS service quotas...
  Error: EC2 instance quota exceeded (need 4, available 0)

Run: aws service-quotas request-service-quota-increase ...
```

This is better than starting cluster creation, waiting 10 minutes, then failing.

**Lesson**: Add pre-flight checks for operations that require external resources.

**4. Status polling needs exponential backoff**

Fixed 10-second polling for a 40-minute operation = 240 API calls. Exponential backoff reduces this to ~50 calls:

```yaml
polling:
  interval: 10
  backoff:
    enabled: true
    multiplier: 1.5
    max-interval: 300
```

**Lesson**: Always use exponential backoff for polling.

**5. Nested resources need parent context**

Users think hierarchically: "I want to list machine pools *for this cluster*"

```bash
# Good
$ rosa list machinepools --cluster prod

# Bad (confusing)
$ rosa list machinepools
Error: which cluster?
```

**Lesson**: Use `x-cli-parent-resource` for all nested resources.

**6. Error messages make or break UX**

Generic errors frustrate users:
```
Error: Bad Request
```

Specific, actionable errors help:
```
Error: Invalid cluster name "PROD_CLUSTER"
  Cluster names must be lowercase alphanumeric with hyphens
  Example: prod-cluster
```

**Lesson**: Use `x-cli-error-mapping` to provide contextual error messages.

### Performance Metrics

**Build time:**
- OpenAPI spec: 4 hours
- CLI implementation: 2 hours (mostly generated)
- Mock server: 3 hours
- Testing: 4 hours
- **Total: ~13 hours** vs. weeks for manual implementation

**Generated code:**
- 2,500+ lines of Go code generated from spec
- 500 lines of hand-written code (main.go, utilities)
- **83% generated, 17% custom**

**Command execution time:**
- List clusters: 45ms (API call) + 5ms (formatting) = 50ms
- Create cluster: 200ms (pre-flight) + 150ms (API call) = 350ms
- Interactive cluster creation: 5-10s (user input time)

**API calls:**
- Non-interactive create: 3 calls (pre-flight × 2, create × 1)
- Interactive create: 5 calls (regions, versions, pre-flight × 2, create)
- Async polling (40 min op): 50 calls (exponential backoff)

### Real-World Usage

**Simplified cluster creation:**

```bash
# Old way (manual AWS console + CLI)
1. Create VPC, subnets (AWS console - 15 min)
2. Create IAM roles (AWS console - 10 min)
3. Run rosa create cluster with 20+ flags (5 min)
4. Wait and check status manually (40 min)
Total: 70 minutes of active work

# New way (CliForge-generated CLI)
1. Run: rosa create cluster (interactive - 2 min)
2. CLI handles everything, polls status automatically
Total: 2 minutes of active work
```

**Multi-cluster deployments:**

```bash
# Create clusters in 3 regions (parallel)
for region in us-east-1 us-west-2 eu-west-1; do
  rosa create cluster \
    --cluster-name "prod-${region}" \
    --region "${region}" \
    --multi-az \
    --compute-replicas 4 \
    --yes &
done
wait

# All clusters created in parallel with status polling
```

**CI/CD integration:**

```yaml
# GitHub Actions workflow
- name: Deploy cluster
  env:
    ROSA_TOKEN: ${{ secrets.ROSA_TOKEN }}
  run: |
    rosa create cluster \
      --cluster-name ${{ github.event.inputs.cluster_name }} \
      --region ${{ github.event.inputs.region }} \
      --version ${{ github.event.inputs.version }} \
      --yes

    rosa describe cluster ${{ github.event.inputs.cluster_name }} -o json \
      > cluster-info.json
```

**Infrastructure as Code:**

```bash
# Terraform external data source
data "external" "cluster" {
  program = ["rosa", "describe", "cluster", "prod-cluster", "-o", "json"]
}

output "cluster_api_url" {
  value = data.external.cluster.result.api_url
}
```

---

## Conclusion

CliForge transforms CLI development from a months-long engineering project into a days-long configuration exercise. By making your OpenAPI specification the single source of truth, you get:

- **Consistency** - API and CLI always in sync
- **Speed** - Build production CLIs in days, not months
- **Quality** - Enterprise features (OAuth2, async polling, interactive mode) built-in
- **Maintainability** - One spec to maintain, not separate codebases

The ROSA-like CLI example proves CliForge can handle real-world enterprise complexity:
- 20+ commands
- OAuth2 authentication
- Nested resource hierarchies
- Interactive workflows
- Async operations
- Multiple output formats

**Your turn**: Take your OpenAPI spec, add `x-cli-*` extensions, and generate a production-ready CLI in hours instead of months.

---

## Additional Resources

- [CliForge Documentation](../README.md)
- [OpenAPI Extensions Reference](../openapi-extensions-reference.md)
- [Configuration DSL Guide](../configuration-dsl.md)
- [ROSA Example Source Code](../../examples/rosa-like/)
- [Technical Specification](../technical-specification.md)
- [Migration Guides](../migration-guides.md)

## Next Steps

1. **Study the ROSA example** - See real-world patterns in action
2. **Design your OpenAPI spec** - Use patterns from this guide
3. **Add x-cli-* extensions** - Start with basic commands, add features incrementally
4. **Generate your CLI** - Run CliForge generator
5. **Test with mock server** - Validate behavior before production
6. **Deploy** - Ship your CLI to users

**Questions?** Open an issue on GitHub or join our community discussions.

---

*Last updated: 2025-11-29*
*CliForge version: 0.10.0*
