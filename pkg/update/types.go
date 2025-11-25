package update

import (
	"time"
)

// ReleaseInfo contains information about an available release.
type ReleaseInfo struct {
	Version      string            `json:"version"`
	URL          string            `json:"url"`
	Checksum     string            `json:"checksum"`
	ChecksumAlgo string            `json:"checksum_algo,omitempty"` // defaults to sha256
	Size         int64             `json:"size,omitempty"`
	ReleaseDate  time.Time         `json:"release_date,omitempty"`
	Changelog    string            `json:"changelog,omitempty"`
	Critical     bool              `json:"critical,omitempty"` // if true, strongly recommend update
	Platform     map[string]string `json:"platform,omitempty"` // platform-specific URLs
}

// UpdateStatus represents the current update status.
type UpdateStatus int

const (
	// UpdateStatusUnknown indicates the update status is unknown.
	UpdateStatusUnknown UpdateStatus = iota

	// UpdateStatusUpToDate indicates the current version is up to date.
	UpdateStatusUpToDate

	// UpdateStatusAvailable indicates a new version is available.
	UpdateStatusAvailable

	// UpdateStatusDownloading indicates an update is being downloaded.
	UpdateStatusDownloading

	// UpdateStatusInstalling indicates an update is being installed.
	UpdateStatusInstalling

	// UpdateStatusFailed indicates an update failed.
	UpdateStatusFailed
)

// String returns the string representation of the update status.
func (s UpdateStatus) String() string {
	switch s {
	case UpdateStatusUnknown:
		return "unknown"
	case UpdateStatusUpToDate:
		return "up-to-date"
	case UpdateStatusAvailable:
		return "available"
	case UpdateStatusDownloading:
		return "downloading"
	case UpdateStatusInstalling:
		return "installing"
	case UpdateStatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// CheckResult represents the result of an update check.
type CheckResult struct {
	Status         UpdateStatus
	CurrentVersion *Version
	LatestVersion  *Version
	Release        *ReleaseInfo
	Error          error
	CheckedAt      time.Time
}

// UpdateAvailable returns true if an update is available.
func (r *CheckResult) UpdateAvailable() bool {
	return r.Status == UpdateStatusAvailable &&
		r.LatestVersion != nil &&
		r.CurrentVersion != nil &&
		r.LatestVersion.IsNewer(r.CurrentVersion)
}

// DownloadProgress represents download progress information.
type DownloadProgress struct {
	BytesDownloaded int64
	TotalBytes      int64
	Percentage      float64
	Speed           int64 // bytes per second
	ETA             time.Duration
}

// IsComplete returns true if the download is complete.
func (p *DownloadProgress) IsComplete() bool {
	return p.TotalBytes > 0 && p.BytesDownloaded >= p.TotalBytes
}

// ProgressCallback is called during download to report progress.
type ProgressCallback func(progress *DownloadProgress)

// UpdateConfig contains configuration for the update process.
type UpdateConfig struct {
	// CurrentVersion is the current version of the CLI.
	CurrentVersion string

	// UpdateURL is the URL to check for updates.
	UpdateURL string

	// CheckInterval is how often to check for updates.
	CheckInterval time.Duration

	// AutoUpdate enables automatic updates.
	AutoUpdate bool

	// RequireConfirmation requires user confirmation before updating.
	RequireConfirmation bool

	// AllowPrerelease allows updating to prerelease versions.
	AllowPrerelease bool

	// HTTPTimeout is the timeout for HTTP requests.
	HTTPTimeout time.Duration

	// StateDir is the directory to store update state.
	StateDir string

	// CacheDir is the directory to cache downloaded updates.
	CacheDir string
}

// DefaultUpdateConfig returns the default update configuration.
func DefaultUpdateConfig() *UpdateConfig {
	return &UpdateConfig{
		CheckInterval:       24 * time.Hour,
		AutoUpdate:          false,
		RequireConfirmation: true,
		AllowPrerelease:     false,
		HTTPTimeout:         30 * time.Second,
	}
}

// LastCheckInfo stores information about the last update check.
type LastCheckInfo struct {
	CheckedAt      time.Time     `json:"checked_at"`
	LatestVersion  string        `json:"latest_version"`
	UpdateSkipped  bool          `json:"update_skipped,omitempty"`
	SkippedVersion string        `json:"skipped_version,omitempty"`
	SkippedAt      time.Time     `json:"skipped_at,omitempty"`
}

// ShouldCheck returns true if enough time has passed since the last check.
func (l *LastCheckInfo) ShouldCheck(interval time.Duration) bool {
	if l.CheckedAt.IsZero() {
		return true
	}
	return time.Since(l.CheckedAt) >= interval
}
