package core

import (
	"fmt"
	"sync"
)

// serviceManager owns the service registry and lifecycle tracking.
// It is an unexported component used internally by Core.
type serviceManager struct {
	mu          sync.RWMutex
	services    map[string]any
	startables  []Startable
	stoppables  []Stoppable
	lockEnabled bool // WithServiceLock was called
	locked      bool // lock applied after New() completes
}

// newServiceManager creates an empty service manager.
func newServiceManager() *serviceManager {
	return &serviceManager{
		services: make(map[string]any),
	}
}

// registerService adds a named service to the registry.
// It also appends to startables/stoppables if the service implements those interfaces.
func (m *serviceManager) registerService(name string, svc any) error {
	if name == "" {
		return fmt.Errorf("core: service name cannot be empty")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.locked {
		return fmt.Errorf("core: service %q is not permitted by the serviceLock setting", name)
	}
	if _, exists := m.services[name]; exists {
		return fmt.Errorf("core: service %q already registered", name)
	}
	m.services[name] = svc

	if s, ok := svc.(Startable); ok {
		m.startables = append(m.startables, s)
	}
	if s, ok := svc.(Stoppable); ok {
		m.stoppables = append(m.stoppables, s)
	}

	return nil
}

// service retrieves a registered service by name, or nil if not found.
func (m *serviceManager) service(name string) any {
	m.mu.RLock()
	svc, ok := m.services[name]
	m.mu.RUnlock()
	if !ok {
		return nil
	}
	return svc
}

// enableLock marks that the lock should be applied after initialisation.
func (m *serviceManager) enableLock() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lockEnabled = true
}

// applyLock activates the service lock if it was enabled.
// Called once during New() after all options have been processed.
func (m *serviceManager) applyLock() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.lockEnabled {
		m.locked = true
	}
}

// getStartables returns a snapshot copy of the startables slice.
func (m *serviceManager) getStartables() []Startable {
	m.mu.RLock()
	out := append([]Startable(nil), m.startables...)
	m.mu.RUnlock()
	return out
}

// getStoppables returns a snapshot copy of the stoppables slice.
func (m *serviceManager) getStoppables() []Stoppable {
	m.mu.RLock()
	out := append([]Stoppable(nil), m.stoppables...)
	m.mu.RUnlock()
	return out
}
