package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/progress"
	"github.com/getkin/kin-openapi/openapi3"
)

// mockAuthenticator is a test authenticator
type mockAuthenticator struct {
	authFunc    func(ctx context.Context) (*auth.Token, error)
	refreshFunc func(ctx context.Context, token *auth.Token) (*auth.Token, error)
	headersFunc func(token *auth.Token) map[string]string
	authType    auth.AuthType
}

func (m *mockAuthenticator) Type() auth.AuthType {
	if m.authType != "" {
		return m.authType
	}
	return auth.AuthTypeOAuth2
}

func (m *mockAuthenticator) Authenticate(ctx context.Context) (*auth.Token, error) {
	if m.authFunc != nil {
		return m.authFunc(ctx)
	}
	return &auth.Token{
		AccessToken: "test-token",
		TokenType:   "Bearer",
	}, nil
}

func (m *mockAuthenticator) RefreshToken(ctx context.Context, token *auth.Token) (*auth.Token, error) {
	if m.refreshFunc != nil {
		return m.refreshFunc(ctx, token)
	}
	return token, nil
}

func (m *mockAuthenticator) GetHeaders(token *auth.Token) map[string]string {
	if m.headersFunc != nil {
		return m.headersFunc(token)
	}
	return map[string]string{
		"Authorization": "Bearer " + token.AccessToken,
	}
}

func (m *mockAuthenticator) Validate() error {
	return nil
}

