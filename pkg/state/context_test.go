package state

import (
	"testing"
	"time"
)

func TestNewContext(t *testing.T) {
	ctx := NewContext("test")

	if ctx.Name != "test" {
		t.Errorf("Expected name to be 'test', got %s", ctx.Name)
	}

	if ctx.Fields == nil {
		t.Error("Expected fields to be initialized")
	}

	if ctx.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
}

func TestContextSetGet(t *testing.T) {
	ctx := NewContext("test")

	ctx.Set("cluster", "cluster-123")
	ctx.Set("region", "us-east-1")

	cluster, exists := ctx.Get("cluster")
	if !exists {
		t.Error("Expected cluster to exist")
	}
	if cluster != "cluster-123" {
		t.Errorf("Expected cluster to be 'cluster-123', got %s", cluster)
	}

	region, exists := ctx.Get("region")
	if !exists {
		t.Error("Expected region to exist")
	}
	if region != "us-east-1" {
		t.Errorf("Expected region to be 'us-east-1', got %s", region)
	}

	_, exists = ctx.Get("nonexistent")
	if exists {
		t.Error("Expected nonexistent key to not exist")
	}
}

func TestContextHas(t *testing.T) {
	ctx := NewContext("test")
	ctx.Set("cluster", "cluster-123")

	if !ctx.Has("cluster") {
		t.Error("Expected Has to return true for existing key")
	}

	if ctx.Has("nonexistent") {
		t.Error("Expected Has to return false for nonexistent key")
	}
}

func TestContextDelete(t *testing.T) {
	ctx := NewContext("test")
	ctx.Set("cluster", "cluster-123")

	if !ctx.Has("cluster") {
		t.Error("Expected cluster to exist before delete")
	}

	ctx.Delete("cluster")

	if ctx.Has("cluster") {
		t.Error("Expected cluster to not exist after delete")
	}
}

func TestContextClear(t *testing.T) {
	ctx := NewContext("test")
	ctx.Set("cluster", "cluster-123")
	ctx.Set("region", "us-east-1")

	ctx.Clear()

	if ctx.Has("cluster") || ctx.Has("region") {
		t.Error("Expected all fields to be cleared")
	}

	if len(ctx.Fields) != 0 {
		t.Error("Expected fields map to be empty")
	}
}

func TestContextClone(t *testing.T) {
	ctx := NewContext("test")
	ctx.Description = "Test context"
	ctx.Set("cluster", "cluster-123")
	ctx.Set("region", "us-east-1")

	clone := ctx.Clone()

	if clone.Name != ctx.Name {
		t.Error("Expected clone to have same name")
	}

	if clone.Description != ctx.Description {
		t.Error("Expected clone to have same description")
	}

	cluster, _ := clone.Get("cluster")
	if cluster != "cluster-123" {
		t.Error("Expected clone to have same cluster value")
	}

	// Modify clone and verify original is unchanged
	clone.Set("cluster", "different-cluster")
	originalCluster, _ := ctx.Get("cluster")
	if originalCluster != "cluster-123" {
		t.Error("Expected original context to be unchanged after modifying clone")
	}
}

func TestContextMerge(t *testing.T) {
	ctx1 := NewContext("test1")
	ctx1.Set("cluster", "cluster-123")
	ctx1.Set("region", "us-east-1")

	ctx2 := NewContext("test2")
	ctx2.Set("region", "us-west-2")
	ctx2.Set("profile", "production")

	ctx1.Merge(ctx2)

	// Check that region was overwritten
	region, _ := ctx1.Get("region")
	if region != "us-west-2" {
		t.Errorf("Expected region to be overwritten to 'us-west-2', got %s", region)
	}

	// Check that cluster remained
	cluster, _ := ctx1.Get("cluster")
	if cluster != "cluster-123" {
		t.Error("Expected cluster to remain unchanged")
	}

	// Check that profile was added
	profile, _ := ctx1.Get("profile")
	if profile != "production" {
		t.Error("Expected profile to be added from merged context")
	}
}

func TestContextMarkUsed(t *testing.T) {
	ctx := NewContext("test")

	if ctx.UseCount != 0 {
		t.Error("Expected initial use count to be 0")
	}

	ctx.MarkUsed()

	if ctx.UseCount != 1 {
		t.Errorf("Expected use count to be 1, got %d", ctx.UseCount)
	}

	if ctx.LastUsed.IsZero() {
		t.Error("Expected LastUsed to be set")
	}

	lastUsed1 := ctx.LastUsed
	time.Sleep(10 * time.Millisecond)
	ctx.MarkUsed()

	if ctx.UseCount != 2 {
		t.Errorf("Expected use count to be 2, got %d", ctx.UseCount)
	}

	if !ctx.LastUsed.After(lastUsed1) {
		t.Error("Expected LastUsed to be updated")
	}
}

func TestContextValidate(t *testing.T) {
	ctx := NewContext("test")
	if err := ctx.Validate(); err != nil {
		t.Errorf("Expected valid context to pass validation, got error: %v", err)
	}

	emptyCtx := &Context{}
	if err := emptyCtx.Validate(); err == nil {
		t.Error("Expected context with empty name to fail validation")
	}
}

