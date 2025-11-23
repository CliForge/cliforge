package update

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

)

// Downloader handles downloading updates.
type Downloader struct {
	config     *UpdateConfig
	httpClient *http.Client
}

// NewDownloader creates a new downloader.
func NewDownloader(config *UpdateConfig) *Downloader {
	if config == nil {
		config = DefaultUpdateConfig()
	}

	return &Downloader{
		config: config,
		httpClient: &http.Client{
			Timeout: config.HTTPTimeout,
		},
	}
}

// Download downloads a release to a temporary file and verifies its checksum.
// Returns the path to the downloaded file.
func (d *Downloader) Download(ctx context.Context, release *ReleaseInfo, progressCallback ProgressCallback) (string, error) {
	// Ensure cache directory exists
	if err := os.MkdirAll(d.config.CacheDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp(d.config.CacheDir, "cliforge-update-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()
	defer tempFile.Close()

	// Download the file
	if err := d.downloadToFile(ctx, release.URL, tempFile, release.Size, progressCallback); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to download: %w", err)
	}

	// Verify checksum
	if err := d.verifyChecksum(tempPath, release.Checksum, release.ChecksumAlgo); err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("checksum verification failed: %w", err)
	}

	return tempPath, nil
}

// downloadToFile downloads content from URL to the given file.
func (d *Downloader) downloadToFile(ctx context.Context, url string, dest io.Writer, size int64, progressCallback ProgressCallback) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", fmt.Sprintf("CliForge-Updater/%s", d.config.CurrentVersion))

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Use Content-Length if size not provided
	if size == 0 {
		size = resp.ContentLength
	}

	// Copy to destination with progress tracking
	if progressCallback != nil {
		reader := newProgressReader(resp.Body, size, progressCallback)
		_, err = io.Copy(dest, reader)
		if err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
	} else {
		_, err = io.Copy(dest, resp.Body)
		if err != nil {
			return fmt.Errorf("failed to write: %w", err)
		}
	}

	return nil
}

// verifyChecksum verifies the file's checksum.
func (d *Downloader) verifyChecksum(filePath, expectedChecksum, algo string) error {
	if expectedChecksum == "" {
		return fmt.Errorf("no checksum provided")
	}

	// Default to sha256 if not specified
	if algo == "" {
		algo = "sha256"
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create hasher
	var hasher hash.Hash
	switch algo {
	case "sha256":
		hasher = sha256.New()
	case "sha512":
		hasher = sha512.New()
	default:
		return fmt.Errorf("unsupported checksum algorithm: %s", algo)
	}

	// Calculate checksum
	if _, err := io.Copy(hasher, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))

	// Compare checksums
	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// DownloadWithProgress downloads a release with progress reporting.
func (d *Downloader) DownloadWithProgress(ctx context.Context, release *ReleaseInfo) (string, error) {
	var downloadedBytes int64
	var startTime = time.Now()
	var lastPrint time.Time

	callback := func(p *DownloadProgress) {
		downloadedBytes = p.BytesDownloaded

		// Print progress every second to avoid flooding the output
		if time.Since(lastPrint) >= time.Second || p.IsComplete() {
			fmt.Printf("\rDownloading: %s / %s (%.1f%%)...",
				formatBytes(p.BytesDownloaded),
				formatBytes(p.TotalBytes),
				p.Percentage)
			lastPrint = time.Now()
		}
	}

	fmt.Println("Downloading update...")

	// Download with progress callback
	tempPath, err := d.Download(ctx, release, callback)
	if err != nil {
		fmt.Printf("\nDownload failed: %v\n", err)
		return "", err
	}

	// Mark as complete
	duration := time.Since(startTime)
	fmt.Printf("\nDownloaded %s in %s\n", formatBytes(downloadedBytes), duration.Round(time.Millisecond))

	return tempPath, nil
}

// CleanupOldDownloads removes old downloaded files from the cache.
func (d *Downloader) CleanupOldDownloads() error {
	if d.config.CacheDir == "" {
		return nil
	}

	entries, err := os.ReadDir(d.config.CacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	now := time.Now()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Remove files older than 7 days
		info, err := entry.Info()
		if err != nil {
			continue
		}

		if now.Sub(info.ModTime()) > 7*24*time.Hour {
			path := filepath.Join(d.config.CacheDir, entry.Name())
			os.Remove(path)
		}
	}

	return nil
}

// progressReader wraps an io.Reader to report download progress.
type progressReader struct {
	reader    io.Reader
	total     int64
	current   int64
	callback  ProgressCallback
	startTime time.Time
	lastSpeed int64
}

// newProgressReader creates a new progress reader.
func newProgressReader(reader io.Reader, total int64, callback ProgressCallback) io.Reader {
	return &progressReader{
		reader:    reader,
		total:     total,
		callback:  callback,
		startTime: time.Now(),
	}
}

// Read implements io.Reader.
func (pr *progressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)

	if n > 0 {
		pr.current += int64(n)

		// Calculate progress
		percentage := 0.0
		if pr.total > 0 {
			percentage = float64(pr.current) / float64(pr.total) * 100
		}

		// Calculate speed
		elapsed := time.Since(pr.startTime).Seconds()
		speed := int64(0)
		if elapsed > 0 {
			speed = int64(float64(pr.current) / elapsed)
		}
		pr.lastSpeed = speed

		// Calculate ETA
		eta := time.Duration(0)
		if speed > 0 && pr.total > 0 {
			remaining := pr.total - pr.current
			eta = time.Duration(float64(remaining) / float64(speed) * float64(time.Second))
		}

		// Call callback
		if pr.callback != nil {
			pr.callback(&DownloadProgress{
				BytesDownloaded: pr.current,
				TotalBytes:      pr.total,
				Percentage:      percentage,
				Speed:           speed,
				ETA:             eta,
			})
		}
	}

	return n, err
}

// formatBytes formats bytes as human-readable string.
func formatBytes(bytes int64) string {
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
