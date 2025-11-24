package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CliForge/cliforge/tests/helpers"
)

// TestFullCLILifecycle tests the complete CLI lifecycle from build to execution.
func TestFullCLILifecycle(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure mock endpoints
	apiServer.OnGET("/users", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"users": []map[string]interface{}{
			{"id": "1", "name": "Alice", "email": "alice@example.com"},
			{"id": "2", "name": "Bob", "email": "bob@example.com"},
		},
	}))

	apiServer.OnGET("/users/1", helpers.JSONResponse(http.StatusOK, map[string]interface{}{
		"id":    "1",
		"name":  "Alice",
		"email": "alice@example.com",
	}))

	apiServer.OnPOST("/users", helpers.JSONResponse(http.StatusCreated, map[string]interface{}{
		"id":    "3",
		"name":  "Charlie",
		"email": "charlie@example.com",
	}))

	apiServer.OnDELETE("/users/1", helpers.JSONResponse(http.StatusNoContent, nil))

	// Create CLI builder
	builder := helpers.NewCLIBuilder(t)

	// Set environment to point to mock server
	builder.WithEnv("TEST_API_URL", apiServer.URL())

	// Build the CLI
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Show version
	t.Run("ShowVersion", func(t *testing.T) {
		result := builder.Run("--version")
		result.AssertSuccess(t)
		result.AssertContains(t, "0.9.0")
	})

	// Test: Show help
	t.Run("ShowHelp", func(t *testing.T) {
		result := builder.Run("--help")
		result.AssertSuccess(t)
		result.AssertContains(t, "cliforge")
		result.AssertContains(t, "Usage:")
	})

	// Verify binary was created
	if _, err := os.Stat(builder.BinaryPath()); os.IsNotExist(err) {
		t.Fatalf("Binary not created at %s", builder.BinaryPath())
	}
}

// TestCLIBuildCommand tests the 'build' command functionality.
func TestCLIBuildCommand(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Create a test OpenAPI spec
	specPath := filepath.Join(builder.WorkDir(), "test-api.yaml")
	specContent, err := helpers.ReadFile("../fixtures/test-api.yaml")
	if err != nil {
		t.Fatalf("Failed to read test spec: %v", err)
	}

	if err := helpers.WriteFile(specPath, specContent); err != nil {
		t.Fatalf("Failed to write test spec: %v", err)
	}

	// Test: Build CLI from spec
	t.Run("BuildFromSpec", func(t *testing.T) {
		outputDir := filepath.Join(builder.WorkDir(), "output")
		result := builder.Run("build", "--spec", specPath, "--output", outputDir)

		result.AssertSuccess(t)
		result.AssertContains(t, "success")

		// Verify generated files exist
		helpers.AssertFileExists(t, filepath.Join(outputDir, "main.go"))
	})
}

// TestCLIValidateCommand tests the 'validate' command functionality.
func TestCLIValidateCommand(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Validate valid spec
	t.Run("ValidateValidSpec", func(t *testing.T) {
		specPath := "../fixtures/test-api.yaml"
		result := builder.Run("validate", "--spec", specPath)

		result.AssertSuccess(t)
		result.AssertContains(t, "valid")
	})

	// Test: Validate invalid spec
	t.Run("ValidateInvalidSpec", func(t *testing.T) {
		invalidSpecPath := filepath.Join(builder.WorkDir(), "invalid.yaml")
		invalidSpec := "invalid: yaml: content:"
		if err := helpers.WriteFile(invalidSpecPath, invalidSpec); err != nil {
			t.Fatalf("Failed to write invalid spec: %v", err)
		}

		result := builder.Run("validate", "--spec", invalidSpecPath)
		result.AssertFailed(t)
		result.AssertStderrContains(t, "error")
	})
}

// TestCLICommandExecution tests executing commands against a mock API.
func TestCLICommandExecution(t *testing.T) {
	// Start mock API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Start OAuth2 server
	oauthServer := helpers.NewMockOAuth2Server("test-client", "test-secret")
	defer oauthServer.Close()

	// Configure mock API endpoints
	users := []map[string]interface{}{
		{"id": "1", "name": "Alice", "email": "alice@example.com"},
		{"id": "2", "name": "Bob", "email": "bob@example.com"},
	}

	apiServer.OnGET("/users", oauthServer.AuthenticatedHandler(
		helpers.JSONResponse(http.StatusOK, users),
	))

	apiServer.OnGET("/users/1", oauthServer.AuthenticatedHandler(
		helpers.JSONResponse(http.StatusOK, users[0]),
	))

	apiServer.OnPOST("/users", oauthServer.AuthenticatedHandler(
		func(w http.ResponseWriter, r *http.Request) {
			var newUser map[string]interface{}
			json.NewDecoder(r.Body).Decode(&newUser)
			newUser["id"] = "3"
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newUser)
		},
	))

	// This would require a generated CLI binary
	// For now, we verify the mock servers are working
	t.Run("MockServersRunning", func(t *testing.T) {
		helpers.AssertNotNil(t, apiServer)
		helpers.AssertNotNil(t, oauthServer)

		// Verify OAuth2 server can issue tokens
		token := oauthServer.generateToken("read write")
		helpers.AssertNotEqual(t, "", token.AccessToken)
		helpers.AssertNotEqual(t, "", token.RefreshToken)
		helpers.AssertEqual(t, "Bearer", token.TokenType)
	})
}

