package workflow

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Executor executes workflows.
type Executor struct {
	workflow     *Workflow
	dag          *DAG
	stepExecutor *StepExecutor
	rollback     *RollbackManager
	state        *StateManager
}

// NewExecutor creates a new workflow executor.
func NewExecutor(workflow *Workflow, httpClient *http.Client, pluginExecutor interface{}) (*Executor, error) {
	// Parse workflow and build DAG
	parser := NewParser(workflow)
	dag, err := parser.Parse()
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %w", err)
	}

	executor := &Executor{
		workflow:     workflow,
		dag:          dag,
		stepExecutor: NewStepExecutor(httpClient, pluginExecutor),
		rollback:     NewRollbackManager(),
		state:        NewStateManager(),
	}

	return executor, nil
}

// Execute executes the workflow.
func (e *Executor) Execute(ctx *ExecutionContext) (*ExecutionState, error) {
	state := &ExecutionState{
		WorkflowID:     fmt.Sprintf("workflow-%d", time.Now().Unix()),
		StartTime:      time.Now(),
		Status:         ExecutionStatusRunning,
		CompletedSteps: make([]*StepResult, 0),
	}

	// Get execution order from the already-parsed DAG
	// Create a temporary parser to get execution order
	parser := &Parser{
		workflow: e.workflow,
		dag:      e.dag,
	}
	executionOrder := parser.GetExecutionOrder()

	// Execute steps level by level
	for _, levelSteps := range executionOrder {
		if len(levelSteps) == 0 {
			continue
		}

		// Check if parallel execution is enabled
		parallelEnabled := e.workflow.Settings != nil && e.workflow.Settings.ParallelExecution

		var levelErr error
		if parallelEnabled && len(levelSteps) > 1 {
			// Execute level in parallel
			levelErr = e.executeLevelParallel(levelSteps, ctx, state)
		} else {
			// Execute level sequentially
			levelErr = e.executeLevelSequential(levelSteps, ctx, state)
		}

		if levelErr != nil {
			// Handle failure
			state.Status = ExecutionStatusFailed
			state.Error = levelErr

			// Trigger rollback
			if err := e.rollback.ExecuteRollback(ctx, e.stepExecutor); err != nil {
				state.Error = fmt.Errorf("workflow failed and rollback failed: %w (rollback error: %v)", levelErr, err)
			} else {
				state.Status = ExecutionStatusRolledBack
			}

			return state, levelErr
		}

		// Check if should fail fast
		if e.workflow.Settings != nil && e.workflow.Settings.FailFast {
			for _, step := range levelSteps {
				result, exists := ctx.GetStepResult(step.ID)
				if exists && !result.Success {
					state.Status = ExecutionStatusFailed
					state.Error = fmt.Errorf("step %s failed (fail-fast enabled)", step.ID)

					// Trigger rollback
					if err := e.rollback.ExecuteRollback(ctx, e.stepExecutor); err != nil {
						state.Error = fmt.Errorf("workflow failed and rollback failed: %w (rollback error: %v)", state.Error, err)
					} else {
						state.Status = ExecutionStatusRolledBack
					}

					return state, state.Error
				}
			}
		}

		// Check timeout
		if e.workflow.Settings != nil && e.workflow.Settings.Timeout > 0 {
			elapsed := time.Since(state.StartTime).Seconds()
			if elapsed > float64(e.workflow.Settings.Timeout) {
				state.Status = ExecutionStatusFailed
				state.Error = fmt.Errorf("workflow timeout after %d seconds", e.workflow.Settings.Timeout)

				// Trigger rollback
				if err := e.rollback.ExecuteRollback(ctx, e.stepExecutor); err != nil {
					state.Error = fmt.Errorf("workflow timeout and rollback failed: %w (rollback error: %v)", state.Error, err)
				} else {
					state.Status = ExecutionStatusRolledBack
				}

				return state, state.Error
			}
		}

		// Save state after each level
		if err := e.state.SaveState(state); err != nil {
			// Log error but continue
			fmt.Printf("Warning: failed to save state: %v\n", err)
		}
	}

	// All steps completed successfully
	state.Status = ExecutionStatusCompleted
	state.CompletedSteps = ctx.CompletedSteps

	if err := e.state.SaveState(state); err != nil {
		fmt.Printf("Warning: failed to save final state: %v\n", err)
	}

	return state, nil
}

