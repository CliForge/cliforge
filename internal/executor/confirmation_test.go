package executor

import (
	"context"
	"strings"
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

func TestExecutor_CheckConfirmation(t *testing.T) {
	tests := []struct {
		name           string
		confirmation   *openapi.CLIConfirmation
		setupFlags     func(*cobra.Command)
		expectProceed  bool
		expectPrompted bool
	}{
		{
			name:           "no confirmation config",
			confirmation:   nil,
			setupFlags:     func(cmd *cobra.Command) {},
			expectProceed:  true,
			expectPrompted: false,
		},
		{
			name: "confirmation disabled",
			confirmation: &openapi.CLIConfirmation{
				Enabled: false,
				Message: "Are you sure?",
			},
			setupFlags:     func(cmd *cobra.Command) {},
			expectProceed:  true,
			expectPrompted: false,
		},
		{
			name: "bypass with yes flag",
			confirmation: &openapi.CLIConfirmation{
				Enabled: true,
				Message: "Delete resource?",
				Flag:    "--yes",
			},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Bool("yes", false, "Bypass confirmation")
				_ = cmd.Flags().Set("yes", "true")
			},
			expectProceed:  true,
			expectPrompted: false,
		},
		{
			name: "bypass with custom flag name",
			confirmation: &openapi.CLIConfirmation{
				Enabled: true,
				Message: "Delete resource?",
				Flag:    "--force",
			},
			setupFlags: func(cmd *cobra.Command) {
				cmd.Flags().Bool("force", false, "Force operation")
				_ = cmd.Flags().Set("force", "true")
			},
			expectProceed:  true,
			expectPrompted: false,
		},
		// Note: Interactive prompt test case removed as it blocks in CI/non-TTY environments
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor
			parser := openapi.NewParser()
			spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
			if err != nil {
				t.Fatalf("Failed to parse spec: %v", err)
			}

			executor, err := NewExecutor(spec, &ExecutorConfig{
				BaseURL: "https://api.example.com",
			})
			if err != nil {
				t.Fatalf("Failed to create executor: %v", err)
			}

			// Create command with flags
			cmd := &cobra.Command{
				Use: "test",
			}
			tt.setupFlags(cmd)

			// Create operation
			op := &openapi.Operation{
				CLIConfirmation: tt.confirmation,
			}

			// Check confirmation
			proceed, err := executor.CheckConfirmation(cmd, op)

			if err != nil {
				t.Fatalf("CheckConfirmation() unexpected error: %v", err)
			}

			if proceed != tt.expectProceed {
				t.Errorf("CheckConfirmation() proceed = %v, want %v", proceed, tt.expectProceed)
			}
		})
	}
}

func TestExecutor_SubstituteParameters(t *testing.T) {
	tests := []struct {
		name     string
		message  string
		flags    map[string]string
		expected string
	}{
		{
			name:     "no substitution needed",
			message:  "Are you sure you want to proceed?",
			flags:    map[string]string{},
			expected: "Are you sure you want to proceed?",
		},
		{
			name:    "single parameter substitution",
			message: "Delete cluster '{cluster-id}'?",
			flags: map[string]string{
				"cluster-id": "my-cluster",
			},
			expected: "Delete cluster 'my-cluster'?",
		},
		{
			name:    "multiple parameter substitution",
			message: "Delete {count} resources in '{region}'?",
			flags: map[string]string{
				"count":  "5",
				"region": "us-east-1",
			},
			expected: "Delete 5 resources in 'us-east-1'?",
		},
		{
			name:    "camelCase parameter name",
			message: "Delete cluster '{clusterId}'?",
			flags: map[string]string{
				"cluster-id": "my-cluster",
			},
			expected: "Delete cluster 'my-cluster'?",
		},
		{
			name:    "snake_case parameter name",
			message: "Delete cluster '{cluster_id}'?",
			flags: map[string]string{
				"cluster-id": "my-cluster",
			},
			expected: "Delete cluster 'my-cluster'?",
		},
		{
			name:     "parameter not set",
			message:  "Delete cluster '{cluster-id}'?",
			flags:    map[string]string{},
			expected: "Delete cluster '{cluster-id}'?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create executor
			parser := openapi.NewParser()
			spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
			if err != nil {
				t.Fatalf("Failed to parse spec: %v", err)
			}

			executor, err := NewExecutor(spec, &ExecutorConfig{
				BaseURL: "https://api.example.com",
			})
			if err != nil {
				t.Fatalf("Failed to create executor: %v", err)
			}

			// Create command with flags
			cmd := &cobra.Command{
				Use: "test",
			}

			for name, value := range tt.flags {
				cmd.Flags().String(name, "", "")
				_ = cmd.Flags().Set(name, value)
			}

			// Substitute parameters
			result := executor.substituteParameters(cmd, tt.message)

			if result != tt.expected {
				t.Errorf("substituteParameters() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCaseConversions(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCamel string
		expectedSnake string
	}{
		{
			name:          "kebab-case",
			input:         "cluster-id",
			expectedCamel: "clusterId",
			expectedSnake: "cluster_id",
		},
		{
			name:          "snake_case",
			input:         "cluster_id",
			expectedCamel: "clusterId",
			expectedSnake: "cluster_id",
		},
		{
			name:          "single word",
			input:         "cluster",
			expectedCamel: "cluster",
			expectedSnake: "cluster",
		},
		{
			name:          "multiple words",
			input:         "very-long-parameter-name",
			expectedCamel: "veryLongParameterName",
			expectedSnake: "very_long_parameter_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			camel := toCamelCase(tt.input)
			if camel != tt.expectedCamel {
				t.Errorf("toCamelCase(%q) = %q, want %q", tt.input, camel, tt.expectedCamel)
			}

			snake := toSnakeCase(tt.input)
			if snake != tt.expectedSnake {
				t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, snake, tt.expectedSnake)
			}
		})
	}
}

func TestShowConfirmationPrompt_NonInteractive(t *testing.T) {
	t.Skip("Skipping interactive prompt test - requires TTY")
	// This test verifies the function signature and basic error handling
	// Interactive tests would require a TTY, which isn't available in CI

	message := "Are you sure you want to delete this resource?"

	// In non-interactive environment, this should fail gracefully
	_, err := ShowConfirmationPrompt(message)
	if err != nil {
		// Expected - no TTY available
		if !strings.Contains(err.Error(), "confirmation") {
			t.Errorf("Expected confirmation-related error, got: %v", err)
		}
	}
}

func TestShowExplicitConfirmationPrompt_NonInteractive(t *testing.T) {
	t.Skip("Skipping interactive prompt test - requires TTY")
	// This test verifies the function signature and basic error handling
	// Interactive tests would require a TTY, which isn't available in CI

	message := "⚠️  This will permanently delete all data. Type 'yes' to confirm."

	// In non-interactive environment, this should fail gracefully
	_, err := ShowExplicitConfirmationPrompt(message)
	if err != nil {
		// Expected - no TTY available
		if !strings.Contains(err.Error(), "confirmation") {
			t.Errorf("Expected confirmation-related error, got: %v", err)
		}
	}
}

// TestConfirmationIntegration tests the confirmation flow in a realistic scenario
func TestConfirmationIntegration(t *testing.T) {
	// Parse spec with confirmation extension
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Get operations
	operations, err := spec.GetOperations()
	if err != nil {
		t.Fatalf("Failed to get operations: %v", err)
	}

	// Find delete operation (which should have confirmation)
	var deleteOp *openapi.Operation
	for _, op := range operations {
		if strings.Contains(strings.ToLower(op.OperationID), "delete") {
			deleteOp = op
			break
		}
	}

	if deleteOp == nil {
		t.Skip("No delete operation found in test spec")
	}

	// Create executor
	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL: "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	t.Run("with bypass flag", func(t *testing.T) {
		cmd := &cobra.Command{Use: "delete"}
		cmd.Flags().Bool("yes", false, "Bypass confirmation")
		_ = cmd.Flags().Set("yes", "true")

		proceed, err := executor.CheckConfirmation(cmd, deleteOp)
		if err != nil {
			t.Fatalf("CheckConfirmation() error: %v", err)
		}

		if !proceed {
			t.Error("Expected operation to proceed with --yes flag")
		}
	})

	// Note: Interactive prompt test removed as it blocks in CI/non-TTY environments
}

