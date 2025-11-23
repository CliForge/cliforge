package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// PermissionManager handles plugin permission approval and storage.
type PermissionManager struct {
	configDir string
	store     *PermissionStore
	mu        sync.RWMutex
	approver  PermissionApprover
}

// PermissionApprover is an interface for requesting user approval.
type PermissionApprover interface {
	// RequestApproval asks the user to approve permissions.
	RequestApproval(pluginName string, permissions []Permission) (bool, error)
}

// PermissionStore stores approved permissions.
type PermissionStore struct {
	// Plugins maps plugin names to their approved permissions.
	Plugins map[string]*ApprovedPlugin `yaml:"plugins"`
}

// ApprovedPlugin contains approved permissions for a plugin.
type ApprovedPlugin struct {
	// ApprovedPermissions is the list of approved permission strings.
	ApprovedPermissions []string `yaml:"approved_permissions"`

	// ApprovedAt is when the permissions were approved.
	ApprovedAt time.Time `yaml:"approved_at"`

	// LastUsed is when the plugin was last executed.
	LastUsed time.Time `yaml:"last_used,omitempty"`

	// Version is the plugin version these permissions apply to.
	Version string `yaml:"version,omitempty"`
}

// NewPermissionManager creates a new permission manager.
func NewPermissionManager(configDir string, approver PermissionApprover) (*PermissionManager, error) {
	pm := &PermissionManager{
		configDir: configDir,
		approver:  approver,
		store: &PermissionStore{
			Plugins: make(map[string]*ApprovedPlugin),
		},
	}

	// Load existing permissions
	if err := pm.load(); err != nil {
		// If file doesn't exist, that's OK - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load permissions: %w", err)
		}
	}

	return pm, nil
}

// CheckPermissions verifies that all required permissions are approved.
// If not approved, requests user approval.
func (pm *PermissionManager) CheckPermissions(pluginName string, permissions []Permission) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Built-in plugins are trusted by default
	if pm.isBuiltinPlugin(pluginName) {
		return nil
	}

	approved, exists := pm.store.Plugins[pluginName]
	if !exists {
		// No permissions approved yet, request approval
		return pm.requestApproval(pluginName, permissions)
	}

	// Check if all required permissions are approved
	missing := pm.findMissingPermissions(permissions, approved.ApprovedPermissions)
	if len(missing) > 0 {
		// Request approval for missing permissions
		return pm.requestApproval(pluginName, missing)
	}

	// Update last used time
	approved.LastUsed = time.Now()
	if err := pm.save(); err != nil {
		// Don't fail on save error, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to update last used time: %v\n", err)
	}

	return nil
}

// GrantPermissions grants specific permissions to a plugin.
func (pm *PermissionManager) GrantPermissions(pluginName string, permissions []Permission, version string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	approved, exists := pm.store.Plugins[pluginName]
	if !exists {
		approved = &ApprovedPlugin{
			ApprovedPermissions: make([]string, 0),
			ApprovedAt:          time.Now(),
			Version:             version,
		}
		pm.store.Plugins[pluginName] = approved
	}

	// Add new permissions
	for _, perm := range permissions {
		permStr := perm.String()
		if !pm.containsPermission(approved.ApprovedPermissions, permStr) {
			approved.ApprovedPermissions = append(approved.ApprovedPermissions, permStr)
		}
	}

	approved.ApprovedAt = time.Now()
	approved.Version = version

	return pm.save()
}

// RevokePermissions removes all permissions for a plugin.
func (pm *PermissionManager) RevokePermissions(pluginName string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	delete(pm.store.Plugins, pluginName)
	return pm.save()
}

// ListApprovedPlugins returns all plugins with approved permissions.
func (pm *PermissionManager) ListApprovedPlugins() map[string]*ApprovedPlugin {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]*ApprovedPlugin)
	for name, plugin := range pm.store.Plugins {
		pluginCopy := *plugin
		result[name] = &pluginCopy
	}
	return result
}

// GetApprovedPermissions returns approved permissions for a plugin.
func (pm *PermissionManager) GetApprovedPermissions(pluginName string) ([]string, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	approved, exists := pm.store.Plugins[pluginName]
	if !exists {
		return nil, false
	}
	return approved.ApprovedPermissions, true
}

// requestApproval requests user approval for permissions.
func (pm *PermissionManager) requestApproval(pluginName string, permissions []Permission) error {
	if pm.approver == nil {
		return fmt.Errorf("no permission approver configured")
	}

	approved, err := pm.approver.RequestApproval(pluginName, permissions)
	if err != nil {
		return fmt.Errorf("failed to request approval: %w", err)
	}

	if !approved {
		return fmt.Errorf("permission denied by user")
	}

	// Store approved permissions
	return pm.GrantPermissions(pluginName, permissions, "")
}

