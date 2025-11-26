package benchmarks

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/cache"
	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/CliForge/cliforge/pkg/output"
	"github.com/CliForge/cliforge/pkg/workflow"
)

// BenchmarkMemorySpecCaching measures memory allocation for spec caching
func BenchmarkMemorySpecCaching(b *testing.B) {
	tmpDir := b.TempDir()
	specCache := &cache.SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
	}

	ctx := context.Background()
	specData := generateMediumSpec()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		testSpec := &cache.CachedSpec{
			Data:      specData,
			FetchedAt: time.Now(),
			URL:       "https://example.com/openapi.json",
		}

		err := specCache.Set(ctx, "https://example.com/openapi.json", testSpec)
		if err != nil {
			b.Fatalf("failed to set cache: %v", err)
		}
	}
}

// BenchmarkMemorySpecLoading measures memory allocation during spec loading
func BenchmarkMemorySpecLoading(b *testing.B) {
	tmpDir := b.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")
	specData := generateMediumSpec()

	if err := os.WriteFile(specPath, specData, 0644); err != nil {
		b.Fatalf("failed to write test spec: %v", err)
	}

	loader := openapi.NewLoader(nil)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadFromFile(ctx, specPath)
		if err != nil {
			b.Fatalf("failed to load spec: %v", err)
		}
	}
}

// BenchmarkMemorySpecLoadingLarge measures memory with large specs
func BenchmarkMemorySpecLoadingLarge(b *testing.B) {
	tmpDir := b.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.json")
	specData := generateLargeSpec()

	if err := os.WriteFile(specPath, specData, 0644); err != nil {
		b.Fatalf("failed to write test spec: %v", err)
	}

	loader := openapi.NewLoader(nil)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := loader.LoadFromFile(ctx, specPath)
		if err != nil {
			b.Fatalf("failed to load spec: %v", err)
		}
	}
}

// BenchmarkMemoryWorkflowExecution measures memory allocation during workflow execution
func BenchmarkMemoryWorkflowExecution(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:       "step1",
				Type:     workflow.StepTypeAPICall,
				Required: true,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL,
				},
			},
		},
	}

	executor, err := workflow.NewExecutor(wf, server.Client(), nil)
	if err != nil {
		b.Fatalf("failed to create executor: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := workflow.NewExecutionContext(nil)
		_, err := executor.Execute(ctx)
		if err != nil {
			b.Fatalf("failed to execute workflow: %v", err)
		}
	}
}

// BenchmarkMemoryWorkflowMultiStep measures memory for multi-step workflows
func BenchmarkMemoryWorkflowMultiStep(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "123",
			"data": "test data",
		})
	}))
	defer server.Close()

	wf := &workflow.Workflow{
		Steps: []*workflow.Step{
			{
				ID:       "step1",
				Type:     workflow.StepTypeAPICall,
				Required: true,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/step1",
				},
			},
			{
				ID:        "step2",
				Type:      workflow.StepTypeAPICall,
				DependsOn: []string{"step1"},
				Required:  true,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/step2",
				},
			},
			{
				ID:        "step3",
				Type:      workflow.StepTypeAPICall,
				DependsOn: []string{"step2"},
				Required:  true,
				APICall: &workflow.APICallStep{
					Method:   "GET",
					Endpoint: server.URL + "/step3",
				},
			},
		},
	}

	executor, err := workflow.NewExecutor(wf, server.Client(), nil)
	if err != nil {
		b.Fatalf("failed to create executor: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx := workflow.NewExecutionContext(nil)
		_, err := executor.Execute(ctx)
		if err != nil {
			b.Fatalf("failed to execute workflow: %v", err)
		}
	}
}

// BenchmarkMemoryOutputFormattingJSON measures memory for JSON formatting
func BenchmarkMemoryOutputFormattingJSON(b *testing.B) {
	data := generateTestData(100)
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := manager.Format(buf, data, "json")
		if err != nil {
			b.Fatalf("failed to format JSON: %v", err)
		}
	}
}

