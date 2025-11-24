package integration

import (
	"context"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/plugin"
	"github.com/CliForge/cliforge/tests/helpers"
)

// TestPluginExecution tests basic plugin execution.
func TestPluginExecution(t *testing.T) {
	// Test: Plugin interface implementation
	t.Run("PluginInterface", func(t *testing.T) {
		// Create a mock plugin
		mockPlugin := &MockPlugin{
			name:    "test-plugin",
			version: "1.0.0",
		}

		// Verify plugin implements interface
		var _ plugin.Plugin = mockPlugin

		info := mockPlugin.Describe()
		helpers.AssertEqual(t, "test-plugin", info.Manifest.Name)
		helpers.AssertEqual(t, "1.0.0", info.Manifest.Version)
	})
}

// TestPluginRegistry tests the plugin registry.
func TestPluginRegistry(t *testing.T) {
	// Test: Register and retrieve plugins
	t.Run("RegisterPlugin", func(t *testing.T) {
		registry := plugin.NewRegistry()
		helpers.AssertNotNil(t, registry)

		mockPlugin := &MockPlugin{
			name:    "registered-plugin",
			version: "1.0.0",
		}

		err := registry.Register(mockPlugin)
		helpers.AssertNoError(t, err)

		// Retrieve plugin
		retrievedPlugin, err := registry.Get("registered-plugin")
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, retrievedPlugin)

		info := retrievedPlugin.Describe()
		helpers.AssertEqual(t, "registered-plugin", info.Manifest.Name)
	})

	// Test: Duplicate registration
	t.Run("DuplicateRegistration", func(t *testing.T) {
		registry := plugin.NewRegistry()

		mockPlugin1 := &MockPlugin{name: "duplicate", version: "1.0.0"}
		mockPlugin2 := &MockPlugin{name: "duplicate", version: "2.0.0"}

		err := registry.Register(mockPlugin1)
		helpers.AssertNoError(t, err)

		err = registry.Register(mockPlugin2)
		helpers.AssertError(t, err)
		helpers.AssertErrorContains(t, err, "already registered")
	})
}

// TestPluginPermissions tests plugin permission system.
func TestPluginPermissions(t *testing.T) {
	// Test: Permission validation
	t.Run("PermissionValidation", func(t *testing.T) {
		permissions := []plugin.Permission{
			{Type: plugin.PermissionExecute, Resource: "aws"},
			{Type: plugin.PermissionReadFile, Resource: "*.yaml"},
			{Type: plugin.PermissionWriteFile, Resource: "/tmp/*"},
			{Type: plugin.PermissionNetwork},
		}

		for _, perm := range permissions {
			helpers.AssertNotEqual(t, "", string(perm.Type))
		}

		// Test permission string representation
		execPerm := permissions[0]
		helpers.AssertEqual(t, "execute:aws", execPerm.String())

		networkPerm := permissions[3]
		helpers.AssertEqual(t, "network", networkPerm.String())
	})
}

// TestPluginManifest tests plugin manifest parsing and validation.
func TestPluginManifest(t *testing.T) {
	// Test: Valid manifest
	t.Run("ValidManifest", func(t *testing.T) {
		manifest := plugin.PluginManifest{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Type:        plugin.PluginTypeBuiltin,
			Description: "Test plugin",
			Author:      "Test Author",
			Permissions: []plugin.Permission{
				{Type: plugin.PermissionExecute, Resource: "*"},
			},
		}

		helpers.AssertEqual(t, "test-plugin", manifest.Name)
		helpers.AssertEqual(t, "1.0.0", manifest.Version)
		helpers.AssertEqual(t, plugin.PluginTypeBuiltin, manifest.Type)
		helpers.AssertEqual(t, 1, len(manifest.Permissions))
	})

	// Test: Load manifest from file
	t.Run("LoadManifest", func(t *testing.T) {
		manifestPath := "../fixtures/plugin-manifest.yaml"
		helpers.AssertFileExists(t, manifestPath)

		// In a real implementation, this would parse the YAML
		content, err := helpers.ReadFile(manifestPath)
		helpers.AssertNoError(t, err)
		helpers.AssertStringContains(t, content, "test-plugin")
	})
}