func TestExecutor_executePreflightCheck(t *testing.T) {
	tests := []struct {
		name       string
		check      *openapi.CLIPreflight
		serverFunc func(w http.ResponseWriter, r *http.Request)
		wantPassed bool
		wantError  bool
	}{
		{
			name: "successful GET check",
			check: &openapi.CLIPreflight{
				Name:        "test-check",
				Description: "Test preflight check",
				Endpoint:    "/api/v1/check",
				Method:      "GET",
				Required:    true,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			},
			wantPassed: true,
			wantError:  false,
		},
		{
			name: "successful POST check",
			check: &openapi.CLIPreflight{
				Name:        "post-check",
				Description: "POST preflight check",
				Endpoint:    "/api/v1/verify",
				Method:      "POST",
				Required:    true,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				// Verify content type
				if ct := r.Header.Get("Content-Type"); ct != "application/json" {
					t.Errorf("Expected Content-Type: application/json, got %s", ct)
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"verified": true}`))
			},
			wantPassed: true,
			wantError:  false,
		},
		{
			name: "default GET when method not specified",
			check: &openapi.CLIPreflight{
				Name:        "default-method",
				Description: "Check with default method",
				Endpoint:    "/api/v1/status",
				Required:    false,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET request (default), got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
			},
			wantPassed: true,
			wantError:  false,
		},
		{
			name: "failed check with error message",
			check: &openapi.CLIPreflight{
				Name:        "fail-check",
				Description: "Check that fails",
				Endpoint:    "/api/v1/fail",
				Method:      "GET",
				Required:    false,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"message": "Validation failed: insufficient quota"}`))
			},
			wantPassed: false,
			wantError:  true,
		},
		{
			name: "failed check with error field",
			check: &openapi.CLIPreflight{
				Name:        "error-field",
				Description: "Check with error field",
				Endpoint:    "/api/v1/check",
				Method:      "GET",
				Required:    false,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"error": "Access denied"}`))
			},
			wantPassed: false,
			wantError:  true,
		},
		{
			name: "failed check with detail field",
			check: &openapi.CLIPreflight{
				Name:        "detail-field",
				Description: "Check with detail field",
				Endpoint:    "/api/v1/check",
				Method:      "GET",
				Required:    false,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"detail": "Invalid configuration"}`))
			},
			wantPassed: false,
			wantError:  true,
		},
		{
			name: "500 server error",
			check: &openapi.CLIPreflight{
				Name:        "server-error",
				Description: "Server error check",
				Endpoint:    "/api/v1/check",
				Method:      "GET",
				Required:    false,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte("Internal server error"))
			},
			wantPassed: false,
			wantError:  true,
		},
		{
			name: "201 created is success",
			check: &openapi.CLIPreflight{
				Name:        "created-check",
				Description: "Check returning 201",
				Endpoint:    "/api/v1/check",
				Method:      "POST",
				Required:    true,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"status": "created"}`))
			},
			wantPassed: true,
			wantError:  false,
		},
		{
			name: "204 no content is success",
			check: &openapi.CLIPreflight{
				Name:        "no-content-check",
				Description: "Check returning 204",
				Endpoint:    "/api/v1/check",
				Method:      "GET",
				Required:    true,
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			wantPassed: true,
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			// Create executor
			spec := &openapi.ParsedSpec{
				Spec: &openapi3.T{
					Servers: openapi3.Servers{
						{URL: server.URL},
					},
				},
			}

			executor, err := NewExecutor(spec, &ExecutorConfig{
				BaseURL:    server.URL,
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			})
			if err != nil {
				t.Fatalf("Failed to create executor: %v", err)
			}

			// Execute preflight check
			result := executor.executePreflightCheck(context.Background(), tt.check)

			// Verify result
			if result.Passed != tt.wantPassed {
				t.Errorf("executePreflightCheck() passed = %v, want %v", result.Passed, tt.wantPassed)
			}

			if (result.Error != nil) != tt.wantError {
				t.Errorf("executePreflightCheck() error = %v, wantError %v", result.Error, tt.wantError)
			}

			if result.Name != tt.check.Name {
				t.Errorf("Result name = %v, want %v", result.Name, tt.check.Name)
			}

			if result.Required != tt.check.Required {
				t.Errorf("Result required = %v, want %v", result.Required, tt.check.Required)
			}
		})
	}
}

func TestExecutor_executePreflightChecks(t *testing.T) {
	tests := []struct {
		name         string
		checks       []*openapi.CLIPreflight
		serverFunc   func(w http.ResponseWriter, r *http.Request)
		wantAllPass  bool
		wantErr      bool
		wantFailures int
	}{
		{
			name:   "no checks",
			checks: []*openapi.CLIPreflight{},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantAllPass:  true,
			wantErr:      false,
			wantFailures: 0,
		},
		{
			name: "all checks pass",
			checks: []*openapi.CLIPreflight{
				{
					Name:        "check1",
					Description: "First check",
					Endpoint:    "/check1",
					Required:    true,
				},
				{
					Name:        "check2",
					Description: "Second check",
					Endpoint:    "/check2",
					Required:    false,
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte(`{"status": "ok"}`))
			},
			wantAllPass:  true,
			wantErr:      false,
			wantFailures: 0,
		},
		{
			name: "optional check fails - continues",
			checks: []*openapi.CLIPreflight{
				{
					Name:        "check1",
					Description: "Required check",
					Endpoint:    "/check1",
					Required:    true,
				},
				{
					Name:        "check2",
					Description: "Optional check",
					Endpoint:    "/check2",
					Required:    false,
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/check1" {
					w.WriteHeader(http.StatusOK)
				} else {
					w.WriteHeader(http.StatusBadRequest)
					_, _ = w.Write([]byte(`{"error": "Optional check failed"}`))
				}
			},
			wantAllPass:  false,
			wantErr:      false,
			wantFailures: 1,
		},
		{
			name: "required check fails - stops",
			checks: []*openapi.CLIPreflight{
				{
					Name:        "check1",
					Description: "Required check",
					Endpoint:    "/check1",
					Required:    true,
				},
				{
					Name:        "check2",
					Description: "Another check",
					Endpoint:    "/check2",
					Required:    false,
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/check1" {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusForbidden)
					_, _ = w.Write([]byte(`{"message": "Insufficient permissions"}`))
				} else {
					w.WriteHeader(http.StatusOK)
				}
			},
			wantAllPass:  false,
			wantErr:      true,
			wantFailures: 1,
		},
		{
			name: "multiple required checks - first fails",
			checks: []*openapi.CLIPreflight{
				{
					Name:        "check1",
					Description: "First required",
					Endpoint:    "/check1",
					Required:    true,
				},
				{
					Name:        "check2",
					Description: "Second required",
					Endpoint:    "/check2",
					Required:    true,
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/check1" {
					w.WriteHeader(http.StatusServiceUnavailable)
				} else {
					w.WriteHeader(http.StatusOK)
				}
			},
			wantAllPass:  false,
			wantErr:      true,
			wantFailures: 1, // Should stop after first required failure
		},
		{
			name: "mixed checks - all pass",
			checks: []*openapi.CLIPreflight{
				{
					Name:        "aws-creds",
					Description: "Verifying AWS credentials",
					Endpoint:    "/api/v1/aws/credentials/verify",
					Method:      "POST",
					Required:    false,
				},
				{
					Name:        "quota",
					Description: "Checking AWS quotas",
					Endpoint:    "/api/v1/aws/quotas/verify",
					Method:      "POST",
					Required:    false,
				},
			},
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if strings.Contains(r.URL.Path, "credentials") {
					_, _ = w.Write([]byte(`{"valid": true}`))
				} else {
					_, _ = w.Write([]byte(`{"sufficient": true}`))
				}
			},
			wantAllPass:  true,
			wantErr:      false,
			wantFailures: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test server
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			// Create executor
			spec := &openapi.ParsedSpec{
				Spec: &openapi3.T{
					Servers: openapi3.Servers{
						{URL: server.URL},
					},
				},
			}

			executor, err := NewExecutor(spec, &ExecutorConfig{
				BaseURL:    server.URL,
				HTTPClient: &http.Client{Timeout: 5 * time.Second},
			})
			if err != nil {
				t.Fatalf("Failed to create executor: %v", err)
			}

			// Execute preflight checks
			results, err := executor.executePreflightChecks(context.Background(), tt.checks)

			// Verify error
			if (err != nil) != tt.wantErr {
				t.Errorf("executePreflightChecks() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify results
			if results.AllPass != tt.wantAllPass {
				t.Errorf("AllPass = %v, want %v", results.AllPass, tt.wantAllPass)
			}

			if len(results.Failed) != tt.wantFailures {
				t.Errorf("Failed count = %v, want %v", len(results.Failed), tt.wantFailures)
			}

			if len(results.Checks) != len(tt.checks) && !tt.wantErr {
				// If no error, all checks should have been executed
				t.Errorf("Checks count = %v, want %v", len(results.Checks), len(tt.checks))
			}
		})
	}
}

func TestExecutor_executePreflightChecksWithAuth(t *testing.T) {
	// Create test server that requires auth
	authToken := "test-token-123"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+authToken {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error": "Unauthorized"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create executor with auth
	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
		},
	}

	authMgr := auth.NewManager("test")
	authenticator := &mockAuthenticator{
		headersFunc: func(token *auth.Token) map[string]string {
			return map[string]string{
				"Authorization": "Bearer " + authToken,
			}
		},
	}
	_ = authMgr.RegisterAuthenticator("default", authenticator)

	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL:     server.URL,
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
		AuthManager: authMgr,
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Execute preflight check
	check := &openapi.CLIPreflight{
		Name:        "auth-check",
		Description: "Check with authentication",
		Endpoint:    "/api/v1/check",
		Method:      "GET",
		Required:    true,
	}

	result := executor.executePreflightCheck(context.Background(), check)

	if !result.Passed {
		t.Errorf("Check with auth failed: %v", result.Error)
	}
}

func TestExecutor_executePreflightChecksWithProgress(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Create executor with progress manager
	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
		},
	}

	progressMgr := progress.NewManager(&progress.Config{
		Type:    progress.TypeSpinner,
		Enabled: false, // Disable for test
	})

	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL:     server.URL,
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
		ProgressMgr: progressMgr,
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	checks := []*openapi.CLIPreflight{
		{
			Name:        "check1",
			Description: "First check",
			Endpoint:    "/check1",
			Required:    true,
		},
	}

	// Execute with progress
	results, err := executor.executePreflightChecksWithProgress(context.Background(), checks)
	if err != nil {
		t.Errorf("executePreflightChecksWithProgress() error = %v", err)
	}

	if !results.AllPass {
		t.Errorf("Expected all checks to pass")
	}
}

func TestExecutor_extractPreflightError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		body           []byte
		expectedErrMsg string
	}{
		{
			name:           "message field",
			statusCode:     400,
			body:           []byte(`{"message": "Invalid request"}`),
			expectedErrMsg: "HTTP 400: Invalid request",
		},
		{
			name:           "error field",
			statusCode:     403,
			body:           []byte(`{"error": "Forbidden"}`),
			expectedErrMsg: "HTTP 403: Forbidden",
		},
		{
			name:           "detail field",
			statusCode:     422,
			body:           []byte(`{"detail": "Validation error"}`),
			expectedErrMsg: "HTTP 422: Validation error",
		},
		{
			name:           "non-JSON body",
			statusCode:     500,
			body:           []byte("Internal server error"),
			expectedErrMsg: "HTTP 500: Internal server error",
		},
		{
			name:           "long non-JSON body",
			statusCode:     500,
			body:           []byte(strings.Repeat("x", 300)),
			expectedErrMsg: "HTTP 500",
		},
		{
			name:           "empty body",
			statusCode:     404,
			body:           []byte{},
			expectedErrMsg: "HTTP 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &Executor{}

			resp := &http.Response{
				StatusCode: tt.statusCode,
			}

			err := executor.extractPreflightError(resp, tt.body)

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Errorf("Error message = %q, want to contain %q", err.Error(), tt.expectedErrMsg)
			}
		})
	}
}

func TestExecutor_preflightCheckRelativeURL(t *testing.T) {
	// Test that relative URLs are properly resolved
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/check" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
		},
	}

	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	check := &openapi.CLIPreflight{
		Name:        "relative-url",
		Description: "Check with relative URL",
		Endpoint:    "/api/v1/check", // Relative URL
		Method:      "GET",
		Required:    true,
	}

	result := executor.executePreflightCheck(context.Background(), check)

	if !result.Passed {
		t.Errorf("Check failed: %v", result.Error)
	}
}

func TestExecutor_preflightCheckAbsoluteURL(t *testing.T) {
	// Test that absolute URLs are used as-is
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{},
	}

	executor, err := NewExecutor(spec, &ExecutorConfig{
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	check := &openapi.CLIPreflight{
		Name:        "absolute-url",
		Description: "Check with absolute URL",
		Endpoint:    server.URL + "/api/v1/check", // Absolute URL
		Method:      "GET",
		Required:    true,
	}

	result := executor.executePreflightCheck(context.Background(), check)

	if !result.Passed {
		t.Errorf("Check failed: %v", result.Error)
	}
}

func TestExecutor_preflightCheckTimeout(t *testing.T) {
	// Create server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
		},
	}

	// Create executor with short timeout
	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 100 * time.Millisecond},
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	check := &openapi.CLIPreflight{
		Name:        "timeout-check",
		Description: "Check that times out",
		Endpoint:    "/api/v1/check",
		Method:      "GET",
		Required:    true,
	}

	result := executor.executePreflightCheck(context.Background(), check)

	if result.Passed {
		t.Error("Expected check to fail due to timeout")
	}

	if result.Error == nil {
		t.Error("Expected error for timeout")
	}
}

// Benchmark preflight checks
func BenchmarkExecutor_executePreflightCheck(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
		},
	}

	executor, _ := NewExecutor(spec, &ExecutorConfig{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})

	check := &openapi.CLIPreflight{
		Name:        "bench-check",
		Description: "Benchmark check",
		Endpoint:    "/api/v1/check",
		Method:      "GET",
		Required:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = executor.executePreflightCheck(context.Background(), check)
	}
}

func TestExecutor_preflightCheckNetworkError(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{},
	}

	// Create executor with mock client that returns network error
	executor := &Executor{
		spec: spec,
		httpClient: &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return nil, fmt.Errorf("network error: connection refused")
				},
			},
		},
		baseURL: "http://localhost:9999",
	}

	check := &openapi.CLIPreflight{
		Name:        "network-error",
		Description: "Check with network error",
		Endpoint:    "/api/v1/check",
		Method:      "GET",
		Required:    true,
	}

	result := executor.executePreflightCheck(context.Background(), check)

	if result.Passed {
		t.Error("Expected check to fail due to network error")
	}

	if result.Error == nil {
		t.Error("Expected error for network failure")
	}

	if !strings.Contains(result.Error.Error(), "request failed") {
		t.Errorf("Error message should mention request failure: %v", result.Error)
	}
}

func TestExecutor_preflightCheckBodyReadError(t *testing.T) {
	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{},
	}

	// Create executor with mock client that returns invalid body
	executor := &Executor{
		spec: spec,
		httpClient: &http.Client{
			Transport: &mockTransport{
				roundTripFunc: func(req *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       io.NopCloser(&errorReader{}),
					}, nil
				},
			},
		},
		baseURL: "http://localhost:9999",
	}

	check := &openapi.CLIPreflight{
		Name:        "read-error",
		Description: "Check with body read error",
		Endpoint:    "/api/v1/check",
		Method:      "GET",
		Required:    true,
	}

	result := executor.executePreflightCheck(context.Background(), check)

	if result.Passed {
		t.Error("Expected check to fail due to read error")
	}

	if result.Error == nil {
		t.Error("Expected error for body read failure")
	}
}
