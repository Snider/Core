package display

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// EventType represents the type of event.
type EventType string

const (
	EventWindowFocus  EventType = "window.focus"
	EventWindowBlur   EventType = "window.blur"
	EventWindowMove   EventType = "window.move"
	EventWindowResize EventType = "window.resize"
	EventWindowClose  EventType = "window.close"
	EventWindowCreate EventType = "window.create"
	EventThemeChange  EventType = "theme.change"
	EventScreenChange EventType = "screen.change"
)

// Event represents a display event sent to subscribers.
type Event struct {
	Type      EventType      `json:"type"`
	Timestamp int64          `json:"timestamp"`
	Window    string         `json:"window,omitempty"`
	Data      map[string]any `json:"data,omitempty"`
}

// Subscription represents a client subscription to events.
type Subscription struct {
	ID         string      `json:"id"`
	EventTypes []EventType `json:"eventTypes"`
}

// WSEventManager manages WebSocket connections and event subscriptions.
type WSEventManager struct {
	upgrader    websocket.Upgrader
	clients     map[*websocket.Conn]*clientState
	mu          sync.RWMutex
	display     *Service
	nextSubID   int
	eventBuffer chan Event
}

// clientState tracks a client's subscriptions.
type clientState struct {
	subscriptions map[string]*Subscription
	mu            sync.RWMutex
}

// NewWSEventManager creates a new event manager.
func NewWSEventManager(display *Service) *WSEventManager {
	em := &WSEventManager{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins for local dev
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:     make(map[*websocket.Conn]*clientState),
		display:     display,
		eventBuffer: make(chan Event, 100),
	}

	// Start event broadcaster
	go em.broadcaster()

	return em
}

// broadcaster sends events to all subscribed clients.
func (em *WSEventManager) broadcaster() {
	for event := range em.eventBuffer {
		em.mu.RLock()
		for conn, state := range em.clients {
			if em.clientSubscribed(state, event.Type) {
				go em.sendEvent(conn, event)
			}
		}
		em.mu.RUnlock()
	}
}

// clientSubscribed checks if a client is subscribed to an event type.
func (em *WSEventManager) clientSubscribed(state *clientState, eventType EventType) bool {
	state.mu.RLock()
	defer state.mu.RUnlock()

	for _, sub := range state.subscriptions {
		for _, et := range sub.EventTypes {
			if et == eventType || et == "*" {
				return true
			}
		}
	}
	return false
}

// sendEvent sends an event to a specific client.
func (em *WSEventManager) sendEvent(conn *websocket.Conn, event Event) {
	em.mu.RLock()
	_, exists := em.clients[conn]
	em.mu.RUnlock()

	if !exists {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		em.removeClient(conn)
	}
}

// HandleWebSocket handles WebSocket upgrade and connection.
func (em *WSEventManager) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := em.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	em.mu.Lock()
	em.clients[conn] = &clientState{
		subscriptions: make(map[string]*Subscription),
	}
	em.mu.Unlock()

	// Handle incoming messages
	go em.handleMessages(conn)
}

// handleMessages processes incoming WebSocket messages.
func (em *WSEventManager) handleMessages(conn *websocket.Conn) {
	defer em.removeClient(conn)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var msg struct {
			Action     string      `json:"action"`
			ID         string      `json:"id,omitempty"`
			EventTypes []EventType `json:"eventTypes,omitempty"`
		}

		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}

		switch msg.Action {
		case "subscribe":
			em.subscribe(conn, msg.ID, msg.EventTypes)
		case "unsubscribe":
			em.unsubscribe(conn, msg.ID)
		case "list":
			em.listSubscriptions(conn)
		}
	}
}