// TestMessageTemplating tests parameter substitution in confirmation messages
func TestMessageTemplating(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	executor, err := NewExecutor(spec, &ExecutorConfig{
		BaseURL: "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("resource-id", "", "Resource ID")
	cmd.Flags().String("region", "", "Region")
	_ = cmd.Flags().Set("resource-id", "res-12345")
	_ = cmd.Flags().Set("region", "us-west-2")

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "kebab-case placeholder",
			message:  "Delete resource '{resource-id}' in {region}?",
			expected: "Delete resource 'res-12345' in us-west-2?",
		},
		{
			name:     "camelCase placeholder",
			message:  "Delete resource '{resourceId}' in {region}?",
			expected: "Delete resource 'res-12345' in us-west-2?",
		},
		{
			name:     "snake_case placeholder",
			message:  "Delete resource '{resource_id}' in {region}?",
			expected: "Delete resource 'res-12345' in us-west-2?",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.substituteParameters(cmd, tt.message)
			if result != tt.expected {
				t.Errorf("substituteParameters() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Benchmark parameter substitution
func BenchmarkSubstituteParameters(b *testing.B) {
	parser := openapi.NewParser()
	spec, _ := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	executor, _ := NewExecutor(spec, &ExecutorConfig{BaseURL: "https://api.example.com"})

	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("cluster-id", "", "")
	cmd.Flags().String("region", "", "")
	cmd.Flags().String("count", "", "")
	_ = cmd.Flags().Set("cluster-id", "my-cluster")
	_ = cmd.Flags().Set("region", "us-east-1")
	_ = cmd.Flags().Set("count", "5")

	message := "Delete {count} resources in cluster '{cluster-id}' ({region})?"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.substituteParameters(cmd, message)
	}
}

// Example usage of confirmation system
func ExampleExecutor_CheckConfirmation() {
	// Parse spec
	parser := openapi.NewParser()
	spec, _ := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")

	// Create executor
	executor, _ := NewExecutor(spec, &ExecutorConfig{
		BaseURL: "https://api.example.com",
	})

	// Create operation with confirmation
	op := &openapi.Operation{
		CLIConfirmation: &openapi.CLIConfirmation{
			Enabled: true,
			Message: "Are you sure you want to delete this resource?",
			Flag:    "--yes",
		},
	}

	// Create command with bypass flag
	cmd := &cobra.Command{Use: "delete"}
	cmd.Flags().Bool("yes", false, "Bypass confirmation")
	_ = cmd.Flags().Set("yes", "true")

	// Check confirmation (bypassed by flag)
	proceed, err := executor.CheckConfirmation(cmd, op)
	if err != nil {
		panic(err)
	}

	if proceed {
		// Operation proceeds
		println("Operation confirmed")
	}
}
