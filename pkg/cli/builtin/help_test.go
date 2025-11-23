package builtin

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestNewHelpCommand(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}
	rootCmd.AddCommand(&cobra.Command{Use: "subcommand", Short: "A subcommand"})

	opts := &HelpOptions{
		ShowExamples: true,
		ShowAliases:  true,
		Output:       output,
	}

	cmd := NewHelpCommand(opts, rootCmd)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "help [command]" {
		t.Errorf("expected Use 'help [command]', got %q", cmd.Use)
	}
}

func TestShowCustomHelp(t *testing.T) {
	output := &bytes.Buffer{}
	targetCmd := &cobra.Command{
		Use:     "test",
		Short:   "Test command",
		Example: "test --flag value",
		Aliases: []string{"t", "tst"},
	}

	opts := &HelpOptions{
		ShowExamples: true,
		ShowAliases:  true,
		Output:       output,
	}

	err := showCustomHelp(targetCmd, opts)
	if err != nil {
		t.Fatalf("showCustomHelp failed: %v", err)
	}

	result := output.String()

	if !strings.Contains(result, "Examples:") {
		t.Errorf("expected 'Examples:' in output, got: %s", result)
	}

	if !strings.Contains(result, "test --flag value") {
		t.Errorf("expected example in output, got: %s", result)
	}

	if !strings.Contains(result, "Aliases:") {
		t.Errorf("expected 'Aliases:' in output, got: %s", result)
	}

	if !strings.Contains(result, "t, tst") {
		t.Errorf("expected aliases in output, got: %s", result)
	}
}

func TestShowCustomHelp_NoExamples(t *testing.T) {
	output := &bytes.Buffer{}
	targetCmd := &cobra.Command{
		Use:   "test",
		Short: "Test command",
	}

	opts := &HelpOptions{
		ShowExamples: true,
		ShowAliases:  false,
		Output:       output,
	}

	err := showCustomHelp(targetCmd, opts)
	if err != nil {
		t.Fatalf("showCustomHelp failed: %v", err)
	}

	result := output.String()

	// Should not show Examples section if no examples exist
	if strings.Contains(result, "Examples:\n\n") {
		t.Errorf("did not expect empty 'Examples:' section in output")
	}
}

func TestShowCustomHelp_DisabledOptions(t *testing.T) {
	output := &bytes.Buffer{}
	targetCmd := &cobra.Command{
		Use:     "test",
		Short:   "Test command",
		Example: "test --flag value",
		Aliases: []string{"t"},
	}

	opts := &HelpOptions{
		ShowExamples: false,
		ShowAliases:  false,
		Output:       output,
	}

	err := showCustomHelp(targetCmd, opts)
	if err != nil {
		t.Fatalf("showCustomHelp failed: %v", err)
	}

	result := output.String()

	// Should not show examples or aliases when disabled
	if strings.Contains(result, "Examples:\n") {
		t.Errorf("expected no 'Examples:' section when disabled, got: %s", result)
	}

	if strings.Contains(result, "Aliases:\n") {
		t.Errorf("expected no 'Aliases:' section when disabled, got: %s", result)
	}
}

func TestGetCommandNames(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(&cobra.Command{Use: "foo", Short: "Foo command"})
	rootCmd.AddCommand(&cobra.Command{Use: "bar", Short: "Bar command"})
	rootCmd.AddCommand(&cobra.Command{Use: "baz", Short: "Baz command", Hidden: true})

	names := getCommandNames(rootCmd, "")

	if len(names) != 2 {
		t.Errorf("expected 2 command names (excluding hidden), got %d", len(names))
	}

	// Verify hidden command is not included
	for _, name := range names {
		if name == "baz" {
			t.Error("hidden command should not be included")
		}
	}
}

func TestGetCommandNames_WithPrefix(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}
	rootCmd.AddCommand(&cobra.Command{Use: "foo", Short: "Foo command"})
	rootCmd.AddCommand(&cobra.Command{Use: "bar", Short: "Bar command"})
	rootCmd.AddCommand(&cobra.Command{Use: "far", Short: "Far command"})

	names := getCommandNames(rootCmd, "f")

	if len(names) != 2 {
		t.Errorf("expected 2 command names starting with 'f', got %d", len(names))
	}

	for _, name := range names {
		if !strings.HasPrefix(name, "f") {
			t.Errorf("expected command name to start with 'f', got %q", name)
		}
	}
}

func TestCustomizeHelpTemplate(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	CustomizeHelpTemplate(cmd, "Test CLI")

	template := cmd.HelpTemplate()

	if !strings.Contains(template, "Test CLI") {
		t.Error("expected branding in help template")
	}

	if !strings.Contains(template, "Usage:") {
		t.Error("expected 'Usage:' in help template")
	}

	if !strings.Contains(template, "Available Commands:") {
		t.Error("expected 'Available Commands:' in help template")
	}
}

func TestCustomizeHelpTemplate_NoBranding(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	CustomizeHelpTemplate(cmd, "")

	template := cmd.HelpTemplate()

	// Should still have standard sections even without branding
	if !strings.Contains(template, "Usage:") {
		t.Error("expected 'Usage:' in help template")
	}
}

func TestCustomizeUsageTemplate(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	CustomizeUsageTemplate(cmd)

	template := cmd.UsageTemplate()

	if !strings.Contains(template, "Usage:") {
		t.Error("expected 'Usage:' in usage template")
	}

	if !strings.Contains(template, "Available Commands:") {
		t.Error("expected 'Available Commands:' in usage template")
	}

	if !strings.Contains(template, "Flags:") {
		t.Error("expected 'Flags:' in usage template")
	}
}

func TestHelpCommand_Integration(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	subCmd := &cobra.Command{
		Use:   "subcommand",
		Short: "A test subcommand",
	}
	rootCmd.AddCommand(subCmd)

	opts := &HelpOptions{
		ShowExamples: true,
		ShowAliases:  true,
		Output:       output,
	}

	helpCmd := NewHelpCommand(opts, rootCmd)
	rootCmd.AddCommand(helpCmd)

	// Test help with no args (should show root help)
	helpCmd.SetArgs([]string{})
	err := helpCmd.Execute()
	if err != nil {
		t.Fatalf("help command failed: %v", err)
	}

	// Output should contain something about testcli
	result := output.String()
	if len(result) == 0 {
		t.Error("expected help output, got empty string")
	}
}

func TestHelpCommand_UnknownCommand(t *testing.T) {
	output := &bytes.Buffer{}
	rootCmd := &cobra.Command{Use: "testcli"}

	opts := &HelpOptions{
		ShowExamples: true,
		ShowAliases:  true,
		Output:       output,
	}

	helpCmd := NewHelpCommand(opts, rootCmd)

	// Manually call RunE with unknown command
	err := helpCmd.RunE(helpCmd, []string{"nonexistent"})
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}

	if !strings.Contains(err.Error(), "unknown command") {
		t.Errorf("expected 'unknown command' error, got: %v", err)
	}
}
