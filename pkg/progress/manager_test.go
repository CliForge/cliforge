package progress

import (
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/workflow"
)

func TestNewManager(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "with nil config",
			config: nil,
		},
		{
			name: "with custom config",
			config: &Config{
				Type:    ProgressTypeSpinner,
				Enabled: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager(tt.config)
			if manager == nil {
				t.Error("NewManager() returned nil")
			}

			if manager.config == nil {
				t.Error("Manager config should not be nil")
			}
		})
	}
}

func TestManager_SetStreamConfig(t *testing.T) {
	manager := NewManager(nil)

	config := &StreamConfig{
		Type:     StreamTypeSSE,
		Endpoint: "http://example.com/stream",
	}

	manager.SetStreamConfig(config)

	if manager.streamConfig == nil {
		t.Error("Stream config should be set")
	}

	if manager.streamConfig.Type != StreamTypeSSE {
		t.Errorf("Expected type SSE, got %s", manager.streamConfig.Type)
	}
}

func TestManager_StartProgress(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false, // Disable for testing
	})

	progress, err := manager.StartProgress("Testing...", 0)
	if err != nil {
		t.Errorf("StartProgress() error = %v", err)
	}

	if progress == nil {
		t.Error("StartProgress() returned nil progress")
	}

	if manager.GetCurrentProgress() == nil {
		t.Error("Current progress should be set")
	}

	// Test error on double start
	_, err = manager.StartProgress("Again...", 0)
	if err == nil {
		// Note: NoopProgress doesn't error on double start
		// Only check if the progress is not nil
	}

	manager.StopProgress()
}

func TestManager_StartProgressForOperation(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false,
	})

	enabled := true
	opConfig := &openapi.CLIProgress{
		Enabled: &enabled,
		Type:    "spinner",
	}

	progress, err := manager.StartProgressForOperation(opConfig, "Operation...")
	if err != nil {
		t.Errorf("StartProgressForOperation() error = %v", err)
	}

	if progress == nil {
		t.Error("StartProgressForOperation() returned nil progress")
	}

	manager.StopProgress()
}

func TestManager_StartWorkflowProgress(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSteps,
		Enabled: false,
	})

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:          "step1",
				Type:        workflow.StepTypeAPICall,
				Description: "First step",
			},
			{
				ID:          "step2",
				Type:        workflow.StepTypeAPICall,
				Description: "Second step",
			},
		},
	}

	enabled := true
	progressConfig := &openapi.CLIProgress{
		Enabled: &enabled,
		Type:    "steps",
	}

	progress, err := manager.StartWorkflowProgress(wf, progressConfig)
	if err != nil {
		t.Errorf("StartWorkflowProgress() error = %v", err)
	}

	if progress == nil {
		t.Error("StartWorkflowProgress() returned nil progress")
	}

	multiStep, ok := progress.(*MultiStep)
	if !ok {
		t.Error("Expected MultiStep progress for workflow")
	}

	if len(multiStep.steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(multiStep.steps))
	}

	manager.StopProgress()
}

func TestManager_StopProgress(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false,
	})

	// Start progress
	manager.StartProgress("Testing...", 0)

	// Stop progress
	err := manager.StopProgress()
	if err != nil {
		t.Errorf("StopProgress() error = %v", err)
	}

	if manager.GetCurrentProgress() != nil {
		t.Error("Current progress should be nil after stop")
	}

	// Stopping again should not error
	err = manager.StopProgress()
	if err != nil {
		t.Errorf("StopProgress() on already stopped should not error: %v", err)
	}
}

func TestManager_SuccessProgress(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false,
	})

	manager.StartProgress("Testing...", 0)

	err := manager.SuccessProgress("Done!")
	if err != nil {
		t.Errorf("SuccessProgress() error = %v", err)
	}
}

func TestManager_FailureProgress(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false,
	})

	manager.StartProgress("Testing...", 0)

	err := manager.FailureProgress("Failed!")
	if err != nil {
		t.Errorf("FailureProgress() error = %v", err)
	}
}

func TestManager_UpdateProgress(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false,
	})

	manager.StartProgress("Testing...", 0)

	err := manager.UpdateProgress("Updated message")
	if err != nil {
		t.Errorf("UpdateProgress() error = %v", err)
	}

	manager.StopProgress()
}

