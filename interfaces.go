package mongodb

// Logger defines the interface for pluggable logging in the MongoDB client.
// Users can implement this interface to integrate their preferred logging solution.
type Logger interface {
	// Info logs an informational message with optional structured fields
	Info(msg string, fields ...any)
	// Warn logs a warning message with optional structured fields
	Warn(msg string, fields ...any)
	// Error logs an error message with optional structured fields
	Error(msg string, fields ...any)
	// Debug logs a debug message with optional structured fields
	Debug(msg string, fields ...any)
}

// NopLogger is a no-operation logger that discards all log messages.
// This is used as the default logger when no logger is provided via WithLogger.
type NopLogger struct{}

// Info implements Logger.Info by doing nothing
func (NopLogger) Info(msg string, fields ...any) {}

// Warn implements Logger.Warn by doing nothing
func (NopLogger) Warn(msg string, fields ...any) {}

// Error implements Logger.Error by doing nothing
func (NopLogger) Error(msg string, fields ...any) {}

// Debug implements Logger.Debug by doing nothing
func (NopLogger) Debug(msg string, fields ...any) {}
