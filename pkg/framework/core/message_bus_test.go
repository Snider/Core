package core

import (
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMessageBus_Action_Good(t *testing.T) {
	c, _ := New()

	var received []Message
	c.bus.registerAction(func(_ *Core, msg Message) error {
		received = append(received, msg)
		return nil
	})
	c.bus.registerAction(func(_ *Core, msg Message) error {
		received = append(received, msg)
		return nil
	})

	err := c.bus.action("hello")
	assert.NoError(t, err)
	assert.Len(t, received, 2)
}

func TestMessageBus_Action_Bad(t *testing.T) {
	c, _ := New()

	err1 := errors.New("handler1 failed")
	err2 := errors.New("handler2 failed")

	c.bus.registerAction(func(_ *Core, msg Message) error { return err1 })
	c.bus.registerAction(func(_ *Core, msg Message) error { return nil })
	c.bus.registerAction(func(_ *Core, msg Message) error { return err2 })

	err := c.bus.action("test")
	assert.Error(t, err)
	assert.ErrorIs(t, err, err1)
	assert.ErrorIs(t, err, err2)
}

func TestMessageBus_RegisterAction_Good(t *testing.T) {
	c, _ := New()

	var coreRef *Core
	c.bus.registerAction(func(core *Core, msg Message) error {
		coreRef = core
		return nil
	})

	_ = c.bus.action(nil)
	assert.Same(t, c, coreRef, "handler should receive the Core reference")
}

func TestMessageBus_Query_Good(t *testing.T) {
	c, _ := New()

	c.bus.registerQuery(func(_ *Core, q Query) (any, bool, error) {
		return "first", true, nil
	})

	result, handled, err := c.bus.query(TestQuery{Value: "test"})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "first", result)
}

func TestMessageBus_QueryAll_Good(t *testing.T) {
	c, _ := New()

	c.bus.registerQuery(func(_ *Core, q Query) (any, bool, error) {
		return "a", true, nil
	})
	c.bus.registerQuery(func(_ *Core, q Query) (any, bool, error) {
		return nil, false, nil // skips
	})
	c.bus.registerQuery(func(_ *Core, q Query) (any, bool, error) {
		return "b", true, nil
	})

	results, err := c.bus.queryAll(TestQuery{})
	assert.NoError(t, err)
	assert.Equal(t, []any{"a", "b"}, results)
}

func TestMessageBus_Perform_Good(t *testing.T) {
	c, _ := New()

	c.bus.registerTask(func(_ *Core, t Task) (any, bool, error) {
		return "done", true, nil
	})

	result, handled, err := c.bus.perform(TestTask{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "done", result)
}

func TestMessageBus_ConcurrentAccess_Good(t *testing.T) {
	c, _ := New()

	var wg sync.WaitGroup
	const goroutines = 20

	// Concurrent register + dispatch
	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.bus.registerAction(func(_ *Core, msg Message) error { return nil })
		}()
		go func() {
			defer wg.Done()
			_ = c.bus.action("ping")
		}()
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.bus.registerQuery(func(_ *Core, q Query) (any, bool, error) { return nil, false, nil })
		}()
		go func() {
			defer wg.Done()
			_, _ = c.bus.queryAll(TestQuery{})
		}()
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			c.bus.registerTask(func(_ *Core, t Task) (any, bool, error) { return nil, false, nil })
		}()
		go func() {
			defer wg.Done()
			_, _, _ = c.bus.perform(TestTask{})
		}()
	}

	wg.Wait()
}

func TestMessageBus_Action_NoHandlers(t *testing.T) {
	c, _ := New()
	// Should not error if no handlers are registered
	err := c.bus.action("no one listening")
	assert.NoError(t, err)
}

func TestMessageBus_Query_NoHandlers(t *testing.T) {
	c, _ := New()
	result, handled, err := c.bus.query(TestQuery{})
	assert.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, result)
}

func TestMessageBus_QueryAll_NoHandlers(t *testing.T) {
	c, _ := New()
	results, err := c.bus.queryAll(TestQuery{})
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestMessageBus_Perform_NoHandlers(t *testing.T) {
	c, _ := New()
	result, handled, err := c.bus.perform(TestTask{})
	assert.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, result)
}
