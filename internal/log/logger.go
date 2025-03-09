package log

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

// Logger provides an interface into the underlying logging system for Hisame's purposes.
type Logger struct {
	logger *slog.Logger
	file   *os.File
}

// Config contains logging information used to set up the logging framework
type Config struct {
	// Log Level.  One of: debug, info, warn, error
	Level string
	// Path to the file to log into
	FilePath string
}

func New(config Config) (*Logger, error) {
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return nil, err
	}

	opts := &slog.HandlerOptions{
		Level: parseLogLevel(config.Level),
	}

	handler := slog.NewJSONHandler(file, opts)

	logger := &Logger{
		logger: slog.New(handler),
		file:   file,
	}

	return logger, nil
}

// Close the log file
func (l *Logger) Close() {
	err := l.file.Close()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error closing logger: %v\n", err)
	}
}

// Debug logs a message a debug Level
func (l *Logger) Debug(msg string, args ...any) {
	l.logger.Debug(msg, args...)
}

// Info logs a message at info Level
func (l *Logger) Info(msg string, args ...any) {
	l.logger.Info(msg, args...)
}

// Warn logs a message at info Level
func (l *Logger) Warn(msg string, args ...any) {
	l.logger.Warn(msg, args...)
}

// Error logs a message at error Level.
func (l *Logger) Error(msg string, args ...any) {
	l.logger.Error(msg, args...)
}

// FatalExit logs a message at error Level and then terminated the application.
func (l *Logger) FatalExit(msg string, err error, args ...any) {
	l.logger.Error(msg, args...)
	l.logger.Error("Fatal error.  Terminating application.")
	os.Exit(1)
}

// parseLogLevel is a helper to convert a string log Level into the slog version.  Defaults to info if a matching log
// Level cannot be found.
func parseLogLevel(lvl string) slog.Level {
	switch lvl {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
