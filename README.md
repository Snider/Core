# Core

Core is a Web3 Framework, written in Go using Wails.io to replace Electron and the bloat of browsers that, at their core, still live in their mum's basement.

- Discord: http://discord.dappco.re
- Repo: https://github.com/Snider/Core

## Vision

Core is an **opinionated Web3 desktop application framework** providing:

1. **Service-Oriented Architecture** - Pluggable services with dependency injection
2. **Encrypted Workspaces** - Each workspace gets its own PGP keypair, files are obfuscated
3. **Cross-Platform Storage** - Abstract storage backends (local, SFTP, WebDAV) behind a `Medium` interface
4. **Multi-Brand Support** - Same codebase powers different "hub" apps (AdminHub, ServerHub, GatewayHub, DeveloperHub, ClientHub)
5. **Built-in Crypto** - PGP encryption/signing, hashing, checksums as first-class citizens

**Mental model:** A secure, encrypted workspace manager where each "workspace" is a cryptographically isolated environment. The framework handles windows, menus, trays, config, and i18n.

## Quick Start

```go
import core "github.com/Snider/Core"

app := core.New(
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
import { ACTION, Config, Service } from './bindings/github.com/Snider/Core/pkg/core'

ACTION(msg)              // Broadcast IPC message
Config()                 // Get config service reference
Service("workspace")     // Get service by name (returns any)
```

**NOT exposed:** Direct calls like `workspace.CreateWorkspace()` or `crypt.Hash()`.

### The IPC Bridge Pattern (Chosen Architecture)

Sub-services are accessed via Core's **IPC/ACTION system**, not direct Wails bindings:

```typescript
// Frontend calls Core.ACTION() with typed messages
import { ACTION } from './bindings/github.com/Snider/Core/pkg/core'

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

```bash
cd cmd/core-gui
wails3 generate bindings    # Regenerate after Go changes
```

Bindings output to `cmd/core-gui/public/bindings/github.com/Snider/Core/` mirroring Go package structure.

---

### Service Interfaces (`pkg/core/interfaces.go`)

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
| `pkg/core` | Service container, DI, thread-safe - solid |
| `pkg/config` | JSON persistence, XDG paths - solid |
| `pkg/crypt` | Hashing, checksums, PGP - solid, well-tested |
| `pkg/help` | Embedded docs, Show/ShowAt - solid |
| `pkg/i18n` | Multi-language with go-i18n - solid |
| `pkg/io` | Medium interface + local backend - solid |
| `pkg/workspace` | Workspace creation, switching, file ops - functional |

### Partial

| Package | Issues |
|---------|--------|
| `pkg/display` | Window creation works; menu/tray handlers are TODOs |

---

## Priority Work Items

### 1. IMPLEMENT: System Tray Brand Support

`pkg/display/tray.go:52-63` - Commented brand-specific menu items need implementation.

### 2. ADD: Integration Tests

| Package | Notes |
|---------|-------|
| `pkg/display` | Integration tests requiring Wails runtime (27% unit coverage) |

---

## Package Deep Dives

### pkg/workspace - The Core Feature

Each workspace is:
1. Identified by LTHN hash of user identifier
2. Has directory structure: `config/`, `log/`, `data/`, `files/`, `keys/`
3. Gets a PGP keypair generated on creation
4. Files accessed via obfuscated paths

The `workspaceList` maps workspace IDs to public keys.

### pkg/crypt/openpgp

Full PGP using `github.com/ProtonMail/go-crypto`:
- `CreateKeyPair(name, passphrase)` - RSA-4096 with revocation cert
- `EncryptPGP()` - Encrypt + optional signing
- `DecryptPGP()` - Decrypt + optional signature verification

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

## For New Contributors

1. Run `task test` to verify all tests pass
2. Follow TDD: `task test-gen` creates stubs, implement to pass
3. The dual-constructor pattern is intentional: `New(deps)` for tests, `Register()` for runtime
4. See `cmd/core-gui/main.go` for how services wire together
5. IPC handlers in each service's `HandleIPCEvents()` are the frontend bridge
