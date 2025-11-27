package executor

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/CliForge/cliforge/internal/builder"
	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/auth/storage"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/output"
	"github.com/CliForge/cliforge/pkg/plugin"
	"github.com/CliForge/cliforge/pkg/progress"
	"github.com/CliForge/cliforge/pkg/state"
	"github.com/spf13/cobra"
)

// Runtime wires together all subsystems and manages the CLI lifecycle.
type Runtime struct {
	// Core components
	spec    *openapi.ParsedSpec
	rootCmd *cobra.Command

	// Managers
	authManager     *auth.Manager
	stateManager    *state.Manager
	outputManager   *output.Manager
	progressManager *progress.Manager
	pluginRegistry  *plugin.Registry

	// Builders
	commandBuilder *builder.Builder
	flagBuilder    *builder.FlagBuilder

	// Executor
	executor *Executor

	// HTTP client
	httpClient *http.Client
}

// RuntimeConfig configures the runtime.
type RuntimeConfig struct {
	CLIName    string
	SpecPath   string
	ConfigPath string
	BaseURL    string
}

// NewRuntime creates a new runtime instance.
func NewRuntime(ctx context.Context, runtimeConfig *RuntimeConfig) (*Runtime, error) {
	rt := &Runtime{}

	// Parse OpenAPI spec
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(ctx, runtimeConfig.SpecPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OpenAPI spec: %w", err)
	}
	rt.spec = spec

	// Override CLI name from spec extensions
	if spec.Extensions.Config != nil {
		if spec.Extensions.Config.Name != "" {
			runtimeConfig.CLIName = spec.Extensions.Config.Name
		}
	}

	// Initialize managers
	if err := rt.initializeManagers(runtimeConfig.CLIName); err != nil {
		return nil, fmt.Errorf("failed to initialize managers: %w", err)
	}

	// Create HTTP client with authentication
	if err := rt.createHTTPClient(); err != nil {
		return nil, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Build command tree
	if err := rt.buildCommandTree(runtimeConfig); err != nil {
		return nil, fmt.Errorf("failed to build command tree: %w", err)
	}

	return rt, nil
}

// initializeManagers initializes all manager components.
func (rt *Runtime) initializeManagers(cliName string) error {
	// State manager
	var err error
	rt.stateManager, err = state.NewManager(cliName)
	if err != nil {
		return fmt.Errorf("failed to create state manager: %w", err)
	}

	// Output manager
	rt.outputManager = output.NewManager()
	if rt.spec.Extensions.Config != nil && rt.spec.Extensions.Config.Output != nil {
		rt.outputManager.SetDefaultFormat(rt.spec.Extensions.Config.Output.DefaultFormat)
	}

	// Progress manager
	progressConfig := progress.DefaultConfig()
	rt.progressManager = progress.NewManager(progressConfig)

	// Auth manager
	rt.authManager = auth.NewManager(cliName)
	if err := rt.initializeAuth(); err != nil {
		return fmt.Errorf("failed to initialize auth: %w", err)
	}

	// Plugin registry - create with basic permissions
	permManager, _ := plugin.NewPermissionManager("", nil)
	rt.pluginRegistry = plugin.NewRegistry(cliName, permManager)
	if err := rt.initializePlugins(); err != nil {
		return fmt.Errorf("failed to initialize plugins: %w", err)
	}

	return nil
}

// initializeAuth initializes authentication from spec and config.
func (rt *Runtime) initializeAuth() error {
	// Check for auth configuration in spec extensions
	if rt.spec.Extensions.Config != nil && rt.spec.Extensions.Config.Auth != nil {
		authSettings := rt.spec.Extensions.Config.Auth

		// Create auth config based on type
		authConfig := &auth.Config{
			Type: auth.AuthType(authSettings.Type),
		}

		// Configure storage
		if authSettings.Storage != "" {
			authConfig.Storage = &auth.StorageConfig{
				Type: auth.StorageType(authSettings.Storage),
			}
		}

		// Register authenticator
		authConfigs := map[string]*auth.Config{
			"default": authConfig,
		}

		if err := rt.authManager.CreateFromConfig(authConfigs); err != nil {
			return err
		}
	} else {
		// Register no-auth as default
		if err := rt.authManager.RegisterAuthenticator("default", &auth.NoneAuth{}); err != nil {
			return err
		}
	}

	// Register memory storage as fallback
	rt.authManager.RegisterStorage("memory", storage.NewMemoryStorage())

	return nil
}

// initializePlugins initializes built-in plugins.
func (rt *Runtime) initializePlugins() error {
	// Register built-in plugins would go here
	// For now, just initialize the registry
	return nil
}

// createHTTPClient creates an authenticated HTTP client.
func (rt *Runtime) createHTTPClient() error {
	// Create base client with timeout
	baseClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	// For now, use base client
	// TODO: Wrap with authenticated client when needed
	rt.httpClient = baseClient

	return nil
}

// buildCommandTree builds the Cobra command tree.
func (rt *Runtime) buildCommandTree(runtimeConfig *RuntimeConfig) error {
	// Create builder config
	builderConfig := &builder.BuilderConfig{
		RootName:                runtimeConfig.CLIName,
		GroupByTags:             true,
		FlattenSingleOperations: true,
	}

	// Override from spec config
	if rt.spec.Extensions.Config != nil {
		if rt.spec.Extensions.Config.Name != "" {
			builderConfig.RootName = rt.spec.Extensions.Config.Name
		}
		if rt.spec.Extensions.Config.Description != "" {
			builderConfig.RootDescription = rt.spec.Extensions.Config.Description
		}
	}

	// Create command builder
	rt.commandBuilder = builder.NewBuilder(rt.spec, builderConfig)

	// Create executor
	execConfig := &ExecutorConfig{
		BaseURL:       runtimeConfig.BaseURL,
		HTTPClient:    rt.httpClient,
		AuthManager:   rt.authManager,
		OutputManager: rt.outputManager,
		StateManager:  rt.stateManager,
		ProgressMgr:   rt.progressManager,
	}

	var err error
	rt.executor, err = NewExecutor(rt.spec, execConfig)
	if err != nil {
		return err
	}

	// Set executor as default
	builderConfig.DefaultExecutor = rt.executor.Execute

	// Build command tree
	rootCmd, err := rt.commandBuilder.Build()
	if err != nil {
		return err
	}
	rt.rootCmd = rootCmd

	// Create flag builder
	rt.flagBuilder = builder.NewFlagBuilder(rt.spec.Extensions.Config)

	// Add global flags
	rt.flagBuilder.AddGlobalFlags(rootCmd)

	// Add operation-specific flags to all operation commands
	if err := rt.addOperationFlags(rootCmd); err != nil {
		return err
	}

	return nil
}

// addOperationFlags adds operation-specific flags to commands.
func (rt *Runtime) addOperationFlags(cmd *cobra.Command) error {
	// Check if this command has an operation
	if cmd.Annotations != nil {
		if operationID, ok := cmd.Annotations["operationID"]; ok {
			// Find the operation
			operations, err := rt.spec.GetOperations()
			if err != nil {
				return err
			}

			for _, op := range operations {
				if op.OperationID == operationID {
					// Add flags for this operation
					if err := rt.flagBuilder.AddOperationFlags(cmd, op); err != nil {
						return fmt.Errorf("failed to add flags for operation %s: %w", operationID, err)
					}
					break
				}
			}
		}
	}

	// Recursively process subcommands
	for _, subCmd := range cmd.Commands() {
		if err := rt.addOperationFlags(subCmd); err != nil {
			return err
		}
	}

	return nil
}

// GetRootCommand returns the root Cobra command.
func (rt *Runtime) GetRootCommand() *cobra.Command {
	return rt.rootCmd
}

// Execute executes the CLI.
func (rt *Runtime) Execute() error {
	return rt.rootCmd.Execute()
}

// ExecuteContext executes the CLI with a context.
func (rt *Runtime) ExecuteContext(ctx context.Context) error {
	return rt.rootCmd.ExecuteContext(ctx)
}

// GetAuthManager returns the auth manager.
func (rt *Runtime) GetAuthManager() *auth.Manager {
	return rt.authManager
}

// GetStateManager returns the state manager.
func (rt *Runtime) GetStateManager() *state.Manager {
	return rt.stateManager
}

// GetOutputManager returns the output manager.
func (rt *Runtime) GetOutputManager() *output.Manager {
	return rt.outputManager
}

// GetProgressManager returns the progress manager.
func (rt *Runtime) GetProgressManager() *progress.Manager {
	return rt.progressManager
}

// GetPluginRegistry returns the plugin registry.
func (rt *Runtime) GetPluginRegistry() *plugin.Registry {
	return rt.pluginRegistry
}

// Shutdown performs cleanup operations.
func (rt *Runtime) Shutdown() error {
	// Save state
	if rt.stateManager != nil {
		if err := rt.stateManager.Save(); err != nil {
			return fmt.Errorf("failed to save state: %w", err)
		}
	}

	return nil
}
