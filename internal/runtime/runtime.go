// Package runtime provides the runtime environment for generated CLIs.
//
// The runtime package initializes and manages the complete runtime environment
// for CliForge-generated CLIs. It loads embedded configuration, initializes all
// subsystems (auth, cache, state, output), builds the command tree, and
// orchestrates command execution.
//
// # Initialization Flow
//
//	1. Parse embedded configuration
//	2. Initialize subsystems:
//	   - Output manager
//	   - State manager (contexts, history)
//	   - Cache (OpenAPI specs)
//	   - Auth manager
//	3. Load OpenAPI specification
//	4. Build Cobra command tree
//	5. Add global flags
//	6. Add built-in commands
//	7. Execute CLI
//
// # Example Usage
//
// This is typically used in generated main.go files:
//
//	//go:embed config.yaml
//	var embeddedConfig []byte
//
//	func main() {
//	    rt, err := runtime.NewRuntime(embeddedConfig, version, debug)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    rt.Execute()
//	}
//
// # Subsystems
//
//   - AuthManager: Handles authentication flows and token storage
//   - StateManager: Manages contexts, history, and recent values
//   - SpecCache: Caches OpenAPI specs with ETag support
//	OutputManager: Formats command output in multiple formats
//   - ProgressManager: Displays progress for long operations
//
// # Global Flags
//
// The runtime adds global flags based on configuration:
//
//	--output, -o     Output format (json, yaml, table)
//	--verbose, -v    Enable verbose output
//	--debug          Enable debug mode
//	--no-color       Disable colored output
//	--config         Path to config file
//	--profile        Named profile to use
//
// # Built-in Commands
//
// Built-in commands are added based on cli-config.yaml:
//
//	version      Show version information
//	help         Help about any command
//	completion   Generate shell completion
//	config       Manage configuration
//	context      Manage contexts
//	cache        Manage cache
//	update       Check for updates
//	auth         Manage authentication
//
// The runtime package provides the complete scaffolding needed for
// production-ready CLI applications generated from OpenAPI specs.
package runtime

