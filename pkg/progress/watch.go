package progress

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/CliForge/cliforge/pkg/workflow"
	"github.com/pterm/pterm"
)

// Watch implements the watch mode coordinator.
type Watch struct {
	config       *WatchConfig
	streamClient StreamClient
	progress     Progress
	running      bool
	cancel       context.CancelFunc
	mu           sync.RWMutex
	exprEval     ExpressionEvaluator
}

// ExpressionEvaluator evaluates exit conditions.
type ExpressionEvaluator interface {
	Evaluate(condition string, data map[string]interface{}) (bool, error)
}

// NewWatch creates a new watch coordinator.
func NewWatch(config *WatchConfig) (*Watch, error) {
	if config == nil {
		return nil, fmt.Errorf("watch config is required")
	}

	// Create stream client
	streamClient := NewStreamClient(config.StreamConfig)

	// Create progress indicator
	progress := New(config.ProgressConfig, 0)

	return &Watch{
		config:       config,
		streamClient: streamClient,
		progress:     progress,
	}, nil
}

// SetExpressionEvaluator sets the expression evaluator for exit conditions.
func (w *Watch) SetExpressionEvaluator(eval ExpressionEvaluator) {
	w.exprEval = eval
}

// Start starts watch mode.
func (w *Watch) Start(ctx context.Context) error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return fmt.Errorf("watch mode already running")
	}
	w.mu.Unlock()

	if !w.config.Enabled {
		return fmt.Errorf("watch mode is not enabled")
	}

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start progress indicator
	if err := w.progress.Start("Watching for updates..."); err != nil {
		return fmt.Errorf("failed to start progress: %w", err)
	}

	// Connect to stream
	if err := w.streamClient.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect stream: %w", err)
	}

	// Subscribe to configured events
	for _, eventType := range w.config.StreamConfig.Events {
		if err := w.streamClient.Subscribe(eventType, w.handleEvent); err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", eventType, err)
		}
	}

	// Process events
	go w.processEvents(ctx)

	// Wait for completion or interrupt
	select {
	case <-ctx.Done():
		w.cleanup()
		return ctx.Err()
	case <-sigChan:
		pterm.Println("\nReceived interrupt signal, shutting down gracefully...")
		w.cleanup()
		return nil
	}
}

// processEvents processes incoming events.
func (w *Watch) processEvents(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return

		case event := <-w.streamClient.Events():
			if event == nil {
				continue
			}

			// Handle the event
			if err := w.handleEvent(event); err != nil {
				pterm.Error.Printf("Error handling event: %v\n", err)
			}

		case err := <-w.streamClient.Errors():
			if err == nil {
				continue
			}

			pterm.Error.Printf("Stream error: %v\n", err)

			// Check if error is fatal
			if w.isFatalError(err) {
				w.Stop()
				return
			}
		}
	}
}

// handleEvent handles a single event.
func (w *Watch) handleEvent(event *Event) error {
	// Display log events
	if w.config.ShowLogs && (event.Type == "log" || event.Type == "message") {
		w.displayLog(event)
	}

	// Update progress for status events
	if event.Type == "status" {
		w.updateProgress(event)
	}

	// Handle error events
	if event.Type == "error" {
		w.displayError(event)
	}

	// Check exit conditions
	if w.shouldExit(event) {
		// Found exit condition, stop watching
		if w.cancel != nil {
			w.cancel()
		}
	}

	return nil
}

// displayLog displays a log event.
func (w *Watch) displayLog(event *Event) {
	prefix := w.config.LogPrefix
	if prefix == "" {
		prefix = time.Now().Format("2006-01-02 15:04:05")
	}

	pterm.Printf("%s %s\n", pterm.Gray(prefix), event.Data)
}

// updateProgress updates the progress indicator with status information.
func (w *Watch) updateProgress(event *Event) {
	if event.Data != "" {
		w.progress.Update(event.Data)
	}
}

// displayError displays an error event.
func (w *Watch) displayError(event *Event) {
	pterm.Error.Printf("Error: %s\n", event.Data)
}

// shouldExit checks if an event matches an exit condition.
func (w *Watch) shouldExit(event *Event) bool {
	if len(w.config.ExitConditions) == 0 {
		return false
	}

	for _, condition := range w.config.ExitConditions {
		// Check if event type matches
		if condition.EventType != "" && condition.EventType != event.Type {
			continue
		}

		// Evaluate condition if present
		if condition.Condition != "" && w.exprEval != nil {
			data := map[string]interface{}{
				"event": map[string]interface{}{
					"type": event.Type,
					"data": event.Data,
					"id":   event.ID,
				},
			}

			matched, err := w.exprEval.Evaluate(condition.Condition, data)
			if err != nil {
				pterm.Warning.Printf("Failed to evaluate exit condition: %v\n", err)
				continue
			}

			if matched {
				if condition.Message != "" {
					pterm.Success.Println(condition.Message)
				}
				return true
			}
		} else if condition.Condition == "" {
			// No condition, just event type match
			if condition.Message != "" {
				pterm.Success.Println(condition.Message)
			}
			return true
		}
	}

	return false
}

// isFatalError determines if an error should terminate watch mode.
func (w *Watch) isFatalError(err error) bool {
	// Add logic to determine if error is fatal
	// For now, no errors are considered fatal (will retry)
	return false
}

