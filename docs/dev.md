# core dev

Portable development environment with 100+ embedded tools.

## Overview

Core DevOps provides a sandboxed, immutable development environment based on LinuxKit. It includes AI tools (Claude, Gemini), runtimes (Go, Node, PHP, Python, Rust), and infrastructure tools (Docker, Kubernetes, Terraform).

## Commands

| Command | Description |
|---------|-------------|
| `install` | Download the core-devops image for your platform |
| `boot` | Start the development environment |
| `stop` | Stop the running environment |
| `status` | Show environment status |
| `shell` | Open a shell in the environment |
| `serve` | Mount project and start dev server |
| `test` | Run tests inside the environment |
| `claude` | Start sandboxed Claude session |
| `update` | Update to latest image |

## Quick Start

```bash
# First time setup
core dev install
core dev boot

# Open shell
core dev shell

# Or mount current project and serve
core dev serve
```

## dev install

Download the core-devops image for your platform.

```bash
core dev install [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--source` | Image source: `github`, `registry`, `cdn` (default: auto) |
| `--force` | Force re-download even if exists |

### Examples

```bash
# Download image (auto-detects platform)
core dev install

# Force re-download
core dev install --force
```

## dev boot

Start the development environment.

```bash
core dev boot [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--memory` | Memory allocation in MB (default: 4096) |
| `--cpus` | Number of CPUs (default: 4) |
| `--name` | Container name (default: core-dev) |

### Examples

```bash
# Start with defaults
core dev boot

# More resources
core dev boot --memory 8192 --cpus 8
```

## dev shell

Open a shell in the running environment.

```bash
core dev shell [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--console` | Use serial console instead of SSH |

### Examples

```bash
# SSH into environment
core dev shell

# Serial console (for debugging)
core dev shell --console
```

## dev serve

Mount current directory and start the appropriate dev server.

```bash
core dev serve [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--port` | Port to expose (default: 8000) |
| `--path` | Subdirectory to serve |

### Auto-Detection

| Project | Server Command |
|---------|---------------|
| Laravel (`artisan`) | `php artisan octane:start` |
| Node (`package.json` with `dev` script) | `npm run dev` |
| PHP (`composer.json`) | `frankenphp php-server` |
| Other | `python -m http.server` |

### Examples

```bash
# Auto-detect and serve
core dev serve

# Custom port
core dev serve --port 3000
```

## dev test

Run tests inside the environment.

```bash
core dev test [flags] [-- custom command]
```

### Flags

| Flag | Description |
|------|-------------|
| `--unit` | Run only unit tests |

### Test Detection

Core auto-detects the test framework:

1. `.core/test.yaml` - Custom config
2. `composer.json` → `composer test`
3. `package.json` → `npm test`
4. `go.mod` → `go test ./...`
5. `pytest.ini` → `pytest`
6. `Taskfile.yaml` → `task test`

### Examples

```bash
# Auto-detect and run tests
core dev test

# Custom command
core dev test -- go test -v ./pkg/...
```

### Test Configuration

Create `.core/test.yaml` for custom test setup:

```yaml
version: 1

commands:
  - name: unit
    run: vendor/bin/pest --parallel
  - name: types
    run: vendor/bin/phpstan analyse
  - name: lint
    run: vendor/bin/pint --test

env:
  APP_ENV: testing
  DB_CONNECTION: sqlite
```

## dev claude

Start a sandboxed Claude session with your project mounted.

```bash
core dev claude [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--no-auth` | Clean session without host credentials |
| `--auth` | Selective auth forwarding (e.g., `gh,anthropic`) |

### What Gets Forwarded

By default, these are forwarded to the sandbox:
- `~/.anthropic/` or `ANTHROPIC_API_KEY`
- `~/.config/gh/` (GitHub CLI auth)
- SSH agent
- Git config (name, email)

### Examples

```bash
# Full auth forwarding (default)
core dev claude

# Clean sandbox
core dev claude --no-auth

# Only GitHub auth
core dev claude --auth=gh
```

### Why Use This?

- **Immutable base** - Reset anytime with `core dev boot --fresh`
- **Safe experimentation** - Claude can install packages, make mistakes
- **Host system untouched** - All changes stay in the sandbox
- **Real credentials** - Can still push code, create PRs
- **Full tooling** - 100+ tools available in the image

## dev status

Show the current state of the development environment.

```bash
core dev status
```

Output includes:
- Running/stopped state
- Resource usage (CPU, memory)
- Exposed ports
- Mounted directories

## dev update

Check for and download newer images.

```bash
core dev update [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--force` | Force download even if up to date |

## Embedded Tools

The core-devops image includes 100+ tools:

| Category | Tools |
|----------|-------|
| **AI/LLM** | claude, gemini, aider, ollama, llm |
| **VCS** | git, gh, glab, lazygit, delta, git-lfs |
| **Runtimes** | frankenphp, node, bun, deno, go, python3, rustc |
| **Package Mgrs** | composer, npm, pnpm, yarn, pip, uv, cargo |
| **Build** | task, make, just, nx, turbo |
| **Linting** | pint, phpstan, prettier, eslint, biome, golangci-lint, ruff |
| **Testing** | phpunit, pest, vitest, playwright, k6 |
| **Infra** | docker, kubectl, k9s, helm, terraform, ansible |
| **Databases** | sqlite3, mysql, psql, redis-cli, mongosh, usql |
| **HTTP/Net** | curl, httpie, xh, websocat, grpcurl, mkcert, ngrok |
| **Data** | jq, yq, fx, gron, miller, dasel |
| **Security** | age, sops, cosign, trivy, trufflehog, vault |
| **Files** | fd, rg, fzf, bat, eza, tree, zoxide, broot |
| **Editors** | nvim, helix, micro |

## Configuration

Global config in `~/.core/config.yaml`:

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

## Image Storage

Images are stored in `~/.core/images/`:

```
~/.core/
├── config.yaml
└── images/
    ├── core-devops-darwin-arm64.qcow2
    ├── core-devops-linux-amd64.qcow2
    └── manifest.json
```
