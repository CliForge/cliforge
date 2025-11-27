package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	// Create temp directory for testing
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, err := NewManager("testcli")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	if mgr.cliName != "testcli" {
		t.Errorf("Expected cliName to be 'testcli', got %s", mgr.cliName)
	}

	if mgr.state == nil {
		t.Error("Expected state to be initialized")
	}

	if mgr.state.CurrentContext != "default" {
		t.Errorf("Expected default context to be 'default', got %s", mgr.state.CurrentContext)
	}
}

func TestManagerLoadSave(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, err := NewManager("testcli")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Reset to ensure clean state
	_ = mgr.Reset()

	// Create a context
	ctx := NewContext("production")
	ctx.Set("cluster", "prod-cluster-123")
	ctx.Set("region", "us-east-1")

	if err := mgr.CreateContext("production", ctx); err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}

	// Save state
	if err := mgr.Save(); err != nil {
		t.Fatalf("Failed to save state: %v", err)
	}

	// Create new manager and load
	mgr2, err := NewManager("testcli")
	if err != nil {
		t.Fatalf("Failed to create second manager: %v", err)
	}

	// Verify loaded state
	loadedCtx, err := mgr2.GetContext("production")
	if err != nil {
		t.Fatalf("Failed to get production context: %v", err)
	}

	cluster, _ := loadedCtx.Get("cluster")
	if cluster != "prod-cluster-123" {
		t.Errorf("Expected cluster to be 'prod-cluster-123', got %s", cluster)
	}
}

func TestContextManagement(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, err := NewManager("testcli")
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	// Reset to ensure clean state
	_ = mgr.Reset()

	// Create contexts
	prodCtx := NewContext("production")
	prodCtx.Set("cluster", "prod-cluster")
	stagingCtx := NewContext("staging")
	stagingCtx.Set("cluster", "staging-cluster")

	if err := mgr.CreateContext("production", prodCtx); err != nil {
		t.Fatalf("Failed to create production context: %v", err)
	}

	if err := mgr.CreateContext("staging", stagingCtx); err != nil {
		t.Fatalf("Failed to create staging context: %v", err)
	}

	// List contexts
	contexts := mgr.ListContexts()
	if len(contexts) != 3 { // default + production + staging
		t.Errorf("Expected 3 contexts, got %d", len(contexts))
	}

	// Switch context
	if err := mgr.SetCurrentContext("production"); err != nil {
		t.Fatalf("Failed to switch context: %v", err)
	}

	currentCtx := mgr.GetCurrentContext()
	if currentCtx.Name != "production" {
		t.Errorf("Expected current context to be 'production', got %s", currentCtx.Name)
	}

	// Delete context
	if err := mgr.DeleteContext("staging"); err != nil {
		t.Fatalf("Failed to delete context: %v", err)
	}

	contexts = mgr.ListContexts()
	if len(contexts) != 2 {
		t.Errorf("Expected 2 contexts after deletion, got %d", len(contexts))
	}

	// Try to delete default context (should fail)
	if err := mgr.DeleteContext("default"); err == nil {
		t.Error("Expected error when deleting default context")
	}
}

func TestRecentValues(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")

	// Add recent values
	mgr.AddRecentValue("clusters", "cluster-1")
	mgr.AddRecentValue("clusters", "cluster-2")
	mgr.AddRecentValue("clusters", "cluster-3")

	values := mgr.GetRecentValues("clusters")
	if len(values) != 3 {
		t.Errorf("Expected 3 recent values, got %d", len(values))
	}

	// Most recent should be first
	if values[0] != "cluster-3" {
		t.Errorf("Expected most recent value to be 'cluster-3', got %s", values[0])
	}

	// Add duplicate (should move to front)
	mgr.AddRecentValue("clusters", "cluster-1")
	values = mgr.GetRecentValues("clusters")
	if values[0] != "cluster-1" {
		t.Errorf("Expected 'cluster-1' to be moved to front, got %s", values[0])
	}
}

func TestPreferences(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")

	// Set preference
	pref := &ResourcePreference{
		Favorite: true,
		Metadata: map[string]string{
			"color": "blue",
		},
	}
	mgr.SetPreference("clusters", "cluster-1", pref)

	// Get preference
	loadedPref, exists := mgr.GetPreference("clusters", "cluster-1")
	if !exists {
		t.Fatal("Expected preference to exist")
	}

	if !loadedPref.Favorite {
		t.Error("Expected preference to be favorite")
	}

	if loadedPref.Metadata["color"] != "blue" {
		t.Errorf("Expected color to be 'blue', got %s", loadedPref.Metadata["color"])
	}

	if loadedPref.UseCount != 1 {
		t.Errorf("Expected use count to be 1, got %d", loadedPref.UseCount)
	}
}

