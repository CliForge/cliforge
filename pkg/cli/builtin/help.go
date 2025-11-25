package builtin

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

// HelpOptions configures the help command behavior.
type HelpOptions struct {
	ShowExamples bool
	ShowAliases  bool
	Output       io.Writer
}

// NewHelpCommand creates a new help command.
// Note: Cobra provides built-in help functionality, but this allows customization.
func NewHelpCommand(opts *HelpOptions, rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type ` + rootCmd.Name() + ` help [path to command] for full details.`,
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return getCommandNames(rootCmd, toComplete), cobra.ShellCompDirectiveNoFileComp
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no args, show root help
			if len(args) == 0 {
				return rootCmd.Help()
			}

			// Find the command
			targetCmd, _, err := rootCmd.Find(args)
			if err != nil {
				return fmt.Errorf("unknown command %q: %w", strings.Join(args, " "), err)
			}

			// Show help for the target command
			return showCustomHelp(targetCmd, opts)
		},
	}

	return cmd
}

// showCustomHelp displays custom help for a command.
func showCustomHelp(cmd *cobra.Command, opts *HelpOptions) error {
	// Use the command's default help
	if err := cmd.Help(); err != nil {
		return err
	}

	// Add examples if enabled
	if opts.ShowExamples && cmd.Example != "" {
		fmt.Fprintln(opts.Output, "\nExamples:")
		fmt.Fprintln(opts.Output, cmd.Example)
	}

	// Add aliases if enabled
	if opts.ShowAliases && len(cmd.Aliases) > 0 {
		fmt.Fprintf(opts.Output, "\nAliases: %s\n", strings.Join(cmd.Aliases, ", "))
	}

	return nil
}

// getCommandNames returns a list of all command names for completion.
func getCommandNames(cmd *cobra.Command, prefix string) []string {
	var names []string

	for _, subCmd := range cmd.Commands() {
		if !subCmd.Hidden && strings.HasPrefix(subCmd.Name(), prefix) {
			names = append(names, subCmd.Name())
		}
	}

	return names
}

// CustomizeHelpTemplate customizes the help template for a command.
func CustomizeHelpTemplate(cmd *cobra.Command, branding string) {
	helpTemplate := `{{if .Long}}{{.Long | trimTrailingWhitespaces}}{{else}}{{.Short | trimTrailingWhitespaces}}{{end}}

{{if .Runnable}}Usage:
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

	if branding != "" {
		helpTemplate = branding + "\n\n" + helpTemplate
	}

	cmd.SetHelpTemplate(helpTemplate)
}

// CustomizeUsageTemplate customizes the usage template for a command.
func CustomizeUsageTemplate(cmd *cobra.Command) {
	usageTemplate := `Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`

	cmd.SetUsageTemplate(usageTemplate)
}
