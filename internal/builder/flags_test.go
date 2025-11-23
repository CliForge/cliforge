package builder

import (
	"context"
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

func TestFlagBuilder_AddOperationFlags(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Get listUsers operation
	operations, err := spec.GetOperations()
	if err != nil {
		t.Fatalf("Failed to get operations: %v", err)
	}

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

	// Create flag builder
	flagBuilder := NewFlagBuilder(spec.Extensions.Config)

	// Create test command
	cmd := &cobra.Command{
		Use: "list",
	}

	// Add operation flags
	if err := flagBuilder.AddOperationFlags(cmd, listOp); err != nil {
		t.Fatalf("Failed to add operation flags: %v", err)
	}

	// Verify limit flag was added (query parameter)
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("Expected 'limit' flag not found")
	}
}

func TestFlagBuilder_AddGlobalFlags(t *testing.T) {
	flagBuilder := NewFlagBuilder(nil)
	cmd := &cobra.Command{Use: "root"}

	flagBuilder.AddGlobalFlags(cmd)

	// Verify global flags exist
	expectedFlags := []string{"output", "verbose", "no-color", "config", "profile", "dry-run", "debug", "interactive"}
	for _, flagName := range expectedFlags {
		flag := cmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected global flag '%s' not found", flagName)
		}
	}
}

func TestBuildRequestParams(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Annotations = make(map[string]string)

	// Add a parameter flag
	cmd.Flags().String("limit", "100", "Limit")
	cmd.Annotations["param:limit"] = "limit"
	cmd.Annotations["param:limit:in"] = "query"

	// Set the flag
	cmd.Flags().Set("limit", "50")

	// Build params
	params, err := BuildRequestParams(cmd)
	if err != nil {
		t.Fatalf("Failed to build request params: %v", err)
	}

	// Verify param was extracted
	if val, ok := params["limit"]; !ok {
		t.Error("Expected 'limit' parameter not found")
	} else if val != "50" {
		t.Errorf("Expected limit value '50', got '%v'", val)
	}
}

func TestBuildRequestBody(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Annotations = make(map[string]string)

	// Add body field flags
	cmd.Flags().String("cluster-name", "", "Cluster name")
	cmd.Flags().String("region", "", "Region")
	cmd.Annotations["body:cluster-name"] = "name"
	cmd.Annotations["body:region"] = "region"

	// Set the flags
	cmd.Flags().Set("cluster-name", "my-cluster")
	cmd.Flags().Set("region", "us-east-1")

	// Build body
	body, err := BuildRequestBody(cmd)
	if err != nil {
		t.Fatalf("Failed to build request body: %v", err)
	}

	// Verify body fields
	if val, ok := body["name"]; !ok {
		t.Error("Expected 'name' field not found")
	} else if val != "my-cluster" {
		t.Errorf("Expected name value 'my-cluster', got '%v'", val)
	}

	if val, ok := body["region"]; !ok {
		t.Error("Expected 'region' field not found")
	} else if val != "us-east-1" {
		t.Errorf("Expected region value 'us-east-1', got '%v'", val)
	}
}

func TestToFlagName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"clusterName", "clustername"},
		{"cluster_name", "cluster-name"},
		{"CLUSTER NAME", "cluster-name"},
		{"multi-AZ", "multi-az"},
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

func TestValidateEnumFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	// Add enum flag
	cmd.Flags().String("state", "", "State")
	cmd.Flags().SetAnnotation("state", "enum", []string{"pending", "ready", "error"})

	// Test valid value
	cmd.Flags().Set("state", "ready")
	if err := ValidateEnumFlags(cmd); err != nil {
		t.Errorf("Expected validation to pass for valid enum value, got error: %v", err)
	}

	// Test invalid value
	cmd.Flags().Set("state", "invalid")
	if err := ValidateEnumFlags(cmd); err == nil {
		t.Error("Expected validation to fail for invalid enum value")
	}
}
