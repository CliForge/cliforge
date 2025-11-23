package progress

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// SSEClient implements a Server-Sent Events client.
type SSEClient struct {
	config   *StreamConfig
	client   *http.Client
	handlers map[string][]EventHandler
	events   chan *Event
	errors   chan error
	connected bool
	cancel   context.CancelFunc
	mu       sync.RWMutex
}

// NewSSEClient creates a new SSE client.
func NewSSEClient(config *StreamConfig) *SSEClient {
	if config == nil {
		config = DefaultStreamConfig()
		config.Type = StreamTypeSSE
	}

	return &SSEClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		handlers: make(map[string][]EventHandler),
		events:   make(chan *Event, 100),
		errors:   make(chan error, 10),
	}
}

// Connect establishes a connection to the SSE endpoint.
func (s *SSEClient) Connect(ctx context.Context) error {
	s.mu.Lock()
	if s.connected {
		s.mu.Unlock()
		return fmt.Errorf("already connected")
	}
	s.mu.Unlock()

	// Create cancellable context
	ctx, cancel := context.WithCancel(ctx)
	s.cancel = cancel

	// Start connection with retry logic
	go s.connectWithRetry(ctx)

	return nil
}

// connectWithRetry attempts to connect with automatic reconnection.
func (s *SSEClient) connectWithRetry(ctx context.Context) {
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if s.config.MaxReconnectAttempts > 0 && attempts >= s.config.MaxReconnectAttempts {
			s.errors <- fmt.Errorf("max reconnect attempts reached")
			return
		}

		err := s.doConnect(ctx)
		if err != nil {
			s.errors <- fmt.Errorf("connection error: %w", err)

			// Wait before retrying
			select {
			case <-ctx.Done():
				return
			case <-time.After(s.config.ReconnectInterval):
				attempts++
				continue
			}
		}

		// Connection closed normally, exit
		return
	}
}

// doConnect performs the actual SSE connection.
func (s *SSEClient) doConnect(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.config.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set SSE headers
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Connection", "keep-alive")

	// Add custom headers
	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	s.mu.Lock()
	s.connected = true
	s.mu.Unlock()

	// Read events from the stream
	return s.readEvents(ctx, resp.Body)
}

// readEvents reads and parses SSE events from the stream.
func (s *SSEClient) readEvents(ctx context.Context, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	var event *Event

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			s.mu.Lock()
			s.connected = false
			s.mu.Unlock()
			return ctx.Err()
		default:
		}

		line := scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if event != nil {
				event.Timestamp = time.Now()
				s.dispatchEvent(event)
				event = nil
			}
			continue
		}

		// Parse SSE field
		if !strings.Contains(line, ":") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		field := parts[0]
		value := strings.TrimSpace(parts[1])

		if event == nil {
			event = &Event{}
		}

		switch field {
		case "event":
			event.Type = value
		case "data":
			if event.Data != "" {
				event.Data += "\n"
			}
			event.Data += value
		case "id":
			event.ID = value
		}
	}

	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()

	return scanner.Err()
}

// dispatchEvent dispatches an event to handlers and the events channel.
func (s *SSEClient) dispatchEvent(event *Event) {
	// Send to events channel
	select {
	case s.events <- event:
	default:
		// Channel full, skip event
	}

	// Call registered handlers
	s.mu.RLock()
	handlers := s.handlers[event.Type]
	s.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			s.errors <- fmt.Errorf("handler error: %w", err)
		}
	}
}

// Subscribe subscribes to events with a handler.
func (s *SSEClient) Subscribe(eventType string, handler EventHandler) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.handlers[eventType] = append(s.handlers[eventType], handler)
	return nil
}

// Unsubscribe removes a subscription.
func (s *SSEClient) Unsubscribe(eventType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.handlers, eventType)
	return nil
}

// Close closes the connection.
func (s *SSEClient) Close() error {
	if s.cancel != nil {
		s.cancel()
	}

	s.mu.Lock()
	s.connected = false
	s.mu.Unlock()

	return nil
}

// IsConnected returns true if the client is connected.
func (s *SSEClient) IsConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.connected
}

