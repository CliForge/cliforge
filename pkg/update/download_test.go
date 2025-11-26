package update

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestDownloader_Download(t *testing.T) {
	testData := []byte("test binary content")
	checksum := sha256.Sum256(testData)
	checksumStr := hex.EncodeToString(checksum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	tests := []struct {
		name         string
		checksum     string
		checksumAlgo string
		wantErr      bool
	}{
		{
			name:         "valid download with checksum",
			checksum:     checksumStr,
			checksumAlgo: "sha256",
			wantErr:      false,
		},
		{
			name:         "invalid checksum",
			checksum:     "invalid_checksum",
			checksumAlgo: "sha256",
			wantErr:      true,
		},
		{
			name:         "empty checksum",
			checksum:     "",
			checksumAlgo: "sha256",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			config := &UpdateConfig{
				CurrentVersion: "1.0.0",
				CacheDir:       tmpDir,
				HTTPTimeout:    5 * time.Second,
			}

			downloader := NewDownloader(config)

			release := &ReleaseInfo{
				URL:          server.URL,
				Checksum:     tt.checksum,
				ChecksumAlgo: tt.checksumAlgo,
				Size:         int64(len(testData)),
			}

			path, err := downloader.Download(context.Background(), release, nil)

			if (err != nil) != tt.wantErr {
				t.Errorf("Downloader.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				defer os.Remove(path)

				// Verify file exists and has correct content
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("Failed to read downloaded file: %v", err)
				}

				if string(data) != string(testData) {
					t.Errorf("Downloaded content = %s, want %s", string(data), string(testData))
				}
			}
		})
	}
}

func TestDownloader_VerifyChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	testData := []byte("test content")
	testFile := tmpDir + "/test.bin"

	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum := sha256.Sum256(testData)
	validChecksum := hex.EncodeToString(checksum[:])

	config := &UpdateConfig{
		CacheDir: tmpDir,
	}
	downloader := NewDownloader(config)

	tests := []struct {
		name     string
		checksum string
		algo     string
		wantErr  bool
	}{
		{
			name:     "valid sha256",
			checksum: validChecksum,
			algo:     "sha256",
			wantErr:  false,
		},
		{
			name:     "invalid checksum",
			checksum: "invalid",
			algo:     "sha256",
			wantErr:  true,
		},
		{
			name:     "unsupported algorithm",
			checksum: validChecksum,
			algo:     "md5",
			wantErr:  true,
		},
		{
			name:     "empty checksum",
			checksum: "",
			algo:     "sha256",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := downloader.verifyChecksum(testFile, tt.checksum, tt.algo)

			if (err != nil) != tt.wantErr {
				t.Errorf("verifyChecksum() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDownloadProgress_IsComplete(t *testing.T) {
	tests := []struct {
		name            string
		bytesDownloaded int64
		totalBytes      int64
		want            bool
	}{
		{
			name:            "complete",
			bytesDownloaded: 100,
			totalBytes:      100,
			want:            true,
		},
		{
			name:            "incomplete",
			bytesDownloaded: 50,
			totalBytes:      100,
			want:            false,
		},
		{
			name:            "over-complete",
			bytesDownloaded: 150,
			totalBytes:      100,
			want:            true,
		},
		{
			name:            "zero total",
			bytesDownloaded: 50,
			totalBytes:      0,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DownloadProgress{
				BytesDownloaded: tt.bytesDownloaded,
				TotalBytes:      tt.totalBytes,
			}

			if got := p.IsComplete(); got != tt.want {
				t.Errorf("DownloadProgress.IsComplete() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 500,
			want:  "500 B",
		},
		{
			name:  "kilobytes",
			bytes: 1024,
			want:  "1.0 KiB",
		},
		{
			name:  "megabytes",
			bytes: 1024 * 1024,
			want:  "1.0 MiB",
		},
		{
			name:  "gigabytes",
			bytes: 1024 * 1024 * 1024,
			want:  "1.0 GiB",
		},
		{
			name:  "fractional",
			bytes: 1536,
			want:  "1.5 KiB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatBytes(tt.bytes); got != tt.want {
				t.Errorf("formatBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDownloader_CleanupOldDownloads(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some old files
	oldFile := tmpDir + "/old.tmp"
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("Failed to create old file: %v", err)
	}

	// Set modification time to 8 days ago
	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(oldFile, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set file time: %v", err)
	}

	// Create a new file
	newFile := tmpDir + "/new.tmp"
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatalf("Failed to create new file: %v", err)
	}

	config := &UpdateConfig{
		CacheDir: tmpDir,
	}
	downloader := NewDownloader(config)

	if err := downloader.CleanupOldDownloads(); err != nil {
		t.Fatalf("CleanupOldDownloads() error = %v", err)
	}

	// Old file should be removed
	if _, err := os.Stat(oldFile); !os.IsNotExist(err) {
		t.Error("Old file should have been removed")
	}

	// New file should still exist
	if _, err := os.Stat(newFile); err != nil {
		t.Error("New file should still exist")
	}
}

func TestDownloader_Download_HTTPError(t *testing.T) {
	tmpDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		CacheDir:       tmpDir,
		HTTPTimeout:    5 * time.Second,
	}

	downloader := NewDownloader(config)

	release := &ReleaseInfo{
		URL:          server.URL,
		Checksum:     "abc123",
		ChecksumAlgo: "sha256",
		Size:         100,
	}

	_, err := downloader.Download(context.Background(), release, nil)
	if err == nil {
		t.Error("Download() should return error for HTTP error")
	}
}

func TestDownloader_Download_ContextCanceled(t *testing.T) {
	tmpDir := t.TempDir()

	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.Write([]byte("data"))
	}))
	defer server.Close()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		CacheDir:       tmpDir,
		HTTPTimeout:    5 * time.Second,
	}

	downloader := NewDownloader(config)

	release := &ReleaseInfo{
		URL:          server.URL,
		Checksum:     "abc123",
		ChecksumAlgo: "sha256",
		Size:         100,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := downloader.Download(ctx, release, nil)
	if err == nil {
		t.Error("Download() should return error when context is canceled")
	}
}

func TestDownloader_Download_WithProgress(t *testing.T) {
	testData := []byte("test binary content for progress tracking")
	checksum := sha256.Sum256(testData)
	checksumStr := hex.EncodeToString(checksum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		CacheDir:       tmpDir,
		HTTPTimeout:    5 * time.Second,
	}

	downloader := NewDownloader(config)

	release := &ReleaseInfo{
		URL:          server.URL,
		Checksum:     checksumStr,
		ChecksumAlgo: "sha256",
		Size:         int64(len(testData)),
	}

	progressCalled := false
	callback := func(p *DownloadProgress) {
		progressCalled = true
		if p.TotalBytes != int64(len(testData)) {
			t.Errorf("Progress TotalBytes = %d, want %d", p.TotalBytes, len(testData))
		}
	}

	path, err := downloader.Download(context.Background(), release, callback)
	if err != nil {
		t.Fatalf("Download() error = %v", err)
	}
	defer os.Remove(path)

	if !progressCalled {
		t.Error("Progress callback was not called")
	}
}

func TestDownloader_DownloadWithProgress(t *testing.T) {
	testData := []byte("test binary content")
	checksum := sha256.Sum256(testData)
	checksumStr := hex.EncodeToString(checksum[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(testData)
	}))
	defer server.Close()

	tmpDir := t.TempDir()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		CacheDir:       tmpDir,
		HTTPTimeout:    5 * time.Second,
	}

	downloader := NewDownloader(config)

	release := &ReleaseInfo{
		URL:          server.URL,
		Checksum:     checksumStr,
		ChecksumAlgo: "sha256",
		Size:         int64(len(testData)),
	}

	path, err := downloader.DownloadWithProgress(context.Background(), release)
	if err != nil {
		t.Fatalf("DownloadWithProgress() error = %v", err)
	}
	defer os.Remove(path)

	// Verify file exists
	if _, err := os.Stat(path); err != nil {
		t.Errorf("Downloaded file does not exist: %v", err)
	}
}

