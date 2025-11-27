package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/output"
	"github.com/CliForge/cliforge/pkg/workflow"
)

// BenchmarkHTTPRequestSimple benchmarks a simple HTTP GET request
func BenchmarkHTTPRequestSimple(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := server.Client()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("failed to make request: %v", err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// BenchmarkHTTPRequestWithAuth benchmarks HTTP request with authentication
func BenchmarkHTTPRequestWithAuth(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	apiKeyAuth, err := auth.NewAPIKeyAuth(&auth.APIKeyConfig{
		Key:      "test-key",
		Location: auth.APIKeyLocationHeader,
		Name:     "Authorization",
		Prefix:   "Bearer",
	})
	if err != nil {
		b.Fatalf("failed to create auth: %v", err)
	}

	client := auth.NewAuthenticatedClient(server.Client(), apiKeyAuth, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequestWithContext(ctx, "GET", server.URL, nil)
		resp, err := client.Do(req)
		if err != nil {
			b.Fatalf("failed to make request: %v", err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// BenchmarkHTTPRequestLargePayload benchmarks HTTP request with large JSON payload
func BenchmarkHTTPRequestLargePayload(b *testing.B) {
	// Generate large payload
	largePayload := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largePayload[fmt.Sprintf("field%d", i)] = map[string]interface{}{
			"id":    i,
			"name":  fmt.Sprintf("Item %d", i),
			"value": i * 100,
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(largePayload)
	}))
	defer server.Close()

	client := server.Client()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("failed to make request: %v", err)
		}
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}
}

// BenchmarkOutputFormatJSON benchmarks JSON output formatting
func BenchmarkOutputFormatJSON(b *testing.B) {
	data := generateTestData(100)
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := manager.Format(buf, data, "json")
		if err != nil {
			b.Fatalf("failed to format JSON: %v", err)
		}
	}
}

// BenchmarkOutputFormatYAML benchmarks YAML output formatting
func BenchmarkOutputFormatYAML(b *testing.B) {
	data := generateTestData(100)
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := manager.Format(buf, data, "yaml")
		if err != nil {
			b.Fatalf("failed to format YAML: %v", err)
		}
	}
}

// BenchmarkOutputFormatTable benchmarks table output formatting
func BenchmarkOutputFormatTable(b *testing.B) {
	data := generateTestData(100)
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := manager.Format(buf, data, "table")
		if err != nil {
			b.Fatalf("failed to format table: %v", err)
		}
	}
}

// BenchmarkOutputFormatJSONLarge benchmarks JSON formatting with large dataset
func BenchmarkOutputFormatJSONLarge(b *testing.B) {
	data := generateTestData(1000)
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := manager.Format(buf, data, "json")
		if err != nil {
			b.Fatalf("failed to format JSON: %v", err)
		}
	}
}

// BenchmarkAuthTokenHandling benchmarks authentication token handling
func BenchmarkAuthTokenHandling(b *testing.B) {
	ctx := context.Background()
	manager := auth.NewManager("test-cli")

	// Create and register an API key authenticator
	apiKeyAuth, _ := auth.NewAPIKeyAuth(&auth.APIKeyConfig{
		Key:      "test-key",
		Location: auth.APIKeyLocationHeader,
		Name:     "Authorization",
		Prefix:   "Bearer",
	})
	_ = manager.RegisterAuthenticator("apikey", apiKeyAuth)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := manager.GetToken(ctx, "apikey")
		if err != nil {
			b.Fatalf("failed to get token: %v", err)
		}
	}
}

// BenchmarkWorkflowExecutionSimple benchmarks simple workflow execution
func BenchmarkWorkflowExecutionSimple(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:       "step1",
				Type:     workflow.StepTypeAPICall,
				Required: true,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL,
				},
			},
		},
	}

	executor, err := workflow.NewExecutor(wf, server.Client(), nil)
	if err != nil {
		b.Fatalf("failed to create executor: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := workflow.NewExecutionContext(nil)
		_, err := executor.Execute(ctx)
		if err != nil {
			b.Fatalf("failed to execute workflow: %v", err)
		}
	}
}

// BenchmarkWorkflowExecutionMultiStep benchmarks multi-step workflow execution
func BenchmarkWorkflowExecutionMultiStep(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "123",
			"name": "Test",
		})
	}))
	defer server.Close()

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:       "step1",
				Type:     workflow.StepTypeAPICall,
				Required: true,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/users",
				},
			},
			{
				ID:        "step2",
				Type:      workflow.StepTypeAPICall,
				DependsOn: []string{"step1"},
				Required:  true,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/users/123",
				},
			},
			{
				ID:        "step3",
				Type:      workflow.StepTypeAPICall,
				DependsOn: []string{"step2"},
				Required:  true,
				APICall: &workflow.APICallStep{
					Method:   "POST",
					Endpoint: server.URL + "/users",
				},
			},
		},
	}

	executor, err := workflow.NewExecutor(wf, server.Client(), nil)
	if err != nil {
		b.Fatalf("failed to create executor: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := workflow.NewExecutionContext(nil)
		_, err := executor.Execute(ctx)
		if err != nil {
			b.Fatalf("failed to execute workflow: %v", err)
		}
	}
}

// BenchmarkWorkflowExecutionParallel benchmarks parallel workflow execution
func BenchmarkWorkflowExecutionParallel(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate some processing time
		time.Sleep(10 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	wf := &workflow.Workflow{
		Settings: &workflow.Settings{
			ParallelExecution: true,
		},
		Steps: []*workflow.Step{
			{
				ID:   "step1",
				Type: workflow.StepTypeAPICall,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/endpoint1",
				},
			},
			{
				ID:   "step2",
				Type: workflow.StepTypeAPICall,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/endpoint2",
				},
			},
			{
				ID:   "step3",
				Type: workflow.StepTypeAPICall,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/endpoint3",
				},
			},
		},
	}

	executor, err := workflow.NewExecutor(wf, server.Client(), nil)
	if err != nil {
		b.Fatalf("failed to create executor: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := workflow.NewExecutionContext(nil)
		_, err := executor.Execute(ctx)
		if err != nil {
			b.Fatalf("failed to execute workflow: %v", err)
		}
	}
}

// Helper function to generate test data
func generateTestData(count int) []map[string]interface{} {
	data := make([]map[string]interface{}, count)
	for i := 0; i < count; i++ {
		data[i] = map[string]interface{}{
			"id":        i,
			"name":      fmt.Sprintf("Item %d", i),
			"value":     i * 100,
			"active":    i%2 == 0,
			"timestamp": time.Now().Unix(),
		}
	}
	return data
}
