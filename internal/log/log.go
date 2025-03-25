package log

import "sync"

var (
	defaultLogger *Logger
	mu            sync.RWMutex
)

// SetDefaultLogger sets the default global logger that will be used if calling logging functions directly exported by this package
func SetDefaultLogger(logger *Logger) {
	mu.Lock()
	defaultLogger = logger
	mu.Unlock()
}

// DefaultLogger returns the current default logger
func DefaultLogger() *Logger {
	mu.RLock()
	defer mu.RUnlock()
	return defaultLogger
}

// Debug logs at debug Level using the default logger.
// See (*Logger).Debug for more information.
func Debug(msg string, args ...any) {
	if logger := DefaultLogger(); logger != nil {
		logger.Debug(msg, args...)
	}
}

// Info logs at info Level using the default logger.
// See (*Logger).Info for more information.
func Info(msg string, args ...any) {
	if logger := DefaultLogger(); logger != nil {
		logger.Info(msg, args...)
	}
}

// Warn logs at warn Level using the default logger.
// See (*Logger).Warn for more information.
func Warn(msg string, args ...any) {
	if logger := DefaultLogger(); logger != nil {
		logger.Warn(msg, args...)
	}
}

// Error logs at error Level using the default logger.
// See (*Logger).Error for more information.
func Error(msg string, args ...any) {
	if logger := DefaultLogger(); logger != nil {
		logger.Error(msg, args...)
	}
}

// Trace logs at debug level, but only if trace logging is enabled.
// This is a 'fake' trace level.
func Trace(msg string, args ...any) {
	if logger := DefaultLogger(); logger != nil && logger.traceEnabled {
		logger.Debug("TRACE: "+msg, args...)
	}
}
