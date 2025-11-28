package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSpecCache_SetAndGet(t *testing.T) {
	// Create temporary cache directory
	tmpDir := filepath.Join(os.TempDir(), "cliforge-test-cache")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Ensure directory exists
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	testSpec := &CachedSpec{
		Data:         []byte(`{"openapi":"3.0.0"}`),
		ETag:         "test-etag",
		LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
		FetchedAt:    time.Now(),
		URL:          "https://example.com/openapi.json",
		Version:      "1.0.0",
	}

	// Test Set
	err := cache.Set(ctx, "https://example.com/openapi.json", testSpec)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Test Get
	cached, err := cache.Get(ctx, "https://example.com/openapi.json")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	if cached == nil {
		t.Fatal("expected cached spec, got nil")
	}

	if string(cached.Data) != string(testSpec.Data) {
		t.Errorf("expected data %s, got %s", testSpec.Data, cached.Data)
	}
	if cached.ETag != testSpec.ETag {
		t.Errorf("expected ETag %s, got %s", testSpec.ETag, cached.ETag)
	}
	if cached.URL != testSpec.URL {
		t.Errorf("expected URL %s, got %s", testSpec.URL, cached.URL)
	}
}

func TestSpecCache_CacheMiss(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "cliforge-test-cache-miss")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()

	// Try to get non-existent cache entry
	cached, err := cache.Get(ctx, "https://example.com/nonexistent.json")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss, got %v", err)
	}
	if cached != nil {
		t.Error("expected nil cached spec")
	}
}

func TestSpecCache_Invalidate(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "cliforge-test-cache-invalidate")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Ensure directory exists
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	testSpec := &CachedSpec{
		Data:      []byte(`{"openapi":"3.0.0"}`),
		FetchedAt: time.Now(),
		URL:       "https://example.com/openapi.json",
	}

	// Set cache
	err := cache.Set(ctx, "https://example.com/openapi.json", testSpec)
	if err != nil {
		t.Fatalf("failed to set cache: %v", err)
	}

	// Verify it exists
	_, err = cache.Get(ctx, "https://example.com/openapi.json")
	if err != nil {
		t.Fatalf("failed to get cache: %v", err)
	}

	// Invalidate
	err = cache.Invalidate(ctx, "https://example.com/openapi.json")
	if err != nil {
		t.Fatalf("failed to invalidate cache: %v", err)
	}

	// Verify it's gone
	_, err = cache.Get(ctx, "https://example.com/openapi.json")
	if err != ErrCacheMiss {
		t.Errorf("expected ErrCacheMiss after invalidation, got %v", err)
	}
}

func TestSpecCache_IsValid(t *testing.T) {
	cache := &SpecCache{
		DefaultTTL: 5 * time.Minute,
	}

	tests := []struct {
		name     string
		spec     *CachedSpec
		ttl      time.Duration
		expected bool
	}{
		{
			name: "Valid cache",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(-1 * time.Minute),
			},
			ttl:      5 * time.Minute,
			expected: true,
		},
		{
			name: "Expired cache",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(-10 * time.Minute),
			},
			ttl:      5 * time.Minute,
			expected: false,
		},
		{
			name: "Just expired",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(-6 * time.Minute),
			},
			ttl:      5 * time.Minute,
			expected: false,
		},
		{
			name:     "Nil spec",
			spec:     nil,
			ttl:      5 * time.Minute,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := cache.IsValid(tt.spec, tt.ttl)
			if valid != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, valid)
			}
		})
	}
}

func TestSpecCache_Clear(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "cliforge-test-cache-clear")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Ensure directory exists
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()

	// Add multiple cache entries
	for i := 0; i < 3; i++ {
		testSpec := &CachedSpec{
			Data:      []byte(`{"openapi":"3.0.0"}`),
			FetchedAt: time.Now(),
			URL:       "https://example.com/openapi" + string(rune(i)) + ".json",
		}
		err := cache.Set(ctx, testSpec.URL, testSpec)
		if err != nil {
			t.Fatalf("failed to set cache: %v", err)
		}
	}

	// Clear all
	err := cache.Clear(ctx)
	if err != nil {
		t.Fatalf("failed to clear cache: %v", err)
	}

	// Verify all are gone
	stats, err := cache.GetStats(ctx)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}
	if stats.TotalEntries != 0 {
		t.Errorf("expected 0 entries after clear, got %d", stats.TotalEntries)
	}
}

func TestSpecCache_Prune(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "cliforge-test-cache-prune")
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Ensure directory exists
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()

	// Add fresh entry
	freshSpec := &CachedSpec{
		Data:      []byte(`{"openapi":"3.0.0"}`),
		FetchedAt: time.Now(),
		URL:       "https://example.com/fresh.json",
	}
	err := cache.Set(ctx, freshSpec.URL, freshSpec)
	if err != nil {
		t.Fatalf("failed to set fresh cache: %v", err)
	}

	// Add stale entry
	staleSpec := &CachedSpec{
		Data:      []byte(`{"openapi":"3.0.0"}`),
		FetchedAt: time.Now().Add(-10 * time.Minute),
		URL:       "https://example.com/stale.json",
	}
	err = cache.Set(ctx, staleSpec.URL, staleSpec)
	if err != nil {
		t.Fatalf("failed to set stale cache: %v", err)
	}

	// Prune with 5 minute TTL
	pruned, err := cache.Prune(ctx, 5*time.Minute)
	if err != nil {
		t.Fatalf("failed to prune cache: %v", err)
	}

	if pruned != 1 {
		t.Errorf("expected 1 entry pruned, got %d", pruned)
	}

	// Verify fresh entry still exists
	_, err = cache.Get(ctx, freshSpec.URL)
	if err != nil {
		t.Errorf("fresh entry should still exist: %v", err)
	}

	// Verify stale entry is gone
	_, err = cache.Get(ctx, staleSpec.URL)
	if err != ErrCacheMiss {
		t.Errorf("stale entry should be pruned: %v", err)
	}
}

func TestGetCacheDir(t *testing.T) {
	dir := GetCacheDir("test-app")
	if dir == "" {
		t.Error("expected non-empty cache directory")
	}

	// Should contain the app name
	if !filepath.IsAbs(dir) {
		t.Error("expected absolute path")
	}
}

func TestGetConfigDir(t *testing.T) {
	dir := GetConfigDir("test-app")
	if dir == "" {
		t.Error("expected non-empty config directory")
	}

	if !filepath.IsAbs(dir) {
		t.Error("expected absolute path")
	}
}

func TestGetDataDir(t *testing.T) {
	dir := GetDataDir("test-app")
	if dir == "" {
		t.Error("expected non-empty data directory")
	}

	if !filepath.IsAbs(dir) {
		t.Error("expected absolute path")
	}
}

func TestGetStateDir(t *testing.T) {
	dir := GetStateDir("test-app")
	if dir == "" {
		t.Error("expected non-empty state directory")
	}

	if !filepath.IsAbs(dir) {
		t.Error("expected absolute path")
	}
}
