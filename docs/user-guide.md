# User Guide

This guide provides a comprehensive overview of how to use the Core CLI to manage your development workflow.

## Key Concepts

### Projects
A Project is a single repository containing code (Go, PHP, or Wails). Core helps you test, build, and release these projects using a consistent set of commands.

### Workspaces
A Workspace is a collection of related projects. Core is designed to work across multiple repositories, allowing you to perform actions (like checking status or committing changes) on all of them at once.

### Registry (`repos.yaml`)
The Registry is a configuration file that defines the repositories in your workspace. It includes information about where they are located on GitHub, their dependencies, and their purpose.

---

## Daily Workflow

### Working with a Single Project

For a typical day-to-day development on a single project:

1. **Verify your environment**:
   ```bash
   core doctor
   ```
2. **Run tests while you work**:
   ```bash
   core go test
   ```
3. **Keep code clean**:
   ```bash
   core go fmt --fix
   core go lint
   ```
4. **Build and preview**:
   ```bash
   core build
   ```

### Working with Multiple Repositories

If you are working across many repositories in a workspace:

1. **Check status of all repos**:
   ```bash
   core dev work --status
   ```
2. **Sync all changes**:
   ```bash
   core dev pull --all
   ```
3. **Commit and push everything**:
   ```bash
   core dev work
   ```

---

## Building and Releasing

Core separates the building of artifacts from the releasing of those artifacts.

### 1. Build
The `core build` command detects your project type and builds binaries for your configured targets. Artifacts are placed in the `dist/` directory.

### 2. Preview Release
Use `core ci` to see a summary of what would be included in a release (changelog, artifacts, etc.). This is a dry-run by default.

### 3. Publish Release
When you are ready to publish to GitHub:
```bash
core ci --we-are-go-for-launch
```

---

## PHP and Laravel Development

Core provides a unified development server for Laravel projects that orchestrates several services:

```bash
core php dev
```
This starts FrankenPHP, Vite, Horizon, Reverb, and Redis as configured in your `.core/php.yaml`.

---

## Common Workflows

For detailed examples of common end-to-end workflows, see the [Workflows](workflows.md) page.

---

## Getting More Help

- Use the `--help` flag with any command: `core build --help`
- Check the [FAQ](faq.md) for common questions.
- If you run into trouble, see the [Troubleshooting Guide](troubleshooting.md).
