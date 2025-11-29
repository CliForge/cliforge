package workflow

import (
	"fmt"
	"testing"
)

func TestRollbackManager_ExecuteRollback_NoActions(t *testing.T) {
	rm := NewRollbackManager()
	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	err := rm.ExecuteRollback(ctx, executor)
	if err != nil {
		t.Errorf("expected no error for empty rollback actions, got: %v", err)
	}
}

func TestRollbackManager_ExecuteRollback_Success(t *testing.T) {
	rm := NewRollbackManager()
	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add rollback actions
	action1 := &RollbackAction{
		StepID: "step1",
		Action: &Step{
			ID:   "rollback-step1",
			Type: StepTypeNoop,
		},
		Result: &StepResult{Success: true},
	}
	action2 := &RollbackAction{
		StepID: "step2",
		Action: &Step{
			ID:   "rollback-step2",
			Type: StepTypeNoop,
		},
		Result: &StepResult{Success: true},
	}

	ctx.AddRollbackAction(action1)
	ctx.AddRollbackAction(action2)

	err := rm.ExecuteRollback(ctx, executor)
	if err != nil {
		t.Errorf("expected successful rollback, got error: %v", err)
	}
}

func TestRollbackManager_ExecuteRollback_NilAction(t *testing.T) {
	rm := NewRollbackManager()
	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add rollback action with nil Action
	action := &RollbackAction{
		StepID: "step1",
		Action: nil,
		Result: &StepResult{Success: true},
	}
	ctx.AddRollbackAction(action)

	err := rm.ExecuteRollback(ctx, executor)
	if err != nil {
		t.Errorf("expected successful rollback with nil action, got error: %v", err)
	}
}

func TestRollbackManager_ExecuteRollback_ContinueOnError(t *testing.T) {
	rm := NewRollbackManager()
	rm.SetContinueOnError(true)

	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add a failing rollback action (missing configuration)
	action1 := &RollbackAction{
		StepID: "step1",
		Action: &Step{
			ID:   "rollback-step1",
			Type: StepTypeAPICall, // Will fail due to missing APICall config
		},
		Result: &StepResult{Success: true},
	}
	// Add a successful rollback action
	action2 := &RollbackAction{
		StepID: "step2",
		Action: &Step{
			ID:   "rollback-step2",
			Type: StepTypeNoop,
		},
		Result: &StepResult{Success: true},
	}

	ctx.AddRollbackAction(action1)
	ctx.AddRollbackAction(action2)

	err := rm.ExecuteRollback(ctx, executor)
	// Should complete with errors but not stop
	if err == nil {
		t.Error("expected error from failed rollback action")
	}
	if !contains(err.Error(), "error(s)") {
		t.Errorf("expected combined errors message, got: %v", err)
	}
}

func TestRollbackManager_ExecuteRollback_StopOnError(t *testing.T) {
	rm := NewRollbackManager()
	rm.SetContinueOnError(false)

	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add a failing rollback action
	action1 := &RollbackAction{
		StepID: "step1",
		Action: &Step{
			ID:   "rollback-step1",
			Type: StepTypeAPICall, // Will fail due to missing APICall config
		},
		Result: &StepResult{Success: true},
	}
	// Add another action that won't be executed
	action2 := &RollbackAction{
		StepID: "step2",
		Action: &Step{
			ID:   "rollback-step2",
			Type: StepTypeNoop,
		},
		Result: &StepResult{Success: true},
	}

	ctx.AddRollbackAction(action1)
	ctx.AddRollbackAction(action2)

	err := rm.ExecuteRollback(ctx, executor)
	if err == nil {
		t.Error("expected error from rollback")
	}
	if !contains(err.Error(), "aborted") {
		t.Errorf("expected 'aborted' in error message, got: %v", err)
	}
}

func TestRollbackManager_ExecuteRollbackWithStatus_NoActions(t *testing.T) {
	rm := NewRollbackManager()
	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	status, err := rm.ExecuteRollbackWithStatus(ctx, executor)
	if err != nil {
		t.Errorf("expected no error for empty rollback actions, got: %v", err)
	}
	if status.TotalActions != 0 {
		t.Errorf("expected TotalActions to be 0, got %d", status.TotalActions)
	}
}

func TestRollbackManager_ExecuteRollbackWithStatus_Success(t *testing.T) {
	rm := NewRollbackManager()
	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add successful rollback actions
	action1 := &RollbackAction{
		StepID: "step1",
		Action: &Step{
			ID:   "rollback-step1",
			Type: StepTypeNoop,
		},
		Result: &StepResult{Success: true},
	}
	action2 := &RollbackAction{
		StepID: "step2",
		Action: nil, // Nil action should be counted as successful
		Result: &StepResult{Success: true},
	}

	ctx.AddRollbackAction(action1)
	ctx.AddRollbackAction(action2)

	status, err := rm.ExecuteRollbackWithStatus(ctx, executor)
	if err != nil {
		t.Errorf("expected successful rollback, got error: %v", err)
	}
	if status.TotalActions != 2 {
		t.Errorf("expected TotalActions to be 2, got %d", status.TotalActions)
	}
	if status.ExecutedActions != 2 {
		t.Errorf("expected ExecutedActions to be 2, got %d", status.ExecutedActions)
	}
	if status.SuccessfulActions != 2 {
		t.Errorf("expected SuccessfulActions to be 2, got %d", status.SuccessfulActions)
	}
	if status.FailedActions != 0 {
		t.Errorf("expected FailedActions to be 0, got %d", status.FailedActions)
	}
}

