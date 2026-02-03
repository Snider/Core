# Core

<<<<<<< HEAD
> NOTICE
> The current version here is not the current version, which is in letheanVPN/desktop/services/core (from memory). This was a nice idea, but I'm reorganising the code bases. Check back later.


[![codecov](https://codecov.io/github/Snider/Core/branch/dev/graph/badge.svg?token=I4DVF9746V)](https://codecov.io/github/Snider/Core)
=======
[![codecov](https://codecov.io/gh/host-uk/core/branch/dev/graph/badge.svg)](https://codecov.io/gh/host-uk/core)
[![Go Test Coverage](https://github.com/host-uk/core/actions/workflows/coverage.yml/badge.svg)](https://github.com/host-uk/core/actions/workflows/coverage.yml)
[![Code Scanning](https://github.com/host-uk/core/actions/workflows/codescan.yml/badge.svg)](https://github.com/host-uk/core/actions/workflows/codescan.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/host-uk/core)](https://go.dev/)
[![License](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](https://opensource.org/licenses/EUPL-1.2)
>>>>>>> 8c93abb (docs: add CI and coverage badges to README)

Core is a Web3 Framework, written in Go using Wails.io to replace Electron and the bloat of browsers that, at their core, still live in their mum's basement.

<<<<<<< HEAD
More to come, follow us on Discord http://discord.dappco.re
=======
- Repo: https://github.com/host-uk/core
>>>>>>> b553afa (feat: CI improvements and release channels (#266))


Repo: https://github.com/Snider/Core

## Quick start

```go
import core "github.com/host-uk/core"

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

- Go 1.25+ (this repo targets Go 1.25 and uses workspaces)
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

<<<<<<< HEAD
To build the CLI application:
=======
| Task | Description |
|------|-------------|
| `task test` | Run all Go tests |
| `task test-gen` | Generate test stubs for public API |
| `task check` | go mod tidy + tests + review |
| `task review` | CodeRabbit review |
| `task cov` | Generate coverage.txt |
| `task cov-view` | Open HTML coverage report |
| `task sync` | Update public API Go files |

---

## Architecture

### Project Structure

```
.
├── core.go              # Facade re-exporting pkg/core
├── pkg/
│   ├── core/            # Service container, DI, Runtime[T]
│   ├── config/          # JSON persistence, XDG paths
│   ├── display/         # Windows, tray, menus (Wails)
│   ├── crypt/           # Hashing, checksums, PGP
│   │   └── openpgp/     # Full PGP implementation
│   ├── io/              # Medium interface + backends
│   ├── workspace/       # Encrypted workspace management
│   ├── help/            # In-app documentation
│   └── i18n/            # Internationalization
├── cmd/
│   ├── core/            # CLI application
│   └── core-gui/        # Wails GUI application
└── go.work              # Links root, cmd/core, cmd/core-gui
```

### Service Pattern (Dual-Constructor DI)

Every service follows this pattern:

```go
// Static DI - standalone use/testing (no core.Runtime)
func New() (*Service, error)

// Dynamic DI - for core.WithService() registration
func Register(c *core.Core) (any, error)
```

Services embed `*core.Runtime[Options]` for access to `Core()` and `Config()`.

### IPC/Action System

Services implement `HandleIPCEvents(c *core.Core, msg core.Message) error` - auto-discovered via reflection. Handles typed actions like `core.ActionServiceStartup`.

---

## Wails v3 Frontend Bindings

Core uses [Wails v3](https://v3alpha.wails.io/) to expose Go methods to a WebView2 browser runtime. Wails automatically generates TypeScript bindings for registered services.

**Documentation:** [Wails v3 Method Bindings](https://v3alpha.wails.io/features/bindings/methods/)

### How It Works

1. **Go services** with exported methods are registered with Wails
2. Run `wails3 generate bindings` (or `wails3 dev` / `wails3 build`)
3. **TypeScript SDK** is generated in `frontend/bindings/`
4. Frontend calls Go methods with full type safety, no HTTP overhead

### Current Binding Architecture

```go
// cmd/core-gui/main.go
app.RegisterService(application.NewService(coreService))  // Only Core is registered
```

**Problem:** Only `Core` is registered with Wails. Sub-services (crypt, workspace, display, etc.) are internal to Core's service map - their methods aren't directly exposed to JS.

**Currently exposed** (see `cmd/core-gui/public/bindings/`):
```typescript
// From frontend:
import { ACTION, Config, Service } from './bindings/github.com/host-uk/core/pkg/core'

ACTION(msg)              // Broadcast IPC message
Config()                 // Get config service reference
Service("workspace")     // Get service by name (returns any)
```

**NOT exposed:** Direct calls like `workspace.CreateWorkspace()` or `crypt.Hash()`.

### The IPC Bridge Pattern (Chosen Architecture)

Sub-services are accessed via Core's **IPC/ACTION system**, not direct Wails bindings:

```typescript
// Frontend calls Core.ACTION() with typed messages
import { ACTION } from './bindings/github.com/host-uk/core/pkg/core'

// Open a window
ACTION({ action: "display.open_window", name: "settings", options: { Title: "Settings", Width: 800 } })

// Switch workspace
ACTION({ action: "workspace.switch_workspace", name: "myworkspace" })
```

Each service implements `HandleIPCEvents(c *core.Core, msg core.Message)` to process these messages:

```go
// pkg/display/display.go
func (s *Service) HandleIPCEvents(c *core.Core, msg core.Message) error {
    switch m := msg.(type) {
    case map[string]any:
        if action, ok := m["action"].(string); ok && action == "display.open_window" {
            return s.handleOpenWindowAction(m)
        }
    }
    return nil
}
```

**Why this pattern:**
- Single Wails service (Core) = simpler binding generation
- Services remain decoupled from Wails
- Centralized message routing via `ACTION()`
- Services can communicate internally using same pattern

**Current gap:** Not all service methods have IPC handlers yet. See `HandleIPCEvents` in each service to understand what's wired up.

### Generating Bindings
>>>>>>> b553afa (feat: CI improvements and release channels (#266))

```bash
task cli:build
```

<<<<<<< HEAD
The executable will be located in the `cmd/core/bin` directory.
=======
Bindings output to `cmd/core-gui/public/bindings/github.com/host-uk/core/` mirroring Go package structure.
>>>>>>> b553afa (feat: CI improvements and release channels (#266))

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
The documentation site is powered by MkDocs Material and lives under `docs/` with configuration in `mkdocs.yml`.

- Install docs tooling:
  - `pip install -r docs/requirements.txt`
- Live preview from repository root:
  - `mkdocs serve -o -c`
- Build static site:
  - `mkdocs build --clean`

## Releasing (GoReleaser)
This repo includes a minimal GoReleaser config (`.goreleaser.yaml`). Tagged pushes like `v1.2.3` will build and publish archives via GitHub Actions (see `.github/workflows/release.yml`).

- Local dry run: `goreleaser release --snapshot --clean`
- Real release: create and push a version tag `vX.Y.Z`.

## Go Workspaces
This repository uses Go workspaces (`go.work`) targeting Go 1.25.

- Add/remove modules with `go work use`.
- Typical workflow:
  - `go work sync`
  - `go mod tidy` in modules as needed
