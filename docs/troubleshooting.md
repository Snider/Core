# Troubleshooting

Common issues and how to resolve them.

## Installation Issues

### "command not found: core"

**Cause:** Go's bin directory is not in your PATH.

**Fix:**

```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:$(go env GOPATH)/bin"

# Then reload
source ~/.bashrc  # or ~/.zshrc
```

### "go: module github.com/host-uk/core: no matching versions"

**Cause:** Go module proxy hasn't cached the latest version yet.

**Fix:**

```bash
# Bypass proxy
GOPROXY=direct go install github.com/host-uk/core/cmd/core@latest
```

---

## Build Issues

### "no Go files in..."

**Cause:** Core couldn't find a main package to build.

**Fix:**

1. Check you're in the correct directory
2. Ensure `.core/build.yaml` has the correct `main` path:

```yaml
project:
  main: ./cmd/myapp  # Path to main package
```

### "CGO_ENABLED=1 but no C compiler"

**Cause:** Build requires CGO but no C compiler is available.

**Fix:**

```bash
# Option 1: Disable CGO (if not needed)
core build  # Core disables CGO by default

# Option 2: Install a C compiler
# macOS
xcode-select --install

# Ubuntu/Debian
sudo apt install build-essential

# Windows
# Install MinGW or use WSL
```

### Build succeeds but binary doesn't run

**Cause:** Built for wrong architecture.

**Fix:**

```bash
# Check what you built
file dist/myapp-*

# Build for your current platform
core build --targets $(go env GOOS)/$(go env GOARCH)
```

---

## Release Issues

### "dry-run mode, use --we-are-go-for-launch to publish"

**This is expected behaviour.** Core runs in dry-run mode by default for safety.

**To actually publish:**

```bash
core ci --we-are-go-for-launch
```

### "failed to create release: 401 Unauthorized"

**Cause:** GitHub token missing or invalid.

**Fix:**

```bash
# Authenticate with GitHub CLI
gh auth login

# Or set token directly
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx
```

### "no artifacts found in dist/"

**Cause:** You need to build before releasing.

**Fix:**

```bash
# Build first
core build

# Then release
core ci --we-are-go-for-launch
```

### "tag already exists"

**Cause:** Trying to release a version that's already been released.

**Fix:**

1. Update version in your code/config
2. Or delete the existing tag (if intentional):

```bash
git tag -d v1.0.0
git push origin :refs/tags/v1.0.0
```

---

## Multi-Repo Issues

### "repos.yaml not found"

**Cause:** Core can't find the package registry.

**Fix:**

Core looks for `repos.yaml` in:
1. Current directory
2. Parent directories (walking up to root)
3. `~/Code/host-uk/repos.yaml`
4. `~/.config/core/repos.yaml`

Either:
- Run commands from a directory with `repos.yaml`
- Use `--registry /path/to/repos.yaml`
- Run `core setup` to bootstrap a new workspace

### "failed to clone: Permission denied (publickey)"

**Cause:** SSH key not configured for GitHub.

**Fix:**

```bash
# Check SSH connection
ssh -T git@github.com

# If that fails, add your key
ssh-add ~/.ssh/id_ed25519

# Or configure SSH
# See: https://docs.github.com/en/authentication/connecting-to-github-with-ssh
```

### "repository not found" during setup

**Cause:** You don't have access to the repository, or it doesn't exist.

**Fix:**

1. Check you're authenticated: `gh auth status`
2. Verify the repo exists and you have access
3. For private repos, ensure your token has `repo` scope

---

## GitHub Integration Issues

### "gh: command not found"

**Cause:** GitHub CLI not installed.

**Fix:**

```bash
# macOS
brew install gh

# Ubuntu/Debian
sudo apt install gh

# Windows
winget install GitHub.cli

# Then authenticate
gh auth login
```

### "core dev issues" shows nothing

**Possible causes:**

1. No open issues exist
2. Not authenticated with GitHub
3. Not in a directory with `repos.yaml`

**Fix:**

```bash
# Check auth
gh auth status

# Check you're in a workspace
ls repos.yaml

# Show all issues including closed
core dev issues --all
```

---

## PHP Issues

### "frankenphp: command not found"

**Cause:** FrankenPHP not installed.

**Fix:**

```bash
# macOS
brew install frankenphp

# Or use Docker
core php dev --docker
```

### "core php dev" exits immediately

**Cause:** Usually a port conflict or missing dependency.

**Fix:**

```bash
# Check if port 8000 is in use
lsof -i :8000

# Try a different port
core php dev --port 9000

# Check logs for errors
core php logs
```

---

## Performance Issues

### Commands are slow

**Possible causes:**

1. Large number of repositories
2. Network latency to GitHub
3. Go module downloads

**Fix:**

```bash
# For multi-repo commands, use health for quick check
core dev health  # Fast summary

# Instead of
core dev work --status  # Full table (slower)

# Pre-download Go modules
go mod download
```

---

## Getting More Help

### Enable Verbose Output

Most commands support `-v` or `--verbose`:

```bash
core build -v
core go test -v
```

### Check Environment

```bash
core doctor
```

This verifies all required tools are installed and configured.

### Report Issues

If you've found a bug:

1. Check existing issues: https://github.com/host-uk/core/issues
2. Create a new issue with:
   - Core version (`core --version`)
   - OS and architecture (`go env GOOS GOARCH`)
   - Command that failed
   - Full error output

---

## See Also

- [Getting Started](getting-started.md) - Installation and first steps
- [Configuration](configuration.md) - Config file reference
- [doctor](cmd/doctor/) - Environment verification
