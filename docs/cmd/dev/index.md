# core dev

Multi-repo workflow and portable development environment.

## Multi-Repo Commands

| Command | Description |
|---------|-------------|
| [work](work/) | Full workflow: status + commit + push |
| `health` | Quick health check across repos |
| `commit` | Claude-assisted commits |
| `push` | Push repos with unpushed commits |
| `pull` | Pull repos that are behind |
| `issues` | List open issues |
| `reviews` | List PRs needing review |
| `ci` | Check CI status |
| `impact` | Show dependency impact |
| `api` | Tools for managing service APIs |
| `sync` | Synchronize public service APIs |

## Task Management Commands

> **Note:** Task management commands have moved to [`core ai`](../ai/).

| Command | Description |
|---------|-------------|
| [`ai tasks`](../ai/) | List available tasks from core-agentic |
| [`ai task`](../ai/) | Show task details or auto-select a task |
| [`ai task:update`](../ai/) | Update task status or progress |
| [`ai task:complete`](../ai/) | Mark a task as completed |
| [`ai task:commit`](../ai/) | Auto-commit changes with task reference |
| [`ai task:pr`](../ai/) | Create a pull request for a task |

## Dev Environment Commands

| Command | Description |
|---------|-------------|
| `install` | Download the core-devops image |
| `boot` | Start the environment |
| `stop` | Stop the environment |
| `status` | Show status |
| `shell` | Open shell |
| `serve` | Start dev server |
| `test` | Run tests |
| `claude` | Sandboxed Claude |
| `update` | Update image |

---

## Dev Environment Overview

Core DevOps provides a sandboxed, immutable development environment based on LinuxKit with 100+ embedded tools.

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
core dev install
```

Downloads the platform-specific dev environment image including Go, PHP, Node.js, Python, Docker, and Claude CLI. Downloads are cached at `~/.core/images/`.

### Examples

```bash
# Download image (auto-detects platform)
core dev install
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
| `--cpus` | Number of CPUs (default: 2) |
| `--fresh` | Stop existing and start fresh |

### Examples

```bash
# Start with defaults
core dev boot

# More resources
core dev boot --memory 8192 --cpus 4

# Fresh start
core dev boot --fresh
```

## dev shell

Open a shell in the running environment.

```bash
core dev shell [flags] [-- command]
```

Uses SSH by default, or serial console with `--console`.

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

# Run a command
core dev shell -- ls -la
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
| `--name` | Run named test command from `.core/test.yaml` |

### Test Detection

Core auto-detects the test framework or uses `.core/test.yaml`:

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

# Run named test from config
core dev test --name integration

# Custom command
core dev test -- go test -v ./pkg/...
```

### Test Configuration

Create `.core/test.yaml` for custom test setup - see [Configuration](example.md#configuration) for examples.

## dev claude

Start a sandboxed Claude session with your project mounted.

```bash
core dev claude [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--model` | Model to use (`opus`, `sonnet`) |
| `--no-auth` | Don't forward any auth credentials |
| `--auth` | Selective auth forwarding (`gh`, `anthropic`, `ssh`, `git`) |

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

# Use Opus model
core dev claude --model opus

# Clean sandbox
core dev claude --no-auth

# Only GitHub and Anthropic auth
core dev claude --auth gh,anthropic
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

Check for and apply updates.

```bash
core dev update [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--apply` | Download and apply the update |

### Examples

```bash
# Check for updates
core dev update

# Apply available update
core dev update --apply
```

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

Global config in `~/.core/config.yaml` - see [Configuration](example.md#configuration) for examples.

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

## Multi-Repo Commands

See the [work](work/) page for detailed documentation on multi-repo commands.

### dev ci

Check GitHub Actions workflow status across all repos.

```bash
core dev ci [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--registry` | Path to `repos.yaml` (auto-detected if not specified) |
| `--branch` | Filter by branch (default: main) |
| `--failed` | Show only failed runs |

Requires the `gh` CLI to be installed and authenticated.

### dev api

Tools for managing service APIs.

```bash
core dev api sync
```

Synchronizes the public service APIs with their internal implementations.

### dev sync

Alias for `core dev api sync`. Synchronizes the public service APIs with their internal implementations.

```bash
core dev sync
```

This command scans the `pkg` directory for services and ensures that the top-level public API for each service is in sync with its internal implementation. It automatically generates the necessary Go files with type aliases.

## See Also

- [work](work/) - Multi-repo workflow commands (`core dev work`, `core dev health`, etc.)
- [ai](../ai/) - Task management commands (`core ai tasks`, `core ai task`, etc.)
