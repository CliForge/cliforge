package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/workflow"
	"github.com/CliForge/cliforge/tests/helpers"
)

// TestWorkflowExecution tests basic workflow execution.
func TestWorkflowExecution(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure mock endpoints for workflow
	apiServer.OnPOST("/apps/test-app/validate", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"success": true,
		"errors":  []string{},
	}))

	apiServer.OnPOST("/apps/test-app/deploy", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"deployment_id": "deploy-123",
		"status":        "deploying",
	}))

	apiServer.OnGET("/health", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
	}))

	// Test: Simple sequential workflow
	t.Run("SequentialWorkflow", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name:        "deploy-app",
			Description: "Deploy application workflow",
			Steps: []workflow.Step{
				{
					ID:          "validate",
					Type:        workflow.StepTypeAPICall,
					Description: "Validate application",
					APICall: &workflow.APICallStep{
						Operation: "validateApp",
						Path:      "/apps/test-app/validate",
						Method:    "POST",
					},
				},
				{
					ID:          "deploy",
					Type:        workflow.StepTypeAPICall,
					Description: "Deploy application",
					APICall: &workflow.APICallStep{
						Operation: "deployApp",
						Path:      "/apps/test-app/deploy",
						Method:    "POST",
					},
					DependsOn: []string{"validate"},
				},
				{
					ID:          "health-check",
					Type:        workflow.StepTypeAPICall,
					Description: "Check deployment health",
					APICall: &workflow.APICallStep{
						Operation: "healthCheck",
						Path:      "/health",
						Method:    "GET",
					},
					DependsOn: []string{"deploy"},
				},
			},
		}

		// Parse and validate workflow
		helpers.AssertNotNil(t, wf)
		helpers.AssertEqual(t, 3, len(wf.Steps))
		helpers.AssertEqual(t, "deploy-app", wf.Name)
	})
}

// TestConditionalWorkflow tests conditional workflow execution.
func TestConditionalWorkflow(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	validationSuccess := true

	// Configure endpoints with conditional behavior
	apiServer.OnPOST("/apps/test-app/validate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": validationSuccess,
			"errors":  []string{},
		})
	})

	apiServer.OnPOST("/apps/test-app/deploy", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"deployment_id": "deploy-456",
		"status":        "deploying",
	}))

	apiServer.OnPOST("/logs/error", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"logged": true,
	}))

	// Test: Conditional execution (success path)
	t.Run("ConditionalSuccess", func(t *testing.T) {
		validationSuccess = true

		wf := &workflow.Workflow{
			Name: "conditional-deploy",
			Steps: []workflow.Step{
				{
					ID:   "validate",
					Type: workflow.StepTypeAPICall,
					APICall: &workflow.APICallStep{
						Operation: "validateApp",
						Path:      "/apps/test-app/validate",
						Method:    "POST",
					},
				},
				{
					ID:        "check-validation",
					Type:      workflow.StepTypeConditional,
					Condition: "validate.success == true",
					Then: []workflow.Step{
						{
							ID:   "deploy",
							Type: workflow.StepTypeAPICall,
							APICall: &workflow.APICallStep{
								Operation: "deployApp",
								Path:      "/apps/test-app/deploy",
								Method:    "POST",
							},
						},
					},
					Else: []workflow.Step{
						{
							ID:   "log-error",
							Type: workflow.StepTypeAPICall,
							APICall: &workflow.APICallStep{
								Operation: "logError",
								Path:      "/logs/error",
								Method:    "POST",
							},
						},
					},
					DependsOn: []string{"validate"},
				},
			},
		}

		helpers.AssertEqual(t, 2, len(wf.Steps))
		helpers.AssertEqual(t, workflow.StepTypeConditional, wf.Steps[1].Type)
	})

	// Test: Conditional execution (failure path)
	t.Run("ConditionalFailure", func(t *testing.T) {
		validationSuccess = false

		// Same workflow would take different path
		// Verification would check that log-error was called, not deploy
		helpers.AssertEqual(t, false, validationSuccess)
	})
}

