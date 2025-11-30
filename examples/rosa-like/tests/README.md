# ROSA CLI Integration Tests

This directory contains the integration test harness for the ROSA CLI.

## Overview

The test harness provides infrastructure for end-to-end testing of the ROSA CLI against a mock API server. It handles:

- Mock server lifecycle management (start/stop, port allocation)
- CLI execution and output capture
- Output parsing (JSON, YAML, table formats)
- Credential management
- Test environment isolation

## Files

- **harness.go** - Core test harness implementation with helper functions
- **main_test.go** - TestMain setup/teardown and basic validation tests
- **types.go** - Shared type definitions for API responses

## Running Tests

```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestHarnessSetup

# Run with timeout
go test -v -timeout 5m
```

## Test Harness Features

### 1. Mock Server Management

The harness automatically:
- Builds the mock server from source
- Finds an available port to avoid conflicts
- Starts the server before tests run
- Waits for the server to be ready
- Stops the server and cleans up after tests

### 2. CLI Execution Helpers

```go
// Run CLI command and capture output
stdout, stderr, err := h.RunCLI("list", "clusters")

// Run CLI with stdin input
stdout, stderr, err := h.RunCLIWithInput("yes\n", "delete", "cluster", "abc123")
```

### 3. Output Parsing

```go
// Parse JSON output
var cluster map[string]interface{}
err := h.ParseJSONOutput(stdout, &cluster)

// Parse YAML output
var data map[string]interface{}
err := h.ParseYAMLOutput(stdout, &data)

// Parse table output
rows, err := h.ParseTableOutput(stdout)
// rows is []map[string]string, e.g., [{"ID": "abc123", "NAME": "cluster1"}]
```

### 4. Test Utilities

```go
// Clean credentials before test
h.CleanCredentials()

// Setup mock authentication
h.MockAuthenticate()

// Wait for async operations
err := h.WaitForClusterState("cluster-id", "ready", 30*time.Second)

// Verify API was called
err := h.VerifyAPICall("GET", "/api/v1/clusters")
```

## Writing New Tests

### Basic Test Structure

```go
func TestMyFeature(t *testing.T) {
    h := getHarness(t)

    // Clean state before test
    cleanCredentials(t)

    // Setup authentication if needed
    mockAuthenticate(t)

    // Run CLI command
    stdout, stderr, err := h.RunCLI("list", "clusters")
    if err != nil {
        t.Fatalf("Command failed: %v\nStderr: %s", err, stderr)
    }

    // Parse and verify output
    var result map[string]interface{}
    if err := h.ParseJSONOutput(stdout, &result); err != nil {
        t.Fatalf("Failed to parse output: %v", err)
    }

    // Make assertions
    if result["total"] != 0 {
        t.Errorf("Expected 0 clusters, got %v", result["total"])
    }
}
```

### Using Subtests

```go
func TestClusters(t *testing.T) {
    h := getHarness(t)

    t.Run("create", func(t *testing.T) {
        mockAuthenticate(t)
        stdout, _, err := h.RunCLI("create", "cluster", "--name", "test")
        // ...
    })

    t.Run("list", func(t *testing.T) {
        mockAuthenticate(t)
        stdout, _, err := h.RunCLI("list", "clusters")
        // ...
    })

    t.Run("delete", func(t *testing.T) {
        mockAuthenticate(t)
        stdout, _, err := h.RunCLI("delete", "cluster", "test")
        // ...
    })
}
```

## Environment Variables

The harness sets these environment variables for CLI execution:

- `HOME` - Temporary test directory
- `XDG_CONFIG_HOME` - Test config directory
- `ROSA_API_URL` - Mock server URL (random port)
- `NO_COLOR=1` - Disable colored output
- `ROSA_DISABLE_KEYRING=1` - Use file storage instead of keyring
- `TERM=dumb` - Disable interactive features

## Test Isolation

Each test run gets:
- A unique temporary directory for config/credentials
- An isolated mock server instance on a random port
- Clean credential state (if using `cleanCredentials()`)

The harness ensures cleanup happens even if tests fail or panic.

## Debugging Tests

```bash
# Run with verbose output
go test -v

# Run specific test with output
go test -v -run TestHarnessSetup

# Keep test artifacts (disable cleanup)
# Modify main_test.go to comment out cleanup in TestMain

# Check mock server logs
# Server logs are printed to stdout during test execution
```

## CI/CD Integration

Tests can run in CI with:

```bash
# Build CLI first
make build

# Run tests
cd tests && go test -v -timeout 10m
```

Requirements:
- Go 1.24 or later
- CLI binary at `../rosa`
- Mock server source at `../mock-server/`

## Troubleshooting

### Port Already in Use

The harness automatically finds available ports. If you see port conflicts, ensure:
- No other test instances are running
- No manual mock server processes are running

### Server Failed to Start

Check that:
- Mock server builds successfully: `cd ../mock-server && go build`
- Port is not blocked by firewall
- `go` is in PATH

### CLI Binary Not Found

Ensure the CLI is built:
```bash
cd .. && make build
# Or: go build -o rosa ./cmd/rosa
```

### Tests Hang

If tests hang, check:
- Mock server is responding: `curl http://localhost:PORT/health`
- No deadlocks in CLI code
- Use `-timeout` flag to prevent infinite hangs

## Next Steps

After the harness is working, add tests for:

1. **Authentication flows** - login, logout, token refresh
2. **Cluster operations** - create, list, describe, delete
3. **Nested resources** - machine pools, identity providers, addons
4. **Error handling** - invalid inputs, API errors, network failures
5. **Output formats** - JSON, YAML, table formatting
6. **Interactive prompts** - confirmation dialogs, selection menus

See the task implementation plan for detailed test scenarios.
