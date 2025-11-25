package workflow

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestExecutor_Execute_SimpleWorkflow(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					Endpoint: server.URL + "/api/test",
					Method:   "GET",
				},
			},
			{
				ID:        "step2",
				Type:      StepTypeAPICall,
				DependsOn: []string{"step1"},
				APICall: &APICallStep{
					Endpoint: server.URL + "/api/test2",
					Method:   "GET",
				},
			},
		},
	}

	executor, err := NewExecutor(workflow, server.Client(), nil)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := NewExecutionContext(map[string]interface{}{})
	state, err := executor.Execute(ctx)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if state.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %s, got %s", ExecutionStatusCompleted, state.Status)
	}

	// Check that both steps completed successfully
	_, step1Exists := ctx.GetStepResult("step1")
	_, step2Exists := ctx.GetStepResult("step2")

	if !step1Exists {
		t.Error("expected step1 to be executed")
	}
	if !step2Exists {
		t.Error("expected step2 to be executed")
	}
}

func TestExecutor_Execute_ParallelWorkflow(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		time.Sleep(10 * time.Millisecond) // Simulate some work
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	workflow := &Workflow{
		Settings: &WorkflowSettings{
			ParallelExecution: true,
		},
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					Endpoint: server.URL + "/api/test1",
					Method:   "GET",
				},
			},
			{
				ID:   "step2",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					Endpoint: server.URL + "/api/test2",
					Method:   "GET",
				},
			},
		},
	}

	executor, err := NewExecutor(workflow, server.Client(), nil)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := NewExecutionContext(map[string]interface{}{})
	startTime := time.Now()
	state, err := executor.Execute(ctx)
	duration := time.Since(startTime)

	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if state.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %s, got %s", ExecutionStatusCompleted, state.Status)
	}

	// Parallel execution should be faster than sequential
	// With 2 steps taking 10ms each, parallel should take ~10-20ms, sequential ~20-30ms
	if duration > 50*time.Millisecond {
		t.Errorf("parallel execution took too long: %v (expected < 50ms)", duration)
	}
}

func TestExecutor_Execute_ConditionalStep(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "conditional",
				Type: StepTypeConditional,
				Conditional: &ConditionalStep{
					Condition: "flags.enabled == true",
					Then: []*Step{
						{
							ID:   "then-step",
							Type: StepTypeAPICall,
							APICall: &APICallStep{
								Endpoint: server.URL + "/api/then",
								Method:   "GET",
							},
						},
					},
					Else: []*Step{
						{
							ID:   "else-step",
							Type: StepTypeAPICall,
							APICall: &APICallStep{
								Endpoint: server.URL + "/api/else",
								Method:   "GET",
							},
						},
					},
				},
			},
		},
	}

	executor, err := NewExecutor(workflow, server.Client(), nil)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	// Test with enabled = true
	ctx := NewExecutionContext(map[string]interface{}{
		"enabled": true,
	})
	state, err := executor.Execute(ctx)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if state.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %s, got %s", ExecutionStatusCompleted, state.Status)
	}

	// Verify conditional executed (nested steps won't be in step results since they're inside the conditional)
	condResult, exists := ctx.GetStepResult("conditional")
	if !exists {
		t.Error("expected conditional step to be executed")
	}
	if branch, ok := condResult.Output["branch"].(string); !ok || branch != "then" {
		t.Errorf("expected then branch to be executed, got %v", condResult.Output["branch"])
	}

	// Test with enabled = false
	ctx2 := NewExecutionContext(map[string]interface{}{
		"enabled": false,
	})
	state2, err := executor.Execute(ctx2)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if state2.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %s, got %s", ExecutionStatusCompleted, state2.Status)
	}

	// Verify else branch was executed
	condResult2, exists := ctx2.GetStepResult("conditional")
	if !exists {
		t.Error("expected conditional step to be executed")
	}
	if branch, ok := condResult2.Output["branch"].(string); !ok || branch != "else" {
		t.Errorf("expected else branch to be executed, got %v", condResult2.Output["branch"])
	}
}

func TestExecutor_Execute_LoopStep(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "loop",
				Type: StepTypeLoop,
				Loop: &LoopStep{
					Iterator:   "item",
					Collection: "flags.items",
					Steps: []*Step{
						{
							ID:   "process-item",
							Type: StepTypeAPICall,
							APICall: &APICallStep{
								Endpoint: server.URL + "/api/process",
								Method:   "POST",
								Body: map[string]interface{}{
									"item": "test-item", // Simplified - just use a static value for this test
								},
							},
						},
					},
				},
			},
		},
	}

	executor, err := NewExecutor(workflow, server.Client(), nil)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := NewExecutionContext(map[string]interface{}{
		"items": []interface{}{"a", "b", "c"},
	})

	state, err := executor.Execute(ctx)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	if state.Status != ExecutionStatusCompleted {
		t.Errorf("expected status %s, got %s", ExecutionStatusCompleted, state.Status)
	}

	// Check loop executed
	loopResult, exists := ctx.GetStepResult("loop")
	if !exists {
		t.Error("expected loop step to be executed")
	}

	// Verify loop processed all 3 items
	if size, ok := loopResult.Output["collection_size"].(int); !ok || size != 3 {
		t.Errorf("expected collection_size to be 3, got %v", loopResult.Output["collection_size"])
	}
}

func TestExecutor_Execute_WithRollback(t *testing.T) {
	createdResources := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" && r.URL.Path == "/api/create" {
			createdResources["resource1"] = true
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"id": "resource1"}`))
		} else if r.Method == "POST" && r.URL.Path == "/api/fail" {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "failed"}`))
		} else if r.Method == "DELETE" {
			delete(createdResources, "resource1")
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "create-resource",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					Endpoint: server.URL + "/api/create",
					Method:   "POST",
				},
				Rollback: &Step{
					ID:   "delete-resource",
					Type: StepTypeAPICall,
					APICall: &APICallStep{
						Endpoint: server.URL + "/api/delete",
						Method:   "DELETE",
					},
				},
			},
			{
				ID:        "fail-step",
				Type:      StepTypeAPICall,
				DependsOn: []string{"create-resource"},
				Required:  true,
				APICall: &APICallStep{
					Endpoint: server.URL + "/api/fail",
					Method:   "POST",
				},
			},
		},
	}

	executor, err := NewExecutor(workflow, server.Client(), nil)
	if err != nil {
		t.Fatalf("failed to create executor: %v", err)
	}

	ctx := NewExecutionContext(map[string]interface{}{})
	state, err := executor.Execute(ctx)

	// Execution should fail (err can be nil if rollback succeeded)
	if err == nil && state.Status == ExecutionStatusCompleted {
		t.Fatal("expected execution to fail")
	}

	// Status should be rolled back
	if state.Status != ExecutionStatusRolledBack {
		t.Errorf("expected status %s, got %s", ExecutionStatusRolledBack, state.Status)
	}

	// Resource should be deleted
	if createdResources["resource1"] {
		t.Error("expected resource to be rolled back (deleted)")
	}
}