// BenchmarkMemoryOutputFormattingYAML measures memory for YAML formatting
func BenchmarkMemoryOutputFormattingYAML(b *testing.B) {
	data := generateTestData(100)
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := manager.Format(buf, data, "yaml")
		if err != nil {
			b.Fatalf("failed to format YAML: %v", err)
		}
	}
}

// BenchmarkMemoryOutputFormattingTable measures memory for table formatting
func BenchmarkMemoryOutputFormattingTable(b *testing.B) {
	data := generateTestData(100)
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		buf.Reset()
		err := manager.Format(buf, data, "table")
		if err != nil {
			b.Fatalf("failed to format table: %v", err)
		}
	}
}

// BenchmarkMemoryLargeResponseHandling measures memory for handling large API responses
func BenchmarkMemoryLargeResponseHandling(b *testing.B) {
	// Generate a large response payload
	largeData := make([]map[string]interface{}, 10000)
	for i := 0; i < 10000; i++ {
		largeData[i] = map[string]interface{}{
			"id":          i,
			"name":        fmt.Sprintf("Item %d", i),
			"description": fmt.Sprintf("This is a test description for item %d", i),
			"value":       i * 100,
			"tags":        []string{"tag1", "tag2", "tag3"},
			"metadata": map[string]interface{}{
				"created":  time.Now().Unix(),
				"modified": time.Now().Unix(),
				"version":  1,
			},
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(largeData)
	}))
	defer server.Close()

	client := server.Client()
	manager := output.NewManager()
	buf := new(bytes.Buffer)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		resp, err := client.Get(server.URL)
		if err != nil {
			b.Fatalf("failed to make request: %v", err)
		}

		var result []map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			b.Fatalf("failed to decode response: %v", err)
		}
		resp.Body.Close()

		buf.Reset()
		if err := manager.Format(buf, result, "json"); err != nil {
			b.Fatalf("failed to format output: %v", err)
		}
	}
}

// BenchmarkMemoryCacheFootprint measures overall cache memory footprint
func BenchmarkMemoryCacheFootprint(b *testing.B) {
	tmpDir := b.TempDir()
	specCache := &cache.SpecCache{
		BaseDir:    tmpDir,
		AppName:    "test",
	}

	ctx := context.Background()

	// Pre-populate cache with multiple specs
	specs := []struct {
		url  string
		data []byte
	}{
		{"https://example.com/api1.json", generateSmallSpec()},
		{"https://example.com/api2.json", generateMediumSpec()},
		{"https://example.com/api3.json", generateLargeSpec()},
	}

	for _, spec := range specs {
		testSpec := &cache.CachedSpec{
			Data:      spec.data,
			FetchedAt: time.Now(),
			URL:       spec.url,
		}
		if err := specCache.Set(ctx, spec.url, testSpec); err != nil {
			b.Fatalf("failed to set cache: %v", err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, spec := range specs {
			_, err := specCache.Get(ctx, spec.url)
			if err != nil {
				b.Fatalf("failed to get cache: %v", err)
			}
		}
	}
}

// BenchmarkMemoryJSONParsing measures memory allocation during JSON parsing
func BenchmarkMemoryJSONParsing(b *testing.B) {
	jsonData := generateMediumSpec()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result map[string]interface{}
		if err := json.Unmarshal(jsonData, &result); err != nil {
			b.Fatalf("failed to parse JSON: %v", err)
		}
	}
}

// BenchmarkMemoryStringConcatenation measures memory for string operations
func BenchmarkMemoryStringConcatenation(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var result string
		for j := 0; j < 100; j++ {
			result += fmt.Sprintf("test string %d", j)
		}
		_ = result
	}
}

// BenchmarkMemoryStringBuffer measures memory for buffer-based string building
func BenchmarkMemoryStringBuffer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var buf bytes.Buffer
		for j := 0; j < 100; j++ {
			buf.WriteString("test string ")
			buf.WriteString(fmt.Sprintf("%d", j))
		}
		_ = buf.String()
	}
}
