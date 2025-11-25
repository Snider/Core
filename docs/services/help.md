---
title: help
---
# Service: `help`

The `help` service manages the in-app documentation and help system, allowing users to view guides and context-sensitive help.

## Methods

### `func Show() error`

`Show` opens the main help window or interface.

### `func ShowAt(anchor string) error`

`ShowAt` opens the help window and navigates directly to the section specified by the `anchor`.

### `func HandleIPCEvents(c *core.Core, msg core.Message) error`

`HandleIPCEvents` processes help-related IPC messages, allowing other services to trigger help displays.
