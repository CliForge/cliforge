package openapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

// mockCache implements SpecCache for testing
type mockCache struct {
	data map[string]*CachedSpec
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]*CachedSpec),
	}
}

func (m *mockCache) Get(ctx context.Context, key string) (*CachedSpec, error) {
	if spec, ok := m.data[key]; ok {
		return spec, nil
	}
	return nil, nil
}

func (m *mockCache) Set(ctx context.Context, key string, spec *CachedSpec) error {
	m.data[key] = spec
	return nil
}

func (m *mockCache) Invalidate(ctx context.Context, key string) error {
	delete(m.data, key)
	return nil
}

func TestNewLoader(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)

	if loader == nil {
		t.Fatal("NewLoader returned nil")
	}
	if loader.Parser == nil {
		t.Error("Parser not initialized")
	}
	if loader.Cache == nil {
		t.Error("Cache not set")
	}
	if loader.HTTPClient == nil {
		t.Error("HTTPClient not initialized")
	}
	if loader.CacheTTL != 5*time.Minute {
		t.Errorf("expected CacheTTL 5m, got %v", loader.CacheTTL)
	}
}

func TestLoader_LoadFromFile(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	spec, err := loader.LoadFromFile(ctx, "testdata/simple_openapi3.json")
	if err != nil {
		t.Fatalf("failed to load from file: %v", err)
	}

	if spec == nil {
		t.Fatal("spec is nil")
	}

	info := spec.GetInfo()
	if info.Title != "Test API" {
		t.Errorf("expected title 'Test API', got '%s'", info.Title)
	}
}

func TestLoader_LoadFromFile_NotFound(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	_, err := loader.LoadFromFile(ctx, "testdata/nonexistent.json")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestLoader_LoadFromReader(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	specData := `{
		"openapi": "3.0.0",
		"info": {"title": "Reader Test", "version": "1.0.0"},
		"paths": {}
	}`

	reader := strings.NewReader(specData)
	spec, err := loader.LoadFromReader(ctx, reader)
	if err != nil {
		t.Fatalf("failed to load from reader: %v", err)
	}

	info := spec.GetInfo()
	if info.Title != "Reader Test" {
		t.Errorf("expected title 'Reader Test', got '%s'", info.Title)
	}
}

func TestLoader_LoadFromData(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	specData := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Data Test", "version": "1.0.0"},
		"paths": {}
	}`)

	spec, err := loader.LoadFromData(ctx, specData)
	if err != nil {
		t.Fatalf("failed to load from data: %v", err)
	}

	info := spec.GetInfo()
	if info.Title != "Data Test" {
		t.Errorf("expected title 'Data Test', got '%s'", info.Title)
	}
}

func TestLoader_LoadFromURL(t *testing.T) {
	specData := `{
		"openapi": "3.0.0",
		"info": {"title": "URL Test", "version": "1.0.0"},
		"paths": {}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", `"test-etag"`)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(specData))
	}))
	defer server.Close()

	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	spec, err := loader.LoadFromURL(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("failed to load from URL: %v", err)
	}

	info := spec.GetInfo()
	if info.Title != "URL Test" {
		t.Errorf("expected title 'URL Test', got '%s'", info.Title)
	}

	// Verify cache was populated
	cached, _ := cache.Get(ctx, server.URL)
	if cached == nil {
		t.Error("spec not cached")
	}
	if cached.ETag != `"test-etag"` {
		t.Errorf("expected ETag '\"test-etag\"', got '%s'", cached.ETag)
	}
}

func TestLoader_LoadFromURL_Cached(t *testing.T) {
	specData := `{
		"openapi": "3.0.0",
		"info": {"title": "Cached Test", "version": "1.0.0"},
		"paths": {}
	}`

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(specData))
	}))
	defer server.Close()

	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	// First load - should hit server
	_, err := loader.LoadFromURL(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}

	if requestCount != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}

	// Second load - should use cache
	_, err = loader.LoadFromURL(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}

	// Should still be 1 because cache is fresh
	if requestCount != 1 {
		t.Errorf("expected 1 request (cached), got %d", requestCount)
	}
}