// Events returns a channel of events.
func (s *SSEClient) Events() <-chan *Event {
	return s.events
}

// Errors returns a channel of errors.
func (s *SSEClient) Errors() <-chan error {
	return s.errors
}

// WebSocketClient implements a WebSocket client.
type WebSocketClient struct {
	config    *StreamConfig
	conn      *websocket.Conn
	handlers  map[string][]EventHandler
	events    chan *Event
	errors    chan error
	connected bool
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// NewWebSocketClient creates a new WebSocket client.
func NewWebSocketClient(config *StreamConfig) *WebSocketClient {
	if config == nil {
		config = DefaultStreamConfig()
		config.Type = StreamTypeWebSocket
	}

	return &WebSocketClient{
		config:   config,
		handlers: make(map[string][]EventHandler),
		events:   make(chan *Event, 100),
		errors:   make(chan error, 10),
	}
}

// Connect establishes a WebSocket connection.
func (w *WebSocketClient) Connect(ctx context.Context) error {
	w.mu.Lock()
	if w.connected {
		w.mu.Unlock()
		return fmt.Errorf("already connected")
	}
	w.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	w.cancel = cancel

	go w.connectWithRetry(ctx)

	return nil
}

// connectWithRetry attempts to connect with automatic reconnection.
func (w *WebSocketClient) connectWithRetry(ctx context.Context) {
	attempts := 0

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if w.config.MaxReconnectAttempts > 0 && attempts >= w.config.MaxReconnectAttempts {
			w.errors <- fmt.Errorf("max reconnect attempts reached")
			return
		}

		err := w.doConnect(ctx)
		if err != nil {
			w.errors <- fmt.Errorf("connection error: %w", err)

			select {
			case <-ctx.Done():
				return
			case <-time.After(w.config.ReconnectInterval):
				attempts++
				continue
			}
		}

		return
	}
}

// doConnect performs the actual WebSocket connection.
func (w *WebSocketClient) doConnect(ctx context.Context) error {
	headers := http.Header{}
	for key, value := range w.config.Headers {
		headers.Set(key, value)
	}

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = w.config.Timeout

	conn, _, err := dialer.DialContext(ctx, w.config.Endpoint, headers)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	w.mu.Lock()
	w.conn = conn
	w.connected = true
	w.mu.Unlock()

	// Read messages
	go w.readMessages(ctx)

	return nil
}

// readMessages reads messages from the WebSocket connection.
func (w *WebSocketClient) readMessages(ctx context.Context) {
	defer func() {
		w.mu.Lock()
		w.connected = false
		if w.conn != nil {
			w.conn.Close()
		}
		w.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		_, message, err := w.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				w.errors <- fmt.Errorf("read error: %w", err)
			}
			return
		}

		event := &Event{
			Type:      "message",
			Data:      string(message),
			Timestamp: time.Now(),
			Raw:       message,
		}

		w.dispatchEvent(event)
	}
}

// dispatchEvent dispatches an event to handlers and the events channel.
func (w *WebSocketClient) dispatchEvent(event *Event) {
	select {
	case w.events <- event:
	default:
	}

	w.mu.RLock()
	handlers := w.handlers[event.Type]
	w.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			w.errors <- fmt.Errorf("handler error: %w", err)
		}
	}
}

// Subscribe subscribes to events with a handler.
func (w *WebSocketClient) Subscribe(eventType string, handler EventHandler) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.handlers[eventType] = append(w.handlers[eventType], handler)
	return nil
}

// Unsubscribe removes a subscription.
func (w *WebSocketClient) Unsubscribe(eventType string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.handlers, eventType)
	return nil
}

// Send sends a message through the WebSocket.
func (w *WebSocketClient) Send(data []byte) error {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if !w.connected || w.conn == nil {
		return fmt.Errorf("not connected")
	}

	return w.conn.WriteMessage(websocket.TextMessage, data)
}

