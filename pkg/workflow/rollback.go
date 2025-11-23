package workflow

import (
	"fmt"
)

// RollbackManager manages rollback execution.
type RollbackManager struct {
	// Configuration
	continueOnError bool
}

// NewRollbackManager creates a new rollback manager.
func NewRollbackManager() *RollbackManager {
	return &RollbackManager{
		continueOnError: true, // Continue rolling back even if one rollback fails
	}
}

// ExecuteRollback executes rollback actions in reverse order.
func (rm *RollbackManager) ExecuteRollback(ctx *ExecutionContext, executor *StepExecutor) error {
	actions := ctx.GetRollbackActions()

	if len(actions) == 0 {
		// No rollback actions to execute
		return nil
	}

	fmt.Printf("Executing rollback for %d steps...\n", len(actions))

	var errors []error

	// Execute rollback actions in reverse order (LIFO)
	for i, action := range actions {
		fmt.Printf("Rolling back step %s (%d/%d)...\n", action.StepID, i+1, len(actions))

		if action.Action == nil {
			// No rollback action defined for this step
			continue
		}

		// Execute the rollback action
		result, err := executor.ExecuteStep(action.Action, ctx)

		if err != nil || (result != nil && !result.Success) {
			rollbackErr := fmt.Errorf("rollback of step %s failed: %w", action.StepID, err)
			errors = append(errors, rollbackErr)

			// Log the error
			fmt.Printf("Warning: %v\n", rollbackErr)

			if !rm.continueOnError {
				// Stop rollback on first error
				return fmt.Errorf("rollback aborted after %d actions: %w", i+1, rollbackErr)
			}
		} else {
			fmt.Printf("Successfully rolled back step %s\n", action.StepID)
		}
	}

	// Return combined errors if any occurred
	if len(errors) > 0 {
		return fmt.Errorf("rollback completed with %d error(s): %v", len(errors), errors)
	}

	fmt.Printf("Rollback completed successfully\n")
	return nil
}

// SetContinueOnError sets whether to continue rollback on errors.
func (rm *RollbackManager) SetContinueOnError(continueOnError bool) {
	rm.continueOnError = continueOnError
}

// RollbackStatus represents the status of a rollback operation.
type RollbackStatus struct {
	TotalActions      int
	ExecutedActions   int
	SuccessfulActions int
	FailedActions     int
	Errors            []error
}

// ExecuteRollbackWithStatus executes rollback and returns detailed status.
func (rm *RollbackManager) ExecuteRollbackWithStatus(ctx *ExecutionContext, executor *StepExecutor) (*RollbackStatus, error) {
	actions := ctx.GetRollbackActions()

	status := &RollbackStatus{
		TotalActions: len(actions),
		Errors:       make([]error, 0),
	}

	if len(actions) == 0 {
		return status, nil
	}

	fmt.Printf("Executing rollback for %d steps...\n", len(actions))

	// Execute rollback actions in reverse order (LIFO)
	for i, action := range actions {
		fmt.Printf("Rolling back step %s (%d/%d)...\n", action.StepID, i+1, len(actions))
		status.ExecutedActions++

		if action.Action == nil {
			// No rollback action defined for this step
			status.SuccessfulActions++
			continue
		}

		// Execute the rollback action
		result, err := executor.ExecuteStep(action.Action, ctx)

		if err != nil || (result != nil && !result.Success) {
			rollbackErr := fmt.Errorf("rollback of step %s failed: %w", action.StepID, err)
			status.Errors = append(status.Errors, rollbackErr)
			status.FailedActions++

			// Log the error
			fmt.Printf("Warning: %v\n", rollbackErr)

			if !rm.continueOnError {
				// Stop rollback on first error
				return status, fmt.Errorf("rollback aborted after %d actions: %w", i+1, rollbackErr)
			}
		} else {
			fmt.Printf("Successfully rolled back step %s\n", action.StepID)
			status.SuccessfulActions++
		}
	}

	// Return error if any rollbacks failed
	if status.FailedActions > 0 {
		return status, fmt.Errorf("rollback completed with %d error(s)", status.FailedActions)
	}

	fmt.Printf("Rollback completed successfully\n")
	return status, nil
}
