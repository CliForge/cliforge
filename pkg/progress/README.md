# Progress and Streaming Package

This package provides comprehensive progress indicators and streaming support for CliForge v0.9.0, enabling rich real-time feedback during long-running operations.

## Features

### Progress Indicators

- **Spinner**: For single operations with unknown duration
- **Progress Bar**: For operations with known step counts
- **Multi-Step Display**: For complex workflows with hierarchical steps
- **Configurable**: Control appearance, colors, timestamps, and more

### Streaming Support

- **SSE (Server-Sent Events)**: One-way real-time updates from server
- **WebSocket**: Bidirectional communication
- **Polling**: HTTP polling fallback
- **Automatic Reconnection**: Built-in retry logic
- **Event Handlers**: Subscribe to specific event types

### Watch Mode

- **Real-time Monitoring**: Stream logs and status updates
- **Exit Conditions**: Define when to stop watching
- **Workflow Integration**: Track multi-step operations
- **Graceful Shutdown**: Handle Ctrl+C cleanly

## Usage

### Basic Spinner

```go
import "github.com/CliForge/cliforge/pkg/progress"

// Create and start a spinner
spinner := progress.NewSpinner(&progress.Config{
    Type:    progress.ProgressTypeSpinner,
    Enabled: true,
})

spinner.Start("Processing...")
// ... do work ...
spinner.Success("Done!")
```

### Progress Bar

```go
bar := progress.NewProgressBar(&progress.Config{
    Type:    progress.ProgressTypeBar,
    Enabled: true,
}, 10) // 10 total steps

bar.Start("Processing items...")

for i := 0; i < 10; i++ {
    // ... process item ...
    bar.Increment()
}

bar.Success("All items processed!")
```

### Multi-Step Workflow

```go
multiStep := progress.NewMultiStep(&progress.Config{
    Type:                 progress.ProgressTypeSteps,
    Enabled:              true,
    ShowTimestamps:       true,
    ShowStepDescriptions: true,
})

multiStep.Start("Deploying application...")

// Add steps
steps := []*progress.StepInfo{
    {ID: "validate", Description: "Validating configuration", Status: progress.StepStatusPending},
    {ID: "build", Description: "Building application", Status: progress.StepStatusPending},
    {ID: "deploy", Description: "Deploying to production", Status: progress.StepStatusPending},
}

for _, step := range steps {
    multiStep.AddStep(step)
}

// Update steps as they progress
multiStep.UpdateStep("validate", progress.StepStatusRunning, "Validating...")
// ...
multiStep.UpdateStep("validate", progress.StepStatusCompleted, "Validation complete")

multiStep.Success("Deployment successful!")
```

### Streaming with SSE

```go
config := &progress.StreamConfig{
    Type:                 progress.StreamTypeSSE,
    Endpoint:             "https://api.example.com/stream",
    Events:               []string{"log", "status", "error"},
    ReconnectInterval:    5 * time.Second,
    MaxReconnectAttempts: 3,
}

client := progress.NewSSEClient(config)

// Subscribe to events
client.Subscribe("log", func(event *progress.Event) error {
    fmt.Println("LOG:", event.Data)
    return nil
})

client.Subscribe("status", func(event *progress.Event) error {
    fmt.Println("STATUS:", event.Data)
    return nil
})

ctx := context.Background()
client.Connect(ctx)

// Wait for events
for event := range client.Events() {
    // Process event
}

client.Close()
```

### Watch Mode

```go
watchConfig := &progress.WatchConfig{
    Enabled: true,
    StreamConfig: &progress.StreamConfig{
        Type:     progress.StreamTypeSSE,
        Endpoint: "https://api.example.com/cluster/logs",
        Events:   []string{"log", "status"},
    },
    ProgressConfig: &progress.Config{
        Type:    progress.ProgressTypeSpinner,
        Enabled: true,
    },
    ExitConditions: []*progress.ExitCondition{
        {
            EventType: "status",
            Condition: "event.data.state == 'completed'",
            Message:   "Cluster creation completed",
        },
    },
    ShowLogs:  true,
    LogPrefix: "2006-01-02 15:04:05",
}

watch, err := progress.NewWatch(watchConfig)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
watch.Start(ctx) // Blocks until completion or error
```

### Progress Manager

The progress manager provides a high-level API for managing progress indicators:

```go
manager := progress.NewManager(&progress.Config{
    Type:    progress.ProgressTypeSpinner,
    Enabled: true,
})

// Start progress
prog, _ := manager.StartProgress("Initializing...", 0)

// Update
manager.UpdateProgress("Processing...")

// Complete
manager.SuccessProgress("Done!")

// Cleanup
manager.StopProgress()
```

### Workflow Integration

Integrate progress tracking with workflow execution:

```go
manager := progress.NewManager(&progress.Config{
    Type:    progress.ProgressTypeSteps,
    Enabled: true,
})

integration := progress.NewWorkflowIntegration(manager)

// Workflow lifecycle
integration.OnWorkflowStart(workflow)
integration.OnStepStart("step1")
integration.OnStepComplete("step1")
integration.OnWorkflowComplete(true, "Success!")
```

