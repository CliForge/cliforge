package workflow

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"sync"
	"time"
)

// StepExecutor coordinates execution of all step types.
type StepExecutor struct {
	httpClient     *http.Client
	pluginExecutor interface{}
}

// NewStepExecutor creates a new step executor.
func NewStepExecutor(httpClient *http.Client, pluginExecutor interface{}) *StepExecutor {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &StepExecutor{
		httpClient:     httpClient,
		pluginExecutor: pluginExecutor,
	}
}

// ExecuteStep executes a single step with retry logic.
func (e *StepExecutor) ExecuteStep(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	// Check condition
	if step.Condition != "" {
		evaluator := NewExprEvaluator(ctx)
		shouldExecute, err := evaluator.EvaluateCondition(step.Condition)
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate step condition: %w", err)
		}

		if !shouldExecute {
			// Skip this step
			result := &StepResult{
				StepID:    step.ID,
				Success:   true,
				Output:    map[string]interface{}{"skipped": true, "reason": "condition not met"},
				StartTime: time.Now(),
			}
			result.EndTime = result.StartTime
			result.Duration = 0
			return result, nil
		}
	}

	// Execute with retry
	var result *StepResult
	var err error

	maxAttempts := 1
	if step.Retry != nil && step.Retry.MaxAttempts > 0 {
		maxAttempts = step.Retry.MaxAttempts
	}

	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			// Wait before retry
			backoffDuration := e.calculateBackoff(step, attempt)
			time.Sleep(backoffDuration)
		}

		// Execute the step
		result, err = e.executeStepByType(step, ctx)
		if result != nil {
			result.Retries = attempt
		}

		// Check if successful
		if err == nil && result != nil && result.Success {
			return result, nil
		}

		// Check if should retry
		if attempt < maxAttempts-1 {
			if !e.shouldRetry(step, result) {
				break
			}
		}
	}

	// All retries exhausted
	return result, err
}

// executeStepByType executes a step based on its type.
func (e *StepExecutor) executeStepByType(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	switch step.Type {
	case StepTypeAPICall:
		return e.executeAPICall(step, ctx)
	case StepTypePlugin:
		return e.executePlugin(step, ctx)
	case StepTypeConditional:
		return e.executeConditional(step, ctx)
	case StepTypeLoop:
		return e.executeLoop(step, ctx)
	case StepTypeWait:
		return e.executeWait(step, ctx)
	case StepTypeParallel:
		return e.executeParallel(step, ctx)
	case StepTypeNoop:
		return e.executeNoop(step)
	default:
		return nil, fmt.Errorf("unknown step type: %s", step.Type)
	}
}

// executeNoop executes a no-op step.
func (e *StepExecutor) executeNoop(step *Step) (*StepResult, error) {
	result := &StepResult{
		StepID:    step.ID,
		Success:   true,
		Output:    make(map[string]interface{}),
		StartTime: time.Now(),
	}
	result.EndTime = result.StartTime
	result.Duration = 0
	return result, nil
}

// shouldRetry determines if a step should be retried.
func (e *StepExecutor) shouldRetry(step *Step, result *StepResult) bool {
	if step.Retry == nil || result == nil {
		return false
	}

	switch step.Type {
	case StepTypeAPICall:
		return e.shouldRetryAPICall(result, step)
	case StepTypePlugin:
		return e.shouldRetryPlugin(result, step)
	default:
		return false
	}
}

// calculateBackoff calculates the backoff duration for a retry attempt.
func (e *StepExecutor) calculateBackoff(step *Step, attempt int) time.Duration {
	if step.Retry == nil || step.Retry.Backoff == nil {
		return time.Second
	}

	backoff := step.Retry.Backoff
	initialInterval := time.Duration(backoff.InitialInterval) * time.Second
	if initialInterval == 0 {
		initialInterval = time.Second
	}

	var duration time.Duration

	switch backoff.Type {
	case BackoffFixed:
		duration = initialInterval
	case BackoffLinear:
		duration = initialInterval * time.Duration(attempt+1)
	case BackoffExponential:
		multiplier := backoff.Multiplier
		if multiplier == 0 {
			multiplier = 2.0
		}
		duration = time.Duration(float64(initialInterval) * math.Pow(multiplier, float64(attempt)))
	default:
		duration = initialInterval
	}

	// Apply max interval
	if backoff.MaxInterval > 0 {
		maxDuration := time.Duration(backoff.MaxInterval) * time.Second
		if duration > maxDuration {
			duration = maxDuration
		}
	}

	return duration
}

// API Call execution

