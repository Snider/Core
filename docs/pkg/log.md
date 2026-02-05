# Log Retention Policy

The `log` package provides structured logging with automatic log rotation and retention management.

## Retention Policy

By default, the following log retention policy is applied when log rotation is enabled:

- **Max Size**: 100 MB per log file.
- **Max Backups**: 5 old log files are retained.
- **Max Age**: 28 days. Old log files beyond this age are automatically deleted. (Set to -1 to disable age-based retention).
- **Compression**: Rotated log files can be compressed (future feature).

## Configuration

Logging can be configured using the `log.Options` struct. To enable log rotation to a file, provide a `RotationOptions` struct. If both `Output` and `Rotation` are provided, `Rotation` takes precedence and `Output` is ignored.

### Standalone Usage

```go
logger := log.New(log.Options{
    Level: log.LevelInfo,
    Rotation: &log.RotationOptions{
        Filename:   "app.log",
        MaxSize:    100, // MB
        MaxBackups: 5,
        MaxAge:     28, // days
    },
})

logger.Info("application started")
```

### Framework Integration

When using the Core framework, logging is usually configured during application initialization:

```go
app := core.New(
    framework.WithName("my-app", log.NewService(log.Options{
        Level: log.LevelDebug,
        Rotation: &log.RotationOptions{
            Filename: "/var/log/my-app.log",
        },
    })),
)
```

## How It Works

1.  **Rotation**: When the current log file exceeds `MaxSize`, it is rotated. The current file is renamed to `filename.1`, `filename.1` is renamed to `filename.2`, and so on.
2.  **Retention**:
    -   Files beyond `MaxBackups` are automatically deleted during rotation.
    -   Files older than `MaxAge` days are automatically deleted during the cleanup process.
3.  **Appends**: When an application restarts, it appends to the existing log file instead of truncating it.
