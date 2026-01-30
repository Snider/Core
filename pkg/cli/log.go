package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/host-uk/core/pkg/framework"
)

// LogLevel defines logging verbosity.
type LogLevel int

const (
	LogLevelQuiet LogLevel = iota
	LogLevelError
	LogLevelWarn
	LogLevelInfo
	LogLevelDebug
)

// LogService provides structured logging for the CLI.
type LogService struct {
	*framework.ServiceRuntime[LogOptions]
	mu     sync.RWMutex
	level  LogLevel
	output io.Writer
}

// LogOptions configures the log service.
type LogOptions struct {
	Level  LogLevel
	Output io.Writer // defaults to os.Stderr
}

// NewLogService creates a log service factory.
func NewLogService(opts LogOptions) func(*framework.Core) (any, error) {
	return func(c *framework.Core) (any, error) {
		output := opts.Output
		if output == nil {
			output = os.Stderr
		}

		return &LogService{
			ServiceRuntime: framework.NewServiceRuntime(c, opts),
			level:          opts.Level,
			output:         output,
		}, nil
	}
}

// OnStartup registers query handlers.
func (s *LogService) OnStartup(ctx context.Context) error {
	s.Core().RegisterQuery(s.handleQuery)
	s.Core().RegisterTask(s.handleTask)
	return nil
}

// Queries and tasks for log service

// QueryLogLevel returns the current log level.
type QueryLogLevel struct{}

// TaskSetLogLevel changes the log level.
type TaskSetLogLevel struct {
	Level LogLevel
}

func (s *LogService) handleQuery(c *framework.Core, q framework.Query) (any, bool, error) {
	switch q.(type) {
	case QueryLogLevel:
		s.mu.RLock()
		defer s.mu.RUnlock()
		return s.level, true, nil
	}
	return nil, false, nil
}

func (s *LogService) handleTask(c *framework.Core, t framework.Task) (any, bool, error) {
	switch m := t.(type) {
	case TaskSetLogLevel:
		s.mu.Lock()
		s.level = m.Level
		s.mu.Unlock()
		return nil, true, nil
	}
	return nil, false, nil
}

// SetLevel changes the log level.
func (s *LogService) SetLevel(level LogLevel) {
	s.mu.Lock()
	s.level = level
	s.mu.Unlock()
}

// Level returns the current log level.
func (s *LogService) Level() LogLevel {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.level
}

func (s *LogService) shouldLog(level LogLevel) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return level <= s.level
}

func (s *LogService) log(level, prefix, msg string) {
	timestamp := time.Now().Format("15:04:05")
	fmt.Fprintf(s.output, "%s %s %s\n", DimStyle.Render(timestamp), prefix, msg)
}

// Debug logs a debug message.
func (s *LogService) Debug(msg string) {
	if s.shouldLog(LogLevelDebug) {
		s.log("debug", DimStyle.Render("[DBG]"), msg)
	}
}

// Infof logs an info message.
func (s *LogService) Infof(msg string) {
	if s.shouldLog(LogLevelInfo) {
		s.log("info", InfoStyle.Render("[INF]"), msg)
	}
}

// Warnf logs a warning message.
func (s *LogService) Warnf(msg string) {
	if s.shouldLog(LogLevelWarn) {
		s.log("warn", WarningStyle.Render("[WRN]"), msg)
	}
}

// Errorf logs an error message.
func (s *LogService) Errorf(msg string) {
	if s.shouldLog(LogLevelError) {
		s.log("error", ErrorStyle.Render("[ERR]"), msg)
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
		l.Infof(msg)
	}
}

// LogWarn logs a warning message if log service is available.
func LogWarn(msg string) {
	if l := Log(); l != nil {
		l.Warnf(msg)
	}
}

// LogError logs an error message if log service is available.
func LogError(msg string) {
	if l := Log(); l != nil {
		l.Errorf(msg)
	}
}
