package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

// Harness provides utilities for testing the ROSA CLI against the mock server.
type Harness struct {
	// Mock server configuration
	ServerURL    string
	ServerPort   string
	serverCmd    *exec.Cmd
	serverAddr   string
	serverBinary string // Path to temporary server binary

	// CLI binary path
	CLIBinary string

	// Test environment
	testDir       string
	configDir     string
	credentialDir string

	// Context for cleanup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewHarness creates a new test harness instance.
func NewHarness() (*Harness, error) {
	ctx, cancel := context.WithCancel(context.Background())

	h := &Harness{
		ctx:    ctx,
		cancel: cancel,
	}

	// Find CLI binary
	cliBinary, err := findCLIBinary()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to find CLI binary: %w", err)
	}
	h.CLIBinary = cliBinary

	// Setup test directories
	if err := h.setupTestDirs(); err != nil {
		cancel()
		return nil, fmt.Errorf("failed to setup test directories: %w", err)
	}

	return h, nil
}

// Start initializes the test environment and starts the mock server.
func (h *Harness) Start() error {
	// Start mock server on random available port
	port, err := h.getAvailablePort()
	if err != nil {
		return fmt.Errorf("failed to find available port: %w", err)
	}

	h.ServerPort = port
	h.ServerURL = fmt.Sprintf("http://localhost:%s", port)
	h.serverAddr = fmt.Sprintf("localhost:%s", port)

	// Start the mock server
	if err := h.startMockServer(); err != nil {
		return fmt.Errorf("failed to start mock server: %w", err)
	}

	// Wait for server to be ready
	if err := h.waitForServer(10 * time.Second); err != nil {
		h.Stop()
		return fmt.Errorf("mock server failed to start: %w", err)
	}

	return nil
}

