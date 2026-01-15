# Copilot Instructions for Help Module

## Project Overview

This repository contains the `help` module, which was formerly part of the `Snider/Core` framework. The module provides assistance and documentation functionality for applications, allowing them to embed and display help documentation.

The project consists of three main components:

1. **Go Library**: A Go module that provides a help service for displaying documentation
2. **Documentation**: MkDocs-based documentation that can be embedded in applications
3. **Angular UI Component**: A custom HTML element built with Angular Elements for the help interface

## Technology Stack

- **Go Library**: Go 1.25 with embedded file systems
- **Documentation**: MkDocs Material theme with Python
- **UI Component**: Angular 20.3+ with Angular Elements
- **Build Tools**: 
  - GoReleaser for Go library releases
  - Angular CLI for UI builds
  - Task (Taskfile) for documentation builds
- **Package Management**: 
  - Go modules for Go dependencies
  - pip for Python/MkDocs dependencies
  - npm for Node.js/Angular dependencies

## Project Structure

```
.
├── help.go                # Main Go library implementation
├── help_test.go           # Go tests
├── go.mod                 # Go module definition
├── examples/              # Example applications
│   ├── show_help/         # Example: showing help
│   └── show_at/           # Example: showing help at specific anchor
├── src/                   # MkDocs documentation source
│   ├── index.md           # Main documentation page
│   ├── images/            # Documentation images
│   └── stylesheets/       # Custom CSS
├── public/                # Built documentation (embedded in Go binary)
├── ui/                    # Angular help UI component
│   ├── src/               # Angular source code
│   ├── public/            # Static assets
│   └── package.json       # Node.js dependencies
├── mkdocs.yml             # MkDocs configuration
├── taskfile.dist.yml      # Task runner configuration
└── .github/               # GitHub workflows and configurations
```

## Development Setup

### Prerequisites

- Go 1.25 or later
- Python 3 with pip
- Node.js and npm
- Task (task runner) - optional but recommended

### Initial Setup

1. Install Go dependencies:
   ```bash
   go mod tidy
   ```

2. Install documentation dependencies:
   ```bash
   pip install -r requirements.txt
   ```

3. Install UI dependencies (if working on the Angular component):
   ```bash
   cd ui
   npm install
   ```

### Running the Documentation Server

Using Task (recommended):
```bash
task dev
```

Or using MkDocs directly:
```bash
mkdocs serve
```

This starts a live-reloading development server for the documentation at `http://localhost:8000`.

## Building

### Build Documentation

The documentation is built into the `public` directory and embedded in the Go binary:

```bash
task build
# or
mkdocs build --clean -d public
```

### Build UI Component

```bash
cd ui
npm run build
```

This creates the Angular custom element in the `dist` directory.

## Testing

### Go Tests

Run all Go tests with coverage:
```bash
go test -v -coverprofile=coverage.out ./...
```

The test suite includes:
- Service initialization tests
- Display functionality tests
- Error handling tests
- Custom source and asset loading tests

### UI Tests

Run Angular tests:
```bash
cd ui
npm test
```

## Code Style and Conventions

### Go Code

- Follow standard Go conventions and formatting (use `gofmt` or `go fmt`)
- Use meaningful variable and function names
- Keep functions focused and single-purpose
- Follow the existing patterns for interfaces (Logger, App, Core, Display, Help)
- Use context.Context for service lifecycle management
- Embed static assets using `//go:embed` directives

### Documentation (Markdown)

- Follow MkDocs Material conventions
- Use clear, concise language
- Include code examples where appropriate
- Keep line length reasonable for readability
- Use proper heading hierarchy

### Angular/TypeScript Code

- Follow Angular style guide
- Use Prettier for code formatting (configuration in `ui/package.json`)
- Settings:
  - Print width: 100 characters
  - Single quotes preferred
  - Angular parser for HTML files

### General Guidelines

- Write clear, self-documenting code
- Add comments for complex logic or non-obvious decisions
- Keep commits atomic and well-described
- Follow the existing code patterns in the repository

## API Overview

### Main Interfaces

- **Help**: Main interface for the help service with `Show()`, `ShowAt(anchor string)`, and `ServiceStartup(ctx context.Context)` methods
- **Core**: Application core interface providing ACTION dispatch and App access
- **App**: Application interface providing Logger access
- **Logger**: Logging interface with Info and Error methods
- **Display**: Marker interface for display service dependency checking

### Options

The `Options` struct allows configuration:
- `Source`: Path to custom static site directory
- `Assets`: Custom `fs.FS` for documentation assets (can be `embed.FS` or any `fs.FS` implementation)

### Usage Pattern

```go
// Initialize with default embedded documentation
helpService, err := help.New(help.Options{})

// Or with custom embedded assets
helpService, err := help.New(help.Options{
    Assets: myDocs,
})

// Or with custom directory source
helpService, err := help.New(help.Options{
    Source: "path/to/docs",
})

// Start the service
err = helpService.ServiceStartup(ctx)

// Show help
err = helpService.Show()

// Show help at specific location
err = helpService.ShowAt("ui/how/settings#resetPassword")
```

## Important Notes for AI Assistants

1. **Minimal Changes**: Make the smallest possible changes to achieve the goal
2. **No Breaking Changes**: Don't modify working code unless necessary
3. **Testing**: Always run `go test ./...` before and after changes to ensure nothing breaks
4. **Build Verification**: Ensure `go build` succeeds after making changes
5. **Dependencies**: Check for security vulnerabilities before adding new dependencies using `govulncheck` or similar security scanning tools
6. **Module Name**: The Go module is `github.com/Snider/help` - do not change this
7. **Embedded Assets**: The `public` directory is embedded in the Go binary via `//go:embed all:public/*`
8. **Documentation**: When changing documentation, rebuild with `task build` or `mkdocs build`

## CI/CD Workflows

- **Go CI** (`.github/workflows/go.yml`): Runs tests and uploads coverage on push/PR to main
- **Release** (`.github/workflows/release.yml`): Uses GoReleaser to build and publish releases

## Contributing

When contributing to this repository:

1. Run tests before making changes: `go test ./...`
2. Make minimal, focused changes
3. Run tests after changes to verify nothing broke
4. Follow the existing code style
5. Update documentation if the API changes
6. Ensure all CI checks pass
7. Keep commits atomic with clear messages

## Common Tasks

- **Run tests**: `go test -v ./...`
- **Run tests with coverage**: `go test -v -coverprofile=coverage.out ./...`
- **Build documentation**: `task build` or `mkdocs build --clean -d public`
- **Serve documentation**: `task dev` or `mkdocs serve`
- **Format Go code**: `go fmt ./...`
- **Tidy dependencies**: `go mod tidy`