func (e *StepExecutor) executeAPICall(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	if step.APICall == nil {
		return nil, fmt.Errorf("api-call step %s missing configuration", step.ID)
	}

	result := &StepResult{
		StepID:    step.ID,
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	evaluator := NewExprEvaluator(ctx)

	endpoint, err := evaluator.InterpolateString(step.APICall.Endpoint)
	if err != nil {
		result.Error = fmt.Errorf("failed to interpolate endpoint: %w", err)
		result.Success = false
		return result, result.Error
	}

	method := step.APICall.Method
	if method == "" {
		method = "GET"
	}

	headers, err := evaluator.InterpolateStringMap(step.APICall.Headers)
	if err != nil {
		result.Error = fmt.Errorf("failed to interpolate headers: %w", err)
		result.Success = false
		return result, result.Error
	}

	query, err := evaluator.InterpolateStringMap(step.APICall.Query)
	if err != nil {
		result.Error = fmt.Errorf("failed to interpolate query: %w", err)
		result.Success = false
		return result, result.Error
	}

	var bodyReader io.Reader
	if step.APICall.Body != nil {
		interpolatedBody, err := evaluator.InterpolateMap(step.APICall.Body)
		if err != nil {
			result.Error = fmt.Errorf("failed to interpolate body: %w", err)
			result.Success = false
			return result, result.Error
		}

		bodyBytes, err := json.Marshal(interpolatedBody)
		if err != nil {
			result.Error = fmt.Errorf("failed to marshal body: %w", err)
			result.Success = false
			return result, result.Error
		}

		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		result.Success = false
		return result, result.Error
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	if len(query) > 0 {
		q := req.URL.Query()
		for key, value := range query {
			q.Set(key, value)
		}
		req.URL.RawQuery = q.Encode()
	}

	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("failed to execute request: %w", err)
		result.Success = false
		return result, result.Error
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response: %w", err)
		result.Success = false
		return result, result.Error
	}

	var responseData interface{}
	if len(responseBody) > 0 {
		if err := json.Unmarshal(responseBody, &responseData); err != nil {
			responseData = string(responseBody)
		}
	}

	result.Output["response"] = responseData
	result.Output["status_code"] = resp.StatusCode
	result.Output["headers"] = resp.Header

	if resp.StatusCode >= 400 {
		result.Error = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(responseBody))
		result.Success = false
		return result, result.Error
	}

	if step.Output != nil {
		for key, expr := range step.Output {
			value, err := evaluator.EvaluateExpression(expr)
			if err != nil {
				result.Output[key] = nil
			} else {
				result.Output[key] = value
			}
		}
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

func (e *StepExecutor) shouldRetryAPICall(result *StepResult, step *Step) bool {
	if step.Retry == nil {
		return false
	}

	if result.Retries >= step.Retry.MaxAttempts {
		return false
	}

	if step.Retry.RetryableErrors != nil {
		statusCode, ok := result.Output["status_code"].(int)
		if !ok {
			return false
		}

		for _, errorMatch := range step.Retry.RetryableErrors {
			if errorMatch.HTTPStatus != nil && *errorMatch.HTTPStatus == statusCode {
				return true
			}

			if errorMatch.HTTPStatus != nil && *errorMatch.HTTPStatus/100 == 5 && statusCode >= 500 && statusCode < 600 {
				return true
			}
		}

		return false
	}

	if statusCode, ok := result.Output["status_code"].(int); ok {
		return statusCode >= 500 && statusCode < 600
	}

	return false
}

// Plugin execution

func (e *StepExecutor) executePlugin(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	if step.Plugin == nil {
		return nil, fmt.Errorf("plugin step %s missing configuration", step.ID)
	}

	result := &StepResult{
		StepID:    step.ID,
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	evaluator := NewExprEvaluator(ctx)

	pluginName, err := evaluator.InterpolateString(step.Plugin.Plugin)
	if err != nil {
		result.Error = fmt.Errorf("failed to interpolate plugin name: %w", err)
		result.Success = false
		return result, result.Error
	}

	command, err := evaluator.InterpolateString(step.Plugin.Command)
	if err != nil {
		result.Error = fmt.Errorf("failed to interpolate command: %w", err)
		result.Success = false
		return result, result.Error
	}

	var input map[string]interface{}
	if step.Plugin.Input != nil {
		input, err = evaluator.InterpolateMap(step.Plugin.Input)
		if err != nil {
			result.Error = fmt.Errorf("failed to interpolate input: %w", err)
			result.Success = false
			return result, result.Error
		}
	}

	result.Output["plugin"] = pluginName
	result.Output["command"] = command
	result.Output["input"] = input
	result.Output["result"] = map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Plugin %s.%s executed (placeholder)", pluginName, command),
	}

	if step.Output != nil {
		for key, expr := range step.Output {
			value, err := evaluator.EvaluateExpression(expr)
			if err != nil {
				result.Output[key] = nil
			} else {
				result.Output[key] = value
			}
		}
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

func (e *StepExecutor) shouldRetryPlugin(result *StepResult, step *Step) bool {
	if step.Retry == nil {
		return false
	}

	if result.Retries >= step.Retry.MaxAttempts {
		return false
	}

	if step.Retry.RetryableErrors != nil {
		for _, errorMatch := range step.Retry.RetryableErrors {
			if errorMatch.ErrorType != "" {
				return true
			}
		}
		return false
	}

	return true
}

// Conditional execution

func (e *StepExecutor) executeConditional(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	if step.Conditional == nil {
		return nil, fmt.Errorf("conditional step %s missing configuration", step.ID)
	}

	result := &StepResult{
		StepID:    step.ID,
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	evaluator := NewExprEvaluator(ctx)

	conditionResult, err := evaluator.EvaluateCondition(step.Conditional.Condition)
	if err != nil {
		result.Error = fmt.Errorf("failed to evaluate condition: %w", err)
		result.Success = false
		return result, result.Error
	}

	result.Output["condition"] = conditionResult

	var branchSteps []*Step
	var branchName string

	if conditionResult {
		branchSteps = step.Conditional.Then
		branchName = "then"
	} else {
		branchSteps = step.Conditional.Else
		branchName = "else"
	}

	result.Output["branch"] = branchName

	branchResults := make([]*StepResult, 0)
	for _, branchStep := range branchSteps {
		stepResult, err := e.ExecuteStep(branchStep, ctx)
		if err != nil {
			result.Error = err
			result.Success = false
			return result, err
		}

		branchResults = append(branchResults, stepResult)

		if !stepResult.Success && branchStep.Required {
			result.Error = fmt.Errorf("required step %s in %s branch failed", branchStep.ID, branchName)
			result.Success = false
			return result, result.Error
		}
	}

	result.Output["branch_results"] = branchResults
	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// Loop execution

func (e *StepExecutor) executeLoop(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	if step.Loop == nil {
		return nil, fmt.Errorf("loop step %s missing configuration", step.ID)
	}

	result := &StepResult{
		StepID:    step.ID,
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	evaluator := NewExprEvaluator(ctx)

	collectionValue, err := evaluator.EvaluateExpression(step.Loop.Collection)
	if err != nil {
		result.Error = fmt.Errorf("failed to evaluate collection: %w", err)
		result.Success = false
		return result, result.Error
	}

	var collection []interface{}
	switch v := collectionValue.(type) {
	case []interface{}:
		collection = v
	case []string:
		collection = make([]interface{}, len(v))
		for i, item := range v {
			collection[i] = item
		}
	case []int:
		collection = make([]interface{}, len(v))
		for i, item := range v {
			collection[i] = item
		}
	default:
		result.Error = fmt.Errorf("collection is not iterable: %T", collectionValue)
		result.Success = false
		return result, result.Error
	}

	result.Output["collection_size"] = len(collection)

	iterationResults := make([]interface{}, 0)
	for i, item := range collection {
		iterCtx := ctx.Clone()
		iterCtx.SetVariable(step.Loop.Iterator, item)
		iterCtx.SetVariable(fmt.Sprintf("%s_index", step.Loop.Iterator), i)

		for _, loopStep := range step.Loop.Steps {
			stepResult, err := e.ExecuteStep(loopStep, iterCtx)
			if err != nil {
				result.Error = fmt.Errorf("iteration %d failed: %w", i, err)
				result.Success = false
				result.Output["iteration_results"] = iterationResults
				return result, result.Error
			}

			if !stepResult.Success && loopStep.Required {
				result.Error = fmt.Errorf("required step %s failed in iteration %d", loopStep.ID, i)
				result.Success = false
				result.Output["iteration_results"] = iterationResults
				return result, result.Error
			}

			iterationResults = append(iterationResults, map[string]interface{}{
				"index":  i,
				"item":   item,
				"result": stepResult,
			})
		}
	}

	result.Output["iteration_results"] = iterationResults
	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}

// Wait execution

func (e *StepExecutor) executeWait(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	if step.Wait == nil {
		return nil, fmt.Errorf("wait step %s missing configuration", step.ID)
	}

	result := &StepResult{
		StepID:    step.ID,
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	if step.Wait.Duration > 0 && step.Wait.Polling == nil {
		time.Sleep(time.Duration(step.Wait.Duration) * time.Second)
		result.Output["waited_seconds"] = step.Wait.Duration
		result.Success = true
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, nil
	}

	if step.Wait.Polling != nil {
		return e.executePolling(step, ctx, result)
	}

	result.Error = fmt.Errorf("wait step must have either duration or polling configuration")
	result.Success = false
	return result, result.Error
}

func (e *StepExecutor) executePolling(step *Step, ctx *ExecutionContext, result *StepResult) (*StepResult, error) {
	polling := step.Wait.Polling
	evaluator := NewExprEvaluator(ctx)

	endpoint, err := evaluator.InterpolateString(polling.Endpoint)
	if err != nil {
		result.Error = fmt.Errorf("failed to interpolate endpoint: %w", err)
		result.Success = false
		return result, result.Error
	}

	interval := polling.Interval
	if interval == 0 {
		interval = 10
	}

	timeout := polling.Timeout
	if timeout == 0 {
		timeout = 600
	}

	startTime := time.Now()
	pollCount := 0

	for {
		pollCount++

		if time.Since(startTime).Seconds() > float64(timeout) {
			result.Error = fmt.Errorf("polling timeout after %d seconds", timeout)
			result.Success = false
			result.Output["poll_count"] = pollCount
			return result, result.Error
		}

		resp, err := e.httpClient.Get(endpoint)
		if err != nil {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		var responseData map[string]interface{}
		if err := json.Unmarshal(body, &responseData); err != nil {
			time.Sleep(time.Duration(interval) * time.Second)
			continue
		}

		result.Output["response"] = responseData

		if polling.StatusField != "" && len(polling.TerminalStates) > 0 {
			status, ok := responseData[polling.StatusField].(string)
			if ok {
				for _, terminalState := range polling.TerminalStates {
					if status == terminalState {
						result.Output["final_status"] = status
						result.Output["poll_count"] = pollCount
						result.Success = true
						result.EndTime = time.Now()
						result.Duration = result.EndTime.Sub(result.StartTime)

						if step.Output != nil {
							for key, expr := range step.Output {
								value, err := evaluator.EvaluateExpression(expr)
								if err != nil {
									result.Output[key] = nil
								} else {
									result.Output[key] = value
								}
							}
						}

						return result, nil
					}
				}
			}
		}

		if step.Wait.Condition != "" {
			ctx.SetVariable("response", responseData)

			conditionMet, err := evaluator.EvaluateCondition(step.Wait.Condition)
			if err == nil && conditionMet {
				result.Output["poll_count"] = pollCount
				result.Success = true
				result.EndTime = time.Now()
				result.Duration = result.EndTime.Sub(result.StartTime)

				if step.Output != nil {
					for key, expr := range step.Output {
						value, err := evaluator.EvaluateExpression(expr)
						if err != nil {
							result.Output[key] = nil
						} else {
							result.Output[key] = value
						}
					}
				}

				return result, nil
			}
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

// Parallel execution

func (e *StepExecutor) executeParallel(step *Step, ctx *ExecutionContext) (*StepResult, error) {
	if step.Parallel == nil {
		return nil, fmt.Errorf("parallel step %s missing configuration", step.ID)
	}

	result := &StepResult{
		StepID:    step.ID,
		StartTime: time.Now(),
		Output:    make(map[string]interface{}),
	}

	if len(step.Parallel.Steps) == 0 {
		result.Success = true
		result.EndTime = time.Now()
		result.Duration = result.EndTime.Sub(result.StartTime)
		return result, nil
	}

	var wg sync.WaitGroup
	resultsChan := make(chan *stepExecutionResult, len(step.Parallel.Steps))

	for _, parallelStep := range step.Parallel.Steps {
		wg.Add(1)
		go func(s *Step) {
			defer wg.Done()

			parallelCtx := ctx.Clone()

			stepResult, err := e.ExecuteStep(s, parallelCtx)
			resultsChan <- &stepExecutionResult{
				step:   s,
				result: stepResult,
				err:    err,
			}
		}(parallelStep)
	}

	wg.Wait()
	close(resultsChan)

	parallelResults := make(map[string]*StepResult)
	var errors []error
	allSuccess := true

	for execResult := range resultsChan {
		parallelResults[execResult.step.ID] = execResult.result

		if execResult.err != nil || !execResult.result.Success {
			allSuccess = false
			if execResult.err != nil {
				errors = append(errors, execResult.err)
			}
		}

		ctx.SetStepResult(execResult.step.ID, execResult.result)
	}

	result.Output["parallel_results"] = parallelResults
	result.Output["step_count"] = len(step.Parallel.Steps)

	if !allSuccess {
		if len(errors) > 0 {
			result.Error = fmt.Errorf("parallel execution failed: %v", errors[0])
		} else {
			result.Error = fmt.Errorf("one or more parallel steps failed")
		}
		result.Success = false
		return result, result.Error
	}

	result.Success = true
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)

	return result, nil
}