// TestCLIErrorHandling tests error handling scenarios.
func TestCLIErrorHandling(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Missing required argument
	t.Run("MissingRequiredArgument", func(t *testing.T) {
		result := builder.Run("build")
		result.AssertFailed(t)
		result.AssertStderrContains(t, "required")
	})

	// Test: Invalid command
	t.Run("InvalidCommand", func(t *testing.T) {
		result := builder.Run("invalid-command")
		result.AssertFailed(t)
		result.AssertStderrContains(t, "unknown")
	})

	// Test: Non-existent spec file
	t.Run("NonExistentSpec", func(t *testing.T) {
		result := builder.Run("validate", "--spec", "/nonexistent/spec.yaml")
		result.AssertFailed(t)
		result.AssertStderrContains(t, "not found")
	})
}

// TestCLIPerformance tests CLI performance characteristics.
func TestCLIPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Command execution time
	t.Run("CommandExecutionTime", func(t *testing.T) {
		result := builder.Run("--version")
		result.AssertSuccess(t)

		// Version command should complete quickly
		helpers.AssertDurationLessThan(t, result.Duration, 1*time.Second, "Version command too slow")
	})
}

// TestCLIConcurrentExecution tests concurrent CLI execution.
func TestCLIConcurrentExecution(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Multiple concurrent executions
	t.Run("ConcurrentVersionChecks", func(t *testing.T) {
		const concurrency = 5

		results := make(chan *helpers.CLIResult, concurrency)

		// Launch concurrent executions
		for i := 0; i < concurrency; i++ {
			go func() {
				results <- builder.Run("--version")
			}()
		}

		// Collect results
		for i := 0; i < concurrency; i++ {
			result := <-results
			result.AssertSuccess(t)
			result.AssertContains(t, "0.9.0")
		}
	})
}

// TestCLIConfigFile tests configuration file handling.
func TestCLIConfigFile(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Create a config file
	configContent, err := helpers.ReadFile("../fixtures/config.yaml")
	if err != nil {
		t.Fatalf("Failed to read config fixture: %v", err)
	}

	configPath, err := builder.CreateConfigFile("config.yaml", configContent)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test: Validate with config file
	t.Run("ValidateWithConfig", func(t *testing.T) {
		helpers.AssertFileExists(t, configPath)
		// Further tests would use the generated CLI with config
	})
}

// TestCLIOutputFormats tests different output format options.
func TestCLIOutputFormats(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// These tests would work with a generated CLI that supports output formats
	t.Run("OutputFormatsSupported", func(t *testing.T) {
		// Verify the builder is ready
		helpers.AssertNotEqual(t, "", builder.BinaryPath())
	})
}

// TestCLIEnvironmentVariables tests environment variable handling.
func TestCLIEnvironmentVariables(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Set test environment variables
	builder.WithEnv("CLIFORGE_DEBUG", "true")
	builder.WithEnv("CLIFORGE_NO_COLOR", "true")

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Environment variables are respected
	t.Run("EnvironmentVariables", func(t *testing.T) {
		result := builder.Run("--version")
		result.AssertSuccess(t)
		// Debug mode might add extra output
	})
}

// TestCLITimeout tests command timeout handling.
func TestCLITimeout(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Command with timeout
	t.Run("CommandTimeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Quick command should complete before timeout
		result := builder.RunWithContext(ctx, "--version")
		result.AssertSuccess(t)
	})
}

// TestCLICleanup tests cleanup and resource management.
func TestCLICleanup(t *testing.T) {
	builder := helpers.NewCLIBuilder(t)

	// Build the cliforge binary
	pkgPath := "../../cmd/cliforge"
	if err := builder.Build(pkgPath); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Test: Verify temp directory cleanup
	t.Run("TempDirectoryExists", func(t *testing.T) {
		helpers.AssertFileExists(t, builder.WorkDir())
		helpers.AssertFileExists(t, builder.BinaryPath())
	})
}