func TestDownloader_DownloadWithProgress_Error(t *testing.T) {
	tmpDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := &UpdateConfig{
		CurrentVersion: "1.0.0",
		CacheDir:       tmpDir,
		HTTPTimeout:    5 * time.Second,
	}

	downloader := NewDownloader(config)

	release := &ReleaseInfo{
		URL:          server.URL,
		Checksum:     "abc123",
		ChecksumAlgo: "sha256",
		Size:         100,
	}

	_, err := downloader.DownloadWithProgress(context.Background(), release)
	if err == nil {
		t.Error("DownloadWithProgress() should return error on HTTP error")
	}
}

func TestDownloader_VerifyChecksum_SHA512(t *testing.T) {
	tmpDir := t.TempDir()
	testData := []byte("test content for sha512")
	testFile := tmpDir + "/test.bin"

	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum := sha512.Sum512(testData)
	validChecksum := hex.EncodeToString(checksum[:])

	config := &UpdateConfig{
		CacheDir: tmpDir,
	}
	downloader := NewDownloader(config)

	err := downloader.verifyChecksum(testFile, validChecksum, "sha512")
	if err != nil {
		t.Errorf("verifyChecksum() with sha512 error = %v", err)
	}
}

func TestDownloader_CleanupOldDownloads_NonExistentDir(t *testing.T) {
	config := &UpdateConfig{
		CacheDir: "/nonexistent/path",
	}
	downloader := NewDownloader(config)

	// Should not error on non-existent directory
	err := downloader.CleanupOldDownloads()
	if err != nil {
		t.Errorf("CleanupOldDownloads() should not error on non-existent directory, got: %v", err)
	}
}

func TestDownloader_CleanupOldDownloads_EmptyCacheDir(t *testing.T) {
	config := &UpdateConfig{
		CacheDir: "",
	}
	downloader := NewDownloader(config)

	err := downloader.CleanupOldDownloads()
	if err != nil {
		t.Errorf("CleanupOldDownloads() error = %v", err)
	}
}

func TestDownloader_CleanupOldDownloads_SkipsDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an old directory
	oldDir := tmpDir + "/olddir"
	if err := os.Mkdir(oldDir, 0755); err != nil {
		t.Fatalf("Failed to create old directory: %v", err)
	}

	oldTime := time.Now().Add(-8 * 24 * time.Hour)
	if err := os.Chtimes(oldDir, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set directory time: %v", err)
	}

	config := &UpdateConfig{
		CacheDir: tmpDir,
	}
	downloader := NewDownloader(config)

	if err := downloader.CleanupOldDownloads(); err != nil {
		t.Fatalf("CleanupOldDownloads() error = %v", err)
	}

	// Directory should still exist (not removed)
	if _, err := os.Stat(oldDir); os.IsNotExist(err) {
		t.Error("Old directory should not be removed")
	}
}
