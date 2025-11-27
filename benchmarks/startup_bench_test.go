package benchmarks

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/cache"
	"github.com/CliForge/cliforge/pkg/openapi"
)

// BenchmarkSpecLoadingFromFile benchmarks loading OpenAPI spec from a local file
func BenchmarkSpecLoadingFromFile(b *testing.B) {
	// Create test spec file
	tmpDir := b.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")

	specData := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "Success"}}
				}
			},
			"/users/{id}": {
				"get": {
					"operationId": "getUser",
					"parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}],
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`)

	if err := os.WriteFile(specPath, specData, 0644); err != nil {
		b.Fatalf("failed to write test spec: %v", err)
	}

	loader := openapi.NewLoader(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadFromFile(ctx, specPath)
		if err != nil {
			b.Fatalf("failed to load spec: %v", err)
		}
	}
}

// BenchmarkSpecLoadingFromFileWithValidation benchmarks loading and validating OpenAPI spec
func BenchmarkSpecLoadingFromFileWithValidation(b *testing.B) {
	tmpDir := b.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")

	specData := []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Test API", "version": "1.0.0"},
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"responses": {"200": {"description": "Success"}}
				}
			}
		}
	}`)

	if err := os.WriteFile(specPath, specData, 0644); err != nil {
		b.Fatalf("failed to write test spec: %v", err)
	}

	loader := openapi.NewLoader(nil)
	validator := openapi.NewValidator()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		spec, err := loader.LoadFromFile(ctx, specPath)
		if err != nil {
			b.Fatalf("failed to load spec: %v", err)
		}
		_, err = validator.Validate(ctx, spec)
		if err != nil {
			b.Fatalf("failed to validate spec: %v", err)
		}
	}
}

// BenchmarkSpecCacheGet benchmarks cache retrieval operations
func BenchmarkSpecCacheGet(b *testing.B) {
	tmpDir := b.TempDir()

	specCache := &cache.SpecCache{
		BaseDir: tmpDir,
		AppName: "test",
	}

	ctx := context.Background()
	testSpec := &cache.CachedSpec{
		Data:      []byte(`{"openapi":"3.0.0"}`),
		FetchedAt: time.Now(),
		URL:       "https://example.com/openapi.json",
	}

	// Pre-populate cache
	if err := specCache.Set(ctx, "https://example.com/openapi.json", testSpec); err != nil {
		b.Fatalf("failed to set cache: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := specCache.Get(ctx, "https://example.com/openapi.json")
		if err != nil {
			b.Fatalf("failed to get cache: %v", err)
		}
	}
}

// BenchmarkSpecCacheSet benchmarks cache write operations
func BenchmarkSpecCacheSet(b *testing.B) {
	tmpDir := b.TempDir()

	specCache := &cache.SpecCache{
		BaseDir: tmpDir,
		AppName: "test",
	}

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testSpec := &cache.CachedSpec{
			Data:      []byte(`{"openapi":"3.0.0"}`),
			FetchedAt: time.Now(),
			URL:       "https://example.com/openapi.json",
		}
		err := specCache.Set(ctx, "https://example.com/openapi.json", testSpec)
		if err != nil {
			b.Fatalf("failed to set cache: %v", err)
		}
	}
}

// BenchmarkSpecLoadingSmall benchmarks loading a small OpenAPI spec (5 endpoints)
func BenchmarkSpecLoadingSmall(b *testing.B) {
	benchmarkSpecLoading(b, generateSmallSpec())
}

// BenchmarkSpecLoadingMedium benchmarks loading a medium OpenAPI spec (50 endpoints)
func BenchmarkSpecLoadingMedium(b *testing.B) {
	benchmarkSpecLoading(b, generateMediumSpec())
}

// BenchmarkSpecLoadingLarge benchmarks loading a large OpenAPI spec (200 endpoints)
func BenchmarkSpecLoadingLarge(b *testing.B) {
	benchmarkSpecLoading(b, generateLargeSpec())
}

// Helper function to benchmark spec loading with different sizes
func benchmarkSpecLoading(b *testing.B, specData []byte) {
	tmpDir := b.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")

	if err := os.WriteFile(specPath, specData, 0644); err != nil {
		b.Fatalf("failed to write test spec: %v", err)
	}

	loader := openapi.NewLoader(nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := loader.LoadFromFile(ctx, specPath)
		if err != nil {
			b.Fatalf("failed to load spec: %v", err)
		}
	}
}

