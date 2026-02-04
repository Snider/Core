package collect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDispatcher_Emit_Good(t *testing.T) {
	d := NewDispatcher()

	var received Event
	d.On(EventStart, func(e Event) {
		received = e
	})

	d.Emit(Event{
		Type:    EventStart,
		Source:  "test",
		Message: "hello",
	})

	assert.Equal(t, EventStart, received.Type)
	assert.Equal(t, "test", received.Source)
	assert.Equal(t, "hello", received.Message)
	assert.False(t, received.Time.IsZero(), "Time should be set automatically")
}

func TestDispatcher_On_Good(t *testing.T) {
	d := NewDispatcher()

	var count int
	handler := func(e Event) { count++ }

	d.On(EventProgress, handler)
	d.On(EventProgress, handler)
	d.On(EventProgress, handler)

	d.Emit(Event{Type: EventProgress, Source: "test"})
	assert.Equal(t, 3, count, "All three handlers should be called")
}

func TestDispatcher_Emit_Good_NoHandlers(t *testing.T) {
	d := NewDispatcher()

	// Should not panic when emitting an event with no handlers
	assert.NotPanics(t, func() {
		d.Emit(Event{
			Type:    "unknown-event",
			Source:  "test",
			Message: "this should be silently dropped",
		})
	})
}

func TestDispatcher_Emit_Good_MultipleEventTypes(t *testing.T) {
	d := NewDispatcher()

	var starts, errors int
	d.On(EventStart, func(e Event) { starts++ })
	d.On(EventError, func(e Event) { errors++ })

	d.Emit(Event{Type: EventStart, Source: "test"})
	d.Emit(Event{Type: EventStart, Source: "test"})
	d.Emit(Event{Type: EventError, Source: "test"})

	assert.Equal(t, 2, starts)
	assert.Equal(t, 1, errors)
}

func TestDispatcher_Emit_Good_SetsTime(t *testing.T) {
	d := NewDispatcher()

	var received Event
	d.On(EventItem, func(e Event) {
		received = e
	})

	before := time.Now()
	d.Emit(Event{Type: EventItem, Source: "test"})
	after := time.Now()

	assert.True(t, received.Time.After(before) || received.Time.Equal(before))
	assert.True(t, received.Time.Before(after) || received.Time.Equal(after))
}

func TestDispatcher_Emit_Good_PreservesExistingTime(t *testing.T) {
	d := NewDispatcher()

	customTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	var received Event
	d.On(EventItem, func(e Event) {
		received = e
	})

	d.Emit(Event{Type: EventItem, Source: "test", Time: customTime})
	assert.True(t, customTime.Equal(received.Time))
}

func TestDispatcher_EmitHelpers_Good(t *testing.T) {
	d := NewDispatcher()

	events := make(map[string]Event)
	for _, eventType := range []string{EventStart, EventProgress, EventItem, EventError, EventComplete} {
		et := eventType
		d.On(et, func(e Event) {
			events[et] = e
		})
	}

	d.EmitStart("s1", "started")
	d.EmitProgress("s2", "progressing", map[string]int{"count": 5})
	d.EmitItem("s3", "got item", nil)
	d.EmitError("s4", "something failed", nil)
	d.EmitComplete("s5", "done", nil)

	assert.Equal(t, "s1", events[EventStart].Source)
	assert.Equal(t, "started", events[EventStart].Message)

	assert.Equal(t, "s2", events[EventProgress].Source)
	assert.NotNil(t, events[EventProgress].Data)

	assert.Equal(t, "s3", events[EventItem].Source)
	assert.Equal(t, "s4", events[EventError].Source)
	assert.Equal(t, "s5", events[EventComplete].Source)
}

func TestNewDispatcher_Good(t *testing.T) {
	d := NewDispatcher()
	assert.NotNil(t, d)
	assert.NotNil(t, d.handlers)
}
