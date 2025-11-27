// Package workflow provides multi-step API workflow orchestration for CliForge.
//
// The workflow package enables complex multi-step operations that go beyond
// simple HTTP requests, including sequential and parallel execution, conditional
// branching, loops, polling, and automatic rollback on failures.
//
// # Step Types
//
//   - api-call: Execute HTTP API requests
//   - plugin: Invoke external tools or plugins
//   - conditional: Branch execution based on conditions
//   - loop: Iterate over collections
//   - wait: Delay or poll for status changes
//   - parallel: Execute multiple steps concurrently
//
// # Workflow Features
//
//   - DAG-based dependency resolution
//   - Automatic retry with exponential backoff
//   - Rollback actions for failed steps
//   - Output mapping between steps
//   - Expression evaluation for conditions and data transformation
//
// # Example Workflow Definition
//
//	workflow:
//	  steps:
//	    - id: create-resource
//	      type: api-call
//	      api-call:
//	        method: POST
//	        endpoint: /api/resources
//	    - id: wait-ready
//	      type: wait
//	      depends-on: [create-resource]
//	      wait:
//	        polling:
//	          endpoint: /api/resources/{id}/status
//	          interval: 5
//	          terminal-states: [ready, failed]
//
// # OpenAPI Integration
//
// Workflows are defined via x-cli-workflow extensions in OpenAPI specs
// and executed automatically when a command is invoked.
package workflow

import (
	"time"
)

// Workflow represents a complete workflow definition.
type Workflow struct {
	Steps    []*Step   `json:"steps"`
	Settings *Settings `json:"settings,omitempty"`
}

// Settings contains workflow-level configuration.
type Settings struct {
	ParallelExecution bool `json:"parallel-execution,omitempty"`
	FailFast          bool `json:"fail-fast,omitempty"`
	Timeout           int  `json:"timeout,omitempty"` // seconds
	DryRunSupported   bool `json:"dry-run-supported,omitempty"`
}

// Step represents a single workflow step.
type Step struct {
	// Common fields
	ID          string   `json:"id"`
	Type        StepType `json:"type"`
	Description string   `json:"description,omitempty"`
	DependsOn   []string `json:"depends-on,omitempty"`
	Condition   string   `json:"condition,omitempty"`
	Required    bool     `json:"required,omitempty"`

	// Retry configuration
	Retry *RetryConfig `json:"retry,omitempty"`

	// Rollback action
	Rollback *Step `json:"rollback,omitempty"`

	// Output mapping
	Output map[string]string `json:"output,omitempty"`

	// Type-specific fields
	APICall     *APICallStep     `json:"api-call,omitempty"`
	Plugin      *PluginStep      `json:"plugin,omitempty"`
	Conditional *ConditionalStep `json:"conditional,omitempty"`
	Loop        *LoopStep        `json:"loop,omitempty"`
	Wait        *WaitStep        `json:"wait,omitempty"`
	Parallel    *ParallelStep    `json:"parallel,omitempty"`
}

// StepType defines the type of workflow step.
type StepType string

const (
	// StepTypeAPICall represents an API call step.
	StepTypeAPICall StepType = "api-call"
	// StepTypePlugin represents a plugin execution step.
	StepTypePlugin StepType = "plugin"
	// StepTypeConditional represents a conditional execution step.
	StepTypeConditional StepType = "conditional"
	// StepTypeLoop represents a loop execution step.
	StepTypeLoop StepType = "loop"
	// StepTypeWait represents a wait/delay step.
	StepTypeWait StepType = "wait"
	// StepTypeParallel represents parallel execution of sub-steps.
	StepTypeParallel StepType = "parallel"
	// StepTypeNoop represents a no-operation step.
	StepTypeNoop StepType = "noop"
)

// RetryConfig defines retry behavior for a step.
type RetryConfig struct {
	MaxAttempts     int            `json:"max-attempts,omitempty"`
	Backoff         *BackoffConfig `json:"backoff,omitempty"`
	RetryableErrors []*ErrorMatch  `json:"retryable-errors,omitempty"`
}

// BackoffConfig defines exponential backoff settings.
type BackoffConfig struct {
	Type            BackoffType `json:"type,omitempty"`
	InitialInterval int         `json:"initial-interval,omitempty"` // seconds
	Multiplier      float64     `json:"multiplier,omitempty"`
	MaxInterval     int         `json:"max-interval,omitempty"` // seconds
}

