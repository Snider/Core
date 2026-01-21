package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ClaudeBridge forwards messages between GUI clients and the MCP core WebSocket.
type ClaudeBridge struct {
	mcpConn     *websocket.Conn
	mcpURL      string
	clients     map[*websocket.Conn]bool
	clientsMu   sync.RWMutex
	broadcast   chan []byte
	reconnectMu sync.Mutex
}

// NewClaudeBridge creates a new bridge to the MCP core WebSocket.
func NewClaudeBridge(mcpURL string) *ClaudeBridge {
	return &ClaudeBridge{
		mcpURL:    mcpURL,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan []byte, 256),
	}
}

// Start connects to the MCP WebSocket and starts the bridge.
func (cb *ClaudeBridge) Start() {
	go cb.connectToMCP()
	go cb.broadcastLoop()
}

// connectToMCP establishes connection to the MCP core WebSocket.
func (cb *ClaudeBridge) connectToMCP() {
	for {
		cb.reconnectMu.Lock()
		if cb.mcpConn != nil {
			cb.mcpConn.Close()
		}

		log.Printf("Claude bridge connecting to MCP at %s", cb.mcpURL)
		conn, _, err := websocket.DefaultDialer.Dial(cb.mcpURL, nil)
		if err != nil {
			log.Printf("Claude bridge failed to connect to MCP: %v", err)
			cb.reconnectMu.Unlock()
			time.Sleep(5 * time.Second)
			continue
		}

		cb.mcpConn = conn
		cb.reconnectMu.Unlock()
		log.Printf("Claude bridge connected to MCP")

		// Read messages from MCP and broadcast to clients
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Claude bridge MCP read error: %v", err)
				break
			}
			cb.broadcast <- message
		}

		// Connection lost, retry
		time.Sleep(2 * time.Second)
	}
}

// broadcastLoop sends messages from MCP to all connected clients.
func (cb *ClaudeBridge) broadcastLoop() {
	for message := range cb.broadcast {
		cb.clientsMu.RLock()
		for client := range cb.clients {
			err := client.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Claude bridge client write error: %v", err)
			}
		}
		cb.clientsMu.RUnlock()
	}
}

// HandleWebSocket handles WebSocket connections from GUI clients.
func (cb *ClaudeBridge) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Claude bridge upgrade error: %v", err)
		return
	}

	cb.clientsMu.Lock()
	cb.clients[conn] = true
	cb.clientsMu.Unlock()

	// Send connected message
	connMsg, _ := json.Marshal(map[string]any{
		"type":      "system",
		"data":      "Connected to Claude bridge",
		"timestamp": time.Now(),
	})
	conn.WriteMessage(websocket.TextMessage, connMsg)

	defer func() {
		cb.clientsMu.Lock()
		delete(cb.clients, conn)
		cb.clientsMu.Unlock()
		conn.Close()
	}()

	// Read messages from client and forward to MCP
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Parse the message to check type
		var msg map[string]any
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		// Forward claude_message to MCP
		if msgType, ok := msg["type"].(string); ok && msgType == "claude_message" {
			cb.sendToMCP(message)
		}
	}
}

// sendToMCP sends a message to the MCP WebSocket.
func (cb *ClaudeBridge) sendToMCP(message []byte) {
	cb.reconnectMu.Lock()
	defer cb.reconnectMu.Unlock()

	if cb.mcpConn == nil {
		log.Printf("Claude bridge: MCP not connected")
		return
	}

	err := cb.mcpConn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		log.Printf("Claude bridge MCP write error: %v", err)
	}
}
