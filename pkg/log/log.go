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
}

// Options configures a Logger.
type Options struct {
	Level  Level
	Output io.Writer // defaults to os.Stderr
}

// New creates a new Logger with the given options.
func New(opts Options) *Logger {
	output := opts.Output
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

	// Automatically extract context from error if present in keyvals
	origLen := len(keyvals)
	for i := 0; i < origLen; i += 2 {
		if i+1 < origLen {
			if err, ok := keyvals[i+1].(error); ok {
				if op := Op(err); op != "" {
					// Check if op is already in keyvals
					hasOp := false
					for j := 0; j < len(keyvals); j += 2 {
						if keyvals[j] == "op" {
							hasOp = true
							break
						}
					}
					if !hasOp {
						keyvals = append(keyvals, "op", op)
					}
				}
				if stack := FormatStackTrace(err); stack != "" {
					// Check if stack is already in keyvals
					hasStack := false
					for j := 0; j < len(keyvals); j += 2 {
						if keyvals[j] == "stack" {
							hasStack = true
							break
						}
					}
					if !hasStack {
						keyvals = append(keyvals, "stack", stack)
					}
				}
			}
		}
	}

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