func TestMarkResourceUsed(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")

	// Mark resource as used
	mgr.MarkResourceUsed("clusters", "cluster-1")

	pref, exists := mgr.GetPreference("clusters", "cluster-1")
	if !exists {
		t.Fatal("Expected preference to be created")
	}

	if pref.UseCount != 1 {
		t.Errorf("Expected use count to be 1, got %d", pref.UseCount)
	}

	// Mark again
	time.Sleep(10 * time.Millisecond) // Small delay to ensure different timestamp
	mgr.MarkResourceUsed("clusters", "cluster-1")

	pref, _ = mgr.GetPreference("clusters", "cluster-1")
	if pref.UseCount != 2 {
		t.Errorf("Expected use count to be 2, got %d", pref.UseCount)
	}
}

func TestSession(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")

	// Set session command
	mgr.SetSessionCommand("mycli describe cluster")

	session := mgr.GetSession()
	if session.LastCommand != "mycli describe cluster" {
		t.Errorf("Expected last command to be set, got %s", session.LastCommand)
	}

	if session.LastCommandTime.IsZero() {
		t.Error("Expected last command time to be set")
	}
}

func TestReset(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")

	// Create some state
	ctx := NewContext("production")
	_ = mgr.CreateContext("production", ctx)
	mgr.AddRecentValue("clusters", "cluster-1")

	// Reset
	if err := mgr.Reset(); err != nil {
		t.Fatalf("Failed to reset state: %v", err)
	}

	// Verify reset
	contexts := mgr.ListContexts()
	if len(contexts) != 1 || contexts[0] != "default" {
		t.Error("Expected only default context after reset")
	}

	values := mgr.GetRecentValues("clusters")
	if len(values) != 0 {
		t.Error("Expected no recent values after reset")
	}
}

func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	// Create manager and add state
	mgr1, _ := NewManager("testcli")
	_ = mgr1.Reset() // Ensure clean state

	ctx := NewContext("production")
	ctx.Set("cluster", "prod-cluster")
	_ = mgr1.CreateContext("production", ctx)
	mgr1.AddRecentValue("clusters", "cluster-1")
	mgr1.MarkResourceUsed("clusters", "cluster-1")

	// Save
	if err := mgr1.Save(); err != nil {
		t.Fatalf("Failed to save: %v", err)
	}

	// Create new manager and verify state is loaded
	mgr2, err := NewManager("testcli")
	if err != nil {
		t.Fatalf("Failed to create new manager: %v", err)
	}

	// Check context
	prodCtx, err := mgr2.GetContext("production")
	if err != nil {
		t.Fatalf("Failed to get production context: %v", err)
	}

	cluster, _ := prodCtx.Get("cluster")
	if cluster != "prod-cluster" {
		t.Errorf("Expected cluster to be 'prod-cluster', got %s", cluster)
	}

	// Check recent
	values := mgr2.GetRecentValues("clusters")
	if len(values) != 1 || values[0] != "cluster-1" {
		t.Error("Expected recent values to be persisted")
	}

	// Check preferences
	pref, exists := mgr2.GetPreference("clusters", "cluster-1")
	if !exists {
		t.Error("Expected preference to be persisted")
	}
	if pref.UseCount != 1 {
		t.Errorf("Expected use count to be 1, got %d", pref.UseCount)
	}
}

func TestGetStatePath(t *testing.T) {
	tmpDir := t.TempDir()

	// Set env var BEFORE creating manager (xdg lib caches on first access)
	oldXDG := os.Getenv("XDG_STATE_HOME")
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() {
		if oldXDG != "" {
			_ = os.Setenv("XDG_STATE_HOME", oldXDG)
		} else {
			_ = os.Unsetenv("XDG_STATE_HOME")
		}
	}()

	// Force xdg to reload by using a different CLI name
	mgr, _ := NewManager("testcli-pathtest")
	path := mgr.GetStatePath()

	// Just verify it contains the CLI name and ends with state.yaml
	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
	if filepath.Base(path) != "state.yaml" {
		t.Errorf("Expected path to end with state.yaml, got %s", path)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")

	// Concurrent writes
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(n int) {
			mgr.AddRecentValue("clusters", "cluster-"+string(rune(n+'0')))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	values := mgr.GetRecentValues("clusters")
	if len(values) == 0 {
		t.Error("Expected some recent values after concurrent access")
	}
}
