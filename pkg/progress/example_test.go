package progress_test

import (
	"context"
	"fmt"
	"time"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/progress"
	"github.com/CliForge/cliforge/pkg/workflow"
)

// Example_spinner demonstrates using a spinner for a single operation.
func Example_spinner() {
	// Create a spinner config
	config := &progress.Config{
		Type:    progress.TypeSpinner,
		Enabled: false, // Set to false for example to avoid terminal output
	}

	// Create a spinner
	spinner := progress.NewSpinner(config)

	// Start the spinner
	_ = spinner.Start("Processing...")

	// Simulate work
	time.Sleep(100 * time.Millisecond)

	// Update the message
	_ = spinner.Update("Almost done...")

	// Mark as successful
	_ = spinner.Success("Completed successfully!")

	fmt.Println("Operation completed")
	// Output: Operation completed
}

// Example_progressBar demonstrates using a progress bar for known steps.
func Example_progressBar() {
	config := &progress.Config{
		Type:    progress.TypeBar,
		Enabled: false,
	}

	bar := progress.NewProgressBar(config, 5)

	_ = bar.Start("Processing items...")

	for i := 0; i < 5; i++ {
		time.Sleep(50 * time.Millisecond)
		_ = bar.Increment()
	}

	_ = bar.Success("All items processed!")

	fmt.Println("Processing complete")
	// Output: Processing complete
}

// Example_multiStep demonstrates using multi-step progress for workflows.
func Example_multiStep() {
	config := &progress.Config{
		Type:    progress.TypeSteps,
		Enabled: false,
	}

	multiStep := progress.NewMultiStep(config)

	_ = multiStep.Start("Deploying application...")

	// Add steps
	steps := []*progress.StepInfo{
		{
			ID:          "validate",
			Description: "Validating configuration",
			Status:      progress.StepStatusPending,
		},
		{
			ID:          "build",
			Description: "Building application",
			Status:      progress.StepStatusPending,
		},
		{
			ID:          "deploy",
			Description: "Deploying to production",
			Status:      progress.StepStatusPending,
		},
	}

	for _, step := range steps {
		_ = multiStep.AddStep(step)
	}

	// Execute steps
	for _, step := range steps {
		_ = multiStep.UpdateStep(step.ID, progress.StepStatusRunning, step.Description)
		time.Sleep(50 * time.Millisecond)
		_ = multiStep.UpdateStep(step.ID, progress.StepStatusCompleted, step.Description)
	}

	_ = multiStep.Success("Application deployed successfully!")

	fmt.Println("Deployment complete")
	// Output: Deployment complete
}

// Example_manager demonstrates using the progress manager.
func Example_manager() {
	// Create a progress manager
	manager := progress.NewManager(&progress.Config{
		Type:    progress.TypeSpinner,
		Enabled: false,
	})

	// Start progress
	prog, _ := manager.StartProgress("Initializing...", 0)

	// Update progress
	_ = prog.Update("Processing data...")

	time.Sleep(50 * time.Millisecond)

	// Mark as successful
	_ = manager.SuccessProgress("Done!")

	// Stop progress
	_ = manager.StopProgress()

	fmt.Println("Manager example complete")
	// Output: Manager example complete
}

// Example_workflowIntegration demonstrates integrating progress with workflows.
func Example_workflowIntegration() {
	// Create a workflow
	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:          "step1",
				Type:        workflow.StepTypeAPICall,
				Description: "Fetch data",
			},
			{
				ID:          "step2",
				Type:        workflow.StepTypeAPICall,
				Description: "Process data",
			},
		},
	}

	// Create progress manager
	manager := progress.NewManager(&progress.Config{
		Type:    progress.TypeSteps,
		Enabled: false,
	})

	// Create workflow integration
	integration := progress.NewWorkflowIntegration(manager)

	// Workflow lifecycle
	_ = integration.OnWorkflowStart(wf)

	// Simulate step execution
	_ = integration.OnStepStart("step1")
	time.Sleep(50 * time.Millisecond)
	_ = integration.OnStepComplete("step1")

	_ = integration.OnStepStart("step2")
	time.Sleep(50 * time.Millisecond)
	_ = integration.OnStepComplete("step2")

	_ = integration.OnWorkflowComplete(true, "Workflow completed successfully")

	fmt.Println("Workflow integration complete")
	// Output: Workflow integration complete
}

