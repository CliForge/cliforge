package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"

	"github.com/CliForge/cliforge/internal/embed"
	"github.com/CliForge/cliforge/internal/runtime"
	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type buildTarget struct {
	OS   string
	Arch string
}

var defaultTargets = []buildTarget{
	{"linux", "amd64"},
	{"linux", "arm64"},
	{"darwin", "amd64"},
	{"darwin", "arm64"},
	{"windows", "amd64"},
}

func newBuildCmd() *cobra.Command {
	var (
		configPath  string
		outputDir   string
		platforms   []string
		allPlatforms bool
		skipChecksums bool
	)

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build CLI binary from configuration",
		Long: `Build a branded CLI binary from your configuration.

This command:
  1. Loads and validates your CLI configuration
  2. Fetches and validates the OpenAPI specification
  3. Embeds configuration into the binary
  4. Compiles the CLI for specified platforms
  5. Generates checksums for verification`,
		RunE: func(cmd *cobra.Command, args []string) error {
			verbose, _ := cmd.Flags().GetBool("verbose")
			debug, _ := cmd.Flags().GetBool("debug")

			ctx := context.Background()

			// Load configuration
			fmt.Println("Loading configuration...")
			config, err := loadConfig(configPath)
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			if verbose {
				fmt.Printf("✓ Loaded config for %s v%s\n", config.Metadata.Name, config.Metadata.Version)
			}

			// Validate configuration
			if err := validateConfig(config); err != nil {
				return fmt.Errorf("config validation failed: %w", err)
			}

			// Load OpenAPI spec
			fmt.Println("Loading OpenAPI specification...")
			spec, err := loadOpenAPISpec(ctx, config.API.OpenAPIURL, verbose)
			if err != nil {
				return fmt.Errorf("failed to load OpenAPI spec: %w", err)
			}

			if verbose {
				info := spec.GetInfo()
				fmt.Printf("✓ Loaded OpenAPI spec: %s v%s\n", info.Title, info.Version)
			}

			// Determine build targets
			targets, err := getBuildTargets(platforms, allPlatforms)
			if err != nil {
				return err
			}

			if verbose {
				fmt.Printf("Building for %d platform(s)\n", len(targets))
			}

			// Create output directory
			if outputDir == "" {
				outputDir = "dist"
			}
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Generate runtime code
			fmt.Println("Generating runtime code...")
			buildDir, err := generateRuntimeCode(config, debug, verbose)
			if err != nil {
				return fmt.Errorf("failed to generate runtime: %w", err)
			}
			defer os.RemoveAll(buildDir)

			if verbose {
				fmt.Printf("✓ Generated runtime in %s\n", buildDir)
			}

			// Build for each target
			checksumFile := filepath.Join(outputDir, "checksums.txt")
			var checksumData strings.Builder

			for _, target := range targets {
				fmt.Printf("Building for %s/%s...\n", target.OS, target.Arch)

				outputPath, err := buildBinary(buildDir, outputDir, config.Metadata.Name, target, config.Metadata.Version, debug)
				if err != nil {
					return fmt.Errorf("build failed for %s/%s: %w", target.OS, target.Arch, err)
				}

				fmt.Printf("✓ Built %s\n", filepath.Base(outputPath))

				// Generate checksum
				if !skipChecksums {
					checksum, err := generateChecksum(outputPath)
					if err != nil {
						return fmt.Errorf("failed to generate checksum: %w", err)
					}
					checksumData.WriteString(fmt.Sprintf("%s  %s\n", checksum, filepath.Base(outputPath)))
				}
			}

			// Write checksums file
			if !skipChecksums {
				if err := os.WriteFile(checksumFile, []byte(checksumData.String()), 0644); err != nil {
					return fmt.Errorf("failed to write checksums: %w", err)
				}
				fmt.Printf("✓ Generated %s\n", checksumFile)
			}

			fmt.Printf("\n✓ Build complete! Binaries are in %s\n", outputDir)
			return nil
		},
	}

	cmd.Flags().StringVarP(&configPath, "config", "c", "cli-config.yaml", "Path to CLI configuration file")
	cmd.Flags().StringVarP(&outputDir, "output", "o", "dist", "Output directory for binaries")
	cmd.Flags().StringSliceVarP(&platforms, "platform", "p", nil, "Target platforms (e.g., linux/amd64,darwin/arm64)")
	cmd.Flags().BoolVarP(&allPlatforms, "all", "a", false, "Build for all supported platforms")
	cmd.Flags().BoolVar(&skipChecksums, "skip-checksums", false, "Skip checksum generation")

	return cmd
}