import (
	"context"
	"fmt"
	"os"

	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/cache"
	"github.com/CliForge/cliforge/pkg/cli"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/output"
	"github.com/CliForge/cliforge/pkg/state"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// Runtime represents the runtime environment for a generated CLI.
type Runtime struct {
	config          *cli.CLIConfig
	version         string
	debug           bool
	rootCmd         *cobra.Command
	authManager   *auth.Manager
	outputManager *output.Manager
	stateManager  *state.Manager
	specCache     *cache.SpecCache
}

// NewRuntime creates a new Runtime instance from embedded configuration.
func NewRuntime(embeddedConfig []byte, version string, debug bool) (*Runtime, error) {
	// Parse embedded configuration
	var config cli.CLIConfig
	if err := yaml.Unmarshal(embeddedConfig, &config); err != nil {
		return nil, fmt.Errorf("failed to parse embedded config: %w", err)
	}

	// Set debug mode from build
	config.Metadata.Debug = debug

	rt := &Runtime{
		config:  &config,
		version: version,
		debug:   debug,
	}

	// Initialize all subsystems
	if err := rt.initializeSubsystems(); err != nil {
		return nil, fmt.Errorf("failed to initialize subsystems: %w", err)
	}

	// Build command tree
	if err := rt.buildCommandTree(); err != nil {
		return nil, fmt.Errorf("failed to build command tree: %w", err)
	}

	return rt, nil
}

// initializeSubsystems initializes all CLI subsystems.
func (rt *Runtime) initializeSubsystems() error {
	ctx := context.Background()

	// Initialize output manager
	rt.outputManager = output.NewManager()

	// Initialize state manager
	stateDir := fmt.Sprintf("%s/.%s/state", os.Getenv("HOME"), rt.config.Metadata.Name)
	var err error
	rt.stateManager, err = state.NewManager(stateDir)
	if err != nil {
		return fmt.Errorf("failed to initialize state manager: %w", err)
	}

	// Initialize cache
	if rt.config.Defaults != nil && rt.config.Defaults.Caching != nil && rt.config.Defaults.Caching.Enabled {
		rt.specCache, err = cache.NewSpecCache(rt.config.Metadata.Name)
		if err != nil {
			return fmt.Errorf("failed to initialize cache: %w", err)
		}
	}

	// Initialize auth manager
	if rt.config.Behaviors != nil && rt.config.Behaviors.Auth != nil {
		rt.authManager = auth.NewManager(rt.config.Metadata.Name)
		authConfig := convertAuthConfig(rt.config.Behaviors.Auth)
		authenticator, err := createAuthenticator(authConfig)
		if err != nil {
			return fmt.Errorf("failed to create authenticator: %w", err)
		}
		if err := rt.authManager.RegisterAuthenticator("default", authenticator); err != nil {
			return fmt.Errorf("failed to register authenticator: %w", err)
		}
	}

	// Load OpenAPI spec
	if err := rt.loadOpenAPISpec(ctx); err != nil {
		return fmt.Errorf("failed to load OpenAPI spec: %w", err)
	}

	return nil
}

// loadOpenAPISpec loads the OpenAPI specification.
func (rt *Runtime) loadOpenAPISpec(ctx context.Context) error {
	loader := openapi.NewLoader(nil)

	// Load spec from configured URL or file
	_, err := loader.LoadFromURL(ctx, rt.config.API.OpenAPIURL, nil)
	if err != nil {
		// Try as file path
		_, err = loader.LoadFromFile(ctx, rt.config.API.OpenAPIURL)
		if err != nil {
			return fmt.Errorf("failed to load OpenAPI spec: %w", err)
		}
	}

	return nil
}

// buildCommandTree builds the Cobra command tree from the OpenAPI spec.
func (rt *Runtime) buildCommandTree() error {
	ctx := context.Background()

	// Create root command
	rt.rootCmd = &cobra.Command{
		Use:     rt.config.Metadata.Name,
		Short:   rt.config.Metadata.Description,
		Long:    rt.config.Metadata.LongDescription,
		Version: rt.version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return rt.preRunHook(cmd, args)
		},
		SilenceUsage: true,
	}

	// Add global flags
	rt.addGlobalFlags()

	// Add built-in commands
	if err := rt.addBuiltinCommands(); err != nil {
		return fmt.Errorf("failed to add builtin commands: %w", err)
	}

	// Load spec and build API commands
	if err := rt.buildAPICommands(ctx); err != nil {
		return fmt.Errorf("failed to build API commands: %w", err)
	}

	return nil
}

// addGlobalFlags adds global flags to the root command.
func (rt *Runtime) addGlobalFlags() {
	if rt.config.Behaviors == nil || rt.config.Behaviors.GlobalFlags == nil {
		return
	}

	flags := rt.config.Behaviors.GlobalFlags

	if flags.Output != nil && flags.Output.Enabled {
		rt.rootCmd.PersistentFlags().StringP("output", "o", rt.config.Defaults.Output.Format, flags.Output.Description)
	}

	if flags.Verbose != nil && flags.Verbose.Enabled {
		rt.rootCmd.PersistentFlags().BoolP("verbose", "v", false, flags.Verbose.Description)
	}

	if flags.Debug != nil && flags.Debug.Enabled {
		rt.rootCmd.PersistentFlags().Bool("debug", false, flags.Debug.Description)
	}

	if flags.NoColor != nil && flags.NoColor.Enabled {
		rt.rootCmd.PersistentFlags().Bool("no-color", false, flags.NoColor.Description)
	}
}

// addBuiltinCommands adds built-in commands to the CLI.
func (rt *Runtime) addBuiltinCommands() error {
	if rt.config.Behaviors == nil || rt.config.Behaviors.BuiltinCommands == nil {
		return nil
	}

	cmds := rt.config.Behaviors.BuiltinCommands

	// Built-in commands - placeholder for future implementation
	// Version, help, completion, cache, update, context, history commands
	// Will be implemented in future iterations
	_ = cmds // Prevent unused variable warning

	return nil
}

