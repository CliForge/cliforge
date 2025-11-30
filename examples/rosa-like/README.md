# ROSA-like CLI Example

> **A production-grade example showcasing CliForge's ability to generate enterprise-class CLIs from OpenAPI specifications.**

This example recreates core functionality of the [Red Hat OpenShift Service on AWS (ROSA) CLI](https://github.com/openshift/rosa), demonstrating how CliForge transforms a single OpenAPI specification into a fully-featured, production-ready command-line interface.

---

## Table of Contents

- [Quick Start](#quick-start)
- [What This Demonstrates](#what-this-demonstrates)
- [Architecture](#architecture)
- [Commands Reference](#commands-reference)
- [CliForge Features Demonstrated](#cliforge-features-demonstrated)
- [Usage Examples](#usage-examples)
- [OpenAPI to CLI Mapping](#openapi-to-cli-mapping)
- [File Structure](#file-structure)
- [Development Guide](#development-guide)
- [Comparison with Real ROSA](#comparison-with-real-rosa)
- [Troubleshooting](#troubleshooting)

---

## Quick Start

Get the CLI running in under 5 minutes:

```bash
# 1. Start the mock API server (in terminal 1)
cd examples/rosa-like
make mock-server

# 2. Build the CLI (in terminal 2)
make build

# 3. Try it out!
./rosa --help
./rosa version
./rosa login
./rosa clusters list

# 4. Run integration tests
make test
```

That's it! You now have a working enterprise CLI generated from an OpenAPI spec.

---

## Authentication Methods

The ROSA CLI supports multiple authentication methods, automatically selecting the best available source.

### Method 1: Token from Red Hat Console (Recommended for Getting Started)

Get your offline token from the [Red Hat Hybrid Cloud Console](https://console.redhat.com/openshift/token/rosa):

```bash
# Use --token flag
rosa login --token="eyJhbGciOiJIUzI1..."

# Or set environment variable (recommended for CI/CD)
export ROSA_TOKEN="eyJhbGciOiJIUzI1..."
rosa login

# Alternative environment variable (OCM compatibility)
export OCM_TOKEN="eyJhbGciOiJIUzI1..."
rosa login
```

**Token Resolution Order:**
1. `--token` flag (highest priority)
2. `ROSA_TOKEN` environment variable
3. `OCM_TOKEN` environment variable
4. Saved token in config file
5. Interactive prompt (lowest priority)

### Method 2: Browser OAuth Flow (Interactive)

Opens your browser for authentication with Red Hat SSO:

```bash
rosa login --use-auth-code
# Browser opens automatically
# Redirects back to CLI after login
```

**How it works:**
- CLI starts local callback server on port 9998
- Opens browser to Red Hat SSO login page
- After successful login, browser redirects to `http://localhost:9998/callback`
- CLI receives authorization code and exchanges for tokens
- Tokens stored securely in system keyring

### Method 3: Device Code Flow (Headless/Remote)

For environments without a browser (servers, containers):

```bash
rosa login --use-device-code
# Shows: Visit https://sso.redhat.com/device and enter code: ABCD-1234
# Enter code on another device
# CLI polls until authorization complete
```

**When to use:**
- SSH sessions without X11 forwarding
- CI/CD runners without browser
- Containers and remote servers
- Automated testing environments

### Method 4: Token from Config File (Auto-Resume)

If you've logged in before, the CLI automatically uses stored tokens:

```bash
# First login (token stored)
rosa login --token=$OFFLINE_TOKEN

# Later sessions (token auto-loaded)
rosa whoami
# Output: Logged in: Yes

# Token auto-refreshes when expired
rosa clusters list
# Silently refreshes token if needed
```

**Storage locations:**
- **Keyring** (secure): macOS Keychain, GNOME Keyring, Windows Credential Manager
- **File** (fallback): `~/.config/rosa/auth.json` (XDG standard)

### Environment Variables Reference

| Variable | Purpose | Example |
|----------|---------|---------|
| `ROSA_TOKEN` | Provide token directly (highest priority env var) | `export ROSA_TOKEN=eyJ...` |
| `OCM_TOKEN` | Fallback token variable (OCM compatibility) | `export OCM_TOKEN=eyJ...` |
| `ROSA_API_URL` | Override API endpoint | `export ROSA_API_URL=https://api.openshift.com` |

### Authentication Examples

**Example 1: CI/CD Pipeline**
```yaml
# GitHub Actions workflow
- name: Authenticate with ROSA
  env:
    ROSA_TOKEN: ${{ secrets.ROSA_OFFLINE_TOKEN }}
  run: |
    rosa login
    rosa clusters list
```

**Example 2: Interactive Development**
```bash
# First time setup (browser)
rosa login --use-auth-code
# ✓ Login successful
# Token stored in keyring

# Future sessions auto-resume
rosa whoami
# Logged in: Yes
# Username: john.doe@company.com
```

**Example 3: Server/Container**
```bash
# Use device code flow
rosa login --use-device-code
# Visit https://sso.redhat.com/device
# Enter code: WXYZ-5678
# Waiting for authorization...
# ✓ Login successful
```

**Example 4: Token Refresh**
```bash
# View current token
rosa token

# Force refresh if needed
rosa token --refresh
# Refreshing token...
# ✓ Token refreshed
# eyJhbGciOiJSUzI1NiIs...
```

### Troubleshooting Authentication

**"Not logged in" error:**
```bash
rosa clusters list
# Error: not authenticated - run 'rosa login' first

# Solution: Login first
rosa login --token=$OFFLINE_TOKEN
```

**Token expired:**
```bash
# Tokens auto-refresh, but you can force it:
rosa token --refresh
```

**Check authentication status:**
```bash
rosa whoami
# Logged in: Yes
# Username: john.doe@company.com
# Token expires: 2025-12-01T12:00:00Z
```

**Logout and clear tokens:**
```bash
rosa logout
# ✓ Logged out
# Credentials removed
```

---

## What This Demonstrates

This example proves CliForge can handle **real-world enterprise CLI complexity**:

| Feature | Description | Impact |
|---------|-------------|--------|
| **OAuth2 Authentication** | Full authorization code flow with PKCE, token refresh, and secure keyring storage | Production-grade security out of the box |
| **Complex CRUD Operations** | Complete cluster lifecycle: create, read, update, delete with validation | Handles sophisticated business logic |
| **Nested Resource Hierarchies** | Machine pools, IDPs, addons scoped to parent clusters | Natural modeling of complex domain relationships |
| **Interactive Mode** | Smart prompts that fetch options from API endpoints | Superior UX for complex operations |
| **Async Operation Polling** | Status monitoring for long-running cluster operations | Proper handling of real-world async workflows |
| **Multiple Output Formats** | Table, JSON, and YAML with consistent formatting | Integration-friendly and human-readable |
| **Pre-flight Validation** | Check AWS credentials and quotas before operations | Fail fast, fail clearly |
| **Confirmation Prompts** | Required confirmations for destructive operations | Safety and compliance built-in |
| **Comprehensive Help** | Auto-generated from OpenAPI descriptions | Self-documenting CLI |

**All of this from a single OpenAPI specification with CliForge extensions.**

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         User Interface                          │
│                                                                 │
│  $ rosa clusters create --cluster-name prod --region us-east-1 │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      CliForge Runtime                           │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐   │
│  │   Command    │  │     Auth     │  │      Output        │   │
│  │   Builder    │  │   Manager    │  │     Manager        │   │
│  └──────────────┘  └──────────────┘  └────────────────────┘   │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐   │
│  │  Interactive │  │   Progress   │  │       State        │   │
│  │   Prompter   │  │   Spinner    │  │      Manager       │   │
│  └──────────────┘  └──────────────┘  └────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │           OpenAPI Spec Parser & Executor               │   │
│  │         (Reads api/rosa-api.yaml at runtime)           │   │
│  └─────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │
                             │ HTTP/REST
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│                      Mock API Server                            │
│                    (localhost:8080)                             │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │         Thread-Safe In-Memory Data Store               │   │
│  │  • Clusters  • Machine Pools  • IDPs  • Add-ons        │   │
│  └─────────────────────────────────────────────────────────┘   │
│                                                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │              OAuth2 Token Management                    │   │
│  │     (Authorization code flow + refresh tokens)          │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

### How It Works

1. **OpenAPI Specification** (`api/rosa-api.yaml`) defines the API contract and CLI behavior using CliForge's `x-cli-*` extensions
2. **CliForge Runtime** parses the spec at build time and constructs the command tree dynamically
3. **Mock API Server** provides realistic API responses for testing and development
4. **Generated CLI** provides a professional user experience with zero hand-written command code

**The magic:** Change the OpenAPI spec, rebuild, and your CLI automatically reflects the new API structure.

---

## Commands Reference

### Authentication Commands

```bash
# Login with token from Red Hat console
rosa login --token=$OFFLINE_TOKEN

# Login with browser (default)
rosa login --use-auth-code

# Login for headless environments
rosa login --use-device-code

# Or use environment variables
export ROSA_TOKEN=$OFFLINE_TOKEN
rosa login

# Other auth commands
rosa logout             # Clear stored credentials
rosa whoami             # Display authentication status and username
rosa token              # Print access token (with auto-refresh)
rosa token --refresh    # Force token refresh
```

### Cluster Management

```bash
# List all clusters
rosa clusters list
rosa clusters list -o json
rosa clusters list --search "name like 'prod%'"

# Create cluster (interactive mode)
rosa clusters create --interactive

# Create cluster (declarative)
rosa clusters create \
  --cluster-name my-cluster \
  --region us-east-1 \
  --version 4.14.5 \
  --multi-az \
  --compute-machine-type m5.xlarge \
  --replicas 3

# Get cluster details
rosa clusters describe --cluster my-cluster
rosa clusters describe --cluster abc123 -o yaml

# Update cluster
rosa clusters edit --cluster my-cluster --replicas 5

# Delete cluster (with confirmation)
rosa clusters delete --cluster my-cluster
rosa clusters delete --cluster my-cluster --yes  # Skip confirmation
```

### Nested Resources

```bash
# Machine pools
rosa machine-pools create --cluster my-cluster \
  --name workers \
  --machine-type m5.2xlarge \
  --replicas 5

rosa machine-pools list --cluster my-cluster

# Identity providers
rosa identity-providers create --cluster my-cluster \
  --type github \
  --name github-idp

rosa identity-providers list --cluster my-cluster

# Upgrade policies
rosa upgrades upgrade --cluster my-cluster --version 4.14.6
rosa upgrades list --cluster my-cluster
```

### Utilities

```bash
# Available OpenShift versions
rosa versions list
rosa versions list --channel-group stable

# Available AWS regions
rosa regions list

# Pre-flight checks (used internally during cluster creation)
# These are typically called automatically, not manually
```

---

## CliForge Features Demonstrated

### 1. OpenAPI Extensions (`x-cli-*`)

The `api/rosa-api.yaml` spec uses CliForge's custom extensions to control CLI behavior:

```yaml
paths:
  /api/v1/clusters:
    post:
      operationId: createCluster
      x-cli-command: "create cluster"
      x-cli-flags:
        - name: cluster-name
          source: name
          flag: "--cluster-name"
          required: true
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
      x-cli-preflight:
        - name: verify-aws-credentials
          endpoint: "/api/v1/aws/credentials/verify"
```

This declarative approach means **no code generation required** - the CLI behavior is defined entirely in the OpenAPI spec.

### 2. Interactive Prompts

When creating a cluster with `--interactive` flag, CliForge automatically:
- Fetches available regions from the API
- Presents them as a selectable list
- Validates input according to spec constraints
- Populates the request payload

```bash
$ rosa clusters create --interactive

Cluster name: my-prod-cluster
Select AWS region:
  ❯ us-east-1 (US East - N. Virginia)
    us-west-2 (US West - Oregon)
    eu-west-1 (EU - Ireland)
Select OpenShift version:
  ❯ 4.14.5 (stable)
    4.14.4 (stable)
    4.15.0 (candidate)
Deploy to multiple availability zones? (y/N): y
Number of worker nodes: 3

✓ Verifying AWS credentials...
✓ Checking AWS service quotas...
✓ Cluster 'my-prod-cluster' is being created
⠋ Waiting for cluster to become ready...
```

### 3. Authentication Integration

OAuth2 authentication is configured in the OpenAPI spec:

```yaml
components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      x-auth-config:
        flows:
          authorizationCode:
            authorizationUrl: https://sso.example.com/auth
            tokenUrl: https://sso.example.com/token
        token-storage:
          - type: keyring
            service: "rosa-like-cli"
```

CliForge automatically:
- Opens browser for OAuth2 authorization
- Handles PKCE flow
- Stores tokens securely in system keyring
- Auto-refreshes expired tokens
- Injects `Authorization: Bearer <token>` headers

### 4. Output Formatting

All commands support multiple output formats:

```bash
# Human-readable table (default)
$ rosa clusters list
ID        NAME           STATE   REGION      VERSION
abc-123   prod-cluster   ready   us-east-1   4.14.5
def-456   dev-cluster    ready   us-west-2   4.14.4

# JSON (for scripting)
$ rosa clusters list -o json
{
  "items": [
    {
      "id": "abc-123",
      "name": "prod-cluster",
      "state": "ready",
      "region": "us-east-1",
      "version": "4.14.5"
    }
  ]
}

# YAML (for configuration)
$ rosa clusters describe --cluster prod -o yaml
id: abc-123
name: prod-cluster
state: ready
region: us-east-1
version: 4.14.5
```

### 5. Async Operation Polling

Long-running operations automatically poll for completion:

```yaml
x-cli-output:
  watch-status: true
  status-field: "state"
  status-endpoint: "/api/v1/clusters/{id}"
  terminal-states: [ready, error]
  polling:
    interval: 30
    timeout: 3600
```

```bash
$ rosa clusters create --cluster-name my-cluster --region us-east-1
✓ Cluster 'my-cluster' creation initiated
⠋ Waiting for cluster to become ready... (state: installing)
  Elapsed: 5m 30s
```

### 6. Pre-flight Validation

Before executing expensive operations, CliForge runs pre-flight checks:

```yaml
x-cli-preflight:
  - name: verify-aws-credentials
    description: "Verifying AWS credentials..."
    endpoint: "/api/v1/aws/credentials/verify"
  - name: verify-quota
    description: "Checking AWS service quotas..."
    endpoint: "/api/v1/aws/quotas/verify"
```

If checks fail, the operation is aborted before API calls:

```bash
$ rosa clusters create --cluster-name my-cluster
✓ Verifying AWS credentials...
✗ Checking AWS service quotas...
Error: Insufficient EC2 quota in us-east-1
  Current: 10 vCPUs
  Required: 48 vCPUs

Please request a quota increase at:
https://console.aws.amazon.com/servicequotas/
```

---

## Usage Examples

### Example 1: Create a Development Cluster

```bash
# Interactive mode walks you through all options
$ rosa clusters create --interactive

Cluster name: dev-testing
Select AWS region: us-east-1
Select OpenShift version: 4.14.5
Deploy to multiple availability zones? n
Number of worker nodes: 2

✓ Cluster 'dev-testing' is being created
⠋ Waiting for cluster to become ready...
```

### Example 2: Create a Production Cluster

```bash
# Declarative mode for automation/IaC
$ rosa clusters create \
    --cluster-name production \
    --region us-east-1 \
    --multi-az \
    --version 4.14.5 \
    --compute-machine-type m5.2xlarge \
    --enable-autoscaling \
    --min-replicas 3 \
    --max-replicas 10

✓ Verifying AWS credentials...
✓ Checking AWS service quotas...
✓ Cluster 'production' is being created

Cluster Details:
  ID: prod-abc-123
  API URL: https://api.production.abc123.p1.openshiftapps.com:6443
  Console: https://console-openshift-console.apps.production.abc123.p1.openshiftapps.com
```

### Example 3: Add a Machine Pool

```bash
# Add GPU nodes for ML workloads
$ rosa machine-pools create \
    --cluster production \
    --name gpu-pool \
    --machine-type p3.2xlarge \
    --replicas 2 \
    --labels workload=ml,gpu=true

✓ Machine pool 'gpu-pool' created
```

### Example 4: Configure GitHub Authentication

```bash
# Add GitHub as identity provider
$ rosa identity-providers create \
    --cluster production \
    --type github \
    --name github-sso \
    --client-id abc123 \
    --client-secret $GITHUB_CLIENT_SECRET \
    --organizations mycompany

✓ Identity provider 'github-sso' configured
  Users from 'mycompany' GitHub org can now log in
```

### Example 5: Scripting with JSON Output

```bash
#!/bin/bash
# Get all ready clusters in us-east-1
clusters=$(rosa clusters list -o json | \
  jq -r '.items[] | select(.state == "ready" and .region == "us-east-1") | .name')

for cluster in $clusters; do
  echo "Backing up $cluster..."
  # ... backup logic ...
done
```

---

## OpenAPI to CLI Mapping

Here's how CliForge transforms OpenAPI definitions into CLI commands:

### API Operation → CLI Command

**OpenAPI:**
```yaml
paths:
  /api/v1/clusters:
    get:
      operationId: listClusters
      summary: List clusters
      x-cli-command: "list clusters"
```

**Generated CLI:**
```bash
$ rosa clusters list
# Calls: GET /api/v1/clusters
```

### Path Parameters → Required Flags

**OpenAPI:**
```yaml
paths:
  /api/v1/clusters/{cluster_id}:
    parameters:
      - name: cluster_id
        in: path
        required: true
        x-cli-flag: "--cluster"
```

**Generated CLI:**
```bash
$ rosa clusters describe --cluster my-cluster
# Calls: GET /api/v1/clusters/my-cluster
```

### Request Body → Flags

**OpenAPI:**
```yaml
requestBody:
  content:
    application/json:
      schema:
        properties:
          name:
            type: string
          region:
            type: string
x-cli-flags:
  - name: cluster-name
    source: name
    flag: "--cluster-name"
  - name: region
    source: region
    flag: "--region"
```

**Generated CLI:**
```bash
$ rosa clusters create --cluster-name test --region us-east-1
# Calls: POST /api/v1/clusters
# Body: {"name": "test", "region": "us-east-1"}
```

### Nested Resources → Hierarchical Commands

**OpenAPI:**
```yaml
paths:
  /api/v1/clusters/{cluster_id}/machine_pools:
    post:
      operationId: createMachinePool
      x-cli-command: "create machinepool"
      x-cli-parent-resource: "cluster"
```

**Generated CLI:**
```bash
$ rosa machine-pools create --cluster my-cluster --name workers
# Calls: POST /api/v1/clusters/my-cluster/machine_pools
```

---

## File Structure

```
rosa-like/
├── api/
│   └── rosa-api.yaml           # OpenAPI 3.0 spec with x-cli-* extensions
│                               # SINGLE SOURCE OF TRUTH for CLI behavior
│
├── cmd/
│   └── rosa/
│       ├── main.go             # CLI entry point
│       │                       # - Loads OpenAPI spec
│       │                       # - Sets up auth, output, state managers
│       │                       # - Builds command tree via CliForge
│       │
│       └── builtin.go          # Built-in commands (login, logout, whoami, token)
│                               # These don't come from the OpenAPI spec
│
├── mock-server/
│   ├── main.go                 # HTTP server and routing
│   ├── types.go                # API data structures
│   ├── store.go                # Thread-safe in-memory storage
│   ├── auth.go                 # OAuth2 token handling
│   ├── test.sh                 # Comprehensive API test script
│   └── README.md               # Mock server documentation
│
├── tests/
│   └── (integration tests)     # End-to-end CLI tests
│
├── Makefile                    # Build, run, and test targets
│
├── README.md                   # This file
│
└── rosa                        # Built binary (created by `make build`)
```

### Key Files Explained

**`api/rosa-api.yaml`** - The heart of the example. This OpenAPI specification:
- Defines all API endpoints, request/response schemas
- Uses `x-cli-*` extensions to control CLI generation
- Configures authentication, interactive prompts, validation
- **No Go code needed** - just YAML configuration

**`cmd/rosa/main.go`** - Minimal boilerplate that:
- Loads the OpenAPI spec
- Instantiates CliForge managers (auth, output, state, progress)
- Calls `builder.NewBuilder(spec, config).Build()`
- Registers built-in commands
- The command tree is **generated at runtime** from the spec

**`mock-server/`** - Production-quality mock API that:
- Implements the exact API contract from `rosa-api.yaml`
- Provides realistic OAuth2 flow
- Returns proper error codes and validation messages
- Enables development and testing without real infrastructure

---

## Development Guide

### Prerequisites

- Go 1.21 or later
- Make
- jq (for mock server tests)

### Building from Source

```bash
# From examples/rosa-like directory

# Build CLI binary
make build

# Install to $GOPATH/bin
make install

# Run CLI with arguments
make run ARGS="clusters list"
```

### Running the Mock Server

```bash
# Foreground (see logs)
make mock-server

# Background
make mock-server-bg

# Check server health
curl http://localhost:8080/health

# Stop background server
make mock-server-stop
```

### Testing

```bash
# Run integration tests
make test

# Test mock server API
cd mock-server && ./test.sh

# Quick smoke test
make quick
```

### Modifying the CLI

**To add a new command:**

1. Add the endpoint to `api/rosa-api.yaml`:
   ```yaml
   /api/v1/clusters/{cluster_id}/logs:
     get:
       operationId: getClusterLogs
       x-cli-command: "logs"
       x-cli-parent-resource: "cluster"
   ```

2. Rebuild:
   ```bash
   make build
   ```

3. The command is now available:
   ```bash
   ./rosa clusters logs --cluster my-cluster
   ```

**To add interactive prompts:**

Add `x-cli-interactive` to the operation:
```yaml
x-cli-interactive:
  enabled: true
  prompts:
    - parameter: log_level
      type: select
      message: "Select log level:"
      options: [debug, info, warn, error]
```

**To add pre-flight checks:**

Add `x-cli-preflight` to validate before execution:
```yaml
x-cli-preflight:
  - name: check-cluster-ready
    description: "Verifying cluster is ready..."
    endpoint: "/api/v1/clusters/{cluster_id}/status"
```

### Extending the Mock Server

To add a new endpoint:

1. Add the struct to `mock-server/types.go`
2. Add storage methods to `mock-server/store.go`
3. Add HTTP handler to `mock-server/main.go`
4. Update `api/rosa-api.yaml` to match

---

## Comparison with Real ROSA

This example implements a **representative subset** of ROSA CLI functionality:

| Aspect | Real ROSA CLI | This Example |
|--------|---------------|--------------|
| **Commands** | 100+ commands covering full OpenShift lifecycle | ~20 core commands demonstrating key patterns |
| **Authentication** | OAuth2 (device code + auth code flows), offline tokens | OAuth2 authorization code flow with PKCE |
| **API Backend** | OpenShift Cluster Manager API (api.openshift.com) | Mock server (localhost:8080) |
| **AWS Integration** | Real AWS SDK calls, STS token generation | Simulated/mocked responses |
| **Cluster Creation** | Provisions real EC2 instances, VPCs, load balancers | Returns mock cluster ID, simulates states |
| **Resource Management** | Full CRUD for 20+ resource types | Clusters, machine pools, IDPs, add-ons, upgrade policies |
| **Validation** | Deep AWS quota checks, IAM policy validation | Basic schema validation, mock pre-flight checks |
| **Output** | Table, JSON, YAML | Table, JSON, YAML |
| **Error Handling** | Detailed AWS/OpenShift error messages | Structured error responses from OpenAPI spec |

### What's the Same?

- **Command structure and UX** - Nearly identical command patterns
- **Interactive mode** - Same pattern of prompting for complex operations
- **OAuth2 flow** - Real authentication with token refresh
- **Output formatting** - Professional tables, JSON, YAML
- **Async operations** - Polling and status monitoring

### What's Different?

- **Scope** - Subset focused on demonstrating CliForge capabilities
- **Backend** - Mock server instead of real OpenShift/AWS APIs
- **Validation** - Simplified (no actual AWS quota/permission checks)
- **State management** - In-memory only (real ROSA persists to API)

### Why This Example Matters

This example proves that **CliForge can generate a CLI comparable to a mature, production tool** used by thousands of OpenShift customers worldwide. The real ROSA CLI has evolved over years with multiple contributors. This example was built in days by defining an OpenAPI spec.

**Key takeaway:** If CliForge can replicate ROSA's complexity, it can handle your API.

---

## Troubleshooting

### CLI won't start / "could not find rosa-api.yaml"

**Problem:** The CLI can't locate the OpenAPI spec file.

**Solution:** Run the CLI from the `examples/rosa-like` directory, or use an absolute path:
```bash
cd examples/rosa-like
./rosa --help
```

### "Authentication failed" / "Not logged in"

**Problem:** No valid credentials stored.

**Solution:**
```bash
# Make sure mock server is running
make mock-server

# In another terminal, log in
./rosa login
```

The mock server accepts any authorization code for testing.

### "Connection refused" errors

**Problem:** Mock server not running or listening on wrong port.

**Solution:**
```bash
# Check if server is running
curl http://localhost:8080/health

# If not, start it
make mock-server

# Check for port conflicts
lsof -i :8080
```

### Commands return empty results

**Problem:** No data in mock server (in-memory storage cleared on restart).

**Solution:** Create some test data:
```bash
./rosa login
./rosa clusters create --cluster-name test --region us-east-1
./rosa clusters list
```

### Interactive prompts don't work

**Problem:** API server not reachable when fetching prompt options.

**Solution:**
```bash
# Verify server is running
curl http://localhost:8080/api/v1/regions

# Check CLI can reach it
./rosa regions list

# If working, interactive mode should work
./rosa clusters create --interactive
```

### Token expired errors

**Problem:** Access token expired (default: 1 hour).

**Solution:**
```bash
# Refresh token
./rosa token --refresh

# Or log in again
./rosa login
```

### Build errors

**Problem:** Missing dependencies or wrong Go version.

**Solution:**
```bash
# Check Go version (requires 1.21+)
go version

# From repo root, ensure dependencies are available
cd ../..
go mod tidy
go mod download

# Build again
cd examples/rosa-like
make build
```

### Tests fail

**Problem:** Test expects specific state or server behavior.

**Solution:**
```bash
# Clean all state
make clean

# Start fresh server
make mock-server-bg

# Run tests
make test

# Check test output for specific failures
cd tests && go test -v ./...
```

---

## Learn More

- **[CliForge Documentation](../../docs/)** - Complete CliForge guides and references
- **[OpenAPI Extensions Reference](../../docs/openapi-extensions-reference.md)** - All `x-cli-*` extensions explained
- **[Real ROSA CLI](https://github.com/openshift/rosa)** - The inspiration for this example
- **[OpenAPI Specification](https://spec.openapis.org/oas/v3.0.3)** - OpenAPI 3.0 standard

---

## License

This example is part of the CliForge project and is licensed under the MIT License.

---

**Built with CliForge** - Dynamic CLI generation from OpenAPI specifications.
