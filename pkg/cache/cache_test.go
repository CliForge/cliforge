package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewSpecCache(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-new")
	defer os.RemoveAll(tmpDir)

	// Set custom cache dir for testing
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	cache, err := NewSpecCache("test-app")
	if err != nil {
		t.Fatalf("NewSpecCache() error = %v", err)
	}

	if cache == nil {
		t.Fatal("NewSpecCache() returned nil")
	}

	if cache.AppName != "test-app" {
		t.Errorf("AppName = %v, want test-app", cache.AppName)
	}

	if cache.DefaultTTL != 5*time.Minute {
		t.Errorf("DefaultTTL = %v, want 5m", cache.DefaultTTL)
	}

	// Verify directory was created
	if _, err := os.Stat(cache.BaseDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestSpecCache_SetGetCycle(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-setget")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()

	// Test with complete spec data
	spec := &CachedSpec{
		Data:         []byte(`{"openapi":"3.1.0","info":{"title":"Test API"}}`),
		ETag:         "test-etag-123",
		LastModified: "Wed, 21 Oct 2023 07:28:00 GMT",
		FetchedAt:    time.Now(),
		URL:          "https://api.example.com/v1/openapi.yaml",
		Version:      "1.0.0",
	}

	// Test Set
	err := cache.Set(ctx, spec.URL, spec)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Test Get
	retrieved, err := cache.Get(ctx, spec.URL)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	// Verify all fields
	if string(retrieved.Data) != string(spec.Data) {
		t.Errorf("Data mismatch")
	}

	if retrieved.ETag != spec.ETag {
		t.Errorf("ETag = %v, want %v", retrieved.ETag, spec.ETag)
	}

	if retrieved.LastModified != spec.LastModified {
		t.Errorf("LastModified = %v, want %v", retrieved.LastModified, spec.LastModified)
	}

	if retrieved.URL != spec.URL {
		t.Errorf("URL = %v, want %v", retrieved.URL, spec.URL)
	}

	if retrieved.Version != spec.Version {
		t.Errorf("Version = %v, want %v", retrieved.Version, spec.Version)
	}
}

func TestSpecCache_IsValidWithTTL(t *testing.T) {
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
			name: "fresh cache within TTL",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(-2 * time.Minute),
			},
			ttl:      5 * time.Minute,
			expected: true,
		},
		{
			name: "expired cache beyond TTL",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(-10 * time.Minute),
			},
			ttl:      5 * time.Minute,
			expected: false,
		},
		{
			name: "exact TTL boundary",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(-5 * time.Minute),
			},
			ttl:      5 * time.Minute,
			expected: false,
		},
		{
			name:     "nil spec is invalid",
			spec:     nil,
			ttl:      5 * time.Minute,
			expected: false,
		},
		{
			name: "zero TTL uses default",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(-3 * time.Minute),
			},
			ttl:      0,
			expected: true,
		},
		{
			name: "future FetchedAt is valid",
			spec: &CachedSpec{
				FetchedAt: time.Now().Add(1 * time.Minute),
			},
			ttl:      5 * time.Minute,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := cache.IsValid(tt.spec, tt.ttl)
			if valid != tt.expected {
				t.Errorf("IsValid() = %v, want %v", valid, tt.expected)
			}
		})
	}
}

func TestSpecCache_PruneDefaultTTL(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-prune-default")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 2 * time.Minute,
	}

	ctx := context.Background()

	// Add stale entry (older than default TTL)
	spec := &CachedSpec{
		Data:      []byte(`{"openapi":"3.0.0"}`),
		FetchedAt: time.Now().Add(-5 * time.Minute),
		URL:       "https://example.com/old.yaml",
	}
	cache.Set(ctx, spec.URL, spec)

	// Prune with zero TTL (should use default)
	pruned, err := cache.Prune(ctx, 0)
	if err != nil {
		t.Fatalf("Prune() error = %v", err)
	}

	if pruned != 1 {
		t.Errorf("Pruned count = %v, want 1", pruned)
	}
}

