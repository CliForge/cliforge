package executor

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/output"
	"github.com/CliForge/cliforge/pkg/progress"
	"github.com/spf13/cobra"
)

// Mock HTTP transport for testing
type mockTransport struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.roundTripFunc(req)
}

func newMockHTTPClient(handler func(*http.Request) (*http.Response, error)) *http.Client {
	return &http.Client{
		Transport: &mockTransport{roundTripFunc: handler},
	}
}

func TestNewExecutor(t *testing.T) {
	tests := []struct {
		name    string
		spec    *openapi.ParsedSpec
		config  *ExecutorConfig
		wantErr bool
	}{
		{
			name: "valid config",
			spec: &openapi.ParsedSpec{},
			config: &ExecutorConfig{
				BaseURL: "https://api.example.com",
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			spec:    &openapi.ParsedSpec{},
			config:  nil,
			wantErr: true,
		},
		{
			name: "config with custom HTTP client",
			spec: &openapi.ParsedSpec{},
			config: &ExecutorConfig{
				BaseURL:    "https://api.example.com",
				HTTPClient: &http.Client{Timeout: 10 * time.Second},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewExecutor(tt.spec, tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewExecutor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && executor == nil {
				t.Error("Expected executor to be non-nil")
			}
		})
	}
}

func TestExecutor_Execute(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items": [{"id": "1", "name": "test"}]}`))
	}))
	defer server.Close()

	// Parse spec
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Create executor
	authMgr := auth.NewManager("test")
	_ = authMgr.RegisterAuthenticator("default", &auth.NoneAuth{})

	config := &ExecutorConfig{
		BaseURL:       server.URL,
		OutputManager: output.NewManager(),
		AuthManager:   authMgr,
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		wantErr     bool
		errContains string
	}{
		{
			name: "successful execution",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "list"}
				cmd.SetContext(context.Background())
				cmd.Annotations = map[string]string{
					"operationID": "listUsers",
				}
				cmd.Flags().String("output", "json", "Output format")
				return cmd
			},
			wantErr: false,
		},
		{
			name: "missing operation ID",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Annotations = map[string]string{}
				return cmd
			},
			wantErr:     true,
			errContains: "no operationID annotation",
		},
		{
			name: "operation not found",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Annotations = map[string]string{
					"operationID": "nonExistentOperation",
				}
				return cmd
			},
			wantErr:     true,
			errContains: "not found in spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			err := executor.Execute(cmd, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("Execute() error = %v, should contain %v", err, tt.errContains)
			}
		})
	}
}

func TestExecutor_BuildURL(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL: "https://api.example.com/v1",
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp, getOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
		}
		if op.OperationID == "getUser" {
			getOp = op
		}
	}

	tests := []struct {
		name        string
		operation   *openapi.Operation
		setupCmd    func() *cobra.Command
		args        []string
		expectedURL string
		wantErr     bool
	}{
		{
			name:      "simple path with query params",
			operation: listOp,
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "list"}
				cmd.Annotations = map[string]string{
					"param:limit":    "limit",
					"param:limit:in": "query",
				}
				cmd.Flags().Int("limit", 100, "Limit")
				_ = cmd.Flags().Set("limit", "50")
				return cmd
			},
			args:        nil,
			expectedURL: "https://api.example.com/v1/users?limit=50",
			wantErr:     false,
		},
		{
			name:      "path with path parameter from args",
			operation: getOp,
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "get"}
				cmd.Annotations = make(map[string]string)
				return cmd
			},
			args:        []string{"user-123"},
			expectedURL: "https://api.example.com/v1/users/user-123",
			wantErr:     false,
		},
		{
			name:      "path with escaped parameters",
			operation: getOp,
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "get"}
				cmd.Annotations = make(map[string]string)
				return cmd
			},
			args:        []string{"user/with/slashes"},
			expectedURL: "https://api.example.com/v1/users/user%2Fwith%2Fslashes",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			url, err := executor.buildURL(cmd, tt.operation, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && url != tt.expectedURL {
				t.Errorf("buildURL() = %v, want %v", url, tt.expectedURL)
			}
		})
	}
}