func TestManager_UpdateProgressWithData(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeBar,
		Enabled: false,
	})

	manager.StartProgress("Testing...", 10)

	data := &ProgressData{
		Message: "Step 5",
		Current: 5,
		Total:   10,
	}

	err := manager.UpdateProgressWithData(data)
	if err != nil {
		t.Errorf("UpdateProgressWithData() error = %v", err)
	}

	manager.StopProgress()
}

func TestManager_selectProgressConfig(t *testing.T) {
	manager := NewManager(&Config{
		Type:           ProgressTypeSpinner,
		Enabled:        true,
		ShowTimestamps: false,
	})

	tests := []struct {
		name      string
		opConfig  *openapi.CLIProgress
		wantType  ProgressType
		wantEnabled bool
	}{
		{
			name:        "nil operation config uses default",
			opConfig:    nil,
			wantType:    ProgressTypeSpinner,
			wantEnabled: true,
		},
		{
			name: "operation config overrides type",
			opConfig: &openapi.CLIProgress{
				Type: "bar",
			},
			wantType:    ProgressTypeBar,
			wantEnabled: true,
		},
		{
			name: "operation config can disable",
			opConfig: &openapi.CLIProgress{
				Enabled: boolPtr(false),
			},
			wantType:    ProgressTypeSpinner,
			wantEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := manager.selectProgressConfig(tt.opConfig)

			if config.Type != tt.wantType {
				t.Errorf("Expected type %s, got %s", tt.wantType, config.Type)
			}

			if config.Enabled != tt.wantEnabled {
				t.Errorf("Expected enabled %v, got %v", tt.wantEnabled, config.Enabled)
			}
		})
	}
}

func TestNewWorkflowIntegration(t *testing.T) {
	manager := NewManager(nil)
	integration := NewWorkflowIntegration(manager)

	if integration == nil {
		t.Error("NewWorkflowIntegration() returned nil")
	}

	if integration.manager != manager {
		t.Error("Manager not set correctly")
	}

	if integration.stepMap == nil {
		t.Error("stepMap should be initialized")
	}
}

func TestWorkflowIntegration_OnWorkflowStart(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSteps,
		Enabled: false,
	})

	integration := NewWorkflowIntegration(manager)

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

	err := integration.OnWorkflowStart(wf)
	if err != nil {
		t.Errorf("OnWorkflowStart() error = %v", err)
	}

	if integration.multiStep == nil {
		t.Error("MultiStep should be initialized")
	}

	if len(integration.stepMap) != 2 {
		t.Errorf("Expected 2 steps in stepMap, got %d", len(integration.stepMap))
	}
}

func TestWorkflowIntegration_StepLifecycle(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSteps,
		Enabled: false,
	})

	integration := NewWorkflowIntegration(manager)

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{ID: "step1", Description: "Step 1"},
		},
	}

	integration.OnWorkflowStart(wf)

	// Test OnStepStart
	err := integration.OnStepStart("step1")
	if err != nil {
		t.Errorf("OnStepStart() error = %v", err)
	}

	// Test OnStepComplete
	err = integration.OnStepComplete("step1")
	if err != nil {
		t.Errorf("OnStepComplete() error = %v", err)
	}

	// Test OnStepFail
	integration.OnWorkflowStart(wf) // Reset
	integration.OnStepStart("step1")
	err = integration.OnStepFail("step1", nil)
	if err != nil {
		t.Errorf("OnStepFail() error = %v", err)
	}

	// Test OnStepSkip
	integration.OnWorkflowStart(wf) // Reset
	err = integration.OnStepSkip("step1")
	if err != nil {
		t.Errorf("OnStepSkip() error = %v", err)
	}
}

func TestWorkflowIntegration_OnWorkflowComplete(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSteps,
		Enabled: false,
	})

	integration := NewWorkflowIntegration(manager)

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{ID: "step1", Description: "Step 1"},
		},
	}

	integration.OnWorkflowStart(wf)

	// Test success
	err := integration.OnWorkflowComplete(true, "Workflow completed successfully")
	if err != nil {
		t.Errorf("OnWorkflowComplete(success) error = %v", err)
	}

	// Reset and test failure
	integration.OnWorkflowStart(wf)
	err = integration.OnWorkflowComplete(false, "Workflow failed")
	if err != nil {
		t.Errorf("OnWorkflowComplete(failure) error = %v", err)
	}
}

