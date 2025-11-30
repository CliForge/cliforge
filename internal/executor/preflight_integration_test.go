package executor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

// TestPreflightIntegration tests the complete preflight check flow
func TestPreflightIntegration(t *testing.T) {
	// Track which endpoints were called
	calledEndpoints := make(map[string]int)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledEndpoints[r.URL.Path]++

		switch r.URL.Path {
		case "/api/v1/aws/credentials/verify":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"valid": true}`))
		case "/api/v1/aws/quotas/verify":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"sufficient": true}`))
		case "/api/v1/clusters":
			// Main operation endpoint - should only be called after preflight checks pass
			if calledEndpoints["/api/v1/aws/credentials/verify"] == 0 || calledEndpoints["/api/v1/aws/quotas/verify"] == 0 {
				t.Error("Main operation called before preflight checks completed")
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "cluster-123", "status": "creating"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create parsed spec with preflight checks
	createOp := &openapi3.Operation{
		Summary:     "Create cluster",
		OperationID: "createCluster",
		Extensions: map[string]interface{}{
			"x-cli-preflight": []interface{}{
				map[string]interface{}{
					"name":        "verify-aws-credentials",
					"description": "Verifying AWS credentials...",
					"endpoint":    "/api/v1/aws/credentials/verify",
					"method":      "POST",
					"required":    false,
				},
				map[string]interface{}{
					"name":        "verify-quota",
					"description": "Checking AWS service quotas...",
					"endpoint":    "/api/v1/aws/quotas/verify",
					"method":      "POST",
					"required":    false,
				},
			},
		},
	}

	paths := openapi3.NewPaths()
	paths.Set("/api/v1/clusters", &openapi3.PathItem{
		Post: createOp,
	})

	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
			Paths: paths,
		},
	}

	// Create executor
	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Create command
	cmd := &cobra.Command{
		Use: "create",
	}
	cmd.SetContext(context.Background())
	cmd.Annotations = map[string]string{
		"operationID": "createCluster",
	}
	cmd.Flags().String("output", "json", "Output format")

	// Execute command (this should run preflight checks, then the main operation)
	err = executor.Execute(cmd, []string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify all endpoints were called
	expectedCalls := map[string]int{
		"/api/v1/aws/credentials/verify": 1,
		"/api/v1/aws/quotas/verify":      1,
		"/api/v1/clusters":               1,
	}

	for endpoint, expectedCount := range expectedCalls {
		if calledEndpoints[endpoint] != expectedCount {
			t.Errorf("Endpoint %s called %d times, expected %d", endpoint, calledEndpoints[endpoint], expectedCount)
		}
	}
}

// TestPreflightIntegrationWithRequiredFailure tests that required check failure blocks operation
func TestPreflightIntegrationWithRequiredFailure(t *testing.T) {
	// Track which endpoints were called
	calledEndpoints := make(map[string]int)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledEndpoints[r.URL.Path]++

		switch r.URL.Path {
		case "/api/v1/aws/credentials/verify":
			// This required check fails
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			_, _ = w.Write([]byte(`{"message": "Invalid AWS credentials"}`))
		case "/api/v1/aws/quotas/verify":
			// This should not be reached
			w.WriteHeader(http.StatusOK)
		case "/api/v1/clusters":
			// This should not be reached
			t.Error("Main operation called despite required preflight check failure")
			w.WriteHeader(http.StatusCreated)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create parsed spec with required preflight check
	createOp := &openapi3.Operation{
		Summary:     "Create cluster",
		OperationID: "createCluster",
		Extensions: map[string]interface{}{
			"x-cli-preflight": []interface{}{
				map[string]interface{}{
					"name":        "verify-aws-credentials",
					"description": "Verifying AWS credentials...",
					"endpoint":    "/api/v1/aws/credentials/verify",
					"method":      "POST",
					"required":    true, // Required check
				},
				map[string]interface{}{
					"name":        "verify-quota",
					"description": "Checking AWS service quotas...",
					"endpoint":    "/api/v1/aws/quotas/verify",
					"method":      "POST",
					"required":    false,
				},
			},
		},
	}

	paths := openapi3.NewPaths()
	paths.Set("/api/v1/clusters", &openapi3.PathItem{
		Post: createOp,
	})

	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
			Paths: paths,
		},
	}

	// Create executor
	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Create command
	cmd := &cobra.Command{
		Use: "create",
	}
	cmd.SetContext(context.Background())
	cmd.Annotations = map[string]string{
		"operationID": "createCluster",
	}
	cmd.Flags().String("output", "json", "Output format")

	// Execute command - should fail at preflight check
	err = executor.Execute(cmd, []string{})
	if err == nil {
		t.Fatal("Expected error due to required preflight check failure, got nil")
	}

	// Verify only the first check was called
	if calledEndpoints["/api/v1/aws/credentials/verify"] != 1 {
		t.Errorf("Expected credentials check to be called once, got %d", calledEndpoints["/api/v1/aws/credentials/verify"])
	}

	// Verify subsequent checks and main operation were not called
	if calledEndpoints["/api/v1/aws/quotas/verify"] != 0 {
		t.Error("Quota check should not have been called after required check failed")
	}

	if calledEndpoints["/api/v1/clusters"] != 0 {
		t.Error("Main operation should not have been called after required check failed")
	}
}

// TestPreflightIntegrationWithOptionalFailure tests that optional check failure allows operation
func TestPreflightIntegrationWithOptionalFailure(t *testing.T) {
	// Track which endpoints were called
	calledEndpoints := make(map[string]int)

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calledEndpoints[r.URL.Path]++

		switch r.URL.Path {
		case "/api/v1/aws/credentials/verify":
			// This optional check fails
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"message": "AWS API temporarily unavailable"}`))
		case "/api/v1/clusters":
			// Main operation should still proceed
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id": "cluster-123", "status": "creating"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create parsed spec with optional preflight check
	createOp := &openapi3.Operation{
		Summary:     "Create cluster",
		OperationID: "createCluster",
		Extensions: map[string]interface{}{
			"x-cli-preflight": []interface{}{
				map[string]interface{}{
					"name":        "verify-aws-credentials",
					"description": "Verifying AWS credentials...",
					"endpoint":    "/api/v1/aws/credentials/verify",
					"method":      "POST",
					"required":    false, // Optional check
				},
			},
		},
	}

	paths := openapi3.NewPaths()
	paths.Set("/api/v1/clusters", &openapi3.PathItem{
		Post: createOp,
	})

	spec := &openapi.ParsedSpec{
		Spec: &openapi3.T{
			Servers: openapi3.Servers{
				{URL: server.URL},
			},
			Paths: paths,
		},
	}

	// Create executor
	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL:    server.URL,
		HTTPClient: &http.Client{Timeout: 5 * time.Second},
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	// Create command
	cmd := &cobra.Command{
		Use: "create",
	}
	cmd.SetContext(context.Background())
	cmd.Annotations = map[string]string{
		"operationID": "createCluster",
	}
	cmd.Flags().String("output", "json", "Output format")

	// Execute command - should succeed despite optional check failure
	err = executor.Execute(cmd, []string{})
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Verify both endpoints were called
	if calledEndpoints["/api/v1/aws/credentials/verify"] != 1 {
		t.Errorf("Expected credentials check to be called once, got %d", calledEndpoints["/api/v1/aws/credentials/verify"])
	}

	if calledEndpoints["/api/v1/clusters"] != 1 {
		t.Errorf("Expected main operation to be called once despite optional check failure, got %d", calledEndpoints["/api/v1/clusters"])
	}
}
