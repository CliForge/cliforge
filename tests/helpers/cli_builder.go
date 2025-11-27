package helpers

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// CLIBuilder helps build and test CLI binaries.
type CLIBuilder struct {
	t             *testing.T
	workDir       string
	buildDir      string
	binaryPath    string
	env           map[string]string
	buildTimeout  time.Duration
	cleanupOnFail bool
}

// NewCLIBuilder creates a new CLI builder.
func NewCLIBuilder(t *testing.T) *CLIBuilder {
	workDir := t.TempDir()
	buildDir := filepath.Join(workDir, "build")

	return &CLIBuilder{
		t:             t,
		workDir:       workDir,
		buildDir:      buildDir,
		env:           make(map[string]string),
		buildTimeout:  5 * time.Minute,
		cleanupOnFail: true,
	}
}

// WithEnv sets an environment variable for the build.
func (cb *CLIBuilder) WithEnv(key, value string) *CLIBuilder {
	cb.env[key] = value
	return cb
}

// WithBuildTimeout sets the build timeout.
func (cb *CLIBuilder) WithBuildTimeout(timeout time.Duration) *CLIBuilder {
	cb.buildTimeout = timeout
	return cb
}

// Build builds the CLI binary from the specified package.
func (cb *CLIBuilder) Build(pkgPath string) error {
	cb.t.Helper()

	// Create build directory
	if err := os.MkdirAll(cb.buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}

	// Determine binary name
	binaryName := "testcli"
	cb.binaryPath = filepath.Join(cb.buildDir, binaryName)

	// Build command
	ctx, cancel := context.WithTimeout(context.Background(), cb.buildTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "build", "-o", cb.binaryPath, pkgPath)
	cmd.Dir = cb.workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range cb.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run build
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("build failed: %w\nstdout: %s\nstderr: %s", err, stdout.String(), stderr.String())
	}

	// Verify binary exists
	if _, err := os.Stat(cb.binaryPath); os.IsNotExist(err) {
		return fmt.Errorf("binary not found at %s", cb.binaryPath)
	}

	return nil
}

// BinaryPath returns the path to the built binary.
func (cb *CLIBuilder) BinaryPath() string {
	return cb.binaryPath
}

// WorkDir returns the working directory.
func (cb *CLIBuilder) WorkDir() string {
	return cb.workDir
}

// Run executes the CLI with the given arguments.
func (cb *CLIBuilder) Run(args ...string) *CLIResult {
	return cb.RunWithContext(context.Background(), args...)
}

// RunWithContext executes the CLI with the given arguments and context.
func (cb *CLIBuilder) RunWithContext(ctx context.Context, args ...string) *CLIResult {
	cb.t.Helper()

	cmd := exec.CommandContext(ctx, cb.binaryPath, args...)
	cmd.Dir = cb.workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range cb.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &CLIResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
		Error:    err,
	}
}

// RunWithInput executes the CLI with stdin input.
func (cb *CLIBuilder) RunWithInput(stdin string, args ...string) *CLIResult {
	cb.t.Helper()

	cmd := exec.Command(cb.binaryPath, args...)
	cmd.Dir = cb.workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range cb.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set stdin
	cmd.Stdin = strings.NewReader(stdin)

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run command
	start := time.Now()
	err := cmd.Run()
	duration := time.Since(start)

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		}
	}

	return &CLIResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Duration: duration,
		Error:    err,
	}
}

// RunAsync executes the CLI asynchronously.
func (cb *CLIBuilder) RunAsync(args ...string) *AsyncCLI {
	cb.t.Helper()

	cmd := exec.Command(cb.binaryPath, args...)
	cmd.Dir = cb.workDir

	// Set environment
	cmd.Env = os.Environ()
	for k, v := range cb.env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Create pipes
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()
	stdin, _ := cmd.StdinPipe()

	async := &AsyncCLI{
		cmd:    cmd,
		stdout: stdout,
		stderr: stderr,
		stdin:  stdin,
		done:   make(chan error, 1),
	}

	// Start command
	if err := cmd.Start(); err != nil {
		cb.t.Fatalf("failed to start async command: %v", err)
	}

	// Wait for completion in background
	go func() {
		async.done <- cmd.Wait()
	}()

	return async
}

