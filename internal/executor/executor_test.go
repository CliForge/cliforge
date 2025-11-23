package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/output"
	"github.com/spf13/cobra"
)

func TestExecutor_BuildURL(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Create executor
	config := &ExecutorConfig{
		BaseURL: "https://api.example.com",
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Get an operation
	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	if listOp == nil {
		t.Fatal("listUsers operation not found")
	}

	// Create test command with query parameter
	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)
	cmd.Flags().Int("limit", 100, "Limit")
	cmd.Annotations["param:limit"] = "limit"
	cmd.Annotations["param:limit:in"] = "query"
	cmd.Flags().Set("limit", "50")

	// Build URL
	url, err := executor.buildURL(cmd, listOp, nil)
	if err != nil {
		t.Fatalf("Failed to build URL: %v", err)
	}

	expectedURL := "https://api.example.com/v1/users?limit=50"
	if url != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, url)
	}
}

func TestExecutor_BuildURLWithPathParams(t *testing.T) {
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

	// Get getUser operation
	operations, _ := spec.GetOperations()
	var getOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "getUser" {
			getOp = op
			break
		}
	}

	if getOp == nil {
		t.Fatal("getUser operation not found")
	}

	// Create command with path parameter in args
	cmd := &cobra.Command{Use: "get"}
	cmd.Annotations = make(map[string]string)
	args := []string{"user-123"}

	// Build URL
	url, err := executor.buildURL(cmd, getOp, args)
	if err != nil {
		t.Fatalf("Failed to build URL: %v", err)
	}

	expectedURL := "https://api.example.com/v1/users/user-123"
	if url != expectedURL {
		t.Errorf("Expected URL '%s', got '%s'", expectedURL, url)
	}
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

func TestExecutor_ExecuteHTTPOperation(t *testing.T) {
	// Create mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"items": [{"id": "cluster-1", "name": "test-cluster"}]}`))
	}))
	defer server.Close()

	// Parse spec
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Create executor
	config := &ExecutorConfig{
		BaseURL:       server.URL,
		OutputManager: output.NewManager(),
		AuthManager:   auth.NewManager("test"),
	}
	executor, err := NewExecutor(spec, config)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Get operation
	operations, _ := spec.GetOperations()
	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	// Create command
	cmd := &cobra.Command{Use: "list"}
	cmd.Annotations = make(map[string]string)
	cmd.Flags().String("output", "json", "Output format")

	// Execute operation
	ctx := context.Background()
	err = executor.executeHTTPOperation(ctx, cmd, listOp, nil)
	if err != nil {
		t.Errorf("Expected successful execution, got error: %v", err)
	}
}

func TestConvertToWorkflow(t *testing.T) {
	// Skip test as swagger2 example doesn't have workflow
	t.Skip("Swagger2 example doesn't have workflow - would need YAML support")
}
