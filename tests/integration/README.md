# Integration Tests for CliForge v0.9.0

This directory contains comprehensive integration tests for CliForge's end-to-end functionality.

## Structure

- `integration_test.go` - Full CLI lifecycle tests (build → run → commands)
- `workflow_test.go` - Multi-step workflow execution tests
- `auth_test.go` - Authentication flow tests (OAuth2, API key)
- `plugin_test.go` - Plugin execution in workflows
- `streaming_test.go` - SSE/WebSocket streaming tests
- `state_test.go` - Context and history persistence tests

## Test Fixtures

Located in `../fixtures/`:
- `test-api.yaml` - Test OpenAPI specification
- `workflow-api.yaml` - Workflow test API spec
- `config.yaml` - Test configuration
- `plugin-manifest.yaml` - Plugin manifest for testing

## Test Helpers

Located in `../helpers/`:
- `mock_server.go` - Mock HTTP servers
- `oauth_server.go` - Mock OAuth2 servers
- `cli_builder.go` - Test CLI builders and execution
- `assertions.go` - Test assertion helpers
- `utils.go` - Utility functions

## Running Tests

```bash
# Run all integration tests
go test ./tests/integration/...

# Run specific test file
go test ./tests/integration/integration_test.go

# Run with verbose output
go test -v ./tests/integration/...

# Run specific test
go test -v ./tests/integration/... -run TestFullCLILifecycle
```

## Test Coverage

The integration tests cover:

1. **CLI Lifecycle**:
   - Binary building and compilation
   - Command execution
   - Version and help output
   - Error handling
   - Configuration file loading

2. **Workflow Execution**:
   - Sequential workflows
   - Conditional execution
   - Parallel execution
   - Retry logic
   - Rollback functionality

3. **Authentication**:
   - OAuth2 authorization code flow
   - Token refresh
   - Token storage (file, memory, keyring)
   - API key authentication
   - Basic authentication

4. **Plugin System**:
   - Plugin registration
   - Permission validation
   - Plugin execution
   - Workflow integration
   - Timeout handling

5. **Streaming**:
   - Server-Sent Events (SSE)
   - WebSocket connections
   - Progress updates
   - Error handling
   - Reconnection logic

6. **State Management**:
   - Context creation and switching
   - Variable storage
   - Recent values tracking
   - Resource preferences
   - Session tracking
   - State persistence

## Notes

These integration tests verify end-to-end functionality by:
- Building actual CLI binaries
- Executing real commands
- Using mock servers for external dependencies
- Testing with real file I/O
- Verifying state persistence

Test data and mock servers are isolated per test to ensure independence.
