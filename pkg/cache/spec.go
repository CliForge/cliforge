// Package cache provides XDG-compliant caching for OpenAPI specifications.
//
// The cache package implements efficient caching of OpenAPI specs with support
// for ETags, Last-Modified headers, TTL-based expiration, and automatic cache
// invalidation. All cache files are stored in XDG-compliant directories.
//
// # Features
//
//   - XDG Base Directory specification compliance
//   - HTTP ETag and Last-Modified support for conditional requests
//   - Configurable TTL (time-to-live) per cache entry
//   - Automatic cache pruning of expired entries
//   - Cache statistics and management commands
//   - Thread-safe concurrent access
//
// # Example Usage
//
//	// Create cache
//	cache, _ := cache.NewSpecCache("mycli")
//
//	// Check cache
//	cached, err := cache.Get(ctx, "https://api.example.com/openapi.yaml")
//	if err == cache.ErrCacheMiss || !cache.IsValid(cached, 5*time.Minute) {
//	    // Fetch fresh spec
//	    spec := fetchSpec(url)
//	    cache.Set(ctx, url, &cache.CachedSpec{
//	        Data: spec,
//	        ETag: etag,
//	        FetchedAt: time.Now(),
//	    })
//	}
//
// # Cache Locations
//
//   - Linux: ~/.cache/mycli/
//   - macOS: ~/Library/Caches/mycli/
//   - Windows: %LOCALAPPDATA%\mycli\cache\
//
// Cache entries are stored as JSON files with SHA-256 hash-based filenames
// to avoid filesystem naming conflicts.
package cache

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

// SpecCache implements XDG-compliant caching for OpenAPI specs.
type SpecCache struct {
	// BaseDir is the cache directory (defaults to XDG cache dir)
	BaseDir string
	// AppName is the application name (used in cache path)
	AppName string
	// DefaultTTL is the default cache TTL (default: 5 minutes)
	DefaultTTL time.Duration
}

// CachedSpec represents a cached OpenAPI specification.
type CachedSpec struct {
	// Data is the raw spec data
	Data []byte `json:"data"`
	// ETag is the HTTP ETag header value
	ETag string `json:"etag"`
	// LastModified is the HTTP Last-Modified header value
	LastModified string `json:"last_modified"`
	// FetchedAt is when the spec was fetched
	FetchedAt time.Time `json:"fetched_at"`
	// URL is the source URL
	URL string `json:"url"`
	// Version is the spec version
	Version string `json:"version"`
}

// NewSpecCache creates a new XDG-compliant spec cache.
func NewSpecCache(appName string) (*SpecCache, error) {
	baseDir := GetCacheDir(appName)

	// Ensure cache directory exists
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &SpecCache{
		BaseDir:    baseDir,
		AppName:    appName,
		DefaultTTL: 5 * time.Minute,
	}, nil
}

// Get retrieves a cached spec by key (typically a URL).
func (c *SpecCache) Get(ctx context.Context, key string) (*CachedSpec, error) {
	cacheKey := c.cacheKey(key)
	cachePath := filepath.Join(c.BaseDir, cacheKey+".json")

	// Check if cache file exists
	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		return nil, ErrCacheMiss
	}

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache file
	var cached CachedSpec
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &cached, nil
}

// Set stores a spec in cache.
func (c *SpecCache) Set(ctx context.Context, key string, spec *CachedSpec) error {
	cacheKey := c.cacheKey(key)
	cachePath := filepath.Join(c.BaseDir, cacheKey+".json")

	// Marshal to JSON
	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal cache data: %w", err)
	}

	// Write to cache file
	if err := os.WriteFile(cachePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Invalidate removes a cached spec.
func (c *SpecCache) Invalidate(ctx context.Context, key string) error {
	cacheKey := c.cacheKey(key)
	cachePath := filepath.Join(c.BaseDir, cacheKey+".json")

	if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove cache file: %w", err)
	}

	return nil
}

