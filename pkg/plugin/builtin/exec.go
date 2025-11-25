// Package builtin provides built-in plugins that ship with CliForge.
package builtin

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/plugin"
)

// ExecPlugin executes external command-line tools with sandboxing.
type ExecPlugin struct {
	allowedCommands map[string]bool
	sandbox         bool
}

// NewExecPlugin creates a new exec plugin.
func NewExecPlugin(allowedCommands []string, sandbox bool) *ExecPlugin {
	allowed := make(map[string]bool)
	for _, cmd := range allowedCommands {
		allowed[cmd] = true
	}

	return &ExecPlugin{
		allowedCommands: allowed,
		sandbox:         sandbox,
	}
}

// Execute runs an external command.
func (p *ExecPlugin) Execute(ctx context.Context, input *plugin.PluginInput) (*plugin.PluginOutput, error) {
	// Validate input
	if input.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Check if command is allowed
	if len(p.allowedCommands) > 0 {
		cmdBase := filepath.Base(input.Command)
		if !p.allowedCommands[cmdBase] && !p.allowedCommands[input.Command] {
			return nil, fmt.Errorf("command '%s' is not in the allowed list", input.Command)
		}
	}

	// Validate command doesn't contain injection attempts
	if err := validateCommand(input.Command); err != nil {
		return nil, err
	}

	// Validate arguments
	for _, arg := range input.Args {
		if err := validateArgument(arg); err != nil {
			return nil, err
		}
	}

	startTime := time.Now()

	// Create the command
	cmd := exec.CommandContext(ctx, input.Command, input.Args...)

	// Set up stdin
	if input.Stdin != "" {
		cmd.Stdin = strings.NewReader(input.Stdin)
	}

	// Set up stdout/stderr capture
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Set environment variables
	if input.Env != nil {
		env := os.Environ()
		for k, v := range input.Env {
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
		cmd.Env = env
	}

	// Set working directory
	if input.WorkingDir != "" {
		// Validate working directory exists
		if info, err := os.Stat(input.WorkingDir); err != nil || !info.IsDir() {
			return nil, fmt.Errorf("invalid working directory: %s", input.WorkingDir)
		}
		cmd.Dir = input.WorkingDir
	}

	// Apply sandboxing if enabled
	if p.sandbox {
		if err := applySandbox(cmd); err != nil {
			return nil, fmt.Errorf("failed to apply sandbox: %w", err)
		}
	}

	// Execute the command
	err := cmd.Run()
	duration := time.Since(startTime)

	// Determine exit code
	exitCode := 0
	var execError string
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			execError = err.Error()
			exitCode = -1
		}
	}

	return &plugin.PluginOutput{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Error:    execError,
		Duration: duration,
		Metadata: map[string]string{
			"command": input.Command,
		},
	}, nil
}

// Validate checks if the plugin is properly configured.
func (p *ExecPlugin) Validate() error {
	// Exec plugin is always valid
	return nil
}

// Describe returns plugin metadata.
func (p *ExecPlugin) Describe() *plugin.PluginInfo {
	manifest := plugin.PluginManifest{
		Name:        "exec",
		Version:     "1.0.0",
		Type:        plugin.PluginTypeBuiltin,
		Description: "Execute external command-line tools",
		Author:      "CliForge",
		Permissions: []plugin.Permission{
			{
				Type:        plugin.PermissionExecute,
				Resource:    "*",
				Description: "Execute external commands",
			},
		},
	}

	return &plugin.PluginInfo{
		Manifest:     manifest,
		Capabilities: []string{"execute", "capture-output", "sandbox"},
		Status:       plugin.PluginStatusReady,
	}
}

// validateCommand validates a command for injection attempts.
func validateCommand(cmd string) error {
	// Check for shell metacharacters
	dangerous := []string{";", "|", "&", "$", "`", "(", ")", "<", ">", "\n", "\r"}
	for _, char := range dangerous {
		if strings.Contains(cmd, char) {
			return fmt.Errorf("command contains dangerous character: %s", char)
		}
	}
	return nil
}

// validateArgument validates a command argument.
func validateArgument(arg string) error {
	// Allow most characters in arguments, but check for obvious injection attempts
	// This is a basic check - more sophisticated validation may be needed
	if strings.Contains(arg, "\x00") {
		return fmt.Errorf("argument contains null byte")
	}
	return nil
}

// applySandbox applies sandboxing to a command (platform-specific).
func applySandbox(cmd *exec.Cmd) error {
	// Basic sandboxing: limit resource access
	// More sophisticated sandboxing (chroot, namespaces, etc.) would be platform-specific

	// For now, we just ensure no dangerous environment variables are inherited
	if cmd.Env == nil {
		cmd.Env = []string{}
	}

	// Filter out potentially dangerous env vars
	safeEnv := make([]string, 0)
	for _, envVar := range cmd.Env {
		if strings.HasPrefix(envVar, "PATH=") ||
			strings.HasPrefix(envVar, "HOME=") ||
			strings.HasPrefix(envVar, "USER=") ||
			strings.HasPrefix(envVar, "LANG=") {
			safeEnv = append(safeEnv, envVar)
		}
	}
	cmd.Env = safeEnv

	return nil
}

// ExecuteWithShell executes a command through a shell (use with caution).
func ExecuteWithShell(ctx context.Context, command string, shell string) (*plugin.PluginOutput, error) {
	if shell == "" {
		shell = "/bin/sh"
	}

	// Validate shell path
	if !filepath.IsAbs(shell) {
		return nil, fmt.Errorf("shell must be an absolute path")
	}

	startTime := time.Now()

	cmd := exec.CommandContext(ctx, shell, "-c", command)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(startTime)

	exitCode := 0
	var execError string
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			execError = err.Error()
			exitCode = -1
		}
	}

	return &plugin.PluginOutput{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: exitCode,
		Error:    execError,
		Duration: duration,
		Metadata: map[string]string{
			"shell":   shell,
			"command": command,
		},
	}, nil
}