// BackoffType defines the type of backoff strategy.
type BackoffType string

const (
	// BackoffFixed represents a fixed delay between retries.
	BackoffFixed BackoffType = "fixed"
	// BackoffLinear represents linearly increasing delay between retries.
	BackoffLinear BackoffType = "linear"
	// BackoffExponential represents exponentially increasing delay between retries.
	BackoffExponential BackoffType = "exponential"
)

// ErrorMatch defines criteria for matching errors.
type ErrorMatch struct {
	HTTPStatus *int   `json:"http-status,omitempty"`
	ErrorType  string `json:"error-type,omitempty"`
}

// APICallStep defines an HTTP API call step.
type APICallStep struct {
	Endpoint string                 `json:"endpoint"`
	Method   string                 `json:"method"`
	Headers  map[string]string      `json:"headers,omitempty"`
	Body     map[string]interface{} `json:"body,omitempty"`
	Query    map[string]string      `json:"query,omitempty"`
}

// PluginStep defines a plugin execution step.
type PluginStep struct {
	Plugin  string                 `json:"plugin"`
	Command string                 `json:"command"`
	Input   map[string]interface{} `json:"input,omitempty"`
}

// ConditionalStep defines a conditional branching step.
type ConditionalStep struct {
	Condition string  `json:"condition"`
	Then      []*Step `json:"then,omitempty"`
	Else      []*Step `json:"else,omitempty"`
}

// LoopStep defines an iteration step.
type LoopStep struct {
	Iterator   string  `json:"iterator"`
	Collection string  `json:"collection"`
	Steps      []*Step `json:"steps"`
}

// WaitStep defines a waiting/polling step.
type WaitStep struct {
	Condition string         `json:"condition,omitempty"`
	Duration  int            `json:"duration,omitempty"` // seconds
	Polling   *PollingConfig `json:"polling,omitempty"`
}

// PollingConfig defines polling behavior.
type PollingConfig struct {
	Endpoint       string   `json:"endpoint"`
	Interval       int      `json:"interval"` // seconds
	Timeout        int      `json:"timeout"`  // seconds
	TerminalStates []string `json:"terminal-states,omitempty"`
	StatusField    string   `json:"status-field,omitempty"`
}

// ParallelStep defines concurrent execution of multiple steps.
type ParallelStep struct {
	Steps []*Step `json:"steps"`
}

// StepResult represents the result of executing a step.
type StepResult struct {
	StepID    string
	Success   bool
	Output    map[string]interface{}
	Error     error
	Retries   int
	StartTime time.Time
	EndTime   time.Time
	Duration  time.Duration
}

// ExecutionState represents the state of workflow execution.
type ExecutionState struct {
	WorkflowID     string
	StartTime      time.Time
	Status         ExecutionStatus
	CompletedSteps []*StepResult
	CurrentStep    string
	Error          error
}

// ExecutionStatus defines the status of workflow execution.
type ExecutionStatus string

const (
	// ExecutionStatusPending indicates the workflow is waiting to start.
	ExecutionStatusPending ExecutionStatus = "pending"
	// ExecutionStatusRunning indicates the workflow is currently executing.
	ExecutionStatusRunning ExecutionStatus = "running"
	// ExecutionStatusCompleted indicates the workflow completed successfully.
	ExecutionStatusCompleted ExecutionStatus = "completed"
	// ExecutionStatusFailed indicates the workflow failed.
	ExecutionStatusFailed ExecutionStatus = "failed"
	// ExecutionStatusRolledBack indicates the workflow was rolled back.
	ExecutionStatusRolledBack ExecutionStatus = "rolled-back"
)

// DAG represents a Directed Acyclic Graph of workflow steps.
type DAG struct {
	Nodes map[string]*DAGNode
	Edges map[string][]string // stepID -> []dependentStepIDs
}

// DAGNode represents a node in the dependency graph.
type DAGNode struct {
	Step         *Step
	Dependencies []string
	Dependents   []string
	Level        int // Depth in the graph for topological ordering
}

// RollbackAction represents a rollback action for a step.
type RollbackAction struct {
	StepID string
	Action *Step
	Result *StepResult
}
