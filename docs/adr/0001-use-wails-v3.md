# ADR 0001: Use Wails v3 for GUI

* Status: accepted
* Deciders: Project Maintainers
* Date: 2025-05-15

## Context and Problem Statement

The project needs a way to build cross-platform desktop applications with a modern UI. Historically, Electron has been the go-to choice, but it is known for its high resource consumption and large binary sizes.

## Decision Drivers

* Performance and resource efficiency.
* Smaller binary sizes.
* Tight integration with Go.
* Native look and feel.

## Considered Options

* Electron
* Wails (v2)
* Wails (v3)
* Fyne

## Decision Outcome

Chosen option: "Wails (v3)", because it provides the best balance of using web technologies for the UI while keeping the backend in Go with minimal overhead. Wails v3 specifically offers improvements in performance and features over v2.

### Positive Consequences

* Significantly smaller binary sizes compared to Electron.
* Reduced memory usage.
* Ability to use any frontend framework (Vue, React, Svelte, etc.).
* Direct Go-to-JS bindings.

### Negative Consequences

* Wails v3 is still in alpha/beta, which might lead to breaking changes or bugs.
* Smaller ecosystem compared to Electron.