func TestContextBuilder(t *testing.T) {
	ctx := NewContextBuilder("production").
		WithDescription("Production environment").
		WithField("cluster", "prod-cluster").
		WithField("region", "us-east-1").
		WithFields(map[string]string{
			"profile": "prod-profile",
			"zone":    "us-east-1a",
		}).
		Build()

	if ctx.Name != "production" {
		t.Errorf("Expected name to be 'production', got %s", ctx.Name)
	}

	if ctx.Description != "Production environment" {
		t.Error("Expected description to be set")
	}

	cluster, _ := ctx.Get("cluster")
	if cluster != "prod-cluster" {
		t.Error("Expected cluster field to be set")
	}

	profile, _ := ctx.Get("profile")
	if profile != "prod-profile" {
		t.Error("Expected profile field to be set from WithFields")
	}
}

func TestContextManager(t *testing.T) {
	// Create temp state manager
	mgr, _ := NewManager("testcli")
	cm := NewContextManager(mgr)

	// Create context
	err := cm.Create("staging", "Staging environment", map[string]string{
		"cluster": "staging-cluster",
		"region":  "us-west-2",
	})
	if err != nil {
		t.Fatalf("Failed to create context: %v", err)
	}

	// List contexts
	contexts, err := cm.List()
	if err != nil {
		t.Fatalf("Failed to list contexts: %v", err)
	}

	if _, exists := contexts["staging"]; !exists {
		t.Error("Expected staging context to exist in list")
	}

	// Switch context
	if err := cm.SwitchTo("staging"); err != nil {
		t.Fatalf("Failed to switch context: %v", err)
	}

	if cm.CurrentName() != "staging" {
		t.Errorf("Expected current context to be 'staging', got %s", cm.CurrentName())
	}

	// Verify context was marked as used
	ctx, _ := mgr.GetContext("staging")
	if ctx.UseCount != 1 {
		t.Errorf("Expected use count to be 1 after switch, got %d", ctx.UseCount)
	}

	// Update context
	err = cm.Update("staging", map[string]string{
		"profile": "staging-profile",
	})
	if err != nil {
		t.Fatalf("Failed to update context: %v", err)
	}

	ctx, _ = mgr.GetContext("staging")
	profile, _ := ctx.Get("profile")
	if profile != "staging-profile" {
		t.Error("Expected profile to be updated")
	}
}

func TestContextManagerRename(t *testing.T) {
	mgr, _ := NewManager("testcli")
	cm := NewContextManager(mgr)

	// Create and switch to a context
	cm.Create("old-name", "Test context", map[string]string{
		"cluster": "test-cluster",
	})
	cm.SwitchTo("old-name")

	// Rename
	err := cm.Rename("old-name", "new-name")
	if err != nil {
		t.Fatalf("Failed to rename context: %v", err)
	}

	// Verify old name doesn't exist
	_, err = mgr.GetContext("old-name")
	if err == nil {
		t.Error("Expected old context name to not exist")
	}

	// Verify new name exists
	ctx, err := mgr.GetContext("new-name")
	if err != nil {
		t.Fatalf("Failed to get renamed context: %v", err)
	}

	cluster, _ := ctx.Get("cluster")
	if cluster != "test-cluster" {
		t.Error("Expected renamed context to retain fields")
	}

	// Verify current context was updated
	if cm.CurrentName() != "new-name" {
		t.Error("Expected current context to be updated to new name")
	}
}

func TestContextManagerExportImport(t *testing.T) {
	mgr, _ := NewManager("testcli")
	cm := NewContextManager(mgr)

	// Create a context
	cm.Create("export-test", "Test export", map[string]string{
		"cluster": "test-cluster",
		"region":  "us-east-1",
	})

	// Export
	exported, err := cm.Export("export-test")
	if err != nil {
		t.Fatalf("Failed to export context: %v", err)
	}

	if exported["name"] != "export-test" {
		t.Error("Expected exported name to match")
	}

	fields, ok := exported["fields"].(map[string]string)
	if !ok {
		t.Fatal("Expected fields to be map[string]string")
	}

	if fields["cluster"] != "test-cluster" {
		t.Error("Expected exported fields to match")
	}

	// Delete original
	cm.Delete("export-test")

	// Import
	err = cm.Import(exported)
	if err != nil {
		t.Fatalf("Failed to import context: %v", err)
	}

	// Verify imported context
	ctx, err := mgr.GetContext("export-test")
	if err != nil {
		t.Fatalf("Failed to get imported context: %v", err)
	}

	cluster, _ := ctx.Get("cluster")
	if cluster != "test-cluster" {
		t.Error("Expected imported context to have correct fields")
	}
}

func TestContextManagerDelete(t *testing.T) {
	mgr, _ := NewManager("testcli")
	cm := NewContextManager(mgr)

	cm.Create("delete-test", "Test delete", map[string]string{})

	err := cm.Delete("delete-test")
	if err != nil {
		t.Fatalf("Failed to delete context: %v", err)
	}

	_, err = mgr.GetContext("delete-test")
	if err == nil {
		t.Error("Expected deleted context to not exist")
	}
}
