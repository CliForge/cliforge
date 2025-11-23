# Workflow Orchestration Design

**Version**: 1.0.0
**Date**: 2025-11-23
**Status**: Proposed

---

## Table of Contents

1. [Overview](#overview)
2. [Requirements](#requirements)
3. [Architecture](#architecture)
4. [Step Types](#step-types)
5. [Dependencies & Control Flow](#dependencies--control-flow)
6. [Error Handling & Rollback](#error-handling--rollback)
7. [OpenAPI Extension](#openapi-extension)
8. [Examples](#examples)
9. [Implementation](#implementation)

---

## Overview

### Problem Statement

Enterprise CLI operations like ROSA's cluster creation involve complex multi-step workflows:
1. Pre-flight validation (AWS credentials, quotas, permissions)
2. Resource creation in specific order (IAM roles → OIDC provider → cluster)
3. Conditional execution based on previous results
4. Rollback on failures
5. Post-creation configuration

A single HTTP API call cannot represent this complexity. CliForge needs a workflow orchestration system.

### Solution

A declarative workflow engine that:
1. **Executes multi-step operations** defined in OpenAPI specs
2. **Handles dependencies** between steps
3. **Supports conditionals** for branching logic
4. **Manages rollback** on failures
5. **Integrates with plugins** for external tool execution
6. **Provides visibility** into execution progress

---

## Requirements

### Functional Requirements

**FR-1**: Execute steps sequentially or in parallel based on dependencies
**FR-2**: Support conditional execution based on previous step results
**FR-3**: Handle step failures with retry and rollback
**FR-4**: Pass data between steps using expressions
**FR-5**: Support multiple step types (API calls, plugin execution, loops, conditionals)
**FR-6**: Provide real-time progress feedback
**FR-7**: Allow workflow dry-run for validation

### Non-Functional Requirements

**NFR-1**: **Performance** - Parallel steps execute concurrently
**NFR-2**: **Reliability** - Partial execution state persisted for resume
**NFR-3**: **Debuggability** - Full execution trace for troubleshooting
**NFR-4**: **Simplicity** - Declarative YAML, no custom programming

---

## Architecture

### Workflow Execution Engine

```
┌─────────────────────────────────────────────────────────┐
│            Workflow Orchestration Engine                │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │              Workflow Parser                      │ │
│  ├───────────────────────────────────────────────────┤ │
│  │ • Parse x-cli-workflow from OpenAPI               │ │
│  │ • Build execution graph                           │ │
│  │ • Validate dependencies                           │ │
│  │ • Detect cycles                                   │ │
│  └───────────────────────────────────────────────────┘ │
│                          │                             │
│  ┌───────────────────────▼────────────────────────────┐│
│  │            Execution Coordinator                   ││
│  ├────────────────────────────────────────────────────┤│
│  │ • Schedule steps based on dependencies            ││
│  │ • Execute in parallel where possible              ││
│  │ • Evaluate conditions                             ││
│  │ • Handle retries                                  ││
│  │ • Trigger rollback on failures                    ││
│  └────────────────────────────────────────────────────┘│
│                          │                             │
│         ┌────────────────┴────────────────┐            │
│         │                                 │            │
│  ┌──────▼──────┐                  ┌──────▼──────┐     │
│  │    Step     │                  │   Context   │     │
│  │  Executors  │                  │   Manager   │     │
│  └─────────────┘                  └─────────────┘     │
│         │                                 │            │
│  ┌──────▼──────────┐              ┌──────▼──────────┐ │
│  │ - API Call      │              │ • Step outputs  │ │
│  │ - Plugin        │              │ • Variables     │ │
│  │ - Conditional   │              │ • Flags/args    │ │
│  │ - Loop          │              │ • State         │ │
│  │ - Sleep/Wait    │              └─────────────────┘ │
│  └─────────────────┘                                   │
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │              State Persistence                     │ │
│  ├───────────────────────────────────────────────────┤ │
│  │ • Save execution state                            │ │
│  │ • Enable resume on failure                        │ │
│  │ • Audit trail                                     │ │
│  └───────────────────────────────────────────────────┘ │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Workflow Lifecycle

```
Start Workflow
      │
      ▼
┌─────────────┐
│   Parse     │ ── Validate ──> Error: Invalid workflow
│  Workflow   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│    Build    │ ── Detect cycles ──> Error: Circular dependency
│ Exec Graph  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Execute    │
│   Steps     │
└──────┬──────┘
       │
       ├──> Step Success ──> Continue
       │
       ├──> Step Failure ──> Retry? ──Yes──> Retry
       │                        │
       │                       No
       │                        │
       │                        ▼
       │                   Rollback? ──Yes──> Execute Rollback Steps
       │                        │
       │                       No
       │                        │
       ▼                        ▼
 All Steps Done          Workflow Failed
       │
       ▼
┌─────────────┐
│  Cleanup    │
└──────┬──────┘
       │
       ▼
Workflow Complete
```

---

## Step Types

### 1. API Call Step

**Purpose**: Make HTTP request to API endpoint

**Definition**:
```yaml
- id: create-cluster
  type: api-call
  endpoint: /api/v1/clusters
  method: POST
  headers:
    X-Custom-Header: "{flags.custom_value}"
  body:
    name: "{flags.cluster_name}"
    region: "{flags.region}"
  output:
    cluster_id: "{response.id}"
    cluster_state: "{response.state}"
```

### 2. Plugin Step

**Purpose**: Execute plugin for external operations

**Definition**:
```yaml
- id: create-iam-roles
  type: plugin
  plugin: aws-cli
  command: cloudformation-deploy
  input:
    stack-name: "{flags.cluster_name}-roles"
    template: embedded://templates/roles.yaml
  output:
    role_arn: "{result.outputs.RoleArn}"
```

### 3. Conditional Step

**Purpose**: Execute steps conditionally

**Definition**:
```yaml
- id: check-multi-az
  type: conditional
  condition: "{flags.multi_az == true}"
  then:
    - id: validate-azs
      type: plugin
      plugin: aws-cli
      command: validate-availability-zones
  else:
    - id: use-single-az
      type: noop
```

### 4. Loop Step

**Purpose**: Iterate over collection

**Definition**:
```yaml
- id: create-machine-pools
  type: loop
  iterator: machine_pool
  collection: "{flags.machine_pools}"
  steps:
    - id: create-pool
      type: api-call
      endpoint: /api/v1/clusters/{cluster_id}/machine_pools
      body:
        name: "{machine_pool.name}"
        instance_type: "{machine_pool.instance_type}"
```

### 5. Wait/Sleep Step

**Purpose**: Delay execution or wait for condition

**Definition**:
```yaml
- id: wait-for-ready
  type: wait
  condition: "{cluster.state == 'ready'}"
  polling:
    endpoint: /api/v1/clusters/{cluster_id}
    interval: 30
    timeout: 3600
  output:
    final_state: "{response.state}"
```

### 6. Parallel Step

**Purpose**: Execute multiple steps concurrently

**Definition**:
```yaml
- id: setup-networking
  type: parallel
  steps:
    - id: create-vpc
      type: plugin
      plugin: aws-cli
      command: create-vpc

    - id: create-security-groups
      type: plugin
      plugin: aws-cli
      command: create-security-groups

    - id: create-subnets
      type: plugin
      plugin: aws-cli
      command: create-subnets

  # All parallel steps must complete before continuing
```

---

## Dependencies & Control Flow

### Dependency Declaration

**Explicit dependencies**:
```yaml
- id: step-b
  type: api-call
  depends-on: [step-a]  # Wait for step-a to complete
  endpoint: /api/v1/resource
```

**Implicit dependencies** (via output references):
```yaml
- id: step-a
  output:
    value: "{response.id}"

- id: step-b
  # Implicitly depends on step-a because it references step-a.value
  body:
    ref: "{step-a.value}"
```

### Execution Order

The engine builds a DAG (Directed Acyclic Graph) and executes:
1. **Parallel execution** for independent steps
2. **Sequential execution** for dependent steps
3. **Topological sort** determines optimal order

```
Example workflow:
  A (no dependencies)
  B (depends on A)
  C (depends on A)
  D (depends on B, C)

Execution:
  1. Execute A
  2. Execute B and C in parallel (both depend only on A)
  3. Execute D (after B and C complete)
```

### Conditional Execution

**Simple condition**:
```yaml
condition: "{flags.enable_feature == true}"
```

**Complex condition** (using expr language):
```yaml
condition: |
  flags.multi_az == true &&
  len(flags.availability_zones) >= 3 &&
  step_check_quotas.success == true
```

**Condition operators**:
- Comparison: `==`, `!=`, `<`, `>`, `<=`, `>=`
- Logical: `&&`, `||`, `!`
- Membership: `in`, `not in`
- Functions: `len()`, `has()`, `startsWith()`, etc.

---

## Error Handling & Rollback

### Retry Strategy

```yaml
- id: create-resource
  type: api-call
  endpoint: /api/v1/resources
  retry:
    max-attempts: 3
    backoff:
      type: exponential
      initial-interval: 1  # seconds
      multiplier: 2
      max-interval: 30
    retryable-errors:
      - http-status: 429  # Rate limit
      - http-status: 5xx  # Server errors
      - error-type: network-timeout
```

### Rollback on Failure

Each step can define a rollback action:

```yaml
- id: create-iam-role
  type: plugin
  plugin: aws-cli
  command: iam-create-role
  input:
    role-name: "{flags.cluster_name}-installer"
  output:
    role_arn: "{result.RoleArn}"

  # Rollback: delete the role if workflow fails later
  rollback:
    type: plugin
    plugin: aws-cli
    command: iam-delete-role
    input:
      role-name: "{flags.cluster_name}-installer"
```

**Rollback execution**:
- Triggered when any step fails (after retries exhausted)
- Executed in **reverse order** of step execution
- Only rolls back steps that completed successfully
- Rollback failures logged but don't stop other rollbacks

**Example workflow with rollback**:
```
Steps executed:
  1. create-iam-role ✓
  2. create-oidc-provider ✓
  3. create-cluster ✗ (FAILED)

Rollback sequence:
  1. Rollback create-oidc-provider
  2. Rollback create-iam-role
  3. Report workflow failure
```

### Error Context

```yaml
# Step fails
error:
  step-id: create-cluster
  type: api-error
  http-status: 400
  message: "Invalid cluster name"
  retry-count: 3
  rollback-executed: true
  rollback-status:
    - step: create-oidc-provider
      status: success
    - step: create-iam-role
      status: success
```

---

## OpenAPI Extension

### `x-cli-workflow`

**Location**: Operation object
**Purpose**: Define multi-step workflow for operation

**Schema**:
```yaml
x-cli-workflow:
  # List of steps to execute
  steps:
    - id: string               # Unique step identifier
      type: string             # api-call | plugin | conditional | loop | wait | parallel
      description: string      # Step description (for progress display)

      # Execution control
      depends-on: [string]     # List of step IDs this depends on
      condition: string        # Expression to evaluate (skip if false)
      required: boolean        # If false, failure doesn't stop workflow

      # Retry configuration
      retry:
        max-attempts: integer
        backoff:
          type: string         # fixed | linear | exponential
          initial-interval: integer
          multiplier: number
          max-interval: integer

      # Rollback action
      rollback:
        type: string           # Same types as steps
        # ... rollback step definition

      # Step-specific fields
      # (varies by type)

      # Output mapping
      output:
        variable-name: expression

  # Workflow-level settings
  settings:
    parallel-execution: boolean  # Allow parallel step execution
    fail-fast: boolean           # Stop on first failure
    timeout: integer             # Total workflow timeout (seconds)
    dry-run-supported: boolean   # Can simulate without executing
```

### Full Example

```yaml
paths:
  /api/v1/clusters:
    post:
      operationId: createCluster
      summary: Create a new ROSA-like cluster

      x-cli-workflow:
        settings:
          parallel-execution: true
          fail-fast: false  # Try all pre-flights even if one fails
          timeout: 7200  # 2 hours
          dry-run-supported: true

        steps:
          # ===== PRE-FLIGHT CHECKS =====
          - id: check-aws-credentials
            type: plugin
            description: "Verifying AWS credentials"
            plugin: aws-cli
            command: validate-credentials
            input:
              profile: "{flags.aws_profile}"
            output:
              aws_account: "{result.account_id}"

          - id: check-aws-quotas
            type: plugin
            description: "Checking AWS quotas"
            plugin: aws-cli
            command: check-quotas
            input:
              service: ec2
              region: "{flags.region}"
            condition: "!flags.skip_quota_check"

          - id: check-permissions
            type: plugin
            description: "Validating AWS permissions"
            plugin: aws-cli
            command: check-scp-policies

          # ===== IAM ROLE CREATION =====
          - id: create-installer-role
            type: plugin
            description: "Creating installer IAM role"
            depends-on: [check-aws-credentials]
            plugin: aws-cli
            command: iam-create-role
            input:
              role-name: "{flags.cluster_name}-installer"
              trust-policy: embedded://policies/installer-trust.json
            output:
              installer_role_arn: "{result.RoleArn}"
            rollback:
              type: plugin
              plugin: aws-cli
              command: iam-delete-role
              input:
                role-name: "{flags.cluster_name}-installer"

          - id: create-worker-role
            type: plugin
            description: "Creating worker IAM role"
            depends-on: [check-aws-credentials]
            plugin: aws-cli
            command: iam-create-role
            input:
              role-name: "{flags.cluster_name}-worker"
              trust-policy: embedded://policies/worker-trust.json
            output:
              worker_role_arn: "{result.RoleArn}"
            rollback:
              type: plugin
              plugin: aws-cli
              command: iam-delete-role
              input:
                role-name: "{flags.cluster_name}-worker"

          # ===== OIDC PROVIDER =====
          - id: create-oidc-provider
            type: plugin
            description: "Creating OIDC provider"
            depends-on: [create-installer-role]
            plugin: aws-cli
            command: iam-create-oidc-provider
            output:
              oidc_provider_arn: "{result.ProviderArn}"
            rollback:
              type: plugin
              plugin: aws-cli
              command: iam-delete-oidc-provider
              input:
                arn: "{oidc_provider_arn}"

          # ===== CLUSTER CREATION =====
          - id: create-cluster-api
            type: api-call
            description: "Creating cluster via API"
            depends-on:
              - create-installer-role
              - create-worker-role
              - create-oidc-provider
            endpoint: /api/v1/clusters
            method: POST
            body:
              name: "{flags.cluster_name}"
              region: "{flags.region}"
              version: "{flags.version}"
              multi_az: "{flags.multi_az}"
              aws:
                account_id: "{check-aws-credentials.aws_account}"
                installer_role_arn: "{create-installer-role.installer_role_arn}"
                worker_role_arn: "{create-worker-role.worker_role_arn}"
                oidc_provider_arn: "{create-oidc-provider.oidc_provider_arn}"
            output:
              cluster_id: "{response.id}"
              cluster_state: "{response.state}"
            retry:
              max-attempts: 3
              backoff:
                type: exponential
                initial-interval: 5
                multiplier: 2

          # ===== WAIT FOR CLUSTER READY =====
          - id: wait-for-ready
            type: wait
            description: "Waiting for cluster to be ready"
            depends-on: [create-cluster-api]
            condition: "{create-cluster-api.cluster_state != 'ready'}"
            polling:
              endpoint: /api/v1/clusters/{create-cluster-api.cluster_id}
              interval: 30
              timeout: 3600
              terminal-states: [ready, error]
              status-field: state
            output:
              final_state: "{response.state}"

          # ===== POST-CREATION =====
          - id: create-default-ingress
            type: api-call
            description: "Creating default ingress"
            depends-on: [wait-for-ready]
            condition: "{wait-for-ready.final_state == 'ready'}"
            endpoint: /api/v1/clusters/{create-cluster-api.cluster_id}/ingresses
            method: POST
            body:
              default: true
              listening: external
```

---

## Examples

See full examples in the OpenAPI extension section above and in:
- `hack/rosa-cli/alpha-omega-rosa-spec.yaml`
- `hack/rosa-cli/openapi-cli-extensions-spec.md`

---

## Implementation

### Phase 1: Core Engine (v0.8.0)

**Scope**:
- Workflow parser
- Basic step types (api-call, plugin)
- Sequential execution
- Simple conditions
- Basic error handling

**Out of scope**:
- Parallel execution
- Rollback
- Complex conditions
- Loops

### Phase 2: Advanced Features (v0.9.0)

**Scope**:
- Parallel execution
- Rollback mechanism
- Retry with backoff
- Loop and conditional steps
- Wait/polling

### Phase 3: Enterprise Features (v1.0.0)

**Scope**:
- State persistence and resume
- Distributed tracing
- Dry-run mode
- Workflow templates
- Advanced error recovery

---

## Related Documents

- **Plugin Architecture**: `plugin-architecture.md`
- **Gap Analysis**: `gap-analysis-rosa-requirements.md`
- **Technical Specification**: `../../docs/technical-specification.md`
- **ADR-101**: Use expr for Templating
- **ADR-202**: Workflow Orchestration Engine (to be created)

---

*⚒️ Forged with ❤️ by the CliForge team*
