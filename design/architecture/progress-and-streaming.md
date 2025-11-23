# Progress Indicators & Streaming Design

**Version**: 1.0.0
**Date**: 2025-11-23
**Status**: Proposed

---

## Overview

### Problem Statement

ROSA CLI provides excellent real-time feedback during long operations:
- Spinners during API calls ("Creating cluster...")
- Progress bars for multi-step workflows
- Streaming installation logs with `--watch`
- Real-time status updates

CliForge's async polling provides basic status checking but lacks visual feedback and streaming capabilities.

### Solution

A progress/streaming framework that:
1. Shows visual progress indicators (spinners, progress bars)
2. Streams real-time logs via SSE/WebSocket
3. Provides step-by-step workflow progress
4. Gracefully degrades if streaming unavailable

---

## Progress Indicators

### UX Framework

**Library**: pterm (https://github.com/pterm/pterm)
- Rich terminal UI components
- Spinners, progress bars, tables
- Cross-platform support
- No external dependencies

### Indicator Types

#### 1. Spinner

**Use Case**: Single operation in progress

```go
spinner := pterm.DefaultSpinner.Start("Creating cluster...")
// ... operation ...
spinner.Success("Cluster created successfully")
```

**OpenAPI Extension**:
```yaml
x-cli-output:
  progress:
    type: spinner
    message: "Creating cluster..."
    success-message: "Cluster '{name}' created"
```

#### 2. Progress Bar

**Use Case**: Known number of steps

```go
bar := pterm.DefaultProgressbar.
    WithTotal(5).
    WithTitle("Creating cluster")

bar.Increment() // Step 1
bar.Increment() // Step 2
// ...
```

**OpenAPI Extension**:
```yaml
x-cli-workflow:
  settings:
    show-progress: true
    progress-type: bar  # bar or steps

  steps:
    - id: step-1
      description: "Validating credentials"
    - id: step-2
      description: "Creating IAM roles"
    # ... 3 more steps
```

**Display**:
```
Creating cluster [████████░░░░░░░] 40% (2/5)
Current step: Creating IAM roles...
```

#### 3. Multi-Step Display

**Use Case**: Complex workflow with sub-steps

```
✓ Validate AWS credentials
✓ Check quotas
⧗ Create IAM roles
  ✓ Installer role
  ⧗ Worker role
  ☐ Operator roles
☐ Create OIDC provider
☐ Create cluster
```

**OpenAPI Extension**:
```yaml
x-cli-workflow:
  settings:
    show-progress: true
    progress-type: steps
```

---

## Streaming Support

### Protocols

#### 1. Server-Sent Events (SSE)

**Use Case**: One-way real-time updates from server

**Advantages**:
- Simple HTTP-based
- Automatic reconnection
- Text-based, easy to debug
- No special server infrastructure

**Example**:
```yaml
x-cli-watch:
  enabled: true
  type: sse
  endpoint: /api/v1/clusters/{cluster_id}/logs/stream
  events:
    - log  # Event type
    - status
    - error
```

**Client Implementation**:
```go
// Stream logs
eventSource := sse.Connect(streamURL)
for event := range eventSource.Events {
    switch event.Type {
    case "log":
        fmt.Println(event.Data)
    case "status":
        updateStatus(event.Data)
    case "error":
        handleError(event.Data)
    }
}
```

#### 2. WebSocket

**Use Case**: Bidirectional communication needed

**Advantages**:
- Full duplex
- Binary support
- Lower latency

**Use When**:
- Need to send commands during streaming
- Interactive sessions

**Example**:
```yaml
x-cli-watch:
  enabled: true
  type: websocket
  endpoint: ws://api.example.com/v1/clusters/{cluster_id}/console
```

#### 3. Polling Fallback

**Use Case**: Streaming unavailable

**Implementation**:
```yaml
x-cli-async:
  enabled: true
  streaming:
    preferred: sse
    fallback: polling
  polling:
    interval: 30
    timeout: 3600
```

---

## Watch Mode

### Flag: `--watch`

**Behavior**:
- Initiates operation (if applicable)
- Streams real-time updates
- Exits when operation completes or fails
- Handles interrupts (Ctrl+C) gracefully

### Example: Cluster Creation with Watch

```bash
$ mycli create cluster --name test --watch

Creating cluster 'test'...

⧗ Validating AWS credentials... ✓ Done
⧗ Checking AWS quotas... ✓ Done
⧗ Creating IAM roles...
  ✓ Installer role (arn:aws:iam::123:role/test-installer)
  ⧗ Worker role...

--- Streaming installation logs ---
2025-11-23 10:30:15 INFO Starting cluster installation
2025-11-23 10:30:20 INFO Creating control plane nodes
2025-11-23 10:31:45 INFO Control plane healthy
2025-11-23 10:32:10 INFO Creating worker nodes
...
2025-11-23 10:45:00 INFO Cluster installation complete

✓ Cluster 'test' is ready
  API URL: https://api.test.example.com
  Console: https://console.test.example.com
```

---

## OpenAPI Extensions

### `x-cli-watch`

**Location**: Operation level

**Purpose**: Enable watch/streaming mode

```yaml
delete:
  operationId: deleteCluster
  x-cli-watch:
    enabled: true
    type: sse  # sse, websocket, polling
    endpoint: /api/v1/clusters/{cluster_id}/logs/uninstall
    events:
      - log
      - status
    exit-on:
      - event: status
        condition: "{event.data.state in ['deleted', 'error']}"
```

### `x-cli-progress`

**Location**: Operation or workflow level

**Purpose**: Configure progress display

```yaml
x-cli-progress:
  enabled: true
  type: spinner  # spinner, bar, steps
  show-step-descriptions: true
  show-timestamps: true
  color: auto  # auto, always, never
```

---

## Implementation Phases

### Phase 1: Basic Progress (v0.8.0)

**Scope**:
- Spinner for single operations
- Simple step display for workflows

### Phase 2: Streaming (v0.9.0)

**Scope**:
- SSE client
- Polling fallback
- `--watch` flag

### Phase 3: Advanced Features (v1.0.0)

**Scope**:
- WebSocket support
- Interactive streaming
- Advanced progress bars

---

## Related Documents

- **Workflow Orchestration**: `workflow-orchestration.md`
- **Technical Specification**: `../../docs/technical-specification.md`
- **Gap Analysis**: `gap-analysis-rosa-requirements.md`

---

*⚒️ Forged with ❤️ by the CliForge team*