// subscribe adds a subscription for a client.
func (em *WSEventManager) subscribe(conn *websocket.Conn, id string, eventTypes []EventType) {
	em.mu.RLock()
	state, exists := em.clients[conn]
	em.mu.RUnlock()

	if !exists {
		return
	}

	// Generate ID if not provided
	if id == "" {
		em.mu.Lock()
		em.nextSubID++
		id = fmt.Sprintf("sub-%d", em.nextSubID)
		em.mu.Unlock()
	}

	state.mu.Lock()
	state.subscriptions[id] = &Subscription{
		ID:         id,
		EventTypes: eventTypes,
	}
	state.mu.Unlock()

	// Send confirmation
	response := map[string]any{
		"type":       "subscribed",
		"id":         id,
		"eventTypes": eventTypes,
	}
	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

// unsubscribe removes a subscription for a client.
func (em *WSEventManager) unsubscribe(conn *websocket.Conn, id string) {
	em.mu.RLock()
	state, exists := em.clients[conn]
	em.mu.RUnlock()

	if !exists {
		return
	}

	state.mu.Lock()
	delete(state.subscriptions, id)
	state.mu.Unlock()

	// Send confirmation
	response := map[string]any{
		"type": "unsubscribed",
		"id":   id,
	}
	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

// listSubscriptions sends a list of active subscriptions to a client.
func (em *WSEventManager) listSubscriptions(conn *websocket.Conn) {
	em.mu.RLock()
	state, exists := em.clients[conn]
	em.mu.RUnlock()

	if !exists {
		return
	}

	state.mu.RLock()
	subs := make([]*Subscription, 0, len(state.subscriptions))
	for _, sub := range state.subscriptions {
		subs = append(subs, sub)
	}
	state.mu.RUnlock()

	response := map[string]any{
		"type":          "subscriptions",
		"subscriptions": subs,
	}
	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

// removeClient removes a client and its subscriptions.
func (em *WSEventManager) removeClient(conn *websocket.Conn) {
	em.mu.Lock()
	delete(em.clients, conn)
	em.mu.Unlock()
	conn.Close()
}

// Emit sends an event to all subscribed clients.
func (em *WSEventManager) Emit(event Event) {
	event.Timestamp = time.Now().UnixMilli()
	select {
	case em.eventBuffer <- event:
	default:
		// Buffer full, drop event
	}
}

// EmitWindowEvent is a helper to emit window-related events.
func (em *WSEventManager) EmitWindowEvent(eventType EventType, windowName string, data map[string]any) {
	em.Emit(Event{
		Type:   eventType,
		Window: windowName,
		Data:   data,
	})
}

// ConnectedClients returns the number of connected WebSocket clients.
func (em *WSEventManager) ConnectedClients() int {
	em.mu.RLock()
	defer em.mu.RUnlock()
	return len(em.clients)
}

// Close shuts down the event manager.
func (em *WSEventManager) Close() {
	em.mu.Lock()
	for conn := range em.clients {
		conn.Close()
	}
	em.clients = make(map[*websocket.Conn]*clientState)
	em.mu.Unlock()
	close(em.eventBuffer)
}

// SetupWindowEventListeners attaches event listeners to all windows.
func (em *WSEventManager) SetupWindowEventListeners() {
	app := application.Get()
	if app == nil {
		return
	}

	// Listen for theme changes
	app.Event.OnApplicationEvent(events.Common.ThemeChanged, func(event *application.ApplicationEvent) {
		isDark := app.Env.IsDarkMode()
		em.Emit(Event{
			Type: EventThemeChange,
			Data: map[string]any{
				"isDark": isDark,
				"theme":  map[bool]string{true: "dark", false: "light"}[isDark],
			},
		})
	})
}

// AttachWindowListeners attaches event listeners to a specific window.
func (em *WSEventManager) AttachWindowListeners(window *application.WebviewWindow) {
	if window == nil {
		return
	}

	name := window.Name()

	// Window focus
	window.OnWindowEvent(events.Common.WindowFocus, func(event *application.WindowEvent) {
		em.EmitWindowEvent(EventWindowFocus, name, nil)
	})

	// Window blur
	window.OnWindowEvent(events.Common.WindowLostFocus, func(event *application.WindowEvent) {
		em.EmitWindowEvent(EventWindowBlur, name, nil)
	})

	// Window move
	window.OnWindowEvent(events.Common.WindowDidMove, func(event *application.WindowEvent) {
		x, y := window.Position()
		em.EmitWindowEvent(EventWindowMove, name, map[string]any{
			"x": x,
			"y": y,
		})
	})

	// Window resize
	window.OnWindowEvent(events.Common.WindowDidResize, func(event *application.WindowEvent) {
		width, height := window.Size()
		em.EmitWindowEvent(EventWindowResize, name, map[string]any{
			"width":  width,
			"height": height,
		})
	})

	// Window close
	window.OnWindowEvent(events.Common.WindowClosing, func(event *application.WindowEvent) {
		em.EmitWindowEvent(EventWindowClose, name, nil)
	})
}
