# CLI Documentation TODO

Commands and subcommands that need documentation.

## Missing Documentation

| Command | Subcommand | Status |
|---------|------------|--------|
| dev | health | Missing |
| dev | commit | Missing |
| dev | push | Missing |
| dev | pull | Missing |
| dev | issues | Missing |
| dev | reviews | Missing |
| dev | ci | Missing |
| dev | impact | Missing |
| dev | api | Missing |
| dev | api sync | Missing |
| dev | sync | Missing |
| dev | tasks | Missing |
| dev | task | Missing |
| dev | task:update | Missing |
| dev | task:complete | Missing |
| dev | task:commit | Missing |
| dev | task:pr | Missing |
| dev | install | Missing |
| dev | boot | Missing |
| dev | stop | Missing |
| dev | status | Missing |
| dev | shell | Missing |
| dev | serve | Missing |
| dev | test | Missing |
| dev | claude | Missing |
| dev | update | Missing |
| go | work sync | Missing |
| go | work init | Missing |
| go | work use | Missing |
| go | mod tidy | Missing |
| go | mod download | Missing |
| go | mod verify | Missing |
| go | mod graph | Missing |
| php | dev | Missing |
| php | logs | Missing |
| php | stop | Missing |
| php | status | Missing |
| php | ssl | Missing |
| php | build | Missing |
| php | serve | Missing |
| php | shell | Missing |
| php | test | Missing |
| php | fmt | Missing |
| php | analyse | Missing |
| php | packages | Missing |
| php | packages link | Missing |
| php | packages unlink | Missing |
| php | packages update | Missing |
| php | packages list | Missing |
| php | deploy | Missing |
| php | deploy:status | Missing |
| php | deploy:rollback | Missing |
| php | deploy:list | Missing |
| build | from-path | Missing |
| build | pwa | Missing |
| ci | - | Missing (only has subcommands) |
| sdk | diff | Missing |
| sdk | validate | Missing |
| pkg | install | Missing |
| pkg | list | Missing |
| pkg | update | Missing |
| pkg | outdated | Missing |
| vm | run | Missing |
| vm | ps | Missing |
| vm | stop | Missing |
| vm | logs | Missing |
| vm | exec | Missing |
| vm | templates show | Missing |
| vm | templates vars | Missing |
| docs | sync | Missing |
| docs | list | Missing |

## Needs Update

| Command | Issue |
|---------|-------|
| build/sdk | Documentation exists but command has been moved to `build sdk` |
| go/work | Index exists but subcommands (sync, init, use) are undocumented |
| go/mod | Index exists but subcommands (tidy, download, verify, graph) are undocumented |
| vm/templates | Index exists but subcommands (show, vars) are undocumented |
| pkg/search | Index exists but may need updating with new flags |

## Documentation Structure Notes

The following commands have complete documentation:
- `test` - /Users/snider/Code/host-uk/core/docs/cmd/test/index.md
- `doctor` - /Users/snider/Code/host-uk/core/docs/cmd/doctor/index.md
- `setup` - /Users/snider/Code/host-uk/core/docs/cmd/setup/index.md
- `dev/work` - /Users/snider/Code/host-uk/core/docs/cmd/dev/work/index.md
- `dev` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/dev/index.md
- `go` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/go/index.md
- `go/test` - /Users/snider/Code/host-uk/core/docs/cmd/go/test/index.md
- `go/cov` - /Users/snider/Code/host-uk/core/docs/cmd/go/cov/index.md
- `go/fmt` - /Users/snider/Code/host-uk/core/docs/cmd/go/fmt/index.md
- `go/lint` - /Users/snider/Code/host-uk/core/docs/cmd/go/lint/index.md
- `go/install` - /Users/snider/Code/host-uk/core/docs/cmd/go/install/index.md
- `go/mod` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/go/mod/index.md
- `go/work` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/go/work/index.md
- `ci/init` - /Users/snider/Code/host-uk/core/docs/cmd/ci/init/index.md
- `ci/changelog` - /Users/snider/Code/host-uk/core/docs/cmd/ci/changelog/index.md
- `ci/version` - /Users/snider/Code/host-uk/core/docs/cmd/ci/version/index.md
- `build` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/build/index.md
- `build/sdk` - /Users/snider/Code/host-uk/core/docs/cmd/build/sdk/index.md
- `sdk` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/sdk/index.md
- `pkg` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/pkg/index.md
- `pkg/search` - /Users/snider/Code/host-uk/core/docs/cmd/pkg/search/index.md
- `vm` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/vm/index.md
- `vm/templates` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/vm/templates/index.md
- `php` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/php/index.md
- `docs` (parent) - /Users/snider/Code/host-uk/core/docs/cmd/docs/index.md

## Priority Recommendations

High priority (commonly used commands):
1. `dev` subcommands (health, commit, push, pull, issues, reviews, ci, impact)
2. `php` subcommands (dev, build, test, fmt, packages)
3. `go/mod` subcommands (tidy, download, verify)
4. `go/work` subcommands (sync, init, use)
5. `vm` core commands (run, ps, stop, logs)

Medium priority:
1. `dev` task management commands
2. `dev` dev environment commands (install, boot, stop, status, shell)
3. `sdk` validation commands
4. `pkg` package management commands
5. `php` deployment commands
6. `build` alternative builders (from-path, pwa)

Low priority:
1. `dev` advanced commands (api sync, claude, update)
2. `vm/templates` subcommands
3. `docs` management commands

## Issues Found

- There appears to be duplicate documentation under `/Users/snider/Code/host-uk/core/docs/cmd/docs/cmd/` which mirrors the main command structure. This should be cleaned up.
- The `ci` parent command has no index.md, only subcommand documentation exists.
- Many parent commands (dev, go, php, etc.) have good overview documentation but are missing subcommand details.

Last verified: 2026-01-29
