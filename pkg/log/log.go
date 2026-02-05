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
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
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

func (l Level) slogLevel() slog.Level {
	switch l {
	case LevelDebug:
		return slog.LevelDebug
	case LevelInfo:
		return slog.LevelInfo
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	case LevelQuiet:
		return slog.Level(100)
	default:
		return slog.LevelInfo
	}
}

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

// LogFormat defines the output format.
type LogFormat int

const (
	// FormatText outputs human-readable text.
	FormatText LogFormat = iota
	// FormatJSON outputs structured JSON.
	FormatJSON
)

// Logger provides structured logging.
type Logger struct {
	mu     sync.RWMutex
	level  Level
	output io.Writer
	format LogFormat
	slog   *slog.Logger

	// Style functions for formatting (can be overridden)
	StyleTimestamp func(string) string
	StyleDebug     func(string) string
	StyleInfo      func(string) string
	StyleWarn      func(string) string
	StyleError     func(string) string
}

// Options configures a Logger.
type Options struct {
	Level Level

	// Format selects the log output format. Use FormatText for human-readable logs
	// (typically during local development) and FormatJSON for structured logs that
	// are easier to parse and aggregate in log collectors. When FormatJSON is used,
	// errors automatically include captured stack traces when available.
	Format LogFormat

	Output io.Writer // defaults to os.Stderr
}

// New creates a new Logger with the given options.
func New(opts Options) *Logger {
	output := opts.Output
	if output == nil {
		output = os.Stderr
	}

	l := &Logger{
		level:          opts.Level,
		output:         output,
		format:         opts.Format,
		StyleTimestamp: identity,
		StyleDebug:     identity,
		StyleInfo:      identity,
		StyleWarn:      identity,
		StyleError:     identity,
	}
	l.updateSlog()
	return l
}

func (l *Logger) updateSlog() {
	l.mu.Lock()
	defer l.mu.Unlock()

	var handler slog.Handler
	if l.format == FormatJSON {
		handler = slog.NewJSONHandler(l.output, &slog.HandlerOptions{
			Level: l.level.slogLevel(),
		})
	} else {
		handler = &textHandler{
			l:     l,
			level: l.level.slogLevel(),
		}
	}
	l.slog = slog.New(handler)
}

type textHandler struct {
	l           *Logger
	level       slog.Level
	attrs       []slog.Attr
	groupPrefix string
}

func (h *textHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *textHandler) Handle(ctx context.Context, r slog.Record) error {
	h.l.mu.RLock()
	output := h.l.output
	styleTimestamp := h.l.StyleTimestamp
	styleDebug := h.l.StyleDebug
	styleInfo := h.l.StyleInfo
	styleWarn := h.l.StyleWarn
	styleError := h.l.StyleError
	h.l.mu.RUnlock()

	timestamp := styleTimestamp(r.Time.Format("15:04:05"))

	var prefix string
	switch r.Level {
	case slog.LevelDebug:
		prefix = styleDebug("[DBG]")
	case slog.LevelInfo:
		prefix = styleInfo("[INF]")
	case slog.LevelWarn:
		prefix = styleWarn("[WRN]")
	case slog.LevelError:
		prefix = styleError("[ERR]")
	default:
		prefix = "[" + r.Level.String() + "]"
	}

	var kvStr strings.Builder
	for _, a := range h.attrs {
		kvStr.WriteString(fmt.Sprintf(" %s=%v", a.Key, a.Value.Any()))
	}
	r.Attrs(func(a slog.Attr) bool {
		kvStr.WriteString(fmt.Sprintf(" %s%s=%v", h.groupPrefix, a.Key, a.Value.Any()))
		return true
	})

	_, err := fmt.Fprintf(output, "%s %s %s%s\n", timestamp, prefix, r.Message, kvStr.String())
	return err
}