// TestPluginInput tests plugin input handling.
func TestPluginInput(t *testing.T) {
	// Test: Create plugin input
	t.Run("CreateInput", func(t *testing.T) {
		input := &plugin.PluginInput{
			Command: "process",
			Args:    []string{"--input", "data.json"},
			Env: map[string]string{
				"DEBUG": "true",
			},
			Data: map[string]interface{}{
				"config": "test",
			},
			Timeout:    30 * time.Second,
			WorkingDir: "/tmp",
		}

		helpers.AssertEqual(t, "process", input.Command)
		helpers.AssertEqual(t, 2, len(input.Args))
		helpers.AssertEqual(t, "true", input.Env["DEBUG"])
		helpers.AssertEqual(t, 30*time.Second, input.Timeout)
	})
}

// TestPluginOutput tests plugin output handling.
func TestPluginOutput(t *testing.T) {
	// Test: Successful output
	t.Run("SuccessfulOutput", func(t *testing.T) {
		output := &plugin.PluginOutput{
			Stdout:   "Operation completed successfully",
			ExitCode: 0,
			Data: map[string]interface{}{
				"result":  "success",
				"count":   42,
				"enabled": true,
			},
			Duration: 150 * time.Millisecond,
		}

		helpers.AssertTrue(t, output.Success(), "Output should indicate success")

		// Test data extraction
		result, ok := output.GetString("result")
		helpers.AssertTrue(t, ok, "Should retrieve string value")
		helpers.AssertEqual(t, "success", result)

		count, ok := output.GetInt("count")
		helpers.AssertTrue(t, ok, "Should retrieve int value")
		helpers.AssertEqual(t, 42, count)

		enabled, ok := output.GetBool("enabled")
		helpers.AssertTrue(t, ok, "Should retrieve bool value")
		helpers.AssertTrue(t, enabled, "Value should be true")
	})

	// Test: Failed output
	t.Run("FailedOutput", func(t *testing.T) {
		output := &plugin.PluginOutput{
			Stderr:   "Operation failed",
			ExitCode: 1,
			Error:    "Process exited with code 1",
		}

		helpers.AssertFalse(t, output.Success(), "Output should indicate failure")
	})
}

// TestPluginError tests plugin error handling.
func TestPluginError(t *testing.T) {
	// Test: Create plugin error
	t.Run("CreateError", func(t *testing.T) {
		baseErr := plugin.NewPluginError(
			"test-plugin",
			"execution failed",
			nil,
		)

		helpers.AssertNotNil(t, baseErr)
		helpers.AssertStringContains(t, baseErr.Error(), "test-plugin")
		helpers.AssertStringContains(t, baseErr.Error(), "execution failed")
	})

	// Test: Error with suggestion
	t.Run("ErrorWithSuggestion", func(t *testing.T) {
		err := plugin.NewPluginError(
			"test-plugin",
			"command not found",
			nil,
		).WithSuggestion("Install the plugin using: cliforge plugin install test-plugin")

		helpers.AssertNotNil(t, err.Suggestion)
		helpers.AssertStringContains(t, err.Suggestion, "cliforge plugin install")
	})

	// Test: Recoverable error
	t.Run("RecoverableError", func(t *testing.T) {
		err := plugin.NewPluginError(
			"test-plugin",
			"temporary failure",
			nil,
		).AsRecoverable()

		helpers.AssertTrue(t, err.Recoverable, "Error should be marked as recoverable")
	})
}

// TestPluginContext tests plugin context.
func TestPluginContext(t *testing.T) {
	// Test: Create plugin context
	t.Run("CreateContext", func(t *testing.T) {
		ctx := &plugin.PluginContext{
			CLIName:    "testcli",
			CLIVersion: "1.0.0",
			ConfigDir:  "/home/user/.config/testcli",
			CacheDir:   "/home/user/.cache/testcli",
			DataDir:    "/home/user/.local/share/testcli",
			Debug:      true,
			DryRun:     false,
			UserData: map[string]interface{}{
				"custom": "value",
			},
		}

		helpers.AssertEqual(t, "testcli", ctx.CLIName)
		helpers.AssertTrue(t, ctx.Debug, "Debug should be enabled")
		helpers.AssertFalse(t, ctx.DryRun, "DryRun should be disabled")
	})
}

