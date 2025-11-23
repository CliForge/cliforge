# Plugin Architecture Design

**Version**: 1.0.0
**Date**: 2025-11-23
**Status**: Proposed

---

## Table of Contents

1. [Overview](#overview)
2. [Requirements](#requirements)
3. [Architecture](#architecture)
4. [Plugin Types](#plugin-types)
5. [Security Model](#security-model)
6. [Lifecycle](#lifecycle)
7. [OpenAPI Extensions](#openapi-extensions)
8. [Examples](#examples)
9. [Implementation Notes](#implementation-notes)

---

## Overview

### Problem Statement

ROSA CLI and similar enterprise tools need to integrate with external systems that cannot be accessed via pure HTTP APIs:
- AWS CLI operations (CloudFormation, IAM, STS)
- Local file transformations
- Complex validation logic
- Third-party tool execution

CliForge's pure OpenAPI-driven approach cannot handle these requirements.

### Solution

A plugin architecture that:
1. **Extends CliForge capabilities** beyond HTTP APIs
2. **Maintains security** through sandboxing and permissions
3. **Stays optional** - simple CLIs don't need plugins
4. **Integrates cleanly** with OpenAPI workflow specifications

---

## Requirements

### Functional Requirements

**FR-1**: Execute external command-line tools (AWS CLI, kubectl, etc.)
**FR-2**: Perform local file operations (read, parse, transform)
**FR-3**: Execute custom validation logic
**FR-4**: Integrate with multiple authentication providers
**FR-5**: Support user-defined custom commands

### Non-Functional Requirements

**NFR-1**: **Security** - Plugins cannot access user credentials without explicit permission
**NFR-2**: **Isolation** - Plugin failures don't crash the main CLI
**NFR-3**: **Performance** - Plugin overhead < 100ms for typical operations
**NFR-4**: **Portability** - Plugins work across platforms (Linux, macOS, Windows)
**NFR-5**: **Discoverability** - Users can list and inspect available plugins

---

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────┐
│                    CliForge Binary                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │              Plugin Manager                       │ │
│  ├───────────────────────────────────────────────────┤ │
│  │                                                   │ │
│  │  ┌─────────────┐  ┌──────────────┐  ┌─────────┐ │ │
│  │  │   Registry  │  │   Executor   │  │ Sandbox │ │ │
│  │  └─────────────┘  └──────────────┘  └─────────┘ │ │
│  │                                                   │ │
│  │  ┌─────────────┐  ┌──────────────┐              │ │
│  │  │ Permissions │  │ Validation   │              │ │
│  │  └─────────────┘  └──────────────┘              │ │
│  │                                                   │ │
│  └───────────────────────────────────────────────────┘ │
│                          │                             │
│  ┌───────────────────────▼────────────────────────────┐│
│  │              Plugin Interface                      ││
│  ├────────────────────────────────────────────────────┤│
│  │  Execute(ctx, input) → (output, error)            ││
│  │  Validate() → error                                ││
│  │  Describe() → PluginInfo                          ││
│  └────────────────────────────────────────────────────┘│
│                          │                             │
│         ┌────────────────┴────────────────┐            │
│         │                                 │            │
│  ┌──────▼──────┐                  ┌──────▼──────┐     │
│  │   Built-in  │                  │   External  │     │
│  │   Plugins   │                  │   Plugins   │     │
│  └─────────────┘                  └─────────────┘     │
│         │                                 │            │
│  ┌──────▼──────────┐              ┌──────▼──────────┐ │
│  │ - aws-cli       │              │ User-defined    │ │
│  │ - file-ops      │              │ WASM modules    │ │
│  │ - validators    │              │ or              │ │
│  │ - formatters    │              │ executables     │ │
│  └─────────────────┘              └─────────────────┘ │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Plugin Types

CliForge supports three plugin types:

#### 1. Built-in Plugins (Compiled Into Binary)

**Purpose**: Common operations that ship with CliForge

**Examples**:
- `aws-cli` - Execute AWS CLI commands
- `file-ops` - Read, parse, transform files
- `json-ops` - JSON/YAML processing
- `cert-ops` - Certificate parsing and validation

**Pros**:
- No external dependencies
- Fast execution
- Always available

**Cons**:
- Increases binary size
- Harder to update independently

#### 2. External Binary Plugins

**Purpose**: Third-party tools or custom scripts

**Location**: `~/.config/<cli-name>/plugins/`

**Format**: Executable files following plugin protocol

**Examples**:
- Custom validation scripts
- Organization-specific tooling
- Language-specific processors

**Pros**:
- Easy to distribute
- Any language (Python, Bash, etc.)
- User-extensible

**Cons**:
- Security risks (needs sandboxing)
- Platform-specific binaries

#### 3. WebAssembly (WASM) Plugins

**Purpose**: Secure, portable user-defined plugins

**Location**: `~/.config/<cli-name>/plugins/`

**Format**: `.wasm` files compiled from any WASM-compatible language

**Examples**:
- Custom data transformations
- Complex validation logic
- Format converters

**Pros**:
- Secure (sandboxed by design)
- Cross-platform
- Fast execution

**Cons**:
- Limited I/O capabilities
- Requires WASM toolchain

---

## Security Model

### Threat Model

**T-1**: Malicious plugin steals credentials
**T-2**: Plugin modifies system files
**T-3**: Plugin makes unauthorized network calls
**T-4**: Plugin consumes excessive resources (CPU, memory)
**T-5**: Plugin exploits CLI binary vulnerabilities

### Security Controls

#### 1. Permission System

Plugins declare required permissions in manifest:

```yaml
# plugin-manifest.yaml
name: aws-cli
version: 1.0.0
type: builtin
permissions:
  - execute:aws          # Can execute AWS CLI
  - read:env:AWS_*       # Can read AWS_* env vars
  - read:file:~/.aws/*   # Can read AWS config files
  - network:aws.amazon.com # Can make network calls to AWS
```

**Permission Types**:
- `execute:<command>` - Run external commands
- `read:env:<pattern>` - Read environment variables
- `write:env:<pattern>` - Write environment variables
- `read:file:<path-pattern>` - Read files
- `write:file:<path-pattern>` - Write files
- `network:<domain-pattern>` - Make network requests
- `credential:read` - Access stored credentials

#### 2. User Approval

First time a plugin requests a permission:

```
⚠️  Plugin 'aws-cli' requests permission:

  • Execute AWS CLI command
  • Read ~/.aws/config and ~/.aws/credentials

Grant permission? [y/N]:
```

Approved permissions saved to:
```yaml
# ~/.config/mycli/plugin-permissions.yaml
plugins:
  aws-cli:
    approved_permissions:
      - execute:aws
      - read:file:~/.aws/*
    approved_at: 2025-11-23T10:30:00Z
```

#### 3. Sandboxing

**External Binary Plugins**:
- Run in restricted subprocess
- Limited environment variables
- chroot jail (Linux/macOS) or AppContainer (Windows)
- Network access via proxy only
- File access via temporary workspace

**WASM Plugins**:
- WASI runtime with capability-based security
- No host access by default
- Explicit imports only

#### 4. Plugin Signing (Future)

**v2.0 Feature**:
- Plugins signed with developer key
- CliForge verifies signatures
- Trust store for known publishers
- Unsigned plugins require explicit approval

---

## Lifecycle

### 1. Plugin Discovery

**Built-in plugins**:
- Registered at compile time
- Always available

**External plugins**:
```bash
# User installs plugin
mycli plugin install aws-helper

# CliForge downloads to ~/.config/mycli/plugins/aws-helper/
# Validates manifest
# Requests permissions
```

### 2. Plugin Registration

```go
// Plugin manifest
type PluginManifest struct {
    Name        string            `yaml:"name"`
    Version     string            `yaml:"version"`
    Type        string            `yaml:"type"`        // builtin, binary, wasm
    Executable  string            `yaml:"executable"`  // for binary plugins
    Entrypoint  string            `yaml:"entrypoint"`  // for WASM plugins
    Permissions []string          `yaml:"permissions"`
    Metadata    map[string]string `yaml:"metadata"`
}
```

### 3. Plugin Execution

**Plugin Protocol** (JSON-RPC over stdin/stdout):

```json
// Request
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "execute",
  "params": {
    "command": "aws",
    "args": ["sts", "get-caller-identity"],
    "env": {
      "AWS_PROFILE": "default"
    },
    "input": null
  }
}

// Response
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "stdout": "{\"UserId\": \"...\", \"Account\": \"123456789012\", ...}",
    "stderr": "",
    "exit_code": 0
  }
}
```

### 4. Error Handling

```yaml
# Plugin execution fails
error:
  type: plugin_execution_error
  plugin: aws-cli
  message: "AWS CLI not found in PATH"
  suggestion: "Install AWS CLI: brew install awscli"
  recoverable: false
```

---

## OpenAPI Extensions

### `x-cli-plugin`

**Location**: Operation level
**Purpose**: Invoke plugin during workflow

```yaml
paths:
  /api/v1/clusters:
    post:
      operationId: createCluster
      x-cli-workflow:
        steps:
          - id: validate-aws-creds
            type: plugin
            plugin: aws-cli
            command: validate-credentials
            input:
              profile: "{flags.aws_profile}"
            required: true
            error-message: "AWS credentials are not configured"

          - id: create-iam-roles
            type: plugin
            plugin: aws-cli
            command: cloudformation
            input:
              stack-name: "{flags.cluster_name}-roles"
              template-file: "/embedded/cf-templates/roles.yaml"
            condition: "validate-aws-creds.success == true"

          - id: create-cluster-api
            type: api-call
            endpoint: /api/v1/clusters
            method: POST
            body:
              name: "{flags.cluster_name}"
              role_arn: "{create-iam-roles.output.RoleArn}"
            depends-on: [create-iam-roles]
```

### `x-cli-plugin-config`

**Location**: Root level
**Purpose**: Configure plugin behavior

```yaml
x-cli-plugin-config:
  enabled: true
  allow-external: false  # Only built-in plugins
  plugin-dir: "~/.mycli/plugins"
  timeout: 300  # seconds
  plugins:
    aws-cli:
      enabled: true
      config:
        default-region: us-east-1
        profile: default
```

---

## Plugin Types

### Type 1: External Command Executor

**Purpose**: Run external CLI tools

**Built-in Plugin**: `exec`

**Configuration**:
```yaml
x-cli-workflow:
  steps:
    - id: run-aws-cli
      type: plugin
      plugin: exec
      input:
        command: aws
        args:
          - sts
          - get-caller-identity
          - --profile
          - "{flags.aws_profile}"
      capture-output: true
```

**Security**:
- Requires `execute:<command>` permission
- Command must be in allowed list or PATH
- Arguments validated for injection attacks

---

### Type 2: File Operations

**Purpose**: Read, parse, transform files

**Built-in Plugin**: `file-ops`

**Operations**:
- `read` - Read file contents
- `parse` - Parse JSON/YAML/XML/PEM
- `validate` - Validate file format
- `transform` - Apply transformations

**Example**:
```yaml
x-cli-workflow:
  steps:
    - id: read-ca-cert
      type: plugin
      plugin: file-ops
      input:
        operation: parse
        file: "{flags.ca_cert_file}"
        format: pem
        type: x509-certificate
      output:
        cert-data: "{result.base64}"
        issuer: "{result.issuer}"

    - id: upload-cert
      type: api-call
      endpoint: /api/v1/certificates
      body:
        certificate: "{read-ca-cert.cert-data}"
```

---

### Type 3: Validators

**Purpose**: Complex validation logic

**Built-in Plugin**: `validators`

**Example**:
```yaml
x-cli-preflight:
  - name: validate-cluster-name
    type: plugin
    plugin: validators
    input:
      validator: cluster-name
      value: "{flags.cluster_name}"
      rules:
        - type: regex
          pattern: "^[a-z][a-z0-9-]{0,53}$"
        - type: custom
          function: check-dns-availability
    error-message: "Invalid cluster name"
```

---

### Type 4: Data Transformers

**Purpose**: Transform data between formats

**Built-in Plugin**: `transformers`

**Example**:
```yaml
x-cli-workflow:
  steps:
    - id: transform-htpasswd
      type: plugin
      plugin: transformers
      input:
        operation: htpasswd-to-users
        file: "{flags.htpasswd_file}"
      output:
        users: "{result.users}"

    - id: create-idp
      type: api-call
      endpoint: /api/v1/clusters/{cluster_id}/idps
      body:
        type: htpasswd
        users: "{transform-htpasswd.users}"
```

---

## Examples

### Example 1: AWS Integration (ROSA-like)

```yaml
# In OpenAPI spec
paths:
  /api/v1/clusters:
    post:
      operationId: createCluster
      x-cli-workflow:
        steps:
          # Step 1: Validate AWS credentials
          - id: check-aws
            type: plugin
            plugin: aws-cli
            command: validate
            input:
              profile: "{flags.aws_profile}"
              region: "{flags.region}"

          # Step 2: Check AWS quotas
          - id: check-quotas
            type: plugin
            plugin: aws-cli
            command: check-quotas
            input:
              service: ec2
              quotas:
                - vpc-count
                - elastic-ip-count
            skip-if: "{flags.skip_quota_check}"

          # Step 3: Create IAM roles via CloudFormation
          - id: create-roles
            type: plugin
            plugin: aws-cli
            command: cloudformation-deploy
            input:
              stack-name: "{flags.cluster_name}-roles"
              template: embedded://cf-templates/cluster-roles.yaml
              parameters:
                ClusterName: "{flags.cluster_name}"
            output:
              installer-role-arn: "{stack-outputs.InstallerRoleArn}"
              worker-role-arn: "{stack-outputs.WorkerRoleArn}"

          # Step 4: Create cluster via API
          - id: create-cluster
            type: api-call
            endpoint: /api/v1/clusters
            method: POST
            body:
              name: "{flags.cluster_name}"
              region: "{flags.region}"
              aws:
                installer_role_arn: "{create-roles.installer-role-arn}"
                worker_role_arn: "{create-roles.worker-role-arn}"
```

### Example 2: File-Based IDP Creation

```yaml
paths:
  /api/v1/clusters/{cluster_id}/identity_providers:
    post:
      x-cli-flags:
        - name: htpasswd-file
          flag: "--from-file"
          type: file
          description: "Path to htpasswd file"

      x-cli-workflow:
        steps:
          # Parse htpasswd file
          - id: parse-htpasswd
            type: plugin
            plugin: file-ops
            input:
              operation: parse
              file: "{flags.htpasswd_file}"
              format: htpasswd
            output:
              users: "{result.users}"

          # Create IDP with parsed users
          - id: create-idp
            type: api-call
            endpoint: /api/v1/clusters/{cluster_id}/identity_providers
            body:
              name: htpasswd
              type: htpasswd
              htpasswd:
                users: "{parse-htpasswd.users}"
```

---

## Implementation Notes

### Phase 1: Built-in Plugins (v0.8.0)

**Focus**: Core built-in plugins only

**Scope**:
- Plugin interface definition
- Plugin manager
- Built-in plugins: exec, file-ops, validators
- Permission system (basic)
- No external plugins yet

### Phase 2: External Plugins (v0.9.0)

**Focus**: Enable user-defined plugins

**Scope**:
- External binary plugin support
- JSON-RPC protocol
- Sandboxing (basic)
- Plugin discovery and loading
- Permission approval UI

### Phase 3: WASM & Advanced Features (v1.0.0)

**Focus**: Secure, portable plugins

**Scope**:
- WASM plugin runtime
- Plugin signing and verification
- Advanced sandboxing
- Plugin marketplace/registry
- Plugin development SDK

---

## Alternatives Considered

### Alternative 1: No Plugins (Pure OpenAPI)

**Pros**: Simpler, more secure
**Cons**: Cannot replicate ROSA functionality

**Decision**: Rejected - requirement for AWS integration is critical

### Alternative 2: Embedded Scripts (Lua, Starlark)

**Pros**: Secure, portable, simple
**Cons**: Limited capabilities, new language for users

**Decision**: Consider for v2.0, but binary plugins needed first

### Alternative 3: gRPC Plugins (HashiCorp model)

**Pros**: Well-proven, language-agnostic
**Cons**: Complex, requires network stack, heavier weight

**Decision**: Too complex for v1.0, consider for enterprise features

---

## Related Documents

- **Gap Analysis**: `gap-analysis-rosa-requirements.md`
- **Workflow Orchestration**: `workflow-orchestration.md` (to be created)
- **Technical Specification**: `../../docs/technical-specification.md` (to be updated)
- **ADR**: ADR-201 (to be created)

---

## Open Questions

1. Should built-in plugins be configurable or fixed?
2. How to handle plugin versioning and compatibility?
3. Should plugins have access to CLI context (current cluster, etc.)?
4. What's the update mechanism for external plugins?
5. Should there be a plugin development guide/SDK?

---

*⚒️ Forged with ❤️ by the CliForge team*
