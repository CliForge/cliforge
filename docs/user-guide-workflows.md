# Workflow User Guide

**Version**: 0.9.0
**Last Updated**: 2025-11-23

---

## Table of Contents

1. [Introduction](#introduction)
2. [What Are Workflows?](#what-are-workflows)
3. [When to Use Workflows](#when-to-use-workflows)
4. [Step Types](#step-types)
5. [Dependencies and Execution Order](#dependencies-and-execution-order)
6. [Conditions and Expressions](#conditions-and-expressions)
7. [Error Handling and Retry](#error-handling-and-retry)
8. [Rollback on Failure](#rollback-on-failure)
9. [State Persistence and Resume](#state-persistence-and-resume)
10. [Real-World Examples](#real-world-examples)
11. [Debugging Workflows](#debugging-workflows)
12. [Best Practices](#best-practices)

---

## Introduction

Workflows in CliForge v0.9.0 enable you to orchestrate complex, multi-step operations that go beyond simple API calls. This guide teaches you how to define, execute, and debug workflows for your CLI applications.

### Key Benefits

- **Multi-step operations**: Chain together API calls, plugin executions, and conditional logic
- **Dependency management**: Control execution order with explicit or implicit dependencies
- **Error resilience**: Built-in retry mechanisms and rollback support
- **Parallel execution**: Run independent steps concurrently for better performance
- **State management**: Resume failed workflows from the last successful step

---

## What Are Workflows?

A workflow is a declarative specification of multiple steps that should be executed in a coordinated manner. Each workflow is defined in your OpenAPI specification using the `x-cli-workflow` extension.

### Basic Structure

```yaml
paths:
  /api/v1/resources:
    post:
      operationId: createResource
      x-cli-workflow:
        settings:
          parallel-execution: true
          fail-fast: false
          timeout: 3600
          dry-run-supported: true

        steps:
          - id: step-1
            type: api-call
            description: "Creating resource"
            # ... step configuration

          - id: step-2
            type: plugin
            depends-on: [step-1]
            description: "Configuring resource"
            # ... step configuration
```

### Workflow Components

- **Settings**: Workflow-level configuration (timeout, parallel execution, etc.)
- **Steps**: Individual operations to execute
- **Dependencies**: Relationships between steps that control execution order
- **Outputs**: Data passed between steps

---

## When to Use Workflows

Use workflows when your operation requires:

### 1. Pre-flight Validation

Before creating a resource, validate prerequisites:

```yaml
steps:
  - id: check-credentials
    type: plugin
    plugin: aws-cli
    command: validate-credentials

  - id: check-quotas
    type: plugin
    plugin: aws-cli
    command: check-quotas
    depends-on: [check-credentials]

  - id: create-resource
    type: api-call
    endpoint: /api/v1/resources
    depends-on: [check-credentials, check-quotas]
```

### 2. Multi-Stage Resource Creation

Create resources in the correct order:

```yaml
steps:
  - id: create-iam-role
    type: plugin
    description: "Creating IAM role"

  - id: create-oidc-provider
    type: plugin
    description: "Creating OIDC provider"
    depends-on: [create-iam-role]

  - id: create-cluster
    type: api-call
    description: "Creating cluster"
    depends-on: [create-iam-role, create-oidc-provider]
```

### 3. Conditional Operations

Execute steps based on runtime conditions:

```yaml
steps:
  - id: check-mode
    type: conditional
    condition: "{flags.high_availability == true}"
    then:
      - id: configure-ha
        type: plugin
        plugin: setup-ha
    else:
      - id: configure-single
        type: plugin
        plugin: setup-single
```

### 4. Waiting for Asynchronous Operations

Poll until a resource reaches a desired state:

```yaml
steps:
  - id: create-cluster
    type: api-call
    endpoint: /api/v1/clusters

  - id: wait-ready
    type: wait
    description: "Waiting for cluster to be ready"
    condition: "{cluster.state == 'ready'}"
    polling:
      endpoint: /api/v1/clusters/{cluster_id}
      interval: 30
      timeout: 3600
```

### 5. Batch Operations

Process multiple items:

```yaml
steps:
  - id: create-machine-pools
    type: loop
    iterator: pool
    collection: "{flags.machine_pools}"
    steps:
      - id: create-pool
        type: api-call
        endpoint: /api/v1/machine_pools
        body:
          name: "{pool.name}"
          instance_type: "{pool.instance_type}"
```

---

## Step Types

CliForge supports six step types, each designed for specific use cases.

### 1. API Call Step

Execute an HTTP request to your API.

**Use when**: Making REST API calls

**Configuration**:
```yaml
- id: create-cluster
  type: api-call
  description: "Creating cluster via API"
  endpoint: /api/v1/clusters
  method: POST
  headers:
    X-Custom-Header: "{flags.custom_value}"
  query:
    region: "{flags.region}"
  body:
    name: "{flags.cluster_name}"
    version: "{flags.version}"
  output:
    cluster_id: "{response.id}"
    cluster_state: "{response.state}"
```

**Fields**:
- `endpoint` (required): API endpoint path
- `method` (required): HTTP method (GET, POST, PUT, DELETE, PATCH)
- `headers` (optional): Request headers with expression support
- `query` (optional): Query parameters
- `body` (optional): Request body
- `output` (optional): Map response fields to workflow variables

**Expression Support**: All fields support `{expression}` syntax for dynamic values.

### 2. Plugin Step

Execute a plugin for operations outside the API.

**Use when**: Interacting with external tools, files, or custom logic

**Configuration**:
```yaml
- id: create-iam-roles
  type: plugin
  description: "Creating IAM roles via CloudFormation"
  plugin: aws-cli
  command: cloudformation-deploy
  input:
    stack-name: "{flags.cluster_name}-roles"
    template: embedded://templates/roles.yaml
    parameters:
      ClusterName: "{flags.cluster_name}"
  output:
    role_arn: "{result.outputs.RoleArn}"
```

**Fields**:
- `plugin` (required): Plugin name (e.g., `aws-cli`, `file-ops`)
- `command` (required): Plugin command to execute
- `input` (optional): Input data for the plugin
- `output` (optional): Map plugin output to workflow variables

**Available Built-in Plugins**: `exec`, `file-ops`, `validators`, `transformers`

### 3. Conditional Step

Execute steps conditionally based on runtime evaluation.

**Use when**: Branching logic based on flags, previous step results, or environment

**Configuration**:
```yaml
- id: check-multi-az
  type: conditional
  description: "Checking multi-AZ configuration"
  condition: "{flags.multi_az == true}"
  then:
    - id: validate-azs
      type: plugin
      plugin: aws-cli
      command: validate-availability-zones
      input:
        region: "{flags.region}"
        required-count: 3
  else:
    - id: skip-az-check
      type: noop
      description: "Single-AZ cluster, skipping validation"
```

**Fields**:
- `condition` (required): Expression that evaluates to boolean
- `then` (required): Steps to execute if condition is true
- `else` (optional): Steps to execute if condition is false

**Condition Examples**:
```yaml
# Simple boolean check
condition: "{flags.enable_feature == true}"

# Numeric comparison
condition: "{flags.node_count > 3}"

# String comparison
condition: "{flags.region == 'us-east-1'}"

# Complex expression
condition: "{flags.multi_az == true && len(flags.availability_zones) >= 3}"
```

### 4. Loop Step

Iterate over a collection and execute steps for each item.

**Use when**: Processing arrays or multiple items

**Configuration**:
```yaml
- id: create-machine-pools
  type: loop
  description: "Creating machine pools"
  iterator: machine_pool
  collection: "{flags.machine_pools}"
  steps:
    - id: create-pool
      type: api-call
      description: "Creating pool {machine_pool.name}"
      endpoint: /api/v1/clusters/{cluster_id}/machine_pools
      method: POST
      body:
        name: "{machine_pool.name}"
        instance_type: "{machine_pool.instance_type}"
        replicas: "{machine_pool.replicas}"
      output:
        pool_id: "{response.id}"
```

**Fields**:
- `iterator` (required): Variable name for current item
- `collection` (required): Expression that evaluates to an array
- `steps` (required): Steps to execute for each item

**Iterator Access**: Within loop steps, access the current item via `{iterator_name.field}`.

### 5. Wait Step

Pause execution or poll until a condition is met.

**Use when**: Waiting for asynchronous operations to complete

**Configuration**:

**Simple delay**:
```yaml
- id: wait-30-seconds
  type: wait
  description: "Waiting 30 seconds"
  duration: 30
```

**Polling with condition**:
```yaml
- id: wait-for-ready
  type: wait
  description: "Waiting for cluster to be ready"
  condition: "{cluster.state == 'ready'}"
  polling:
    endpoint: /api/v1/clusters/{cluster_id}
    interval: 30
    timeout: 3600
    terminal-states: [ready, error, failed]
    status-field: state
  output:
    final_state: "{response.state}"
```

**Fields**:
- `duration` (optional): Fixed wait time in seconds
- `condition` (optional): Expression to evaluate
- `polling` (optional): Polling configuration
  - `endpoint`: API endpoint to poll
  - `interval`: Seconds between polls
  - `timeout`: Maximum wait time
  - `terminal-states`: States that stop polling
  - `status-field`: Response field containing state

### 6. Parallel Step

Execute multiple steps concurrently.

**Use when**: Independent operations that can run simultaneously

**Configuration**:
```yaml
- id: setup-networking
  type: parallel
  description: "Setting up network infrastructure"
  steps:
    - id: create-vpc
      type: plugin
      plugin: aws-cli
      command: create-vpc
      input:
        cidr: "10.0.0.0/16"

    - id: create-security-groups
      type: plugin
      plugin: aws-cli
      command: create-security-groups

    - id: create-subnets
      type: plugin
      plugin: aws-cli
      command: create-subnets
```

**Fields**:
- `steps` (required): Array of steps to execute in parallel

**Important**: All parallel steps must complete before workflow continues. If any step fails, behavior depends on the `fail-fast` setting.

---

## Dependencies and Execution Order

CliForge builds a Directed Acyclic Graph (DAG) from your workflow steps and executes them in the optimal order.

### Explicit Dependencies

Use `depends-on` to specify that a step requires others to complete first:

```yaml
steps:
  - id: step-a
    type: api-call
    # ... configuration

  - id: step-b
    type: api-call
    depends-on: [step-a]  # Wait for step-a
    # ... configuration

  - id: step-c
    type: api-call
    depends-on: [step-a, step-b]  # Wait for both
    # ... configuration
```

### Implicit Dependencies

Dependencies are automatically detected when you reference another step's output:

```yaml
steps:
  - id: create-cluster
    type: api-call
    output:
      cluster_id: "{response.id}"

  - id: create-ingress
    type: api-call
    # Implicitly depends on create-cluster because it references cluster_id
    endpoint: /api/v1/clusters/{create-cluster.cluster_id}/ingresses
```

### Execution Order Example

```yaml
steps:
  # Level 0: No dependencies
  - id: step-a
    type: api-call

  # Level 1: Depends on step-a
  - id: step-b
    type: api-call
    depends-on: [step-a]

  - id: step-c
    type: api-call
    depends-on: [step-a]

  # Level 2: Depends on step-b and step-c
  - id: step-d
    type: api-call
    depends-on: [step-b, step-c]
```

**Execution**:
1. Execute `step-a`
2. Execute `step-b` and `step-c` in parallel (if `parallel-execution: true`)
3. Execute `step-d` after both complete

### Parallel Execution

Enable parallel execution in workflow settings:

```yaml
x-cli-workflow:
  settings:
    parallel-execution: true  # Enable concurrent execution
```

**When enabled**:
- Steps at the same dependency level run concurrently
- Independent steps execute in parallel
- Improves performance for I/O-bound operations

**When disabled**:
- All steps execute sequentially
- Useful for debugging or when order matters

---

## Conditions and Expressions

CliForge uses the [expr](https://github.com/expr-lang/expr) language for dynamic expressions.

### Expression Syntax

Expressions are wrapped in `{...}`:

```yaml
endpoint: /api/v1/clusters/{cluster_id}
body:
  name: "{flags.cluster_name}"
  enabled: "{flags.enable_feature}"
```

### Available Variables

**`flags`**: CLI flags and arguments
```yaml
"{flags.cluster_name}"
"{flags.region}"
"{flags.node_count}"
```

**`response`**: Current step's API response
```yaml
"{response.id}"
"{response.state}"
"{response.metadata.created_at}"
```

**`step-id`**: Previous step's output
```yaml
"{create-cluster.cluster_id}"
"{validate-creds.aws_account}"
```

**`env`**: Environment variables
```yaml
"{env.AWS_PROFILE}"
"{env.HOME}"
```

### Operators

**Comparison**:
```yaml
"{flags.count > 3}"
"{flags.region == 'us-east-1'}"
"{flags.version != '1.0.0'}"
"{flags.count <= 10}"
```

**Logical**:
```yaml
"{flags.enabled && flags.verified}"
"{flags.skip_validation || flags.force}"
"{!flags.disabled}"
```

**Arithmetic**:
```yaml
"{flags.count + 1}"
"{flags.total - flags.used}"
"{flags.count * 2}"
```

**String**:
```yaml
"{flags.name + '-suffix'}"
"{flags.region in ['us-east-1', 'us-west-2']}"
```

### Functions

**`len()`**: Get collection length
```yaml
"{len(flags.availability_zones) >= 3}"
```

**`has()`**: Check if map has key
```yaml
"{has(response, 'error')}"
```

**`startsWith()`, `endsWith()`, `contains()`**:
```yaml
"{startsWith(flags.name, 'prod-')}"
"{endsWith(flags.name, '-cluster')}"
"{contains(flags.name, 'test')}"
```

### Complex Conditions

```yaml
condition: |
  flags.multi_az == true &&
  len(flags.availability_zones) >= 3 &&
  !flags.skip_validation &&
  validate_quotas.success == true
```

---

## Error Handling and Retry

### Retry Configuration

Configure automatic retries for transient failures:

```yaml
- id: create-resource
  type: api-call
  endpoint: /api/v1/resources
  retry:
    max-attempts: 3
    backoff:
      type: exponential
      initial-interval: 1
      multiplier: 2
      max-interval: 30
    retryable-errors:
      - http-status: 429  # Rate limit
      - http-status: 5xx  # Server errors
      - error-type: network-timeout
```

**Retry Configuration**:
- `max-attempts`: Maximum retry attempts (default: 1, no retry)
- `backoff.type`: Backoff strategy (`fixed`, `linear`, `exponential`)
- `backoff.initial-interval`: Initial wait time in seconds
- `backoff.multiplier`: Multiplier for exponential backoff
- `backoff.max-interval`: Maximum wait time between retries
- `retryable-errors`: Conditions that trigger retry

### Backoff Strategies

**Fixed**: Same delay between retries
```yaml
backoff:
  type: fixed
  initial-interval: 5  # Always wait 5 seconds
```

**Linear**: Incremental increase
```yaml
backoff:
  type: linear
  initial-interval: 2  # 2s, 4s, 6s, 8s...
  multiplier: 2
```

**Exponential**: Exponential increase (recommended)
```yaml
backoff:
  type: exponential
  initial-interval: 1  # 1s, 2s, 4s, 8s, 16s, 30s...
  multiplier: 2
  max-interval: 30
```

### Error Matching

Specify which errors should trigger retry:

**HTTP Status Codes**:
```yaml
retryable-errors:
  - http-status: 429  # Specific status
  - http-status: 5xx  # Pattern matching
  - http-status: 503
```

**Error Types**:
```yaml
retryable-errors:
  - error-type: network-timeout
  - error-type: connection-refused
  - error-type: plugin-execution-error
```

### Required vs Optional Steps

Control whether step failure stops the workflow:

```yaml
- id: optional-validation
  type: plugin
  plugin: validators
  required: false  # Failure won't stop workflow

- id: critical-operation
  type: api-call
  required: true  # Failure stops workflow (default)
```

### Fail-Fast Behavior

Configure workflow-level failure handling:

```yaml
settings:
  fail-fast: true  # Stop on first failure
  # OR
  fail-fast: false  # Continue executing, collect all errors
```

---

## Rollback on Failure

When a workflow fails, CliForge can automatically rollback completed steps.

### Defining Rollback Actions

Each step can specify a rollback action:

```yaml
- id: create-iam-role
  type: plugin
  plugin: aws-cli
  command: iam-create-role
  input:
    role-name: "{flags.cluster_name}-installer"
  output:
    role_arn: "{result.RoleArn}"

  # Rollback: delete the role if workflow fails
  rollback:
    type: plugin
    plugin: aws-cli
    command: iam-delete-role
    input:
      role-name: "{flags.cluster_name}-installer"
```

### Rollback Execution

**Trigger**: Rollback executes when:
- A required step fails after exhausting retries
- Workflow timeout is exceeded
- User cancels workflow execution

**Order**: Rollback steps execute in **reverse order** of creation:
```
Steps executed:
  1. create-vpc ✓
  2. create-security-group ✓
  3. create-cluster ✗ (FAILED)

Rollback sequence:
  1. Rollback create-security-group
  2. Rollback create-vpc
```

**Behavior**:
- Only successful steps are rolled back
- Rollback failures are logged but don't stop other rollbacks
- Workflow state is marked as `rolled-back` if successful

### Rollback Example

```yaml
x-cli-workflow:
  steps:
    - id: create-vpc
      type: plugin
      plugin: aws-cli
      command: create-vpc
      output:
        vpc_id: "{result.VpcId}"
      rollback:
        type: plugin
        plugin: aws-cli
        command: delete-vpc
        input:
          vpc-id: "{create-vpc.vpc_id}"

    - id: create-security-group
      type: plugin
      plugin: aws-cli
      command: create-security-group
      depends-on: [create-vpc]
      input:
        vpc-id: "{create-vpc.vpc_id}"
      output:
        sg_id: "{result.GroupId}"
      rollback:
        type: plugin
        plugin: aws-cli
        command: delete-security-group
        input:
          group-id: "{create-security-group.sg_id}"

    - id: create-cluster
      type: api-call
      endpoint: /api/v1/clusters
      depends-on: [create-vpc, create-security-group]
      body:
        vpc_id: "{create-vpc.vpc_id}"
        security_group_id: "{create-security-group.sg_id}"
      # No rollback needed - cluster deletion handled by API
```

### Partial Rollback

If rollback fails partway through, you'll see detailed output:

```
Workflow failed: step create-cluster failed
Executing rollback...
  ✓ Rolled back create-security-group
  ✗ Rollback failed for create-vpc: VPC still has dependencies

Manual cleanup may be required for:
  - VPC: vpc-12345678
```

---

## State Persistence and Resume

CliForge can save workflow state to enable resuming after failures.

### State Management

**Automatic State Saving**:
- State is saved after each workflow level completes
- Includes all completed steps and their outputs
- Stored in `~/.config/<cli-name>/workflow-state/`

**State Contents**:
```yaml
workflow_id: workflow-1732387200
status: failed
start_time: 2025-11-23T10:00:00Z
current_step: create-cluster
completed_steps:
  - step_id: check-credentials
    success: true
    output:
      aws_account: "123456789012"
  - step_id: create-iam-role
    success: true
    output:
      role_arn: "arn:aws:iam::123456789012:role/..."
error:
  step: create-cluster
  message: "API request failed: 503 Service Unavailable"
```

### Resuming Workflows

**Manual Resume** (future feature):
```bash
mycli cluster create --resume workflow-1732387200
```

**Auto-Resume** (future feature):
```yaml
settings:
  auto-resume: true
  resume-on-errors: [network-timeout, rate-limit]
```

### State Lifecycle

1. **Workflow Start**: Create new state with `pending` status
2. **During Execution**: Update state after each level completes
3. **On Success**: Mark state as `completed`, retain for audit
4. **On Failure**: Mark state as `failed`, enable resume
5. **On Rollback**: Mark state as `rolled-back`

### State Cleanup

States are retained for debugging and audit purposes:

```yaml
x-cli-config:
  workflow:
    state-retention-days: 7  # Delete states older than 7 days
    max-states: 100  # Keep only 100 most recent states
```

---

## Real-World Examples

### Example 1: ROSA-like Cluster Creation

Complete workflow for creating a managed Kubernetes cluster:

```yaml
paths:
  /api/v1/clusters:
    post:
      operationId: createCluster
      summary: Create a new cluster

      x-cli-workflow:
        settings:
          parallel-execution: true
          fail-fast: false
          timeout: 7200
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
            retry:
              max-attempts: 2

          - id: check-aws-quotas
            type: plugin
            description: "Checking AWS quotas"
            plugin: aws-cli
            command: check-quotas
            input:
              service: ec2
              region: "{flags.region}"
            condition: "!flags.skip_quota_check"
            required: false

          - id: check-permissions
            type: plugin
            description: "Validating AWS permissions"
            plugin: aws-cli
            command: check-scp-policies
            depends-on: [check-aws-credentials]

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

### Example 2: File-Based Identity Provider Creation

Workflow that processes a local file and creates API resources:

```yaml
paths:
  /api/v1/clusters/{cluster_id}/identity_providers:
    post:
      operationId: createIdentityProvider
      summary: Create identity provider from htpasswd file

      parameters:
        - name: cluster_id
          in: path
          required: true
          schema:
            type: string

      x-cli-flags:
        - name: htpasswd-file
          flag: "--from-file"
          type: file
          required: true
          description: "Path to htpasswd file"

      x-cli-workflow:
        steps:
          # Parse htpasswd file
          - id: parse-htpasswd
            type: plugin
            description: "Parsing htpasswd file"
            plugin: file-ops
            command: parse
            input:
              operation: parse
              file: "{flags.htpasswd_file}"
              format: htpasswd
            output:
              users: "{result.users}"
              user_count: "{result.count}"

          # Validate user count
          - id: validate-users
            type: conditional
            depends-on: [parse-htpasswd]
            condition: "{parse-htpasswd.user_count == 0}"
            then:
              - id: error-no-users
                type: noop
                description: "Error: No users found in htpasswd file"
            else:
              - id: continue
                type: noop

          # Create IDP with parsed users
          - id: create-idp
            type: api-call
            description: "Creating identity provider"
            depends-on: [validate-users]
            endpoint: /api/v1/clusters/{cluster_id}/identity_providers
            method: POST
            body:
              name: htpasswd
              type: htpasswd
              htpasswd:
                users: "{parse-htpasswd.users}"
```

### Example 3: Multi-Environment Deployment

Deploy to multiple environments in sequence:

```yaml
x-cli-workflow:
  steps:
    - id: deploy-to-dev
      type: api-call
      description: "Deploying to development"
      endpoint: /api/v1/deployments
      body:
        environment: dev
        version: "{flags.version}"
      output:
        dev_deployment_id: "{response.id}"

    - id: wait-dev-ready
      type: wait
      description: "Waiting for dev deployment"
      depends-on: [deploy-to-dev]
      polling:
        endpoint: /api/v1/deployments/{deploy-to-dev.dev_deployment_id}
        interval: 10
        timeout: 300
        terminal-states: [ready, failed]
        status-field: status

    - id: run-integration-tests
      type: plugin
      description: "Running integration tests"
      depends-on: [wait-dev-ready]
      condition: "{wait-dev-ready.final_state == 'ready'}"
      plugin: exec
      input:
        command: npm
        args: [test, --env=dev]

    - id: deploy-to-staging
      type: conditional
      depends-on: [run-integration-tests]
      condition: "{flags.deploy_staging == true && run-integration-tests.exit_code == 0}"
      then:
        - id: staging-deployment
          type: api-call
          endpoint: /api/v1/deployments
          body:
            environment: staging
            version: "{flags.version}"
```

---

## Debugging Workflows

### Dry Run Mode

Execute workflow without side effects:

```yaml
settings:
  dry-run-supported: true
```

```bash
mycli cluster create --cluster-name test --dry-run
```

**Output**:
```
[DRY RUN] Workflow execution plan:
  Level 0:
    ✓ check-aws-credentials
    ✓ check-aws-quotas
  Level 1:
    ✓ create-installer-role
    ✓ create-worker-role
  Level 2:
    ✓ create-oidc-provider
  Level 3:
    ✓ create-cluster-api
  Level 4:
    ✓ wait-for-ready
  Level 5:
    ✓ create-default-ingress

Total steps: 7
Estimated time: ~45 minutes
```

### Verbose Output

Enable detailed execution logs:

```bash
mycli cluster create --cluster-name test --verbose
```

**Output**:
```
[10:00:00] Starting workflow: createCluster
[10:00:00] Level 0: Executing 2 steps in parallel
[10:00:01]   ✓ check-aws-credentials (1.2s)
[10:00:01]     Output: aws_account=123456789012
[10:00:02]   ✓ check-aws-quotas (1.8s)
[10:00:02] Level 1: Executing 2 steps in parallel
[10:00:03]   ✓ create-installer-role (1.1s)
[10:00:03]     Output: installer_role_arn=arn:aws:iam::...
[10:00:04]   ✓ create-worker-role (1.3s)
[10:00:04] Level 2: Executing 1 step
[10:00:05]   ✓ create-oidc-provider (1.0s)
...
```

### Step-Level Debugging

Add debug output to steps:

```yaml
- id: debug-step
  type: plugin
  plugin: exec
  input:
    command: echo
    args:
      - "Cluster ID: {create-cluster.cluster_id}"
      - "State: {create-cluster.cluster_state}"
```

### Workflow Validation

Validate workflow before execution:

```bash
mycli workflow validate petstore-api.yaml
```

**Output**:
```
✓ Workflow is valid
  - 15 steps defined
  - 0 circular dependencies
  - 5 dependency levels
  - Estimated execution time: 30-45 minutes

Warnings:
  - Step 'optional-validation' is marked as optional
  - No rollback defined for step 'create-cluster'
```

### Common Issues

**Circular Dependencies**:
```
Error: Circular dependency detected: step-a → step-b → step-c → step-a
```

**Solution**: Review `depends-on` declarations and remove cycles.

**Missing Outputs**:
```
Error: Step 'create-ingress' references undefined output: create-cluster.cluster_id
```

**Solution**: Ensure referenced step defines output mapping.

**Invalid Expressions**:
```
Error: Invalid expression in step 'check-multi-az': unexpected token '='
  condition: "{flags.multi_az = true}"
                               ^
```

**Solution**: Use `==` for comparison, not `=`.

---

## Best Practices

### 1. Use Descriptive Step IDs and Descriptions

**Good**:
```yaml
- id: create-installer-iam-role
  description: "Creating installer IAM role with CloudFormation"
```

**Bad**:
```yaml
- id: step-1
  description: "Creating stuff"
```

### 2. Add Rollback to Destructive Operations

Always define rollback for resource creation:

```yaml
- id: create-resource
  type: api-call
  endpoint: /api/v1/resources
  rollback:
    type: api-call
    endpoint: /api/v1/resources/{create-resource.resource_id}
    method: DELETE
```

### 3. Use Retry for Transient Failures

Configure retry for network operations:

```yaml
retry:
  max-attempts: 3
  backoff:
    type: exponential
    initial-interval: 1
    multiplier: 2
  retryable-errors:
    - http-status: 429
    - http-status: 5xx
```

### 4. Validate Early

Put validation steps at the beginning:

```yaml
steps:
  - id: validate-credentials
    type: plugin
    # ... validation

  - id: validate-quotas
    type: plugin
    # ... validation

  - id: create-resources
    type: api-call
    depends-on: [validate-credentials, validate-quotas]
```

### 5. Use Parallel Execution

Enable parallel execution for independent steps:

```yaml
settings:
  parallel-execution: true

steps:
  # These run in parallel
  - id: create-vpc
    type: plugin

  - id: create-security-groups
    type: plugin

  - id: create-subnets
    type: plugin
```

### 6. Set Reasonable Timeouts

Configure timeouts to prevent hanging:

```yaml
settings:
  timeout: 3600  # 1 hour total workflow timeout

steps:
  - id: wait-for-ready
    type: wait
    polling:
      timeout: 1800  # 30 minutes for this step
```

### 7. Make Workflows Idempotent

Design steps to be safely re-runnable:

```yaml
- id: create-role
  type: plugin
  plugin: aws-cli
  command: iam-create-role-if-not-exists  # Idempotent operation
  input:
    role-name: "{flags.cluster_name}-installer"
```

### 8. Use Conditions for Optional Features

Make optional features conditional:

```yaml
- id: enable-monitoring
  type: conditional
  condition: "{flags.enable_monitoring == true}"
  then:
    - id: setup-monitoring
      type: plugin
      plugin: monitoring-setup
```

### 9. Document Complex Workflows

Add comments in your OpenAPI spec:

```yaml
# ===== PRE-FLIGHT VALIDATION PHASE =====
# These steps run in parallel to validate prerequisites before
# creating any resources. Failures here prevent resource creation.

- id: check-aws-credentials
  # ... configuration
```

### 10. Test Workflows Thoroughly

Test workflows with various scenarios:

- Success path (all steps succeed)
- Failure scenarios (each step fails)
- Retry behavior (transient failures)
- Rollback behavior (cleanup verification)
- Timeout handling
- Parallel execution
- Dry-run mode

---

## Summary

Workflows in CliForge v0.9.0 enable you to:

- Orchestrate complex multi-step operations
- Handle dependencies and execution order automatically
- Implement robust error handling with retry and rollback
- Execute steps in parallel for better performance
- Persist state for resumability
- Debug with dry-run and verbose modes

For more information:
- **Plugin Development**: See `developer-guide-plugins.md`
- **Architecture Details**: See `design/architecture/workflow-orchestration.md`
- **API Reference**: See `docs/technical-specification.md`

---

*Forged with CliForge v0.9.0*
