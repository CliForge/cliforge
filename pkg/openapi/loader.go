package openapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Loader handles loading OpenAPI specs from various sources with caching.
type Loader struct {
	// Parser is used to parse the loaded spec
	Parser *Parser
	// Cache stores loaded specs
	Cache SpecCache
	// HTTPClient for fetching remote specs
	HTTPClient *http.Client
	// CacheTTL is the cache time-to-live (default: 5 minutes)
	CacheTTL time.Duration
}

// SpecCache defines the interface for caching loaded specs.
type SpecCache interface {
	// Get retrieves a cached spec
	Get(ctx context.Context, key string) (*CachedSpec, error)
	// Set stores a spec in cache
	Set(ctx context.Context, key string, spec *CachedSpec) error
	// Invalidate removes a spec from cache
	Invalidate(ctx context.Context, key string) error
}

// CachedSpec represents a cached OpenAPI spec with metadata.
type CachedSpec struct {
	Data      []byte
	ETag      string
	FetchedAt time.Time
	URL       string
}

// NewLoader creates a new Loader instance.
func NewLoader(cache SpecCache) *Loader {
	return &Loader{
		Parser: NewParser(),
		Cache:  cache,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		CacheTTL: 5 * time.Minute,
	}
}

// LoadFromURL loads an OpenAPI spec from a URL with caching support.
func (l *Loader) LoadFromURL(ctx context.Context, specURL string, options *LoadOptions) (*ParsedSpec, error) {
	if options == nil {
		options = &LoadOptions{}
	}

	// Validate URL
	if _, err := url.Parse(specURL); err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	var data []byte
	var etag string
	var err error

	// Try to load from cache if not forcing refresh
	if !options.ForceRefresh && l.Cache != nil {
		cached, err := l.Cache.Get(ctx, specURL)
		if err == nil && cached != nil {
			// Check if cache is still valid
			if time.Since(cached.FetchedAt) < l.CacheTTL {
				// Use cached data
				data = cached.Data
				etag = cached.ETag

				// If we have an ETag, try conditional request
				if etag != "" && !options.SkipConditional {
					freshData, newETag, notModified, err := l.fetchWithETag(ctx, specURL, etag)
					if err == nil {
						if notModified {
							// Server says not modified, use cached data
							data = cached.Data
						} else {
							// Server returned new data
							data = freshData
							etag = newETag
							// Update cache
							if l.Cache != nil {
								_ = l.Cache.Set(ctx, specURL, &CachedSpec{
									Data:      data,
									ETag:      etag,
									FetchedAt: time.Now(),
									URL:       specURL,
								})
							}
						}
					}
					// If conditional request failed, fall through to use cached data
				}
			} else {
				// Cache expired, fetch fresh
				data, etag, err = l.fetchSpec(ctx, specURL)
				if err != nil {
					// If fetch fails, use stale cache if available
					if cached.Data != nil {
						data = cached.Data
						// Use stale cache data, continue without error
					} else {
						return nil, fmt.Errorf("failed to fetch spec and no cache available: %w", err)
					}
				} else {
					// Update cache with fresh data
					if l.Cache != nil {
						_ = l.Cache.Set(ctx, specURL, &CachedSpec{
							Data:      data,
							ETag:      etag,
							FetchedAt: time.Now(),
							URL:       specURL,
						})
					}
				}
			}
		} else {
			// No cache, fetch fresh
			data, etag, err = l.fetchSpec(ctx, specURL)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch spec: %w", err)
			}

			// Store in cache
			if l.Cache != nil {
				_ = l.Cache.Set(ctx, specURL, &CachedSpec{
					Data:      data,
					ETag:      etag,
					FetchedAt: time.Now(),
					URL:       specURL,
				})
			}
		}
	} else {
		// Force refresh or no cache
		data, etag, err = l.fetchSpec(ctx, specURL)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch spec: %w", err)
		}

		// Store in cache
		if l.Cache != nil {
			_ = l.Cache.Set(ctx, specURL, &CachedSpec{
				Data:      data,
				ETag:      etag,
				FetchedAt: time.Now(),
				URL:       specURL,
			})
		}
	}

	// Parse the spec
	return l.Parser.Parse(ctx, data)
}

// LoadFromFile loads an OpenAPI spec from a file.
func (l *Loader) LoadFromFile(ctx context.Context, path string) (*ParsedSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return l.Parser.Parse(ctx, data)
}

// LoadFromReader loads an OpenAPI spec from an io.Reader.
func (l *Loader) LoadFromReader(ctx context.Context, r io.Reader) (*ParsedSpec, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec: %w", err)
	}

	return l.Parser.Parse(ctx, data)
}

// LoadFromData loads an OpenAPI spec from raw bytes.
func (l *Loader) LoadFromData(ctx context.Context, data []byte) (*ParsedSpec, error) {
	return l.Parser.Parse(ctx, data)
}

// RefreshCache forces a refresh of the cached spec for a URL.
func (l *Loader) RefreshCache(ctx context.Context, specURL string) (*ParsedSpec, error) {
	return l.LoadFromURL(ctx, specURL, &LoadOptions{ForceRefresh: true})
}

// InvalidateCache removes a cached spec.
func (l *Loader) InvalidateCache(ctx context.Context, specURL string) error {
	if l.Cache == nil {
		return nil
	}
	return l.Cache.Invalidate(ctx, specURL)
}

// LoadOptions controls how specs are loaded.
type LoadOptions struct {
	// ForceRefresh bypasses cache and fetches fresh spec
	ForceRefresh bool
	// SkipConditional skips ETag-based conditional requests
	SkipConditional bool
	// Headers to include in HTTP request
	Headers map[string]string
}

// fetchSpec fetches a spec from a URL.
func (l *Loader) fetchSpec(ctx context.Context, specURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, application/yaml, application/x-yaml, text/yaml")
	req.Header.Set("User-Agent", "CliForge/0.9.0 OpenAPI Loader")

	resp, err := l.HTTPClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch spec: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	etag := resp.Header.Get("ETag")
	return data, etag, nil
}

// fetchWithETag performs a conditional GET request using ETag.
func (l *Loader) fetchWithETag(ctx context.Context, specURL, etag string) ([]byte, string, bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, specURL, nil)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json, application/yaml, application/x-yaml, text/yaml")
	req.Header.Set("User-Agent", "CliForge/0.9.0 OpenAPI Loader")
	req.Header.Set("If-None-Match", etag)

	resp, err := l.HTTPClient.Do(req)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to fetch spec: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotModified {
		// Content not modified, use cached version
		return nil, etag, true, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, "", false, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", false, fmt.Errorf("failed to read response: %w", err)
	}

	newETag := resp.Header.Get("ETag")
	return data, newETag, false, nil
}

// SetCacheTTL sets the cache time-to-live duration.
func (l *Loader) SetCacheTTL(ttl time.Duration) {
	l.CacheTTL = ttl
}

// SetHTTPClient sets a custom HTTP client.
func (l *Loader) SetHTTPClient(client *http.Client) {
	l.HTTPClient = client
}
