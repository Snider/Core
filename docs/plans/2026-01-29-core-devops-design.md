# Core DevOps CLI Design (S4.6)

## Summary

Portable development environment CLI commands for the core-devops LinuxKit image. Provides a sandboxed, immutable environment with 100+ embedded tools.

## Design Decisions

- **Image sources**: GitHub Releases + Container Registry + CDN (try in order, configurable)
- **Local storage**: `~/.core/images/` with `CORE_IMAGES_DIR` env override
- **Shell connection**: SSH by default, `--console` for serial fallback
- **Serve**: Mount PWD into VM via 9P/SSHFS, run auto-detected dev server
- **Test**: Auto-detect framework + `.core/test.yaml` config + `--` override
- **Update**: Simple hash/version check, `--force` to always download
- **Claude sandbox**: SSH in with forwarded auth, safe experimentation in immutable image

## Package Structure

```
pkg/devops/
├── devops.go           # DevOps struct, Boot/Stop/Status
├── images.go           # ImageManager, manifest handling
├── mount.go            # Directory mounting (9P, SSHFS)
├── serve.go            # Project detection, serve command
├── test.go             # Test detection, .core/test.yaml parsing
├── config.go           # ~/.core/config.yaml handling
└── sources/
    ├── source.go       # ImageSource interface
    ├── github.go       # GitHub Releases
    ├── registry.go     # Container registry
    └── cdn.go          # CDN/S3

cmd/core/cmd/dev.go     # CLI commands
```

## Image Storage

```
~/.core/
├── config.yaml              # Global config (image source preference, etc.)
└── images/
    ├── core-devops-darwin-arm64.qcow2
    ├── core-devops-darwin-amd64.qcow2
    ├── core-devops-linux-amd64.qcow2
    └── manifest.json        # Tracks versions, hashes, last-updated
```

## ImageSource Interface

```go
type ImageSource interface {
    Name() string
    Available() bool
    LatestVersion() (string, error)
    Download(ctx context.Context, dest string) error
}
```

Sources tried in order: GitHub → Registry → CDN, or respect user preference in config.

## CLI Commands

```go
// cmd/core/cmd/dev.go

func AddDevCommand(app *clir.Cli) {
    devCmd := app.NewSubCommand("dev", "Portable development environment")

    // core dev install [--source github|registry|cdn]
    // Downloads core-devops image for current platform

    // core dev boot [--memory 4096] [--cpus 4] [--name mydev]
    // Boots the dev environment (detached by default)

    // core dev shell [--console]
    // SSH into running dev env (or serial console with --console)

    // core dev serve [--port 8000]
    // Mount PWD → /app, run FrankenPHP, forward port

    // core dev test [-- custom command]
    // Auto-detect tests or use .core/test.yaml or pass custom

    // core dev claude [--auth] [--model opus|sonnet]
    // SSH in with forwarded auth, start Claude in sandbox

    // core dev update [--force]
    // Check for newer image, download if available

    // core dev status
    // Show if dev env is running, resource usage, ports

    // core dev stop
    // Stop the running dev environment
}
```

## Command Flow

```
First time:
  core dev install     → Downloads ~/.core/images/core-devops-{os}-{arch}.qcow2
  core dev boot        → Starts VM in background
  core dev shell       → SSH in

Daily use:
  core dev boot        → Start (if not running)
  core dev serve       → Mount project, start server
  core dev test        → Run tests inside VM
  core dev shell       → Interactive work

AI sandbox:
  core dev claude      → SSH + forward auth + start Claude CLI

Maintenance:
  core dev update      → Get latest image
  core dev status      → Check what's running
```

## `core dev claude` - Sandboxed AI Session

```bash
core dev claude              # Forward all auth by default
core dev claude --no-auth    # Clean session, no host credentials
core dev claude --auth=gh,anthropic  # Selective forwarding
```

**What it does:**
1. Ensures dev VM is running (auto-boots if not)
2. Forwards auth credentials from host:
   - `~/.anthropic/` or `ANTHROPIC_API_KEY`
   - `~/.config/gh/` (GitHub CLI auth)
   - SSH agent forwarding
   - Git config (name, email)
3. SSHs into VM with agent forwarding (`ssh -A`)
4. Starts `claude` CLI inside with forwarded context
5. Current project mounted at `/app`

**Why this is powerful:**
- Immutable base = reset anytime with `core dev boot --fresh`
- Claude can experiment freely, install packages, make mistakes
- Host system untouched
- Still has real credentials to push code, create PRs
- Full 100+ tools available in core-devops image

