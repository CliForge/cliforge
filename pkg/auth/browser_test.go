package auth

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestSystemBrowserOpener_Open(t *testing.T) {
	opener := &SystemBrowserOpener{}

	// We can't actually test that a browser opens without side effects,
	// but we can test that the function doesn't panic with a valid URL
	err := opener.Open("https://example.com")
	// The error might vary by platform (no browser available in CI, etc.)
	// so we just verify it doesn't panic
	if err != nil {
		t.Logf("SystemBrowserOpener.Open() returned error (expected in CI): %v", err)
	}
}

func TestMockBrowserOpener_Success(t *testing.T) {
	mock := &MockBrowserOpener{}

	url1 := "https://example.com/auth"
	url2 := "https://example.com/login"

	// Test opening first URL
	err := mock.Open(url1)
	if err != nil {
		t.Errorf("MockBrowserOpener.Open() unexpected error: %v", err)
	}

	if len(mock.OpenedURLs) != 1 {
		t.Errorf("Expected 1 opened URL, got %d", len(mock.OpenedURLs))
	}
	if mock.OpenedURLs[0] != url1 {
		t.Errorf("Expected URL %s, got %s", url1, mock.OpenedURLs[0])
	}

	// Test opening second URL
	err = mock.Open(url2)
	if err != nil {
		t.Errorf("MockBrowserOpener.Open() unexpected error: %v", err)
	}

	if len(mock.OpenedURLs) != 2 {
		t.Errorf("Expected 2 opened URLs, got %d", len(mock.OpenedURLs))
	}
	if mock.OpenedURLs[1] != url2 {
		t.Errorf("Expected URL %s, got %s", url2, mock.OpenedURLs[1])
	}
}

func TestMockBrowserOpener_Error(t *testing.T) {
	expectedErr := errors.New("browser not available")
	mock := &MockBrowserOpener{
		Err: expectedErr,
	}

	url := "https://example.com/auth"
	err := mock.Open(url)

	if err == nil {
		t.Error("Expected error, got nil")
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}

	// URL should still be recorded even on error
	if len(mock.OpenedURLs) != 1 {
		t.Errorf("Expected 1 opened URL, got %d", len(mock.OpenedURLs))
	}
	if mock.OpenedURLs[0] != url {
		t.Errorf("Expected URL %s, got %s", url, mock.OpenedURLs[0])
	}
}

func TestOpenBrowser(t *testing.T) {
	// Similar to SystemBrowserOpener test, we can't test actual browser opening
	// but we verify the function doesn't panic
	err := OpenBrowser("https://example.com")
	if err != nil {
		t.Logf("OpenBrowser() returned error (expected in CI): %v", err)
	}
}

func TestOpenBrowserWithFallback_Success(t *testing.T) {
	var buf bytes.Buffer
	url := "https://example.com/auth"

	// This will likely fail in CI, but we test the fallback message
	_ = OpenBrowserWithFallback(url, &buf)

	output := buf.String()

	// Should contain the URL
	if !strings.Contains(output, url) {
		t.Errorf("Output should contain URL %s, got: %s", url, output)
	}

	// Should contain opening message
	if !strings.Contains(output, "Opening browser to:") {
		t.Errorf("Output should contain 'Opening browser to:', got: %s", output)
	}
}

func TestOpenBrowserWithFallback_Error(t *testing.T) {
	var buf bytes.Buffer
	// Use an invalid URL scheme that will definitely fail
	url := "invalid://this-will-fail"

	err := OpenBrowserWithFallback(url, &buf)

	// Should return an error
	if err == nil {
		t.Log("Expected error for invalid URL, got nil (browser may have handled it)")
	}

	output := buf.String()

	// Should contain the URL
	if !strings.Contains(output, url) {
		t.Errorf("Output should contain URL %s, got: %s", url, output)
	}

	// Should contain opening message
	if !strings.Contains(output, "Opening browser to:") {
		t.Errorf("Output should contain 'Opening browser to:', got: %s", output)
	}
}
