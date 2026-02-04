package core

import (
	"errors"
	"sync"
)

// messageBus owns the IPC action, query, and task dispatch.
// It is an unexported component used internally by Core.
type messageBus struct {
	core *Core

	ipcMu       sync.RWMutex
	ipcHandlers []func(*Core, Message) error

	queryMu       sync.RWMutex
	queryHandlers []QueryHandler

	taskMu       sync.RWMutex
	taskHandlers []TaskHandler
}

// newMessageBus creates an empty message bus bound to the given Core.
func newMessageBus(c *Core) *messageBus {
	return &messageBus{core: c}
}

// action dispatches a message to all registered IPC handlers.
func (b *messageBus) action(msg Message) error {
	b.ipcMu.RLock()
	handlers := append([]func(*Core, Message) error(nil), b.ipcHandlers...)
	b.ipcMu.RUnlock()

	var agg error
	for _, h := range handlers {
		if err := h(b.core, msg); err != nil {
			agg = errors.Join(agg, err)
		}
	}
	return agg
}

// registerAction adds a single IPC handler.
func (b *messageBus) registerAction(handler func(*Core, Message) error) {
	b.ipcMu.Lock()
	b.ipcHandlers = append(b.ipcHandlers, handler)
	b.ipcMu.Unlock()
}

// registerActions adds multiple IPC handlers.
func (b *messageBus) registerActions(handlers ...func(*Core, Message) error) {
	b.ipcMu.Lock()
	b.ipcHandlers = append(b.ipcHandlers, handlers...)
	b.ipcMu.Unlock()
}

// query dispatches a query to handlers until one responds.
func (b *messageBus) query(q Query) (any, bool, error) {
	b.queryMu.RLock()
	handlers := append([]QueryHandler(nil), b.queryHandlers...)
	b.queryMu.RUnlock()

	for _, h := range handlers {
		result, handled, err := h(b.core, q)
		if handled {
			return result, true, err
		}
	}
	return nil, false, nil
}

// queryAll dispatches a query to all handlers and collects all responses.
func (b *messageBus) queryAll(q Query) ([]any, error) {
	b.queryMu.RLock()
	handlers := append([]QueryHandler(nil), b.queryHandlers...)
	b.queryMu.RUnlock()

	var results []any
	var agg error
	for _, h := range handlers {
		result, handled, err := h(b.core, q)
		if err != nil {
			agg = errors.Join(agg, err)
		}
		if handled && result != nil {
			results = append(results, result)
		}
	}
	return results, agg
}

// registerQuery adds a query handler.
func (b *messageBus) registerQuery(handler QueryHandler) {
	b.queryMu.Lock()
	b.queryHandlers = append(b.queryHandlers, handler)
	b.queryMu.Unlock()
}

// perform dispatches a task to handlers until one executes it.
func (b *messageBus) perform(t Task) (any, bool, error) {
	b.taskMu.RLock()
	handlers := append([]TaskHandler(nil), b.taskHandlers...)
	b.taskMu.RUnlock()

	for _, h := range handlers {
		result, handled, err := h(b.core, t)
		if handled {
			return result, true, err
		}
	}
	return nil, false, nil
}

// registerTask adds a task handler.
func (b *messageBus) registerTask(handler TaskHandler) {
	b.taskMu.Lock()
	b.taskHandlers = append(b.taskHandlers, handler)
	b.taskMu.Unlock()
}
