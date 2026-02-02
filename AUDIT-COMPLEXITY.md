# Code Complexity & Maintainability Audit

This document analyzes the code quality of the Core framework, identifies maintainability issues, and provides recommendations for improvement.

## High-Level Findings

### 1. Centralized "God Object" (`Core` struct)

The `pkg/framework/core/core.go` file defines a `Core` struct that acts as a central hub for managing services, dependency injection, IPC, and more. While this centralizes control, it also creates a "God Object" that is tightly coupled to all other components of the system.

**Impact:**

-   **High Coupling:** Changes to the `Core` struct can have cascading effects throughout the codebase.
-   **Reduced Cohesion:** The `Core` struct has many unrelated responsibilities, making it difficult to understand and maintain.
-   **Testing Complexity:** Testing components that depend on the `Core` struct requires a complex setup.

**Recommendation:**

-   **Refactor to smaller, more focused components:** Break down the `Core` struct's responsibilities into smaller, more cohesive modules. For example, create separate managers for services, IPC, and lifecycle events.

### 2. IPC System Lacks Type Safety

The current IPC system uses `interface{}` for messages, which means that message types are not checked at compile time. This can lead to runtime errors and makes the code more difficult to refactor.

**Impact:**

-   **Runtime Errors:** Type mismatches in IPC messages will only be caught at runtime.
-   **Refactoring Challenges:** It is difficult to safely rename or modify message structures without introducing breaking changes.
-   **Reduced Readability:** The lack of explicit types makes it harder to understand the data that is being passed between services.

**Recommendation:**

-   **Introduce a typed messaging system:** Use generics or code generation to create a type-safe messaging system. This will provide compile-time guarantees and improve the overall robustness of the framework.

---

## Detailed Analysis

### `pkg/framework`

#### `core.go`: The `Core` God Object

The `Core` struct in `pkg/framework/core/interfaces.go` is a classic example of a God Object. It has too many responsibilities, leading to high coupling and low cohesion.

**Code Example (`Core` struct definition):**

```go
type Core struct {
	App            any // GUI runtime
	assets         embed.FS
	Features       *Features
	serviceLock    bool
	ipcMu          sync.RWMethods // Responsibility 1: IPC
	ipcHandlers    []func(*Core, Message) error
	queryMu        sync.RWMutex // Responsibility 2: Queries
	queryHandlers  []QueryHandler
	taskMu         sync.RWMutex // Responsibility 3: Tasks
	taskHandlers   []TaskHandler
	serviceMu      sync.RWMutex // Responsibility 4: Service Management
	services       map[string]any
	servicesLocked bool
	startables     []Startable // Responsibility 5: Lifecycle Management
	stoppables     []Stoppable
}
```

**Maintainability Issues:**

*   **Single Responsibility Principle Violation:** The struct combines service management, event dispatching (IPC, queries, tasks), and lifecycle management.
*   **High Complexity:** Any function that interacts with the `Core` has access to a wide range of functionalities, making it harder to reason about the code.
*   **Difficult to Test:** Mocks for the `Core` struct would be large and complex.

**Refactoring Recommendation:**

Decompose the `Core` struct into smaller, more focused components.

1.  **Introduce a `ServiceManager`:** This component would handle service registration and retrieval.

    ```go
    type ServiceManager struct {
        mu       sync.RWMutex
        services map[string]any
        locked   bool
    }

    func (sm *ServiceManager) Register(name string, service any) error { /* ... */ }
    func (sm *ServiceManager) Get(name string) any { /* ... */ }
    ```

2.  **Introduce a `MessageBus`:** This component would manage IPC, queries, and tasks.

    ```go
    type MessageBus struct {
        ipcMu       sync.RWMutex
        ipcHandlers []func(*Core, Message) error
        // ... and so on for queries and tasks
    }

    func (mb *MessageBus) Dispatch(msg Message) error { /* ... */ }
    ```

3.  **Introduce a `LifecycleManager`:** This component would manage `Startable` and `Stoppable` services.

    ```go
    type LifecycleManager struct {
        startables []Startable
        stoppables []Stoppable
    }

    func (lm *LifecycleManager) Startup(ctx context.Context) error { /* ... */ }
    func (lm *LifecycleManager) Shutdown(ctx context.Context) error { /* ... */ }
    ```

4.  **Compose the `Core` struct from these new components:**

    ```go
    type Core struct {
        App              any
        assets           embed.FS
        Features         *Features
        ServiceManager   *ServiceManager
        MessageBus       *MessageBus
        LifecycleManager *LifecycleManager
    }
    ```

This approach promotes the Single Responsibility Principle, reduces coupling, and makes the system easier to test and maintain.
