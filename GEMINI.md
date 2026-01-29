# GEMINI.md

This file provides guidance for agentic interactions within this repository, specifically for Gemini and other MCP-compliant agents.

## Agentic Context & MCP

This project is built with an **Agentic** design philosophy. It is not exclusive to any single LLM provider (like Claude).

- **MCP Support**: The system is designed to leverage the Model Context Protocol (MCP) to provide rich context and tools to agents.
- **Developer Image**: You are running within a standardized developer image (`host-uk/core` dev environment), ensuring consistent tooling and configuration.

## Core CLI (Agent Interface)

The `core` command is the primary interface for agents to manage the project. Agents should **always** prefer `core` commands over raw shell commands (like `go test`, `php artisan`, etc.).

### Key Commands for Agents

| Task | Command | Notes |
|------|---------|-------|
| **Health Check** | `core doctor` | Verify tools and environment |
| **Repo Status** | `core dev health` | Quick summary of all repos |
| **Work Status** | `core dev work --status` | Detailed dirty/ahead status |
| **Run Tests** | `core go test` | Run Go tests with correct flags |
| **Coverage** | `core go cov` | Generate coverage report |
| **Build** | `core build` | Build the project safely |
| **Search Code** | `core pkg search` | Find packages/repos |

## Project Architecture

Core is a Web3 Framework written in Go using Wails v3.

### Core Framework

- **Services**: Managed via dependency injection (`ServiceFor[T]()`).
- **Lifecycle**: `OnStartup` and `OnShutdown` hooks.
- **IPC**: Message-passing system for service communication.

### Development Workflow

1.  **Check State**: `core dev work --status`
2.  **Make Changes**: Modify code, add tests.
3.  **Verify**: `core go test` (or `core php test` for PHP components).
4.  **Commit**: `core dev commit` (or standard git if automated).
5.  **Push**: `core dev push` (handles multiple repos).

## Testing Standards

- **Suffix Pattern**:
    - `_Good`: Happy path
    - `_Bad`: Expected errors
    - `_Ugly`: Edge cases/panics

## Go Workspace

The project uses Go workspaces (`go.work`). Always run `core go work sync` after modifying modules.
