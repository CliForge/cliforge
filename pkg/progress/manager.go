package progress

import (
	"context"
	"fmt"
	"sync"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/workflow"
)

// Manager manages progress indicators and streaming for operations.
type Manager struct {
	config            *Config
	streamConfig      *StreamConfig
	currentProgress   Progress
	currentWatch      WatchCoordinator
	workflowIntegration *WorkflowIntegration
	mu                sync.RWMutex
}

// NewManager creates a new progress manager.
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	return &Manager{
		config: config,
	}
}

// SetStreamConfig sets the streaming configuration.
func (m *Manager) SetStreamConfig(config *StreamConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.streamConfig = config
}

// SetWorkflowIntegration sets the workflow integration.
func (m *Manager) SetWorkflowIntegration(integration *WorkflowIntegration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.workflowIntegration = integration
}

// StartProgress starts a progress indicator based on configuration.
func (m *Manager) StartProgress(message string, total int) (Progress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentProgress != nil && m.currentProgress.IsActive() {
		return nil, fmt.Errorf("progress already active")
	}

	progress := New(m.config, total)
	if err := progress.Start(message); err != nil {
		return nil, fmt.Errorf("failed to start progress: %w", err)
	}

	m.currentProgress = progress
	return progress, nil
}

// StartProgressForOperation starts progress for an OpenAPI operation.
func (m *Manager) StartProgressForOperation(opConfig *openapi.CLIProgress, message string) (Progress, error) {
	config := m.selectProgressConfig(opConfig)

	m.mu.Lock()
	defer m.mu.Unlock()

	progress := New(config, 0)
	if err := progress.Start(message); err != nil {
		return nil, fmt.Errorf("failed to start progress: %w", err)
	}

	m.currentProgress = progress
	return progress, nil
}

// StartWorkflowProgress starts progress tracking for a workflow.
func (m *Manager) StartWorkflowProgress(wf *workflow.Workflow, progressConfig *openapi.CLIProgress) (Progress, error) {
	config := m.selectProgressConfig(progressConfig)

	// Use multi-step for workflows
	config.Type = ProgressTypeSteps

	m.mu.Lock()
	defer m.mu.Unlock()

	multiStep := NewMultiStep(config)

	// Initialize steps from workflow
	if wf != nil && wf.Steps != nil {
		for _, step := range wf.Steps {
			stepInfo := &StepInfo{
				ID:          step.ID,
				Description: step.Description,
				Status:      StepStatusPending,
			}
			multiStep.AddStep(stepInfo)
		}
	}

	if err := multiStep.Start("Executing workflow..."); err != nil {
		return nil, fmt.Errorf("failed to start workflow progress: %w", err)
	}

	m.currentProgress = multiStep
	return multiStep, nil
}

// StartWatch starts watch mode for streaming updates.
func (m *Manager) StartWatch(ctx context.Context, watchConfig *WatchConfig) (WatchCoordinator, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentWatch != nil && m.currentWatch.IsRunning() {
		return nil, fmt.Errorf("watch mode already active")
	}

	watch, err := NewWatch(watchConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create watch: %w", err)
	}

	// Start watch in background
	go func() {
		if err := watch.Start(ctx); err != nil && err != context.Canceled {
			fmt.Printf("Watch error: %v\n", err)
		}
	}()

	m.currentWatch = watch
	return watch, nil
}

// StartWorkflowWatch starts watch mode for workflow execution.
func (m *Manager) StartWorkflowWatch(ctx context.Context, wf *workflow.Workflow, watchConfig *WatchConfig) (WatchCoordinator, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentWatch != nil && m.currentWatch.IsRunning() {
		return nil, fmt.Errorf("watch mode already active")
	}

	coordinator, err := NewWorkflowWatchCoordinator(watchConfig, wf)
	if err != nil {
		return nil, fmt.Errorf("failed to create workflow watch: %w", err)
	}

	// Start watch in background
	go func() {
		if err := coordinator.Start(ctx); err != nil && err != context.Canceled {
			fmt.Printf("Watch error: %v\n", err)
		}
	}()

	m.currentWatch = coordinator
	return coordinator, nil
}

// GetCurrentProgress returns the currently active progress indicator.
func (m *Manager) GetCurrentProgress() Progress {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentProgress
}

// GetCurrentWatch returns the currently active watch coordinator.
func (m *Manager) GetCurrentWatch() WatchCoordinator {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentWatch
}

// StopProgress stops the current progress indicator.
func (m *Manager) StopProgress() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentProgress == nil {
		return nil
	}

	if err := m.currentProgress.Stop(); err != nil {
		return fmt.Errorf("failed to stop progress: %w", err)
	}

	m.currentProgress = nil
	return nil
}

// StopWatch stops the current watch coordinator.
func (m *Manager) StopWatch() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentWatch == nil {
		return nil
	}

	if err := m.currentWatch.Stop(); err != nil {
		return fmt.Errorf("failed to stop watch: %w", err)
	}

	m.currentWatch = nil
	return nil
}

// SuccessProgress marks the current progress as successful.
func (m *Manager) SuccessProgress(message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentProgress == nil {
		return nil
	}

	if err := m.currentProgress.Success(message); err != nil {
		return fmt.Errorf("failed to mark progress as success: %w", err)
	}

	return nil
}

// FailureProgress marks the current progress as failed.
func (m *Manager) FailureProgress(message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.currentProgress == nil {
		return nil
	}

	if err := m.currentProgress.Failure(message); err != nil {
		return fmt.Errorf("failed to mark progress as failed: %w", err)
	}

	return nil
}