// Example_streaming demonstrates using streaming clients.
func Example_streaming() {
	// Create a streaming config
	config := &progress.StreamConfig{
		Type:            progress.StreamTypeSSE,
		Endpoint:        "http://example.com/status",
		PollingInterval: 1 * time.Second,
		Timeout:         5 * time.Second,
	}

	// Create an SSE client
	client := progress.NewSSEClient(config)

	// Subscribe to events
	_ = client.Subscribe("status", func(event *progress.Event) error {
		fmt.Printf("Received: %s\n", event.Type)
		return nil
	})

	// In a real application, you would:
	// ctx := context.Background()
	// client.Connect(ctx)
	// ... wait for events ...
	// client.Close()

	fmt.Println("Streaming example setup complete")
	// Output: Streaming example setup complete
}

// Example_watch demonstrates using watch mode.
func Example_watch() {
	// Create watch config
	watchConfig := &progress.WatchConfig{
		Enabled: true,
		StreamConfig: &progress.StreamConfig{
			Type:     progress.StreamTypeSSE,
			Endpoint: "http://example.com/logs",
		},
		ProgressConfig: &progress.Config{
			Type:    progress.TypeSpinner,
			Enabled: false,
		},
		ExitConditions: []*progress.ExitCondition{
			{
				EventType: "status",
				Condition: "event.data.state == 'completed'",
				Message:   "Operation completed",
			},
		},
		ShowLogs: true,
	}

	// Create watch coordinator
	watch, _ := progress.NewWatch(watchConfig)

	// In a real application, you would:
	// ctx := context.Background()
	// watch.Start(ctx)

	_ = watch

	fmt.Println("Watch mode example setup complete")
	// Output: Watch mode example setup complete
}

// Example_progressWithOpenAPI demonstrates using progress with OpenAPI config.
func Example_progressWithOpenAPI() {
	// OpenAPI operation config
	enabled := true
	opConfig := &openapi.CLIProgress{
		Enabled:              &enabled,
		Type:                 "spinner",
		ShowStepDescriptions: &enabled,
		ShowTimestamps:       &enabled,
	}

	// Create manager
	manager := progress.NewManager(&progress.Config{
		Type:    progress.TypeSpinner,
		Enabled: false,
	})

	// Start progress for operation
	prog, _ := manager.StartProgressForOperation(opConfig, "Executing API call...")

	_ = prog.Update("Sending request...")
	time.Sleep(50 * time.Millisecond)

	_ = manager.SuccessProgress("API call completed")
	_ = manager.StopProgress()

	fmt.Println("OpenAPI integration complete")
	// Output: OpenAPI integration complete
}

// Example_workflowWatch demonstrates watch mode for workflow execution.
func Example_workflowWatch() {
	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{ID: "step1", Description: "Initialize"},
			{ID: "step2", Description: "Execute"},
			{ID: "step3", Description: "Finalize"},
		},
	}

	watchConfig := &progress.WatchConfig{
		Enabled: true,
		StreamConfig: &progress.StreamConfig{
			Type:     progress.StreamTypePolling,
			Endpoint: "http://example.com/workflow/status",
		},
		ProgressConfig: &progress.Config{
			Type:    progress.TypeSteps,
			Enabled: false,
		},
	}

	coordinator, _ := progress.NewWorkflowWatchCoordinator(watchConfig, wf)

	// In real usage:
	// ctx := context.Background()
	// coordinator.Start(ctx)
	// ... workflow executes ...
	// coordinator.Success("Workflow completed")

	_ = coordinator

	fmt.Println("Workflow watch example complete")
	// Output: Workflow watch example complete
}