func TestSpecCache_CacheKeyGeneration(t *testing.T) {
	cache := &SpecCache{
		BaseDir:    "/tmp",
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	tests := []struct {
		key1 string
		key2 string
		same bool
	}{
		{
			key1: "https://api.example.com/v1/openapi.yaml",
			key2: "https://api.example.com/v1/openapi.yaml",
			same: true,
		},
		{
			key1: "https://api.example.com/v1/openapi.yaml",
			key2: "https://api.example.com/v2/openapi.yaml",
			same: false,
		},
		{
			key1: "test",
			key2: "test",
			same: true,
		},
	}

	for _, tt := range tests {
		hash1 := cache.cacheKey(tt.key1)
		hash2 := cache.cacheKey(tt.key2)

		if (hash1 == hash2) != tt.same {
			t.Errorf("cacheKey(%q, %q) same = %v, want %v", tt.key1, tt.key2, hash1 == hash2, tt.same)
		}

		// Verify hash is hex string
		if len(hash1) != 64 {
			t.Errorf("Cache key length = %d, want 64 (SHA-256 hex)", len(hash1))
		}
	}
}

func TestSpecCache_ClearWithNonJSONFiles(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-clear-mixed")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()

	// Add cache entry
	spec := &CachedSpec{
		Data:      []byte(`{"openapi":"3.0.0"}`),
		FetchedAt: time.Now(),
		URL:       "https://example.com/api.yaml",
	}
	cache.Set(ctx, spec.URL, spec)

	// Add non-JSON file
	nonJSONFile := filepath.Join(tmpDir, "readme.txt")
	os.WriteFile(nonJSONFile, []byte("test"), 0644)

	// Add subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	os.MkdirAll(subDir, 0755)

	// Clear should only remove JSON files
	err := cache.Clear(ctx)
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// Verify cache entry is gone
	_, err = cache.Get(ctx, spec.URL)
	if err != ErrCacheMiss {
		t.Errorf("Expected ErrCacheMiss after clear, got %v", err)
	}

	// Verify non-JSON file still exists
	if _, err := os.Stat(nonJSONFile); os.IsNotExist(err) {
		t.Error("Clear() should not remove non-JSON files")
	}
}

func TestCachedSpec(t *testing.T) {
	now := time.Now()
	spec := &CachedSpec{
		Data:         []byte(`{"openapi":"3.1.0"}`),
		ETag:         "abc123",
		LastModified: "Mon, 01 Jan 2024 00:00:00 GMT",
		FetchedAt:    now,
		URL:          "https://api.example.com/openapi.json",
		Version:      "2.0.0",
	}

	if string(spec.Data) != `{"openapi":"3.1.0"}` {
		t.Error("Data mismatch")
	}

	if spec.ETag != "abc123" {
		t.Errorf("ETag = %v, want abc123", spec.ETag)
	}

	if spec.LastModified != "Mon, 01 Jan 2024 00:00:00 GMT" {
		t.Errorf("LastModified = %v", spec.LastModified)
	}

	if !spec.FetchedAt.Equal(now) {
		t.Error("FetchedAt mismatch")
	}

	if spec.URL != "https://api.example.com/openapi.json" {
		t.Errorf("URL = %v", spec.URL)
	}

	if spec.Version != "2.0.0" {
		t.Errorf("Version = %v, want 2.0.0", spec.Version)
	}
}

func TestGetCacheDir_CustomXDG(t *testing.T) {
	// Save original env
	originalXDG := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_CACHE_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	// Set custom XDG_CACHE_HOME
	customCache := "/custom/cache"
	os.Setenv("XDG_CACHE_HOME", customCache)

	dir := GetCacheDir("myapp")
	expected := filepath.Join(customCache, "myapp")

	if dir != expected {
		t.Errorf("GetCacheDir() = %v, want %v", dir, expected)
	}
}

func TestGetConfigDir_CustomXDG(t *testing.T) {
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_CONFIG_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_CONFIG_HOME")
		}
	}()

	customConfig := "/custom/config"
	os.Setenv("XDG_CONFIG_HOME", customConfig)

	dir := GetConfigDir("myapp")
	expected := filepath.Join(customConfig, "myapp")

	if dir != expected {
		t.Errorf("GetConfigDir() = %v, want %v", dir, expected)
	}
}

func TestGetDataDir_CustomXDG(t *testing.T) {
	originalXDG := os.Getenv("XDG_DATA_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_DATA_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_DATA_HOME")
		}
	}()

	customData := "/custom/data"
	os.Setenv("XDG_DATA_HOME", customData)

	dir := GetDataDir("myapp")
	expected := filepath.Join(customData, "myapp")

	if dir != expected {
		t.Errorf("GetDataDir() = %v, want %v", dir, expected)
	}
}

