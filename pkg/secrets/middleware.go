package secrets

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// HTTPMiddleware wraps HTTP requests and responses to mask secrets.
type HTTPMiddleware struct {
	detector *Detector
	next     http.RoundTripper
}

// NewHTTPMiddleware creates a new HTTP middleware for secret masking.
func NewHTTPMiddleware(detector *Detector, next http.RoundTripper) *HTTPMiddleware {
	if next == nil {
		next = http.DefaultTransport
	}

	return &HTTPMiddleware{
		detector: detector,
		next:     next,
	}
}

// RoundTrip implements http.RoundTripper, masking secrets in debug logging.
func (m *HTTPMiddleware) RoundTrip(req *http.Request) (*http.Response, error) {
	// Execute the request
	resp, err := m.next.RoundTrip(req)
	return resp, err
}

// MaskRequest returns a copy of the request with secrets masked for logging.
func (m *HTTPMiddleware) MaskRequest(req *http.Request) *MaskedRequest {
	if req == nil {
		return nil
	}

	masked := &MaskedRequest{
		Method: req.Method,
		URL:    req.URL.String(),
		Header: make(map[string][]string),
	}

	// Mask headers
	masked.Header = m.detector.MaskHeaders(req.Header)

	return masked
}

// MaskRequestBody masks secrets in a request body for logging.
func (m *HTTPMiddleware) MaskRequestBody(body []byte, contentType string) ([]byte, error) {
	if !m.detector.IsEnabled() {
		return body, nil
	}

	// Handle JSON content
	if strings.Contains(contentType, "application/json") {
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			// If not valid JSON, return as-is
			return body, nil
		}

		masked := m.detector.MaskJSON(data)
		return json.Marshal(masked)
	}

	// Handle plain text - mask using string patterns
	maskedStr := m.detector.MaskString(string(body))
	return []byte(maskedStr), nil
}

// MaskResponse returns a copy of the response with secrets masked for logging.
func (m *HTTPMiddleware) MaskResponse(resp *http.Response) (*MaskedResponse, error) {
	if resp == nil {
		return nil, nil
	}

	masked := &MaskedResponse{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Header:     make(map[string][]string),
	}

	// Mask headers
	masked.Header = m.detector.MaskHeaders(resp.Header)

	return masked, nil
}

// MaskResponseBody masks secrets in a response body for logging.
func (m *HTTPMiddleware) MaskResponseBody(body []byte, contentType string) ([]byte, error) {
	if !m.detector.IsEnabled() {
		return body, nil
	}

	// Handle JSON content
	if strings.Contains(contentType, "application/json") {
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			// If not valid JSON, return as-is
			return body, nil
		}

		masked := m.detector.MaskJSON(data)
		return json.Marshal(masked)
	}

	// Handle plain text - mask using string patterns
	maskedStr := m.detector.MaskString(string(body))
	return []byte(maskedStr), nil
}

// MaskedRequest represents a request with masked secrets.
type MaskedRequest struct {
	Method string
	URL    string
	Header map[string][]string
	Body   interface{} // Can be string or structured data
}

// MaskedResponse represents a response with masked secrets.
type MaskedResponse struct {
	StatusCode int
	Status     string
	Header     map[string][]string
	Body       interface{} // Can be string or structured data
}

// DebugLogger provides secret-aware debug logging for HTTP operations.
type DebugLogger struct {
	detector *Detector
	writer   io.Writer
	enabled  bool
}

// NewDebugLogger creates a new debug logger with secret masking.
func NewDebugLogger(detector *Detector, writer io.Writer) *DebugLogger {
	return &DebugLogger{
		detector: detector,
		writer:   writer,
		enabled:  true,
	}
}

// SetEnabled enables or disables debug logging.
func (l *DebugLogger) SetEnabled(enabled bool) {
	l.enabled = enabled
}

// LogRequest logs an HTTP request with secrets masked.
func (l *DebugLogger) LogRequest(req *http.Request) error {
	if !l.enabled {
		return nil
	}

	// Log request line
	fmt.Fprintf(l.writer, "DEBUG: %s %s\n", req.Method, req.URL.String())

	// Log headers with masking
	maskedHeaders := l.detector.MaskHeaders(req.Header)
	fmt.Fprintf(l.writer, "DEBUG: Headers: %v\n", formatHeaders(maskedHeaders))

	// Log body if present
	if req.Body != nil {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}

		// Restore body for actual request
		req.Body = io.NopCloser(bytes.NewReader(body))

		// Mask and log
		contentType := req.Header.Get("Content-Type")
		maskedBody, err := l.maskBody(body, contentType)
		if err != nil {
			return fmt.Errorf("failed to mask request body: %w", err)
		}

		fmt.Fprintf(l.writer, "DEBUG: Body: %s\n", maskedBody)
	}

	return nil
}