// Close closes the WebSocket connection.
func (w *WebSocketClient) Close() error {
	if w.cancel != nil {
		w.cancel()
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.connected = false
	if w.conn != nil {
		return w.conn.Close()
	}

	return nil
}

// IsConnected returns true if the client is connected.
func (w *WebSocketClient) IsConnected() bool {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.connected
}

// Events returns a channel of events.
func (w *WebSocketClient) Events() <-chan *Event {
	return w.events
}

// Errors returns a channel of errors.
func (w *WebSocketClient) Errors() <-chan error {
	return w.errors
}

// PollingClient implements a polling-based client.
type PollingClient struct {
	config    *StreamConfig
	client    *http.Client
	handlers  map[string][]EventHandler
	events    chan *Event
	errors    chan error
	connected bool
	cancel    context.CancelFunc
	mu        sync.RWMutex
}

// NewPollingClient creates a new polling client.
func NewPollingClient(config *StreamConfig) *PollingClient {
	if config == nil {
		config = DefaultStreamConfig()
		config.Type = StreamTypePolling
	}

	return &PollingClient{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		handlers: make(map[string][]EventHandler),
		events:   make(chan *Event, 100),
		errors:   make(chan error, 10),
	}
}

// Connect starts the polling loop.
func (p *PollingClient) Connect(ctx context.Context) error {
	p.mu.Lock()
	if p.connected {
		p.mu.Unlock()
		return fmt.Errorf("already connected")
	}
	p.mu.Unlock()

	ctx, cancel := context.WithCancel(ctx)
	p.cancel = cancel

	p.mu.Lock()
	p.connected = true
	p.mu.Unlock()

	go p.poll(ctx)

	return nil
}

// poll performs periodic polling.
func (p *PollingClient) poll(ctx context.Context) {
	interval := p.config.PollingInterval
	if interval <= 0 {
		interval = 5 * time.Second // Default to 5 seconds if not set
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.mu.Lock()
			p.connected = false
			p.mu.Unlock()
			return
		case <-ticker.C:
			if err := p.doPoll(ctx); err != nil {
				p.errors <- fmt.Errorf("poll error: %w", err)
			}
		}
	}
}

// doPoll performs a single poll request.
func (p *PollingClient) doPoll(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.config.Endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	for key, value := range p.config.Headers {
		req.Header.Set(key, value)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to poll: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	event := &Event{
		Type:      "poll",
		Data:      string(body),
		Timestamp: time.Now(),
		Raw:       body,
	}

	p.dispatchEvent(event)

	return nil
}

// dispatchEvent dispatches an event to handlers and the events channel.
func (p *PollingClient) dispatchEvent(event *Event) {
	select {
	case p.events <- event:
	default:
	}

	p.mu.RLock()
	handlers := p.handlers[event.Type]
	p.mu.RUnlock()

	for _, handler := range handlers {
		if err := handler(event); err != nil {
			p.errors <- fmt.Errorf("handler error: %w", err)
		}
	}
}

// Subscribe subscribes to events with a handler.
func (p *PollingClient) Subscribe(eventType string, handler EventHandler) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers[eventType] = append(p.handlers[eventType], handler)
	return nil
}

// Unsubscribe removes a subscription.
func (p *PollingClient) Unsubscribe(eventType string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.handlers, eventType)
	return nil
}

// Close stops the polling.
func (p *PollingClient) Close() error {
	if p.cancel != nil {
		p.cancel()
	}

	p.mu.Lock()
	p.connected = false
	p.mu.Unlock()

	return nil
}

// IsConnected returns true if the client is connected.
func (p *PollingClient) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// Events returns a channel of events.
func (p *PollingClient) Events() <-chan *Event {
	return p.events
}

// Errors returns a channel of errors.
func (p *PollingClient) Errors() <-chan error {
	return p.errors
}

// NewStreamClient creates a new stream client based on the config.
func NewStreamClient(config *StreamConfig) StreamClient {
	if config == nil {
		config = DefaultStreamConfig()
	}

	switch config.Type {
	case StreamTypeSSE:
		return NewSSEClient(config)
	case StreamTypeWebSocket:
		return NewWebSocketClient(config)
	case StreamTypePolling:
		return NewPollingClient(config)
	default:
		return NewSSEClient(config)
	}
}
