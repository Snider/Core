# core doctor

Check your development environment for required tools and configuration.

## Usage

```bash
core doctor
```

## What It Checks

### Required Tools

| Tool | Purpose |
|------|---------|
| `git` | Version control |
| `go` | Go compiler |
| `gh` | GitHub CLI |

### Optional Tools

| Tool | Purpose |
|------|---------|
| `node` | Node.js runtime |
| `docker` | Container runtime |
| `wails` | Desktop app framework |
| `qemu` | VM runtime for LinuxKit |
| `gpg` | Code signing |
| `codesign` | macOS signing (macOS only) |

### Configuration

- Git user name and email
- GitHub CLI authentication
- Go workspace setup

## Output

```
Core Doctor
===========

Required:
  [OK] git 2.43.0
  [OK] go 1.23.0
  [OK] gh 2.40.0

Optional:
  [OK] node 20.10.0
  [OK] docker 24.0.7
  [--] wails (not installed)
  [OK] qemu 8.2.0
  [OK] gpg 2.4.3
  [OK] codesign (available)

Configuration:
  [OK] git user.name: Your Name
  [OK] git user.email: you@example.com
  [OK] gh auth status: Logged in

All checks passed!
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | All required checks passed |
| 1 | One or more required checks failed |

## See Also

- [setup command](setup.md) - Clone repos from registry
- [dev install](dev.md) - Install development environment
