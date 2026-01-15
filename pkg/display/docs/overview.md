# Overview

The `display` module is a core component responsible for the visual presentation and system integration of the application. It leverages **Wails v3** to create a desktop application backend in Go and **Angular** for the frontend user interface.

## Architecture

The project consists of two main parts:

1.  **Backend (Go):** Handles window management, system tray integration, application menus, and communication with the frontend. It is located in the root directory and packaged as a Go module.
2.  **Frontend (Angular):** Provides the user interface. It is located in the `ui/` directory and is built as a custom element that interacts with the backend.

## Key Components

### Display Service (`display`)
The core service that manages the application lifecycle. It wraps the Wails application instance and exposes methods to:
- Open and configure windows.
- Manage the system tray.
- Show system dialogs (e.g., environment info).

### System Integration
- **Menu:** A standard application menu (File, Edit, View, etc.) is constructed in `menu.go`.
- **System Tray:** A system tray icon and menu are configured in `tray.go`, allowing quick access to common actions like showing/hiding windows or viewing environment info.

### Demo CLI
A command-line interface (`cmd/demo-cli`) is provided to run and test the display module. It includes a `serve` command for web-based development.
