# ADR 0002: IPC Bridge Pattern

* Status: accepted
* Deciders: Project Maintainers
* Date: 2025-05-15

## Context and Problem Statement

Wails allows direct binding of Go methods to the frontend. However, as the number of services and methods grows, managing individual bindings for every service becomes complex. We need a way to decouple the frontend from the internal service structure.

## Decision Drivers

* Decoupling services from the frontend runtime.
* Simplified binding generation.
* Centralized message routing.
* Uniform internal and external communication.

## Considered Options

* Direct Wails Bindings for all services.
* IPC Bridge Pattern (Centralized ACTION handler).

## Decision Outcome

Chosen option: "IPC Bridge Pattern", because it allows services to remain agnostic of the frontend runtime. Only the `Core` service is registered with Wails, and it exposes a single `ACTION` method that routes messages to the appropriate service based on an IPC handler.

### Positive Consequences

* Only one Wails service needs to be registered.
* Services can be tested independently of Wails.
* Adding new functionality to a service doesn't necessarily require regenerating frontend bindings.
* Consistency between frontend-to-backend and backend-to-backend communication.

### Negative Consequences

* Less type safety out-of-the-box in the frontend for specific service methods (though this can be improved with manual type definitions or codegen).
* Requires services to implement `HandleIPCEvents`.