func TestLoader_LoadFromURL_ForceRefresh(t *testing.T) {
	specData := `{
		"openapi": "3.0.0",
		"info": {"title": "Refresh Test", "version": "1.0.0"},
		"paths": {}
	}`

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(specData))
	}))
	defer server.Close()

	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	// First load
	_, err := loader.LoadFromURL(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}

	// Second load with force refresh
	_, err = loader.LoadFromURL(ctx, server.URL, &LoadOptions{ForceRefresh: true})
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests (force refresh), got %d", requestCount)
	}
}

func TestLoader_LoadFromURL_ETag(t *testing.T) {
	specData := `{
		"openapi": "3.0.0",
		"info": {"title": "ETag Test", "version": "1.0.0"},
		"paths": {}
	}`

	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		if r.Header.Get("If-None-Match") == `"test-etag"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", `"test-etag"`)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(specData))
	}))
	defer server.Close()

	cache := newMockCache()
	loader := NewLoader(cache)
	loader.CacheTTL = 1 * time.Millisecond
	ctx := context.Background()

	// First load
	_, err := loader.LoadFromURL(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("first load failed: %v", err)
	}

	// Wait for cache to expire
	time.Sleep(10 * time.Millisecond)

	// Second load - cache expired, should check ETag
	_, err = loader.LoadFromURL(ctx, server.URL, nil)
	if err != nil {
		t.Fatalf("second load failed: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests, got %d", requestCount)
	}
}

func TestLoader_LoadFromURL_InvalidURL(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	_, err := loader.LoadFromURL(ctx, "://invalid-url", nil)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestLoader_LoadFromURL_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	_, err := loader.LoadFromURL(ctx, server.URL, nil)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestLoader_RefreshCache(t *testing.T) {
	specData := `{
		"openapi": "3.0.0",
		"info": {"title": "Refresh Test", "version": "1.0.0"},
		"paths": {}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(specData))
	}))
	defer server.Close()

	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	spec, err := loader.RefreshCache(ctx, server.URL)
	if err != nil {
		t.Fatalf("RefreshCache failed: %v", err)
	}

	if spec == nil {
		t.Fatal("spec is nil")
	}
}

func TestLoader_InvalidateCache(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	// Add something to cache
	cache.Set(ctx, "test-url", &CachedSpec{
		Data:      []byte("test"),
		FetchedAt: time.Now(),
	})

	// Invalidate
	err := loader.InvalidateCache(ctx, "test-url")
	if err != nil {
		t.Fatalf("InvalidateCache failed: %v", err)
	}

	// Verify it's gone
	cached, _ := cache.Get(ctx, "test-url")
	if cached != nil {
		t.Error("cache not invalidated")
	}
}

func TestLoader_SetCacheTTL(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)

	loader.SetCacheTTL(10 * time.Minute)
	if loader.CacheTTL != 10*time.Minute {
		t.Errorf("expected CacheTTL 10m, got %v", loader.CacheTTL)
	}
}

func TestLoader_SetHTTPClient(t *testing.T) {
	cache := newMockCache()
	loader := NewLoader(cache)

	customClient := &http.Client{Timeout: 60 * time.Second}
	loader.SetHTTPClient(customClient)

	if loader.HTTPClient != customClient {
		t.Error("HTTPClient not set correctly")
	}
}

func TestLoader_LoadFromFile_Integration(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "openapi-test-*.json")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	specData := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Temp Test", "version": "1.0.0"},
		"paths": {}
	}`)

	if _, err := tmpFile.Write(specData); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cache := newMockCache()
	loader := NewLoader(cache)
	ctx := context.Background()

	spec, err := loader.LoadFromFile(ctx, tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	info := spec.GetInfo()
	if info.Title != "Temp Test" {
		t.Errorf("expected title 'Temp Test', got '%s'", info.Title)
	}
}