// UpdateProgress updates the current progress.
func (m *Manager) UpdateProgress(message string) error {
	m.mu.RLock()
	progress := m.currentProgress
	m.mu.RUnlock()

	if progress == nil {
		return nil
	}

	return progress.Update(message)
}

// UpdateProgressWithData updates the current progress with structured data.
func (m *Manager) UpdateProgressWithData(data *ProgressData) error {
	m.mu.RLock()
	progress := m.currentProgress
	m.mu.RUnlock()

	if progress == nil {
		return nil
	}

	return progress.UpdateWithData(data)
}

// selectProgressConfig selects the appropriate progress configuration.
func (m *Manager) selectProgressConfig(opConfig *openapi.CLIProgress) *Config {
	config := &Config{
		Type:                 m.config.Type,
		Enabled:              m.config.Enabled,
		ShowTimestamps:       m.config.ShowTimestamps,
		ShowStepDescriptions: m.config.ShowStepDescriptions,
		Color:                m.config.Color,
		Writer:               m.config.Writer,
		RefreshRate:          m.config.RefreshRate,
	}

	// Apply operation-specific config if present
	if opConfig != nil {
		if opConfig.Enabled != nil {
			config.Enabled = *opConfig.Enabled
		}
		if opConfig.Type != "" {
			config.Type = ProgressType(opConfig.Type)
		}
		if opConfig.ShowTimestamps != nil {
			config.ShowTimestamps = *opConfig.ShowTimestamps
		}
		if opConfig.ShowStepDescriptions != nil {
			config.ShowStepDescriptions = *opConfig.ShowStepDescriptions
		}
		if opConfig.Color != "" {
			config.Color = opConfig.Color
		}
	}

	return config
}

// WorkflowIntegration integrates progress with workflow execution.
type WorkflowIntegration struct {
	manager   *Manager
	multiStep *MultiStep
	stepMap   map[string]*StepInfo
	mu        sync.RWMutex
}

// NewWorkflowIntegration creates a new workflow integration.
func NewWorkflowIntegration(manager *Manager) *WorkflowIntegration {
	return &WorkflowIntegration{
		manager: manager,
		stepMap: make(map[string]*StepInfo),
	}
}

// OnWorkflowStart is called when a workflow starts.
func (w *WorkflowIntegration) OnWorkflowStart(wf *workflow.Workflow) error {
	config := w.manager.config
	config.Type = ProgressTypeSteps

	multiStep := NewMultiStep(config)

	// Initialize steps
	for _, step := range wf.Steps {
		stepInfo := &StepInfo{
			ID:          step.ID,
			Description: step.Description,
			Status:      StepStatusPending,
		}
		w.stepMap[step.ID] = stepInfo
		multiStep.AddStep(stepInfo)
	}

	if err := multiStep.Start("Executing workflow..."); err != nil {
		return err
	}

	w.mu.Lock()
	w.multiStep = multiStep
	w.mu.Unlock()

	return nil
}

// OnStepStart is called when a step starts.
func (w *WorkflowIntegration) OnStepStart(stepID string) error {
	w.mu.RLock()
	multiStep := w.multiStep
	w.mu.RUnlock()

	if multiStep == nil {
		return nil
	}

	return multiStep.UpdateStep(stepID, StepStatusRunning, "")
}

// OnStepComplete is called when a step completes successfully.
func (w *WorkflowIntegration) OnStepComplete(stepID string) error {
	w.mu.RLock()
	multiStep := w.multiStep
	w.mu.RUnlock()

	if multiStep == nil {
		return nil
	}

	return multiStep.UpdateStep(stepID, StepStatusCompleted, "")
}

// OnStepFail is called when a step fails.
func (w *WorkflowIntegration) OnStepFail(stepID string, err error) error {
	w.mu.RLock()
	multiStep := w.multiStep
	w.mu.RUnlock()

	if multiStep == nil {
		return nil
	}

	message := ""
	if err != nil {
		message = err.Error()
	}

	return multiStep.UpdateStep(stepID, StepStatusFailed, message)
}

// OnStepSkip is called when a step is skipped.
func (w *WorkflowIntegration) OnStepSkip(stepID string) error {
	w.mu.RLock()
	multiStep := w.multiStep
	w.mu.RUnlock()

	if multiStep == nil {
		return nil
	}

	return multiStep.UpdateStep(stepID, StepStatusSkipped, "")
}

// OnWorkflowComplete is called when the workflow completes.
func (w *WorkflowIntegration) OnWorkflowComplete(success bool, message string) error {
	w.mu.RLock()
	multiStep := w.multiStep
	w.mu.RUnlock()

	if multiStep == nil {
		return nil
	}

	if success {
		return multiStep.Success(message)
	}

	return multiStep.Failure(message)
}

// DefaultManager is the global default progress manager.
var DefaultManager = NewManager(DefaultConfig())

// Package-level convenience functions

// StartProgress starts a progress indicator using the default manager.
func StartProgress(message string, total int) (Progress, error) {
	return DefaultManager.StartProgress(message, total)
}

// StopProgress stops the current progress using the default manager.
func StopProgress() error {
	return DefaultManager.StopProgress()
}

// UpdateProgress updates the current progress using the default manager.
func UpdateProgress(message string) error {
	return DefaultManager.UpdateProgress(message)
}

// SuccessProgress marks progress as successful using the default manager.
func SuccessProgress(message string) error {
	return DefaultManager.SuccessProgress(message)
}

// FailureProgress marks progress as failed using the default manager.
func FailureProgress(message string) error {
	return DefaultManager.FailureProgress(message)
}
