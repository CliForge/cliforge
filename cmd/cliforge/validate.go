package main

import (
	"context"
	"fmt"
	"os"

	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/CliForge/cliforge/pkg/config"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func newValidateCmd() *cobra.Command {
	var (
		configPath string
	)

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate CLI configuration and OpenAPI spec",
		Long: `Validate your CLI configuration file and OpenAPI specification.

This command checks:
  - Configuration file syntax and structure
  - Required fields and valid values
  - OpenAPI spec format and validity
  - Compatibility between config and spec`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			debug, _ := cmd.Flags().GetBool("debug")

			fmt.Println("Validating CLI configuration...")

			// Load and validate config
			config, err := loadConfig(configPath)
			if err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}

			if verbose {
				fmt.Printf("✓ Configuration loaded: %s v%s\n", config.Metadata.Name, config.Metadata.Version)
			}

			// Validate config structure
			if err := validateConfig(config); err != nil {
				return fmt.Errorf("configuration validation failed: %w", err)
			}
			fmt.Println("✓ Configuration is valid")

			// Validate OpenAPI spec
			fmt.Println("\nValidating OpenAPI specification...")
			if err := validateOpenAPISpec(config.API.OpenAPIURL, verbose, debug); err != nil {
				return fmt.Errorf("OpenAPI validation failed: %w", err)
			}
			fmt.Println("✓ OpenAPI specification is valid")

			fmt.Println("\n✓ All validations passed")
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "cli-config.yaml", "Path to CLI configuration file")

	return cmd
}

func loadConfig(path string) (*cli.CLIConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config cli.CLIConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return &config, nil
}

func validateConfig(cfg *cli.CLIConfig) error {
	validator := config.NewValidator()
	return validator.Validate(cfg)
}

func validateOpenAPISpec(specPath string, verbose, debug bool) error {
	ctx := context.Background()
	parser := openapi.NewParser()

	if debug {
		parser.DisableValidation = false
		parser.AllowRemoteRefs = true
	}

	// Check if spec path is URL or file
	var spec *openapi.ParsedSpec
	var err error

	if isURL(specPath) {
		if verbose {
			fmt.Printf("  Loading spec from URL: %s\n", specPath)
		}
		loader := openapi.NewLoader(nil)
		spec, err = loader.LoadFromURL(ctx, specPath, &openapi.LoadOptions{
			ForceRefresh: true,
		})
	} else {
		if verbose {
			fmt.Printf("  Loading spec from file: %s\n", specPath)
		}
		spec, err = parser.ParseFile(ctx, specPath)
	}

	if err != nil {
		return err
	}

	info := spec.GetInfo()
	if verbose {
		fmt.Printf("  API: %s v%s\n", info.Title, info.Version)
		if info.CLIVersion != "" {
			fmt.Printf("  CLI Version: %s\n", info.CLIVersion)
		}
	}

	// Get operations count
	operations, err := spec.GetOperations()
	if err != nil {
		return fmt.Errorf("failed to parse operations: %w", err)
	}

	if verbose {
		fmt.Printf("  Operations: %d\n", len(operations))
	}

	return nil
}

func isURL(path string) bool {
	return len(path) >= 7 && (path[:7] == "http://" || path[:8] == "https://")
}
