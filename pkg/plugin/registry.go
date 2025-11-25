package plugin

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Registry manages plugin registration, discovery, and lifecycle.
type Registry struct {
	plugins           map[string]Plugin
	manifests         map[string]*PluginManifest
	mu                sync.RWMutex
	pluginDir         string
	permissionManager *PermissionManager
}

// NewRegistry creates a new plugin registry.
func NewRegistry(pluginDir string, permissionManager *PermissionManager) *Registry {
	return &Registry{
		plugins:           make(map[string]Plugin),
		manifests:         make(map[string]*PluginManifest),
		pluginDir:         pluginDir,
		permissionManager: permissionManager,
	}
}

// Register registers a plugin with the registry.
// Built-in plugins should be registered at startup.
func (r *Registry) Register(plugin Plugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	info := plugin.Describe()
	if info == nil {
		return fmt.Errorf("plugin returned nil info")
	}

	name := info.Manifest.Name
	if name == "" {
		return fmt.Errorf("plugin name cannot be empty")
	}

	// Validate the plugin
	if err := plugin.Validate(); err != nil {
		return fmt.Errorf("plugin validation failed: %w", err)
	}

	// Check if already registered
	if _, exists := r.plugins[name]; exists {
		return fmt.Errorf("plugin '%s' is already registered", name)
	}

	r.plugins[name] = plugin
	r.manifests[name] = &info.Manifest

	return nil
}

// Unregister removes a plugin from the registry.
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.plugins[name]; !exists {
		return fmt.Errorf("plugin '%s' not found", name)
	}

	delete(r.plugins, name)
	delete(r.manifests, name)

	return nil
}

// Get retrieves a plugin by name.
func (r *Registry) Get(name string) (Plugin, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	plugin, exists := r.plugins[name]
	if !exists {
		return nil, fmt.Errorf("plugin '%s' not found", name)
	}

	return plugin, nil
}

// List returns all registered plugin names.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.plugins))
	for name := range r.plugins {
		names = append(names, name)
	}
	return names
}

// GetManifest returns the manifest for a plugin.
func (r *Registry) GetManifest(name string) (*PluginManifest, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	manifest, exists := r.manifests[name]
	if !exists {
		return nil, fmt.Errorf("plugin manifest for '%s' not found", name)
	}

	return manifest, nil
}

// Execute executes a plugin with the given input.
// Checks permissions before execution.
func (r *Registry) Execute(ctx context.Context, name string, input *PluginInput) (*PluginOutput, error) {
	// Get the plugin
	plugin, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	// Get manifest for permission check
	manifest, err := r.GetManifest(name)
	if err != nil {
		return nil, err
	}

	// Check permissions
	if r.permissionManager != nil {
		if err := r.permissionManager.CheckPermissions(name, manifest.Permissions); err != nil {
			return nil, NewPluginError(name, "permission denied", err)
		}
	}

	// Execute the plugin
	output, err := plugin.Execute(ctx, input)
	if err != nil {
		return nil, NewPluginError(name, "execution failed", err)
	}

	return output, nil
}

// DiscoverPlugins discovers and loads external plugins from the plugin directory.
func (r *Registry) DiscoverPlugins() error {
	if r.pluginDir == "" {
		return nil
	}

	// Check if plugin directory exists
	if _, err := os.Stat(r.pluginDir); os.IsNotExist(err) {
		// Plugin directory doesn't exist yet, that's OK
		return nil
	}

	// Walk the plugin directory
	return filepath.Walk(r.pluginDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Look for manifest files
		if filepath.Base(path) == "plugin-manifest.yaml" {
			return r.loadExternalPlugin(filepath.Dir(path))
		}

		return nil
	})
}

// loadExternalPlugin loads an external plugin from a directory.
func (r *Registry) loadExternalPlugin(dir string) error {
	manifestPath := filepath.Join(dir, "plugin-manifest.yaml")

	// Read manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Validate manifest
	if err := r.validateManifest(&manifest); err != nil {
		return fmt.Errorf("invalid manifest: %w", err)
	}

	// Create plugin based on type
	var plugin Plugin
	switch manifest.Type {
	case PluginTypeBinary:
		// Resolve executable path
		execPath := manifest.Executable
		if !filepath.IsAbs(execPath) {
			execPath = filepath.Join(dir, execPath)
		}
		plugin = NewBinaryPlugin(manifest, execPath)

	case PluginTypeWASM:
		// WASM plugins are not yet implemented
		return fmt.Errorf("WASM plugins are not yet supported")

	default:
		return fmt.Errorf("unsupported plugin type: %s", manifest.Type)
	}

	// Register the plugin
	return r.Register(plugin)
}