func TestExecutor_BuildURLWithNoBaseURL(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Create executor without base URL
	config := &ExecutorConfig{}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)

	// Should use server URL from spec
	url, err := executor.buildURL(cmd, listOp, nil)
	if err != nil {
		t.Errorf("buildURL() should use spec server URL, got error: %v", err)
	}
	if !strings.Contains(url, "/users") {
		t.Errorf("buildURL() = %v, should contain /users", url)
	}
}

func TestExecutor_BuildRequest(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL: "https://api.example.com",
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var createOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "createUser" {
			createOp = op
			break
		}
	}

	if createOp == nil {
		t.Skip("createUser operation not found in spec")
	}

	tests := []struct {
		name            string
		operation       *openapi.Operation
		setupCmd        func() *cobra.Command
		wantMethod      string
		wantHasBody     bool
		wantContentType string
	}{
		{
			name:      "POST request with body",
			operation: createOp,
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "create"}
				cmd.Annotations = map[string]string{
					"body:name": "name",
				}
				cmd.Flags().String("name", "", "Name")
				_ = cmd.Flags().Set("name", "test-user")
				return cmd
			},
			wantMethod:      "POST",
			wantHasBody:     true,
			wantContentType: "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()
			ctx := context.Background()
			req, err := executor.buildRequest(ctx, cmd, tt.operation, nil)
			if err != nil {
				t.Errorf("buildRequest() error = %v", err)
				return
			}

			if req.Method != tt.wantMethod {
				t.Errorf("buildRequest() method = %v, want %v", req.Method, tt.wantMethod)
			}

			hasBody := req.Body != nil
			if hasBody != tt.wantHasBody {
				t.Errorf("buildRequest() hasBody = %v, want %v", hasBody, tt.wantHasBody)
			}

			if tt.wantContentType != "" {
				ct := req.Header.Get("Content-Type")
				if ct != tt.wantContentType {
					t.Errorf("buildRequest() Content-Type = %v, want %v", ct, tt.wantContentType)
				}
			}
		})
	}
}

