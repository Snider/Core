# Glossary

Definitions of terms used throughout Core CLI documentation.

## A

### Artifact
A file produced by a build, typically a binary, archive, or checksum file. Artifacts are stored in the `dist/` directory and published during releases.

## C

### CGO
Go's mechanism for calling C code. Core disables CGO by default (`CGO_ENABLED=0`) to produce statically-linked binaries that don't depend on system libraries.

### Changelog
Automatically generated list of changes between releases, created from conventional commit messages. Configure in `.core/release.yaml`.

### Conventional Commits
A commit message format: `type(scope): description`. Types include `feat`, `fix`, `docs`, `chore`. Core uses this to generate changelogs.

## D

### Dry-run
A mode where commands show what they would do without actually doing it. `core ci` runs in dry-run mode by default for safety.

## F

### Foundation Package
A core package with no dependencies on other packages. Examples: `core-php`, `core-devops`. These form the base of the dependency tree.

### FrankenPHP
A modern PHP application server used by `core php dev`. Combines PHP with Caddy for high-performance serving.

## G

### `gh`
The GitHub CLI tool. Required for commands that interact with GitHub: `core dev issues`, `core dev reviews`, `core dev ci`.

## L

### LinuxKit
A toolkit for building lightweight, immutable Linux distributions. Core can build LinuxKit images via `core build --type linuxkit`.

## M

### Module (Go)
A collection of Go packages with a `go.mod` file. Core's Go commands operate on modules.

### Module (Package)
A host-uk package that depends on foundation packages. Examples: `core-tenant`, `core-admin`. Compare with **Foundation Package** and **Product**.

## P

### Package
An individual repository in the host-uk ecosystem. Packages are defined in `repos.yaml` and managed with `core pkg` commands.

### Package Index
The `repos.yaml` file that lists all packages in a workspace. Contains metadata like dependencies, type, and description.

### Product
A user-facing application package. Examples: `core-bio`, `core-social`. Products depend on foundation and module packages.

### Publisher
A release target configured in `.core/release.yaml`. Types include `github`, `docker`, `npm`, `homebrew`, `linuxkit`.

## R

### Registry (Docker/npm)
A remote repository for container images or npm packages. Core can publish to registries during releases.

### `repos.yaml`
The package index file defining all repositories in a workspace. Used by multi-repo commands like `core dev work`.

## S

### SDK
Software Development Kit. Core can generate API client SDKs from OpenAPI specs via `core build sdk`.

## T

### Target
A build target specified as `os/arch`, e.g., `linux/amd64`, `darwin/arm64`. Use `--targets` flag to specify.

## W

### Wails
A framework for building desktop applications with Go backends and web frontends. Core detects Wails projects and uses appropriate build commands.

### Workspace (Go)
A Go 1.18+ feature for working with multiple modules simultaneously. Managed via `core go work` commands.

### Workspace (Multi-repo)
A directory containing multiple packages from `repos.yaml`. Created via `core setup` and managed with `core dev` commands.

## Symbols

### `.core/`
Directory containing project configuration files:
- `build.yaml` - Build settings
- `release.yaml` - Release targets
- `test.yaml` - Test configuration
- `linuxkit/` - LinuxKit templates

### `--we-are-go-for-launch`
Flag to disable dry-run mode and actually publish a release. Named as a deliberate friction to prevent accidental releases.

---

## See Also

- [Configuration](configuration.md) - Config file reference
- [Getting Started](getting-started.md) - First-time setup