func TestRollbackManager_ExecuteRollbackWithStatus_WithFailures(t *testing.T) {
	rm := NewRollbackManager()
	rm.SetContinueOnError(true)

	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add a failing rollback action
	action1 := &RollbackAction{
		StepID: "step1",
		Action: &Step{
			ID:   "rollback-step1",
			Type: StepTypeAPICall, // Will fail due to missing APICall config
		},
		Result: &StepResult{Success: true},
	}
	// Add a successful action
	action2 := &RollbackAction{
		StepID: "step2",
		Action: &Step{
			ID:   "rollback-step2",
			Type: StepTypeNoop,
		},
		Result: &StepResult{Success: true},
	}

	ctx.AddRollbackAction(action1)
	ctx.AddRollbackAction(action2)

	status, err := rm.ExecuteRollbackWithStatus(ctx, executor)
	if err == nil {
		t.Error("expected error from failed rollback")
	}
	if status.TotalActions != 2 {
		t.Errorf("expected TotalActions to be 2, got %d", status.TotalActions)
	}
	if status.FailedActions != 1 {
		t.Errorf("expected FailedActions to be 1, got %d", status.FailedActions)
	}
	if status.SuccessfulActions != 1 {
		t.Errorf("expected SuccessfulActions to be 1, got %d", status.SuccessfulActions)
	}
	if len(status.Errors) != 1 {
		t.Errorf("expected 1 error in status, got %d", len(status.Errors))
	}
}

func TestRollbackManager_SetContinueOnError(t *testing.T) {
	rm := NewRollbackManager()

	// Default should be true
	if !rm.continueOnError {
		t.Error("expected default continueOnError to be true")
	}

	// Set to false
	rm.SetContinueOnError(false)
	if rm.continueOnError {
		t.Error("expected continueOnError to be false after setting")
	}

	// Set back to true
	rm.SetContinueOnError(true)
	if !rm.continueOnError {
		t.Error("expected continueOnError to be true after setting")
	}
}

// TestRollbackManager_ExecuteRollbackWithStatus_StopOnError tests that rollback stops on first error when configured
func TestRollbackManager_ExecuteRollbackWithStatus_StopOnError(t *testing.T) {
	rm := NewRollbackManager()
	rm.SetContinueOnError(false)

	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add a failing action first (will be executed last due to reverse order)
	action1 := &RollbackAction{
		StepID: "step1",
		Action: &Step{
			ID:   "rollback-step1",
			Type: StepTypeNoop,
		},
		Result: &StepResult{Success: true},
	}
	// Add a failing action
	action2 := &RollbackAction{
		StepID: "step2",
		Action: &Step{
			ID:      "rollback-step2",
			Type:    StepTypeAPICall,
			APICall: nil, // This will cause an error
		},
		Result: &StepResult{Success: true},
	}

	ctx.AddRollbackAction(action1)
	ctx.AddRollbackAction(action2)

	status, err := rm.ExecuteRollbackWithStatus(ctx, executor)
	if err == nil {
		t.Error("expected error from rollback")
	}
	if !contains(err.Error(), "aborted") {
		t.Errorf("expected 'aborted' in error message, got: %v", err)
	}
	if status.FailedActions == 0 {
		t.Error("expected at least one failed action")
	}
}

// TestRollbackAction_NilResult tests rollback with nil result
func TestRollbackManager_ExecuteRollback_FailedStepResult(t *testing.T) {
	rm := NewRollbackManager()
	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add action that returns failed result
	action := &RollbackAction{
		StepID: "step1",
		Action: &Step{
			ID:   "rollback-step1",
			Type: StepTypeAPICall, // Will fail
		},
		Result: &StepResult{Success: true},
	}
	ctx.AddRollbackAction(action)

	err := rm.ExecuteRollback(ctx, executor)
	if err == nil {
		t.Error("expected error from failed rollback")
	}
}

// Ensure error format matches expected patterns
func TestRollbackManager_ErrorMessages(t *testing.T) {
	rm := NewRollbackManager()
	rm.SetContinueOnError(true)

	ctx := NewExecutionContext(map[string]interface{}{})
	executor := NewStepExecutor(nil, nil)

	// Add failing action
	action := &RollbackAction{
		StepID: "test-step",
		Action: &Step{
			ID:   "rollback-step",
			Type: StepTypeAPICall,
		},
		Result: &StepResult{Success: true},
	}
	ctx.AddRollbackAction(action)

	err := rm.ExecuteRollback(ctx, executor)
	if err == nil {
		t.Fatal("expected error")
	}

	errMsg := err.Error()
	if !contains(errMsg, "error(s)") {
		t.Errorf("expected error message to contain 'error(s)', got: %s", errMsg)
	}
}

// Test ordering of rollback actions (LIFO)
func TestRollbackManager_ExecuteOrder(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	// Add multiple actions
	for i := 1; i <= 3; i++ {
		action := &RollbackAction{
			StepID: fmt.Sprintf("step%d", i),
			Action: &Step{
				ID:   fmt.Sprintf("rollback-step%d", i),
				Type: StepTypeNoop,
			},
			Result: &StepResult{Success: true},
		}
		ctx.AddRollbackAction(action)
	}

	// Verify they're returned in reverse order
	actions := ctx.GetRollbackActions()
	if len(actions) != 3 {
		t.Fatalf("expected 3 actions, got %d", len(actions))
	}
	if actions[0].StepID != "step3" {
		t.Errorf("expected first action to be step3, got %s", actions[0].StepID)
	}
	if actions[1].StepID != "step2" {
		t.Errorf("expected second action to be step2, got %s", actions[1].StepID)
	}
	if actions[2].StepID != "step1" {
		t.Errorf("expected third action to be step1, got %s", actions[2].StepID)
	}
}