## Test Configuration

**`.core/test.yaml` format:**
```yaml
version: 1

# Commands to run (in order)
commands:
  - name: unit
    run: vendor/bin/pest --parallel
  - name: types
    run: vendor/bin/phpstan analyse
  - name: lint
    run: vendor/bin/pint --test

# Or simple single command
command: npm test

# Environment variables
env:
  APP_ENV: testing
  DB_CONNECTION: sqlite
```

**Auto-Detection Priority:**
1. `.core/test.yaml`
2. `composer.json` scripts.test → `composer test`
3. `package.json` scripts.test → `npm test`
4. `go.mod` → `go test ./...`
5. `pytest.ini` or `pyproject.toml` → `pytest`
6. `Taskfile.yaml` → `task test`

**CLI Usage:**
```bash
core dev test              # Auto-detect and run
core dev test --unit       # Run only "unit" from .core/test.yaml
core dev test -- go test -v ./pkg/...  # Override with custom
```

## `core dev serve` - Mount & Serve

**How it works:**
1. Ensure VM is running
2. Mount current directory into VM via 9P virtio-fs (or SSHFS fallback)
3. Start auto-detected dev server on /app inside VM
4. Forward port to host

**Mount Strategy:**
```go
type MountMethod int
const (
    Mount9P      MountMethod = iota  // QEMU virtio-9p (faster)
    MountSSHFS                        // sshfs reverse mount
    MountRSync                        // Fallback: rsync on change
)
```

**CLI Usage:**
```bash
core dev serve                    # Mount PWD, serve on :8000
core dev serve --port 3000        # Custom port
core dev serve --path ./backend   # Serve subdirectory
```

**Project Detection:**
```go
func detectServeCommand(projectDir string) string {
    if exists("artisan") {
        return "php artisan octane:start --host=0.0.0.0 --port=8000"
    }
    if exists("package.json") && hasScript("dev") {
        return "npm run dev -- --host 0.0.0.0"
    }
    if exists("composer.json") {
        return "frankenphp php-server"
    }
    return "python -m http.server 8000"  // Fallback
}
```

## Image Sources & Updates

**~/.core/config.yaml:**
```yaml
version: 1

images:
  source: auto  # auto | github | registry | cdn

  cdn:
    url: https://images.example.com/core-devops

  github:
    repo: host-uk/core-images

  registry:
    image: ghcr.io/host-uk/core-devops
```

**Manifest for Update Checking:**
```json
// ~/.core/images/manifest.json
{
  "core-devops-darwin-arm64.qcow2": {
    "version": "v1.2.0",
    "sha256": "abc123...",
    "downloaded": "2026-01-29T10:00:00Z",
    "source": "github"
  }
}
```

**Update Flow:**
```go
func (d *DevOps) Update(force bool) error {
    local := d.manifest.Get(imageName)
    remote, _ := d.source.LatestVersion()

    if force || local.Version != remote {
        fmt.Printf("Updating %s → %s\n", local.Version, remote)
        return d.source.Download(ctx, imagePath)
    }
    fmt.Println("Already up to date")
    return nil
}
```

## Commands Summary

| Command | Description |
|---------|-------------|
| `core dev install` | Download image for platform |
| `core dev boot` | Start VM (auto-installs if needed) |
| `core dev shell` | SSH in (--console for serial) |
| `core dev serve` | Mount PWD, run dev server |
| `core dev test` | Run tests inside VM |
| `core dev claude` | Start Claude session in sandbox |
| `core dev update` | Check/download newer image |
| `core dev status` | Show VM state, ports, resources |
| `core dev stop` | Stop the VM |

## Dependencies

- Reuse existing `pkg/container` for VM management (LinuxKitManager)
- SSH client for shell/exec (golang.org/x/crypto/ssh)
- Progress bar for downloads (charmbracelet/bubbles or similar)

## Implementation Steps

1. Create `pkg/devops/` package structure
2. Implement ImageSource interface and sources (GitHub, Registry, CDN)
3. Implement image download with manifest tracking
4. Implement config loading (`~/.core/config.yaml`)
5. Add CLI commands to `cmd/core/cmd/dev.go`
6. Implement boot/stop using existing LinuxKitManager
7. Implement shell (SSH + serial console)
8. Implement serve (mount + project detection)
9. Implement test (detection + .core/test.yaml)
10. Implement claude (auth forwarding + sandbox)
11. Implement update (version check + download)
12. Implement status
