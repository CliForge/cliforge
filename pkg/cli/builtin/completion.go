package builtin

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// CompletionOptions configures the completion command behavior.
type CompletionOptions struct {
	CLIName       string
	EnabledShells []string
	Output        io.Writer
}

// NewCompletionCommand creates a new completion command.
func NewCompletionCommand(opts *CompletionOptions, rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: fmt.Sprintf(`Generate shell completion scripts for %s.

The completion command generates shell-specific completion scripts that enable
tab-completion for commands, flags, and arguments.

Installation:

Bash:
  $ %s completion bash > /etc/bash_completion.d/%s
  Or for the current user:
  $ %s completion bash > ~/.local/share/bash-completion/completions/%s

Zsh:
  $ %s completion zsh > ~/.zsh/completion/_%s
  Then add the following to ~/.zshrc:
  fpath=(~/.zsh/completion $fpath)
  autoload -Uz compinit && compinit

Fish:
  $ %s completion fish > ~/.config/fish/completions/%s.fish

PowerShell:
  $ %s completion powershell > %s.ps1
  Then add to your PowerShell profile`, opts.CLIName,
			opts.CLIName, opts.CLIName,
			opts.CLIName, opts.CLIName,
			opts.CLIName, opts.CLIName,
			opts.CLIName, opts.CLIName,
			opts.CLIName, opts.CLIName),
		ValidArgs:             opts.EnabledShells,
		Args:                  cobra.ExactArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCompletion(rootCmd, args[0], opts)
		},
	}

	return cmd
}

// runCompletion generates the completion script for the specified shell.
func runCompletion(rootCmd *cobra.Command, shell string, opts *CompletionOptions) error {
	// Check if shell is enabled
	enabled := false
	for _, s := range opts.EnabledShells {
		if s == shell {
			enabled = true
			break
		}
	}

	if !enabled {
		return fmt.Errorf("completion for %s is not enabled", shell)
	}

	// Generate completion
	switch shell {
	case "bash":
		return rootCmd.GenBashCompletion(opts.Output)
	case "zsh":
		return rootCmd.GenZshCompletion(opts.Output)
	case "fish":
		return rootCmd.GenFishCompletion(opts.Output, true)
	case "powershell":
		return rootCmd.GenPowerShellCompletionWithDesc(opts.Output)
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

// SetupCompletionFunctions configures dynamic completion for a command.
func SetupCompletionFunctions(cmd *cobra.Command) {
	// Add completion for common flags
	_ = cmd.RegisterFlagCompletionFunc("output", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml", "table", "csv", "text"}, cobra.ShellCompDirectiveNoFileComp
	})

	_ = cmd.RegisterFlagCompletionFunc("format", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"json", "yaml", "table", "csv", "text"}, cobra.ShellCompDirectiveNoFileComp
	})
}

// CompletionFunc is a helper type for dynamic completion functions.
type CompletionFunc func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective)

// NoFileCompletion returns a completion function that disables file completion.
func NoFileCompletion() CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

// FixedCompletion returns a completion function with fixed values.
func FixedCompletion(values ...string) CompletionFunc {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return values, cobra.ShellCompDirectiveNoFileComp
	}
}
