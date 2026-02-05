# Core

[![codecov](https://codecov.io/gh/host-uk/core/branch/dev/graph/badge.svg)](https://codecov.io/gh/host-uk/core)
[![Go Test Coverage](https://github.com/host-uk/core/actions/workflows/coverage.yml/badge.svg)](https://github.com/host-uk/core/actions/workflows/coverage.yml)
[![Code Scanning](https://github.com/host-uk/core/actions/workflows/codescan.yml/badge.svg)](https://github.com/host-uk/core/actions/workflows/codescan.yml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/host-uk/core)](https://go.dev/)
[![License](https://img.shields.io/badge/License-EUPL--1.2-blue.svg)](https://opensource.org/licenses/EUPL-1.2)

Core is a Web3 Framework, written in Go using Wails.io to replace Electron and the bloat of browsers that, at their core, still live in their mum's basement.

- Repo: https://github.com/host-uk/core

## Vision

Core is an **opinionated Web3 desktop application framework** providing:

1. **Service-Oriented Architecture** - Pluggable services with dependency injection
2. **Encrypted Workspaces** - Each workspace gets its own PGP keypair, files are obfuscated
3. **Cross-Platform Storage** - Abstract storage backends (local, SFTP, WebDAV) behind a `Medium` interface
4. **Multi-Brand Support** - Same codebase powers different "hub" apps (AdminHub, ServerHub, GatewayHub, DeveloperHub, ClientHub)
5. **Built-in Crypto** - PGP encryption/signing, hashing, checksums as first-class citizens

**Mental model:** A secure, encrypted workspace manager where each "workspace" is a cryptographically isolated environment. The framework handles windows, menus, trays, config, and i18n.

## CLI Quick Start

```bash
# 1. Install Core
go install github.com/host-uk/core/cmd/core@latest

# 2. Verify environment
core doctor

# 3. Run tests in any Go/PHP project
core go test   # or core php test

# 4. Build and preview release
core build
core ci
```

For more details, see the [User Guide](docs/user-guide.md).

## Framework Quick Start (Go)

```go
import core "github.com/host-uk/core/pkg/framework/core"

app, err := core.New(
  core.WithServiceLock(),
)
```

## Prerequisites

- [Go](https://go.dev/) 1.25+
- [Node.js](https://nodejs.org/)
- [Wails](https://wails.io/) v3
- [Task](https://taskfile.dev/)

## Development Workflow (TDD)

```bash
task test-gen    # 1. Generate test stubs
task test        # 2. Run tests (watch them fail)
# 3. Implement your feature
task test        # 4. Run tests (watch them pass)
task review      # 5. CodeRabbit review
```

## Building & Running

```bash
# GUI (Wails)
task gui:dev      # Development with hot-reload
task gui:build    # Production build

# CLI
task cli:build    # Build to cmd/core/bin/core
task cli:run      # Build and run
```

## All Tasks

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
├── main.go              # CLI application entry point
├── pkg/
│   ├── framework/core/  # Service container, DI, Runtime[T]
│   ├── crypt/           # Hashing, checksums, PGP
│   ├── io/              # Medium interface + backends
│   ├── help/            # In-app documentation
│   ├── i18n/            # Internationalization
│   ├── repos/           # Multi-repo registry & management
│   ├── agentic/         # AI agent task management
│   └── mcp/             # Model Context Protocol service
├── internal/
│   ├── cmd/             # CLI command implementations
│   └── variants/        # Build variants (full, minimal, etc.)
└── go.mod               # Go module definition
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

## Configuration Management

Core uses a **centralized configuration service** implemented in `pkg/config`, with YAML-based persistence and layered overrides.

The `pkg/config` package provides:

- YAML-backed persistence at `~/.core/config.yaml`
- Dot-notation key access (for example: `cfg.Set("dev.editor", "vim")`, `cfg.GetString("dev.editor")`)
- Environment variable overlay support (env vars can override persisted values)
- Thread-safe operations for concurrent reads/writes

Application code should treat `pkg/config` as the **primary configuration mechanism**. Direct reads/writes to YAML files should generally be avoided from application logic in favour of using this centralized service.

### Project and Service Configuration Files

In addition to the centralized configuration service, Core uses several YAML files for project-specific build/CI and service configuration. These live alongside (but are distinct from) the centralized configuration:

- **Project Configuration** (in the `.core/` directory of the project root):
    - `build.yaml`: Build targets, flags, and project metadata.
    - `release.yaml`: Release automation, changelog settings, and publishing targets.
    - `ci.yaml`: CI pipeline configuration.
- **Global Configuration** (in the `~/.core/` directory):
    - `config.yaml`: Centralized user/framework settings and defaults, managed via `pkg/config`.
    - `agentic.yaml`: Configuration for agentic services (BaseURL, Token, etc.).
- **Registry Configuration** (`repos.yaml`, auto-discovered):
    - Multi-repo registry definition.
    - Searched in the current directory and its parent directories (walking up).
    - Then in `~/Code/host-uk/repos.yaml`.
    - Finally in `~/.config/core/repos.yaml`.

### Format

All persisted configuration files described above use **YAML** format for readability and nested structure support.

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

Wails v3 bindings are typically generated in the GUI repository (e.g., `core-gui`).

```bash
wails3 generate bindings    # Regenerate after Go changes
```

---

### Service Interfaces (`pkg/framework/core/interfaces.go`)

```go
type Config interface {
    Get(key string, out any) error
    Set(key string, v any) error
}

type Display interface {
    OpenWindow(opts ...WindowOption) error
}

type Workspace interface {
    CreateWorkspace(identifier, password string) (string, error)
    SwitchWorkspace(name string) error
    WorkspaceFileGet(filename string) (string, error)
    WorkspaceFileSet(filename, content string) error
}

type Crypt interface {
    EncryptPGP(writer io.Writer, recipientPath, data string, ...) (string, error)
    DecryptPGP(recipientPath, message, passphrase string, ...) (string, error)
}
```

---

## Current State (Prototype)

### Working

| Package | Notes |
|---------|-------|
| `pkg/framework/core` | Service container, DI, thread-safe - solid |
| `pkg/config` | Layered YAML configuration, XDG paths - solid |
| `pkg/crypt` | Hashing, checksums, symmetric/asymmetric - solid, well-tested |
| `pkg/help` | Embedded docs, full-text search - solid |
| `pkg/i18n` | Multi-language with go-i18n - solid |
| `pkg/io` | Medium interface + local backend - solid |
| `pkg/repos` | Multi-repo registry & management - solid |
| `pkg/agentic` | AI agent task management - solid |
| `pkg/mcp` | Model Context Protocol service - solid |

---

## Package Deep Dives

### pkg/crypt

The crypt package provides a comprehensive suite of cryptographic primitives:
- **Hashing & Checksums**: SHA-256, SHA-512, and CRC32 support.
- **Symmetric Encryption**: AES-GCM and ChaCha20-Poly1305 for secure data at rest.
- **Key Derivation**: Argon2id for secure password hashing.
- **Asymmetric Encryption**: PGP implementation in the `pkg/crypt/openpgp` subpackage using `github.com/ProtonMail/go-crypto`.

### pkg/io - Storage Abstraction

```go
type Medium interface {
    Read(path string) (string, error)
    Write(path, content string) error
    EnsureDir(path string) error
    IsFile(path string) bool
    FileGet(path string) (string, error)
    FileSet(path, content string) error
}
```

Implementations: `local/`, `sftp/`, `webdav/`

---

## Future Work

### Phase 1: Core Stability
- [x] ~~Fix workspace medium injection (critical blocker)~~
- [x] ~~Initialize `io.Local` global~~
- [x] ~~Clean up dead code (orphaned vars, broken wrappers)~~
- [x] ~~Wire up IPC handlers for all services (config, crypt, display, help, i18n, workspace)~~
- [x] ~~Complete display menu handlers (New/List workspace)~~
- [x] ~~Tray icon setup with asset embedding~~
- [x] ~~Test coverage for io packages~~
- [ ] System tray brand-specific menus

### Phase 2: Multi-Brand Support
- [ ] Define brand configuration system (config? build flags?)
- [ ] Implement brand-specific tray menus (AdminHub, ServerHub, GatewayHub, DeveloperHub, ClientHub)
- [ ] Brand-specific theming/assets
- [ ] Per-brand default workspace configurations

### Phase 3: Remote Storage
- [ ] Complete SFTP backend (`pkg/io/sftp/`)
- [ ] Complete WebDAV backend (`pkg/io/webdav/`)
- [ ] Workspace sync across storage backends
- [ ] Conflict resolution for multi-device access

### Phase 4: Enhanced Crypto
- [ ] Key management UI (import/export, key rotation)
- [ ] Multi-recipient encryption
- [ ] Hardware key support (YubiKey, etc.)
- [ ] Encrypted workspace backup/restore

### Phase 5: Developer Experience
- [ ] TypeScript types for IPC messages (codegen from Go structs)
- [ ] Hot-reload for service registration
- [ ] Plugin system for third-party services
- [ ] CLI tooling for workspace management

### Phase 6: Distribution
- [ ] Auto-update mechanism
- [ ] Platform installers (DMG, MSI, AppImage)
- [ ] Signing and notarization
- [ ] Crash reporting integration

---

## Getting Help

- **[User Guide](docs/user-guide.md)**: Detailed usage and concepts.
- **[FAQ](docs/faq.md)**: Frequently asked questions.
- **[Workflows](docs/workflows.md)**: Common task sequences.
- **[Troubleshooting](docs/troubleshooting.md)**: Solving common issues.
- **[Configuration](docs/configuration.md)**: Config file reference.

```bash
# Check environment
core doctor

# Command help
core <command> --help
```

---

## For New Contributors

1. Run `task test` to verify all tests pass
2. Follow TDD: `task test-gen` creates stubs, implement to pass
3. The dual-constructor pattern is intentional: `New(deps)` for tests, `Register()` for runtime
4. IPC handlers in each service's `HandleIPCEvents()` are the frontend bridge