// executeLevelSequential executes steps in a level sequentially.
func (e *Executor) executeLevelSequential(levelSteps []*Step, ctx *ExecutionContext, state *ExecutionState) error {
	for _, step := range levelSteps {
		state.CurrentStep = step.ID

		result, err := e.stepExecutor.ExecuteStep(step, ctx)
		if err != nil {
			return fmt.Errorf("step %s failed: %w", step.ID, err)
		}

		// Store result
		ctx.SetStepResult(step.ID, result)
		state.CompletedSteps = append(state.CompletedSteps, result)

		// Add rollback action if step has rollback
		if step.Rollback != nil && result.Success {
			rollbackAction := &RollbackAction{
				StepID: step.ID,
				Action: step.Rollback,
				Result: result,
			}
			ctx.AddRollbackAction(rollbackAction)
		}

		// Check if step failed and is required
		if !result.Success {
			if step.Required {
				return fmt.Errorf("required step %s failed", step.ID)
			}
		}
	}

	return nil
}

// executeLevelParallel executes steps in a level in parallel.
func (e *Executor) executeLevelParallel(levelSteps []*Step, ctx *ExecutionContext, state *ExecutionState) error {
	var wg sync.WaitGroup
	resultsChan := make(chan *stepExecutionResult, len(levelSteps))
	errorsChan := make(chan error, len(levelSteps))

	for _, step := range levelSteps {
		wg.Add(1)
		go func(s *Step) {
			defer wg.Done()

			result, err := e.stepExecutor.ExecuteStep(s, ctx)
			if err != nil {
				errorsChan <- fmt.Errorf("step %s failed: %w", s.ID, err)
				return
			}

			resultsChan <- &stepExecutionResult{
				step:   s,
				result: result,
				err:    err,
			}
		}(step)
	}

	wg.Wait()
	close(resultsChan)
	close(errorsChan)

	// Check for errors
	for err := range errorsChan {
		return err
	}

	// Process results
	for execResult := range resultsChan {
		ctx.SetStepResult(execResult.step.ID, execResult.result)
		state.CompletedSteps = append(state.CompletedSteps, execResult.result)

		// Add rollback action
		if execResult.step.Rollback != nil && execResult.result.Success {
			rollbackAction := &RollbackAction{
				StepID: execResult.step.ID,
				Action: execResult.step.Rollback,
				Result: execResult.result,
			}
			ctx.AddRollbackAction(rollbackAction)
		}

		// Check if step failed and is required
		if !execResult.result.Success && execResult.step.Required {
			return fmt.Errorf("required step %s failed", execResult.step.ID)
		}
	}

	return nil
}

// stepExecutionResult holds the result of a step execution.
type stepExecutionResult struct {
	step   *Step
	result *StepResult
	err    error
}

// Resume resumes a workflow from a saved state.
func (e *Executor) Resume(stateID string, ctx *ExecutionContext) (*ExecutionState, error) {
	// Load saved state
	savedState, err := e.state.LoadState(stateID)
	if err != nil {
		return nil, fmt.Errorf("failed to load state: %w", err)
	}

	// Restore completed steps to context
	for _, result := range savedState.CompletedSteps {
		ctx.SetStepResult(result.StepID, result)
	}

	// Continue execution from where it left off
	// This is a simplified version - a full implementation would need to:
	// 1. Determine which steps have completed
	// 2. Resume from the next incomplete step
	// 3. Handle partial progress within composite steps

	return e.Execute(ctx)
}