func TestExecutor_ApplyAuth(t *testing.T) {
	tests := []struct {
		name        string
		setupAuth   func() *auth.Manager
		wantErr     bool
		errContains string
	}{
		{
			name: "no auth",
			setupAuth: func() *auth.Manager {
				mgr := auth.NewManager("test")
				_ = mgr.RegisterAuthenticator("default", &auth.NoneAuth{})
				return mgr
			},
			wantErr: false,
		},
		{
			name: "api key auth",
			setupAuth: func() *auth.Manager {
				mgr := auth.NewManager("test")
				apiKeyAuth, _ := auth.NewAPIKeyAuth(&auth.APIKeyConfig{
					Location: auth.APIKeyLocationHeader,
					Name:     "X-API-Key",
					Key:      "test-key-123",
				})
				_ = mgr.RegisterAuthenticator("default", apiKeyAuth)
				return mgr
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authMgr := tt.setupAuth()
			executor := &Executor{
				authManager: authMgr,
			}

			req, _ := http.NewRequest("GET", "https://api.example.com/test", nil)
			err := executor.applyAuth(context.Background(), req)

			if (err != nil) != tt.wantErr {
				t.Errorf("applyAuth() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutor_HandleErrorResponse(t *testing.T) {
	executor := &Executor{}

	tests := []struct {
		name        string
		statusCode  int
		body        []byte
		errContains string
	}{
		{
			name:        "JSON error with message field",
			statusCode:  400,
			body:        []byte(`{"message": "Invalid request"}`),
			errContains: "Invalid request",
		},
		{
			name:        "JSON error with error field",
			statusCode:  500,
			body:        []byte(`{"error": "Internal server error"}`),
			errContains: "Internal server error",
		},
		{
			name:        "non-JSON error",
			statusCode:  404,
			body:        []byte("Not found"),
			errContains: "Not found",
		},
		{
			name:        "JSON without message",
			statusCode:  403,
			body:        []byte(`{"code": 403}`),
			errContains: "HTTP 403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{
				StatusCode: tt.statusCode,
			}

			err := executor.handleErrorResponse(resp, tt.body, nil)
			if err == nil {
				t.Error("handleErrorResponse() should return error")
				return
			}

			if !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("handleErrorResponse() error = %v, should contain %v", err, tt.errContains)
			}
		})
	}
}

func TestExecutor_HandleAsyncOperation(t *testing.T) {
	tests := []struct {
		name          string
		initialBody   string
		operation     *openapi.Operation
		mockResponses []string
		wantErr       bool
		errContains   string
	}{
		{
			name:        "completes immediately",
			initialBody: `{"status": "completed", "id": "123"}`,
			operation: &openapi.Operation{
				CLIAsync: &openapi.CLIAsync{
					Enabled:        true,
					StatusField:    "status",
					TerminalStates: []string{"completed", "failed"},
					Polling: &openapi.PollingConfig{
						Interval: 1,
						Timeout:  10,
					},
				},
			},
			wantErr: false,
		},
		{
			name:        "invalid JSON response",
			initialBody: `invalid json`,
			operation: &openapi.Operation{
				CLIAsync: &openapi.CLIAsync{
					Enabled:        true,
					StatusField:    "status",
					TerminalStates: []string{"completed"},
				},
			},
			wantErr:     true,
			errContains: "failed to parse",
		},
		{
			name:        "missing status field",
			initialBody: `{"id": "123"}`,
			operation: &openapi.Operation{
				CLIAsync: &openapi.CLIAsync{
					Enabled:        true,
					StatusField:    "status",
					TerminalStates: []string{"completed"},
					Polling: &openapi.PollingConfig{
						Interval: 1,
						Timeout:  2,
					},
				},
			},
			wantErr:     true,
			errContains: "status field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &Executor{}
			ctx := context.Background()
			resp := &http.Response{StatusCode: 200}
			body := []byte(tt.initialBody)

			err := executor.handleAsyncOperation(ctx, resp, body, tt.operation, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleAsyncOperation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("handleAsyncOperation() error = %v, should contain %v", err, tt.errContains)
			}
		})
	}
}

func TestExecutor_HandleAsyncOperationWithPolling(t *testing.T) {
	// Test with actual polling
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		if requestCount < 3 {
			_, _ = w.Write([]byte(`{"status": "pending"}`))
		} else {
			_, _ = w.Write([]byte(`{"status": "completed"}`))
		}
	}))
	defer server.Close()

	executor := &Executor{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	operation := &openapi.Operation{
		CLIAsync: &openapi.CLIAsync{
			Enabled:        true,
			StatusField:    "status",
			TerminalStates: []string{"completed"},
			StatusEndpoint: server.URL + "/status",
			Polling: &openapi.PollingConfig{
				Interval: 1,
				Timeout:  10,
			},
		},
	}

	ctx := context.Background()
	resp := &http.Response{StatusCode: 200}
	body := []byte(`{"status": "pending", "id": "123"}`)

	err := executor.handleAsyncOperation(ctx, resp, body, operation, nil)
	if err != nil {
		t.Errorf("handleAsyncOperation() error = %v", err)
	}

	if requestCount < 2 {
		t.Errorf("Expected at least 2 polling requests, got %d", requestCount)
	}
}

func TestExecutor_HandleAsyncOperationTimeout(t *testing.T) {
	executor := &Executor{}

	operation := &openapi.Operation{
		CLIAsync: &openapi.CLIAsync{
			Enabled:        true,
			StatusField:    "status",
			TerminalStates: []string{"completed"},
			Polling: &openapi.PollingConfig{
				Interval: 1,
				Timeout:  1, // Very short timeout
			},
		},
	}

	ctx := context.Background()
	resp := &http.Response{StatusCode: 200}
	body := []byte(`{"status": "pending"}`)

	err := executor.handleAsyncOperation(ctx, resp, body, operation, nil)
	if err == nil {
		t.Error("handleAsyncOperation() should timeout")
	}

	if !strings.Contains(err.Error(), "timed out") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestExecutor_PollStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": "running", "progress": 50}`))
	}))
	defer server.Close()

	executor := &Executor{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	initialResp := map[string]interface{}{"id": "123"}

	result, err := executor.pollStatus(ctx, server.URL, initialResp)
	if err != nil {
		t.Errorf("pollStatus() error = %v", err)
	}

	if result["status"] != "running" {
		t.Errorf("Expected status 'running', got %v", result["status"])
	}
}

