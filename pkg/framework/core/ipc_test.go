package core

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type IPCTestQuery struct{ Value string }
type IPCTestTask struct{ Value string }

func TestIPC_Query(t *testing.T) {
	c, _ := New()

	// No handler
	res, handled, err := c.QUERY(IPCTestQuery{})
	assert.False(t, handled)
	assert.Nil(t, res)
	assert.Nil(t, err)

	// With handler
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		if tq, ok := q.(IPCTestQuery); ok {
			return tq.Value + "-response", true, nil
		}
		return nil, false, nil
	})

	res, handled, err = c.QUERY(IPCTestQuery{Value: "test"})
	assert.True(t, handled)
	assert.Nil(t, err)
	assert.Equal(t, "test-response", res)
}

func TestIPC_QueryAll(t *testing.T) {
	c, _ := New()

	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "h1", true, nil
	})
	c.RegisterQuery(func(c *Core, q Query) (any, bool, error) {
		return "h2", true, nil
	})

	results, err := c.QUERYALL(IPCTestQuery{})
	assert.Nil(t, err)
	assert.Len(t, results, 2)
	assert.Contains(t, results, "h1")
	assert.Contains(t, results, "h2")
}

func TestIPC_Perform(t *testing.T) {
	c, _ := New()

	c.RegisterTask(func(c *Core, task Task) (any, bool, error) {
		if tt, ok := task.(IPCTestTask); ok {
			if tt.Value == "error" {
				return nil, true, errors.New("task error")
			}
			return "done", true, nil
		}
		return nil, false, nil
	})

	// Success
	res, handled, err := c.PERFORM(IPCTestTask{Value: "run"})
	assert.True(t, handled)
	assert.Nil(t, err)
	assert.Equal(t, "done", res)

	// Error
	res, handled, err = c.PERFORM(IPCTestTask{Value: "error"})
	assert.True(t, handled)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func (IPCTestQuery) Response() string { return "" }
func (IPCTestTask) Response() string  { return "" }

func TestIPC_Typed(t *testing.T) {
	c, _ := New()

	RegisterQuery(c, func(c *Core, q IPCTestQuery) (string, bool, error) {
		return q.Value + "-typed", true, nil
	})

	// Using Query
	res, handled, err := Query[string](c, IPCTestQuery{Value: "test"})
	assert.True(t, handled)
	assert.Nil(t, err)
	assert.Equal(t, "test-typed", res)

	// Using DispatchQuery
	res2, handled2, err2 := DispatchQuery(c, IPCTestQuery{Value: "test2"})
	assert.True(t, handled2)
	assert.Nil(t, err2)
	assert.Equal(t, "test2-typed", res2)

	RegisterTask(c, func(c *Core, t IPCTestTask) (string, bool, error) {
		return t.Value + "-done", true, nil
	})

	// Using Perform
	res3, handled3, err3 := Perform[string](c, IPCTestTask{Value: "task"})
	assert.True(t, handled3)
	assert.Nil(t, err3)
	assert.Equal(t, "task-done", res3)

	// Using DispatchTask
	res4, handled4, err4 := DispatchTask(c, IPCTestTask{Value: "task2"})
	assert.True(t, handled4)
	assert.Nil(t, err4)
	assert.Equal(t, "task2-done", res4)
}
