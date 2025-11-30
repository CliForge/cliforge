package auth

import (
	"fmt"
	"io"
	"sync"

	"github.com/skratchdot/open-golang/open"
)

// BrowserOpener defines the interface for opening URLs in a browser.
type BrowserOpener interface {
	Open(url string) error
}

// SystemBrowserOpener opens URLs using the system default browser.
type SystemBrowserOpener struct{}

// Open opens a URL in the system default browser.
func (s *SystemBrowserOpener) Open(url string) error {
	return open.Run(url)
}

// MockBrowserOpener is a mock implementation for testing.
type MockBrowserOpener struct {
	mu         sync.Mutex
	OpenedURLs []string
	Err        error
}

// Open records the URL and returns the configured error.
func (m *MockBrowserOpener) Open(url string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.OpenedURLs = append(m.OpenedURLs, url)
	return m.Err
}

// GetOpenedURLs returns a copy of the opened URLs in a thread-safe manner.
func (m *MockBrowserOpener) GetOpenedURLs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	urls := make([]string, len(m.OpenedURLs))
	copy(urls, m.OpenedURLs)
	return urls
}

// OpenBrowser opens a URL using the system default browser.
// This is a convenience function using the default opener.
func OpenBrowser(url string) error {
	return open.Run(url)
}

// OpenBrowserWithFallback tries to open the browser and prints a fallback message on failure.
func OpenBrowserWithFallback(url string, writer io.Writer) error {
	_, _ = fmt.Fprintf(writer, "\nOpening browser to:\n%s\n\n", url)

	if err := OpenBrowser(url); err != nil {
		_, _ = fmt.Fprintf(writer, "Failed to open browser automatically.\n")
		_, _ = fmt.Fprintf(writer, "Please visit the URL above manually.\n")
		return err
	}

	return nil
}
