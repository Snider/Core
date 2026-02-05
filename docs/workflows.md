# Workflows

Common end-to-end workflows for Core CLI.

## Go Project: Build and Release

Complete workflow from code to GitHub release.

```bash
# 1. Run tests
core go test

# 2. Check coverage
core go cov --threshold 80

# 3. Format and lint
core go fmt --fix
core go lint

# 4. Build for all platforms
core build --targets linux/amd64,linux/arm64,darwin/arm64,windows/amd64

# 5. Preview release (dry-run)
core ci

# 6. Publish
core ci --we-are-go-for-launch
```

**Output structure:**

```
dist/
├── myapp-darwin-arm64.tar.gz
├── myapp-linux-amd64.tar.gz
├── myapp-linux-arm64.tar.gz
├── myapp-windows-amd64.zip
└── CHECKSUMS.txt
```

---

## PHP Project: Development to Deployment

Local development through to production deployment.

```bash
# 1. Start development environment
core php dev

# 2. Run tests (in another terminal)
core php test --parallel

# 3. Check code quality
core php fmt --fix
core php analyse

# 4. Deploy to staging
core php deploy --staging --wait

# 5. Verify staging
# (manual testing)

# 6. Deploy to production
core php deploy --wait

# 7. Monitor
core php deploy:status
```

**Rollback if needed:**

```bash
core php deploy:rollback
```

---

## Multi-Repo: Daily Workflow

Working across the host-uk monorepo.

### Morning: Sync Everything

```bash
# Quick health check
core dev health

# Pull all repos that are behind
core dev pull --all

# Check for issues assigned to you
core dev issues --assignee @me
```

### During Development

```bash
# Work on code...

# Check status across all repos
core dev work --status

# Commit changes (Claude-assisted messages)
core dev commit

# Push when ready
core dev push
```

### End of Day

```bash
# Full workflow: status → commit → push
core dev work

# Check CI status
core dev ci

# Review any failed builds
core dev ci --failed
```

---

## New Developer: Environment Setup

First-time setup for a new team member.

```bash
# 1. Verify prerequisites
core doctor

# 2. Create workspace directory
mkdir ~/Code/host-uk && cd ~/Code/host-uk

# 3. Bootstrap workspace (interactive)
core setup

# 4. Select packages in wizard
# Use arrow keys, space to select, enter to confirm

# 5. Verify setup
core dev health

# 6. Start working
core dev work --status
```

---

## CI Pipeline: Automated Build

Example GitHub Actions workflow.

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Install Core
        run: go install github.com/host-uk/core/cmd/core@latest

      - name: Build
        run: core build --ci

      - name: Release
        run: core ci --we-are-go-for-launch
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## SDK Generation: API Client Updates

Generate SDK clients when API changes.

```bash
# 1. Validate OpenAPI spec
core sdk validate

# 2. Check for breaking changes
core sdk diff --base v1.0.0

# 3. Generate SDKs
core build sdk

# 4. Review generated code
git diff

# 5. Commit if satisfied
git add -A && git commit -m "chore: regenerate SDK clients"
```

---

## Dependency Update: Cross-Repo Change

When updating a shared package like `core-php`.

```bash
# 1. Make changes in core-php
cd ~/Code/host-uk/core-php
# ... edit code ...

# 2. Run tests
core go test  # or core php test

# 3. Check what depends on core-php
core dev impact core-php

# Output:
# core-tenant (direct)
# core-admin (via core-tenant)
# core-api (direct)
# ...

# 4. Commit core-php changes
core dev commit

# 5. Update dependent packages
cd ~/Code/host-uk
for pkg in core-tenant core-admin core-api; do
  cd $pkg
  composer update host-uk/core-php
  core php test
  cd ..
done

# 6. Commit all updates
core dev work
```

---

## Hotfix: Emergency Production Fix

Fast path for critical fixes.

```bash
# 1. Create hotfix branch
git checkout -b hotfix/critical-bug main

# 2. Make fix
# ... edit code ...

# 3. Test
core go test --run TestCriticalPath

# 4. Build
core build

# 5. Preview release
core ci --prerelease

# 6. Publish hotfix
core ci --we-are-go-for-launch --prerelease

# 7. Merge back to main
git checkout main
git merge hotfix/critical-bug
git push
```

---

## Documentation: Sync Across Repos

Keep documentation synchronised.

```bash
# 1. List all docs
core docs list

# 2. Sync to central location
core docs sync --output ./docs-site

# 3. Review changes
git diff docs-site/

# 4. Commit
git add docs-site/
git commit -m "docs: sync from packages"
```

---

## Troubleshooting: Failed Build

When a build fails.

```bash
# 1. Check environment
core doctor

# 2. Clean previous artifacts
rm -rf dist/

# 3. Verbose build
core build -v

# 4. If Go-specific issues
core go mod tidy
core go mod verify

# 5. Check for test failures
core go test -v

# 6. Review configuration
cat .core/build.yaml
```

---

## See Also

- [Getting Started](getting-started.md) - First-time setup
- [Troubleshooting](troubleshooting.md) - When things go wrong
- [Configuration](configuration.md) - Config file reference
