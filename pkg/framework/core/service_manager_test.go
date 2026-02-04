package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceManager_RegisterService_Good(t *testing.T) {
	m := newServiceManager()

	err := m.registerService("svc1", &MockService{Name: "one"})
	assert.NoError(t, err)

	got := m.service("svc1")
	assert.NotNil(t, got)
	assert.Equal(t, "one", got.(*MockService).GetName())
}

func TestServiceManager_RegisterService_Bad(t *testing.T) {
	m := newServiceManager()

	// Empty name
	err := m.registerService("", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be empty")

	// Duplicate
	err = m.registerService("dup", &MockService{})
	assert.NoError(t, err)
	err = m.registerService("dup", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already registered")

	// Locked
	m2 := newServiceManager()
	m2.enableLock()
	m2.applyLock()
	err = m2.registerService("late", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serviceLock")
}

func TestServiceManager_ServiceNotFound_Good(t *testing.T) {
	m := newServiceManager()
	assert.Nil(t, m.service("nonexistent"))
}

func TestServiceManager_Startables_Good(t *testing.T) {
	m := newServiceManager()

	s1 := &MockStartable{}
	s2 := &MockStartable{}

	_ = m.registerService("s1", s1)
	_ = m.registerService("s2", s2)

	startables := m.getStartables()
	assert.Len(t, startables, 2)

	// Verify order matches registration order
	assert.Same(t, s1, startables[0])
	assert.Same(t, s2, startables[1])

	// Verify it's a copy — mutating the slice doesn't affect internal state
	startables[0] = nil
	assert.Len(t, m.getStartables(), 2)
	assert.NotNil(t, m.getStartables()[0])
}

func TestServiceManager_Stoppables_Good(t *testing.T) {
	m := newServiceManager()

	s1 := &MockStoppable{}
	s2 := &MockStoppable{}

	_ = m.registerService("s1", s1)
	_ = m.registerService("s2", s2)

	stoppables := m.getStoppables()
	assert.Len(t, stoppables, 2)

	// Stoppables are returned in registration order; Core.ServiceShutdown reverses them
	assert.Same(t, s1, stoppables[0])
	assert.Same(t, s2, stoppables[1])
}

func TestServiceManager_Lock_Good(t *testing.T) {
	m := newServiceManager()

	// Register before lock — should succeed
	err := m.registerService("early", &MockService{})
	assert.NoError(t, err)

	// Enable and apply lock
	m.enableLock()
	m.applyLock()

	// Register after lock — should fail
	err = m.registerService("late", &MockService{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "serviceLock")

	// Early service is still accessible
	assert.NotNil(t, m.service("early"))
}

func TestServiceManager_LockNotAppliedWithoutEnable_Good(t *testing.T) {
	m := newServiceManager()
	m.applyLock() // applyLock without enableLock should be a no-op

	err := m.registerService("svc", &MockService{})
	assert.NoError(t, err)
}

type mockFullLifecycle struct{}

func (m *mockFullLifecycle) OnStartup(_ context.Context) error  { return nil }
func (m *mockFullLifecycle) OnShutdown(_ context.Context) error { return nil }

func TestServiceManager_LifecycleBoth_Good(t *testing.T) {
	m := newServiceManager()

	svc := &mockFullLifecycle{}
	err := m.registerService("both", svc)
	assert.NoError(t, err)

	// Should appear in both startables and stoppables
	assert.Len(t, m.getStartables(), 1)
	assert.Len(t, m.getStoppables(), 1)
}
