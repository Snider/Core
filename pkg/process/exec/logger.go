package exec

// Logger interface for command execution logging.
// Compatible with pkg/log.Logger and other structured loggers.
type Logger interface {
	// Debug logs a debug-level message with optional key-value pairs.
	Debug(msg string, keyvals ...any)
	// Error logs an error-level message with optional key-value pairs.
	Error(msg string, keyvals ...any)
}

// NopLogger is a no-op logger that discards all messages.
type NopLogger struct{}

// Debug discards the message (no-op implementation).
func (NopLogger) Debug(string, ...any) {}

// Error discards the message (no-op implementation).
func (NopLogger) Error(string, ...any) {}

var defaultLogger Logger = NopLogger{}

// SetDefaultLogger sets the package-level default logger.
// Commands without an explicit logger will use this.
func SetDefaultLogger(l Logger) {
	if l == nil {
		l = NopLogger{}
	}
	defaultLogger = l
}

// DefaultLogger returns the current default logger.
func DefaultLogger() Logger {
	return defaultLogger
}
