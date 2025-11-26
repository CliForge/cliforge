package workflow

import (
	"testing"
	"time"
)

func TestExecutionContext_GetVariable(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{
		"initial_flag": "flag_value",
	})

	// Set a variable
	ctx.SetVariable("test_key", "test_value")

	// Get existing variable
	value, exists := ctx.GetVariable("test_key")
	if !exists {
		t.Error("expected variable to exist")
	}
	if value != "test_value" {
		t.Errorf("expected 'test_value', got %v", value)
	}

	// Get non-existent variable
	_, exists = ctx.GetVariable("non_existent")
	if exists {
		t.Error("expected variable to not exist")
	}
}

func TestExecutionContext_SetAndGetVariable(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	// Test different types
	testCases := []struct {
		key   string
		value interface{}
	}{
		{"string_var", "hello"},
		{"int_var", 42},
		{"float_var", 3.14},
		{"bool_var", true},
		{"map_var", map[string]interface{}{"nested": "value"}},
		{"slice_var", []interface{}{"a", "b", "c"}},
		{"nil_var", nil},
	}

	for _, tc := range testCases {
		ctx.SetVariable(tc.key, tc.value)
		value, exists := ctx.GetVariable(tc.key)
		if !exists {
			t.Errorf("variable %s should exist", tc.key)
		}
		// For map and slice, just check it's not nil
		if value == nil && tc.value != nil {
			t.Errorf("expected non-nil value for %s", tc.key)
		}
	}
}

