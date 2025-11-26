package progress

import (
	"context"
	"errors"
	"testing"

	"github.com/CliForge/cliforge/pkg/workflow"
)

func TestNewWatch(t *testing.T) {
	tests := []struct {
		name    string
		config  *WatchConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: &WatchConfig{
				Enabled: true,
				StreamConfig: &StreamConfig{
					Type:     StreamTypeSSE,
					Endpoint: "http://example.com/stream",
				},
				ProgressConfig: DefaultConfig(),
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watch, err := NewWatch(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if watch == nil {
					t.Error("NewWatch() returned nil")
				}
			}
		})
	}
}

func TestWatch_SetExpressionEvaluator(t *testing.T) {
	config := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: DefaultConfig(),
	}

	watch, err := NewWatch(config)
	if err != nil {
		t.Fatalf("NewWatch() error = %v", err)
	}

	// Create mock evaluator
	mockEval := &MockExpressionEvaluator{}
	watch.SetExpressionEvaluator(mockEval)

	if watch.exprEval == nil {
		t.Error("ExpressionEvaluator should be set")
	}
}

func TestWatch_IsRunning(t *testing.T) {
	config := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &Config{
			Type:    ProgressTypeSpinner,
			Enabled: false, // Disable for testing
		},
	}

	watch, err := NewWatch(config)
	if err != nil {
		t.Fatalf("NewWatch() error = %v", err)
	}

	// Should not be running initially
	if watch.IsRunning() {
		t.Error("Watch should not be running initially")
	}
}

func TestWatch_StartNotEnabled(t *testing.T) {
	config := &WatchConfig{
		Enabled: false,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: DefaultConfig(),
	}

	watch, err := NewWatch(config)
	if err != nil {
		t.Fatalf("NewWatch() error = %v", err)
	}

	ctx := context.Background()
	err = watch.Start(ctx)
	if err == nil {
		t.Error("Start() should fail when watch is not enabled")
	}
}

func TestWatch_handleEvent(t *testing.T) {
	config := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &Config{
			Type:    ProgressTypeSpinner,
			Enabled: false,
		},
		ShowLogs: true,
	}

	watch, err := NewWatch(config)
	if err != nil {
		t.Fatalf("NewWatch() error = %v", err)
	}

	tests := []struct {
		name  string
		event *Event
	}{
		{
			name: "log event",
			event: &Event{
				Type: "log",
				Data: "test log message",
			},
		},
		{
			name: "status event",
			event: &Event{
				Type: "status",
				Data: "running",
			},
		},
		{
			name: "error event",
			event: &Event{
				Type: "error",
				Data: "test error",
			},
		},
		{
			name: "message event",
			event: &Event{
				Type: "message",
				Data: "test message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := watch.handleEvent(tt.event)
			if err != nil {
				t.Errorf("handleEvent() error = %v", err)
			}
		})
	}
}

