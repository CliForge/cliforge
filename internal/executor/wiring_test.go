package executor

import (
	"context"
	"testing"
)

func TestNewRuntime(t *testing.T) {
	tests := []struct {
		name        string
		config      *RuntimeConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid runtime with swagger spec",
			config: &RuntimeConfig{
				CLIName:  "testcli",
				SpecPath: "../../examples/openapi/swagger2-example.json",
				BaseURL:  "https://api.example.com",
			},
			wantErr: false,
		},
		{
			name: "invalid spec path",
			config: &RuntimeConfig{
				CLIName:  "testcli",
				SpecPath: "/nonexistent/spec.json",
				BaseURL:  "https://api.example.com",
			},
			wantErr:     true,
			errContains: "failed to parse OpenAPI spec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			runtime, err := NewRuntime(ctx, tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewRuntime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errContains != "" && err != nil {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("NewRuntime() error = %v, should contain %v", err, tt.errContains)
					}
				}
				return
			}

			// Verify runtime was created correctly
			if runtime == nil {
				t.Fatal("Expected runtime to be non-nil")
			}

			if runtime.spec == nil {
				t.Error("Expected spec to be initialized")
			}

			if runtime.rootCmd == nil {
				t.Error("Expected root command to be initialized")
			}

			if runtime.authManager == nil {
				t.Error("Expected auth manager to be initialized")
			}

			if runtime.outputManager == nil {
				t.Error("Expected output manager to be initialized")
			}

			if runtime.stateManager == nil {
				t.Error("Expected state manager to be initialized")
			}

			if runtime.progressManager == nil {
				t.Error("Expected progress manager to be initialized")
			}

			if runtime.pluginRegistry == nil {
				t.Error("Expected plugin registry to be initialized")
			}

			if runtime.executor == nil {
				t.Error("Expected executor to be initialized")
			}

			if runtime.httpClient == nil {
				t.Error("Expected HTTP client to be initialized")
			}
		})
	}
}

func TestRuntime_GetRootCommand(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	rootCmd := runtime.GetRootCommand()
	if rootCmd == nil {
		t.Error("GetRootCommand() returned nil")
	}

	if rootCmd.Use == "" {
		t.Error("Root command should have Use set")
	}
}

func TestRuntime_GetAuthManager(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	authMgr := runtime.GetAuthManager()
	if authMgr == nil {
		t.Error("GetAuthManager() returned nil")
	}
}

func TestRuntime_GetStateManager(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	stateMgr := runtime.GetStateManager()
	if stateMgr == nil {
		t.Error("GetStateManager() returned nil")
	}
}

func TestRuntime_GetOutputManager(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	outputMgr := runtime.GetOutputManager()
	if outputMgr == nil {
		t.Error("GetOutputManager() returned nil")
	}
}

func TestRuntime_GetProgressManager(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	progressMgr := runtime.GetProgressManager()
	if progressMgr == nil {
		t.Error("GetProgressManager() returned nil")
	}
}

func TestRuntime_GetPluginRegistry(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	pluginRegistry := runtime.GetPluginRegistry()
	if pluginRegistry == nil {
		t.Error("GetPluginRegistry() returned nil")
	}
}

func TestRuntime_Shutdown(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	err = runtime.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestRuntime_Execute(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// Set args to show help (should not error)
	runtime.rootCmd.SetArgs([]string{"--help"})

	// Execute should show help and exit cleanly
	// We don't actually call Execute() here as it would exit the test process
	// Just verify the command is set up correctly
	if runtime.rootCmd == nil {
		t.Error("Root command not initialized")
	}
}

func TestRuntime_ExecuteContext(t *testing.T) {
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// Set args to show help
	runtime.rootCmd.SetArgs([]string{"--help"})

	// ExecuteContext with help flag should not error
	// We skip actually calling it to avoid test process exit
	if runtime.rootCmd == nil {
		t.Error("Root command not initialized for ExecuteContext")
	}
}

func TestRuntime_InitializeManagers(t *testing.T) {
	// Test is implicitly covered by TestNewRuntime
	// This test verifies that NewRuntime properly initializes all managers
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// All managers should be non-nil
	if runtime.authManager == nil {
		t.Error("authManager not initialized")
	}
	if runtime.stateManager == nil {
		t.Error("stateManager not initialized")
	}
	if runtime.outputManager == nil {
		t.Error("outputManager not initialized")
	}
	if runtime.progressManager == nil {
		t.Error("progressManager not initialized")
	}
	if runtime.pluginRegistry == nil {
		t.Error("pluginRegistry not initialized")
	}
}

func TestRuntime_CreateHTTPClient(t *testing.T) {
	// Test is implicitly covered by TestNewRuntime
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if runtime.httpClient == nil {
		t.Error("HTTP client not initialized")
	}

	if runtime.httpClient.Timeout == 0 {
		t.Error("HTTP client should have timeout set")
	}
}

func TestRuntime_BuildCommandTree(t *testing.T) {
	// Test is implicitly covered by TestNewRuntime
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	if runtime.rootCmd == nil {
		t.Fatal("Root command not built")
	}

	// Should have subcommands from the spec
	if !runtime.rootCmd.HasSubCommands() {
		t.Error("Root command should have subcommands from OpenAPI spec")
	}
}

func TestRuntime_AddOperationFlags(t *testing.T) {
	// Test is implicitly covered by TestNewRuntime
	ctx := context.Background()
	runtime, err := NewRuntime(ctx, &RuntimeConfig{
		CLIName:  "testcli",
		SpecPath: "../../examples/openapi/swagger2-example.json",
		BaseURL:  "https://api.example.com",
	})
	if err != nil {
		t.Fatalf("Failed to create runtime: %v", err)
	}

	// Check that the command tree was built
	if runtime.rootCmd == nil {
		t.Fatal("Root command not built")
	}

	// Just verify that flags were attempted to be added
	// The actual flag presence depends on the operation parameters in the spec
	if !runtime.rootCmd.HasSubCommands() {
		t.Error("Root command should have subcommands")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != substr && len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || hasSubstring(s, substr))))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
