package core

// ServiceRuntime is a helper struct embedded in services to provide access to the core application.
// It is generic and can be parameterized with a service-specific options struct.
type ServiceRuntime[T any] struct {
	core *Core
	opts T
}

// NewServiceRuntime creates a new ServiceRuntime instance for a service.
// This is typically called by a service's constructor.
func NewServiceRuntime[T any](c *Core, opts T) *ServiceRuntime[T] {
	return &ServiceRuntime[T]{
		core: c,
		opts: opts,
	}
}

// Core returns the central core instance, providing access to all registered services.
func (r *ServiceRuntime[T]) Core() *Core {
	return r.core
}

// Config returns the registered Config service from the core application.
// This is a convenience method for accessing the application's configuration.
func (r *ServiceRuntime[T]) Config() Config {
	return r.core.Config()
}