// TestPluginInWorkflow tests plugin integration in workflows.
func TestPluginInWorkflow(t *testing.T) {
	// Test: Plugin step in workflow
	t.Run("PluginStep", func(t *testing.T) {
		// This would be tested in workflow_test.go
		// Here we just verify the plugin can be configured for workflow use
		mockPlugin := &MockPlugin{
			name:    "workflow-plugin",
			version: "1.0.0",
		}

		ctx := context.Background()
		input := &plugin.PluginInput{
			Command: "execute",
			Data: map[string]interface{}{
				"workflow_id": "wf-123",
			},
		}

		output, err := mockPlugin.Execute(ctx, input)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, output)
		helpers.AssertTrue(t, output.Success(), "Plugin execution should succeed")
	})
}

// TestBuiltinPlugins tests built-in plugins.
func TestBuiltinPlugins(t *testing.T) {
	// Test: Echo plugin
	t.Run("EchoPlugin", func(t *testing.T) {
		echoPlugin := &EchoPlugin{}
		ctx := context.Background()

		input := &plugin.PluginInput{
			Args: []string{"Hello, World!"},
		}

		output, err := echoPlugin.Execute(ctx, input)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, output)
		helpers.AssertStringContains(t, output.Stdout, "Hello, World!")
	})
}

// TestPluginTimeout tests plugin execution timeout.
func TestPluginTimeout(t *testing.T) {
	// Test: Plugin with timeout
	t.Run("PluginTimeout", func(t *testing.T) {
		slowPlugin := &SlowPlugin{
			delay: 2 * time.Second,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		input := &plugin.PluginInput{
			Command: "slow-operation",
		}

		_, err := slowPlugin.Execute(ctx, input)
		helpers.AssertError(t, err)
		helpers.AssertErrorContains(t, err, "context deadline exceeded")
	})
}

// MockPlugin is a mock plugin for testing.
type MockPlugin struct {
	name    string
	version string
}

func (m *MockPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
	return &plugin.PluginOutput{
		Stdout:   "Mock plugin executed",
		ExitCode: 0,
		Data: map[string]interface{}{
			"executed": true,
		},
		Duration: 10 * time.Millisecond,
	}, nil
}

func (m *MockPlugin) Validate() error {
	return nil
}

func (m *MockPlugin) Describe() *plugin.PluginInfo {
	return &plugin.PluginInfo{
		Manifest: plugin.PluginManifest{
			Name:    m.name,
			Version: m.version,
			Type:    plugin.PluginTypeBuiltin,
		},
		Status: plugin.PluginStatusReady,
	}
}

// EchoPlugin echoes input back.
type EchoPlugin struct{}

func (e *EchoPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
	output := ""
	for _, arg := range input.Args {
		output += arg + " "
	}

	return &plugin.PluginOutput{
		Stdout:   output,
		ExitCode: 0,
	}, nil
}

func (e *EchoPlugin) Validate() error {
	return nil
}

func (e *EchoPlugin) Describe() *plugin.PluginInfo {
	return &plugin.PluginInfo{
		Manifest: plugin.PluginManifest{
			Name:    "echo",
			Version: "1.0.0",
			Type:    plugin.PluginTypeBuiltin,
		},
		Status: plugin.PluginStatusReady,
	}
}

// SlowPlugin simulates a slow operation.
type SlowPlugin struct {
	delay time.Duration
}

func (s *SlowPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
	select {
	case <-time.After(s.delay):
		return &plugin.PluginOutput{
			Stdout:   "Slow operation completed",
			ExitCode: 0,
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *SlowPlugin) Validate() error {
	return nil
}

func (s *SlowPlugin) Describe() *plugin.PluginInfo {
	return &plugin.PluginInfo{
		Manifest: plugin.PluginManifest{
			Name:    "slow",
			Version: "1.0.0",
			Type:    plugin.PluginTypeBuiltin,
		},
		Status: plugin.PluginStatusReady,
	}
}
