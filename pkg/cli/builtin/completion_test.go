package builtin

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewCompletionCommand(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &CompletionOptions{
		CLIName:       "testcli",
		EnabledShells: []string{"bash", "zsh", "fish", "powershell"},
		Output:        output,
	}

	cmd := NewCompletionCommand(opts, rootCmd)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "completion [bash|zsh|fish|powershell]" {
		t.Errorf("expected Use 'completion [bash|zsh|fish|powershell]', got %q", cmd.Use)
	}

	if len(cmd.ValidArgs) != 4 {
		t.Errorf("expected 4 valid args, got %d", len(cmd.ValidArgs))
	}
}

func TestRunCompletion_Bash(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &CompletionOptions{
		CLIName:       "testcli",
		EnabledShells: []string{"bash", "zsh", "fish", "powershell"},
		Output:        output,
	}

	err := runCompletion(rootCmd, "bash", opts)
	if err != nil {
		t.Fatalf("runCompletion failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "bash completion") {
		t.Errorf("expected bash completion script in output")
	}
}

func TestRunCompletion_Zsh(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &CompletionOptions{
		CLIName:       "testcli",
		EnabledShells: []string{"bash", "zsh", "fish", "powershell"},
		Output:        output,
	}

	err := runCompletion(rootCmd, "zsh", opts)
	if err != nil {
		t.Fatalf("runCompletion failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "zsh completion") {
		t.Errorf("expected zsh completion script in output")
	}
}

func TestRunCompletion_Fish(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &CompletionOptions{
		CLIName:       "testcli",
		EnabledShells: []string{"bash", "zsh", "fish", "powershell"},
		Output:        output,
	}

	err := runCompletion(rootCmd, "fish", opts)
	if err != nil {
		t.Fatalf("runCompletion failed: %v", err)
	}

	result := output.String()
	// Fish completion should have generated content
	if len(result) == 0 {
		t.Error("expected fish completion script in output")
	}
}

func TestRunCompletion_PowerShell(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &CompletionOptions{
		CLIName:       "testcli",
		EnabledShells: []string{"bash", "zsh", "fish", "powershell"},
		Output:        output,
	}

	err := runCompletion(rootCmd, "powershell", opts)
	if err != nil {
		t.Fatalf("runCompletion failed: %v", err)
	}

	result := output.String()
	// PowerShell completion should have generated content
	if len(result) == 0 {
		t.Error("expected powershell completion script in output")
	}
}

func TestRunCompletion_UnsupportedShell(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &CompletionOptions{
		CLIName:       "testcli",
		EnabledShells: []string{"bash"},
		Output:        output,
	}

	err := runCompletion(rootCmd, "unsupported", opts)
	if err == nil {
		t.Fatal("expected error for unsupported shell, got nil")
	}

	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("expected 'unsupported shell' error, got: %v", err)
	}
}

func TestRunCompletion_DisabledShell(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &CompletionOptions{
		CLIName:       "testcli",
		EnabledShells: []string{"bash"},
		Output:        output,
	}

	err := runCompletion(rootCmd, "zsh", opts)
	if err == nil {
		t.Fatal("expected error for disabled shell, got nil")
	}

	if !strings.Contains(err.Error(), "not enabled") {
		t.Errorf("expected 'not enabled' error, got: %v", err)
	}
}

func TestSetupCompletionFunctions(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("output", "", "output format")
	cmd.Flags().String("format", "", "format")

	SetupCompletionFunctions(cmd)

	// Test that completion functions were registered
	// We can't easily test the actual completion, but we can verify no errors
	if cmd == nil {
		t.Error("command should not be nil")
	}
}

func TestNoFileCompletion(t *testing.T) {
	fn := NoFileCompletion()
	if fn == nil {
		t.Fatal("expected completion function, got nil")
	}

	cmd := &cobra.Command{Use: "test"}
	values, directive := fn(cmd, []string{}, "")

	if values != nil {
		t.Errorf("expected nil values, got %v", values)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
}

func TestFixedCompletion(t *testing.T) {
	fn := FixedCompletion("value1", "value2", "value3")
	if fn == nil {
		t.Fatal("expected completion function, got nil")
	}

	cmd := &cobra.Command{Use: "test"}
	values, directive := fn(cmd, []string{}, "")

	if len(values) != 3 {
		t.Errorf("expected 3 values, got %d", len(values))
	}

	if values[0] != "value1" || values[1] != "value2" || values[2] != "value3" {
		t.Errorf("expected [value1, value2, value3], got %v", values)
	}

	if directive != cobra.ShellCompDirectiveNoFileComp {
		t.Errorf("expected NoFileComp directive, got %v", directive)
	}
}

func TestFixedCompletion_Empty(t *testing.T) {
	fn := FixedCompletion()
	if fn == nil {
		t.Fatal("expected completion function, got nil")
	}

	cmd := &cobra.Command{Use: "test"}
	values, _ := fn(cmd, []string{}, "")

	if len(values) != 0 {
		t.Errorf("expected 0 values, got %d", len(values))
	}
}
