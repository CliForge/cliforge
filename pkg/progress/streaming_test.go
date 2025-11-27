package progress

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewSSEClient(t *testing.T) {
	config := &StreamConfig{
		Type:     StreamTypeSSE,
		Endpoint: "http://example.com/stream",
	}

	client := NewSSEClient(config)
	if client == nil {
		t.Fatal("NewSSEClient() returned nil")
	}

	if client.config.Type != StreamTypeSSE {
		t.Errorf("Expected type SSE, got %s", client.config.Type)
	}

	if client.handlers == nil {
		t.Error("handlers map should be initialized")
	}

	if client.events == nil {
		t.Error("events channel should be initialized")
	}

	if client.errors == nil {
		t.Error("errors channel should be initialized")
	}
}

func TestSSEClient_Subscribe(t *testing.T) {
	client := NewSSEClient(nil)

	handler := func(event *Event) error {
		return nil
	}

	err := client.Subscribe("test", handler)
	if err != nil {
		t.Errorf("Subscribe() error = %v", err)
	}

	if len(client.handlers["test"]) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(client.handlers["test"]))
	}
}

func TestSSEClient_Unsubscribe(t *testing.T) {
	client := NewSSEClient(nil)

	handler := func(event *Event) error { return nil }
	_ = client.Subscribe("test", handler)

	err := client.Unsubscribe("test")
	if err != nil {
		t.Errorf("Unsubscribe() error = %v", err)
	}

	if len(client.handlers["test"]) != 0 {
		t.Errorf("Expected 0 handlers, got %d", len(client.handlers["test"]))
	}
}

func TestSSEClient_Connect(t *testing.T) {
	// Create a test SSE server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("Expected http.ResponseWriter to be an http.Flusher")
		}

		// Send a test event
		_, _ = w.Write([]byte("event: test\n"))
		_, _ = w.Write([]byte("data: hello\n\n"))
		flusher.Flush()

		// Keep connection open briefly
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	config := &StreamConfig{
		Type:                 StreamTypeSSE,
		Endpoint:             server.URL,
		Events:               []string{"test"},
		Timeout:              5 * time.Second,
		MaxReconnectAttempts: 1,
	}

	client := NewSSEClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	// Wait briefly for connection
	time.Sleep(200 * time.Millisecond)

	// Check if we received the event
	select {
	case event := <-client.Events():
		if event.Type != "test" {
			t.Errorf("Expected event type 'test', got %s", event.Type)
		}
		if event.Data != "hello" {
			t.Errorf("Expected data 'hello', got %s", event.Data)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for event")
	}

	_ = client.Close()
}

func TestSSEClient_readEvents(t *testing.T) {
	sseData := `event: message
data: line1
data: line2

event: status
data: {"state":"running"}
id: 123

`

	client := NewSSEClient(nil)
	ctx := context.Background()

	reader := strings.NewReader(sseData)

	eventReceived := make(chan *Event, 2)
	_ = client.Subscribe("message", func(e *Event) error {
		eventReceived <- e
		return nil
	})
	_ = client.Subscribe("status", func(e *Event) error {
		eventReceived <- e
		return nil
	})

	go func() { _ = client.readEvents(ctx, reader) }()

	// Check first event
	select {
	case event := <-eventReceived:
		if event.Type != "message" {
			t.Errorf("Expected type 'message', got %s", event.Type)
		}
		if event.Data != "line1\nline2" {
			t.Errorf("Expected data 'line1\\nline2', got %s", event.Data)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for first event")
	}

	// Check second event
	select {
	case event := <-eventReceived:
		if event.Type != "status" {
			t.Errorf("Expected type 'status', got %s", event.Type)
		}
		if event.ID != "123" {
			t.Errorf("Expected ID '123', got %s", event.ID)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for second event")
	}
}

func TestNewWebSocketClient(t *testing.T) {
	config := &StreamConfig{
		Type:     StreamTypeWebSocket,
		Endpoint: "ws://example.com/ws",
	}

	client := NewWebSocketClient(config)
	if client == nil {
		t.Fatal("NewWebSocketClient() returned nil")
	}

	if client.config.Type != StreamTypeWebSocket {
		t.Errorf("Expected type WebSocket, got %s", client.config.Type)
	}
}

func TestWebSocketClient_Connect(t *testing.T) {
	// Create test WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("Upgrade error: %v", err)
			return
		}
		defer func() { _ = conn.Close() }()

		// Send a test message
		msg := map[string]string{"type": "test", "data": "hello"}
		data, err := json.Marshal(msg)
		if err != nil {
			t.Errorf("Marshal error: %v", err)
			return
		}
		_ = conn.WriteMessage(websocket.TextMessage, data)

		// Keep connection open briefly
		time.Sleep(200 * time.Millisecond)
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	config := &StreamConfig{
		Type:                 StreamTypeWebSocket,
		Endpoint:             wsURL,
		Timeout:              5 * time.Second,
		MaxReconnectAttempts: 1,
	}

	client := NewWebSocketClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	// Wait for connection
	time.Sleep(300 * time.Millisecond)

	// Check if we received the message
	select {
	case event := <-client.Events():
		if event.Type != "message" {
			t.Errorf("Expected type 'message', got %s", event.Type)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for message")
	}

	_ = client.Close()
}

func TestNewPollingClient(t *testing.T) {
	config := &StreamConfig{
		Type:            StreamTypePolling,
		Endpoint:        "http://example.com/status",
		PollingInterval: 1 * time.Second,
	}

	client := NewPollingClient(config)
	if client == nil {
		t.Fatal("NewPollingClient() returned nil")
	}

	if client.config.Type != StreamTypePolling {
		t.Errorf("Expected type Polling, got %s", client.config.Type)
	}
}

func TestPollingClient_Connect(t *testing.T) {
	var pollCount atomic.Int32

	// Create test HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := pollCount.Add(1)
		response := map[string]interface{}{
			"status": "running",
			"count":  count,
		}
		data, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(data)
	}))
	defer server.Close()

	config := &StreamConfig{
		Type:            StreamTypePolling,
		Endpoint:        server.URL,
		PollingInterval: 100 * time.Millisecond,
		Timeout:         5 * time.Second,
	}

	client := NewPollingClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	// Wait for at least 2 polls
	time.Sleep(300 * time.Millisecond)

	count := pollCount.Load()
	if count < 2 {
		t.Errorf("Expected at least 2 polls, got %d", count)
	}

	// Check if we received events
	eventCount := 0
	timeout := time.After(200 * time.Millisecond)
	for {
		select {
		case <-client.Events():
			eventCount++
			if eventCount >= 2 {
				goto done
			}
		case <-timeout:
			goto done
		}
	}

