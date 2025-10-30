# Core

Core is an opinionated Go framework for building desktop apps with Wails. It provides a small set of focused modules:

- Core — framework bootstrap and service container
- Core.Config — app and UI state persistence
- Core.Crypt — keys, encrypt/decrypt, sign/verify
- Core.Display — windows, tray, window state
- Core.Docs — in‑app help and deep‑links
- Core.IO — local/remote filesystem helpers
- [Core.Workspace](https://core.help/) — projects and paths

Help: https://Core.Help \
Web: https://dAppCo.re \
Repo: https://github.com/Snider/Core

## Quick start
```go
import core "github.com/Snider/Core"

app := core.New(
  core.WithServiceLock(),
)
```

## Tasks

This project uses [Task](https://taskfile.dev/) for task automation. The root `Taskfile.yml` includes the `Taskfile.yml` from `cmd/core-gui` and `cmd/core` under the `gui` and `cli` namespaces respectively. The following tasks are available:

### General

- `task test`: Run all Go tests recursively for the entire project.
- `task review`: Run CodeRabbit review to get feedback on the current changes.
- `task check`: Run `go mod tidy`, the full test suite, and a CodeRabbit review.

### GUI Application (`cmd/core-gui`)

- `task gui:build`: Builds the GUI application.
- `task gui:package`: Packages a production build of the GUI application.
- `task gui:run`: Runs the GUI application.
- `task gui:dev`: Runs the GUI application in development mode.

### CLI Application (`cmd/core`)

- `task cli:build`: Builds the CLI application.
- `task cli:build:dev`: Builds the CLI application for development.
- `task cli:run`: Builds and runs the CLI application.

## Docs (MkDocs)
The help site and in‑app docs are built with MkDocs Material and live under `pkg/v1/core/docs`.

- Live preview: from `pkg/v1/core/docs` run
  - `pip install -r requirements.txt`
  - `mkdocs serve -o -c` (or `task dev` if you use Task)
- Build static site: `mkdocs build --clean -d public` (or `task build`)

The demo app embeds the built docs from `public/` and can open specific sections in new windows using stable, short headings.