func TestGetStateDir_CustomXDG(t *testing.T) {
	originalXDG := os.Getenv("XDG_STATE_HOME")
	defer func() {
		if originalXDG != "" {
			os.Setenv("XDG_STATE_HOME", originalXDG)
		} else {
			os.Unsetenv("XDG_STATE_HOME")
		}
	}()

	customState := "/custom/state"
	os.Setenv("XDG_STATE_HOME", customState)

	dir := GetStateDir("myapp")
	expected := filepath.Join(customState, "myapp")

	if dir != expected {
		t.Errorf("GetStateDir() = %v, want %v", dir, expected)
	}
}

func TestCacheStats(t *testing.T) {
	stats := &CacheStats{
		TotalEntries: 10,
		TotalSize:    1024,
	}

	if stats.TotalEntries != 10 {
		t.Errorf("TotalEntries = %v, want 10", stats.TotalEntries)
	}

	if stats.TotalSize != 1024 {
		t.Errorf("TotalSize = %v, want 1024", stats.TotalSize)
	}
}

func TestErrCacheMiss(t *testing.T) {
	if ErrCacheMiss == nil {
		t.Error("ErrCacheMiss should not be nil")
	}

	if ErrCacheMiss.Error() == "" {
		t.Error("ErrCacheMiss should have an error message")
	}
}

func TestSpecCache_InvalidateMissing(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-invalidate-missing")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()

	// Invalidate non-existent key should not error
	err := cache.Invalidate(ctx, "nonexistent")
	if err != nil {
		t.Errorf("Invalidate() should not error for non-existent key: %v", err)
	}
}

func TestNewSpecCache_DirectoryCreation(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-newcache-dir")
	defer os.RemoveAll(tmpDir)

	// Set custom cache dir that doesn't exist
	os.Setenv("XDG_CACHE_HOME", tmpDir)
	defer os.Unsetenv("XDG_CACHE_HOME")

	cache, err := NewSpecCache("test-app")
	if err != nil {
		t.Fatalf("NewSpecCache() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(cache.BaseDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestSpecCache_GetStatsEmptyDir(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-stats-empty")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	stats, err := cache.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats() error = %v", err)
	}

	if stats.TotalEntries != 0 {
		t.Errorf("TotalEntries = %v, want 0", stats.TotalEntries)
	}

	if stats.TotalSize != 0 {
		t.Errorf("TotalSize = %v, want 0", stats.TotalSize)
	}
}

func TestSpecCache_PruneEmptyCache(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-prune-empty")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	pruned, err := cache.Prune(ctx, 5*time.Minute)
	if err != nil {
		t.Fatalf("Prune() error = %v", err)
	}

	if pruned != 0 {
		t.Errorf("Pruned count = %v, want 0", pruned)
	}
}

func TestSpecCache_ClearEmptyCache(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "test-cache-clear-empty")
	defer os.RemoveAll(tmpDir)

	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	cache := &SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	err := cache.Clear(ctx)
	if err != nil {
		t.Errorf("Clear() error = %v", err)
	}
}

func TestSpecCache_GetStatsNonExistentDir(t *testing.T) {
	cache := &SpecCache{
		BaseDir:    "/nonexistent/path/that/does/not/exist",
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	_, err := cache.GetStats(ctx)
	if err == nil {
		t.Error("GetStats() should error for non-existent directory")
	}
}

func TestSpecCache_ClearNonExistentDir(t *testing.T) {
	cache := &SpecCache{
		BaseDir:    "/nonexistent/path/that/does/not/exist",
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	err := cache.Clear(ctx)
	if err == nil {
		t.Error("Clear() should error for non-existent directory")
	}
}

func TestSpecCache_PruneNonExistentDir(t *testing.T) {
	cache := &SpecCache{
		BaseDir:    "/nonexistent/path/that/does/not/exist",
		AppName:    "test",
		DefaultTTL: 5 * time.Minute,
	}

	ctx := context.Background()
	_, err := cache.Prune(ctx, 5*time.Minute)
	if err == nil {
		t.Error("Prune() should error for non-existent directory")
	}
}
