//go:build !windows

package player

import (
	"context"
	"fmt"
	"github.com/PizzaHomicide/hisame/internal/log"
	"net"
	"os/exec"
	"syscall"
)

// setupPlayerProcess configures the process for detached execution
func setupPlayerProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// releasePlayerProcess handles post-start process management
func releasePlayerProcess(cmd *exec.Cmd) error {
	if cmd.Process != nil {
		return cmd.Process.Release()
	}
	return nil
}

// Connect establishes a connection with MPV for Unix systems
func (c *MPVIPCClient) Connect(ctx context.Context) error {
	// For Unix systems, use Unix domain socket
	log.Debug("Connecting to Unix socket", "path", c.socketPath)
	var d net.Dialer
	conn, err := d.DialContext(ctx, "unix", c.socketPath)
	if err != nil {
		return fmt.Errorf("failed to connect to MPV socket: %w", err)
	}

	c.conn = conn
	go c.readEvents()
	return nil
}
