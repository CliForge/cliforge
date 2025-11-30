# Preflight Check Implementation

This document describes the preflight check system for the ROSA CLI implementation.

## Overview

The preflight check system executes validation checks before the main operation runs. It supports both required and optional checks, with configurable endpoints and user-friendly output.

## Files

- `preflight.go` - Core preflight check implementation (293 lines)
- `preflight_test.go` - Unit tests for preflight checks (903 lines)
- `preflight_integration_test.go` - Integration tests (316 lines)

## Features

### 1. Check Execution
- Executes checks defined in `x-cli-preflight` extension before main operation
- Supports GET, POST, and other HTTP methods
- Handles both relative and absolute endpoint URLs
- Applies authentication headers automatically

### 2. Check Configuration
Each check has:
- **name** - Human-readable identifier
- **description** - What the check is validating
- **endpoint** - API endpoint to call
- **method** - HTTP method (defaults to GET)
- **required** - If true, failure blocks the operation

### 3. Failure Handling
- **Required checks**: Fail fast - stop on first failure, block operation
- **Optional checks**: Continue execution, show warnings, allow operation to proceed
- User-friendly error messages extracted from response (message, error, or detail fields)

### 4. User Interface
- Section header: "Running preflight checks"
- Success indicator: ✓ (green)
- Failure indicator: ✗ (red for required, yellow for optional)
- Summary message for optional failures

### 5. Progress Indicators
Two modes available:
- **executePreflightChecks**: Simple success/failure output
- **executePreflightChecksWithProgress**: Spinner for each check (optional)

## Usage Example

### OpenAPI Specification

```yaml
paths:
  /api/v1/clusters:
    post:
      operationId: createCluster
      summary: Create a new cluster
      x-cli-preflight:
        - name: verify-aws-credentials
          description: Verifying AWS credentials...
          endpoint: /api/v1/aws/credentials/verify
          method: POST
          required: false
        - name: verify-quota
          description: Checking AWS service quotas...
          endpoint: /api/v1/aws/quotas/verify
          method: POST
          required: false
```

### Execution Flow

1. User runs: `rosa-cli cluster create`
2. System executes preflight checks:
   - Verifying AWS credentials... ✓
   - Checking AWS service quotas... ✓
3. Main operation executes: Create cluster
4. Output: Cluster created successfully

### Error Handling

#### Required Check Failure
```
Running preflight checks

✗ verify-credentials - Verifying credentials: HTTP 403: Invalid credentials

Error: preflight checks failed: required preflight check failed: verify-credentials - HTTP 403: Invalid credentials
```

#### Optional Check Failure
```
Running preflight checks

✓ verify-credentials - Verifying credentials
⚠ verify-quota - Checking quotas: HTTP 503: Service temporarily unavailable

⚠ Some optional checks failed but proceeding with operation

Creating cluster...
```

## Integration

The preflight system integrates seamlessly with the executor:

```go
// In executeHTTPOperation
if len(op.CLIPreflight) > 0 {
    if _, err := e.executePreflightChecks(ctx, op.CLIPreflight); err != nil {
        return fmt.Errorf("preflight checks failed: %w", err)
    }
}
```

## Test Coverage

- **Unit tests**: 72.7% statement coverage
- **Test scenarios**:
  - Successful checks (GET, POST, various status codes)
  - Failed checks (error messages, HTTP errors)
  - Required vs optional check behavior
  - Authentication integration
  - Network errors and timeouts
  - Relative and absolute URLs
  - Integration with main operation flow

## Performance

- Checks execute sequentially
- Fail-fast for required checks (no wasted API calls)
- Typical check execution: <100ms per check
- Benchmark: ~3.5ms per check (local mock server)

## Future Enhancements

Potential improvements:
1. Parallel check execution for independent checks
2. Retry logic for transient failures
3. Caching of check results (with TTL)
4. Skip flags for individual checks
5. Dry-run mode (show what checks would run)
6. Custom check timeout configuration