done:
	if eventCount < 2 {
		t.Errorf("Expected at least 2 events, got %d", eventCount)
	}

	_ = client.Close()
}

func TestPollingClient_IsConnected(t *testing.T) {
	client := NewPollingClient(nil)

	if client.IsConnected() {
		t.Error("Client should not be connected initially")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a dummy server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client.config.Endpoint = server.URL
	_ = client.Connect(ctx)

	time.Sleep(50 * time.Millisecond)

	if !client.IsConnected() {
		t.Error("Client should be connected after Connect()")
	}

	_ = client.Close()

	time.Sleep(50 * time.Millisecond)

	if client.IsConnected() {
		t.Error("Client should not be connected after Close()")
	}
}

func TestNewStreamClient(t *testing.T) {
	tests := []struct {
		name     string
		config   *StreamConfig
		wantType string
	}{
		{
			name: "SSE client",
			config: &StreamConfig{
				Type: StreamTypeSSE,
			},
			wantType: "*progress.SSEClient",
		},
		{
			name: "WebSocket client",
			config: &StreamConfig{
				Type: StreamTypeWebSocket,
			},
			wantType: "*progress.WebSocketClient",
		},
		{
			name: "Polling client",
			config: &StreamConfig{
				Type: StreamTypePolling,
			},
			wantType: "*progress.PollingClient",
		},
		{
			name:     "nil config defaults to SSE",
			config:   nil,
			wantType: "*progress.SSEClient",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewStreamClient(tt.config)
			if got == nil {
				t.Error("NewStreamClient() returned nil")
			}
		})
	}
}

func TestDefaultStreamConfig(t *testing.T) {
	config := DefaultStreamConfig()

	if config == nil {
		t.Fatal("DefaultStreamConfig() returned nil")
	}

	if config.Type != StreamTypeSSE {
		t.Errorf("Expected default type SSE, got %s", config.Type)
	}

	if config.ReconnectInterval != 5*time.Second {
		t.Errorf("Expected reconnect interval 5s, got %v", config.ReconnectInterval)
	}

	if config.MaxReconnectAttempts != 3 {
		t.Errorf("Expected max reconnect attempts 3, got %d", config.MaxReconnectAttempts)
	}

	if config.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", config.Timeout)
	}

	if config.PollingInterval != 5*time.Second {
		t.Errorf("Expected polling interval 5s, got %v", config.PollingInterval)
	}

	if config.Headers == nil {
		t.Error("Headers should be initialized")
	}
}

func TestEvent(t *testing.T) {
	now := time.Now()
	event := &Event{
		Type:      "test",
		Data:      "test data",
		ID:        "123",
		Timestamp: now,
		Raw:       []byte("raw data"),
	}

	if event.Type != "test" {
		t.Errorf("Expected type 'test', got %s", event.Type)
	}

	if event.Data != "test data" {
		t.Errorf("Expected data 'test data', got %s", event.Data)
	}

	if event.ID != "123" {
		t.Errorf("Expected ID '123', got %s", event.ID)
	}

	if !event.Timestamp.Equal(now) {
		t.Error("Timestamp mismatch")
	}
}
