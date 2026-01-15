# Installation

## Prerequisites

### Go 1.22+

```bash
# macOS
brew install go

# Linux
sudo apt install golang-go

# Windows - download from https://go.dev/dl/
```

### Wails v3

```bash
go install github.com/wailsapp/wails/v3/cmd/wails3@latest
```

### Task (Build Automation)

```bash
# macOS
brew install go-task

# Linux
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d

# Windows
choco install go-task
```

## Install Core

```bash
go get github.com/Snider/Core@latest
```

## Verify Installation

```bash
# Check Go
go version

# Check Wails
wails3 version

# Check Task
task --version
```

## IDE Setup

### VS Code

Install the Go extension and configure:

```json
{
  "go.useLanguageServer": true,
  "gopls": {
    "ui.semanticTokens": true
  }
}
```

### GoLand / IntelliJ

Go support is built-in. Enable the Wails plugin for additional features.

## Next Steps

Continue to [Quick Start](quickstart.md) to create your first application.