func TestExecutor_PollStatusError(t *testing.T) {
	executor := &Executor{
		httpClient: &http.Client{Timeout: 1 * time.Second},
	}

	ctx := context.Background()
	initialResp := map[string]interface{}{}

	// Invalid URL
	_, err := executor.pollStatus(ctx, "://invalid", initialResp)
	if err == nil {
		t.Error("pollStatus() should return error for invalid URL")
	}

	// Non-existent server
	_, err = executor.pollStatus(ctx, "http://localhost:99999/status", initialResp)
	if err == nil {
		t.Error("pollStatus() should return error for unreachable server")
	}
}

func TestExecutor_FormatOutput(t *testing.T) {
	tests := []struct {
		name         string
		body         []byte
		outputFormat string
		hasOutputMgr bool
		wantErr      bool
	}{
		{
			name:         "JSON output",
			body:         []byte(`{"id": "123", "name": "test"}`),
			outputFormat: "json",
			hasOutputMgr: false,
			wantErr:      false,
		},
		{
			name:         "empty response",
			body:         []byte{},
			outputFormat: "json",
			hasOutputMgr: false,
			wantErr:      false,
		},
		{
			name:         "non-JSON response",
			body:         []byte("plain text response"),
			outputFormat: "json",
			hasOutputMgr: false,
			wantErr:      false,
		},
		{
			name:         "with output manager",
			body:         []byte(`{"result": "success"}`),
			outputFormat: "json",
			hasOutputMgr: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &Executor{}
			if tt.hasOutputMgr {
				executor.outputManager = output.NewManager()
			}

			cmd := &cobra.Command{}
			cmd.Flags().String("output", tt.outputFormat, "Output format")

			resp := &http.Response{
				StatusCode: 200,
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}

			var buf bytes.Buffer
			cmd.SetOut(&buf)

			err := executor.formatOutput(cmd, resp, tt.body, &openapi.Operation{})
			if (err != nil) != tt.wantErr {
				t.Errorf("formatOutput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecutor_ExecuteHTTPOperation(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	authMgr := auth.NewManager("test")
	_ = authMgr.RegisterAuthenticator("default", &auth.NoneAuth{})

	config := &ExecutorConfig{
		BaseURL:       server.URL,
		OutputManager: output.NewManager(),
		AuthManager:   authMgr,
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)
	cmd.Flags().String("output", "json", "Output format")

	ctx := context.Background()
	err = executor.executeHTTPOperation(ctx, cmd, listOp, nil)

	if err != nil {
		t.Errorf("executeHTTPOperation() error = %v", err)
	}

	if callCount != 1 {
		t.Errorf("Expected 1 HTTP call, got %d", callCount)
	}
}

func TestExecutor_ExecuteHTTPOperationWithError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message": "Invalid input"}`))
	}))
	defer server.Close()

	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL:       server.URL,
		OutputManager: output.NewManager(),
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)
	cmd.Flags().String("output", "json", "Output format")

	ctx := context.Background()
	err = executor.executeHTTPOperation(ctx, cmd, listOp, nil)

	if err == nil {
		t.Error("executeHTTPOperation() should return error for 400 status")
	}

	if !strings.Contains(err.Error(), "Invalid input") {
		t.Errorf("Error should contain 'Invalid input', got: %v", err)
	}
}

func TestExecutor_ExecuteHTTPOperationWithProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL:       server.URL,
		OutputManager: output.NewManager(),
		ProgressMgr:   progress.NewManager(progress.DefaultConfig()),
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)
	cmd.Flags().String("output", "json", "Output format")

	ctx := context.Background()
	err = executor.executeHTTPOperation(ctx, cmd, listOp, nil)

	if err != nil {
		t.Errorf("executeHTTPOperation() error = %v", err)
	}
}

