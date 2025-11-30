// Package executor provides preflight check execution for OpenAPI operations.
package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/progress"
	"github.com/pterm/pterm"
)

// PreflightResult represents the result of a single preflight check.
type PreflightResult struct {
	Name        string
	Description string
	Passed      bool
	Required    bool
	Error       error
	Response    *http.Response
	Body        []byte
}

// PreflightResults contains the results of all preflight checks.
type PreflightResults struct {
	Checks  []*PreflightResult
	AllPass bool
	Failed  []*PreflightResult
}

// executePreflightChecks runs all preflight checks for an operation.
// Returns an error if any required check fails; otherwise returns results.
func (e *Executor) executePreflightChecks(ctx context.Context, checks []*openapi.CLIPreflight) (*PreflightResults, error) {
	if len(checks) == 0 {
		return &PreflightResults{
			AllPass: true,
			Checks:  []*PreflightResult{},
			Failed:  []*PreflightResult{},
		}, nil
	}

	results := &PreflightResults{
		Checks:  make([]*PreflightResult, 0, len(checks)),
		Failed:  make([]*PreflightResult, 0),
		AllPass: true,
	}

	// Print section header
	pterm.Println()
	pterm.DefaultSection.Println("Running preflight checks")
	pterm.Println()

	// Execute each check
	for _, check := range checks {
		result := e.executePreflightCheck(ctx, check)
		results.Checks = append(results.Checks, result)

		// Print result
		if result.Passed {
			pterm.Success.Printf("%s - %s\n", check.Name, check.Description)
		} else {
			results.AllPass = false
			results.Failed = append(results.Failed, result)

			if result.Required {
				pterm.Error.Printf("%s - %s: %v\n", check.Name, check.Description, result.Error)
				// Fail fast on required checks
				pterm.Println()
				return results, fmt.Errorf("required preflight check failed: %s - %v", check.Name, result.Error)
			} else {
				pterm.Warning.Printf("%s - %s: %v\n", check.Name, check.Description, result.Error)
			}
		}
	}

	pterm.Println()

	// Summary for optional check failures
	if !results.AllPass && len(results.Failed) > 0 {
		hasRequired := false
		for _, f := range results.Failed {
			if f.Required {
				hasRequired = true
				break
			}
		}
		if !hasRequired {
			pterm.Warning.Printfln("Some optional checks failed but proceeding with operation")
			pterm.Println()
		}
	}

	return results, nil
}

// executePreflightCheck runs a single preflight check.
func (e *Executor) executePreflightCheck(ctx context.Context, check *openapi.CLIPreflight) *PreflightResult {
	result := &PreflightResult{
		Name:        check.Name,
		Description: check.Description,
		Required:    check.Required,
		Passed:      false,
	}

	// Default method to GET if not specified
	method := check.Method
	if method == "" {
		method = "GET"
	}
	method = strings.ToUpper(method)

	// Build URL
	url := check.Endpoint
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		// Relative URL, prepend base URL
		baseURL := e.baseURL
		if baseURL == "" && len(e.spec.Spec.Servers) > 0 {
			baseURL = e.spec.Spec.Servers[0].URL
		}
		url = strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(url, "/")
	}

	// Create request
	var bodyReader io.Reader
	if method == "POST" || method == "PUT" || method == "PATCH" {
		// For write methods, send empty JSON body
		bodyReader = bytes.NewReader([]byte("{}"))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		result.Error = fmt.Errorf("failed to create request: %w", err)
		return result
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Apply authentication
	if e.authManager != nil {
		if err := e.applyAuth(ctx, req); err != nil {
			result.Error = fmt.Errorf("failed to apply authentication: %w", err)
			return result
		}
	}

	// Execute request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		result.Error = fmt.Errorf("request failed: %w", err)
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = fmt.Errorf("failed to read response: %w", err)
		return result
	}

	result.Response = resp
	result.Body = body

	// Check HTTP status
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Passed = true
		return result
	}

	// Check failed - extract error message
	result.Error = e.extractPreflightError(resp, body)
	return result
}

// extractPreflightError extracts a user-friendly error message from the response.
func (e *Executor) extractPreflightError(resp *http.Response, body []byte) error {
	// Try to parse as JSON error
	var errorData map[string]interface{}
	if err := json.Unmarshal(body, &errorData); err == nil {
		// Look for common error message fields
		if msg, ok := errorData["message"].(string); ok {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
		}
		if msg, ok := errorData["error"].(string); ok {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
		}
		if msg, ok := errorData["detail"].(string); ok {
			return fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
		}
	}

	// Fallback to status code and raw body
	if len(body) > 0 && len(body) < 200 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("HTTP %d", resp.StatusCode)
}

// executePreflightChecksWithProgress runs preflight checks with progress indicators.
// This is an alternative to executePreflightChecks that uses spinners for each check.
func (e *Executor) executePreflightChecksWithProgress(ctx context.Context, checks []*openapi.CLIPreflight) (*PreflightResults, error) {
	if len(checks) == 0 {
		return &PreflightResults{
			AllPass: true,
			Checks:  []*PreflightResult{},
			Failed:  []*PreflightResult{},
		}, nil
	}

	results := &PreflightResults{
		Checks:  make([]*PreflightResult, 0, len(checks)),
		Failed:  make([]*PreflightResult, 0),
		AllPass: true,
	}

	// Print section header
	pterm.Println()
	pterm.DefaultSection.Println("Running preflight checks")
	pterm.Println()

	// Execute each check with spinner
	for _, check := range checks {
		var spinner progress.Progress
		if e.progressMgr != nil {
			var err error
			spinner, err = e.progressMgr.StartProgress(check.Description, 0)
			if err == nil && spinner != nil {
				defer func() { _ = spinner.Stop() }()
			}
		}

		result := e.executePreflightCheck(ctx, check)
		results.Checks = append(results.Checks, result)

		// Update spinner and display result
		if result.Passed {
			if spinner != nil {
				_ = spinner.Success(fmt.Sprintf("%s - %s", check.Name, check.Description))
			} else {
				pterm.Success.Printf("%s - %s\n", check.Name, check.Description)
			}
		} else {
			results.AllPass = false
			results.Failed = append(results.Failed, result)

			if spinner != nil {
				_ = spinner.Failure(fmt.Sprintf("%s - %s: %v", check.Name, check.Description, result.Error))
			}

			if result.Required {
				if spinner == nil {
					pterm.Error.Printf("%s - %s: %v\n", check.Name, check.Description, result.Error)
				}
				// Fail fast on required checks
				pterm.Println()
				return results, fmt.Errorf("required preflight check failed: %s - %v", check.Name, result.Error)
			} else {
				if spinner == nil {
					pterm.Warning.Printf("%s - %s: %v\n", check.Name, check.Description, result.Error)
				}
			}
		}
	}

	pterm.Println()

	// Summary for optional check failures
	if !results.AllPass && len(results.Failed) > 0 {
		hasRequired := false
		for _, f := range results.Failed {
			if f.Required {
				hasRequired = true
				break
			}
		}
		if !hasRequired {
			pterm.Warning.Printfln("Some optional checks failed but proceeding with operation")
			pterm.Println()
		}
	}

	return results, nil
}
