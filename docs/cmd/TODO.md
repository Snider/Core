# CLI Documentation TODO

Commands and subcommands that need documentation.

## Summary

| Package | CLI Commands | Documented | Coverage | Status |
|---------|--------------|------------|----------|--------|
| ai | 10 | 10 | 100% | ✓ Complete |
| build | 4 | 2 | 50% | Needs work |
| ci | 4 | 4 | 100% | ✓ Complete |
| dev | 21 | 21 | 100% | ✓ Complete |
| docs | 3 | 3 | 100% | ✓ Complete |
| doctor | 1 | 1 | 100% | ✓ Complete |
| go | 14 | 14 | 100% | ✓ Complete |
| php | 20 | 1 | 5% | Needs work |
| pkg | 5 | 2 | 40% | Needs work |
| sdk | 3 | 1 | 33% | Needs work |
| setup | 1 | 1 | 100% | ✓ Complete |
| test | 1 | 1 | 100% | ✓ Complete |
| vm | 8 | 2 | 25% | Needs work |

## Missing Documentation

### build

| Command | Status |
|---------|--------|
| build from-path | Missing |
| build pwa | Missing |

### php (subcommand detail pages)

All covered in main index.md but no individual pages:

| Command | Status |
|---------|--------|
| php dev | Missing (covered in index) |
| php logs | Missing (covered in index) |
| php stop | Missing (covered in index) |
| php status | Missing (covered in index) |
| php ssl | Missing (covered in index) |
| php build | Missing (covered in index) |
| php serve | Missing (covered in index) |
| php shell | Missing (covered in index) |
| php test | Missing (covered in index) |
| php fmt | Missing (covered in index) |
| php analyse | Missing (covered in index) |
| php packages link | Missing |
| php packages unlink | Missing |
| php packages update | Missing |
| php packages list | Missing |
| php deploy | Missing |
| php deploy:status | Missing |
| php deploy:rollback | Missing |
| php deploy:list | Missing |

### pkg

| Command | Status |
|---------|--------|
| pkg install | Missing |
| pkg list | Missing |
| pkg update | Missing |
| pkg outdated | Missing |

### sdk

| Command | Status |
|---------|--------|
| sdk diff | Missing |
| sdk validate | Missing |

### vm

| Command | Status |
|---------|--------|
| vm run | Missing |
| vm ps | Missing |
| vm stop | Missing |
| vm logs | Missing |
| vm exec | Missing |
| vm templates show | Missing |
| vm templates vars | Missing |

## Recently Completed

- ✓ ai (all commands documented in index.md)
- ✓ dev (health, commit, push, pull, issues, reviews, ci, impact)
- ✓ go/mod (tidy, download, verify, graph)
- ✓ go/work (sync, init, use)
- ✓ ci (init, changelog, version)

## Documentation Structure

All documented packages follow this structure:
```
/docs/cmd/{package}/
├── index.md          (main command overview)
├── example.md        (optional usage examples)
└── {subcommand}/
    ├── index.md      (subcommand docs)
    └── example.md    (optional examples)
```

## Priority

### High
1. `php` subcommands - Most commonly used
2. `vm` core commands - Important for dev environment

### Medium
3. `pkg` package management commands
4. `sdk` validation commands
5. `build` legacy commands (from-path, pwa)

### Low
6. `vm/templates` subcommands

## Notes

- The `ai` package is new and contains task management commands previously documented under `dev`
- The `dev` index.md includes task command docs that now refer to the `ai` package
- Many packages have comprehensive index.md files but lack individual subcommand pages

Last verified: 2026-01-29