// LogResponse logs an HTTP response with secrets masked.
func (l *DebugLogger) LogResponse(resp *http.Response) error {
	if !l.enabled || resp == nil {
		return nil
	}

	// Log status line
	fmt.Fprintf(l.writer, "DEBUG: Response: %s\n", resp.Status)

	// Log headers with masking
	maskedHeaders := l.detector.MaskHeaders(resp.Header)
	fmt.Fprintf(l.writer, "DEBUG: Headers: %v\n", formatHeaders(maskedHeaders))

	// Log body if present
	if resp.Body != nil {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}

		// Restore body for consumption
		resp.Body = io.NopCloser(bytes.NewReader(body))

		// Mask and log
		contentType := resp.Header.Get("Content-Type")
		maskedBody, err := l.maskBody(body, contentType)
		if err != nil {
			return fmt.Errorf("failed to mask response body: %w", err)
		}

		fmt.Fprintf(l.writer, "DEBUG: Body: %s\n", maskedBody)
	}

	return nil
}

// maskBody masks secrets in a body based on content type.
func (l *DebugLogger) maskBody(body []byte, contentType string) (string, error) {
	if !l.detector.IsEnabled() {
		return string(body), nil
	}

	// Handle JSON content
	if strings.Contains(contentType, "application/json") {
		var data interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			// If not valid JSON, mask as plain text
			return l.detector.MaskString(string(body)), nil
		}

		masked := l.detector.MaskJSON(data)
		maskedBytes, err := json.MarshalIndent(masked, "", "  ")
		if err != nil {
			return "", fmt.Errorf("failed to marshal masked JSON: %w", err)
		}

		return string(maskedBytes), nil
	}

	// Handle plain text - mask using string patterns
	return l.detector.MaskString(string(body)), nil
}

// formatHeaders formats headers for display.
func formatHeaders(headers map[string][]string) string {
	if len(headers) == 0 {
		return "{}"
	}

	var parts []string
	for key, values := range headers {
		if len(values) == 1 {
			parts = append(parts, fmt.Sprintf("%q: %q", key, values[0]))
		} else {
			parts = append(parts, fmt.Sprintf("%q: %v", key, values))
		}
	}

	return "{" + strings.Join(parts, ", ") + "}"
}

// ResponseBodyReader wraps a response body and masks secrets when read.
type ResponseBodyReader struct {
	detector    *Detector
	contentType string
	body        io.ReadCloser
	buffer      *bytes.Buffer
	masked      bool
}

// NewResponseBodyReader creates a new response body reader with masking.
func NewResponseBodyReader(detector *Detector, contentType string, body io.ReadCloser) *ResponseBodyReader {
	return &ResponseBodyReader{
		detector:    detector,
		contentType: contentType,
		body:        body,
		buffer:      &bytes.Buffer{},
		masked:      false,
	}
}

// Read implements io.Reader.
func (r *ResponseBodyReader) Read(p []byte) (n int, err error) {
	if !r.masked {
		// Read entire body
		_, err := io.Copy(r.buffer, r.body)
		if err != nil {
			return 0, err
		}

		// Mask the content
		original := r.buffer.Bytes()
		var masked []byte

		if strings.Contains(r.contentType, "application/json") {
			var data interface{}
			if json.Unmarshal(original, &data) == nil {
				maskedData := r.detector.MaskJSON(data)
				masked, _ = json.Marshal(maskedData)
			} else {
				maskedStr := r.detector.MaskString(string(original))
				masked = []byte(maskedStr)
			}
		} else {
			maskedStr := r.detector.MaskString(string(original))
			masked = []byte(maskedStr)
		}

		// Replace buffer with masked content
		r.buffer = bytes.NewBuffer(masked)
		r.masked = true
	}

	return r.buffer.Read(p)
}

// Close implements io.Closer.
func (r *ResponseBodyReader) Close() error {
	return r.body.Close()
}
