package electroncompat

// LoggerService provides structured logging operations.
// This corresponds to the Logger IPC service from the Electron app.
type LoggerService struct{}

// NewLoggerService creates a new LoggerService instance.
func NewLoggerService() *LoggerService {
	return &LoggerService{}
}

// Info logs an info message.
func (s *LoggerService) Info(message string, args ...any) error {
	return notImplemented("Logger", "info")
}

// Warn logs a warning message.
func (s *LoggerService) Warn(message string, args ...any) error {
	return notImplemented("Logger", "warn")
}

// Error logs an error message.
func (s *LoggerService) Error(message string, args ...any) error {
	return notImplemented("Logger", "error")
}

// Log logs a generic message.
func (s *LoggerService) Log(level, message string, args ...any) error {
	return notImplemented("Logger", "log")
}

// Download downloads log files.
func (s *LoggerService) Download() ([]byte, error) {
	return nil, notImplemented("Logger", "download")
}