func TestExecutor_ConvertToWorkflow(t *testing.T) {
	executor := &Executor{}

	cliWorkflow := &openapi.CLIWorkflow{
		Steps: []*openapi.WorkflowStep{
			{
				ID:        "step1",
				Condition: "true",
				Request: &openapi.WorkflowRequest{
					Method:  "POST",
					URL:     "/api/resource",
					Headers: map[string]string{"Content-Type": "application/json"},
					Body:    map[string]interface{}{"key": "value"},
				},
			},
			{
				ID:        "step2",
				Condition: "step1.status == 'success'",
				Request: &openapi.WorkflowRequest{
					Method: "GET",
					URL:    "/api/resource/{{step1.id}}",
				},
			},
		},
	}

	wf, err := executor.convertToWorkflow(cliWorkflow)
	if err != nil {
		t.Errorf("convertToWorkflow() error = %v", err)
	}

	if len(wf.Steps) != 2 {
		t.Errorf("Expected 2 steps, got %d", len(wf.Steps))
	}

	if wf.Steps[0].ID != "step1" {
		t.Errorf("Expected step ID 'step1', got %s", wf.Steps[0].ID)
	}

	if wf.Steps[0].APICall == nil {
		t.Error("Expected APICall to be set")
	}

	if wf.Steps[0].APICall.Method != "POST" {
		t.Errorf("Expected method POST, got %s", wf.Steps[0].APICall.Method)
	}
}

func TestExecutor_ExecuteWorkflow(t *testing.T) {
	// Skip because workflow execution requires complex setup
	t.Skip("Workflow execution requires workflow engine integration")
}

func TestExtractPathParams(t *testing.T) {
	tests := []struct {
		path     string
		expected []string
	}{
		{"/clusters/{id}", []string{"id"}},
		{"/users/{userId}/posts/{postId}", []string{"userId", "postId"}},
		{"/clusters", []string{}},
		{"/api/{version}/resources/{id}", []string{"version", "id"}},
		{"/{a}/{b}/{c}", []string{"a", "b", "c"}},
		{"/no/params/here", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := extractPathParams(tt.path)
			if len(result) != len(tt.expected) {
				t.Errorf("extractPathParams(%s) returned %d params; want %d", tt.path, len(result), len(tt.expected))
				return
			}

			for i, param := range result {
				if param != tt.expected[i] {
					t.Errorf("extractPathParams(%s)[%d] = %s; want %s", tt.path, i, param, tt.expected[i])
				}
			}
		})
	}
}

func TestToFlagName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"userName", "username"},
		{"user_name", "user-name"},
		{"USER_NAME", "user-name"},
		{"userId", "userid"},
		{"user_id", "user-id"},
		{"API_KEY", "api-key"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toFlagName(tt.input)
			if result != tt.expected {
				t.Errorf("toFlagName(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestExecutor_BuildRequestWithHeaders(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL: "https://api.example.com",
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = map[string]string{
		"param:x-custom-header:in": "header",
	}
	cmd.Flags().String("x-custom-header", "", "Custom header")
	_ = cmd.Flags().Set("x-custom-header", "custom-value")

	ctx := context.Background()
	req, err := executor.buildRequest(ctx, cmd, listOp, nil)
	if err != nil {
		t.Fatalf("buildRequest() error = %v", err)
	}

	if req.Header.Get("Accept") != "application/json" {
		t.Error("Expected Accept header to be application/json")
	}
}

func TestExecutor_ExecuteHTTPOperationHTTPError(t *testing.T) {
	// Create executor with client that returns error
	mockClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return nil, io.EOF
	})

	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL:    "https://api.example.com",
		HTTPClient: mockClient,
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)
	cmd.Flags().String("output", "json", "Output format")

	ctx := context.Background()
	err = executor.executeHTTPOperation(ctx, cmd, listOp, nil)

	if err == nil {
		t.Error("executeHTTPOperation() should return error on HTTP failure")
	}

	if !strings.Contains(err.Error(), "HTTP request failed") {
		t.Errorf("Error should mention HTTP request failed, got: %v", err)
	}
}

