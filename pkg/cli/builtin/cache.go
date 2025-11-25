package builtin

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/CliForge/cliforge/pkg/cache"
	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
)

// CacheOptions configures the cache command behavior.
type CacheOptions struct {
	CLIName string
	Output  io.Writer
}

// CacheInfo contains information about the cache.
type CacheInfo struct {
	CacheDir      string    `json:"cache_dir"`
	Size          int64     `json:"size_bytes"`
	SpecCached    bool      `json:"spec_cached"`
	SpecAge       string    `json:"spec_age,omitempty"`
	SpecURL       string    `json:"spec_url,omitempty"`
	LastFetched   time.Time `json:"last_fetched,omitempty"`
}

// NewCacheCommand creates a new cache command group.
func NewCacheCommand(opts *CacheOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cache",
		Long: `Manage the CLI cache.

The cache stores:
- OpenAPI specifications
- Response data (if caching is enabled)
- Temporary files

Available subcommands:
  info   - Show cache information
  clear  - Clear the cache`,
	}

	// Add subcommands
	cmd.AddCommand(newCacheInfoCommand(opts))
	cmd.AddCommand(newCacheClearCommand(opts))

	return cmd
}

// newCacheInfoCommand creates the cache info subcommand.
func newCacheInfoCommand(opts *CacheOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Show cache information",
		Long:  "Display information about the cache including size and contents.",
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := getCacheInfo(opts.CLIName)
			if err != nil {
				return err
			}

			printCacheInfo(info, opts.Output)
			return nil
		},
	}
}

// newCacheClearCommand creates the cache clear subcommand.
func newCacheClearCommand(opts *CacheOptions) *cobra.Command {
	var specOnly bool

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear the cache",
		Long: `Clear the cache.

By default, clears the entire cache directory.
Use --spec-only to clear only the OpenAPI specification cache.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return clearCache(opts.CLIName, specOnly, opts.Output)
		},
	}

	cmd.Flags().BoolVar(&specOnly, "spec-only", false, "Clear only the OpenAPI spec cache")

	return cmd
}

// getCacheInfo retrieves cache information.
func getCacheInfo(cliName string) (*CacheInfo, error) {
	cacheDir := filepath.Join(xdg.CacheHome, cliName)

	info := &CacheInfo{
		CacheDir: cacheDir,
	}

	// Check if cache directory exists
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return info, nil
	}

	// Calculate cache size
	size, err := calculateDirSize(cacheDir)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate cache size: %w", err)
	}
	info.Size = size

	// Check spec cache
	specCache, err := cache.NewSpecCache(cliName)
	if err != nil {
		return info, nil // Continue without spec info
	}

	// Try to get spec info (we don't have the URL here, so we'll check if any spec is cached)
	ctx := context.Background()
	entries, err := listCacheEntries(ctx, specCache)
	if err == nil && len(entries) > 0 {
		info.SpecCached = true
		// Get info from first entry
		if len(entries) > 0 {
			info.SpecURL = entries[0].URL
			info.LastFetched = entries[0].FetchedAt
			age := time.Since(entries[0].FetchedAt)
			info.SpecAge = formatDuration(age)
		}
	}

	return info, nil
}

// calculateDirSize calculates the total size of a directory.
func calculateDirSize(path string) (int64, error) {
	var size int64

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// listCacheEntries lists all cache entries (helper function).
func listCacheEntries(ctx context.Context, specCache *cache.SpecCache) ([]*cache.CachedSpec, error) {
	// This is a simplified implementation
	// In a real implementation, we would have a method to list all cached specs
	return []*cache.CachedSpec{}, nil
}

// printCacheInfo prints cache information.
func printCacheInfo(info *CacheInfo, w io.Writer) {
	fmt.Fprintln(w, "Cache Information:")
	fmt.Fprintf(w, "  Directory: %s\n", info.CacheDir)

	if info.Size > 0 {
		fmt.Fprintf(w, "  Size: %s\n", formatSize(info.Size))
	} else {
		fmt.Fprintln(w, "  Size: Empty")
	}

	fmt.Fprintln(w)

	if info.SpecCached {
		fmt.Fprintln(w, "OpenAPI Spec:")
		if info.SpecURL != "" {
			fmt.Fprintf(w, "  URL: %s\n", info.SpecURL)
		}
		if !info.LastFetched.IsZero() {
			fmt.Fprintf(w, "  Last fetched: %s\n", info.LastFetched.Format(time.RFC3339))
		}
		if info.SpecAge != "" {
			fmt.Fprintf(w, "  Age: %s\n", info.SpecAge)
		}
	} else {
		fmt.Fprintln(w, "OpenAPI Spec: Not cached")
	}
}

// formatSize formats a byte size in a human-readable way.
func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// clearCache clears the cache.
func clearCache(cliName string, specOnly bool, w io.Writer) error {
	return clearCacheWithDir(xdg.CacheHome, cliName, specOnly, w)
}

// clearCacheWithDir clears the cache using a specific cache home (for testing).
func clearCacheWithDir(cacheHome, cliName string, specOnly bool, w io.Writer) error {
	cacheDir := filepath.Join(cacheHome, cliName)

	if specOnly {
		// Clear only spec cache
		specCacheDir := filepath.Join(cacheDir, "specs")
		if err := os.RemoveAll(specCacheDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clear spec cache: %w", err)
		}
		fmt.Fprintln(w, "✓ Spec cache cleared")
	} else {
		// Clear entire cache
		if err := os.RemoveAll(cacheDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to clear cache: %w", err)
		}
		fmt.Fprintln(w, "✓ Cache cleared")
	}

	return nil
}
