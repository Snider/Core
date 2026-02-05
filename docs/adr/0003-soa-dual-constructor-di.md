# ADR 0003: Service-Oriented Architecture with Dual-Constructor DI

* Status: accepted
* Deciders: Project Maintainers
* Date: 2025-05-15

## Context and Problem Statement

The application consists of many components (config, crypt, workspace, etc.) that depend on each other. We need a consistent way to manage these dependencies and allow for easy testing.

## Decision Drivers

* Testability.
* Modularity.
* Ease of service registration.
* Clear lifecycle management.

## Considered Options

* Global variables/singletons.
* Dependency Injection (DI) container.
* Manual Dependency Injection.

## Decision Outcome

Chosen option: "Service-Oriented Architecture with Dual-Constructor DI". Each service follows a pattern where it provides a `New()` constructor for standalone use/testing (static DI) and a `Register()` function for registration with the `Core` service container (dynamic DI).

### Positive Consequences

* Easy to unit test services by passing mock dependencies to `New()`.
* Automatic service discovery and lifecycle management via `Core`.
* Decoupled components.

### Negative Consequences

* Some boilerplate required for each service (`New` and `Register`).
* Dependency on `pkg/core` for `Register`.