// Example_packageFunctions demonstrates package-level convenience functions.
func Example_packageFunctions() {
	// These use the default manager
	prog, _ := progress.StartProgress("Working...", 0)

	_ = progress.UpdateProgress("Still working...")

	time.Sleep(50 * time.Millisecond)

	_ = progress.SuccessProgress("All done!")

	_ = progress.StopProgress()

	_ = prog

	fmt.Println("Package functions example complete")
	// Output: Package functions example complete
}

// Example_customProgress demonstrates creating a custom progress indicator.
func Example_customProgress() {
	// Using the factory function
	config := &progress.Config{
		Type:                 progress.TypeBar,
		Enabled:              false,
		ShowTimestamps:       true,
		ShowStepDescriptions: true,
	}

	// Create appropriate progress type based on config
	prog := progress.New(config, 10)

	_ = prog.Start("Custom progress...")

	// Update with structured data
	data := &progress.Data{
		Message:    "Processing item 5/10",
		Current:    5,
		Total:      10,
		Percentage: 50.0,
		Timestamp:  time.Now(),
	}

	_ = prog.UpdateWithData(data)

	time.Sleep(50 * time.Millisecond)

	_ = prog.Success("Custom progress complete!")

	fmt.Println("Custom progress example complete")
	// Output: Custom progress example complete
}

// Example_stepInfo demonstrates working with step information.
func Example_stepInfo() {
	// Create a step with substeps
	mainStep := &progress.StepInfo{
		ID:          "deploy",
		Description: "Deploy application",
		Status:      progress.StepStatusRunning,
		StartTime:   time.Now(),
		Metadata: map[string]interface{}{
			"environment": "production",
			"version":     "1.0.0",
		},
	}

	// Add substeps
	mainStep.SubSteps = []*progress.StepInfo{
		{
			ID:          "deploy-backend",
			Description: "Deploy backend services",
			Status:      progress.StepStatusCompleted,
		},
		{
			ID:          "deploy-frontend",
			Description: "Deploy frontend",
			Status:      progress.StepStatusRunning,
		},
	}

	fmt.Printf("Main step: %s (%s)\n", mainStep.Description, mainStep.Status)
	fmt.Printf("Substeps: %d\n", len(mainStep.SubSteps))
	// Output:
	// Main step: Deploy application (running)
	// Substeps: 2
}

// Example_errorHandling demonstrates error handling with progress.
func Example_errorHandling() {
	config := &progress.Config{
		Type:    progress.TypeSpinner,
		Enabled: false,
	}

	spinner := progress.NewSpinner(config)

	// Start spinner
	if err := spinner.Start("Attempting operation..."); err != nil {
		fmt.Printf("Error starting spinner: %v\n", err)
		return
	}

	// Simulate an error
	time.Sleep(50 * time.Millisecond)

	// Mark as failed
	if err := spinner.Failure("Operation failed: connection timeout"); err != nil {
		fmt.Printf("Error marking failure: %v\n", err)
	}

	fmt.Println("Error handling example complete")
	// Output: Error handling example complete
}

// Example_contextCancellation demonstrates canceling operations.
func Example_contextCancellation() {
	watchConfig := &progress.WatchConfig{
		Enabled: true,
		StreamConfig: &progress.StreamConfig{
			Type:     progress.StreamTypePolling,
			Endpoint: "http://example.com/stream",
		},
		ProgressConfig: &progress.Config{
			Type:    progress.TypeSpinner,
			Enabled: false,
		},
	}

	watch, _ := progress.NewWatch(watchConfig)

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Start watch (will timeout)
	go func() {
		_ = watch.Start(ctx)
	}()

	// Wait for timeout
	<-ctx.Done()

	fmt.Println("Context cancellation example complete")
	// Output: Context cancellation example complete
}
