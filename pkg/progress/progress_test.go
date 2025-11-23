package progress

import (
	"bytes"
	"testing"
	"time"
)

func TestNewSpinner(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
		want   bool
	}{
		{
			name:   "with nil config",
			config: nil,
			want:   true,
		},
		{
			name: "with custom config",
			config: &Config{
				Type:    ProgressTypeSpinner,
				Enabled: true,
			},
			want: true,
		},
		{
			name: "with disabled config",
			config: &Config{
				Type:    ProgressTypeSpinner,
				Enabled: false,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spinner := NewSpinner(tt.config)
			if spinner == nil {
				t.Error("NewSpinner() returned nil")
			}
		})
	}
}

func TestSpinner_StartStop(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeSpinner,
		Enabled: true,
		Writer:  &buf,
	}

	spinner := NewSpinner(config)

	// Test start
	err := spinner.Start("Testing...")
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if !spinner.IsActive() {
		t.Error("Spinner should be active after Start()")
	}

	// Test double start
	err = spinner.Start("Again...")
	if err == nil {
		t.Error("Expected error when starting already active spinner")
	}

	// Test stop
	err = spinner.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if spinner.IsActive() {
		t.Error("Spinner should not be active after Stop()")
	}
}

func TestSpinner_Update(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeSpinner,
		Enabled: true,
		Writer:  &buf,
	}

	spinner := NewSpinner(config)
	spinner.Start("Initial")

	err := spinner.Update("Updated message")
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	spinner.Stop()
}

func TestSpinner_Success(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeSpinner,
		Enabled: true,
		Writer:  &buf,
	}

	spinner := NewSpinner(config)
	spinner.Start("Working...")

	err := spinner.Success("Done!")
	if err != nil {
		t.Errorf("Success() error = %v", err)
	}

	if spinner.IsActive() {
		t.Error("Spinner should not be active after Success()")
	}
}

func TestSpinner_Failure(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeSpinner,
		Enabled: true,
		Writer:  &buf,
	}

	spinner := NewSpinner(config)
	spinner.Start("Working...")

	err := spinner.Failure("Failed!")
	if err != nil {
		t.Errorf("Failure() error = %v", err)
	}

	if spinner.IsActive() {
		t.Error("Spinner should not be active after Failure()")
	}
}

func TestNewProgressBar(t *testing.T) {
	config := &Config{
		Type:    ProgressTypeBar,
		Enabled: true,
	}

	bar := NewProgressBar(config, 10)
	if bar == nil {
		t.Error("NewProgressBar() returned nil")
	}

	if bar.total != 10 {
		t.Errorf("Expected total = 10, got %d", bar.total)
	}
}

func TestProgressBar_StartStop(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeBar,
		Enabled: true,
		Writer:  &buf,
	}

	bar := NewProgressBar(config, 5)

	err := bar.Start("Processing...")
	if err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if !bar.IsActive() {
		t.Error("ProgressBar should be active after Start()")
	}

	err = bar.Stop()
	if err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if bar.IsActive() {
		t.Error("ProgressBar should not be active after Stop()")
	}
}

func TestProgressBar_Increment(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeBar,
		Enabled: true,
		Writer:  &buf,
	}

	bar := NewProgressBar(config, 5)
	bar.Start("Processing...")

	for i := 0; i < 5; i++ {
		err := bar.Increment()
		if err != nil {
			t.Errorf("Increment() error = %v", err)
		}
	}

	bar.Stop()
}

func TestProgressBar_UpdateWithData(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeBar,
		Enabled: true,
		Writer:  &buf,
	}

	bar := NewProgressBar(config, 10)
	bar.Start("Processing...")

	data := &ProgressData{
		Message: "Step 5",
		Current: 5,
		Total:   10,
	}

	err := bar.UpdateWithData(data)
	if err != nil {
		t.Errorf("UpdateWithData() error = %v", err)
	}

	bar.Stop()
}

func TestNewMultiStep(t *testing.T) {
	config := &Config{
		Type:    ProgressTypeSteps,
		Enabled: true,
	}

	ms := NewMultiStep(config)
	if ms == nil {
		t.Error("NewMultiStep() returned nil")
	}

	if ms.steps == nil {
		t.Error("steps map should be initialized")
	}

	if ms.order == nil {
		t.Error("order slice should be initialized")
	}
}

func TestMultiStep_AddStep(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeSteps,
		Enabled: true,
		Writer:  &buf,
	}

	ms := NewMultiStep(config)
	ms.Start("Workflow")

	step := &StepInfo{
		ID:          "step1",
		Description: "First step",
		Status:      StepStatusPending,
	}

	err := ms.AddStep(step)
	if err != nil {
		t.Errorf("AddStep() error = %v", err)
	}

	if len(ms.steps) != 1 {
		t.Errorf("Expected 1 step, got %d", len(ms.steps))
	}

	if len(ms.order) != 1 {
		t.Errorf("Expected 1 step in order, got %d", len(ms.order))
	}

	ms.Stop()
}

