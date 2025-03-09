package log

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func TestLogging(t *testing.T) {
	// Create temp directory for test
	tempDir, err := os.MkdirTemp("", "hisame-log-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create log file path
	logPath := filepath.Join(tempDir, "test.log")

	// Create logger with debug level
	logger, err := New(Config{
		Level:    "debug",
		FilePath: logPath,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	SetDefaultLogger(logger)

	// Log test messages at different levels
	Debug("Debug message", "test", true)
	Info("Info message", "test", true)
	Warn("Warning message", "test", true)
	Error("Error message", "error", fmt.Errorf("test error"))

	// Close logger to ensure file is written
	logger.Close()

	// Read log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Basic verification
	contentStr := string(content)
	assert.Contains(t, contentStr, "Debug message")
	assert.Contains(t, contentStr, "Info message")
	assert.Contains(t, contentStr, "Warning message")
	assert.Contains(t, contentStr, "Error message")
	assert.Contains(t, contentStr, "test error")
}
