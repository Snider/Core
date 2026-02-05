# Frequently Asked Questions (FAQ)

Common questions and answers about the Core CLI and Framework.

## General

### What is Core?

Core is a unified CLI and framework for building and managing Go, PHP, and Wails applications. It provides an opinionated set of tools for development, testing, building, and releasing projects within the host-uk ecosystem.

### Is Core a CLI or a Framework?

It is both. The Core Framework (`pkg/core`) is a library for building Go desktop applications with Wails. The Core CLI (`cmd/core`) is the tool you use to manage projects, run tests, build binaries, and handle multi-repository workspaces.

---

## Installation

### How do I install the Core CLI?

The recommended way is via Go:

```bash
go install github.com/host-uk/core/cmd/core@latest
```

Ensure your Go bin directory is in your PATH. See [Getting Started](getting-started.md) for more options.

### I get "command not found: core" after installation.

This usually means your Go bin directory is not in your system's PATH. Add it by adding this to your shell profile (`.bashrc`, `.zshrc`, etc.):

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

---

## Usage

### Why does `core ci` not publish anything by default?

Core is designed to be **safe by default**. `core ci` runs in dry-run mode to show you what would be published. To actually publish a release, you must use the `--we-are-go-for-launch` flag:

```bash
core ci --we-are-go-for-launch
```

### How do I run tests for only one package?

You can pass standard Go test flags to `core go test`:

```bash
core go test ./pkg/my-package
```

### What is `core doctor` for?

`core doctor` checks your development environment to ensure all required tools (Go, Git, Docker, etc.) are installed and correctly configured. It's the first thing you should run if something isn't working.

---

## Configuration

### Where is Core's configuration stored?

- **Project-specific**: In the `.core/` directory within your project root.
- **Global**: In `~/.core/` or as defined by `CORE_CONFIG`.
- **Registry**: The `repos.yaml` file defines the multi-repo workspace.

### How do I change the build targets?

You can specify targets in `.core/release.yaml` or use the `--targets` flag with the `core build` command:

```bash
core build --targets linux/amd64,darwin/arm64
```

---

## Workspaces and Registry

### What is a "workspace" in Core?

In the context of the CLI, a workspace is a directory containing multiple repositories defined in a `repos.yaml` file. The `core dev` commands allow you to manage status, commits, and synchronization across all repositories in the workspace at once.

### What is `repos.yaml`?

`repos.yaml` is the "registry" for your workspace. It lists the repositories, their types (foundation, module, product), and their dependencies. Core uses this file to know which repositories to clone during `core setup`.

---

## See Also

- [Getting Started](getting-started.md) - Installation and first steps
- [User Guide](user-guide.md) - Detailed usage information
- [Troubleshooting](troubleshooting.md) - Solving common issues