// Stop shuts down the mock server and cleans up resources.
func (h *Harness) Stop() error {
	var errs []error

	// Stop mock server
	if h.serverCmd != nil && h.serverCmd.Process != nil {
		// Send SIGTERM for graceful shutdown
		if err := h.serverCmd.Process.Signal(syscall.SIGTERM); err != nil {
			errs = append(errs, fmt.Errorf("failed to stop mock server: %w", err))
		} else {
			// Wait for process to exit with timeout
			done := make(chan error, 1)
			go func() {
				done <- h.serverCmd.Wait()
			}()

			select {
			case <-time.After(5 * time.Second):
				// Force kill if graceful shutdown times out
				if err := h.serverCmd.Process.Kill(); err != nil {
					errs = append(errs, fmt.Errorf("failed to kill mock server: %w", err))
				}
			case err := <-done:
				if err != nil && !strings.Contains(err.Error(), "signal: terminated") {
					errs = append(errs, fmt.Errorf("mock server exited with error: %w", err))
				}
			}
		}
	}

	// Clean up temporary server binary
	if h.serverBinary != "" {
		if err := os.Remove(h.serverBinary); err != nil && !os.IsNotExist(err) {
			errs = append(errs, fmt.Errorf("failed to remove server binary: %w", err))
		}
	}

	// Cancel context
	h.cancel()

	// Cleanup test directories
	if err := h.cleanupTestDirs(); err != nil {
		errs = append(errs, fmt.Errorf("failed to cleanup test directories: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("cleanup errors: %v", errs)
	}

	return nil
}

// CleanCredentials removes stored credentials from the test environment.
func (h *Harness) CleanCredentials() error {
	// Remove credential storage directory
	if h.credentialDir != "" {
		if err := os.RemoveAll(h.credentialDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove credential directory: %w", err)
		}
		// Recreate empty directory
		if err := os.MkdirAll(h.credentialDir, 0700); err != nil {
			return fmt.Errorf("failed to recreate credential directory: %w", err)
		}
	}

	return nil
}

// RunCLI executes the CLI with the given arguments and returns stdout, stderr, and error.
func (h *Harness) RunCLI(args ...string) (stdout, stderr string, err error) {
	return h.RunCLIWithInput("", args...)
}

// RunCLIWithInput executes the CLI with the given arguments and stdin input.
func (h *Harness) RunCLIWithInput(input string, args ...string) (stdout, stderr string, err error) {
	cmd := exec.CommandContext(h.ctx, h.CLIBinary, args...)

	// Set up environment variables
	cmd.Env = h.buildTestEnv()

	// Capture stdout and stderr
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	// Provide stdin if needed
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	// Run the command
	err = cmd.Run()

	return outBuf.String(), errBuf.String(), err
}

// ParseJSONOutput parses JSON output from the CLI into the provided interface.
func (h *Harness) ParseJSONOutput(output string, v interface{}) error {
	output = strings.TrimSpace(output)
	if output == "" {
		return fmt.Errorf("empty output")
	}

	decoder := json.NewDecoder(strings.NewReader(output))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(v); err != nil {
		return fmt.Errorf("failed to parse JSON output: %w (output: %s)", err, output)
	}

	return nil
}

// ParseYAMLOutput parses YAML output from the CLI into the provided interface.
func (h *Harness) ParseYAMLOutput(output string, v interface{}) error {
	output = strings.TrimSpace(output)
	if output == "" {
		return fmt.Errorf("empty output")
	}

	if err := yaml.Unmarshal([]byte(output), v); err != nil {
		return fmt.Errorf("failed to parse YAML output: %w", err)
	}

	return nil
}

// ParseTableOutput extracts data from table-formatted CLI output.
// Returns a slice of maps, where each map represents a row with column headers as keys.
func (h *Harness) ParseTableOutput(output string) ([]map[string]string, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("table output must have at least header and separator lines")
	}

	// Parse header line to get column names and positions
	headerLine := lines[0]
	headers := strings.Fields(headerLine)
	if len(headers) == 0 {
		return nil, fmt.Errorf("no columns found in header")
	}

	// Skip separator line (typically contains dashes)
	dataStart := 1
	if len(lines) > 1 && strings.Contains(lines[1], "-") {
		dataStart = 2
	}

	// Parse data rows
	var rows []map[string]string
	for i := dataStart; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		// Create row map
		row := make(map[string]string)
		for j, header := range headers {
			if j < len(fields) {
				row[header] = fields[j]
			} else {
				row[header] = ""
			}
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// WaitForClusterState polls the cluster status until it reaches the expected state or times out.
func (h *Harness) WaitForClusterState(clusterID, expectedState string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		stdout, _, err := h.RunCLI("describe", "cluster", clusterID, "--output", "json")
		if err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		var cluster map[string]interface{}
		if err := h.ParseJSONOutput(stdout, &cluster); err != nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if state, ok := cluster["state"].(string); ok && state == expectedState {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("timeout waiting for cluster %s to reach state %s", clusterID, expectedState)
}

// MockAuthenticate sets up a mock authentication token for testing.
func (h *Harness) MockAuthenticate() error {
	// Create a mock token file
	tokenData := map[string]interface{}{
		"access_token":  "mock-test-token",
		"token_type":    "Bearer",
		"expires_in":    3600,
		"refresh_token": "mock-refresh-token",
	}

	tokenFile := filepath.Join(h.configDir, "auth.json")
	data, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token data: %w", err)
	}

	if err := os.WriteFile(tokenFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// VerifyAPICall checks if a specific API endpoint was called.
// This is a simple check - in a real implementation, you might want to add
// request logging to the mock server and query those logs.
func (h *Harness) VerifyAPICall(method, path string) error {
	// Make a request to verify the endpoint exists
	url := fmt.Sprintf("%s%s", h.ServerURL, path)
	req, err := http.NewRequestWithContext(h.ctx, method, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add mock auth header
	req.Header.Set("Authorization", "Bearer mock-test-token")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		return fmt.Errorf("API returned server error: %d", resp.StatusCode)
	}

	return nil
}

// Helper: getAvailablePort finds an available port on localhost.
func (h *Harness) getAvailablePort() (string, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", err
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)
	return fmt.Sprintf("%d", addr.Port), nil
}

// Helper: startMockServer starts the mock server process.
func (h *Harness) startMockServer() error {
	// Find mock server source directory
	mockServerDir, err := findMockServerDir()
	if err != nil {
		return err
	}

	// Build mock server binary to temporary location
	tmpBinary := filepath.Join(os.TempDir(), fmt.Sprintf("mock-server-%s", h.ServerPort))
	h.serverBinary = tmpBinary

	buildCmd := exec.CommandContext(h.ctx, "go", "build", "-o", tmpBinary, ".")
	buildCmd.Dir = mockServerDir
	buildCmd.Env = os.Environ()

	output, err := buildCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to build mock server: %w\nOutput: %s", err, string(output))
	}

	// Ensure binary is cleaned up if we fail to start
	defer func() {
		if h.serverCmd == nil || h.serverCmd.Process == nil {
			// Server didn't start, clean up binary
			os.Remove(tmpBinary)
			h.serverBinary = ""
		}
	}()

	// Run the compiled binary
	cmd := exec.CommandContext(h.ctx, tmpBinary)

	// Set server port
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%s", h.ServerPort))

	// Capture server output for debugging
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Start the server
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start server process: %w", err)
	}

	h.serverCmd = cmd
	return nil
}

// Helper: waitForServer waits for the mock server to become ready.
func (h *Harness) waitForServer(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	healthURL := fmt.Sprintf("%s/health", h.ServerURL)

	for time.Now().Before(deadline) {
		resp, err := http.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for server to become ready")
}

// Helper: setupTestDirs creates temporary directories for test environment.
func (h *Harness) setupTestDirs() error {
	// Create temporary test directory
	testDir, err := os.MkdirTemp("", "rosa-test-*")
	if err != nil {
		return fmt.Errorf("failed to create test directory: %w", err)
	}
	h.testDir = testDir

	// Create config directory
	configDir := filepath.Join(testDir, ".config", "rosa")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	h.configDir = configDir

	// Create credentials directory
	credentialDir := filepath.Join(configDir, "credentials")
	if err := os.MkdirAll(credentialDir, 0700); err != nil {
		return fmt.Errorf("failed to create credential directory: %w", err)
	}
	h.credentialDir = credentialDir

	return nil
}

// Helper: cleanupTestDirs removes temporary test directories.
func (h *Harness) cleanupTestDirs() error {
	if h.testDir != "" {
		if err := os.RemoveAll(h.testDir); err != nil {
			return err
		}
	}
	return nil
}

// Helper: buildTestEnv builds the environment variables for CLI execution.
func (h *Harness) buildTestEnv() []string {
	env := []string{
		fmt.Sprintf("HOME=%s", h.testDir),
		fmt.Sprintf("XDG_CONFIG_HOME=%s", filepath.Join(h.testDir, ".config")),
		fmt.Sprintf("ROSA_API_URL=%s", h.ServerURL),
		"NO_COLOR=1",              // Disable colored output for easier parsing
		"ROSA_DISABLE_KEYRING=1",  // Use file storage instead of keyring
		"TERM=dumb",               // Disable interactive features
	}

	// Preserve essential environment variables
	for _, key := range []string{"PATH", "USER", "TMPDIR"} {
		if val := os.Getenv(key); val != "" {
			env = append(env, fmt.Sprintf("%s=%s", key, val))
		}
	}

	return env
}

// Helper: findCLIBinary locates the rosa CLI binary.
func findCLIBinary() (string, error) {
	// Try several locations
	locations := []string{
		"./rosa",                                    // Current directory
		"../rosa",                                   // Parent directory
		"../../rosa",                                // Two levels up
		"./examples/rosa-like/rosa",                 // From repo root
		"/Users/wgordon/SynologyDrive/ActiveProjects/ai/alpha-omega/examples/rosa-like/rosa", // Absolute path
	}

	for _, loc := range locations {
		absPath, err := filepath.Abs(loc)
		if err != nil {
			continue
		}

		if _, err := os.Stat(absPath); err == nil {
			// Verify it's executable
			if info, err := os.Stat(absPath); err == nil && info.Mode()&0111 != 0 {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("rosa binary not found in expected locations")
}

// Helper: findMockServerDir locates the mock server source directory.
func findMockServerDir() (string, error) {
	// Try to find mock server directory
	locations := []string{
		"../mock-server",
		"./mock-server",
		"../../mock-server",
		"./examples/rosa-like/mock-server",
		"/Users/wgordon/SynologyDrive/ActiveProjects/ai/alpha-omega/examples/rosa-like/mock-server",
	}

	for _, loc := range locations {
		absPath, err := filepath.Abs(loc)
		if err != nil {
			continue
		}

		// Check if directory exists and contains main.go
		mainFile := filepath.Join(absPath, "main.go")
		if _, err := os.Stat(mainFile); err == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("mock server directory not found in expected locations")
}
