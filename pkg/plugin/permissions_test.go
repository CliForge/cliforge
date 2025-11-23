package plugin

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewPermissionManager(t *testing.T) {
	tmpDir := t.TempDir()

	pm, err := NewPermissionManager(tmpDir, &AutoApprover{})
	if err != nil {
		t.Fatalf("NewPermissionManager() error = %v", err)
	}

	if pm == nil {
		t.Fatal("NewPermissionManager() returned nil")
	}

	if pm.configDir != tmpDir {
		t.Errorf("configDir = %v, want %v", pm.configDir, tmpDir)
	}
}

func TestPermissionManager_GrantPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})

	permissions := []Permission{
		{Type: PermissionExecute, Resource: "aws"},
		{Type: PermissionReadFile, Resource: "/tmp/*"},
	}

	err := pm.GrantPermissions("test-plugin", permissions, "1.0.0")
	if err != nil {
		t.Fatalf("GrantPermissions() error = %v", err)
	}

	// Verify permissions were granted
	approved, exists := pm.GetApprovedPermissions("test-plugin")
	if !exists {
		t.Fatal("Permissions not found after granting")
	}

	if len(approved) != 2 {
		t.Errorf("Approved permissions count = %v, want 2", len(approved))
	}

	// Verify permissions file was created
	permFile := filepath.Join(tmpDir, "plugin-permissions.yaml")
	if _, err := os.Stat(permFile); os.IsNotExist(err) {
		t.Error("Permissions file was not created")
	}
}

func TestPermissionManager_CheckPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})

	// Test built-in plugin (should pass without approval)
	err := pm.CheckPermissions("exec", []Permission{
		{Type: PermissionExecute, Resource: "*"},
	})
	if err != nil {
		t.Errorf("CheckPermissions() for built-in plugin failed: %v", err)
	}

	// Test external plugin without approval (should request approval)
	permissions := []Permission{
		{Type: PermissionExecute, Resource: "test"},
	}

	err = pm.CheckPermissions("external-plugin", permissions)
	if err != nil {
		t.Errorf("CheckPermissions() with auto-approver failed: %v", err)
	}

	// Verify approval was saved
	approved, exists := pm.GetApprovedPermissions("external-plugin")
	if !exists {
		t.Error("Permissions were not saved after approval")
	}

	if len(approved) != 1 {
		t.Errorf("Approved permissions count = %v, want 1", len(approved))
	}
}

func TestPermissionManager_RevokePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})

	// Grant permissions first
	permissions := []Permission{
		{Type: PermissionExecute, Resource: "test"},
	}
	pm.GrantPermissions("test-plugin", permissions, "1.0.0")

	// Verify permissions exist
	_, exists := pm.GetApprovedPermissions("test-plugin")
	if !exists {
		t.Fatal("Permissions not found after granting")
	}

	// Revoke permissions
	err := pm.RevokePermissions("test-plugin")
	if err != nil {
		t.Fatalf("RevokePermissions() error = %v", err)
	}

	// Verify permissions were revoked
	_, exists = pm.GetApprovedPermissions("test-plugin")
	if exists {
		t.Error("Permissions still exist after revocation")
	}
}

func TestPermissionManager_ListApprovedPlugins(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})

	// Grant permissions to multiple plugins
	pm.GrantPermissions("plugin1", []Permission{{Type: PermissionExecute, Resource: "test1"}}, "1.0.0")
	pm.GrantPermissions("plugin2", []Permission{{Type: PermissionExecute, Resource: "test2"}}, "1.0.0")

	approved := pm.ListApprovedPlugins()
	if len(approved) != 2 {
		t.Errorf("ListApprovedPlugins() count = %v, want 2", len(approved))
	}

	if _, exists := approved["plugin1"]; !exists {
		t.Error("plugin1 not in approved list")
	}

	if _, exists := approved["plugin2"]; !exists {
		t.Error("plugin2 not in approved list")
	}
}

