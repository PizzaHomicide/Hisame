//go:build windows

package player

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/log"
	"gopkg.in/natefinch/npipe.v2"
	"os/exec"
	"syscall"
)

// setupPlayerProcess configures the process for detached execution
func setupPlayerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// releasePlayerProcess handles post-start process management
func releasePlayerProcess(cmd *exec.Cmd) error {
	// Windows doesn't need explicit process release
	return nil
}

// Connect establishes a connection with MPV for Windows
func (c *MPVIPCClient) Connect(ctx context.Context) error {
	log.Debug("Connecting to Windows named pipe", "path", c.socketPath)

	// Connect using the Windows-specific named pipe package
	conn, err := npipe.Dial(c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to MPV pipe: %w", err)
	}

	c.conn = conn
	go c.readEvents()
	return nil
}
