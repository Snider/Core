# Core

Core is a Web3 Framework, written in Go using Wails.io to replace Electron and the bloat of browsers that, at their core, still live in their mum's basement.

More to come, follow us on Discord http://discord.dappco.re


Repo: https://github.com/Snider/Core

## Quick start

```go
import core "github.com/Snider/Core"

app := core.New(
  core.WithServiceLock(),
)
```

## Development Workflow

This project follows a Test-Driven Development (TDD) approach. We use [Task](https://taskfile.dev/) for task automation to streamline the development process.

The recommended workflow is:

1.  **Generate Tests**: For any changes to the public API, first generate the necessary test stubs.

    ```bash
    task test-gen
    ```

2.  **Run Tests (and watch them fail)**: Verify that the new tests fail as expected.

    ```bash
    task test
    ```

3.  **Implement Your Feature**: Write the code to make the tests pass.

4.  **Run Tests Again**: Ensure all tests now pass.

    ```bash
    task test
    ```

5.  **Submit for Review**: Once your changes are complete and tests are passing, submit them for a CodeRabbit review.

    ```bash
    task review
    ```

## Project Structure

The project is organized into the following main directories:

- `pkg/`: Contains the core Go packages that make up the framework.
- `cmd/`: Contains the entry points for the two main applications:
  - `core-gui/`: The Wails-based GUI application.
  - `core/`: The command-line interface (CLI) application.

## Prerequisites

- [Go](https://go.dev/)
- [Node.js](https://nodejs.org/)
- [Wails](https://wails.io/)
- [Task](https://taskfile.dev/)

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

## Available Tasks

To run any of the following tasks, open your terminal in the project's root directory and execute the `task` command.

### General Tasks

- `task test`: Runs all Go tests recursively for the entire project.
- `task test-gen`: Generates tests for the public API.
- `task check`: A comprehensive check that runs `go mod tidy`, the full test suite, and a CodeRabbit review.
- `task review`: Submits the current changes for a CodeRabbit review.
- `task cov`: Generates a test coverage profile (`coverage.txt`).
- `task cov-view`: Opens the HTML coverage report in your browser.
- `task sync`: Updates the public API Go files to match the exported interface of the modules.

### GUI Application (`cmd/core-gui`)

These tasks are run from the root directory and operate on the GUI application.

- `task gui:build`: Builds the GUI application.
- `task gui:package`: Packages a production build of the GUI application.
- `task gui:run`: Runs the GUI application.
- `task gui:dev`: Runs the GUI application in development mode, with hot-reloading enabled.

### CLI Application (`cmd/core`)

These tasks are run from the root directory and operate on the CLI application.

- `task cli:build`: Builds the CLI application.
- `task cli:build:dev`: Builds the CLI application for development.
- `task cli:run`: Builds and runs the CLI application.
- `task cli:sync`: Updates the public API Go files.
- `task cli:test-gen`: Generates tests for the public API.

## Docs (MkDocs)
The help site and in‑app docs are built with MkDocs Material and live under `pkg/v1/core/docs`.

- Live preview: from `pkg/v1/core/docs` run
  - `pip install -r requirements.txt`
  - `mkdocs serve -o -c` (or `task dev` if you use Task)
- Build static site: `mkdocs build --clean -d public` (or `task build`)

The demo app embeds the built docs from `public/` and can open specific sections in new windows using stable, short headings.
