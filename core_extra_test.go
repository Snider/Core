package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockServiceWithIPC struct {
	MockService
	handled bool
}

func (m *MockServiceWithIPC) HandleIPCEvents(c *Core, msg Message) error {
	m.handled = true
	return nil
}

func TestCore_WithService_IPC(t *testing.T) {
	svc := &MockServiceWithIPC{MockService: MockService{Name: "ipc-service"}}
	factory := func(c *Core) (any, error) {
		return svc, nil
	}
	c, err := New(WithService(factory))
	assert.NoError(t, err)

	// Trigger ACTION to verify handler was registered
	err = c.ACTION(nil)
	assert.NoError(t, err)
	assert.True(t, svc.handled)
}

func TestCore_ACTION_Bad(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)
	errHandler := func(c *Core, msg Message) error {
		return assert.AnError
	}
	c.RegisterAction(errHandler)
	err = c.ACTION(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), assert.AnError.Error())
}
