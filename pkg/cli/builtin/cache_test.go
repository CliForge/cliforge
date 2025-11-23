package builtin

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewCacheCommand(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &CacheOptions{
		CLIName: "testcli",
		Output:  output,
	}

	cmd := NewCacheCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "cache" {
		t.Errorf("expected Use 'cache', got %q", cmd.Use)
	}

	// Check subcommands exist
	subcommands := []string{"info", "clear"}
	for _, subcmd := range subcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == subcmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", subcmd)
		}
	}
}

func TestGetCacheInfo_EmptyCache(t *testing.T) {
	// Use a non-existent directory
	tempDir := t.TempDir()
	originalXDGCache := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tempDir)
	defer os.Setenv("XDG_CACHE_HOME", originalXDGCache)

	info, err := getCacheInfo("nonexistent")
	if err != nil {
		t.Fatalf("getCacheInfo failed: %v", err)
	}

	if info.Size != 0 {
		t.Errorf("expected size 0 for empty cache, got %d", info.Size)
	}

	if info.SpecCached {
		t.Error("expected SpecCached to be false for empty cache")
	}
}

func TestCalculateDirSize(t *testing.T) {
	tempDir := t.TempDir()

	// Create some test files
	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(tempDir, "file2.txt")

	content1 := []byte("hello world") // 11 bytes
	content2 := []byte("test")         // 4 bytes

	if err := os.WriteFile(file1, content1, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.WriteFile(file2, content2, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	size, err := calculateDirSize(tempDir)
	if err != nil {
		t.Fatalf("calculateDirSize failed: %v", err)
	}

	expectedSize := int64(15) // 11 + 4
	if size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, size)
	}
}

func TestCalculateDirSize_NestedDirs(t *testing.T) {
	tempDir := t.TempDir()

	// Create nested structure
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	file1 := filepath.Join(tempDir, "file1.txt")
	file2 := filepath.Join(subDir, "file2.txt")

	content := []byte("test") // 4 bytes each

	if err := os.WriteFile(file1, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	if err := os.WriteFile(file2, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	size, err := calculateDirSize(tempDir)
	if err != nil {
		t.Fatalf("calculateDirSize failed: %v", err)
	}

	expectedSize := int64(8) // 4 + 4
	if size != expectedSize {
		t.Errorf("expected size %d, got %d", expectedSize, size)
	}
}

func TestPrintCacheInfo(t *testing.T) {
	output := &bytes.Buffer{}

	info := &CacheInfo{
		CacheDir:   "/tmp/testcache",
		Size:       1024 * 1024, // 1 MiB
		SpecCached: true,
		SpecURL:    "https://example.com/spec.json",
		SpecAge:    "2h",
	}

	printCacheInfo(info, output)

	result := output.String()

	if !strings.Contains(result, "/tmp/testcache") {
		t.Errorf("expected cache directory in output, got: %s", result)
	}

	if !strings.Contains(result, "1.0 MiB") {
		t.Errorf("expected formatted size in output, got: %s", result)
	}

	if !strings.Contains(result, "https://example.com/spec.json") {
		t.Errorf("expected spec URL in output, got: %s", result)
	}
}

func TestPrintCacheInfo_EmptyCache(t *testing.T) {
	output := &bytes.Buffer{}

	info := &CacheInfo{
		CacheDir:   "/tmp/testcache",
		Size:       0,
		SpecCached: false,
	}

	printCacheInfo(info, output)

	result := output.String()

	if !strings.Contains(result, "Empty") {
		t.Errorf("expected 'Empty' in output for empty cache, got: %s", result)
	}

	if !strings.Contains(result, "Not cached") {
		t.Errorf("expected 'Not cached' in output, got: %s", result)
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		name     string
		bytes    int64
		expected string
	}{
		{"bytes", 500, "500 B"},
		{"kilobytes", 1024, "1.0 KiB"},
		{"megabytes", 1024 * 1024, "1.0 MiB"},
		{"gigabytes", 1024 * 1024 * 1024, "1.0 GiB"},
		{"terabytes", 1024 * 1024 * 1024 * 1024, "1.0 TiB"},
		{"large", 1536 * 1024, "1.5 MiB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatSize(tt.bytes)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestClearCache_FullCache(t *testing.T) {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "testcli")

	// Create cache directory with files
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		t.Fatalf("failed to create cache dir: %v", err)
	}

	testFile := filepath.Join(cacheDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set XDG_CACHE_HOME to our temp dir
	originalXDGCache := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tempDir)
	defer os.Setenv("XDG_CACHE_HOME", originalXDGCache)

	output := &bytes.Buffer{}
	err := clearCache("testcli", false, output)
	if err != nil {
		t.Fatalf("clearCache failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Cache cleared") {
		t.Errorf("expected 'Cache cleared' in output, got: %s", result)
	}

	// Verify cache was deleted
	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		t.Error("expected cache directory to be deleted")
	}
}

func TestClearCache_SpecOnly(t *testing.T) {
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "testcli")
	specDir := filepath.Join(cacheDir, "specs")

	// Create cache directory structure
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatalf("failed to create spec dir: %v", err)
	}

	specFile := filepath.Join(specDir, "spec.json")
	otherFile := filepath.Join(cacheDir, "other.txt")

	if err := os.WriteFile(specFile, []byte("spec"), 0644); err != nil {
		t.Fatalf("failed to create spec file: %v", err)
	}

	if err := os.WriteFile(otherFile, []byte("other"), 0644); err != nil {
		t.Fatalf("failed to create other file: %v", err)
	}

	// Set XDG_CACHE_HOME to our temp dir
	originalXDGCache := os.Getenv("XDG_CACHE_HOME")
	os.Setenv("XDG_CACHE_HOME", tempDir)
	defer os.Setenv("XDG_CACHE_HOME", originalXDGCache)

	output := &bytes.Buffer{}
	err := clearCache("testcli", true, output)
	if err != nil {
		t.Fatalf("clearCache failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Spec cache cleared") {
		t.Errorf("expected 'Spec cache cleared' in output, got: %s", result)
	}

	// Verify spec directory was deleted
	if _, err := os.Stat(specDir); !os.IsNotExist(err) {
		t.Error("expected spec directory to be deleted")
	}

	// Verify other file still exists
	if _, err := os.Stat(otherFile); err != nil {
		t.Error("expected other file to still exist")
	}
}

func TestNewCacheInfoCommand(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &CacheOptions{
		CLIName: "testcli",
		Output:  output,
	}

	cmd := newCacheInfoCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "info" {
		t.Errorf("expected Use 'info', got %q", cmd.Use)
	}
}

func TestNewCacheClearCommand(t *testing.T) {
	output := &bytes.Buffer{}
	opts := &CacheOptions{
		CLIName: "testcli",
		Output:  output,
	}

	cmd := newCacheClearCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "clear" {
		t.Errorf("expected Use 'clear', got %q", cmd.Use)
	}

	// Check that spec-only flag exists
	flag := cmd.Flags().Lookup("spec-only")
	if flag == nil {
		t.Error("expected 'spec-only' flag to exist")
	}
}