func (h *textHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newAttrs := make([]slog.Attr, len(h.attrs)+len(attrs))
	copy(newAttrs, h.attrs)
	for i, a := range attrs {
		newAttrs[len(h.attrs)+i] = slog.Attr{Key: h.groupPrefix + a.Key, Value: a.Value}
	}
	return &textHandler{
		l:           h.l,
		level:       h.level,
		attrs:       newAttrs,
		groupPrefix: h.groupPrefix,
	}
}

func (h *textHandler) WithGroup(name string) slog.Handler {
	return &textHandler{
		l:           h.l,
		level:       h.level,
		attrs:       h.attrs,
		groupPrefix: h.groupPrefix + name + ".",
	}
}

func identity(s string) string { return s }

// SetLevel changes the log level.
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
	l.updateSlog()
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
	l.updateSlog()
}

// Debug logs a debug message with optional key-value pairs.
func (l *Logger) Debug(msg string, keyvals ...any) {
	l.slog.Debug(msg, keyvals...)
}

// DebugContext logs a debug message with optional key-value pairs.
func (l *Logger) DebugContext(ctx context.Context, msg string, keyvals ...any) {
	l.slog.DebugContext(ctx, msg, keyvals...)
}

// Info logs an info message with optional key-value pairs.
func (l *Logger) Info(msg string, keyvals ...any) {
	l.slog.Info(msg, keyvals...)
}

// InfoContext logs an info message with optional key-value pairs.
func (l *Logger) InfoContext(ctx context.Context, msg string, keyvals ...any) {
	l.slog.InfoContext(ctx, msg, keyvals...)
}

// Warn logs a warning message with optional key-value pairs.
func (l *Logger) Warn(msg string, keyvals ...any) {
	l.slog.Warn(msg, keyvals...)
}

// WarnContext logs a warning message with optional key-value pairs.
func (l *Logger) WarnContext(ctx context.Context, msg string, keyvals ...any) {
	l.slog.WarnContext(ctx, msg, keyvals...)
}

// Error logs an error message with optional key-value pairs.
func (l *Logger) Error(msg string, keyvals ...any) {
	l.ErrorContext(context.Background(), msg, keyvals...)
}

// ErrorContext logs an error message with optional key-value pairs.
func (l *Logger) ErrorContext(ctx context.Context, msg string, keyvals ...any) {
	l.mu.RLock()
	format := l.format
	hndlr := l.slog.Handler()
	l.mu.RUnlock()

	// Add stack trace for errors in JSON mode
	if format == FormatJSON && hndlr.Enabled(ctx, slog.LevelError) {
		buf := stackPool.Get().([]byte)
		n := runtime.Stack(buf, false)
		keyvals = append(keyvals, slog.String("stack", string(buf[:n])))
		stackPool.Put(buf)
	}
	l.slog.ErrorContext(ctx, msg, keyvals...)
}

var stackPool = sync.Pool{
	New: func() any {
		return make([]byte, 1024*8)
	},
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

// DebugContext logs to the default logger.
func DebugContext(ctx context.Context, msg string, keyvals ...any) {
	defaultLogger.DebugContext(ctx, msg, keyvals...)
}

// Info logs to the default logger.
func Info(msg string, keyvals ...any) {
	defaultLogger.Info(msg, keyvals...)
}

// InfoContext logs to the default logger.
func InfoContext(ctx context.Context, msg string, keyvals ...any) {
	defaultLogger.InfoContext(ctx, msg, keyvals...)
}

// Warn logs to the default logger.
func Warn(msg string, keyvals ...any) {
	defaultLogger.Warn(msg, keyvals...)
}

// WarnContext logs to the default logger.
func WarnContext(ctx context.Context, msg string, keyvals ...any) {
	defaultLogger.WarnContext(ctx, msg, keyvals...)
}

// Error logs to the default logger.
func Error(msg string, keyvals ...any) {
	defaultLogger.Error(msg, keyvals...)
}

// ErrorContext logs to the default logger.
func ErrorContext(ctx context.Context, msg string, keyvals ...any) {
	defaultLogger.ErrorContext(ctx, msg, keyvals...)
}
