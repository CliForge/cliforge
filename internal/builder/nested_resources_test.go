package builder

import (
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
)

func TestNestedResourceHandler_ExtractPathParameters(t *testing.T) {
	handler := NewNestedResourceHandler()

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "no parameters",
			path:     "/api/v1/clusters",
			expected: []string{},
		},
		{
			name:     "single parameter",
			path:     "/api/v1/clusters/{cluster_id}",
			expected: []string{"cluster_id"},
		},
		{
			name:     "multiple parameters",
			path:     "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}",
			expected: []string{"cluster_id", "machinepool_id"},
		},
		{
			name:     "camelCase parameters",
			path:     "/api/v1/clusters/{clusterId}/nodes/{nodeId}",
			expected: []string{"clusterId", "nodeId"},
		},
		{
			name:     "mixed format",
			path:     "/api/v1/organizations/{org_id}/clusters/{clusterId}",
			expected: []string{"org_id", "clusterId"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.extractPathParameters(tt.path)

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d parameters, got %d", len(tt.expected), len(result))
				return
			}

			for i, param := range result {
				if param != tt.expected[i] {
					t.Errorf("Parameter %d: expected '%s', got '%s'", i, tt.expected[i], param)
				}
			}
		})
	}
}

func TestNestedResourceHandler_FindParentParameter(t *testing.T) {
	handler := NewNestedResourceHandler()

	tests := []struct {
		name           string
		parentResource string
		pathParams     []string
		expected       string
	}{
		{
			name:           "exact match",
			parentResource: "cluster",
			pathParams:     []string{"cluster", "machinepool_id"},
			expected:       "cluster",
		},
		{
			name:           "match with _id suffix",
			parentResource: "cluster",
			pathParams:     []string{"cluster_id", "machinepool_id"},
			expected:       "cluster_id",
		},
		{
			name:           "match with camelCase Id suffix",
			parentResource: "cluster",
			pathParams:     []string{"clusterId", "machinepoolId"},
			expected:       "clusterId",
		},
		{
			name:           "no match",
			parentResource: "organization",
			pathParams:     []string{"cluster_id", "machinepool_id"},
			expected:       "",
		},
		{
			name:           "match with hyphenated name",
			parentResource: "machine-pool",
			pathParams:     []string{"cluster_id", "machine_pool_id"},
			expected:       "machine_pool_id",
		},
		{
			name:           "prefix match",
			parentResource: "cluster",
			pathParams:     []string{"cluster_uuid", "machinepool_id"},
			expected:       "cluster_uuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.findParentParameter(tt.parentResource, tt.pathParams)

			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNestedResourceHandler_AddParentResourceFlag(t *testing.T) {
	handler := NewNestedResourceHandler()

	tests := []struct {
		name        string
		operation   *openapi.Operation
		expectError bool
		errorMsg    string
		validateFn  func(*testing.T, *cobra.Command)
	}{
		{
			name: "basic nested resource",
			operation: &openapi.Operation{
				Path:         "/api/v1/clusters/{cluster_id}/machine_pools",
				Method:       "POST",
				OperationID:  "createMachinePool",
				CLIParentRes: "cluster",
			},
			expectError: false,
			validateFn: func(t *testing.T, cmd *cobra.Command) {
				// Check flag was added
				flag := cmd.Flags().Lookup("cluster")
				if flag == nil {
					t.Error("Expected --cluster flag to be added")
					return
				}

				// Check flag is required
				annotations := flag.Annotations
				if annotations == nil || annotations[cobra.BashCompOneRequiredFlag] == nil {
					t.Error("Expected --cluster flag to be required")
				}

				// Check annotations
				if cmd.Annotations["parent-resource"] != "cluster" {
					t.Errorf("Expected parent-resource annotation 'cluster', got '%s'", cmd.Annotations["parent-resource"])
				}
				if cmd.Annotations["parent-param"] != "cluster_id" {
					t.Errorf("Expected parent-param annotation 'cluster_id', got '%s'", cmd.Annotations["parent-param"])
				}
			},
		},
		{
			name: "no parent resource",
			operation: &openapi.Operation{
				Path:         "/api/v1/clusters",
				Method:       "GET",
				OperationID:  "listClusters",
				CLIParentRes: "",
			},
			expectError: false,
			validateFn: func(t *testing.T, cmd *cobra.Command) {
				// No flag should be added
				flag := cmd.Flags().Lookup("cluster")
				if flag != nil {
					t.Error("Did not expect any parent resource flag")
				}
			},
		},
		{
			name: "parent resource with no path params",
			operation: &openapi.Operation{
				Path:         "/api/v1/clusters",
				Method:       "GET",
				OperationID:  "listClusters",
				CLIParentRes: "organization",
			},
			expectError: true,
			errorMsg:    "no path parameters",
		},
		{
			name: "parent resource not in path",
			operation: &openapi.Operation{
				Path:         "/api/v1/clusters/{cluster_id}/machine_pools",
				Method:       "GET",
				OperationID:  "listMachinePools",
				CLIParentRes: "organization",
			},
			expectError: true,
			errorMsg:    "not found in path parameters",
		},
		{
			name: "multiple path parameters",
			operation: &openapi.Operation{
				Path:         "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}",
				Method:       "GET",
				OperationID:  "getMachinePool",
				CLIParentRes: "cluster",
			},
			expectError: false,
			validateFn: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("cluster")
				if flag == nil {
					t.Error("Expected --cluster flag to be added")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{
				Use: "test",
			}

			err := handler.AddParentResourceFlag(cmd, tt.operation)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.validateFn != nil {
				tt.validateFn(t, cmd)
			}
		})
	}
}

func TestNestedResourceHandler_SubstituteParentResource(t *testing.T) {
	handler := NewNestedResourceHandler()

	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		path        string
		expected    string
		expectError bool
		errorMsg    string
	}{
		{
			name: "basic substitution",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("cluster", "", "Cluster ID")
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
					"parent-param":    "cluster_id",
				}
				_ = cmd.Flags().Set("cluster", "my-cluster-123")
				return cmd
			},
			path:     "/api/v1/clusters/{cluster_id}/machine_pools",
			expected: "/api/v1/clusters/my-cluster-123/machine_pools",
		},
		{
			name: "multiple parameters with substitution",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("cluster", "", "Cluster ID")
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
					"parent-param":    "cluster_id",
				}
				_ = cmd.Flags().Set("cluster", "cluster-xyz")
				return cmd
			},
			path:     "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}",
			expected: "/api/v1/clusters/cluster-xyz/machine_pools/{machinepool_id}",
		},
		{
			name: "no parent resource annotation",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{Use: "test"}
			},
			path:     "/api/v1/clusters",
			expected: "/api/v1/clusters",
		},
		{
			name: "missing parent-param annotation",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
				}
				return cmd
			},
			path:        "/api/v1/clusters/{cluster_id}/machine_pools",
			expectError: true,
			errorMsg:    "no parent-param annotation",
		},
		{
			name: "empty parent ID",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("cluster", "", "Cluster ID")
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
					"parent-param":    "cluster_id",
				}
				// Don't set the flag value
				return cmd
			},
			path:        "/api/v1/clusters/{cluster_id}/machine_pools",
			expectError: true,
			errorMsg:    "is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()

			result, err := handler.SubstituteParentResource(cmd, tt.path)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result != tt.expected {
				t.Errorf("Expected path '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNestedResourceHandler_GetNestedResourceLevel(t *testing.T) {
	handler := NewNestedResourceHandler()

	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{
			name:     "top-level resource",
			path:     "/api/v1/clusters",
			expected: 0,
		},
		{
			name:     "top-level resource with ID",
			path:     "/api/v1/clusters/{cluster_id}",
			expected: 0,
		},
		{
			name:     "nested resource level 1",
			path:     "/api/v1/clusters/{cluster_id}/machine_pools",
			expected: 1,
		},
		{
			name:     "nested resource level 1 with ID",
			path:     "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}",
			expected: 1,
		},
		{
			name:     "nested resource level 2",
			path:     "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}/nodes",
			expected: 2,
		},
		{
			name:     "nested resource level 2 with ID",
			path:     "/api/v1/clusters/{cluster_id}/machine_pools/{machinepool_id}/nodes/{node_id}",
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.GetNestedResourceLevel(tt.path)

			if result != tt.expected {
				t.Errorf("Expected level %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestNestedResourceHandler_ValidateNestedResourceFlags(t *testing.T) {
	handler := NewNestedResourceHandler()

	tests := []struct {
		name        string
		setupCmd    func() *cobra.Command
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid parent resource flag",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("cluster", "", "Cluster ID")
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
				}
				_ = cmd.Flags().Set("cluster", "my-cluster")
				return cmd
			},
			expectError: false,
		},
		{
			name: "no parent resource",
			setupCmd: func() *cobra.Command {
				return &cobra.Command{Use: "test"}
			},
			expectError: false,
		},
		{
			name: "parent flag not set",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("cluster", "", "Cluster ID")
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
				}
				// Don't set the flag
				return cmd
			},
			expectError: true,
			errorMsg:    "not set",
		},
		{
			name: "parent flag empty",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				cmd.Flags().String("cluster", "", "Cluster ID")
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
				}
				_ = cmd.Flags().Set("cluster", "")
				return cmd
			},
			expectError: true,
			errorMsg:    "cannot be empty",
		},
		{
			name: "parent flag missing",
			setupCmd: func() *cobra.Command {
				cmd := &cobra.Command{Use: "test"}
				// Don't add the flag
				cmd.Annotations = map[string]string{
					"parent-resource": "cluster",
				}
				return cmd
			},
			expectError: true,
			errorMsg:    "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.setupCmd()

			err := handler.ValidateNestedResourceFlags(cmd)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
					return
				}
				if tt.errorMsg != "" && !contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error to contain '%s', got '%s'", tt.errorMsg, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestNestedResourceHandler_Integration(t *testing.T) {
	// Integration test simulating real usage
	handler := NewNestedResourceHandler()

	// Create an operation for creating a machine pool
	op := &openapi.Operation{
		Path:         "/api/v1/clusters/{cluster_id}/machine_pools",
		Method:       "POST",
		OperationID:  "createMachinePool",
		CLIParentRes: "cluster",
		Operation:    &openapi3.Operation{},
	}

	// Create command
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a machine pool",
	}

	// Add parent resource flag
	err := handler.AddParentResourceFlag(cmd, op)
	if err != nil {
		t.Fatalf("Failed to add parent resource flag: %v", err)
	}

	// Verify flag was added and is required
	flag := cmd.Flags().Lookup("cluster")
	if flag == nil {
		t.Fatal("Expected --cluster flag to be added")
	}

	// Set the flag
	err = cmd.Flags().Set("cluster", "my-test-cluster")
	if err != nil {
		t.Fatalf("Failed to set flag: %v", err)
	}

	// Validate flags
	err = handler.ValidateNestedResourceFlags(cmd)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Substitute parent resource in path
	substituted, err := handler.SubstituteParentResource(cmd, op.Path)
	if err != nil {
		t.Fatalf("Failed to substitute parent resource: %v", err)
	}

	expected := "/api/v1/clusters/my-test-cluster/machine_pools"
	if substituted != expected {
		t.Errorf("Expected path '%s', got '%s'", expected, substituted)
	}

	// Check nesting level
	level := handler.GetNestedResourceLevel(op.Path)
	if level != 1 {
		t.Errorf("Expected nesting level 1, got %d", level)
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) &&
		(s == substr || len(s) > len(substr) &&
			(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
				containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
