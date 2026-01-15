// Package ws provides WebSocket support for real-time streaming.
// It enables live process output, events, and bidirectional communication
// between the Go backend and web frontends.
package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for local development
	},
}

// MessageType identifies the type of WebSocket message.
type MessageType string

const (
	TypeProcessOutput MessageType = "process_output"
	TypeProcessStatus MessageType = "process_status"
	TypeEvent         MessageType = "event"
	TypeError         MessageType = "error"
	TypePing          MessageType = "ping"
	TypePong          MessageType = "pong"
	TypeSubscribe     MessageType = "subscribe"
	TypeUnsubscribe   MessageType = "unsubscribe"
)

// Message is the standard WebSocket message format.
type Message struct {
	Type      MessageType `json:"type"`
	Channel   string      `json:"channel,omitempty"`
	ProcessID string      `json:"processId,omitempty"`
	Data      any         `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a connected WebSocket client.
type Client struct {
	hub           *Hub
	conn          *websocket.Conn
	send          chan []byte
	subscriptions map[string]bool
	mu            sync.RWMutex
}

// Hub manages WebSocket connections and message broadcasting.
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	channels   map[string]map[*Client]bool
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		channels:   make(map[string]map[*Client]bool),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				// Remove from all channels
				for channel := range client.subscriptions {
					if clients, ok := h.channels[channel]; ok {
						delete(clients, client)
					}
				}
			}
			h.mu.Unlock()
		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Subscribe adds a client to a channel.
func (h *Hub) Subscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.channels[channel]; !ok {
		h.channels[channel] = make(map[*Client]bool)
	}
	h.channels[channel][client] = true

	client.mu.Lock()
	client.subscriptions[channel] = true
	client.mu.Unlock()
}

// Unsubscribe removes a client from a channel.
func (h *Hub) Unsubscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if clients, ok := h.channels[channel]; ok {
		delete(clients, client)
	}

	client.mu.Lock()
	delete(client.subscriptions, channel)
	client.mu.Unlock()
}

// Broadcast sends a message to all connected clients.
func (h *Hub) Broadcast(msg Message) error {
	msg.Timestamp = time.Now()
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case h.broadcast <- data:
	default:
		return fmt.Errorf("broadcast channel full")
	}
	return nil
}

// SendToChannel sends a message to all clients subscribed to a channel.
func (h *Hub) SendToChannel(channel string, msg Message) error {
	msg.Timestamp = time.Now()
	msg.Channel = channel
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	h.mu.RLock()
	clients, ok := h.channels[channel]
	h.mu.RUnlock()

	if !ok {
		return nil // No subscribers
	}

	for client := range clients {
		select {
		case client.send <- data:
		default:
			// Client buffer full, skip
		}
	}
	return nil
}

// SendProcessOutput sends process output to subscribers.
func (h *Hub) SendProcessOutput(processID string, output string) error {
	return h.SendToChannel("process:"+processID, Message{
		Type:      TypeProcessOutput,
		ProcessID: processID,
		Data:      output,
	})
}

// SendProcessStatus sends process status update to subscribers.
func (h *Hub) SendProcessStatus(processID string, status string, exitCode int) error {
	return h.SendToChannel("process:"+processID, Message{
		Type:      TypeProcessStatus,
		ProcessID: processID,
		Data: map[string]any{
			"status":   status,
			"exitCode": exitCode,
		},
	})
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HubStats contains hub statistics.
type HubStats struct {
	Clients  int `json:"clients"`
	Channels int `json:"channels"`
}

// Stats returns current hub statistics.
func (h *Hub) Stats() HubStats {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return HubStats{
		Clients:  len(h.clients),
		Channels: len(h.channels),
	}
}

// HandleWebSocket is an alias for Handler for clearer API.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	h.Handler()(w, r)
}

// Handler returns an HTTP handler for WebSocket connections.
func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Client{
			hub:           h,
			conn:          conn,
			send:          make(chan []byte, 256),
			subscriptions: make(map[string]bool),
		}

		h.register <- client

		go client.writePump()
		go client.readPump()
	}
}

// readPump handles incoming messages from the client.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(65536)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Type {
		case TypeSubscribe:
			if channel, ok := msg.Data.(string); ok {
				c.hub.Subscribe(c, channel)
			}
		case TypeUnsubscribe:
			if channel, ok := msg.Data.(string); ok {
				c.hub.Unsubscribe(c, channel)
			}
		case TypePing:
			c.send <- mustMarshal(Message{Type: TypePong, Timestamp: time.Now()})
		}
	}
}

// writePump sends messages to the client.
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Batch queued messages
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func mustMarshal(v any) []byte {
	data, _ := json.Marshal(v)
	return data
}
