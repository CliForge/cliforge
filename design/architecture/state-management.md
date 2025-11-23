# State Management & Context Design

**Version**: 1.0.0
**Date**: 2025-11-23
**Status**: Proposed

---

## Overview

### Problem Statement

ROSA CLI maintains state across commands:
- Current cluster context (like kubectl)
- Recently used values (for smart defaults)
- Command history
- User preferences per cluster

CliForge currently has no state management beyond API response caching.

### Solution

A state management system that:
1. Persists context between commands
2. Provides smart defaults based on usage
3. Stores user preferences
4. Maintains command history
5. Follows XDG directory specification

---

## Architecture

### State Storage Location

**XDG-compliant paths**:
```
~/.config/mycli/           # XDG_CONFIG_HOME
├── config.yaml            # User configuration
├── state.yaml             # Current state/context
├── history.json           # Command history
└── cache/                 # API response cache
    └── openapi-spec.json
```

### State File Structure

**`state.yaml`**:
```yaml
# Current context
current:
  cluster: cluster-abc123
  region: us-east-1
  profile: production

# Recent items (for autocomplete)
recent:
  clusters:
    - cluster-abc123
    - cluster-xyz789
  regions:
    - us-east-1
    - us-west-2

# Per-resource preferences
preferences:
  clusters:
    cluster-abc123:
      last-used: 2025-11-23T10:30:00Z
      default-machine-pool: default
      favorite: true

# Session data
session:
  last-command: describe cluster
  last-command-time: 2025-11-23T10:30:00Z
```

---

## Context System

### Setting Context

```bash
# Set current cluster
$ mycli config set-context --cluster cluster-abc123

# Set multiple context values
$ mycli config set-context \
    --cluster cluster-abc123 \
    --region us-east-1 \
    --profile production
```

### Using Context

**Implicit usage** (no cluster flag needed):
```bash
# Uses current cluster from context
$ mycli describe cluster

# Equivalent to:
$ mycli describe cluster --cluster cluster-abc123
```

**Explicit override**:
```bash
# Override context
$ mycli describe cluster --cluster cluster-xyz789
```

### Multiple Contexts

**Named contexts** (like kubeconfig):
```yaml
# ~/.config/mycli/state.yaml
contexts:
  production:
    cluster: prod-cluster
    region: us-east-1
    profile: prod-profile

  staging:
    cluster: staging-cluster
    region: us-west-2
    profile: staging-profile

current-context: production
```

**Switch context**:
```bash
$ mycli config use-context staging
Switched to context "staging"

$ mycli describe cluster
# Now uses staging-cluster
```

---

## Smart Defaults

### Recent Values

**Autocomplete uses recent values**:
```bash
$ mycli create machinepool --cluster <TAB>
cluster-abc123  cluster-xyz789  # From recent list
```

**Flag defaults**:
```yaml
# User ran: mycli create cluster --region us-east-1

# Next time, default offered:
$ mycli create cluster --name test2
? Select region: (us-east-1) ▸ us-east-1  # Default from last time
                                us-west-2
                                eu-west-1
```

### Usage-Based Ordering

```bash
# Sort by last-used
$ mycli list clusters --sort recent
ID              NAME            LAST USED
cluster-abc123  production      2 minutes ago
cluster-xyz789  staging         1 hour ago
cluster-def456  development     2 days ago
```

---

## Command History

### Storage

**`history.json`**:
```json
{
  "history": [
    {
      "command": "mycli create cluster --name test --region us-east-1",
      "timestamp": "2025-11-23T10:30:00Z",
      "exit_code": 0,
      "duration_ms": 45230,
      "user": "alice"
    },
    {
      "command": "mycli describe cluster --cluster test",
      "timestamp": "2025-11-23T10:31:00Z",
      "exit_code": 0,
      "duration_ms": 1250,
      "user": "alice"
    }
  ],
  "max_entries": 1000
}
```

### History Command

```bash
# View recent commands
$ mycli history
1  2025-11-23 10:30  mycli create cluster --name test
2  2025-11-23 10:31  mycli describe cluster --cluster test

# Re-run command
$ mycli history run 1

# Search history
$ mycli history search "create cluster"
```

---

## Workflow State Persistence

### Resume Failed Workflows

**Use Case**: Cluster creation fails at step 5 of 10

**State saved**:
```yaml
workflows:
  create-cluster-abc123:
    workflow-id: wf-12345
    operation: createCluster
    started: 2025-11-23T10:30:00Z
    status: failed
    completed-steps:
      - check-aws-credentials: success
      - check-quotas: success
      - create-iam-roles: success
      - create-oidc-provider: success
      - create-cluster-api: failed
    step-outputs:
      create-iam-roles:
        installer_role_arn: arn:aws:iam::123:role/test-installer
      create-oidc-provider:
        oidc_provider_arn: arn:aws:iam::123:oidc-provider/...
    error:
      step: create-cluster-api
      message: "Network timeout"
```

**Resume command**:
```bash
$ mycli workflow resume wf-12345
Resuming workflow from step 'create-cluster-api'...

Using saved state:
  ✓ IAM roles already created
  ✓ OIDC provider already created

Retrying cluster creation...
```

---

## OpenAPI Extension

### `x-cli-state`

**Location**: Operation level

**Purpose**: Declare state/context requirements

```yaml
get:
  operationId: describeCluster
  x-cli-state:
    # Allow using cluster from context
    context-params:
      - name: cluster_id
        context-key: current.cluster
        required: false  # Can be provided via flag

    # Save to recent list
    save-recent:
      - field: cluster_id
        list: clusters
        max-entries: 10
```

### `x-cli-context`

**Location**: Root level

**Purpose**: Configure context system

```yaml
x-cli-context:
  enabled: true
  state-file: "~/.mycli/state.yaml"
  contexts-supported: true
  default-context: default

  # Context fields
  fields:
    - name: cluster
      description: "Current cluster"
      type: string

    - name: region
      description: "Default AWS region"
      type: string
      env-var: AWS_REGION

    - name: profile
      description: "Configuration profile"
      type: string
```

---

## Implementation Phases

### Phase 1: Basic State (v0.8.0)

**Scope**:
- State file storage
- Current context (single value)
- Recent values tracking

### Phase 2: Advanced Context (v0.9.0)

**Scope**:
- Named contexts
- Context switching
- Smart defaults
- Command history

### Phase 3: Workflow Persistence (v1.0.0)

**Scope**:
- Workflow state persistence
- Resume capability
- Audit trail

---

## Related Documents

- **Workflow Orchestration**: `workflow-orchestration.md`
- **Technical Specification**: `../../docs/technical-specification.md`
- **Gap Analysis**: `gap-analysis-rosa-requirements.md`

---

*⚒️ Forged with ❤️ by the CliForge team*
