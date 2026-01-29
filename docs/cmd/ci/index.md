# core ci

Publish releases to GitHub, Docker, npm, Homebrew, and more.

**Safety:** Dry-run by default. Use `--we-are-go-for-launch` to actually publish.

## Subcommands

| Command | Description |
|---------|-------------|
| [init](init/) | Initialize release config |
| [changelog](changelog/) | Generate changelog |
| [version](version/) | Show determined version |

## Usage

```bash
core ci [flags]
```

## Flags

| Flag | Description |
|------|-------------|
| `--we-are-go-for-launch` | Actually publish (default is dry-run) |
| `--version` | Override version |
| `--draft` | Create as draft release |
| `--prerelease` | Mark as prerelease |

## Examples

```bash
# Preview what would be published (safe)
core ci

# Actually publish
core ci --we-are-go-for-launch

# Publish as draft
core ci --we-are-go-for-launch --draft

# Publish as prerelease
core ci --we-are-go-for-launch --prerelease
```

## Workflow

Build and publish are **separated** to prevent accidents:

```bash
# Step 1: Build artifacts
core build
core build sdk

# Step 2: Preview (dry-run by default)
core ci

# Step 3: Publish (explicit flag required)
core ci --we-are-go-for-launch
```

## Publishers

See [Publisher Examples](example.md#publisher-examples) for configuration.

| Type | Target |
|------|--------|
| `github` | GitHub Releases |
| `docker` | Container registries |
| `linuxkit` | LinuxKit images |
| `npm` | npm registry |
| `homebrew` | Homebrew tap |
| `scoop` | Scoop bucket |
| `aur` | Arch User Repository |
| `chocolatey` | Chocolatey |

## Changelog

Auto-generated from conventional commits. See [Changelog Configuration](example.md#changelog-configuration).