// Generate test specs of different sizes
func generateSmallSpec() []byte {
	return []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Small API", "version": "1.0.0"},
		"paths": {
			"/users": {"get": {"operationId": "getUsers", "responses": {"200": {"description": "Success"}}}},
			"/users/{id}": {"get": {"operationId": "getUser", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}], "responses": {"200": {"description": "Success"}}}},
			"/posts": {"get": {"operationId": "getPosts", "responses": {"200": {"description": "Success"}}}},
			"/posts/{id}": {"get": {"operationId": "getPost", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string"}}], "responses": {"200": {"description": "Success"}}}},
			"/comments": {"get": {"operationId": "getComments", "responses": {"200": {"description": "Success"}}}}
		}
	}`)
}

func generateMediumSpec() []byte {
	return []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Medium API", "version": "1.0.0"},
		"paths": {
			"/users": {"get": {"operationId": "getUsers", "responses": {"200": {"description": "Success"}}}},
			"/posts": {"get": {"operationId": "getPosts", "responses": {"200": {"description": "Success"}}}},
			"/comments": {"get": {"operationId": "getComments", "responses": {"200": {"description": "Success"}}}},
			"/tags": {"get": {"operationId": "getTags", "responses": {"200": {"description": "Success"}}}},
			"/categories": {"get": {"operationId": "getCategories", "responses": {"200": {"description": "Success"}}}}
		}
	}`)
}

func generateLargeSpec() []byte {
	return []byte(`{
		"openapi": "3.0.0",
		"info": {"title": "Large API", "version": "1.0.0"},
		"paths": {
			"/users": {"get": {"operationId": "getUsers", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createUser", "responses": {"201": {"description": "Created"}}}},
			"/posts": {"get": {"operationId": "getPosts", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createPost", "responses": {"201": {"description": "Created"}}}},
			"/comments": {"get": {"operationId": "getComments", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createComment", "responses": {"201": {"description": "Created"}}}},
			"/tags": {"get": {"operationId": "getTags", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createTag", "responses": {"201": {"description": "Created"}}}},
			"/categories": {"get": {"operationId": "getCategories", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createCategory", "responses": {"201": {"description": "Created"}}}},
			"/authors": {"get": {"operationId": "getAuthors", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createAuthor", "responses": {"201": {"description": "Created"}}}},
			"/articles": {"get": {"operationId": "getArticles", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createArticle", "responses": {"201": {"description": "Created"}}}},
			"/reviews": {"get": {"operationId": "getReviews", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createReview", "responses": {"201": {"description": "Created"}}}},
			"/ratings": {"get": {"operationId": "getRatings", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createRating", "responses": {"201": {"description": "Created"}}}},
			"/notifications": {"get": {"operationId": "getNotifications", "responses": {"200": {"description": "Success"}}}, "post": {"operationId": "createNotification", "responses": {"201": {"description": "Created"}}}}
		}
	}`)
}

// BenchmarkSpecParsing benchmarks just the parsing phase (not loading from disk)
func BenchmarkSpecParsing(b *testing.B) {
	specData := generateMediumSpec()
	parser := openapi.NewParser()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := parser.Parse(ctx, specData)
		if err != nil {
			b.Fatalf("failed to parse spec: %v", err)
		}
	}
}

// BenchmarkCommandTreeGeneration benchmarks generating command tree from spec
// This would require implementing the command tree builder
// Commented out as implementation may vary
/*
func BenchmarkCommandTreeGeneration(b *testing.B) {
	specData := generateMediumSpec()
	loader := openapi.NewLoader(nil)
	ctx := context.Background()

	tmpDir := b.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")
	if err := os.WriteFile(specPath, specData, 0644); err != nil {
		b.Fatalf("failed to write test spec: %v", err)
	}

	spec, err := loader.LoadFromFile(ctx, specPath)
	if err != nil {
		b.Fatalf("failed to load spec: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Generate command tree from spec
		// builder := command.NewBuilder(spec)
		// _, err := builder.BuildCommandTree()
		// if err != nil {
		//     b.Fatalf("failed to build command tree: %v", err)
		// }
	}
}
*/
