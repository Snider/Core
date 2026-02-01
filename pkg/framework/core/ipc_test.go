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