### OpenAPI Integration

Use progress indicators configured via OpenAPI extensions:

```go
// From OpenAPI spec:
// x-cli-progress:
//   enabled: true
//   type: spinner
//   show-timestamps: true

opConfig := &openapi.CLIProgress{
    Enabled: &enabled,
    Type:    "spinner",
    ShowTimestamps: &enabled,
}

manager := progress.NewManager(defaultConfig)
prog, _ := manager.StartProgressForOperation(opConfig, "Executing operation...")
```

## Configuration

### Progress Config

```go
type Config struct {
    Type                 ProgressType  // spinner, bar, steps, none
    Enabled              bool          // Enable/disable progress
    ShowTimestamps       bool          // Show timestamps
    ShowStepDescriptions bool          // Show detailed descriptions
    Color                string        // auto, always, never
    Writer               io.Writer     // Output writer
    RefreshRate          time.Duration // Refresh interval
}
```

### Stream Config

```go
type StreamConfig struct {
    Type                 StreamType    // sse, websocket, polling
    Endpoint             string        // Streaming endpoint
    Events               []string      // Event types to listen for
    ReconnectInterval    time.Duration // Time between reconnects
    MaxReconnectAttempts int           // Max reconnection attempts
    Timeout              time.Duration // Connection timeout
    PollingInterval      time.Duration // Polling interval
    Headers              map[string]string // Additional headers
}
```

## Testing

Run tests:

```bash
go test ./pkg/progress/...
```

Run with coverage:

```bash
go test ./pkg/progress/... -cover
```

Run examples:

```bash
go test ./pkg/progress/... -run Example
```

## Architecture

```
pkg/progress/
├── types.go           # Core types and interfaces
├── progress.go        # Progress indicator implementations
├── streaming.go       # Streaming client implementations
├── watch.go          # Watch mode coordinator
├── manager.go        # Progress manager
├── progress_test.go  # Progress tests
├── streaming_test.go # Streaming tests
├── manager_test.go   # Manager tests
├── example_test.go   # Usage examples
└── README.md         # This file
```

## OpenAPI Extensions

### x-cli-progress

Configure progress display for operations:

```yaml
x-cli-progress:
  enabled: true
  type: spinner  # spinner, bar, steps
  show-step-descriptions: true
  show-timestamps: true
  color: auto  # auto, always, never
```

### x-cli-watch

Enable watch mode for operations:

```yaml
x-cli-watch:
  enabled: true
  type: sse  # sse, websocket, polling
  endpoint: /api/v1/operations/{id}/stream
  events:
    - log
    - status
    - error
  exit-on:
    - event: status
      condition: "event.data.state == 'completed'"
      message: "Operation completed successfully"
  reconnect:
    enabled: true
    max-attempts: 3
    interval: 5
```

## Package-Level Functions

Convenience functions using the default manager:

```go
// Start progress
progress.StartProgress("Working...", 0)

// Update progress
progress.UpdateProgress("Still working...")

// Mark successful
progress.SuccessProgress("Done!")

// Mark failed
progress.FailureProgress("Error!")

// Stop progress
progress.StopProgress()
```

## Event Types

### Standard Event Types

- `log`: Log messages
- `status`: Status updates
- `error`: Error messages
- `message`: General messages (WebSocket)
- `poll`: Polling responses

### Event Structure

```go
type Event struct {
    Type      string      // Event type
    Data      string      // Event payload
    ID        string      // Event ID (SSE)
    Timestamp time.Time   // When received
    Raw       interface{} // Raw event data
}
```

## Step Status

```go
const (
    StepStatusPending   // Not started
    StepStatusRunning   // Currently executing
    StepStatusCompleted // Finished successfully
    StepStatusFailed    // Failed
    StepStatusSkipped   // Skipped
)
```

## Best Practices

1. **Choose the Right Progress Type**:
   - Use `spinner` for single operations with unknown duration
   - Use `bar` for operations with known step counts
   - Use `steps` for complex workflows

2. **Handle Errors Gracefully**:
   - Always check for errors when starting progress
   - Use `Failure()` to indicate errors
   - Clean up with `Stop()` in defer blocks

3. **Streaming**:
   - Set reasonable timeouts
   - Implement reconnection logic
   - Handle context cancellation

4. **Watch Mode**:
   - Define clear exit conditions
   - Handle Ctrl+C gracefully
   - Show meaningful log messages

5. **Testing**:
   - Set `Enabled: false` in tests to avoid terminal output
   - Use mock servers for streaming tests
   - Test concurrent access for managers

## Performance

- Progress indicators use minimal CPU when inactive
- Streaming clients use buffered channels to prevent blocking
- Multi-step display uses efficient tree rendering
- All operations are thread-safe with proper locking

## Dependencies

- `github.com/pterm/pterm`: Terminal UI library
- `github.com/gorilla/websocket`: WebSocket support
- Standard library: net/http, context, sync

## License

Part of CliForge - see main project LICENSE

## Version

v0.9.0 - Complete progress indicators and streaming system