func TestExecutor_ExecuteHTTPOperationReadBodyError(t *testing.T) {
	// Create executor with client that returns response but body read fails
	mockClient := newMockHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(&errorReader{}),
			Header:     http.Header{"Content-Type": []string{"application/json"}},
		}, nil
	})

	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL:    "https://api.example.com",
		HTTPClient: mockClient,
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)
	cmd.Flags().String("output", "json", "Output format")

	ctx := context.Background()
	err = executor.executeHTTPOperation(ctx, cmd, listOp, nil)

	if err == nil {
		t.Error("executeHTTPOperation() should return error on body read failure")
	}
}

// errorReader is a helper that always returns errors on read
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestExecutor_HandleAsyncOperationContextCanceled(t *testing.T) {
	executor := &Executor{}

	operation := &openapi.Operation{
		CLIAsync: &openapi.CLIAsync{
			Enabled:        true,
			StatusField:    "status",
			TerminalStates: []string{"completed"},
			Polling: &openapi.PollingConfig{
				Interval: 10,
				Timeout:  100,
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	resp := &http.Response{StatusCode: 200}
	body := []byte(`{"status": "pending"}`)

	err := executor.handleAsyncOperation(ctx, resp, body, operation, nil)
	if err == nil {
		t.Error("handleAsyncOperation() should return error on context cancellation")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got: %v", err)
	}
}

func TestExecutor_PollStatusInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	executor := &Executor{
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	ctx := context.Background()
	_, err := executor.pollStatus(ctx, server.URL, map[string]interface{}{})

	if err == nil {
		t.Error("pollStatus() should return error for invalid JSON")
	}
}

func TestExecutor_BuildURLMissingPathParam(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL: "https://api.example.com",
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	var getOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "getUser" {
			getOp = op
			break
		}
	}

	cmd := &cobra.Command{Use: "get"}
	cmd.Annotations = make(map[string]string)

	// No args provided for path parameter
	_, err = executor.buildURL(cmd, getOp, nil)
	if err == nil {
		t.Error("buildURL() should return error for missing path parameter")
	}

	if !strings.Contains(err.Error(), "missing value for path parameter") {
		t.Errorf("Error should mention missing path parameter, got: %v", err)
	}
}

func TestExecutor_FormatOutputWithCLIOutput(t *testing.T) {
	executor := &Executor{
		outputManager: output.NewManager(),
	}

	operation := &openapi.Operation{
		CLIOutput: &openapi.CLIOutput{
			Table: &openapi.TableConfig{
				Columns: []*openapi.TableColumn{
					{Field: "id", Header: "ID"},
					{Field: "name", Header: "Name"},
				},
			},
		},
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("output", "table", "Output format")

	resp := &http.Response{StatusCode: 200}
	body := []byte(`{"id": "123", "name": "test", "extra": "data"}`)

	var buf bytes.Buffer
	cmd.SetOut(&buf)

	err := executor.formatOutput(cmd, resp, body, operation)
	if err != nil {
		t.Errorf("formatOutput() error = %v", err)
	}
}

func TestExecutor_BuildRequestWithEmptyBody(t *testing.T) {
	// Create a spec with request body
	specJSON := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Test", "version": "1.0.0"},
		"paths": {
			"/test": {
				"post": {
					"operationId": "testOp",
					"requestBody": {
						"required": true,
						"content": {
							"application/json": {
								"schema": {"type": "object"}
							}
						}
					},
					"responses": {"200": {"description": "OK"}}
				}
			}
		}
	}`)

	parser := openapi.NewParser()
	spec, err := parser.Parse(context.Background(), specJSON)
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &ExecutorConfig{
		BaseURL: "https://api.example.com",
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	operations, _ := spec.GetOperations()
	if len(operations) == 0 {
		t.Fatal("No operations found")
	}

	op := operations[0]
	cmd := &cobra.Command{Use: "test"}
	cmd.Annotations = make(map[string]string)

	ctx := context.Background()
	_, err = executor.buildRequest(ctx, cmd, op, nil)
	// Should succeed even with no body data
	if err != nil {
		t.Errorf("buildRequest() error = %v", err)
	}
}