// validateManifest validates a plugin manifest.
func (r *Registry) validateManifest(manifest *PluginManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("name is required")
	}

	if manifest.Version == "" {
		return fmt.Errorf("version is required")
	}

	if manifest.Type == "" {
		return fmt.Errorf("type is required")
	}

	// Validate permissions
	for _, perm := range manifest.Permissions {
		if err := ValidatePermission(perm.String()); err != nil {
			return fmt.Errorf("invalid permission: %w", err)
		}
	}

	// Type-specific validation
	switch manifest.Type {
	case PluginTypeBinary:
		if manifest.Executable == "" {
			return fmt.Errorf("executable is required for binary plugins")
		}

	case PluginTypeWASM:
		if manifest.Entrypoint == "" {
			return fmt.Errorf("entrypoint is required for WASM plugins")
		}

	case PluginTypeBuiltin:
		// Built-in plugins are registered programmatically

	default:
		return fmt.Errorf("unknown plugin type: %s", manifest.Type)
	}

	return nil
}

// GetPluginInfo returns detailed information about a plugin.
func (r *Registry) GetPluginInfo(name string) (*PluginInfo, error) {
	plugin, err := r.Get(name)
	if err != nil {
		return nil, err
	}

	info := plugin.Describe()
	if info == nil {
		return nil, fmt.Errorf("plugin returned nil info")
	}

	// Add permission approval status
	if r.permissionManager != nil {
		if approved, exists := r.permissionManager.GetApprovedPermissions(name); exists {
			if info.Status == PluginStatusReady {
				// Check if all required permissions are approved
				manifest := &info.Manifest
				missing := r.permissionManager.findMissingPermissions(manifest.Permissions, approved)
				if len(missing) > 0 {
					info.Status = PluginStatusPendingApproval
				}
			}
		} else if len(info.Manifest.Permissions) > 0 {
			info.Status = PluginStatusPendingApproval
		}
	}

	return info, nil
}

// ListPluginInfo returns information about all registered plugins.
func (r *Registry) ListPluginInfo() ([]*PluginInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]*PluginInfo, 0, len(r.plugins))
	for name := range r.plugins {
		// Temporarily unlock to call GetPluginInfo
		r.mu.RUnlock()
		info, err := r.GetPluginInfo(name)
		r.mu.RLock()

		if err != nil {
			// Skip plugins that fail to provide info
			continue
		}
		infos = append(infos, info)
	}

	return infos, nil
}

// BinaryPlugin wraps an external binary plugin.
type BinaryPlugin struct {
	manifest PluginManifest
	execPath string
}

// NewBinaryPlugin creates a new binary plugin.
func NewBinaryPlugin(manifest PluginManifest, execPath string) *BinaryPlugin {
	return &BinaryPlugin{
		manifest: manifest,
		execPath: execPath,
	}
}

// Execute runs the binary plugin.
func (p *BinaryPlugin) Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error) {
	// Binary plugin execution is handled by the executor
	// This is a placeholder that will be implemented in executor.go
	return nil, fmt.Errorf("binary plugin execution not yet implemented")
}

// Validate checks if the binary plugin is valid.
func (p *BinaryPlugin) Validate() error {
	// Check if executable exists
	if _, err := os.Stat(p.execPath); err != nil {
		return fmt.Errorf("executable not found: %s", p.execPath)
	}

	// Check if executable is executable
	info, err := os.Stat(p.execPath)
	if err != nil {
		return err
	}

	// Check execute permission
	if info.Mode()&0111 == 0 {
		return fmt.Errorf("file is not executable: %s", p.execPath)
	}

	return nil
}

// Describe returns plugin information.
func (p *BinaryPlugin) Describe() *PluginInfo {
	return &PluginInfo{
		Manifest:     p.manifest,
		Capabilities: []string{"execute"},
		Status:       PluginStatusReady,
	}
}
