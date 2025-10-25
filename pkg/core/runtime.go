package core

// Runtime is a generic struct that provides a common foundation for all
// Core services. It holds shared dependencies (like the Core pointer) and a
// field for module-specific, type-safe configuration.
//
// T is the type of the module-specific options' struct.
//
// Example:
//
//	type MyModuleOptions struct {
//		Retries int
//	}
//
//	// in practice, this would be a core.Core struct
//	c := &core.Core{}
//
//	runtime := core.NewRuntime(c, MyModuleOptions{
//		Retries: 3,
//	})
type Runtime[T any] struct {
	core *Core

	// Options hold the module-specific configuration of type T.
	//
	// Example:
	//
	//	type MyModuleOptions struct {
	//		Retries int
	//	}
	//
	//	// in practice, this would be a core.Core struct
	//	c := &core.Core{}
	//
	//	runtime := core.NewRuntime(c, MyModuleOptions{
	//		Retries: 3,
	//	})
	//
	//	fmt.Println(runtime.Options.Retries) // Output: 3
	Options T
}

// NewRuntime is a constructor for the generic Runtime.
// It initialises a new runtime with the provided Core pointer and module options.
//
// Example:
//
//	type MyModuleOptions struct {
//		Retries int
//	}
//
//	// in practice, this would be a core.Core struct
//	c := &core.Core{}
//
//	runtime := core.NewRuntime(c, MyModuleOptions{
//		Retries: 3,
//	})
func NewRuntime[T any](c *Core, opts T) *Runtime[T] {
	return &Runtime[T]{
		core:    c,
		Options: opts,
	}
}

// Core returns the central Core instance, providing access to all core functionalities.
//
// Example:
//
//	// in practice, this would be a core.Core struct
//	c := &core.Core{}
//
//	runtime := core.NewRuntime(c, struct{}{})
//	coreInstance := runtime.Core()
func (r *Runtime[T]) Core() *Core {
	return r.core
}