// CreateConfigFile creates a config file in the work directory.
func (cb *CLIBuilder) CreateConfigFile(name, content string) (string, error) {
	configPath := filepath.Join(cb.workDir, name)
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to create config file: %w", err)
	}
	return configPath, nil
}

// CLIResult represents the result of a CLI execution.
type CLIResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
	Error    error
}

// Success returns true if the command succeeded (exit code 0).
func (r *CLIResult) Success() bool {
	return r.ExitCode == 0
}

// Failed returns true if the command failed (non-zero exit code).
func (r *CLIResult) Failed() bool {
	return r.ExitCode != 0
}

// AssertSuccess asserts that the command succeeded.
func (r *CLIResult) AssertSuccess(t *testing.T) {
	t.Helper()
	if !r.Success() {
		t.Fatalf("Command failed with exit code %d\nStdout: %s\nStderr: %s", r.ExitCode, r.Stdout, r.Stderr)
	}
}

// AssertFailed asserts that the command failed.
func (r *CLIResult) AssertFailed(t *testing.T) {
	t.Helper()
	if r.Success() {
		t.Fatalf("Expected command to fail, but it succeeded\nStdout: %s\nStderr: %s", r.Stdout, r.Stderr)
	}
}

// AssertContains asserts that stdout contains the given substring.
func (r *CLIResult) AssertContains(t *testing.T, substr string) {
	t.Helper()
	if !strings.Contains(r.Stdout, substr) {
		t.Fatalf("Expected stdout to contain %q, but it didn't\nStdout: %s", substr, r.Stdout)
	}
}

// AssertNotContains asserts that stdout does not contain the given substring.
func (r *CLIResult) AssertNotContains(t *testing.T, substr string) {
	t.Helper()
	if strings.Contains(r.Stdout, substr) {
		t.Fatalf("Expected stdout not to contain %q, but it did\nStdout: %s", substr, r.Stdout)
	}
}

// AssertStderrContains asserts that stderr contains the given substring.
func (r *CLIResult) AssertStderrContains(t *testing.T, substr string) {
	t.Helper()
	if !strings.Contains(r.Stderr, substr) {
		t.Fatalf("Expected stderr to contain %q, but it didn't\nStderr: %s", substr, r.Stderr)
	}
}

// AssertExitCode asserts the exit code matches.
func (r *CLIResult) AssertExitCode(t *testing.T, expected int) {
	t.Helper()
	if r.ExitCode != expected {
		t.Fatalf("Expected exit code %d, got %d\nStdout: %s\nStderr: %s", expected, r.ExitCode, r.Stdout, r.Stderr)
	}
}

// AsyncCLI represents an asynchronously running CLI process.
type AsyncCLI struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
	stderr io.ReadCloser
	stdin  io.WriteCloser
	done   chan error
}

// Wait waits for the command to complete.
func (a *AsyncCLI) Wait() error {
	return <-a.done
}

// WaitWithTimeout waits for the command to complete with a timeout.
func (a *AsyncCLI) WaitWithTimeout(timeout time.Duration) error {
	select {
	case err := <-a.done:
		return err
	case <-time.After(timeout):
		_ = a.Kill()
		return fmt.Errorf("command timed out after %s", timeout)
	}
}

// Kill terminates the process.
func (a *AsyncCLI) Kill() error {
	if a.cmd.Process != nil {
		return a.cmd.Process.Kill()
	}
	return nil
}

// WriteStdin writes to the process stdin.
func (a *AsyncCLI) WriteStdin(data string) error {
	_, err := a.stdin.Write([]byte(data))
	return err
}

// CloseStdin closes the stdin pipe.
func (a *AsyncCLI) CloseStdin() error {
	return a.stdin.Close()
}

// ReadStdout reads from stdout.
func (a *AsyncCLI) ReadStdout() (string, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, a.stdout)
	return buf.String(), err
}

// ReadStderr reads from stderr.
func (a *AsyncCLI) ReadStderr() (string, error) {
	buf := new(bytes.Buffer)
	_, err := io.Copy(buf, a.stderr)
	return buf.String(), err
}