func loadOpenAPISpec(ctx context.Context, specPath string, verbose bool) (*openapi.ParsedSpec, error) {
	parser := openapi.NewParser()

	if isURL(specPath) {
		loader := openapi.NewLoader(nil)
		return loader.LoadFromURL(ctx, specPath, &openapi.LoadOptions{
			ForceRefresh: true,
		})
	}

	return parser.ParseFile(ctx, specPath)
}

func getBuildTargets(platforms []string, allPlatforms bool) ([]buildTarget, error) {
	if allPlatforms {
		return defaultTargets, nil
	}

	if len(platforms) == 0 {
		// Build for current platform only
		return []buildTarget{{OS: goruntime.GOOS, Arch: goruntime.GOARCH}}, nil
	}

	var targets []buildTarget
	for _, platform := range platforms {
		parts := strings.Split(platform, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid platform format: %s (expected: os/arch)", platform)
		}
		targets = append(targets, buildTarget{OS: parts[0], Arch: parts[1]})
	}

	return targets, nil
}

func generateRuntimeCode(config *cli.CLIConfig, debug, verbose bool) (string, error) {
	// Create temporary build directory
	buildDir, err := os.MkdirTemp("", "cliforge-build-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}

	// Generate embedded config
	configData, err := yaml.Marshal(config)
	if err != nil {
		os.RemoveAll(buildDir)
		return "", fmt.Errorf("failed to marshal config: %w", err)
	}

	// Generate main.go using runtime template
	generator := runtime.NewGenerator(config)
	mainCode, err := generator.GenerateMain(configData)
	if err != nil {
		os.RemoveAll(buildDir)
		return "", fmt.Errorf("failed to generate main: %w", err)
	}

	// Write main.go
	mainPath := filepath.Join(buildDir, "main.go")
	if err := os.WriteFile(mainPath, []byte(mainCode), 0644); err != nil {
		os.RemoveAll(buildDir)
		return "", fmt.Errorf("failed to write main.go: %w", err)
	}

	// Generate embedded config file
	embedder := embed.NewEmbedder()
	embedPath := filepath.Join(buildDir, "config_embedded.yaml")
	if err := embedder.WriteEmbeddedConfig(configData, embedPath); err != nil {
		os.RemoveAll(buildDir)
		return "", fmt.Errorf("failed to write embedded config: %w", err)
	}

	// Initialize go.mod
	if err := initGoModule(buildDir, config.Metadata.Name, verbose); err != nil {
		os.RemoveAll(buildDir)
		return "", fmt.Errorf("failed to initialize go module: %w", err)
	}

	return buildDir, nil
}

func initGoModule(buildDir, moduleName string, verbose bool) error {
	// Create go.mod
	goModContent := fmt.Sprintf(`module %s

go 1.25.4

require github.com/CliForge/cliforge v0.9.0
`, moduleName)

	goModPath := filepath.Join(buildDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		return fmt.Errorf("failed to write go.mod: %w", err)
	}

	// Run go mod tidy
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = buildDir
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy failed: %w", err)
	}

	return nil
}

func buildBinary(buildDir, outputDir, name string, target buildTarget, version string, debug bool) (string, error) {
	// Determine binary name
	binaryName := name
	if target.OS == "windows" {
		binaryName += ".exe"
	}

	// Add platform to filename
	outputName := fmt.Sprintf("%s-%s-%s", name, target.OS, target.Arch)
	if target.OS == "windows" {
		outputName += ".exe"
	}

	outputPath := filepath.Join(outputDir, outputName)

	// Build command
	ldflags := fmt.Sprintf("-s -w -X main.version=%s", version)
	if !debug {
		ldflags += " -X main.debug=false"
	}

	args := []string{
		"build",
		"-ldflags", ldflags,
		"-o", outputPath,
		".",
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = buildDir
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", target.OS),
		fmt.Sprintf("GOARCH=%s", target.Arch),
		"CGO_ENABLED=0",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("build failed: %w\n%s", err, output)
	}

	return outputPath, nil
}

func generateChecksum(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
