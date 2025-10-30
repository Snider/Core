# Developer Guide

This guide provides instructions for building, testing, and interacting with the Core project.

## Project Structure

The project is organized into the following main directories:

- `pkg/`: Contains the core Go packages that make up the framework.
- `cmd/`: Contains the entry points for the two main applications:
  - `core-gui/`: The Wails-based GUI application.
  - `core/`: The command-line interface (CLI) application.

## Development Workflow

This project uses [Task](https://taskfile.dev/) for task automation. The `Taskfile.yml` in the root directory defines a set of tasks to streamline common development operations.

### Prerequisites

- [Go](https://go.dev/)
- [Node.js](https://nodejs.org/)
- [Wails](https://wails.io/)
- [Task](https://taskfile.dev/)

### Available Tasks

To run any of the following tasks, open your terminal in the project's root directory and execute the `task` command.

#### General Tasks

- `task test`: Runs all Go tests recursively for the entire project.
- `task check`: A comprehensive check that runs `go mod tidy`, the full test suite, and a CodeRabbit review.
- `task review`: Submits the current changes for a CodeRabbit review.
- `task cov`: Generates a test coverage profile (`coverage.txt`).
- `task cov-view`: Opens the HTML coverage report in your browser.

#### GUI Application (`cmd/core-gui`)

These tasks are run from the root directory and operate on the GUI application.

- `task gui:build`: Builds the GUI application.
- `task gui:package`: Packages a production build of the GUI application.
- `task gui:run`: Runs the GUI application.
- `task gui:dev`: Runs the GUI application in development mode, with hot-reloading enabled.

#### CLI Application (`cmd/core`)

These tasks are run from the root directory and operate on the CLI application.

- `task cli:build`: Builds the CLI application.
- `task cli:build:dev`: Builds the CLI application for development.

## Building and Running

### GUI Application

To run the GUI application in development mode:

```bash
task gui:dev
```

To build the final application for your platform:

```bash
task gui:build
```

### CLI Application

To build the CLI application:

```bash
task cli:build
```

The executable will be located in the `cmd/core/bin` directory.