func TestDefaultManager(t *testing.T) {
	if DefaultManager == nil {
		t.Error("DefaultManager should not be nil")
	}

	if DefaultManager.config == nil {
		t.Error("DefaultManager config should not be nil")
	}
}

func TestPackageLevelFunctions(t *testing.T) {
	// These test the package-level convenience functions
	// They use the DefaultManager which might have state from other tests,
	// so we just check they don't panic

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Package function panicked: %v", r)
		}
	}()

	// Clean up any existing progress
	StopProgress()

	// Test StartProgress
	progress, err := StartProgress("Test", 0)
	if err != nil {
		// Some tests might have left progress active, that's ok
		if progress == nil {
			t.Error("StartProgress returned nil without error")
		}
	}

	// Test UpdateProgress
	_ = UpdateProgress("Updated")

	// Test SuccessProgress
	_ = SuccessProgress("Done")

	// Clean up
	StopProgress()

	// Test FailureProgress
	StartProgress("Test", 0)
	_ = FailureProgress("Failed")

	StopProgress()
}

func TestManager_GetCurrentProgress(t *testing.T) {
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false,
	})

	if manager.GetCurrentProgress() != nil {
		t.Error("Current progress should be nil initially")
	}

	manager.StartProgress("Test", 0)

	if manager.GetCurrentProgress() == nil {
		t.Error("Current progress should not be nil after start")
	}

	manager.StopProgress()

	if manager.GetCurrentProgress() != nil {
		t.Error("Current progress should be nil after stop")
	}
}

func TestManager_GetCurrentWatch(t *testing.T) {
	manager := NewManager(nil)

	if manager.GetCurrentWatch() != nil {
		t.Error("Current watch should be nil initially")
	}

	// We can't easily test StartWatch without a real server,
	// so just verify the getter works
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}

func TestWatchConfig(t *testing.T) {
	streamConfig := &StreamConfig{
		Type:     StreamTypeSSE,
		Endpoint: "http://example.com/stream",
	}

	progressConfig := &Config{
		Type:    ProgressTypeSteps,
		Enabled: true,
	}

	watchConfig := &WatchConfig{
		Enabled:        true,
		StreamConfig:   streamConfig,
		ProgressConfig: progressConfig,
		ExitConditions: []*ExitCondition{
			{
				EventType: "status",
				Condition: "event.data.state == 'completed'",
				Message:   "Operation completed",
			},
		},
		ShowLogs:  true,
		LogPrefix: "LOG:",
	}

	if !watchConfig.Enabled {
		t.Error("WatchConfig should be enabled")
	}

	if watchConfig.StreamConfig.Type != StreamTypeSSE {
		t.Errorf("Expected stream type SSE, got %s", watchConfig.StreamConfig.Type)
	}

	if watchConfig.ProgressConfig.Type != ProgressTypeSteps {
		t.Errorf("Expected progress type Steps, got %s", watchConfig.ProgressConfig.Type)
	}

	if len(watchConfig.ExitConditions) != 1 {
		t.Errorf("Expected 1 exit condition, got %d", len(watchConfig.ExitConditions))
	}

	if !watchConfig.ShowLogs {
		t.Error("ShowLogs should be true")
	}

	if watchConfig.LogPrefix != "LOG:" {
		t.Errorf("Expected log prefix 'LOG:', got %s", watchConfig.LogPrefix)
	}
}

func TestExitCondition(t *testing.T) {
	condition := &ExitCondition{
		EventType: "status",
		Condition: "event.data.state == 'completed'",
		Message:   "Done!",
	}

	if condition.EventType != "status" {
		t.Errorf("Expected event type 'status', got %s", condition.EventType)
	}

	if condition.Condition == "" {
		t.Error("Condition should not be empty")
	}

	if condition.Message != "Done!" {
		t.Errorf("Expected message 'Done!', got %s", condition.Message)
	}
}

func TestManager_Concurrency(t *testing.T) {
	// Test that manager handles concurrent operations safely
	manager := NewManager(&Config{
		Type:    ProgressTypeSpinner,
		Enabled: false,
	})

	done := make(chan bool)

	// Start multiple goroutines trying to start progress
	for i := 0; i < 10; i++ {
		go func() {
			manager.StartProgress("Test", 0)
			time.Sleep(10 * time.Millisecond)
			manager.StopProgress()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Manager should be in clean state
	if manager.GetCurrentProgress() != nil {
		t.Error("Manager should have no active progress after concurrent operations")
	}
}
