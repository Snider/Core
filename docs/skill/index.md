# Claude Code Skill

The `core` skill teaches Claude Code how to use the Core CLI effectively.

## Installation

```bash
curl -fsSL https://raw.githubusercontent.com/host-uk/core/main/.claude/skills/core/install.sh | bash
```

Or if you have the repo cloned:

```bash
./.claude/skills/core/install.sh
```

## What it does

Once installed, Claude Code will:

- Auto-invoke when working in host-uk repositories
- Use `core` commands instead of raw `go`/`php`/`git` commands
- Follow the correct patterns for testing, building, and releasing

## Manual invocation

Type `/core` in Claude Code to invoke the skill manually.

## Updating

Re-run the install command to update to the latest version.

## Location

Skills are installed to `~/.claude/skills/core/SKILL.md`.
