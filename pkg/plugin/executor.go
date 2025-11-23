package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

// Executor handles plugin execution with timeouts and error handling.
type Executor struct {
	defaultTimeout time.Duration
	maxTimeout     time.Duration
}

// NewExecutor creates a new plugin executor.
func NewExecutor(defaultTimeout, maxTimeout time.Duration) *Executor {
	if defaultTimeout == 0 {
		defaultTimeout = 30 * time.Second
	}
	if maxTimeout == 0 {
		maxTimeout = 5 * time.Minute
	}

	return &Executor{
		defaultTimeout: defaultTimeout,
		maxTimeout:     maxTimeout,
	}
}

// Execute executes a plugin with the given input.
// Applies timeout and collects output/errors.
func (e *Executor) Execute(ctx context.Context, plugin Plugin, input *PluginInput) (*PluginOutput, error) {
	// Determine timeout
	timeout := e.defaultTimeout
	if input.Timeout > 0 {
		timeout = input.Timeout
		if timeout > e.maxTimeout {
			timeout = e.maxTimeout
		}
	}

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Track execution time
	startTime := time.Now()

	// Execute the plugin
	output, err := plugin.Execute(execCtx, input)
	duration := time.Since(startTime)

	if err != nil {
		// Check if it was a timeout
		if execCtx.Err() == context.DeadlineExceeded {
			return nil, NewPluginError(
				plugin.Describe().Manifest.Name,
				fmt.Sprintf("execution timed out after %v", timeout),
				err,
			).WithSuggestion("Increase timeout or simplify the operation")
		}
		return nil, err
	}

	// Set duration in output
	if output != nil {
		output.Duration = duration
	}

	return output, nil
}

// ExecuteBinary executes an external binary plugin.
func (e *Executor) ExecuteBinary(ctx context.Context, execPath string, input *PluginInput) (*PluginOutput, error) {
	// Prepare the JSON-RPC request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"id":      "1",
		"method":  "execute",
		"params":  input,
	}

	requestData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create command
	cmd := exec.CommandContext(ctx, execPath)

	// Set up stdin/stdout/stderr
	cmd.Stdin = bytes.NewReader(requestData)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment variables
	if input.Env != nil {
		env := make([]string, 0, len(input.Env))
		for k, v := range input.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// Set working directory
	if input.WorkingDir != "" {
		cmd.Dir = input.WorkingDir
	}

	// Track execution time
	startTime := time.Now()

	// Execute the command
	err = cmd.Run()
	duration := time.Since(startTime)

	// Get exit code
	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute binary: %w", err)
		}
	}

	// Parse JSON-RPC response
	var response struct {
		JSONRPC string `json:"jsonrpc"`
		ID      string `json:"id"`
		Result  struct {
			Stdout   string                 `json:"stdout"`
			Stderr   string                 `json:"stderr"`
			ExitCode int                    `json:"exit_code"`
			Data     map[string]interface{} `json:"data"`
			Error    string                 `json:"error"`
		} `json:"result"`
		Error *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	stdoutData := stdout.Bytes()
	if len(stdoutData) > 0 {
		if err := json.Unmarshal(stdoutData, &response); err != nil {
			// If not valid JSON-RPC, treat stdout as plain output
			return &PluginOutput{
				Stdout:   stdout.String(),
				Stderr:   stderr.String(),
				ExitCode: exitCode,
				Duration: duration,
			}, nil
		}

		// Check for JSON-RPC error
		if response.Error != nil {
			return &PluginOutput{
				Stderr:   response.Error.Message,
				ExitCode: response.Error.Code,
				Error:    response.Error.Message,
				Duration: duration,
			}, nil
		}

		// Return structured response
		return &PluginOutput{
			Stdout:   response.Result.Stdout,
			Stderr:   response.Result.Stderr,
			ExitCode: response.Result.ExitCode,
			Data:     response.Result.Data,
			Error:    response.Result.Error,
			Duration: duration,
		}, nil
	}

	// No stdout, return stderr and exit code
	return &PluginOutput{
		Stdout:   "",
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
	}, nil
}

// ExecuteWithRetry executes a plugin with retry logic.
func (e *Executor) ExecuteWithRetry(ctx context.Context, plugin Plugin, input *PluginInput, maxRetries int) (*PluginOutput, error) {
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		output, err := e.Execute(ctx, plugin, input)
		if err == nil {
			return output, nil
		}

		lastErr = err

		// Check if error is recoverable
		if pluginErr, ok := err.(*PluginError); ok {
			if !pluginErr.Recoverable {
				return nil, err
			}
		}

		// Don't retry on context cancellation
		if ctx.Err() != nil {
			return nil, err
		}

		// Wait before retry (exponential backoff)
		if attempt < maxRetries {
			backoff := time.Duration(1<<uint(attempt)) * time.Second
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// ExecuteParallel executes multiple plugins in parallel.
func (e *Executor) ExecuteParallel(ctx context.Context, executions []PluginExecution) ([]PluginExecutionResult, error) {
	results := make([]PluginExecutionResult, len(executions))
	errChan := make(chan error, len(executions))
	doneChan := make(chan struct{})

	// Execute each plugin in a goroutine
	for i, execution := range executions {
		go func(idx int, exec PluginExecution) {
			output, err := e.Execute(ctx, exec.Plugin, exec.Input)
			results[idx] = PluginExecutionResult{
				Output: output,
				Error:  err,
				Index:  idx,
			}

			if err != nil {
				errChan <- err
			}
		}(i, execution)
	}

	// Wait for all to complete or context cancellation
	go func() {
		for i := 0; i < len(executions); i++ {
			<-errChan
		}
		close(doneChan)
	}()

	select {
	case <-doneChan:
		// All completed
	case <-ctx.Done():
		return results, ctx.Err()
	}

	return results, nil
}

// PluginExecution represents a plugin execution request.
type PluginExecution struct {
	Plugin Plugin
	Input  *PluginInput
}

// PluginExecutionResult represents the result of a plugin execution.
type PluginExecutionResult struct {
	Output *PluginOutput
	Error  error
	Index  int
}

// ExecutionStats tracks plugin execution statistics.
type ExecutionStats struct {
	PluginName    string
	TotalRuns     int64
	SuccessRuns   int64
	FailedRuns    int64
	AverageDuration time.Duration
	LastExecuted  time.Time
}

// StatsCollector collects plugin execution statistics.
type StatsCollector struct {
	stats map[string]*ExecutionStats
}

// NewStatsCollector creates a new stats collector.
func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		stats: make(map[string]*ExecutionStats),
	}
}

// Record records a plugin execution.
func (s *StatsCollector) Record(pluginName string, duration time.Duration, success bool) {
	stats, exists := s.stats[pluginName]
	if !exists {
		stats = &ExecutionStats{
			PluginName: pluginName,
		}
		s.stats[pluginName] = stats
	}

	stats.TotalRuns++
	if success {
		stats.SuccessRuns++
	} else {
		stats.FailedRuns++
	}

	// Update average duration
	if stats.TotalRuns == 1 {
		stats.AverageDuration = duration
	} else {
		total := stats.AverageDuration * time.Duration(stats.TotalRuns-1)
		stats.AverageDuration = (total + duration) / time.Duration(stats.TotalRuns)
	}

	stats.LastExecuted = time.Now()
}

// GetStats returns statistics for a plugin.
func (s *StatsCollector) GetStats(pluginName string) *ExecutionStats {
	return s.stats[pluginName]
}

// GetAllStats returns all statistics.
func (s *StatsCollector) GetAllStats() map[string]*ExecutionStats {
	return s.stats
}
