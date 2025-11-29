// Package executor handles execution of OpenAPI operations from Cobra commands.
//
// The executor package implements the runtime execution engine for generated
// CLIs, handling HTTP requests, authentication, response formatting, async
// operations, workflows, and error handling. It bridges Cobra commands with
// OpenAPI operations.
//
// # Execution Flow
//
//  1. Extract operation metadata from command annotations
//  2. Build HTTP request from command flags and args
//  3. Apply authentication headers
//  4. Execute HTTP request
//  5. Handle response (sync, async, workflow)
//  6. Format and display output
//  7. Update state and history
//
// # Features
//
//   - Automatic request building from OpenAPI specs
//   - Authentication injection (API key, OAuth2, Basic)
//   - Path/query/header parameter mapping
//   - Request body construction from flags
//   - Async operation polling with progress display
//   - Multi-step workflow execution
//   - Response formatting (JSON, YAML, table, etc.)
//   - Error handling with helpful messages
//
// # Example Usage
//
//	// Create executor
//	executor, _ := executor.NewExecutor(spec, &executor.ExecutorConfig{
//	    BaseURL: "https://api.example.com",
//	    AuthManager: authManager,
//	    OutputManager: outputManager,
//	})
//
//	// Wire to command
//	cmd.RunE = func(cmd *cobra.Command, args []string) error {
//	    return executor.Execute(cmd, args)
//	}
//
// # Async Operations
//
// For operations marked with x-cli-async, the executor automatically
// polls for completion:
//
//	x-cli-async:
//	  enabled: true
//	  status-field: status
//	  terminal-states: [completed, failed]
//	  polling:
//	    interval: 5
//	    timeout: 300
//
// # Workflow Execution
//
// For x-cli-workflow operations, the executor delegates to the workflow
// engine for multi-step execution with dependencies, retries, and rollback.
//
// The executor integrates with progress indicators, state management,
// and secrets masking for a complete CLI execution experience.
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/CliForge/cliforge/internal/builder"
	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/output"
	"github.com/CliForge/cliforge/pkg/progress"
	"github.com/CliForge/cliforge/pkg/state"
	"github.com/CliForge/cliforge/pkg/workflow"
	"github.com/spf13/cobra"
)

// Executor executes OpenAPI operations from Cobra commands.
type Executor struct {
	spec          *openapi.ParsedSpec
	httpClient    *http.Client
	baseURL       string
	authManager   *auth.Manager
	outputManager *output.Manager
	stateManager  *state.Manager
	progressMgr   *progress.Manager
}

// ExecutorConfig configures the executor.
type ExecutorConfig struct {
	BaseURL       string
	HTTPClient    *http.Client
	AuthManager   *auth.Manager
	OutputManager *output.Manager
	StateManager  *state.Manager
	ProgressMgr   *progress.Manager
}

// NewExecutor creates a new command executor.
func NewExecutor(spec *openapi.ParsedSpec, config *ExecutorConfig) (*Executor, error) {
	if config == nil {
		return nil, fmt.Errorf("executor config is required")
	}

	// Use provided HTTP client or create default
	httpClient := config.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}

	return &Executor{
		spec:          spec,
		httpClient:    httpClient,
		baseURL:       config.BaseURL,
		authManager:   config.AuthManager,
		outputManager: config.OutputManager,
		stateManager:  config.StateManager,
		progressMgr:   config.ProgressMgr,
	}, nil
}

// Execute executes a command by making the appropriate HTTP request.
func (e *Executor) Execute(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get operation metadata from command annotations
	operationID, ok := cmd.Annotations["operationID"]
	if !ok {
		return fmt.Errorf("command has no operationID annotation")
	}

	// Find the operation in the spec
	operations, err := e.spec.GetOperations()
	if err != nil {
		return fmt.Errorf("failed to get operations: %w", err)
	}

	var operation *openapi.Operation
	for _, op := range operations {
		if op.OperationID == operationID {
			operation = op
			break
		}
	}

	if operation == nil {
		return fmt.Errorf("operation %s not found in spec", operationID)
	}

	// Check if operation uses workflow
	if operation.CLIWorkflow != nil {
		return e.executeWorkflow(ctx, cmd, operation)
	}

	// Execute regular HTTP operation
	return e.executeHTTPOperation(ctx, cmd, operation, args)
}

