# Development Guide

This guide covers how to set up the development environment, build the project, and run the demo.

## Prerequisites

1.  **Go:** Version 1.25 or later.
2.  **Node.js & npm:** For building the Angular frontend.
3.  **Wails Dependencies:**
    - **Linux:** `libgtk-3-dev`, `libwebkit2gtk-4.1-dev`
    - **macOS/Windows:** Standard development tools (Xcode Command Line Tools / build-essential).

## Setup

1.  Clone the repository:
    ```bash
    git clone https://github.com/Snider/display.git
    cd display
    ```

2.  Install Go dependencies:
    ```bash
    go mod tidy
    ```

3.  Install Frontend dependencies:
    ```bash
    cd ui
    npm install
    cd ..
    ```

## Running the Demo

The project includes a CLI to facilitate development.

### Serve Mode (Web Preview)
To start a simple HTTP server that serves the frontend and a mock API:

1.  Build the frontend first:
    ```bash
    cd ui && npm run build && cd ..
    ```
2.  Run the serve command:
    ```bash
    go run ./cmd/demo-cli serve
    ```
    Access the app at `http://localhost:8080`.

## Building the Project

### Frontend
```bash
cd ui
npm run build
```

### Backend / Application
This project is a library/module. However, it can be tested via the demo CLI or by integrating it into a Wails application entry point.

To run the tests:
```bash
go test ./...
```
