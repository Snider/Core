package collect

import (
	"sync"
	"time"
)

// Event types used by the collection subsystem.
const (
	// EventStart is emitted when a collector begins its run.
	EventStart = "start"

	// EventProgress is emitted to report incremental progress.
	EventProgress = "progress"

	// EventItem is emitted when a single item is collected.
	EventItem = "item"

	// EventError is emitted when an error occurs during collection.
	EventError = "error"

	// EventComplete is emitted when a collector finishes its run.
	EventComplete = "complete"
)

// Event represents a collection event.
type Event struct {
	// Type is one of the Event* constants.
	Type string `json:"type"`

	// Source identifies the collector that emitted the event.
	Source string `json:"source"`

	// Message is a human-readable description of the event.
	Message string `json:"message"`

	// Data carries optional event-specific payload.
	Data any `json:"data,omitempty"`

	// Time is when the event occurred.
	Time time.Time `json:"time"`
}

// EventHandler handles collection events.
type EventHandler func(Event)

// Dispatcher manages event dispatch. Handlers are registered per event type
// and are called synchronously when an event is emitted.
type Dispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

// On registers a handler for an event type. Multiple handlers can be
// registered for the same event type and will be called in order.
func (d *Dispatcher) On(eventType string, handler EventHandler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = append(d.handlers[eventType], handler)
}

// Emit dispatches an event to all registered handlers for that event type.
// If no handlers are registered for the event type, the event is silently dropped.
// The event's Time field is set to now if it is zero.
func (d *Dispatcher) Emit(event Event) {
	if event.Time.IsZero() {
		event.Time = time.Now()
	}

	d.mu.RLock()
	handlers := d.handlers[event.Type]
	d.mu.RUnlock()

	for _, h := range handlers {
		h(event)
	}
}

// EmitStart emits a start event for the given source.
func (d *Dispatcher) EmitStart(source, message string) {
	d.Emit(Event{
		Type:    EventStart,
		Source:  source,
		Message: message,
	})
}

// EmitProgress emits a progress event.
func (d *Dispatcher) EmitProgress(source, message string, data any) {
	d.Emit(Event{
		Type:    EventProgress,
		Source:  source,
		Message: message,
		Data:    data,
	})
}

// EmitItem emits an item event.
func (d *Dispatcher) EmitItem(source, message string, data any) {
	d.Emit(Event{
		Type:    EventItem,
		Source:  source,
		Message: message,
		Data:    data,
	})
}

// EmitError emits an error event.
func (d *Dispatcher) EmitError(source, message string, data any) {
	d.Emit(Event{
		Type:    EventError,
		Source:  source,
		Message: message,
		Data:    data,
	})
}

// EmitComplete emits a complete event.
func (d *Dispatcher) EmitComplete(source, message string, data any) {
	d.Emit(Event{
		Type:    EventComplete,
		Source:  source,
		Message: message,
		Data:    data,
	})
}