// executeHTTPOperation executes a single HTTP operation.
func (e *Executor) executeHTTPOperation(ctx context.Context, cmd *cobra.Command, op *openapi.Operation, args []string) error {
	// Start progress indicator
	var prog progress.Progress
	if e.progressMgr != nil {
		var err error
		prog, err = e.progressMgr.StartProgress(op.Summary, 0)
		if err == nil {
			defer func() { _ = prog.Stop() }()
		}
	}

	// Build request
	req, err := e.buildRequest(ctx, cmd, op, args)
	if err != nil {
		if prog != nil {
			_ = prog.Failure("Failed to build request")
		}
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Apply authentication
	if e.authManager != nil {
		if err := e.applyAuth(ctx, req); err != nil {
			if prog != nil {
				_ = prog.Failure("Authentication failed")
			}
			return fmt.Errorf("failed to apply authentication: %w", err)
		}
	}

	// Execute request
	if prog != nil {
		_ = prog.Update("Sending request...")
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		if prog != nil {
			_ = prog.Failure("Request failed")
		}
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		if prog != nil {
			_ = prog.Failure("Failed to read response")
		}
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		if prog != nil {
			_ = prog.Failure(fmt.Sprintf("Request failed with status %d", resp.StatusCode))
		}
		return e.handleErrorResponse(resp, body, op)
	}

	// Handle async operations
	if op.CLIAsync != nil && op.CLIAsync.Enabled {
		return e.handleAsyncOperation(ctx, resp, body, op, prog)
	}

	// Success
	if prog != nil {
		_ = prog.Success("Request completed")
	}

	// Format and output response
	return e.formatOutput(cmd, resp, body, op)
}

// buildRequest builds an HTTP request from command flags and operation.
func (e *Executor) buildRequest(ctx context.Context, cmd *cobra.Command, op *openapi.Operation, args []string) (*http.Request, error) {
	// Build URL
	reqURL, err := e.buildURL(cmd, op, args)
	if err != nil {
		return nil, err
	}

	// Build request body
	var bodyReader io.Reader
	if op.Operation.RequestBody != nil {
		bodyData, err := builder.BuildRequestBody(cmd)
		if err != nil {
			return nil, fmt.Errorf("failed to build request body: %w", err)
		}

		if len(bodyData) > 0 {
			bodyBytes, err := json.Marshal(bodyData)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			bodyReader = bytes.NewReader(bodyBytes)
		}
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, op.Method, reqURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	// Add custom headers from parameters
	params, _ := builder.BuildRequestParams(cmd)
	for key, value := range params {
		// Check if this is a header parameter
		paramKey := fmt.Sprintf("param:%s:in", toFlagName(key))
		if cmd.Annotations != nil {
			if in, ok := cmd.Annotations[paramKey]; ok && in == "header" {
				req.Header.Set(key, fmt.Sprintf("%v", value))
			}
		}
	}

	return req, nil
}

// buildURL builds the request URL from operation path and parameters.
func (e *Executor) buildURL(cmd *cobra.Command, op *openapi.Operation, args []string) (string, error) {
	// Start with base URL
	baseURL := e.baseURL
	if baseURL == "" {
		// Try to get from spec servers
		if len(e.spec.Spec.Servers) > 0 {
			baseURL = e.spec.Spec.Servers[0].URL
		} else {
			return "", fmt.Errorf("no base URL configured")
		}
	}

	// Build path with path parameters
	path := op.Path
	pathParams := extractPathParams(op.Path)

	// Get parameter values from flags or args
	params, _ := builder.BuildRequestParams(cmd)

	// Replace path parameters
	argIndex := 0
	for _, paramName := range pathParams {
		var paramValue string

		// Try to get from flags first
		if val, ok := params[paramName]; ok {
			paramValue = fmt.Sprintf("%v", val)
		} else if argIndex < len(args) {
			// Use positional argument
			paramValue = args[argIndex]
			argIndex++
		} else {
			return "", fmt.Errorf("missing value for path parameter: %s", paramName)
		}

		// Replace in path
		path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", paramName), url.PathEscape(paramValue))
	}

	// Build full URL
	fullURL := strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")

	// Add query parameters
	parsedURL, err := url.Parse(fullURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	query := parsedURL.Query()
	for key, value := range params {
		// Check if this is a query parameter
		paramKey := fmt.Sprintf("param:%s:in", toFlagName(key))
		if cmd.Annotations != nil {
			if in, ok := cmd.Annotations[paramKey]; ok && in == "query" {
				query.Set(key, fmt.Sprintf("%v", value))
			}
		}
	}
	parsedURL.RawQuery = query.Encode()

	return parsedURL.String(), nil
}

// applyAuth applies authentication to the request.
func (e *Executor) applyAuth(ctx context.Context, req *http.Request) error {
	// Get token from auth manager
	token, err := e.authManager.GetToken(ctx, "")
	if err != nil {
		return err
	}

	// Get authenticator to apply headers
	authenticator, err := e.authManager.GetAuthenticator("")
	if err != nil {
		return err
	}

	// Apply auth headers
	headers := authenticator.GetHeaders(token)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return nil
}

// handleErrorResponse handles error responses.
func (e *Executor) handleErrorResponse(resp *http.Response, body []byte, op *openapi.Operation) error {
	// Try to parse error response
	var errorData map[string]interface{}
	if err := json.Unmarshal(body, &errorData); err != nil {
		// Not JSON, return raw error
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Format error message
	errorMsg := fmt.Sprintf("HTTP %d", resp.StatusCode)
	if msg, ok := errorData["message"].(string); ok {
		errorMsg = fmt.Sprintf("%s: %s", errorMsg, msg)
	} else if msg, ok := errorData["error"].(string); ok {
		errorMsg = fmt.Sprintf("%s: %s", errorMsg, msg)
	}

	return fmt.Errorf("%s", errorMsg)
}

// handleAsyncOperation handles async operations with polling.
func (e *Executor) handleAsyncOperation(ctx context.Context, resp *http.Response, body []byte, op *openapi.Operation, prog progress.Progress) error {
	// Parse initial response
	var respData map[string]interface{}
	if err := json.Unmarshal(body, &respData); err != nil {
		return fmt.Errorf("failed to parse async response: %w", err)
	}

	// Get status field
	statusField := op.CLIAsync.StatusField
	if statusField == "" {
		statusField = "status"
	}

	// Poll for completion
	pollConfig := op.CLIAsync.Polling
	if pollConfig == nil {
		pollConfig = &openapi.PollingConfig{
			Interval: 5,
			Timeout:  300,
		}
	}

	ticker := time.NewTicker(time.Duration(pollConfig.Interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(pollConfig.Timeout) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-timeout:
			return fmt.Errorf("operation timed out after %d seconds", pollConfig.Timeout)

		case <-ticker.C:
			// Check status
			status, ok := respData[statusField].(string)
			if !ok {
				return fmt.Errorf("status field %s not found in response", statusField)
			}

			// Update progress
			if prog != nil {
				_ = prog.Update(fmt.Sprintf("Status: %s", status))
			}

			// Check if terminal state
			isTerminal := false
			for _, terminalState := range op.CLIAsync.TerminalStates {
				if status == terminalState {
					isTerminal = true
					break
				}
			}

			if isTerminal {
				if prog != nil {
					_ = prog.Success(fmt.Sprintf("Operation completed with status: %s", status))
				}
				// Return the final response data
				return nil
			}

			// Poll status endpoint
			if op.CLIAsync.StatusEndpoint != "" {
				statusResp, err := e.pollStatus(ctx, op.CLIAsync.StatusEndpoint, respData)
				if err != nil {
					return fmt.Errorf("failed to poll status: %w", err)
				}
				respData = statusResp
			}
		}
	}
}

// pollStatus polls the status endpoint.
func (e *Executor) pollStatus(ctx context.Context, statusEndpoint string, initialResp map[string]interface{}) (map[string]interface{}, error) {
	// Build status URL (may contain template variables)
	statusURL := statusEndpoint
	// TODO: Template variable replacement from initialResp

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var statusData map[string]interface{}
	if err := json.Unmarshal(body, &statusData); err != nil {
		return nil, err
	}

	return statusData, nil
}

// formatOutput formats and displays the response.
func (e *Executor) formatOutput(cmd *cobra.Command, resp *http.Response, body []byte, op *openapi.Operation) error {
	// Parse response body
	var data interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &data); err != nil {
			// Not JSON, use raw string
			data = string(body)
		}
	}

	// Get output format from flags
	outputFormat, _ := cmd.Flags().GetString("output")

	// Use output manager to format
	if e.outputManager != nil {
		// Apply output configuration from operation
		formatConfig := e.outputManager.ApplyOutputRules(op.CLIOutput)

		return e.outputManager.FormatWithConfig(cmd.OutOrStdout(), data, outputFormat, formatConfig)
	}

	// Fallback to simple JSON output
	formatted, _ := json.MarshalIndent(data, "", "  ")
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(formatted))

	return nil
}

// executeWorkflow executes a workflow operation.
func (e *Executor) executeWorkflow(ctx context.Context, cmd *cobra.Command, op *openapi.Operation) error {
	// Convert CLI workflow to workflow engine format
	wf, err := e.convertToWorkflow(op.CLIWorkflow)
	if err != nil {
		return fmt.Errorf("failed to convert workflow: %w", err)
	}

	// Create workflow executor
	workflowExec, err := workflow.NewExecutor(wf, e.httpClient, nil)
	if err != nil {
		return fmt.Errorf("failed to create workflow executor: %w", err)
	}

	// Start workflow progress
	var prog progress.Progress
	if e.progressMgr != nil {
		prog, _ = e.progressMgr.StartProgress("Executing workflow", len(wf.Steps))
		if prog != nil {
			defer func() { _ = prog.Stop() }()
		}
	}

	// Execute workflow
	execCtx := workflow.NewExecutionContext(nil)
	state, err := workflowExec.Execute(execCtx)
	if err != nil {
		if prog != nil {
			_ = prog.Failure("Workflow failed")
		}
		return fmt.Errorf("workflow execution failed: %w", err)
	}

	// Success
	if prog != nil {
		_ = prog.Success("Workflow completed")
	}

	// Format output
	if e.outputManager != nil {
		outputFormat, _ := cmd.Flags().GetString("output")
		return e.outputManager.Format(cmd.OutOrStdout(), state, outputFormat)
	}

	return nil
}

// convertToWorkflow converts CLI workflow to workflow engine format.
func (e *Executor) convertToWorkflow(cliWorkflow *openapi.CLIWorkflow) (*workflow.Workflow, error) {
	wf := &workflow.Workflow{
		Steps: make([]*workflow.Step, 0, len(cliWorkflow.Steps)),
	}

	for _, cliStep := range cliWorkflow.Steps {
		step := &workflow.Step{
			ID:        cliStep.ID,
			Type:      workflow.StepTypeAPICall,
			Condition: cliStep.Condition,
		}

		if cliStep.Request != nil {
			step.APICall = &workflow.APICallStep{
				Method:   cliStep.Request.Method,
				Endpoint: cliStep.Request.URL,
				Headers:  cliStep.Request.Headers,
				Body:     cliStep.Request.Body,
			}
		}

		wf.Steps = append(wf.Steps, step)
	}

	return wf, nil
}

// extractPathParams extracts path parameter names from a path template.
func extractPathParams(path string) []string {
	var params []string
	parts := strings.Split(path, "/")

	for _, part := range parts {
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			param := strings.TrimSuffix(strings.TrimPrefix(part, "{"), "}")
			params = append(params, param)
		}
	}

	return params
}

// toFlagName converts a string to flag name format.
func toFlagName(s string) string {
	s = strings.ReplaceAll(s, "_", "-")
	s = strings.ToLower(s)
	return s
}