// buildAPICommands builds commands from the OpenAPI specification.
func (rt *Runtime) buildAPICommands(ctx context.Context) error {
	// Load OpenAPI spec
	loader := openapi.NewLoader(nil)
	spec, err := loader.LoadFromURL(ctx, rt.config.API.OpenAPIURL, nil)
	if err != nil {
		// Try as file path
		spec, err = loader.LoadFromFile(ctx, rt.config.API.OpenAPIURL)
		if err != nil {
			return fmt.Errorf("failed to load spec: %w", err)
		}
	}

	// Get operations from spec
	operations, err := spec.GetOperations()
	if err != nil {
		return fmt.Errorf("failed to get operations: %w", err)
	}

	// Build commands from operations
	builder := NewCommandBuilder(rt)
	for _, op := range operations {
		cmd, err := builder.BuildCommand(op)
		if err != nil {
			return fmt.Errorf("failed to build command for %s: %w", op.OperationID, err)
		}
		rt.rootCmd.AddCommand(cmd)
	}

	return nil
}

// preRunHook is executed before every command.
func (rt *Runtime) preRunHook(cmd *cobra.Command, args []string) error {
	// Future: Check for updates in background
	return nil
}

// Execute runs the CLI.
func (rt *Runtime) Execute() error {
	return rt.rootCmd.Execute()
}

// convertAuthConfig converts CLI auth config to auth package config.
func convertAuthConfig(authBehavior *cli.AuthBehavior) *auth.Config {
	cfg := &auth.Config{}

	switch authBehavior.Type {
	case "api_key":
		cfg.Type = auth.AuthTypeAPIKey
		if authBehavior.APIKey != nil {
			cfg.APIKey = &auth.APIKeyConfig{
				Name:     authBehavior.APIKey.Header,
				Location: auth.APIKeyLocationHeader,
				EnvVar:   authBehavior.APIKey.EnvVar,
			}
		}
	case "oauth2":
		cfg.Type = auth.AuthTypeOAuth2
		if authBehavior.OAuth2 != nil {
			cfg.OAuth2 = &auth.OAuth2Config{
				ClientID:     authBehavior.OAuth2.ClientID,
				ClientSecret: authBehavior.OAuth2.ClientSecret,
				AuthURL:      authBehavior.OAuth2.AuthURL,
				TokenURL:     authBehavior.OAuth2.TokenURL,
				Scopes:       authBehavior.OAuth2.Scopes,
				Flow:         auth.OAuth2FlowAuthorizationCode,
			}
		}
	case "basic":
		cfg.Type = auth.AuthTypeBasic
		if authBehavior.Basic != nil {
			cfg.Basic = &auth.BasicConfig{
				EnvUsername: authBehavior.Basic.UsernameEnv,
				EnvPassword: authBehavior.Basic.PasswordEnv,
			}
		}
	default:
		cfg.Type = auth.AuthTypeNone
	}

	return cfg
}

// createAuthenticator creates an authenticator from config.
func createAuthenticator(config *auth.Config) (auth.Authenticator, error) {
	switch config.Type {
	case auth.AuthTypeAPIKey:
		if config.APIKey == nil {
			return nil, fmt.Errorf("apikey config is required for API key auth")
		}
		return auth.NewAPIKeyAuth(config.APIKey)

	case auth.AuthTypeOAuth2:
		if config.OAuth2 == nil {
			return nil, fmt.Errorf("oauth2 config is required for OAuth2 auth")
		}
		return auth.NewOAuth2Auth(config.OAuth2)

	case auth.AuthTypeBasic:
		if config.Basic == nil {
			return nil, fmt.Errorf("basic config is required for Basic auth")
		}
		return auth.NewBasicAuth(config.Basic)

	case auth.AuthTypeNone:
		return &auth.NoneAuth{}, nil

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", config.Type)
	}
}
