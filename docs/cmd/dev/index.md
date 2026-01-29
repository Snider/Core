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

| Command | Description |
|---------|-------------|
| `tasks` | List available tasks from core-agentic |
| `task` | Show task details or auto-select a task |
| `task:update` | Update task status or progress |
| `task:complete` | Mark a task as completed |
| `task:commit` | Auto-commit changes with task reference |
| `task:pr` | Create a pull request for a task |

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

## Task Management Commands

The task commands integrate with the core-agentic service for AI-powered task management.

### Configuration

Task commands load configuration from:
1. Environment variables (`AGENTIC_TOKEN`, `AGENTIC_BASE_URL`)
2. `.env` file in current directory
3. `~/.core/agentic.yaml`

### dev tasks

List available tasks from core-agentic.

```bash
core dev tasks [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--status` | Filter by status (`pending`, `in_progress`, `completed`, `blocked`) |
| `--priority` | Filter by priority (`critical`, `high`, `medium`, `low`) |
| `--labels` | Filter by labels (comma-separated) |
| `--project` | Filter by project |
| `--limit` | Max number of tasks to return (default: 20) |

#### Examples

```bash
core dev tasks
core dev tasks --status pending --priority high
core dev tasks --labels bug,urgent
```

### dev task

Show task details or auto-select a task.

```bash
core dev task [task-id] [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--auto` | Auto-select highest priority pending task |
| `--claim` | Claim the task after showing details |
| `--context` | Show gathered context for AI collaboration |

#### Examples

```bash
# Show task details
core dev task abc123

# Show and claim
core dev task abc123 --claim

# Show with context
core dev task abc123 --context

# Auto-select highest priority pending task
core dev task --auto
```

### dev task:update

Update a task's status, progress, or notes.

```bash
core dev task:update <task-id> [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--status` | New status (`pending`, `in_progress`, `completed`, `blocked`) |
| `--progress` | Progress percentage (0-100) |
| `--notes` | Notes about the update |

#### Examples

```bash
core dev task:update abc123 --status in_progress
core dev task:update abc123 --progress 50 --notes 'Halfway done'
```

### dev task:complete

Mark a task as completed with optional output and artifacts.

```bash
core dev task:complete <task-id> [flags]
```

#### Flags

| Flag | Description |
|------|-------------|
| `--output` | Summary of the completed work |
| `--failed` | Mark the task as failed |
| `--error` | Error message if failed |

#### Examples

```bash
core dev task:complete abc123 --output 'Feature implemented'
core dev task:complete abc123 --failed --error 'Build failed'
```

### dev task:commit

Create a git commit with a task reference and co-author attribution.

```bash
core dev task:commit <task-id> [flags]
```

Commit message format:
```
feat(scope): description

Task: #123
Co-Authored-By: Claude <noreply@anthropic.com>
```

#### Flags

| Flag | Description |
|------|-------------|
| `-m`, `--message` | Commit message (without task reference) |
| `--scope` | Scope for the commit type (e.g., `auth`, `api`, `ui`) |
| `--push` | Push changes after committing |

#### Examples

```bash
core dev task:commit abc123 --message 'add user authentication'
core dev task:commit abc123 -m 'fix login bug' --scope auth
core dev task:commit abc123 -m 'update docs' --push
```

### dev task:pr

Create a GitHub pull request linked to a task.

```bash
core dev task:pr <task-id> [flags]
```

Requires the GitHub CLI (`gh`) to be installed and authenticated.

#### Flags

| Flag | Description |
|------|-------------|
| `--title` | PR title (defaults to task title) |
| `--base` | Base branch (defaults to main) |
| `--draft` | Create as draft PR |
| `--labels` | Labels to add (comma-separated) |

#### Examples

```bash
core dev task:pr abc123
core dev task:pr abc123 --title 'Add authentication feature'
core dev task:pr abc123 --draft --labels 'enhancement,needs-review'
core dev task:pr abc123 --base develop
```

## See Also

- [work](work/) - Multi-repo workflow commands (`core dev work`, `core dev health`, etc.)
