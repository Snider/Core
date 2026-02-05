# BugSETI

**Distributed Bug Fixing - like SETI@home but for code**

BugSETI is a system tray application that helps developers contribute to open source by fixing bugs in their spare CPU cycles. It fetches issues from GitHub repositories, prepares context using AI, and guides you through the fix-and-submit workflow.

## Features

- **System Tray Integration**: Runs quietly in the background, ready when you are
- **Issue Queue**: Automatically fetches and queues issues from configured repositories
- **AI Context Seeding**: Prepares relevant code context for each issue using pattern matching
- **Workbench UI**: Full-featured interface for reviewing issues and submitting fixes
- **Automated PR Submission**: Streamlined workflow from fix to pull request
- **Stats & Leaderboard**: Track your contributions and compete with the community

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/host-uk/core.git
cd core

# Build BugSETI
task bugseti:build

# The binary will be in build/bin/bugseti
```

### Prerequisites

- Go 1.25 or later
- Node.js 18+ and npm (for frontend)
- GitHub CLI (`gh`) authenticated
- Chrome/Chromium (optional, for webview features)

## Configuration

On first launch, BugSETI will show an onboarding wizard to configure:

1. **GitHub Token**: For fetching issues and submitting PRs
2. **Repositories**: Which repos to fetch issues from
3. **Filters**: Issue labels, difficulty levels, languages
4. **Notifications**: How to alert you about new issues

### Configuration File

Settings are stored in `~/.config/bugseti/config.json`:

```json
{
  "github_token": "ghp_...",
  "repositories": [
    "host-uk/core",
    "example/repo"
  ],
  "filters": {
    "labels": ["good first issue", "help wanted", "bug"],
    "languages": ["go", "typescript"],
    "max_age_days": 30
  },
  "notifications": {
    "enabled": true,
    "sound": true
  },
  "fetch_interval_minutes": 30
}
```

## Usage

### Starting BugSETI

```bash
# Run the application
./bugseti

# Or use task runner
task bugseti:run
```

The app will appear in your system tray. Click the icon to see the quick menu or open the workbench.

### Workflow

1. **Browse Issues**: Click the tray icon to see available issues
2. **Select an Issue**: Choose one to work on from the queue
3. **Review Context**: BugSETI shows relevant files and patterns
4. **Fix the Bug**: Make your changes in your preferred editor
5. **Submit PR**: Use the workbench to create and submit your pull request

### Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Ctrl+Shift+B` | Open workbench |
| `Ctrl+Shift+N` | Next issue |
| `Ctrl+Shift+S` | Submit PR |

## Architecture

```
cmd/bugseti/
  main.go          # Application entry point
  tray.go          # System tray service
  icons/           # Tray icons (light/dark/template)
  frontend/        # Angular frontend
    src/
      app/
        tray/      # Tray panel component
        workbench/ # Main workbench
        settings/  # Settings panel
        onboarding/ # First-run wizard

internal/bugseti/
  config.go        # Configuration service
  fetcher.go       # GitHub issue fetcher
  queue.go         # Issue queue management
  seeder.go        # Context seeding via AI
  submit.go        # PR submission
  notify.go        # Notification service
  stats.go         # Statistics tracking
```

## Contributing

We welcome contributions! Here's how to get involved:

### Development Setup

```bash
# Install dependencies
cd cmd/bugseti/frontend
npm install

# Run in development mode
task bugseti:dev
```

### Running Tests

```bash
# Go tests
go test ./cmd/bugseti/... ./internal/bugseti/...

# Frontend tests
cd cmd/bugseti/frontend
npm test
```

### Submitting Changes

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and add tests
4. Run the test suite: `task test`
5. Submit a pull request

### Code Style

- Go: Follow standard Go conventions, run `go fmt`
- TypeScript/Angular: Follow Angular style guide
- Commits: Use conventional commit messages

## Roadmap

- [ ] Auto-update mechanism
- [ ] Team/organization support
- [ ] Integration with more issue trackers (GitLab, Jira)
- [ ] AI-assisted code review
- [ ] Mobile companion app

## License

MIT License - see [LICENSE](../../LICENSE) for details.

## Acknowledgments

- Inspired by SETI@home and distributed computing projects
- Built with [Wails v3](https://wails.io/) for native desktop integration
- Uses [Angular](https://angular.io/) for the frontend

---

**Happy Bug Hunting!**
