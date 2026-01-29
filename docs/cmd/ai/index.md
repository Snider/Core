# core ai

AI agent task management and Claude Code integration.

## Task Management Commands

| Command | Description |
|---------|-------------|
| `tasks` | List available tasks from core-agentic |
| `task` | View task details or auto-select |
| `task:update` | Update task status or progress |
| `task:complete` | Mark task as completed or failed |
| `task:commit` | Create git commit with task reference |
| `task:pr` | Create GitHub PR linked to task |

## Claude Integration

| Command | Description |
|---------|-------------|
| `claude run` | Run Claude Code in current directory |
| `claude config` | Manage Claude configuration |

---

## Configuration

Task commands load configuration from:
1. Environment variables (`AGENTIC_TOKEN`, `AGENTIC_BASE_URL`)
2. `.env` file in current directory
3. `~/.core/agentic.yaml`

---

## ai tasks

List available tasks from core-agentic.

```bash
core ai tasks [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--status` | Filter by status (`pending`, `in_progress`, `completed`, `blocked`) |
| `--priority` | Filter by priority (`critical`, `high`, `medium`, `low`) |
| `--labels` | Filter by labels (comma-separated) |
| `--project` | Filter by project |
| `--limit` | Max number of tasks to return (default: 20) |

### Examples

```bash
# List all pending tasks
core ai tasks

# Filter by status and priority
core ai tasks --status pending --priority high

# Filter by labels
core ai tasks --labels bug,urgent
```

---

## ai task

View task details or auto-select a task.

```bash
core ai task [task-id] [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--auto` | Auto-select highest priority pending task |
| `--claim` | Claim the task after showing details |
| `--context` | Show gathered context for AI collaboration |

### Examples

```bash
# Show task details
core ai task abc123

# Show and claim
core ai task abc123 --claim

# Show with context
core ai task abc123 --context

# Auto-select highest priority pending task
core ai task --auto
```

---

## ai task:update

Update a task's status, progress, or notes.

```bash
core ai task:update <task-id> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--status` | New status (`pending`, `in_progress`, `completed`, `blocked`) |
| `--progress` | Progress percentage (0-100) |
| `--notes` | Notes about the update |

### Examples

```bash
# Set task to in progress
core ai task:update abc123 --status in_progress

# Update progress with notes
core ai task:update abc123 --progress 50 --notes 'Halfway done'
```

---

## ai task:complete

Mark a task as completed with optional output and artifacts.

```bash
core ai task:complete <task-id> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--output` | Summary of the completed work |
| `--failed` | Mark the task as failed |
| `--error` | Error message if failed |

### Examples

```bash
# Complete successfully
core ai task:complete abc123 --output 'Feature implemented'

# Mark as failed
core ai task:complete abc123 --failed --error 'Build failed'
```

---

## ai task:commit

Create a git commit with a task reference and co-author attribution.

```bash
core ai task:commit <task-id> [flags]
```

Commit message format:
```
feat(scope): description

Task: #123
Co-Authored-By: Claude <noreply@anthropic.com>
```

### Flags

| Flag | Description |
|------|-------------|
| `-m`, `--message` | Commit message (without task reference) |
| `--scope` | Scope for the commit type (e.g., `auth`, `api`, `ui`) |
| `--push` | Push changes after committing |

### Examples

```bash
# Commit with message
core ai task:commit abc123 --message 'add user authentication'

# With scope
core ai task:commit abc123 -m 'fix login bug' --scope auth

# Commit and push
core ai task:commit abc123 -m 'update docs' --push
```

---

## ai task:pr

Create a GitHub pull request linked to a task.

```bash
core ai task:pr <task-id> [flags]
```

Requires the GitHub CLI (`gh`) to be installed and authenticated.

### Flags

| Flag | Description |
|------|-------------|
| `--title` | PR title (defaults to task title) |
| `--base` | Base branch (defaults to main) |
| `--draft` | Create as draft PR |
| `--labels` | Labels to add (comma-separated) |

### Examples

```bash
# Create PR with defaults
core ai task:pr abc123

# Custom title
core ai task:pr abc123 --title 'Add authentication feature'

# Draft PR with labels
core ai task:pr abc123 --draft --labels 'enhancement,needs-review'

# Target different base branch
core ai task:pr abc123 --base develop
```

---

## ai claude

Claude Code integration commands.

### ai claude run

Run Claude Code in the current directory.

```bash
core ai claude run
```

### ai claude config

Manage Claude configuration.

```bash
core ai claude config
```

---

## Workflow Example

See [Workflow Example](example.md#workflow-example) for a complete task management workflow.

## See Also

- [dev](../dev/) - Multi-repo workflow commands
- [Claude Code documentation](https://claude.ai/code)
