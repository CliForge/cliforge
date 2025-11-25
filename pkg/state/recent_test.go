package state

import (
	"testing"
	"time"
)

func TestNewRecent(t *testing.T) {
	r := NewRecent()

	if r.Lists == nil {
		t.Error("Expected lists to be initialized")
	}

	if r.MaxPerList != DefaultMaxRecentEntries {
		t.Errorf("Expected maxPerList to be %d, got %d", DefaultMaxRecentEntries, r.MaxPerList)
	}
}

func TestRecentAdd(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")

	values := r.Get("clusters")
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	// Most recent should be first
	if values[0] != "cluster-2" {
		t.Errorf("Expected most recent to be 'cluster-2', got %s", values[0])
	}
}

func TestRecentAddDuplicate(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")
	r.Add("clusters", "cluster-3")

	// Add duplicate
	r.Add("clusters", "cluster-1")

	values := r.Get("clusters")

	// Should still have 3 values
	if len(values) != 3 {
		t.Errorf("Expected 3 values, got %d", len(values))
	}

	// cluster-1 should be moved to front
	if values[0] != "cluster-1" {
		t.Errorf("Expected 'cluster-1' to be moved to front, got %s", values[0])
	}

	// Verify use count was incremented
	items := r.GetWithMetadata("clusters")
	for _, item := range items {
		if item.Value == "cluster-1" {
			if item.UseCount != 2 {
				t.Errorf("Expected use count to be 2, got %d", item.UseCount)
			}
			break
		}
	}
}

func TestRecentMaxEntries(t *testing.T) {
	r := NewRecentWithMax(3)

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")
	r.Add("clusters", "cluster-3")
	r.Add("clusters", "cluster-4")

	values := r.Get("clusters")

	if len(values) != 3 {
		t.Errorf("Expected max 3 values, got %d", len(values))
	}

	// cluster-1 (oldest) should be removed
	for _, v := range values {
		if v == "cluster-1" {
			t.Error("Expected oldest value to be removed")
		}
	}

	// cluster-4 (newest) should be first
	if values[0] != "cluster-4" {
		t.Errorf("Expected newest value to be first, got %s", values[0])
	}
}

func TestRecentGet(t *testing.T) {
	r := NewRecent()

	values := r.Get("nonexistent")
	if len(values) != 0 {
		t.Error("Expected empty slice for non-existent list")
	}

	r.Add("clusters", "cluster-1")
	values = r.Get("clusters")
	if len(values) != 1 {
		t.Error("Expected 1 value")
	}
}

func TestRecentGetWithMetadata(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	time.Sleep(10 * time.Millisecond)
	r.Add("clusters", "cluster-2")

	items := r.GetWithMetadata("clusters")

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	// Most recent first
	if items[0].Value != "cluster-2" {
		t.Error("Expected most recent item first")
	}

	if items[0].LastUsed.IsZero() {
		t.Error("Expected LastUsed to be set")
	}

	if items[0].UseCount != 1 {
		t.Errorf("Expected use count to be 1, got %d", items[0].UseCount)
	}
}

func TestRecentGetTop(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")
	r.Add("clusters", "cluster-3")
	r.Add("clusters", "cluster-4")

	top2 := r.GetTop("clusters", 2)

	if len(top2) != 2 {
		t.Errorf("Expected 2 values, got %d", len(top2))
	}

	if top2[0] != "cluster-4" || top2[1] != "cluster-3" {
		t.Error("Expected top 2 most recent values")
	}
}

func TestRecentRemove(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")

	r.Remove("clusters", "cluster-1")

	values := r.Get("clusters")
	if len(values) != 1 {
		t.Errorf("Expected 1 value after removal, got %d", len(values))
	}

	if values[0] != "cluster-2" {
		t.Error("Expected correct value to remain")
	}
}

func TestRecentClear(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")

	r.Clear("clusters")

	values := r.Get("clusters")
	if len(values) != 0 {
		t.Error("Expected empty list after clear")
	}
}

func TestRecentClearAll(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("regions", "us-east-1")

	r.ClearAll()

	if len(r.Get("clusters")) != 0 {
		t.Error("Expected clusters list to be empty")
	}

	if len(r.Get("regions")) != 0 {
		t.Error("Expected regions list to be empty")
	}
}

func TestRecentListNames(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("regions", "us-east-1")
	r.Add("profiles", "production")

	names := r.ListNames()

	if len(names) != 3 {
		t.Errorf("Expected 3 list names, got %d", len(names))
	}

	// Check all names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	if !nameMap["clusters"] || !nameMap["regions"] || !nameMap["profiles"] {
		t.Error("Expected all list names to be present")
	}
}

func TestRecentGetList(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")

	list, exists := r.GetList("clusters")
	if !exists {
		t.Error("Expected list to exist")
	}

	if list.Name != "clusters" {
		t.Error("Expected list name to match")
	}

	if len(list.Entries) != 1 {
		t.Error("Expected 1 entry in list")
	}

	// Verify it's a copy
	list.Entries[0].Value = "modified"
	original := r.Get("clusters")
	if original[0] == "modified" {
		t.Error("Expected GetList to return a copy")
	}
}