func TestPermissionManager_Persistence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create first manager and grant permissions
	pm1, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	permissions := []Permission{
		{Type: PermissionExecute, Resource: "aws"},
	}
	pm1.GrantPermissions("test-plugin", permissions, "1.0.0")

	// Create second manager and verify permissions persisted
	pm2, err := NewPermissionManager(tmpDir, &AutoApprover{})
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}

	approved, exists := pm2.GetApprovedPermissions("test-plugin")
	if !exists {
		t.Fatal("Permissions not persisted")
	}

	if len(approved) != 1 {
		t.Errorf("Persisted permissions count = %v, want 1", len(approved))
	}

	if approved[0] != "execute:aws" {
		t.Errorf("Persisted permission = %v, want execute:aws", approved[0])
	}
}

func TestPermissionManager_MissingPermissions(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})

	// Grant initial permissions
	pm.GrantPermissions("test-plugin", []Permission{
		{Type: PermissionExecute, Resource: "aws"},
	}, "1.0.0")

	// Request additional permissions
	allPermissions := []Permission{
		{Type: PermissionExecute, Resource: "aws"},
		{Type: PermissionReadFile, Resource: "/tmp/*"},
	}

	err := pm.CheckPermissions("test-plugin", allPermissions)
	if err != nil {
		t.Fatalf("CheckPermissions() error = %v", err)
	}

	// Verify new permission was added
	approved, _ := pm.GetApprovedPermissions("test-plugin")
	if len(approved) != 2 {
		t.Errorf("Approved permissions count = %v, want 2", len(approved))
	}
}

func TestApprovedPlugin(t *testing.T) {
	now := time.Now()
	plugin := &ApprovedPlugin{
		ApprovedPermissions: []string{"execute:aws", "read:file:/tmp/*"},
		ApprovedAt:          now,
		LastUsed:            now.Add(1 * time.Hour),
		Version:             "1.0.0",
	}

	if len(plugin.ApprovedPermissions) != 2 {
		t.Errorf("ApprovedPermissions count = %v, want 2", len(plugin.ApprovedPermissions))
	}

	if plugin.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", plugin.Version)
	}

	if plugin.LastUsed.Before(plugin.ApprovedAt) {
		t.Error("LastUsed should be after ApprovedAt")
	}
}

func TestDefaultApprover(t *testing.T) {
	// DefaultApprover requires user input, so we can't fully test it here
	// Just verify it implements the interface
	var _ PermissionApprover = &DefaultApprover{}
}

func TestAutoApprover(t *testing.T) {
	approver := &AutoApprover{}

	permissions := []Permission{
		{Type: PermissionExecute, Resource: "test"},
	}

	approved, err := approver.RequestApproval("test-plugin", permissions)
	if err != nil {
		t.Fatalf("RequestApproval() error = %v", err)
	}

	if !approved {
		t.Error("AutoApprover should always approve")
	}
}

func TestMatchWildcard(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		str      string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "execute:aws",
			str:      "execute:aws",
			expected: true,
		},
		{
			name:     "wildcard at end",
			pattern:  "execute:*",
			str:      "execute:aws",
			expected: true,
		},
		{
			name:     "wildcard at start",
			pattern:  "*:aws",
			str:      "execute:aws",
			expected: true,
		},
		{
			name:     "wildcard in middle",
			pattern:  "read:file:/home/*/data",
			str:      "read:file:/home/user/data",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "execute:aws",
			str:      "execute:kubectl",
			expected: false,
		},
		{
			name:     "partial match should fail",
			pattern:  "execute:aw",
			str:      "execute:aws",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchWildcard(tt.pattern, tt.str); got != tt.expected {
				t.Errorf("matchWildcard(%q, %q) = %v, want %v", tt.pattern, tt.str, got, tt.expected)
			}
		})
	}
}

func TestPermissionStore(t *testing.T) {
	store := &PermissionStore{
		Plugins: make(map[string]*ApprovedPlugin),
	}

	now := time.Now()
	store.Plugins["test-plugin"] = &ApprovedPlugin{
		ApprovedPermissions: []string{"execute:test"},
		ApprovedAt:          now,
		Version:             "1.0.0",
	}

	if len(store.Plugins) != 1 {
		t.Errorf("Plugins count = %v, want 1", len(store.Plugins))
	}

	plugin, exists := store.Plugins["test-plugin"]
	if !exists {
		t.Fatal("test-plugin not found in store")
	}

	if plugin.Version != "1.0.0" {
		t.Errorf("Version = %v, want 1.0.0", plugin.Version)
	}
}
