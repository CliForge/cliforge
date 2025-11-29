// Package helpers provides test helper functions for integration tests.
package helpers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

// MockServer represents a mock HTTP server for testing.
type MockServer struct {
	server   *httptest.Server
	handlers map[string]http.HandlerFunc
	mu       sync.RWMutex
	requests []*RecordedRequest
}

// RecordedRequest stores details of a received request.
type RecordedRequest struct {
	Method  string
	Path    string
	Headers http.Header
	Body    []byte
	Time    time.Time
}

// NewMockServer creates a new mock HTTP server.
func NewMockServer() *MockServer {
	ms := &MockServer{
		handlers: make(map[string]http.HandlerFunc),
		requests: make([]*RecordedRequest, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", ms.handleRequest)

	ms.server = httptest.NewServer(mux)
	return ms
}

// URL returns the server URL.
func (ms *MockServer) URL() string {
	return ms.server.URL
}

// Close shuts down the server.
func (ms *MockServer) Close() {
	ms.server.Close()
}

// On registers a handler for a specific method and path.
func (ms *MockServer) On(method, path string, handler http.HandlerFunc) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	key := fmt.Sprintf("%s %s", method, path)
	ms.handlers[key] = handler
}

// OnGET is a convenience method for GET requests.
func (ms *MockServer) OnGET(path string, handler http.HandlerFunc) {
	ms.On(http.MethodGet, path, handler)
}

// OnPOST is a convenience method for POST requests.
func (ms *MockServer) OnPOST(path string, handler http.HandlerFunc) {
	ms.On(http.MethodPost, path, handler)
}

// OnDELETE is a convenience method for DELETE requests.
func (ms *MockServer) OnDELETE(path string, handler http.HandlerFunc) {
	ms.On(http.MethodDelete, path, handler)
}

// handleRequest routes incoming requests to registered handlers.
func (ms *MockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	// Record the request
	ms.recordRequest(r)

	// Find and execute handler
	ms.mu.RLock()
	key := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
	handler, exists := ms.handlers[key]
	ms.mu.RUnlock()

	if exists {
		handler(w, r)
	} else {
		http.NotFound(w, r)
	}
}

// recordRequest stores request details.
func (ms *MockServer) recordRequest(r *http.Request) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	// Read body
	var body []byte
	if r.Body != nil {
		body, _ = readBody(r)
	}

	ms.requests = append(ms.requests, &RecordedRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: r.Header.Clone(),
		Body:    body,
		Time:    time.Now(),
	})
}

// GetRequests returns all recorded requests.
func (ms *MockServer) GetRequests() []*RecordedRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return append([]*RecordedRequest{}, ms.requests...)
}

// ClearRequests clears recorded requests.
func (ms *MockServer) ClearRequests() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.requests = make([]*RecordedRequest, 0)
}

// GetRequestCount returns the number of recorded requests.
func (ms *MockServer) GetRequestCount() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.requests)
}

// JSONResponse creates a JSON response handler.
func JSONResponse(statusCode int, data interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(data)
	}
}

// ErrorResponse creates an error response handler.
func ErrorResponse(statusCode int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error": message,
		})
	}
}

// DelayedResponse wraps a handler with a delay.
func DelayedResponse(delay time.Duration, handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(delay)
		handler(w, r)
	}
}

// readBody reads request body and restores it.
func readBody(r *http.Request) ([]byte, error) {
	if r.Body == nil {
		return []byte{}, nil
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// SSEServer creates a Server-Sent Events test server.
type SSEServer struct {
	server  *httptest.Server
	events  chan SSEEvent
	clients sync.Map
}

// SSEEvent represents a server-sent event.
type SSEEvent struct {
	Event string
	Data  string
	ID    string
}

// NewSSEServer creates a new SSE server.
func NewSSEServer() *SSEServer {
	ss := &SSEServer{
		events: make(chan SSEEvent, 10),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/events", ss.handleSSE)

	ss.server = httptest.NewServer(mux)
	return ss
}

// URL returns the SSE server URL.
func (ss *SSEServer) URL() string {
	return ss.server.URL
}

// Close shuts down the SSE server.
func (ss *SSEServer) Close() {
	close(ss.events)
	ss.server.Close()
}

// SendEvent sends an event to all connected clients.
func (ss *SSEServer) SendEvent(event SSEEvent) {
	ss.events <- event
}

// handleSSE handles SSE connections.
func (ss *SSEServer) handleSSE(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	clientID := fmt.Sprintf("client-%d", time.Now().UnixNano())
	done := make(chan bool)
	ss.clients.Store(clientID, done)

	defer func() {
		ss.clients.Delete(clientID)
		close(done)
	}()

	for {
		select {
		case event := <-ss.events:
			if event.Event != "" {
				_, _ = fmt.Fprintf(w, "event: %s\n", event.Event)
			}
			if event.ID != "" {
				_, _ = fmt.Fprintf(w, "id: %s\n", event.ID)
			}
			_, _ = fmt.Fprintf(w, "data: %s\n\n", event.Data)
			flusher.Flush()

		case <-r.Context().Done():
			return

		case <-done:
			return
		}
	}
}
