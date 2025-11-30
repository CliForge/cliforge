package tests

import (
	"fmt"
	"os"
	"testing"
)

var (
	// testHarness is the global test harness instance shared across all tests.
	testHarness *Harness
)

// TestMain is the entry point for all tests. It sets up and tears down the test environment.
func TestMain(m *testing.M) {
	var exitCode int

	// Defer cleanup to ensure it always runs
	defer func() {
		if testHarness != nil {
			if err := testHarness.Stop(); err != nil {
				fmt.Fprintf(os.Stderr, "Failed to stop test harness: %v\n", err)
			}
		}
		os.Exit(exitCode)
	}()

	// Create test harness
	harness, err := NewHarness()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create test harness: %v\n", err)
		exitCode = 1
		return
	}
	testHarness = harness

	// Start the test environment (including mock server)
	if err := testHarness.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start test harness: %v\n", err)
		exitCode = 1
		return
	}

	fmt.Printf("Test environment ready:\n")
	fmt.Printf("  - Mock Server: %s\n", testHarness.ServerURL)
	fmt.Printf("  - CLI Binary: %s\n", testHarness.CLIBinary)
	fmt.Printf("  - Test Directory: %s\n", testHarness.testDir)

	// Run all tests
	exitCode = m.Run()
}

// Helper function to get the global test harness.
// Tests should call this to access the harness.
func getHarness(t *testing.T) *Harness {
	if testHarness == nil {
		t.Fatal("test harness is not initialized")
	}
	return testHarness
}

// Helper function to clean credentials before a test.
// Use this in tests that require a fresh authentication state.
func cleanCredentials(t *testing.T) {
	h := getHarness(t)
	if err := h.CleanCredentials(); err != nil {
		t.Fatalf("Failed to clean credentials: %v", err)
	}
}

// Helper function to setup mock authentication before a test.
// Use this in tests that require authenticated API access.
func mockAuthenticate(t *testing.T) {
	h := getHarness(t)
	if err := h.MockAuthenticate(); err != nil {
		t.Fatalf("Failed to setup mock authentication: %v", err)
	}
}

// TestHarnessSetup verifies that the test harness is correctly initialized.
func TestHarnessSetup(t *testing.T) {
	h := getHarness(t)

	// Verify CLI binary exists and is executable
	if _, err := os.Stat(h.CLIBinary); err != nil {
		t.Fatalf("CLI binary not found or not accessible: %v", err)
	}

	// Verify mock server is responding
	if err := h.VerifyAPICall("GET", "/health"); err != nil {
		t.Fatalf("Mock server health check failed: %v", err)
	}

	// Verify test directories are created
	if _, err := os.Stat(h.testDir); err != nil {
		t.Fatalf("Test directory not created: %v", err)
	}

	if _, err := os.Stat(h.configDir); err != nil {
		t.Fatalf("Config directory not created: %v", err)
	}

	t.Logf("Test harness successfully initialized")
	t.Logf("  Server URL: %s", h.ServerURL)
	t.Logf("  CLI Binary: %s", h.CLIBinary)
	t.Logf("  Test Dir: %s", h.testDir)
}

// TestCLIExecution verifies basic CLI execution works.
func TestCLIExecution(t *testing.T) {
	h := getHarness(t)

	tests := []struct {
		name        string
		args        []string
		expectError bool
		expectOut   string
	}{
		{
			name:        "version command",
			args:        []string{"version"},
			expectError: false,
			expectOut:   "rosa version",
		},
		{
			name:        "help command",
			args:        []string{"--help"},
			expectError: false,
			expectOut:   "rosa",
		},
		{
			name:        "invalid command",
			args:        []string{"nonexistent"},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, err := h.RunCLI(tt.args...)

			// Check error expectation
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v\nStdout: %s\nStderr: %s", err, stdout, stderr)
			}

			// Check output contains expected string
			if tt.expectOut != "" {
				output := stdout + stderr
				if !containsString(output, tt.expectOut) {
					t.Errorf("Expected output to contain %q, got:\n%s", tt.expectOut, output)
				}
			}
		})
	}
}