func TestWatch_shouldExit(t *testing.T) {
	tests := []struct {
		name           string
		exitConditions []*ExitCondition
		event          *Event
		expected       bool
	}{
		{
			name:           "no exit conditions",
			exitConditions: nil,
			event: &Event{
				Type: "status",
				Data: "completed",
			},
			expected: false,
		},
		{
			name: "event type match",
			exitConditions: []*ExitCondition{
				{
					EventType: "complete",
					Message:   "Done!",
				},
			},
			event: &Event{
				Type: "complete",
				Data: "finished",
			},
			expected: true,
		},
		{
			name: "event type mismatch",
			exitConditions: []*ExitCondition{
				{
					EventType: "complete",
				},
			},
			event: &Event{
				Type: "status",
				Data: "running",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &WatchConfig{
				Enabled:        true,
				StreamConfig:   DefaultStreamConfig(),
				ProgressConfig: DefaultConfig(),
				ExitConditions: tt.exitConditions,
			}

			watch, err := NewWatch(config)
			if err != nil {
				t.Fatalf("NewWatch() error = %v", err)
			}

			result := watch.shouldExit(tt.event)
			if result != tt.expected {
				t.Errorf("shouldExit() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWatch_isFatalError(t *testing.T) {
	config := &WatchConfig{
		Enabled:        true,
		StreamConfig:   DefaultStreamConfig(),
		ProgressConfig: DefaultConfig(),
	}

	watch, err := NewWatch(config)
	if err != nil {
		t.Fatalf("NewWatch() error = %v", err)
	}

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "generic error is not fatal",
			err:      errors.New("test error"),
			expected: false,
		},
		{
			name:     "context canceled is not fatal",
			err:      context.Canceled,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := watch.isFatalError(tt.err)
			if result != tt.expected {
				t.Errorf("isFatalError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSimpleExpressionEvaluator(t *testing.T) {
	mockEval := &MockExprEvaluator{
		result: true,
	}

	evaluator := NewSimpleExpressionEvaluator(mockEval)
	if evaluator == nil {
		t.Fatal("NewSimpleExpressionEvaluator() returned nil")
	}

	data := map[string]interface{}{
		"event": map[string]interface{}{
			"type": "status",
		},
	}

	result, err := evaluator.Evaluate("event.type == 'status'", data)
	if err != nil {
		t.Errorf("Evaluate() error = %v", err)
	}

	if !result {
		t.Error("Evaluate() should return true")
	}
}

func TestSimpleExpressionEvaluator_NoEvaluator(t *testing.T) {
	evaluator := NewSimpleExpressionEvaluator(nil)

	_, err := evaluator.Evaluate("test", nil)
	if err == nil {
		t.Error("Evaluate() should error when no evaluator configured")
	}
}

func TestSimpleExpressionEvaluator_NonBooleanResult(t *testing.T) {
	mockEval := &MockExprEvaluator{
		result: "not a boolean",
	}

	evaluator := NewSimpleExpressionEvaluator(mockEval)

	_, err := evaluator.Evaluate("test", nil)
	if err == nil {
		t.Error("Evaluate() should error when result is not boolean")
	}
}

func TestNewWorkflowWatchCoordinator(t *testing.T) {
	watchConfig := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &Config{
			Type:    ProgressTypeSteps,
			Enabled: false,
		},
	}

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:          "step1",
				Description: "Step 1",
			},
			{
				ID:          "step2",
				Description: "Step 2",
			},
		},
	}

	coordinator, err := NewWorkflowWatchCoordinator(watchConfig, wf)
	if err != nil {
		t.Fatalf("NewWorkflowWatchCoordinator() error = %v", err)
	}

	if coordinator == nil {
		t.Fatal("NewWorkflowWatchCoordinator() returned nil")
	}

	if len(coordinator.stepMap) != 2 {
		t.Errorf("Expected 2 steps in stepMap, got %d", len(coordinator.stepMap))
	}

	if coordinator.workflow != wf {
		t.Error("Workflow not set correctly")
	}

	if coordinator.multiStep == nil {
		t.Error("MultiStep should be initialized")
	}
}

func TestWorkflowWatchCoordinator_UpdateStepStatus(t *testing.T) {
	watchConfig := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &Config{
			Type:    ProgressTypeSteps,
			Enabled: false,
		},
	}

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:          "step1",
				Description: "Step 1",
			},
		},
	}

	coordinator, err := NewWorkflowWatchCoordinator(watchConfig, wf)
	if err != nil {
		t.Fatalf("NewWorkflowWatchCoordinator() error = %v", err)
	}

	// Test updating step status
	err = coordinator.UpdateStepStatus("step1", StepStatusRunning, "Running step 1")
	if err != nil {
		t.Errorf("UpdateStepStatus() error = %v", err)
	}

	// Test updating non-existent step
	err = coordinator.UpdateStepStatus("nonexistent", StepStatusRunning, "")
	if err == nil {
		t.Error("UpdateStepStatus() should error for non-existent step")
	}
}

