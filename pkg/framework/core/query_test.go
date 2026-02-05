package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestQuery struct {
	Value string
}

type TestTask struct {
	Value string
}

func TestCore_QUERY_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Register a handler that responds to TestQuery
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		if tq, ok := q.(TestQuery); ok {
			return "result-" + tq.Value, true, nil
		}
		return nil, false, nil
	})

	result, handled, err := c.QUERY(TestQuery{Value: "test"})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "result-test", result)
}

func TestCore_QUERY_NotHandled(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// No handlers registered
	result, handled, err := c.QUERY(TestQuery{Value: "test"})
	assert.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, result)
}

func TestCore_QUERY_FirstResponder(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// First handler responds
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "first", true, nil
	})

	// Second handler would respond but shouldn't be called
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "second", true, nil
	})

	result, handled, err := c.QUERY(TestQuery{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "first", result)
}

func TestCore_QUERY_SkipsNonHandlers(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// First handler doesn't handle
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return nil, false, nil
	})

	// Second handler responds
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "second", true, nil
	})

	result, handled, err := c.QUERY(TestQuery{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "second", result)
}

func TestCore_QUERYALL_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Multiple handlers respond
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "first", true, nil
	})

	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "second", true, nil
	})

	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return nil, false, nil // Doesn't handle
	})

	results, err := c.QUERYALL(TestQuery{})
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, results, "first")
	assert.Contains(t, results, "second")
}

func TestCore_QUERYALL_AggregatesErrors(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	err1 := errors.New("error1")
	err2 := errors.New("error2")

	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "result1", true, err1
	})

	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "result2", true, err2
	})

	results, err := c.QUERYALL(TestQuery{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, err1)
	assert.ErrorIs(t, err, err2)
	assert.Len(t, results, 2)
}

func TestCore_PERFORM_Good(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	executed := false
	c.RegisterTask(func(c *Core, t Task) (any, bool, error) {
		if tt, ok := t.(TestTask); ok {
			executed = true
			return "done-" + tt.Value, true, nil
		}
		return nil, false, nil
	})

	result, handled, err := c.PERFORM(TestTask{Value: "work"})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.True(t, executed)
	assert.Equal(t, "done-work", result)
}

func TestCore_PERFORM_NotHandled(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// No handlers registered
	result, handled, err := c.PERFORM(TestTask{Value: "work"})
	assert.NoError(t, err)
	assert.False(t, handled)
	assert.Nil(t, result)
}

func TestCore_PERFORM_FirstResponder(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	callCount := 0

	c.RegisterTask(func(c *Core, t Task) (any, bool, error) {
		callCount++
		return "first", true, nil
	})

	c.RegisterTask(func(c *Core, t Task) (any, bool, error) {
		callCount++
		return "second", true, nil
	})

	result, handled, err := c.PERFORM(TestTask{})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "first", result)
	assert.Equal(t, 1, callCount) // Only first handler called
}

func TestCore_PERFORM_WithError(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	expectedErr := errors.New("task failed")
	c.RegisterTask(func(c *Core, t Task) (any, bool, error) {
		return nil, true, expectedErr
	})

	result, handled, err := c.PERFORM(TestTask{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
	assert.True(t, handled)
	assert.Nil(t, result)
}

func (TestQuery) Response() string { return "" }
func (TestTask) Response() string  { return "" }

func TestTypedIPC(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Test RegisterQuery with types
	RegisterQuery(c, func(c *Core, q TestQuery) (string, bool, error) {
		return "result-" + q.Value, true, nil
	})

	// Test Query with type parameter
	res, handled, err := Query[string](c, TestQuery{Value: "query"})
	assert.NoError(t, err)
	assert.True(t, handled)
	assert.Equal(t, "result-query", res)

	// Test DispatchQuery with inference
	res2, handled2, err2 := DispatchQuery(c, TestQuery{Value: "dispatch"})
	assert.NoError(t, err2)
	assert.True(t, handled2)
	assert.Equal(t, "result-dispatch", res2)

	// Test RegisterTask with types
	RegisterTask(c, func(c *Core, t TestTask) (string, bool, error) {
		return "done-" + t.Value, true, nil
	})

	// Test Perform with type parameter
	res3, handled3, err3 := Perform[string](c, TestTask{Value: "perform"})
	assert.NoError(t, err3)
	assert.True(t, handled3)
	assert.Equal(t, "done-perform", res3)

	// Test DispatchTask with inference
	res4, handled4, err4 := DispatchTask(c, TestTask{Value: "task-dispatch"})
	assert.NoError(t, err4)
	assert.True(t, handled4)
	assert.Equal(t, "done-task-dispatch", res4)
}

type ActionMsg struct {
	Value string
}

func TestTypedAction(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	var received string
	RegisterAction(c, func(c *Core, msg ActionMsg) error {
		received = msg.Value
		return nil
	})

	err = Action(c, ActionMsg{Value: "hello"})
	assert.NoError(t, err)
	assert.Equal(t, "hello", received)
}

func TestTypedQueryAll(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	RegisterQuery(c, func(c *Core, q TestQuery) (string, bool, error) {
		return "one", true, nil
	})
	RegisterQuery(c, func(c *Core, q TestQuery) (string, bool, error) {
		return "two", true, nil
	})

	results, err := QueryAll[string](c, TestQuery{})
	assert.NoError(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, results, "one")
	assert.Contains(t, results, "two")
}