func TestMultiStep_UpdateStep(t *testing.T) {
	var buf bytes.Buffer
	config := &Config{
		Type:    ProgressTypeSteps,
		Enabled: true,
		Writer:  &buf,
	}

	ms := NewMultiStep(config)
	ms.Start("Workflow")

	step := &StepInfo{
		ID:          "step1",
		Description: "First step",
		Status:      StepStatusPending,
	}

	ms.AddStep(step)

	err := ms.UpdateStep("step1", StepStatusRunning, "Running...")
	if err != nil {
		t.Errorf("UpdateStep() error = %v", err)
	}

	if ms.steps["step1"].Status != StepStatusRunning {
		t.Errorf("Expected status Running, got %s", ms.steps["step1"].Status)
	}

	err = ms.UpdateStep("nonexistent", StepStatusRunning, "")
	if err == nil {
		t.Error("Expected error for non-existent step")
	}

	ms.Stop()
}

func TestMultiStep_StatusIcons(t *testing.T) {
	config := &Config{
		Type:    ProgressTypeSteps,
		Enabled: true,
	}

	ms := NewMultiStep(config)

	tests := []struct {
		status StepStatus
		want   string
	}{
		{StepStatusPending, "☐"},
		{StepStatusRunning, "⧗"},
		{StepStatusCompleted, "✓"},
		{StepStatusFailed, "✗"},
		{StepStatusSkipped, "○"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			icon := ms.getStatusIcon(tt.status)
			// Just check that we get a non-empty string
			if icon == "" {
				t.Errorf("Expected non-empty icon for status %s", tt.status)
			}
		})
	}
}

func TestNoopProgress(t *testing.T) {
	noop := NewNoopProgress()

	// All operations should succeed and do nothing
	if err := noop.Start("test"); err != nil {
		t.Errorf("Start() error = %v", err)
	}

	if err := noop.Update("test"); err != nil {
		t.Errorf("Update() error = %v", err)
	}

	if err := noop.UpdateWithData(&ProgressData{}); err != nil {
		t.Errorf("UpdateWithData() error = %v", err)
	}

	if err := noop.Success("test"); err != nil {
		t.Errorf("Success() error = %v", err)
	}

	if err := noop.Failure("test"); err != nil {
		t.Errorf("Failure() error = %v", err)
	}

	if err := noop.Stop(); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	if noop.IsActive() {
		t.Error("NoopProgress should never be active")
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		total       int
		wantType    string
	}{
		{
			name: "spinner type",
			config: &Config{
				Type:    ProgressTypeSpinner,
				Enabled: true,
			},
			total:    0,
			wantType: "*progress.Spinner",
		},
		{
			name: "bar type",
			config: &Config{
				Type:    ProgressTypeBar,
				Enabled: true,
			},
			total:    10,
			wantType: "*progress.ProgressBar",
		},
		{
			name: "steps type",
			config: &Config{
				Type:    ProgressTypeSteps,
				Enabled: true,
			},
			total:    0,
			wantType: "*progress.MultiStep",
		},
		{
			name: "none type",
			config: &Config{
				Type:    ProgressTypeNone,
				Enabled: true,
			},
			total:    0,
			wantType: "*progress.NoopProgress",
		},
		{
			name: "disabled",
			config: &Config{
				Type:    ProgressTypeSpinner,
				Enabled: false,
			},
			total:    0,
			wantType: "*progress.NoopProgress",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.config, tt.total)
			if got == nil {
				t.Error("New() returned nil")
			}
		})
	}
}

func TestProgressData(t *testing.T) {
	now := time.Now()
	data := &ProgressData{
		Message:    "Test message",
		Current:    5,
		Total:      10,
		Percentage: 50.0,
		Timestamp:  now,
		Metadata: map[string]interface{}{
			"key": "value",
		},
	}

	if data.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got %s", data.Message)
	}

	if data.Current != 5 {
		t.Errorf("Expected current 5, got %d", data.Current)
	}

	if data.Total != 10 {
		t.Errorf("Expected total 10, got %d", data.Total)
	}

	if data.Percentage != 50.0 {
		t.Errorf("Expected percentage 50.0, got %f", data.Percentage)
	}

	if !data.Timestamp.Equal(now) {
		t.Error("Timestamp mismatch")
	}

	if data.Metadata["key"] != "value" {
		t.Error("Metadata not set correctly")
	}
}

func TestStepInfo(t *testing.T) {
	step := &StepInfo{
		ID:          "test-step",
		Description: "Test step",
		Status:      StepStatusPending,
		Metadata:    make(map[string]interface{}),
	}

	if step.ID != "test-step" {
		t.Errorf("Expected ID 'test-step', got %s", step.ID)
	}

	if step.Status != StepStatusPending {
		t.Errorf("Expected status Pending, got %s", step.Status)
	}

	// Test with substeps
	substep := &StepInfo{
		ID:          "substep-1",
		Description: "Substep",
		Status:      StepStatusPending,
	}

	step.SubSteps = append(step.SubSteps, substep)

	if len(step.SubSteps) != 1 {
		t.Errorf("Expected 1 substep, got %d", len(step.SubSteps))
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig() returned nil")
	}

	if config.Type != ProgressTypeSpinner {
		t.Errorf("Expected default type Spinner, got %s", config.Type)
	}

	if !config.Enabled {
		t.Error("Expected enabled to be true")
	}

	if config.ShowTimestamps {
		t.Error("Expected ShowTimestamps to be false by default")
	}

	if !config.ShowStepDescriptions {
		t.Error("Expected ShowStepDescriptions to be true by default")
	}

	if config.Color != "auto" {
		t.Errorf("Expected color 'auto', got %s", config.Color)
	}
}
