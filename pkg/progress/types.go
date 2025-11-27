package progress

import (
	"context"
	"io"
	"time"
)

// Type defines the type of progress indicator.
type Type string

const (
	// TypeSpinner shows a spinner for single operations.
	TypeSpinner Type = "spinner"
	// TypeBar shows a progress bar for known step counts.
	TypeBar Type = "bar"
	// TypeSteps shows multi-step tree display for workflows.
	TypeSteps Type = "steps"
	// TypeNone disables progress indicators.
	TypeNone Type = "none"
)

// Progress is the interface for all progress indicators.
type Progress interface {
	// Start starts the progress indicator with a message.
	Start(message string) error

	// Update updates the progress with a message or percentage.
	Update(message string) error

	// UpdateWithData updates progress with structured data.
	UpdateWithData(data *Data) error

	// Success marks the progress as successful.
	Success(message string) error

	// Failure marks the progress as failed.
	Failure(message string) error

	// Stop stops the progress indicator.
	Stop() error

	// IsActive returns true if the progress indicator is active.
	IsActive() bool
}

// Data contains structured progress information.
type Data struct {
	// Message is the current progress message.
	Message string

	// Current is the current step/item number.
	Current int

	// Total is the total number of steps/items.
	Total int

	// Percentage is the completion percentage (0-100).
	Percentage float64

	// Timestamp is when this update occurred.
	Timestamp time.Time

	// Metadata contains additional context.
	Metadata map[string]interface{}
}

// StepStatus defines the status of a workflow step.
type StepStatus string

const (
	// StepStatusPending indicates the step has not started.
	StepStatusPending StepStatus = "pending"

	// StepStatusRunning indicates the step is currently executing.
	StepStatusRunning StepStatus = "running"

	// StepStatusCompleted indicates the step completed successfully.
	StepStatusCompleted StepStatus = "completed"

	// StepStatusFailed indicates the step failed.
	StepStatusFailed StepStatus = "failed"

	// StepStatusSkipped indicates the step was skipped.
	StepStatusSkipped StepStatus = "skipped"
)

// StepInfo contains information about a workflow step.
type StepInfo struct {
	// ID is the unique identifier for the step.
	ID string

	// Description is the human-readable description.
	Description string

	// Status is the current status of the step.
	Status StepStatus

	// StartTime is when the step started.
	StartTime time.Time

	// EndTime is when the step completed.
	EndTime time.Time

	// Error is any error that occurred during the step.
	Error error

	// SubSteps contains nested steps.
	SubSteps []*StepInfo

	// Metadata contains additional step information.
	Metadata map[string]interface{}
}

// Config contains configuration for progress indicators.
type Config struct {
	// Type is the type of progress indicator to use.
	Type Type

	// Enabled determines if progress indicators are shown.
	Enabled bool

	// ShowTimestamps shows timestamps with step updates.
	ShowTimestamps bool

	// ShowStepDescriptions shows detailed step descriptions.
	ShowStepDescriptions bool

	// Color determines color usage (auto, always, never).
	Color string

	// Writer is where to write progress output.
	Writer io.Writer

	// RefreshRate is how often to refresh the display (for spinners).
	RefreshRate time.Duration
}

// DefaultConfig returns a default configuration.
func DefaultConfig() *Config {
	return &Config{
		Type:                 TypeSpinner,
		Enabled:              true,
		ShowTimestamps:       false,
		ShowStepDescriptions: true,
		Color:                "auto",
		RefreshRate:          100 * time.Millisecond,
	}
}

// StreamType defines the type of streaming connection.
type StreamType string

const (
	// StreamTypeSSE uses Server-Sent Events.
	StreamTypeSSE StreamType = "sse"

	// StreamTypeWebSocket uses WebSocket.
	StreamTypeWebSocket StreamType = "websocket"

	// StreamTypePolling uses HTTP polling.
	StreamTypePolling StreamType = "polling"
)

// StreamConfig contains configuration for streaming.
type StreamConfig struct {
	// Type is the type of streaming to use.
	Type StreamType

	// Endpoint is the streaming endpoint URL.
	Endpoint string

	// Events is the list of event types to listen for.
	Events []string

	// ReconnectInterval is how long to wait before reconnecting.
	ReconnectInterval time.Duration

	// MaxReconnectAttempts is the maximum number of reconnect attempts.
	MaxReconnectAttempts int

	// Timeout is the connection timeout.
	Timeout time.Duration

	// PollingInterval is the interval for polling mode.
	PollingInterval time.Duration

	// Headers are additional HTTP headers to send.
	Headers map[string]string
}

// DefaultStreamConfig returns a default stream configuration.
func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		Type:                 StreamTypeSSE,
		ReconnectInterval:    5 * time.Second,
		MaxReconnectAttempts: 3,
		Timeout:              30 * time.Second,
		PollingInterval:      5 * time.Second,
		Headers:              make(map[string]string),
	}
}

// Event represents a streaming event.
type Event struct {
	// Type is the event type.
	Type string

	// Data is the event payload.
	Data string

	// ID is the event ID (for SSE).
	ID string

	// Timestamp is when the event was received.
	Timestamp time.Time

	// Raw is the raw event data.
	Raw interface{}
}

// EventHandler handles streaming events.
type EventHandler func(event *Event) error

// StreamClient is the interface for streaming clients.
type StreamClient interface {
	// Connect establishes a connection to the streaming endpoint.
	Connect(ctx context.Context) error

	// Subscribe subscribes to events with a handler.
	Subscribe(eventType string, handler EventHandler) error

	// Unsubscribe removes a subscription.
	Unsubscribe(eventType string) error

	// Close closes the connection.
	Close() error

	// IsConnected returns true if the client is connected.
	IsConnected() bool

	// Events returns a channel of events.
	Events() <-chan *Event

	// Errors returns a channel of errors.
	Errors() <-chan error
}

// WatchConfig contains configuration for watch mode.
type WatchConfig struct {
	// Enabled determines if watch mode is enabled.
	Enabled bool

	// StreamConfig is the streaming configuration.
	StreamConfig *StreamConfig

	// ProgressConfig is the progress indicator configuration.
	ProgressConfig *Config

	// ExitConditions defines when to exit watch mode.
	ExitConditions []*ExitCondition

	// ShowLogs determines if logs should be displayed.
	ShowLogs bool

	// LogPrefix is the prefix to use for log lines.
	LogPrefix string
}

// ExitCondition defines a condition for exiting watch mode.
type ExitCondition struct {
	// EventType is the event type to check.
	EventType string

	// Condition is an expression to evaluate.
	Condition string

	// Message is the message to display when exiting.
	Message string
}

// WatchCoordinator coordinates watch mode operation.
type WatchCoordinator interface {
	// Start starts watch mode.
	Start(ctx context.Context) error

	// Stop stops watch mode.
	Stop() error

	// IsRunning returns true if watch mode is active.
	IsRunning() bool
}
