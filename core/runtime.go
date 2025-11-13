package core

// Runtime is a helper struct embedded in services to provide access to the core application.
type Runtime[T any] struct {
	core *Core
	opts T
}

// NewRuntime creates a new Runtime instance for a service.
func NewRuntime[T any](c *Core, opts T) *Runtime[T] {
	return &Runtime[T]{
		core: c,
		opts: opts,
	}
}

// Core returns the central core instance.
func (r *Runtime[T]) Core() *Core {
	return r.core
}

// Config returns the registered Config service from the core application.
func (r *Runtime[T]) Config() Config {
	return r.core.Config()
}