// Stop stops watch mode.
func (w *Watch) Stop() error {
	if w.cancel != nil {
		w.cancel()
	}

	w.cleanup()

	return nil
}

// cleanup performs cleanup operations.
func (w *Watch) cleanup() {
	w.mu.Lock()
	defer w.mu.Unlock()

	if !w.running {
		return
	}

	// Close stream connection
	if w.streamClient != nil {
		w.streamClient.Close()
	}

	// Stop progress indicator
	if w.progress != nil {
		w.progress.Stop()
	}

	w.running = false
}

// IsRunning returns true if watch mode is active.
func (w *Watch) IsRunning() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.running
}

// SimpleExpressionEvaluator is a basic expression evaluator.
type SimpleExpressionEvaluator struct {
	evaluator ExprEvaluator
}

// ExprEvaluator is an interface for evaluating expressions.
type ExprEvaluator interface {
	Eval(expr string, env map[string]interface{}) (interface{}, error)
}

// NewSimpleExpressionEvaluator creates a new simple expression evaluator.
func NewSimpleExpressionEvaluator(evaluator ExprEvaluator) *SimpleExpressionEvaluator {
	return &SimpleExpressionEvaluator{
		evaluator: evaluator,
	}
}

// Evaluate evaluates an expression and returns a boolean result.
func (s *SimpleExpressionEvaluator) Evaluate(condition string, data map[string]interface{}) (bool, error) {
	if s.evaluator == nil {
		return false, fmt.Errorf("no evaluator configured")
	}

	result, err := s.evaluator.Eval(condition, data)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	// Convert result to boolean
	boolResult, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expression did not return a boolean value")
	}

	return boolResult, nil
}

// WorkflowWatchCoordinator coordinates watch mode for workflow execution.
type WorkflowWatchCoordinator struct {
	watch      *Watch
	workflow   *workflow.Workflow
	multiStep  *MultiStep
	stepMap    map[string]*StepInfo
	mu         sync.RWMutex
}

// NewWorkflowWatchCoordinator creates a coordinator for workflow watch mode.
func NewWorkflowWatchCoordinator(watchConfig *WatchConfig, wf *workflow.Workflow) (*WorkflowWatchCoordinator, error) {
	// Create multi-step progress display
	progressConfig := watchConfig.ProgressConfig
	if progressConfig == nil {
		progressConfig = DefaultConfig()
	}
	progressConfig.Type = ProgressTypeSteps

	multiStep := NewMultiStep(progressConfig)

	// Update watch config to use multi-step
	watchConfig.ProgressConfig = progressConfig

	watch, err := NewWatch(watchConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create watch: %w", err)
	}

	// Replace progress with multi-step
	watch.progress = multiStep

	coordinator := &WorkflowWatchCoordinator{
		watch:     watch,
		workflow:  wf,
		multiStep: multiStep,
		stepMap:   make(map[string]*StepInfo),
	}

	// Initialize steps from workflow
	coordinator.initializeSteps()

	return coordinator, nil
}

// initializeSteps initializes step tracking from the workflow.
func (w *WorkflowWatchCoordinator) initializeSteps() {
	if w.workflow == nil || w.workflow.Steps == nil {
		return
	}

	for _, step := range w.workflow.Steps {
		stepInfo := &StepInfo{
			ID:          step.ID,
			Description: step.Description,
			Status:      StepStatusPending,
			Metadata:    make(map[string]interface{}),
		}

		w.stepMap[step.ID] = stepInfo
		w.multiStep.AddStep(stepInfo)
	}
}

// UpdateStepStatus updates the status of a workflow step.
func (w *WorkflowWatchCoordinator) UpdateStepStatus(stepID string, status StepStatus, message string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	step, exists := w.stepMap[stepID]
	if !exists {
		return fmt.Errorf("step %s not found", stepID)
	}

	step.Status = status
	if message != "" {
		step.Description = message
	}

	return w.multiStep.UpdateStep(stepID, status, step.Description)
}

// AddSubStep adds a sub-step to a parent step.
func (w *WorkflowWatchCoordinator) AddSubStep(parentID string, substep *StepInfo) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	parent, exists := w.stepMap[parentID]
	if !exists {
		return fmt.Errorf("parent step %s not found", parentID)
	}

	parent.SubSteps = append(parent.SubSteps, substep)
	w.stepMap[substep.ID] = substep

	// Re-render
	return w.multiStep.UpdateStep(parentID, parent.Status, parent.Description)
}

// Start starts the workflow watch coordinator.
func (w *WorkflowWatchCoordinator) Start(ctx context.Context) error {
	return w.watch.Start(ctx)
}

// Stop stops the workflow watch coordinator.
func (w *WorkflowWatchCoordinator) Stop() error {
	return w.watch.Stop()
}

// IsRunning returns true if the coordinator is running.
func (w *WorkflowWatchCoordinator) IsRunning() bool {
	return w.watch.IsRunning()
}

// Success marks the workflow as successful.
func (w *WorkflowWatchCoordinator) Success(message string) error {
	return w.multiStep.Success(message)
}

// Failure marks the workflow as failed.
func (w *WorkflowWatchCoordinator) Failure(message string) error {
	return w.multiStep.Failure(message)
}