func TestRecentSetMax(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")
	r.Add("clusters", "cluster-3")
	r.Add("clusters", "cluster-4")

	r.SetMax("clusters", 2)

	values := r.Get("clusters")
	if len(values) != 2 {
		t.Errorf("Expected 2 values after setting max, got %d", len(values))
	}

	// Should keep most recent
	if values[0] != "cluster-4" {
		t.Error("Expected most recent values to be kept")
	}
}

func TestRecentGetMostUsed(t *testing.T) {
	r := NewRecent()

	// Add with different use counts
	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")
	r.Add("clusters", "cluster-1") // cluster-1 used twice
	r.Add("clusters", "cluster-3")
	r.Add("clusters", "cluster-1") // cluster-1 used three times

	mostUsed := r.GetMostUsed("clusters", 0)

	// cluster-1 should be first (most used)
	if mostUsed[0] != "cluster-1" {
		t.Errorf("Expected 'cluster-1' to be most used, got %s", mostUsed[0])
	}

	// Verify use count via metadata
	items := r.GetWithMetadata("clusters")
	for _, item := range items {
		if item.Value == "cluster-1" {
			if item.UseCount != 3 {
				t.Errorf("Expected cluster-1 use count to be 3, got %d", item.UseCount)
			}
		}
	}
}

func TestRecentGetByPattern(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "prod-cluster-123")
	r.Add("clusters", "staging-cluster-456")
	r.Add("clusters", "dev-cluster-789")

	matches := r.GetByPattern("clusters", "prod")
	if len(matches) != 1 {
		t.Errorf("Expected 1 match for 'prod', got %d", len(matches))
	}

	if matches[0] != "prod-cluster-123" {
		t.Error("Expected correct match")
	}

	// Case insensitive
	matches = r.GetByPattern("clusters", "PROD")
	if len(matches) != 1 {
		t.Error("Expected pattern matching to be case insensitive")
	}

	// Substring match
	matches = r.GetByPattern("clusters", "cluster")
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches for 'cluster', got %d", len(matches))
	}
}

func TestRecentPrune(t *testing.T) {
	r := NewRecent()

	now := time.Now()
	past := now.Add(-2 * time.Hour)

	// Manually create items with specific timestamps
	r.Add("clusters", "old-cluster")
	list, _ := r.GetList("clusters")
	list.Entries[0].LastUsed = past

	time.Sleep(10 * time.Millisecond)
	r.Add("clusters", "new-cluster")

	// Update the list
	r.Lists["clusters"] = list
	r.Add("clusters", "new-cluster") // Re-add to update

	r.Prune("clusters", 1*time.Hour)

	values := r.Get("clusters")
	if len(values) != 1 {
		t.Errorf("Expected 1 value after pruning, got %d", len(values))
	}

	if values[0] != "new-cluster" {
		t.Error("Expected old entries to be pruned")
	}
}

func TestRecentPruneAll(t *testing.T) {
	r := NewRecent()

	now := time.Now()
	past := now.Add(-2 * time.Hour)

	r.Add("clusters", "cluster-1")
	r.Add("regions", "region-1")

	// Set old timestamps
	if list, exists := r.GetList("clusters"); exists {
		list.Entries[0].LastUsed = past
		r.Lists["clusters"] = list
	}

	if list, exists := r.GetList("regions"); exists {
		list.Entries[0].LastUsed = past
		r.Lists["regions"] = list
	}

	r.PruneAll(1 * time.Hour)

	if len(r.Get("clusters")) != 0 {
		t.Error("Expected clusters to be pruned")
	}

	if len(r.Get("regions")) != 0 {
		t.Error("Expected regions to be pruned")
	}
}

func TestRecentGetStats(t *testing.T) {
	r := NewRecent()

	r.Add("clusters", "cluster-1")
	r.Add("clusters", "cluster-2")
	r.Add("clusters", "cluster-1") // Use cluster-1 again

	stats := r.GetStats("clusters")

	if stats.ListName != "clusters" {
		t.Error("Expected list name to match")
	}

	if stats.TotalItems != 2 {
		t.Errorf("Expected 2 total items, got %d", stats.TotalItems)
	}

	if stats.MostUsedValue != "cluster-1" {
		t.Errorf("Expected most used to be 'cluster-1', got %s", stats.MostUsedValue)
	}

	if stats.MostUsedCount != 2 {
		t.Errorf("Expected most used count to be 2, got %d", stats.MostUsedCount)
	}

	if stats.MostRecentValue != "cluster-1" {
		t.Errorf("Expected most recent to be 'cluster-1', got %s", stats.MostRecentValue)
	}

	if stats.AverageUseCount != 1.5 { // (2 + 1) / 2
		t.Errorf("Expected average use count to be 1.5, got %f", stats.AverageUseCount)
	}
}

func TestRecentStatsEmptyList(t *testing.T) {
	r := NewRecent()

	stats := r.GetStats("nonexistent")

	if stats.ListName != "nonexistent" {
		t.Error("Expected list name to be set even for nonexistent list")
	}

	if stats.TotalItems != 0 {
		t.Error("Expected 0 total items for nonexistent list")
	}
}

func TestRecentConcurrentAccess(t *testing.T) {
	r := NewRecent()

	done := make(chan bool, 10)

	// Concurrent adds
	for i := 0; i < 10; i++ {
		go func(n int) {
			r.Add("clusters", "cluster-"+string(rune(n+'0')))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	values := r.Get("clusters")
	if len(values) == 0 {
		t.Error("Expected some values after concurrent access")
	}
}
