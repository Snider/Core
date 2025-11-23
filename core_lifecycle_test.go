package core

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wailsapp/wails/v3/pkg/application"
)

type MockStartable struct {
	started bool
	err     error
}

func (m *MockStartable) OnStartup(ctx context.Context) error {
	m.started = true
	return m.err
}

type MockStoppable struct {
	stopped bool
	err     error
}

func (m *MockStoppable) OnShutdown(ctx context.Context) error {
	m.stopped = true
	return m.err
}

type MockLifecycle struct {
	MockStartable
	MockStoppable
}

func (m *MockLifecycle) OnStartup(ctx context.Context) error {
	return m.MockStartable.OnStartup(ctx)
}

func (m *MockLifecycle) OnShutdown(ctx context.Context) error {
	return m.MockStoppable.OnShutdown(ctx)
}

func TestCore_LifecycleInterfaces(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	startable := &MockStartable{}
	stoppable := &MockStoppable{}
	lifecycle := &MockLifecycle{}

	// Register services
	err = c.RegisterService("startable", startable)
	assert.NoError(t, err)
	err = c.RegisterService("stoppable", stoppable)
	assert.NoError(t, err)
	err = c.RegisterService("lifecycle", lifecycle)
	assert.NoError(t, err)

	// Startup
	err = c.ServiceStartup(context.Background(), application.ServiceOptions{})
	assert.NoError(t, err)
	assert.True(t, startable.started)
	assert.True(t, lifecycle.started)
	assert.False(t, stoppable.stopped)

	// Shutdown
	err = c.ServiceShutdown(context.Background())
	assert.NoError(t, err)
	assert.True(t, stoppable.stopped)
	assert.True(t, lifecycle.stopped)
}

type MockLifecycleWithLog struct {
	id  string
	log *[]string
}

func (m *MockLifecycleWithLog) OnStartup(ctx context.Context) error {
	*m.log = append(*m.log, "start-"+m.id)
	return nil
}

func (m *MockLifecycleWithLog) OnShutdown(ctx context.Context) error {
	*m.log = append(*m.log, "stop-"+m.id)
	return nil
}

func TestCore_LifecycleOrder(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	var callOrder []string

	s1 := &MockLifecycleWithLog{id: "1", log: &callOrder}
	s2 := &MockLifecycleWithLog{id: "2", log: &callOrder}

	err = c.RegisterService("s1", s1)
	assert.NoError(t, err)
	err = c.RegisterService("s2", s2)
	assert.NoError(t, err)

	// Startup
	err = c.ServiceStartup(context.Background(), application.ServiceOptions{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"start-1", "start-2"}, callOrder)

	// Reset log
	callOrder = nil

	// Shutdown
	err = c.ServiceShutdown(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, []string{"stop-2", "stop-1"}, callOrder)
}

func TestCore_LifecycleErrors(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	s1 := &MockStartable{err: assert.AnError}
	s2 := &MockStoppable{err: assert.AnError}

	c.RegisterService("s1", s1)
	c.RegisterService("s2", s2)

	err = c.ServiceStartup(context.Background(), application.ServiceOptions{})
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)

	err = c.ServiceShutdown(context.Background())
	assert.Error(t, err)
	assert.ErrorIs(t, err, assert.AnError)
}

func TestCore_LifecycleErrors_Aggregated(t *testing.T) {
	c, err := New()
	assert.NoError(t, err)

	// Register action that fails
	c.RegisterAction(func(c *Core, msg Message) error {
		if _, ok := msg.(ActionServiceStartup); ok {
			return errors.New("startup action error")
		}
		if _, ok := msg.(ActionServiceShutdown); ok {
			return errors.New("shutdown action error")
		}
		return nil
	})

	// Register service that fails
	s1 := &MockStartable{err: errors.New("startup service error")}
	s2 := &MockStoppable{err: errors.New("shutdown service error")}

	err = c.RegisterService("s1", s1)
	assert.NoError(t, err)
	err = c.RegisterService("s2", s2)
	assert.NoError(t, err)

	// Startup
	err = c.ServiceStartup(context.Background(), application.ServiceOptions{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "startup action error")
	assert.Contains(t, err.Error(), "startup service error")

	// Shutdown
	err = c.ServiceShutdown(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "shutdown action error")
	assert.Contains(t, err.Error(), "shutdown service error")
}