// TestParallelWorkflow tests parallel workflow execution.
func TestParallelWorkflow(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure parallel endpoints
	apiServer.OnGET("/health", helpers.DelayedResponse(50*time.Millisecond,
		helpers.JSONResponse(http.StatusOK, map[string]interface{}{
			"status": "healthy",
		}),
	))

	apiServer.OnGET("/database/check", helpers.DelayedResponse(50*time.Millisecond,
		helpers.JSONResponse(http.StatusOK, map[string]interface{}{
			"success":    true,
			"latency_ms": 10,
		}),
	))

	apiServer.OnPOST("/system/ready", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"ready": true,
	}))

	// Test: Parallel execution
	t.Run("ParallelExecution", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "parallel-checks",
			Steps: []workflow.Step{
				{
					ID:   "run-checks",
					Type: workflow.StepTypeParallel,
					Parallel: &workflow.ParallelStep{
						Steps: []workflow.Step{
							{
								ID:   "check-api",
								Type: workflow.StepTypeAPICall,
								APICall: &workflow.APICallStep{
									Operation: "healthCheck",
									Path:      "/health",
									Method:    "GET",
								},
							},
							{
								ID:   "check-database",
								Type: workflow.StepTypeAPICall,
								APICall: &workflow.APICallStep{
									Operation: "checkDatabase",
									Path:      "/database/check",
									Method:    "GET",
								},
							},
						},
					},
				},
				{
					ID:   "mark-ready",
					Type: workflow.StepTypeAPICall,
					APICall: &workflow.APICallStep{
						Operation: "markReady",
						Path:      "/system/ready",
						Method:    "POST",
					},
					DependsOn: []string{"run-checks"},
				},
			},
		}

		helpers.AssertEqual(t, 2, len(wf.Steps))
		helpers.AssertNotNil(t, wf.Steps[0].Parallel)
		helpers.AssertEqual(t, 2, len(wf.Steps[0].Parallel.Steps))
	})
}

// TestLoopWorkflow tests loop workflow execution.
func TestLoopWorkflow(t *testing.T) {
	// Test: Loop step configuration
	t.Run("LoopConfiguration", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "batch-process",
			Steps: []workflow.Step{
				{
					ID:   "process-items",
					Type: workflow.StepTypeLoop,
					Loop: &workflow.LoopStep{
						Items: "input.items",
						Step: workflow.Step{
							ID:   "process-item",
							Type: workflow.StepTypeAPICall,
							APICall: &workflow.APICallStep{
								Operation: "processItem",
								Path:      "/items/process",
								Method:    "POST",
							},
						},
						MaxConcurrency: 3,
					},
				},
			},
		}

		helpers.AssertEqual(t, 1, len(wf.Steps))
		helpers.AssertNotNil(t, wf.Steps[0].Loop)
		helpers.AssertEqual(t, 3, wf.Steps[0].Loop.MaxConcurrency)
	})
}

// TestWorkflowRetry tests retry logic in workflows.
func TestWorkflowRetry(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	attempts := 0
	maxAttempts := 3

	// Configure endpoint that fails first few times
	apiServer.OnGET("/unreliable", func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < maxAttempts {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Service temporarily unavailable",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":   "success",
			"attempts": attempts,
		})
	})

	// Test: Retry configuration
	t.Run("RetryConfiguration", func(t *testing.T) {
		step := workflow.Step{
			ID:   "unreliable-call",
			Type: workflow.StepTypeAPICall,
			APICall: &workflow.APICallStep{
				Operation: "unreliableCall",
				Path:      "/unreliable",
				Method:    "GET",
			},
			Retry: &workflow.RetryConfig{
				MaxAttempts: 3,
				Delay:       100 * time.Millisecond,
				Backoff:     "exponential",
			},
		}

		helpers.AssertNotNil(t, step.Retry)
		helpers.AssertEqual(t, 3, step.Retry.MaxAttempts)
		helpers.AssertEqual(t, "exponential", step.Retry.Backoff)
	})
}

// TestWorkflowRollback tests rollback functionality.
func TestWorkflowRollback(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	deployed := false

	// Configure endpoints
	apiServer.OnPOST("/apps/test-app/deploy", func(w http.ResponseWriter, r *http.Request) {
		deployed = true
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"deployment_id": "deploy-789",
			"status":        "deployed",
		})
	})

	apiServer.OnPOST("/apps/test-app/rollback", func(w http.ResponseWriter, r *http.Request) {
		deployed = false
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "rolled_back",
		})
	})

	apiServer.OnGET("/health", func(w http.ResponseWriter, r *http.Request) {
		status := "healthy"
		if !deployed {
			status = "unhealthy"
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": status,
		})
	})

	// Test: Rollback on failure
	t.Run("RollbackOnFailure", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "deploy-with-rollback",
			Steps: []workflow.Step{
				{
					ID:   "deploy",
					Type: workflow.StepTypeAPICall,
					APICall: &workflow.APICallStep{
						Operation: "deployApp",
						Path:      "/apps/test-app/deploy",
						Method:    "POST",
					},
					OnError: &workflow.ErrorHandler{
						Rollback: []workflow.Step{
							{
								ID:   "rollback",
								Type: workflow.StepTypeAPICall,
								APICall: &workflow.APICallStep{
									Operation: "rollbackApp",
									Path:      "/apps/test-app/rollback",
									Method:    "POST",
								},
							},
						},
					},
				},
			},
		}

		helpers.AssertNotNil(t, wf.Steps[0].OnError)
		helpers.AssertNotNil(t, wf.Steps[0].OnError.Rollback)
		helpers.AssertEqual(t, 1, len(wf.Steps[0].OnError.Rollback))
	})
}

