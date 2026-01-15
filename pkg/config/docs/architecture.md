# Architecture Overview

This project implements a modular configuration system with a Go backend and an Angular frontend, orchestrated by a CLI runner.

## Components

The architecture consists of three main components:

1.  **Backend Library (`pkg/config`)**
    - **Responsibility**: Manages configuration loading, saving, and persistence.
    - **Tech Stack**: Go.
    - **Features**:
        - Struct-based configuration with JSON persistence.
        - Generic key-value storage supporting JSON, YAML, INI, and XML.
        - XDG-compliant directory management.
    - **Integration**: Can be used via static (`New`) or dynamic (`Register`) dependency injection.

2.  **Frontend (`ui`)**
    - **Responsibility**: Provides a user interface for the application.
    - **Tech Stack**: Angular.
    - **Deployment**: Built as a static asset (or Custom Element) served by the backend.

3.  **CLI Runner (`cmd/demo-cli`)**
    - **Responsibility**: Entry point for the application.
    - **Tech Stack**: Go (`spf13/cobra`).
    - **Functions**:
        - `serve`: Starts a web server that hosts the compiled Angular frontend and API endpoints.

## Data Flow

1.  **Initialization**: The CLI starts and initializes the `pkg/config` service to load settings from disk.
2.  **Serving**: The CLI's `serve` command spins up an HTTP server.
3.  **Interaction**: Users interact with the Angular UI in the browser.
4.  **API Communication**: The UI communicates with the backend via REST API endpoints (e.g., `/api/v1/demo`).
5.  **Persistence**: Configuration changes made via the backend service are persisted to the filesystem.

## Directory Structure

```
├── cmd/
│   └── demo-cli/       # CLI Application entry point
├── pkg/
│   └── config/         # Core configuration logic (Go)
├── ui/                 # Frontend application (Angular)
└── docs/               # Project documentation
```
