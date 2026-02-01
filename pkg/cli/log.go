package cli

import (
	"github.com/host-uk/core/pkg/framework"
	"github.com/host-uk/core/pkg/log"
)

// LogLevel aliases for backwards compatibility.
type LogLevel = log.Level

// Log level constants aliased from the log package.
const (
	// LogLevelQuiet suppresses all output.
	LogLevelQuiet = log.LevelQuiet
	// LogLevelError shows only error messages.
	LogLevelError = log.LevelError
	// LogLevelWarn shows warnings and errors.
	LogLevelWarn = log.LevelWarn
	// LogLevelInfo shows info, warnings, and errors.
	LogLevelInfo = log.LevelInfo
	// LogLevelDebug shows all messages including debug.
	LogLevelDebug = log.LevelDebug
)

// LogService wraps log.Service with CLI styling.
type LogService struct {
	*log.Service
}

// LogOptions configures the log service.
type LogOptions = log.Options

// NewLogService creates a log service factory with CLI styling.
func NewLogService(opts LogOptions) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		// Create the underlying service
		factory := log.NewService(opts)
		svc, err := factory(c)
		if err != nil {
			return nil, err
		}

		logSvc := svc.(*log.Service)

		// Apply CLI styles
		logSvc.StyleTimestamp = func(s string) string { return DimStyle.Render(s) }
		logSvc.StyleDebug = func(s string) string { return DimStyle.Render(s) }
		logSvc.StyleInfo = func(s string) string { return InfoStyle.Render(s) }
		logSvc.StyleWarn = func(s string) string { return WarningStyle.Render(s) }
		logSvc.StyleError = func(s string) string { return ErrorStyle.Render(s) }

		return &LogService{Service: logSvc}, nil
	}
}

// --- Package-level convenience ---

// Log returns the CLI's log service, or nil if not available.
func Log() *LogService {
	if instance == nil {
		return nil
	}
	svc, err := framework.ServiceFor[*LogService](instance.core, "log")
	if err != nil {
		return nil
	}
	return svc
}

// LogDebug logs a debug message if log service is available.
func LogDebug(msg string) {
	if l := Log(); l != nil {
		l.Debug(msg)
	}
}

// LogInfo logs an info message if log service is available.
func LogInfo(msg string) {
	if l := Log(); l != nil {
		l.Info(msg)
	}
}

// LogWarn logs a warning message if log service is available.
func LogWarn(msg string) {
	if l := Log(); l != nil {
		l.Warn(msg)
	}
}

// LogError logs an error message if log service is available.
func LogError(msg string) {
	if l := Log(); l != nil {
		l.Error(msg)
	}
}
