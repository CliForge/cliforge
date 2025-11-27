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
		RootName:    "test-cli",
		GroupByTags: true,
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
		{"Delete User", "delete--user"},   // Space becomes double hyphen after processing
		{"{clusterId}", "cluster-id"},     // camelCase with braces removed
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

func TestBuilder_PathGrouping(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &BuilderConfig{
		GroupByTags: false, // Use path-based grouping
	}
	builder := NewBuilder(spec, config)
	rootCmd, err := builder.Build()
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	if rootCmd == nil {
		t.Fatal("Root command is nil")
	}

	// Verify path-based structure was created
	commands := rootCmd.Commands()
	if len(commands) == 0 {
		t.Error("Expected subcommands for path-based grouping, got none")
	}
}

func TestBuilder_GetCommandByPath(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	config := &BuilderConfig{
		GroupByTags: false, // Use path-based grouping to create path commands
	}
	builder := NewBuilder(spec, config)
	_, err = builder.Build()
	if err != nil {
		t.Fatalf("Failed to build: %v", err)
	}

	// Test getting root command
	cmd, ok := builder.GetCommandByPath("")
	if !ok || cmd == nil {
		t.Error("Expected to find root command by empty path")
	}

	// Test getting a path that exists
	_, ok = builder.GetCommandByPath("/users")
	if !ok {
		t.Error("Expected to find command for /users path")
	}

	// Test getting a path that doesn't exist
	_, ok = builder.GetCommandByPath("/nonexistent")
	if ok {
		t.Error("Expected not to find command for non-existent path")
	}
}

func TestDetermineCommandName(t *testing.T) {
	tests := []struct {
		name     string
		op       *openapi.Operation
		expected string
	}{
		{
			name: "uses x-cli-command if present",
			op: &openapi.Operation{
				CLICommand:  "list-all",
				OperationID: "listUsers",
				Method:      "GET",
			},
			expected: "list-all",
		},
		{
			name: "uses operationID if x-cli-command not present",
			op: &openapi.Operation{
				CLICommand:  "",
				OperationID: "getUserById",
				Method:      "GET",
			},
			expected: "get-user-by-id",
		},
		{
			name: "falls back to method if no operationID",
			op: &openapi.Operation{
				CLICommand:  "",
				OperationID: "",
				Method:      "POST",
			},
			expected: "post",
		},
		{
			name: "handles empty x-cli-command",
			op: &openapi.Operation{
				CLICommand:  "",
				OperationID: "deleteResource",
				Method:      "DELETE",
			},
			expected: "delete-resource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &Builder{}
			result := builder.determineCommandName(tt.op)
			if result != tt.expected {
				t.Errorf("determineCommandName() = %s; want %s", result, tt.expected)
			}
		})
	}
}

func TestGetOrCreatePathCommand(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		expectCreated bool
	}{
		{
			name:          "simple path",
			path:          "/users",
			expectCreated: true,
		},
		{
			name:          "nested path",
			path:          "/api/v1/users",
			expectCreated: true,
		},
		{
			name:          "path with parameter",
			path:          "/users/{id}",
			expectCreated: true,
		},
		{
			name:          "path with multiple parameters",
			path:          "/users/{userId}/posts/{postId}",
			expectCreated: true,
		},
		{
			name:          "empty path returns root",
			path:          "",
			expectCreated: false,
		},
		{
			name:          "root path returns root",
			path:          "/",
			expectCreated: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := &Builder{
				commandMap: make(map[string]*cobra.Command),
			}
			rootCmd := &cobra.Command{Use: "root"}
			builder.commandMap[""] = rootCmd

			result := builder.getOrCreatePathCommand(rootCmd, tt.path)
			if result == nil {
				t.Fatal("getOrCreatePathCommand returned nil")
			}

			if !tt.expectCreated && result != rootCmd {
				t.Error("Expected to return root command for empty/root path")
			}
		})
	}
}

func TestCamelToKebab(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simpleCase", "simple-Case"},
		{"HTTPServer", "HTTPServer"}, // Consecutive caps stay together
		{"getUserByID", "get-User-By-ID"},
		{"myAPIKey", "my-APIKey"},
		{"ABC", "ABC"}, // All caps
		{"a", "a"},     // Single char
		{"", ""},       // Empty
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := camelToKebab(tt.input)
			if result != tt.expected {
				t.Errorf("camelToKebab(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestBuildRootCommand_WithExtensions(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Test with custom config
	config := &BuilderConfig{
		RootName:        "custom-cli",
		RootDescription: "Custom description",
	}
	builder := &Builder{
		spec:       spec,
		config:     config,
		commandMap: make(map[string]*cobra.Command),
	}

	rootCmd := builder.buildRootCommand()

	if rootCmd == nil {
		t.Fatal("Root command is nil")
	}

	// Verify custom config was applied
	if rootCmd.Use != "custom-cli" {
		t.Errorf("Expected root command Use 'custom-cli', got '%s'", rootCmd.Use)
	}
}

func TestBuilder_Build_ErrorHandling(t *testing.T) {
	// Test with a minimal valid spec
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Test with default config
	builder := NewBuilder(spec, nil)
	rootCmd, err := builder.Build()
	if err != nil {
		t.Fatalf("Build failed with default config: %v", err)
	}
	if rootCmd == nil {
		t.Error("Expected non-nil root command")
	}

	// Verify default config was used
	if builder.config == nil {
		t.Error("Expected default config to be set")
	}
}

func TestBuildOperationCommand_Annotations(t *testing.T) {
	builder := &Builder{
		config: DefaultBuilderConfig(),
	}

	op := &openapi.Operation{
		OperationID: "testOp",
		Method:      "GET",
		Path:        "/test",
		Summary:     "Test operation",
		Description: "Test description",
		CLICommand:  "custom-test",
		CLIAliases:  []string{"ct", "test-alias"},
	}

	cmd, err := builder.buildOperationCommand(op)
	if err != nil {
		t.Fatalf("Failed to build operation command: %v", err)
	}

	// Verify custom command name
	if cmd.Use != "custom-test" {
		t.Errorf("Expected Use 'custom-test', got '%s'", cmd.Use)
	}

	// Verify aliases
	if len(cmd.Aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(cmd.Aliases))
	}

	// Verify annotations
	if cmd.Annotations["operationID"] != "testOp" {
		t.Errorf("Expected operationID 'testOp', got '%s'", cmd.Annotations["operationID"])
	}
	if cmd.Annotations["method"] != "GET" {
		t.Errorf("Expected method 'GET', got '%s'", cmd.Annotations["method"])
	}
	if cmd.Annotations["path"] != "/test" {
		t.Errorf("Expected path '/test', got '%s'", cmd.Annotations["path"])
	}
}

func TestBuildOperationCommand_WithExecutor(t *testing.T) {
	executorCalled := false
	executor := func(cmd *cobra.Command, args []string) error {
		executorCalled = true
		return nil
	}

	builder := &Builder{
		config: &BuilderConfig{
			DefaultExecutor: executor,
		},
	}

	op := &openapi.Operation{
		OperationID: "testOp",
		Method:      "GET",
		Path:        "/test",
	}

	cmd, err := builder.buildOperationCommand(op)
	if err != nil {
		t.Fatalf("Failed to build operation command: %v", err)
	}

	// Verify executor was set
	if cmd.RunE == nil {
		t.Error("Expected RunE to be set")
	}

	// Execute and verify executor is called
	if cmd.RunE != nil {
		_ = cmd.RunE(cmd, []string{})
		if !executorCalled {
			t.Error("Expected executor to be called")
		}
	}
}