// findMissingPermissions returns permissions that are not approved.
func (pm *PermissionManager) findMissingPermissions(required []Permission, approved []string) []Permission {
	missing := make([]Permission, 0)
	for _, perm := range required {
		permStr := perm.String()
		if !pm.containsPermission(approved, permStr) {
			missing = append(missing, perm)
		}
	}
	return missing
}

// containsPermission checks if a permission string is in the list.
func (pm *PermissionManager) containsPermission(permissions []string, permission string) bool {
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

// isBuiltinPlugin checks if a plugin is a built-in plugin.
func (pm *PermissionManager) isBuiltinPlugin(pluginName string) bool {
	// Built-in plugins have specific names
	builtinPlugins := []string{
		"exec",
		"file-ops",
		"validators",
		"transformers",
	}
	for _, name := range builtinPlugins {
		if pluginName == name {
			return true
		}
	}
	return false
}

// load loads permissions from the config file.
func (pm *PermissionManager) load() error {
	path := pm.getPermissionsPath()

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var store PermissionStore
	if err := yaml.Unmarshal(data, &store); err != nil {
		return fmt.Errorf("failed to parse permissions file: %w", err)
	}

	if store.Plugins == nil {
		store.Plugins = make(map[string]*ApprovedPlugin)
	}

	pm.store = &store
	return nil
}

// save saves permissions to the config file.
func (pm *PermissionManager) save() error {
	path := pm.getPermissionsPath()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(pm.store)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write permissions file: %w", err)
	}

	return nil
}

// getPermissionsPath returns the path to the permissions file.
func (pm *PermissionManager) getPermissionsPath() string {
	return filepath.Join(pm.configDir, "plugin-permissions.yaml")
}

// ValidatePermission checks if a permission string is valid.
func ValidatePermission(permStr string) error {
	parts := strings.SplitN(permStr, ":", 2)
	if len(parts) == 0 {
		return fmt.Errorf("empty permission string")
	}

	permType := PermissionType(parts[0])
	validTypes := []PermissionType{
		PermissionExecute,
		PermissionReadEnv,
		PermissionWriteEnv,
		PermissionReadFile,
		PermissionWriteFile,
		PermissionNetwork,
		PermissionCredential,
	}

	valid := false
	for _, vt := range validTypes {
		if permType == vt {
			valid = true
			break
		}
	}

	if !valid {
		return fmt.Errorf("invalid permission type: %s", parts[0])
	}

	// Some permissions require a resource
	requiresResource := []PermissionType{
		PermissionExecute,
		PermissionReadEnv,
		PermissionWriteEnv,
		PermissionReadFile,
		PermissionWriteFile,
		PermissionNetwork,
	}

	needsResource := false
	for _, rt := range requiresResource {
		if permType == rt {
			needsResource = true
			break
		}
	}

	if needsResource && len(parts) < 2 {
		return fmt.Errorf("permission type %s requires a resource", permType)
	}

	return nil
}

// MatchPermission checks if a request matches an approved permission pattern.
// Supports wildcards (*) in patterns.
func MatchPermission(pattern, request string) bool {
	// Exact match
	if pattern == request {
		return true
	}

	// Wildcard match
	if strings.Contains(pattern, "*") {
		return matchWildcard(pattern, request)
	}

	return false
}

// matchWildcard performs wildcard pattern matching.
func matchWildcard(pattern, str string) bool {
	// Simple wildcard matching (supports * only)
	parts := strings.Split(pattern, "*")
	if len(parts) == 1 {
		return pattern == str
	}

	// Check prefix
	if !strings.HasPrefix(str, parts[0]) {
		return false
	}
	str = str[len(parts[0]):]

	// Check suffix
	if !strings.HasSuffix(str, parts[len(parts)-1]) {
		return false
	}
	str = str[:len(str)-len(parts[len(parts)-1])]

	// Check middle parts
	for i := 1; i < len(parts)-1; i++ {
		idx := strings.Index(str, parts[i])
		if idx < 0 {
			return false
		}
		str = str[idx+len(parts[i]):]
	}

	return true
}

// DefaultApprover provides a CLI-based permission approval interface.
type DefaultApprover struct{}

// RequestApproval asks the user to approve permissions via CLI.
func (a *DefaultApprover) RequestApproval(pluginName string, permissions []Permission) (bool, error) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "⚠️  Plugin '%s' requests the following permissions:\n", pluginName)
	fmt.Fprintf(os.Stderr, "\n")

	for _, perm := range permissions {
		fmt.Fprintf(os.Stderr, "  • %s", perm.String())
		if perm.Description != "" {
			fmt.Fprintf(os.Stderr, " - %s", perm.Description)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "Grant these permissions? [y/N]: ")

	var response string
	if _, err := fmt.Scanln(&response); err != nil {
		return false, nil
	}

	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// AutoApprover automatically approves all permissions (for testing only).
type AutoApprover struct{}

// RequestApproval automatically approves all permissions.
func (a *AutoApprover) RequestApproval(pluginName string, permissions []Permission) (bool, error) {
	return true, nil
}
