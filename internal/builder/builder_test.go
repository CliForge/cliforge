package builder

import (
	"context"
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

func TestBuilder_Build(t *testing.T) {
	// Parse example spec - use swagger2 example for now as it's JSON
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Create builder
	config := &BuilderConfig{
		RootName:        "test-cli",
		GroupByTags:     true,
	}
	builder := NewBuilder(spec, config)

	// Build command tree
	rootCmd, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build command tree: %v", err)
	}

	// Verify root command
	if rootCmd == nil {
		t.Fatal("Root command is nil")
	}

	if rootCmd.Use != "test-cli" {
		t.Errorf("Expected root command name 'test-cli', got '%s'", rootCmd.Use)
	}

	// Verify subcommands exist
	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Error("Expected subcommands, got none")
	}

	// Check for Users group (swagger2 example)
	var usersCmd *cobra.Command
	for _, cmd := range commands {
		if cmd.Use == "users" {
			usersCmd = cmd
			break
		}
	}

	if usersCmd == nil {
		t.Error("Expected 'users' subcommand not found")
	}
}

func TestBuilder_GetCommandByOperationID(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	builder := NewBuilder(spec, nil)
	rootCmd, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	// Find listUsers operation
	cmd := builder.GetCommandByOperationID(rootCmd, "listUsers")
	if cmd == nil {
		t.Error("Expected to find command for listUsers operation")
	}

	// Verify annotations
	if cmd != nil {
		if cmd.Annotations["operationID"] != "listUsers" {
			t.Errorf("Expected operationID annotation 'listUsers', got '%s'", cmd.Annotations["operationID"])
		}
	}
}

func TestToCommandName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"listClusters", "list-clusters"},
		{"CreateCluster", "create-cluster"},
		{"get_users", "get-users"},
		{"Delete User", "delete--user"}, // Space becomes double hyphen after processing
		{"{clusterId}", "cluster-id"}, // camelCase with braces removed
		{"MyAPICommand", "my-apicommand"}, // Consecutive caps stay together
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toCommandName(tt.input)
			if result != tt.expected {
				t.Errorf("toCommandName(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParsePathSegments(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"/clusters", []string{"clusters"}},
		{"/clusters/{id}", []string{"clusters", "{id}"}},
		{"/api/v1/users/{userId}/posts", []string{"api", "v1", "users", "{userId}", "posts"}},
		{"/", []string{}},
		{"", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parsePathSegments(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parsePathSegments(%s) returned %d segments; want %d", tt.input, len(result), len(tt.expected))
				return
			}

			for i, seg := range result {
				if seg != tt.expected[i] {
					t.Errorf("parsePathSegments(%s)[%d] = %s; want %s", tt.input, i, seg, tt.expected[i])
				}
			}
		})
	}
}

func TestBuilder_TagGrouping(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &BuilderConfig{
		GroupByTags: true,
	}
	builder := NewBuilder(spec, config)
	rootCmd, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	// Verify tag-based groups exist
	expectedGroups := []string{"users"}
	commands := rootCmd.Commands()

	for _, expectedGroup := range expectedGroups {
		found := false
		for _, cmd := range commands {
			if cmd.Use == expectedGroup {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected tag group '%s' not found", expectedGroup)
		}
	}
}