func TestWorkflowWatchCoordinator_AddSubStep(t *testing.T) {
	watchConfig := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &Config{
			Type:    ProgressTypeSteps,
			Enabled: false,
		},
	}

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:          "step1",
				Description: "Step 1",
			},
		},
	}

	coordinator, err := NewWorkflowWatchCoordinator(watchConfig, wf)
	if err != nil {
		t.Fatalf("NewWorkflowWatchCoordinator() error = %v", err)
	}

	substep := &StepInfo{
		ID:          "substep1",
		Description: "Sub Step 1",
		Status:      StepStatusPending,
	}

	err = coordinator.AddSubStep("step1", substep)
	if err != nil {
		t.Errorf("AddSubStep() error = %v", err)
	}

	// Verify substep was added to stepMap
	if _, exists := coordinator.stepMap["substep1"]; !exists {
		t.Error("Substep should be added to stepMap")
	}

	// Test adding substep to non-existent parent
	err = coordinator.AddSubStep("nonexistent", substep)
	if err == nil {
		t.Error("AddSubStep() should error for non-existent parent")
	}
}

func TestWorkflowWatchCoordinator_IsRunning(t *testing.T) {
	watchConfig := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &Config{
			Type:    ProgressTypeSteps,
			Enabled: false,
		},
	}

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:          "step1",
				Description: "Step 1",
			},
		},
	}

	coordinator, err := NewWorkflowWatchCoordinator(watchConfig, wf)
	if err != nil {
		t.Fatalf("NewWorkflowWatchCoordinator() error = %v", err)
	}

	if coordinator.IsRunning() {
		t.Error("Coordinator should not be running initially")
	}
}

func TestWorkflowWatchCoordinator_SuccessFailure(t *testing.T) {
	watchConfig := &WatchConfig{
		Enabled: true,
		StreamConfig: &StreamConfig{
			Type:     StreamTypeSSE,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &Config{
			Type:    ProgressTypeSteps,
			Enabled: false,
		},
	}

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:          "step1",
				Description: "Step 1",
			},
		},
	}

	coordinator, err := NewWorkflowWatchCoordinator(watchConfig, wf)
	if err != nil {
		t.Fatalf("NewWorkflowWatchCoordinator() error = %v", err)
	}

	// Test Success
	err = coordinator.Success("Workflow completed successfully")
	if err != nil {
		t.Errorf("Success() error = %v", err)
	}

	// Test Failure
	err = coordinator.Failure("Workflow failed")
	if err != nil {
		t.Errorf("Failure() error = %v", err)
	}
}

func TestWorkflowWatchCoordinator_initializeSteps(t *testing.T) {
	watchConfig := &WatchConfig{
		Enabled:        true,
		StreamConfig:   DefaultStreamConfig(),
		ProgressConfig: &Config{Type: ProgressTypeSteps, Enabled: false},
	}

	tests := []struct {
		name     string
		workflow *workflow.Workflow
		wantLen  int
	}{
		{
			name: "with steps",
			workflow: &workflow.Workflow{
				Steps: []*workflow.Step{
					{ID: "s1", Description: "Step 1"},
					{ID: "s2", Description: "Step 2"},
				},
			},
			wantLen: 2,
		},
		{
			name:     "nil workflow",
			workflow: nil,
			wantLen:  0,
		},
		{
			name: "nil steps",
			workflow: &workflow.Workflow{
				Steps: nil,
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coordinator, err := NewWorkflowWatchCoordinator(watchConfig, tt.workflow)
			if err != nil {
				t.Fatalf("NewWorkflowWatchCoordinator() error = %v", err)
			}

			if len(coordinator.stepMap) != tt.wantLen {
				t.Errorf("stepMap length = %d, want %d", len(coordinator.stepMap), tt.wantLen)
			}
		})
	}
}

// Mock implementations

type MockExpressionEvaluator struct {
	result bool
	err    error
}

func (m *MockExpressionEvaluator) Evaluate(condition string, data map[string]interface{}) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	return m.result, nil
}

type MockExprEvaluator struct {
	result interface{}
	err    error
}

func (m *MockExprEvaluator) Eval(expr string, env map[string]interface{}) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}