func TestExecutionContext_Variables_ThreadSafe(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(n int) {
			ctx.SetVariable("key", n)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			ctx.GetVariable("key")
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestExecutionContext_Clone_PreservesVariables(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{
		"flag1": "value1",
	})

	ctx.SetVariable("var1", "original")
	ctx.SetVariable("var2", 42)

	clone := ctx.Clone()

	// Check that variables are copied
	val1, exists := clone.GetVariable("var1")
	if !exists || val1 != "original" {
		t.Error("expected var1 to be copied to clone")
	}

	val2, exists := clone.GetVariable("var2")
	if !exists || val2 != 42 {
		t.Error("expected var2 to be copied to clone")
	}

	// Modify clone - should not affect original
	clone.SetVariable("var1", "modified")
	clone.SetVariable("var3", "new")

	// Check original is unchanged
	origVal, _ := ctx.GetVariable("var1")
	if origVal != "original" {
		t.Error("modifying clone should not affect original")
	}

	// Check clone has new variable
	_, exists = ctx.GetVariable("var3")
	if exists {
		t.Error("new variable in clone should not appear in original")
	}

	// Check flags are shared (same reference)
	if len(clone.Flags) != len(ctx.Flags) {
		t.Error("expected flags to be shared")
	}
}

func TestExecutionContext_Clone_SharesStepResults(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	result := &StepResult{
		StepID:  "step1",
		Success: true,
		Output:  map[string]interface{}{"key": "value"},
	}
	ctx.SetStepResult("step1", result)

	clone := ctx.Clone()

	// Step results should be shared
	cloneResult, exists := clone.GetStepResult("step1")
	if !exists {
		t.Error("expected step result to be shared with clone")
	}
	if cloneResult.StepID != "step1" {
		t.Error("expected same step result in clone")
	}

	// Modifying in clone should affect original (shared reference)
	newResult := &StepResult{
		StepID:  "step2",
		Success: true,
	}
	clone.SetStepResult("step2", newResult)

	_, exists = ctx.GetStepResult("step2")
	if !exists {
		t.Error("expected step result added in clone to appear in original (shared)")
	}
}

func TestExecutionContext_SetAndGetStepResult(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	result := &StepResult{
		StepID:    "test-step",
		Success:   true,
		Output:    map[string]interface{}{"result": "success"},
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Duration:  100 * time.Millisecond,
	}

	ctx.SetStepResult("test-step", result)

	// Get existing result
	retrieved, exists := ctx.GetStepResult("test-step")
	if !exists {
		t.Error("expected step result to exist")
	}
	if retrieved.StepID != "test-step" {
		t.Errorf("expected StepID 'test-step', got %s", retrieved.StepID)
	}
	if !retrieved.Success {
		t.Error("expected result to be successful")
	}

	// Get non-existent result
	_, exists = ctx.GetStepResult("non-existent")
	if exists {
		t.Error("expected step result to not exist")
	}

	// Check CompletedSteps
	if len(ctx.CompletedSteps) != 1 {
		t.Errorf("expected 1 completed step, got %d", len(ctx.CompletedSteps))
	}
	if ctx.CompletedSteps[0].StepID != "test-step" {
		t.Error("expected completed step to match")
	}
}

func TestExecutionContext_AddAndGetRollbackActions(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	action1 := &RollbackAction{
		StepID: "step1",
		Action: &Step{ID: "rollback1"},
		Result: &StepResult{Success: true},
	}
	action2 := &RollbackAction{
		StepID: "step2",
		Action: &Step{ID: "rollback2"},
		Result: &StepResult{Success: true},
	}
	action3 := &RollbackAction{
		StepID: "step3",
		Action: &Step{ID: "rollback3"},
		Result: &StepResult{Success: true},
	}

	ctx.AddRollbackAction(action1)
	ctx.AddRollbackAction(action2)
	ctx.AddRollbackAction(action3)

	// Get rollback actions - should be in reverse order (LIFO)
	actions := ctx.GetRollbackActions()
	if len(actions) != 3 {
		t.Errorf("expected 3 rollback actions, got %d", len(actions))
	}

	// Check reverse order
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

func TestExecutionContext_RollbackActions_ThreadSafe(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	done := make(chan bool)

	// Concurrent adds
	for i := 0; i < 10; i++ {
		go func(n int) {
			action := &RollbackAction{
				StepID: "step",
				Action: &Step{ID: "rollback"},
			}
			ctx.AddRollbackAction(action)
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			ctx.GetRollbackActions()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}

func TestExecutionContext_NewExecutionContext(t *testing.T) {
	flags := map[string]interface{}{
		"flag1": "value1",
		"flag2": 42,
	}

	ctx := NewExecutionContext(flags)

	if ctx.Flags == nil {
		t.Error("expected Flags to be initialized")
	}
	if ctx.Variables == nil {
		t.Error("expected Variables to be initialized")
	}
	if ctx.StepResults == nil {
		t.Error("expected StepResults to be initialized")
	}
	if ctx.CompletedSteps == nil {
		t.Error("expected CompletedSteps to be initialized")
	}
	if ctx.RollbackActions == nil {
		t.Error("expected RollbackActions to be initialized")
	}

	if ctx.Flags["flag1"] != "value1" {
		t.Error("expected flags to be set")
	}
}

func TestExecutionContext_MultipleStepResults(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	results := []*StepResult{
		{StepID: "step1", Success: true},
		{StepID: "step2", Success: true},
		{StepID: "step3", Success: false},
	}

	for _, result := range results {
		ctx.SetStepResult(result.StepID, result)
	}

	// Check all results exist
	for _, result := range results {
		retrieved, exists := ctx.GetStepResult(result.StepID)
		if !exists {
			t.Errorf("expected step %s to exist", result.StepID)
		}
		if retrieved.Success != result.Success {
			t.Errorf("expected success=%v for %s", result.Success, result.StepID)
		}
	}

	// Check completed steps order
	if len(ctx.CompletedSteps) != 3 {
		t.Errorf("expected 3 completed steps, got %d", len(ctx.CompletedSteps))
	}
	for i, result := range results {
		if ctx.CompletedSteps[i].StepID != result.StepID {
			t.Errorf("expected completed step %d to be %s, got %s",
				i, result.StepID, ctx.CompletedSteps[i].StepID)
		}
	}
}

func TestExecutionContext_EmptyRollbackActions(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	actions := ctx.GetRollbackActions()
	if len(actions) != 0 {
		t.Errorf("expected 0 rollback actions, got %d", len(actions))
	}
}

func TestExecutionContext_Clone_SharesRollbackActions(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{})

	action := &RollbackAction{
		StepID: "step1",
		Action: &Step{ID: "rollback1"},
	}
	ctx.AddRollbackAction(action)

	clone := ctx.Clone()

	// Rollback actions should be shared
	cloneActions := clone.GetRollbackActions()
	if len(cloneActions) != 1 {
		t.Errorf("expected 1 rollback action in clone, got %d", len(cloneActions))
	}

	// Note: The slice reference is shared in Clone, so modifications affect both
	// This is by design for rollback actions
}