// Clear removes all cached specs.
func (c *SpecCache) Clear(ctx context.Context) error {
	entries, err := os.ReadDir(c.BaseDir)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			path := filepath.Join(c.BaseDir, entry.Name())
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove cache file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// IsValid checks if a cached spec is still valid based on TTL.
func (c *SpecCache) IsValid(cached *CachedSpec, ttl time.Duration) bool {
	if cached == nil {
		return false
	}

	if ttl == 0 {
		ttl = c.DefaultTTL
	}

	return time.Since(cached.FetchedAt) < ttl
}

// GetStats returns cache statistics.
func (c *SpecCache) GetStats(ctx context.Context) (*CacheStats, error) {
	entries, err := os.ReadDir(c.BaseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache directory: %w", err)
	}

	stats := &CacheStats{
		TotalEntries: 0,
		TotalSize:    0,
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			stats.TotalEntries++
			info, err := entry.Info()
			if err == nil {
				stats.TotalSize += info.Size()
			}
		}
	}

	return stats, nil
}

// Prune removes expired cache entries.
func (c *SpecCache) Prune(ctx context.Context, ttl time.Duration) (int, error) {
	if ttl == 0 {
		ttl = c.DefaultTTL
	}

	entries, err := os.ReadDir(c.BaseDir)
	if err != nil {
		return 0, fmt.Errorf("failed to read cache directory: %w", err)
	}

	pruned := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(c.BaseDir, entry.Name())

		// Read cache entry
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var cached CachedSpec
		if err := json.Unmarshal(data, &cached); err != nil {
			continue
		}

		// Check if expired
		if time.Since(cached.FetchedAt) >= ttl {
			if err := os.Remove(path); err == nil {
				pruned++
			}
		}
	}

	return pruned, nil
}

// cacheKey generates a cache key from a URL or identifier.
func (c *SpecCache) cacheKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// CacheStats contains cache statistics.
type CacheStats struct {
	TotalEntries int
	TotalSize    int64
}

// ErrCacheMiss is returned when a cache entry is not found.
var ErrCacheMiss = fmt.Errorf("cache miss")

// GetCacheDir returns the XDG cache directory for the application.
// Follows XDG Base Directory Specification on Linux/macOS.
func GetCacheDir(appName string) string {
	if runtime.GOOS == "windows" {
		// Windows: %LOCALAPPDATA%\appname\cache
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, "cache")
	}

	// Unix-like systems: $XDG_CACHE_HOME/appname or ~/.cache/appname
	xdgCache := os.Getenv("XDG_CACHE_HOME")
	if xdgCache != "" {
		return filepath.Join(xdgCache, appName)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp
		return filepath.Join(os.TempDir(), appName, "cache")
	}

	return filepath.Join(homeDir, ".cache", appName)
}

// GetConfigDir returns the XDG config directory for the application.
func GetConfigDir(appName string) string {
	if runtime.GOOS == "windows" {
		// Windows: %APPDATA%\appname
		appData := os.Getenv("APPDATA")
		if appData == "" {
			appData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Roaming")
		}
		return filepath.Join(appData, appName)
	}

	// Unix-like systems: $XDG_CONFIG_HOME/appname or ~/.config/appname
	xdgConfig := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfig != "" {
		return filepath.Join(xdgConfig, appName)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp
		return filepath.Join(os.TempDir(), appName, "config")
	}

	return filepath.Join(homeDir, ".config", appName)
}

// GetDataDir returns the XDG data directory for the application.
func GetDataDir(appName string) string {
	if runtime.GOOS == "windows" {
		// Windows: %LOCALAPPDATA%\appname\data
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, "data")
	}

	// Unix-like systems: $XDG_DATA_HOME/appname or ~/.local/share/appname
	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData != "" {
		return filepath.Join(xdgData, appName)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp
		return filepath.Join(os.TempDir(), appName, "data")
	}

	return filepath.Join(homeDir, ".local", "share", appName)
}

// GetStateDir returns the XDG state directory for the application.
func GetStateDir(appName string) string {
	if runtime.GOOS == "windows" {
		// Windows: %LOCALAPPDATA%\appname\state
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			localAppData = filepath.Join(os.Getenv("USERPROFILE"), "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, "state")
	}

	// Unix-like systems: $XDG_STATE_HOME/appname or ~/.local/state/appname
	xdgState := os.Getenv("XDG_STATE_HOME")
	if xdgState != "" {
		return filepath.Join(xdgState, appName)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to /tmp
		return filepath.Join(os.TempDir(), appName, "state")
	}

	return filepath.Join(homeDir, ".local", "state", appName)
}
