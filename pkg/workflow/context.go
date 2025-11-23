package workflow

import (
	"sync"
)

// ExecutionContext holds the state and data for workflow execution.
type ExecutionContext struct {
	// Flags from CLI arguments
	Flags map[string]interface{}

	// Variables for temporary storage
	Variables map[string]interface{}

	// Step results keyed by step ID
	StepResults map[string]*StepResult

	// Completed steps in order
	CompletedSteps []*StepResult

	// Rollback actions to execute if workflow fails
	RollbackActions []*RollbackAction

	// HTTP client for API calls
	HTTPClient interface{} // Will be *http.Client

	// Plugin executor interface
	PluginExecutor interface{}

	// Mutex for thread-safe access
	mu sync.RWMutex
}

// NewExecutionContext creates a new execution context.
func NewExecutionContext(flags map[string]interface{}) *ExecutionContext {
	return &ExecutionContext{
		Flags:           flags,
		Variables:       make(map[string]interface{}),
		StepResults:     make(map[string]*StepResult),
		CompletedSteps:  make([]*StepResult, 0),
		RollbackActions: make([]*RollbackAction, 0),
	}
}

// SetStepResult stores the result of a step execution.
func (c *ExecutionContext) SetStepResult(stepID string, result *StepResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.StepResults[stepID] = result
	c.CompletedSteps = append(c.CompletedSteps, result)
}

// GetStepResult retrieves the result of a step.
func (c *ExecutionContext) GetStepResult(stepID string) (*StepResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, exists := c.StepResults[stepID]
	return result, exists
}

// SetVariable sets a context variable.
func (c *ExecutionContext) SetVariable(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Variables[key] = value
}

// GetVariable retrieves a context variable.
func (c *ExecutionContext) GetVariable(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	value, exists := c.Variables[key]
	return value, exists
}

// AddRollbackAction adds a rollback action to the stack.
func (c *ExecutionContext) AddRollbackAction(action *RollbackAction) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.RollbackActions = append(c.RollbackActions, action)
}

// GetRollbackActions returns all rollback actions in reverse order.
func (c *ExecutionContext) GetRollbackActions() []*RollbackAction {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return in reverse order (LIFO)
	reversed := make([]*RollbackAction, len(c.RollbackActions))
	for i, action := range c.RollbackActions {
		reversed[len(c.RollbackActions)-1-i] = action
	}

	return reversed
}

// Clone creates a shallow copy of the context for nested execution.
func (c *ExecutionContext) Clone() *ExecutionContext {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := &ExecutionContext{
		Flags:           c.Flags,
		Variables:       make(map[string]interface{}),
		StepResults:     c.StepResults, // Share step results
		CompletedSteps:  c.CompletedSteps,
		RollbackActions: c.RollbackActions,
		HTTPClient:      c.HTTPClient,
		PluginExecutor:  c.PluginExecutor,
	}

	// Copy variables
	for k, v := range c.Variables {
		clone.Variables[k] = v
	}

	return clone
}