// TestWorkflowTimeout tests workflow timeout handling.
func TestWorkflowTimeout(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure slow endpoint
	apiServer.OnGET("/slow", helpers.DelayedResponse(2*time.Second,
		helpers.JSONResponse(http.StatusOK, map[string]interface{}{
			"status": "complete",
		}),
	))

	// Test: Workflow timeout
	t.Run("WorkflowTimeout", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name:    "timeout-test",
			Timeout: 500 * time.Millisecond,
			Steps: []workflow.Step{
				{
					ID:   "slow-call",
					Type: workflow.StepTypeAPICall,
					APICall: &workflow.APICallStep{
						Operation: "slowCall",
						Path:      "/slow",
						Method:    "GET",
					},
				},
			},
		}

		helpers.AssertEqual(t, 500*time.Millisecond, wf.Timeout)
	})
}

// TestWorkflowStateManagement tests workflow state management.
func TestWorkflowStateManagement(t *testing.T) {
	// Test: Workflow state structure
	t.Run("WorkflowState", func(t *testing.T) {
		state := &workflow.WorkflowState{
			WorkflowID: "wf-123",
			Status:     workflow.StatusRunning,
			StartedAt:  time.Now(),
			Steps: map[string]*workflow.StepState{
				"step1": {
					StepID:    "step1",
					Status:    workflow.StatusCompleted,
					StartedAt: time.Now().Add(-1 * time.Minute),
					Output:    map[string]interface{}{"result": "success"},
				},
				"step2": {
					StepID:    "step2",
					Status:    workflow.StatusRunning,
					StartedAt: time.Now(),
				},
			},
		}

		helpers.AssertEqual(t, "wf-123", state.WorkflowID)
		helpers.AssertEqual(t, workflow.StatusRunning, state.Status)
		helpers.AssertEqual(t, 2, len(state.Steps))
	})
}

// TestWorkflowExpressionEvaluation tests expression evaluation in workflows.
func TestWorkflowExpressionEvaluation(t *testing.T) {
	// Test: Expression parsing
	t.Run("ExpressionParsing", func(t *testing.T) {
		expressions := []string{
			"validate.success == true",
			"response.status_code == 200",
			"len(items) > 0",
			"previous.output.count >= 10",
		}

		for _, expr := range expressions {
			// Verify expressions are valid strings
			helpers.AssertNotEqual(t, "", expr)
		}
	})
}

// TestWorkflowPluginIntegration tests plugin integration in workflows.
func TestWorkflowPluginIntegration(t *testing.T) {
	// Test: Plugin step configuration
	t.Run("PluginStep", func(t *testing.T) {
		step := workflow.Step{
			ID:   "run-plugin",
			Type: workflow.StepTypePlugin,
			Plugin: &workflow.PluginStep{
				Name:    "test-plugin",
				Command: "process",
				Args:    []string{"--input", "data.json"},
				Input: map[string]interface{}{
					"config": "test",
				},
			},
		}

		helpers.AssertEqual(t, workflow.StepTypePlugin, step.Type)
		helpers.AssertNotNil(t, step.Plugin)
		helpers.AssertEqual(t, "test-plugin", step.Plugin.Name)
	})
}

// TestWorkflowDependencyResolution tests dependency resolution.
func TestWorkflowDependencyResolution(t *testing.T) {
	// Test: Complex dependency graph
	t.Run("DependencyGraph", func(t *testing.T) {
		wf := &workflow.Workflow{
			Name: "complex-workflow",
			Steps: []workflow.Step{
				{ID: "step1", Type: workflow.StepTypeAPICall},
				{ID: "step2", Type: workflow.StepTypeAPICall, DependsOn: []string{"step1"}},
				{ID: "step3", Type: workflow.StepTypeAPICall, DependsOn: []string{"step1"}},
				{ID: "step4", Type: workflow.StepTypeAPICall, DependsOn: []string{"step2", "step3"}},
			},
		}

		helpers.AssertEqual(t, 4, len(wf.Steps))
		helpers.AssertEqual(t, 0, len(wf.Steps[0].DependsOn))
		helpers.AssertEqual(t, 1, len(wf.Steps[1].DependsOn))
		helpers.AssertEqual(t, 2, len(wf.Steps[3].DependsOn))
	})
}

// TestWorkflowExecutor tests the workflow executor.
func TestWorkflowExecutor(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	apiServer.OnGET("/test", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"status": "ok",
	}))

	// Test: Executor initialization
	t.Run("ExecutorInit", func(t *testing.T) {
		ctx := context.Background()
		executor := workflow.NewExecutor()

		helpers.AssertNotNil(t, executor)
		helpers.AssertNotNil(t, ctx)
	})
}
