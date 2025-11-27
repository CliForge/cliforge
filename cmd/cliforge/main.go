// Package main implements the CliForge generator CLI.
//
// CliForge is a powerful tool that generates production-ready, branded CLIs
// from OpenAPI specifications. It provides a complete CLI generation workflow
// including initialization, validation, and binary building.
//
// # Commands
//
//   - init: Initialize a new CLI project with configuration templates
//   - build: Generate and compile a CLI binary from config and spec
//   - validate: Validate OpenAPI spec and CLI configuration
//
// # Example Usage
//
//	# Initialize a new CLI project
//	cliforge init --name mycli --spec api.yaml
//
//	# Validate configuration and spec
//	cliforge validate
//
//	# Build the CLI binary
//	cliforge build --output ./bin/mycli
//
// # Configuration
//
// CliForge uses a cli-config.yaml file that defines CLI metadata, branding,
// authentication, behaviors, and features. The configuration is embedded
// into the generated binary and controls all runtime behavior.
//
// For more information, visit: https://github.com/CliForge/cliforge
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Version is set at build time
	version = "0.9.0"
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
		Version:      version,
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