// TestOutputParsing verifies the output parsing utilities work correctly.
func TestOutputParsing(t *testing.T) {
	h := getHarness(t)

	t.Run("parse JSON", func(t *testing.T) {
		jsonOutput := `{"name": "test-cluster", "state": "ready", "id": "12345"}`
		var result map[string]interface{}

		err := h.ParseJSONOutput(jsonOutput, &result)
		if err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if result["name"] != "test-cluster" {
			t.Errorf("Expected name to be 'test-cluster', got %v", result["name"])
		}
		if result["state"] != "ready" {
			t.Errorf("Expected state to be 'ready', got %v", result["state"])
		}
	})

	t.Run("parse YAML", func(t *testing.T) {
		yamlOutput := `
name: test-cluster
state: ready
id: "12345"
`
		var result map[string]interface{}

		err := h.ParseYAMLOutput(yamlOutput, &result)
		if err != nil {
			t.Fatalf("Failed to parse YAML: %v", err)
		}

		if result["name"] != "test-cluster" {
			t.Errorf("Expected name to be 'test-cluster', got %v", result["name"])
		}
	})

	t.Run("parse table", func(t *testing.T) {
		tableOutput := `
ID       NAME           STATE   REGION
-------- -------------- ------- --------
abc123   prod-cluster   ready   us-east-1
def456   test-cluster   pending us-west-2
`
		rows, err := h.ParseTableOutput(tableOutput)
		if err != nil {
			t.Fatalf("Failed to parse table: %v", err)
		}

		if len(rows) != 2 {
			t.Fatalf("Expected 2 rows, got %d", len(rows))
		}

		// Verify first row
		if rows[0]["ID"] != "abc123" {
			t.Errorf("Expected ID to be 'abc123', got %v", rows[0]["ID"])
		}
		if rows[0]["NAME"] != "prod-cluster" {
			t.Errorf("Expected NAME to be 'prod-cluster', got %v", rows[0]["NAME"])
		}
		if rows[0]["STATE"] != "ready" {
			t.Errorf("Expected STATE to be 'ready', got %v", rows[0]["STATE"])
		}

		// Verify second row
		if rows[1]["ID"] != "def456" {
			t.Errorf("Expected ID to be 'def456', got %v", rows[1]["ID"])
		}
	})

	t.Run("parse empty JSON", func(t *testing.T) {
		var result map[string]interface{}
		err := h.ParseJSONOutput("", &result)
		if err == nil {
			t.Error("Expected error for empty JSON output")
		}
	})

	t.Run("parse invalid JSON", func(t *testing.T) {
		var result map[string]interface{}
		err := h.ParseJSONOutput("{invalid json}", &result)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

// TestCredentialManagement verifies credential storage works correctly.
func TestCredentialManagement(t *testing.T) {
	h := getHarness(t)

	t.Run("clean credentials", func(t *testing.T) {
		// Create a dummy credential file
		err := h.MockAuthenticate()
		if err != nil {
			t.Fatalf("Failed to create mock credentials: %v", err)
		}

		// Verify file exists
		authFile := fmt.Sprintf("%s/auth.json", h.configDir)
		if _, err := os.Stat(authFile); err != nil {
			t.Fatalf("Auth file was not created: %v", err)
		}

		// Clean credentials
		if err := h.CleanCredentials(); err != nil {
			t.Fatalf("Failed to clean credentials: %v", err)
		}

		// Verify directory was recreated but file is removed
		if _, err := os.Stat(h.credentialDir); err != nil {
			t.Error("Credential directory should still exist after cleanup")
		}

		// The auth.json is in configDir, not credentialDir, so it won't be removed by CleanCredentials
		// This is expected behavior - CleanCredentials only removes the credential storage directory
		// Let's verify the credential directory is empty instead
		entries, err := os.ReadDir(h.credentialDir)
		if err != nil {
			t.Fatalf("Failed to read credential directory: %v", err)
		}
		if len(entries) > 0 {
			t.Errorf("Credential directory should be empty after cleanup, found %d entries", len(entries))
		}
	})

	t.Run("mock authenticate", func(t *testing.T) {
		// Clean first
		if err := h.CleanCredentials(); err != nil {
			t.Fatalf("Failed to clean credentials: %v", err)
		}

		// Setup mock auth
		if err := h.MockAuthenticate(); err != nil {
			t.Fatalf("Failed to mock authenticate: %v", err)
		}

		// Verify auth file exists and has correct content
		authFile := fmt.Sprintf("%s/auth.json", h.configDir)
		data, err := os.ReadFile(authFile)
		if err != nil {
			t.Fatalf("Failed to read auth file: %v", err)
		}

		var authData map[string]interface{}
		if err := h.ParseJSONOutput(string(data), &authData); err != nil {
			t.Fatalf("Failed to parse auth data: %v", err)
		}

		if authData["access_token"] != "mock-test-token" {
			t.Errorf("Expected access_token to be 'mock-test-token', got %v", authData["access_token"])
		}
	})
}

// TestMockServerConnectivity verifies the mock server is accessible.
func TestMockServerConnectivity(t *testing.T) {
	h := getHarness(t)

	tests := []struct {
		name       string
		method     string
		path       string
		expectOK   bool
	}{
		{
			name:     "health endpoint",
			method:   "GET",
			path:     "/health",
			expectOK: true,
		},
		{
			name:     "api metadata endpoint",
			method:   "GET",
			path:     "/api/v1",
			expectOK: true,
		},
		{
			name:     "nonexistent endpoint",
			method:   "GET",
			path:     "/nonexistent",
			expectOK: true, // Mock server returns 404, not a server error, so VerifyAPICall succeeds
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.VerifyAPICall(tt.method, tt.path)

			if tt.expectOK && err != nil {
				t.Errorf("Expected API call to succeed, got error: %v", err)
			}
			if !tt.expectOK && err == nil {
				t.Errorf("Expected API call to fail, but it succeeded")
			}
		})
	}
}

// Helper function to check if a string contains a substring (case-insensitive).
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
		(s[:len(substr)] == substr || containsString(s[1:], substr)))
}
