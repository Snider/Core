# Docs Sync Setup - Next Steps

After moving repo to `~/Code/host-uk/core`:

## 1. Add to repos.yaml

Add this to `/Users/snider/Code/host-uk/repos.yaml` under `repos:`:

```yaml
  # CLI (Go)
  core:
    type: foundation
    description: Core CLI - build, release, deploy for Go/Wails/PHP/containers
    docs: true
    ci: github-actions
```

## 2. Test docs sync

```bash
cd ~/Code/host-uk
core docs list          # Should show "core" with docs
core docs sync --dry-run  # Preview what syncs
```

## 3. Add CLI section to VitePress (core-php)

Edit `core-php/docs/.vitepress/config.js`:
- Add `/cli/` to nav
- Add sidebar for CLI commands

## 4. Sync and verify

```bash
core docs sync --output ../core-php/docs/cli
```

---

Current state:
- CLI docs written in `docs/cmd/*.md` (12 files)
- `docs/index.md` updated with command table
- All committed to git
