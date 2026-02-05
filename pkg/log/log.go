// Package log provides structured logging for Core applications.
//
// The package works standalone or integrated with the Core framework:
//
//	// Standalone usage
//	log.SetLevel(log.LevelDebug)
//	log.Info("server started", "port", 8080)
//	log.Error("failed to connect", "err", err)
//
//	// With Core framework
//	core.New(
//	    framework.WithName("log", log.NewService(log.Options{Level: log.LevelInfo})),
//	)
package log

import (
	"fmt"
	"io"
	"os"
	"os/user"
	"sync"
	"time"
)

// Level defines logging verbosity.
type Level int

// Logging level constants ordered by increasing verbosity.
const (
	// LevelQuiet suppresses all log output.
	LevelQuiet Level = iota
	// LevelError shows only error messages.
	LevelError
	// LevelWarn shows warnings and errors.
	LevelWarn
	// LevelInfo shows informational messages, warnings, and errors.
	LevelInfo
	// LevelDebug shows all messages including debug details.
	LevelDebug
)

// String returns the level name.
func (l Level) String() string {
	switch l {
	case LevelQuiet:
		return "quiet"
	case LevelError:
		return "error"
	case LevelWarn:
		return "warn"
	case LevelInfo:
		return "info"
	case LevelDebug:
		return "debug"
	default:
		return "unknown"
	}
}

// Logger provides structured logging.
type Logger struct {
	mu     sync.RWMutex
	level  Level
	output io.Writer

	// Style functions for formatting (can be overridden)
	StyleTimestamp func(string) string
	StyleDebug     func(string) string
	StyleInfo      func(string) string
	StyleWarn      func(string) string
	StyleError     func(string) string
	StyleSecurity  func(string) string
}

// RotationOptions defines the log rotation and retention policy.
type RotationOptions struct {
	// Filename is the log file path. If empty, rotation is disabled.
	Filename string

	// MaxSize is the maximum size of the log file in megabytes before it gets rotated.
	// It defaults to 100 megabytes.
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on their
	// file modification time. It defaults to 28 days.
	// Note: set to a negative value to disable age-based retention.
	MaxAge int

	// MaxBackups is the maximum number of old log files to retain.
	// It defaults to 5 backups.
	MaxBackups int

	// Compress determines if the rotated log files should be compressed using gzip.
	// It defaults to true.
	Compress bool
}

// Options configures a Logger.
type Options struct {
	Level Level
	// Output is the destination for log messages. If Rotation is provided,
	// Output is ignored and logs are written to the rotating file instead.
	Output io.Writer
	// Rotation enables log rotation to file. If provided, Filename must be set.
	Rotation *RotationOptions
}

// New creates a new Logger with the given options.
func New(opts Options) *Logger {
	output := opts.Output
	if opts.Rotation != nil && opts.Rotation.Filename != "" {
		output = NewRotatingWriter(*opts.Rotation, nil)
	}
	if output == nil {
		output = os.Stderr
	}

	return &Logger{
		level:          opts.Level,
		output:         output,
		StyleTimestamp: identity,
		StyleDebug:     identity,
		StyleInfo:      identity,
		StyleWarn:      identity,
		StyleError:     identity,
		StyleSecurity:  identity,
	}
}

func identity(s string) string { return s }

// SetLevel changes the log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// Level returns the current log level.
func (l *Logger) Level() Level {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.level
}

// SetOutput changes the output writer.
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	l.output = w
	l.mu.Unlock()
}

func (l *Logger) shouldLog(level Level) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return level <= l.level
}

func (l *Logger) log(level Level, prefix, msg string, keyvals ...any) {
	l.mu.RLock()
	output := l.output
	styleTimestamp := l.StyleTimestamp
	l.mu.RUnlock()

	timestamp := styleTimestamp(time.Now().Format("15:04:05"))

	// Format key-value pairs
	var kvStr string
	if len(keyvals) > 0 {
		kvStr = " "
		for i := 0; i < len(keyvals); i += 2 {
			if i > 0 {
				kvStr += " "
			}
			key := keyvals[i]
			var val any
			if i+1 < len(keyvals) {
				val = keyvals[i+1]
			}
			kvStr += fmt.Sprintf("%v=%v", key, val)
		}
	}

	_, _ = fmt.Fprintf(output, "%s %s %s%s\n", timestamp, prefix, msg, kvStr)
}

// Debug logs a debug message with optional key-value pairs.
func (l *Logger) Debug(msg string, keyvals ...any) {
	if l.shouldLog(LevelDebug) {
		l.log(LevelDebug, l.StyleDebug("[DBG]"), msg, keyvals...)
	}
}

// Info logs an info message with optional key-value pairs.
func (l *Logger) Info(msg string, keyvals ...any) {
	if l.shouldLog(LevelInfo) {
		l.log(LevelInfo, l.StyleInfo("[INF]"), msg, keyvals...)
	}
}

// Warn logs a warning message with optional key-value pairs.
func (l *Logger) Warn(msg string, keyvals ...any) {
	if l.shouldLog(LevelWarn) {
		l.log(LevelWarn, l.StyleWarn("[WRN]"), msg, keyvals...)
	}
}

// Error logs an error message with optional key-value pairs.
func (l *Logger) Error(msg string, keyvals ...any) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, l.StyleError("[ERR]"), msg, keyvals...)
	}
}

// Security logs a security event with optional key-value pairs.
// It uses LevelError to ensure security events are visible even in restrictive
// log configurations.
func (l *Logger) Security(msg string, keyvals ...any) {
	if l.shouldLog(LevelError) {
		l.log(LevelError, l.StyleSecurity("[SEC]"), msg, keyvals...)
	}
}

// Username returns the current system username.
// It uses os/user for reliability and falls back to environment variables.
func Username() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	// Fallback for environments where user lookup might fail
	if u := os.Getenv("USER"); u != "" {
		return u
	}
	return os.Getenv("USERNAME")
}

// --- Default logger ---

var defaultLogger = New(Options{Level: LevelInfo})

// Default returns the default logger.
func Default() *Logger {
	return defaultLogger
}

// SetDefault sets the default logger.
func SetDefault(l *Logger) {
	defaultLogger = l
}

// SetLevel sets the default logger's level.
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// Debug logs to the default logger.
func Debug(msg string, keyvals ...any) {
	defaultLogger.Debug(msg, keyvals...)
}

// Info logs to the default logger.
func Info(msg string, keyvals ...any) {
	defaultLogger.Info(msg, keyvals...)
}

// Warn logs to the default logger.
func Warn(msg string, keyvals ...any) {
	defaultLogger.Warn(msg, keyvals...)
}

// Error logs to the default logger.
func Error(msg string, keyvals ...any) {
	defaultLogger.Error(msg, keyvals...)
}

// Security logs to the default logger.
func Security(msg string, keyvals ...any) {
	defaultLogger.Security(msg, keyvals...)
}
