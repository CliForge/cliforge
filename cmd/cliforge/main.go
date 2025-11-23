// Package main implements the CliForge generator CLI.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	version = "0.9.0"
	// BuildDate is set at build time
	buildDate = "unknown"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cliforge",
		Short: "CliForge - Generate branded CLIs from OpenAPI specs",
		Long: `CliForge is a powerful CLI generator that creates branded,
production-ready command-line tools from OpenAPI specifications.

It supports authentication, caching, updates, and many other
enterprise features out of the box.`,
		Version: version,
		SilenceUsage: true,
	}

	// Add global flags
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
	cmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug mode")

	// Add subcommands
	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newBuildCmd())
	cmd.AddCommand(newValidateCmd())

	return cmd
}
